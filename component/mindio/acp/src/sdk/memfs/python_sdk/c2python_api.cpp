/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#include <fcntl.h>
#include <unistd.h>
#include <sys/stat.h>
#include <algorithm>
#include <sstream>

#include "file_check_utils.h"
#include "common_includes.h"
#include "memfs_sdk_api.h"
#include "c2python_api.h"

using namespace ock::common;

namespace ock {
namespace memfs {
namespace sdk {
constexpr int DICT_MAX_LEN = 100;
constexpr int DICT_VALUE_MAX_COUNT = 1000;

static CallbackReceiver *g_callbackHandler = nullptr;
NdsFileDriver *NdsReadableFile::g_ndsDriver = NdsFileDriver::Instance();

static bool CheckFsFdValid(int32_t fd) noexcept
{
    int32_t ret = MemFsCntl(fd, F_GETFL);
    if (ret == -1) {
        return false;
    }
    return true;
}

static int CloseMemFsFd(int &fd) noexcept
{
    if (fd < 0) {
        return 0;
    }
    auto ret = MemFsClose(fd);
    if (ret != 0) {
        LOG_ERROR("memfs close failed: " << errno << ":" << strerror(errno));
        return -1;
    }

    fd = -1;
    return ret;
}

static int CloseSystemFd(FILE *fp) noexcept
{
    if (fp == nullptr) {
        return 0;
    }
    if (fclose(fp) != 0) {
        LOG_ERROR("fclose failed: " << errno << ":" << strerror(errno));
        return -1;
    }
    fp = nullptr;
    return 0;
}

CloseableFile::CloseableFile(std::string path) noexcept : filePath{ std::move(path) }, fd{ -1 } {}

CloseableFile::~CloseableFile() noexcept = default;

Readable::~Readable() noexcept = default;

Writeable::Writeable(std::string path, mode_t mode) noexcept : CloseableFile{ std::move(path) }, fileMode{ mode } {}

Writeable::~Writeable() noexcept = default;

int WriteableFile::Drop() noexcept
{
    auto ret = MemFsCloseWithUnlink(fd);
    if (ret != 0) {
        LOG_ERROR("MemFs client close with unlink failed: " << errno << ":" << strerror(errno));
        return -1;
    }
    fd = -1;
    return ret;
}

int WriteableFile::Flush() noexcept
{
    auto ret = MemFsFlush(fd);
    if (ret != 0) {
        LOG_ERROR("MemFs client flush data failed: " << errno << ":" << strerror(errno));
        return -1;
    }
    return ret;
}

int ReadableFile::Initialize() noexcept
{
    if (fd >= 0) {
        return 0;
    }
    if (MemFsIsForkedProcess()) {
        LOG_ERROR("MemFs client is forked process, not support.");
        return -1;
    }
    auto result = MemFsOpenFile(filePath.c_str(), O_RDONLY);
    if (result < 0) {
        LOG_ERROR("MemFs client open file failed: " << errno << ":" << strerror(errno));
        return -1;
    }

    fd = result;
    fileSize = MemFsGetSize(fd);
    if (fileSize == -1) {
        return -1;
    }
    return 0;
}

int ReadableFile::Close() noexcept
{
    return CloseMemFsFd(fd);
}

ssize_t ReadableFile::Read(void *buffer, size_t count, off_t offset) noexcept
{
    if (!CheckFsFdValid(fd)) {
        return -1;
    }
    if (MemFsIsForkedProcess()) {
        LOG_ERROR("MemFs client is forked process, not support.");
        return -1;
    }
    auto retSize = MemFsRead(fd, reinterpret_cast<uintptr_t>(buffer), offset, count);
    if (retSize == -1) {
        LOG_ERROR("memfs read@" << FileCheckUtils::RemovePrefixPath(filePath) << " failed: " << errno << ":" <<
            strerror(errno));
        return -1;
    }

    LOG_DEBUG("ReadableFile read success");
    return retSize;
}

ssize_t ReadableFile::ReadV(std::vector<ock::memfs::ReadBuffer> readVector) noexcept
{
    if (!CheckFsFdValid(fd)) {
        return -1;
    }

    if (MemFsIsForkedProcess()) {
        LOG_ERROR("MemFs client is forked process, not support.");
        return -1;
    }

    ssize_t retSize = MemFsReadM(fd, readVector);
    if (retSize == -1) {
        LOG_ERROR("MemFs client multi read@" << FileCheckUtils::RemovePrefixPath(filePath) << " failed: " << errno <<
            ":" << strerror(errno));
        return -1;
    }

    LOG_DEBUG("ReadableFile readv success");
    return 0;
}

int NdsReadableFile::Initialize() noexcept
{
    if (fd >= 0) {
        return 0;
    }

    if (g_ndsDriver == nullptr) {
        LOG_DEBUG("[nds] Nds driver instance is null.");
        return -1;
    }

    if (!g_ndsDriver->NdsAvailable()) {
        LOG_DEBUG("[nds] Nds file driver is not available");
        return -1;
    }

    fd = open(filePath.c_str(), O_RDONLY | O_DIRECT);
    if (fd < 0) {
        LOG_DEBUG("[nds] open file read failed:" << errno << ":" << strerror(errno));
        return -1;
    }

    NdsFileError_t status = g_ndsDriver->RegisterHandler(fd);
    if (status.err != NDS_FILE_SUCCESS) {
        LOG_ERROR("[nds] Register handler fd: " << fd << " error:" << status.err);
        return -1;
    }

    struct stat fileStat {};
    if (fstat(fd, &fileStat) != 0) {
        LOG_ERROR("[nds] Get file stat failed:" << errno << ":" << strerror(errno));
        return -1;
    }
    fileSize = fileStat.st_size;

    LOG_DEBUG("NdsReadableFile with type rb success");
    return 0;
}

int NdsReadableFile::Close() noexcept
{
    if (g_ndsDriver == nullptr) {
        LOG_DEBUG("[nds] Nds driver instance is null.");
        return -1;
    }

    if (!g_ndsDriver->NdsAvailable()) {
        LOG_DEBUG("[nds] Nds file driver is not available");
        return -1;
    }

    g_ndsDriver->NdsFileHandleDeregister(fd);
    close(fd);
    fd = -1;
    LOG_DEBUG("[nds] Fclose and deregister fd success");
    return 0;
}

ssize_t NdsReadableFile::Read(void *buffer, size_t count, off_t offset) noexcept
{
    if (g_ndsDriver == nullptr) {
        LOG_DEBUG("[nds] Nds driver instance is null.");
        return -1;
    }

    if (!g_ndsDriver->NdsAvailable()) {
        LOG_DEBUG("[nds] Nds file driver is not available");
        return -1;
    }

    if (g_ndsDriver->NdsRead(fd, reinterpret_cast<char *>(buffer), offset, count) != 0) {
        LOG_ERROR("[nds] file read failed:" << errno << ":" << strerror(errno));
        return -1;
    }

    LOG_DEBUG("[nds] FileRead success");
    return 0;
}

ssize_t NdsReadableFile::ReadV(std::vector<ock::memfs::ReadBuffer> readVector) noexcept
{
    if (g_ndsDriver == nullptr) {
        LOG_DEBUG("[nds] Nds driver instance is null.");
        return -1;
    }

    if (!g_ndsDriver->NdsAvailable()) {
        LOG_DEBUG("[nds] Nds file driver is not available");
        return -1;
    }

    for (const auto vector : readVector) {
        if (g_ndsDriver->NdsRead(fd, vector.buffer, vector.start, vector.size) != 0) {
            LOG_ERROR("[nds] file read failed:" << errno << ":" << strerror(errno));
            return -1;
        }
    }

    LOG_DEBUG("[nds] Fread success");
    return 0;
}

int FReadableFile::Initialize() noexcept
{
    if (fp != nullptr) {
        return 0;
    }

    struct stat statBuf {};
    auto ret = stat(filePath.c_str(), &statBuf);
    if (ret < 0) {
        return -1;
    }

    fp = fopen(filePath.c_str(), "rb");
    if (fp == nullptr) {
        return -1;
    }

    fileSize = statBuf.st_size;

    LOG_DEBUG("FReadableFile with type rb success");
    return 0;
}

int FReadableFile::Close() noexcept
{
    return CloseSystemFd(fp);
}

ssize_t FReadableFile::Read(void *buffer, size_t count, off_t offset) noexcept
{
    if (fp == nullptr) {
        LOG_ERROR("file not initialized");
        return -1;
    }
    auto ret = fseek(fp, offset, SEEK_SET);
    if (ret != 0) {
        LOG_ERROR("Stdio seek failed.");
        return -1;
    }
    auto result = fread(buffer, count, 1, fp);
    if (result != 1) {
        LOG_ERROR("fread@" << FileCheckUtils::RemovePrefixPath(filePath) << " failed: " << errno << ":" <<
            strerror(errno));
        return -1;
    }
    LOG_DEBUG("fread success");
    return static_cast<ssize_t>(result);
}

ssize_t FReadableFile::ReadV(std::vector<ock::memfs::ReadBuffer> readVector) noexcept
{
    if (fp == nullptr) {
        LOG_ERROR("file not initialized");
        return -1;
    }
    for (const auto vector : readVector) {
        auto ret = fseek(fp, static_cast<long>(vector.start), SEEK_SET);
        if (ret != 0) {
            LOG_ERROR("Stdio seek failed.");
            return -1;
        }
        auto readSize = fread(reinterpret_cast<void *>(vector.buffer), vector.size, 1, fp);
        if (readSize != 1) {
            LOG_ERROR("Stdio read failed.");
            return -1;
        }
    }
    LOG_DEBUG("Fread success");
    return 0;
}

int WriteableFile::Initialize() noexcept
{
    if (fd >= 0) {
        return 0;
    }
    if (MemFsIsForkedProcess()) {
        LOG_ERROR("MemFs client is forked process, not support.");
        return -1;
    }
    auto result = MemFsOpenFile(filePath.c_str(), O_WRONLY | O_CREAT | O_TRUNC);
    if (result < 0) {
        LOG_ERROR("MemFs open file@" << FileCheckUtils::RemovePrefixPath(filePath) <<
            " with type write failed error: " << errno << ":" << strerror(errno));
        return -1;
    }
    fd = result;
    LOG_DEBUG("MemFsOpenFile with type wb success");
    return 0;
}

int WriteableFile::Close() noexcept
{
    return CloseMemFsFd(fd);
}

ssize_t WriteableFile::Write(const void *buffer, size_t count) noexcept
{
    if (!CheckFsFdValid(fd)) {
        return -1;
    }
    if (MemFsIsForkedProcess()) {
        LOG_ERROR("MemFs client is forked process, not support.");
        return -1;
    }
    auto result = MemFsWrite(fd, reinterpret_cast<uintptr_t>(buffer), count);
    if (result == -1) {
        LOG_ERROR("memfs write@" << FileCheckUtils::RemovePrefixPath(filePath) << " failed: " << errno << ":" <<
            strerror(errno));
        return -1;
    }

    LOG_DEBUG("MemFsWrite success");
    return 0;
}

ssize_t WriteableFile::WriteV(std::vector<ock::memfs::Buffer> buffers) noexcept
{
    if (!CheckFsFdValid(fd)) {
        LOG_ERROR("The fd [" << fd << "] is wrong.");
        return -1;
    }

    if (MemFsIsForkedProcess()) {
        LOG_ERROR("MemFs client is forked process, not support.");
        return -1;
    }

    auto outSize = MemFsWriteV(fd, buffers);
    if (outSize == -1) {
        LOG_ERROR("MemFs write failed, the fd is " << fd);
        (void)MemFsCloseWithUnlink(fd);
        return -1;
    }

    LOG_DEBUG("MemFsWriteV success");
    return 0;
}

int FWriteableFile::Initialize() noexcept
{
    if (fp != nullptr) {
        return 0;
    }
    std::string stagePath = std::string(filePath).append(".m.stg");
    auto stgFd = open(stagePath.c_str(), O_CREAT | O_WRONLY | O_TRUNC, S_IRUSR | S_IWUSR);
    if (stgFd < 0) {
        LOG_ERROR("create stage file(" << FileCheckUtils::RemovePrefixPath(stagePath) << ") failed: " << errno <<
            " : " << strerror(errno));
        return -1;
    }
    close(stgFd);
    fp = fopen(filePath.c_str(), "wb");
    if (fp == nullptr) {
        LOG_ERROR("open file@" << FileCheckUtils::RemovePrefixPath(filePath) << " failed: " << errno << ":" <<
            strerror(errno));
        return -1;
    }

    LOG_DEBUG("FWriteableFile with type wb success");
    return 0;
}

int FWriteableFile::Close() noexcept
{
    auto ret = CloseSystemFd(fp);
    return RemoveStgFile() ? ret : -1;
}

int FWriteableFile::Flush() noexcept
{
    return 0;
}

int FWriteableFile::Drop() noexcept
{
    if (remove(filePath.c_str()) != 0) {
        LOG_ERROR("system drop file " << FileCheckUtils::RemovePrefixPath(filePath) << " failed: " << errno << ":" <<
            strerror(errno));
        return -1;
    }
    return RemoveStgFile() ? 0 : -1;
}

bool FWriteableFile::RemoveStgFile() noexcept
{
    std::string stagePath = std::string(filePath).append(".m.stg");
    if (!FileUtil::Remove(stagePath) && errno != ENOENT) {
        LOG_ERROR("Remove stage file " << FileCheckUtils::RemovePrefixPath(stagePath) << " failed: " << errno <<
            " : " << strerror(errno));
        return false;
    }
    return true;
}

ssize_t FWriteableFile::Write(const void *buffer, size_t count) noexcept
{
    if (fp == nullptr) {
        LOG_ERROR("file not initialized");
        return -1;
    }
    auto result = fwrite(buffer, count, 1, fp);
    if (result != 1) {
        LOG_ERROR("fwrite@" << FileCheckUtils::RemovePrefixPath(filePath) << " failed: " << errno << ":" <<
            strerror(errno));
        return -1;
    }

    LOG_DEBUG("FWriteableFile write success");
    return static_cast<ssize_t>(result);
}

ssize_t FWriteableFile::WriteV(std::vector<ock::memfs::Buffer> buffers) noexcept
{
    if (fp == nullptr) {
        LOG_ERROR("file not initialized");
        return -1;
    }
    for (auto &buffer : buffers) {
        auto tempRes = fwrite(buffer.buffer, buffer.size, 1, fp);
        if (tempRes != 1) {
            LOG_ERROR("fwrite@" << FileCheckUtils::RemovePrefixPath(filePath) << " failed: " << errno << ":" <<
                strerror(errno));
            return -1;
        }
    }

    LOG_DEBUG("FWriteableFile writev success");
    return 0;
}

int LinkFile(const std::string &source, const std::string &target) noexcept
{
    if (MemFsIsForkedProcess()) {
        LOG_ERROR("MemFs client is forked process, not support.");
        return -1;
    }

    if (MemFsLinkFile(source.c_str(), target.c_str()) != 0) {
        LOG_ERROR("Link source path to destination path failed.");
        return -1;
    }

    LOG_DEBUG("Link source path to destination path success");
    return 0;
}

int PreloadFile(const std::vector<std::string> &paths) noexcept
{
    if (MemFsIsForkedProcess() != 0) {
        LOG_ERROR("MemFs client is forked process, not support.");
        return -1;
    }

    bool allSuccess = true;
    for (const auto &path : paths) {
        int result = MemfsPreloadFile(path.c_str());
        if (result < 0) {
            LOG_WARN("MemFs preload file(" << FileCheckUtils::RemovePrefixPath(path) << ") failed:" << errno << ":" <<
                strerror(errno) << ", preload next, skip.");
            allSuccess = false;
        }
    }

    if (!allSuccess) {
        return -1;
    }

    LOG_DEBUG("MemFsPreload task submit success");
    return 0;
}

int Initialize(const std::map<std::string, std::string> &serverInfoParam) noexcept
{
    return MemFsClientInitialize(serverInfoParam);
}

void UnInitialize() noexcept
{
    MemFsClientUnInitialize();
}

void InitCallbackHandler(CallbackReceiver *handler)
{
    if (handler == nullptr) {
        LOG_ERROR("init callback handler failed, input param is nullptr.");
        return;
    }
    g_callbackHandler = handler;
}

void HandleCheckDirCallback(int32_t result, PyObject *callbackParam)
{
    if (g_callbackHandler == nullptr) {
        LOG_ERROR("callback handler is null");
        return;
    }
    g_callbackHandler->CheckDirCallback(result, callbackParam);
}

int RegisterChecker(const std::unordered_map<std::string, uint64_t> &dirInfo, PyObject *callbackParam,
    uint64_t timeoutSec)
{
    if (g_callbackHandler == nullptr) {
        LOG_ERROR("callback handler is null");
        return -1;
    }

    if (dirInfo.size() > DICT_MAX_LEN) {
        LOG_ERROR("The check dictionary length [" << dirInfo.size() << "] is not correct.");
        return -1;
    }

    for (auto &item : dirInfo) {
        if (item.second < 1 || item.second > DICT_VALUE_MAX_COUNT) {
            LOG_ERROR("The check dictionary value [" << item.second << "] is not correct.");
            return -1;
        }
    }

    uint64_t eventId = 0;
    CheckDirCallback callback = &HandleCheckDirCallback;
    auto innerCallback = [callback, callbackParam](uint64_t eventId, int32_t result) {
        LOG_INFO("Call callback function, result: [" << result << "], eventId: [" << eventId << "].");
        callback(result, callbackParam);
    };

    int result = MemfsRegisterWatchDir(dirInfo, timeoutSec, innerCallback, eventId);
    if (result != 0) {
        LOG_ERROR("MemFs register watch dir failed.");
        return -1;
    }
    return 0;
}

int CheckBackgroundTask() noexcept
{
    if (MemFsIsForkedProcess()) {
        LOG_ERROR("MemFs client is forked process, not support.");
        return -1;
    }

    int result = MemFsCheckBackgroundTask();
    if (result < 0) {
        LOG_ERROR("MemFs check background task failed:" << errno << ":" << strerror(errno));
        return result;
    }

    LOG_INFO("MemFs all background tasks are finished.");
    return 0;
}
} // namespace sdk
} // namespace memfs
} // namespace ock