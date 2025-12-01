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
#include "user_group_cache.h"

using namespace ock::memfs;

class TestSystemUserGroupLoader : public testing::Test {
protected:
    SystemUserGroupLoader loader;
};

namespace {

TEST_F(TestSystemUserGroupLoader, root_user_load)
{
    auto user = loader.LoadUser(0);
    ASSERT_TRUE(user != nullptr);
    ASSERT_EQ(0U, user->GetUserId());
    ASSERT_EQ(0U, user->GetGroupId());

    user = loader.LoadUser("root");
    ASSERT_TRUE(user != nullptr);
    ASSERT_EQ(0U, user->GetUserId());
    ASSERT_EQ(0U, user->GetGroupId());
}

TEST_F(TestSystemUserGroupLoader, root_group_load)
{
    auto group = loader.LoadGroup(0);
    ASSERT_TRUE(group != nullptr);
    ASSERT_EQ(0U, group->GetGroupId());

    group = loader.LoadGroup("root");
    ASSERT_TRUE(group != nullptr);
    ASSERT_EQ(0U, group->GetGroupId());
}

TEST_F(TestSystemUserGroupLoader, not_exist_user)
{
    uid_t userId = 88887777U;
    auto user = loader.LoadUser(userId);
    ASSERT_TRUE(user == nullptr);
}

TEST_F(TestSystemUserGroupLoader, not_exist_group)
{
    uid_t groupid = 88889999U;
    auto group = loader.LoadGroup(groupid);
    ASSERT_TRUE(group == nullptr);
}
}