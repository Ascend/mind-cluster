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
#include <csignal>
#include <unordered_set>
#include <unordered_map>
#include "ttp_logger.h"
#include "processor.h"

namespace ock {
namespace ttp {

constexpr uint32_t UCE_NO_REBUILD = 2;
constexpr uint32_t PRELOCK_RETRY_TIMES = 30 * 1000;
constexpr uint32_t DUMP_WAIT_TIME  = 30 * 60;
constexpr uint32_t REPAIR_WAIT_TIME = 5;
constexpr uint32_t REPAIR_RETRY_TIMES = 5;
constexpr uint32_t REGISTER_REPORT_WAIT_TIME = 6;
constexpr uint32_t REGISTER_REPORT_RETRY_TIMES = 5;

constexpr EnvVarValue ConnectRetryTimes = {.minVal = 1, .maxVal = 300, .defaultVal = 10};
constexpr EnvVarValue PortInfo = {.minVal = 1024, .maxVal = 65535, .defaultVal = 6000};

const std::unordered_set<MsgOpCode> allowRetryOp { TTP_MSG_OP_COLLECTION_REPLY };

ProcessorPtr Processor::GetInstance(bool destroy)
{
    static std::mutex gMutex;
    static ProcessorPtr gInstance;

    if (gInstance == nullptr) {
        std::lock_guard<std::mutex> guard(gMutex);
        if (gInstance.Get() == nullptr) {
            // logger must not nullptr
            if (OutLogger::Instance() == nullptr) {
                throw std::bad_alloc();
            }

            gInstance = MakeRef<Processor>();
            if (gInstance == nullptr) {
                TTP_LOG_ERROR("Create Processor failed,out of memory");
                throw std::bad_alloc();
            }
        }
    } else if (destroy) {
        std::lock_guard<std::mutex> guard(gMutex);
        gInstance = nullptr;
    }

    return gInstance;
}

Processor::Processor() : rank_{0}, controllerIdx_{0}, port_{0}, worldSize_{0}, localCopySwitch_{false}
{
    rankList_.clear();
    controllerIps_.clear();
    replyMsgBackup_.sn = 0;
    replyMsgBackup_.rank = 0;
    replyMsgBackup_.status = TTP_BUTT;
}

bool Processor::ProcessorIsRunning()
{
    uint32_t st = processorStatus_.load();
    return st == PS_NORMAL || st == PS_PAUSE || st == PS_DUMP;
}

void SignalHandler(int32_t signum)
{
    TTP_LOG_WARN("==== this processor" << " receive sig " << strsignal(signum));
    return;
}

void Processor::RequestHandleRegister()
{
    auto ckptMethod = [this](uint8_t *data, uint32_t len) { return HandleDumpCkpt(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_CKPT_SEND, ckptMethod);
    auto renameMethod = [this](uint8_t *data, uint32_t len) { return HandleRename(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_RENAME, renameMethod);
    auto exitMethod = [this](uint8_t *data, uint32_t len) { return HandleExit(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_EXIT, exitMethod);
    auto destroyMethod = [this](uint8_t *data, uint32_t len) { return HandleDestroyNofity(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_DESTROY_NOTIFY, destroyMethod);
    auto notifyMethod = [this](uint8_t *data, uint32_t len) { return HandleBroadcast(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_CTRL_NOTIFY, notifyMethod);
    auto collectMethod = [this](uint8_t *data, uint32_t len) { return HandleCollection(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_COLLECTION, collectMethod);
    auto stopMethod = [this](uint8_t *data, uint32_t len) { return HandleDeviceStop(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_DEVICE_STOP, stopMethod);
    auto cleanMethod = [this](uint8_t *data, uint32_t len) { return HandleDeviceClean(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_DEVICE_CLEAN, cleanMethod);
    auto repairMethod = [this](uint8_t *data, uint32_t len) { return HandleRepair(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_REPAIR, repairMethod);
    auto normalMethod = [this](uint8_t *data, uint32_t len) { return HandleNotifyNormal(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_NOTIFY_NORMAL, normalMethod);
    auto lockMethod = [this](uint8_t *data, uint32_t len) { return HandlePrelock(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_PRELOCK, lockMethod);
    auto pauseMethod = [this](uint8_t *data, uint32_t len) { return HandlePause(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_PAUSE, pauseMethod);
    auto continueMethod = [this](uint8_t *data, uint32_t len) { return HandleContinue(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_CONTINUE, continueMethod);
    auto rollbackMethod = [this](uint8_t *data, uint32_t len) { return HandleRollback(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_ROLLBACK, rollbackMethod);
    auto rebuildMethod = [this](uint8_t *data, uint32_t len) { return HandleDowngradeRebuild(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_DOWNGRADE_REBUILD, rebuildMethod);
    auto ptcommMethod = [this](uint8_t *data, uint32_t len) { return HandlePtComm(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_PT_COMM, ptcommMethod);
    auto upgradeRebuildMethod = [this](uint8_t *data, uint32_t len) { return HandleUpgradeRebuild(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_UPGRADE_REBUILD, upgradeRebuildMethod);
    auto upgradeRepairMethod = [this](uint8_t *data, uint32_t len) { return HandleUpgradeRepair(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_UPGRADE_REPAIR, upgradeRepairMethod);
    auto upgradeRollbackMethod = [this](uint8_t *data, uint32_t len) { return HandleUpgradeRollback(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_UPGRADE_ROLLBACK, upgradeRollbackMethod);
    auto launchServerMethod = [this](uint8_t *data, uint32_t len) { return HandleLaunchTcpStoreServer(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_LAUNCH_STORE_SERVER, launchServerMethod);

    auto replyMethod = [this](uint8_t *data, uint32_t len) { return HandleControllerReply(data, len); };
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_REGISTER, replyMethod);
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_INIT_REPORT, replyMethod);
    newClient_->RegisterNewRequestHandler(TTP_MSG_OP_DP_REPORT, replyMethod);
}

void Processor::ExtraReplyHandleRegister()
{
    auto extraReplyMethod = [this](uint16_t sn, MsgOpCode op) { return ProcessResultAndHBReply(sn, op); };
    extraReplyHandleList_[TTP_MSG_OP_COLLECTION_REPLY] = extraReplyMethod;
    extraReplyHandleList_[TTP_MSG_OP_PRELOCK_REPLY] = extraReplyMethod;
    extraReplyHandleList_[TTP_MSG_OP_NORMAL_REPLY] = extraReplyMethod;
}

TResult Processor::Initialize(int32_t rank, int32_t worldSize, bool enableLocalCopy,
    const AccTlsOption &tlsOption, bool enableUce, bool enableArf, bool enableZit)
{
    std::lock_guard<std::mutex> guard(initOrDestroyMutex_);

    TTPLogger::Init();
    TTP_LOG_DEBUG("Start to init processor, rank:" << rank);

    TTP_ASSERT_RETURN(rank >= 0 && rank < TTP_MAX_WORLD_SIZE, TTP_ERROR);
    TTP_ASSERT_RETURN(worldSize > 0 && worldSize <= TTP_MAX_WORLD_SIZE, TTP_ERROR);
    TTP_ASSERT_RETURN(rank < worldSize, TTP_ERROR);
    mindSpore_ = GetMsEnv("MINDIO_FOR_MINDSPORE");
    if (mindSpore_ && enableLocalCopy) {
        TTP_LOG_ERROR("MS does not support local copy");
        return TTP_ERROR;
    }

#ifndef UT_ENABLED  // ut测试跳过signal register操作
    struct sigaction act;
    act.sa_handler = SignalHandler;
    sigemptyset(&act.sa_mask);
    act.sa_flags = SA_RESTART;
    sigaction(SIGINT, &act, nullptr);
    sigaction(SIGTERM, &act, nullptr);
#endif

    uint32_t expectedStatus = PS_BEGIN;
    if (!processorStatus_.compare_exchange_strong(expectedStatus, PS_INITING)) {
        TTP_LOG_ERROR("rank:" << rank_ << " initialize failed, has inited_");
        return TTP_ERROR;
    }

    rank_ = rank;
    worldSize_ = worldSize;
    localCopySwitch_ = enableLocalCopy;
    arfSwitch_ = enableArf;
    uceSwitch_ = enableUce;
    zitSwitch_ = enableZit;
    newClient_ = AccTcpClient::Create();
    TTP_ASSERT_RETURN(newClient_ != nullptr, TTP_ERROR);
    if (InitTlsOption(tlsOption) != TTP_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " initialize tls options failed");
        return TTP_ERROR;
    }
    if (dumpCond_.Initialize() != TTP_OK || replySem_.Initialize() != TTP_OK ||
        repairWaitCond_.Initialize() != TTP_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " initialize sem failed");
        return TTP_ERROR;
    }

    RequestHandleRegister();
    ExtraReplyHandleRegister();
    TTP_LOG_INFO("processor:" << rank_ << " initialize success, uce:" << enableUce << ", arf:" << enableArf <<
        ", zit:" << enableZit);

    sem_init(&waitSem_, 0, 0);
    processorStatus_.store(PS_INITED);
    return TTP_OK;
}

TResult Processor::InitTlsOption(const AccTlsOption &tlsOption)
{
    if (!tlsOption.enableTls) {
        return TTP_OK;
    }

    auto ret = newClient_->LoadDynamicLib(tlsOption.packagePath);
    if (ret != ACC_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " load ssl dynamic lib failed");
        return TTP_ERROR;
    }
    newClient_->SetSslOption(tlsOption);
    tlsOption_ = tlsOption;
    return TTP_OK;
}

// must be used in statusMutex_
// backup_step <= step, so pair.first <= pair.second
std::pair<int64_t, int64_t> Processor::GetNowStep()
{
    if (localCopySwitch_) {
        if (trainStatus_.data_status == Updated) {
            return std::make_pair(trainStatus_.backup_step, trainStatus_.step);
        } else if (trainStatus_.data_status == Copying) {
            return std::make_pair(trainStatus_.step, trainStatus_.step);
        } else if (trainStatus_.data_status == Updating) {
            return std::make_pair(trainStatus_.backup_step, trainStatus_.backup_step);
        }
    } else {
        return std::make_pair(trainStatus_.step, trainStatus_.step);
    }
}

void Processor::WaitForLimitStepRelease(int64_t nowStep)
{
    TTP_ASSERT_RET_VOID(nowStep >= -1 && nowStep < INT64_MAX);
    // when nowStep > limitStep_, keep waiting and print log
    while (true) {
        if (nowStep < limitStep_.load()) {
            break;
        }
        TTP_LOG_LIMIT_WARN(LOG_PRINT_INTERVAL, "processor:" << rank_ <<
            " locked. now_step: " << trainStatus_.step << " limit_step: " << limitStep_);
        usleep(TTP_WAIT_TIME_1MS);
    }
}

TResult Processor::BeginCopying()
{
    if (processorStatus_.load() == PS_DUMP) {
        TTP_LOG_ERROR("processor:" << rank_ << " start copying os failed, rank_ dump now");
        return TTP_ERROR;
    }

    std::unique_lock<std::mutex> lck(statusMutex_);
    trainStatus_.data_status = Copying;
    TTP_LOG_DEBUG("processor:" << rank_ << " start copying os success");

    return TTP_OK;
}

TResult Processor::BeginUpdating(int64_t backupStep)
{
    TTP_ASSERT_RETURN(backupStep >= -1 && backupStep < INT64_MAX, TTP_ERROR);
    TTP_LOG_DEBUG("processor:" << rank_ << " begin to update, step: " << backupStep);

    int64_t nowStep = 0;
    {
        std::unique_lock<std::mutex> lck(statusMutex_);
        auto range = GetNowStep();
        nowStep = range.second;
    }

    WaitForLimitStepRelease(nowStep);
    if (isPrelockOk_) {
        isPrelockOk_.store(false);
        if (!mindSpore_) {
            TTP_LOG_INFO("rank:" << rank_ << " throw force stop exception during Begin Updating.");
            throw std::runtime_error("FORCE STOP. By mindio-ttp BeginUpdating");
        } else {
            return TTP_OK;
        }
    }

    if (processorStatus_.load() == PS_DUMP) {
        TTP_LOG_ERROR("processor:" << rank_ << " start updating os failed, rank_ dump now");
        return TTP_ERROR;
    }

    std::unique_lock<std::mutex> lck(statusMutex_);
    trainStatus_.data_status = Updating;
    trainStatus_.backup_step = backupStep;
    TTP_LOG_DEBUG("processor:" << rank_ << " start updating os success");

    return TTP_OK;
}

TResult Processor::FinishedUpdate(int64_t step)
{
    TTP_ASSERT_RETURN(step > 0 && step < INT64_MAX, TTP_ERROR);
    TTP_LOG_DEBUG("processor:" << rank_ << " end to update, step: " << step);
    int64_t nowStep;
    {
        std::unique_lock<std::mutex> lck(statusMutex_);
        if (trainStatus_.data_status == Updating) {
            trainStatus_.data_status = Updated;
            trainStatus_.step = step;
            TTP_LOG_DEBUG("processor:" << rank_ << " updated os success, step: " << step);
        }
        auto range = GetNowStep();
        nowStep = range.second;
    }

    WaitForLimitStepRelease(nowStep);
    return TTP_OK;
}

TResult Processor::ResetLimitStep()
{
    isPrelockOk_.store(true);
    limitStep_.store(INT64_MAX); // clear prelock
    TTP_LOG_DEBUG("rank:" << rank_ << " call reset limit step");
    return TTP_OK;
}

TResult Processor::ReportReplicaInfo(std::vector<std::vector<int32_t>> &groups,
    std::vector<int32_t> replicaCnt, std::vector<int32_t> replicaShift)
{
    uint32_t groupNum = groups.size();
    if (replicaCnt.size() != groupNum ||
        replicaShift.size() != groupNum ||
        groupNum == 0 ||
        groupNum > TTP_MAX_OPTIM_NUM) {
        TTP_LOG_ERROR("size not match, groupNum:" << groupNum << " cntSize:"
                                                  << replicaCnt.size() << " shiftSize:" << replicaShift.size());
        return TTP_ERROR;
    }

    for (int32_t i = 0; i < static_cast<int32_t>(groupNum); i++) {
        if (mindSpore_) {
            if (replicaCnt[i] < 1 || replicaCnt[i] != static_cast<int32_t>(groups[i].size())) {  // rank list size > 1
                TTP_LOG_ERROR("MS ranks error, cnt:" << replicaCnt[i] << " ranks:" << IntVec2String(groups[i]));
                return TTP_ERROR;
            }
        } else if (replicaCnt[i] < 1 || replicaCnt[i] > static_cast<int32_t>(groups[i].size())) {
            TTP_LOG_ERROR("param error, groupSize:" << groups[i].size() << " replicaCnt:" << replicaCnt[i]);
            return TTP_ERROR;
        }
    }

    groupTypeNum_ = static_cast<int32_t>(groupNum);
    rankList_ = groups;
    replicaCnt_ = replicaCnt;
    replicaShift_ = replicaShift;
    TResult ret = ReportInfo2Controller(replicaCnt, replicaShift);
    if (ret != TTP_OK) {
        return ret;
    }

    return TTP_OK;
}

TResult Processor::ReportDpInfo(std::vector<int32_t> &dpRankList)
{
    if (!zitSwitch_) {    // only need to report dp group when zit is on
        return TTP_OK;
    }

    TTP_ASSERT_RETURN(dpRankList.size() <= worldSize_, TTP_ERROR);
    bool containCurRank = false;
    for (auto rk : dpRankList) {
        if (rk == rank_) {
            containCurRank = true;
        }
        TTP_ASSERT_RETURN(rk >= 0 && rk < worldSize_, TTP_ERROR);
    }
    TTP_ASSERT_RETURN(containCurRank, TTP_ERROR);
    dpList_ = dpRankList;

    // if ep_size < cp_size, zit needs report origin dp group.
    // dp_cp_size = dp_size * cp_size
    // dp_ep_size = dp_cp_size / ep_size
    for (const auto &group : rankList_) {
        if (group.size() <= dpRankList.size()) {
            return TTP_OK;
        }
    }
    auto ret = ReportDp2Controller();
    return ret;
}

TResult Processor::Start(std::string &masterIp, int32_t port, std::string localIp)
{
    std::lock_guard<std::mutex> guard(initOrDestroyMutex_);

    TTP_ASSERT_RETURN(IsValidIpV4(masterIp), TTP_ERROR);
    if (localIp != "" && !IsValidIpV4(localIp)) {
        TTP_LOG_ERROR("localIp invalid!");
        return TTP_ERROR;
    }
    if (port < PortInfo.minVal || port > PortInfo.maxVal) {
        TTP_LOG_ERROR("rank:" << rank_ << " initialize processor failed, invaild port: " << port);
        return TTP_ERROR;
    }

    uint32_t expectedStatus = PS_INITED;
    if (!processorStatus_.compare_exchange_strong(expectedStatus, PS_START)) {
        TTP_LOG_ERROR("rank:" << rank_ << " start failed, now status: " << processorStatus_.load());
        return TTP_ERROR;
    }

    controllerIps_.push_back(masterIp);
    controllerIdx_ = 0;
    port_ = port;
    newClient_->SetServerIpAndPort(controllerIps_[controllerIdx_], port);
    newClient_->SetLocalIp(localIp);

    // 1. connect to controller
    TResult result = Connect2Controller();
    if (result == TTP_ERROR) {
        processorStatus_.store(PS_INITED);
        return TTP_ERROR;
    }

    TTP_LOG_INFO("rank:" << rank_ << " start msg polling thread......");
    newClient_->StartPolling();

    // 2. register to controller
    result = Register2Controller();
    if (result != TTP_OK) {
        return result;
    }

    // 3. start heartbeat thread
    isStarted_.store(false);
    isStopped_.store(false);
    TTP_LOG_INFO("rank:" << rank_ << " start heartbeat thread......");
    std::thread tmpThread(&Processor::HeartbeatThread, this);
    thread_ = std::move(tmpThread);
    std::string tName = "ttp_processor";
    if (pthread_setname_np(thread_.native_handle(), tName.c_str()) != 0) {
        TTP_LOG_WARN("rank:" << rank_ << " set ttp processor thread name failed");
    }

    while (!isStarted_.load()) {
        usleep(100L);
    }

    processorStatus_.store(PS_NORMAL);
    return TTP_OK;
}

void Processor::Destroy(bool inInner)
{
    std::lock_guard<std::mutex> guard(initOrDestroyMutex_);

    auto st = processorStatus_.load();
    bool invalidStatus = (st == PS_BEGIN) || (st == PS_END);
    if (invalidStatus) {
        TTP_LOG_DEBUG("rank:" << rank_ << " status:" << st << " ...");
        return;
    }

    if (!inInner && !readyToExit_.load()) {
        ReportBeforeDestroy();
    }

    isStopped_.store(true);
    if (newClient_ != nullptr) {
        bool needWait = !inInner; // 由Controller发起的destroy操作不等待tcpclient退出
        newClient_->Destroy(needWait);
    }

    if (thread_.joinable()) {
        thread_.join();
    }

    std::fill(std::begin(eventHandleList_), std::end(eventHandleList_), nullptr);

    if (startBackup_ != -1 && startBackup_ != controllerIdx_) {
        Controller::GetInstance()->Destroy();
    }

    processorStatus_.store(PS_END);
    TTP_LOG_INFO("rank:" << rank_ << " processor exit done.");
}

void Processor::ReportBeforeDestroy()
{
    readyToExit_.store(true);

    HeartBeatMsg hbMsg;
    {
        std::unique_lock<std::mutex> lck(statusMutex_);
        trainStatus_.run_status = TTP_STATUS_EXIT;
        hbMsg.rank = rank_;
        hbMsg.status = trainStatus_;
    }
    auto result = newClient_->Send(TTP_MSG_OP_HEARTBEAT_SEND,
        reinterpret_cast<uint8_t *>(&hbMsg), sizeof(HeartBeatMsg));
    if (result != ACC_OK) {
        TTP_LOG_WARN("rank:" << rank_ << " send exit status msg to controller: " << IpPort() << " failed");
    }
    TTP_LOG_DEBUG("rank:" << rank_ << " report exit status: " << int32_t(trainStatus_.npu_status) <<
                         ", run_status: " << int32_t(trainStatus_.run_status));
}

TResult Processor::HeartbeatSend()
{
    HeartBeatMsg hbMsg;
    hbMsg.rank = rank_;

    int32_t times = 0;
    while (times < TTP_HEARTBEAT_CONNECT_THRESHOLD) {
        {
            std::unique_lock<std::mutex> lck(statusMutex_);
            hbMsg.status = trainStatus_;
        }
        hbMsg.repairId = repairId_;

        auto result = newClient_->Send(TTP_MSG_OP_HEARTBEAT_SEND,
            reinterpret_cast<uint8_t *>(&hbMsg), sizeof(HeartBeatMsg));
        if (result == ACC_OK) {
            TTP_LOG_DEBUG("rank:" << rank_ << " send heartbeat msg to controller: " << IpPort() << " success.");
            return TTP_OK;
        }
        times++;
        sleep(1);
    }

    return TTP_TIMEOUT;
}

void Processor::HeartbeatThread()
{
    isStarted_.store(true);
    TResult ret;

    while (!isStopped_.load() && !readyToExit_.load()) {
        ret = HeartbeatSend();
        if (ret == TTP_TIMEOUT) {
            TTP_LOG_ERROR("rank:" << rank_ << " send heartbeat msg to controller: " << IpPort() << " failed");
            int32_t result = ReStart();
            if (result != TTP_OK) {
                TTP_LOG_ERROR("rank:" << rank_ << " send heartbeat msg to all controller failed, exit...");
                break;
            }
            TTP_LOG_INFO("rank:" << rank_ << " send heartbeat msg to [backup controller]: " << IpPort());
        } else if (ret != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " send heartbeat msg to controller failed, unprocessed");
        }
        sleep(TTP_SLEEP_TIME);
    }
    TTP_LOG_DEBUG("rank:" << rank_ << " heartbeat thread exit...");
}

void Processor::HandleDumpCkpt(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_LOG_INFO("rank:" << rank_ << " receive dump message");
    // check sn before changing processorStatus_
    CkptMsg *msg = reinterpret_cast<CkptMsg *>(data);
    TTP_ASSERT_RET_VOID(msg->JudgeVariableValid(len, worldSize_));
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msg->sn, TTP_MSG_OP_CKPT_REPLY) != TTP_ERROR);

    TResult ret = TTP_OK;
    uint32_t nowStatus;
    do {
        nowStatus = processorStatus_.load();
        if (nowStatus < PS_NORMAL) { // PS_NORMAL,PS_PAUSE,PS_DUMP can transfer to PS_DUMP
            TTP_LOG_ERROR("rank:" << rank_ << " handle dump failed, now status:" << nowStatus);
            ret = TTP_ERROR;
            break;
        }
    } while (!processorStatus_.compare_exchange_strong(nowStatus, PS_DUMP));

    if (!envClearFlag_) {
        ret = StopAndCleanBeforeDump();
    } else {
        TTP_LOG_INFO("rank:" << rank_ << " stop&clean has succeeded before! skip!");
        ret = TTP_OK;
    }

    TTPReplyMsg replyMsg;
    if (ret == TTP_OK) {
        ret = DumpCkpt(data);
    }

    if (ret == TTP_OK) {
        if (dumpRet_ == -1) { // execute callback success, wait dump finish and notify
            auto ret = dumpCond_.PthreadTimedwaitSecs(DUMP_WAIT_TIME);
            if (ret == ETIMEDOUT) {
                TTP_LOG_WARN("rank:" << rank_ << " wait to finish dump timeout, dumpStatus:" << dumpRet_);
                dumpRet_ = TTP_ERROR;
            }
        }
        replyMsg.status = dumpRet_;
    } else { // execute callback failed, directly to send to controller
        TTP_LOG_WARN("rank:" << rank_ << " execute dump callback failed:" << dumpRet_);
        replyMsg.status = TTP_ERROR;
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_CKPT_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply dump result success, dump ret:" << replyMsg.status);
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply dump result failed, dump ret:" << replyMsg.status);
    }
}

TResult Processor::StopAndCleanBeforeDump()
{
    bool flag = GetEnvValue2Uint32("TTP_STOP_CLEAN_BEFORE_DUMP", 1, 1, 0);
    if (!flag) {
        return TTP_OK;
    }

    auto ret = EventProcess(PROCESSOR_EVENT_DEVICE_STOP, &rank_, sizeof(rank_));
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " device stop failed before dump, ret: " << ret);
    } else {
        TTP_LOG_INFO("rank:" << rank_ << " device stop success before dump.");
    }

    ret = EventProcess(PROCESSOR_EVENT_DEVICE_CLEAN, &rank_, sizeof(rank_));
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " clean before dump failed, ret: " << ret);
    }
#ifndef UT_ENABLED
    if (!mindSpore_) {
        waitResult_ = TTP_OK;
        sem_post(&waitSem_);
        TTP_LOG_INFO("rank:" << rank_ << " stop and clean before dump end wait clean action.. ret: " << waitResult_);
    }
    RepairBarrierWithFramework();
#endif
    return ret;
}

TResult Processor::DumpCkpt(uint8_t *data)
{
    TTP_ASSERT_RETURN(data != nullptr, TTP_ERROR);
    CkptMsg *msg = reinterpret_cast<CkptMsg *>(data);
    std::vector<int32_t> groupIdx(msg->num);
    std::vector<std::vector<int32_t>> rankVec(msg->num);
    int32_t pIdx = 0;
    int32_t foundCount = 0;
    for (uint32_t idx = 0; idx < msg->num; idx++) {
        groupIdx[idx] = msg->ranks[pIdx++];
        for (int32_t i = 1; i <= msg->ranks[pIdx]; i++) {
            rankVec[idx].push_back(msg->ranks[pIdx + i]);
            foundCount += (msg->ranks[pIdx + i] == rank_);
        }
        pIdx += msg->ranks[pIdx] + 1;
    }

    // local rank_ not in ranklist, rank_ which to save not in ranklist
    bool err = (foundCount != static_cast<int32_t>(msg->num)) || (!CheckOptimStateOK(msg->step));
    if (err) {
        TTP_LOG_ERROR("rank:" << rank_ << " dump param check failed, receive step: " << msg->step
                              << ", current step: " << trainStatus_.step << ", receive dump group num: " << msg->num);
        return TTP_ERROR;
    }

    if (!msg->isTcpStoreOK && !mindSpore_) {
        if (LaunchTcpStoreClient() != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " detect tcp store client error and start failed!");
            return TTP_ERROR;
        }
    }

    SaveCkptContext scc = { msg->step, msg->repairId, groupIdx, rankVec };
    TResult ret = EventProcess(PROCESSOR_EVENT_SAVE_CKPT, &scc, sizeof(SaveCkptContext));

    return ret != 0 ? TTP_ERROR : TTP_OK;
}

bool Processor::CheckOptimStateOK(int64_t step)
{
    std::unique_lock<std::mutex> lck(statusMutex_);
    bool isErr = (step <= 0) ||
        (step != trainStatus_.step && !localCopySwitch_) ||
        (step != trainStatus_.step && localCopySwitch_ && trainStatus_.data_status == Copying) ||
        (step != trainStatus_.backup_step && localCopySwitch_ && trainStatus_.data_status == Updating) ||
        (step != trainStatus_.step && step != trainStatus_.backup_step && localCopySwitch_ &&
        trainStatus_.data_status == Updated);

    if (isErr) {
        TTP_LOG_ERROR("rank:" << rank_ << " repair param check failed, receive step: " << step <<
            ", current step: " << trainStatus_.step);
    }
    return !isErr;
}

void Processor::HandleRename(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(CommonMsg));
    TTP_LOG_INFO("rank:" << rank_ << " receive rename message");
    CommonMsg *msg = reinterpret_cast<CommonMsg *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msg->sn, TTP_MSG_OP_RENAME_REPLY) != TTP_ERROR);

    TTPReplyMsg replyMsg;
    replyMsg.rank = rank_;
    TResult ret = Rename(data);
    replyMsg.status = (ret == TTP_OK) ? TTP_OK : TTP_ERROR;
    replyMsg.sn = actionSn_.load();
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_RENAME_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply rename result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply rename result failed");
    }
}

TResult Processor::Rename(uint8_t *data)
{
    TTP_ASSERT_RETURN(data != nullptr, TTP_ERROR);
    CommonMsg *renameMsg = reinterpret_cast<CommonMsg *>(data);
    if (renameMsg->rank != rank_) {
        TTP_LOG_ERROR("local rank_ not match rank_ received, rename failed");
        return TTP_ERROR;
    }

    TResult ret = EventProcess(PROCESSOR_EVENT_RENAME, nullptr, 0);
    if (ret != 0) {
        TTP_LOG_ERROR("rank:" << rank_ << " rename failed");
        return TTP_ERROR;
    }

    TTP_LOG_INFO("rank:" << rank_ << " rename success");
    return TTP_OK;
}

void Processor::HandleDeviceStop(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(uint16_t) + sizeof(uint32_t));
    TTP_LOG_INFO("rank:" << rank_ << " receive device stop message");
    uint16_t *receiveSn = reinterpret_cast<uint16_t *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(*receiveSn, TTP_MSG_OP_STOP_REPLY) != TTP_ERROR);

    uint32_t *repairType = reinterpret_cast<uint32_t *>(data + sizeof(uint16_t));
    repairType_ = *repairType;

    TTPReplyMsg replyMsg;
    if (processorStatus_.load() != PS_PAUSE) {
        TTP_LOG_ERROR("rank:" << rank_ << " device stop failed, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else {
        replyMsg.status = EventProcess(PROCESSOR_EVENT_DEVICE_STOP, &rank_, sizeof(rank_));
        if (replyMsg.status != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " device stop failed, ret: " << replyMsg.status);
        }
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_STOP_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply device stop result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply device stop result failed");
    }
}

void Processor::HandleDeviceClean(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(uint16_t) + sizeof(uint32_t));
    TTP_LOG_INFO("rank:" << rank_ << " receive device clean message");
    uint16_t *receiveSn = reinterpret_cast<uint16_t *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(*receiveSn, TTP_MSG_OP_CLEAN_REPLY) != TTP_ERROR);

