/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */
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

class ControllerBaseTest : public ControllerTest {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();
};

void ControllerBaseTest::SetUpTestCase() {}

void ControllerBaseTest::TearDownTestCase()
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

TEST_F(ControllerBaseTest, invalid_param)
{
    ASSERT_EQ(setenv("TTP_NORMAL_ACTION_TIME_LIMIT", "\0", 1), 0);
    ControllerBaseTest::InitSource();
    int32_t ret = controller1->Initialize(0, WORLD_SIZE, enableLocalCopy); // 重复初始化
    ASSERT_EQ(ret, TTP_OK);
    std::string ip = "0.0.0.0";
    uint32_t port = CONTROLLER_PORT; // 正常端口
    uint32_t invalidPort = 70000; // 非法端口

    ret = controller1->Start(ip, port, testTlsOption); // 重复启动
    ASSERT_EQ(ret, TTP_OK);

    ret = controller1->Start(ip, invalidPort, testTlsOption); // 端口非法
    ASSERT_EQ(ret, TTP_ERROR);

    ret = Controller::GetInstance()->Initialize(-2, invalidPort, enableLocalCopy); // rank非法 -2
    ASSERT_EQ(ret, TTP_ERROR);

    ret = Controller::GetInstance()->Initialize(0, -1, enableLocalCopy); // 总卡数（worldSize）非法
    ASSERT_EQ(ret, TTP_ERROR);

    ret = Processor::GetInstance()->Initialize(0, -1, enableLocalCopy, testTlsOption); // 非法参数
    ASSERT_EQ(ret, TTP_ERROR);

    testTlsOption.enableTls = true;
    ProcessorPtr procLocal = nullptr;
    ControllerBaseTest::InitProcessor(procLocal);
    ret = procLocal->Initialize(0, WORLD_SIZE, enableLocalCopy, testTlsOption);
    ASSERT_EQ(ret, TTP_ERROR);
    procLocal->Destroy(true);
    testTlsOption.enableTls = false;
    unsetenv("TTP_NORMAL_ACTION_TIME_LIMIT");
}

TEST_F(ControllerBaseTest, replica_param_error)
{
    ControllerBaseTest::InitSource();
    std::vector<int32_t> testReplicaCnt = { 2 };
    std::vector<int32_t> testReplicaOffset = { 0 };
    std::vector<int32_t> ranks0 = {0, 2};
    std::vector<int32_t> ranks1 = {1, 3};
    int32_t ret;
    std::vector<std::vector<int32_t>> errorGroups = { ranks0, ranks1 };

    ret = processor1->Initialize(0, WORLD_SIZE, enableLocalCopy, testTlsOption); // 正常重复初始化
    ASSERT_EQ(ret, TTP_ERROR);

    std::string ip = "0.0.0.0";
    uint32_t port = CONTROLLER_PORT; // 正常端口
    uint32_t invalidPort = 70000; // 非法端口

    ret = processor1->Start(ip, port); // 正常重复启动
    ASSERT_EQ(ret, TTP_ERROR);

    ret = processor1->ReportReplicaInfo(errorGroups, testReplicaCnt, testReplicaOffset); // Groups错误
    ASSERT_EQ(ret, TTP_ERROR);

    std::vector<int32_t> errorReplicaCnt = { 0 };
    std::vector<std::vector<int32_t>> correctGroups = { ranks0 };
    ret = processor2->ReportReplicaInfo(correctGroups, errorReplicaCnt, testReplicaOffset); // replicaCnt错误
    ASSERT_EQ(ret, TTP_ERROR);

    processor3->mindSpore_=true;
    ret = processor3->ReportReplicaInfo(correctGroups, errorReplicaCnt, testReplicaOffset);
    // replicaCnt错误且mindSpore_为true
    ASSERT_EQ(ret, TTP_ERROR);
}

// BroadcastCrtlIps check
std::string g_broadcastIp;
uint32_t g_broadcastMask, g_broadcastType;
int32_t SendMsgStub(Controller *control, int16_t msgType, const AccDataBufferPtr &d, std::vector<int32_t> &targetRanks)
{
    g_broadcastType = msgType;
    g_broadcastMask = 0;
    for (int i = 0; i < targetRanks.size(); i++) {
        g_broadcastMask |= 1 << (targetRanks[i]);
    }

    BroadcastIpMsg *msg = reinterpret_cast<BroadcastIpMsg *>(d->DataPtr());
    std::string ipStr(msg->arr, msg->ipLen - 1);
    g_broadcastIp = ipStr;

    return TTP_ERROR;
}

