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
#include "controller.h"
#include <chrono>
#include <cstring>
#include <netdb.h>
#include <arpa/inet.h>
#include <dirent.h>
#include <set>
#include <sys/stat.h>
#include <algorithm>
#include "mindx_engine.h"
#include "ttp_logger.h"

namespace ock {
namespace ttp {

using GroupRecord = std::pair<int32_t, std::vector<int32_t>>;

constexpr uint32_t MAX_MSG_LEN = TTP_MAX_WORLD_SIZE * 8 + 100;
constexpr uint32_t HEART_BEAT_MAX_LOSS = 3;
constexpr uint32_t TIME_CHECKER_INTERVAL = 2000;  // unit:ms
constexpr uint32_t BACKUP_CONTROLLER_NUM = 2;
constexpr uint32_t SLEEP_FOR_PROCESSOR_CONNECT = 4;

constexpr uint32_t PRELOCK_RETRY_TIME = 3;
constexpr uint32_t ARF_WAIT_ADD_TIME = 10 * 60 * 1000;
constexpr uint16_t CONTROLLER_SN_GENERATION = 20000;

constexpr int32_t UCE_ERROR = 0;
constexpr int32_t PROCESSES_ERROR = 1;
constexpr int32_t HCCL_ERROR = 2;

constexpr EnvVarValue WaitMindxTime = {.minVal = 1, .maxVal = 3600, .defaultVal = 30};
constexpr EnvVarValue PortInfo = {.minVal = 1024, .maxVal = 65535, .defaultVal = 6000};

inline int64_t GetNowTime()
{
    auto now = std::chrono::system_clock::now();
    auto now_ms = std::chrono::time_point_cast<std::chrono::milliseconds>(now);
    auto value = now_ms.time_since_epoch().count();
    return static_cast<int64_t>(value);
}

#define STATUS_MAP_VAL_PRINT(status)   " step: " << (status).step \
                                    << " npu_status: " << static_cast<int32_t>((status).npu_status) \
                                    << " run_status: " << static_cast<int32_t>((status).run_status) \
                                    << " data_aval: " << static_cast<int32_t>((status).data_aval) \
                                    << " data_status: " << static_cast<int32_t>((status).data_status) \
                                    << " diff_time : " << (GetNowTime() - ((status).lastUpdateTime))


Controller::Controller() {};

ControllerPtr Controller::GetInstance(bool destroy)
{
    static std::mutex gMutex;
    static ControllerPtr gInstance;

    if (gInstance == nullptr) {
        std::lock_guard<std::mutex> guard(gMutex);
        if (gInstance.Get() == nullptr) {
            // logger must not nullptr
            if (OutLogger::Instance() == nullptr) {
                throw std::bad_alloc();
            }

            gInstance = MakeRef<Controller>();
            if (gInstance == nullptr) {
                TTP_LOG_ERROR("Create controller failed, out of memory");
                throw std::bad_alloc();
            }
        }
    } else if (destroy) {
        std::lock_guard<std::mutex> guard(gMutex);
        gInstance = nullptr;
    }

    return gInstance;
}

// init
TResult Controller::Initialize(int32_t rank, int32_t worldSize, bool enableLocalCopy, bool enableARF, bool enableZIT)
{
    TTPLogger::Init();
    TTP_LOG_DEBUG("Start to init controller, rank:" << rank);

    TTP_ASSERT_RETURN(rank >= -1, TTP_ERROR);
    TTP_ASSERT_RETURN(worldSize > 0 && worldSize <= TTP_MAX_WORLD_SIZE, TTP_ERROR);
    TTP_ASSERT_RETURN(rank < worldSize, TTP_ERROR);
    mindSpore_ = GetMsEnv("MINDIO_FOR_MINDSPORE");
    if (mindSpore_ && (enableLocalCopy || enableZIT)) {
        TTP_LOG_ERROR("MS: does not support localCopy/ZIT");
        return TTP_ERROR;
    }

    std::lock_guard<std::mutex> guard(initOrDestroyMutex_);
    if (isInited_.load()) { /* already inited_ */
        TTP_LOG_INFO("controller has already inited, rank:" << rank);
        return TTP_OK;
    }

    rankToRename_ = rank;
    rank_ = rank;
    worldSize_ = worldSize;
    localCopySwitch_ = enableLocalCopy;
    arfSwitch_ = enableARF;
    zitSwitch_ = enableZIT;

    TResult ret = TcpServerInit();
    if (ret != TTP_OK) { return ret; }

    ret = ActionEngineInit();
    if (ret != TTP_OK) { return ret; }

    if (rank == 0 || rank == -1) {
        isMasterCtrl_.store(true);
        if (rank == -1) {
            isSupportBackupToMaster_.store(false);
        }
    }
    isInited_.store(true);
    repairId_.store(0);
    ret = StateMachineInit();
    if (ret != TTP_OK) { return ret; }

    mindXEngine_ = MindXEngine::GetInstance();
    waitMindxTimes_ = GetEnvValue2Uint32("MINDIO_WAIT_MINDX_TIME",
    WaitMindxTime.minVal, WaitMindxTime.maxVal, WaitMindxTime.defaultVal);
    TTP_LOG_DEBUG("[env] MINDIO_WAIT_MINDX_TIME:" << waitMindxTimes_);

    TTP_LOG_INFO("Init controller success, rank:" << rank_ << ", world size:" << worldSize_ \
        << ", is master controller:" << isMasterCtrl_.load() << ", arf:" << arfSwitch_ << ", zit:" << zitSwitch_);
    return TTP_OK;
}

TResult Controller::TcpServerInit()
{
    mServer_ = AccTcpServer::Create();
    if (mServer_ == nullptr) {
        TTP_LOG_ERROR("controller:" << rank_ << " initialize AccTcpServer failed");
        return TTP_ERROR;
    }

    // add server handler
    auto hbMethod = [this](const AccTcpRequestContext &context) { return HandleHeartBeat(context); };
    mServer_->RegisterNewRequestHandler(TTP_MSG_OP_HEARTBEAT_SEND, hbMethod);
    auto registerMethod = [this](const AccTcpRequestContext &context) { return HandleRegister(context); };
    mServer_->RegisterNewRequestHandler(TTP_MSG_OP_REGISTER, registerMethod);
    auto reportMethod = [this](const AccTcpRequestContext &context) { return HandleReportInfo(context); };
    mServer_->RegisterNewRequestHandler(TTP_MSG_OP_INIT_REPORT, reportMethod);
    auto dpMethod = [this](const AccTcpRequestContext &context) { return HandleReportDp(context); };
    mServer_->RegisterNewRequestHandler(TTP_MSG_OP_DP_REPORT, dpMethod);

    // add link handle
    auto linkMethod = [this](const AccConnReq &req, const AccTcpLinkComplexPtr &link) {
        return HandleNewConnection(req, link);
    };
    mServer_->RegisterNewLinkHandler(linkMethod);

    auto linkBrokenMethod = [this](const AccTcpLinkComplexPtr &link) { return HandleLinkBroken(link); };
    mServer_->RegisterLinkBrokenHandler(linkBrokenMethod);

    return TTP_OK;
}

TResult Controller::ActionEngineInit()
{
    TTP_ASSERT_RETURN(worldSize_ > 0 && worldSize_ <= TTP_MAX_WORLD_SIZE, TTP_ERROR);
    engine_ = MakeRef<ActionEngine>();
    if (engine_ == nullptr) {
        TTP_LOG_ERROR("controller:" << rank_ << " create ActionEngine failed");
        return TTP_ERROR;
    }

    auto sendMsgMethod = [this](int16_t msgType, const AccDataBufferPtr &d,
            std::vector<int32_t> &targetRanks, const std::vector<AccDataBufferPtr> &cbCtx) {
        return SendMsg(msgType, d, targetRanks, cbCtx);
    };
    TResult ret = engine_->Initialize(mServer_, sendMsgMethod, worldSize_);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " initialize ActionEngine failed");
        return ret;
    }

    auto collectionReply = [this](const AccTcpRequestContext &context) { return HandleCollectionReply(context); };
    engine_->ReplyRegister(TTP_MSG_OP_COLLECTION_REPLY, collectionReply);
    auto prelockReply = [this](const AccTcpRequestContext &context) { return HandlePrelockReply(context); };
    engine_->ReplyRegister(TTP_MSG_OP_PRELOCK_REPLY, prelockReply);
    auto notifyReply = [this](const AccTcpRequestContext &context) { return HandleNotifyNormalReply(context); };
    engine_->ReplyRegister(TTP_MSG_OP_NORMAL_REPLY, notifyReply);
    auto cleanReply = [this](const AccTcpRequestContext &context) { return HandleCleanReply(context); };
    engine_->ReplyRegister(TTP_MSG_OP_CLEAN_REPLY, cleanReply);

    auto prelockExtraParseReply = [this](const AccTcpRequestContext &context, TTPReplyMsg &msg) {
        return PrelockResultAndHbReplyParse(context, msg);
    };
    engine_->ReplyParseRegister(TTP_MSG_OP_PRELOCK_REPLY, prelockExtraParseReply);

    // Extra Reply Parse Handle for sn check; if reply format is NOT TTPReplyMsg, need register.
    auto extraParseReply = [this](const AccTcpRequestContext &context, TTPReplyMsg &msg) {
        return ResultAndHbReplyParse(context, msg);
    };
    engine_->ReplyParseRegister(TTP_MSG_OP_COLLECTION_REPLY, extraParseReply);
    engine_->ReplyParseRegister(TTP_MSG_OP_NORMAL_REPLY, extraParseReply);

    auto statusMakrMethod = [this](const std::vector<int32_t> &ranks) { return MarkNoReponseRanks(ranks); };
    engine_->RankStatusRegister(statusMakrMethod);
    return TTP_OK;
}

TResult Controller::StateMachineInit()
{
    TResult ret = pthreadTimeChecker_ .Initialize();
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("init pthreadTimeChecker_ failed, rank:" << rank_);
        return ret;
    }

    stateMachine_ = MakeRef<controllerStateMachine>();
    if (stateMachine_ == nullptr) {
        TTP_LOG_ERROR("controller:" << rank_ << " create stateMachine_ failed");
        return TTP_ERROR;
    }

    ret = stateMachine_->Initialize(rank_);
    if (ret != TTP_OK) {
        TTP_LOG_INFO("init controller stateMachine_ failed, rank:" << rank_ <<
                     " is master controller: " << isMasterCtrl_.load());
        return ret;
    }

    // add stateMachine_ handle
    auto initMethod = [this]() { return InitCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_INIT, initMethod);

    auto normalMethod = [this]() { return NormalCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_NORMAL, normalMethod);

    auto pauseMethod = [this]() { return PauseCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_PAUSE, pauseMethod);

    auto stepFinishMethod = [this]() { return StepFinishCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_STEP_FINISH, stepFinishMethod);

    auto migrationMethod = [this]() { return MigrationCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_MIGRATION, migrationMethod);

    auto abnormalMethod = [this]() { return AbnormalCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_ABNORMAL, abnormalMethod);

    auto dumpMethod = [this]() { return DumpCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_DUMP, dumpMethod);

    auto exitMethod = [this]() { return ExitCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_EXIT, exitMethod);

    auto clearMethod = [this]() { return EnvClearCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_ENV_CLEAR, clearMethod);

    auto repairMethod = [this]() { return RepairCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_REPAIR, repairMethod);

    auto dgRepairMethod = [this]() { return DowngradeRepairCallback(); };
    stateMachine_->ControllerActionRegister(STATE_OP_DG_REPAIR, dgRepairMethod);

    auto dgMethod = [this]() { return DowngradeRunning(); };
    stateMachine_->ControllerActionRegister(STATE_OP_DOWNGRADE, dgMethod);

    auto ugMethod = [this]() { return UpgradeRunning(); };
    stateMachine_->ControllerActionRegister(STATE_OP_UPGRADE, ugMethod);

    return TTP_OK;
}

// init controller
TResult Controller::Start(std::string &masterIp, int32_t port, const AccTlsOption &tlsOpts, uint32_t controllerIdx)
{
    TTP_ASSERT_RETURN(IsValidIpV4(masterIp), TTP_ERROR);
    TTP_ASSERT_RETURN(controllerIdx >= 0 && controllerIdx <= BACKUP_CONTROLLER_NUM, TTP_ERROR);
    TTP_ASSERT_RETURN(port >= PortInfo.minVal && port <= PortInfo.maxVal, TTP_ERROR);
    std::lock_guard<std::mutex> guard(initOrDestroyMutex_);
    if (!isInited_.load()) {
        TTP_LOG_ERROR("controller not init, master ip: " << masterIp << ", port: " << port);
        return TTP_ERROR;
    }
    if (isStarted_.load()) {
        TTP_LOG_INFO("controller has started, ip: " << masterIp << ", port: " << port);
        return TTP_OK;
    }
    TTP_LOG_INFO("Begin start server, master ip: " << masterIp << ", port: " << port);  // 补充日志信息
    AccTcpServerOptions opts; // 1.start net server & bind port & listen
    opts.enableListener = true;
    opts.linkSendQueueSize = TTP_LINK_SEND_QUEUE_SIZE;
    opts.listenIp = masterIp;
    opts.listenPort = port;
    opts.reusePort = true;
    opts.magic = 0;
    opts.version = TTP_DEFAULT_START_VERSION;
    opts.workerCount = TTP_SERVER_WORKER_COUNT;
    opts.maxWorldSize = TTP_MAX_WORLD_SIZE;

    controllerIp_ = masterIp;
    controllerPort_ = port;
    controllerIdx_ = controllerIdx;

    if (tlsOpts.enableTls) {
        auto ret = mServer_->LoadDynamicLib(tlsOpts.packagePath);
        if (ret != ACC_OK) {
            TTP_LOG_ERROR("Controller:" << rank_ << " load ssl dynamic lib failed");
            return TTP_ERROR;
        }
    }
    if (mServer_->Start(opts, tlsOpts) != TTP_OK) {
        TTP_LOG_ERROR("controller start server failed, master ip: " << masterIp << " port: " << port);
        return TTP_ERROR;
    }

    isStarted_.store(false);
    isStopped_.store(false);
    stateMachine_->StartStateMachine(); // 2.start state machine thread
    while (!isStarted_.load()) {
        TTP_LOG_LIMIT_INFO(LOG_PRINT_INTERVAL, "sleep for started, isStarted: " << isStarted_.load());
        usleep(TTP_WAIT_TIME_1MS);
    }
    TTP_LOG_INFO("Start controller success! ip: " << masterIp << ", port: " << port);
    return TTP_OK;
}