    uint32_t *repairType = reinterpret_cast<uint32_t *>(data + sizeof(uint16_t));
    repairType_ = *repairType;

    TTPReplyMsg replyMsg;
    if (processorStatus_.load() != PS_PAUSE) {
        TTP_LOG_ERROR("rank:" << rank_ << " device clean failed, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else {
        replyMsg.status = EventProcess(PROCESSOR_EVENT_DEVICE_CLEAN, &rank_, sizeof(rank_));
        if (replyMsg.status == UCE_NO_REBUILD) { // uce low_level
            TTP_LOG_INFO("rank:" << rank_ << " receive low level uce type..");
            {
                std::unique_lock<std::mutex> lck(statusMutex_);
                trainStatus_.npu_status = TTP_STATUS_UCE_LOW;
            }
            replyMsg.status = HeartbeatSend();
        }
        if (replyMsg.status != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " device clean failed, ret: " << replyMsg.status);
        } else {
            envClearFlag_ = true;
        }
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_CLEAN_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply device clean result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply device clean result failed");
    }

#ifndef UT_ENABLED
    if (!mindSpore_) {
        waitResult_ = TTP_OK;
        sem_post(&waitSem_);
        TTP_LOG_INFO("rank:" << rank_ << " end wait clean action.. waitRet: " << waitResult_);
    }
#endif
}

void Processor::HandleRepair(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_LOG_INFO("rank:" << rank_ << " receive repair message");

    RepairMsg *msg = reinterpret_cast<RepairMsg *>(data);
    TTP_ASSERT_RET_VOID(msg->JudgeVariableValid(len, worldSize_));
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msg->sn, TTP_MSG_OP_REPAIR_REPLY) != TTP_ERROR);

    // wait for npu cache clear finish
    TTP_ASSERT_RET_VOID(RepairBarrierWithFramework() == TTP_OK);

    TTPReplyMsg replyMsg {};
    RepairContext rc;

    RepairMsgUnit *uptr = reinterpret_cast<RepairMsgUnit *>(&msg->arr[msg->rankNum]);
    rc.type = uptr[0].type;
    for (uint32_t i = 0; i < static_cast<uint32_t>(msg->repairNum); i++) {
        rc.srcRank.push_back(uptr[i].srcRank);
        rc.dstRank.push_back(uptr[i].dstRank);
        rc.replicaIdx.push_back(uptr[i].replicaIdx);
        rc.groupIdx.push_back(uptr[i].groupType);
        if (rc.type != uptr[i].type) {
            TTP_LOG_ERROR("rank:" << rank_ << " recv repair msg error, type:"
                << static_cast<int32_t>(rc.type) << " " << static_cast<int32_t>(uptr[i].type));
            replyMsg.status = TTP_ERROR;
        }
    }
    // check step in repair msg whether exist in processor
    TTP_ASSERT_RET_VOID(rc.type != RepairType::RT_SEND || CheckOptimStateOK(msg->step));

    repairId_ = msg->repairId;
    repairType_ = msg->repairType;
    rc.repairId = repairId_;
    rc.step = msg->step;
    for (uint32_t i = 0; i < msg->rankNum; i++) {
        rc.ranks.push_back(msg->arr[i]);
    }

    if (processorStatus_.load() != PS_PAUSE) {
        TTP_LOG_ERROR("rank:" << rank_ << " repair falied, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else if (replyMsg.status == TTP_OK) {
        replyMsg.status = EventProcess(PROCESSOR_EVENT_REPAIR, &rc, sizeof(RepairContext));
        if (replyMsg.status != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " repair failed, ret: " << replyMsg.status);
        }
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_REPAIR_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply repair result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply repair result failed");
    }
}

