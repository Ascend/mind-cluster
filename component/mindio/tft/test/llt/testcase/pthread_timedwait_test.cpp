/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 */

#include <thread>
#include <unistd.h>
#include <gtest/gtest.h>
#include <mockcpp/mockcpp.hpp>
#include "common.h"

using namespace ock::ttp;

class TestPthreadTimedwait : public testing::Test {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();

    void SetUp() override
    {
        timeWait.Initialize();
    };

    void TearDown() override
    {
        if (caller.joinable()) {
            caller.join();
        }
    };

    void WakeUp(int delay)
    {
        if (delay != 0) {
            usleep(delay);
        }

        auto ret = timeWait.PthreadSignal();
        ASSERT_EQ(ret, TTP_OK);
    }

public:
    PthreadTimedwait timeWait;
    std::thread caller;
};

void TestPthreadTimedwait::SetUpTestCase() {}

void TestPthreadTimedwait::TearDownTestCase() {}

namespace {
TEST_F(TestPthreadTimedwait, clean_wait_awake_success)
{
    constexpr int delay = 500 * 1000;    // us
    constexpr int waitTime = 1;          // s

    // 1. clean
    timeWait.SignalClean();

    // 3. delay and awake
    std::thread thd(&TestPthreadTimedwait::WakeUp, this, delay);
    caller = std::move(thd);

    // 2. wait
    auto ret = timeWait.PthreadTimedwaitSecs(waitTime);
    ASSERT_EQ(ret, TTP_OK);
}

TEST_F(TestPthreadTimedwait, clean_wait_awake_timeout)
{
    constexpr int delay = 1200 * 1000;    // us
    constexpr int waitTime = 1;           // s

    // 1. clean
    timeWait.SignalClean();

    // 3. delay > wait
    std::thread thd(&TestPthreadTimedwait::WakeUp, this, delay);
    caller = std::move(thd);

    // 2. wait timeout
    auto ret = timeWait.PthreadTimedwaitSecs(waitTime);
    ASSERT_EQ(ret, ETIMEDOUT);
}

TEST_F(TestPthreadTimedwait, clean_awake_wait_success)
{
    constexpr int delay = 10000;    // us
    constexpr int waitTime = 1;     // s

    // 1. clean
    timeWait.SignalClean();

    // 2. awake
    std::thread thd(&TestPthreadTimedwait::WakeUp, this, 0);
    caller = std::move(thd);

    // 3. wait
    usleep(delay);
    auto ret = timeWait.PthreadTimedwaitSecs(waitTime);
    ASSERT_EQ(ret, TTP_OK);
}
}