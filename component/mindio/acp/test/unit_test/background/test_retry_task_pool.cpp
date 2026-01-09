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

#include <gtest/gtest.h>
#include <mockcpp/mokc.h>

#include "memfs_api.h"
#include "service_configure.h"
#include "retry_task_pool.h"

using namespace ock::memfs;
using namespace ock::common::config;
using namespace ock::bg::util;

namespace {

bool TestFunc()
{
    return true;
}

TEST(TestRetryTaskPool, test_RetryTaskPool_submit_start_should_return_success)
{
    RetryTaskPool::RetryTaskConfig config;
    config.name = "MindBG";
    config.autoEvictFile = 0;
    config.thCnt = 1;
    config.maxFailCntForUnserviceable = 1;
    config.retryTimes = 1;
    config.retryIntervalSec = 1;
    config.firstWaitMs = 1;
    RetryTaskPool *pool = new RetryTaskPool(config);
    ASSERT_NE(nullptr, pool);

    pool->Submit(TestFunc);

    auto ret = pool->Start();
    ASSERT_EQ(0, ret);

    pool->ReportCCAE(true);
    pool->ReportCCAE(false);

    delete pool;
}

TEST(TestRetryTaskPool, report_ccae_normal)
{
    RetryTaskPool::RetryTaskConfig config;
    config.name = "MindBG";
    config.autoEvictFile = 0;
    config.thCnt = 1;
    config.maxFailCntForUnserviceable = 1;
    config.retryTimes = 1;
    config.retryIntervalSec = 1;
    config.firstWaitMs = 1;
    RetryTaskPool pool(config);
    auto ret = pool.Start();
    ASSERT_EQ(0, ret);

    auto &instance = ServiceConfigure::GetInstance();
    auto workPath = instance.GetWorkPath();

    auto ccaeDir = workPath + "/ccae";
    mkdir(ccaeDir.c_str(), S_IRUSR | S_IWUSR);

    pool.ReportCCAE(true);
    pool.ReportCCAE(false);

    rmdir(ccaeDir.c_str());
}

TEST(TestRetryTaskPool, process_abnormal_retry_task)
{
    RetryTaskPool::RetryTaskConfig config;
    config.name = "MindBG";
    config.autoEvictFile = 0;
    config.thCnt = 1;
    config.maxFailCntForUnserviceable = 1;
    config.retryTimes = 1;
    config.retryIntervalSec = 1;
    config.firstWaitMs = 1;
    RetryTaskPool pool(config);
    auto ret = pool.Start();
    ASSERT_EQ(0, ret);

    MOCKER((void(*)(bool))MemFsApi::Serviceable).stubs().will(ignoreReturnValue());

    auto successFunc = []() { return true; };
    auto failedFunc = []() { return false; };

    RetryTask::Process(std::make_shared<RetryTask>(failedFunc, pool));
    RetryTask::Process(std::make_shared<RetryTask>(failedFunc, pool));
    RetryTask::Process(std::make_shared<RetryTask>(successFunc, pool));
}
}