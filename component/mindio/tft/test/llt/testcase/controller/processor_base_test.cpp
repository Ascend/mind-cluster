/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */
#include <cstdlib>
#include "common.h"
#include "mindx_engine.h"
#include "acc_tcp_server_default.h"
#include "acc_tcp_client_default.h"
#include "controller_test.h"
namespace {
using namespace ock::ttp;
#ifndef MOCKER_CPP
#define MOCKER_CPP(api, TT) MOCKCPP_NS::mockAPI(#api, reinterpret_cast<TT>(api))
#endif
#ifndef MOCKCPP_RESET
#define MOCKCPP_RESET GlobalMockObject::reset()
#endif

uint32_t g_logNum = 0;
const char* PORT = "6000";
uint16_t g_portStore = 6000;
const char* ADDRESS = "0.0.0.0";
std::string g_rankIp = "rankIp";
const char* INVALID_PORT = "70000";
std::atomic<uint32_t> g_broadIpCount { 0 };

const std::string BACKUP_IP = "127.0.0.1";
const std::string BACKUP_PORT = "1234";
const std::string CONTROLLER_IP = "0.0.0.0";
constexpr uint32_t CONTROLLER_PORT = 8555;
constexpr int64_t BACKUP_STEP = 1;

class ProcessorBaseTest : public ControllerTest {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();
};

void ProcessorBaseTest::SetUpTestCase() {}

void ProcessorBaseTest::TearDownTestCase()
{
    ControllerPtr ctrl = Controller::GetInstance();
    std::vector<int32_t> replicaCnt = { 2 };
    std::vector<int32_t> replicaOffset = { 0 };
    ctrl->Initialize(0, WORLD_SIZE, false, false, false);
    mkdir("logs", 0750); // 日志文件夹权限为0750
    std::ofstream file("logs/ttp_log.log");
    file << "This is a test file." << std::endl;
    file.close();
    ctrl->Destroy();
}

TEST_F(ProcessorBaseTest, processor_GetInstance)
{
    ProcessorBaseTest::InitSource();
    ProcessorPtr proc1 = Processor::GetInstance(false);
    ProcessorPtr proc2 = Processor::GetInstance(true);
    ASSERT_EQ(proc2, nullptr);
}

TEST_F(ProcessorBaseTest, processor_ProcessorIsRunning)
{
    ProcessorBaseTest::InitSource();
    uint32_t oldStatus = processor1->processorStatus_.load();

    processor1->processorStatus_.store(PS_DUMP);
    int32_t ret = processor1->ProcessorIsRunning();
    ASSERT_EQ(ret, true);

    processor1->processorStatus_.store(PS_END);
    ret = processor1->ProcessorIsRunning();
    ASSERT_EQ(ret, false);

    processor1->processorStatus_.store(oldStatus);
}

TEST_F(ProcessorBaseTest, processor_Initialize)
{
    ProcessorBaseTest::InitSource();
    processor1->localCopySwitch_ = true;
    processor1->trainStatus_.data_status = Updated;
    processor1->trainStatus_.step = 1;
    auto range = processor1->GetNowStep();
    int64_t nowStep = range.second;
    ASSERT_EQ(nowStep, 1);
    
    processor1->trainStatus_.data_status = Copying;
    processor1->trainStatus_.step = 2;
    range = processor1->GetNowStep();
    nowStep = range.second;
    ASSERT_EQ(nowStep, 2);
    
    processor1->trainStatus_.data_status = Updating;
    processor1->trainStatus_.backup_step = 3;
    range = processor1->GetNowStep();
    nowStep = range.second;
    ASSERT_EQ(nowStep, 3);
}

TEST_F(ProcessorBaseTest, processor_BeginUpdating)
{
    ProcessorBaseTest::InitSource();
    processor1->isPrelockOk_.store(true);
    processor1->mindSpore_ = true;
    int32_t ret = processor1->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, TTP_OK);

    try {
        processor1->isPrelockOk_.store(true);
        processor1->mindSpore_ = false;
        ret = processor1->BeginUpdating(BACKUP_STEP);
        FAIL() << "Expected exception not thrown";
    } catch (const std::runtime_error& e) {
        EXPECT_STREQ(e.what(), "FORCE STOP. By mindio-ttp BeginUpdating");
    } catch (...) {
        FAIL() << "Unexpected exception type";
    }
}

TEST_F(ProcessorBaseTest, processor_ReportDpInfo)
{
    std::vector<int32_t> dpRankList = {0, 1};
    ProcessorBaseTest::InitSource();
    processor1->zitSwitch_ = false;
    int32_t ret = processor1->ReportDpInfo(dpRankList);
    ASSERT_EQ(ret, TTP_OK);
}

TEST_F(ProcessorBaseTest, processor_start_test)
{
    std::string ip = CONTROLLER_IP;
    int32_t port = CONTROLLER_PORT;
    std::string localIp = BACKUP_IP;
    ProcessorBaseTest::InitSource();
    int32_t ret = processor1->Start(ip, port, localIp);
    ASSERT_EQ(ret, TTP_ERROR);
    
    ret = processor1->Start(ip, port, "123");
    ASSERT_EQ(ret, TTP_ERROR);
}

TEST_F(ProcessorBaseTest, processor_Destroy)
{
    ProcessorBaseTest::InitSource();
    processor1->startBackup_ = processor1->controllerIdx_ + 1;
    processor1->Destroy();
}

TEST_F(ProcessorBaseTest, processor_Rename)
{
    ProcessorBaseTest::InitSource();
    processor1->rank_ = 1;
    CommonMsg msg;
    msg.rank = 2;
    uint8_t* data = reinterpret_cast<uint8_t*>(&msg);
    int32_t ret = processor1->Rename(data);
    ASSERT_EQ(ret, TTP_ERROR);
}

TEST_F(ProcessorBaseTest, processor_HandleLaunchTcpStoreServer)
{
    setenv("MASTER_PORT", "6000", 1);
    setenv("MASTER_ADDR", "0.0.0.0", 1);

    ProcessorBaseTest::InitSource();

    uint8_t data[3] = {0xAA, 0xBB, 0xFF};
    uint32_t len = 2;

    processor1->processorStatus_.store(PS_PAUSE);
    processor1->HandleLaunchTcpStoreServer(data, len); // PS_PAUSE且IP正确
    ASSERT_EQ(processor1->processorStatus_, PS_PAUSE);
    unsetenv("MASTER_PORT");
    unsetenv("MASTER_ADDR");
}

TEST_F(ProcessorBaseTest, processor_ReportDp2Controller)
{
    std::vector<int32_t> dpRankList = {0, 1};
    ProcessorBaseTest::InitSource();
    processor1->dpList_ = dpRankList;
    processor1->rank_ = 1;
    int32_t ret = processor1->ReportDp2Controller();
    ASSERT_EQ(ret, TTP_OK);
}

TEST_F(ProcessorBaseTest, processor_ReportStatus)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ProcessorBaseTest::ReportStrategies, int(*)(void *ctx, int ctxSize)).
        stubs().will(returnValue(400)); // invalid return 400

    ProcessorBaseTest::InitSource();
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ret = processor4->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, 0);

    ReportState state = ReportState::RS_STEP_FINISH;
    ret = processor4->ReportStatus(state);
    sleep(1);
    ASSERT_EQ(ret, TTP_OK);

    state = ReportState::RS_NORMAL;
    ret = processor4->ReportStatus(state);
    ASSERT_EQ(ret, TTP_OK);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

}