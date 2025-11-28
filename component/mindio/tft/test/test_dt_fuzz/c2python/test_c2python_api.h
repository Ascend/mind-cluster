/*
* Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
*/

#ifndef OCK_TTP_TEST_C2PYTHON_H
#define OCK_TTP_TEST_C2PYTHON_H
#include "gtest/gtest.h"

using namespace ock::ttp;

class C2PythonDtFuzz : public testing::Test {
public:
    static void SetUpTestSuite();
    static void TearDownTestSuite();

public:
    int32_t CallBackFunc(void *ctx, int ctxSize)
    {
        return 0;
    }

    int ExitFunc(void *ctx, int ctxSize)
    {
        return 0;
    }

    int RenameFunc(void *ctx, int ctxSize)
    {
        return 0;
    }

    int Register(void *ctx, int ctxSize)
    {
        return 0;
    }

    void InitProcessor(ProcessorPtr& proc)
    {
        proc = MakeRef<Processor>();
        proc->RegisterEventHandler(PROCESSOR_EVENT_EXIT, std::bind(&C2PythonDtFuzz::ExitFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        proc->RegisterEventHandler(PROCESSOR_EVENT_SAVE_CKPT, std::bind(&C2PythonDtFuzz::CallBackFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        proc->RegisterEventHandler(PROCESSOR_EVENT_RENAME, std::bind(&C2PythonDtFuzz::RenameFunc,
            this, std::placeholders::_1, std::placeholders::_2));
    }

    void InitController(ControllerPtr& ctrl)
    {
        ctrl = MakeRef<Controller>();
        MindXEnginePtr engine = MindXEngine::GetInstance();
        engine->RegisterEventHandler(MindXEvent::MINDX_EVENT_REGISTER, std::bind(&C2PythonDtFuzz::Register,
            this, std::placeholders::_1, std::placeholders::_2));
    }

    void Init(bool enableARF, bool enableZIT);

    void Destroy();
};
#endif // OCK_TTP_TEST_C2PYTHON_H