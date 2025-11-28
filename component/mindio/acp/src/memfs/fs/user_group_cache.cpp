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
#include "common_includes.h"
#include "user_group_cache.h"

using namespace ock::memfs;

static constexpr uint32_t READ_BUF_START_SIZE = 4096;
static constexpr uint32_t READ_BUF_MAX_SIZE = 4U * 1024U * 1024U;
static constexpr int64_t CACHE_ITEM_EXPIRED_MACRO_SECONDS = 5L * 1000L * 1000L;

UserInfo::UserInfo(uid_t uid, gid_t gid) noexcept : userId{ uid }, groupId{ gid } {}

uid_t UserInfo::GetUserId() const noexcept
{
    return userId;
}

gid_t UserInfo::GetGroupId() const noexcept
{
    return groupId;
}

GroupInfo::GroupInfo(gid_t gid, std::unordered_set<uid_t> users) noexcept
    : groupId{ gid }, usersInGroup{ std::move(users) }
{}


gid_t GroupInfo::GetGroupId() const noexcept
{
    return groupId;
}

bool GroupInfo::Contains(uid_t uid) const noexcept
{
    return usersInGroup.find(uid) != usersInGroup.end();
}

std::shared_ptr<UserGroupCache> UserGroupCache::GetSystemInstance() noexcept
{
    auto systemLoader = std::make_shared<SystemUserGroupLoader>();
    static auto instance = MakeRef<UserGroupCache>(systemLoader);
    return instance;
}

UserGroupCache::UserGroupCache(const std::shared_ptr<UserGroupLoader> &loader, int64_t expiresUs) noexcept
    : userGroupLoader{ loader },
      expiresMacroSeconds{ expiresUs <= 0L ? CACHE_ITEM_EXPIRED_MACRO_SECONDS : expiresUs },
      users{ [this](uid_t uid) { return userGroupLoader->LoadUser(uid); } },
      groups{ [this](gid_t gid) { return userGroupLoader->LoadGroup(gid); } }
{}

UserGroupCache::~UserGroupCache() noexcept = default;

bool UserGroupCache::UserInGroup(uid_t user, gid_t group) noexcept
{
    auto userInfo = GetUsersWithCache(user);
    if (userInfo == nullptr) {
        LOG_ERROR("user id:" << user << " not found!");
        return false;
    }

    if (userInfo->GetGroupId() == group) {
        return true;
    }

    auto groupInfo = GetGroupWithCache(group);
    if (groupInfo == nullptr) {
        LOG_ERROR("group id:" << group << " not found!");
        return false;
    }

    return groupInfo->Contains(user);
}

UserInfoPtr UserGroupCache::GetUsersWithCache(uid_t user) noexcept
{
    return users.Get(user, expiresMacroSeconds);
}

GroupInfoPtr UserGroupCache::GetGroupWithCache(gid_t group) noexcept
{
    return groups.Get(group, expiresMacroSeconds);
}

UserInfoPtr SystemUserGroupLoader::LoadUser(uid_t uid) const noexcept
{
    return GetUser([&uid](struct passwd &user, char *buffer, size_t bufferSize,
        struct passwd *&pUser) { return SystemUserGroupWrapper::GetPwUid(uid, &user, buffer, bufferSize, &pUser); },
        std::to_string(uid));
}

UserInfoPtr SystemUserGroupLoader::LoadUser(const std::string &userName) const noexcept
{
    return GetUser(
        [&userName](struct passwd &user, char *buffer, size_t bufferSize, struct passwd *&pUser) {
            return SystemUserGroupWrapper::GetPwNam(userName, &user, buffer, bufferSize, &pUser);
        },
        userName);
}

GroupInfoPtr SystemUserGroupLoader::LoadGroup(gid_t gid) const noexcept
{
    return GetGroup([&gid](struct group &group, char *buffer, size_t bufferSize,
        struct group *&pGroup) { return SystemUserGroupWrapper::GetGrGid(gid, &group, buffer, bufferSize, &pGroup); },
        std::to_string(gid));
}

GroupInfoPtr SystemUserGroupLoader::LoadGroup(const std::string &groupName) const noexcept
{
    return GetGroup(
        [&groupName](struct group &group, char *buffer, size_t bufferSize, struct group *&pGroup) {
            return SystemUserGroupWrapper::GetGrNam(groupName, &group, buffer, bufferSize, &pGroup);
        },
        groupName);
}

