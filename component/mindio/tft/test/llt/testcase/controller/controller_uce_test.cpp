/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */
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

class ControllerUCETest : public ControllerTest {};

TEST_F(ControllerUCETest, handle_mindx_reject)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ControllerUCETest::ReportStrategies, int(*)(void *ctx, int ctxSize)).
        stubs().will(returnValue(400)); // invalid return 400

    ControllerUCETest::InitSource();
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ret = processor4->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, 0);

    ReportState state = ReportState::RS_UCE;
    processor4->ReportStatus(state);

    sleep(1); // wait for dump 1s

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(repairSendCount.load(), 0);
    ASSERT_EQ(repairUCECount.load(), 0);
    ASSERT_EQ(repairRollbackCount.load(), 0);
    ASSERT_EQ(repairRanks.size(), 0);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerUCETest, handle_uce_failed)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    ControllerUCETest::InitSource();
    repairFlag.store(false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ret = processor4->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, 0);

    ReportState state = ReportState::RS_UCE;
    processor4->ReportStatus(state);

    ChangeStrategy(STRATEGY_RETRY);
    sleep(1); // wait 1s
    repairFlag.store(true);
    std::map<int32_t, int32_t> ranks;
    NotifyRankInfo rankInfo {ranks, TTP_WAIT_TIME_1MS};
    MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_NOTIFY_FAULT_RANKS,
                                             &rankInfo, sizeof(NotifyRankInfo));
    ChangeStrategy(STRATEGY_EXIT);

    ret = processor1->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor2->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor3->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor4->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);

    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairSendCount.load(), 1);
    ASSERT_EQ(repairUCECount.load(), 1);

    unsetenv("MINDX_TASK_ID");
}

// UCE
TEST_F(ControllerUCETest, engine_process_fail_in_uce_fail)
{
    ControllerUCETest::InitSource(REPLICA_NUM_TWO);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ret = processor4->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, 0);

    MOCKER_CPP(&ActionEngine::Process, int32_t(*)(uint32_t, std::vector<ActionInfo>, bool))
        .stubs().will(returnValue(1));
    MOCKER_CPP(&AccTcpClientDefault::Send, int32_t(*)(int16_t msgType, uint8_t *data, uint32_t len))
        .stubs().will(returnValue(0));
    MOCKER_CPP(&AccTcpClientDefault::Connect, int32_t (*)(const AccConnReq &connReq,
        uint32_t maxConnRetryTimes)).stubs().will(returnValue(-1));

    ReportState state = ReportState::RS_UCE;
    processor4->ReportStatus(state);

    sleep(1);

    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);
    ASSERT_EQ(exitCount.load(), 0);
    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(repairSendCount.load(), 0);
    ASSERT_EQ(repairUCECount.load(), 0);
    ASSERT_EQ(repairRollbackCount.load(), 0);

    MOCKCPP_RESET;
}

TEST_F(ControllerUCETest, replica_euqals_zero)
{
    ControllerUCETest::InitSource(1);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ret = processor4->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, 0);

    // InitSource副本为1，一张卡故障触发无副本流程，走周期ckpt修复
    ReportState state = ReportState::RS_UCE;
    processor4->ReportStatus(state);
    sleep(1);

    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);
    ASSERT_EQ(exitCount.load(), 0);
    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairLoadCkpt.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(repairLoadRebuild.load(), CHECK_COUNT_ONE);
    ASSERT_EQ(repairRollbackCount.load(), WORLD_SIZE);
}