void Processor::HandleUpgradeRepair(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_LOG_INFO("rank:" << rank_ << " receive repair message");

    RepairMsg *msg = reinterpret_cast<RepairMsg *>(data);
    TTP_ASSERT_RET_VOID(msg->JudgeVariableValid(len, worldSize_));
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msg->sn, TTP_MSG_OP_UPGRADE_REPAIR_REPLY) != TTP_ERROR);

    // wait for npu cache clear finish
    TTP_ASSERT_RET_VOID(RepairBarrierWithFramework() == TTP_OK);

    TTPReplyMsg replyMsg {};
    RepairContext rc;
    int32_t ret = DeSerializedUpgradeRepairMsg(msg, replyMsg, rc);
    TTP_ASSERT_RET_VOID(ret == TTP_OK);
    if (processorStatus_.load() != PS_PAUSE) {
        TTP_LOG_ERROR("rank:" << rank_ << " repair falied, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else if (replyMsg.status == TTP_OK) {
        replyMsg.status = EventProcess(PROCESSOR_EVENT_UPGRADE_REPAIR, &rc, sizeof(RepairContext));
        if (replyMsg.status != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " repair failed, ret: " << replyMsg.status);
        }
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_UPGRADE_REPAIR_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply repair result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply repair result failed");
    }
}

