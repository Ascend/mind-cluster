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
#ifndef OCK_DFS_INODE_EVICTOR_H
#define OCK_DFS_INODE_EVICTOR_H

#include "dpax_list.h"
#include "non_copyable.h"

namespace ock {
namespace memfs {
class InodeEvictor : public common::NonCopyable {
public:
    static InodeEvictor &GetInstance() noexcept;

public:
    int Initialize() noexcept;
    void Destroy() noexcept;
    void RecycleInodes(uint64_t bytes) noexcept;

private:
    static bool RemoveOneInode(list_head *node, uint64_t &recycleBytes) noexcept;

private:
    InodeEvictor() noexcept;
    ~InodeEvictor() override = default;

private:
    list_head inodeHead;
};
}
}


#endif // OCK_DFS_INODE_EVICTOR_H