TEST_F(ControllerUCETest, handle_uce_lowlevel_success)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ControllerUCETest::Register, TResult(*)(void)).
        expects(once()).will(returnValue(TTP_ERROR));

    ControllerUCETest::InitSource();
    int32_t ret;
    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    // 构造rank0, rank3 uce low level 故障
    lowLevelRanks = {0, 3};

    ReportState state = ReportState::RS_UCE;
    processor1->ReportStatus(state);
    processor4->ReportStatus(state);

    ret = processor1->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor4->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairSendCount.load(), 0);
    ASSERT_EQ(repairUCECount.load(), 0);
    ASSERT_EQ(repairRollbackCount.load(), WORLD_SIZE);

    ASSERT_TRUE(repairRankInfos.empty());

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerUCETest, handle_hccl_success)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ControllerUCETest::Register, TResult(*)(void)).
    expects(once()).will(returnValue(TTP_ERROR));

    ControllerUCETest::InitSource();
    int32_t ret;
    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_HCCL_FAILED;
    processor1->ReportStatus(state);
    processor4->ReportStatus(state);

    ret = processor1->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor4->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairSendCount.load(), 0);
    ASSERT_EQ(repairUCECount.load(), 0);
    ASSERT_EQ(repairRollbackCount.load(), WORLD_SIZE);

    ASSERT_TRUE(repairRankInfos.empty());

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerUCETest, handle_uce_highlevel_success)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ControllerUCETest::Register, TResult(*)(void)).
        expects(once()).will(returnValue(TTP_ERROR));

    ControllerUCETest::InitSource();
    int32_t ret;
    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UCE;
    processor1->ReportStatus(state);
    processor4->ReportStatus(state);

    ret = processor1->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor4->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(processor2->GetRepairType(), "retry");
    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairSendCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(repairUCECount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(repairRollbackCount.load(), WORLD_SIZE);

    // 构造的0,3号卡故障，因为dp组为[0,1,2,3]，两副本，因此通知repair时，0和2收到的[send:2,recv:0], 1和3收到的[send:1,recv:3]
    std::map<std::string, std::set<std::vector<int32_t>>> expect = {
        {"send", {{1}, {2}}},
        {"ucerecv", {{0}, {3}}}
    };
    ASSERT_EQ(repairRankInfos, expect);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerUCETest, handle_x1_uce_highlevel_success)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    ASSERT_EQ(setenv("TTP_FRAMEWORK_TYPE", "1", 1), 0);
    MOCKER_CPP(&ControllerUCETest::Register, TResult(*)(void)).
    expects(once()).will(returnValue(TTP_ERROR));

    ControllerUCETest::InitSource();
    int32_t ret;
    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UCE;
    processor1->ReportStatus(state);
    processor4->ReportStatus(state);

    ret = processor1->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor4->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(processor2->GetRepairType(), "retry");
    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairSendCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(repairUCECount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(repairRollbackCount.load(), WORLD_SIZE);

    // 构造的0,3号卡故障，因为dp组为[0,1,2,3]，两副本，因此通知repair时，0和2收到的[send:2,recv:0], 1和3收到的[send:1,recv:3]
    std::map<std::string, std::set<std::vector<int32_t>>> expect = {
        {"send", {{1}, {2}}},
        {"ucerecv", {{0}, {3}}}
    };
    ASSERT_EQ(repairRankInfos, expect);

    unsetenv("MINDX_TASK_ID");
    unsetenv("TTP_FRAMEWORK_TYPE");
    MOCKCPP_RESET;
}

TEST_F(ControllerUCETest, repair_msg_overflow_fail)
{
    ControllerUCETest::InitSource();
    repairFlag.store(false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ret = processor4->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, 0);

    ReportState state = ReportState::RS_UCE;
    processor4->ReportStatus(state);
    MOCKER_CPP(&IsOverflow, bool(*)(uint32_t, uint32_t)).stubs().will(returnValue(true));

    ret = processor1->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor2->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor3->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor4->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    repairFlag.store(true);

    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairSendCount.load(), 0);
    ASSERT_EQ(repairUCECount.load(), 0);

    MOCKCPP_RESET;
}

TEST_F(ControllerUCETest, clean_failed_update_errorRanks)
{
    MOCKER_CPP(&ControllerUCETest::CleanFunc, int32_t(*)(void *, int)).stubs().will(returnValue(1));
    ControllerUCETest::InitSource();
    int32_t ret;
    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UCE;
    processor4->ReportStatus(state);

    ret = processor1->WaitRepairAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor2->WaitRepairAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor3->WaitRepairAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor4->WaitRepairAction();
    ASSERT_EQ(ret, TTP_ERROR);

    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(repairSendCount.load(), 0);
    ASSERT_EQ(repairUCECount.load(), 0);
    ASSERT_EQ(repairRollbackCount.load(), 0);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerUCETest, handle_uce_corrupted_load_ckpt_success)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ControllerUCETest::Register, TResult(*)(void)).
    expects(once()).will(returnValue(TTP_ERROR));

    ControllerUCETest::InitSource();
    int32_t ret;
    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UCE_CORRUPTED;
    processor1->ReportStatus(state);
    processor4->ReportStatus(state);

    ret = processor1->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor4->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(processor2->GetRepairType(), "retry");
    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairLoadCkpt.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(repairRollbackCount.load(), WORLD_SIZE);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerUCETest, handle_precision_error_load_ckpt_success)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ControllerUCETest::Register, TResult(*)(void)).
    expects(once()).will(returnValue(TTP_ERROR));

    ControllerUCETest::InitSource();
    int32_t ret;
    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_RETRY;
    processor1->ReportStatus(state);
    processor2->ReportStatus(state);
    processor3->ReportStatus(state);
    processor4->ReportStatus(state);

    ret = processor1->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor4->WaitRepairAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(processor2->GetRepairType(), "retry");
    ASSERT_EQ(stopCount.load(), WORLD_SIZE);
    ASSERT_EQ(cleanCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairLoadCkpt.load(), CHECK_COUNT_FOUR);
    ASSERT_EQ(repairRollbackCount.load(), WORLD_SIZE);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}
}