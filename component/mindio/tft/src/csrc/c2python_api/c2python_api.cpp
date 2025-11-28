/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#include <sys/syscall.h>
#include "controller.h"
#include "processor.h"
#include "mindx_engine.h"
#include "ttp_logger.h"

using namespace ock::ttp;

using CkptCallback = std::function<int(int, int, std::vector<int32_t>&, std::vector<std::vector<int32_t>>&)>;
using RepairCallback = std::function<int(RepairContext&)>;
using RollbackCallback = std::function<int(RepairContext&)>;
using PauseCallback = std::function<int(int64_t, bool)>;
using ContinueCallback = std::function<int(void)>;
using DowngradeRebuildCallback = std::function<int(std::vector<int32_t>&, std::vector<std::vector<int32_t>>&,
    int32_t repairId, std::string&)>;
using ReportStopCompleteCallback = std::function<int(int, std::string&, std::map<int32_t, int32_t>&)>;
using ReportStrategiesCallback = std::function<int(std::map<int32_t, int32_t>&, std::vector<std::string>&)>;
using ReportResultCallback = std::function<int(int, std::string&, std::map<int32_t, int32_t>&, std::string&)>;
using ReportFaultRanksCallback = std::function<int(std::map<int32_t, int32_t>&)>;
using RenameCallback = std::function<int(void)>;
using ExitCallback = std::function<void(void)>;
using DeviceStopCallback = std::function<int(void)>;
using DeviceCleanCallback = std::function<int(void)>;
using RegisterCallback = std::function<int(void)>;
using CommunicationOperateCallback = std::function<int(std::vector<int32_t>&, int)>;
using UpgradeRebuildCallback = std::function<int(std::vector<int32_t>&, int32_t, std::string&)>;
using LaunchTcpStoreClientCallback = std::function<int(std::string&, int, int)>;
using LaunchTcpStoreServerCallback = std::function<int(std::string&, int)>;
using DecryptCallback = std::function<std::string(const std::string&)>;

class CallbackReceiver {
public:
    virtual int CkptCallback(int step, int repairId, std::vector<int> groupIdx,
                             std::vector<std::vector<int>> ranks) = 0;
    virtual int RepairCallback(RepairContext context) = 0;
    virtual int UpgradeRepairCallback(RepairContext context) = 0;
    virtual int RollbackCallback(RepairContext context) = 0;
    virtual int PauseCallback(int64_t pauseStep, bool hotSwitch) = 0;
    virtual int UpgradeRollbackCallback(RepairContext context) = 0;
    virtual int ContinueCallback() = 0;
    virtual int DowngradeRebuildCallback(std::vector<int32_t> commGroupIdx,
        std::vector<std::vector<int32_t>> commGroups, int32_t repairId, std::string param) = 0;
    virtual int UpgradeRebuildCallback(std::vector<int32_t> rankList, int repairId, std::string param) = 0;
    virtual int ReportStopCompleteCallback(int code, std::string msg, std::map<int32_t, int32_t> errorInfoMap) = 0;
    virtual int ReportStrategiesCallback(std::map<int32_t, int32_t> errorInfoMap,
                                              std::vector<std::string> strategies) = 0;
    virtual int ReportResultCallback(int code, std::string msg, std::map<int32_t, int32_t> errorInfoMap,
                                            std::string strategy) = 0;
    virtual int ReportFaultRanksCallback(std::map<int32_t, int32_t> errorInfoMap) = 0;
    virtual int RenameCallback(void) = 0;
    virtual void ExitCallback(void) {};
    virtual int DeviceStopCallback(void) = 0;
    virtual int DeviceCleanCallback(void) = 0;
    virtual int RegisterCallback(void) = 0;
    virtual int CommunicationOperateCallback(std::vector<int32_t> rankList, int repairId) = 0;
    virtual int LaunchTcpStoreClientCallback(std::string url, int rank, int worldSize) = 0;
    virtual int LaunchTcpStoreServerCallback(std::string url, int worldSize) = 0;
    virtual std::string DecryptCallback(std::string cipherText) = 0;
    virtual ~CallbackReceiver() {};
};

static CallbackReceiver *receiver_ptr = nullptr;

static void SetCallbackReceiver(CallbackReceiver *handler)
{
    TTP_ASSERT_RET_VOID(handler != nullptr);
    receiver_ptr = handler;
}

