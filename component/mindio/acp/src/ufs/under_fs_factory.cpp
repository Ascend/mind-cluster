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
#include "under_fs_factory.h"

using namespace ock::ufs;

UnderFsFactory &UnderFsFactory::GetInstance() noexcept
{
    static UnderFsFactory instance;
    return instance;
}

void UnderFsFactory::Set(const std::string &name, const std::shared_ptr<BaseFileService> &fs) noexcept
{
    underFileSystems[name] = fs;
}

std::shared_ptr<BaseFileService> UnderFsFactory::Get(const std::string &name) noexcept
{
    auto pos = underFileSystems.find(name);
    if (pos == underFileSystems.end()) {
        return nullptr;
    }

    return pos->second;
}

void UnderFsFactory::SetDefault(const std::shared_ptr<BaseFileService> &fs) noexcept
{
    defaultFileSystem = fs;
}

std::shared_ptr<BaseFileService> UnderFsFactory::GetDefault() noexcept
{
    return defaultFileSystem;
}
