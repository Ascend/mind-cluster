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
#include <gtest/gtest.h>
#include <mutex>
#include <thread>

#include "user_group_cache.h"
#include "data_base_user_group_loader.h"

using namespace ock::memfs;

static constexpr int64_t EXPIRES_MACRO_SECONDS = 10000L;

class TestUserGroupCache : public testing::Test {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();

protected:
    static std::shared_ptr<DataBaseUserGroupLoader> loader;
    static std::shared_ptr<UserGroupCache> cacheInstance;
};

std::shared_ptr<DataBaseUserGroupLoader> TestUserGroupCache::loader;
std::shared_ptr<UserGroupCache> TestUserGroupCache::cacheInstance;

void TestUserGroupCache::SetUpTestCase()
{
    loader = std::make_shared<DataBaseUserGroupLoader>();
    for (auto i = 1; i <= 3; i++) {
        for (auto j = 0; j < 4; j++) {
            auto userId = static_cast<uid_t>(i * 100 + j);
            auto groupId = static_cast<gid_t>(userId);
            auto userName = std::string("user-").append(std::to_string(userId));
            loader->AddUser(userName, userId, groupId);

            if (i <= 2) {
                auto group = static_cast<gid_t>(i * 1000);
                auto groupName = std::string("group-").append(std::to_string(group));
                loader->AddGroup(groupName, group);
                loader->AddUserToGroup(group, userId);
            }
        }
    }
    cacheInstance = MakeRef<UserGroupCache>(loader, EXPIRES_MACRO_SECONDS);
}

void TestUserGroupCache::TearDownTestCase()
{
    loader = nullptr;
    cacheInstance = nullptr;
}

TEST_F(TestUserGroupCache, user_info_get_uid_gid)
{
    uid_t uid = 1001;
    gid_t gid = 2002;
    UserInfo userInfo{uid, gid};

    ASSERT_EQ(uid, userInfo.GetUserId());
    ASSERT_EQ(gid, userInfo.GetGroupId());
}
TEST_F(TestUserGroupCache, user_info_get_uid_gid_ptr)
{
    uid_t uid = 1001;
    gid_t gid = 2002;
    auto userInfo = MakeRef<UserInfo>(uid, gid);

    ASSERT_TRUE(userInfo != nullptr);
    ASSERT_EQ(uid, userInfo->GetUserId());
    ASSERT_EQ(gid, userInfo->GetGroupId());

    userInfo = nullptr;
}

TEST_F(TestUserGroupCache, group_info_get_gid)
{
    std::unordered_set<uid_t> users;
    gid_t gid = 2002;
    GroupInfo groupInfo{gid, users};

    ASSERT_EQ(gid, groupInfo.GetGroupId());
}

TEST_F(TestUserGroupCache, group_info_get_gid_ptr)
{
    std::unordered_set<uid_t> users;
    gid_t gid = 2002;
    auto groupInfo = MakeRef<GroupInfo>(gid, users);

    ASSERT_TRUE(groupInfo != nullptr);
    ASSERT_EQ(gid, groupInfo->GetGroupId());

    groupInfo = nullptr;
}

TEST_F(TestUserGroupCache, load_user_info_user_name)
{
    for (auto i = 1; i <= 3; i++) {
        for (auto j = 0; j < 4; j++) {
            auto userId = static_cast<uid_t>(i * 100 + j);
            auto userName = std::string("user-").append(std::to_string(userId));
            auto userInfo = loader->LoadUser(userName);

            ASSERT_TRUE(userInfo != nullptr);
            ASSERT_EQ(userId, userInfo->GetUserId());
        }
    }
}

TEST_F(TestUserGroupCache, load_group_info_group_name)
{
    for (auto i = 1; i <= 2; i++) {
        for (auto j = 0; j < 4; j++) {
            auto groupId = static_cast<gid_t>(i * 1000);
            auto groupName = std::string("group-").append(std::to_string(groupId));
            auto groupInfo = loader->LoadGroup(groupName);

            ASSERT_TRUE(groupInfo != nullptr);
            ASSERT_EQ(groupId, groupInfo->GetGroupId());
        }
    }
}

