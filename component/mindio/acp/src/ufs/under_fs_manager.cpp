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
#include "ufs_log.h"
#include "service_configure.h"
#include "under_fs_factory.h"
#include "pacific_adapter.h"
#include "under_fs_manager.h"

using namespace ock::common;
using namespace ock::ufs;

UnderFsManager &UnderFsManager::GetInstance() noexcept
{
    static UnderFsManager instance;
    return instance;
}

int UnderFsManager::Initialize() noexcept
{
    auto &ufsConfig = config::ServiceConfigure::GetInstance().GetUnderFileSystemConfig();
    for (auto &inst : ufsConfig.instances) {
        if (inst.second.type != "pacific") {
            UFS_LOG_ERROR("invalid under fs name(" << inst.second.name.c_str() << ") type(" <<
                inst.second.type.c_str() << ")");
            return -1;
        }

        auto &ops = inst.second.options;
        auto it = ops.find("mount_path");
        if (it == ops.end()) {
            UFS_LOG_ERROR("under fs(" << inst.second.name.c_str() << ") no mount_path");
            return -1;
        }

        auto fs = std::make_shared<PacificAdapter>(it->second);
        UnderFsFactory::GetInstance().Set(inst.second.name, fs);
        UFS_LOG_INFO("create pacific instance name(" << inst.second.name.c_str() <<
            ") mount_path(" << it->second.c_str() << ")");
    }

    auto defaultFs = UnderFsFactory::GetInstance().Get(ufsConfig.defaultName);
    if (defaultFs == nullptr) {
        UFS_LOG_ERROR("default under fs(" << ufsConfig.defaultName.c_str() << ") not found");
        return -1;
    }

    UnderFsFactory::GetInstance().SetDefault(defaultFs);
    UFS_LOG_INFO("default under fs(" << ufsConfig.defaultName.c_str() << ")");
    return 0;
}

void UnderFsManager::Destroy() noexcept {}