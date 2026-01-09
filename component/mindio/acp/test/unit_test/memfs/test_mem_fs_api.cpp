/**
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
#include <semaphore.h>

#include <gtest/gtest.h>

#include <cstdint>
#include <thread>
#include <iostream>
#include <vector>
#include "memfs_api.h"
#include "service_configure.h"

#include "memfs_sdk_api.h"
#include "mem_file_system.h"

using namespace ock::memfs;
using namespace ock::common::config;

class TestMemFsApi : public testing::Test {
public:
    static void SetUpTestSuite();
    static void TearDownTestSuite();
};

void TestMemFsApi::SetUpTestSuite()
{
    int confRet = ServiceConfigure::GetInstance().Initialize();
    if (confRet != 0) {
        std::cout << "service configure init failed." << std::endl;
        ASSERT_EQ(0, confRet);
        return;
    }

    int ret = MemFsApi::Initialize();
    std::cout << "test_mem_fs_api set up, ret is " << ret << std::endl;
    ASSERT_EQ(0, ret);
}

void TestMemFsApi::TearDownTestSuite()
{
    MemFsApi::Destroy();
    ServiceConfigure::GetInstance().Destroy();
    std::cout << "test_mem_fs_api tear down" << std::endl;
}

namespace {

TEST_F(TestMemFsApi, test_open_file_create_success)
{
    std::string path = "/TestOpenFile_Create_OK.txt";
    int validFlags = O_CREAT | O_TRUNC | O_WRONLY;

    mode_t mode = 0644;
    auto fd = MemFsApi::OpenFile(path, validFlags, mode);
    ASSERT_TRUE(fd > 0);

    int backup = MemFsApi::SetBackupFinished(fd);
    ASSERT_EQ(backup, 0);

    int discard = MemFsApi::DiscardFile(path, fd);
    ASSERT_EQ(discard, 0);

    int close = MemFsApi::CloseFile(fd);
    ASSERT_EQ(close, -1);
}

TEST_F(TestMemFsApi, test_open_file_create_invalid_flag)
{
    std::string path = "/TestOpenFile_Create_Invalid_Flag.txt";
    int invalidFlags = 100;
    mode_t mode = 0644;
    auto fd = MemFsApi::OpenFile(path, invalidFlags, mode);
    ASSERT_EQ(fd, -1);
}

TEST_F(TestMemFsApi, test_alloc_data_block_one)
{
    std::string path = "/TestAllocDataBlock_One.txt";
    int validFlags = O_CREAT | O_TRUNC | O_WRONLY;
    mode_t mode = 0644;
    auto fd = MemFsApi::OpenFile(path, validFlags, mode);
    ASSERT_TRUE(fd > 0);

    uint64_t blockId;
    uint64_t blockSize;
    int ret = MemFsApi::AllocDataBlock(fd, blockId, blockSize);
    ASSERT_EQ(ret, 0);
    ASSERT_EQ(blockSize, 128 * 1024 * 1024);
    ASSERT_TRUE(blockId > 0);

    int offset = MemFsApi::GetBlockOffset(blockId);
    ASSERT_TRUE(offset >= 0);

    int close = MemFsApi::CloseFile(fd);
    ASSERT_EQ(close, 0);
}

TEST_F(TestMemFsApi, test_alloc_data_blocks_multi)
{
    std::string path = "/TestAllocDataBlocks_Multi.txt";
    int validFlags = O_CREAT | O_TRUNC | O_WRONLY;
    mode_t mode = 0644;
    auto fd = MemFsApi::OpenFile(path, validFlags, mode);
    ASSERT_TRUE(fd > 0);

    uint64_t bytes = 129 * 1024 * 1024;
    std::vector<uint64_t> blocks;
    uint64_t blockSize;
    int ret = MemFsApi::AllocDataBlocks(fd, bytes, blocks, blockSize);
    ASSERT_EQ(ret, 0);
    ASSERT_EQ(blockSize, 128 * 1024 * 1024);
    ASSERT_EQ(blocks.size(), 2);
    ASSERT_TRUE(blocks[0] > 0);
    ASSERT_TRUE(blocks[1] > 0);

    int close = MemFsApi::CloseFile(fd);
    ASSERT_EQ(close, 0);
}

TEST_F(TestMemFsApi, test_oprate_file_success)
{
    std::string path = "/TestAllocDataBlocks_fuse.txt";
    int validFlags = O_CREAT | O_TRUNC | O_WRONLY;
    mode_t mode = 0644;
    auto fd = MemFsApi::OpenFile(path, validFlags, mode);
    ASSERT_TRUE(fd > 0);

    struct stat statBuf {};
    int result = MemFsApi::GetFileMeta(fd, statBuf);
    ASSERT_EQ(result, 0);

    int fdShare = MemFsApi::GetShareMemoryFd();
    ASSERT_TRUE(fdShare > 0);

    std::vector<uint64_t> blocks;
    result = MemFsApi::GetFileBlocks(fd, blocks);
    ASSERT_EQ(result, 1);

    result = MemFsApi::Chmod(path, 0777);
    ASSERT_EQ(result, 0);

    result = MemFsApi::Chown(path, 1001, 2002);
    ASSERT_EQ(result, 0);

    result = MemFsApi::TruncateFile(fd, 1);
    ASSERT_EQ(result, 0);
}

TEST_F(TestMemFsApi, test_create_directory)
{
    std::string path = "/test_create_directory";
    auto ret = MemFsApi::CreateDirectory(path, 0755, false);
    ASSERT_EQ(0, ret) << "create directory failed:" << errno << " : " << strerror(errno);

    struct stat statBuf {};
    int result = MemFsApi::GetMeta(path, statBuf);
    ASSERT_EQ(0, result);

    ret = MemFsApi::RemoveDirectory(path);
    ASSERT_EQ(0, ret) << "remove directory failed:" << errno << " : " << strerror(errno);
    result = MemFsApi::GetMeta(path, statBuf);
    ASSERT_EQ(-1, result);
    ASSERT_EQ(ENOENT, errno);
}

TEST_F(TestMemFsApi, test_stat_fs)
{
    struct statvfs buf {};
    auto ret = MemFsApi::GetFileSystemStat(buf);
    ASSERT_EQ(0, ret) << "stat filesystem failed:" << errno << " : " << strerror(errno);
}

TEST_F(TestMemFsApi, test_truncate_exist_file)
{
    std::string path = "/test_truncate_exist_file.txt";
    int validFlags = O_CREAT | O_TRUNC | O_WRONLY;
    mode_t mode = 0644;
    auto fd = MemFsApi::OpenFile(path, validFlags, mode);
    ASSERT_TRUE(fd > 0);

    uint64_t bytes = 129 * 1024 * 1024;
    std::vector<uint64_t> blocks;
    uint64_t blockSize;
    auto ret = MemFsApi::AllocDataBlocks(fd, bytes, blocks, blockSize);
    ASSERT_EQ(0, ret);
    ASSERT_EQ(128 * 1024 * 1024, blockSize);
    ASSERT_EQ(2, blocks.size());
    ASSERT_TRUE(blocks[0] > 0);
    ASSERT_TRUE(blocks[1] > 0);
    ret = MemFsApi::TruncateFile(fd, bytes);
    ASSERT_EQ(0, ret) << "TruncateFile failed: " << errno << " : " << strerror(errno);

    auto close = MemFsApi::CloseFile(fd);
    ASSERT_EQ(close, 0);

    struct stat statBuf {};
    ret = MemFsApi::GetMeta(path, statBuf);
    ASSERT_EQ(0, ret);
    ASSERT_EQ(static_cast<int64_t>(bytes), statBuf.st_size);

    fd = MemFsApi::OpenFile(path, validFlags, mode);
    ASSERT_GT(fd, 0);

    close = MemFsApi::CloseFile(fd);
    ASSERT_EQ(close, 0);

    ret = MemFsApi::GetMeta(path, statBuf);
    ASSERT_EQ(0, ret);
    ASSERT_EQ(0L, statBuf.st_size);
}

TEST_F(TestMemFsApi, test_open_exist_file_read)
{
    std::string path = "/test_truncate_exist_file.txt";
    int validFlags = O_CREAT | O_TRUNC | O_WRONLY;
    mode_t mode = 0644;
    auto fd = MemFsApi::OpenFile(path, validFlags, mode);
    ASSERT_TRUE(fd > 0);

    uint64_t bytes = 129 * 1024 * 1024;
    std::vector<uint64_t> blocks;
    uint64_t blockSize;
    auto ret = MemFsApi::AllocDataBlocks(fd, bytes, blocks, blockSize);
    ASSERT_EQ(0, ret);
    ASSERT_EQ(128 * 1024 * 1024, blockSize);
    ASSERT_EQ(2, blocks.size());
    ASSERT_TRUE(blocks[0] > 0);
    ASSERT_TRUE(blocks[1] > 0);
    ret = MemFsApi::TruncateFile(fd, bytes);
    ASSERT_EQ(0, ret) << "TruncateFile failed: " << errno << " : " << strerror(errno);

    auto close = MemFsApi::CloseFile(fd);
    ASSERT_EQ(close, 0);

    struct stat statBuf {};
    ret = MemFsApi::GetMeta(path, statBuf);
    ASSERT_EQ(0, ret);
    ASSERT_EQ(static_cast<int64_t>(bytes), statBuf.st_size);

    fd = MemFsApi::OpenFile(path, O_RDONLY);
    ASSERT_GT(fd, 0);

    ret = MemFsApi::GetFileMeta(fd, statBuf);
    EXPECT_EQ(0, ret);
    EXPECT_EQ(static_cast<int64_t>(bytes), statBuf.st_size);

    close = MemFsApi::CloseFile(fd);
    EXPECT_EQ(close, 0);
}

TEST_F(TestMemFsApi, test_link_file_source_is_dir)
{
    std::string source = "/test_link_file_source_is_dir.txt";
    std::string target = "/test_truncate_exist_file_target.txt";

    auto ret = MemFsApi::CreateDirectory(source, 0755);
    ASSERT_EQ(0, ret);

    ret = MemFsApi::Link(source, target);
    EXPECT_NE(0, ret);
}

TEST_F(TestMemFsApi, test_link_file_dest_dir_not_exist)
{
    std::string source = "/test_link_file_dest_dir_not_exist.txt";
    std::string target = "/dest_not_exist_dir/test_truncate_exist_file_target.txt";

    auto fd = MemFsApi::OpenFile(source, O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_GT(fd, 0);
    MemFsApi::CloseFile(fd);

    auto ret = MemFsApi::Link(source, target);
    EXPECT_NE(0, ret);
}

TEST_F(TestMemFsApi, test_link_file_target_exist)
{
    std::string source = "/test_link_file_target_exist.txt";
    std::string target = "/test_link_file_target_exist_target.txt";

    auto fd = MemFsApi::OpenFile(source, O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_GT(fd, 0);
    MemFsApi::CloseFile(fd);

    fd = MemFsApi::OpenFile(target, O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_GT(fd, 0);
    MemFsApi::CloseFile(fd);

    auto ret = MemFsApi::Link(source, target);
    EXPECT_NE(0, ret);
}

TEST_F(TestMemFsApi, test_read_dir_not_exist)
{
    std::string path = "/test_read_dir_not_exist";
    std::vector<std::pair<std::string, bool>> entries;
    auto ret = MemFsApi::ReadDirectory(path, entries);
    ASSERT_NE(0, ret);
    ASSERT_EQ(ENOENT, errno) << "errno = " << errno << ", " << strerror(errno);
}

TEST_F(TestMemFsApi, test_rename_parent_not_exist)
{
    std::string source = "/test_rename_parent_not_exist/path1";
    std::string target = "/test_rename_parent_not_exist/path2";

    auto ret = MemFsApi::Rename(source, target);
    EXPECT_NE(0, ret);
}

TEST_F(TestMemFsApi, test_rename_source_not_exist)
{
    std::string source = "/test_rename_source_not_exist/path1";
    std::string target = "/test_rename_source_not_exist/path2";

    auto ret = MemFsApi::CreateDirectory("/test_rename_source_not_exist", 0755);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Rename(source, target);
    EXPECT_NE(0, ret);
}

TEST_F(TestMemFsApi, test_rename_sucess)
{
    std::string source = "/test_rename_sucess/path1";
    std::string target = "/test_rename_sucess/path2";

    auto ret = MemFsApi::CreateDirectory(source, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Rename(source, target);
    ASSERT_EQ(0, ret) << "rename path failed: " << errno << ":" << strerror(errno);
}

TEST_F(TestMemFsApi, test_rename_target_exist_empty)
{
    std::string source = "/test_rename_target_exist_empty/path1";
    std::string target = "/test_rename_target_exist_empty/path2";

    auto ret = MemFsApi::CreateDirectory(source, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    auto fd = MemFsApi::OpenFile(source + "/simple-file.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CreateDirectory(target, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Rename(source, target);
    ASSERT_EQ(0, ret) << "rename path failed: " << errno << ":" << strerror(errno);
}

TEST_F(TestMemFsApi, test_rename_target_exist_not_empty_failed)
{
    std::string source = "/test_rename_target_exist_not_empty_failed/path1";
    std::string target = "/test_rename_target_exist_not_empty_failed/path2";

    auto ret = MemFsApi::CreateDirectory(source, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    auto fd = MemFsApi::OpenFile(source + "/simple-file-1.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CreateDirectory(target, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    fd = MemFsApi::OpenFile(target + "/simple-file-2.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in target path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in target path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Rename(source, target);
    ASSERT_NE(0, ret) << "rename path should not success";
    ASSERT_EQ(ENOTEMPTY, errno) << "rename path should errno: " << strerror(errno);
}

TEST_F(TestMemFsApi, test_rename_target_exist_not_empty_force_success)
{
    std::string source = "/test_rename_target_exist_not_empty_force_success/path1";
    std::string target = "/test_rename_target_exist_not_empty_force_success/path2";

    auto ret = MemFsApi::CreateDirectory(source, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    auto fd = MemFsApi::OpenFile(source + "/simple-file-1.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CreateDirectory(target, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    fd = MemFsApi::OpenFile(target + "/simple-file-2.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in target path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in target path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Rename(source, target, MemFsConstants::RENAME_FLAG_FORCE);
    ASSERT_EQ(0, ret) << "rename path force failed: " << errno << ":" << strerror(errno);
}

TEST_F(TestMemFsApi, test_rename_diff_target_exist_not_empty_force_success)
{
    std::string source = "/test_rename_diff_target_exist_not_empty_force_success_1/path1";
    std::string target = "/test_rename_diff_target_exist_not_empty_force_success_2/path2";

    auto ret = MemFsApi::CreateDirectory(source, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    auto fd = MemFsApi::OpenFile(source + "/simple-file-1.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CreateDirectory(target, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    fd = MemFsApi::OpenFile(target + "/simple-file-2.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in target path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in target path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Rename(source, target, MemFsConstants::RENAME_FLAG_FORCE);
    ASSERT_EQ(0, ret) << "rename path force failed: " << errno << ":" << strerror(errno);
}

TEST_F(TestMemFsApi, test_exchange_diff_target_exist_not_empty_force_success)
{
    std::string source = "/test_exchange_diff_target_exist_not_empty_force_success_1/path1";
    std::string target = "/test_exchange_diff_target_exist_not_empty_force_success_2/path2";

    auto ret = MemFsApi::CreateDirectory(source, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    auto fd = MemFsApi::OpenFile(source + "/simple-file-1.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in source path failed: " << errno << ":" << strerror(errno);

    struct stat oldSourceStat {};
    ret = MemFsApi::GetMeta(source, oldSourceStat);
    ASSERT_EQ(0, ret) << "stat source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CreateDirectory(target, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    fd = MemFsApi::OpenFile(target + "/simple-file-2.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in target path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in target path failed: " << errno << ":" << strerror(errno);

    struct stat oldTargetStat {};
    ret = MemFsApi::GetMeta(target, oldTargetStat);
    ASSERT_EQ(0, ret) << "stat target path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Rename(source, target, MemFsConstants::RENAME_FLAG_EXCHANGE);
    ASSERT_EQ(0, ret) << "rename path force failed: " << errno << ":" << strerror(errno);

    struct stat newSourceStat {};
    ret = MemFsApi::GetMeta(source, newSourceStat);
    ASSERT_EQ(0, ret) << "stat source path failed: " << errno << ":" << strerror(errno);

    struct stat newTargetStat {};
    ret = MemFsApi::GetMeta(target, newTargetStat);
    ASSERT_EQ(0, ret) << "stat target path failed: " << errno << ":" << strerror(errno);

    EXPECT_EQ(oldSourceStat.st_ino, newTargetStat.st_ino);
    EXPECT_EQ(oldTargetStat.st_ino, newSourceStat.st_ino);
}

TEST_F(TestMemFsApi, test_rename_self_success)
{
    std::string source = "/test_rename_self_success/path1";

    auto ret = MemFsApi::CreateDirectory(source, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    auto fd = MemFsApi::OpenFile(source + "/simple-file.txt", O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Rename(source, source);
    ASSERT_EQ(0, ret) << "rename self path failed: " << errno << ":" << strerror(errno);
}

TEST_F(TestMemFsApi, test_rename_linked_file_success)
{
    std::string sourceDir = "/test_rename_linked_file_success/path1";
    std::string targetDir = "/test_rename_linked_file_success/path2";
    std::string source = sourceDir + "/simple-file-1.txt";
    std::string target = targetDir + "/simple-file-2.txt";

    auto ret = MemFsApi::CreateDirectory(sourceDir, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    auto fd = MemFsApi::OpenFile(source, O_CREAT | O_TRUNC | O_WRONLY, 0644);
    ASSERT_TRUE(fd >= 0) << "create file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, ret) << "close file in source path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::CreateDirectory(targetDir, 0755, true);
    ASSERT_EQ(0, ret) << "create parent path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Link(source, target);
    ASSERT_EQ(0, ret) << "link path failed: " << errno << ":" << strerror(errno);

    ret = MemFsApi::Rename(source, source);
    ASSERT_EQ(0, ret) << "rename self path failed: " << errno << ":" << strerror(errno);
}

TEST_F(TestMemFsApi, test_block_to_address_success)
{
    uint64_t blockId = 1;
    auto ret = MemFsApi::BlockToAddress(blockId);
    ASSERT_TRUE(ret != 0);
}

TEST_F(TestMemFsApi, test_unlink_file_exit_success)
{
    std::string name = "/test_unlink_file_exit_success";
    auto fd = MemFsApi::OpenFile(name, O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_GT(fd, 0);
    MemFsApi::CloseFile(fd);
    auto ret = MemFsApi::Unlink(name);
    ASSERT_TRUE(ret == 0);
}

TEST_F(TestMemFsApi, test_unlink_file_no_exit_failed)
{
    std::string name = "/test_unlink_file_no_exit_failed";
    auto ret = MemFsApi::Unlink(name);
    ASSERT_TRUE(ret != 0);
}

TEST_F(TestMemFsApi, test_service_is_able_success)
{
    bool state = true;
    auto ret = MemFsApi::Serviceable();
    ASSERT_TRUE(ret != 0);
    MemFsApi::Serviceable(state);
    state = false;
    MemFsApi::Serviceable(state);
}

TEST_F(TestMemFsApi, test_get_share_file_cfg_success)
{
    uint64_t blockSize = 128 * 1024 * 1024;
    uint64_t blockCnt = 10;
    MemFsApi::GetShareFileCfg(blockSize, blockCnt);
}

TEST_F(TestMemFsApi, test_register_file_op_notify_success)
{
    FileOpNotify fileOpNotify;
    auto ret = MemFsApi::RegisterFileOpNotify(fileOpNotify);
    ASSERT_TRUE(ret == 0);
}

TEST_F(TestMemFsApi, test_read_dir_is_exist)
{
    std::string path = "/test_create_directory";
    auto ret = MemFsApi::CreateDirectory(path, 0755, false);
    ASSERT_EQ(0, ret) << "create directory failed:" << errno << " : " << strerror(errno);

    struct stat statBuf {};
    int result = MemFsApi::GetMeta(path, statBuf);
    ASSERT_EQ(0, result);

    std::vector<std::pair<std::string, bool>> entries;
    ret = MemFsApi::ReadDirectory(path, entries);
    ASSERT_EQ(0, ret);
    ASSERT_EQ(ENOENT, errno) << "errno = " << errno << ", " << strerror(errno);
}

TEST_F(TestMemFsApi, test_set_external_stat)
{
    ExternalStat stat = [](const std::string &, struct stat &, MemfsFileAcl &) { return 0;};
    std::string source = "/test_set_external_stat.txt";
    auto fd = MemFsApi::OpenFile(source, O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_GT(fd, 0);
    MemFsApi::CloseFile(fd);
    MemFsApi::SetExternalStat(stat);
}

TEST_F(TestMemFsApi, test_create_and_open_file)
{
    std::string path = "/test/create/test_file.txt";
    uint64_t inodeNum = 1;
    MemFsApi::CreateAndOpenFile(path, inodeNum, 0);

    path = "test_file.txt";
    MemFsApi::CreateAndOpenFile(path, inodeNum, 0);
}

TEST_F(TestMemFsApi, test_preload_exist_file)
{
    std::string path = "/test_preload_exist_file.txt";
    int validFlags = O_CREAT | O_TRUNC | O_WRONLY;

    auto fd = MemFsApi::OpenFile(path, validFlags);
    ASSERT_TRUE(fd > 0) << "create file failed:" << errno << " : " << strerror(errno);

    auto result = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, result);

    result = MemFsApi::PreloadFile(path);
    ASSERT_EQ(0, result);
}

TEST_F(TestMemFsApi, test_preload_file_success)
{
    std::string path = "/test_preload_file.txt";
    int validFlags = O_CREAT | O_TRUNC | O_WRONLY;

    auto fd = MemFsApi::OpenFile(path, validFlags);
    ASSERT_TRUE(fd > 0) << "create file failed:" << errno << " : " << strerror(errno);

    auto result = MemFsApi::CloseFile(fd);
    ASSERT_EQ(0, result);

    result = MemFsApi::Unlink(path);
    ASSERT_EQ(0, result);

    result = MemFsApi::PreloadFile(path);
    ASSERT_EQ(0, result);
}

TEST_F(TestMemFsApi, test_preload_progress_view_interface)
{
    std::string path = "/test_preload_file_not_exist";
    auto ret = PreloadProgressView::PathExist(path);
    ASSERT_FALSE(ret);

    PreloadProgressView::InsertPath(path);
    PreloadProgressView::RemovePath(path);

    PreloadProgressView::Wait(1, path);
}
}