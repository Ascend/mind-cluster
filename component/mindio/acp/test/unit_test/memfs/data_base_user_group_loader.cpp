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
#include <pwd.h>
#include <cstring>
#include "securec.h"
#include "data_base_user_group_loader.h"

using namespace ock::memfs;

static constexpr auto DEFAULT_INFO_SIZE = 128UL;

DataBaseUserGroupLoader DataBaseUserGroupLoader::defaultDataBase;

DataBaseUserGroupLoader::DataBaseUserGroupLoader() noexcept
    : userInfoNeedSize{ DEFAULT_INFO_SIZE }, groupInfoNeedSize{ DEFAULT_INFO_SIZE }
{
    users.emplace(0U, std::make_pair("root", 0U));
    groups.emplace(0U, std::make_pair("root", std::unordered_set<gid_t>()));
}


void DataBaseUserGroupLoader::AddUser(const std::string &name, uid_t user, gid_t group) noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    if (users.find(user) == users.end() && userNames.find(name) == userNames.end()) {
        userNames.emplace(name, user);
        users.emplace(user, std::make_pair(name, group));
    }
}

void DataBaseUserGroupLoader::AddGroup(const std::string &name, gid_t group) noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    if (groups.find(group) == groups.end() && groupNames.find(name) == groupNames.end()) {
        groupNames.emplace(name, group);
        groups.emplace(group, std::make_pair(name, std::unordered_set<uid_t>()));
    }
}

void DataBaseUserGroupLoader::AddUserToGroup(gid_t group, uid_t user) noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    groups[group].second.emplace(user);
}

void DataBaseUserGroupLoader::Clear() noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    userNames.clear();
    users.clear();
    groupNames.clear();
    groups.clear();
    userInfoNeedSize = DEFAULT_INFO_SIZE;
    groupInfoNeedSize = DEFAULT_INFO_SIZE;
}

void DataBaseUserGroupLoader::SetGetUserBufferRange(size_t needSize) noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    userInfoNeedSize = needSize;
}

void DataBaseUserGroupLoader::SetGetGroupBufferRange(size_t needSize) noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    groupInfoNeedSize = needSize;
}

DataBaseUserGroupLoader &DataBaseUserGroupLoader::GetDefault() noexcept
{
    return defaultDataBase;
}

UserInfoPtr DataBaseUserGroupLoader::LoadUser(uid_t uid) const noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    auto pos = users.find(uid);
    if (pos == users.end()) {
        return nullptr;
    }
    return MakeRef<UserInfo>(uid, pos->second.second);
}

UserInfoPtr DataBaseUserGroupLoader::LoadUser(const std::string &userName) const noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    auto namePos = userNames.find(userName);
    if (namePos == userNames.end()) {
        return nullptr;
    }

    auto pos = users.find(namePos->second);
    if (pos == users.end()) {
        return nullptr;
    }

    return MakeRef<UserInfo>(pos->first, pos->second.second);
}

GroupInfoPtr DataBaseUserGroupLoader::LoadGroup(gid_t gid) const noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    auto pos = groups.find(gid);
    if (pos == groups.end()) {
        return nullptr;
    }
    return MakeRef<GroupInfo>(gid, pos->second.second);
}

GroupInfoPtr DataBaseUserGroupLoader::LoadGroup(const std::string &groupName) const noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    auto namePos = groupNames.find(groupName);
    if (namePos == groupNames.end()) {
        return nullptr;
    }

    auto pos = groups.find(namePos->second);
    if (pos == groups.end()) {
        return nullptr;
    }

    return MakeRef<GroupInfo>(pos->first, pos->second.second);
}

int DataBaseUserGroupLoader::GetPwNam(const std::string &name, struct passwd *pwd, char *buf, size_t bufLen,
    struct passwd **result)
{
    auto &loader = DataBaseUserGroupLoader::GetDefault();
    auto userInfo = loader.LoadUser(name);
    if (userInfo == nullptr) {
        *result = nullptr;
        return 0;
    }

    return GetPwUid(userInfo->GetUserId(), pwd, buf, bufLen, result);
}

