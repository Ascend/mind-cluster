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
#include <aio.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <vector>
#include <thread>
#include "background_log.h"
#include "memfs_api.h"
#include "service_configure.h"
#include "aio_sync.h"
#include "mem_fs_backup_initiator.h"

using namespace ock::memfs;
using namespace ock::bg::backup;
using namespace ock::ufs;
using namespace ock::common::config;

__thread bool MemFsBackupInitiator::marked;
static constexpr auto MIN_WRITE_PARALLEL_THREAD_COUNT = 2U;
static constexpr uint32_t SINGLE_THREAD_WRITE_STANDARD_SIZE = 16 * 1024 * 1024U;
std::atomic<uint64_t> MemFsBackupInitiator::taskIdGen{ 0x4321abUL };


int MemFsBackupInitiator::GetAttribute(uint64_t taskId, const std::string &path, struct stat &buf) noexcept
{
    BKG_LOG_DEBUG("task(" << taskId << ") get attribute for file(" << path.c_str() << ")");
    auto ret = MemFsApi::GetMeta(path, buf);
    if (ret != 0) {
        BKG_LOG_INFO("task(" << taskId << ") stat for file(" << path.c_str() << ") failed(" << errno << " : " <<
            strerror(errno) << ")");
        return -errno;
    }

    return 0;
}

int MemFsBackupInitiator::RemoveStageFileFromUfs(const std::string &path, const UFS &ufs) noexcept
{
    auto ret = MemFsApi::Unlink(path);
    if (ret != 0) {
        return -1;
    }
    BKG_LOG_INFO("memfs file write finished, and copy file to ufs finished, remove memfs file success, path: " <<
        path.c_str());
    std::string stagePath = path;
    stagePath.append(".m.stg");
    ret = ufs->RemoveFile(stagePath);
    if (ret != 0 && errno != ENOENT) {
        BKG_LOG_ERROR("remove stage file(" << stagePath.c_str() << ") failed(" << errno << " : " << strerror(errno) <<
            "),ret:" << ret);
        return -1;
    }
    return ret;
}

