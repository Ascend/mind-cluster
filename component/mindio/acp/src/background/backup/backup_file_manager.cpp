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
#include <set>

#include "service_configure.h"
#include "memfs_api.h"
#include "under_fs_factory.h"
#include "mem_fs_backup_initiator.h"
#include "dfs_backup_initiator.h"
#include "background_log.h"
#include "background_constants.h"
#include "retry_task_pool.h"
#include "backup_file_manager.h"

using namespace ock::bg::util;
using namespace ock::bg::backup;
using BC = ock::bg::BackgroundConstants;

BackupFileManager &BackupFileManager::GetInstance() noexcept
{
    static BackupFileManager instance;
    return instance;
}

std::shared_ptr<BackupInitiator> BackupFileManager::GetInitiator(const std::string &name) noexcept
{
    auto pos = initiators.find(name);
    if (pos == initiators.end()) {
        return nullptr;
    }

    return pos->second;
}

int BackupFileManager::Validate(const common::config::BackupServiceConfig &backupConfig) noexcept
{
    if (backupConfig.threadNum < BC::TASK_MIN_THREAD_NUM || backupConfig.threadNum > BC::TASK_MAX_THREAD_NUM) {
        BKG_LOG_ERROR("backup service thread number(" << backupConfig.threadNum << ") invalid.");
        return -1;
    }

    if (backupConfig.maxFailCntForUnserviceable == 0) {
        BKG_LOG_ERROR("backup service max fail cnt number(" <<
            backupConfig.maxFailCntForUnserviceable << ") invalid, must > 0.");
        return -1;
    }

    return 0;
}

int BackupFileManager::Initialize(const common::config::BackupServiceConfig &backupConfig) noexcept
{
    if (!backupConfig.enabled) {
        BKG_LOG_ERROR("backup service not enabled.");
        return -1;
    }

    if (Validate(backupConfig) != 0) {
        BKG_LOG_ERROR("backup service validate config failed.");
        return -1;
    }

    std::map<std::string, TargetsConf> backupSourceTargets;
    for (auto &backup : backupConfig.backups) {
        if (!backup.opened) {
            continue;
        }

        if (!CheckBackupInstanceConfig(backup)) {
            return -1;
        }
        backupSourceTargets[backup.source].push_back(std::make_pair(backup.destType, backup.destName));
    }

    if (backupSourceTargets.empty()) {
        BKG_LOG_ERROR("no backups valid.");
        return -1;
    }

    if (!CheckBackupFullConfig(backupSourceTargets)) {
        BKG_LOG_ERROR("configure for backups contains same target.");
        return -1;
    }

    dfs = std::make_shared<DfsAdapter>();
    auto ret = CreateThreadPool(backupConfig);
    if (ret != 0) {
        dfs.reset();
        return -1;
    }

    if (CreateBackupInstances(backupSourceTargets) != 0) {
        pool.reset();
        dfs.reset();
        return -1;
    }

    return 0;
}

void BackupFileManager::Destroy() noexcept
{
    pool.reset();
    dfs.reset();
    initiators.clear();
}

int BackupFileManager::CreateBackupInstances(const std::map<std::string, TargetsConf> &sourceTarget) noexcept
{
    for (auto &instance : sourceTarget) {
        auto target = CreateBackupTarget(instance.first, instance.second);
        if (target == nullptr) {
            BKG_LOG_ERROR("create backup for source(" << instance.first.c_str() << ") failed!");
            return -1;
        }

        std::shared_ptr<BackupInitiator> initiator;
        if (instance.first == "mfs") {
            initiator = std::make_shared<MemFsBackupInitiator>();
            memfs::MemFsApi::SetExternalStat(
                [target](const std::string &path, struct stat &buf, memfs::MemfsFileAcl &acl) -> int {
                    ufs::FileAcl ufsAcl;
                    auto ret = target->StatFile(path, buf, ufsAcl);
                    if (ret != 0) {
                        return -errno;
                    }

                    acl.ownerPerm = ufsAcl.ownerPerm;
                    acl.groupPerm = ufsAcl.groupPerm;
                    acl.otherPerm = ufsAcl.otherPerm;
                    acl.permMask = ufsAcl.permMask;
                    acl.usersAcl = std::move(ufsAcl.users);
                    acl.groupsAcl = std::move(ufsAcl.groups);
                    return 0;
                });
        } else {
#ifdef __BUILD_BOTH_MFS_DFS__
            initiator = std::make_shared<DfsBackupInitiator>();
#else
            BKG_LOG_ERROR("unsupported backup initiator(" << instance.first.c_str() << ")");
            return -1;
#endif
        }

        auto ret = initiator->Initialize(pool, target);
        if (ret != 0) {
            BKG_LOG_ERROR("initialize backup initiator for(" << instance.first.c_str() << ") failed!");
            return -1;
        }

        initiators[instance.first] = initiator;
    }

    return 0;
}

