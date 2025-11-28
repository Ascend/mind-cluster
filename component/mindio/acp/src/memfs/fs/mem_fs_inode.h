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
#ifndef OCK_DFS_MEM_FS_INODE_H
#define OCK_DFS_MEM_FS_INODE_H

#include <cstdint>
#include <ctime>
#include <limits>
#include <utility>
#include <vector>
#include <list>
#include <string>
#include <unordered_map>
#include "mem_fs_constants.h"
#include "memfs_api.h"
#include "bmm.h"
#include "common_locker.h"
#include "evict_helper.h"
#include "inode_permission.h"

namespace ock {
namespace memfs {
enum InodeType : int {
    INODE_REG = 1000, // regular file
    INODE_DIR,        // directory
};

struct Dentry {
    uint64_t inode;
    InodeType type;

    Dentry(uint64_t ino, InodeType tp) noexcept : inode(ino), type(tp) {}
    Dentry() noexcept : inode(MemFsConstants::INODE_INVALID), type{ INODE_REG } {}
};

using InodeAcl = MemfsFileAcl;
using InodeParents = std::map<uint64_t, std::set<std::string>>;

struct MemFsInode {
    const uint64_t inode;
    const InodeType type;
    MemFsBMM *fsBmm = nullptr;

    uid_t uid{ 0 };
    gid_t gid{ 0 };
    mode_t accessMode{ 0777 };
    struct timespec accessTime {
        0, 0
    };
    struct timespec modifiedTime {
        0, 0
    };
    struct timespec changedTime {
        0, 0
    };
    ReadWriteLock inodeLock;
    uint64_t fileSize{ 0UL };
    uint64_t blockSize{ 0UL };
    int32_t childrenCount{ 0 };
    bool removed{ false };
    bool writing{ false };
    uint8_t backup{ 0 };
    int32_t openCount{ 0 };
    EvictHelper evictHelper{};

    InodeParents *parents = nullptr;
    std::unordered_map<std::string, Dentry> *entries{ nullptr };
    std::vector<uint64_t> *blocks{ nullptr };
    InodeAcl *acl{ nullptr };

    MemFsInode(InodeParents parents, uint64_t ino, InodeType tp, mode_t mode, MemFsBMM &bmm) noexcept;
    MemFsInode(uint64_t pIno, uint64_t ino, std::string name, InodeType tp, mode_t mode, MemFsBMM &bmm) noexcept;
    ~MemFsInode() noexcept;

    inline InodeType GetType() const noexcept
    {
        return type;
    }

    inline uint64_t GetInodeNumber() const noexcept
    {
        return inode;
    }

    inline bool ActiveFile() noexcept
    {
        RwLockGuard lockGuard{ inodeLock, false };
        return writing || openCount > 0;
    }

    inline bool Valid() const noexcept
    {
        if (parents == nullptr) {
            return false;
        }
        if (type == InodeType::INODE_DIR) {
            return entries != nullptr;
        }
        return blocks != nullptr;
    }

    inline bool AddParent(uint64_t pIno, std::string name) noexcept
    {
        RwLockGuard lockGuard{ inodeLock, false };
        auto pos = parents->find(pIno);
        if (pos == parents->end()) {
            std::set<std::string> values;
            values.emplace(std::move(name));
            parents->emplace(pIno, std::move(values));
            return true;
        }

        auto valPos = pos->second.find(name);
        if (valPos == pos->second.end()) {
            pos->second.emplace(std::move(name));
            return true;
        }

        return false;
    }

    inline bool RemoveParent(uint64_t pIno, const std::string &name, uint64_t &leftParentCount) noexcept
    {
        bool found = false;
        RwLockGuard lockGuard{ inodeLock, false };
        auto pos = parents->find(pIno);
        if (pos != parents->end()) {
            auto valPos = pos->second.find(name);
            if (valPos != pos->second.end()) {
                pos->second.erase(valPos);
                if (pos->second.empty()) {
                    parents->erase(pos);
                }
                found = true;
            }
        }

        if (!found) {
            return false;
        }

        leftParentCount = 0;
        for (auto &inoPos : *parents) {
            leftParentCount += inoPos.second.size();
        }

        return true;
    }

    inline bool RemoveParent(uint64_t pIno, const std::string &name) noexcept
    {
        uint64_t temp = 0;
        return RemoveParent(pIno, name, temp);
    }

    inline uint32_t GetLinkCountInLock() const noexcept
    {
        auto leftParentCount = 0UL;
        for (auto &inoPos : *parents) {
            for (auto &name : inoPos.second) {
                leftParentCount++;
            }
        }

        return leftParentCount;
    }

    static inline InodeParents MakeSingleParent(uint64_t pIno, std::string name) noexcept
    {
        InodeParents result;
        std::set<std::string> values;
        values.emplace(std::move(name));
        result.emplace(pIno, std::move(values));
        return result;
    }

    uint64_t GetFileSize() noexcept;

    bool GetDentry(const std::string &dname, Dentry &dentry) noexcept;
    bool PutDentry(const std::string &dname, uint64_t ino, InodeType tp, int &errorCode) noexcept;
    bool RenameDentry(const std::string &src, const std::string &dst, int &errorCode) noexcept;
    bool ExchangeDentry(const std::string &src, const std::string &dst, int &errorCode) noexcept;
    bool RenameDentry(const std::string &src, const std::shared_ptr<MemFsInode> &pDest, const std::string &dst,
        int &errorCode) noexcept;
    bool ExchangeDentry(const std::string &src, const std::shared_ptr<MemFsInode> &pDest, const std::string &dst,
        int &errorCode) noexcept;
    bool DeleteDentry(const std::string &dname, Dentry &dentry) noexcept;
    bool TryRemove() noexcept;
    InodePermission GetPermInfo() noexcept;
};
}
}

#endif // OCK_DFS_MEM_FS_INODE_H
