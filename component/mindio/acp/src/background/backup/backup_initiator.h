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
#ifndef OCK_DFS_BACKUP_INITIATOR_H
#define OCK_DFS_BACKUP_INITIATOR_H

#include <sys/types.h>
#include <sys/stat.h>

#include <cstdint>
#include <string>
#include <atomic>

#include "ufs_api.h"
#include "backup_file_tracer.h"
#include "retry_task_pool.h"
#include "backup_file_view.h"
#include "backup_target.h"

namespace ock {
namespace bg {
namespace backup {
enum CompareFileResult : int {
    FILE_SAME,
    FILE_NOT_EXIST,
    IO_FAILED,
    INODE_NOT_MATCH
};

class BackupInitiator {
public:
    using UFS = std::shared_ptr<ufs::BaseFileService>;
    using TaskPool = std::shared_ptr<util::RetryTaskPool>;

    int Initialize(const TaskPool &pool, const std::shared_ptr<BackupTarget> &target) noexcept;
    void Destroy() noexcept;
    CompareFileResult CompareFile(const std::string &path, int64_t inode, struct stat &buf) noexcept;

    virtual int GetAttribute(uint64_t taskId, const std::string &path, struct stat &buf) noexcept = 0;
    virtual int MultiCopyFileToUfs(uint64_t taskId, const std::string &path, UFS &ufs) noexcept = 0;
    virtual int CopyFileToMemfs(uint64_t taskId, const std::string &path, UFS &ufs,
        const TaskInfo &taskInfo) noexcept = 0;
    virtual int RecordToMemfsTaskResult(uint64_t taskId, const std::string &path, int taskResult,
        const TaskInfo &taskInfo) noexcept = 0;
    virtual void SetProcessingMark() noexcept = 0;
    virtual void ClearProcessingMark() noexcept = 0;
    virtual bool CheckProcessingMark() noexcept = 0;

protected:
    virtual int Prepare() noexcept = 0;
    void SubmitRemoveFileTask(const std::string &path, int64_t inode) noexcept;
    int CompareTraceFile(const FileTrace &trace, struct stat &fileStat) noexcept;

protected:
    TaskPool taskPool;
    std::shared_ptr<BackupTarget> backupTarget;
    BackupFileTracer fileTracer;
};

class NotifyProcessMark {
public:
    explicit NotifyProcessMark(BackupInitiator *init) noexcept;
    virtual ~NotifyProcessMark() noexcept;

private:
    BackupInitiator *initiator;
};
}
}
}


#endif // OCK_DFS_BACKUP_INITIATOR_H
