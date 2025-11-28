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

#include "backup_file_manager.h"
#include "mem_fs_backup_initiator.h"
#include "backup_target.h"
#include "memfs_api.h"

using namespace ock::bg::backup;
using namespace ock::common;
using namespace ock::memfs;

class TestBackupFileManager : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;
};

void TestBackupFileManager::SetUp() {}

void TestBackupFileManager::TearDown()
{
    GlobalMockObject::verify();
}

TEST_F(TestBackupFileManager, Initialize)
{
    int ret;

    config::BackupServiceConfig config;
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    config.enabled = true;
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    config.enabled = false;
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    config.threadNum = 1;
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    config::BackupInstance backupInstance;
    config.backups.push_back(backupInstance);
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    config.backups.clear();
    backupInstance.opened = true;
    config.backups.push_back(backupInstance);
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    config.backups.clear();
    backupInstance.source = "mfs";
    config.backups.push_back(backupInstance);
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    config.backups.clear();
    backupInstance.destType = "mfs";
    config.backups.push_back(backupInstance);
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    config.backups.clear();
    backupInstance.destType = "xxx";
    config.backups.push_back(backupInstance);
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    config.backups.clear();
    backupInstance.destName = "dest1";
    backupInstance.source = "dfs";
    backupInstance.destType = "under_fs";
    config.backups.push_back(backupInstance);
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_NE(0, ret);

    std::string name = "/test_name";
    auto ptr = BackupFileManager::GetInstance().GetInitiator(name);
    ASSERT_EQ(nullptr, ptr);

    BackupFileManager::GetInstance().Destroy();
}

static FileOpNotify g_testFileOpNotify;
int RegisterFileOpNotifyStub(const FileOpNotify &notify)
{
    g_testFileOpNotify = notify;
    return 0;
}

constexpr uint64_t TEST_INODE_ID = 1;

union MockHelper {
    int (BackupTarget::*createFile)(const FileTrace &trace, const struct stat &buf);
    int (*mockCeateFile)(BackupTarget*, const FileTrace &trace, const struct stat &buf);
};

TEST_F(TestBackupFileManager, UpFs)
{
    int ret;
    config::BackupServiceConfig config;
    config.enabled = true;
    config.threadNum = 1;
    config.maxFailCntForUnserviceable = 1;
    config::BackupInstance backupInstance;
    backupInstance.destName = "dest1";
    backupInstance.destType = "dfs";
    backupInstance.source = "mfs";
    backupInstance.opened = true;
    config.backups.push_back(backupInstance);

    MOCKER(MemFsApi::RegisterFileOpNotify).stubs().will(invoke(RegisterFileOpNotifyStub));
    ret = BackupFileManager::GetInstance().Initialize(config);
    ASSERT_EQ(0, ret);

    MockHelper helper{};
    // helper.createFile = &BackupTarget::CreateFileAndStageSync;
    // MOCKCPP_NS:mockAPI("&BackupTarget::CreateFileAndStageSync", helper.mockCeateFile).stubs().will(returnValue(0));
    MOCKER(MemFsApi::GetMeta).stubs().will(returnValue(0));
    
    g_testFileOpNotify.openNotify(1, "abc", 0, TEST_INODE_ID);
    g_testFileOpNotify.closeNotify(1, false);
    g_testFileOpNotify.closeNotify(1, true);

    g_testFileOpNotify.mkdirNotify("/abc", 0755, 0, 0);
    g_testFileOpNotify.unlinkNotify("/abc", TEST_INODE_ID);

    g_testFileOpNotify.preloadFileNotify("/abcde");
    g_testFileOpNotify.newFileNotify("/abcd", TEST_INODE_ID);
    BackupFileManager::GetInstance().Destroy();
}