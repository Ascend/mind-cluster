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
#ifndef OCK_DFS_UNDER_FS_FACTORY_H
#define OCK_DFS_UNDER_FS_FACTORY_H

#include "non_copyable.h"
#include "ufs_api.h"

namespace ock {
namespace ufs {
class UnderFsFactory : public common::NonCopyable {
public:
    static UnderFsFactory &GetInstance() noexcept;

public:
    void Set(const std::string &name, const std::shared_ptr<BaseFileService> &fs) noexcept;
    std::shared_ptr<BaseFileService> Get(const std::string &name) noexcept;

    void SetDefault(const std::shared_ptr<BaseFileService> &fs) noexcept;
    std::shared_ptr<BaseFileService> GetDefault() noexcept;

private:
    explicit UnderFsFactory() = default;
    ~UnderFsFactory() override = default;

private:
    std::map<std::string, std::shared_ptr<BaseFileService>> underFileSystems;
    std::shared_ptr<BaseFileService> defaultFileSystem{ nullptr };
};
}
}


#endif // OCK_DFS_UNDER_FS_FACTORY_H
