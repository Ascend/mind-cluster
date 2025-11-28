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
#ifndef OCK_DFS_BACKUP_TARGET_H
#define OCK_DFS_BACKUP_TARGET_H

#include <list>
#include <utility>

#include "ufs_api.h"
#include "retry_task_pool.h"
#include "backup_file_tracer.h"
#include "backup_file_view.h"

namespace ock {
namespace bg {
namespace backup {
enum LockFileResult : int {
    TRYLOCK_SUCCESS = 0,
    LOCK_ERROR = 1,
    FILE_BEEN_REMOVED = 2,
    MTIME_NO_CHANGE_TIMEOUT = 3
};

struct UnderFsFileView {
    std::shared_ptr<ufs::BaseFileService> underFs;
    std::shared_ptr<BackupFileView> backupFileView;

    explicit UnderFsFileView(std::shared_ptr<ufs::BaseFileService> ufs) noexcept
        : underFs{ std::move(ufs) }, backupFileView{ std::make_shared<BackupFileView>() }
    {}

    bool AddUploadFileToView(const FileTrace &trace, const struct stat &buf, bool force = false) noexcept;
    bool DoRemoveFile(uint64_t taskId, const std::string &path, int64_t inode, int64_t ufsInode, bool file) noexcept;
};

/**
 * @brief 当启用多线程并行写入时，用于控制多线程的上下文对象。
 */
struct ParallelLoadContext {
public:
    std::unordered_map<uint64_t, uint32_t> taskRetryCntMap;
    std::atomic<uint32_t> succeedCnt{ 0 }; /* succeed task count */
    std::atomic<uint32_t> failedCnt{ 0 };  /* failed task count */

    explicit ParallelLoadContext(uint32_t threadNum) noexcept : totalTaskCnt{ threadNum } {};

    void RecordTaskOffset(uint64_t offset)
    {
        taskRetryCntMap[offset] = 0;
    }

    bool AllTaskFinished() const
    {
        return this->totalTaskCnt <= failedCnt + succeedCnt;
    }

private:
    uint32_t totalTaskCnt{ 0 }; /* split task count(thread count) */
};

struct TaskInfo {
    int fd{ -1 };
    uint64_t threadSize{ 0 };
    uint64_t offset{ 0 };
    uint64_t fileTotalSize{ 0 };
    std::shared_ptr<ParallelLoadContext> paraLoadCtx;

    explicit TaskInfo(int fd, uint64_t thSize, uint64_t offset, uint64_t totalSize,
        const std::shared_ptr<ParallelLoadContext> &paraLoadCtx) noexcept
        : fd(fd), threadSize(thSize), offset(offset), fileTotalSize(totalSize), paraLoadCtx(paraLoadCtx){};
};

class BackupTarget {
public:
    using TaskPool = std::shared_ptr<util::RetryTaskPool>;
    using MUFS = std::list<std::shared_ptr<ufs::BaseFileService>>;

    int Initialize(const std::string &srcName, const TaskPool &pool, const MUFS &mufs) noexcept;
    void Destroy() noexcept;

    void UploadFile(const FileTrace &trace, const struct stat &fileStat, bool force = false) noexcept;
    void RemoveFile(const std::string &path, int64_t inode) noexcept;
    void MakeFileCache(const FileTrace &trace, const TaskInfo &taskInfo) noexcept;

    int CreateDir(const std::string &name, mode_t mode, uid_t owner, gid_t group) noexcept;

    int StatFile(const std::string &path, struct stat &buf) noexcept;
    int StatFile(const std::string &path, struct stat &buf, struct ufs::FileAcl &acl) noexcept;
    int CreateFileAndStageSync(const FileTrace &trace, const struct stat &buf) noexcept;
    int RemoveFileAndStageSync(const FileTrace &trace) noexcept;

    inline bool IsTaskPoolEmpty() noexcept
    {
        return taskPool == nullptr || taskPool->IsTaskPoolEmpty();
    }

private:
    int CheckStgMtime(UnderFsFileView &view, const std::string &stgPath) noexcept;
    int TryLockStg(uint64_t taskId, UnderFsFileView &view, std::shared_ptr<ock::ufs::FileLock> &fileLock,
        const std::string &stgPath, const struct stat &stgLockBuf) noexcept;
    int UnLockStg(std::shared_ptr<ock::ufs::FileLock> &fileLock, const std::string &stgPath) noexcept;
    bool DoBackupFile(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view) noexcept;
    bool DoRealBackupFile(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view,
        struct stat &stgLockBuf) noexcept;
    bool DoBackupFileWrapper(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view) noexcept;
    bool RealBackupAllParentDirectory(uint64_t taskId, const std::string &path, UnderFsFileView &view) noexcept;
    bool RealBackupFile(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view) noexcept;
    static bool CreateOneParent(uint64_t taskId, const std::string &path, const ufs::FileMode &mode,
        UnderFsFileView &view) noexcept;
    static bool CorrectOneParent(uint64_t taskId, const std::string &path, const struct stat &buf,
        const ufs::FileMeta &meta, UnderFsFileView &view) noexcept;
    bool DoMakeFileCache(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view,
        const TaskInfo &taskInfo) noexcept;

private:
    TaskPool taskPool;
    std::list<UnderFsFileView> underFsFileView;
    std::string sourceName;
    static std::atomic<uint64_t> taskIdGen;
};
}
}
}


#endif // OCK_DFS_BACKUP_TARGET_H
