/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */
#include <gtest/gtest.h>
#include <mockcpp/mockcpp.hpp>
#include <dlfcn.h>
#include "common.h"
#include "controller.h"
#include "processor.h"

using namespace ock::ttp;

#ifndef MOCKER_CPP
#define MOCKER_CPP(api, TT) MOCKCPP_NS::mockAPI(#api, reinterpret_cast<TT>(api))
#endif
#ifndef MOCKCPP_RESET
#define MOCKCPP_RESET GlobalMockObject::reset()
#endif

static void *handle;

constexpr uint32_t CHECK_COUNT_ONE = 1;
constexpr uint32_t CHECK_COUNT_TWO = 2;
constexpr uint32_t CHECK_COUNT_THREE = 3;
constexpr uint32_t CHECK_COUNT_FOUR = 4;

namesapce {
ProcessorPtr g_processor2_ = nullptr;
ProcessorPtr g_processor3_ = nullptr;
ProcessorPtr g_processor4_ = nullptr;

std::string g_backupIp_ = "127.0.0.1";
std::string g_backupPort_ = "1234";
std::string g_masterIp_ = "0.0.0.0";
uint32_t g_masterPort_ = 8555;
bool g_enableLocalCopy_ = false;
int64_t g_step_ = 2;
int64_t g_backupStep_ = 1;
AccTlsOption g_tlsOption_1;
bool g_lowUceFlag = false;
constexpr uint32_t UCE_NO_REBUILD = 2;

int (*SetUpdating)(int64_t);
int (*SetFinished)(int64_t);

class TestCAPI : public testing::Test {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();
    void SetUp() override;
    void TearDown() override;
public:
    int32_t CallBackFunc(void *ctx, int ctxSize)
    {
        ckptCount.fetch_add(1);
        return 0;
    }

    int StopFunc(void *ctx, int ctxSize)
    {
        stopCount.fetch_add(1);
        return 0;
    }

    int CleanFunc(void *ctx, int ctxSize)
    {
        cleanCount.fetch_add(1);
        if (g_lowUceFlag) {
            return UCE_NO_REBUILD; // UCE_NO_REBUILD
        }
        return 0;
    }

    int RenameFunc(void *ctx, int ctxSize)
    {
        renameCount.fetch_add(1);
        return 0;
    }

    int ExitFunc(void *ctx, int ctxSize)
    {
        exitCount.fetch_add(1);
        return 0;
    }

    void ExitFunc2()
    {
        exitCount.fetch_add(1);
    }

    void InitTestProcessor(ProcessorPtr& proc)
    {
        int32_t ret;
        proc = MakeRef<Processor>();
        ASSERT_TRUE(proc != nullptr);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_EXIT, std::bind(&TestCAPI::ExitFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_RENAME, std::bind(&TestCAPI::RenameFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_SAVE_CKPT, std::bind(&TestCAPI::CallBackFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_DEVICE_STOP, std::bind(&TestCAPI::StopFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_DEVICE_CLEAN, std::bind(&TestCAPI::CleanFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
    }

    void InitStaticProcessor()
    {
        int32_t ret;
        ProcessorPtr proc = Processor::GetInstance();
        ASSERT_TRUE(proc != nullptr);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_EXIT, std::bind(&TestCAPI::ExitFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_RENAME, std::bind(&TestCAPI::RenameFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_SAVE_CKPT, std::bind(&TestCAPI::CallBackFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_DEVICE_STOP, std::bind(&TestCAPI::StopFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_DEVICE_CLEAN, std::bind(&TestCAPI::CleanFunc,
            this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
    }

    void InitSource(int32_t replicaRankLen, bool enableARF, bool enableZIT);

public:
    std::atomic<uint32_t> ckptCount;
    std::atomic<uint32_t> renameCount;
    std::atomic<uint32_t> exitCount;
    std::atomic<uint32_t> stopCount;
    std::atomic<uint32_t> cleanCount;
};

void TestCAPI::InitSource(int32_t replicaRankLen = 2, bool enableARF = false, bool enableZIT = false)
{
    ControllerPtr ctrl = Controller::GetInstance();
    if (ctrl == nullptr) {
        printf("ControllerPtr is NULL!\n");
    }

    ckptCount.store(0);
    renameCount.store(0);
    exitCount.store(0);
    std::vector<int32_t> replicaCnt = { replicaRankLen };
    std::vector<int32_t> replicaOffset = { 0 };

    ctrl->Initialize(0, 4, g_enableLocalCopy_);

    std::string ip = g_masterIp_;
    int32_t port = g_masterPort_;
    int32_t ret = ctrl->Start(ip, port, g_tlsOption_1);
    ASSERT_EQ(ret, 0);

    ProcessorPtr proc = Processor::GetInstance();
    if (proc == nullptr) {
        printf("ProcessorPtr is NULL!\n");
    }

    std::vector<int32_t> ranks = {0, 1, 2, 3};
    std::vector<std::vector<int32_t>> groups = { ranks };
    InitStaticProcessor();
    TestCAPI::InitTestProcessor(g_processor2_);
    TestCAPI::InitTestProcessor(g_processor3_);
    TestCAPI::InitTestProcessor(g_processor4_);

    ret = proc->Initialize(0, 4, g_enableLocalCopy_, g_tlsOption_1);
    ASSERT_EQ(ret, 0);
    ret = g_processor2_->Initialize(1, 4, g_enableLocalCopy_, g_tlsOption_1);
    ASSERT_EQ(ret, 0);
    ret = g_processor3_->Initialize(2, 4, g_enableLocalCopy_, g_tlsOption_1);
    ASSERT_EQ(ret, 0);
    ret = g_processor4_->Initialize(3, 4, g_enableLocalCopy_, g_tlsOption_1);
    ASSERT_EQ(ret, 0);

    ret = proc->Start(ip, port);
    ASSERT_EQ(ret, 0);
    ret = g_processor2_->Start(ip, port);
    ASSERT_EQ(ret, 0);
    ret = g_processor3_->Start(ip, port);
    ASSERT_EQ(ret, 0);
    ret = g_processor4_->Start(ip, port);
    ASSERT_EQ(ret, 0);

    ret = proc->ReportReplicaInfo(groups, replicaCnt, replicaOffset);
    ASSERT_EQ(ret, 0);
    ret = g_processor2_->ReportReplicaInfo(groups, replicaCnt, replicaOffset);
    ASSERT_EQ(ret, 0);
    ret = g_processor3_->ReportReplicaInfo(groups, replicaCnt, replicaOffset);
    ASSERT_EQ(ret, 0);
    ret = g_processor4_->ReportReplicaInfo(groups, replicaCnt, replicaOffset);
    ASSERT_EQ(ret, 0);
}
void TestCAPI::SetUpTestCase() {}

void TestCAPI::TearDownTestCase() {}

void TestCAPI::SetUp()
{
}

void TestCAPI::TearDown()
{
    constexpr int delay = 1; // s
    ProcessorPtr proc = Processor::GetInstance();
    proc->Destroy(true);
    g_processor2_->Destroy(true);
    g_processor3_->Destroy(true);
    g_processor4_->Destroy(true);

    ControllerPtr ctrl = Controller::GetInstance();
    ctrl->Destroy();
    sleep(delay);
    dlclose(handle);
    GlobalMockObject::verify();
}

TEST_F(TestCAPI, dump_success)
{
    constexpr int delay = 1; // s
    OutLogger::Instance()->SetLogLevel(DEBUG_LEVEL);
    handle = dlopen("../../../output/lib/libttp_c_api.so", RTLD_LAZY);
    if (handle == nullptr) {
        printf("Open Error:%s.\n", dlerror());
    }

    TestCAPI::InitSource();
    int32_t ret;

    auto updating = reinterpret_cast<void **>(&SetUpdating);
    *updating = dlsym(handle, "MindioTtpSetOptimStatusUpdating");
    if (SetUpdating == nullptr) {
        printf("Open Error:%s.\n", dlerror());
    }
    ret = SetUpdating(g_backupStep_);
    ASSERT_EQ(ret, 0);
    auto finished = reinterpret_cast<void **>(&SetFinished);
    *finished = dlsym(handle, "MindioTtpSetOptimStatusFinished");
    if (SetFinished == nullptr) {
        printf("Open Error:%s.\n", dlerror());
    }
    ret = SetFinished(g_step_);
    ASSERT_EQ(ret, 0);

    ret = g_processor2_->BeginUpdating(g_backupStep_);
    ASSERT_EQ(ret, 0);
    ret = g_processor2_->FinishedUpdate(g_step_);
    ASSERT_EQ(ret, 0);

    ret = g_processor3_->BeginUpdating(g_backupStep_);
    ASSERT_EQ(ret, 0);
    ret = g_processor3_->FinishedUpdate(g_step_);
    ASSERT_EQ(ret, 0);

    ret = g_processor4_->BeginUpdating(g_backupStep_);
    ASSERT_EQ(ret, 0);
    ret = g_processor4_->FinishedUpdate(g_step_);
    ASSERT_EQ(ret, 0);
    Processor::GetInstance()->ReportStatus(ReportState::RS_UNKNOWN);
    sleep(delay);
    g_processor2_->SetDumpResult(0);
    g_processor3_->SetDumpResult(0);
    ret = g_processor2_->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = g_processor3_->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = g_processor4_->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);

    ASSERT_EQ(ckptCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(renameCount.load(), CHECK_COUNT_ONE);
    ASSERT_EQ(exitCount.load(), CHECK_COUNT_FOUR);
}
}