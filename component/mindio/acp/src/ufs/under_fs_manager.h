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
#ifndef OCK_DFS_UNDER_FS_MANAGER_H
#define OCK_DFS_UNDER_FS_MANAGER_H

#include "non_copyable.h"

namespace ock {
namespace ufs {
class UnderFsManager : public common::NonCopyable {
public:
    static UnderFsManager &GetInstance() noexcept;

public:
    int Initialize() noexcept;
    void Destroy() noexcept;

private:
    explicit UnderFsManager() = default;
    ~UnderFsManager() override = default;
};
}
}


#endif // OCK_DFS_UNDER_FS_MANAGER_H
