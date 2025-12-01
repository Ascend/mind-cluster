/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2025. All rights reserved.
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

#include "error_code.h"
#include "test_memfs_sdk_api.h"

std::string TestMemfsSdkApi::testPrefix = "/mnt/dpc01/test_sdk_api/";

constexpr uint32_t TEST_BLK_SIZE = 4096;

namespace {

TEST_F(TestMemfsSdkApi, test_make_dir_should_return_success)
{
    std::string path = testPrefix + "test";
    auto result = MemFsMkDir(path.c_str(), 0, true);
    ASSERT_EQ(result, MFS_OK);
}

TEST_F(TestMemfsSdkApi, test_make_dir_twice_should_return_)
{
    std::string path = testPrefix + "test1/";
    auto result = MemFsMkDir(path.c_str(), 0);
    ASSERT_EQ(result, 0);
    result = MemFsMkDir(path.c_str(), 0);
    ASSERT_EQ(result, -1);
}

TEST_F(TestMemfsSdkApi, test_open_and_close_file_should_return_success)
{
    std::string path = testPrefix + "test2";
    auto result = MemFsMkDir(path.c_str(), 0);
    ASSERT_EQ(result, 0);
    path = testPrefix + "test2/1";
    auto fd = MemFsOpenFile(path.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_TRUE(fd >= 0);
    result = MemFsClose(fd);
    ASSERT_EQ(result, 0);
}

TEST_F(TestMemfsSdkApi, test_open_file_twice_should_return_error)
{
    std::string path = testPrefix + "test3";
    auto result = MemFsMkDir(path.c_str(), 0);
    ASSERT_EQ(result, 0);
    path = testPrefix + "test3/1";
    auto fd = MemFsOpenFile(path.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_TRUE(fd >= 0);
    auto errfd = MemFsOpenFile(path.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_TRUE(errfd < 0);
    result = MemFsClose(fd);
    ASSERT_EQ(result, 0);
}

TEST_F(TestMemfsSdkApi, test_open_file_with_empty_path_should_return_error)
{
    std::string path = testPrefix + "test4";
    auto result = MemFsMkDir(path.c_str(), 0);
    ASSERT_EQ(result, 0);
    path = "";
    auto fd = MemFsOpenFile(path.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_TRUE(fd < 0);
}

TEST_F(TestMemfsSdkApi, test_open_file_with_parent_not_exist_should_return_success)
{
    std::string path = testPrefix + "test5/1";
    auto fd = MemFsOpenFile(path.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_TRUE(fd > 0);
}

TEST_F(TestMemfsSdkApi, test_get_file_size_should_return_success)
{
    // make directory.
    std::string path = testPrefix + "test6";
    auto result = MemFsMkDir(path.c_str(), 0);
    ASSERT_EQ(result, 0);

    // create file and write data.
    path = testPrefix + "test6/1";
    auto fd = MemFsOpenFile(path.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_TRUE(fd >= 0);
    uint32_t size = TEST_BLK_SIZE;
    char data[size];
    auto writenSize = MemFsWrite(fd, (uintptr_t)data, size);
    ASSERT_EQ(writenSize, size);

    // close file.
    result = MemFsClose(fd);
    ASSERT_EQ(result, 0);

    // open file again and get file size.
    fd = MemFsOpenFile(path.c_str(), O_RDONLY);
    auto fileSize = MemFsGetSize(fd);
    ASSERT_EQ(fileSize, writenSize);
}

TEST_F(TestMemfsSdkApi, test_seek_and_tell_file_should_return_success)
{
    // make directory.
    std::string path = testPrefix + "test7";
    auto result = MemFsMkDir(path.c_str(), 0);
    ASSERT_EQ(result, 0);

    // create file and write data.
    path = testPrefix + "test/1";
    auto fd = MemFsOpenFile(path.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_TRUE(fd >= 0);
    uint32_t size = TEST_BLK_SIZE * 10;
    char data[size];
    auto writenSize = MemFsWrite(fd, (uintptr_t)data, size);
    ASSERT_EQ(writenSize, size);

    // close file.
    MemFsClose(fd);

    // open file again and get file size.
    fd = MemFsOpenFile(path.c_str(), O_RDONLY);
    result = MemFsSeek(fd, 0, SEEK_SET);
    ASSERT_EQ(result, 0);
    auto fileSize = MemFsGetSize(fd);
    ASSERT_EQ(fileSize, writenSize);

    // normal case
    result = MemFsSeek(fd, 512, SEEK_SET);
    ASSERT_EQ(result, 0);
    auto tell = MemFsTell(fd);
    ASSERT_EQ(tell, 512);

    // error offset, more than file size
    result = MemFsSeek(fd, fileSize + 1, SEEK_SET);
    ASSERT_EQ(result, -1);
    tell = MemFsTell(fd);
    ASSERT_EQ(tell, 512);

    // error whence
    result = MemFsSeek(fd, 512, 9);
    ASSERT_EQ(result, -1);
    tell = MemFsTell(fd);
    ASSERT_EQ(tell, 512);

    // normal case
    result = MemFsSeek(fd, 513, SEEK_CUR);
    ASSERT_EQ(result, 0);
    tell = MemFsTell(fd);
    ASSERT_EQ(tell, 1025);

    // error offset, more than file size
    result = MemFsSeek(fd, 512, SEEK_END);
    ASSERT_EQ(result, -1);
    tell = MemFsTell(fd);
    ASSERT_EQ(tell, 1025);

    // normal case
    result = MemFsSeek(fd, 0, SEEK_END);
    ASSERT_EQ(result, 0);
    tell = MemFsTell(fd);
    ASSERT_EQ(tell, fileSize);

    // error offset, less than 0
    result = MemFsSeek(fd, - (fileSize + 1), SEEK_END);
    ASSERT_EQ(result, -1);
    tell = MemFsTell(fd);
    ASSERT_EQ(tell, fileSize);
}

TEST_F(TestMemfsSdkApi, test_write_and_read_file_should_return_success)
{
    // make directory.
    std::string path = testPrefix + "test9";
    auto result = MemFsMkDir(path.c_str(), 0);
    ASSERT_EQ(result, 0);

    // create file and write data.
    path = testPrefix + "test9/1";
    auto fd = MemFsOpenFile(path.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_TRUE(fd >= 0);
    uint32_t size = TEST_BLK_SIZE * 10 + 13;
    char writeData[size];
    for (int i = 0; i < size; ++i) {
        writeData[i] = 'a' + size % 26;
    }

    uint32_t writeTime = 100;
    for (int i = 0; i < writeTime; ++i) {
        auto writenSize = MemFsWrite(fd, (uintptr_t)writeData, size);
        ASSERT_EQ(writenSize, size);
    }

    result = MemFsClose(fd);
    ASSERT_EQ(result, 0);

    fd = MemFsOpenFile(path.c_str(), O_RDONLY);
    ASSERT_TRUE(fd >= 0);

    char readData[size];
    uint64_t offset = 0;
    for (int i = 0; i < writeTime; ++i) {
        auto readSize = MemFsRead(fd, (uintptr_t)readData, offset, size);
        ASSERT_EQ(readSize, size);
        for (int j = 0; j < size; ++j) {
            if (readData[i] != writeData[i]) {
                std::cout << "not match data index:" << i << std::endl;
                ASSERT_TRUE(readData[i] == writeData[i]);
            }
        }
        offset += readSize;
    }
}

TEST_F(TestMemfsSdkApi, test_preload_file_should_return_success)
{
    // create file for write.
    std::string path = testPrefix + "test10/1";

    auto fd = MemFsOpenFile(path.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
    ASSERT_TRUE(fd >= 0);

    auto result = MemFsClose(fd);
    ASSERT_EQ(result, 0);

    result = MemfsPreloadFile(path.c_str());
    ASSERT_EQ(result, 0);
}
}