TResult Processor::DeSerializedUpgradeRepairMsg(RepairMsg *msg, TTPReplyMsg &replyMsg, RepairContext &rc)
{
    RepairMsgUnit *uptr = reinterpret_cast<RepairMsgUnit *>(&msg->arr[msg->rankNum]);
    rc.type = uptr[0].type;
    for (uint32_t i = 0; i < msg->repairNum; i++) {
        rc.srcRank.push_back(uptr[i].srcRank);
        rc.dstRank.push_back(uptr[i].dstRank);
        rc.replicaIdx.push_back(uptr[i].replicaIdx);
        rc.groupIdx.push_back(uptr[i].groupType);
        if (rc.type != uptr[i].type) {
            TTP_LOG_ERROR("rank:" << rank_ << " recv repair msg error, type:"
                << static_cast<int32_t>(rc.type) << " " << static_cast<int32_t>(uptr[i].type));
            replyMsg.status = TTP_ERROR;
        }
    }
    int32_t *param = reinterpret_cast<int32_t*>(reinterpret_cast<char*>(uptr) + sizeof(RepairMsgUnit) * msg->repairNum);
    int32_t zitParamLen = *param;
    rc.zitParam = std::string(reinterpret_cast<char*>(param + 1));
    TTP_ASSERT_RETURN(zitParamLen == rc.zitParam.length() + 1, TTP_ERROR);
    // check step in repair msg whether exist in processor
    TTP_ASSERT_RETURN(rc.type != RepairType::RT_SEND || CheckOptimStateOK(msg->step), TTP_ERROR);

    repairId_ = msg->repairId;
    repairType_ = msg->repairType;
    rc.repairId = repairId_;
    rc.step = msg->step;
    for (uint32_t i = 0; i < msg->rankNum; i++) {
        rc.ranks.push_back(msg->arr[i]);
    }
    return TTP_OK;
}

TResult Processor::RepairBarrierWithFramework()
{
    if (mindSpore_) {  // mindspore无需处理repair等待
        return TTP_OK;
    }
    uint32_t repairWaitTime = 0;
    while (repairWaitTime < REPAIR_RETRY_TIMES) {
        int32_t waitRet = 0;
#ifndef UT_ENABLED   // ut测试跳过
        waitRet = repairWaitCond_.PthreadTimedwaitSecs(REPAIR_WAIT_TIME);
#endif
        if (waitRet == 0) {
            break;
        }
        ++repairWaitTime;
    }
    if (repairWaitTime >= REPAIR_RETRY_TIMES) {
        TTP_LOG_ERROR("rank:" << rank_ << ", framework enter repair-wait state over-time, repair failed.");
        return TTP_ERROR;
    }
    return TTP_OK;
}

void Processor::HandleRollback(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(RollbackMsg));
    TTP_LOG_INFO("rank:" << rank_ << " receive rollback message");
    RollbackMsg *msg = reinterpret_cast<RollbackMsg *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msg->sn, TTP_MSG_OP_ROLLBACK_REPLY) != TTP_ERROR);
    TTPReplyMsg replyMsg;

    TTP_ASSERT_RET_VOID(msg->type >= RepairType::RT_SEND && msg->type < RepairType::RT_BUTT);
    TTP_ASSERT_RET_VOID(msg->step > 0 && msg->step < INT64_MAX);

    if (processorStatus_.load() != PS_PAUSE) {
        TTP_LOG_ERROR("rank:" << rank_ << " rollback failed, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else {
        RepairContext rc;
        rc.srcRank.push_back(rank_);
        rc.dstRank.push_back(rank_);
        rc.type = msg->type;
        rc.step = msg->step;
        lockStep_ = msg->step;

        repairId_ = msg->repairId;
        rc.repairId = repairId_;

        replyMsg.status = EventProcess(PROCESSOR_EVENT_ROLLBACK, &rc, sizeof(RepairContext));
        if (replyMsg.status != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " rollback failed, ret: " << replyMsg.status);
        }
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_ROLLBACK_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply rollback result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply rollback result failed");
    }
}

void Processor::HandleUpgradeRollback(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len >= sizeof(RollbackMsg));
    TTP_LOG_INFO("rank:" << rank_ << " receive rollback message");
    RollbackMsg *msg = reinterpret_cast<RollbackMsg *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msg->sn, TTP_MSG_OP_UPGRADE_ROLLBACK_REPLY) != TTP_ERROR);
    TTPReplyMsg replyMsg;

    TTP_ASSERT_RET_VOID(msg->type >= RepairType::RT_SEND && msg->type < RepairType::RT_BUTT);
    TTP_ASSERT_RET_VOID(msg->step > 0 && msg->step < INT64_MAX);

    if (processorStatus_.load() != PS_PAUSE) {
        TTP_LOG_ERROR("rank:" << rank_ << " rollback failed, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else {
        RepairContext rc;
        rc.srcRank.push_back(rank_);
        rc.dstRank.push_back(rank_);
        rc.type = msg->type;
        rc.step = msg->step;
        lockStep_ = msg->step;
        repairId_ = msg->repairId;
        rc.repairId = repairId_;

        uint32_t paramLen = msg->dataLen;
        rc.zitParam = std::string(&msg->data[0]);
        TTP_ASSERT_RET_VOID(paramLen == rc.zitParam.length() + 1);
        replyMsg.status = EventProcess(PROCESSOR_EVENT_UPGRADE_ROLLBACK, &rc, sizeof(RepairContext));
        if (replyMsg.status != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " rollback failed, ret: " << replyMsg.status);
        }
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_UPGRADE_ROLLBACK_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply rollback result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply rollback result failed");
    }
}