static int HandleCkptCallback(int step, int repairId, std::vector<int> &groupIdx,
                              std::vector<std::vector<int>> &ranks)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->CkptCallback(step, repairId, groupIdx, ranks);
}

static int HandleRepairCallback(RepairContext &context)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->RepairCallback(context);
}

static int HandleUpgradeRepairCallback(RepairContext &context)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->UpgradeRepairCallback(context);
}

static int HandleRollbackCallback(RepairContext &context)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->RollbackCallback(context);
}

static int HandleUpgradeRollbackCallback(RepairContext &context)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->UpgradeRollbackCallback(context);
}

static int HandlePauseCallback(int64_t pauseStep, bool hotSwitch)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->PauseCallback(pauseStep, hotSwitch);
}

static int HandleContinueCallback()
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->ContinueCallback();
}

static int HandleDowngradeRebuildCallback(std::vector<int32_t> &commGroupIdx,
    std::vector<std::vector<int32_t>> &commGroups, int32_t repairId, std::string &param)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->DowngradeRebuildCallback(commGroupIdx, commGroups, repairId, param);
}

static int HandleUpgradeRebuildCallback(std::vector<int32_t> &rankList, int repairId, std::string &param)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->UpgradeRebuildCallback(rankList, repairId, param);
}

static int HandleReportStopCompleteCallback(int code, std::string &msg, std::map<int32_t, int32_t> &errorInfoMap)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->ReportStopCompleteCallback(code, msg, errorInfoMap);
}

static int HandleReportStrategiesCallback(std::map<int32_t, int32_t> &errorInfoMap,
    std::vector<std::string> &strategies)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->ReportStrategiesCallback(errorInfoMap, strategies);
}

static int HandleReportResultCallback(int code, std::string &msg,
    std::map<int32_t, int32_t> errorInfoMap, std::string strategy)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->ReportResultCallback(code, msg, errorInfoMap, strategy);
}

static int HandleReportFaultRanksCallback(std::map<int32_t, int32_t> &errorInfoMap)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->ReportFaultRanksCallback(errorInfoMap);
}

static int HandleRenameCallback()
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->RenameCallback();
}

static void HandleExitCallback()
{
    TTP_ASSERT_RET_VOID(receiver_ptr != nullptr);
    return receiver_ptr->ExitCallback();
}

static int HandleDeviceStopCallback()
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->DeviceStopCallback();
}

static int HandleDeviceCleanCallback()
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->DeviceCleanCallback();
}

static int HandleRegisterCallback()
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->RegisterCallback();
}

static int HandleCommunicationOperateCallback(std::vector<int32_t> &rankList, int repairId)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->CommunicationOperateCallback(rankList, repairId);
}

static int HandleLaunchTcpStoreClientCallback(std::string &url, int rank, int worldSize)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->LaunchTcpStoreClientCallback(url, rank, worldSize);
}

static int HandleLaunchTcpStoreServerCallback(std::string &url, int worldSize)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, TTP_ERROR);
    return receiver_ptr->LaunchTcpStoreServerCallback(url, worldSize);
}

static std::string HandleDecryptCallback(const std::string &cipherText)
{
    TTP_ASSERT_RETURN(receiver_ptr != nullptr, "");
    return receiver_ptr->DecryptCallback(cipherText);
}

// =================================== call controller inner api ===================================

static int InitController(int rank, int worldSize, bool enableLocalCopy, bool enableARF, bool enableZIT)
{
    return Controller::GetInstance()->Initialize(rank, worldSize, enableLocalCopy, enableARF, enableZIT);
}

static int StartController(std::string masterIp, int port, bool enableSsl, std::string tlsInfo)
{
    AccTlsOption tlsOpts;
    if (enableSsl) {
        bool ret = FileUtils::ParseTlsInfo(tlsInfo, tlsOpts);
        if (!ret) {
            TTP_LOG_ERROR("Tls option set error, start controller failed!");
            return TTP_ERROR;
        }
        tlsOpts.enableTls = enableSsl;
    }
    return Controller::GetInstance()->Start(masterIp, port, tlsOpts);
}

static bool QueryHighAvailabilitySwitch()
{
    return Controller::GetInstance()->GetHighAvailabilitySwitch();
}

