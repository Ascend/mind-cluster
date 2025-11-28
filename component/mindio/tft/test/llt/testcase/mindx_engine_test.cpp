/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */
#include <gtest/gtest.h>
#include <mockcpp/mockcpp.hpp>
#include "common.h"
#include "mindx_engine.h"
namespace {
using namespace ock::ttp;

#ifndef MOCKER_CPP
#define MOCKER_CPP(api, TT) MOCKCPP_NS::mockAPI(#api, reinterpret_cast<TT>(api))
#endif
#ifndef MOCKCPP_RESET
#define MOCKCPP_RESET GlobalMockObject::reset()
#endif
MindXEnginePtr g_mindxEngine1 = nullptr;


class TestMindxEngine : public testing::Test {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();
    void SetUp() override;
    void TearDown() override;

public:
    void InitSource();

    int Register2MindX(void *ctx, int ctxSize)
    {
        registerCount.fetch_add(1);
        return 0;
    }

    void InitMindxEngine(MindXEnginePtr& ctrl)
    {
        int32_t ret;
        ctrl = MakeRef<MindXEngine>();
        ASSERT_TRUE(ctrl != nullptr);
        ctrl = MindXEngine::GetInstance();
        ret = ctrl->RegisterEventHandler(MindXEvent::MINDX_EVENT_REGISTER,
                                         std::bind(&TestMindxEngine::Register2MindX, this,
                                                   std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
    }

public:

    std::atomic<uint32_t> registerCount = { 0 };
};

void TestMindxEngine::InitSource()
{
    registerCount.store(0);
    TestMindxEngine::InitMindxEngine(g_mindxEngine1);
}


TEST_F(TestMindxEngine, register_faild)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&MindXEngine::EventProcess, TResult(*)(void)).expects(once()).will(returnValue(TResult::TTP_ERROR));
    TestMindxEngine::InitSource();
    TResult ret = g_mindxEngine1->EventProcess(MindXEvent::MINDX_EVENT_REGISTER, nullptr, 0);

    g_mindxEngine1->Register2MindX();
    ASSERT_EQ(ret, TTP_ERROR);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

void TestMindxEngine::SetUpTestCase() {}

void TestMindxEngine::TearDownTestCase()
{
    MindXEnginePtr ctrl = MindXEngine::GetInstance();
    ctrl->Destroy();
}

void TestMindxEngine::SetUp() {}

void TestMindxEngine::TearDown()
{
    g_mindxEngine1->Destroy();
}
}