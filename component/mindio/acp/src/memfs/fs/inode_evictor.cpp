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
#include <list>

#include "memfs_logger.h"
#include "evict_helper.h"
#include "mem_file_system.h"
#include "mem_fs_inode.h"
#include "inode_evictor.h"

using namespace ock::memfs;

InodeEvictor &InodeEvictor::GetInstance() noexcept
{
    static InodeEvictor instance;
    return instance;
}

InodeEvictor::InodeEvictor() noexcept : inodeHead{ &inodeHead, &inodeHead } {}

int InodeEvictor::Initialize() noexcept
{
    auto ret = EvictHelper::Initialize(&inodeHead);
    if (ret != 0) {
        MFS_LOG_ERROR("initialize EvictHelper failed "<< ret);
        return -1;
    }

    return 0;
}

void InodeEvictor::Destroy() noexcept
{
    EvictHelper::Destroy();
}

void InodeEvictor::RecycleInodes(uint64_t bytes) noexcept
{
    list_head *pos;
    list_head *next;

    uint64_t totalBytes = 0UL;
    std::list<uint64_t> removedInodes;
    auto &locker = EvictHelper::GetLock();
    locker.lock();
    dpax_list_for_each_safe(pos, next, &inodeHead)
    {
        auto helper = dpax_list_entry(pos, EvictHelper, listNode);
        auto iNode = container_of(helper, MemFsInode, evictHelper);

        uint64_t recycleBytes = 0;
        if (RemoveOneInode(pos, recycleBytes)) {
            removedInodes.push_back(iNode->inode);
            dpax_list_del_init(pos);
            totalBytes += recycleBytes;
        }

        if (totalBytes >= bytes) {
            break;
        }
    }
    locker.unlock();

    std::list<std::shared_ptr<MemFsInode>> evictInodes;
    {
        std::unique_lock<std::mutex> lk(MemFileSystem::mfsInstance->inodeMappingLock);
        for (auto &ino : removedInodes) {
            auto it = MemFileSystem::mfsInstance->inodeMapping.find(ino);
            if (it != MemFileSystem::mfsInstance->inodeMapping.end()) {
                evictInodes.emplace_back(it->second);
                MemFileSystem::mfsInstance->inodeMapping.erase(it);
            }
        }
    }

    MFS_LOG_INFO("run recycle inodes need(" << bytes << ") bytes, can remove(" << evictInodes.size() << ") inodes.");
    for (auto &evictInode: evictInodes) {
        LOG_INFO("inode-life(ino-" << evictInode->inode << ") evict self.");
    }
    evictInodes.clear();
}

bool InodeEvictor::RemoveOneInode(list_head *node, uint64_t &recycleBytes) noexcept
{
    auto helper = dpax_list_entry(node, EvictHelper, listNode);
    auto iNode = container_of(helper, MemFsInode, evictHelper);

    auto fs = MemFileSystem::mfsInstance;
    if (fs == nullptr) {
        return false;
    }

    if (iNode->type == InodeType::INODE_DIR) {
        if (!iNode->TryRemove()) {
            return false;
        }
    } else {
        if (__sync_fetch_and_or(&iNode->backup, static_cast<uint8_t>(0)) == static_cast<uint8_t>(0)) {
            return false;
        }
    }

    Dentry d;
    bool allSuccess = true;
    for (auto &parentPos : *iNode->parents) {
        auto parentInode = fs->GetInode(parentPos.first);
        if (parentInode == nullptr) {
            LOG_ERROR("FIX inode-life(ino-" << iNode->inode << ") parent(" << parentPos.first << ") not exist.");
            continue;
        }
        for (auto &name : parentPos.second) {
            if (parentInode->ActiveFile()) {
                LOG_WARN("inode:" << parentInode->inode << " is opening or writing, can not evict");
                allSuccess = false;
                continue;
            }
            LOG_INFO("remove parent:" << parentInode->inode << ", inode:" << iNode->inode << ", name:" << name);
            if (!parentInode->DeleteDentry(name, d)) {
                allSuccess = false;
            }
        }
    }

    if (!allSuccess) {
        return false;
    }

    recycleBytes = 0UL;
    if (iNode->type == InodeType::INODE_REG) {
        recycleBytes = ((iNode->fileSize + iNode->blockSize - 1) / iNode->blockSize) * iNode->blockSize;
    }
    return true;
}