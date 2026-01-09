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
#include <mockcpp/mokc.h>

#define private public
#include "mem_file_system.h"
#undef private

using namespace ock::memfs;

namespace {

class TestMemFileSystem : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;
protected:
    uint64_t blkSize = 10UL << 10;
    std::string sysName = "test";
};

void TestMemFileSystem::SetUp() {}

void TestMemFileSystem::TearDown()
{
    GlobalMockObject::verify();
}

void MockBmmInit()
{
    union MockerHelper {
        int32_t (MemFsBMM::*initialize)(const MemFsBMMOptions &opt) noexcept;
        int32_t (*mockInitialize)(MemFsBMM *self, const MemFsBMMOptions &opt) noexcept;
    };
    MockerHelper helper{};
    helper.initialize = &MemFsBMM::Initialize;
    auto mocker = MOCKCPP_NS::mockAPI("&MemFsBMM::Initialize", helper.mockInitialize);
    mocker.defaults().will(returnValue(0));
}

TEST_F(TestMemFileSystem, opened_file_init_falied)
{
    OpenedFile file{};
    file.allocatedFlag = true;
    
    MemFsBMM bmm{};
    auto inode = std::make_shared<MemFsInode>(0, 0, "/", InodeType::INODE_DIR, 0, bmm);
    auto ret = file.Initialize(inode, true);
    ASSERT_FALSE(ret);
}

TEST_F(TestMemFileSystem, opened_file_release_falied)
{
    OpenedFile file{};
    file.allocatedFlag = false;
    auto ret = file.Release();
    ASSERT_FALSE(ret);
}

TEST_F(TestMemFileSystem, mem_file_system_init_no_blk)
{
    MemFileSystem memSys{ blkSize, 0UL, sysName };
    auto ret = memSys.Initialize();
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFileSystem, mem_file_system_init_no_mem)
{
    MemFileSystem memSys{ blkSize, 1UL, sysName };
    if (memSys.openedFiles != nullptr) {
        delete[] memSys.openedFiles;
        memSys.openedFiles = nullptr;
    }
    auto ret = memSys.Initialize();
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFileSystem, mem_file_system_init_bmm_failed)
{
    union MockerHelper {
        int32_t (MemFsBMM::*initialize)(const MemFsBMMOptions &opt) noexcept;
        int32_t (*mockInitialize)(MemFsBMM *self, const MemFsBMMOptions &opt) noexcept;
    };
    MockerHelper helper{};
    helper.initialize = &MemFsBMM::Initialize;
    auto mocker = MOCKCPP_NS::mockAPI("&MemFsBMM::Initialize", helper.mockInitialize);
    mocker.defaults().will(returnValue(1));

    MemFileSystem memSys{ blkSize, 1UL, sysName };
    auto ret = memSys.Initialize();
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFileSystem, mem_file_system_open_abnormal_branch)
{
    MockBmmInit();
    MemFileSystem memSys{ blkSize, 1UL, sysName };
    int ret = memSys.Initialize();
    ASSERT_EQ(0, ret);

    std::string path = "/tmp/mem_file_system_open_abnormal_branch";
    uint64_t inode = 0;

    union MockerHelper {
        int (MemFileSystem::*getInodeWithPath)(const std::string&, uint64_t&, uint64_t&, std::string&) noexcept;
        int (*mockGetInodeWithPath)(MemFileSystem *self,
                                    const std::string&, uint64_t&, uint64_t&, std::string&) noexcept;

        std::shared_ptr<MemFsInode> (MemFileSystem::*getInodeWithPermCheck)(uint64_t, PermitType) noexcept;
        std::shared_ptr<MemFsInode> (*mockGetInodeWithPermCheck)(MemFileSystem *self, uint64_t, PermitType) noexcept;

        int (MemFileSystem::*allocateFd)() noexcept;
        int (*mockAllocateFd)(MemFileSystem *self) noexcept;

        OpenedFile *(MemFileSystem::*getNewOpenedFile)(int) noexcept;
        OpenedFile *(*mockGetNewOpenedFile)(MemFileSystem *self, int) noexcept;
    };
    MockerHelper helper{};
    helper.getInodeWithPath = &MemFileSystem::GetInodeWithPath;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetInodeWithPath",
        helper.mockGetInodeWithPath).defaults().will(returnValue(0));

    // get inode with perm check failed
    ret = memSys.Open(path, inode);
    ASSERT_EQ(-1, ret);

    // allocate fd failed
    memSys.freeFds.clear();
    MemFsBMM bmm{};
    auto mockInode = std::make_shared<MemFsInode>(0, 0, "/", InodeType::INODE_DIR, 0, bmm);
    mockInode->writing = true;
    helper.getInodeWithPermCheck = &MemFileSystem::GetInodeWithPermCheck;
    MOCKCPP_NS::mockAPI("&MemFileSystem::GetInodeWithPermCheck",
        helper.mockGetInodeWithPermCheck).defaults().will(returnValue(mockInode));
    ret = memSys.Open(path, inode);
    ASSERT_EQ(-1, ret);

    // get new opened file failed
    helper.allocateFd = &MemFileSystem::AllocateFd;
    auto mockerAllocateFd = MOCKCPP_NS::mockAPI("&MemFileSystem::AllocateFd", helper.mockAllocateFd);
    mockerAllocateFd.defaults().will(returnValue(1));
    ret = memSys.Open(path, inode);
    ASSERT_EQ(-1, ret);

    // normal
    OpenedFile *file = new OpenedFile();
    helper.getNewOpenedFile = &MemFileSystem::GetNewOpenedFile;
    auto mockerGetNewOpenedFile = MOCKCPP_NS::mockAPI("&MemFileSystem::GetNewOpenedFile", helper.mockGetNewOpenedFile);
    mockerGetNewOpenedFile.defaults().will(returnValue(file));
    ret = memSys.Open(path, inode);
    ASSERT_EQ(1, ret);
    delete file;
    file = nullptr;
}

TEST_F(TestMemFileSystem, get_inode_with_path_empty)
{
    MemFileSystem memSys{ blkSize, 1UL, sysName };
    uint64_t ino = 0UL;
    uint64_t pino = 0UL;
    std::string lastToken = "";
    auto ret = memSys.GetInodeWithPath("", ino, pino, lastToken);
    ASSERT_EQ(0, ret);
}

TEST_F(TestMemFileSystem, check_rename_flags_exchange_failed)
{
    MemFileSystem memSys{ blkSize, 1UL, sysName };
    RenameContext context{ "src", "tgt", 0 };
    auto ret = memSys.CheckRenameFlagsExchange(context);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFileSystem, check_rename_flags_no_replace)
{
    MemFileSystem memSys{ blkSize, 1UL, sysName };
    RenameContext context{ "src", "tgt", 0 };
    auto ret = memSys.CheckRenameFlagsNoReplace(context);
    ASSERT_EQ(0, ret);

    MemFsBMM bmm;
    context.targetInode = std::make_shared<MemFsInode>(0, 0, "/", InodeType::INODE_DIR, 0, bmm);
    ret = memSys.CheckRenameFlagsNoReplace(context);
    ASSERT_EQ(-1, ret);
}
}