void Processor::HandlePause(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(PauseMsg));
    TTP_LOG_INFO("rank:" << rank_ << " receive pause message");
    PauseMsg *msg = reinterpret_cast<PauseMsg *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msg->sn, TTP_MSG_OP_PAUSE_REPLY) != TTP_ERROR);
    TTP_ASSERT_RET_VOID(msg->step > 0 && msg->step < INT64_MAX);

    TTPReplyMsg replyMsg;
    if (processorStatus_.load() != PS_NORMAL) {
        TTP_LOG_ERROR("rank:" << rank_ << " pause train failed, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else {
        PauseContext pc {msg->step, msg->hotSwitch};
        replyMsg.status = EventProcess(PROCESSOR_EVENT_PAUSE, &pc, sizeof(PauseContext));
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_PAUSE_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply device pause result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply device pause result failed");
    }
}

void Processor::HandleContinue(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(PauseMsg));
    TTP_LOG_INFO("rank:" << rank_ << " receive continue message");
    PauseMsg *msg = reinterpret_cast<PauseMsg *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msg->sn, TTP_MSG_OP_CONTINUE_REPLY) != TTP_ERROR);

    TTPReplyMsg replyMsg;
    if (processorStatus_.load() != PS_NORMAL) {
        TTP_LOG_ERROR("rank:" << rank_ << " device continue failed, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else {
        int64_t pauseStep = msg->step;
        replyMsg.status = EventProcess(PROCESSOR_EVENT_CONTINUE, &pauseStep, sizeof(pauseStep));
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);

    auto result = newClient_->Send(TTP_MSG_OP_PAUSE_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply device continue result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply device continue result failed");
    }
}

void Processor::HandleNotifyNormal(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(uint16_t));
    TTP_LOG_INFO("rank:" << rank_ << " receive notify normal message");
    uint16_t *receiveSn = reinterpret_cast<uint16_t *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(*receiveSn, TTP_MSG_OP_NORMAL_REPLY) != TTP_ERROR);

    ResultAndHBReplyMsg reply;
    reply.ret = TTP_OK;

    uint32_t nowStatus;
    do {
        nowStatus = processorStatus_.load();
        if (nowStatus != PS_PAUSE) {
            TTP_LOG_ERROR("rank:" << rank_ << " not pause, now status:" << nowStatus);
            reply.ret = TTP_ERROR;
            break;
        }
    } while (!processorStatus_.compare_exchange_strong(nowStatus, PS_NORMAL));

    if (reply.ret == TTP_OK) {
        std::unique_lock<std::mutex> lck(statusMutex_); // 异常状态清理
        trainStatus_.step = lockStep_;
        trainStatus_.npu_status = TTP_STATUS_NORMAL;
        trainStatus_.run_status = TTP_STATUS_NORMAL;
        reply.hb.rank = rank_;
        reply.hb.status = trainStatus_;
        reply.repairStep = repairStep_;
        envClearFlag_ = false;
        repairStep_ = -1;
        isPrelockOk_.store(false);
        waitResult_ = TTP_OK;
        sem_post(&waitSem_); // sem_up wait_next_action
    }
    reply.sn = actionSn_.load();
    SaveBackupStatus({reply.ret, reply.sn, reply.hb.rank});

    auto result = newClient_->Send(TTP_MSG_OP_NORMAL_REPLY,
        reinterpret_cast<uint8_t *>(&reply), sizeof(ResultAndHBReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply notify normal result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply notify normal result failed");
    }
}

void Processor::HandleCollection(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(uint16_t));
    uint16_t *receiveSn = reinterpret_cast<uint16_t *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(*receiveSn, TTP_MSG_OP_COLLECTION_REPLY) != TTP_ERROR);

    ResultAndHBReplyMsg reply;
    reply.ret = TTP_OK;
    reply.hb.rank = rank_;
    reply.hb.status = trainStatus_; // 带回状态,保证controller拥有最新状态
    reply.sn = actionSn_.load();
    SaveBackupStatus({reply.ret, reply.sn, reply.hb.rank});

    auto result = newClient_->Send(TTP_MSG_OP_COLLECTION_REPLY,
        reinterpret_cast<uint8_t *>(&reply), sizeof(ResultAndHBReplyMsg));
    if (result != ACC_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " send collection reply failed, ret: " << result);
    }
}

void Processor::HandlePrelock(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(PrelockMsg));
    TTP_LOG_INFO("rank:" << rank_ << " receive prelock message");
    PrelockMsg *recvMsg = reinterpret_cast<PrelockMsg *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(recvMsg->sn, TTP_MSG_OP_PRELOCK_REPLY) != TTP_ERROR);
    TTP_ASSERT_RET_VOID(recvMsg->step > 0 && recvMsg->step < INT64_MAX);

    ResultAndHBReplyMsg reply;
    reply.ret = TTP_ERROR;

    uint32_t nowStatus;
    do {
        nowStatus = processorStatus_.load();
        if (nowStatus != PS_NORMAL && nowStatus != PS_PAUSE) {
            reply.ret = TTP_ERROR;
            ReplyPrelock(reply, nowStatus);
            return;
        }
    } while (!processorStatus_.compare_exchange_strong(nowStatus, PS_PAUSE));

    statusMutex_.lock();
    limitStep_.store(recvMsg->step);
    lockStep_ = recvMsg->step;
    statusMutex_.unlock();

    uint32_t sleepTime = 0;
    while (sleepTime < PRELOCK_RETRY_TIMES) { // 死等,理论上不用等待太久
        statusMutex_.lock();
        auto range = GetNowStep();
        statusMutex_.unlock();

        // 保证 卡被锁住了且step满足 或 卡状态异常 break
        if ((trainStatus_.data_status == Updated && (range.first == (recvMsg->step) || range.second == (recvMsg->step)))
            || trainStatus_.npu_status != TTP_STATUS_NORMAL || trainStatus_.run_status != TTP_STATUS_NORMAL) {
            reply.ret = TTP_OK;
            break;
        }

        if (range.first > (recvMsg->step)) {
            TTP_LOG_WARN("rank:" << rank_ << " prelock return retry, now_step: "
                << range.first << " limit_step" << (recvMsg->step));
            reply.ret = TTP_NEED_RETRY;
            break;
        }

        sleepTime++;
        usleep(TTP_WAIT_TIME_1MS);
    }

    if (sleepTime >= PRELOCK_RETRY_TIMES) {
        TTP_LOG_ERROR("rank:" << rank_ << " do prelock timeout");
        reply.ret = TTP_TIMEOUT;
    }

    ReplyPrelock(reply, nowStatus);
}

void Processor::ReplyPrelock(ResultAndHBReplyMsg &reply, uint32_t nowStatus)
{
    if (reply.ret == TTP_ERROR) {
        TTP_LOG_ERROR("rank:" << rank_ << " not normal, now status:" << nowStatus);
    }

    reply.hb.rank = rank_;
    reply.hb.status = trainStatus_; // 带回状态,保证controller拥有最新状态
    reply.sn = actionSn_.load();
    SaveBackupStatus({reply.ret, reply.sn, reply.hb.rank});
    auto result = newClient_->Send(TTP_MSG_OP_PRELOCK_REPLY,
        reinterpret_cast<uint8_t *>(&reply), sizeof(ResultAndHBReplyMsg));
    if (result != ACC_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " send prelock reply failed, ret: " << result);
    } else {
        TTP_LOG_INFO("rank:" << rank_ << " reply prelock msg success");
    }
    TTP_LOG_INFO("rank:" << rank_ << " status after prelock: " << " step: " << trainStatus_.step <<
        " npu_status: " << static_cast<int32_t>(trainStatus_.npu_status) <<
        " run_status: " << static_cast<int32_t>(trainStatus_.run_status) <<
        " data_aval: " << static_cast<int32_t>(trainStatus_.data_aval) <<
        " data_status: " << static_cast<int32_t>(trainStatus_.data_status));
}

void Processor::HandleControllerReply(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(RegisterReply));
    RegisterReply *reply = reinterpret_cast<RegisterReply *>(data);
    TTP_ASSERT_RET_VOID(reply->ret >= TTP_OK && reply->ret < TTP_BUTT);
    replyRet_ = static_cast<TResult>(reply->ret);
    repairId_ = reply->repairId;
    hotSwitch_ = reply->hotSwitch;
    TTP_LOG_DEBUG("processor receive reply, ret:" << (reply->ret));
    replySem_.PthreadSignal();
    return;
}

void Processor::HandleExit(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(uint16_t));
    readyToExit_.store(true);
    TTP_LOG_INFO("rank:" << rank_ << " receive exit message");
    uint16_t *receiveSn = reinterpret_cast<uint16_t *>(data);
    TResult snCheckRet = UpdateSerialNumber(*receiveSn);
    if (snCheckRet == TTP_ERROR) {
        TTP_LOG_WARN("rank:" << rank_ << " exit message sn:" << *receiveSn <<", less than actionSn: " <<
            actionSn_.load() << ", drop msg, no response.");
    }

    EventProcess(PROCESSOR_EVENT_EXIT, nullptr, 0);
    Destroy(true);

    waitResult_ = TTP_ERROR;
    sem_post(&waitSem_); // sem_up wait_next_action, maybe transfer from uce
}

void Processor::HandleDestroyNofity(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(uint16_t));
    readyToExit_.store(true);
    newClient_->SetMaxReconnCnt(0);
    TTP_LOG_INFO("rank:" << rank_ << " receive controller destroy notify message");
}

void Processor::HandleBroadcast(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_LOG_INFO("rank:" << rank_ << " receive broadcast message");
    BroadcastIpMsg *msg = reinterpret_cast<BroadcastIpMsg *>(data);
    TTP_ASSERT_RET_VOID(msg->JudgeVariableValid(len));
    TResult snCheckRet = UpdateSerialNumber(msg->sn);
    if (snCheckRet == TTP_ERROR) {
        TTP_LOG_WARN("rank:" << rank_ << " broadcast backup controller message sn:" << msg->sn <<
            ", less than actionSn: " << actionSn_.load() << ", drop msg, no response.");
        return;
    }
    HandleBackupCtrlList(data);
    StartBackupController(data);
}

