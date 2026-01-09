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
#include <fstream>
#include <gtest/gtest.h>
#include <mockcpp/mokc.h>

#include "memfs_api.h"
#include "memfs_file_util.h"
#include "pacific_adapter.h"
#include "service_configure.h"

#define private public
#include "mem_fs_backup_initiator.h"
#undef private

using namespace ock::ufs;
using namespace ock::memfs;
using namespace ock::bg::backup;

namespace {

class TestMemFsBackupInitiator : public testing::Test {
public:
    static void SetUpTestSuite();
    static void TearDownTestSuite();

    void SetUp() override;
    void TearDown() override;

    static void RestSize();

    static int MockStatFile(BackupTarget *self, const std::string &path, struct stat &buf);
    static void MockerGetShareFileCfg(uint64_t &blockSize, uint64_t &blockCnt);
    static int MockGetMeta(const std::string &path, struct stat &statBuf);
    static int MockGetFileMeta(int fd, struct stat &statBuf);
    static int MockGetFileBlocks(int fd, std::vector<uint64_t> &blocks);

protected:
    static std::string mountPath;
    static std::shared_ptr<BaseFileService> mockUfs;
    static uint64_t mockSize;
    static uint64_t mockBlkSize;
    static uint64_t mockBlkCnt;
    static uint64_t mockInode;
};

uint64_t TestMemFsBackupInitiator::mockSize;
uint64_t TestMemFsBackupInitiator::mockBlkSize;
uint64_t TestMemFsBackupInitiator::mockBlkCnt;
uint64_t TestMemFsBackupInitiator::mockInode;
std::string TestMemFsBackupInitiator::mountPath;
std::shared_ptr<BaseFileService> TestMemFsBackupInitiator::mockUfs;

void TestMemFsBackupInitiator::SetUpTestSuite()
{
    mountPath = "./test_memfs_backup_initiator";
    if (FileUtil::Exist(mountPath)) {
        ASSERT_TRUE(FileUtil::RemoveDirRecursive(mountPath))
            << "remove file failed " << errno << ": " << strerror(errno);
    }
    ASSERT_EQ(0, mkdir(mountPath.c_str(), S_IRWXU));

    mockUfs = std::make_shared<PacificAdapter>(mountPath);
    ASSERT_TRUE(mockUfs != nullptr);
    ASSERT_EQ(0, ock::common::config::ServiceConfigure::GetInstance().Initialize());
}

void TestMemFsBackupInitiator::TearDownTestSuite()
{
    mockUfs.reset();
    ASSERT_TRUE(FileUtil::RemoveDirRecursive(mountPath)) << "remove file failed " << errno << ": " << strerror(errno);
    ock::common::config::ServiceConfigure::GetInstance().Destroy();
}

void TestMemFsBackupInitiator::SetUp() {}
void TestMemFsBackupInitiator::TearDown()
{
    RestSize();
    GlobalMockObject::verify();
}

int TestMemFsBackupInitiator::MockStatFile(BackupTarget *self, const std::string &path, struct stat &buf)
{
    buf.st_size = mockSize;
    buf.st_ino = 0UL;
    return 0;
}

void TestMemFsBackupInitiator::MockerGetShareFileCfg(uint64_t &blockSize, uint64_t &blockCnt)
{
    blockSize = mockBlkSize;
    blockCnt = mockBlkCnt;
}

int TestMemFsBackupInitiator::MockGetMeta(const std::string &path, struct stat &statBuf)
{
    statBuf.st_ino = mockInode;
    return 0;
}

int TestMemFsBackupInitiator::MockGetFileMeta(int fd, struct stat &statBuf)
{
    statBuf.st_blksize = mockBlkSize;
    return 0;
}

int TestMemFsBackupInitiator::MockGetFileBlocks(int fd, std::vector<uint64_t> &blocks)
{
    blocks.emplace_back(0UL);
    return 0;
}

void TestMemFsBackupInitiator::RestSize()
{
    mockSize = 0UL;
    mockBlkSize = 0UL;
    mockBlkCnt = 0UL;
    mockInode = 0UL;
}

TEST_F(TestMemFsBackupInitiator, get_attribute_get_meta_failed)
{
    MemFsBackupInitiator initiator{};
    struct stat buf{};
    auto path = mountPath + "/get_attribute_get_meta_failed";
    MOCKER(MemFsApi::GetMeta).stubs().will(returnValue(-1));
    errno = EBUSY;
    auto ret = initiator.GetAttribute(0, path, buf);
    ASSERT_NE(0, ret);
}

TEST_F(TestMemFsBackupInitiator, remove_stg_file_unlink_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/remove_stg_file_unlink_failed";

    MOCKER(MemFsApi::Unlink).stubs().will(returnValue(-1));
    auto ret = initiator.RemoveStageFileFromUfs(path, mockUfs);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, remove_stg_file_name_too_long)
{
    MemFsBackupInitiator initiator{};
    std::string path(4096, 'a');

    MOCKER(MemFsApi::Unlink).stubs().will(returnValue(0));
    auto ret = initiator.RemoveStageFileFromUfs(path, mockUfs);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, multi_task_write_open_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/multi_task_write_open_failed";
    TaskInfo taskInfo(0, 1, 0, 1, std::make_shared<ParallelLoadContext>(1));
    struct stat buf{};

    MOCKER(MemFsApi::OpenFile).stubs().will(returnValue(-1));
    auto ret = initiator.MultiTasksDoWrite(path, taskInfo, buf, mockUfs);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, multi_task_write_get_blk_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = "/multi_task_write_get_blk_failed";
    auto fullPath = mountPath + path;
    std::ofstream{ fullPath };
    TaskInfo taskInfo(0, 1, 0, 1, std::make_shared<ParallelLoadContext>(1));
    struct stat buf{};

    MOCKER(MemFsApi::OpenFile).stubs().will(returnValue(0));
    MOCKER(MemFsApi::GetFileBlocks).stubs().will(returnValue(-1));
    auto ret = initiator.MultiTasksDoWrite(path, taskInfo, buf, mockUfs);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, split_upload_file_task_not_all_success)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/split_upload_file_task_not_all_success";
    struct stat buf{};
    ASSERT_EQ(0, ock::common::config::ServiceConfigure::GetInstance().Initialize());

