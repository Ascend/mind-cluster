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
#include "memfs_logger.h"
#include "evict_helper.h"

namespace ock {
namespace memfs {

std::recursive_mutex EvictHelper::mutexLock;
list_head *EvictHelper::inodeHeader{ nullptr };

int EvictHelper::Initialize(list_head *head) noexcept
{
    inodeHeader = head;
    MFS_LOG_INFO("initialize evict helper success");
    return 0;
}

void EvictHelper::Destroy() noexcept
{
    inodeHeader = nullptr;
}

std::recursive_mutex &EvictHelper::GetLock() noexcept
{
    return mutexLock;
}

void EvictHelper::AddToTail() noexcept
{
    if (inodeHeader == nullptr) {
        return;
    }

    std::unique_lock<std::recursive_mutex> lk(mutexLock);
    dpax_list_add_tail(&listNode, inodeHeader);
}

void EvictHelper::MoveToTail() noexcept
{
    if (inodeHeader == nullptr) {
        return;
    }

    std::unique_lock<std::recursive_mutex> lk(mutexLock);
    dpax_list_del_init(&listNode);
    dpax_list_add_tail(&listNode, inodeHeader);
}

void EvictHelper::RemoveSelf() noexcept
{
    std::unique_lock<std::recursive_mutex> lk(mutexLock);
    dpax_list_del_init(&listNode);
}

}
}