int MemFsBackupInitiator::MultiTasksDoWrite(const std::string &path, const TaskInfo &taskInfo,
    const struct stat &fileStat, UFS ufs) noexcept
{
    auto taskId = taskIdGen.fetch_add(1UL);
    BKG_LOG_DEBUG("execute subtask task id: " << taskId << ", task offset: " << taskInfo.offset << ", task size: " <<
        taskInfo.threadSize << ", file total size: " << taskInfo.fileTotalSize);
    auto fd = MemFsApi::OpenFile(path, O_RDONLY);
    if (fd < 0) {
        BKG_LOG_ERROR("open file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }
    auto outputStream = ufs->PutFile(path, ufs::FileMode(fileStat.st_mode, fileStat.st_uid, fileStat.st_gid),
        FileRange{ taskInfo.offset, taskInfo.threadSize, taskInfo.fileTotalSize });
    if (outputStream == nullptr) {
        BKG_LOG_ERROR("open file(" << path.c_str() << ") write on UFS failed(" << errno << " : " << strerror(errno) <<
            ")");
        MemFsApi::CloseFile(fd);
        return -1;
    }

    std::vector<uint64_t> blocks;
    auto ret = MemFsApi::GetFileBlocks(fd, blocks);
    if (ret < 0) {
        BKG_LOG_ERROR("get blocks for file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) <<
            ")");
        MemFsApi::CloseFile(fd);
        return -1;
    }

    auto blockSize = static_cast<uint64_t>(fileStat.st_blksize);
    auto leftBytes = taskInfo.threadSize;
    auto offset = taskInfo.offset;
    while (leftBytes > 0UL) {
        auto blockIndex = offset / blockSize;
        auto blockOffset = offset % blockSize;
        auto leftBlockSize = blockSize - blockOffset;
        auto writeSize = std::min(leftBlockSize, leftBytes);
        auto address = (uint8_t *)((uintptr_t)MemFsApi::BlockToAddress(blocks[blockIndex]) + blockOffset);
        if (outputStream->Write(address, writeSize) < 0) {
            BKG_LOG_WARN("write file(" << path.c_str() << ") on UFS failed(" << errno << " : " << strerror(errno) <<
                ")");
            MemFsApi::CloseFile(fd);
            return -1;
        }

        leftBytes -= writeSize;
        offset += writeSize;
    }
    MemFsApi::CloseFile(fd);
    outputStream.reset();

    taskInfo.paraLoadCtx->succeedCnt.fetch_add(1);
    BKG_LOG_DEBUG("execute subtask task id: " << taskId << "write finished.");

    return 0;
}

int MemFsBackupInitiator::SplitUploadFileTask(const std::string &path, const struct stat &fileStat, UFS &ufs) noexcept
{
    uint64_t dataSize = fileStat.st_size;
    uint32_t totalThreadNum = ServiceConfigure::GetInstance().GetBackgroundConfig().backupServiceConfig.threadNum;
    uint32_t threadNum;

    auto multiThreadMinSize = MIN_WRITE_PARALLEL_THREAD_COUNT * SINGLE_THREAD_WRITE_STANDARD_SIZE;
    if (dataSize < multiThreadMinSize) {
        threadNum = 1;
    } else {
        threadNum = (dataSize + SINGLE_THREAD_WRITE_STANDARD_SIZE - 1UL) / SINGLE_THREAD_WRITE_STANDARD_SIZE;
        if (threadNum > totalThreadNum) {
            threadNum = totalThreadNum;
        }
    }

    auto thSize = dataSize / threadNum;
    auto paraLoadCtxPtr = std::make_shared<ParallelLoadContext>(threadNum);

    std::vector<std::thread> taskThreads;
    taskThreads.reserve(threadNum);

    try {
        for (uint32_t i = 0; i < threadNum; ++i) {
            uint64_t offset = i * thSize;
            if (i == threadNum - 1) {
                thSize = dataSize - offset;
            }
            paraLoadCtxPtr->RecordTaskOffset(offset);
            TaskInfo taskInfo(-1, thSize, offset, dataSize, paraLoadCtxPtr);
            taskThreads.emplace_back(
                [this, path, taskInfo, fileStat, ufs]() { MultiTasksDoWrite(path, taskInfo, fileStat, ufs); });
        }

        for (auto &taskThread : taskThreads) {
            taskThread.join();
        }
    } catch (const std::exception &e) {
        BKG_LOG_ERROR("Exception occurred: " << e.what());
        return -1;
    }

    if (paraLoadCtxPtr->succeedCnt.load() != threadNum) {
        auto ret = ufs->RemoveFile(path);
        if (ret != 0 && errno != ENOENT) {
            BKG_LOG_ERROR("remove file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) <<
                "),ret:" << ret);
            return -1;
        }
        BKG_LOG_ERROR("copy file to ufs failed, successCount = " << paraLoadCtxPtr->succeedCnt);
        return -1;
    }
    return MultiTasksWriteFinish(path, ufs);
}

int MemFsBackupInitiator::MultiCopyFileToUfs(uint64_t taskId, const std::string &path, UFS &ufs) noexcept
{
    BKG_LOG_INFO("multi task start (" << taskId << ") copy file(" << path.c_str() << ") to ufs");
    auto fd = MemFsApi::OpenFile(path, O_RDONLY);
    if (fd < 0) {
        BKG_LOG_ERROR("open file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    struct stat statBuf {};
    auto ret = MemFsApi::GetFileMeta(fd, statBuf);
    if (ret != 0) {
        BKG_LOG_ERROR("stat file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        MemFsApi::CloseFile(fd);
        return -1;
    }
    MemFsApi::CloseFile(fd);

    return SplitUploadFileTask(path, statBuf, ufs);
}

int MemFsBackupInitiator::RecordToMemfsTaskResult(uint64_t taskId, const std::string &path, int taskResult,
    const TaskInfo &taskInfo) noexcept
{
    auto result = RecordTaskResult(taskId, path, taskResult, taskInfo);
    if (result != 0) {
        return -1;
    }
    auto paraLoadCtxPtr = taskInfo.paraLoadCtx;
    if (!paraLoadCtxPtr->AllTaskFinished()) {
        return 0;
    }
    BKG_LOG_INFO("copy file to memfs tasks all be executed, succeedCnt = " << paraLoadCtxPtr->succeedCnt <<
        ", finishedCnt = " << paraLoadCtxPtr->succeedCnt + paraLoadCtxPtr->failedCnt);

    if (paraLoadCtxPtr->failedCnt > 0) {
        MemFsApi::DiscardFile(path, taskInfo.fd);
        BKG_LOG_ERROR("copy file to memfs failed, failedCount = " << paraLoadCtxPtr->failedCnt);
        return -1;
    }

    taskResult = MemFsApi::TruncateFile(taskInfo.fd, taskInfo.fileTotalSize);
    if (taskResult != 0) {
        BKG_LOG_ERROR("task(" << taskId << ")truncate file(" << path << ") failed");
        return -1;
    }

    MemFsApi::SetBackupFinished(taskInfo.fd);
    taskResult = MemFsApi::CloseFile(taskInfo.fd);
    if (taskResult != 0) {
        BKG_LOG_ERROR("task(" << taskId << ")close file(" << path << ") failed.");
        return -1;
    }

    PreloadProgressView::RemovePath(path);

    BKG_LOG_INFO("preload file path(" << path << ") succeed.");

    return 0;
}

int MemFsBackupInitiator::CopyFileToMemfs(uint64_t taskId, const std::string &path, UFS &ufs,
    const TaskInfo &taskInfo) noexcept
{
    BKG_LOG_DEBUG("task(" << taskId << ") copy file(" << path.c_str() << ") to memfs");

    struct stat statBuf {};
    auto ret = MemFsApi::GetFileMeta(taskInfo.fd, statBuf);
    if (ret != 0) {
        BKG_LOG_WARN("stat file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        MemFsApi::CloseFile(taskInfo.fd);
        return -1;
    }

    auto inputStream = ufs->GetFile(path, FileRange{ taskInfo.offset, taskInfo.threadSize, taskInfo.fileTotalSize });
    if (inputStream == nullptr) {
        BKG_LOG_WARN("open file(" << path.c_str() << ") write to MemFs failed(" << errno << " : " << strerror(errno) <<
            ")");
        MemFsApi::CloseFile(taskInfo.fd);
        return -1;
    }

    std::vector<uint64_t> blocks;
    ret = MemFsApi::GetFileBlocks(taskInfo.fd, blocks);
    if (ret < 0) {
        BKG_LOG_WARN("get blocks for file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        MemFsApi::CloseFile(taskInfo.fd);
        return -1;
    }

    auto blockSize = static_cast<uint64_t>(statBuf.st_blksize);
    auto leftBytes = taskInfo.threadSize;
    auto offset = taskInfo.offset;

    while (leftBytes > 0UL) {
        auto blockIndex = offset / blockSize;
        auto blockOffset = offset % blockSize;
        auto leftBlockSize = blockSize - blockOffset;

        auto readSize = std::min(leftBlockSize, leftBytes);
        auto address = (uint8_t *)((uintptr_t)MemFsApi::BlockToAddress(blocks[blockIndex]) + blockOffset);
        if (inputStream->Read(address, readSize) < 0) {
            BKG_LOG_WARN("read file(" << path.c_str() << ") from UFS failed(" << errno << " : " << strerror(errno) <<
                ")");
            MemFsApi::CloseFile(taskInfo.fd);
            return -1;
        }

        leftBytes -= readSize;
        offset += readSize;
    }

    inputStream.reset();

    return 0;
}

void MemFsBackupInitiator::SplitAndSubmitTask(int fd, const struct stat &ufsBuf, const FileTrace &trace,
    const std::string &path) noexcept
{
    uint64_t dataSize = ufsBuf.st_size;
    uint32_t totalThreadNum = ServiceConfigure::GetInstance().GetBackgroundConfig().backupServiceConfig.threadNum;
    uint32_t threadNum;

    auto multiThreadMinSize = MIN_WRITE_PARALLEL_THREAD_COUNT * SINGLE_THREAD_WRITE_STANDARD_SIZE;
    if (dataSize < multiThreadMinSize) {
        threadNum = 1;
    } else {
        threadNum = (dataSize + SINGLE_THREAD_WRITE_STANDARD_SIZE - 1UL) / SINGLE_THREAD_WRITE_STANDARD_SIZE;
        if (threadNum > totalThreadNum) {
            threadNum = totalThreadNum;
        }
    }

    auto thSize = dataSize / threadNum;
    auto paraLoadCtxPtr = std::make_shared<ParallelLoadContext>(threadNum);

    for (uint32_t i = 0; i < threadNum; ++i) {
        uint64_t offset = i * thSize;
        if (i == threadNum - 1) {
            thSize = dataSize - offset;
        }
        paraLoadCtxPtr->RecordTaskOffset(offset);
        TaskInfo taskInfo(fd, thSize, offset, dataSize, paraLoadCtxPtr);

        backupTarget->MakeFileCache(trace, taskInfo);
    }

    BKG_LOG_INFO("all tasks for make cache already submit, path(" << path << "), task count(" << threadNum << ")");
}

int MemFsBackupInitiator::Prepare() noexcept
{
    FileOpNotify notify;

    notify.openNotify = [this](int fd, const std::string &name, int flags, int64_t inode) -> int {
        return OpenFileNotify(fd, name, flags, inode);
    };

    notify.closeNotify = [this](int fd, bool abnormal) { CloseFileNotify(fd, abnormal); };

    notify.newFileNotify = [this](const std::string &name, int64_t inode) { return NewFileNotify(name, inode); };

    notify.mkdirNotify = [this](const std::string &name, mode_t mode, uid_t owner, gid_t group) {
        return backupTarget->CreateDir(name, mode, owner, group);
    };

    notify.unlinkNotify = [](const std::string &name, int64_t inode) {};

    notify.preloadFileNotify = [this](const std::string &name) { return PreloadFileNotify(name); };

    notify.bgTaskEmptyNotify = [this]() { return backupTarget->IsTaskPoolEmpty(); };

    auto ret = MemFsApi::RegisterFileOpNotify(notify);
    if (ret != 0) {
        BKG_LOG_ERROR("register file operate notify filed(" << ret << ")");
        return -1;
    }

    BKG_LOG_INFO("register file operate notify success.");
    return 0;
}

void MemFsBackupInitiator::SetProcessingMark() noexcept
{
    marked = true;
}

void MemFsBackupInitiator::ClearProcessingMark() noexcept
{
    marked = false;
}

bool MemFsBackupInitiator::CheckProcessingMark() noexcept
{
    return marked;
}

int MemFsBackupInitiator::OpenFileNotify(int fd, const std::string &path, int flags, uint64_t inode) noexcept
{
    if (CheckProcessingMark()) {
        return 0;
    }

    struct stat buf {};
    auto ret = GetAttribute(0, path, buf);
    if (ret != 0) {
        BKG_LOG_ERROR("get file(" << path.c_str() << ") attribute failed : " << ret << " : " << strerror(-ret) << ".");
        return -1;
    }

    FileTrace trace(path, static_cast<int64_t>(inode));
    ret = backupTarget->CreateFileAndStageSync(trace, buf);
    if (ret != 0) {
        BKG_LOG_ERROR("sync file(" << path.c_str() << ") failed : " << ret << " : " << strerror(-ret) << ".");
        return -1;
    }

    BKG_LOG_DEBUG("open file(" << path.c_str() << ") fd(" << fd << ") flags(0" << flags << ") inode(" << inode << ")");
    fileTracer.TraceOpen(fd, path, static_cast<int64_t>(inode));
    return 0;
}

void MemFsBackupInitiator::CloseFileNotify(int fd, bool abnormal) noexcept
{
    if (CheckProcessingMark()) {
        return;
    }

    FileTrace trace;
    BKG_LOG_DEBUG("close file(" << fd << ")");
    if (!fileTracer.CloseFind(fd, trace)) {
        BKG_LOG_DEBUG("close file(" << fd << ") not opened.");
        return;
    }

    if (abnormal) {
        // no need to flush data into under fs
        BKG_LOG_INFO("Ignore close under fs file " << fd << ", trigger by discard, path:" << trace.path);
        backupTarget->RemoveFileAndStageSync(trace);
        return;
    }
    struct stat fileStat {};
    auto ret = CompareTraceFile(trace, fileStat);
    if (ret != 0) {
        return;
    }

    NotifyProcessMark marker(this);
    backupTarget->UploadFile(trace, fileStat, true);
}

int MemFsBackupInitiator::NewFileNotify(const std::string &path, uint64_t inode) noexcept
{
    if (CheckProcessingMark()) {
        return 0;
    }

    struct stat buf {};
    auto ret = GetAttribute(0, path, buf);
    if (ret != 0) {
        BKG_LOG_ERROR("get file(" << path.c_str() << ") attribute failed : " << ret << " : " << strerror(-ret) << ".");
        return -1;
    }

    FileTrace trace{ path, static_cast<int64_t>(inode) };
    ret = backupTarget->CreateFileAndStageSync(trace, buf);
    if (ret != 0) {
        BKG_LOG_ERROR("sync file(" << path.c_str() << ") failed : " << ret << " : " << strerror(-ret) << ".");
        return -1;
    }

    NotifyProcessMark marker(this);
    BKG_LOG_INFO("submit new file task : " << path);

    struct stat fileStat {};
    ret = CompareTraceFile(trace, fileStat);
    if (ret != 0) {
        return -1;
    }
    backupTarget->UploadFile(trace, fileStat, true);
    return 0;
}

int MemFsBackupInitiator::PreloadFileNotify(const std::string &path) noexcept
{
    if (CheckProcessingMark()) {
        return 0;
    }

    /* open the file to check permissions */
    int tempFd = open(path.c_str(), O_RDONLY | O_NOFOLLOW);
    if (tempFd < 0) {
        BKG_LOG_ERROR("open file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }
    close(tempFd);
    tempFd = -1;

    /* read file fileMeta from ufs */
    struct stat ufsBuf {};
    auto ret = backupTarget->StatFile(path, ufsBuf);
    if (ret != 0) {
        BKG_LOG_ERROR("read file(" << path.c_str() << ") meta failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    uint64_t blkSize = 0UL;
    uint64_t blkCnt = 0UL;
    MemFsApi::GetShareFileCfg(blkSize, blkCnt);
    if (static_cast<uint64_t>(ufsBuf.st_size) > blkSize * blkCnt) {
        errno = ENOMEM;
        BKG_LOG_ERROR("preload file size(" << ufsBuf.st_size << ") is out of memfs memory capacity.");
        return -1;
    }

    /* create and open file for write (memfs) */
    uint64_t inode;
    auto fd = MemFsApi::CreateAndOpenFile(path, inode, ufsBuf.st_mode);
    if (fd < 0) {
        BKG_LOG_ERROR("open file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    /* allocate data blocks */
    uint64_t blockSize = 0;
    std::vector<uint64_t> blockIds;

    ret = MemFsApi::AllocDataBlocks(fd, static_cast<uint64_t>(ufsBuf.st_size), blockIds, blockSize);
    if (ret != 0) {
        BKG_LOG_ERROR("allocate data blocks for file(" << path.c_str() << ") failed(" << errno << " : " <<
            strerror(errno) << ")");
        MemFsApi::DiscardFile(path, fd);
        return -1;
    }

    FileTrace trace{ path, static_cast<int64_t>(inode) };
    NotifyProcessMark marker(this);

    struct stat buf {};
    ret = CompareTraceFile(trace, buf);
    if (ret != 0) {
        return -1;
    }
    SplitAndSubmitTask(fd, ufsBuf, trace, path);

    return 0;
}

int MemFsBackupInitiator::RecordTaskResult(uint64_t taskId, const std::string &path, int taskResult,
    const TaskInfo &taskInfo) noexcept
{
    uint32_t retryCnt = 0;
    auto paraLoadCtxPtr = taskInfo.paraLoadCtx;
    auto pos = paraLoadCtxPtr->taskRetryCntMap.find(taskInfo.offset);
    if (pos == paraLoadCtxPtr->taskRetryCntMap.end()) {
        BKG_LOG_WARN("record task(" << taskId << ") failed, task path(" << path << ") invalid.");
        taskResult = -1;
    } else {
        retryCnt = pos->second;
    }

    if (taskResult < 0) {
        if (retryCnt >= ServiceConfigure::GetInstance().GetBackgroundConfig().backupServiceConfig.retryTimes) {
            paraLoadCtxPtr->failedCnt.fetch_add(1);
        } else {
            paraLoadCtxPtr->taskRetryCntMap[taskInfo.offset] = ++retryCnt;
            return -1;
        }
    } else {
        paraLoadCtxPtr->succeedCnt.fetch_add(1);
    }

    return 0;
}

int MemFsBackupInitiator::MultiTasksWriteFinish(const std::string &path, UFS &ufs) noexcept
{
    auto underFd = open(path.c_str(), O_CREAT | O_WRONLY | O_NOFOLLOW, 0600);
    if (underFd < 0) {
        BKG_LOG_ERROR("open file(" << path << ") to write failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    AioSync(underFd, path);
    close(underFd);

    auto fd = MemFsApi::OpenFile(path, O_RDONLY);
    if (fd < 0) {
        BKG_LOG_ERROR("open file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }
    MemFsApi::SetBackupFinished(fd);
    MemFsApi::CloseFile(fd);
    return RemoveStageFileFromUfs(path, ufs);
}