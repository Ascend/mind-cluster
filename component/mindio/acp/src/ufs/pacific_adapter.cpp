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
#include <sys/stat.h>
#include <sys/xattr.h>
#include <fcntl.h>
#include <unistd.h>
#include <libgen.h>

#include <aio.h>
#include <algorithm>
#include <chrono>
#include <climits>
#include <cstdlib>
#include <thread>

#include "file_utils.h"
#include "mem_fs_constants.h"
#include "ufs_log.h"
#include "pacific_adapter.h"

using namespace ock::ufs;

static constexpr auto MAX_ITEM_ONE_PAGE = 512U;
static constexpr auto IO_BUF_SIZE = 16U * 1024U;
static constexpr auto SMALL_FILE_MAX_SIZE = 1024L * 1024L;
static constexpr auto HEALTHY_CHECK_BUF_SIZE = 16U * 1024U;
static constexpr auto HEALTHY_CHECK_DIR = "/.ock_fs_healthy_check";
static constexpr auto DEFAULT_MODE = 0600;

static inline bool ShouldSkipName(const std::string &name) noexcept
{
    return (name == "." || name == "..");
}

static bool PrepareHealthyCheckDir(const std::string &fullPath) noexcept
{
    auto dir = opendir(fullPath.c_str());
    if (dir != nullptr) {
        closedir(dir);
        return true;
    }

    if (errno != ENOENT) {
        UFS_LOG_ERROR("open healthy dir failed(" << errno << " : " << strerror(errno) << ")");
        return false;
    }

    auto ret = mkdir(fullPath.c_str(), 0700);
    if (ret == 0) {
        return true;
    }

    if (errno != EEXIST) {
        UFS_LOG_ERROR("mkdir healthy dir failed(" << errno << " : " << strerror(errno) << ")");
        return false;
    }

    return true;
}

static void GenerateRandomBuffer(utils::ByteBuffer &buffer) noexcept
{
    auto fd = open("/dev/urandom", O_RDONLY);
    if (fd >= 0) {
        read(fd, buffer.Data(), buffer.Capacity());
        close(fd);
    } else {
        UFS_LOG_WARN("open urandom device for read failed : " << errno << " : " << strerror(errno));
    }
}

static bool WriteCheckerFile(const std::string &path, utils::ByteBuffer &buffer) noexcept
{
    auto fd = open(path.c_str(), O_CREAT | O_WRONLY | O_TRUNC | O_NOFOLLOW, 0600);
    if (fd < 0) {
        UFS_LOG_ERROR("open healthy file to write failed(" << errno << " : " << strerror(errno) << ")");
        return false;
    }

    auto count = utils::FileUtils::WriteFull(fd, buffer.Data(), buffer.Capacity());
    if (count < 0) {
        UFS_LOG_ERROR("write healthy file failed(" << errno << " : " << strerror(errno) << ")");
        close(fd);
        unlink(path.c_str());
        return false;
    }

    if (static_cast<uint32_t>(count) != buffer.Capacity()) {
        UFS_LOG_ERROR("write healthy file return bytes failed(" << count << " vs. " << buffer.Capacity() << ")");
        close(fd);
        unlink(path.c_str());
        return false;
    }

    close(fd);
    return true;
}

static bool ReadCheckerFile(const std::string &path, utils::ByteBuffer &buffer) noexcept
{
    auto fd = open(path.c_str(), O_RDONLY | O_NOFOLLOW);
    if (fd < 0) {
        UFS_LOG_ERROR("open healthy file to read failed(" << errno << " : " << strerror(errno) << ")");
        return false;
    }

    auto count = utils::FileUtils::ReadFull(fd, buffer.Data(), buffer.Capacity());
    if (count < 0) {
        UFS_LOG_ERROR("read healthy file failed(" << errno << " : " << strerror(errno) << ")");
        close(fd);
        return false;
    }

    if (static_cast<uint64_t>(count) != buffer.Capacity()) {
        UFS_LOG_ERROR("read healthy file return bytes failed(" << count << " vs. " << buffer.Capacity() << ")");
        close(fd);
        return false;
    }

    close(fd);
    return true;
}

PacificAdapter::PacificAdapter(std::string root) noexcept : mountPath{ std::move(root) } {}

PacificAdapter::~PacificAdapter() noexcept
{
    mountPath.clear();
}