    union MockerHelper {
        int (MemFsBackupInitiator::*multiTasksDoWrite)(const std::string&, const TaskInfo&, const struct stat&,
            BackupInitiator::UFS);
        int (*mockMultiTasksDoWrite)(MemFsBackupInitiator*, const std::string&, const TaskInfo&, const struct stat&,
            BackupInitiator::UFS);
    };

    MockerHelper helper{};
    helper.multiTasksDoWrite = &MemFsBackupInitiator::MultiTasksDoWrite;
    MOCKCPP_NS::mockAPI("&MemFsBackupInitiator::MultiTasksDoWrite",
        helper.mockMultiTasksDoWrite).defaults().will(returnValue(0));
    MOCKER(unlink).stubs().will(returnValue(-1));

    buf.st_size = 32UL << 30;
    auto ret = initiator.SplitUploadFileTask(path, buf, mockUfs);
    ASSERT_EQ(-1, ret);

    // name too long
    path = std::string(4096, 'a');
    ret = initiator.SplitUploadFileTask(path, buf, mockUfs);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, multi_copy_file_to_ufs_open_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/multi_copy_file_to_ufs_open_failed";
    MOCKER(MemFsApi::OpenFile).stubs().will(returnValue(-1));
    auto ret = initiator.MultiCopyFileToUfs(0, path, mockUfs);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, multi_copy_file_to_ufs_get_meta_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/multi_copy_file_to_ufs_get_meta_failed";
    MOCKER(MemFsApi::OpenFile).stubs().will(returnValue(0));
    MOCKER(MemFsApi::GetFileMeta).stubs().will(returnValue(-1));
    auto ret = initiator.MultiCopyFileToUfs(0, path, mockUfs);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, record_to_memfs_task_ret_task_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/record_to_memfs_task_ret_task_failed";
    TaskInfo taskInfo(0, 1, 0, 1, std::make_shared<ParallelLoadContext>(1));