TResult Controller::InitCallback()
{
    isAlreadyBrod_.store(false);
    isBackupToMaster_.store(false);
    errorRankMsg_.clear();

    if (isMasterCtrl_.load()) {
        mindXEngine_->Register2MindX();
    }

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_STOP_TRAIN,
        [this] (void *ctx, int ctxSize) { return MindXNotifyStopTrain(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_PAUSE_TRAIN,
        [this] (void *ctx, int ctxSize) { return MindXNotifyPauseTrain(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_CONTINUE_TRAIN,
        [this] (void *ctx, int ctxSize) { return MindXNotifyContinueTrain(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_NOTIFY_FAULT_RANKS,
        [this] (void *ctx, int ctxSize) { return MindXNotifyFaultRanks(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_HOT_SWITCH,
        [this] (void *ctx, int ctxSize) { return MindXNotifyHotSwitch(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_STOP_SWITCH,
        [this] (void *ctx, int ctxSize) { return MindXNotifyStopSwitch(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_MIGRATION,
        [this] (void *ctx, int ctxSize) { return MindXNotifyMigration(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_DOWNGRADE,
        [this] (void *ctx, int ctxSize) { return MindXNotifyDownGradeRepair(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_UPGRADE,
        [this] (void *ctx, int ctxSize) { return MindXNotifyUpGradeRepair(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_ARF,
        [this] (void *ctx, int ctxSize) { return MindXNotifyArfRepair(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_UCE,
        [this] (void *ctx, int ctxSize) { return MindXNotifyUceRepair(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_DUMP,
        [this] (void *ctx, int ctxSize) { return MindXNotifyDump(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_EXIT,
        [this] (void *ctx, int ctxSize) { return MindXNotifyExit(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_INVALID,
        [this] (void *ctx, int ctxSize) { return MindXInvalidNotify(ctx, ctxSize); });

    mindXEngine_->RegisterEventHandler(MindXEvent::MINDX_EVENT_ELEGANT_DUMP,
        [this] (void *ctx, int ctxSize) { return MindXNotifyElegantDump(ctx, ctxSize); });

    return TTP_OK;
}

TResult Controller::ErrorRankMsgModify(std::map<int32_t, int32_t> &errorRankMap, std::string option)
{
    AutoLock statusMapLock(statusMapLock_, TYPE_READ);
    AutoLock errorRankLock(errorRankLock_, TYPE_WRITE);

    std::string msg = "[";
    for (auto [rank, errType] : errorRankMap) {
        TTP_ASSERT_RETURN(rank >= 0 && rank < worldSize_, TTP_ERROR);
        TTP_ASSERT_RETURN(errType >= UCE_ERROR && errType <= HCCL_ERROR, TTP_ERROR);
        TTP_ASSERT_RETURN(statusMap_.find(rank) != statusMap_.end(), TTP_ERROR);

        if (errorRankMsg_.count(rank) > 0 && errorRankMsg_[rank] == PROCESSES_ERROR) {
            continue;
        }
        errorRankMsg_[rank] = errType;
        msg += std::to_string(rank) + ":" + std::to_string(errType) + ", ";
    }
    msg += "]";
    TTP_LOG_INFO("Mindx notify " <<  option << ", error_rank: " << msg);
    return TTP_OK;
}

TResult Controller::MindXNotifyStopTrain(void *ctx, int32_t ctxSize)
{
    TTP_ASSERT_RETURN(ctx != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(ctxSize <= worldSize_, TTP_ERROR);

    auto rankList = static_cast<std::map<int32_t, int32_t> *>(ctx);
    TResult result = ErrorRankMsgModify(*rankList, "stop train");
    if (result != TTP_OK) {
        TTP_LOG_ERROR("handle mindx notify stop train failed");
        return result;
    }

    repairEvent_ = MindXEvent::MINDX_EVENT_STOP_TRAIN;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify stop train, ret:" << ret);

    pthreadTimeChecker_ .PthreadSignal();
    return ret;
}

TResult Controller::MindXNotifyFaultRanks(void *ctx, int32_t ctxSize)
{
    TTP_ASSERT_RETURN(ctx != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(ctxSize == sizeof(NotifyRankInfo), TTP_ERROR);

    NotifyRankInfo *rankInfo = static_cast<NotifyRankInfo *>(ctx);
    auto &[rankList, hcclTime] = *rankInfo;
    TTP_ASSERT_RETURN(rankList.size() <= TTP_MAX_WORLD_SIZE, TTP_ERROR);
    TTP_ASSERT_RETURN(0 < hcclTime && hcclTime <= WaitMindxTime.maxVal, TTP_ERROR);

    waitHcclTime_ = hcclTime;
    TResult result = ErrorRankMsgModify(rankList, "fault ranks");
    if (result != TTP_OK) {
        TTP_LOG_ERROR("handle mindx notify fault ranks failed");
        return result;
    }

    repairEvent_ = MindXEvent::MINDX_EVENT_NOTIFY_FAULT_RANKS;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify fault ranks, ret:" << ret);

    return ret;
}

TResult Controller::MindXNotifyDump(void *ctx, int32_t ctxSize)
{
    repairEvent_ = MindXEvent::MINDX_EVENT_DUMP;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify dump, ret:" << ret);

    return ret;
}

TResult Controller::MindXNotifyElegantDump(void *ctx, int32_t ctxSize)
{
    repairEvent_ = MindXEvent::MINDX_EVENT_ELEGANT_DUMP;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify elegant dump, ret:" << ret);

    pthreadTimeChecker_ .PthreadSignal();
    return ret;
}

TResult Controller::MindXNotifyPauseTrain(void *ctx, int32_t ctxSize)
{
    TTP_ASSERT_RETURN(ctx != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(ctxSize == sizeof(uint32_t), TTP_ERROR);

    uint32_t* timeout = static_cast<uint32_t*>(ctx);
    waitPauseTimes_ = *timeout;
    repairEvent_ = MindXEvent::MINDX_EVENT_PAUSE_TRAIN;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify pause train, ret:" << ret);

    pthreadTimeChecker_ .PthreadSignal();
    return ret;
}

TResult Controller::MindXNotifyContinueTrain(void *ctx, int32_t ctxSize)
{
    repairEvent_ = MindXEvent::MINDX_EVENT_CONTINUE_TRAIN;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify continue train, ret:" << ret);

    pthreadTimeChecker_ .PthreadSignal();
    return ret;
}

TResult Controller::MindXNotifyHotSwitch(void *ctx, int32_t ctxSize)
{
    TTP_ASSERT_RETURN(ctx != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(ctxSize <= worldSize_, TTP_ERROR);

    auto faultRanks = static_cast<std::map<int32_t, int32_t> *>(ctx);
    hotSwitchRanks_ = GetMapKeysToSet(*faultRanks);
    repairId_.fetch_add(1);
    repairEvent_ = MindXEvent::MINDX_EVENT_HOT_SWITCH;

    pthreadTimeChecker_ .PthreadSignal(); // for NormalCallback()
    return TTP_OK;
}

TResult Controller::MindXNotifyStopSwitch(void *ctx, int32_t ctxSize)
{
    hotSwitchRanks_.clear();
    repairEvent_ = MindXEvent::MINDX_EVENT_STOP_SWITCH;

    auto ret = mindXEngine_->WakeUp(); // for StepFinishCallback()
    TTP_RET_LOG(ret, "mindx notify stop switch, ret:" << ret);

    pthreadTimeChecker_ .PthreadSignal(); // for NormalCallback()
    return ret;
}

TResult Controller::MindXNotifyMigration(void *ctx, int32_t ctxSize)
{
    repairEvent_ = MindXEvent::MINDX_EVENT_MIGRATION;

    auto ret = mindXEngine_->WakeUp(); // for StepFinishCallback()
    TTP_RET_LOG(ret, "mindx notify migration, ret:" << ret);

    return ret;
}

TResult Controller::MindXNotifyArfRepair(void *ctx, int32_t ctxSize)
{
    repairEvent_ = MindXEvent::MINDX_EVENT_ARF;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify arf, ret:" << ret);

    return ret;
}

TResult Controller::MindXNotifyUceRepair(void *ctx, int32_t ctxSize)
{
    repairEvent_ = MindXEvent::MINDX_EVENT_UCE;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify uce, ret:" << ret);

    return ret;
}

TResult Controller::MindXNotifyDownGradeRepair(void *ctx, int32_t ctxSize)
{
    TTP_ASSERT_RETURN(ctx != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(ctxSize > 0 && ctxSize < TTP_MAX_ZIT_PARAM_LEN, TTP_ERROR);
    std::string* inputStr = static_cast<std::string*>(ctx);

    zitParam_.strategyParm = *inputStr;
    TTP_ASSERT_RETURN(zitParam_.strategyParm.length() == ctxSize, TTP_ERROR);
    repairEvent_ = MindXEvent::MINDX_EVENT_DOWNGRADE;
    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify down grade, ret:" << ret);

    return ret;
}

TResult Controller::MindXNotifyUpGradeRepair(void *ctx, int32_t ctxSize)
{
    TTP_ASSERT_RETURN(ctx != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(ctxSize > 0 && ctxSize < TTP_MAX_ZIT_PARAM_LEN, TTP_ERROR);
    std::string* inputStr = static_cast<std::string*>(ctx);

    zitParam_.strategyParm = *inputStr;
    TTP_ASSERT_RETURN(zitParam_.strategyParm.length() == ctxSize, TTP_ERROR);

    repairEvent_ = MindXEvent::MINDX_EVENT_UPGRADE;

    return TTP_OK;
}

TResult Controller::MindXNotifyExit(void *ctx, int32_t ctxSize)
{
    repairEvent_ = MindXEvent::MINDX_EVENT_EXIT;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify exit, ret:" << ret);

    return ret;
}

TResult Controller::MindXInvalidNotify(void *ctx, int32_t ctxSize)
{
    repairEvent_ = MindXEvent::MINDX_EVENT_INVALID;

    auto ret = mindXEngine_->WakeUp();
    TTP_RET_LOG(ret, "mindx notify invalid, ret:" << ret);

    return ret;
}

bool Controller::GetHighAvailabilitySwitch()
{
    AutoLock lock(linkMapLock_, TYPE_READ);
    return !rankLinkMap_.empty();
}

void Controller::InitializeVariables()
{
    isStarted_.store(true);
    repairType_ = ControllerRepairType::CRT_BUTT;
    canRetryCleanFlag_.store(false);
    repairEvent_ = MindXEvent::MINDX_EVENT_BUTT;
    unableRepair_ = false;
    hcclFlag_ = 0;
    if (loadCkptRepairStep_.load() > 0 && loadCkptRepairStep_.load() < INT64_MAX) {
        repairStep_ = loadCkptRepairStep_.load();
    }
    zitParam_.strategyParm = "";
    zitParam_.isolateRanks.clear();
    AutoLock errorRankLock(errorRankLock_, TYPE_WRITE);
    errorRankMsg_.clear();

    hotSwitchRanks_.clear();
    statusMapTmp_.clear();
    rankLinkMapTmp_.clear();
    linkIdMapTmp_.clear();
}

TResult Controller::NormalCallback()
{
    InitializeVariables();
    TResult ret = TTP_ERROR;

    auto checkMindxEvent = [this] () {
        switch (repairEvent_) {
            case MindXEvent::MINDX_EVENT_STOP_TRAIN:
            case MindXEvent::MINDX_EVENT_ELEGANT_DUMP:
                return TTP_ERROR;
            case MindXEvent::MINDX_EVENT_PAUSE_TRAIN:
                return TTP_PAUSE;
            case MindXEvent::MINDX_EVENT_HOT_SWITCH:
                return CheckHotSwitchRegister() ? TTP_SWITCH : TTP_OK;
            default:
                return TTP_OK;
        }
    };

    while (!isStopped_.load()) {
        ret = TTP_OK;
        if (isBackupToMaster_.load() || IsBackupToMaster()) {
            ret = CheckTrainStatus();
        } else if (isMasterCtrl_.load()) {
            // if mindx is on, no backup controller, skip broadcast
            if (!isAlreadyBrod_.load() && isSupportBackupToMaster_.load()) {
                isAlreadyBrod_.store(BroadcastCrtlIps() == TTP_OK);
            }
            ret = CheckTrainStatus();
        }

        // 1.先检查是否有Mindx主动通知跳转
        auto eventRet = checkMindxEvent();
        if (eventRet != TTP_OK) {
            TTP_LOG_WARN("NormalCallback found mindx event:" << static_cast<int>(repairEvent_) << ", ret:" << eventRet);
            return eventRet;
        }

        // 2.检查完Mindx通知后，检查statusMap_状态信息
        if (ret == TTP_ERROR) {
#ifndef UT_ENABLED
            sleep(TTP_SLEEP_TIME);
            auto againRet = CheckTrainStatus();
            TTP_LOG_WARN("NormalCallback found abnormal, ret: " << ret << ", againRet: " << againRet);
#endif
            return ret;
        }

        pthreadTimeChecker_.PthreadTimedwaitSecs(TTP_SLEEP_TIME);
    }

    if (isStopped_.load()) { // forced exit
        return TTP_STOP_SERVICE;
    }
    return ret;
}

bool Controller::CheckHotSwitchRegister()
{
    AutoLock statusMapLock(statusMapLock_, TYPE_READ);
    // 检查新热切节点注册情况
    for (auto rank : hotSwitchRanks_) {
        auto it = statusMapTmp_.find(rank);
        if (it == statusMapTmp_.end() ||
            it->second.run_status !=
            TTP_STATUS_PREREPAIR_FINISH) {
            return false;
        }
    }
    return true;
}

TResult Controller::StepFinishCallback()
{
    TResult ret = PauseTrain();
    TTP_RET_LOG(ret, "controller pause train ret:" << ret);
    std::vector<std::string> strategy {ret == TTP_OK ? STRATEGY_MIGRATION : STRATEGY_EXIT};
    ret = mindXEngine_->ReportStrategies(strategy, errorRankMsg_, errorRankLock_);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }
    ret = mindXEngine_->Wait(waitMindxTimes_);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }
    return repairEvent_ == MindXEvent::MINDX_EVENT_MIGRATION ? TTP_OK : TTP_STOP_SERVICE;
}

void Controller::SwapHotSwitchRankInfo()
{
    AutoLock linkLock(linkMapLock_, TYPE_WRITE);
    // rankLinkMap
    SwapMapWithKeys(rankLinkMap_, rankLinkMapTmp_);
    // linkIdMap
    SwapMapWithVals(linkIdMap_, linkIdMapTmp_);

    AutoLock statusMapLock(statusMapLock_, TYPE_WRITE);
    // statusMap
    SwapMapWithKeys(statusMap_, statusMapTmp_);
}

TResult Controller::MigrationCallback()
{
    TTP_LOG_INFO("start notify hot switch migration");
    SwapHotSwitchRankInfo();

    auto migration = [this] () {
        auto ret = HcclCommGroupRepair(hotSwitchRanks_);
        TTP_ASSERT_RETURN(ret == TTP_OK, TTP_ERROR);

        std::vector<RepairInfo> info;
        ret = PrepareRepairMsg(info, hotSwitchRanks_);
        TTP_ASSERT_RETURN(ret == TTP_OK, TTP_ERROR);

        ret = RepairProcess(info);
        TTP_ASSERT_RETURN(ret == TTP_OK, TTP_ERROR);

        auto ranks = GetAllLinkRanks();
        ret = NotifyRankRollback(ranks, RepairType::RT_ROLLBACK);
        TTP_ASSERT_RETURN(ret == TTP_OK, TTP_ERROR);

        ret = DoNoDataAction(ranks, TTP_MSG_OP_NOTIFY_NORMAL, ACTION_OP_NOTIFY_NORMAL, true);
        TTP_ASSERT_RETURN(ret == TTP_OK, TTP_ERROR);

        return TTP_OK;
    };

    auto ret = migration();
    if (ret != TTP_OK) {
        std::string msg = "notify hot switch migration failed";
        mindXEngine_->ReportResult(RepairResult::FAILED_ALLOW_DUMP, msg, errorRankMsg_, errorRankLock_);
        return WaitDumpOrExitStrategy();
    }
    TTP_LOG_INFO("notify hot switch migration success");
    isNeedToReportResult_.store(true);
    return TTP_OK;
}

TResult Controller::PauseCallback()
{
    std::string msg = "Controller pause train success!";

    TResult ret = PauseTrain();
    if (ret != TTP_OK) {
        msg = "Controller pause train failed.";
        TTP_LOG_ERROR(msg);
        mindXEngine_->ReportStopComplete(RepairResult::REPAIR_COMMON_ERROR, msg, errorRankMsg_, errorRankLock_);
        return TTP_STOP_SERVICE;
    }

    TTP_LOG_INFO(msg);
    mindXEngine_->ReportStopComplete(RepairResult::REPAIR_SUCCESS, msg, errorRankMsg_, errorRankLock_);

    ret = PauseWait();
    if (ret != TTP_OK) {
        return ret;
    }

    ret = ContinueTrain();
    if (ret != TTP_OK) {
        msg = "Controller continue training failed.";
        TTP_LOG_WARN(msg);
        mindXEngine_->ReportResult(RepairResult::REPAIR_COMMON_ERROR, msg, errorRankMsg_, errorRankLock_);
        return TTP_STOP_SERVICE;
    }

    msg = "Controller continue training success.";
    mindXEngine_->ReportResult(RepairResult::REPAIR_SUCCESS, msg, errorRankMsg_, errorRankLock_);

    return ret;
}

TResult Controller::PauseWait()
{
    TTP_LOG_INFO("Waiting for MindXCluster notify continue...");

    uint32_t waitTimes = 0;
    while (waitTimes < waitPauseTimes_) {
        if (repairEvent_ == MindXEvent::MINDX_EVENT_CONTINUE_TRAIN) {
            return TTP_OK;
        }
        if (repairEvent_ == MindXEvent::MINDX_EVENT_STOP_TRAIN) {
            return TTP_ERROR;
        }

        auto ret = CheckTrainStatus();
        if (ret == TTP_ERROR) {
#ifndef UT_ENABLED
            sleep(TTP_SLEEP_TIME);
            auto againRet = CheckTrainStatus();
            TTP_LOG_WARN("PauseWait found abnormal, ret: " << ret << ", againRet: " << againRet);
#endif
            break;
        }

        waitTimes += TTP_SLEEP_TIME;
        pthreadTimeChecker_.PthreadTimedwaitSecs(TTP_SLEEP_TIME);
    }

    auto msg = "Controller do pause wait failed.";
    TTP_LOG_WARN(msg);
    mindXEngine_->ReportResult(RepairResult::FAILED_ALLOW_DUMP, msg, errorRankMsg_, errorRankLock_);

    return TTP_ERROR;
}

TResult Controller::PauseTrain()
{
    auto ranks = GetAllLinkRanks();
    auto ret = DoNoDataAction(ranks, TTP_MSG_OP_COLLECTION, ACTION_OP_COLLECTION, true);
    TTP_LOG_INFO("controller collect latest status end...");
    if (ret != TTP_OK) {
        TTP_LOG_WARN("controller collect latest status found some worker error");
    }

    auto [pauseStep, _] = SelectLockStep();
    pauseStep += 1;
    repairStep_ = pauseStep;

    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(PauseMsg));
    if (buffer == nullptr) {
        return TTP_ERROR;
    }

    PauseMsg *ptr = static_cast<PauseMsg *>(buffer->DataPtrVoid());
    ptr->step = pauseStep;
    ptr->sn = actionSn_.fetch_add(1);
    ptr->hotSwitch = repairEvent_ == MindXEvent::MINDX_EVENT_HOT_SWITCH;
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    std::vector<ActionInfo> info {{TTP_MSG_OP_PAUSE, buffer, ranks}};
    ret = engine_->Process(ACTION_OP_PAUSE, info, true, ptr->sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " pause train found network error");
        return TTP_ERROR;
    }
    return ret;
}

TResult Controller::ContinueTrain()
{
    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(PauseMsg));
    if (buffer == nullptr) {
        return TTP_ERROR;
    }

    PauseMsg *ptr = static_cast<PauseMsg *>(buffer->DataPtrVoid());
    ptr->sn = actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    auto ranks = GetAllLinkRanks();
    std::vector<ActionInfo> info {{TTP_MSG_OP_CONTINUE, buffer, ranks}};
    auto ret = engine_->Process(ACTION_OP_CONTINUE, info, true, ptr->sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " continue train found network error");
        return TTP_ERROR;
    }
    return ret;
}

TResult Controller::MindXInnerInteraction()
{
    // 故障触发两种场景：1.normal状态机检测转到abnormal，上报停止需要等待mindx  2.mindx先通知stoptrain转到abnormal，因此不能SignalClean。
    auto response = mindXEngine_->ReportFaultRanks(errorRankMsg_, errorRankLock_);
    if (response != TTP_OK) {
        return TTP_STOP_SERVICE;
    }

    TTP_LOG_WARN("wait Mindx notify stop train...");
    auto time = std::max(waitPauseTimes_, waitMindxTimes_);
    auto ret = mindXEngine_->Wait(time);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }

    if (repairEvent_ == MindXEvent::MINDX_EVENT_ELEGANT_DUMP) {
        TTP_LOG_WARN("mindx calling notify to dump!");
        return TTP_ERROR;
    }
    if (repairEvent_ != MindXEvent::MINDX_EVENT_STOP_TRAIN) {
        TTP_LOG_WARN("mindx calling action unexpected! prepare to exit");
        return TTP_STOP_SERVICE;
    }

    RepairResult reportCode = !dpGroupListMap_.empty() ? RepairResult::REPAIR_SUCCESS : RepairResult::NO_PROCESSORS;
    response = mindXEngine_->ReportStopComplete(reportCode, "stop ok", errorRankMsg_, errorRankLock_);
    if (response != TTP_OK || reportCode == RepairResult::NO_PROCESSORS) {
        return TTP_STOP_SERVICE;
    }

    TTP_LOG_WARN("wait Mindx notify all fault ranks...");
    ret = mindXEngine_->Wait(waitMindxTimes_);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }

    if (repairEvent_ != MindXEvent::MINDX_EVENT_NOTIFY_FAULT_RANKS) {
        TTP_LOG_WARN("mindx calling action unexpected! prepare to exit");
        return TTP_STOP_SERVICE;
    }
    return TTP_OK;
}

TResult Controller::ProcessRepairFlow(bool isPreLocked)
{
    if (mindXEngine_->IsRegistered()) {
        if (isNeedToReportResult_.load()) {
            std::string msg = "do last repair failed!";
            mindXEngine_->ReportResult(RepairResult::FAILED_ALLOW_DUMP, msg, errorRankMsg_, errorRankLock_);
            isNeedToReportResult_.store(false);
            return WaitDumpOrExitStrategy();
        }

        if (repairEvent_ == MindXEvent::MINDX_EVENT_ELEGANT_DUMP) {
            TTP_LOG_WARN("Mindx Cluster notify to do dump, try to do dump...");
            return TTP_ERROR;
        }

        TResult ret = MindXInnerInteraction();
        if (ret != TTP_OK) {
            return ret;
        }

        ret = MindXConfirmStrategy(isPreLocked);
        if (ret != TTP_OK) {
            return ret;
        }
    } else {
        // MindSpore场景需要提前校验副本，其他场景只需校验DP组情况
        bool canRepair = mindSpore_ ? CheckDpGroup() && CheckCanRepair() : CheckDpGroup();
        if (!canRepair || !isPreLocked) {
            return TTP_STOP_SERVICE; // exit
        }

        repairType_ = ConfirmRepairType();
        if (repairType_ == ControllerRepairType::CRT_DUMP) {
            return TTP_ERROR;  // dump
        }
    }
    return TTP_OK;
}

TResult Controller::MindXConfirmStrategy(bool isPreLocked)
{
    std::vector<std::string> strategies {STRATEGY_EXIT};
    // MindSpore场景需要提前校验副本，其他场景只需校验DP组情况
    bool canRepair = mindSpore_ ? CheckDpGroup() && CheckCanRepair() : CheckDpGroup();
    if (canRepair && isPreLocked) {
        strategies.push_back(STRATEGY_DUMP);

        repairType_ = ConfirmRepairType();
        if (repairType_ == ControllerRepairType::CRT_RETRY) {
            strategies.push_back(STRATEGY_UCE);
            if (arfSwitch_) {
                strategies.push_back(STRATEGY_ARF);
            }
        } else if (repairType_ == ControllerRepairType::CRT_ARF) {
            strategies.push_back(STRATEGY_ARF);
            // 支持ARF 转 ZIT
            if (zitSwitch_) {
                strategies.push_back(STRATEGY_DOWNGRADE);
                strategies.push_back(STRATEGY_UPGRADE);
            }
        } else if (repairType_ == ControllerRepairType::CRT_DOWNGRADE) {
            strategies.push_back(STRATEGY_DOWNGRADE);
            strategies.push_back(STRATEGY_UPGRADE);
        }
    }

    auto ret = mindXEngine_->ReportStrategies(strategies, errorRankMsg_, errorRankLock_);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }

    TTP_LOG_WARN("wait Mindx notify repair strategy ...");
    auto time = std::max(waitHcclTime_, waitMindxTimes_);
    ret = mindXEngine_->Wait(time);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }

    return HandleRecoverStrategy();
}

TResult Controller::HandleRecoverStrategy()
{
    switch (repairEvent_) {
        case MindXEvent::MINDX_EVENT_ARF:
            repairType_ = ControllerRepairType::CRT_ARF;
            return TTP_OK;
        case MindXEvent::MINDX_EVENT_UCE:
            repairType_ = ControllerRepairType::CRT_RETRY;
            return TTP_OK;
        case MindXEvent::MINDX_EVENT_DOWNGRADE:
            repairType_ = ControllerRepairType::CRT_DOWNGRADE;
            return TTP_OK;
        case MindXEvent::MINDX_EVENT_UPGRADE:
            repairType_ = ControllerRepairType::CRT_UPGRADE;
            return TTP_OK;
        case MindXEvent::MINDX_EVENT_DUMP:
            repairType_ = ControllerRepairType::CRT_DUMP;
            return TTP_ERROR;
        case MindXEvent::MINDX_EVENT_EXIT:
            return TTP_STOP_SERVICE;
        case MindXEvent::MINDX_EVENT_INVALID:
            mindXEngine_->ReportResult(RepairResult::REPAIR_COMMON_ERROR,
                "invalid strategy", errorRankMsg_, errorRankLock_);
            return TTP_STOP_SERVICE;
        default:
            return TTP_STOP_SERVICE;
    }
}

TResult Controller::AbnormalCallback()
{
    if (isPorcessorExit_) {
        return TTP_STOP_SERVICE;  // found process exit, do nothing but exit
    }

    repairId_.fetch_add(1);
    auto ret = DoPause();
    SelectErrorRanks();

    // backup no need to notify mindx
    if (isBackupToMaster_.load()) {
        return TTP_ERROR;
    }

    return ProcessRepairFlow(ret == TTP_OK);
}

TResult Controller::DumpCallback()
{
    TResult ret = TTP_ERROR;
    std::string msg = "DumpCallback return TTP_OK";
    if (!CheckDpGroup() || unableRepair_) {
        msg = "do not support dump!";
        mindXEngine_->ReportResult(RepairResult::REPAIR_COMMON_ERROR, msg, errorRankMsg_, errorRankLock_);
        TTP_LOG_ERROR(msg);
        return ret;
    }

    auto errRanks = GetErrorRanks();
    ret = HandleDumpStatus(errRanks);
    if (ret != TTP_OK) {
        msg = "DumpCallback: save ckpt failed, ret: " + std::to_string(ret);
        TTP_LOG_ERROR(msg);
        mindXEngine_->ReportResult(RepairResult::REPAIR_COMMON_ERROR, msg, errorRankMsg_, errorRankLock_);
        return ret;
    }

    ret = Rename();
    if (ret != TTP_OK) {
        msg = "DumpCallback: rename failed, ret: " + std::to_string(ret);
        TTP_LOG_ERROR(msg);
        mindXEngine_->ReportResult(RepairResult::REPAIR_COMMON_ERROR, msg, errorRankMsg_, errorRankLock_);
        return ret;
    }
    TTP_LOG_INFO(msg);
    msg = "do dump success";
    mindXEngine_->ReportResult(RepairResult::REPAIR_SUCCESS, msg, errorRankMsg_, errorRankLock_);

    return TTP_OK;
}

TResult Controller::EnvClearCallback()
{
    TResult ret = TTP_ERROR;
    std::string msg;

    auto ranks = GetAllLinkRanks(true);
    RepairResult repairResult = RepairResult::REPAIR_COMMON_ERROR;
    do {
        TTP_LOG_INFO("start notify all device stop...");
        ret = DoNoDataAction(ranks, TTP_MSG_OP_DEVICE_STOP, ACTION_OP_DEVICE_STOP, true, true);
        if (ret != TTP_OK) {
            msg = "controller:" + std::to_string(rank_) + " notify device stop failed";
            break;
        }

        TTP_LOG_INFO("start notify all device clean...");
        ret = DoNoDataAction(ranks, TTP_MSG_OP_DEVICE_CLEAN, ACTION_OP_DEVICE_CLEAN, true, true);
        if (ret != TTP_OK) {
            msg = "controller:" + std::to_string(rank_) + " notify device clean failed";
            repairResult = canRetryCleanFlag_ ? RepairResult::FAILED_ALLOW_REOCVER : RepairResult::REPAIR_COMMON_ERROR;
            break;
        }
    } while (0);

    if (ret != TTP_OK) {
        TTP_LOG_ERROR(msg);
        mindXEngine_->ReportResult(repairResult, msg, errorRankMsg_, errorRankLock_);
        if (repairEvent_ == MindXEvent::MINDX_EVENT_UCE && canRetryCleanFlag_) {
            ret = MindXReConfirmStrategy();
        } else {
            ret = WaitDumpOrExitStrategy();
        }
        if (ret == TTP_ERROR) {
            msg = "stop & clean failed! not allowed to dump!";
            mindXEngine_->ReportResult(RepairResult::REPAIR_COMMON_ERROR, msg, errorRankMsg_, errorRankLock_);
            ret = TTP_STOP_SERVICE;
        }

        return ret;
    }

    TTP_LOG_INFO("notify all device stop & clean ok!");
    if (repairType_ == ControllerRepairType::CRT_DOWNGRADE) {
        return TTP_DOWNGRADE;
    }
    return TTP_OK;
}

TResult Controller::RepairCallback()
{
    TResult ret = TTP_OK;
    if (repairType_ == ControllerRepairType::CRT_RETRY) {
        if (!ONLY_HCCL_BIT(hcclFlag_)) {
            ret = UCERepair();
        }
    } else if (repairType_ == ControllerRepairType::CRT_UPGRADE) {
        ret = UpGradeRepair();
    } else if (repairType_ == ControllerRepairType::CRT_ARF) {
        ret = ARFRepair();
        if (repairEvent_ == MindXEvent::MINDX_EVENT_DUMP || repairEvent_ == MindXEvent::MINDX_EVENT_EXIT ||
            repairEvent_ == MindXEvent::MINDX_EVENT_DOWNGRADE) {
            return ret;
        }
    } else {
        TTP_LOG_ERROR("Unknown repair type.. " << static_cast<int32_t>(repairType_));
        return TTP_ERROR;
    }
    std::string msg;
    do {
        if (ret != TTP_OK) {
            msg = "do repair failed, ret: " + std::to_string(ret);
            break;
        }

        auto ranks = GetAllLinkRanks();
        TTP_LOG_INFO("start rollback");
        ret = NotifyRankRollback(ranks, RepairType::RT_ROLLBACK);
        if (ret != TTP_OK) {
            msg = "controller:" + std::to_string(rank_) + " notify processor rollback failed";
            break;
        }

        TTP_LOG_INFO("start notify all rank normal...");
        ret = DoNoDataAction(ranks, TTP_MSG_OP_NOTIFY_NORMAL, ACTION_OP_NOTIFY_NORMAL, true);
        msg = "controller:" + std::to_string(rank_) + " notify processor normal failed"; // print when error
    } while (0);

    if (ret != TTP_OK) {
        TTP_LOG_ERROR(msg);
        auto result =
            (repairType_ == ControllerRepairType::CRT_ARF || repairType_ == ControllerRepairType::CRT_UPGRADE) ?
            RepairResult::FAILED_ALLOW_DUMP : RepairResult::REPAIR_COMMON_ERROR;
        mindXEngine_->ReportResult(result, msg, errorRankMsg_, errorRankLock_);
        return WaitDumpOrExitStrategy();
    }

    if (mindXEngine_->IsRegistered()) {
        isNeedToReportResult_.store(true);
    }
    return TTP_OK;
}

TResult Controller::DowngradeRepairCallback()
{
    ChooseIsolateRanks();
    auto isolateRanks = GetIsolateRanks();
    // 上报mindx隔离卡，当前全杀, mindx后续支持隔离卡不杀；全杀情况下isolateRanks和errRanks一样
    {
        AutoLock errorRankLock(errorRankLock_, TYPE_WRITE);
        for (auto rk : isolateRanks) {
            if (errorRankMsg_.count(rk) == 0) {
                errorRankMsg_[rk] = PROCESSES_ERROR;
            }
        }
    }
    std::string msg = "controller notify isolate ranks.";
    bool canDowngrade = CheckCanRepair(true);
    if (!canDowngrade) {
        msg = "controller find no complete duplication to do downgrade running.";
        TTP_LOG_ERROR(msg);
        mindXEngine_->ReportResult(RepairResult::REPAIR_COMMON_ERROR, msg, errorRankMsg_, errorRankLock_);
        return TTP_STOP_SERVICE;
    }

    if (errorRankMsg_.count(0)) {
        msg = "controller find rank 0 in error ranks, not support downgrade.";
        TTP_LOG_ERROR(msg);
        mindXEngine_->ReportResult(RepairResult::FAILED_ALLOW_DUMP, msg, errorRankMsg_, errorRankLock_);
        return WaitDumpOrExitStrategy();
    }

    mindXEngine_->ReportResult(RepairResult::CHOOSE_ZIT_RANK_SUCCESS, msg, errorRankMsg_, errorRankLock_);
    auto errRanks = GetErrorRanks();
    IsolateRanksSetStatus(isolateRanks, errRanks);

    TResult ret = DowngradeNotifyNormalRanks();
    if (ret != TTP_OK) {
        msg = "controller do downgarde repair failed";
        mindXEngine_->ReportResult(RepairResult::FAILED_ALLOW_DUMP, msg, errorRankMsg_, errorRankLock_);
        return WaitDumpOrExitStrategy();
    }
    isNeedToReportResult_.store(true);
    return TTP_OK;
}

TResult Controller::ZitHandleStrategy()
{
    std::vector<std::string> strategies {STRATEGY_EXIT, STRATEGY_DUMP};

    auto ret = mindXEngine_->ReportStrategies(strategies, errorRankMsg_, errorRankLock_);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }

    TTP_LOG_WARN("wait Mindx notify repair strategy ...");
    ret = mindXEngine_->Wait(waitMindxTimes_);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }

    if (repairEvent_ == MindXEvent::MINDX_EVENT_DUMP) {
        repairType_ = ControllerRepairType::CRT_DUMP;
        return TTP_ERROR;
    } else if (repairEvent_ == MindXEvent::MINDX_EVENT_EXIT) {
        return TTP_STOP_SERVICE;
    } else {
        TTP_LOG_ERROR("controller recieve invalid strategy in zit, repairEvent:" <<
            static_cast<uint32_t>(repairEvent_));
        mindXEngine_->ReportResult(RepairResult::REPAIR_COMMON_ERROR, "invalid strategy in zit",
            errorRankMsg_, errorRankLock_);
        return TTP_STOP_SERVICE;
    }
}

TResult Controller::ZitHandleNewFault()
{
    TResult ret = DoPause();

    SelectErrorRanks();
    if (ret != TTP_OK) {
        TTP_LOG_WARN("prelock failed failed in zit, try notify mindx do dump or exit...");
    }
    ret = MindXInnerInteraction();
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }
    return ZitHandleStrategy();
}

TResult Controller::DowngradeRunning()
{
    TTP_LOG_INFO("controller:" << rank_ << " start downgrade running...");
    TResult ret = TTP_ERROR;
    while (!isStopped_.load()) {
        ret = CheckTrainStatus();
        if (ret == TTP_NEED_RETRY && repairEvent_ == MindXEvent::MINDX_EVENT_UPGRADE) {
            repairType_ = ControllerRepairType::CRT_UPGRADE;
            TTP_LOG_INFO("mindx notify upgrade, isolated ranks have finished pretrain, ret: " << ret);
            return TTP_NEED_RETRY;
        }
        // 降级训练时TTP_REPAIR不处理，等待都PREREPAIR_FINISH后再升级
        if (ret == TTP_ERROR || repairEvent_ == MindXEvent::MINDX_EVENT_STOP_TRAIN) {
            TTP_LOG_ERROR("controller find abnormal in downgrade running, ret: " << ret << ", repairEvent_:"
                << static_cast<uint32_t>(repairEvent_));
            return ZitHandleNewFault();
        }
        pthreadTimeChecker_.PthreadTimedwaitSecs(TTP_SLEEP_TIME);
    }
    if (isStopped_.load()) {
        return TTP_STOP_SERVICE;
    }
    return ret;
}

TResult Controller::UpgradeRunning()
{
    TTP_LOG_INFO("controller:" << rank_ << " start upgrade run...");
    auto ret = DoPause();
    TTP_RET_LOG(ret, "controller upgrade do prelock, ret:" << ret);
    return TTP_OK;
}

TResult Controller::ExitCallback()
{
#ifndef UT_ENABLED  // UT Test skip waiting
    if (mindXEngine_->IsRegistered()) {
        TTP_LOG_INFO("controller start wait mindx exit ...");
        auto ret = mindXEngine_->Wait(waitMindxTimes_);
        if (ret != TTP_OK) {
            TTP_LOG_WARN("controller start wait mindx exit timeout, notify all processor exit.");
        } else {
            return TTP_STOP_SERVICE;
        }
    }
#endif

    auto ranks = GetAllLinkRanks();
    (void) DoNoDataAction(ranks, TTP_MSG_OP_EXIT, ACTION_OP_EXIT, false);
    Destroy(true);

    TTP_LOG_INFO("ExitCallBack return TTP_STOP_SERVICE, controller exit done...");
    return TTP_STOP_SERVICE;
}

void Controller::ExitNotify()
{
    auto ranks = GetAllLinkRanks();
    auto ret = DoNoDataAction(ranks, TTP_MSG_OP_DESTROY_NOTIFY, ACTION_OP_DESTROY_NOTIFY, false);
    TTP_LOG_INFO("controller send exit msg to all processors...");
    if (ret != TTP_OK) {
        TTP_LOG_WARN("controller send exit msg to all processors failed, do not retry...");
    }
}

TResult Controller::Destroy(bool isInner)
{
    // 1. 先外部接口调用拿锁，然后状态机刚好切换到exit状态，内部调用尝试拿锁，拿不到返回成功，防止死锁。
    // 2. 先状态机切换到exit状态，内部尝试拿锁，然后外部调用拿锁，拿不到需要阻塞等待，最后调用状态机stop关闭线程。
    if (!isInner) {
        initOrDestroyMutex_.lock();
    } else if (!initOrDestroyMutex_.try_lock()) {
        return TTP_OK;
    }

    if (!isInited_.load()) {
        if (!isInner && stateMachine_ != nullptr) {
            stateMachine_->Stop();
            stateMachine_ = nullptr;
        }
        initOrDestroyMutex_.unlock();
        TTP_LOG_DEBUG("controller is not inited, do not destroy ...");
        return TTP_OK;
    }

    if (mindXEngine_ != nullptr) {
        mindXEngine_->WakeUp();
    }

    if (!isInner) {
        ExitNotify();
    }
    isStopped_.store(true);
    mServer_->Stop();
    isStarted_.store(false);
    isInited_.store(false);
    isAlreadyBrod_.store(false);
    isBackupToMaster_.store(false);
    if (!isInner && stateMachine_ != nullptr) {
        stateMachine_->Stop();
        stateMachine_ = nullptr;
    }
    TTP_LOG_DEBUG("controller Destroy: isStopped_ set true");

    ExitLogsHandler();
    initOrDestroyMutex_.unlock();
    return TTP_OK;
}

TResult Controller::ExitLogsHandler()
{
    std::string logPath = "logs/ttp_log.log";
    const char *tempLogPathEnv = std::getenv("TTP_LOG_PATH");
    std::string errMsg = "";
    if (tempLogPathEnv != nullptr) {
        logPath = tempLogPathEnv;
    }

    std::string purePath = "./";
    auto found = logPath.find_last_of('/');
    if (found != std::string::npos) {
        purePath = logPath.substr(0, found + 1);
    }
    DIR *dir = opendir(purePath.c_str());
    TTP_ASSERT_RETURN(dir != nullptr, TTP_ERROR);
    std::set<std::string> fileSet;
    struct dirent *entry;
    TResult ret = TTP_OK;
    while ((entry = readdir(dir)) != nullptr) {
        std::string fileName = entry->d_name;
        if (fileName == "." || fileName == ".." || fileName.find(".log") == std::string::npos) {
            continue;
        }
        fileSet.insert(purePath + fileName);
    }
    constexpr __mode_t fileMode = 0440;
    for (const auto& filePath : fileSet) {
        if (FileUtils::RegularFilePath(filePath, errMsg) &&
            chmod(filePath.c_str(), fileMode)) {
            ret = TTP_ERROR;
        }
    }
    closedir(dir);
    return ret;
}

TResult Controller::Prelock()
{
    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(PrelockMsg));
    if (buffer == nullptr) {
        return TTP_ERROR;
    }

    auto [step, ranks] = SelectLockStep();
    repairStep_ = step;

    PrelockMsg *ptr = static_cast<PrelockMsg *>(buffer->DataPtrVoid());
    ptr->step = step;
    ptr->sn = actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    std::vector<ActionInfo> info {{TTP_MSG_OP_PRELOCK, buffer, ranks}};
    prelockRet_.store(TTP_OK);
    TResult ret = engine_->Process(ACTION_OP_PRELOCK, info, true, ptr->sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " prelock found network error");
        return TTP_NEED_RETRY;
    }
    return static_cast<TResult>(prelockRet_.load());
}

TResult Controller::DoPause()
{
    auto ranks = GetAllLinkRanks(true);
    auto ret = DoNoDataAction(ranks, TTP_MSG_OP_COLLECTION, ACTION_OP_COLLECTION, true);
    TTP_LOG_INFO("controller collect latest status end...");
    if (ret != TTP_OK) {
        TTP_LOG_WARN("controller collect latest status found some worker error");
    }

    // select step & check no dump
    TTP_LOG_INFO("controller:" << rank_ << " pre lock start.");
    uint32_t tryTimes = 0;
    while (tryTimes++ < PRELOCK_RETRY_TIME) {
        ret = Prelock();
        TTP_LOG_INFO("controller prelock step: " << repairStep_ << " try times: " << tryTimes);
        if (ret == TTP_OK) {
            TTP_LOG_INFO("controller:" << rank_ << " pre lock success.");
            return TTP_OK;
        } else if (ret != TTP_NEED_RETRY) {
            TTP_LOG_ERROR("controller:" << rank_ << " pre lock error.");
            return TTP_ERROR;
        }
    }

    TTP_LOG_ERROR("controller:" << rank_ << " pre lock max_step error, try limit.");
    return TTP_TIMEOUT;
}

TResult Controller::UpdateStatus(HeartBeatMsg *originHeartBeatMsg)
{
    TTP_ASSERT_RETURN(originHeartBeatMsg != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(originHeartBeatMsg->JudgeVariableValid(), TTP_ERROR);

    AutoLock statusMapLock(statusMapLock_, TYPE_WRITE);
    auto isHotSwitchRank = originHeartBeatMsg->repairId == repairId_ &&
        repairEvent_ == MindXEvent::MINDX_EVENT_HOT_SWITCH;
    auto &statusMap = isHotSwitchRank ? statusMapTmp_ : statusMap_;
    auto it = statusMap.find(originHeartBeatMsg->rank);
    if (it == statusMap.end()) {
        TTP_LOG_ERROR("heartbeat rank:" << originHeartBeatMsg->rank << " not exist in statusMap.");
        return TTP_ERROR;
    }

    // arf wait时将异常卡置为offline,若mindx kill较晚,异常卡还在发送心跳,避免覆盖offline状态
    // 仅有register可以修改offline状态
    if (it->second.run_status != TTP_STATUS_OFFLINE) {
        it->second = originHeartBeatMsg->status;
    }
    it->second.lastUpdateTime = GetNowTime();

    return TTP_OK;
}

TResult Controller::HandleCollectionReply(const AccTcpRequestContext &context)
{
    if (isStopped_.load()) {
        TTP_LOG_ERROR("this controller:" << rank_ << " has stopped service.");
        return TTP_ERROR;
    }

    TTP_ASSERT_RETURN(context.DataLen() == sizeof(ResultAndHBReplyMsg), TTP_ERROR);
    ResultAndHBReplyMsg *reply = static_cast<ResultAndHBReplyMsg *>(context.DataPtr());
    TTP_ASSERT_RETURN(reply != nullptr, TTP_ERROR);

    if (UpdateStatus(&(reply->hb)) != TTP_OK) {
        return TTP_ERROR;
    }
    return TTP_OK;
}

TResult Controller::HandlePrelockReply(const AccTcpRequestContext &context)
{
    TTP_ASSERT_RETURN(context.DataLen() == sizeof(ResultAndHBReplyMsg), TTP_ERROR);
    ResultAndHBReplyMsg *reply = static_cast<ResultAndHBReplyMsg *>(context.DataPtr());
    TTP_ASSERT_RETURN(reply != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(reply->ret >= TTP_OK && reply->ret < TTP_BUTT, TTP_ERROR);

    if (reply->ret != TTP_NEED_RETRY && reply->ret != TTP_OK) {
        prelockRet_.store(TTP_ERROR);
        return TTP_OK;
    }

    if (UpdateStatus(&reply->hb) != TTP_OK) {
        TTP_LOG_ERROR("Prelock reply update status failed.");
        return TTP_ERROR;
    }

    if (reply->ret == TTP_NEED_RETRY) {
        int32_t nowRet = TTP_OK;
        prelockRet_.compare_exchange_strong(nowRet, TTP_NEED_RETRY); // not care failed
    }
    return TTP_OK;
}

TResult Controller::HandleNotifyNormalReply(const AccTcpRequestContext &context)
{
    TTP_ASSERT_RETURN(context.DataLen() == sizeof(ResultAndHBReplyMsg), TTP_ERROR);
    ResultAndHBReplyMsg *reply = static_cast<ResultAndHBReplyMsg *>(context.DataPtr());
    TTP_ASSERT_RETURN(reply != nullptr, TTP_ERROR);

    TTP_ASSERT_RETURN(reply->ret >= TTP_OK && reply->ret < TTP_BUTT, TTP_ERROR);
    TTP_ASSERT_RETURN(reply->repairStep >= -1 && reply->repairStep < INT64_MAX, TTP_ERROR);
    if (reply->repairStep > 0) {
        loadCkptRepairStep_.store(reply->repairStep);
    }
    if (reply->ret != TTP_OK) {
        return reply->ret;
    }

    if (UpdateStatus(&reply->hb) != TTP_OK) {
        TTP_LOG_ERROR("notify normal update latest status failed.");
        return TTP_ERROR;
    }

    return TTP_OK;
}

TResult Controller::HandleCleanReply(const AccTcpRequestContext &context)
{
    TTP_ASSERT_RETURN(context.DataLen() == sizeof(TTPReplyMsg), TTP_ERROR);
    TTPReplyMsg *replyMsg = static_cast<TTPReplyMsg *>(context.DataPtr());
    TTP_ASSERT_RETURN(replyMsg != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(replyMsg->status >= TTP_OK && replyMsg->status < TTP_BUTT, TTP_ERROR);

    // clean failed with TTP_ERROR during UCE repair, support transform to do ARF
    if (replyMsg->status == TTP_ERROR && arfSwitch_) {
        canRetryCleanFlag_.store(true);
        AutoLock lock(errorRankLock_, TYPE_WRITE);
        errorRankMsg_[replyMsg->rank] = PROCESSES_ERROR;
    }

    return TTP_OK;
}

void Controller::RecordRankIp(int32_t rankIn)
{
    AutoLock linkLock(linkMapLock_, TYPE_READ);
    auto it = rankLinkMap_.find(rankIn);
    if (it == rankLinkMap_.end()) {
        TTP_LOG_WARN("current rankLink not found in rankLinkMap_.");
        return;
    }
    std::string ipPort = it->second->GetLinkRemoteIpPort();
    std::string ip = ipPort.substr(0, ipPort.find_last_of(":"));

    AutoLock ipLock(ipMapLock_, TYPE_WRITE);
    rankIpMap_[rankIn] = ip;
}

TResult Controller::ResultAndHbReplyParse(const AccTcpRequestContext &context, TTPReplyMsg &msg)
{
    TTP_ASSERT_RETURN(context.DataLen() == sizeof(ResultAndHBReplyMsg), TTP_ERROR);
    ResultAndHBReplyMsg *reply = static_cast<ResultAndHBReplyMsg *>(context.DataPtr());
    TTP_ASSERT_RETURN(reply != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(reply->ret >= TTP_OK && reply->ret < TTP_BUTT, TTP_ERROR);
    TTP_ASSERT_RETURN(reply->hb.rank >= 0 && reply->hb.rank < TTP_MAX_WORLD_SIZE, TTP_ERROR);

    msg.sn = reply->sn;
    msg.status = reply->ret;
    msg.rank = reply->hb.rank;
    return TTP_OK;
}

TResult Controller::PrelockResultAndHbReplyParse(const AccTcpRequestContext &context, TTPReplyMsg &msg)
{
    TTP_ASSERT_RETURN(context.DataLen() == sizeof(ResultAndHBReplyMsg), TTP_ERROR);
    ResultAndHBReplyMsg *reply = static_cast<ResultAndHBReplyMsg *>(context.DataPtr());
    TTP_ASSERT_RETURN(reply != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(reply->ret >= TTP_OK && reply->ret < TTP_BUTT, TTP_ERROR);
    TTP_ASSERT_RETURN(reply->hb.rank >= 0 && reply->hb.rank < TTP_MAX_WORLD_SIZE, TTP_ERROR);

    msg.sn = reply->sn;
    msg.status = (reply->ret == TTP_ERROR) ? TTP_ERROR : TTP_OK;
    msg.rank = reply->hb.rank;
    return TTP_OK;
}

inline bool InRepairStatus(StateOp status)
{
    return (status != STATE_OP_INIT) && (status != STATE_OP_NORMAL) && (status != STATE_OP_DUMP) &&
        (status != STATE_OP_EXIT) && (status != STATE_OP_FINAL);
}

TResult Controller::RegisterStatus(RegisterMsg *registerMsg)
{
    AutoLock statusMapLock(statusMapLock_, TYPE_WRITE);

    auto rank = registerMsg->rank;
    auto &statusMap = repairEvent_ == MindXEvent::MINDX_EVENT_HOT_SWITCH ? statusMapTmp_ : statusMap_;
    auto it = statusMap.find(rank);
    if (it == statusMap.end()) {
        statusMap[rank].lastUpdateTime = GetNowTime();
        return TTP_OK;
    }

    if (!InRepairStatus(stateMachine_->GetCurrentState())) {
        statusMap[rank].lastUpdateTime = GetNowTime();
        TTP_LOG_WARN("repeated registration. rank:" << rank);
        return TTP_OK;
    }

    // run_status = {normal, abnormal, offline, isolate, finish}
    AutoLock errorRankLock(errorRankLock_, TYPE_READ);
    bool isErrorRank = errorRankMsg_.count(rank) > 0;
    if ((it->second.run_status == TTP_STATUS_ABNORMAL || it->second.run_status == TTP_STATUS_OFFLINE) ||
        (it->second.data_aval != TTP_STATUS_NORMAL) || (it->second.npu_status != TTP_STATUS_NORMAL) ||
        isErrorRank) { // 重启太快,statusMap未更新,只能从errorRanks中查询
        it->second.run_status = TTP_STATUS_ISOLATE;
        it->second.data_aval = it->second.npu_status = TTP_STATUS_NORMAL;
        it->second.lastUpdateTime = GetNowTime();
        it->second.step = it->second.backup_step = 0;
        return TTP_REPAIR;
    } else {
        TTP_LOG_ERROR("rank status is ok. rank:" << rank);
        return TTP_ERROR;
    }
}

TResult Controller::HandleRegister(const AccTcpRequestContext &context)
{
    TTP_ASSERT_RETURN(worldSize_ > 0 && worldSize_ <= TTP_MAX_WORLD_SIZE, TTP_ERROR);
    TResult ret = TTP_ERROR;
    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(RegisterReply));
    if (buffer == nullptr) {
        return TTP_ERROR;
    }
    RegisterReply *reply = static_cast<RegisterReply *>(buffer->DataPtrVoid());
    reply->repairId = repairId_.load();
    reply->hotSwitch = repairEvent_ == MindXEvent::MINDX_EVENT_HOT_SWITCH;

    do {
        if (!isInited_.load()) {
            TTP_LOG_ERROR("controller has not start work.");
            break;
        }

        if (context.DataLen() != sizeof(RegisterMsg) || context.DataPtr() == nullptr) {
            TTP_LOG_ERROR("register message is invalid");
            break;
        }

        RegisterMsg *registerMsg = static_cast<RegisterMsg *>(context.DataPtr());
        if (registerMsg->rank >= worldSize_ || registerMsg->rank < 0) {
            ret = TTP_WAIT_CHECK; // UPDATE: hot_backup
            break;
        }

        ret = RegisterStatus(registerMsg); // return ok,error,repair
        if (ret == TTP_ERROR) {
            break;
        }

        RecordRankIp(registerMsg->rank);
        TTP_LOG_DEBUG("controller finish handle register from rank:" << registerMsg->rank);
    } while (0);

    reply->ret = static_cast<int32_t>(ret);
    return static_cast<TResult>(context.Reply(TTP_OK, buffer));
}

TResult Controller::HandleReportInfo(const AccTcpRequestContext &context)
{
    TResult ret = TTP_ERROR;
    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(RegisterReply));
    if (buffer == nullptr) {
        return TTP_ERROR;
    }
    RegisterReply *reply = static_cast<RegisterReply *>(buffer->DataPtrVoid());
    reply->repairId = repairId_.load();

    do {
        if (!isInited_.load()) {
            TTP_LOG_ERROR("controller has not inited.");
            break;
        }

        ReplicaMsg *replicaMsg = static_cast<ReplicaMsg *>(context.DataPtr());
        TTP_ASSERT_RETURN(replicaMsg != nullptr, TTP_ERROR);
        if (!replicaMsg->JudgeVariableValid(context.DataLen(), worldSize_)) {
            break;
        }

        if (InitDpGroupMap(replicaMsg) != TTP_OK) {
            TTP_LOG_ERROR("controller init dp group map failed.");
            break;
        }

        ret = TTP_OK;
        reportedCnt_.fetch_add(1);
        TTP_LOG_INFO("controller finish handle report info from rank:" << replicaMsg->rank);
    } while (0);
    reply->ret = static_cast<int32_t>(ret);
    return static_cast<TResult>(context.Reply(TTP_OK, buffer));
}

TResult Controller::HandleReportDp(const AccTcpRequestContext &context)
{
    TResult ret = TTP_ERROR;
    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(RegisterReply));
    if (buffer == nullptr) {
        TTP_LOG_ERROR("malloc report reply buffer failed");
        return TTP_ERROR;
    }
    RegisterReply *reply = static_cast<RegisterReply *>(buffer->DataPtrVoid());
    reply->repairId = repairId_.load();

    do {
        if (!isInited_.load()) {
            TTP_LOG_ERROR("controller has not inited.");
            break;
        }

        DpMsg *dpMsg = static_cast<DpMsg *>(context.DataPtr());
        TTP_ASSERT_RETURN(dpMsg != nullptr, TTP_ERROR);
        if (!dpMsg->JudgeVariableValid(context.DataLen(), worldSize_)) {
            break;
        }

        originDpLock_.LockWrite();
        originDpSet_.insert(std::vector<int32_t>(dpMsg->dpList, dpMsg->dpList + dpMsg->dpNum));
        originDpLock_.UnLock();

        ret = TTP_OK;
        TTP_LOG_INFO("controller finish handle report dp group, rank:" << dpMsg->rank);
    } while (0);
    reply->ret = static_cast<int32_t>(ret);
    return static_cast<TResult>(context.Reply(TTP_OK, buffer));
}

TResult Controller::InitDpGroupMap(const ReplicaMsg *replicaMsg)
{
    AutoLock lock(dpGroupMapLock_, TYPE_WRITE);

    // 没有初始化过大小则初始化大小
    if (dpGroupListMap_.empty()) {
        dpGroupListMap_.resize(replicaMsg->num);
    }

    // 不同卡注册的数量不一致
    auto expectSize = static_cast<int32_t>(dpGroupListMap_.size());
    if (replicaMsg->num != expectSize) {
        TTP_LOG_ERROR("report error, group num:" << replicaMsg->num << " expect:" << expectSize <<
                      " rank:" << replicaMsg->rank);
        return TTP_ERROR;
    }

    arfSwitch_ = arfSwitch_ & replicaMsg->enableArf;
    uceSwitch_ = replicaMsg->enableUce;
    zitSwitch_ = zitSwitch_ & replicaMsg->enableZit;
    int32_t pIdx = replicaMsg->num + replicaMsg->num;
    for (int32_t groupIdx = 0; groupIdx < replicaMsg->num; groupIdx++) {
        std::vector<int32_t> rankVec;
        uint32_t count = static_cast<uint32_t>(worldSize_ / replicaMsg->ranks[pIdx]);
        for (int32_t i = 1; i <= replicaMsg->ranks[pIdx]; i++) {
            rankVec.push_back(replicaMsg->ranks[pIdx + i]);
        }
        sort(rankVec.begin(), rankVec.end());

        auto &[repCnt, dpGroups] = dpGroupListMap_[groupIdx];
        if (dpGroups.find(rankVec) == dpGroups.end()) {
            repCnt = replicaMsg->ranks[groupIdx];
            dpGroups.insert(rankVec);
            TTP_LOG_INFO("rank:" << replicaMsg->rank << ", report group list: " << IntVec2String(rankVec));
        }

        pIdx += replicaMsg->ranks[pIdx] + 1;
    }

    return TTP_OK;
}

// one master and two slave, distributed in different node
std::vector<BackupInfo> Controller::SelectBackUpController()
{
    std::set<std::string> tmpSet {controllerIp_};
    std::vector<BackupInfo> infoList;
    uint32_t chooseBackUpCount = 0;

    AutoLock linkLock(linkMapLock_, TYPE_READ);
    auto its = rankLinkMap_.find(rank_); // maybe rank0 controllerIp is different from processorIp
    if (its != rankLinkMap_.end()) {
        std::string ipPort = its->second->GetLinkRemoteIpPort();
        std::string ip = ipPort.substr(0, ipPort.find_last_of(":"));
        tmpSet.insert(ip);
    }
    for (auto it = rankLinkMap_.begin(); chooseBackUpCount < BACKUP_CONTROLLER_NUM && it != rankLinkMap_.end(); ++it) {
        std::string ipPort = it->second->GetLinkRemoteIpPort();
        std::string ip = ipPort.substr(0, ipPort.find_last_of(":"));
        if (tmpSet.find(ip) != tmpSet.end()) {
            continue;
        }

        BackupInfo info;
        info.ip = ip;
        info.port = ipPort.substr(ipPort.find_last_of(":") + 1);
        info.rank = it->first;
        infoList.push_back(info);
        tmpSet.insert(ip);
        chooseBackUpCount++;
    }

    return infoList;
}

TResult Controller::HandleHeartBeat(const AccTcpRequestContext &context)
{
    if (isStopped_.load()) {
        TTP_LOG_ERROR("this controller has stopped service.");
        return TTP_ERROR;
    }
    TTP_ASSERT_RETURN(context.DataLen() == sizeof(HeartBeatMsg), TTP_ERROR);

    HeartBeatMsg *originHeartBeatMsg = static_cast<HeartBeatMsg *>(context.DataPtr());
    TTP_ASSERT_RETURN(originHeartBeatMsg != nullptr, TTP_ERROR);
    if (UpdateStatus(originHeartBeatMsg) != TTP_OK) {
        return TTP_ERROR;
    }

    // UCE错误主动唤醒
    if (originHeartBeatMsg->status.npu_status != TTP_STATUS_NORMAL
        || originHeartBeatMsg->status.run_status == TTP_STATUS_ABNORMAL) {
        pthreadTimeChecker_ .PthreadSignal();
    }

    TTP_LOG_DEBUG("finish HandleHeartBeat, rank:" << originHeartBeatMsg->rank <<", ret:" << TTP_OK);
    return TTP_OK;
}

enum MaskStatusEnum : uint8_t {
    MASK_NORMAL = 0,
    MASK_ERROR,
    MASK_CHOSEN,
    // uce used
    MASK_UCE_HIGH,
    MASK_UCE_LOW,
    MASK_UCE_LOCAL,
};

TResult Controller::ChooseRank(const RankChooseInfo &rankChooseInfo, std::vector<int32_t> &tmpRankVec, uint32_t repCnt)
{
    TTP_ASSERT_RETURN(repCnt != 0, TTP_ERROR);

    auto rankMask = GenerateRankMask(rankChooseInfo);
    auto rankSize = rankMask.size();
    // 1. 每个rank都有全量数据,选一个好的就行
    if (rankSize == repCnt) {
        for (auto [rank, mask] : rankMask) {
            if (mask == MASK_NORMAL) {
                tmpRankVec.push_back(rank);
                return TTP_OK;
            }
        }
        return TTP_ERROR;
    }

    // 2. 两副本或多副本, 在dp组内要选出全量信息
    auto offset = rankSize / repCnt;
    for (auto i = 0U; i < offset; i++) {
        for (auto idx = i; idx < rankSize; idx += offset) {
            auto [rank, mask] = rankMask[idx];
            if (mask == MASK_NORMAL) {
                tmpRankVec.push_back(rank);
                break;
            }
        }
    }

    return tmpRankVec.size() == offset ? TTP_OK : TTP_ERROR;
}

RankMask Controller::GenerateRankMask(const RankChooseInfo &rankChooseInfo)
{
    auto &[step, errorRanks, rankVec] = rankChooseInfo;

    RankMask rankMask;
    for (auto rank : rankVec) {
        rankMask.emplace_back(rank, MASK_NORMAL);
    }

    AutoLock statusMapLock(statusMapLock_, TYPE_READ);
    for (auto &[curRank, mask] : rankMask) {
        if (errorRanks.find(curRank) != errorRanks.end()) {
            mask = MASK_ERROR;
            continue;
        }

        auto it = statusMap_.find(curRank);
        if (it == statusMap_.end()) {
            mask = MASK_ERROR;
            continue;
        }

        if (it->second.data_aval != TTP_STATUS_NORMAL) {
            mask = MASK_ERROR;
            continue;
        }

        bool err = (it->second.data_status != Updated || it->second.step != step);
        mask = err ? MASK_ERROR : MASK_NORMAL;

        if (err) {
            TTP_LOG_WARN("rank mask error, rank:" << curRank << ", expect step:" << step
                         << ", actual step:" << it->second.step
                         << ", data_status:" << static_cast<int>(it->second.data_status));
        }
    }

    return std::move(rankMask);
}

TResult Controller::BeginExceptionCkpt(const std::set<int32_t> &errorRanks, bool isTcpStoreOK)
{
    TTP_ASSERT_RETURN(repairStep_ > 0, TTP_ERROR);

    // select dp rank to dump
    std::vector<ActionInfo> info;
    std::unordered_map<int32_t, std::vector<GroupRecord>> sendGroup;
    TResult ret;
    uint16_t sn = actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    rankToRename_ = 0;
    int32_t idx = 0;
    AutoLock lock(dpGroupMapLock_, TYPE_READ);
    // 遍历每种优化器, 在当前种类时遍历全量dp组
    for (auto &[repCnt, dpGroups] : dpGroupListMap_) {
        for (auto &dpGroup : dpGroups) {
            std::vector<int32_t> tmpRankVec;
            RankChooseInfo rankChooseInfo {repairStep_, errorRanks, dpGroup};

            ret = ChooseRank(rankChooseInfo, tmpRankVec, repCnt);
            if (ret != TTP_OK) {
                TTP_LOG_ERROR("dp:" << IntVec2String(dpGroup) << " has no complete data." << " ret:" << ret);
                return ret;
            }

            for (auto rank : tmpRankVec) {
                sendGroup[rank].emplace_back(std::make_pair(idx, tmpRankVec));
                rankToRename_ = rankToRename_ == 0 ? rank : rankToRename_;
            }
        }
        idx++;
    }

    const uint32_t additionalUnit = 2; // record.first and record.second.size() occupy another 2 int space
    for (auto &[rank, rankList] : sendGroup) {
        uint32_t unitNum = 0;
        for (auto &record : rankList) {
            unitNum += record.second.size() + additionalUnit;
        }
        // msg format: dp1_idx dp1_size dp1_list dp2_idx dp2_size dp2_list ...
        uint32_t ckptMsgLen = sizeof(CkptMsg) + sizeof(int32_t) * unitNum;
        TTP_ASSERT_RETURN(IsMsgLenValid(ckptMsgLen), TTP_ERROR);

        AccDataBufferPtr buffer = AccDataBuffer::Create(ckptMsgLen);
        if (buffer == nullptr) {
            return TTP_ERROR;
        }
        CkptMsg *ckptMsg = static_cast<CkptMsg *>(buffer->DataPtrVoid());

        ckptMsg->step = repairStep_;
        ckptMsg->num = rankList.size();
        ckptMsg->sn = sn;
        ckptMsg->isTcpStoreOK = isTcpStoreOK;
        ckptMsg->repairId = repairId_.load();
        int32_t idx = 0;
        for (auto &record : rankList) {
            ckptMsg->ranks[idx++] = record.first;
            ckptMsg->ranks[idx++] = static_cast<int32_t>(record.second.size());
            for (auto rk : record.second) {
                ckptMsg->ranks[idx++] = rk;
            }
        }

        std::vector<int32_t> msgRank = { rank };
        info.push_back({TTP_MSG_OP_CKPT_SEND, buffer, msgRank});
    }

    ret = engine_->Process(ACTION_OP_SAVECKPT, info, true, sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("do save ckpt failed, ret: " << ret);
    }
    return ret;
}

TResult Controller::GenerateRepairMsg(const std::vector<RepairInfo> &rInfo, std::vector<ActionInfo> &info)
{
    std::map<int32_t, std::vector<RepairInfo>> repairMap;
    for (auto &ri : rInfo) {
        repairMap[ri.msgRank].push_back(ri);
    }
    auto rankList = GetMapKeysToVector(repairMap);
    uint32_t rankLen = rankList.size();
    uint32_t unitSize = sizeof(RepairMsgUnit);

    for (auto &it : repairMap) {
        uint32_t unitNum = it.second.size();
        if (IsOverflow(sizeof(int32_t), rankLen) || IsOverflow(unitSize, unitNum)) {
            TTP_LOG_ERROR("multiply overflow");
            return TTP_ERROR;
        }
        uint32_t zitParamLen = 0;
        if (repairType_ == ControllerRepairType::CRT_UPGRADE) {
            zitParamLen = sizeof(int32_t) + zitParam_.strategyParm.length() + 1; // len + char[]
        }
        uint32_t msgLen = sizeof(RepairMsg) + sizeof(int32_t) * rankLen + unitSize * unitNum + zitParamLen;
        AccDataBufferPtr buffer = AccDataBuffer::Create(msgLen);
        if (buffer == nullptr) {
            return TTP_ERROR;
        }
        RepairMsg *msg = static_cast<RepairMsg *>(buffer->DataPtrVoid());

        std::vector<int32_t> ranks = { it.first };
        msg->repairId = repairId_.load();
        msg->repairType = repairType_;
        msg->repairNum = unitNum;
        msg->step = repairStep_;
        msg->rankNum = rankLen;
        msg->sn = actionSn_.load();
        for (uint32_t i = 0; i < rankLen; i++) {
            msg->arr[i] = rankList[i];
        }
        RepairMsgUnit *uptr = reinterpret_cast<RepairMsgUnit *>(&msg->arr[rankLen]);
        for (uint32_t i = 0; i < unitNum; i++) {
            uptr[i].srcRank = it.second[i].srcRank;
            uptr[i].dstRank = it.second[i].dstRank;
            uptr[i].replicaIdx = it.second[i].replicaIdx;
            uptr[i].groupType = it.second[i].groupType;
            uptr[i].type = it.second[i].type;
        }
        if (repairType_ == ControllerRepairType::CRT_UPGRADE) {
            int32_t *ptr = reinterpret_cast<int32_t*>(reinterpret_cast<char*>(uptr) + unitNum * sizeof(RepairMsgUnit));
            *ptr = zitParam_.strategyParm.length() + 1;
            int32_t ret = strncpy_s(reinterpret_cast<char*>(ptr + 1), zitParam_.strategyParm.length() + 1,
                zitParam_.strategyParm.c_str(), zitParam_.strategyParm.length() +1);
            TTP_ASSERT_RETURN(ret == TTP_OK, TTP_ERROR);
        }
        int16_t msgType = repairType_ ==
            ControllerRepairType::CRT_UPGRADE ? TTP_MSG_OP_UPGRADE_REPAIR : TTP_MSG_OP_REPAIR;
        info.push_back({msgType, buffer, ranks});
    }
    return TTP_OK;
}

void Controller::GetAllRepairInfo(RankMask &rankMask, std::vector<RepairInfo> &rInfo, int16_t groupIdx)
{
    for (auto [rank, mask] : rankMask) {
        RepairType rt = mask == MASK_UCE_HIGH ? RepairType::RT_LOAD_REBUILD : RepairType::RT_LOAD_CKPT;
        rInfo.push_back(RepairInfo{rank, rank, rank, groupIdx, -1, rt});
    }
}

TResult Controller::RepairSelectReplica(RankMask &rankMask, std::vector<RepairInfo> &rInfo,
                                        uint32_t repCnt, int16_t groupIdx)
{
    TTP_ASSERT_RETURN(repCnt != 0, TTP_ERROR);

    auto rankSize = rankMask.size();
    for (auto i = 0U; i < rankSize; i++) {
        auto [curRank, mask] = rankMask[i];
        if (mask == MASK_NORMAL || mask == MASK_UCE_LOW) {
            continue;
        }

        RepairType rt = mask == MASK_UCE_HIGH ? RepairType::RT_UCE_HIGHLEVEL : RepairType::RT_RECV_REPAIR;

        auto offset = rankSize / repCnt;
        bool find = false;
        for (uint32_t idx = (i + offset) % rankSize; idx != i; idx = (idx + offset) % rankSize) {
            auto [repRank, repMask] = rankMask[idx];
            if (repMask == MASK_NORMAL) {
                find = true;
                rInfo.push_back(RepairInfo{curRank, repRank, curRank, groupIdx, -1, rt});
                rInfo.push_back(RepairInfo{repRank, repRank, curRank, groupIdx, -1, RepairType::RT_SEND});
                break;
            }
        }

        if (!find) {
            TTP_LOG_ERROR("all rank is abnormal! rank:" << curRank);
            return TTP_ERROR;
        }
    }
    return TTP_OK;
}

RankMask Controller::RepairCheckStatus(const std::vector<int32_t> &rankVec)
{
    RankMask rankMask;
    for (auto rank : rankVec) {
        rankMask.emplace_back(rank, MASK_NORMAL);
    }

    AutoLock statusMapLock(statusMapLock_, TYPE_READ);
    for (auto &[curRank, mask] : rankMask) {
        auto it = statusMap_.find(curRank);
        if (it == statusMap_.end()) {
            mask = MASK_ERROR;
            continue;
        }

        if (it->second.npu_status != TTP_STATUS_NORMAL) {
            if (it->second.npu_status == TTP_STATUS_UCE_HIGH || it->second.npu_status == TTP_STATUS_UCE_CORRUPTED) {
                mask = MASK_UCE_HIGH;
            } else if (it->second.npu_status == TTP_STATUS_UCE_LOW) {
                mask = MASK_UCE_LOW;
            } else {
                mask = MASK_ERROR;
            }
        }
    }

    return std::move(rankMask);
}

TResult Controller::UCERepair()
{
    TTP_ASSERT_RETURN(worldSize_ > 0 && worldSize_ <= TTP_MAX_WORLD_SIZE, TTP_ERROR);
    TTP_LOG_INFO("Start do uce repair");
    bool canRepair = (!unableRepair_) && CheckCanRepair();
    std::vector<RepairInfo> rInfo;
    int16_t idx = 0;

    AutoLock lock(dpGroupMapLock_, TYPE_READ);
    // 遍历每种优化器, 在当前种类时遍历全量dp组
    for (auto &[repCnt, dpGroups] : dpGroupListMap_) {
        for (auto &dpGroup : dpGroups) {
            auto rankMask = RepairCheckStatus(dpGroup);
            // mindspore 暂不支持周期ckpt修复
            if (!mindSpore_ && !canRepair) {
                GetAllRepairInfo(rankMask, rInfo, idx);
                continue;
            }
            auto ret = RepairSelectReplica(rankMask, rInfo, repCnt, idx);
            if (ret != TTP_OK) {
                return TTP_ERROR;
            }
        }
        idx++;
    }

    auto ret = RepairProcess(rInfo);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " do uce repair failed, ret:" << ret);
        return TTP_ERROR;
    }
    return TTP_OK;
}

TResult Controller::UpGradeRepair()
{
    auto isolateRanks = GetIsolateRanks();
    TResult ret = UpGradeCommGroupRepair(isolateRanks);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("Upgrading rebuild common group failed.");
        return TTP_ERROR;
    }
    std::vector<RepairInfo> rInfo;
    ret = PrepareRepairMsg(rInfo, isolateRanks, true);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("Upgrading prepair msg failed.");
        return TTP_ERROR;
    }

    ret = RepairProcess(rInfo);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << this->rank_ << " do upgrade recovery repair failed, ret:" << ret);
        return TTP_ERROR;
    }
    return TTP_OK;
}

TResult Controller::ARFWait(bool isFirst)
{
    uint32_t waitTimes = 0;
    while (waitTimes <= ARF_WAIT_ADD_TIME) {
        // arf等待过程中mindx通知做dump或exit情况
        if (repairEvent_ == MindXEvent::MINDX_EVENT_DUMP) {
            TTP_LOG_WARN("mindx notify to dump, controller do arf repair failed");
            return TTP_ERROR;
        }

        if (repairEvent_ == MindXEvent::MINDX_EVENT_EXIT) {
            TTP_LOG_WARN("mindx notify to exit, controller do arf repair failed");
            return TTP_STOP_SERVICE;
        }

        // arf等待过程中mindx通知做降级
        if (repairEvent_ == MindXEvent::MINDX_EVENT_DOWNGRADE) {
            TTP_LOG_WARN("mindx notify to downgrade, controller do arf repair failed");
            return TTP_DOWNGRADE;
        }
        auto result = CheckTrainStatus();
        if ((isFirst && result == TTP_REPAIR) || result == TTP_NEED_RETRY) {
            return TTP_OK;
        }

        if (result == TTP_ERROR) {
            TTP_LOG_ERROR("controller do arf repair failed, some rank error");
            return TTP_ERROR;
        }

        waitTimes++;
        usleep(TTP_WAIT_TIME_1MS);
        TTP_LOG_LIMIT_WARN(LOG_PRINT_INTERVAL, "arf wait new workers register... ");
    }

    TTP_LOG_ERROR("controller do arf repair failed, wait timeout");
    return TTP_ERROR;
}

// 不感知是节点重启还是worker重启,由mindx控制,需要重启的rank都在errorRanks中
TResult Controller::ARFRepair()
{
    TTP_LOG_INFO("Start do arf repair");

    auto errRanks = GetErrorRanks();
    IsolateRanksSetStatus(errRanks, errRanks);

    auto ret = ARFWait(true);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("arf wait new workers register failed");
        return ret;
    }

    ret = HcclCommGroupRepair(errRanks);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("rebuild HCCL comm group failed");
        return ret;
    }

    ret = ARFWait(false);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("arf wait all workers finish failed");
        return ret;
    }

    std::vector<RepairInfo> info;
    ret = PrepareRepairMsg(info, errRanks);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("prepare repair msg failed");
        return ret;
    }

    ret = RepairProcess(info);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller do arf repair failed");
        return ret;
    }

    return TTP_OK;
}

TResult Controller::MindXReConfirmStrategy()
{
    TTP_LOG_WARN("wait mindx renotify all fault ranks...");
    auto ret = mindXEngine_->Wait(waitMindxTimes_);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }

    if (repairEvent_ != MindXEvent::MINDX_EVENT_NOTIFY_FAULT_RANKS) {
        TTP_LOG_WARN("mindx calling action unexpected! prepare to exit");
        return TTP_STOP_SERVICE;
    }

    return MindXConfirmStrategy();
}

TResult Controller::WaitDumpOrExitStrategy()
{
    if (!mindXEngine_->IsRegistered()) {
        return TTP_STOP_SERVICE;
    }

    auto ret = mindXEngine_->Wait(waitMindxTimes_);
    if (ret != TTP_OK) {
        return TTP_STOP_SERVICE;
    }

    if (repairEvent_ == MindXEvent::MINDX_EVENT_DUMP) {
        TTP_LOG_WARN("controller do last repair op failed, mindx notify to dump");
        return TTP_ERROR;
    }

    if (repairEvent_ == MindXEvent::MINDX_EVENT_EXIT) {
        TTP_LOG_WARN("controller do last repair op failed, mindx notify to exit");
        return TTP_STOP_SERVICE;
    }

    TTP_LOG_WARN("controller do last repair op failed, mindx notify to " << static_cast<int>(repairEvent_));
    return TTP_STOP_SERVICE;
}

TResult Controller::PrepareRepairMsg(std::vector<RepairInfo> &rInfo, const std::set<int32_t> &errRanks, bool isZit)
{
    RankChooseInfo rankChooseInfo;
    rankChooseInfo.step = repairStep_;
    rankChooseInfo.errorRanks = errRanks;
    bool canRepair = CheckCanRepair(isZit);
    int16_t idx = 0;

    AutoLock lock(dpGroupMapLock_, TYPE_READ);
    // 遍历每种优化器, 在当前种类时遍历全量dp组
    for (auto &[repCnt, dpGroups] : dpGroupListMap_) {
        for (auto &dpGroup : dpGroups) {
            rankChooseInfo.rankVec = dpGroup;
            auto rankMask = GenerateRankMask(rankChooseInfo);
            // mindspore 暂不支持周期ckpt修复
            if (!mindSpore_ && !canRepair) {
                GetAllRepairInfo(rankMask, rInfo, idx);
                continue;
            }
            TResult ret = RepairSelectReplica(rankMask, rInfo, repCnt, idx);
            if (ret != TTP_OK) {
                return TTP_ERROR;
            }
        }
        idx++;
    }

    return TTP_OK;
}

TResult Controller::RepairProcess(const std::vector<RepairInfo> &rInfo)
{
    std::vector<ActionInfo> info;
    TResult ret = GenerateRepairMsg(rInfo, info);
    if (ret != TTP_OK) {
        return TTP_ERROR;
    }

    uint16_t sn = actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }
    ActionOp action_type = repairType_ ==
        ControllerRepairType::CRT_UPGRADE ? ACTION_OP_UPGRADE_REPAIR : ACTION_OP_REPAIR;
    ret = engine_->Process(action_type, info, true, sn, 0);

    return ret != TTP_OK ? TTP_ERROR : TTP_OK;
}

// ip list拼接为string
std::string Controller::BuildStr4BackupCtrl()
{
    std::string ipList = "";
    for (uint32_t i = 0; i < backupInfoList_.size(); i++) {
        BackupInfo &info = backupInfoList_[i];
        ipList += std::to_string(info.rank) + ":" + info.ip + "|";
    }
    // remove last "|"
    if (ipList.size() > 1) {
        ipList = ipList.substr(0, ipList.length() - 1);
    }
    return ipList;
}

TResult Controller::BroadcastMsgStuff(BroadcastIpMsg *broadcastIpMsg, std::string &ipList)
{
    if (broadcastIpMsg == nullptr) {
        TTP_LOG_ERROR("BroadcastIpMsg pointer is null.");
        return TTP_ERROR;
    }
    broadcastIpMsg->sn = actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    const char strEndFlag = '\0';
    broadcastIpMsg->ipLen = ipList.length() + sizeof(strEndFlag);
    broadcastIpMsg->enableZIT = (zitSwitch_ ? 1 : 0);
    broadcastIpMsg->enableARF = (arfSwitch_ ? 1 : 0);
    std::copy_n(ipList.c_str(), broadcastIpMsg->ipLen, broadcastIpMsg->arr);
    return TTP_OK;
}

TResult Controller::BroadcastCrtlIps()
{
    TTP_ASSERT_RETURN(worldSize_ > 0 && worldSize_ <= TTP_MAX_WORLD_SIZE, TTP_ERROR);

    // mindSpore的reportinfo会在第一步训练开始以后才调用，因此不能使用reportedCnt_进行判断
    if (mindSpore_) {
        AutoLock statusMapLock(statusMapLock_, TYPE_READ);
        if (statusMap_.size() != static_cast<uint32_t>(worldSize_)) {
            return TTP_WAIT_CHECK;
        }
    } else {
        // pytorch的broadcast需要在reportinfo以后才通知，否则processor启动备controller时可能跟数据集创建并发触发glibc错误
        if (reportedCnt_ != worldSize_) {
            return TTP_WAIT_CHECK;
        }
    }

    backupInfoList_ = SelectBackUpController();
    if (backupInfoList_.size() == 0) {
        TTP_LOG_WARN("can't select backup controller, maybe only one node..");
        return TTP_OK;
    }

    std::string ipList = BuildStr4BackupCtrl();
    TTP_LOG_INFO("start broadcast, ip: " << ipList);
    uint32_t broadcastIpMsgLen = sizeof(BroadcastIpMsg) + ipList.length() + 1;

    AccDataBufferPtr buffer = AccDataBuffer::Create(broadcastIpMsgLen);
    if (buffer == nullptr) {
        return TTP_ERROR;
    }
    BroadcastIpMsg *broadcastIpMsg = static_cast<BroadcastIpMsg *>(buffer->DataPtrVoid());

    if (BroadcastMsgStuff(broadcastIpMsg, ipList) != TTP_OK) {
        return TTP_ERROR;
    }

    auto ranks = GetAllLinkRanks();
    std::vector<ActionInfo> info {{TTP_MSG_OP_CTRL_NOTIFY, buffer, ranks}};
    TResult ret = engine_->Process(ACTION_OP_BROADCAST_IP, info, false, broadcastIpMsg->sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " broadcast ip failed");
        return TTP_ERROR;
    }
    TTP_LOG_INFO("finish broadcast, ip: " << ipList);
    return TTP_OK;
}

TResult Controller::Rename()
{
    TTP_LOG_INFO("Start rename.");
    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(CommonMsg));
    if (buffer == nullptr) {
        return TTP_ERROR;
    }

    CommonMsg *renameMsg = static_cast<CommonMsg *>(buffer->DataPtrVoid());
    renameMsg->rank = rankToRename_;
    renameMsg->sn = actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    std::vector<int32_t> ranks {rankToRename_};
    std::vector<ActionInfo> info {{TTP_MSG_OP_RENAME, buffer, ranks}};
    TResult ret = engine_->Process(ACTION_OP_RENAME, info, true, renameMsg->sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " rename failed");
        return TTP_ERROR;
    }

    TTP_LOG_INFO("finish rename, rank:" << rankToRename_);
    return TTP_OK;
}

TResult Controller::CheckTrainStatus()
{
    int64_t curTime = GetNowTime();
    bool errorFlag = false;
    uint32_t preTrainFinishRankNum = 0;
    uint32_t initRankNum = 0;
    uint32_t normalRankNum = 0;
    int64_t maxStep = 0;

    statusMapLock_.LockRead();
    for (auto &[rank, status] : statusMap_) {
        if (status.run_status == TTP_STATUS_EXIT) {
            isPorcessorExit_ = true;
            continue;
        } else if (status.run_status == TTP_STATUS_ISOLATE || status.run_status == TTP_STATUS_OFFLINE) {
            continue;
        } else if (status.run_status == TTP_STATUS_PREREPAIR_FINISH) {
            preTrainFinishRankNum++;
            continue;
        } else if (status.run_status == TTP_STATUS_INIT_FINISH) {
            initRankNum++;
            continue;
        }

        if ((curTime - (status.lastUpdateTime)) / TIME_CHECKER_INTERVAL > HEART_BEAT_MAX_LOSS) {
            status.data_aval = TTP_STATUS_ABNORMAL;
        }

        if ((status.run_status != TTP_STATUS_NORMAL) ||
            (status.data_aval != TTP_STATUS_NORMAL) ||
            (status.npu_status != TTP_STATUS_NORMAL)) {
            unableRepair_ |= (status.npu_status == TTP_STATUS_UCE_CORRUPTED);
            hcclFlag_ |= SET_HCCL_BIT(status.npu_status == TTP_STATUS_HCCL_FAILED);
            hcclFlag_ |= SET_UCE_BIT(status.npu_status == TTP_STATUS_UCE_HIGH);
            errorFlag = true;
            TTP_LOG_WARN("status error, rank:" << rank << STATUS_MAP_VAL_PRINT(status));
            continue;
        }

        maxStep = maxStep > status.step ? maxStep : status.step;
        normalRankNum++;
    }
    statusMapLock_.UnLock();

    initRankNum += preTrainFinishRankNum;
    bool isReadyForUpgrade = preTrainFinishRankNum > 0 && (preTrainFinishRankNum + normalRankNum == statusMap_.size());
    bool canRebuildGroup = initRankNum > 0 && (initRankNum + normalRankNum == statusMap_.size());

    ReportMindXRepairResult(maxStep);

    if (errorFlag) {
        return TTP_ERROR;
    } else if (isReadyForUpgrade) {
        return TTP_NEED_RETRY; // can repair + rollback
    } else if (canRebuildGroup) {
        return TTP_REPAIR; // can rebuild group
    } else {
        return TTP_OK;
    }
}

void Controller::ReportMindXRepairResult(int64_t &step)
{
    // 复训迭代一步后再上报修复成功消息
    if (step > repairStep_ && isNeedToReportResult_.load()) {
        mindXEngine_->ReportResult(
            RepairResult::REPAIR_SUCCESS, "Mindio do repair operation ok", errorRankMsg_, errorRankLock_);
        isNeedToReportResult_.store(false);
        loadCkptRepairStep_.store(-1);
    }
}

std::pair<int64_t, std::vector<int32_t>> Controller::SelectLockStep()
{
    AutoLock statusMapLock(statusMapLock_, TYPE_READ);
    AutoLock errorRankLock(errorRankLock_, TYPE_READ);

    int64_t step = -1;
    int64_t curTime = GetNowTime();
    std::vector<int32_t> ranks;

    for (auto &[rank, status] : statusMap_) {
        step = step > status.step ? step : status.step;

        if ((status.run_status != TTP_STATUS_NORMAL) ||
            ((curTime - status.lastUpdateTime) / TIME_CHECKER_INTERVAL > HEART_BEAT_MAX_LOSS) ||
            (status.data_aval != TTP_STATUS_NORMAL) ||
            (status.npu_status != TTP_STATUS_NORMAL) ||
            (errorRankMsg_.count(rank))) {
            continue;
        }

        ranks.push_back(rank);
    }

    return {step, ranks};
}

void Controller::SelectErrorRanks()
{
    int64_t curTime = GetNowTime();

    // Check Heartbeat
    AutoLock statusMapLock(statusMapLock_, TYPE_WRITE);
    AutoLock errorRankLock(errorRankLock_, TYPE_WRITE);

    for (auto &[rank, status] : statusMap_) {
        if ((curTime - status.lastUpdateTime) / TIME_CHECKER_INTERVAL > HEART_BEAT_MAX_LOSS) {
            status.data_aval = TTP_STATUS_ABNORMAL;
        }

        if (errorRankMsg_.find(rank) != errorRankMsg_.end()) {
            continue;
        }
        if ((status.npu_status != TTP_STATUS_NORMAL) ||
            (status.run_status != TTP_STATUS_NORMAL) ||
            (status.data_aval != TTP_STATUS_NORMAL)) {
            TTP_LOG_WARN("detected abnormal heartbeat, rank:" << rank << STATUS_MAP_VAL_PRINT(status));
            if (status.npu_status == TTP_STATUS_HCCL_FAILED) {
                errorRankMsg_[rank] = HCCL_ERROR;
            } else if (status.npu_status == TTP_STATUS_UCE_HIGH || status.npu_status == TTP_STATUS_UCE_LOW ||
                       status.npu_status == TTP_STATUS_UCE_CORRUPTED) {
                errorRankMsg_[rank] = UCE_ERROR;
            } else {
                errorRankMsg_[rank] = PROCESSES_ERROR;
            }
        }
    }

    auto errRanks = GetMapKeysToVector(errorRankMsg_);
    TTP_LOG_INFO("selected error ranks: " << IntVec2String(errRanks) << ", step: " << repairStep_);
}

bool Controller::CheckDpGroup()
{
    AutoLock lock(dpGroupMapLock_, TYPE_READ);

    TTP_ASSERT_RETURN(repairStep_ > 0, false);
    TTP_ASSERT_RETURN(!dpGroupListMap_.empty(), false);

    for (auto &[repCnt, dpGroups] : dpGroupListMap_) {
        TTP_ASSERT_RETURN(!dpGroups.empty(), false);

        auto dpsize = dpGroups.begin()->size();
        for (auto &dpGroup : dpGroups) {
            // 一种优化器，每个dp组大小要一样
            TTP_ASSERT_RETURN(dpGroup.size() == dpsize, false);
        }

        // 一种优化器，所有dp组大小加起来要等于worldsize
        TTP_ASSERT_RETURN(dpsize * dpGroups.size() == worldSize_, false);

        // 目前不支持副本数不整除情况
        TTP_ASSERT_RETURN(dpsize % repCnt == 0, false);
    }

    return true;
}

bool Controller::CheckCanRepair(bool isZit)
{
    // hccl重计算场景不计算副本，暂不考虑故障叠加
    if (ONLY_HCCL_BIT(hcclFlag_)) {
        return true;
    }

    // 复用临终遗言的判断逻辑
    AutoLock lock(dpGroupMapLock_, TYPE_READ);
    std::set<int32_t> errRanks;
    if (isZit) {
        errRanks = GetIsolateRanks();
    } else {
        errRanks = GetErrorRanks();
    }
    // 遍历每种优化器, 在当前种类时遍历全量dp组
    for (auto &[repCnt, dpGroups] : dpGroupListMap_) {
        for (auto &dpGroup : dpGroups) {
            std::vector<int32_t> tmpRankVec;
            RankChooseInfo rankChooseInfo {repairStep_, errRanks, dpGroup};
            auto ret = ChooseRank(rankChooseInfo, tmpRankVec, repCnt);
            if (ret != TTP_OK) {
                TTP_LOG_ERROR("dp:" << IntVec2String(dpGroup) << " has no complete data." << " ret:" << ret);
                return false;
            }
        }
    }

    return true;
}

ControllerRepairType Controller::ConfirmRepairType()
{
    ControllerRepairType type = ControllerRepairType::CRT_DUMP;
    AutoLock errorRankLock(errorRankLock_, TYPE_READ);
    AutoLock statusMapLock(statusMapLock_, TYPE_READ);

    auto [retryFlag, nodeErrorFlag] = ConfirmRepairFlag();
    TTP_LOG_INFO("uceSwitch:" << uceSwitch_ << ", arfSwitch:" << arfSwitch_ << ", zitSwitch:" << zitSwitch_ <<
        ", retryFlag:" << retryFlag << ", nodeErrorFlag:" << nodeErrorFlag << ", hcclFlag:" <<hcclFlag_);

    if (nodeErrorFlag || !retryFlag) {
        if (arfSwitch_ && mindXEngine_->IsRegistered()) {
            type = ControllerRepairType::CRT_ARF;  // ARF need mindx registration success
        } else if (zitSwitch_ && mindXEngine_->IsRegistered()) {
            type = ControllerRepairType::CRT_DOWNGRADE;
        } else {
            type = ControllerRepairType::CRT_DUMP;
        }
    } else if (retryFlag && (uceSwitch_ || hcclFlag_)) {
        type = ControllerRepairType::CRT_RETRY;
    }

    return type;
}

std::pair<bool, bool> Controller::ConfirmRepairFlag()
{
    bool retryFlag = false;
    bool nodeErrorFlag = false;

    for (auto [rank, errType] : errorRankMsg_) {
        auto &status = statusMap_[rank];

        // statusMap_中npu状态异常，先认为走retry
        if (status.npu_status != TTP_STATUS_NORMAL) {
            retryFlag = true;
        }

        // statusMap_中心跳或软件异常，认为走节点级
        if ((status.run_status != TTP_STATUS_NORMAL) || (status.data_aval != TTP_STATUS_NORMAL)) {
            nodeErrorFlag = true;
        }

        // errorRankMsg_中错误类型再进行一次覆盖
        if (errType == PROCESSES_ERROR) {
            nodeErrorFlag = true;
        } else {
            retryFlag = true;
        }
    }

    return {retryFlag, nodeErrorFlag};
}

TResult Controller::HandleDumpStatus(const std::set<int32_t> &errorRanks)
{
    // check if agent is available
    bool isTcpStoreOK = false;
    TResult ret;
    if (!mindSpore_) {
        ret = HandleTcpStoreError(isTcpStoreOK);
        if (ret != TTP_OK) {
            TTP_LOG_ERROR("Controller: " << rank_ << " find tcp store server failed and start failed!");
            return TTP_ERROR;
        }
    }

    return BeginExceptionCkpt(errorRanks, isTcpStoreOK);
}

bool Controller::IsBackupToMaster()
{
    if (isMasterCtrl_.load()) {
        return false;
    }
    TTP_ASSERT_RETURN(worldSize_ > 0 && worldSize_ <= TTP_MAX_WORLD_SIZE, false);
    TTP_ASSERT_RETURN(controllerIdx_ <= BACKUP_CONTROLLER_NUM, false);

    AutoLock statusMapLock(statusMapLock_, TYPE_WRITE);
    uint32_t statusMapSize = statusMap_.size();
    if (statusMapSize < (static_cast<uint32_t>(worldSize_) >> 1)) {
        return false;
    }

    isMasterCtrl_.store(true);
    isBackupToMaster_.store(true);
    actionSn_.fetch_add(CONTROLLER_SN_GENERATION * controllerIdx_);
    repairId_.fetch_add(CONTROLLER_SN_GENERATION * controllerIdx_);
    TTP_LOG_INFO("Controller: BackupToMaster, actionSn: " << actionSn_.load());
    if (IsActionSnOK() == TTP_ERROR) {
        TTP_LOG_ERROR("IsBackupToMaster: action serial number occur integer wrap.");
    }
    mindXEngine_->Register2MindX();
    sleep(SLEEP_FOR_PROCESSOR_CONNECT);

    for (int32_t i = 0; i < worldSize_; i++) { // set the initial status of the error node
        if (statusMap_.find(i) == statusMap_.end()) {
            statusMap_[i].data_aval = TTP_STATUS_ABNORMAL;
        }
    }

    TTP_LOG_INFO("backup controller to master, controller rank:" << rank_ <<
                 ", world size: " << worldSize_);
    return true;
}

TResult Controller::HandleNewConnection(const AccConnReq &req, const AccTcpLinkComplexPtr &link)
{
    TTP_LOG_DEBUG("handle new connection from rank:" << req.rankId);
    if (req.rankId < 0 || req.rankId >= worldSize_) {
        TTP_LOG_ERROR("controller:" << rank_ << " invalid rankId:" << req.rankId << " from new connection request.");
        return TTP_ERROR;
    }

    AutoLock linkLock(linkMapLock_, TYPE_WRITE);
    auto &rankLinkMap = repairEvent_ == MindXEvent::MINDX_EVENT_HOT_SWITCH ? rankLinkMapTmp_ : rankLinkMap_;
    auto &linkIdMap = repairEvent_ == MindXEvent::MINDX_EVENT_HOT_SWITCH ? linkIdMapTmp_ : linkIdMap_;
    auto it = rankLinkMap.find(req.rankId);
    if (it != rankLinkMap.end()) {
        TTP_LOG_LIMIT_WARN(LOG_PRINT_INTERVAL, "rank:" << req.rankId << " already exist one link in map");
        it->second->Close();
    }

    rankLinkMap[req.rankId] = link;
    linkIdMap[link->Id()] = req.rankId;

    return TTP_OK;
}

TResult Controller::HandleLinkBroken(const AccTcpLinkComplexPtr &link)
{
    if (link.Get() == nullptr) {
        TTP_LOG_ERROR("invalid tcp link");
        return TTP_ERROR;
    }

    AutoLock lock(linkMapLock_, TYPE_WRITE);
    auto &rankLinkMap = repairEvent_ == MindXEvent::MINDX_EVENT_HOT_SWITCH ? rankLinkMapTmp_ : rankLinkMap_;
    auto &linkIdMap = repairEvent_ == MindXEvent::MINDX_EVENT_HOT_SWITCH ? linkIdMapTmp_ : linkIdMap_;
    auto linkId = link->Id();
    auto iter = linkIdMap.find(linkId);
    if (iter == linkIdMap.end()) {
        TTP_LOG_WARN("remove broken link, linkId:" << linkId << ", but not find in linkIdMap");
        return TTP_ERROR;
    }

    auto rank = iter->second;
    auto itr = rankLinkMap.find(rank);
    if (itr == rankLinkMap.end()) {
        TTP_LOG_WARN("remove broken link, linkId:" << linkId << ", rank:" << rank << " , but not find in rankLinkMap");
        return TTP_ERROR;
    }

    linkIdMap.erase(iter);
    rankLinkMap.erase(itr);
    TTP_LOG_DEBUG("Remove broken link, linkId:" << linkId << ", rank:" << rank);

    return TTP_OK;
}

std::vector<int32_t> Controller::GetAllLinkRanks(bool excludeError)
{
    AutoLock linkLock(linkMapLock_, TYPE_READ);
    AutoLock errorRankLock(errorRankLock_, TYPE_READ);
    std::vector<int32_t> ranks;

    for (auto &[rank, _] : rankLinkMap_) {
        if (excludeError) {
            // 属于故障卡，并且标记为进程类故障，不发送
            auto itr = errorRankMsg_.find(rank);
            if (itr != errorRankMsg_.end() && itr->second == PROCESSES_ERROR) {
                continue;
            }
        }
        ranks.push_back(rank);
    }

    return std::move(ranks);
}

TResult Controller::SendMsg(int16_t msgType, const AccDataBufferPtr &d, std::vector<int32_t> &targetRanks,
                            const std::vector<AccDataBufferPtr> &cbCtx)
{
    std::vector<int32_t> failedList;
    uint32_t rankNum = targetRanks.size();

    AutoLock linkLock(linkMapLock_, TYPE_READ);
    for (uint32_t i = 0; i < rankNum; i++) {
        int32_t curRank = targetRanks.at(i);
        if (curRank == INVALID_RANK_ID) {
            continue;
        }

        auto it = rankLinkMap_.find(curRank);
        if (it == rankLinkMap_.end()) {
            TTP_LOG_WARN("not find rank link, rank:" << curRank << " msg_type:" << StrMsgOpCode(msgType));
            failedList.push_back(curRank);
            continue;
        }

        if (it->second->NonBlockSend(msgType, d, cbCtx.at(i)) != 0) {
            TTP_LOG_WARN("send  msg failed, rank:" << curRank << " msg_type:" << StrMsgOpCode(msgType));
            failedList.push_back(curRank);
            continue;
        }

        targetRanks[i] = INVALID_RANK_ID;
    }

    if (!failedList.empty()) {
        TTP_LOG_INFO("not all msg send ok, msg_type:" << StrMsgOpCode(msgType) <<
                     " failed_ranks:" << IntVec2String(failedList));
        return TTP_ERROR;
    }
    return TTP_OK;
}

void Controller::ChooseIsolateRanks()
{
    // Use minimum of NONE-ZERO {origin_dp_size, dp_cp_size, dp_ep_size} will be ok; but this will couple parallel logic
    // accumulate origin-dp, dp-cp and dp-ep groups to simplify traversal and lock/unlock
    std::vector<GroupMap> allDpGroups;
    dpGroupMapLock_.LockRead();
    for (const auto& group : dpGroupListMap_) {
        allDpGroups.push_back(group.dpGroups);
    }
    dpGroupMapLock_ .UnLock();
    originDpLock_.LockRead();
    if (!originDpSet_.empty()) {
        allDpGroups.push_back(originDpSet_);
    }
    originDpLock_.UnLock();

    // 转换每张卡在dp组内的idx
    std::vector<std::vector<uint32_t>> dpIdx(allDpGroups.size(), std::vector<uint32_t>(worldSize_));
    for (uint32_t groupIdx = 0; groupIdx < allDpGroups.size(); groupIdx++) {
        for (const auto &rankVec : allDpGroups[groupIdx]) {
            uint32_t i = 0;
            for (auto rank : rankVec) {
                dpIdx[groupIdx][rank] = i++;
            }
        }
    }

    // 得到一共坏了哪几路dp
    std::vector<std::set<uint32_t>> errorIdx(allDpGroups.size());
    auto errRanks = GetErrorRanks();
    for (auto rk : errRanks) {
        for (uint32_t groupIdx = 0; groupIdx < allDpGroups.size(); groupIdx++) {
            errorIdx[groupIdx].insert(dpIdx[groupIdx][rk]);
        }
    }

    // 遍历每个dp组，取坏掉的idx的卡，得到隔离的所有卡
    std::set<int32_t> tmpIsolateRanks;
    for (uint32_t groupIdx = 0; groupIdx < allDpGroups.size(); groupIdx++) {
        for (auto &rankVec : allDpGroups[groupIdx]) {
            for (auto idx : errorIdx[groupIdx]) {
                tmpIsolateRanks.insert(rankVec[idx]);
            }
        }
    }
    zitParam_.isolateRanks = tmpIsolateRanks;
    TTP_LOG_INFO("controller:" << rank_ << " choose isolate ranks: " << IntSet2String(zitParam_.isolateRanks));
}

TResult Controller::GenerateDownGradeMsgInner(std::vector<ActionInfo> &info, const std::vector<int32_t> &msgRanks,
                                              const std::vector<std::vector<int32_t>> &dpGroups,
                                              const std::vector<int32_t> &normalRanks, uint16_t sn)
{
    TTP_ASSERT_RETURN(!dpGroups.empty(), TTP_ERROR);
    uint32_t rankNum = msgRanks.size();
    uint32_t normalSize = normalRanks.size();
    TTP_ASSERT_RETURN(rankNum > 0, TTP_ERROR);

    // groupIdx: [-1]->Global Group; [0]->dpcp; [1]->dpep;
    const uint32_t additionalUnit = 2; // groupIdx and size occupy another 2 int space
    uint32_t msgRanksNum = normalSize + additionalUnit;
    for (auto &rankVec : dpGroups) {
        msgRanksNum += rankVec.size() + additionalUnit;
    }
    if (IsOverflow(sizeof(int32_t), msgRanksNum)) {
        TTP_LOG_ERROR("Overflow calculation occurred before malloc downgrade running msg.");
        return TTP_ERROR;
    }
    uint32_t zitParamLen = sizeof(int32_t) + zitParam_.strategyParm.length() + 1; // len + char[]
    uint32_t msgLen = sizeof(DowngradeRunMsg) + sizeof(int32_t) * msgRanksNum + zitParamLen;
    TTP_ASSERT_RETURN(IsMsgLenValid(msgLen), TTP_ERROR);

    AccDataBufferPtr buffer = AccDataBuffer::Create(msgLen);
    if (buffer == nullptr) {
        return TTP_ERROR;
    }
    DowngradeRunMsg *msg = static_cast<DowngradeRunMsg *>(buffer->DataPtrVoid());

    msg->num = 1 + dpGroups.size();
    msg->repairId = repairId_.load();
    msg->sn = sn;
    uint32_t idx = 0;
    msg->ranks[idx++] = -1;  // global comm group
    msg->ranks[idx++] = static_cast<int32_t>(normalSize);
    for (auto rk : normalRanks) {
        msg->ranks[idx++] = rk;
    }
    for (uint32_t dpIdx = 0; dpIdx != dpGroups.size(); dpIdx++) {
        msg->ranks[idx++] = static_cast<int32_t>(dpIdx);
        msg->ranks[idx++] = static_cast<int32_t>(dpGroups[dpIdx].size());
        for (auto rk : dpGroups[dpIdx]) {
            msg->ranks[idx++] = rk;
        }
    }
    msg->ranks[idx++] = static_cast<int32_t>(zitParam_.strategyParm.length() +1);
    int32_t ret = strncpy_s(reinterpret_cast<char*>(&msg->ranks[idx]), zitParam_.strategyParm.length() +1,
        zitParam_.strategyParm.c_str(), zitParam_.strategyParm.length() +1);
    TTP_ASSERT_RETURN(ret == TTP_OK, TTP_ERROR);
    info.push_back({TTP_MSG_OP_DOWNGRADE_REBUILD, buffer, msgRanks});
    return TTP_OK;
}

TResult Controller::GenerateDownGradeMsg(std::vector<ActionInfo> &info, std::vector<int32_t> &normalRanks, uint16_t sn)
{
    TTP_ASSERT_RETURN(worldSize_ > 0 && worldSize_ <= TTP_MAX_WORLD_SIZE, TTP_ERROR);
    auto isolateRanks = GetIsolateRanks();
    uint32_t normalSize = static_cast<uint32_t>(worldSize_) - isolateRanks.size();

    // rankDpMap: map<rank, vector<dpGroup>>; default order of vector<dpGroup>: 0, dpcp; 1, dpep.
    std::unordered_map<int32_t, std::vector<std::vector<int32_t>>> rankDpMap;
    AutoLock lock(dpGroupMapLock_, TYPE_READ);
    for (uint32_t groupIdx = 0; groupIdx != dpGroupListMap_.size(); groupIdx++) {
        for (auto rankVec : dpGroupListMap_[groupIdx].dpGroups) {
            std::vector<int32_t> dpRanks;

            std::copy_if(rankVec.begin(), rankVec.end(), std::back_inserter(dpRanks),
                [&isolateRanks](int rk) { return isolateRanks.count(rk) == 0; });

            for (auto rk : dpRanks) {
                rankDpMap[rk].push_back(dpRanks);
            }

            if (groupIdx == 0) {
                normalRanks.insert(normalRanks.end(), dpRanks.begin(), dpRanks.end());
            }
        }
    }

    if (normalRanks.size() != normalSize) {
        TTP_LOG_ERROR("downgrade run found normal ranks num not meet expectations, exptected: " <<
            normalSize << " actual: " << normalRanks.size());
        return TTP_ERROR;
    }
    std::sort(normalRanks.begin(), normalRanks.end());

    std::map<std::vector<std::vector<int32_t>>, std::vector<int32_t>> dpRankMap;
    for (auto &[rk, dpGroups] : rankDpMap) {
        dpRankMap[dpGroups].push_back(rk);
    }

    for (auto &[dpGroups, msgRanks] : dpRankMap) {
        if (GenerateDownGradeMsgInner(info, msgRanks, dpGroups, normalRanks, sn) != TTP_OK) {
            return TTP_ERROR;
        }
    }
    return TTP_OK;
}

TResult Controller::DowngradeNotifyNormalRanks()
{
    std::vector<ActionInfo> info;
    std::vector<int32_t> normalRanks;
    uint16_t sn = this->actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    TResult ret = GenerateDownGradeMsg(info, normalRanks, sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("do downgrade running failed, generate msg error, ret: " << ret);
        return ret;
    }

    ret = engine_->Process(ACTION_OP_DOWNGRADE_REBUILD, info, true, sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("do downgrade running failed, ret: " << ret);
        return ret;
    }
    TTP_LOG_INFO("controller:" << rank_ << " notify normal rank rebuild success.");

    ret = DoNoDataAction(normalRanks, TTP_MSG_OP_NOTIFY_NORMAL, ACTION_OP_NOTIFY_NORMAL, true);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " notify processor normal failed");
        return TTP_ERROR;
    }
    TTP_LOG_INFO("controller:" << rank_ << " notify downgrade running success.");
    return TTP_OK;
}