TEST_F(ControllerBaseTest, start_before_init)
{
    ControllerBaseTest::InitSource();
    controller1->Destroy();
    std::string ip = "0.0.0.0";
    int32_t ret = controller1->Start(ip, 8555, testTlsOption);
    ASSERT_EQ(ret, TTP_ERROR);
}

TEST_F(ControllerBaseTest, acc_tcp_server_fail)
{
    ControllerBaseTest::InitSource();

    controller2 = MakeRef<Controller>();
    ASSERT_TRUE(controller2 != nullptr);
    controller2->Initialize(0, WORLD_SIZE, enableLocalCopy);
    int32_t ret;
    MOCKER_CPP(&AccTcpServerDefault::Start,
               int32_t(*)(const AccTcpServerOptions &opt, const AccTlsOption &tlsOption)).stubs().will(returnValue(1));
    std::string ip = "0.0.0.0";
    int32_t port = 8555;
    ret = controller2->Start(ip, port, testTlsOption);
    ASSERT_EQ(ret, TTP_ERROR);

    MOCKCPP_RESET;
}

TEST_F(ControllerBaseTest, acc_tcp_client_offline)
{
    ControllerBaseTest::InitSource();
    MOCKER_CPP(&AccTcpClientDefault::Connect, int32_t(*)(const AccConnReq &connReq)).stubs().will(returnValue(1));
    ProcessorPtr g_processor5 = nullptr;
    ControllerBaseTest::InitProcessor(g_processor5);
    int32_t ret = g_processor5->Initialize(WORLD_SIZE, 5, enableLocalCopy, testTlsOption);
    ASSERT_EQ(ret, 0);
    std::string ip = "0.0.0.0";
    int32_t port = 8555;
    ret = g_processor5->Start(ip, port);
    ASSERT_EQ(ret, TTP_ERROR);
    g_processor5->Destroy(true);
    MOCKCPP_RESET;
}

TEST_F(ControllerBaseTest, copy_test)
{
    ControllerBaseTest::InitSource(REPLICA_NUM_TWO);
    int32_t ret = 0;
    ret = processor1->BeginCopying();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->BeginCopying();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->BeginCopying();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor4->BeginCopying();
    ASSERT_EQ(ret, TTP_OK);
}

TEST_F(ControllerBaseTest, engine_init_fail)
{
    controller2 = MakeRef<Controller>();
    MOCKER_CPP(&ActionEngine::Initialize, int32_t(*)(AccTcpServerPtr, const ActionMsgSend &))
        .stubs().will(returnValue(1));
    int32_t ret = controller2->Initialize(1, WORLD_SIZE, enableLocalCopy);
    ASSERT_EQ(ret, 1);
    MOCKCPP_RESET;
}

TEST_F(ControllerBaseTest, machine_init_fail)
{
    controller2 = MakeRef<Controller>();
    MOCKER_CPP(&controllerStateMachine::Initialize, int32_t(*)(uint32_t))
        .stubs().will(returnValue(1));
    int32_t ret = controller2->Initialize(1, WORLD_SIZE, enableLocalCopy);
    ASSERT_EQ(ret, 1);
    MOCKCPP_RESET;
}

int32_t SendMsgStub4Retry(Controller *control, int16_t msgType, const AccDataBufferPtr &d,
                          std::vector<int32_t> &targetRanks, const std::vector<AccDataBufferPtr> &cbCtx)
{
    g_broadIpCount.fetch_add(1);
    return TTP_ERROR;
}

TEST_F(ControllerBaseTest, GetTcpStoreUrlAndTransforHostNameToIp)
{
    setenv("MASTER_PORT", PORT, 1);
    setenv("MASTER_ADDR", ADDRESS, 1);
    Controller controller;
    std::set<std::string> ipList;

    // Test case 1: Normal case
    ASSERT_EQ(controller.GetTcpStoreUrl(ipList, g_portStore), TTP_OK);
    ASSERT_EQ(ipList.size(), 1);
    ASSERT_TRUE(ipList.find(ADDRESS) != ipList.end());
    ASSERT_TRUE(ipList.find(g_rankIp) == ipList.end());

    // Clean up for next test case
    ipList.clear();

    // Test case 2: MASTER_ADDR not set
    unsetenv("MASTER_ADDR");
    ASSERT_EQ(controller.GetTcpStoreUrl(ipList, g_portStore), TTP_ERROR);
    ASSERT_EQ(ipList.size(), 0);
    ASSERT_TRUE(ipList.find(g_rankIp) == ipList.end());

    // Clean up for next test case
    ipList.clear();

    // Test case 3: Invalid MASTER_PORT
    setenv("MASTER_PORT", INVALID_PORT, 1); // Port number out of range
    auto ret = controller.GetTcpStoreUrl(ipList, g_portStore);
    ASSERT_EQ(ret, TTP_ERROR);
    ASSERT_EQ(ipList.size(), 0);
    auto ans = ipList.find(ADDRESS);
    ASSERT_TRUE(ans == ipList.end());
    ans = ipList.find(g_rankIp);
    ASSERT_TRUE(ans == ipList.end());

    // Clean up for next test case
    ipList.clear();

    // Test case 4: Invalid hostname in TRANSFORM_HOST_NAME_TO_IP
    std::string invalidIp;
    ret = controller.TransforHostNameToIp("invalid_hostname", invalidIp);
    ASSERT_EQ(ret, TTP_ERROR);
    ASSERT_TRUE(invalidIp.empty());
    unsetenv("MASTER_PORT");
    unsetenv("MASTER_ADDR");
}