    auto ret = initiator.RecordToMemfsTaskResult(0, path, 0, taskInfo);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, record_to_memfs_task_ret_not_all_finished)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/record_to_memfs_task_ret_not_all_finished";
    auto paraLoadCtx = std::make_shared<ParallelLoadContext>(2);
    paraLoadCtx->taskRetryCntMap[0] = 0;
    TaskInfo taskInfo(0, 1, 0, 1, paraLoadCtx);

    auto ret = initiator.RecordToMemfsTaskResult(0, path, 0, taskInfo);
    ASSERT_EQ(0, ret);
}

TEST_F(TestMemFsBackupInitiator, record_to_memfs_task_ret_failedCnt_gt_0)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/record_to_memfs_task_ret_failedCnt_gt_0";
    auto paraLoadCtx = std::make_shared<ParallelLoadContext>(1);
    paraLoadCtx->taskRetryCntMap[0] = 0;
    paraLoadCtx->failedCnt.fetch_add(1U);
    TaskInfo taskInfo(0, 1, 0, 1, paraLoadCtx);

    auto ret = initiator.RecordToMemfsTaskResult(0, path, 0, taskInfo);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, record_to_memfs_task_ret_close_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/record_to_memfs_task_ret_close_failed";
    auto paraLoadCtx = std::make_shared<ParallelLoadContext>(1);
    paraLoadCtx->taskRetryCntMap[0] = 0;
    TaskInfo taskInfo(0, 1, 0, 1, paraLoadCtx);

    MOCKER(MemFsApi::TruncateFile).stubs().will(returnValue(0));
    auto ret = initiator.RecordToMemfsTaskResult(0, path, 0, taskInfo);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, record_to_memfs_task_ret_normal)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/record_to_memfs_task_ret_normal";
    auto paraLoadCtx = std::make_shared<ParallelLoadContext>(1);
    paraLoadCtx->taskRetryCntMap[0] = 0;
    TaskInfo taskInfo(0, 1, 0, 1, paraLoadCtx);

    MOCKER(MemFsApi::TruncateFile).stubs().will(returnValue(0));
    MOCKER(MemFsApi::CloseFile).stubs().will(returnValue(0));
    auto ret = initiator.RecordToMemfsTaskResult(0, path, 0, taskInfo);
    ASSERT_EQ(0, ret);
}

TEST_F(TestMemFsBackupInitiator, copy_file_to_memfs_get_meta_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = "/copy_file_to_memfs_get_meta_failed";
    auto fullPath = mountPath + path;
    TaskInfo taskInfo(0, 1, 0, 1, std::make_shared<ParallelLoadContext>(1));

    MOCKER(MemFsApi::GetFileMeta).stubs().will(returnValue(-1));
    auto ret = initiator.CopyFileToMemfs(0, path, mockUfs, taskInfo);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, copy_file_to_memfs_get_file_no_exist)
{
    MemFsBackupInitiator initiator{};
    auto path = "/copy_file_to_memfs_get_file_no_exist";
    auto fullPath = mountPath + path;
    TaskInfo taskInfo(0, 1, 0, 1, std::make_shared<ParallelLoadContext>(1));

    MOCKER(MemFsApi::GetFileMeta).stubs().will(returnValue(0));
    auto ret = initiator.CopyFileToMemfs(0, path, mockUfs, taskInfo);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, copy_file_to_memfs_get_blk_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = "/copy_file_to_memfs_get_blk_failed";
    auto fullPath = mountPath + path;
    std::ofstream{ fullPath };
    TaskInfo taskInfo(0, 1, 0, 1, std::make_shared<ParallelLoadContext>(1));

    MOCKER(MemFsApi::GetFileMeta).stubs().will(returnValue(0));
    MOCKER(MemFsApi::GetFileBlocks).stubs().will(returnValue(-1));
    auto ret = initiator.CopyFileToMemfs(0, path, mockUfs, taskInfo);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, copy_file_to_memfs_normal)
{
    MemFsBackupInitiator initiator{};
    auto path = "/copy_file_to_memfs_normal";
    auto fullPath = mountPath + path;
    std::ofstream ofs(fullPath);
    ofs << "hello";
    ofs.close();
    TaskInfo taskInfo(0, 1, 0, 5UL, std::make_shared<ParallelLoadContext>(1));
    char *buf = new char[10];

    MOCKER(MemFsApi::GetFileMeta).stubs().will(invoke(MockGetFileMeta));
    MOCKER(MemFsApi::GetFileBlocks).stubs().will(invoke(MockGetFileBlocks));
    MOCKER(MemFsApi::BlockToAddress).stubs().will(returnValue(static_cast<void *>(buf)));

    mockBlkSize = 10UL;
    auto ret = initiator.CopyFileToMemfs(0, path, mockUfs, taskInfo);
    ASSERT_EQ(0, ret);
    delete[] buf;
}

