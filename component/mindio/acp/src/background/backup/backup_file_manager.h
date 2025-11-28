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
#ifndef OCK_DFS_BACKUP_FILE_MANAGER_H
#define OCK_DFS_BACKUP_FILE_MANAGER_H

#include <cstdint>

#include <utility>
#include <list>
#include <map>
#include <string>
#include <memory>

#include "service_configure.h"
#include "non_copyable.h"
#include "backup_initiator.h"
#include "backup_target.h"
#include "dfs_adapter.h"

namespace ock {
namespace bg {
namespace backup {
class BackupFileManager : public common::NonCopyable {
public:
    using TargetConf = std::pair<std::string, std::string>;
    using TargetsConf = std::list<TargetConf>;
    static BackupFileManager &GetInstance() noexcept;
    std::shared_ptr<BackupInitiator> GetInitiator(const std::string &name) noexcept;

public:
    int Initialize(const common::config::BackupServiceConfig &backupConfig) noexcept;
    void Destroy() noexcept;

private:
    BackupFileManager() = default;
    ~BackupFileManager() override = default;

    int Validate(const common::config::BackupServiceConfig &backupConfig) noexcept;

private:
    int CreateBackupInstances(const std::map<std::string, TargetsConf> &sourceTarget) noexcept;
    int CreateThreadPool(const common::config::BackupServiceConfig &backupConfig) noexcept;
    std::shared_ptr<BackupTarget> CreateBackupTarget(const std::string &src, const TargetsConf &targetInfo) noexcept;
    static bool CheckBackupInstanceConfig(const common::config::BackupInstance &backup) noexcept;
    static bool CheckBackupFullConfig(const std::map<std::string, TargetsConf> &sourceTarget) noexcept;

private:
    std::shared_ptr<util::RetryTaskPool> pool { nullptr };
    std::shared_ptr<ufs::BaseFileService> dfs { nullptr };
    std::map<std::string, std::shared_ptr<BackupInitiator>> initiators;
};
}
}
}

#endif // OCK_DFS_BACKUP_FILE_MANAGER_H