TResult Controller::IsolateRanksSetStatus(std::set<int32_t> &isolateRanks, std::set<int32_t> &errRanks)
{
    AutoLock statusMapLock(statusMapLock_, TYPE_WRITE);
    for (auto rank : isolateRanks) {
        auto it = statusMap_.find(rank);
        if (it == statusMap_.end()) {
            TTP_LOG_ERROR("ranks status not find! rank:" << rank);
            return TTP_ERROR;
        }

        // pretrain finish
        if (it->second.run_status == TTP_STATUS_ISOLATE || it->second.run_status == TTP_STATUS_INIT_FINISH
            || it->second.run_status == TTP_STATUS_PREREPAIR_FINISH) {
            continue;
        }

        if (errRanks.find(rank) != errRanks.end()) {
            it->second.run_status = TTP_STATUS_OFFLINE; // error rank set offline
        } else {
            it->second.run_status = TTP_STATUS_PREREPAIR_FINISH; // isolate rank set pre repair finish
        }
    }

    return TTP_OK;
}

TResult Controller::IsActionSnOK()    // Check sn range, detect integer wrap
{
    TTP_ASSERT_RETURN(controllerIdx_ <= BACKUP_CONTROLLER_NUM, TTP_ERROR);
    uint16_t backupOffset = CONTROLLER_SN_GENERATION * controllerIdx_;
    auto sn = actionSn_.load();
    if (sn < 1 + backupOffset || sn > CONTROLLER_SN_GENERATION + backupOffset) {
        TTP_LOG_ERROR("action serial number occurs integer wrap. isBackupToMaster: " << isBackupToMaster_.load());
        return TTP_ERROR;
    }
    return TTP_OK;
}