int DataBaseUserGroupLoader::GetPwUid(uid_t uid, struct passwd *pwd, char *buf, size_t bufLen, struct passwd **result)
{
    if (pwd == nullptr || buf == nullptr || result == nullptr) {
        return EINVAL;
    }

    auto &loader = DataBaseUserGroupLoader::GetDefault();
    std::unique_lock<std::mutex> lockGuard{ loader.mutex };
    auto uidPos = loader.users.find(uid);
    if (uidPos == loader.users.end()) {
        *result = nullptr;
        return 0;
    }

    if (bufLen < loader.userInfoNeedSize) {
        *result = nullptr;
        return ERANGE;
    }

    const auto &name = uidPos->second.first;
    auto groupId = uidPos->second.second;
    auto pos = buf;

    auto ret = strcpy_s(pos, bufLen, name.c_str());
    if (ret != EOK) {
        return -1;
    }
    pwd->pw_name = pos;
    pos += name.length() + 1;

    ret = strcpy_s(pos, bufLen, "X");
    if (ret != EOK) {
        return -1;
    }
    pwd->pw_passwd = pos;
    pos += 2;

    pwd->pw_uid = uid;
    pwd->pw_gid = groupId;

    ret = strcpy_s(pos, bufLen, "this is user for test.");
    if (ret != EOK) {
        return -1;
    }
    pwd->pw_gecos = pos;
    pos += (strlen(pos) + 1);

    std::string homePath = "/home/" + name;
    ret = strcpy_s(pos, bufLen, homePath.c_str());
    if (ret != EOK) {
        return -1;
    }
    pwd->pw_dir = pos;
    pos += homePath.length() + 1;

    ret = strcpy_s(pos, bufLen, "/bin/bash");
    if (ret != EOK) {
        return -1;
    }
    pwd->pw_shell = pos;

    *result = pwd;
    return 0;
}

int DataBaseUserGroupLoader::GetGrNam(const std::string &name, struct group *grp, char *buf, size_t bufLen,
    struct group **result)
{
    auto &loader = DataBaseUserGroupLoader::GetDefault();
    auto groupInfo = loader.LoadGroup(name);
    if (groupInfo == nullptr) {
        *result = nullptr;
        return 0;
    }

    return GetGrGid(groupInfo->GetGroupId(), grp, buf, bufLen, result);
}

int DataBaseUserGroupLoader::GetGrGid(gid_t gid, struct group *grp, char *buf, size_t bufLen, struct group **result)
{
    if (grp == nullptr || buf == nullptr || result == nullptr) {
        return EINVAL;
    }

    auto &loader = DataBaseUserGroupLoader::GetDefault();
    std::unique_lock<std::mutex> lockGuard{ loader.mutex };
    auto gidPos = loader.groups.find(gid);
    if (gidPos == loader.groups.end()) {
        *result = nullptr;
        return 0;
    }

    if (bufLen < loader.groupInfoNeedSize) {
        *result = nullptr;
        return ERANGE;
    }

    const auto &name = gidPos->second.first;
    const auto &users = gidPos->second.second;
    auto pos = buf;

    auto ret = strcpy_s(pos, bufLen, name.c_str());
    if (ret != EOK) {
        return -1;
    }
    grp->gr_name = pos;
    pos += name.length() + 1;

    ret = strcpy_s(pos, bufLen, "X");
    if (ret != EOK) {
        return -1;
    }
    grp->gr_passwd = pos;
    pos += 2;

    grp->gr_gid = gid;

    grp->gr_mem = reinterpret_cast<char **>(pos);
    auto count = users.size();
    auto arrayTotalSize = sizeof(char *) * (count + 1);
    grp->gr_mem[count] = nullptr;
    auto index = 0;
    for (auto user_id : users) {
        const auto &user_name = loader.users[user_id].first;
        ret = strcpy_s(pos, bufLen, user_name.c_str());
        if (ret != EOK) {
            return -1;
        }
        grp->gr_mem[index] = pos;
        pos += user_name.length() + 1;
    }

    *result = grp;
    return 0;
}