bool PacificAdapter::HealthyCheck() noexcept
{
    std::string rootPath = mountPath + HEALTHY_CHECK_DIR;
    if (!PrepareHealthyCheckDir(rootPath)) {
        return false;
    }

    utils::ByteBuffer writeBuffer(HEALTHY_CHECK_BUF_SIZE);
    if (!writeBuffer.Valid()) {
        return false;
    }
    GenerateRandomBuffer(writeBuffer);

    auto fileName = rootPath + "/" + std::to_string(std::chrono::system_clock::now().time_since_epoch().count());
    if (!WriteCheckerFile(fileName, writeBuffer)) {
        return false;
    }

    utils::ByteBuffer readBuffer(HEALTHY_CHECK_BUF_SIZE);
    if (!readBuffer.Valid()) {
        return false;
    }

    if (!ReadCheckerFile(fileName, readBuffer)) {
        unlink(fileName.c_str());
        return false;
    }

    if (memcmp(readBuffer.Data(), writeBuffer.Data(), HEALTHY_CHECK_BUF_SIZE) != 0) {
        UFS_LOG_ERROR("healthy check, read write data do not match");
        unlink(fileName.c_str());
        return false;
    }

    unlink(fileName.c_str());
    return true;
}

int PacificAdapter::PutFile(const std::string &path, int flags, const FileMode &mode,
    utils::ByteBuffer &dataBuffer) noexcept
{
    auto len = dataBuffer.Capacity() - dataBuffer.Offset();
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }

    auto realFlag = SupportedFlags(flags);
    UFS_LOG_DEBUG("put file " << path << " with flags(" << flags << ") mode(0o" << std::oct << mode.mode << ") size(" <<
        std::dec << len << ") start");
    auto fd = open(fullPath.c_str(), realFlag | O_WRONLY | O_TRUNC | O_NOFOLLOW, static_cast<mode_t>(mode.mode) & 0777);
    if (fd < 0) {
        UFS_LOG_ERROR("open file " << path << " with flags" << flags << " to write failed" << errno << " : " <<
            strerror(errno));
        return -1;
    }

    auto ret = WriteSmallFile(fd, fullPath, mode, dataBuffer);
    close(fd);

    UFS_LOG_DEBUG("put file(" << path << ") with flags(0x" << flags << ") size(" << dataBuffer.Capacity() <<
        ") success");
    return ret;
}

int PacificAdapter::PutFile(const std::string &path, const FileMode &mode, utils::ByteBuffer &dataBuffer) noexcept
{
    return PutFile(path, O_CREAT, mode, dataBuffer);
}

int PacificAdapter::PutFile(const std::string &path, int flags, const FileMode &mode, InputStream &inputStream) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }

    auto realFlag = SupportedFlags(flags);
    UFS_LOG_DEBUG("put file(" << path << ") with mode(0o" << std::oct << mode.mode << ") input stream size(" <<
        std::dec << inputStream.TotalSize() << ") start");
    auto fd = open(fullPath.c_str(), realFlag | O_WRONLY | O_TRUNC | O_NOFOLLOW, static_cast<mode_t>(mode.mode) & 0777);
    if (fd < 0) {
        UFS_LOG_ERROR("open file(" << path << ") to write failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    if (CorrectFileMetadata(fd, mode) != 0) {
        UFS_LOG_ERROR("set owner for file(" << path << ") failed(" << errno << " : " << strerror(errno) << ")");
        close(fd);
        return -1;
    }

    utils::ByteBuffer buf{ IO_BUF_SIZE };
    if (!buf.Valid()) {
        close(fd);
        return -1;
    }

    int64_t count;
    while ((count = inputStream.Read(buf)) > 0L) {
        auto bytes = utils::FileUtils::WriteFull(fd, buf.Data(), count);
        if (bytes < 0) {
            UFS_LOG_ERROR("open file(" << path << ") to write failed(" << errno << " : " << strerror(errno) << ")");
            close(fd);
            return -1;
        }

        if (bytes != count) {
            UFS_LOG_ERROR("write file(" << path << ") want bytes(" << count << ") but real(" << bytes << ")");
            close(fd);
            return -1;
        }

        buf.Offset(0U);
    }

    close(fd);

    UFS_LOG_DEBUG("put file(" << path << ") with input stream size(" << inputStream.TotalSize() << ") success");
    return 0;
}

int PacificAdapter::PutFile(const std::string &path, const FileMode &mode, InputStream &inputStream) noexcept
{
    return PutFile(path, O_CREAT, mode, inputStream);
}

