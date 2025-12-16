/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */
#ifndef OCK_TTP_CONTROLLER_TEST_COMMON_H
#define OCK_TTP_CONTROLLER_TEST_COMMON_H

#include <fstream>
#include <mutex>
#include <sys/stat.h>
#include <gtest/gtest.h>
#include <mockcpp/mockcpp.hpp>
#define private public
#define protected public
#include "common.h"
#include "controller.h"
#include "processor.h"
#include "mindx_engine.h"
#include "ttp_logger.h"
#undef protected
#undef private

using namespace ock::ttp;

const std::string BACKUP_IP = "127.0.0.1";
const std::string BACKUP_PORT = "1234";
const std::string CONTROLLER_IP = "0.0.0.0";
constexpr uint32_t CONTROLLER_PORT = 8555;
constexpr const int32_t WORLD_SIZE = 4;
constexpr int32_t REPLICA_NUM_TWO = 2;
constexpr int64_t COMMON_STEP = 2;
constexpr int64_t BACKUP_STEP = 1;
constexpr uint32_t UCE_NO_REBUILD = 2;
constexpr uint8_t MASK_NORMAL = 0;
constexpr uint8_t MASK_ERROR = 1;

constexpr uint32_t CHECK_COUNT_ONE = 1;
constexpr uint32_t CHECK_COUNT_TWO = 2;
constexpr uint32_t CHECK_COUNT_THREE = 3;
constexpr uint32_t CHECK_COUNT_FOUR = 4;
constexpr uint32_t CHECK_COUNT_FIVE = 5;
constexpr uint32_t CHECK_COUNT_SIX = 6;
constexpr uint32_t CHECK_COUNT_SEVEN = 7;

constexpr uint32_t SLEEP_TWO = 2;

class ControllerTest : public testing::Test {
public:
    void SetUp() override {}

    void TearDown() override
    {
        usleep(TTP_WAIT_TIME_1MS);
        if (controller1 != nullptr) {
            controller1->Destroy();
            controller1 = nullptr;
        }
        if (processor1 != nullptr) {
            processor1->Destroy();
            processor1 = nullptr;
        }
        if (processor2 != nullptr) {
            processor2->Destroy();
            processor2 = nullptr;
        }
        if (processor3 != nullptr) {
            processor3->Destroy();
            processor3 = nullptr;
        }
        if (processor4 != nullptr) {
            processor4->Destroy();
            processor4 = nullptr;
        }
        if (MindXEngine::GetInstance() != nullptr) {
            MindXEngine::GetInstance()->Destroy();
        }
        GlobalMockObject::verify();
        GlobalMockObject::reset();
    }

