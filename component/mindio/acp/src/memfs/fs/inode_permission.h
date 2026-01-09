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
#ifndef OCK_DFS_INODE_PERMISSION_H
#define OCK_DFS_INODE_PERMISSION_H

#include <sys/types.h>
#include <cstdint>
#include "user_group_cache.h"

namespace ock {
namespace memfs {
enum PermitType : uint8_t {
    PERM_EXECUTE = 01,
    PERM_WRITE = 02,
    PERM_READ = 04,
    PERM_MASK = PERM_EXECUTE | PERM_WRITE | PERM_READ
};

enum PermitActor : int8_t {
    PERM_OTHERS_SHIFT = 0,
    PERM_GROUP_SHIFT = 3,
    PERM_OWNER_SHIFT = 6
};

class InodePermission {
public:
    InodePermission(bool d, mode_t m, uid_t u, gid_t g) noexcept;
    InodePermission(bool d, mode_t m, uid_t u, gid_t g, std::map<uid_t, uint16_t> uAcl,
        std::map<gid_t, uint16_t> gAcl) noexcept;

public:
    bool ContainsPermission(PermitType permit) const noexcept;
    static void SetUserGroupCache(const std::shared_ptr<UserGroupCache> &cache) noexcept
    {
        if (cache != nullptr) {
            userGroupCache = cache;
        }
    }

private:
    bool CheckPermissionForRoot(PermitType permit) const noexcept;
    bool CheckPermissionForActor(PermitType permit, PermitActor actor) const noexcept;
    bool CheckPermissionInAcl(uid_t opUid, PermitType permit) const noexcept;

private:
    const bool dir;
    const mode_t mode;
    const uid_t owner;
    const gid_t group;
    const std::map<uid_t, uint16_t> aclUsers;
    const std::map<gid_t, uint16_t> aclGroups;
    static std::shared_ptr<UserGroupCache> userGroupCache;
};
}
}


#endif // OCK_DFS_INODE_PERMISSION_H