TResult Controller::MarkNoReponseRanks(const std::vector<int32_t> &noResponseRanks)
{
    if (noResponseRanks.empty()) {
        return TTP_OK;
    }

    AutoLock statusMapLock(statusMapLock_, TYPE_WRITE);
    for (auto rank : noResponseRanks) {
        auto it = statusMap_.find(rank);
        if (it != statusMap_.end()) {
            it->second.data_aval = TTP_STATUS_ABNORMAL;
            TTP_LOG_INFO("rank:" << rank << ", no response after msg sent, data_aval is set to TTP_STATUS_ABNORMAL");
        }
    }

    return TTP_OK;
}

TResult Controller::HcclCommGroupRepair(const std::set<int32_t> &isolateRanks)
{
    uint32_t msgLen = sizeof(RebuildGroupMsg) + isolateRanks.size() * sizeof(int32_t);
    TTP_ASSERT_RETURN(IsMsgLenValid(msgLen), TTP_ERROR);

    AccDataBufferPtr buffer = AccDataBuffer::Create(msgLen);
    if (buffer == nullptr) {
        return TTP_ERROR;
    }
    RebuildGroupMsg *ptcommMsg = static_cast<RebuildGroupMsg *>(buffer->DataPtrVoid());

    uint16_t sn = actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    ptcommMsg->sn = sn;
    ptcommMsg->repairId = repairId_.load();
    ptcommMsg->rankNum = static_cast<uint32_t>(isolateRanks.size());
    int32_t idx = 0;
    for (auto &rk : isolateRanks) {
        ptcommMsg->ranks[idx++] = rk;
    }
    auto allRanksVec = GetAllLinkRanks();
    std::vector<ActionInfo> info {{TTP_MSG_OP_PT_COMM, buffer, allRanksVec}};
    TResult ret = engine_->Process(ACTION_OP_PT_COMM, info, true, sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " hccl communication failed, return " << ret);
        return TTP_ERROR;
    }

    TTP_LOG_INFO("controller:" << rank_ << " execute hccl process group destroy & rebuid success");
    return TTP_OK;
}

