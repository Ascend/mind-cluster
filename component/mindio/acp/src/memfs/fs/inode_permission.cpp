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
#include <sys/stat.h>
#include <unistd.h>
#include <algorithm>
#include "inode_permission.h"

using namespace ock::memfs;

std::shared_ptr<UserGroupCache> InodePermission::userGroupCache{UserGroupCache::GetSystemInstance()};

InodePermission::InodePermission(bool d, mode_t m, uid_t u, gid_t g) noexcept
    : dir{ d }, mode{ m }, owner{ u }, group{ g }
{}

InodePermission::InodePermission(bool d, mode_t m, uid_t u, gid_t g, std::map<uid_t, uint16_t> uAcl,
    std::map<gid_t, uint16_t> gAcl) noexcept
    : dir{ d }, mode{ m }, owner{ u }, group{ g }, aclUsers{ std::move(uAcl) }, aclGroups{ std::move(gAcl) }
{}

bool InodePermission::ContainsPermission(PermitType permit) const noexcept
{
    uid_t opUid;
    if ((opUid = getuid()) == 0) {
        return CheckPermissionForRoot(permit);
    }

    bool hasPermission;
    if (opUid == owner) {
        hasPermission = CheckPermissionForActor(permit, PERM_OWNER_SHIFT);
    } else if (getgid() == group ||
               (userGroupCache != nullptr && userGroupCache->UserInGroup(opUid, group))) {
        hasPermission = CheckPermissionForActor(permit, PERM_GROUP_SHIFT);
    } else {
        hasPermission = CheckPermissionForActor(permit, PERM_OTHERS_SHIFT);
    }

    if (hasPermission) {
        return true;
    }

    // 运行到此处，说明检查的非root用户，且ugo mode中没有授权，需要检查acl中是否授权
    return CheckPermissionInAcl(opUid, permit);
}

bool InodePermission::CheckPermissionForRoot(PermitType permit) const noexcept
{
    /*
     * 对于普通文件，如果不是判断执行权限，root肯定是有的
     * 对于目录root肯定连执行权限也有
     */
    if ((permit & PERM_EXECUTE) == 0 || dir) {
        return true;
    }

    /*
     * 剩下的就是检查root对于普通文件是否有执行权限了
     * 如果ugo三者任何一方有执行权限，root都有执行权限
     */
    if ((mode & (S_IXUSR | S_IXGRP | S_IXOTH)) != 0) {
        return true;
    }

    // ACL中的user列表中任何一方有执行权限，root都有执行权限
    if (std::any_of(aclUsers.begin(), aclUsers.end(),
        [](const std::pair<uid_t, uint16_t> &pair) { return (pair.second & PERM_EXECUTE) == PERM_EXECUTE; })) {
        return true;
    }

    // ACL中的group列表中任何一方有执行权限，root都有执行权限
    return std::any_of(aclGroups.begin(), aclGroups.end(),
        [](const std::pair<gid_t, uint16_t> &pair) { return (pair.second & PERM_EXECUTE) == PERM_EXECUTE; });
}

bool InodePermission::CheckPermissionForActor(PermitType permit, PermitActor actor) const noexcept
{
    return ((mode >> actor) & permit) == permit;
}

bool InodePermission::CheckPermissionInAcl(uid_t opUid, PermitType permit) const noexcept
{
    auto pos = aclUsers.find(opUid);
    if (pos != aclUsers.end() && (pos->second & permit) == permit) {
        return true;
    }

    if (userGroupCache == nullptr) {
        return false;
    }

    return std::any_of(aclGroups.begin(), aclGroups.end(), [opUid, permit](const std::pair<gid_t, uint16_t> &pair) {
        return userGroupCache->UserInGroup(opUid, pair.first) && (pair.second & permit) == permit;
    });
}