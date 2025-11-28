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
#include "service_configure.h"
#include "background_log.h"
#include "background_constants.h"
#include "backup_file_manager.h"
#include "background_manager.h"

using namespace ock::bg;
using namespace ock::common;
using BC = BackgroundConstants;

BackgroundManager &BackgroundManager::GetInstance() noexcept
{
    static BackgroundManager instance;
    return instance;
}

int BackgroundManager::Initialize() noexcept
{
    auto &config = config::ServiceConfigure::GetInstance().GetBackgroundConfig();
    backupServiceEnabled = config.backupServiceConfig.enabled;
    if (!backupServiceEnabled) {
        BKG_LOG_INFO("backup service not enabled!");
        return 0;
    }

    auto ret = backup::BackupFileManager::GetInstance().Initialize(config.backupServiceConfig);
    if (ret != 0) {
        BKG_LOG_ERROR("initialize backup service failed(" << ret << ")");
        return -1;
    }

    return 0;
}

void BackgroundManager::Destroy() noexcept
{
    if (backupServiceEnabled) {
        backup::BackupFileManager::GetInstance().Destroy();
        backupServiceEnabled = false;
    }
}