 /*
  * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
#include <gtest/gtest.h>
#include <memory>
#include <thread>
#include <chrono>

#include "service_configure.h"
#include "memfs_sdk_api.h"
#include "c2python_api.h"

using namespace ock::memfs;
using namespace ock::memfs::sdk;
using namespace ock::common::config;

class TestC2pythonApi : public testing::Test {
public:
    static void SetUpTestSuite();
    static void TearDownTestSuite();

protected:
    static std::string pathPrefix;
};

std::string TestC2pythonApi::pathPrefix = "/mnt/dpc01/test_c2python/";

void TestC2pythonApi::SetUpTestSuite()
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

void TestC2pythonApi::TearDownTestSuite()
{
    MemFsClientUnInitialize();
    std::this_thread::sleep_for(std::chrono::seconds(3));
    if (pathPrefix.empty() || pathPrefix == "/") {
        return;
    }
    std::string cmd = "rm -rf " + pathPrefix;
    system(cmd.c_str());
}

namespace {

TEST_F(TestC2pythonApi, test_save_file_direct_e2e)
{
    std::string path = pathPrefix + "save_direct.pt";
    WriteableFile writeHandle(path, S_IRUSR | S_IWUSR);
    auto result = writeHandle.Initialize();
    ASSERT_EQ(0, result);
    uint64_t fileSize = 1U << 20;
    char *writeContent = new char[fileSize];
    std::unique_ptr<char[]> writeContentPtr(writeContent);
    for (uint64_t i = 0; i < fileSize; ++i) {
        writeContent[i] = 'a';
    }
    result = writeHandle.Write(writeContent, fileSize);
    ASSERT_EQ(0, result);
    result = writeHandle.Close();
    ASSERT_EQ(0, result);
}

TEST_F(TestC2pythonApi, test_save_file_vector_e2e)
{
    std::string path = pathPrefix + "save_vec.pt";
    WriteableFile writeHandle(path, S_IRUSR | S_IWUSR);
    auto result = writeHandle.Initialize();
    ASSERT_EQ(0, result);
    uint64_t fileSize = 1U << 20;
    char *writeContent = new char[fileSize];
    std::unique_ptr<char[]> writeContentPtr(writeContent);
    for (uint64_t j = 0; j < fileSize; ++j) {
        writeContent[j] = 'a';
    }
    int vectorNum = 3;
    std::vector<Buffer> buffers;
    for (int i = 0; i < vectorNum; ++i) {
        buffers.emplace_back(writeContent, fileSize);
    }
    result = writeHandle.WriteV(buffers);
    ASSERT_EQ(0, result);
    result = writeHandle.Close();
    ASSERT_EQ(0, result);
}
}