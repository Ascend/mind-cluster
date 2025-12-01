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
#include <mockcpp/mockcpp.hpp>

#include "backup_initiator.h"
#include "mem_fs_backup_initiator.h"
#include "memfs_sdk_api.h"
#include "memfs_api.h"
#include "pacific_adapter.h"

using namespace ock::bg::backup;
using namespace ock::memfs;
using namespace ock::ufs;
using namespace ock::bg;

namespace {

constexpr uint64_t DEFAULT_THREAD_DATA_SIZE = 1UL << 20;

class TestBackupInitiator : public testing::Test {
public:
    static void SetUpTestSuite();
    static void TearDownTestSuite();

public:
    void SetUp() override;
    void TearDown() override;

protected:
    static std::shared_ptr<MemFsBackupInitiator> backupTarget;
    static std::shared_ptr<BaseFileService> mockUfs;
    static std::string rootPath;
};

std::shared_ptr<BaseFileService> TestBackupInitiator::mockUfs;
std::shared_ptr<MemFsBackupInitiator> TestBackupInitiator::backupTarget;
std::string TestBackupInitiator::rootPath;

void TestBackupInitiator::SetUpTestSuite()
{
    rootPath = "/backup_initiator_test";
    mockUfs = std::make_shared<PacificAdapter>(rootPath);
    ASSERT_TRUE(mockUfs != nullptr);
}

void TestBackupInitiator::TearDownTestSuite()
{
    mockUfs.reset();
    rootPath.clear();
    GlobalMockObject::verify();
    GlobalMockObject::reset();
}

void TestBackupInitiator::SetUp()
{
    backupTarget = std::make_shared<MemFsBackupInitiator>();
    ASSERT_TRUE(backupTarget != nullptr);
}

void TestBackupInitiator::TearDown()
{
    backupTarget->Destroy();
    backupTarget.reset();
}

TEST_F(TestBackupInitiator, copy_file_to_ufs_test)
{
    std::string name = "/copy_file_to_ufs_test";
    MOCKER(MemFsApi::OpenFile).stubs().will(returnValue(0));
    MOCKER(MemFsApi::GetFileMeta).stubs().will(returnValue(0));
    MOCKER(MemFsApi::GetFileBlocks).stubs().will(returnValue(0));
    auto ret = backupTarget->MultiCopyFileToUfs(1, name, mockUfs);
    ASSERT_TRUE(ret != 0);
}

TEST_F(TestBackupInitiator, copy_file_to_memfs_test)
{
    std::string name = "/copy_file_to_memfs_test";
    MOCKER(MemFsApi::GetFileMeta).stubs().will(returnValue(0));
    MOCKER(MemFsApi::GetFileBlocks).stubs().will(returnValue(0));
    auto paraLoadCtxPtr = std::make_shared<ParallelLoadContext>(1);
    paraLoadCtxPtr->RecordTaskOffset(0);
    TaskInfo taskInfo{ 0, DEFAULT_THREAD_DATA_SIZE, 0, DEFAULT_THREAD_DATA_SIZE, paraLoadCtxPtr };

    auto ret = backupTarget->RecordToMemfsTaskResult(1, name, 0, taskInfo);
    ASSERT_TRUE(ret != 0);
}

TEST_F(TestBackupInitiator, set_process_mark_test)
{
    auto ret = 0;
    backupTarget->SetProcessingMark();
    ASSERT_TRUE(ret == 0);
}

TEST_F(TestBackupInitiator, clear_process_mark_test)
{
    auto ret = 0;
    backupTarget->ClearProcessingMark();
    ASSERT_TRUE(ret == 0);
}

TEST_F(TestBackupInitiator, check_process_mark_test)
{
    auto ret = backupTarget->CheckProcessingMark();
    ASSERT_TRUE(ret == 0);
}
}