// rank1:ip1|rank2:ip2|...
void Processor::HandleBackupCtrlList(uint8_t *data)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    BroadcastIpMsg *msg = reinterpret_cast<BroadcastIpMsg *>(data);
    std::string ipStr(msg->arr, msg->ipLen - 1);
    // split string & add to controllerIps_
    std::istringstream ss(ipStr);
    std::string eachIp;
    while (std::getline(ss, eachIp, '|')) {
        size_t index = eachIp.find(':');
        // 校验输入rank:ip是否合法
        if (index == std::string::npos || index == 0) {
            TTP_LOG_WARN("processor:" << rank_ << " receive invalid rank or controller ip:" << eachIp);
            continue;
        }
        // 1，校验IP部分
        std::string pureIp = eachIp.substr(index + 1);
        if (!IsValidIpV4(pureIp)) {
            TTP_LOG_WARN("processor:" << rank_ << " receive invalid controller ip:" << eachIp);
            continue;
        }
        // 2，校验rank部分
        std::string pureRankStr = eachIp.substr(0, index);
        int32_t pureRankIdx = -1;
        try {
            pureRankIdx = std::stoi(pureRankStr);
        }
        catch (...) {
            TTP_LOG_WARN("processor:" << rank_ << " receive invalid backup controller rank:" << eachIp);
            continue;
        }
        if (pureRankIdx < 0 || pureRankIdx >= worldSize_) {
            TTP_LOG_WARN("processor:" << rank_ << " receive invalid backup controller rank:" << eachIp);
            continue;
        }

        controllerIps_.push_back(eachIp);
    }
    TTP_LOG_INFO("rank:" << rank_ << " update backup controller list success, controllerIps,size: " <<
        controllerIps_.size());
}

void Processor::HandleDowngradeRebuild(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_LOG_INFO("rank:" << rank_ << " receive downgrade rebuild message");
    TTPReplyMsg replyMsg;
    ZitRebuildContext rc;
    DowngradeRunMsg *msg = reinterpret_cast<DowngradeRunMsg *>(data);
    TTP_ASSERT_RET_VOID(msg->JudgeVariableValid(len, worldSize_));
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msg->sn, TTP_MSG_OP_DOWNGRADE_REBUILD_REPLY) != TTP_ERROR);
    TTP_ASSERT_RET_VOID(RepairBarrierWithFramework() == TTP_OK);

    rc.commGroupIdx.resize(msg->num);
    rc.commGroups.resize(msg->num);
    repairId_ = msg->repairId;
    rc.repairId = repairId_;
    int32_t pIdx = 0;
    int32_t foundCount = 0;
    for (uint32_t idx = 0; idx < msg->num; idx++) {
        rc.commGroupIdx[idx] = msg->ranks[pIdx++];
        for (int32_t i = 1; i <= msg->ranks[pIdx]; i++) {
            rc.commGroups[idx].push_back(msg->ranks[pIdx + i]);
            foundCount += (msg->ranks[pIdx + i] == rank_);
        }
        pIdx += msg->ranks[pIdx] + 1;
    }
    int32_t paramLen = msg->ranks[pIdx++];
    rc.zitParam = std::string(reinterpret_cast<char*>(&msg->ranks[pIdx]));
    TTP_ASSERT_RET_VOID(paramLen == rc.zitParam.length() + 1);
    // local rank_ not in dpGroup, can not rebuild comm groups
    if (foundCount != static_cast<int32_t>(msg->num)) {
        TTP_LOG_ERROR("rank:" << rank_ << " receive invalid comm group in downgrade rebuild.");
        replyMsg.status = TTP_ERROR;
    } else if (processorStatus_.load() != PS_PAUSE) {
        TTP_LOG_ERROR("rank:" << rank_ << " downgrade rebuild falied, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else {
        replyMsg.status = EventProcess(PROCESSOR_EVENT_DOWNGRADE_REBUILD, &rc, sizeof(ZitRebuildContext));
        if (replyMsg.status != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " downgrade rebuild failed, ret: " << replyMsg.status);
        }
    }

    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);
    auto result = newClient_->Send(TTP_MSG_OP_DOWNGRADE_REBUILD_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply downgrade rebuild result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply downgrade rebuild result failed");
    }
}

void Processor::HandleUpgradeRebuild(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_LOG_INFO("rank:" << rank_ << " receive upgrade comm operation message");
    RebuildGroupMsg *msgTmp = reinterpret_cast<RebuildGroupMsg *>(data);
    TTP_ASSERT_RET_VOID(msgTmp->JudgeVariableValid(len, worldSize_));
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msgTmp->sn, TTP_MSG_OP_UPGRADE_REBUILD_REPLY) != TTP_ERROR);
    TTPReplyMsg replyMsg;
    replyMsg.status = EventProcess(PROCESSOR_EVENT_UPGRADE_REBUILD, data, sizeof(RebuildGroupMsg));
    if (replyMsg.status != TTP_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << ", pt communication operate failed, ret: " << replyMsg.status);
    }

    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);
    auto result = newClient_->Send(TTP_MSG_OP_UPGRADE_REBUILD_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply upgrade comm operation result success");
        if (trainStatus_.run_status == TTP_STATUS_INIT_FINISH) {
            waitResult_ = TTP_OK;
            sem_post(&waitSem_); // sem_up wait_next_action
            TTP_LOG_INFO("rank:" << rank_ << " sem up training thread after upgrade rebuild group");
        }
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply upgrade comm operation result failed");
    }
}

void Processor::HandlePtComm(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_LOG_INFO("rank:" << rank_ << " receive pt comm operation message");
    RebuildGroupMsg *msgTmp = reinterpret_cast<RebuildGroupMsg *>(data);
    TTP_ASSERT_RET_VOID(msgTmp->JudgeVariableValid(len, worldSize_));
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(msgTmp->sn, TTP_MSG_OP_PT_COMM_REPLY) != TTP_ERROR);

    TTPReplyMsg replyMsg;
    replyMsg.status = EventProcess(PROCESSOR_EVENT_PT_COMM_OPERATE, data, sizeof(RebuildGroupMsg));
    if (replyMsg.status != TTP_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << ", pt communication operate failed, ret: " << replyMsg.status);
    }

    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);
    auto result = newClient_->Send(TTP_MSG_OP_PT_COMM_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply pt comm operation result success");
        if (trainStatus_.run_status == TTP_STATUS_INIT_FINISH) {
            waitResult_ = TTP_OK;
            sem_post(&waitSem_); // sem_up wait_next_action
            TTP_LOG_INFO("rank:" << rank_ << " sem up training thread after rebuild group");
        }
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply pt comm operation result failed");
    }
}

void Processor::HandleLaunchTcpStoreServer(uint8_t *data, uint32_t len)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(len == sizeof(uint16_t));
    TTP_ASSERT_RET_VOID(worldSize_ > 0 && worldSize_ <= TTP_MAX_WORLD_SIZE);
    TTP_LOG_INFO("rank:" << rank_ << " receive launch tcp store server message");
    uint16_t *receiveSn = reinterpret_cast<uint16_t *>(data);
    TTP_ASSERT_RET_VOID_NO_LOG(CheckMsgSnAndReply(*receiveSn, TTP_MSG_OP_STOP_REPLY) != TTP_ERROR);

    TTPReplyMsg replyMsg;
    if (processorStatus_.load() != PS_PAUSE) {
        TTP_LOG_ERROR("rank:" << rank_ << " launch tcp store server failed, status error: " << processorStatus_.load());
        replyMsg.status = TTP_ERROR;
    } else {
        // Get tcp store server port
        std::string url;
        TResult ret = GetTcpStoreUrl(url);
        if (ret != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " launch tcp store server failed, url:" << url);
            replyMsg.status = TTP_ERROR;
        } else {
            LaunchTcpStoreMsg serverInfo {0, worldSize_, url};
            std::thread tmpThread(&Processor::LaunchTcpStoreServerThread, this, &serverInfo);
            std::thread tcpStoreServer = std::move(tmpThread);
            tcpStoreServer.detach();
            sleep(1);  // wait for tcp store server start
            replyMsg.status = ret;
        }
    }
    replyMsg.sn = actionSn_.load();
    replyMsg.rank = rank_;
    SaveBackupStatus(replyMsg);
    auto result = newClient_->Send(TTP_MSG_OP_LAUNCH_STORE_SERVER_REPLY,
        reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
    if (result == ACC_OK) {
        TTP_LOG_INFO("rank:" << rank_ << " reply launch tcp store server result success");
    } else {
        TTP_LOG_ERROR("rank:" << rank_ << " reply launch tcp store server result failed");
    }
}

TResult Processor::SetDumpResult(int32_t result)
{
    TTP_ASSERT_RETURN(result >= TTP_OK && result < TTP_BUTT, TTP_ERROR);
    dumpRet_ = result;
    dumpCond_.PthreadSignal();
    return TTP_OK;
}

TResult Processor::Register2Controller()
{
    int32_t regMsgLen = sizeof(RegisterMsg);
    void *mem = malloc(regMsgLen);
    TTP_ASSERT_RETURN(mem != nullptr, TTP_ERROR);
    RegisterMsg *msg = reinterpret_cast<RegisterMsg *>(mem);
    msg->rank = rank_;

    uint32_t timeRetried = 0;
    TTP_LOG_INFO("rank:" << rank_ << " register to controller: " << IpPort());
    replySem_.SignalClean();
    while (timeRetried < REGISTER_REPORT_RETRY_TIMES) {
        replyRet_ = TTP_ERROR;
        auto result = newClient_->Send(TTP_MSG_OP_REGISTER, reinterpret_cast<uint8_t *>(msg), regMsgLen);
        if (result == ACC_OK) {
            auto ret = replySem_.PthreadTimedwaitSecs(REGISTER_REPORT_WAIT_TIME);
            if (ret != ETIMEDOUT && replyRet_ != TTP_ERROR) {
                break;
            }
        }

        timeRetried++;
        TTP_LOG_WARN("rank:" << rank_ << " register to controller: " << IpPort() << ", retry times:" << timeRetried);
    }

    free(mem);
    mem = nullptr;
    if (timeRetried == REGISTER_REPORT_RETRY_TIMES) {
        TTP_LOG_ERROR("rank:" << rank_ << " register to controller: " << IpPort() << " failed, after 5 times");
        return TTP_ERROR;
    }
    TTP_LOG_INFO("rank:" << rank_ << " register to: " << IpPort() << " success, after " << timeRetried << " times");
    if (replyRet_ == TTP_REPAIR) { // repair node start
        trainStatus_.run_status = TTP_STATUS_ISOLATE;
    }
    return TTP_OK;
}