TEST_F(TestMemFsBackupInitiator, split_and_submit_task_thread_threshold_test)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/split_and_submit_task_thread_threshold_test";
    struct stat buf{};
    FileTrace trace(path, 0);

    union MockerHelper {
        void (BackupTarget::*makeFileCache)(const FileTrace &trace, const TaskInfo &taskInfo);
        void (*mockMakeFileCache)(BackupTarget *self, const FileTrace &trace, const TaskInfo &taskInfo);
    };

    MockerHelper helper{};
    helper.makeFileCache = &BackupTarget::MakeFileCache;
    MOCKCPP_NS::mockAPI("&BackupTarget::MakeFileCache", helper.mockMakeFileCache).stubs().will(ignoreReturnValue());

    // single thread
    buf.st_size = 1UL;
    initiator.SplitAndSubmitTask(0, buf, trace, path);

    // max thread
    buf.st_size = 32UL << 30;
    initiator.SplitAndSubmitTask(0, buf, trace, path);
}

TEST_F(TestMemFsBackupInitiator, open_file_notify_marked)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/open_file_notify_marked";
    initiator.marked = true;
    auto ret = initiator.OpenFileNotify(0, path, 0, 0);
    ASSERT_EQ(0, ret);
}

TEST_F(TestMemFsBackupInitiator, open_file_notify_get_meta_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/open_file_notify_get_meta_failed";
    initiator.marked = false;

    MOCKER(MemFsApi::GetMeta).stubs().will(returnValue(-1));
    errno = EBUSY;
    auto ret = initiator.OpenFileNotify(0, path, 0, 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, new_file_notify_marked)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/new_file_notify_marked";
    initiator.marked = true;
    auto ret = initiator.NewFileNotify(path, 0);
    ASSERT_EQ(0, ret);
}

TEST_F(TestMemFsBackupInitiator, new_file_notify_get_meta_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/new_file_notify_get_meta_failed";
    initiator.marked = false;

    MOCKER(MemFsApi::GetMeta).stubs().will(returnValue(-1));
    errno = EBUSY;
    auto ret = initiator.NewFileNotify(path, 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, new_file_notify_compare_diff_inode)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/new_file_notify_compare_diff_inode";
    initiator.marked = false;
    mockInode = 1UL;

    union MockerHelper {
        int (BackupTarget::*createFileAndStageSync)(const FileTrace &trace, const struct stat &buf);
        int (*mockCreateFileAndStageSync)(BackupTarget *self, const FileTrace &trace, const struct stat &buf);
    };

    MockerHelper helper{};
    helper.createFileAndStageSync = &BackupTarget::CreateFileAndStageSync;
    MOCKCPP_NS::mockAPI("&BackupTarget::CreateFileAndStageSync",
        helper.mockCreateFileAndStageSync).stubs().will(returnValue(0));

    MOCKER(MemFsApi::GetMeta).stubs().will(invoke(MockGetMeta));
    auto ret = initiator.NewFileNotify(path, 0);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, new_file_notify_normal)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/new_file_notify_get_meta_failed";
    initiator.marked = false;

    union MockerHelper {
        int (BackupTarget::*createFileAndStageSync)(const FileTrace &trace, const struct stat &buf);
        int (*mockCreateFileAndStageSync)(BackupTarget *self, const FileTrace &trace, const struct stat &buf);

        void (BackupTarget::*uploadFile)(const FileTrace &trace, const struct stat &fileStat, bool force);
        void (*mockUploadFile)(BackupTarget *self, const FileTrace &trace, const struct stat &fileStat, bool force);
    };

    MockerHelper helper{};
    helper.createFileAndStageSync = &BackupTarget::CreateFileAndStageSync;
    MOCKCPP_NS::mockAPI("&BackupTarget::CreateFileAndStageSync",
        helper.mockCreateFileAndStageSync).stubs().will(returnValue(0));
    helper.uploadFile = &BackupTarget::UploadFile;
    MOCKCPP_NS::mockAPI("&BackupTarget::UploadFile", helper.mockUploadFile).stubs().will(ignoreReturnValue());

    MOCKER(MemFsApi::GetMeta).stubs().will(returnValue(0));
    auto ret = initiator.NewFileNotify(path, 0);
    ASSERT_EQ(0, ret);
}