    int32_t CallBackFunc(void *ctx, int ctxSize)
    {
        SaveCkptContext *info = static_cast<SaveCkptContext *>(ctx);
        if (info != nullptr) {
            {
                std::lock_guard<std::mutex> lock(ckptRankInfosRanksMutex);
                ckptRankInfos.emplace(info->ranks);
            }
        }

        ckptCount.fetch_add(1);
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

    int StopFunc(void *ctx, int ctxSize)
    {
        stopCount.fetch_add(1);
        return 0;
    }

    int CleanFunc(void *ctx, int ctxSize)
    {
        cleanCount.fetch_add(1);
        int32_t rank = *(static_cast<int32_t *>(ctx));
        auto itr = lowLevelRanks.find(rank);
        return itr != lowLevelRanks.end() ? UCE_NO_REBUILD : 0;
    }

    int RepairFunc(void *ctx, int ctxSize)
    {
        RepairContext *rc = static_cast<RepairContext *>(ctx);
        if (rc->type == RepairType::RT_SEND) {
            repairSendCount.fetch_add(1);
            {
                std::lock_guard<std::mutex> lock(repairRanksMutex);
                repairRanks = rc->ranks;
            }
            {
                std::lock_guard<std::mutex> lock(repairRankInfosMutex);
                repairRankInfos["send"].emplace(rc->srcRank);
            }
        } else if (rc->type == RepairType::RT_UCE_HIGHLEVEL || rc->type == RepairType::RT_UCE_LOWLEVEL) {
            repairUCECount.fetch_add(1);
            std::lock_guard<std::mutex> lock(repairRankInfosMutex);
            repairRankInfos["ucerecv"].emplace(rc->dstRank);
        } else if (rc->type == RepairType::RT_RECV_REPAIR) {
            repairZitRecvCount.fetch_add(1);
            std::lock_guard<std::mutex> lock(repairRankInfosMutex);
            repairRankInfos["otherrecv"].emplace(rc->dstRank);
        } else if (rc->type == RepairType::RT_LOAD_CKPT) {
            repairLoadCkpt.fetch_add(1);
        } else if (rc->type == RepairType::RT_LOAD_REBUILD) {
            repairLoadRebuild.fetch_add(1);
        }

        if (repairFlag.load() == true) {
            return 0;
        } else {
            return 1;
        }
    }

    int RollBackFunc(void *ctx, int ctxSize)
    {
        RepairContext *rc = reinterpret_cast<RepairContext *>(ctx);
        repairRollbackCount.fetch_add(1);
        return 0;
    }

    int Register(void *ctx, int ctxSize)
    {
        registerCount.fetch_add(1);
        return 0;
    }

    int RebuildFunc(void *ctx, int ctxSize)
    {
        ZitRebuildContext *rc = reinterpret_cast<ZitRebuildContext *>(ctx);
        std::lock_guard<std::mutex> lock(repairRanksMutex);
        auto it = std::find(rc->commGroupIdx.begin(), rc->commGroupIdx.end(), 0);
        if (it != rc->commGroupIdx.end()) {
            int32_t dpcpIndex = std::distance(rc->commGroupIdx.begin(), it);
            repairRanks = rc->commGroups[dpcpIndex];
        }
        return 0;
    }

    int PtCommFunc(void *ctx, int ctxSize)
    {
        ptCommCount.fetch_add(1);
        return 0;
    }

    int UpPtCommFunc(void *ctx, int ctxSize)
    {
        upPtCommCount.fetch_add(1);
        return 0;
    }

    int PauseTrainFunc(void *ctx, int ctxSize)
    {
        pauseTrainCount.fetch_add(1);
        return 0;
    }

    int ContinueTrainFunc(void *ctx, int ctxSize)
    {
        continueTrainCount.fetch_add(1);
        return 0;
    }

    static int ReportFaultRanks(void *ctx, int ctxSize)
    {
        ProcessFaultContext *nrsc = static_cast<ProcessFaultContext *>(ctx);
        std::map<int32_t, int32_t> ranks = nrsc->errorInfoMap;
        MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_STOP_TRAIN, &ranks, ranks.size());
        return 0;
    }

    static int ReportFaultRanksUnexcepted(void *ctx, int ctxSize)
    {
        ProcessFaultContext *nrsc = static_cast<ProcessFaultContext *>(ctx);
        NotifyRankInfo rankInfo {nrsc->errorInfoMap, TTP_WAIT_TIME_1MS};
        MindXEngine::GetInstance()->EventProcess(
            MindXEvent::MINDX_EVENT_NOTIFY_FAULT_RANKS, &rankInfo, sizeof(NotifyRankInfo));
        return 0;
    }