static int DestroyController()
{
    auto ret = Controller::GetInstance()->Destroy();
    Controller::GetInstance(true);
    MindXEngine::GetInstance()->Destroy();
    MindXEngine::GetInstance(true);
    return ret;
}

// =================================== call processor inner api ====================================

static int InitProcessor(int rank, int32_t worldSize, bool enableLocalCopy, bool enableSsl,
    std::string tlsInfo, bool enableUce, bool enableArf, bool enableZit = false)
{
    AccTlsOption tlsOpts;
    if (enableSsl) {
        bool ret = FileUtils::ParseTlsInfo(tlsInfo, tlsOpts);
        if (!ret) {
            TTP_LOG_ERROR("Tls option set error, init processor failed!");
            return TTP_ERROR;
        }
        tlsOpts.enableTls = enableSsl;
    }
    return Processor::GetInstance()->Initialize(rank,
        worldSize, enableLocalCopy, tlsOpts, enableUce, enableArf, enableZit);
}

static int StartProcessor(std::string masterIp, int32_t port, std::string localIp)
{
    return Processor::GetInstance()->Start(masterIp, port, localIp);
}

static int GetRepairId()
{
    return Processor::GetInstance()->GetRepairId();
}

static bool GetHotSwitch()
{
    return Processor::GetInstance()->GetHotSwitch();
}

static std::string GetRepairType()
{
    return Processor::GetInstance()->GetRepairType();
}

static int SetOptimizerReplica(std::vector<std::vector<int32_t>> rankList,
    std::vector<int32_t> replicaCnt, std::vector<int32_t> replicaShift)
{
    return Processor::GetInstance()->ReportReplicaInfo(rankList, replicaCnt, replicaShift);
}

static int SetDpGroupInfo(std::vector<int32_t> rankList)
{
    return Processor::GetInstance()->ReportDpInfo(rankList);
}

static int SetCopying()
{
    return Processor::GetInstance()->BeginCopying();
}

static int SetUpdating(int64_t backupStep)
{
    return Processor::GetInstance()->BeginUpdating(backupStep);
}

static int ReportLoadCkptStep(int64_t loadCkptStep)
{
    return Processor::GetInstance()->ReportLoadCkptStep(loadCkptStep);
}

static int ResetLimitStep()
{
    return Processor::GetInstance()->ResetLimitStep();
}

static int SetFinished(int64_t step)
{
    return Processor::GetInstance()->FinishedUpdate(step);
}

static int SetDumpResult(int result)
{
    return Processor::GetInstance()->SetDumpResult(result);
}

static int ReportStatus(int32_t state)
{
    return Processor::GetInstance()->ReportStatus(static_cast<ReportState>(state));
}

static int WaitNextAction()
{
    return Processor::GetInstance()->WaitNextAction();
}

static int WaitRepairAction()
{
    return Processor::GetInstance()->WaitRepairAction();
}

static int DestroyProcessor()
{
    Processor::GetInstance()->Destroy();
    Processor::GetInstance(true);
    return 0;
}

// ================================== register processor callback ==================================

static int SetCkptCallback()
{
    CkptCallback callback = &HandleCkptCallback;
    ProcessorEventHandle func = [callback](void* ctx, int ctxSize) -> int {
        SaveCkptContext *scc = static_cast<SaveCkptContext *>(ctx);
        int ret = callback(scc->step, scc->repairId, scc->groupIdx, scc->ranks);
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_SAVE_CKPT, func);
}

static int SetRenameCallback()
{
    RenameCallback callback = &HandleRenameCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        int ret = callback();
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_RENAME, func);
}

static int SetExitCallback()
{
    ExitCallback callback = &HandleExitCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        callback();
        return 0;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_EXIT, func);
}

static int SetDeviceStopCallback()
{
    DeviceStopCallback callback = &HandleDeviceStopCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        int ret = callback();
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_DEVICE_STOP, func);
}

static int SetDeviceCleanCallback()
{
    DeviceCleanCallback callback = &HandleDeviceCleanCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        int ret = callback();
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_DEVICE_CLEAN, func);
}

static int SetRepairCallback()
{
    RepairCallback callback = &HandleRepairCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        RepairContext *rc = static_cast<RepairContext *>(ctx);
        int ret = callback(*rc);
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_REPAIR, func);
}