TEST_F(TestUserGroupCache, root_in_group_root)
{
    auto in = cacheInstance->UserInGroup(0, 0);
    EXPECT_TRUE(in) << "root not in root group";
}

TEST_F(TestUserGroupCache, user_in_its_own_group)
{
    for (auto i = 1; i <= 3; i++) {
        for (auto j = 0; j < 4; j++) {
            auto userId = static_cast<uid_t>(i * 100 + j);
            auto groupId = static_cast<gid_t>(userId);
            auto in = cacheInstance->UserInGroup(userId, groupId);
            EXPECT_TRUE(in) << "user(" << userId << ") not in itself group";
        }
    }
}

TEST_F(TestUserGroupCache, user_in_its_belong_group)
{
    for (auto i = 1; i <= 2; i++) {
        for (auto j = 0; j < 4; j++) {
            auto userId = static_cast<uid_t>(i * 100 + j);
            auto groupId = static_cast<gid_t>(i * 1000);
            auto in = cacheInstance->UserInGroup(userId, groupId);
            EXPECT_TRUE(in) << "user(" << userId << ") not in belong group(" << groupId << ")";
        }
    }
}

TEST_F(TestUserGroupCache, exist_user_not_exist_group)
{
    gid_t group = 8888;
    for (auto i = 1; i <= 3; i++) {
        for (auto j = 0; j < 4; j++) {
            auto userId = static_cast<uid_t>(i * 100 + j);
            auto in = cacheInstance->UserInGroup(userId, group);
            EXPECT_FALSE(in) << "user(" << userId << ") in not exist group(" << group << ")";
        }
    }

    auto in = cacheInstance->UserInGroup(0, group);
    EXPECT_FALSE(in) << "user(" << 0 << ") in not exist group(" << group << ")";
}

TEST_F(TestUserGroupCache, nonexist_user_exist_group)
{
    uid_t user = 8888;
    for (auto i = 1; i <= 3; i++) {
        for (auto j = 0; j < 4; j++) {
            auto groupId = static_cast<gid_t>(i * 100 + j);
            auto in = cacheInstance->UserInGroup(user, groupId);
            EXPECT_FALSE(in) << "non-exist user(" << user << ") in group(" << groupId << ")";
        }
    }

    auto in = cacheInstance->UserInGroup(user, 0);
    EXPECT_FALSE(in) << "non-exist user(" << user << ") in group(" << 0 << ")";
}

TEST_F(TestUserGroupCache, user_not_in_its_other_group)
{
    auto i = 3;
    for (auto j = 0; j < 4; j++) {
        auto userId = static_cast<uid_t>(i * 100 + j);
        auto groupId = static_cast<gid_t>(i * 1000);
        auto in = cacheInstance->UserInGroup(userId, groupId);
        EXPECT_FALSE(in) << "user(" << userId << ") should not in other group(" << groupId << ")";
    }
}

TEST_F(TestUserGroupCache, expires_time)
{
    uid_t userId = 301;
    gid_t groupId = 2000;
    auto in = cacheInstance->UserInGroup(userId, groupId);
    ASSERT_FALSE(in) << "user (" << userId << ") already in group(" << groupId << ")";

    auto groupName = std::string("group-").append(std::to_string(groupId));
    loader->AddGroup(groupName, groupId);
    loader->AddUserToGroup(groupId, userId);

    std::this_thread::sleep_for(std::chrono::microseconds(EXPIRES_MACRO_SECONDS));
    in = cacheInstance->UserInGroup(userId, groupId);
    ASSERT_TRUE(in) << "user (" << userId << ") not in group(" << groupId << ")";
}