std::shared_ptr<BackupTarget> BackupFileManager::CreateBackupTarget(const std::string &src,
    const TargetsConf &targetInfo) noexcept
{
    std::list<std::shared_ptr<ufs::BaseFileService>> targetFs;
    for (auto &e : targetInfo) {
        if (e.first == "dfs") {
            targetFs.push_back(dfs);
        } else {
            targetFs.push_back(ufs::UnderFsFactory::GetInstance().Get(e.second));
        }
    }

    auto backupTarget = std::make_shared<BackupTarget>();
    auto ret = backupTarget->Initialize(src, pool, targetFs);
    if (ret != 0) {
        return nullptr;
    }

    return backupTarget;
}

int BackupFileManager::CreateThreadPool(const common::config::BackupServiceConfig &backupConfig) noexcept
{
    RetryTaskPool::RetryTaskConfig config;
    config.name = "MindBG";
    config.autoEvictFile = backupConfig.autoEvictFile;
    config.thCnt = backupConfig.threadNum;
    config.maxFailCntForUnserviceable = backupConfig.maxFailCntForUnserviceable;
    config.retryTimes = backupConfig.retryTimes;
    config.retryIntervalSec = backupConfig.retryIntervalSec;
    config.firstWaitMs = BC::TASK_RETRY_FIRST_WAIT_MILL_SECONDS;

    if (config.thCnt == 0) {
        BKG_LOG_ERROR("create thread pool with size(0)");
        return -1;
    }

    pool = std::make_shared<RetryTaskPool>(config);
    if (pool == nullptr) {
        BKG_LOG_ERROR("create thread pool with size(" << config.thCnt << ") failed.");
        return -1;
    }

    auto ret = pool->Start();
    if (ret != 0) {
        BKG_LOG_ERROR("start thread pool failed: " << ret);
        pool.reset();
        return -1;
    }

    return 0;
}

bool BackupFileManager::CheckBackupInstanceConfig(const common::config::BackupInstance &backup) noexcept
{
    if (backup.source != "mfs" && backup.source != "dfs") {
        BKG_LOG_ERROR("unknown backup source(" << backup.source.c_str() << ")");
        return false;
    }

    if (backup.source == backup.destType) {
        BKG_LOG_ERROR("backup cannot to itself(" << backup.source.c_str() << ")");
        return false;
    }

    if (backup.destType != "under_fs" && backup.destType != "dfs") {
        BKG_LOG_ERROR("unknown backup source(" << backup.destType.c_str() << ")");
        return false;
    }

    if (backup.destType == "dfs") {
        return true;
    }

    auto ufs = ock::ufs::UnderFsFactory::GetInstance().Get(backup.destName);
    if (ufs == nullptr) {
        BKG_LOG_ERROR("backup target ufs(" << backup.destName.c_str() << ") not exist");
        return false;
    }

    return true;
}

bool BackupFileManager::CheckBackupFullConfig(const std::map<std::string, TargetsConf> &sourceTarget) noexcept
{
    for (auto &e : sourceTarget) {
        int dfsCount = 0;
        std::map<std::string, int> ufsNameCount;
        for (auto &d : e.second) {
            if (d.first == "dfs") {
                dfsCount++;
            } else {
                ufsNameCount[d.second]++;
            }
        }

        if (dfsCount > 1) {
            return false;
        }

        for (auto &count : ufsNameCount) {
            if (count.second > 1) {
                return false;
            }
        }
    }

    return true;
}
