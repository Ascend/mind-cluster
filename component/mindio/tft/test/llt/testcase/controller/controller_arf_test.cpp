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

class ControllerARFTest : public ControllerTest {
public:
    void WaitNormal()
    {
        int32_t ret = processor1->WaitRepairAction();
        ASSERT_EQ(ret, TTP_OK);
        ret = processor2->WaitRepairAction();
        ASSERT_EQ(ret, TTP_OK);
        ret = processor3->WaitRepairAction();
        ASSERT_EQ(ret, TTP_OK);
        ret = processor4->WaitRepairAction();
        ASSERT_EQ(ret, TTP_OK);
        usleep(TTP_WAIT_TIME_1MS);
    }
};

TEST_F(ControllerARFTest, handle_downgrade_running)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    setenv("TTP_LOG_LEVEL", "DEBUG", 1);
    setenv("TTP_LOG_MODE", "ONLY_ONE", 1);
    setenv("TTP_LOG_SIZE", "4096", 1);
    setenv("TEST_LOG_OPEN", "1", 1);
    ControllerARFTest::InitSource(REPLICA_NUM_TWO, false, true);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ret = processor4->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, 0);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    ChangeStrategy(STRATEGY_DOWNGRADE);
    ret = processor1->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(stopCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(cleanCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(repairRanks.size(), CHECK_COUNT_THREE);
    ASSERT_EQ(repairRanks[0], 0);
    ASSERT_EQ(repairRanks[1], 1);
    ASSERT_EQ(repairRanks[2], CHECK_COUNT_TWO); // rank 2
    unsetenv("TTP_LOG_LEVEL");
    unsetenv("TTP_LOG_MODE");
    unsetenv("TTP_LOG_SIZE");
    unsetenv("TEST_LOG_OPEN");
    setenv("TTP_LOG_STDOUT", "1", 1);
    unsetenv("MINDX_TASK_ID");
    OutLogger::Instance()->SetExternalLogFunction(nullptr);
}

TEST_F(ControllerARFTest, handle_downgrade_upgrade_recover)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&Controller::SelectBackUpController,
               std::vector<BackupInfo>(*)(void)).stubs().will(invoke(ControllerARFTest::SelectBackUpController));
    ControllerARFTest::InitSource(REPLICA_NUM_TWO, false, true);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ret = processor4->BeginUpdating(BACKUP_STEP);
    ASSERT_EQ(ret, 0);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);

    sleep(1);
    ChangeStrategy(STRATEGY_DOWNGRADE);
    ret = processor1->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor2->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);
    ret = processor3->WaitNextAction();
    ASSERT_EQ(ret, TTP_OK);

    ASSERT_EQ(stopCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(cleanCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(repairRanks.size(), CHECK_COUNT_THREE);
    ASSERT_EQ(repairRanks[0], 0);
    ASSERT_EQ(repairRanks[1], 1);
    ASSERT_EQ(repairRanks[2], CHECK_COUNT_TWO); // rank 2

    processor4->Destroy(true);
    sleep(1);
    // Raise Processor4
    std::string ip = CONTROLLER_IP;
    int32_t port = CONTROLLER_PORT;
    std::vector<int32_t> ranks = {0, 1, 2, 3};
    std::vector<std::vector<int32_t>> groups = { ranks };
    std::vector<int32_t> replicaCnt = { 2 };
    std::vector<int32_t> replicaOffset = { 0 };
    ControllerARFTest::InitProcessor(processor4);
    processor4->Initialize(3, WORLD_SIZE, enableLocalCopy, testTlsOption, "", false, true); // rank_id is 3
    processor4->Start(ip, port);
    processor4->ReportReplicaInfo(groups, replicaCnt, replicaOffset);
    processor4->ReportDpInfo(ranks);
    // Set status
    state = ReportState::RS_PREREPAIR_FINISH;
    ret = processor4->ReportStatus(state);
    sleep(1);
    ChangeStrategy(STRATEGY_UPGRADE);
    WaitNormal();
    ASSERT_EQ(stopCount.load(), CHECK_COUNT_SIX);
    ASSERT_EQ(cleanCount.load(), CHECK_COUNT_SIX);
    ASSERT_EQ(upPtCommCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairZitRecvCount.load(), 1);
    ASSERT_EQ(repairRollbackCount.load(), CHECK_COUNT_FOUR); // WORLD_SIZE

    MOCKCPP_RESET;
    unsetenv("MINDX_TASK_ID");
}

TEST_F(ControllerARFTest, arf_repair)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    ASSERT_EQ(setenv("TTP_RETRY_TIMES", "30", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

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
    std::vector<int32_t> ranks = {0, 1, 2, 3};
    std::vector<std::vector<int32_t>> groups = { ranks };
    std::vector<int32_t> replicaCnt = { 2 };
    std::vector<int32_t> replicaOffset = { 0 };
    ChangeStrategy(STRATEGY_ARF);
    ControllerARFTest::InitProcessor(processor4);
    processor4->Initialize(3, WORLD_SIZE, enableLocalCopy, testTlsOption, "", true, true); // rank_id is 3
    processor4->Start(ip, port);
    processor4->ReportReplicaInfo(groups, replicaCnt, replicaOffset);

    state = ReportState::RS_PREREPAIR_FINISH;
    ret = processor4->ReportStatus(state);
    WaitNormal();

    ASSERT_EQ(processor1->GetRepairType(), "recover");
    ASSERT_EQ(stopCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(cleanCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(ptCommCount.load(), WORLD_SIZE);
    ASSERT_EQ(repairSendCount.load(), CHECK_COUNT_ONE);
    ASSERT_EQ(repairZitRecvCount.load(), CHECK_COUNT_ONE);
    ASSERT_EQ(repairRollbackCount.load(), WORLD_SIZE);

    // 构造的3号卡故障，因为dp组为[0,1,2,3]，两副本，因此通知repair时，1和3收到的[send:1,recv:3]
    std::map<std::string, std::set<std::vector<int32_t>>> expect = {
        {"send", {{1}}},
        {"otherrecv", {{3}}}
    };
    ASSERT_EQ(repairRankInfos, expect);

    unsetenv("MINDX_TASK_ID");
    unsetenv("TTP_RETRY_TIMES");
}

TEST_F(ControllerARFTest, arf_wait_change_strategy_exit)
{
    MOCKER_CPP(&ControllerARFTest::StopFunc, int32_t(*)(void *, int)).stubs().will(returnValue(1));
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    processor4->Destroy(true);
    ChangeStrategy(STRATEGY_EXIT);

    ret = processor1->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor2->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor3->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerARFTest, arf_wait_change_strategy_dump)
{
    MOCKER_CPP(&ControllerARFTest::StopFunc, int32_t(*)(void *, int)).stubs().will(returnValue(1));
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);

    ChangeStrategy(STRATEGY_ARF);
    ChangeStrategy(STRATEGY_DUMP);
    processor1->SetDumpResult(0);
    processor2->SetDumpResult(0);
    ret = processor1->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor2->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);
    ret = processor3->WaitNextAction();
    ASSERT_NE(ret, TTP_OK);

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerARFTest, arf_notify_dump)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);
    ChangeStrategy(STRATEGY_DUMP);

    processor1->SetDumpResult(0);
    processor2->SetDumpResult(0);
    processor3->SetDumpResult(0);

    ret = processor1->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor2->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor3->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(renameCount.load(), 1);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
}

