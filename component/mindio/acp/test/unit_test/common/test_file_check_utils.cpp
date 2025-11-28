/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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
#include <fcntl.h>
#include <unistd.h>

#include "file_check_utils.h"
#include "mem_file_system.h"
#include "mem_fs_inode.h"

using namespace ock::common;
using namespace ock::memfs;


TEST(TestFileCheckUtils, test_check_file_exist_should_return_success)
{
    const char testCheckFile[25] = "/tmp/test_ock_check_file";
    // create file
    auto fd = open(testCheckFile, O_CREAT | O_WRONLY);
    ASSERT_TRUE(fd > 0);

    // write data
    uint32_t dataSize = 1024;
    char data[dataSize];
    for (int i = 0; i < dataSize; ++i) {
        data[i] = 'a';
    }

    auto count = write(fd, data, dataSize);
    ASSERT_EQ(count, dataSize);

    // after write, close file
    int32_t ret = close(fd);
    ASSERT_EQ(ret, 0);

    bool exist = FileCheckUtils::CheckFileExists(testCheckFile);
    ASSERT_EQ(exist, true);

    unlink(testCheckFile);
}

TEST(TestFileCheckUtils, test_check_dir)
{
    const char tmpDIr[5] = "/tmp";
    bool exist = FileCheckUtils::CheckDirectoryExists(tmpDIr);
    ASSERT_EQ(exist, true);

    exist = FileCheckUtils::CheckDirectoryExists("/foo/bar");
    ASSERT_EQ(exist, false);
}

TEST(TestFileCheckUtils, test_check_is_symlink)
{
    const char testCheckFile[25] = "/tmp/test_ock_check_file";
    // create file
    auto fd = open(testCheckFile, O_CREAT | O_WRONLY);
    ASSERT_TRUE(fd > 0);

    // write data
    char data[] = {1, 2, 3};
    auto count = write(fd, data, sizeof(data));
    ASSERT_EQ(count, sizeof(data));

    // after write, close file
    int32_t ret = close(fd);
    ASSERT_EQ(ret, 0);

    bool isLink = FileCheckUtils::IsSymlink(testCheckFile);
    ASSERT_EQ(isLink, false);

    // create a link
    const char testCheckFileLink[30] = "/tmp/test_ock_check_file_link";
    int32_t retLink = symlink(testCheckFile, testCheckFileLink);
    ASSERT_EQ(retLink, 0);

    isLink = FileCheckUtils::IsSymlink(testCheckFileLink);
    ASSERT_EQ(isLink, true);

    unlink(testCheckFile);
    unlink(testCheckFileLink);
}

TEST(TestFileCheckUtils, test_check_regular_file_path)
{
    std::string errMsg{};
    bool isFile = FileCheckUtils::RegularFilePath("", "/tmp", errMsg);
    ASSERT_EQ(errMsg, "The file path:  is empty.");
    ASSERT_EQ(isFile, false);

    isFile = FileCheckUtils::RegularFilePath("/foo/bar.txt", "", errMsg);
    ASSERT_EQ(errMsg, "The file path basedir:  is empty.");
    ASSERT_EQ(isFile, false);

    const char testCheckFile[25] = "/tmp/test_ock_check_file";
    const char tmpDIr[5] = "/tmp";
    // create file
    auto fd = open(testCheckFile, O_CREAT | O_WRONLY);
    ASSERT_TRUE(fd > 0);

    // write data
    char data[] = {1, 2, 3};
    auto count = write(fd, data, sizeof(data));
    ASSERT_EQ(count, sizeof(data));

    // after write, close file
    int32_t ret = close(fd);
    ASSERT_EQ(ret, 0);

    isFile = FileCheckUtils::RegularFilePath(testCheckFile, tmpDIr, errMsg);
    ASSERT_EQ(isFile, true);

    unlink(testCheckFile);
}

TEST(TestFileCheckUtils, test_check_file_valid)
{
    const char testCheckFile[25] = "/tmp/test_ock_check_file";
    // create file
    auto fd = open(testCheckFile, O_CREAT | O_WRONLY);
    ASSERT_TRUE(fd > 0);

    // write data
    char data[] = {1, 2, 3};
    auto count = write(fd, data, sizeof(data));
    ASSERT_EQ(count, sizeof(data));

    // after write, close file
    int32_t ret = close(fd);
    ASSERT_EQ(ret, 0);

    std::string errMsg{};
    bool isValid = FileCheckUtils::IsFileValid(testCheckFile, errMsg, true, FileCheckUtils::FILE_MODE_400, true, false);
    ASSERT_EQ(isValid, true);

    std::string errMsgTwo{};
    isValid = FileCheckUtils::IsFileValid(testCheckFile, errMsgTwo, true, FileCheckUtils::FILE_MODE_400, true, true);
    ASSERT_EQ(isValid, false);

    unlink(testCheckFile);
}

TEST(TestFileCheckUtils, test_recycle_inodes)
{
    auto &instance = InodeEvictor::GetInstance();
    auto ret = instance.Initialize();
    ASSERT_EQ(ret, 0);
    uint64_t blockSize = 16UL << 20;
    MemFileSystem memFileSystem(blockSize, 1, "test");
    ret = memFileSystem.Initialize();
    ASSERT_EQ(ret, 0);

    instance.RecycleInodes(blockSize);
    memFileSystem.Destroy();
    InodeEvictor::GetInstance().Destroy();
}