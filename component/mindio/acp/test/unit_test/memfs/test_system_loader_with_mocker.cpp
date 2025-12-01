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
#include <mockcpp/mokc.h>
#include "user_group_cache.h"
#include "data_base_user_group_loader.h"

using namespace ock::memfs;

class TestSystemLoaderWithMocker : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;

protected:
    SystemUserGroupLoader loader;
};

void TestSystemLoaderWithMocker::SetUp()
{
}

void TestSystemLoaderWithMocker::TearDown()
{
    GlobalMockObject::verify();
    DataBaseUserGroupLoader::GetDefault().Clear();
}

namespace {

TEST_F(TestSystemLoaderWithMocker, getUidByName_normal)
{
    std::string userName = "dog";
    uid_t userId = 1001;
    gid_t groupId = 2002;
    MOCKER(SystemUserGroupWrapper::GetPwNam).stubs().will(invoke(DataBaseUserGroupLoader::GetPwNam));
    auto &defaultLoader = DataBaseUserGroupLoader::GetDefault();
    defaultLoader.AddUser(userName, userId, groupId);

    uid_t outUserId = 0;
    auto ret = SystemUserGroupLoader::GetUidByName(userName, outUserId);
    ASSERT_EQ(0, ret);
    ASSERT_EQ(userId, outUserId);
}

TEST_F(TestSystemLoaderWithMocker, getUidByName_need_size_large)
{
    std::string userName = "long_name_user";
    uid_t userId = 1001;
    gid_t groupId = 2002;
    MOCKER(SystemUserGroupWrapper::GetPwNam).stubs().will(invoke(DataBaseUserGroupLoader::GetPwNam));
    auto &defaultLoader = DataBaseUserGroupLoader::GetDefault();
    defaultLoader.AddUser(userName, userId, groupId);
    defaultLoader.SetGetUserBufferRange(128UL * 1024UL);

    uid_t outUserId = 0;
    auto ret = SystemUserGroupLoader::GetUidByName(userName, outUserId);
    ASSERT_EQ(0, ret);
    ASSERT_EQ(userId, outUserId);
}

TEST_F(TestSystemLoaderWithMocker, getUidByName_need_size_outof_range)
{
    std::string userName = "long_long_name_user";
    uid_t userId = 1001;
    gid_t groupId = 2002;
    MOCKER(SystemUserGroupWrapper::GetPwNam).stubs().will(invoke(DataBaseUserGroupLoader::GetPwNam));
    auto &defaultLoader = DataBaseUserGroupLoader::GetDefault();
    defaultLoader.AddUser(userName, userId, groupId);
    defaultLoader.SetGetUserBufferRange(128UL * 1024UL * 1024UL);

    uid_t outUserId = 0;
    auto ret = SystemUserGroupLoader::GetUidByName(userName, outUserId);
    ASSERT_EQ(ERANGE, ret);
}

TEST_F(TestSystemLoaderWithMocker, getUidByName_not_exist)
{
    std::string userName = "not_exist_user";
    MOCKER(SystemUserGroupWrapper::GetPwNam).stubs().will(invoke(DataBaseUserGroupLoader::GetPwNam));

    uid_t outUserId = 0;
    auto ret = SystemUserGroupLoader::GetUidByName(userName, outUserId);
    ASSERT_EQ(ENOENT, ret);
}

TEST_F(TestSystemLoaderWithMocker, getUidByName_io_failed)
{
    std::string userName = "io_failed_user";
    MOCKER(SystemUserGroupWrapper::GetPwNam).stubs().will(returnValue(EIO));

    uid_t outUserId = 0;
    auto ret = SystemUserGroupLoader::GetUidByName(userName, outUserId);
    ASSERT_NE(0, ret);
    ASSERT_NE(ENOENT, ret);
}

TEST_F(TestSystemLoaderWithMocker, loadUser_need_size_large)
{
    std::string userName = "long_name_user";
    uid_t userId = 11001;
    gid_t groupId = 12002;
    MOCKER(SystemUserGroupWrapper::GetPwNam).stubs().will(invoke(DataBaseUserGroupLoader::GetPwNam));
    auto &defaultLoader = DataBaseUserGroupLoader::GetDefault();
    defaultLoader.AddUser(userName, userId, groupId);
    defaultLoader.SetGetUserBufferRange(128UL * 1024UL);

    auto userInfo = loader.LoadUser(userName);
    ASSERT_TRUE(userInfo != nullptr);
    ASSERT_EQ(userId, userInfo->GetUserId());
    ASSERT_EQ(groupId, userInfo->GetGroupId());
}

TEST_F(TestSystemLoaderWithMocker, loadUser_need_size_outof_range)
{
    std::string userName = "long_long_name_user";
    uid_t userId = 21001;
    gid_t groupId = 22002;
    MOCKER(SystemUserGroupWrapper::GetPwNam).stubs().will(invoke(DataBaseUserGroupLoader::GetPwNam));
    auto &defaultLoader = DataBaseUserGroupLoader::GetDefault();
    defaultLoader.AddUser(userName, userId, groupId);
    defaultLoader.SetGetUserBufferRange(128UL * 1024UL * 1024UL);

    auto userInfo = loader.LoadUser(userName);
    ASSERT_TRUE(userInfo == nullptr);
}

TEST_F(TestSystemLoaderWithMocker, loadUser_not_exist)
{
    std::string userName = "not_exist_user";
    MOCKER(SystemUserGroupWrapper::GetPwNam).stubs().will(invoke(DataBaseUserGroupLoader::GetPwNam));

    auto userInfo = loader.LoadUser(userName);
    ASSERT_TRUE(userInfo == nullptr);
}

TEST_F(TestSystemLoaderWithMocker, loadUser_io_failed)
{
    std::string userName = "io_failed_user";
    MOCKER(SystemUserGroupWrapper::GetPwNam).stubs().will(returnValue(EIO));

    auto userInfo = loader.LoadUser(userName);
    ASSERT_TRUE(userInfo == nullptr);
}

TEST_F(TestSystemLoaderWithMocker, loadGroup_need_size_large)
{
    std::string groupName = "long_name_group";
    gid_t groupId = 22002;
    MOCKER(SystemUserGroupWrapper::GetGrNam).stubs().will(invoke(DataBaseUserGroupLoader::GetGrNam));
    auto &defaultLoader = DataBaseUserGroupLoader::GetDefault();
    defaultLoader.AddGroup(groupName, groupId);
    defaultLoader.SetGetGroupBufferRange(128UL * 1024UL);

    auto groupInfo = loader.LoadGroup(groupName);
    ASSERT_TRUE(groupInfo != nullptr);
    ASSERT_EQ(groupId, groupInfo->GetGroupId());
}

TEST_F(TestSystemLoaderWithMocker, loadGroup_need_size_outof_range)
{
    std::string groupName = "long_long_name_group";
    gid_t groupId = 23002;
    MOCKER(SystemUserGroupWrapper::GetGrNam).stubs().will(invoke(DataBaseUserGroupLoader::GetGrNam));
    auto &defaultLoader = DataBaseUserGroupLoader::GetDefault();
    defaultLoader.AddGroup(groupName, groupId);
    defaultLoader.SetGetGroupBufferRange(128UL * 1024UL * 1024UL);

    auto groupInfo = loader.LoadGroup(groupName);
    ASSERT_TRUE(groupInfo == nullptr);
}

TEST_F(TestSystemLoaderWithMocker, loadGroup_not_exist)
{
    std::string groupName = "not_exist_group";
    MOCKER(SystemUserGroupWrapper::GetGrNam).stubs().will(invoke(DataBaseUserGroupLoader::GetGrNam));

    auto groupInfo = loader.LoadGroup(groupName);
    ASSERT_TRUE(groupInfo == nullptr);
}

TEST_F(TestSystemLoaderWithMocker, loadGroup_io_failed)
{
    std::string groupName = "io_failed_group";
    MOCKER(SystemUserGroupWrapper::GetGrNam).stubs().will(returnValue(EIO));

    auto groupInfo = loader.LoadGroup(groupName);
    ASSERT_TRUE(groupInfo == nullptr);
}
}