TResult Processor::ReportInfo2Controller(std::vector<int32_t> replicaCnt, std::vector<int32_t> replicaShift)
{
    TTP_ASSERT_RETURN(ProcessorIsRunning(), TTP_ERROR);
    int32_t groupNum = static_cast<int32_t>(rankList_.size());
    int32_t unitNum = groupNum + groupNum;
    for (auto &group : rankList_) {
        unitNum += static_cast<int32_t>(group.size()) + 1;
    }
    TTP_ASSERT_RETURN(!IsOverflow(unitNum, sizeof(int32_t)), TTP_ERROR);
    int32_t replicaMsgLen = sizeof(ReplicaMsg) + unitNum * sizeof(int32_t);
    uint8_t *mem = new (std::nothrow) uint8_t[replicaMsgLen];
    TTP_ASSERT_RETURN(mem != nullptr, TTP_ERROR);
    ReplicaMsg *msg = reinterpret_cast<ReplicaMsg *>(mem);
    msg->num = groupNum;
    msg->rank = rank_;
    msg->enableArf = arfSwitch_;
    msg->enableUce = uceSwitch_;
    msg->enableZit = zitSwitch_;
    int32_t idx = groupNum + groupNum;
    for (int32_t i = 0; i < groupNum; i++) {
        msg->ranks[i] = replicaCnt[i];
        msg->ranks[i + groupNum] = replicaShift[i];
        msg->ranks[idx++] = static_cast<int32_t>(rankList_[i].size());
        for (auto rk : rankList_[i]) {
            msg->ranks[idx++] = rk;
        }
    }

    uint32_t timeRetried = 0;
    TTP_LOG_INFO("rank:" << rank_ << " report info to controller: " << IpPort());
    replySem_.SignalClean();
    while (timeRetried < REGISTER_REPORT_RETRY_TIMES) {
        replyRet_ = TTP_ERROR;
        auto result = newClient_->Send(TTP_MSG_OP_INIT_REPORT, reinterpret_cast<uint8_t *>(msg), replicaMsgLen);
        if (result == ACC_OK) {
            auto ret = replySem_.PthreadTimedwaitSecs(REGISTER_REPORT_WAIT_TIME);
            if (ret != ETIMEDOUT && replyRet_ != TTP_ERROR) {
                break;
            }
        }

        timeRetried++;
        TTP_LOG_WARN("rank:" << rank_ << " report info to controller:" << IpPort() << ", retry times:" << timeRetried);
    }

    delete[] mem;
    mem = nullptr;
    return IsMaxRetryTime(timeRetried);
}

TResult Processor::IsMaxRetryTime(uint32_t timeRetried)
{
    if (timeRetried == REGISTER_REPORT_RETRY_TIMES) {
        TTP_LOG_ERROR("rank:" << rank_ << " report info to controller: " << IpPort() <<
            " failed, after " << REGISTER_REPORT_RETRY_TIMES << " times");
        return TTP_ERROR;
    }
    TTP_LOG_INFO("rank:" << rank_ << " report info to: " << IpPort() << " success, after " << timeRetried << " times");
    return TTP_OK;
}

TResult Processor::ReportDp2Controller()
{
    TTP_ASSERT_RETURN(ProcessorIsRunning(), TTP_ERROR);
    uint32_t msgLen = sizeof(DpMsg) + dpList_.size() * sizeof(int32_t);
    void *mem = malloc(msgLen);
    TTP_ASSERT_RETURN(mem != nullptr, TTP_ERROR);
    DpMsg *msg = reinterpret_cast<DpMsg *>(mem);
    msg->rank = rank_;
    msg->dpNum = dpList_.size();
    for (uint32_t i = 0; i != dpList_.size(); i++) {
        msg->dpList[i] = dpList_[i];
    }

    uint32_t timeRetried = 0;
    replySem_.SignalClean();
    while (timeRetried < REGISTER_REPORT_RETRY_TIMES) {
        replyRet_ = TTP_ERROR;
        auto result = newClient_->Send(TTP_MSG_OP_DP_REPORT, reinterpret_cast<uint8_t *>(msg), msgLen);
        if (result == ACC_OK) {
            auto ret = replySem_.PthreadTimedwaitSecs(REGISTER_REPORT_WAIT_TIME);
            if (ret != ETIMEDOUT && replyRet_ != TTP_ERROR) {
                break;
            }
        }

        timeRetried++;
        TTP_LOG_WARN("rank:" << rank_ << " report dp group to controller:" << IpPort() <<
            ", retry times:" << timeRetried);
    }

    free(mem);
    mem = nullptr;
    if (timeRetried == REGISTER_REPORT_RETRY_TIMES) {
        TTP_LOG_ERROR("rank:" << rank_ << " report dp group to controller: " << IpPort() << " failed, after " <<
                              REGISTER_REPORT_RETRY_TIMES << " times");
        return TTP_ERROR;
    }
    TTP_LOG_INFO("rank:" << rank_ << " report dp group to: " << IpPort() << " success, after " <<
        timeRetried << " times");
    return TTP_OK;
}

// reconnect retry realize inside
TResult Processor::Connect2Controller()
{
    AccConnReq req{};
    req.rankId = rank_;
    req.magic = 0;
    req.version = 1;
    TTP_LOG_DEBUG("AccConnReq, rank:" << rank_ << ", version: " << req.version);

    uint32_t maxRetryTimes = GetEnvValue2Uint32("TTP_RETRY_TIMES", ConnectRetryTimes.minVal,
                                                ConnectRetryTimes.maxVal, ConnectRetryTimes.defaultVal);
    TTP_LOG_INFO("[env] TTP_RETRY_TIMES:" << maxRetryTimes);
    auto result = newClient_->Connect(req, maxRetryTimes);
    if (result != ACC_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " connect to: " << IpPort() <<
            " failed. Maybe environment variable TTP_RETRY_TIMES should be set a larger value.");
        return TTP_ERROR;
    }
    TTP_LOG_INFO("rank:" << rank_ << " connect to: " << IpPort() << " success");
    return TTP_OK;
}

TResult Processor::ReStart()
{
    int32_t nums = static_cast<int32_t>(controllerIps_.size());
    // ip用尽 || 第一次心跳发送失败，没有备份ip
    if ((controllerIdx_ + 1 == nums) || (nums == 1)) {
        TTP_LOG_ERROR("rank:" << rank_ << " out of ip, connect2backup controller failed");
        return TTP_ERROR;
    }

    TResult result = TTP_OK;
    std::string backupIp = "";
    std::size_t index = 0;
    while (controllerIdx_ < nums - 1 && !readyToExit_.load()) {
        controllerIdx_++;
        std::string rankIp = controllerIps_[controllerIdx_];
        index = rankIp.find(':');
        if (index == std::string::npos) {
            TTP_LOG_ERROR("rank:" << rank_ << " receive invalid controller ip:" << rankIp);
            continue;
        }

        backupIp = rankIp.substr(index + 1);
        newClient_->Disconnect();
        newClient_->SetServerIpAndPort(backupIp, port_);

        // 1. connect to controller
        TTP_LOG_INFO("rank:" << rank_ << " connect to backup controller: " << backupIp);
        result = Connect2Controller();
        if (result != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " connect to backup controller failed");
            continue;
        }

        // 2. register to controller
        TTP_LOG_INFO("rank:" << rank_ << " register to backup controller: " << backupIp);
        result = Register2Controller();
        if (result != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " register to backup controller failed");
            continue;
        }

        // 3. report to controller
        TTP_LOG_INFO("rank:" << rank_ << " report to backup controller: " << backupIp);
        result = ReportInfo2Controller(replicaCnt_, replicaShift_);
        if (result != TTP_OK) {
            TTP_LOG_ERROR("rank:" << rank_ << " report to backup controller failed");
            continue;
        }
        break;
    }

    if (result != TTP_OK) {
        TTP_LOG_ERROR("restart failed, rank:" << rank_ << ", backup controller:" << backupIp);
    }

    return result;
}

void Processor::StartBackupController(uint8_t *data)
{
    TTP_ASSERT_RET_VOID(data != nullptr);
    TTP_ASSERT_RET_VOID(worldSize_ > 0 && worldSize_ <= TTP_MAX_WORLD_SIZE);
    BroadcastIpMsg *msg = reinterpret_cast<BroadcastIpMsg *>(data);
    size_t nums = controllerIps_.size();
    if (nums == 1) {
        TTP_LOG_ERROR("backup controllerIps_ is empty, no backup controller to start");
        return;
    }

    bool enableARF = (msg->enableARF != 0);
    bool enableZIT = (msg->enableZIT != 0);

    // rank:ip|rank:ip|...
    for (size_t i = 1; i < nums; i++) {
        std::istringstream rankIp(controllerIps_[i]);
        std::string rankStr;
        std::string ipStr;
        std::getline(rankIp, rankStr, ':');
        std::getline(rankIp, ipStr, ':');
        TTP_LOG_INFO("rank:" << rank_ << ", [rank:ip]:" << controllerIps_[i]
            << ", backup controller nums:" << (nums - 1));
        int32_t rankIn = 0;
        try {
            rankIn = std::stoi(rankStr);
        } catch (...) {
            TTP_LOG_WARN("recv invalid rank from controller!");
            continue;
        }
        if (rank_ != rankIn) {
            continue;
        }

        TTP_LOG_INFO("init backup controller, backup rank:" << rankIn << ", worldSize_: " << worldSize_);
        auto result = Controller::GetInstance()->Initialize(rankIn, worldSize_, localCopySwitch_, enableARF, enableZIT);
        if (result != TTP_OK) {
            TTP_LOG_ERROR("backup controller " << controllerIps_[i] << " initialize failed");
            break;
        }
        result = Controller::GetInstance()->Start(ipStr, port_, tlsOption_, i);
        if (result != TTP_OK) {
            TTP_LOG_ERROR("backup controller " << controllerIps_[i] << " start failed");
            break;
        }

        TTP_LOG_INFO("processor:" << rank_ << " backup controller " << controllerIps_[i] << " start success");
        startBackup_ = static_cast<int32_t>(i);
        break;
    }
}

