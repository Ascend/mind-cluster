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
#include <fstream>
#include <mockcpp/mokc.h>

#include "file_check_utils.h"
#include "mem_file_system.h"
#include "mem_fs_inode.h"

using namespace ock::common;
using namespace ock::memfs;

namespace {

class TestFileCheckUtils : public testing::Test {
public:
    void SetUp() override;
    void TearDown() override;
};

void TestFileCheckUtils::SetUp() {}
void TestFileCheckUtils::TearDown()
{
    GlobalMockObject::verify();
}

TEST_F(TestFileCheckUtils, test_check_file_exist_should_return_success)
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

TEST_F(TestFileCheckUtils, test_check_dir)
{
    const char tmpDIr[5] = "/tmp";
    bool exist = FileCheckUtils::CheckDirectoryExists(tmpDIr);
    ASSERT_EQ(exist, true);

    exist = FileCheckUtils::CheckDirectoryExists("/foo/bar");
    ASSERT_EQ(exist, false);
}

TEST_F(TestFileCheckUtils, test_check_is_symlink)
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

TEST_F(TestFileCheckUtils, test_check_regular_file_path)
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

TEST_F(TestFileCheckUtils, test_check_file_valid)
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

TEST_F(TestFileCheckUtils, test_sym_link_no_exist)
{
    auto ret = FileCheckUtils::IsSymlink("///");
    ASSERT_FALSE(ret);
}

TEST_F(TestFileCheckUtils, test_regular_file_long_path)
{
    int len = 4097;
    std::string path(len, 'a');
    std::string errMsg;
    auto ret = FileCheckUtils::RegularFilePath(path, "/tmp", errMsg);
    ASSERT_FALSE(ret);
}

TEST_F(TestFileCheckUtils, test_regular_file_invalid)
{
    std::string file = "/tmp/test_regular_file_invalid";
    std::string errMsg;
    std::ofstream{ file };

    // prefix error
    ASSERT_FALSE(FileCheckUtils::RegularFilePath(file, "/temp", errMsg));

    std::string linkFile = "/tmp/test_regular_file_invalid_link";
    int32_t retLink = symlink(file.c_str(), linkFile.c_str());
    ASSERT_EQ(0, retLink);

    // symlink error
    ASSERT_FALSE(FileCheckUtils::RegularFilePath(linkFile, "/tmp", errMsg));

    MOCKER(realpath).stubs().will(returnValue(static_cast<char*>(nullptr)));
    // realpath error
    ASSERT_FALSE(FileCheckUtils::RegularFilePath(file, "/tmp", errMsg));

    unlink(linkFile.c_str());
    unlink(file.c_str());
}

TEST_F(TestFileCheckUtils, test_file_is_invalid)
{
    std::string path = "/tmp/test_file_is_invalid";
    std::string errMsg;
    // no exist
    auto ret = FileCheckUtils::IsFileValid(path, errMsg);
    ASSERT_FALSE(ret);

    std::ofstream ofs(path);
    ASSERT_TRUE(ofs.is_open());

    // file size == 0
    ret = FileCheckUtils::IsFileValid(path, errMsg);
    ASSERT_FALSE(ret);

    // checkDataSize invalid
    ofs << "hello";
    ofs.close();
    MOCKER(FileCheckUtils::CheckDataSize).stubs().will(returnValue(false));
    ret = FileCheckUtils::IsFileValid(path, errMsg);
    ASSERT_FALSE(ret);
    GlobalMockObject::verify();

    // no permission
    MOCKER(FileCheckUtils::CheckDataSize).stubs().will(returnValue(true));
    MOCKER(FileCheckUtils::ConstrainOwner).stubs().will(returnValue(false));
    ret = FileCheckUtils::IsFileValid(path, errMsg);
    ASSERT_FALSE(ret);
    unlink(path.c_str());
}

TEST_F(TestFileCheckUtils, test_constrain_owner_invalid)
{
    std::string file = "/tmp/test_constrain_owner_invalid";
    std::string errMsg;
    auto ret = FileCheckUtils::ConstrainOwner(file, errMsg);
    ASSERT_FALSE(ret);

    std::ofstream{ file };

    auto uid = getuid();
    MOCKER(getuid).stubs().will(returnValue(uid + 1));
    ret = FileCheckUtils::ConstrainOwner(file, errMsg);
    ASSERT_FALSE(ret);

    unlink(file.c_str());
}

TEST_F(TestFileCheckUtils, test_constrain_permission_no_exist)
{
    std::string file = "/tmp/test_constrain_permission_invalid";
    std::string errMsg;
    auto ret = FileCheckUtils::ConstrainPermission(file, 0, errMsg);
    ASSERT_FALSE(ret);
}

TEST_F(TestFileCheckUtils, test_get_file_size_invalid)
{
    std::string file = "/tmp/test_get_file_size_invalid";

    auto ret = FileCheckUtils::GetFileSize(file);
    ASSERT_EQ(0, ret);

    MOCKER(FileCheckUtils::CheckFileExists).stubs().will(returnValue(true));
    ret = FileCheckUtils::GetFileSize(file);
    ASSERT_EQ(0, ret);

    MOCKER(FileCheckUtils::RegularFilePath).stubs().will(returnValue(true));
    ret = FileCheckUtils::GetFileSize(file);
    ASSERT_EQ(0, ret);

    std::ofstream{ file };

    MOCKER(fseek).stubs().will(returnValue(1));
    ret = FileCheckUtils::GetFileSize(file);
    ASSERT_EQ(0, ret);

    unlink(file.c_str());
}

TEST_F(TestFileCheckUtils, test_get_base_file_name_end_backslash)
{
    std::string path = "/tmp/test_get_base_file_name_no_exist/";
    auto baseName = FileCheckUtils::GetBaseFileName(path);
    ASSERT_EQ(baseName, "test_get_base_file_name_no_exist");
}

TEST_F(TestFileCheckUtils, test_check_data_size_invalid)
{
    uint64_t maxFileSize = 1UL;

    auto ret = FileCheckUtils::CheckDataSize(0, maxFileSize);
    ASSERT_FALSE(ret);

    maxFileSize = 10UL;
    ret = FileCheckUtils::CheckDataSize(0, maxFileSize);
    ASSERT_FALSE(ret);
}

TEST_F(TestFileCheckUtils, test_remove_prefix_path_normal)
{
    std::string base = "test_remove_prefix_path_normal";
    auto lessPath = FileCheckUtils::RemovePrefixPath(base);
    ASSERT_EQ(base, lessPath);

    lessPath = FileCheckUtils::RemovePrefixPath("/" + base);
    ASSERT_EQ(base, lessPath);

    lessPath = FileCheckUtils::RemovePrefixPath("/tmp/" + base);
    ASSERT_EQ(base, lessPath);
}
}