TResult Controller::UpGradeCommGroupRepair(const std::set<int32_t> &isolateRanks)
{
    uint32_t zitParamLen = sizeof(int32_t) + zitParam_.strategyParm.length() + 1; // len + char[]
    uint32_t baseLen = sizeof(RebuildGroupMsg) + isolateRanks.size() * sizeof(int32_t);
    uint32_t msgLen =  baseLen + zitParamLen;
    TTP_ASSERT_RETURN(IsMsgLenValid(msgLen), TTP_ERROR);

    AccDataBufferPtr buffer = AccDataBuffer::Create(msgLen);
    if (buffer == nullptr) {
        return TTP_ERROR;
    }
    RebuildGroupMsg *ptcommMsg = static_cast<RebuildGroupMsg *>(buffer->DataPtrVoid());

    uint16_t sn = actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    ptcommMsg->sn = sn;
    ptcommMsg->repairId = repairId_.load();
    ptcommMsg->rankNum = static_cast<uint32_t>(isolateRanks.size());
    int32_t idx = 0;
    for (auto &rk : isolateRanks) {
        ptcommMsg->ranks[idx++] = rk;
    }
    ptcommMsg->ranks[idx++] = static_cast<int32_t>(zitParam_.strategyParm.length() + 1);
    int32_t ret = strncpy_s(reinterpret_cast<char *>(&ptcommMsg->ranks[idx]), zitParam_.strategyParm.length() + 1,
        zitParam_.strategyParm.c_str(), zitParam_.strategyParm.length() + 1);
    TTP_ASSERT_RETURN(ret == TTP_OK, TTP_ERROR);
    auto allRanksVec = GetAllLinkRanks();
    std::vector<ActionInfo> info {{TTP_MSG_OP_UPGRADE_REBUILD, buffer, allRanksVec}};
    ret = engine_->Process(ACTION_OP_UPGRADE_REBUILD, info, true, sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " hccl communication failed, return " << ret);
        return TTP_ERROR;
    }

    TTP_LOG_INFO("controller:" << rank_ << " execute hccl process group destroy & rebuid success");
    return TTP_OK;
}