    static int ReportStopComplete(void *ctx, int ctxSize)
    {
        StopCompleteContext *nrsc = static_cast<StopCompleteContext *>(ctx);
        NotifyRankInfo rankInfo {nrsc->errorInfoMap, TTP_WAIT_TIME_1MS};
        MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_NOTIFY_FAULT_RANKS,
                                                 &rankInfo, sizeof(NotifyRankInfo));
        return 0;
    }

    static int ReportStopCompletePause(void *ctx, int ctxSize)
    {
        StopCompleteContext *nrsc = static_cast<StopCompleteContext *>(ctx);
        std::map<int32_t, int32_t> ranks = nrsc->errorInfoMap;
        return 0;
    }

    static int ReportStopCompleteUnexcepted(void *ctx, int ctxSize)
    {
        StopCompleteContext *nrsc = static_cast<StopCompleteContext *>(ctx);
        std::map<int32_t, int32_t> ranks = nrsc->errorInfoMap;
        MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_STOP_TRAIN,
                                                 &ranks, ranks.size());
        return 0;
    }

    int ReportStrategies(void *ctx, int ctxSize)
    {
        canChange.store(true);
        return 0;
    }

    int ReportResult(void *ctx, int ctxSize)
    {
        canChange.store(true);
        reportResultCount.fetch_add(1);
        return 0;
    }

    void ChangeStrategy(std::string strategy)
    {
        while (!canChange.load()) {
            usleep(TTP_WAIT_TIME_1MS);
        }
        canChange.store(false);
        std::string param  = "zit test";
        MindXEngine::GetInstance()->ChangeStrategy(strategy, param);
    }

    void InitProcessor(ProcessorPtr &proc)
    {
        int32_t ret;
        proc = MakeRef<Processor>();
        ASSERT_TRUE(proc != nullptr);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_EXIT, std::bind(&ControllerTest::ExitFunc,
                                                                         this, std::placeholders::_1,
                                                                         std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_RENAME, std::bind(&ControllerTest::RenameFunc,
                                                                           this, std::placeholders::_1,
                                                                           std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_SAVE_CKPT, std::bind(&ControllerTest::CallBackFunc,
                                                                              this, std::placeholders::_1,
                                                                              std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_DEVICE_STOP, std::bind(&ControllerTest::StopFunc,
                                                                                this, std::placeholders::_1,
                                                                                std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_DEVICE_CLEAN, std::bind(&ControllerTest::CleanFunc,
                                                                                 this, std::placeholders::_1,
                                                                                 std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_REPAIR, std::bind(&ControllerTest::RepairFunc,
                                                                           this, std::placeholders::_1,
                                                                           std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_ROLLBACK, std::bind(&ControllerTest::RollBackFunc,
                                                                             this, std::placeholders::_1,
                                                                             std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_DOWNGRADE_REBUILD, std::bind(&ControllerTest::RebuildFunc,
                                                                                      this, std::placeholders::_1,
                                                                                      std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_PT_COMM_OPERATE,
                                         std::bind(&ControllerTest::PtCommFunc,
                                                   this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);

        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_UPGRADE_REBUILD, std::bind(&ControllerTest::UpPtCommFunc,
                                                                                    this, std::placeholders::_1,
                                                                                    std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_UPGRADE_REPAIR, std::bind(&ControllerTest::RepairFunc,
                                                                                   this, std::placeholders::_1,
                                                                                   std::placeholders::_2));
        ASSERT_EQ(ret, 0);

        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_UPGRADE_ROLLBACK, std::bind(&ControllerTest::RollBackFunc,
                                                                                     this, std::placeholders::_1,
                                                                                     std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_PAUSE,
                                         std::bind(&ControllerTest::PauseTrainFunc,
                                                   this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = proc->RegisterEventHandler(PROCESSOR_EVENT_CONTINUE,
                                         std::bind(&ControllerTest::ContinueTrainFunc,
                                                   this, std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
    }

    void InitController(ControllerPtr &ctrl)
    {
        int32_t ret;
        ctrl = MakeRef<Controller>();
        ASSERT_TRUE(ctrl != nullptr);
        if (std::getenv("MINDX_TASK_ID") == nullptr) {
            return;
        }
        MindXEnginePtr engine = MindXEngine::GetInstance();
        ret = engine->RegisterEventHandler(MindXEvent::MINDX_EVENT_REGISTER, std::bind(&ControllerTest::Register,
                                                                                       this, std::placeholders::_1,
                                                                                       std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = engine->RegisterEventHandler(MindXEvent::MINDX_EVENT_REPORT_FAULT_RANKS,
                                           &ControllerTest::ReportFaultRanks);
        ASSERT_EQ(ret, 0);
        ret = engine->RegisterEventHandler(MindXEvent::MINDX_EVENT_REPORT_STOP_COMPLETE,
                                           &ControllerTest::ReportStopComplete);
        ASSERT_EQ(ret, 0);
        ret = engine->RegisterEventHandler(MindXEvent::MINDX_EVENT_REPORT_STRATEGIES,
                                           std::bind(&ControllerTest::ReportStrategies, this,
                                                     std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
        ret = engine->RegisterEventHandler(MindXEvent::MINDX_EVENT_REPORT_RESULT,
                                           std::bind(&ControllerTest::ReportResult, this,
                                                     std::placeholders::_1, std::placeholders::_2));
        ASSERT_EQ(ret, 0);
    }

    static std::vector<BackupInfo> SelectBackUpController()
    {
        std::vector<BackupInfo> backUps;
        BackupInfo info;
        info.rank = 1;
        info.ip = BACKUP_IP;
        info.port = BACKUP_PORT; // unused
        backUps.push_back(info);
        return backUps;
    }

    void CountClean()
    {
        ckptCount.store(0);
        renameCount.store(0);
        exitCount.store(0);
        stopCount.store(0);
        cleanCount.store(0);
        repairSendCount.store(0);
        repairUCECount.store(0);
        repairLoadCkpt.store(0);
        repairLoadRebuild.store(0);
        repairRollbackCount.store(0);
        registerCount.store(0);
        ptCommCount.store(0);
        upPtCommCount.store(0);
        pauseTrainCount.store(0);
        continueTrainCount.store(0);
        repairZitRecvCount.store(0);
        canChange.store(false);
    }

    void MapInfoClean()
    {
        lowLevelRanks.clear();
        {
            std::lock_guard<std::mutex> lock(repairRanksMutex);
            repairRanks.clear();
        }
        {
            std::lock_guard<std::mutex> lock(ckptRankInfosRanksMutex);
            ckptRankInfos.clear();
        }
        {
            std::lock_guard<std::mutex> lock(repairRankInfosMutex);
            repairRankInfos.clear();
        }
    }

    void InitSource(int32_t controllerReplica = 2, bool enableARF = false, bool enableZIT = false)
    {
        ControllerTest::CountClean();
        ControllerTest::MapInfoClean();

        ControllerTest::InitController(controller1);
        std::vector<int32_t> replicaCnt = { controllerReplica };
        std::vector<int32_t> replicaOffset = { 0 };
        int32_t ret = controller1->Initialize(0, WORLD_SIZE, enableLocalCopy, enableARF, enableZIT);
        controller1->retrySwitch_ = true;
        ASSERT_EQ(ret, 0);

        std::string ip = CONTROLLER_IP;
        int32_t port = CONTROLLER_PORT;
        ret = controller1->Start(ip, port, testTlsOption);
        ASSERT_EQ(ret, 0);

        std::vector<int32_t> ranks = {0, 1, 2, 3};
        std::vector<std::vector<int32_t>> groups = { ranks };
        ControllerTest::InitProcessor(processor1);
        ControllerTest::InitProcessor(processor2);
        ControllerTest::InitProcessor(processor3);
        ControllerTest::InitProcessor(processor4);

        ret = processor1->Initialize(0, WORLD_SIZE, enableLocalCopy, testTlsOption, true, enableARF, enableZIT);
        ASSERT_EQ(ret, 0);
        ret = processor2->Initialize(1, WORLD_SIZE, enableLocalCopy, testTlsOption, true, enableARF, enableZIT);
        ASSERT_EQ(ret, 0);
        ret = processor3->Initialize(2, WORLD_SIZE, enableLocalCopy, testTlsOption, true, enableARF, enableZIT);
        ASSERT_EQ(ret, 0);
        ret = processor4->Initialize(3, WORLD_SIZE, enableLocalCopy, testTlsOption, true, enableARF, enableZIT);
        ASSERT_EQ(ret, 0);

        ret = processor1->Start(ip, port);
        ASSERT_EQ(ret, 0);
        ret = processor2->Start(ip, port);
        ASSERT_EQ(ret, 0);
        ret = processor3->Start(ip, port);
        ASSERT_EQ(ret, 0);
        ret = processor4->Start(ip, port);
        ASSERT_EQ(ret, 0);

        ret = processor1->ReportReplicaInfo(groups, replicaCnt, replicaOffset);
        ASSERT_EQ(ret, 0);
        ret = processor2->ReportReplicaInfo(groups, replicaCnt, replicaOffset);
        ASSERT_EQ(ret, 0);
        ret = processor3->ReportReplicaInfo(groups, replicaCnt, replicaOffset);
        ASSERT_EQ(ret, 0);
        ret = processor4->ReportReplicaInfo(groups, replicaCnt, replicaOffset);
        ASSERT_EQ(ret, 0);
    }

    void ProcessorUpdate(ProcessorPtr &proc)
    {
        int32_t ret = proc->BeginUpdating(BACKUP_STEP);
        ASSERT_EQ(ret, 0);
        ret = proc->FinishedUpdate(COMMON_STEP);
        ASSERT_EQ(ret, 0);
    }

    void HeartbeatUpdate()
    {
        int32_t ret = processor1->HeartbeatSend();
        ASSERT_EQ(ret, 0);
        ret = processor2->HeartbeatSend();
        ASSERT_EQ(ret, 0);
        ret = processor3->HeartbeatSend();
        ASSERT_EQ(ret, 0);
        ret = processor4->HeartbeatSend();
        ASSERT_EQ(ret, 0);
    }
public:
    std::atomic<uint32_t> ckptCount;
    std::atomic<uint32_t> renameCount;
    std::atomic<uint32_t> exitCount;
    std::atomic<uint32_t> stopCount;
    std::atomic<uint32_t> cleanCount;
    std::atomic<uint32_t> repairSendCount;
    std::atomic<uint32_t> repairUCECount;
    std::atomic<uint32_t> repairRollbackCount;
    std::atomic<uint32_t> registerCount;
    std::atomic<uint32_t> reportResultCount;
    std::atomic<uint32_t> ptCommCount;
    std::atomic<uint32_t> upPtCommCount;
    std::atomic<uint32_t> pauseTrainCount;
    std::atomic<uint32_t> continueTrainCount;
    std::atomic<uint32_t> repairZitRecvCount;
    std::atomic<uint32_t> repairLoadCkpt;
    std::atomic<uint32_t> repairLoadRebuild;
    std::mutex repairRankInfosMutex;
    std::mutex repairRanksMutex;
    std::mutex ckptRankInfosRanksMutex;

    std::set<int32_t> lowLevelRanks;
    std::vector<int32_t> repairRanks;
    std::set<std::vector<std::vector<int32_t>>> ckptRankInfos;
    std::map<std::string, std::set<std::vector<int32_t>>> repairRankInfos;

    ProcessorPtr processor1 = nullptr;
    ProcessorPtr processor2 = nullptr;
    ProcessorPtr processor3 = nullptr;
    ProcessorPtr processor4 = nullptr;
    ControllerPtr controller1 = nullptr;
    ControllerPtr controller2 = nullptr;
    AccTlsOption testTlsOption;
    bool enableLocalCopy = false;
    std::atomic<bool> repairFlag = { true };
    std::atomic<bool> canChange = { false };
};
#endif // OCK_TTP_CONTROLLER_TEST_COMMON_H
