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

#ifndef OCK_DFS_TEST_FILE_UTILS_H
#define OCK_DFS_TEST_FILE_UTILS_H

#include <thread>
#include <semaphore.h>
#include <gtest/gtest.h>
#include <fcntl.h>
#include <cerrno>
#include <chrono>
#include <algorithm>
#include <random>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>

#include "ufs_log.h"
#include "file_utils.h"

class TestFileUtils : public testing::Test {
public:
    static void RemoveDirectory(const std::string &path);

protected:
    const std::string testFile = "/tmp/test_file";
    const char testFile1[15] = "/tmp/test_file";
};

#endif // OCK_DFS_TEST_FILE_UTILS_H