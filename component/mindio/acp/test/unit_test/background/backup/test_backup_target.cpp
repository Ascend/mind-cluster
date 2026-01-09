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
#include <fcntl.h>
#include <fstream>
#include <gtest/gtest.h>
#include <gmock/gmock.h>
#include <mockcpp/mockcpp.hpp>
#include <mockcpp/mokc.h>

#include "retry_task_pool.h"
#include "pacific_adapter.h"
#include "backup_file_tracer.h"
#include "ufs_api.h"
#include "memfs_api.h"
#include "service_configure.h"

#define private public
#include "mem_fs_backup_initiator.h"
#include "backup_file_manager.h"
#include "backup_target.h"
#undef private

using namespace ock::memfs;
using namespace ock::common::config;
using namespace ock::ufs;
using namespace ock::common;
using namespace ock::bg;
using namespace ock::bg::backup;
using ::testing::_;
using ::testing::Return;

namespace {

constexpr uint64_t DEFAULT_THREAD_DATA_SIZE = 1UL << 20;

class TestBackupTarget : public testing::Test {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();

    void SetUp() override;
    void TearDown() override;

protected:
    std::shared_ptr<BackupTarget> backupTarget;

protected:
    static std::shared_ptr<BaseFileService> mockUfs;
    static std::string ufsPath;
};

std::shared_ptr<BaseFileService> TestBackupTarget::mockUfs;
std::string TestBackupTarget::ufsPath;

void TestBackupTarget::SetUpTestCase()
{
    ufsPath = "./mock.fs.test_backup_target";
    std::string command = "mkdir -p " + ufsPath;

    auto ret = system(command.c_str());
    ASSERT_EQ(0, ret) << " ret = " << ret << ", err=" << errno << ", str=" << strerror(errno);

    mockUfs = std::make_shared<PacificAdapter>(ufsPath);
    ASSERT_TRUE(mockUfs != nullptr);

    int confRet = ServiceConfigure::GetInstance().Initialize();
    if (confRet != 0) {
        std::cout << "service configure init failed." << std::endl;
        ASSERT_EQ(0, confRet);
        return;
    }

    ret = MemFsApi::Initialize();
    std::cout << "test_mem_fs_api set up, ret is " << ret << std::endl;
    ASSERT_EQ(0, ret);
}

void TestBackupTarget::TearDownTestCase()
{
    mockUfs.reset();
    std::string command = "rm -rf " + ufsPath;

    auto ret = system(command.c_str());
    ASSERT_EQ(0, ret) << " ret = " << ret << ", err=" << errno << ", str=" << strerror(errno);

    ufsPath.clear();

    MemFsApi::Destroy();
    ServiceConfigure::GetInstance().Destroy();
    std::cout << "test_mem_fs_api tear down" << std::endl;
}

void TestBackupTarget::SetUp()
{
    util::RetryTaskPool::RetryTaskConfig config;
    config.name = "mock";
    auto taskPool = std::make_shared<util::RetryTaskPool>(config);

    backupTarget = std::make_shared<BackupTarget>();
    ASSERT_TRUE(backupTarget != nullptr);

    std::list<std::shared_ptr<BaseFileService>> ufss;
    ufss.emplace_back(mockUfs);
    auto ret = backupTarget->Initialize("mock_target", taskPool, ufss);
    ASSERT_EQ(0, ret);
}

void TestBackupTarget::TearDown()
{
    backupTarget->Destroy();
    backupTarget.reset();
    GlobalMockObject::verify();
}

TEST_F(TestBackupTarget, create_dir_when_not_exist)
{
    std::string name = "/create_dir_when_not_exist";

    auto ret = backupTarget->CreateDir(name, 0700, 0, 0);
    ASSERT_EQ(0, ret) << " ret = " << ret << ", err=" << errno << ", str=" << strerror(errno);
}

TEST_F(TestBackupTarget, create_dir_already_exist)
{
    std::string name = "/create_dir_already_exist";
    auto ret = mockUfs->CreateDirectory(name, FileMode{ 0700 });
    ASSERT_EQ(0, ret) << " ret = " << ret << ", err=" << errno << ", str=" << strerror(errno);

    ret = backupTarget->CreateDir(name, 0700, 0, 0);
    ASSERT_EQ(0, ret) << " ret = " << ret << ", err=" << errno << ", str=" << strerror(errno);
}

TEST_F(TestBackupTarget, create_dir_exist_file)
{
    std::string name = "/create_dir_exist_file";
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    output.reset();

    auto ret = backupTarget->CreateDir(name, 0700, 0, 0);
}

TEST_F(TestBackupTarget, stat_not_exist_file)
{
    struct stat statBuf {};
    std::string name = "/stat_not_exist_file";
    auto ret = backupTarget->StatFile(name, statBuf);
    ASSERT_NE(0, ret);
    EXPECT_EQ(ENOENT, errno);
}

TEST_F(TestBackupTarget, stat_exist_file)
{
    struct stat statBuf {};
    std::string name = "/stat_exist_file";

    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    output.reset();

    auto ret = backupTarget->StatFile(name, statBuf);
    ASSERT_EQ(0, ret) << " ret = " << ret << ", err=" << errno << ", str=" << strerror(errno);
}

TEST_F(TestBackupTarget, stat_exist_dir)
{
    struct stat statBuf {};
    std::string name = "/stat_exist_dir";

    auto ret = mockUfs->CreateDirectory(name, FileMode{ 0755 });
    ASSERT_EQ(0, ret) << " ret = " << ret << ", err=" << errno << ", str=" << strerror(errno);

    ret = backupTarget->StatFile(name, statBuf);
    ASSERT_EQ(0, ret) << " ret = " << ret << ", err=" << errno << ", str=" << strerror(errno);
}

TEST_F(TestBackupTarget, remove_file_test)
{
    std::string name = "/remove_file_test";
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    output.reset();
    backupTarget->RemoveFile(name, 0);
}

TEST_F(TestBackupTarget, remove_file_and_sync_test)
{
    std::string name = "/remove_file_and_sync_test";
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    output.reset();
    auto ret = backupTarget->RemoveFileAndStageSync(FileTrace{ name, 0 });
    ASSERT_EQ(0, ret) << "err=" << errno << ", str=" << strerror(errno);
}

TEST_F(TestBackupTarget, upload_file_test)
{
    std::string name = "/upload_file_test";
    struct stat statBuf {};
    bool force = true;
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    output.reset();
    backupTarget->UploadFile(FileTrace{ name, 0 }, statBuf, force);
}

TEST_F(TestBackupTarget, upload_no_force_file_test)
{
    std::string name = "/upload_no_force_file_test";
    struct stat statBuf {};
    bool force = false;
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    output.reset();
    backupTarget->UploadFile(FileTrace{ name, 0 }, statBuf, force);
}

TEST_F(TestBackupTarget, create_file_sync_test)
{
    std::string name = "/create_file_sync_test";
    struct stat statBuf {};
    auto ret = backupTarget->CreateFileAndStageSync(FileTrace{ name, 0 }, statBuf);
    ASSERT_EQ(0, ret) << "err=" << errno << ", str=" << strerror(errno);
}

TEST_F(TestBackupTarget, make_file_cache_test)
{
    std::string name = "/make_file_cache_test";
    struct stat statBuf {};
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    auto paraLoadCtxPtr = std::make_shared<ParallelLoadContext>(1);
    paraLoadCtxPtr->RecordTaskOffset(0);
    TaskInfo taskInfo{ 0, DEFAULT_THREAD_DATA_SIZE, 0, DEFAULT_THREAD_DATA_SIZE, paraLoadCtxPtr };

    backupTarget->MakeFileCache(FileTrace{ name, 0 }, taskInfo);
}

TEST_F(TestBackupTarget, upload_no_force_file_three_times_test)
{
    std::string name = "/upload_no_force_file_three_times_test";
    struct stat statBuf {};
    statBuf.st_mtime = 0;
    bool force = false;
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    output.reset();
    backupTarget->UploadFile(FileTrace{ name, 0 }, statBuf, force);
    backupTarget->UploadFile(FileTrace{ name, 0 }, statBuf, force);
    statBuf.st_mtime = 1;
    backupTarget->UploadFile(FileTrace{ name, 0 }, statBuf, force);
}

TEST_F(TestBackupTarget, do_bakup_file_test)
{
    std::string name = "/do_bakup_file_test";
    uint64_t taskId = 1;
    FileTrace trace{ name, 0 };
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    std::shared_ptr<BaseFileService> tmpUfs { nullptr };
    auto view = UnderFsFileView(tmpUfs);

    EXPECT_FALSE(backupTarget->DoBackupFile(taskId, trace, view));
}

TEST_F(TestBackupTarget, real_bakup_all_parent_directory_test)
{
    std::string name = "/real_bakup_all_parent_directory_test";
    uint64_t taskId = 1;
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    std::shared_ptr<BaseFileService> tmpUfs { nullptr };
    auto view = UnderFsFileView(tmpUfs);

    EXPECT_FALSE(backupTarget->RealBackupAllParentDirectory(taskId, name, view));
}

TEST_F(TestBackupTarget, real_bakup_file_test)
{
    std::string name = "/real_bakup_file_test";
    uint64_t taskId = 1;
    FileTrace trace{ name, 0 };
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    std::shared_ptr<BaseFileService> tmpUfs { nullptr };
    auto view = UnderFsFileView(tmpUfs);

    EXPECT_FALSE(backupTarget->RealBackupFile(taskId, trace, view));
}

TEST_F(TestBackupTarget, do_make_file_cache_test)
{
    std::string name = "/do_make_file_cache_test";
    uint64_t taskId = 1;
    FileTrace trace{ name, 0 };
    auto output = mockUfs->PutFile(name, FileMode{ 0600 });
    ASSERT_TRUE(output != nullptr) << "err=" << errno << ", str=" << strerror(errno);
    std::shared_ptr<BaseFileService> tmpUfs { nullptr };
    auto view = UnderFsFileView(tmpUfs);
    auto paraLoadCtxPtr = std::make_shared<ParallelLoadContext>(1);
    paraLoadCtxPtr->RecordTaskOffset(0);
    TaskInfo taskInfo{ 0, DEFAULT_THREAD_DATA_SIZE, 0, DEFAULT_THREAD_DATA_SIZE, paraLoadCtxPtr };

    backupTarget->DoMakeFileCache(taskId, trace, view, taskInfo);
}

TEST_F(TestBackupTarget, add_view_abnormal_branch)
{
    std::string name = "/add_view_abnormal_branch";
    FileTrace trace{ name, 0 };
    struct stat buf{};

    std::shared_ptr<BaseFileService> tmpUfs { nullptr };
    auto view = UnderFsFileView(tmpUfs);
    auto ret = view.AddUploadFileToView(trace, buf);
    ASSERT_TRUE(ret);

    clock_gettime(CLOCK_REALTIME_COARSE, &buf.st_mtim);
    ret = view.AddUploadFileToView(trace, buf);
    ASSERT_TRUE(ret);

    FileTrace trace2{ name, 1 };
    ret = view.AddUploadFileToView(trace2, buf);
    ASSERT_FALSE(ret);

    FileTrace trace3{ name, 1 };
    buf.st_ino = 1;
    ret = view.AddUploadFileToView(trace3, buf);
    ASSERT_TRUE(ret);
}

TEST_F(TestBackupTarget, remove_view_no_exist)
{
    auto view = UnderFsFileView(mockUfs);
    auto ret = view.DoRemoveFile(0, "test", 0, 0, true);
    ASSERT_TRUE(ret);
}

TEST_F(TestBackupTarget, remove_view_file_name_too_long)
{
    std::string path(4096, 'a');
    auto view = UnderFsFileView(mockUfs);
    auto ret = view.DoRemoveFile(0, path, 0, 0, true);
    ASSERT_FALSE(ret);
}

TEST_F(TestBackupTarget, remove_view_file_inode_not_equal)
{
    std::string name = "/remove_view_file_inode_not_equal";
    std::string fullName = ufsPath + name;

    std::ofstream{ fullName };
    auto view = UnderFsFileView(mockUfs);
    auto ret = view.DoRemoveFile(0, name, 0, 1, true);
    ASSERT_TRUE(ret);
}

TEST_F(TestBackupTarget, remove_view_file)
{
    std::string name = "/remove_view_file";
    std::string fullName = ufsPath + name;

    auto fd = open(fullName.c_str(), O_CREAT | O_TRUNC | O_RDWR);
    ASSERT_GT(fd, 0) << "open failed, errno: " << errno << ": " << strerror(errno);

    struct stat st{};
    ASSERT_EQ(fstat(fd, &st), 0) << "stat failed, errno: " << errno << ": " << strerror(errno);

    close(fd);
    auto view = UnderFsFileView(mockUfs);
    auto ret = view.DoRemoveFile(0, name, 0, st.st_ino, true);
    ASSERT_TRUE(ret);
}

TEST_F(TestBackupTarget, remove_view_dir)
{
    std::string name = "/remove_view_dir";
    std::string fullName = ufsPath + name;

    auto fd = open(fullName.c_str(), O_CREAT | O_TRUNC | O_RDWR);
    ASSERT_GT(fd, 0) << "open failed, errno: " << errno << ": " << strerror(errno);

    struct stat st{};
    ASSERT_EQ(fstat(fd, &st), 0) << "stat failed, errno: " << errno << ": " << strerror(errno);

    close(fd);
    auto view = UnderFsFileView(mockUfs);
    auto ret = view.DoRemoveFile(0, name, 0, st.st_ino, false);
    ASSERT_FALSE(ret);
}

TEST_F(TestBackupTarget, backup_target_init_empty_ufs)
{
    BackupTarget target;
    util::RetryTaskPool::RetryTaskConfig config{};
    BackupTarget::TaskPool pool = std::make_shared<util::RetryTaskPool>(config);
    std::list<std::shared_ptr<BaseFileService>> mufs;
    auto ret = target.Initialize("test", pool, mufs);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestBackupTarget, backup_target_remove_file_is_dir)
{
    std::string file = "/backup_target_remove_file_is_dir";
    std::string stageFile = file + ".m.stg";

    std::string fullFile = ufsPath + file;
    std::string fullStageFile = ufsPath + stageFile;

    ASSERT_EQ(mkdir(fullFile.c_str(), S_IRUSR), 0) << "mkdir failed, errno: " << errno << ": " << strerror(errno);
    ASSERT_EQ(mkdir(fullStageFile.c_str(), S_IRUSR), 0) << "mkdir failed, errno: " << errno << ": " << strerror(errno);

    // .stg is dir
    auto ret = backupTarget->RemoveFileAndStageSync(FileTrace(file, 0));
    ASSERT_EQ(-1, ret);

    rmdir(fullStageFile.c_str());
    std::ofstream{ fullStageFile };

    // origin path is dir
    ret = backupTarget->RemoveFileAndStageSync(FileTrace(file, 0));
    ASSERT_EQ(0, ret);
}

TEST_F(TestBackupTarget, backup_target_remove_file_get_meta_failed)
{
    auto file = "/backup_target_remove_file_get_meta_failed";
    std::ofstream{ file };
    MOCKER(unlink).stubs().will(returnValue(0));
    MOCKER(lstat).stubs().will(returnValue(-1));
    errno = EBUSY;
    auto ret = backupTarget->RemoveFileAndStageSync(FileTrace(file, 0));
    ASSERT_EQ(-1, ret);
}

TEST_F(TestBackupTarget, backup_target_create_dir_name_too_long)
{
    std::string file(4096, 'a');
    auto ret = backupTarget->CreateDir(file, 0, 0, 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestBackupTarget, backup_target_check_stg_mtime_normal)
{
    std::string file = "/backup_target_check_stg_mtime_normal";
    std::string stageFile = file + ".m.stg";

    for (auto view : backupTarget->underFsFileView) {
        auto ret = backupTarget->CheckStgMtime(view, stageFile);
        ASSERT_EQ(FILE_BEEN_REMOVED, ret);
    }
}

TEST_F(TestBackupTarget, backup_target_check_stg_mtime_lstat_faile)
{
    std::string file = "/backup_target_check_stg_mtime_lstat_faile";
    std::string stageFile = file + ".m.stg";

    for (auto view : backupTarget->underFsFileView) {
        MOCKER(lstat).stubs().will(returnValue(-1));
        errno = EBUSY;
        auto ret = backupTarget->CheckStgMtime(view, stageFile);
        ASSERT_EQ(MTIME_NO_CHANGE_TIMEOUT, ret);
    }
}

TEST_F(TestBackupTarget, backup_target_check_stg_mtime_no_update)
{
    std::string file = "/backup_target_check_stg_mtime_no_update";
    std::string stageFile = file + ".m.stg";
    std::ofstream{ ufsPath + stageFile };

    for (auto view : backupTarget->underFsFileView) {
        auto ret = backupTarget->CheckStgMtime(view, stageFile);
        ASSERT_EQ(MTIME_NO_CHANGE_TIMEOUT, ret);
    }
}

TEST_F(TestBackupTarget, backup_target_try_lock_file_no_exist)
{
    std::string path = "/backup_target_try_lock_file_no_exist";
    std::shared_ptr<ock::ufs::FileLock> fileLock;
    struct stat buf;
    for (auto view : backupTarget->underFsFileView) {
        auto ret = backupTarget->TryLockStg(0, view, fileLock, path, buf);
        ASSERT_EQ(LOCK_ERROR, ret);
    }
}

TEST_F(TestBackupTarget, create_one_parent_normal)
{
    for (auto view : backupTarget->underFsFileView) {
        auto ret = backupTarget->CreateOneParent(0, "/create_one_parent_failed", FileMode(S_IRUSR), view);
        ASSERT_TRUE(ret);
    }
}

TEST_F(TestBackupTarget, create_one_parent_is_file)
{
    auto path = "/create_one_parent_is_file";
    std::ofstream{ ufsPath + path };
    for (auto view : backupTarget->underFsFileView) {
        auto ret = backupTarget->CreateOneParent(0, path, FileMode(0), view);
        ASSERT_TRUE(ret);
    }
}
}