static int SetZitUpgradeRepairCallback()
{
    RepairCallback callback = &HandleUpgradeRepairCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        RepairContext *rc = static_cast<RepairContext *>(ctx);
        int ret = callback(*rc);
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_UPGRADE_REPAIR, func);
}

static int SetRollbackCallback()
{
    RollbackCallback callback = &HandleRollbackCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        RepairContext *rc = static_cast<RepairContext *>(ctx);
        int ret = callback(*rc);
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_ROLLBACK, func);
}

static int SetZitUpgradeRollbackCallback()
{
    RollbackCallback callback = &HandleUpgradeRollbackCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        RepairContext *rc = static_cast<RepairContext *>(ctx);
        int ret = callback(*rc);
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_UPGRADE_ROLLBACK, func);
}

static int SetPauseCallback()
{
    PauseCallback callback = &HandlePauseCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        PauseContext *pc = static_cast<PauseContext *>(ctx);
        int ret = callback(pc->step, pc->hotSwitch);
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_PAUSE, func);
}

static int SetContinueCallback()
{
    ContinueCallback callback = &HandleContinueCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        int ret = callback();
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_CONTINUE, func);
}

static int SetCommunicationOperateCallback()
{
    CommunicationOperateCallback callback = &HandleCommunicationOperateCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        RebuildGroupMsg *msg = static_cast<RebuildGroupMsg *>(ctx);
        std::vector<int32_t> rankList(msg->ranks, msg->ranks + msg->rankNum);
        int ret = callback(rankList, msg->repairId);
        return ret;
    };
    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_PT_COMM_OPERATE, func);
}

static int SetZitDowngradeRebuildCallback()
{
    DowngradeRebuildCallback callback = &HandleDowngradeRebuildCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        ZitRebuildContext *rc = static_cast<ZitRebuildContext *>(ctx);
        int ret = callback(rc->commGroupIdx, rc->commGroups, rc->repairId, rc->zitParam);
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_DOWNGRADE_REBUILD, func);
}

static int SetZitUpgradeRebuildCallback()
{
    UpgradeRebuildCallback callback = &HandleUpgradeRebuildCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        RebuildGroupMsg *msg = static_cast<RebuildGroupMsg *>(ctx);
        std::vector<int32_t> rankList(msg->ranks, msg->ranks + msg->rankNum);
        int32_t zitParamLen = msg->ranks[msg->rankNum];
        std::string zitParam = std::string(reinterpret_cast<char*>(&msg->ranks[msg->rankNum + 1]));
        TTP_ASSERT_RETURN(zitParamLen == zitParam.length() + 1, TTP_ERROR);
        int ret = callback(rankList, msg->repairId, zitParam);
        return ret;
    };
    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_UPGRADE_REBUILD, func);
}

static int SetLaunchTcpStoreClientCallback()
{
    LaunchTcpStoreClientCallback callback = &HandleLaunchTcpStoreClientCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        LaunchTcpStoreMsg *msg = static_cast<LaunchTcpStoreMsg *>(ctx);
        int ret = callback(msg->url, msg->rank, msg->worldSize);
        return ret;
    };
    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_LAUNCH_TCP_STORE_CLIENT, func);
}

static int SetLaunchTcpStoreServerCallback()
{
    LaunchTcpStoreServerCallback callback = &HandleLaunchTcpStoreServerCallback;
    ProcessorEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        LaunchTcpStoreMsg *msg = static_cast<LaunchTcpStoreMsg *>(ctx);
        int ret = callback(msg->url, msg->worldSize);
        return ret;
    };

    return Processor::GetInstance()->RegisterEventHandler(PROCESSOR_EVENT_LAUNCH_TCP_STORE_SERVER, func);
}

// ================================= register mindxengine callback =================================

static int SetRegisterCallback()
{
    RegisterCallback callback = &HandleRegisterCallback;
    MindXEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        int ret = callback();
        return ret;
    };

    return MindXEngine::GetInstance()->RegisterEventHandler(MindXEvent::MINDX_EVENT_REGISTER, func);
}

static int SetReportStopCompleteCallback()
{
    ReportStopCompleteCallback callback = &HandleReportStopCompleteCallback;
    MindXEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        StopCompleteContext *nrsc = static_cast<StopCompleteContext *>(ctx);
        int ret = callback(nrsc->code, nrsc->msg, nrsc->errorInfoMap);
        return ret;
    };

    return MindXEngine::GetInstance()->RegisterEventHandler(MindXEvent::MINDX_EVENT_REPORT_STOP_COMPLETE, func);
}