TResult Controller::NotifyRankRollback(const std::vector<int32_t> &targetRanks, RepairType type)
{
    uint32_t zitParamLen  = 0;
    int16_t msgType = TTP_MSG_OP_ROLLBACK;
    ActionOp opcode = ACTION_OP_ROLLBACK;
    if (repairType_ == ControllerRepairType::CRT_UPGRADE) {
        zitParamLen = zitParam_.strategyParm.length() + 1;
        msgType = TTP_MSG_OP_UPGRADE_ROLLBACK;
        opcode = ACTION_OP_UPGRADE_ROLLBACK;
    }
    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(RollbackMsg) + zitParamLen);
    if (buffer == nullptr) {
        return TTP_ERROR;
    }

    RollbackMsg *rbMsg = static_cast<RollbackMsg *>(buffer->DataPtrVoid());
    rbMsg->step = repairStep_;
    rbMsg->repairId = repairId_;
    rbMsg->type = type;
    rbMsg->sn = actionSn_.fetch_add(1);
    rbMsg->dataLen = zitParamLen;
    if (zitParamLen > 0) {
        int32_t result = strncpy_s(&rbMsg->data[0], zitParam_.strategyParm.length() + 1,
            zitParam_.strategyParm.c_str(), zitParam_.strategyParm.length() + 1);
        TTP_ASSERT_RETURN(result == TTP_OK, TTP_ERROR);
    }
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    std::vector<ActionInfo> info {{msgType, buffer, targetRanks}};
    TResult ret = engine_->Process(opcode, info, true, rbMsg->sn);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " do rollback failed, ret:" << ret);
        return TTP_ERROR;
    }
    TTP_LOG_INFO("Controller:" << rank_ << " rollback success");
    return TTP_OK;
}