TEST_F(ControllerARFTest, notify_dump_with_stop_clean)
{
    ASSERT_EQ(setenv("TTP_STOP_CLEAN_BEFORE_DUMP", "1", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);

    processor1->SetDumpResult(0);
    processor2->SetDumpResult(0);
    processor3->SetDumpResult(0);

    ret = processor1->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor2->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor3->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);

    ASSERT_EQ(stopCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(cleanCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(ckptCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(renameCount.load(), CHECK_COUNT_ONE);
    ASSERT_EQ(registerCount, 0);

    unsetenv("TTP_STOP_CLEAN_BEFORE_DUMP");
}

TEST_F(ControllerARFTest, notify_dump_without_stop_clean)
{
    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);

    processor1->SetDumpResult(0);
    processor2->SetDumpResult(0);
    processor3->SetDumpResult(0);

    ret = processor1->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor2->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor3->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(renameCount.load(), CHECK_COUNT_ONE);
    ASSERT_EQ(registerCount, 0);
}

TEST_F(ControllerARFTest, arf_clean_faild)
{
    MOCKER_CPP(&ControllerARFTest::CleanFunc, int32_t(*)(void *, int)).stubs().will(returnValue(1));
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);
    ChangeStrategy(STRATEGY_ARF);

    sleep(1);
    ChangeStrategy(STRATEGY_EXIT);
    ret = processor1->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor2->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor3->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);

    ASSERT_EQ(stopCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(cleanCount.load(), 0);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerARFTest, arf_notify_dump_failed)
{
    MOCKER_CPP(&Controller::HandleDumpStatus, int32_t(*)(const std::set<int32_t>&)).stubs().will(returnValue(1));
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);
    ChangeStrategy(STRATEGY_DUMP);

    ret = processor1->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor2->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor3->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerARFTest, arf_notify_dump_rename_failed)
{
    MOCKER_CPP(&Controller::Rename, int32_t(*)(std::set<int32_t>, int64_t)).stubs().will(returnValue(1));
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);
    ChangeStrategy(STRATEGY_DUMP);

    processor1->SetDumpResult(0);
    processor2->SetDumpResult(0);
    processor3->SetDumpResult(0);

    ret = processor1->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor2->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);
    ret = processor3->WaitNextAction();
    ASSERT_EQ(ret, TTP_ERROR);

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), CHECK_COUNT_TWO);
    ASSERT_EQ(renameCount.load(), 0);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerARFTest, arf_unexpected_report_process_fault_calling)
{
    MOCKER_CPP(&ControllerARFTest::ReportFaultRanks, int(*)(void *, int)).
        expects(once()).
        will(invoke(&ControllerARFTest::ReportFaultRanksUnexcepted));
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerARFTest, arf_unexpected_report_stop_calling)
{
    MOCKER_CPP(&ControllerARFTest::ReportStopComplete, int(*)(void *, int)).
        expects(once()).
        will(invoke(&ControllerARFTest::ReportStopCompleteUnexcepted));
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
    MOCKCPP_RESET;
}

