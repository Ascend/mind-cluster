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

#ifndef OCK_DFS_TEST_MEMFS_SDK_API_H
#define OCK_DFS_TEST_MEMFS_SDK_API_H

#include <gtest/gtest.h>
#include <thread>
#include <chrono>

#include "service_configure.h"
#include "memfs_sdk_api.h"

using namespace ock::memfs;
using namespace ock::common::config;

class TestMemfsSdkApi : public testing::Test {
public:
    static void SetUpTestSuite();
    static void TearDownTestSuite();

protected:
    static std::string testPrefix;
};

void TestMemfsSdkApi::SetUpTestSuite()
{
    std::this_thread::sleep_for(std::chrono::seconds(3));
    std::string workPath = ServiceConfigure::GetInstance().GetWorkPath();
    std::map<std::string, std::string> serverInfo {
        { "memfs.data_block_pool_capacity_in_gb", "8" },
        { "memfs.data_block_size_in_mb", "16" },
        { "memfs.write.parallel.enabled", "true" },
        { "memfs.write.parallel.thread_num", "16" },
        { "memfs.write.parallel.slice_in_mb", "16" },
        { "background.backup.thread_num", "16" },
        { "background.backup.failed_auto_evict_file", "true"},
        { "background.backup.failed_max_cnt_for_unserviceable", "11"},
        { "server.worker.path", workPath },
        { "server.ockiod.path", workPath + "/bin/ockiod" }
    };
    MemFsClientInitialize(serverInfo);
}

void TestMemfsSdkApi::TearDownTestSuite()
{
    MemFsClientUnInitialize();
    std::this_thread::sleep_for(std::chrono::seconds(3));
    if (testPrefix.empty() || testPrefix == "/") {
        return;
    }
    std::string cmd = "rm -rf " + testPrefix;
    system(cmd.c_str());
}

#endif // OCK_DFS_TEST_MEMFS_SDK_API_H
