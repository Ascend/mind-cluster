/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */
#include "controller_test.h"
namespace {
using namespace ock::ttp;
#ifndef MOCKER_CPP
#define MOCKER_CPP(api, TT) MOCKCPP_NS::mockAPI(#api, reinterpret_cast<TT>(api))
#endif
#ifndef MOCKCPP_RESET
#define MOCKCPP_RESET GlobalMockObject::reset()
#endif

class ControllerMSTest : public ControllerTest {
public:
    void MsInitSource(int32_t replicaRankLen = 2, bool enableARF = false, bool enableZIT = false);
};

void ControllerMSTest::MsInitSource(int32_t replicaRankLen, bool enableARF, bool enableZIT)
{
    ControllerTest::CountClean();

    ControllerMSTest::InitController(controller1);
    std::vector<int32_t> replicaCnt = { replicaRankLen };
    std::vector<int32_t> replicaOffset = { 2 };
    int32_t ret = controller1->Initialize(0, WORLD_SIZE, enableLocalCopy, enableARF, enableZIT);
    ASSERT_EQ(ret, 0);

    std::string ip = CONTROLLER_IP;
    int32_t port = CONTROLLER_PORT;
    ret = controller1->Start(ip, port, testTlsOption);
    ASSERT_EQ(ret, 0);

    std::vector<int32_t> ranks0 = {0, 2};
    std::vector<int32_t> ranks1 = {1, 3};
    std::vector<std::vector<int32_t>> groups0 = { ranks0 };
    std::vector<std::vector<int32_t>> groups1 = { ranks1 };
    ControllerMSTest::InitProcessor(processor1);
    ControllerMSTest::InitProcessor(processor2);
    ControllerMSTest::InitProcessor(processor3);
    ControllerMSTest::InitProcessor(processor4);

    ret = processor1->Initialize(0, WORLD_SIZE, enableLocalCopy, testTlsOption, true, enableARF);
    ASSERT_EQ(ret, 0);
    ret = processor2->Initialize(1, WORLD_SIZE, enableLocalCopy, testTlsOption, true, enableARF);
    ASSERT_EQ(ret, 0);
    ret = processor3->Initialize(2, WORLD_SIZE, enableLocalCopy, testTlsOption, true, enableARF);
    ASSERT_EQ(ret, 0);
    ret = processor4->Initialize(3, WORLD_SIZE, enableLocalCopy, testTlsOption, true, enableARF);
    ASSERT_EQ(ret, 0);

    ret = processor1->Start(ip, port);
    ASSERT_EQ(ret, 0);
    ret = processor2->Start(ip, port);
    ASSERT_EQ(ret, 0);
    ret = processor3->Start(ip, port);
    ASSERT_EQ(ret, 0);
    ret = processor4->Start(ip, port);
    ASSERT_EQ(ret, 0);

    ret = processor1->ReportReplicaInfo(groups0, replicaCnt, replicaOffset);
    ASSERT_EQ(ret, 0);
    ret = processor2->ReportReplicaInfo(groups1, replicaCnt, replicaOffset);
    ASSERT_EQ(ret, 0);
    ret = processor3->ReportReplicaInfo(groups0, replicaCnt, replicaOffset);
    ASSERT_EQ(ret, 0);
    ret = processor4->ReportReplicaInfo(groups1, replicaCnt, replicaOffset);
    ASSERT_EQ(ret, 0);
}

TEST_F(ControllerMSTest, ms_invalid_param)
{
    ControllerMSTest::MsInitSource();
    std::vector<int32_t> replicaCnt = { 2 };
    std::vector<int32_t> replicaOffset = { 0 };
    int32_t ret;

    ret = Controller::GetInstance()->Initialize(0, WORLD_SIZE, enableLocalCopy); // 重复初始化
    ASSERT_EQ(ret, TTP_OK);

    std::string ip = "0.0.0.0";
    uint32_t port = CONTROLLER_PORT; // 正常端口
    uint32_t invalidPort = 70000; // 非法端口
    bool localCopy = true;  // 不支持Local Copy

    ProcessorPtr procsser = nullptr;
    ControllerMSTest::InitProcessor(procsser);
    ret = Processor::GetInstance()->Initialize(-1, WORLD_SIZE, enableLocalCopy, testTlsOption); // 非法参数
    ASSERT_EQ(ret, TTP_ERROR);

    ret = Processor::GetInstance()->Initialize(0, -1, enableLocalCopy, testTlsOption); // 非法参数
    ASSERT_EQ(ret, TTP_ERROR);

    ret = Processor::GetInstance()->Initialize(5, WORLD_SIZE, enableLocalCopy, testTlsOption); // 非法参数 rank5
    ASSERT_EQ(ret, TTP_ERROR);

    ret = Processor::GetInstance()->Start(ip, invalidPort); // 端口非法
    ASSERT_EQ(ret, TTP_ERROR);

    ret = Processor::GetInstance()->Start(ip, port); // 处理器的状态更改失败
    ASSERT_EQ(ret, TTP_ERROR);
}

TEST_F(ControllerMSTest, ms_loss_heartbeat_ok)
{
    OutLogger::Instance()->SetLogLevel(DEBUG_LEVEL);
    ControllerMSTest::MsInitSource();
    int32_t ret = processor1->BeginUpdating(-1);
    ASSERT_EQ(ret, 0);
    ret = processor1->FinishedUpdate(COMMON_STEP);
    ASSERT_EQ(ret, 0);

    ret = processor2->BeginUpdating(-1);
    ASSERT_EQ(ret, 0);
    ret = processor2->FinishedUpdate(COMMON_STEP);
    ASSERT_EQ(ret, 0);

    ret = processor3->BeginUpdating(-1);
    ASSERT_EQ(ret, 0);
    ret = processor3->FinishedUpdate(COMMON_STEP);
    ASSERT_EQ(ret, 0);

    ret = processor4->BeginUpdating(-1);
    ASSERT_EQ(ret, 0);
    ret = processor4->FinishedUpdate(COMMON_STEP);
    ASSERT_EQ(ret, 0);

    HeartbeatUpdate();
    processor2->Destroy(true);
    sleep(SLEEP_TWO); // wait controller check;
    processor1->SetDumpResult(0);
    processor4->SetDumpResult(0);
    ret = processor1->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor3->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor4->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);