TEST_F(ControllerARFTest, arf_rollback_failed)
{
    MOCKER_CPP(&Controller::NotifyRankRollback, int(*)(const std::vector<int32_t> &, RepairType)).
        expects(once()).
        will(returnValue(1));
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

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
    std::vector<int32_t> ranks = {0, 1, 2, 3};
    std::vector<std::vector<int32_t>> groups = { ranks };
    std::vector<int32_t> replicaCnt = { 2 };
    std::vector<int32_t> replicaOffset = { 0 };
    ChangeStrategy(STRATEGY_ARF);
    ControllerARFTest::InitProcessor(processor4);
    processor4->Initialize(3, WORLD_SIZE, enableLocalCopy, testTlsOption, "", true, true); // rank_id is 3
    processor4->Start(ip, port);
    processor4->ReportReplicaInfo(groups, replicaCnt, replicaOffset);

    state = ReportState::RS_PREREPAIR_FINISH;
    ret = processor4->ReportStatus(state);
    ChangeStrategy(STRATEGY_EXIT);

    ASSERT_EQ(stopCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(cleanCount.load(), CHECK_COUNT_THREE);
    ASSERT_EQ(ptCommCount.load(), WORLD_SIZE);
    ASSERT_EQ(registerCount, 1);

    unsetenv("MINDX_TASK_ID");
    unsetenv("TTP_RETRY_TIMES");
    MOCKCPP_RESET;
}

TEST_F(ControllerARFTest, uninitialized_process_error)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerTest::CountClean();
    ControllerTest::InitController(controller1);
    int32_t ret = controller1->Initialize(0, WORLD_SIZE, false, true, false);
    ASSERT_EQ(ret, 0);

    std::string ip = CONTROLLER_IP;
    int32_t port = CONTROLLER_PORT;
    ret = controller1->Start(ip, port, testTlsOption);

    std::map<int32_t, int32_t> rankList;
    rankList[0] = 1;
    controller1->MindXNotifyStopTrain(&rankList, rankList.size());
    sleep(SLEEP_TWO);

    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);

    unsetenv("MINDX_TASK_ID");
    unsetenv("TTP_RETRY_TIMES");
}

