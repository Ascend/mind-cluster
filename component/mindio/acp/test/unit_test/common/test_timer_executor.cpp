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

#define private public
#include "timer_executor.h"
#undef private

using namespace ock::common;

void TestTimerFun()
{
    return;
}

TEST(TestTimerExecutor, test_RetryTaskPool_start_should_return_success)
{
    std::string timerName = "timer";
    TimerExecutor *timer = new TimerExecutor(timerName);
    ASSERT_NE(nullptr, timer);

    auto ret = timer->Submit(TestTimerFun, 10);
    ASSERT_NE(nullptr, ret);

    timer->Wait();

    timer->Shutdown();

    delete timer;
    timer = nullptr;
}

TEST(TestTimerExecutor, test_TaskFuture_finish_and_cancel)
{
    std::string timerName = "timer2";
    TimerExecutor *timer = new TimerExecutor(timerName);
    ASSERT_NE(nullptr, timer);

    auto future = timer->Submit(TestTimerFun, 10);
    std::shared_ptr<TaskFuture> testFuture = std::static_pointer_cast<TaskFuture>(future);
    testFuture->SetStarted();
    testFuture->SetFinished();
    auto finishRet = testFuture->Finished();
    ASSERT_EQ(true, finishRet);

    auto future2 = timer->Submit(TestTimerFun, 10);
    std::shared_ptr<TaskFuture> testFuture2 = std::static_pointer_cast<TaskFuture>(future2);
    testFuture2->SetStarted();
    auto cancelRet = testFuture2->Cancel();
    ASSERT_EQ(true, cancelRet);
    testFuture2->SetCancelled();
    cancelRet = testFuture2->IsCancelled();
    ASSERT_EQ(true, cancelRet);

    timer->Shutdown();

    delete timer;
    timer = nullptr;
}

TEST(TestTimerExecutor, test_TimerExecutor_get_out_one_task)
{
    std::string timerName = "timer3";
    TimerExecutor *timer = new TimerExecutor(timerName);
    ASSERT_NE(nullptr, timer);

    auto ret = timer->Submit(TestTimerFun, 10);

    RunTask* task = timer->GetOutOneTask();
    ASSERT_NE(nullptr, task);

    delete task;
    task = nullptr;

    timer->Shutdown();

    delete timer;
    timer = nullptr;
}