TEST_F(ControllerBaseTest, process_result_and_hb_reply)
{
    uint16_t sn = 1; // wrong sn
    MsgOpCode op = MsgOpCode::TTP_MSG_OP_HEARTBEAT_SEND;

    ControllerBaseTest::InitSource();

    int32_t ret = processor1->ProcessResultAndHBReply(sn, op);
    EXPECT_EQ(ret, TTP_ERROR);
    sn = 0; // right sn
    ret = processor1->ProcessResultAndHBReply(sn, op);
    EXPECT_EQ(ret, TTP_OK);
}

TEST_F(ControllerBaseTest, launch_tcp_store_client)
{
    setenv("MASTER_PORT", "6000", 1);
    setenv("MASTER_ADDR", "0.0.0.0", 1);

    ControllerBaseTest::InitSource();

    int32_t ret = processor1->LaunchTcpStoreClient();

    processor2->controllerIps_ = {"0.0.0.0", "127.0.0.1"};
    processor2->controllerIdx_ = 1;
    ret = processor2->LaunchTcpStoreClient();

    processor3->controllerIps_ = {"0.0.0.0", "0.0.0.1:6000", "127.0.0.1"};
    processor3->controllerIdx_ = 1;
    ret = processor3->LaunchTcpStoreClient();
    EXPECT_EQ(ret, TTP_ERROR);
    unsetenv("MASTER_PORT");
    unsetenv("MASTER_ADDR");
}

TEST_F(ControllerBaseTest, restart_processor)
{
    setenv("TTP_RETRY_TIMES", "1", 1);

    ControllerBaseTest::InitSource();

    int32_t ret = processor1->ReStart(); // ip用尽
    ASSERT_EQ(ret, TTP_ERROR);

    processor2->controllerIps_ = {"0.0.0.0", "127.0.0.1"}; // 错误IP
    ret = processor2->ReStart();
    ASSERT_EQ(ret, TTP_OK);

    processor3->controllerIps_ = {"1234:127.0.0.1", "8555:0.0.0.0"}; // 正确IP
    processor3->controllerIdx_ = 0; // controllerIdx_复位
    ret = processor3->ReStart(); // 目前允许重复Processor注册，该用例会成功
    ASSERT_EQ(ret, TTP_OK);

    processor4->controllerIps_ = {"1234:127.0.0.1", "8555:125.4.6.110"};
    processor4->controllerIdx_ = 0; // controllerIdx_复位
    ret = processor4->ReStart(); // 连接超时
    ASSERT_EQ(ret, TTP_ERROR);

    unsetenv("TTP_RETRY_TIMES");
}

TEST_F(ControllerBaseTest, handle_function_testing)
{
    setenv("MASTER_PORT", "6000", 1);
    setenv("MASTER_ADDR", "0.0.0.0", 1);

    ControllerBaseTest::InitSource();

    uint8_t data[3] = {0xAA, 0xBB, 0xFF};
    uint32_t len = 2;

    processor1->HandleLaunchTcpStoreServer(data, len); // PS_NORMAL

    processor1->processorStatus_.store(PS_PAUSE);
    processor1->HandleLaunchTcpStoreServer(data, len); // PS_PAUSE且IP正确
    processor1->controllerIps_ = {"0.0.0.0", "127.0.0.1"};
    processor1->HandleLaunchTcpStoreServer(data, len); // PS_PAUSE但是IP错误
    ASSERT_EQ(processor1->processorStatus_, PS_PAUSE);
    unsetenv("MASTER_PORT");
    unsetenv("MASTER_ADDR");
}

