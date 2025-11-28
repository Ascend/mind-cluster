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
#include "mem_fs_inode.h"

using namespace ock::memfs;

MemFsInode::MemFsInode(InodeParents ps, uint64_t ino, InodeType tp, mode_t mode, MemFsBMM &bmm) noexcept
    : inode{ ino },
      type{ tp },
      fsBmm{ &bmm },
      accessMode{ mode & 0777 },
      parents{ new (std::nothrow) InodeParents(std::move(ps)) }
{
    if (type == InodeType::INODE_DIR) {
        entries = new (std::nothrow) std::unordered_map<std::string, Dentry>;
    } else {
        blocks = new (std::nothrow) std::vector<uint64_t>;
        RwLockGuard lockGuard{ inodeLock, false };
        writing = true;
        openCount = 0;
    }

    uid = getuid();
    gid = getgid();
    clock_gettime(CLOCK_REALTIME_COARSE, &changedTime);
    accessTime = modifiedTime = changedTime;
}

MemFsInode::MemFsInode(uint64_t pIno, uint64_t ino, std::string name, InodeType tp, mode_t mode, MemFsBMM &bmm) noexcept
    : MemFsInode{ MakeSingleParent(pIno, std::move(name)), ino, tp, mode, bmm }
{}

MemFsInode::~MemFsInode() noexcept
{
    evictHelper.RemoveSelf();
    if (type == InodeType::INODE_DIR) {
        delete entries;
        entries = nullptr;
    } else {
        for (auto &block : *blocks) {
            fsBmm->ReleaseOne(block);
        }
        delete blocks;
        blocks = nullptr;
    }
    delete parents;
    delete acl;
    parents = nullptr;
    fsBmm = nullptr;
    acl = nullptr;
    LOG_INFO("inode-life(ino-" << inode << ") destructor");
}

uint64_t MemFsInode::GetFileSize() noexcept
{
    RwLockGuard lockGuard{ inodeLock, true };
    return fileSize;
}

bool MemFsInode::GetDentry(const std::string &dname, Dentry &dentry) noexcept
{
    if (type != InodeType::INODE_DIR) {
        return false;
    }

    bool found = false;
    RwLockGuard lockGuard{ inodeLock, true };
    auto pos = entries->find(dname);
    if (pos != entries->end()) {
        dentry = pos->second;
        found = true;
    }

    return found;
}

bool MemFsInode::PutDentry(const std::string &dname, uint64_t ino, InodeType tp, int &errorCode) noexcept
{
    if (type != InodeType::INODE_DIR) {
        return false;
    }

    bool success = false;
    RwLockGuard lockGuard{ inodeLock, false };
    if (!removed) {
        auto pos = entries->find(dname);
        if (pos == entries->end()) {
            entries->insert({ dname, Dentry{ ino, tp } });
            __sync_fetch_and_add(&childrenCount, 1);
            success = true;
        } else {
            errorCode = EEXIST;
        }
    } else {
        errorCode = EAGAIN;
    }

    return success;
}

bool MemFsInode::RenameDentry(const std::string &src, const std::string &dst, int &errorCode) noexcept
{
    if (type != InodeType::INODE_DIR) {
        errorCode = ENOTDIR;
        return false;
    }

    RwLockGuard lockGuard{ inodeLock, false };
    if (removed) {
        errorCode = EBUSY;
        return false;
    }

    auto srcPos = entries->find(src);
    if (srcPos == entries->end()) {
        errorCode = ENOENT;
        return false;
    }

    entries->erase(dst);
    entries->emplace(dst, srcPos->second);
    entries->erase(srcPos);

    return true;
}

bool MemFsInode::ExchangeDentry(const std::string &src, const std::string &dst, int &errorCode) noexcept
{
    if (type != InodeType::INODE_DIR) {
        errorCode = ENOTDIR;
        return false;
    }

    RwLockGuard lockGuard{ inodeLock, false };
    if (removed) {
        errorCode = EBUSY;
        return false;
    }

    auto srcPos = entries->find(src);
    if (srcPos == entries->end()) {
        errorCode = ENOENT;
        return false;
    }

    auto dstPos = entries->find(dst);
    if (dstPos == entries->end()) {
        errorCode = ENOENT;
        return false;
    }

    auto temp = srcPos->second;
    srcPos->second = dstPos->second;
    dstPos->second = temp;

    return true;
}