TEST_F(TestMemFsBackupInitiator, preload_file_notify_marked)
{
    MemFsBackupInitiator initiator{};
    auto path = mountPath + "/preload_file_notify_marked";
    initiator.marked = true;
    auto ret = initiator.PreloadFileNotify(path);
    ASSERT_EQ(0, ret);
}

TEST_F(TestMemFsBackupInitiator, preload_file_notify_read_meta_failed)
{
    MemFsBackupInitiator initiator{};
    initiator.marked = false;
    auto path = mountPath + "/preload_file_notify_read_meta_failed";
    std::ofstream{ path };

    union MockerHelper {
        int (BackupTarget::*statFile)(const std::string &path, struct stat &buf);
        int (*mockStatFile)(BackupTarget *self, const std::string &path, struct stat &buf);
    };

    MockerHelper helper{};
    helper.statFile = &BackupTarget::StatFile;
    MOCKCPP_NS::mockAPI("&BackupTarget::StatFile", helper.mockStatFile).stubs().will(returnValue(-1));

    auto ret = initiator.PreloadFileNotify(path);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, preload_file_notify_out_of_mem)
{
    MemFsBackupInitiator initiator{};
    initiator.marked = false;
    auto path = mountPath + "/preload_file_notify_out_of_mem";
    std::ofstream{ path };
    mockSize = 10UL;

    union MockerHelper {
        int (BackupTarget::*statFile)(const std::string &path, struct stat &buf);
        int (*mockStatFile)(BackupTarget *self, const std::string &path, struct stat &buf);
    };

    MockerHelper helper{};
    helper.statFile = &BackupTarget::StatFile;
    MOCKCPP_NS::mockAPI("&BackupTarget::StatFile", helper.mockStatFile).stubs().will(invoke(MockStatFile));
    MOCKER(MemFsApi::GetShareFileCfg).stubs().will(invoke(MockerGetShareFileCfg));

    auto ret = initiator.PreloadFileNotify(path);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, preload_file_notify_open_memfs_file_failed)
{
    MemFsBackupInitiator initiator{};
    initiator.marked = false;
    auto path = mountPath + "/preload_file_notify_open_memfs_file_failed";
    std::ofstream{ path };

    union MockerHelper {
        int (BackupTarget::*statFile)(const std::string &path, struct stat &buf);
        int (*mockStatFile)(BackupTarget *self, const std::string &path, struct stat &buf);
    };

    MockerHelper helper{};
    helper.statFile = &BackupTarget::StatFile;
    MOCKCPP_NS::mockAPI("&BackupTarget::StatFile", helper.mockStatFile).stubs().will(invoke(MockStatFile));
    MOCKER(MemFsApi::GetShareFileCfg).stubs().will(invoke(MockerGetShareFileCfg));
    MOCKER(MemFsApi::CreateAndOpenFile).stubs().will(returnValue(-1));

    auto ret = initiator.PreloadFileNotify(path);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, preload_file_notify_alloc_blk_failed)
{
    MemFsBackupInitiator initiator{};
    initiator.marked = false;
    auto path = mountPath + "/preload_file_notify_alloc_blk_failed";
    std::ofstream{ path };

    union MockerHelper {
        int (BackupTarget::*statFile)(const std::string &path, struct stat &buf);
        int (*mockStatFile)(BackupTarget *self, const std::string &path, struct stat &buf);
    };

    MockerHelper helper{};
    helper.statFile = &BackupTarget::StatFile;
    MOCKCPP_NS::mockAPI("&BackupTarget::StatFile", helper.mockStatFile).stubs().will(invoke(MockStatFile));
    MOCKER(MemFsApi::GetShareFileCfg).stubs().will(invoke(MockerGetShareFileCfg));
    MOCKER(MemFsApi::CreateAndOpenFile).stubs().will(returnValue(0));
    MOCKER(MemFsApi::AllocDataBlocks).stubs().will(returnValue(-1));

    auto ret = initiator.PreloadFileNotify(path);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, preload_file_notify_compare_diff_inode)
{
    MemFsBackupInitiator initiator{};
    initiator.marked = false;
    auto path = mountPath + "/preload_file_notify_compare_diff_inode";
    std::ofstream{ path };
    mockInode = 1UL;

    union MockerHelper {
        int (BackupTarget::*statFile)(const std::string &path, struct stat &buf);
        int (*mockStatFile)(BackupTarget *self, const std::string &path, struct stat &buf);
    };

    MockerHelper helper{};
    helper.statFile = &BackupTarget::StatFile;
    MOCKCPP_NS::mockAPI("&BackupTarget::StatFile", helper.mockStatFile).stubs().will(invoke(MockStatFile));
    MOCKER(MemFsApi::GetMeta).stubs().will(invoke(MockGetMeta));
    MOCKER(MemFsApi::GetShareFileCfg).stubs().will(invoke(MockerGetShareFileCfg));
    MOCKER(MemFsApi::CreateAndOpenFile).stubs().will(returnValue(0));
    MOCKER(MemFsApi::AllocDataBlocks).stubs().will(returnValue(0));

    auto ret = initiator.PreloadFileNotify(path);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, preload_file_notify_normal)
{
    MemFsBackupInitiator initiator{};
    initiator.marked = false;
    auto path = mountPath + "/preload_file_notify_normal";
    std::ofstream{ path };
    mockSize = 32UL << 20;
    mockBlkSize = 32UL << 20;
    mockBlkCnt = 128UL;

    union MockerHelper {
        int (BackupTarget::*statFile)(const std::string &path, struct stat &buf);
        int (*mockStatFile)(BackupTarget *self, const std::string &path, struct stat &buf);

        void (BackupTarget::*makeFileCache)(const FileTrace &trace, const TaskInfo &taskInfo);
        void (*mockMakeFileCache)(BackupTarget *self, const FileTrace &trace, const TaskInfo &taskInfo);
    };

    MockerHelper helper{};
    helper.statFile = &BackupTarget::StatFile;
    MOCKCPP_NS::mockAPI("&BackupTarget::StatFile", helper.mockStatFile).stubs().will(invoke(MockStatFile));
    helper.makeFileCache = &BackupTarget::MakeFileCache;
    MOCKCPP_NS::mockAPI("&BackupTarget::MakeFileCache", helper.mockMakeFileCache).stubs().will(ignoreReturnValue());
    MOCKER(MemFsApi::GetMeta).stubs().will(returnValue(0));
    MOCKER(MemFsApi::GetShareFileCfg).stubs().will(invoke(MockerGetShareFileCfg));
    MOCKER(MemFsApi::CreateAndOpenFile).stubs().will(returnValue(0));
    MOCKER(MemFsApi::AllocDataBlocks).stubs().will(returnValue(0));

    auto ret = initiator.PreloadFileNotify(path);
    ASSERT_EQ(0, ret);
}

TEST_F(TestMemFsBackupInitiator, multi_tasks_write_finish_file_name_too_long)
{
    MemFsBackupInitiator initiator{};
    std::string path(4096, 'a');

    auto ret = initiator.MultiTasksWriteFinish(path, mockUfs);
    ASSERT_EQ(-1, ret);
}

TEST_F(TestMemFsBackupInitiator, multi_tasks_write_finish_open_memfs_failed)
{
    MemFsBackupInitiator initiator{};
    auto path = "/multi_tasks_write_finish_open_memfs_failed";
    auto fullPath = mountPath + path;

    MOCKER(MemFsApi::OpenFile).stubs().will(returnValue(-1));
    auto ret = initiator.MultiTasksWriteFinish(path, mockUfs);
    ASSERT_EQ(-1, ret);
}
}