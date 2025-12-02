/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
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
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>

#include <thread>
#include <chrono>
#include <list>

#include "file_check_utils.h"
#include "service_configure.h"
#include "memfs_logger.h"
#include "mem_file_system.h"
#include "mem_fs_constants.h"
#include "inode_evictor.h"
#include "memfs_api.h"

using namespace ock::common;
using namespace ock::common::config;
namespace ock {
namespace memfs {
std::vector<std::string> PreloadProgressView::g_loadPathVec;
std::mutex PreloadProgressView::gViewMutex;
std::condition_variable PreloadProgressView::gCond;

static constexpr int MAX_ALLOC_RETRY_TIMES = 60;

static FileOpNotify g_fileOpNotify;
static MemFileSystem *g_fileSystem;
static ExternalStat g_externalStat;
static uint64_t g_dataBlockSize = MemFsConstants::DATA_BLOCK_SIZE;

static void ReservePhysicalPage(void *mappedAddress, uint64_t reserveSize)
{
    auto pageSize = 4096UL;
    /* reserve physical page task */
    uint64_t setLength = 0;
    auto startPos = static_cast<uint8_t *>(mappedAddress);
    while (setLength < reserveSize) {
        *startPos = 0;
        setLength += pageSize;
        startPos += pageSize;
    }
}

static void ReleaseMultiBlocks(MemFsBMM &bmm, const std::vector<uint64_t> &blocks)
{
    for (auto &blockId : blocks) {
        bmm.ReleaseOne(blockId);
    }
}

static bool AllocateMultiBlocks(MemFsBMM &bmm, uint64_t count, std::vector<uint64_t> &blocks)
{
    uint64_t blockId;
    blocks.clear();
    blocks.reserve(count);
    auto retryCount = 0;
    for (auto i = 0UL; i < count; i++) {
        auto ret = bmm.AllocateOne(blockId);
        while (ret != MFS_OK) {
            InodeEvictor::GetInstance().RecycleInodes(g_dataBlockSize * (count - i));
            ret = bmm.AllocateOne(blockId);
            if (ret == MFS_OK) {
                retryCount = 0;
                break;
            }

            if (++retryCount > MAX_ALLOC_RETRY_TIMES) {
                break;
            }

            std::this_thread::sleep_for(std::chrono::seconds(1));
        }
        if (ret != MFS_OK) {
            MFS_LOG_ERROR("allocate block(" << i << " of " << count << ") failed(" << ret << ")");
            ReleaseMultiBlocks(bmm, blocks);
            blocks.clear();
            return false;
        }
        blocks.push_back(blockId);
    }

    return true;
}

int MemFsApi::Initialize() noexcept
{
    g_dataBlockSize = ServiceConfigure::GetInstance().GetMemFsConfig().dataBlockSize;
    auto datBlockCount = ServiceConfigure::GetInstance().GetMemFsConfig().dataBlockCount;
    g_fileSystem = new (std::nothrow) MemFileSystem(g_dataBlockSize, datBlockCount, "mfs");
    if (g_fileSystem == nullptr) {
        MFS_LOG_ERROR("create memfs failed.");
        return -1;
    }

    auto ret = g_fileSystem->Initialize();
    if (ret != 0) {
        MFS_LOG_ERROR("initialize memfs failed(" << ret << ").");
        delete g_fileSystem;
        g_fileSystem = nullptr;
        return -1;
    }

    ret = InodeEvictor::GetInstance().Initialize();
    if (ret != 0) {
        MFS_LOG_ERROR("initialize InodeEvictor failed(" << ret << ").");
        g_fileSystem->Destroy();
        delete g_fileSystem;
        g_fileSystem = nullptr;
        return -1;
    }

    return 0;
}

void MemFsApi::Destroy() noexcept
{
    if (g_fileSystem != nullptr) {
        g_fileSystem->Destroy();
        delete g_fileSystem;
        g_fileSystem = nullptr;
    }
}

int MemFsApi::GetShareMemoryFd() noexcept
{
    if (g_fileSystem == nullptr) {
        return -1;
    }

    return g_fileSystem->GetMemFsBMM().GetFD();
}

bool MemFsApi::BackgroundTaskEmpty() noexcept
{
    return g_fileOpNotify.bgTaskEmptyNotify();
}

int MemFsApi::RegisterFileOpNotify(const FileOpNotify &notify) noexcept
{
    g_fileOpNotify = notify;
    return 0;
}

int MemFsApi::OpenFile(const std::string &path, int flags, mode_t mode) noexcept
{
    int fd = -1;
    uint64_t inodeNum;
    if (flags == (O_CREAT | O_TRUNC | O_WRONLY)) {
        fd = g_fileSystem->Create(path, mode, inodeNum);
        if (fd >= 0) {
            auto ret = g_fileOpNotify.openNotify(fd, path, flags, inodeNum);
            if (ret != 0) {
                DiscardFile(path, fd);
                fd = -1;
                errno = EIO;
                MFS_LOG_ERROR("notify file:" << path << " failed(" << ret << "), errno:" << errno);
            }
        }
    } else if (flags == O_RDONLY) {
        fd = g_fileSystem->Open(path, inodeNum);
    } else {
        errno = EINVAL;
    }

    return fd;
}

int MemFsApi::CreateAndOpenFile(const std::string &path, uint64_t &inodeNum, mode_t mode) noexcept
{
    /* mkdir */
    auto pos = path.find_last_of('/');
    if (pos == std::string::npos) {
        MFS_LOG_ERROR("cannot find parent path");
        return -EINVAL;
    }

    auto parent = path.substr(0, pos);
    if (!parent.empty() && parent != "/") {
        auto ret = MemFsApi::CreateDirectoryWithParents(parent, mode);
        if (ret != 0) {
            return -errno;
        }
    }

    return g_fileSystem->Create(path, mode, inodeNum);
}

int MemFsApi::CloseFile(int fd) noexcept
{
    auto ret = g_fileSystem->Close(fd);
    g_fileOpNotify.closeNotify(fd, false);
    return ret;
}

int MemFsApi::DiscardFile(const std::string &path, int fd) noexcept
{
    struct stat fdStat {};
    struct stat pathStat {};

    /*
     * check path matches fd
     * compare inode number in stat for path and fd
     */
    auto ret = g_fileSystem->GetFileMeta(fd, fdStat);
    if (ret != 0) {
        MFS_LOG_WARN("discard file(" << fd << ":" << path << ") stat fd failed : " << errno << " : " <<
            strerror(errno));
        return -1;
    }

    ret = g_fileSystem->GetMeta(path, pathStat);
    if (ret != 0) {
        MFS_LOG_WARN("discard file(" << fd << ":" << path << ") stat path failed : " << errno << " : " <<
            strerror(errno));
        return -1;
    }

    if (fdStat.st_ino != pathStat.st_ino) {
        MFS_LOG_WARN("discard file(" << fd << ":" << path << ") stat path failed, fd inode:" << fdStat.st_ino <<
            ", path inode:" << pathStat.st_ino);
        errno = EINVAL;
        return -1;
    }

    ret = g_fileSystem->Close(fd);
    g_fileOpNotify.closeNotify(fd, true);
    if (ret != MFS_OK) {
        MFS_LOG_WARN("Discard file close failed as " << ret);
        return -1;
    }

    uint64_t inode;
    ret = g_fileSystem->RemoveFile(path, inode);
    if (ret != MFS_OK) {
        MFS_LOG_WARN("Discard file remove failed as " << ret);
        return -1;
    }
    return 0;
}

int MemFsApi::SetBackupFinished(int fd) noexcept
{
    auto ret = g_fileSystem->SetBackupFinished(fd);
    return ret;
}

int MemFsApi::AllocDataBlock(int fd, uint64_t &blockId, uint64_t &blockSize) noexcept
{
    auto &bmm = g_fileSystem->GetMemFsBMM();
    auto allocateResult = bmm.AllocateOne(blockId);
    auto retryCount = 0;
    while (allocateResult != MFS_OK) {
        InodeEvictor::GetInstance().RecycleInodes(g_dataBlockSize);
        allocateResult = bmm.AllocateOne(blockId);
        if (allocateResult == MFS_OK) {
            break;
        }

        if (++retryCount > MAX_ALLOC_RETRY_TIMES) {
            break;
        }

        std::this_thread::sleep_for(std::chrono::seconds(1));
    }
    if (allocateResult != MFS_OK) {
        MFS_LOG_ERROR("allocate block failed(" << allocateResult << ")");
        errno = ENOMEM;
        return -1;
    }

    blockSize = g_fileSystem->GetBlockSize();
    ReservePhysicalPage(MemFsApi::BlockToAddress(blockId), blockSize);
    auto ret = g_fileSystem->AppendFileDataBlock(fd, blockId);
    if (ret != 0) {
        bmm.ReleaseOne(blockId);
        return -1;
    }
    return 0;
}

int MemFsApi::AllocDataBlocks(int fd, uint64_t bytes, std::vector<uint64_t> &blocks, uint64_t &blockSize) noexcept
{
    auto &bmm = g_fileSystem->GetMemFsBMM();
    blockSize = g_fileSystem->GetBlockSize();
    if (blockSize == 0) {
        return -1;
    }

    auto blockCount = (bytes + blockSize - 1UL) / blockSize;
    if (!AllocateMultiBlocks(bmm, blockCount, blocks)) {
        errno = ENOMEM;
        return -1;
    }

    auto reserveTask = [blockSize](const std::vector<uint64_t> &blocks) {
        for (auto &block : blocks) {
            ReservePhysicalPage(MemFsApi::BlockToAddress(block), blockSize);
        }
    };

    auto threadNum = ock::common::config::ServiceConfigure::GetInstance().GetMemFsConfig().writeParallel.threadNum;
    auto everyThreadBlk = blockCount / threadNum;
    auto leftBlk = blockCount % threadNum;
    if (everyThreadBlk == 0) {
        everyThreadBlk = 1;
        threadNum = leftBlk;
    }

    uint64_t pos = 0UL;
    std::vector<std::thread> threadPool;
    threadPool.reserve(threadNum);
    for (uint32_t i = 0UL; i < threadNum; ++i) {
        auto curBlk = everyThreadBlk;
        if (threadNum != leftBlk && i < leftBlk) {
            curBlk++;
        }
        threadPool.emplace_back(reserveTask,
            std::vector<uint64_t>(blocks.begin() + pos, blocks.begin() + pos + curBlk));
        pos += curBlk;
    }

    for (auto &th : threadPool) {
        th.join();
    }

    auto ret = g_fileSystem->AppendFileDataBlocks(fd, blocks);
    if (ret != 0) {
        ReleaseMultiBlocks(bmm, blocks);
        blocks.clear();
        return -1;
    }

    return 0;
}

int MemFsApi::TruncateFile(int fd, uint64_t length) noexcept
{
    return g_fileSystem->TruncateFile(fd, length);
}

int MemFsApi::GetFileMeta(int fd, struct stat &statBuf) noexcept
{
    return g_fileSystem->GetFileMeta(fd, statBuf);
}

int MemFsApi::GetFileBlocks(int fd, std::vector<uint64_t> &blocks) noexcept
{
    auto ret = g_fileSystem->GetFileDataBlocks(fd, blocks);
    return blocks.empty() ? MFS_ERROR : ret;
}

void *MemFsApi::BlockToAddress(uint64_t blockId) noexcept
{
    BmmBlkId bmmBlkId;
    bmmBlkId.whole = blockId;
    return reinterpret_cast<void *>(bmmBlkId.blkAddress);
}

uint64_t MemFsApi::GetBlockOffset(uint64_t blockId) noexcept
{
    return g_fileSystem->GetMemFsBMM().GetBlockOffset(blockId);
}

int MemFsApi::CreateDirectory(const std::string &path, mode_t mode, bool recursive) noexcept
{
    if (recursive) {
        return CreateDirectoryWithParents(path, mode);
    }

    {
        auto ret = g_fileOpNotify.mkdirNotify(path, mode, getuid(), getgid());
        if (ret != 0) {
            return ret;
        }
    }

    uint64_t inode;
    return g_fileSystem->MakeDirectory(path, mode, inode);
}

int MemFsApi::CreateDirectoryWithParents(const std::string &path, mode_t mode) noexcept
{
    std::string currentPath;

    auto items = MemFileSystem::GetPathNameTokens(path);
    for (auto &item : items) {
        currentPath.append("/").append(item);
        auto ret = CreateOneLevelDirectory(currentPath, mode);
        if (ret != 0) {
            return -errno;
        }
    }

    return 0;
}

int MemFsApi::RemoveDirectory(const std::string &path) noexcept
{
    uint64_t inode;
    auto ret = g_fileSystem->RemoveDirectory(path, inode);
    if (ret == 0) {
        g_fileOpNotify.unlinkNotify(path, inode);
    }
    return ret;
}

int MemFsApi::ReadDirectory(const std::string &path, std::vector<std::pair<std::string, bool>> &entries) noexcept
{
    std::vector<std::pair<std::string, InodeType>> fsEntries;
    auto ret = g_fileSystem->ListDirectory(path, fsEntries);
    if (ret != 0) {
        return ret;
    }

    entries.clear();
    entries.reserve(fsEntries.size());
    for (auto &e : fsEntries) {
        entries.emplace_back(e.first, e.second == INODE_REG);
    }

    return 0;
}

int MemFsApi::GetMeta(const std::string &path, struct stat &statBuf) noexcept
{
    return g_fileSystem->GetMeta(path, statBuf);
}

int MemFsApi::GetMetaAcl(const std::string &path, struct stat &statBuf, MemfsFileAcl &acl) noexcept
{
    return g_fileSystem->GetMeta(path, statBuf, acl);
}

int MemFsApi::Link(const std::string &oldPath, const std::string &newPath) noexcept
{
    uint64_t inode;
    auto ret = g_fileSystem->LinkFile(oldPath, newPath, inode);
    if (ret == 0) {
        auto result = g_fileOpNotify.newFileNotify(newPath, inode);
        if (result != 0) {
            LOG_ERROR("link path: " << newPath << ", notify failed: " << result);
            errno = EIO;
            g_fileSystem->RemoveFile(newPath, inode);
            return -1;
        }
    }
    return ret;
}

int MemFsApi::Rename(const std::string &oldPath, const std::string &newPath, uint32_t flags) noexcept
{
    if (oldPath == newPath) {
        return 0;
    }

    auto ret = g_fileSystem->Rename(oldPath, newPath, flags);
    if (ret != 0) {
        LOG_ERROR("rename(" << FileCheckUtils::RemovePrefixPath(oldPath) << ") to(" <<
            FileCheckUtils::RemovePrefixPath(newPath) << ") with flags: 0x" << std::hex << flags << std::oct <<
            "failed: " << errno << ":" << strerror(errno));
    }
    return ret;
}

int MemFsApi::Unlink(const std::string &path) noexcept
{
    uint64_t inode;
    auto ret = g_fileSystem->RemoveFile(path, inode);
    if (ret == 0) {
        g_fileOpNotify.unlinkNotify(path, inode);
    }
    return ret;
}

int MemFsApi::Chmod(const std::string &path, mode_t mode) noexcept
{
    return g_fileSystem->Chmod(path, mode);
}

int MemFsApi::Chown(const std::string &path, uid_t uid, gid_t gid) noexcept
{
    return g_fileSystem->Chown(path, uid, gid);
}

int MemFsApi::GetFileSystemStat(struct statvfs &statBuf) noexcept
{
    auto ret = g_fileSystem->GetFileSystemStat(statBuf);
    return ret;
}

void MemFsApi::SetExternalStat(const ExternalStat &externalStat) noexcept
{
    if (externalStat == nullptr || g_fileSystem == nullptr) {
        return;
    }

    g_externalStat = externalStat;
    g_fileSystem->SetInjectMkdirCheck(
        [](MemFileSystem *fs, const std::string &path, struct stat &statBuf, InodeAcl &acl) {
            if (fs == g_fileSystem) {
                auto ret = g_externalStat(path, statBuf, acl);
                if (ret != 0) {
                    return ret;
                }
                return 0;
            }
            return -EINVAL;
        });
}

int MemFsApi::CreateOneLevelDirectory(const std::string &path, mode_t mode) noexcept
{
    struct stat statBuf {};
    auto ret = g_fileSystem->GetMeta(path, statBuf);
    if (ret == 0) {
        if (S_ISDIR(statBuf.st_mode)) {
            return 0;
        }
        errno = ENOTDIR;
        return -1;
    }

    if (errno != ENOENT) {
        return -1;
    }

    ret = CreateDirectory(path, mode, false);
    if (ret == 0) {
        return 0;
    }

    if (errno == EEXIST) {
        return 0;
    }

    return -1;
}

void MemFsApi::Serviceable(bool state) noexcept
{
    g_fileSystem->Serviceable(state);
}

bool MemFsApi::Serviceable() noexcept
{
    return g_fileSystem->Serviceable();
}

void MemFsApi::GetShareFileCfg(uint64_t &blockSize, uint64_t &blockCnt) noexcept
{
    blockSize = g_fileSystem->GetBlockSize();
    blockCnt = g_fileSystem->GetBlockCount();
}

int MemFsApi::PreloadFile(const std::string &path) noexcept
{
    int ret = 0;
    if (g_fileSystem->Exist(path)) {
        LOG_INFO("preload path(" << path << ") already exist, no need load again.");
    } else {
        ret = g_fileOpNotify.preloadFileNotify(path);
        if (ret < 0) {
            LOG_ERROR("preload path(" << path << ") notify failed: " << ret);
        } else {
            PreloadProgressView::InsertPath(path);
        }
    }
    return ret;
}

}
}