GroupInfoPtr SystemUserGroupLoader::ParseGroupInfo(struct group *group) noexcept
{
    if (group == nullptr) {
        return nullptr;
    }

    std::unordered_set<uid_t> usersInGroup;
    for (auto name = group->gr_mem; *name != nullptr; name++) {
        uid_t userId;
        auto ret = GetUidByName(*name, userId);
        if (ret != 0) {
            return nullptr;
        }

        usersInGroup.emplace(userId);
    }

    return MakeRef<GroupInfo>(group->gr_gid, usersInGroup);
}

int SystemUserGroupLoader::GetUidByName(const std::string &name, uid_t &userId) noexcept
{
    struct passwd pw {};
    struct passwd *pwp;
    auto bufferSize = READ_BUF_START_SIZE;
    while (true) {
        if (bufferSize > READ_BUF_MAX_SIZE) {
            LOG_ERROR("read for user:" << name << " size too large");
            return ERANGE;
        }

        auto buffer = new (std::nothrow) char[bufferSize];
        if (buffer == nullptr) {
            LOG_ERROR("allocate buffer with size : " << bufferSize << " failed");
            return ENOMEM;
        }

        std::unique_ptr<char[]> holder(buffer);
        auto ret = SystemUserGroupWrapper::GetPwNam(name, &pw, buffer, bufferSize, &pwp);
        if (ret == 0) {
            if (pwp != nullptr) {
                userId = pwp->pw_uid;
                return 0;
            }

            return ENOENT;
        }

        if (ret == ERANGE) {
            bufferSize <<= 1;
            continue;
        }

        LOG_ERROR("read for user name:" << name << " unknown error: " << ret << ":" << strerror(ret));
        return ret;
    }
}

GroupInfoPtr SystemUserGroupLoader::GetGroup(const LowGetGroupFun &low, const std::string &info) noexcept
{
    struct group grp {};
    struct group *pgrp;
    auto bufferSize = READ_BUF_START_SIZE;
    while (true) {
        if (bufferSize > READ_BUF_MAX_SIZE) {
            LOG_ERROR("read for group(" << info << ") size too large");
            return nullptr;
        }

        auto buffer = new (std::nothrow) char[bufferSize];
        if (buffer == nullptr) {
            LOG_ERROR("allocate buffer with size : " << bufferSize << " failed");
            return nullptr;
        }

        std::unique_ptr<char[]> holder(buffer);
        auto ret = low(grp, buffer, bufferSize, pgrp);
        if (ret == 0) {
            return ParseGroupInfo(pgrp);
        }

        if (ret == ERANGE) {
            bufferSize <<= 1;
            continue;
        }

        LOG_ERROR("read for group(" << info << ") unknown error: " << ret << ":" << strerror(ret));
        return nullptr;
    }
}

UserInfoPtr SystemUserGroupLoader::GetUser(const LowGetUserFun &low, const std::string &info) noexcept
{
    struct passwd pw {};
    struct passwd *pwp;
    auto bufferSize = READ_BUF_START_SIZE;

    while (true) {
        if (bufferSize > READ_BUF_MAX_SIZE) {
            LOG_ERROR("read for user(" << info << ") size too large");
            return nullptr;
        }

        auto buffer = new (std::nothrow) char[bufferSize];
        if (buffer == nullptr) {
            LOG_ERROR("allocate buffer with size : " << bufferSize << " failed");
            return nullptr;
        }

        std::unique_ptr<char[]> holder(buffer);
        auto ret = low(pw, buffer, bufferSize, pwp);
        if (ret == 0) {
            if (pwp == nullptr) {
                return nullptr;
            }
            return MakeRef<UserInfo>(pwp->pw_uid, pwp->pw_gid);
        }

        if (ret == ERANGE) {
            bufferSize <<= 1;
            continue;
        }

        LOG_ERROR("read for user(" << info << ") unknown error: " << ret << ":" << strerror(ret));
        return nullptr;
    }
}

int SystemUserGroupWrapper::GetPwNam(const std::string &name, struct passwd *pwd, char *buf, size_t bufLen,
    struct passwd **result)
{
    return getpwnam_r(name.c_str(), pwd, buf, bufLen, result);
}

int SystemUserGroupWrapper::GetPwUid(uid_t uid, struct passwd *pwd, char *buf, size_t bufLen, struct passwd **result)
{
    return getpwuid_r(uid, pwd, buf, bufLen, result);
}

int SystemUserGroupWrapper::GetGrNam(const std::string &name, struct group *grp, char *buf, size_t bufLen,
    struct group **result)
{
    return getgrnam_r(name.c_str(), grp, buf, bufLen, result);
}

int SystemUserGroupWrapper::GetGrGid(gid_t gid, struct group *grp, char *buf, size_t bufLen, struct group **result)
{
    return getgrgid_r(gid, grp, buf, bufLen, result);
}
