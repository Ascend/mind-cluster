/**
 * Copyright (c) Huawei Technologies Co., Ltd. 2026. All rights reserved.
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

#include "mem_fs_inode.h"

using namespace ock::memfs;

namespace {

class TestMemFsInode : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;

protected:
    std::shared_ptr<MemFsInode> root;
    uint64_t ignInode = MemFsConstants::INODE_INVALID;
    uint64_t rootInode = MemFsConstants::ROOT_INODE_NUMBER;
};

void TestMemFsInode::SetUp()
{
    MemFsBMM bmm{};
    root = std::make_shared<MemFsInode>(ignInode, rootInode, "/", InodeType::INODE_DIR, 0755, bmm);
}

void TestMemFsInode::TearDown() {}

TEST_F(TestMemFsInode, test_get_file_size)
{
    auto size = root->GetFileSize();
    ASSERT_EQ(0UL, size);
}

TEST_F(TestMemFsInode, test_get_dentry_with_file)
{
    MemFsBMM bmm{};
    auto fileInode = std::make_shared<MemFsInode>(ignInode, rootInode, "/", InodeType::INODE_REG, 0755, bmm);
    Dentry dentry;
    auto ret = fileInode->GetDentry("", dentry);
    ASSERT_FALSE(ret);
}

TEST_F(TestMemFsInode, test_put_dentry_with_file)
{
    MemFsBMM bmm{};
    auto fileInode = std::make_shared<MemFsInode>(ignInode, rootInode, "/", InodeType::INODE_REG, 0755, bmm);
    Dentry dentry;
    int errorCode = 0;
    auto ret = fileInode->PutDentry("", 0, InodeType::INODE_DIR, errorCode);
    ASSERT_FALSE(ret);
}

TEST_F(TestMemFsInode, test_rename_dentry_invalid)
{
    MemFsBMM bmm{};
    auto fileInode = std::make_shared<MemFsInode>(ignInode, 0, "/", InodeType::INODE_REG, 0755, bmm);
    Dentry dentry;
    int errorCode = 0;

    // bool RenameDentry(const std::string &src, const std::string &dst, int &errorCode) noexcept;
    // file try rename dentry
    auto ret = fileInode->RenameDentry("src", "dst", errorCode);
    ASSERT_FALSE(ret);

    root->removed = true;
    // inode is removed
    ret = root->RenameDentry("src", "dst", errorCode);
    ASSERT_FALSE(ret);

    root->removed = false;
    // src no exist
    ret = root->RenameDentry("src", "dst", errorCode);
    ASSERT_FALSE(ret);
    
    // bool RenameDentry(const std::string&, const std::shared_ptr<MemFsInode>&,const std::string&, int&) noexcept;
    auto pDest = std::make_shared<MemFsInode>(rootInode, 0, "dir", InodeType::INODE_DIR, 0755, bmm);

    // srcInode is nullptr
    ret = root->RenameDentry("src", nullptr, "dst", errorCode);
    ASSERT_FALSE(ret);

    // common inode
    ret = root->RenameDentry("src", root, "dst", errorCode);
    ASSERT_FALSE(ret);

    // file try rename dentry
    ret = root->RenameDentry("src", fileInode, "dst", errorCode);
    ASSERT_FALSE(ret);

    root->removed = true;
    // inode is removed
    ret = root->RenameDentry("src", pDest, "dst", errorCode);
    ASSERT_FALSE(ret);

    root->removed = false;
    // src no exist
    ret = root->RenameDentry("src", pDest, "dst", errorCode);
    ASSERT_FALSE(ret);
}

TEST_F(TestMemFsInode, test_exchange_dentry_invalid)
{
    std::string src = "src";
    std::string dst = "dst";
    int errorCode = 0;

    // bool ExchangeDentry(const std::string &src, const std::string &dst, int &errorCode) noexcept;
    // inode is file
    MemFsBMM bmm{};
    auto fileInode = std::make_shared<MemFsInode>(ignInode, 0, "/", InodeType::INODE_REG, 0755, bmm);
    auto ret = fileInode->ExchangeDentry(src, dst, errorCode);
    ASSERT_FALSE(ret);

    // is removed
    root->removed = true;
    ret = root->ExchangeDentry(src, dst, errorCode);
    ASSERT_FALSE(ret);

    root->removed = false;
    // src not exist
    ret = root->ExchangeDentry(src, dst, errorCode);
    ASSERT_FALSE(ret);

    auto srcInode = std::make_shared<MemFsInode>(rootInode, 0, src, InodeType::INODE_DIR, 0755, bmm);
    root->PutDentry(src, 0, srcInode->type, errorCode);
    // dst not exist
    ret = root->ExchangeDentry(src, dst, errorCode);
    ASSERT_FALSE(ret);

    auto dstInode = std::make_shared<MemFsInode>(rootInode, 1, dst, InodeType::INODE_DIR, 0755, bmm);
    root->PutDentry(dst, 1, dstInode->type, errorCode);

    ret = root->ExchangeDentry(src, dst, errorCode);
    ASSERT_TRUE(ret);

    // bool ExchangeDentry(const std::string &src, const std::shared_ptr<MemFsInode> &pDest, const std::string &dst,
    //                     int &errorCode) noexcept;
    // pDest == nullptr
    ret = root->ExchangeDentry(src, nullptr, dst, errorCode);
    ASSERT_FALSE(ret);

    // common inode
    ret = root->ExchangeDentry(src, root, dst, errorCode);
    ASSERT_TRUE(ret);

    // is file
    ret = root->ExchangeDentry(src, fileInode, dst, errorCode);
    ASSERT_FALSE(ret);

    auto pDest = std::make_shared<MemFsInode>(rootInode, 0, "dir", InodeType::INODE_DIR, 0755, bmm);
    // is removed
    root->removed = true;
    ret = root->ExchangeDentry(src, pDest, dst, errorCode);
    ASSERT_FALSE(ret);

    root->removed = false;

    root->PutDentry(src, 0, srcInode->type, errorCode);
    // dst not exist
    ret = root->ExchangeDentry(src, pDest, dst, errorCode);
    ASSERT_FALSE(ret);

    Dentry dentry;
    root->DeleteDentry(src, dentry);
    // src not exist
    ret = root->ExchangeDentry(src, pDest, dst, errorCode);
    ASSERT_FALSE(ret);
}

TEST_F(TestMemFsInode, test_delete_dentry_with_file)
{
    MemFsBMM bmm{};
    auto fileInode = std::make_shared<MemFsInode>(ignInode, rootInode, "/", InodeType::INODE_REG, 0755, bmm);
    Dentry dentry;
    auto ret = fileInode->DeleteDentry("", dentry);
    ASSERT_FALSE(ret);
}

TEST_F(TestMemFsInode, test_get_perm_info_with_acl)
{
    InodeAcl *acl = new InodeAcl();
    acl->usersAcl.emplace(0, 1);
    acl->groupsAcl.emplace(0, 1);
    root->acl = acl;

    auto permission = root->GetPermInfo();
    ASSERT_TRUE(permission.ContainsPermission(PermitType::PERM_MASK));
}
}