TResult Controller::UpgradePreCheck(const std::set<int32_t> &isolateRanks)
{
    // check mindx repairEvent，mindx已经下发命令，wait保证notify、wait正常配套完成
    TResult ret = mindXEngine_->Wait(waitMindxTimes_);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("wait mindx notify upgrade failed, ret:" << ret);
        return TTP_STOP_SERVICE;
    }
    if (repairEvent_ != MindXEvent::MINDX_EVENT_UPGRADE || repairType_ != ControllerRepairType::CRT_UPGRADE) {
        TTP_LOG_ERROR("controller:" << rank_ << " upgrade condition not match, repairEvent:" <<
            static_cast<uint32_t>(repairEvent_) << ", repairType:" << repairType_);
        return TTP_ERROR;
    }
    bool isAllPreTrainFinish = true;
    AutoLock statusMapLock(statusMapLock_, TYPE_READ);
    for (auto isoRank : isolateRanks) {
        auto it = statusMap_.find(isoRank);
        if (it == statusMap_.end()) {
            TTP_LOG_ERROR("controller: rank:" << isoRank << " not exist in statusMap_.");
            return TTP_ERROR;
        }
        if (it->second.run_status != TTP_STATUS_PREREPAIR_FINISH) {
            isAllPreTrainFinish = false;
            TTP_LOG_ERROR("controller: find rank:" << isoRank << " is not ready.");
            break;
        }
    }

    if (!isAllPreTrainFinish) {
        TTP_LOG_ERROR("controller: find error rank during upgrade run status, go to dump status.");
    }
    return isAllPreTrainFinish ? TTP_OK : TTP_ERROR;
}