TEST_F(ControllerARFTest, arf_notify_invalid_strategy)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    ReportState state = ReportState::RS_UNKNOWN;
    processor4->ReportStatus(state);
    sleep(1);
    processor4->Destroy(true);
    auto isHighAvailability = controller1->GetHighAvailabilitySwitch();
    MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_INVALID, nullptr, 0);

    ASSERT_EQ(isHighAvailability, true);
    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);

    unsetenv("MINDX_TASK_ID");
}

TEST_F(ControllerARFTest, arf_notify_pause_train)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ControllerARFTest::ReportStopComplete, int(*)(void *, int)).
        expects(once()).
        will(invoke(&ControllerARFTest::ReportStopCompletePause));

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    auto isHighAvailability = controller1->GetHighAvailabilitySwitch();
    uint32_t timeout = 30;
    MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_PAUSE_TRAIN, &timeout, sizeof(timeout));
    usleep(TTP_WAIT_TIME_1MS);

    MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_CONTINUE_TRAIN, nullptr, 0);
    usleep(TTP_WAIT_TIME_1MS);

    ASSERT_EQ(isHighAvailability, true);
    ASSERT_EQ(pauseTrainCount.load(), CHECK_COUNT_FOUR);
    ASSERT_EQ(continueTrainCount.load(), CHECK_COUNT_FOUR);
    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);

    unsetenv("MINDX_TASK_ID");
}

TEST_F(ControllerARFTest, arf_notify_continue_train_failed)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ControllerARFTest::ReportStopComplete, int(*)(void *, int)).
        expects(atLeast(1)).
        will(invoke(&ControllerARFTest::ReportStopCompletePause));
    MOCKER_CPP(&Controller::ContinueTrain, int(*)()).
        expects(once()).
        will(returnValue(1));

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    auto isHighAvailability = controller1->GetHighAvailabilitySwitch();
    uint32_t timeout = 30;
    MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_PAUSE_TRAIN, &timeout, sizeof(timeout));
    usleep(TTP_WAIT_TIME_1MS);

    MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_CONTINUE_TRAIN, nullptr, 0);
    usleep(TTP_WAIT_TIME_1MS);

    ASSERT_EQ(isHighAvailability, true);
    ASSERT_EQ(pauseTrainCount.load(), CHECK_COUNT_FOUR);
    ASSERT_EQ(continueTrainCount.load(), 0);
    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);

    unsetenv("MINDX_TASK_ID");
}

TEST_F(ControllerARFTest, arf_notify_pause_wait_failed)
{
    ASSERT_EQ(setenv("MINDX_TASK_ID", "0", 1), 0);
    MOCKER_CPP(&ControllerARFTest::ReportStopComplete, int(*)(void *, int)).
        expects(atLeast(1)).
        will(invoke(&ControllerARFTest::ReportStopCompletePause));
    MOCKER_CPP(&Controller::PauseWait, int(*)(RepairType)).expects(once()).will(returnValue(1));

    ControllerARFTest::InitSource(REPLICA_NUM_TWO, true, false);
    int32_t ret;

    ProcessorUpdate(processor1);
    ProcessorUpdate(processor2);
    ProcessorUpdate(processor3);
    ProcessorUpdate(processor4);

    auto isHighAvailability = controller1->GetHighAvailabilitySwitch();
    uint32_t timeout = 30;
    MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_PAUSE_TRAIN, &timeout, sizeof(uint32_t));
    usleep(TTP_WAIT_TIME_1MS);

    ASSERT_EQ(isHighAvailability, true);
    ASSERT_EQ(pauseTrainCount.load(), CHECK_COUNT_FOUR);
    ASSERT_EQ(continueTrainCount.load(), 0);
    ASSERT_EQ(stopCount.load(), 0);
    ASSERT_EQ(cleanCount.load(), 0);
    ASSERT_EQ(ckptCount.load(), 0);
    ASSERT_EQ(renameCount.load(), 0);

    unsetenv("MINDX_TASK_ID");
}
}