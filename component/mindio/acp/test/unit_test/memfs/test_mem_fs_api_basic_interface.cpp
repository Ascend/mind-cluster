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
#include <fcntl.h>
#include <gtest/gtest.h>
#include <mockcpp/mokc.h>

#define private public
#include "mem_file_system.h"
#undef private

#include "memfs_api.h"

namespace {

class TestMemFsApiBasicInterface : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;
};

void TestMemFsApiBasicInterface::SetUp() {}
void TestMemFsApiBasicInterface::TearDown()
{
    GlobalMockObject::verify();
}

union MockerHelper {
    int (MemFileSystem::*initFileSys)() noexcept;
    int (*mockInitFileSys)(MemFileSystem *self) noexcept;

    int (InodeEvictor::*initInodeEvict)() noexcept;
    int (*mockInitInodeEvict)(MemFileSystem *self) noexcept;

    int (MemFileSystem::*getFileMeta)(int fd, struct stat &metadata) noexcept;
    int (*mockGetFileMeta)(MemFileSystem *self, int fd, struct stat &metadata) noexcept;

    int (MemFileSystem::*getMeta)(const std::string &path, struct stat &metadata) noexcept;
    int (*mockGetMeta)(MemFileSystem *self, const std::string &path, struct stat &metadata) noexcept;

    int (MemFileSystem::*create)(const std::string &path, mode_t mode, uint64_t &outInode) noexcept;
    int (*mockCreate)(MemFileSystem *self, const std::string &path, mode_t mode, uint64_t &outInode) noexcept;

    int (MemFileSystem::*close)(int fd) noexcept;
    int (*mockClose)(MemFileSystem *self, int fd) noexcept;

    int (MemFileSystem::*remove)(const std::string &path, uint64_t &outInode) noexcept;
    int (*mockRemove)(MemFileSystem *self, const std::string &path, uint64_t &outInode) noexcept;

    int (MemFileSystem::*link)(const std::string &oldPath, const std::string &newPath, uint64_t &outInode) noexcept;
    int (*mockLink)(MemFileSystem *self,
                    const std::string &oldPath, const std::string &newPath, uint64_t &outInode) noexcept;
};

int MockGetFileMeta(MemFileSystem *self, int fd, struct stat &metadata)
{
    metadata.st_ino = 0;
    return 0;
};

int MockGetMetaInode0(MemFileSystem *self, const std::string &path, struct stat &metadata)
{
    metadata.st_ino = 0;
    return 0;
};

int MockGetMetaInode1(MemFileSystem *self, const std::string &path, struct stat &metadata)
{
    metadata.st_ino = 1;
    return 0;
};

TEST_F(TestMemFsApiBasicInterface, test_init_mem_file_system_failed)
{
    MockerHelper helper{};
    helper.initFileSys = &MemFileSystem::Initialize;
    MOCKCPP_NS::mockAPI("&MemFileSystem::Initialize", helper.mockInitFileSys).defaults().will(returnValue(-1));

    auto ret = MemFsApi::Initialize();
    ASSERT_EQ(-1, ret);
    ret = MemFsApi::GetShareMemoryFd();
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsApiBasicInterface, test_init_inode_evictor_failed)
{
    MockerHelper helper{};
    helper.initFileSys = &MemFileSystem::Initialize;
    MOCKCPP_NS::mockAPI("&MemFileSystem::Initialize", helper.mockInitFileSys).defaults().will(returnValue(0));
    helper.initInodeEvict = &InodeEvictor::Initialize;
    MOCKCPP_NS::mockAPI("&InodeEvictor::Initialize", helper.mockInitInodeEvict).defaults().will(returnValue(-1));
    auto ret = MemFsApi::Initialize();
    ASSERT_EQ(-1, ret);
}