TEST_F(ControllerBaseTest, serial_num_testing)
{
    setenv("MASTER_PORT", "6000", 1);
    setenv("MASTER_ADDR", "0.0.0.0", 1);

    ControllerBaseTest::InitSource();
    uint16_t testSn = 9;
    uint16_t outdatedSn = 10;
    MsgOpCode testOp = MsgOpCode::TTP_MSG_OP_CKPT_SEND;

    processor1->actionSn_.store(outdatedSn);
    int32_t ret = processor1->CheckMsgSnAndReply(testSn, testOp); // outdated actionSn
    ASSERT_EQ(ret, TTP_ERROR);

    processor1->actionSn_.store(testSn);
    processor1->replyMsgBackup_ = {TTP_OK, 9, 0};
    ret = processor1->CheckMsgSnAndReply(testSn, testOp); // repeated actionSn
    ASSERT_EQ(ret, TTP_ERROR);

    unsetenv("MASTER_PORT");
    unsetenv("MASTER_ADDR");
}

TEST_F(ControllerBaseTest, check_ip_port_ok)
{
    std::string ip = CONTROLLER_IP;
    uint16_t port = 8555;
    setenv("MASTER_PORT", "8555", 1);
    setenv("MASTER_ADDR", "0.0.0.0", 1);

    ControllerBaseTest::InitSource();
    bool ret = controller1->CheckIpPortAccessible(ip, port);
    ASSERT_EQ(ret, true);
    unsetenv("MASTER_PORT");
    unsetenv("MASTER_ADDR");
}

TEST_F(ControllerBaseTest, check_ip_port_error_ip)
{
    uint16_t port = 8555;
    std::string errorIp = "invalid_ip";
    setenv("MASTER_PORT", "8555", 1);
    setenv("MASTER_ADDR", "0.0.0.0", 1);

    ControllerBaseTest::InitSource();
    bool ret = controller1->CheckIpPortAccessible(errorIp, port);
    ASSERT_EQ(ret, false);
    unsetenv("MASTER_PORT");
    unsetenv("MASTER_ADDR");
}

TEST_F(ControllerBaseTest, check_ip_port_error_port)
{
    std::string ip = CONTROLLER_IP;
    uint16_t errorPort = 0;
    setenv("MASTER_PORT", "8555", 1);
    setenv("MASTER_ADDR", "0.0.0.0", 1);

    ControllerBaseTest::InitSource();
    bool ret = controller1->CheckIpPortAccessible(ip, errorPort);
    ASSERT_EQ(ret, false);
    unsetenv("MASTER_PORT");
    unsetenv("MASTER_ADDR");
}

TEST_F(ControllerBaseTest, choose_rank_all_failed)
{
    ControllerBaseTest::InitSource();
    const RankChooseInfo testRankChooseInfo = {COMMON_STEP, {0, 1, 2, 3}, {0, 1, 2, 3}};
    std::vector<int32_t> testRankVec;
    int32_t ret = controller1->ChooseRank(testRankChooseInfo, testRankVec, 2);
    ASSERT_EQ(ret, TTP_ERROR);
}

TEST_F(ControllerBaseTest, misc_test) // 杂项测试，未来需要进一步分类
{
    ControllerBaseTest::InitSource();
    processor1->processorStatus_.store(PS_DUMP); // PS_DUMP
    int32_t ret = processor1->BeginCopying();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor1->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, TTP_ERROR);

    processor1->controllerIps_ = {"1:0.0.0.0"};
    uint8_t data[100];
    processor1->StartBackupController(data);

    ret = controller1->MindXNotifyExit(nullptr, 0);
    ASSERT_EQ(ret, TTP_OK);

    controller1->repairType_ = ControllerRepairType::CRT_BUTT;
    ret = controller1->RepairCallback();
    ASSERT_EQ(ret, TTP_ERROR);

    controller1->isMasterCtrl_.store(false);
    bool boolRet = controller1->IsBackupToMaster();
    ASSERT_EQ(boolRet, true);
}

TEST_F(ControllerBaseTest, multi_replica_overflow_fail)
{
    MOCKER_CPP(&Controller::ChooseRank, TResult(*)(const RankChooseInfo, std::vector<int32_t>,
        uint32_t)).stubs().will(returnValue(TResult::TTP_ERROR));
    ControllerBaseTest::InitSource(WORLD_SIZE);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    HeartbeatUpdate();
    ReportState state = ReportState::RS_UNKNOWN;
    processor2->ReportStatus(state);
    processor1->SetDumpResult(1);
    processor3->SetDumpResult(1);
    processor4->SetDumpResult(1);
    ret = processor1->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor3->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor4->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);

    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);
    ASSERT_EQ(exitCount.load(), WORLD_SIZE);

    MOCKCPP_RESET;
}

TEST_F(ControllerBaseTest, overflow_multiply)
{
    uint32_t a = INT32_MAX;
    uint32_t b = 2; // 2
    bool ret = IsOverflow(a, b);
    ASSERT_EQ(ret, false);
}

}