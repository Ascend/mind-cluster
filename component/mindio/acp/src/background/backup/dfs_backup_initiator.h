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
#ifndef OCK_DFS_DFS_BACKUP_INITIATOR_H
#define OCK_DFS_DFS_BACKUP_INITIATOR_H

#include "backup_initiator.h"
#include "backup_target.h"

namespace ock {
namespace bg {
namespace backup {
class DfsBackupInitiator : public BackupInitiator {
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
    int OpenFileNotify(int fd, const std::string &path, int64_t inode, bool dir) noexcept;
    void CloseFileNotify(int fd, bool dir) noexcept;
    void CreateDirectoryNotify(const std::string &path, int64_t inode) noexcept;
    void RemoveFileNotify(const std::string &path, int64_t inode) noexcept;

private:
    static __thread bool marked;
};
}
}
}


#endif // OCK_DFS_DFS_BACKUP_INITIATOR_H