TResult Controller::DoNoDataAction(const std::vector<int32_t> &ranks, MsgOpCode mOp, ActionOp aOp, bool reply,
                                   bool sendRepairType)
{
    uint32_t dataLen = sendRepairType ? sizeof(uint16_t) + sizeof(uint32_t) : sizeof(uint16_t);

    AccDataBufferPtr buffer = AccDataBuffer::Create(dataLen);
    if (buffer == nullptr) {
        return TTP_ERROR;
    }

    uint16_t *snPtr = static_cast<uint16_t *>(buffer->DataPtrVoid());
    *snPtr = actionSn_.fetch_add(1);
    if (IsActionSnOK() == TTP_ERROR) {
        return TTP_ERROR;
    }

    if (sendRepairType) {
        uint32_t *repairType = reinterpret_cast<uint32_t *>(buffer->DataPtr() + sizeof(uint16_t));
        *repairType = repairType_;
    }

    std::vector<ActionInfo> info {{mOp, buffer, ranks}};
    TResult ret = engine_->Process(aOp, info, reply, *snPtr);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("controller:" << rank_ << " do action failed. msg_op:" << static_cast<int32_t>(mOp));
        return TTP_ERROR;
    }
    return ret;
}

bool Controller::CheckTcpStoreServerAvailable()
{
#ifdef UT_ENABLED  // ut测试跳过tcp store server检测步骤
    TTP_LOG_INFO("skip tcp store server check...");
    return true;
#endif
    // 获取TCPStore的server
    std::set<std::string> tcpStoreServerIP;
    uint16_t tcpStoreServerPort;
    TResult ret = GetTcpStoreUrl(tcpStoreServerIP, tcpStoreServerPort);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("Controller: " << rank_ << " get tcp store url failed!");
        return false;
    }

    bool isTcpStoreServerOK = false;
    for (auto &ip : tcpStoreServerIP) {
        if (CheckIpPortAccessible(ip, tcpStoreServerPort)) {
            isTcpStoreServerOK = true;  // Any one server is OK, no need launch new tch store
            break;
        }
    }
    return isTcpStoreServerOK;
}

bool Controller::CheckIpPortAccessible(const std::string &ip, uint16_t port)
{
    // socket创建，失败的话就返回false
    int sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock < 0) {
        TTP_LOG_ERROR("Controller:" << rank_ << " error creating socket");
        return false;
    }

    // 建立sockaddr_in结构.
    struct sockaddr_in serverAddr{};
    serverAddr.sin_family = AF_INET;
    serverAddr.sin_port = htons(port);  // port已校验范围[1024, uint16_max]
    serverAddr.sin_addr.s_addr = inet_addr(ip.data());

    // 尝试建立socket连接
    if (connect(sock, reinterpret_cast<struct sockaddr *>(&serverAddr), sizeof(serverAddr)) < 0) {
        TTP_LOG_INFO("Controller:" << rank_ << " connection failed, ip:" << ip << ':' << port);
        close(sock);
        return false;
    }

    // 使用getsockopt函数检查连接状态
    int optval;
    socklen_t optlen = sizeof(optval);
    bool retVal;
    if (getsockopt(sock, SOL_SOCKET, SO_ERROR, &optval, &optlen) == 0) {
        if (optval == 0) {
            TTP_LOG_INFO("Controller:" << rank_ << " connected to the server successfully!");
            retVal = true;
        } else {
            TTP_LOG_ERROR("Controller:" << rank_ << " socket error " << strerror(optval));
            retVal = false;
        }
    } else {
        TTP_LOG_ERROR("Controller:" << rank_ << " getsockopt failed");
        retVal = false;
    }

    // 关闭socket
    close(sock);
    return retVal;
}

TResult Controller::GetTcpStoreUrl(std::set<std::string> &ipList, uint16_t &port)
{
    ipList.clear();
    // Get tcp store server port
    uint32_t portTmp = GetEnvValue2Uint32("MASTER_PORT", PortInfo.minVal, PortInfo.maxVal, PortInfo.defaultVal);
    TTP_LOG_INFO("[env] MASTER_PORT:" << portTmp);

    port = static_cast<uint16_t>(portTmp);

    // Get master ip
    const char *masterIp = std::getenv("MASTER_ADDR");
    if (masterIp != nullptr) {
        std::string ipTemp = "";
        TResult ret = TransforHostNameToIp(masterIp, ipTemp);
        if (ret == TTP_OK) {
            ipList.insert(ipTemp);
        }
    } else {
        TTP_LOG_WARN("Environment variable: MASTER_ADDR not set!");
    }

    // Get controller ip
    AutoLock ipLock(ipMapLock_, TYPE_READ);
    // Get current controller rank ip
    auto it = rankIpMap_.find(rank_);
    if (it == rankIpMap_.end()) {
        if (ipList.empty()) {
            TTP_LOG_ERROR("Controller: " << rank_ << ", not found in rankIpMap_.");
            return TTP_ERROR;
        }
        return TTP_OK;
    }
    ipList.insert(it->second);

    return TTP_OK;
}

TResult Controller::TransforHostNameToIp(const char *hostName, std::string &ip)
{
    ip = "";
    if (hostName == nullptr) {
        TTP_LOG_WARN("Input null hostname!");
        return TTP_ERROR;
    }

    struct hostent *masterHostent = nullptr;
    if (inet_addr(hostName) == INADDR_NONE) {
        if ((masterHostent = gethostbyname(hostName)) == nullptr) {
            TTP_LOG_WARN("Input invalid hostname: " << std::string(hostName) << ", convert to ip failed.");
            return TTP_ERROR;
        }
        const char *ipaddr = inet_ntoa(*reinterpret_cast<struct in_addr *>(masterHostent->h_addr));
        if (ipaddr != nullptr) {
            ip = ipaddr;
            return TTP_OK;
        }
        TTP_LOG_WARN("hostname:" << std::string(hostName) << " can not convert to ip");
        return TTP_ERROR;
    }
    TTP_ASSERT_RETURN(IsValidIpV4(hostName), TTP_ERROR);  // check input length in IsValidIpV4
    ip = hostName;
    return TTP_OK;
}

TResult Controller::HandleTcpStoreError(bool &tcpStoreOK)
{
    if (rank_ == -1 || CheckTcpStoreServerAvailable()) {
        tcpStoreOK = true;
        return TTP_OK;
    }
    tcpStoreOK = false;
    // use controller rank to launch tcp store server
    std::vector<int32_t> ranks { rank_ };
    TResult ret = DoNoDataAction(ranks, TTP_MSG_OP_LAUNCH_STORE_SERVER, ACTION_OP_LAUNCH_STORE_SERVER, true);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("Controller: " << rank_ << " launch tcp store server failed, ret:" << ret);
        return TTP_ERROR;
    }
    return ret;
}

inline bool Controller::IsMsgLenValid(uint32_t len)
{
    return (len <= MAX_MSG_LEN);
}

std::set<int32_t> Controller::GetErrorRanks()
{
    AutoLock errorRankLock(errorRankLock_, TYPE_READ);
    auto errRanks = GetMapKeysToSet(errorRankMsg_);
    return std::move(errRanks);
}

std::set<int32_t> Controller::GetIsolateRanks()
{
    auto errRanks = zitParam_.isolateRanks;
    return std::move(errRanks);
}

}
}