bool MemFsInode::RenameDentry(const std::string &src, const std::shared_ptr<MemFsInode> &pDest, const std::string &dst,
    int &errorCode) noexcept
{
    if (!pDest) {
        errorCode = EINVAL;
        return false;
    }
    if (pDest->inode == inode) {
        return RenameDentry(src, dst, errorCode);
    }

    if (type != InodeType::INODE_DIR || pDest->type != InodeType::INODE_DIR) {
        errorCode = ENOTDIR;
        return false;
    }

    MultiLockGuard lockGuard{ { { inode, &inodeLock }, { pDest->inode, &pDest->inodeLock } }, false };
    if (removed || pDest->removed) {
        errorCode = EBUSY;
        return false;
    }

    auto srcPos = entries->find(src);
    if (srcPos == entries->end()) {
        errorCode = ENOENT;
        return false;
    }

    pDest->entries->erase(dst);
    pDest->entries->emplace(dst, srcPos->second);
    entries->erase(srcPos);

    return true;
}

bool MemFsInode::ExchangeDentry(const std::string &src, const std::shared_ptr<MemFsInode> &pDest,
    const std::string &dst, int &errorCode) noexcept
{
    if (pDest == nullptr) {
        errorCode = EINVAL;
        return false;
    }

    if (pDest->inode == inode) {
        return ExchangeDentry(src, dst, errorCode);
    }

    if (type != InodeType::INODE_DIR || pDest->type != InodeType::INODE_DIR) {
        errorCode = ENOTDIR;
        return false;
    }

    MultiLockGuard lockGuard{ { { inode, &inodeLock }, { pDest->inode, &pDest->inodeLock } }, false };
    if (removed || pDest->removed) {
        errorCode = EBUSY;
        return false;
    }

    auto srcPos = entries->find(src);
    if (srcPos == entries->end()) {
        errorCode = ENOENT;
        return false;
    }

    auto dstPos = pDest->entries->find(dst);
    if (dstPos == pDest->entries->end()) {
        errorCode = ENOENT;
        return false;
    }

    auto temp = srcPos->second;
    srcPos->second = dstPos->second;
    dstPos->second = temp;

    return true;
}

bool MemFsInode::DeleteDentry(const std::string &dname, Dentry &dentry) noexcept
{
    if (type != InodeType::INODE_DIR) {
        return false;
    }

    bool success = false;
    RwLockGuard lockGuard{ inodeLock, false };
    auto pos = entries->find(dname);
    if (pos != entries->end()) {
        dentry = pos->second;
        entries->erase(pos);
        __sync_fetch_and_sub(&childrenCount, 1);
        success = true;
    }

    return success;
}

bool MemFsInode::TryRemove() noexcept
{
    if (type != InodeType::INODE_DIR) {
        return true;
    }

    bool success = false;
    RwLockGuard lockGuard{ inodeLock, false };
    if (entries->empty()) {
        removed = true;
        success = true;
    }

    return success;
}

InodePermission MemFsInode::GetPermInfo() noexcept
{
    std::map<uid_t, uint16_t> usersInAcl;
    std::map<gid_t, uint16_t> groupsInAcl;

    RwLockGuard lockGuard{ inodeLock, true };
    if (acl != nullptr) {
        for (const auto &pair : acl->usersAcl) {
            usersInAcl.emplace(pair.first, pair.second & acl->permMask);
        }
        for (const auto &pair : acl->groupsAcl) {
            groupsInAcl.emplace(pair.first, pair.second & acl->permMask);
        }
    }
    return InodePermission{ type == INODE_DIR, accessMode, uid, gid, usersInAcl, groupsInAcl };
}