TResult Processor::RegisterEventHandler(ProcessorEvent event, ProcessorEventHandle handle)
{
    TTP_ASSERT_RETURN(event < PROCESSOR_EVENT_BUTT, TTP_ERROR);
    TTP_ASSERT_RETURN(handle != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(eventHandleList_[event] == nullptr, TTP_ERROR);
    eventHandleList_[event] = handle;
    return TTP_OK;
}

TResult Processor::WaitNextAction()
{
    TTP_LOG_INFO("rank:" << rank_ << " start wait next action..");
    int32_t waitRet = 0;
    int32_t errorNumber = 0;
    do {
        waitRet = sem_wait(&waitSem_);
        errorNumber = errno;
        if (waitRet != 0) {
            TTP_LOG_WARN("rank:" << rank_ << "wait next action abnormal.. ret: " <<
                          waitRet <<", errno " << errorNumber);
        }
        if (errorNumber == EINTR) {
            sleep(1);
        }
    } while (waitRet != 0 && errorNumber == EINTR);

    TTP_LOG_INFO("rank:" << rank_ << " end wait next action.. ret: " << waitResult_ << ",waitRet " << waitRet);
    return static_cast<TResult>(waitResult_);
}

TResult Processor::WaitRepairAction()
{
    TTP_LOG_INFO("rank:" << rank_ << " the next action is: repair..");
    repairWaitCond_.PthreadSignal();
    return WaitNextAction();
}

TResult Processor::ReportStatus(ReportState state)
{
    TTP_ASSERT_RETURN(ProcessorIsRunning(), TTP_ERROR);

    statusMutex_.lock();
    if (state == ReportState::RS_UCE) {
        trainStatus_.npu_status = TTP_STATUS_UCE_HIGH;
    } else if (state == ReportState::RS_UCE_CORRUPTED) {
        trainStatus_.npu_status = TTP_STATUS_UCE_CORRUPTED;
    } else if (state == ReportState::RS_UNKNOWN) {
        trainStatus_.run_status = TTP_STATUS_ABNORMAL;
    } else if (state == ReportState::RS_PREREPAIR_FINISH) {
        trainStatus_.run_status = TTP_STATUS_PREREPAIR_FINISH;
    } else if (state == ReportState::RS_INIT_FINISH) {
        trainStatus_.run_status = TTP_STATUS_INIT_FINISH;
    } else if (state == ReportState::RS_HCCL_FAILED) {
        trainStatus_.npu_status = TTP_STATUS_HCCL_FAILED;
    } else if (state == ReportState::RS_STEP_FINISH) {
        trainStatus_.run_status = TTP_STATUS_STEP_FINISH;
    } else if (state != ReportState::RS_NORMAL) {
        TTP_LOG_ERROR("rank:" << rank_ << " report invalid status: " << static_cast<int>(state));
        return TTP_ERROR;
    }
    if (readyToExit_.load()) {
        trainStatus_.run_status = TTP_STATUS_EXIT;
    }
    statusMutex_.unlock();

    processorStatus_.store(PS_PAUSE);
    if (HeartbeatSend() != TTP_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " send status msg to controller: " << IpPort() << " failed");
        return TTP_ERROR;
    }
    TTP_LOG_INFO("rank:" << rank_ << " report status: " << int32_t(trainStatus_.npu_status) <<
        ", run_status: " << int32_t(trainStatus_.run_status));
    return TTP_OK;
}

TResult Processor::ReportLoadCkptStep(int64_t step)
{
    TTP_ASSERT_RETURN(step > 0 && step < INT64_MAX, TTP_ERROR);
    repairStep_ = step;
    return TTP_OK;
}

TResult Processor::EventProcess(ProcessorEvent eventCode, void *ctx, int ctxSize)
{
    if (eventHandleList_[eventCode] == nullptr) {
        TTP_LOG_WARN("event handle is null! event: " << eventCode);
        return TTP_ERROR;
    }

    return static_cast<TResult>(eventHandleList_[eventCode](ctx, ctxSize));
}

TResult Processor::CheckMsgSnAndReply(uint16_t sn, MsgOpCode replyOp)
{
    TResult snCheckRet = UpdateSerialNumber(sn, replyOp);
    if (snCheckRet == TTP_ERROR) {
        TTP_LOG_WARN("rank:" << rank_ << ", reply MsgOpCode: "<< replyOp << ", drop msg, no response.");
        return TTP_ERROR;
    }
    if (snCheckRet == TTP_NEED_RETRY) {
        TTP_LOG_WARN("rank:" << rank_ << " msg sn:" << sn <<", equal to actionSn: " <<
            actionSn_.load() << ", which was done success before, send reply only.");

        int32_t result = ACC_OK;
        if (extraReplyHandleList_[replyOp] == nullptr) {
            TTPReplyMsg replyMsg;
            GetBackupStatus(sn, replyMsg);
            result = newClient_->Send(replyOp, reinterpret_cast<uint8_t *>(&replyMsg), sizeof(TTPReplyMsg));
        } else {
            result = extraReplyHandleList_[replyOp](sn, replyOp);
        }

        if (result == ACC_OK) {
            TTP_LOG_INFO("rank:" << rank_ << " reply result success, MsgOp: " << replyOp <<
                ", sn equal only re-sent reply.");
        } else {
            TTP_LOG_ERROR("rank:" << rank_ << " reply result failed, MsgOp: " << replyOp);
        }
        return TTP_ERROR;
    }
    return TTP_OK;
}

TResult Processor::UpdateSerialNumber(uint16_t sn, MsgOpCode replyOp)
{
    if (sn < actionSn_.load()) {
        TTP_LOG_INFO("rank:"<< rank_ << ", outdated actionSn, will not response! msg sn: " <<
            sn << ", current sn: " << actionSn_.load());
        return TTP_ERROR;   // skip this Action
    }
    if (sn > actionSn_.load()) {
        TTP_LOG_DEBUG("rank:"<< rank_ << ", normal actionSn increase, will response and update actionSn!"  <<
            " msg sn: " << sn << ", current sn: " << actionSn_.load());
        actionSn_.store(sn);
        return TTP_OK;      // do this Action
    }
    TTPReplyMsg msg;
    TResult ret = GetBackupStatus(sn, msg);
    if (ret != TTP_OK) {  // can't find reply with same sn
        TTP_LOG_INFO("rank:"<< rank_ << ", repeated actionSn, reply msg not found," <<
            " will retry! msg sn: " << sn << ", current sn: " << actionSn_.load());
        return TTP_OK;    // do this action
    }
    // do this action success before, or can't execute op twice, only send reply
    if (msg.status == TTP_OK || allowRetryOp.count(replyOp) == 0) {
        TTP_LOG_INFO("rank:" << rank_ << ", repeated actionSn, reply status: " << msg.status <<
            ", will only send reply! msg sn: " << sn << ", current sn: " << actionSn_.load());
        return TTP_NEED_RETRY; // only send reply
    }

    TTP_LOG_INFO("rank:"<< rank_ << ", repeated actionSn, failed before and support retry," <<
        " will retry! msg sn: " << sn << ", current sn: " << actionSn_.load());
    return TTP_OK;   // do this Action
}

TResult Processor::GetBackupStatus(uint16_t sn, TTPReplyMsg& msg)
{
    TResult ret = TTP_ERROR;
    if (sn == replyMsgBackup_.sn) {
        msg = replyMsgBackup_;
        ret = TTP_OK;
    }
    return ret;
}

void Processor::SaveBackupStatus(const TTPReplyMsg& msg)
{
    replyMsgBackup_ = msg;
}

int32_t Processor::ProcessResultAndHBReply(uint16_t sn, MsgOpCode op)
{
    TTPReplyMsg replyMsg;
    TResult ret = GetBackupStatus(sn, replyMsg);
    if (ret != TTP_OK) {
        TTP_LOG_INFO("can not find replyMsgBackup_, rank:" << rank_ << ", MsgOp: " << op << ", sn: " << sn);
        return TTP_ERROR;
    }

    ResultAndHBReplyMsg msg;
    msg.ret = static_cast<TResult>(replyMsg.status);
    msg.sn = sn;
    msg.hb.rank = rank_;
    statusMutex_.lock();
    msg.hb.status = trainStatus_; // 带回状态,保证controller拥有最新状态
    statusMutex_.unlock();

    auto result = newClient_->Send(op, reinterpret_cast<uint8_t *>(&msg), sizeof(ResultAndHBReplyMsg));
    return result;
}

TResult Processor::LaunchTcpStoreClient()
{
    TTP_ASSERT_RETURN(worldSize_ > 0 && worldSize_ <= TTP_MAX_WORLD_SIZE, TTP_ERROR);
    std::string url;
    TResult ret = GetTcpStoreUrl(url);
    if (ret != TTP_OK) {
        TTP_LOG_ERROR("rank:" << rank_ << " launch tcp store client failed, url:" << url);
        return ret;
    }
    LaunchTcpStoreMsg msg;
    msg.rank = rank_;
    msg.worldSize = worldSize_;
    msg.url = url;

    TTP_LOG_INFO("rank:" << rank_ << " begin to launch tcp_store_client, url:" << msg.url <<
        ", world_size:" << worldSize_);
    return EventProcess(PROCESSOR_EVENT_LAUNCH_TCP_STORE_CLIENT, &msg, sizeof(LaunchTcpStoreMsg));
}

TResult Processor::GetTcpStoreUrl(std::string &url)
{
    TTP_ASSERT_RETURN(controllerIdx_ >= 0 && controllerIdx_ < static_cast<int32_t>(controllerIps_.size()), TTP_ERROR);

    std::string ipPort = controllerIps_.at(controllerIdx_);
    std::size_t index = 0;
    if (controllerIdx_ > 0) {
        index = ipPort.find(':');
        if (index == std::string::npos) {
            TTP_LOG_ERROR("rank:" << rank_ << " receive invalid controller ip:" << ipPort);
            return TTP_ERROR;
        }
        index += 1;
    }
    std::string ip = ipPort.substr(index);

    // Get tcp store server port
    uint32_t port = GetEnvValue2Uint32("MASTER_PORT", PortInfo.minVal, PortInfo.maxVal, PortInfo.defaultVal);
    TTP_LOG_INFO("[env] MASTER_PORT:" << port);

    url = "tcp://" + ip + ":" + std::to_string(port);
    return TTP_OK;
}

void Processor::LaunchTcpStoreServerThread(LaunchTcpStoreMsg *serverInfo)
{
    EventProcess(PROCESSOR_EVENT_LAUNCH_TCP_STORE_SERVER, serverInfo, sizeof(LaunchTcpStoreMsg));
}

std::string Processor::GetRepairType()
{
    static std::unordered_map<uint32_t, std::string> type2str = {
        {ControllerRepairType::CRT_DUMP,  "dump"},
        {ControllerRepairType::CRT_RETRY,   "retry"},
        {ControllerRepairType::CRT_ARF,   "recover"},
    };

    auto itr = type2str.find(repairType_);
    return itr != type2str.end() ? itr->second : "unknow";
}

}
}