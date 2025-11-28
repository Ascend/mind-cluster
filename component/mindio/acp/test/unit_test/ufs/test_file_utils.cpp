/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
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

#include <dirent.h>
#include "test_file_utils.h"

using namespace ock::ufs;

void TestFileUtils::RemoveDirectory(const std::string &path)
{
    auto dir = opendir(path.c_str());
    if (dir == nullptr) {
        return;
    }

    struct dirent *entry;
    while ((entry = readdir(dir)) != nullptr) {
        auto name = std::string(entry->d_name);
        if (name == "." || name == "..") {
            continue;
        }

        auto fullName = path;
        fullName.append("/").append(name);
        if (entry->d_type != DT_DIR) {
            unlink(fullName.c_str());
        } else {
            RemoveDirectory(fullName);
        }
    }

    closedir(dir);
    rmdir(path.c_str());
}

TEST_F(TestFileUtils, test_read_full_should_return_success)
{
    // create file
    auto fd = open(testFile1, O_CREAT | O_WRONLY);
    ASSERT_TRUE(fd > 0);

    // write data
    uint32_t dataSize = 1024;
    char data[dataSize];
    for (int i = 0; i < dataSize; ++i) {
        data[i] = 'a';
    }

    auto count = write(fd, data, dataSize);
    ASSERT_EQ(count, dataSize);

    // after write, close file, and then reopen it
    int32_t ret = close(fd);
    ASSERT_EQ(ret, 0);

    fd = open(testFile1,  O_RDONLY);
    ASSERT_TRUE(fd > 0);

    // read data
    uint8_t readBuffer[dataSize];
    auto readBytes = utils::FileUtils::ReadFull(fd, readBuffer, dataSize);
    ASSERT_EQ(readBytes, dataSize);
    for (int j = 0; j < dataSize; ++j) {
        ASSERT_EQ(readBuffer[j], data[j]);
    }

    unlink(testFile.c_str());
}

TEST_F(TestFileUtils, test_read_full_should_return_fail)
{
    // read data
    uint32_t dataSize = 1024;
    uint8_t readBuffer[dataSize];
    auto readBytes = utils::FileUtils::ReadFull(-1, readBuffer, dataSize);
    ASSERT_EQ(readBytes, -1);
}

TEST_F(TestFileUtils, test_write_full_should_return_success)
{
    // create file
    auto fd = open(testFile1, O_CREAT | O_WRONLY);
    ASSERT_TRUE(fd > 0);

    // write data
    uint32_t dataSize = 1024;
    uint8_t data[dataSize];
    for (int i = 0; i < dataSize; ++i) {
        data[i] = 'a';
    }

    auto count = utils::FileUtils::WriteFull(fd, data, dataSize);
    ASSERT_EQ(count, dataSize);

    // after write, close file, and then reopen it
    int32_t ret = close(fd);
    ASSERT_EQ(ret, 0);

    fd = open(testFile1,  O_RDONLY);
    ASSERT_TRUE(fd > 0);

    // read data
    char readBuffer[dataSize];
    auto readBytes = read(fd, readBuffer, dataSize);
    ASSERT_EQ(readBytes, dataSize);
    for (int j = 0; j < dataSize; ++j) {
        ASSERT_EQ(readBuffer[j], data[j]);
    }

    unlink(testFile.c_str());
}

TEST_F(TestFileUtils, test_write_full_should_return_fail)
{
    // read data
    uint32_t dataSize = 1024;
    uint8_t readBuffer[dataSize];
    auto readBytes = utils::FileUtils::WriteFull(-1, readBuffer, dataSize);
    ASSERT_EQ(readBytes, -1);
}