std::shared_ptr<OutputStream> PacificAdapter::PutFile(const std::string &path, int flags, const FileMode &mode,
    const FileRange &range) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return nullptr;
    }

    auto realFlag = SupportedFlags(flags);
    struct stat statBuf {};
    FileMode newMode = mode;
    // adapter mindspore write finish set mode (only read)
    auto ret = stat(fullPath.c_str(), &statBuf);
    if (ret == 0) {
        newMode.mode = static_cast<int64_t>(statBuf.st_mode);
        auto result = chmod(fullPath.c_str(), static_cast<mode_t>(DEFAULT_MODE & 0777));
        if (result < 0) {
            UFS_LOG_ERROR("set file meta path(" << path << ") failed(" << errno << ":" << strerror(errno) << ")");
            return nullptr;
        }
    } else {
        if (errno == ENOENT) {
            newMode.mode = DEFAULT_MODE;
        } else {
            UFS_LOG_ERROR("open file(" << path << ") to get meta failed(" << errno << " : " << strerror(errno) << ")");
            return nullptr;
        }
    }
    auto fd = open(fullPath.c_str(), realFlag | O_WRONLY | O_NOFOLLOW, static_cast<mode_t>(newMode.mode & 0777));
    if (fd < 0) {
        UFS_LOG_ERROR("open file(" << path << ") to write failed(" << errno << " : " << strerror(errno) << ")");
        return nullptr;
    }

    if (CorrectFileMetadata(fd, newMode) != 0) {
        UFS_LOG_ERROR("set owner for file(" << path << ") failed(" << errno << " : " << strerror(errno) << ")");
        close(fd);
        return nullptr;
    }

    if (range.fileTotalSize == 0L) {
        UFS_LOG_DEBUG("file(" << path << ") is empty, return null output stream");
        close(fd);
        return std::make_shared<NullOutputStream>();
    }

    auto wantWriteSize = MovePosForRange2Write(fullPath, fd, range);
    if (wantWriteSize < 0L) {
        close(fd);
        return nullptr;
    }

    if (wantWriteSize == 0L) {
        close(fd);
        return std::make_shared<NullOutputStream>();
    }
    return std::make_shared<FileOutputStream>(fullPath, fd, wantWriteSize);
}

std::shared_ptr<OutputStream> PacificAdapter::PutFile(const std::string &path, const FileMode &mode,
    const FileRange &range) noexcept
{
    return PutFile(path, O_CREAT, mode, range);
}

std::shared_ptr<OutputStream> PacificAdapter::PutFile(const std::string &path, const FileMode &mode) noexcept
{
    return PutFile(path, O_CREAT | O_TRUNC, mode, FileRange{});
}

int PacificAdapter::GetFile(const std::string &path, utils::ByteBuffer &dataBuffer, const FileRange &range) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("get file(" << path << ") start");
    auto fd = open(fullPath.c_str(), O_RDONLY | O_NOFOLLOW);
    if (fd < 0) {
        UFS_LOG_ERROR("open file(" << path << ") for read failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    auto wantReadSize = MovePosForRange(fullPath, fd, range);
    if (wantReadSize <= 0L) {
        close(fd);
        return static_cast<int>(wantReadSize);
    }

    if (wantReadSize > SMALL_FILE_MAX_SIZE) {
        UFS_LOG_ERROR("open file(" << path << ") for read all, file too large(" << wantReadSize << ")");
        close(fd);
        errno = EOVERFLOW;
        return -1;
    }

    utils::ByteBuffer buffer{ static_cast<uint64_t>(wantReadSize) };
    if (!buffer.Valid()) {
        close(fd);
        return -1;
    }

    auto count = utils::FileUtils::ReadFull(fd, buffer.Data(), buffer.Capacity());
    if (count < 0) {
        UFS_LOG_ERROR("read file(" << path << ") failed(" << errno << " : " << strerror(errno) << ")");
        close(fd);
        return -1;
    }

    close(fd);
    buffer.AddOffset(static_cast<uint64_t>(count));
    dataBuffer = std::move(buffer);
    UFS_LOG_DEBUG("get file(" << path << ") finished data = " << count);

    return 0;
}

std::shared_ptr<InputStream> PacificAdapter::GetFile(const std::string &path, const FileRange &range) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return nullptr;
    }
    UFS_LOG_DEBUG("get file(" << path << ") return input stream start");
    auto fd = open(fullPath.c_str(), O_RDONLY | O_NOFOLLOW);
    if (fd < 0) {
        UFS_LOG_ERROR("open file(" << path << ") for read failed(" << errno << " : " << strerror(errno) << ")");
        return nullptr;
    }

    auto wantReadSize = MovePosForRange(fullPath, fd, range);
    if (wantReadSize < 0L) {
        close(fd);
        return nullptr;
    }

    if (wantReadSize == 0L) {
        close(fd);
        return std::make_shared<NullInputStream>();
    }

    return std::make_shared<FileInputStream>(fullPath, fd, wantReadSize);
}

