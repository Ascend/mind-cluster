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

#include "inode_permission.h"
#include "user_group_cache.h"

using namespace ock::memfs;

class MockUserGroupLoader : public UserGroupLoader {
public:
    UserInfoPtr LoadUser(uid_t uid) const noexcept override
    {
        return MakeRef<UserInfo>(uid, uid * 2U);
    }

    UserInfoPtr LoadUser(const std::string &userName) const noexcept override
    {
        return nullptr;
    }

    GroupInfoPtr LoadGroup(gid_t gid) const noexcept override
    {
        std::unordered_set<uid_t> users;
        users.emplace(gid);
        return MakeRef<GroupInfo>(gid, users);
    }

    GroupInfoPtr LoadGroup(const std::string &groupName) const noexcept override
    {
        return nullptr;
    }
};

class TestInodePermission : public testing::Test {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();
    void SetUp() override;
    void TearDown() override;
};

void TestInodePermission::SetUpTestCase()
{
    InodePermission::SetUserGroupCache(MakeRef<UserGroupCache>(std::make_shared<MockUserGroupLoader>()));
}

void TestInodePermission::TearDownTestCase()
{
    InodePermission::SetUserGroupCache(UserGroupCache::GetSystemInstance());
}

void TestInodePermission::SetUp()
{
}

void TestInodePermission::TearDown()
{
    GlobalMockObject::verify();
}

namespace {

TEST_F(TestInodePermission, common_file_owner_read_permission)
{
    auto inodePerm = std::make_shared<InodePermission>(false, S_IRUSR, 1001, 2002);
    auto ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR | S_IWUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR | S_IWUSR | S_IXUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    EXPECT_TRUE(ok);
}

TEST_F(TestInodePermission, common_file_owner_write_permission)
{
    auto inodePerm = std::make_shared<InodePermission>(false, S_IWUSR, 1001, 2002);
    auto ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR | S_IWUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR | S_IWUSR | S_IXUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    EXPECT_TRUE(ok);
}

TEST_F(TestInodePermission, common_file_owner_execute_permission)
{
    auto inodePerm = std::make_shared<InodePermission>(false, S_IXUSR, 1001, 2002);
    auto ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR | S_IXUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR | S_IWUSR | S_IXUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, 0, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_FALSE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_FALSE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR | S_IWUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_FALSE(ok);
}

TEST_F(TestInodePermission, common_file_group_read_permission)
{
    auto inodePerm = std::make_shared<InodePermission>(false, S_IRGRP, 1001, 2002);
    auto ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRGRP | S_IWGRP, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRGRP | S_IWGRP | S_IXGRP, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    EXPECT_TRUE(ok);
}

TEST_F(TestInodePermission, root_can_read_all_file)
{
    auto inodePerm = std::make_shared<InodePermission>(false, 0, 1001, 2002);
    auto ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    EXPECT_TRUE(ok);
}

TEST_F(TestInodePermission, root_can_write_all_file)
{
    auto inodePerm = std::make_shared<InodePermission>(false, 0, 1001, 2002);
    auto ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IWUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    EXPECT_TRUE(ok);
}

TEST_F(TestInodePermission, root_can_execute_all_file_contains_exec_bit)
{
    auto inodePerm = std::make_shared<InodePermission>(false, S_IXUSR, 1001, 2002);
    auto ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IXGRP, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IXOTH, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    EXPECT_TRUE(ok);
}

TEST_F(TestInodePermission, root_can_execute_all_dir)
{
    auto inodePerm = std::make_shared<InodePermission>(true, 0, 1001, 2002);
    auto ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(true, S_IRUSR | S_IWUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_TRUE(ok);

    inodePerm = std::make_shared<InodePermission>(true, 0666, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_TRUE(ok);
}

TEST_F(TestInodePermission, root_cannot_exec_file_no_exec_bit)
{
    auto inodePerm = std::make_shared<InodePermission>(false, 0, 1001, 2002);
    auto ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_FALSE(ok);

    inodePerm = std::make_shared<InodePermission>(false, S_IRUSR | S_IWUSR, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_FALSE(ok);

    inodePerm = std::make_shared<InodePermission>(false, 0666, 1001, 2002);
    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    EXPECT_FALSE(ok);
}

TEST_F(TestInodePermission, acl_root_permit_with_users)
{
    std::map<uid_t, uint16_t> usersAcl{ { 3003, PERM_WRITE } };
    std::map<gid_t, uint16_t> groupsAcl;
    auto inodePerm = std::make_shared<InodePermission>(false, 0400, 1001, 2002, usersAcl, groupsAcl);

    auto ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    ASSERT_TRUE(ok);

    ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    ASSERT_TRUE(ok);

    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    ASSERT_FALSE(ok);
}

TEST_F(TestInodePermission, acl_root_permit_with_groups)
{
    std::map<uid_t, uint16_t> usersAcl;
    std::map<gid_t, uint16_t> groupsAcl{ { 3003, PERM_WRITE } };
    auto inodePerm = std::make_shared<InodePermission>(false, 0400, 1001, 2002, usersAcl, groupsAcl);

    auto ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    ASSERT_TRUE(ok);

    ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    ASSERT_TRUE(ok);

    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    ASSERT_FALSE(ok);
}

TEST_F(TestInodePermission, acl_to_single_user)
{
    std::map<uid_t, uint16_t> usersAcl{ { 3003, PERM_WRITE } };
    std::map<gid_t, uint16_t> groupsAcl;
    auto inodePerm = std::make_shared<InodePermission>(false, 0644, 1001, 2002, usersAcl, groupsAcl);

    auto ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    ASSERT_TRUE(ok);

    ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    ASSERT_TRUE(ok);

    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    ASSERT_FALSE(ok);
}

TEST_F(TestInodePermission, acl_to_single_group)
{
    std::map<uid_t, uint16_t> usersAcl;
    std::map<gid_t, uint16_t> groupsAcl{ { 3003, PERM_WRITE } };
    auto inodePerm = std::make_shared<InodePermission>(false, 0644, 1001, 2002, usersAcl, groupsAcl);

    auto ok = inodePerm->ContainsPermission(PermitType::PERM_READ);
    ASSERT_TRUE(ok);

    union MockerHelper {
        bool (UserGroupCache::*userInGroupFun)(uid_t user, gid_t group) noexcept;
        bool (*mockUserInGroup)(UserGroupCache *self, uid_t user, gid_t group) noexcept;
    };
    MockerHelper helper{};
    helper.userInGroupFun = &UserGroupCache::UserInGroup;
    auto mocker = MOCKCPP_NS::mockAPI("&UserGroupCache::UserInGroup", helper.mockUserInGroup);
    mocker.defaults().will(returnValue(false));
    mocker.stubs().with(any(), eq(8008U), eq(3003U)).will(returnValue(true));
    ok = inodePerm->ContainsPermission(PermitType::PERM_WRITE);
    ASSERT_TRUE(ok);

    ok = inodePerm->ContainsPermission(PermitType::PERM_EXECUTE);
    ASSERT_FALSE(ok);
}
}