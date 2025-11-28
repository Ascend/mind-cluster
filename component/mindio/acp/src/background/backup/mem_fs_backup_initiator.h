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
#ifndef OCK_DFS_MEM_FS_BACKUP_INITIATOR_H
#define OCK_DFS_MEM_FS_BACKUP_INITIATOR_H

#include "backup_initiator.h"

namespace ock {
namespace bg {
namespace backup {
class MemFsBackupInitiator : public BackupInitiator {
public:
    int GetAttribute(uint64_t taskId, const std::string &path, struct stat &buf) noexcept override;
    int MultiCopyFileToUfs(uint64_t taskId, const std::string &path, UFS &ufs) noexcept override;
    int CopyFileToMemfs(uint64_t taskId, const std::string &path, UFS &ufs, const TaskInfo &taskInfo) noexcept override;
    int RecordToMemfsTaskResult(uint64_t taskId, const std::string &path, int taskResult,
        const TaskInfo &taskInfo) noexcept override;
    void SetProcessingMark() noexcept override;
    void ClearProcessingMark() noexcept override;
    bool CheckProcessingMark() noexcept override;

protected:
    int Prepare() noexcept override;

private:
    int MultiTasksWriteFinish(const std::string &path, UFS &ufs) noexcept;
    int MultiTasksDoWrite(const std::string &path, const TaskInfo &taskInfo, const struct stat &fileStat,
        UFS ufs) noexcept;
    int SplitUploadFileTask(const std::string &path, const struct stat &fileStat, UFS &ufs) noexcept;
    int RecordTaskResult(uint64_t taskId, const std::string &path, int taskResult, const TaskInfo &taskInfo) noexcept;
    int OpenFileNotify(int fd, const std::string &path, int flags, uint64_t inode) noexcept;
    void CloseFileNotify(int fd, bool abnormal) noexcept;
    int NewFileNotify(const std::string &path, uint64_t inode) noexcept;
    int RemoveStageFileFromUfs(const std::string &path, const UFS &ufs) noexcept;
    int PreloadFileNotify(const std::string &path) noexcept;
    void SplitAndSubmitTask(int fd, const struct stat &ufsBuf, const FileTrace &trace,
        const std::string &path) noexcept;

private:
    static __thread bool marked;
    static std::atomic<uint64_t> taskIdGen;
};
}
}
}

#endif // OCK_DFS_MEM_FS_BACKUP_INITIATOR_H
