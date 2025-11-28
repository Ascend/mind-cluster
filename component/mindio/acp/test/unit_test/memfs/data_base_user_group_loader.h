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

#ifndef OCKIO_DATA_BASE_USER_GROUP_LOADER_H
#define OCKIO_DATA_BASE_USER_GROUP_LOADER_H

#include "user_group_cache.h"

class DataBaseUserGroupLoader : public ock::memfs::UserGroupLoader {
public:
    DataBaseUserGroupLoader() noexcept;
    void AddUser(const std::string &name, uid_t user, gid_t group) noexcept;
    void AddGroup(const std::string &name, gid_t group) noexcept;
    void AddUserToGroup(gid_t group, uid_t user) noexcept;
    void Clear() noexcept;
    ock::memfs::UserInfoPtr LoadUser(uid_t uid) const noexcept override;
    ock::memfs::UserInfoPtr LoadUser(const std::string &userName) const noexcept override;
    ock::memfs::GroupInfoPtr LoadGroup(gid_t gid) const noexcept override;
    ock::memfs::GroupInfoPtr LoadGroup(const std::string &groupName) const noexcept override;

    void SetGetUserBufferRange(size_t needSize) noexcept;
    void SetGetGroupBufferRange(size_t needSize) noexcept;

public:
    static DataBaseUserGroupLoader &GetDefault() noexcept;

public:
    static int GetPwNam(const std::string &name, struct passwd *pwd, char *buf, size_t bufLen, struct passwd **result);
    static int GetPwUid(uid_t uid, struct passwd *pwd, char *buf, size_t bufLen, struct passwd **result);
    static int GetGrNam(const std::string &name, struct group *grp, char *buf, size_t bufLen, struct group **result);
    static int GetGrGid(gid_t gid, struct group *grp, char *buf, size_t bufLen, struct group **result);

private:
    mutable std::mutex mutex;
    std::map<std::string, uid_t> userNames;
    std::map<uid_t, std::pair<std::string, gid_t>> users;
    std::map<std::string, gid_t> groupNames;
    std::map<gid_t, std::pair<std::string, std::unordered_set<uid_t>>> groups;
    size_t userInfoNeedSize;
    size_t groupInfoNeedSize;
    static DataBaseUserGroupLoader defaultDataBase;
};

#endif // OCKIO_DATA_BASE_USER_GROUP_LOADER_H