int PacificAdapter::GetFile(const std::string &path, const FileRange &range, OutputStream &outputStream) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("get file(" << path << ") with stream start");
    auto fd = open(fullPath.c_str(), O_RDONLY | O_NOFOLLOW);
    if (fd < 0) {
        UFS_LOG_ERROR("open file(" << path << ") for read failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    auto wantReadSize = MovePosForRange(fullPath, fd, range);
    if (wantReadSize <= 0L) {
        close(fd);
        return static_cast<int>(wantReadSize);
    }

    utils::ByteBuffer buf{ IO_BUF_SIZE };
    if (!buf.Valid()) {
        close(fd);
        return -1;
    }

    int64_t count;
    while ((count = utils::FileUtils::ReadFull(fd, buf.Data(), buf.Capacity())) > 0) {
        auto bytes = outputStream.Write(buf.Data(), count);
        if (bytes < 0L) {
            UFS_LOG_ERROR("open file(" << path << ") to write failed(" << errno << " : " << strerror(errno) << ")");
            close(fd);
            return -1;
        }

        if (bytes != count) {
            UFS_LOG_ERROR("read file(" << path << ") want bytes(" << count << ") but real(" << bytes << ")");
            close(fd);
            return -1;
        }
    }

    close(fd);

    UFS_LOG_DEBUG("get file(" << path << ") with output stream success");
    return 0;
}

int PacificAdapter::MoveFile(const std::string &source, const std::string &destination) noexcept
{
    auto fullSource = GetRealPath(source);
    if (fullSource.empty()) {
        return -1;
    }
    auto fullDest = GetRealPath(destination);
    if (fullDest.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("move file(" << source << ") to be(" << destination << ") start");

    auto ret = rename(fullSource.c_str(), fullDest.c_str());
    if (ret < 0) {
        UFS_LOG_ERROR("move file(" << source << ") to be(" << destination << ") failed(" << errno << " : " <<
            strerror(errno) << ")");
        return -1;
    }

    UFS_LOG_DEBUG("move file(" << source << ") to be(" << destination << ") success");
    return 0;
}

int PacificAdapter::CopyFile(const std::string &source, const std::string &destination) noexcept
{
    auto fullSource = GetRealPath(source);
    if (fullSource.empty()) {
        return -1;
    }
    auto fullDest = GetRealPath(destination);
    if (fullDest.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("link file(" << source << ") to be(" << destination << ") start");

    auto ret = link(fullSource.c_str(), fullDest.c_str());
    if (ret < 0) {
        UFS_LOG_ERROR("link file(" << source << ") to be(" << destination << ") failed(" << errno << " : " <<
            strerror(errno) << ")");
        return -1;
    }

    UFS_LOG_DEBUG("link file(" << source << ") to be(" << destination << ") success");
    return 0;
}

int PacificAdapter::RemoveFile(const std::string &path) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("remove file(" << path << ") start");

    auto ret = unlink(fullPath.c_str());
    if (ret < 0) {
        UFS_LOG_ERROR("remove file(" << path << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    UFS_LOG_DEBUG("remove file(" << path << ") success");
    return 0;
}

int PacificAdapter::CreateDirectory(const std::string &path, const FileMode &mode) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("mkdir path(" << path << ") with mode(0o" << std::oct << mode.mode << ") start");

    int ret = mkdir(fullPath.c_str(), static_cast<mode_t>(mode.mode) & 0777);
    if (ret < 0 && errno != EEXIST) {
        UFS_LOG_ERROR("mkdir path(" << path << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    if (!IsDefaultOwner(mode)) {
        ret = chown(fullPath.c_str(), mode.owner, mode.group);
        if (ret < 0) {
            UFS_LOG_WARN("chown for path(" << path << ") failed(" << errno << " : " << strerror(errno) << "), ignore");
        }
    }

    UFS_LOG_DEBUG("mkdir path(" << path << ") with mode(0o" << std::oct << mode.mode << ") success");
    return 0;
}

int PacificAdapter::RemoveDirectory(const std::string &path) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("remove directory path(" << path << ") start");

    auto ret = rmdir(fullPath.c_str());
    if (ret < 0) {
        UFS_LOG_ERROR("remove directory path(" << path << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    UFS_LOG_DEBUG("remove directory path(" << path << ") success");
    return 0;
}

int PacificAdapter::ListFiles(const std::string &path, ListFileResult &result) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("list files from path(" << path << ") start");

    auto dir = opendir(fullPath.c_str());
    if (dir == nullptr) {
        UFS_LOG_ERROR("list files open dir(" << path << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    struct dirent *entry;
    result.files.clear();
    auto count = 0U;
    while (count < MAX_ITEM_ONE_PAGE && (entry = readdir(dir)) != nullptr) {
        if (ShouldSkipName(entry->d_name)) {
            continue;
        }

        result.files.emplace_back(std::string{ entry->d_name }, entry->d_type != DT_DIR);
        count++;
    }

    if (entry == nullptr) {
        result.marker.reset(new FinishedMarker);
        closedir(dir);
        UFS_LOG_DEBUG("list files from path(" << path << ") finished");
    } else {
        result.marker.reset(new ReadDirMarker(fullPath, dir, entry->d_off));
        UFS_LOG_DEBUG("list files from path(" << path << ") return not finished");
    }

    return 0;
}

int PacificAdapter::ListFiles(const std::string &path, ListFileResult &result,
    std::shared_ptr<ListFilePageMarker> marker) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("continue list files from path(" << path << ") start");

    auto m = dynamic_cast<ReadDirMarker *>(marker.get());
    if (m == nullptr) {
        UFS_LOG_ERROR("continue list files from path(" << path << ") failed, marker invalid.");
        return -1;
    }

    auto dir = m->dir;
    struct dirent *entry;
    result.files.clear();
    auto count = 0U;

    seekdir(dir, m->position);
    while (count < MAX_ITEM_ONE_PAGE && (entry = readdir(dir)) != nullptr) {
        if (ShouldSkipName(entry->d_name)) {
            continue;
        }

        result.files.emplace_back(std::string{ entry->d_name }, entry->d_type != DT_DIR);
        count++;
    }

    if (entry == nullptr) {
        result.marker.reset(new FinishedMarker);
        closedir(dir);
        m->dir = nullptr;
        m->finished = true;
        UFS_LOG_DEBUG("continue list files from path(" << path << ") finished");
    } else {
        result.marker.reset(new ReadDirMarker(fullPath, dir, entry->d_off));
        UFS_LOG_DEBUG("continue list files from path(" << path << ") return not finished");
    }

    return 0;
}

int PacificAdapter::GetFileMeta(const std::string &path, FileMeta &meta) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("get file meta for path(" << path << ") start");

    struct stat statBuf {};
    auto ret = lstat(fullPath.c_str(), &statBuf);
    if (ret < 0) {
        UFS_LOG_DEBUG("get file meta for path(" << path << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    meta.name = fullPath;
    meta.mode = statBuf.st_mode;
    meta.size = static_cast<uint64_t>(statBuf.st_size);
    meta.mtime = statBuf.st_mtim;
    meta.meta["st_dev"] = std::to_string(statBuf.st_dev);
    meta.meta["st_ino"] = std::to_string(statBuf.st_ino);
    meta.meta["st_mode"] = std::to_string(statBuf.st_mode);
    meta.meta["st_nlink"] = std::to_string(statBuf.st_nlink);
    meta.meta["st_uid"] = std::to_string(statBuf.st_uid);
    meta.meta["st_gid"] = std::to_string(statBuf.st_gid);
    meta.meta["st_rdev"] = std::to_string(statBuf.st_rdev);
    meta.meta["st_size"] = std::to_string(statBuf.st_size);
    meta.meta["st_blksize"] = std::to_string(statBuf.st_blksize);
    meta.meta["st_blocks"] = std::to_string(statBuf.st_blocks);
    meta.meta["st_atime"] = std::to_string(statBuf.st_atime);
    meta.meta["st_mtime"] = std::to_string(statBuf.st_mtime);
    meta.meta["st_ctime"] = std::to_string(statBuf.st_ctime);
    UFS_LOG_DEBUG("get file meta for path(" << path << ") success, mode(0" << statBuf.st_mode << "), owner(" <<
        statBuf.st_uid << ":" << statBuf.st_gid << ")");

    static constexpr auto ownerPermShift = 6;
    static constexpr auto groupPermShift = 3;
    meta.acl.users.clear();
    meta.acl.groups.clear();
    meta.acl.ownerPerm = ((statBuf.st_mode & 0700U) >> ownerPermShift);
    meta.acl.groupPerm = ((statBuf.st_mode & 070U) >> groupPermShift);
    meta.acl.otherPerm = (statBuf.st_mode & 07U);
    meta.acl.permMask = memfs::MemFsConstants::ACL_DEFAULT_MASK;
    return GetFileAcl(fullPath, meta.acl);
}

int PacificAdapter::SetFileMeta(const std::string &path, std::map<std::string, std::string> &meta) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return -1;
    }
    UFS_LOG_DEBUG("set file meta for path(" << path << ") start");

    auto pos = meta.find("st_mode");
    if (pos != meta.end()) {
        auto mode = static_cast<mode_t>((std::strtol(pos->second.c_str(), nullptr, 10) & 0777));
        auto ret = chmod(fullPath.c_str(), mode);
        if (ret < 0) {
            UFS_LOG_ERROR("set file meta path(" << path << ") mode(" << pos->second << " : 0" << mode << ") failed(" <<
                errno << " : " << strerror(errno) << ")");
            return -1;
        }
        UFS_LOG_ERROR("set file meta path(" << path << ") mode(" << pos->second << " : 0" << mode << ") success");
    }

    UFS_LOG_DEBUG("set file meta for path(" << path << ") finished");
    return 0;
}

std::shared_ptr<FileLock> PacificAdapter::GetFileLock(const std::string &path) noexcept
{
    auto fullPath = GetRealPath(path);
    if (fullPath.empty()) {
        return nullptr;
    }
    UFS_LOG_DEBUG("get file lock for path(" << path << ") start");

    auto fd = open(fullPath.c_str(), O_WRONLY | O_NOFOLLOW);
    if (fd < 0) {
        UFS_LOG_ERROR("get file lock for path(" << path << ") failed(" << errno << " : " << strerror(errno) << ")");
        return nullptr;
    }

    return std::make_shared<PacificFileLock>(path, fd);
}

std::string PacificAdapter::GetRealPath(const std::string &path) noexcept
{
    auto originPath = mountPath + path;
    if (originPath.length() >= PATH_MAX) {
        UFS_LOG_ERROR("path name(" << path << ") too long");
        errno = EINVAL;
        return "";
    }

    char tempPath[PATH_MAX + 1];
    if (strcpy_s(tempPath, PATH_MAX, originPath.c_str()) != EOK) {
        UFS_LOG_ERROR("copy path name(" << path << ") failed");
        return "";
    }

    tempPath[PATH_MAX] = '\0';
    auto baseName = std::string(basename(tempPath));
    auto dirName = std::string(dirname(tempPath));

    auto str = realpath(dirName.c_str(), tempPath);
    if (str == nullptr) {
        UFS_LOG_ERROR("call realpath for dir name(" << dirName << ") failed(" << errno << " : " << strerror(errno) <<
            ")");
        return "";
    }

    return std::string(str).append("/").append(baseName);
}

int64_t PacificAdapter::MovePosForRange(const std::string &path, int fd, const FileRange &range) noexcept
{
    struct stat statBuf {};
    auto ret = fstat(fd, &statBuf);
    if (ret < 0) {
        UFS_LOG_ERROR("stat file for check size failed(" << errno << " : " << strerror(errno) << ")");
        return -1L;
    }

    if (range.begin >= static_cast<uint64_t>(statBuf.st_size)) {
        UFS_LOG_ERROR("file with range(" << range.begin << ", " << range.count << ") no data to read");
        return 0L;
    }

    auto leftSize = static_cast<uint64_t>(statBuf.st_size) - range.begin;
    auto readSize = std::min(leftSize, range.count);
    if (range.begin > 0UL) {
        auto off = lseek(fd, static_cast<off_t>(range.begin), SEEK_SET);
        if (off == static_cast<off_t>(-1)) {
            UFS_LOG_ERROR("seek file to (" << range.begin << ") failed(" << errno << " : " << strerror(errno) << ")");
            return -1L;
        }
    }

    return static_cast<int64_t>(readSize);
}

int64_t PacificAdapter::MovePosForRange2Write(const std::string &path, int fd,
    const ock::ufs::FileRange &range) noexcept
{
    if (range.begin >= static_cast<uint64_t>(range.fileTotalSize)) {
        UFS_LOG_WARN("file with range(" << range.begin << ", " << range.count << ") no data to write.");
        return 0L;
    }

    auto leftSize = static_cast<uint64_t>(range.fileTotalSize) - range.begin;
    auto writeSize = std::min(leftSize, range.count);
    if (range.begin > 0UL) {
        auto off = lseek(fd, static_cast<off_t>(range.begin), SEEK_SET);
        if (off == static_cast<off_t>(-1)) {
            UFS_LOG_ERROR("seek file to (" << range.begin << ") failed(" << errno << " : " << strerror(errno) << ")");
            return -1L;
        }
    }

    return static_cast<int64_t>(writeSize);
}

bool PacificAdapter::IsDefaultOwner(const FileMode &mode) noexcept
{
    return mode.owner == 0U && mode.group == 0U;
}

int PacificAdapter::CorrectFileMetadata(int fd, const FileMode &mode) noexcept
{
    if (fchmod(fd, static_cast<mode_t>(mode.mode)) != 0) {
        return -1;
    }

    if (IsDefaultOwner(mode)) {
        return 0;
    }

    return fchown(fd, mode.owner, mode.group);
}

int PacificAdapter::WriteSmallFile(int fd, const std::string &fullPath, const FileMode &mode,
    utils::ByteBuffer &dataBuffer) noexcept
{
    auto len = dataBuffer.Capacity() - dataBuffer.Offset();

    if (CorrectFileMetadata(fd, mode) != 0) {
        UFS_LOG_ERROR("set owner for file failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    if (len == 0U) {
        UFS_LOG_DEBUG("put file with size(" << dataBuffer.Capacity() << ") success");
        return 0;
    }

    auto buf = dataBuffer.Data() + dataBuffer.Offset();
    auto count = utils::FileUtils::WriteFull(fd, buf, len);
    if (count < 0) {
        UFS_LOG_ERROR("open file to write failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    if (count != static_cast<int>(len)) {
        UFS_LOG_ERROR("write file want bytes(" << len << ") but real(" << count << ")");
        return -1;
    }

    return 0;
}

int PacificAdapter::SupportedFlags(int flags) noexcept
{
    return (flags & (O_CREAT | O_EXCL | O_SYNC | O_DSYNC | O_TRUNC));
}

int PacificAdapter::GetFileAcl(const std::string &path, FileAcl &acl) noexcept
{
    auto size = getxattr(path.c_str(), memfs::MemFsConstants::ACL_XATTR_KEY, nullptr, 0);
    if (size <= 0) {
        return 0;
    }

    auto valueSize = static_cast<uint64_t>(size);
    if (valueSize < sizeof(FileAclHeader)) {
        UFS_LOG_ERROR("get file acl xattr length(" << valueSize << ") invalid.");
        return 0;
    }

    valueSize -= sizeof(FileAclHeader);
    if (valueSize % sizeof(FileAclEntry)) {
        UFS_LOG_ERROR("get file acl xattr value length(" << valueSize << ") invalid.");
        return 0;
    }

    utils::ByteBuffer buffer(size);
    if (!buffer.Valid()) {
        UFS_LOG_ERROR("allocate buffer with size(" << size << ") failed.");
        errno = ENOMEM;
        return -1;
    }

    auto realSize = getxattr(path.c_str(), memfs::MemFsConstants::ACL_XATTR_KEY, buffer.Data(), buffer.Capacity());
    if (realSize != size) {
        UFS_LOG_ERROR("getxattr return (" << size << ") vs << (" << realSize << ")");
        return 0;
    }

    auto valueCount = valueSize / sizeof(FileAclEntry);
    auto aclData = static_cast<FileAclHeader *>(static_cast<void *>(buffer.Data()));
    ParseFileAcl(aclData, valueCount, acl);
    return 0;
}

void PacificAdapter::ParseFileAcl(const ock::ufs::FileAclHeader *header, uint32_t entryCount, FileAcl &acl) noexcept
{
    for (auto i = 0U; i < entryCount; i++) {
        if (header->entries[i].tag == memfs::MemFsConstants::ACL_TAG_USER) {
            acl.users.emplace(header->entries[i].id, header->entries[i].perm);
        } else if (header->entries[i].tag == memfs::MemFsConstants::ACL_TAG_USER_OBJ) {
            acl.ownerPerm = header->entries[i].perm;
        } else if (header->entries[i].tag == memfs::MemFsConstants::ACL_TAG_GROUP) {
            acl.groups.emplace(header->entries[i].id, header->entries[i].perm);
        } else if (header->entries[i].tag == memfs::MemFsConstants::ACL_TAG_GROUP_OBJ) {
            acl.groupPerm = header->entries[i].perm;
        } else if (header->entries[i].tag == memfs::MemFsConstants::ACL_TAG_MASK) {
            acl.permMask = header->entries[i].perm;
        } else if (header->entries[i].tag == memfs::MemFsConstants::ACL_TAG_OTHER) {
            acl.otherPerm = header->entries[i].perm;
        }
    }
}

FileInputStream::FileInputStream(std::string path, int fd, int64_t limit) noexcept
    : filePath{ std::move(path) }, fileDesc{ fd }, maxLeftBytes{ limit }
{}

FileInputStream::~FileInputStream() noexcept
{
    FileInputStream::Close();
}

int FileInputStream::Read(uint8_t &byte) noexcept
{
    return static_cast<int>(Read(&byte, 1UL));
}

int64_t FileInputStream::Read(uint8_t *buf, uint64_t count) noexcept
{
    if (maxLeftBytes <= 0L) {
        return 0L;
    }

    if (static_cast<int64_t>(count) > maxLeftBytes) {
        count = static_cast<uint64_t>(maxLeftBytes);
    }

    auto bytes = utils::FileUtils::ReadFull(fileDesc, buf, count);
    if (bytes > 0) {
        maxLeftBytes -= bytes;
    }

    return bytes;
}

int FileInputStream::Close() noexcept
{
    if (fileDesc >= 0) {
        close(fileDesc);
        UFS_LOG_DEBUG("close file for input stream finished");
    }
    fileDesc = -1;
    filePath.clear();
    return 0;
}

FileOutputStream::FileOutputStream(std::string path, int fd, int64_t count) noexcept
    : fileDesc{ fd }, filePath{ std::move(path) }, maxLeftBytes{ count }
{}

FileOutputStream::~FileOutputStream() noexcept
{
    FileOutputStream::Close();
}

int FileOutputStream::Write(uint8_t byte) noexcept
{
    return static_cast<int>(Write(&byte, 1U));
}

int64_t FileOutputStream::Write(const uint8_t *buf, uint64_t count) noexcept
{
    if (maxLeftBytes <= 0L) {
        return 0L;
    }

    if (static_cast<int64_t>(count) > maxLeftBytes) {
        count = static_cast<uint64_t>(maxLeftBytes);
    }
    auto bytes = utils::FileUtils::WriteFull(fileDesc, buf, count);
    if (bytes > 0) {
        maxLeftBytes -= bytes;
    }
    return bytes;
}

int FileOutputStream::Sync() noexcept
{
    return fsync(fileDesc);
}

int FileOutputStream::Close() noexcept
{
    if (fileDesc >= 0) {
        close(fileDesc);
        UFS_LOG_DEBUG("close file for output stream finished");
    }
    fileDesc = -1;
    filePath.clear();
    return 0;
}

PacificFileLock::PacificFileLock(std::string path, int fd) noexcept : filePath{ std::move(path) }, fileDesc{ fd } {}

PacificFileLock::~PacificFileLock() noexcept
{
    close(fileDesc);
    fileDesc = -1;
}

int PacificFileLock::Lock() noexcept
{
    struct flock lock {};
    lock.l_type = F_WRLCK;
    lock.l_start = 0;
    lock.l_whence = SEEK_SET;
    lock.l_len = 0;

    UFS_LOG_DEBUG("lock file start");
    auto ret = fcntl(fileDesc, F_SETLKW, &lock);
    if (ret < 0) {
        UFS_LOG_ERROR("lock file failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    UFS_LOG_DEBUG("lock file success");
    return 0;
}

int PacificFileLock::TryLock() noexcept
{
    struct flock lock {};
    lock.l_type = F_WRLCK;
    lock.l_start = 0;
    lock.l_whence = SEEK_SET;
    lock.l_len = 0;

    UFS_LOG_DEBUG("try lock file start");
    auto ret = fcntl(fileDesc, F_SETLK, &lock);
    if (ret < 0) {
        if (errno != EACCES && errno != EAGAIN) {
            UFS_LOG_ERROR("lock file failed(" << errno << " : " << strerror(errno) << ")");
        } else {
            UFS_LOG_DEBUG("try lock file not get lock");
        }
        return -1;
    }

    UFS_LOG_DEBUG("try lock file success");
    return 0;
}

int PacificFileLock::Unlock() noexcept
{
    struct flock lock {};
    lock.l_type = F_UNLCK;
    lock.l_start = 0;
    lock.l_whence = SEEK_SET;
    lock.l_len = 0;

    UFS_LOG_DEBUG("unlock file start");
    auto ret = fcntl(fileDesc, F_SETLK, &lock);
    if (ret < 0) {
        UFS_LOG_ERROR("unlock file failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    UFS_LOG_DEBUG("unlock file success");
    return 0;
}