TEST_F(TestMemFsApiBasicInterface, test_open_file_notify_failed)
{
    MockerHelper helper{};
    helper.create = &MemFileSystem::Create;
    MOCKCPP_NS::mockAPI("&MemFileSystem::Create", helper.mockCreate).defaults().will(returnValue(0));

    FileOpNotify notify{};
    notify.openNotify = [](int fd, const std::string &name, int flags, int64_t inode) {
        return -1;
    };
    ASSERT_EQ(MemFsApi::RegisterFileOpNotify(notify), 0);
    auto ret = MemFsApi::OpenFile("/tmp", O_CREAT | O_TRUNC | O_WRONLY, 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsApiBasicInterface, test_discard_file_get_file_meta_failed)
{
    MockerHelper helper{};
    helper.getFileMeta = &MemFileSystem::GetFileMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetFileMeta", helper.mockGetFileMeta).defaults().will(returnValue(-1));
    auto ret = MemFsApi::DiscardFile("/tmp", 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsApiBasicInterface, test_discard_file_get_meta_failed)
{
    MockerHelper helper{};
    helper.getFileMeta = &MemFileSystem::GetFileMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetFileMeta", helper.mockGetFileMeta).defaults().will(returnValue(0));
    helper.getMeta = &MemFileSystem::GetMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetMeta", helper.mockGetMeta).defaults().will(returnValue(-1));
    auto ret = MemFsApi::DiscardFile("/tmp", 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsApiBasicInterface, test_discard_file_stat_not_equal)
{
    MockerHelper helper{};
    helper.getFileMeta = &MemFileSystem::GetFileMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetFileMeta", helper.mockGetFileMeta).stubs().will(invoke(MockGetFileMeta));
    helper.getMeta = &MemFileSystem::GetMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetMeta", helper.mockGetMeta).stubs().will(invoke(MockGetMetaInode1));
    auto ret = MemFsApi::DiscardFile("/tmp", 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsApiBasicInterface, test_discard_file_close_failed)
{
    MockerHelper helper{};
    helper.getFileMeta = &MemFileSystem::GetFileMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetFileMeta", helper.mockGetFileMeta).defaults().will(invoke(MockGetFileMeta));
    helper.getMeta = &MemFileSystem::GetMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetMeta", helper.mockGetMeta).defaults().will(invoke(MockGetMetaInode0));
    helper.close = &MemFileSystem::Close;
    MOCKCPP_NS::mockAPI("&MemFileSystem::Close", helper.mockClose).defaults().will(returnValue(-1));
    auto ret = MemFsApi::DiscardFile("/tmp", 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsApiBasicInterface, test_discard_file_remove_failed)
{
    MockerHelper helper{};
    helper.getFileMeta = &MemFileSystem::GetFileMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetFileMeta", helper.mockGetFileMeta).defaults().will(invoke(MockGetFileMeta));
    helper.getMeta = &MemFileSystem::GetMeta;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetMeta", helper.mockGetMeta).defaults().will(invoke(MockGetMetaInode0));
    helper.close = &MemFileSystem::Close;
    MOCKCPP_NS::mockAPI("&MemFileSystem::Close", helper.mockClose).defaults().will(returnValue(0));
    helper.remove = &MemFileSystem::RemoveFile;
    MOCKCPP_NS::mockAPI("&MemFileSystem::RemoveFile", helper.mockRemove).defaults().will(returnValue(-1));
    auto ret = MemFsApi::DiscardFile("/tmp", 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsApiBasicInterface, test_link_notify_failed)
{
    MockerHelper helper{};
    helper.link = &MemFileSystem::LinkFile;
    MOCKCPP_NS::mockAPI("&MemFileSystem::LinkFile", helper.mockLink).defaults().will(returnValue(0));
    helper.remove = &MemFileSystem::RemoveFile;
    MOCKCPP_NS::mockAPI("&MemFileSystem::RemoveFile", helper.mockRemove).defaults().will(returnValue(-1));

    FileOpNotify notify{};
    notify.newFileNotify = [](const std::string &path, int64_t inode) {
        return -1;
    };
    ASSERT_EQ(MemFsApi::RegisterFileOpNotify(notify), 0);
    auto ret = MemFsApi::Link("src", "dst");
    ASSERT_EQ(-1, ret);
}
}