    ASSERT_EQ(ckptCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(renameCount.load(), 1);
    ASSERT_EQ(exitCount.load(), CHECK_COUNT_THREE);
}

TEST_F(ControllerMSTest, dump_success)
{
    ASSERT_EQ(setenv("MINDIO_FOR_MINDSPORE", "true", 1), 0);

    // ms 每张卡拥有dp组内全量数据，因此副本数与dp大小一致
    ControllerMSTest::MsInitSource();

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    // 构造故障
    ReportState state = ReportState::RS_UNKNOWN;
    processor2->ReportStatus(state);

    sleep(1); // wait for dump 1s

    processor1->SetDumpResult(0);
    processor4->SetDumpResult(0);

    ASSERT_EQ(ckptCount.load(), CHECK_COUNT_TWO);

    // 构造的1号卡故障，因为[0,2],[1,3]互为副本，因此通知dump的集合应该为{{0}},{{3}}
    std::set<std::vector<std::vector<int32_t>>> expect = { {{0}}, {{3}} };
    ASSERT_EQ(ckptRankInfos, expect);

    unsetenv("MINDIO_FOR_MINDSPORE");
}

TEST_F(ControllerMSTest, uce_success)
{
    ASSERT_EQ(setenv("MINDIO_FOR_MINDSPORE", "true", 1), 0);

    // ms 每张卡拥有dp组内全量数据，因此副本数与dp大小一致
    ControllerMSTest::MsInitSource();

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    // 构造rank1,rank2同时uce
    ReportState state = ReportState::RS_UCE;
    processor2->ReportStatus(state);
    processor3->ReportStatus(state);

    auto ret = processor1->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor4->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairSendCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(repairUCECount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(repairRollbackCount.load(), WORLD_SIZE);

    // 构造的1,2号卡故障，因为[0,2],[1,3]互为副本，因此通知repair时，0和2收到的[send:0,recv:2], 1和3收到的[send:3,recv:1]
    std::map<std::string, std::set<std::vector<int32_t>>> expect = {
        {"send", {{0}, {3}}},
        {"ucerecv", {{1}, {2}}}
    };
    ASSERT_EQ(repairRankInfos, expect);

    unsetenv("MINDIO_FOR_MINDSPORE");
}

TEST_F(ControllerMSTest, replica_euqals_zero)
{
    ASSERT_EQ(setenv("MINDIO_FOR_MINDSPORE", "true", 1), 0);

    ControllerMSTest::InitSource(WORLD_SIZE);

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    // 构造全部同时uce，无副本修复
    ReportState state = ReportState::RS_UCE;
    processor1->ReportStatus(state);
    processor2->ReportStatus(state);
    processor3->ReportStatus(state);
    processor4->ReportStatus(state);
    sleep(1);

    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);
    ASSERT_EQ(exitCount.load(), WORLD_SIZE);
    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(repairSendCount.load(), 0);
    ASSERT_EQ(repairUCECount.load(), 0);
    ASSERT_EQ(repairRollbackCount.load(), 0);

    unsetenv("MINDIO_FOR_MINDSPORE");
}

TEST_F(ControllerMSTest, ms_arf_ok)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    ASSERT_EQ(setenv("TTP_RETRY_TIMES", "30", 1), 0);
    ControllerMSTest::MsInitSource(REPLICA_NUM_TWO, true, false);

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);

    std::string ip = CONTROLLER_IP;
    int32_t port = CONTROLLER_PORT;
    std::vector<int32_t> ranks = {1, 3};
    std::vector<std::vector<int32_t>> groups = { ranks };
    std::vector<int32_t> replicaCnt = { 2 };
    std::vector<int32_t> replicaOffset = { 0 };
    ChangeStrategy(STRATEGY_ARF);
    ControllerTest::InitProcessor(processor4);
    processor4->Initialize(3, WORLD_SIZE, enableLocalCopy, testTlsOption, "", true, true); // rank_id is 3
    processor4->Start(ip, port);
    processor4->ReportReplicaInfo(groups, replicaCnt, replicaOffset);

    state = ReportState::RS_INIT_FINISH;
    int32_t ret = processor4->ReportStatus(state);
    ret = processor4->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);

    state = ReportState::RS_PREREPAIR_FINISH;
    ret = processor4->ReportStatus(state);

    ret = processor1->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor4->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(stopCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(cleanCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(ptCommCount.load(), WORLD_SIZE);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
    unsetenv("TTP_RETRY_TIMES");
}

}