static int SetReportStrategiesCallback()
{
    ReportStrategiesCallback callback = &HandleReportStrategiesCallback;
    MindXEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        RecoverStrategyContext *nrsc = static_cast<RecoverStrategyContext *>(ctx);
        int ret = callback(nrsc->errorInfoMap, nrsc->strategies);
        return ret;
    };

    return MindXEngine::GetInstance()->RegisterEventHandler(MindXEvent::MINDX_EVENT_REPORT_STRATEGIES, func);
}

static int SetReportResultCallback()
{
    ReportResultCallback callback = &HandleReportResultCallback;
    MindXEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        RecoverStatusContext *nrsc = static_cast<RecoverStatusContext *>(ctx);
        int ret = callback(nrsc->code, nrsc->msg, nrsc->errorInfoMap, nrsc->strategy);
        return ret;
    };

    return MindXEngine::GetInstance()->RegisterEventHandler(MindXEvent::MINDX_EVENT_REPORT_RESULT, func);
}

static int SetReportFaultRanksCallback()
{
    ReportFaultRanksCallback callback = &HandleReportFaultRanksCallback;
    MindXEventHandle func = [callback](void *ctx, int ctxSize) -> int {
        ProcessFaultContext *nrsc = static_cast<ProcessFaultContext *>(ctx);
        int ret = callback(nrsc->errorInfoMap);
        return ret;
    };

    return MindXEngine::GetInstance()->RegisterEventHandler(MindXEvent::MINDX_EVENT_REPORT_FAULT_RANKS, func);
}

// ======================================== mindx call api =========================================

static int MindxStopTrainCallback(std::map<int32_t, int32_t> rankList)
{
    TTP_ASSERT_RETURN(rankList.size() <= TTP_MAX_WORLD_SIZE, TTP_ERROR);
    return MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_STOP_TRAIN, &rankList, rankList.size());
}

static int MindxPauseTrainCallback(uint32_t timeout)
{
    return MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_PAUSE_TRAIN, &timeout, sizeof(timeout));
}

static int MindxNotifyFaultRanksCallback(std::map<int32_t, int32_t> rankList, uint32_t hcclTime)
{
    NotifyRankInfo rankInfo {rankList, hcclTime};
    return MindXEngine::GetInstance()->EventProcess(MindXEvent::MINDX_EVENT_NOTIFY_FAULT_RANKS,
                                                    &rankInfo, sizeof(NotifyRankInfo));
}

static int MindxPrepareActionCallback(std::string action, std::map<int32_t, int32_t> faultRanks)
{
    return MindXEngine::GetInstance()->PrepareAction(action, faultRanks);
}

static int MindxChangeStrategyCallback(std::string strategy, std::string param = "")
{
    return MindXEngine::GetInstance()->ChangeStrategy(strategy, param);
}

static int MindxNotifyDumpCallback()
{
    return MindXEngine::GetInstance()->MindXNotifyDump();
}

static void Log(int level, std::string msg)
{
    OutLogger::Instance()->Log(level, std::ostringstream(msg));
}

static int SetDecryptCallback()
{
    DecryptCallback callback = &HandleDecryptCallback;
    AccDecryptHandler func = [callback](const std::string &cipherText,
        char *plainText, size_t &plainTextLen) -> int {
        if (cipherText.size() > MAX_CIPHER_LEN) {
            TTP_LOG_ERROR("input cipher len is too long");
            return TTP_ERROR;
        }
        try {
            std::string plain = callback(cipherText);
            if (plain.size() >= plainTextLen) {
                TTP_LOG_ERROR("output cipher len is too long");
                std::fill(plain.begin(), plain.end(), 0);
                return TTP_ERROR;
            }

            std::copy(plain.begin(), plain.end(), plainText);
            plainText[plain.size()] = '\0';
            plainTextLen = plain.size();
            std::fill(plain.begin(), plain.end(), 0);
            return 0;
        } catch (...) {
            return TTP_ERROR;
        }
    };

    AccTcpServer::RegisterDecryptHandler(func);
    AccTcpClient::RegisterDecryptHandler(func);

    return TTP_OK;
}
