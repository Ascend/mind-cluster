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
#ifndef OCK_TTP_PROCESSOR_H
#define OCK_TTP_PROCESSOR_H

#include <cstdint>
#include <mutex>
#include <string>
#include <vector>
#include <thread>
#include "common.h"
#include "controller.h"
#include "acc_tcp_client.h"

using namespace ock::acc;
namespace ock {
namespace ttp {

enum ProcessorEvent : uint32_t {
    PROCESSOR_EVENT_SAVE_CKPT,
    PROCESSOR_EVENT_RENAME,
    PROCESSOR_EVENT_DEVICE_STOP,
    PROCESSOR_EVENT_DEVICE_CLEAN,
    PROCESSOR_EVENT_REPAIR,
    PROCESSOR_EVENT_ROLLBACK,
    PROCESSOR_EVENT_PAUSE,
    PROCESSOR_EVENT_CONTINUE,
    PROCESSOR_EVENT_PT_COMM_OPERATE,
    PROCESSOR_EVENT_UPGRADE_REBUILD,
    PROCESSOR_EVENT_UPGRADE_REPAIR,
    PROCESSOR_EVENT_UPGRADE_ROLLBACK,
    PROCESSOR_EVENT_DOWNGRADE_REBUILD,
    PROCESSOR_EVENT_LAUNCH_TCP_STORE_CLIENT,
    PROCESSOR_EVENT_LAUNCH_TCP_STORE_SERVER,
    PROCESSOR_EVENT_EXIT,
    PROCESSOR_EVENT_BUTT,
};

struct SaveCkptContext {
    int64_t step;
    int32_t repairId;
    std::vector<int32_t> groupIdx;
    std::vector<std::vector<int32_t>> ranks;
};

enum class ReportState : int32_t {
    RS_NORMAL = 0,
    RS_UCE,
    RS_UCE_CORRUPTED,
    RS_HCCL_FAILED,
    RS_INIT_FINISH,
    RS_PREREPAIR_FINISH,
    RS_STEP_FINISH,
    RS_UNKNOWN,
};

struct RepairContext {
    RepairType type;
    std::vector<int32_t> srcRank;
    std::vector<int32_t> dstRank;
    std::vector<int32_t> replicaIdx;
    std::vector<int32_t> groupIdx;
    int32_t repairId;
    int64_t step;
    std::vector<int32_t> ranks;
    std::string zitParam;
};

struct PauseContext {
    int64_t step;
    bool hotSwitch;
};

struct ZitRebuildContext {
    int32_t repairId;
    std::vector<int32_t> commGroupIdx;
    std::vector<std::vector<int32_t>> commGroups;
    std::string zitParam;
};

struct LaunchTcpStoreMsg {
    int32_t rank;
    int32_t worldSize;
    std::string url;
};

// 使用了比较运算,注意顺序
enum ProcessorStatus : uint32_t {
    PS_BEGIN,
    PS_INITING,
    PS_INITED,
    PS_START,
    PS_RESTART,
    PS_END,
    PS_NORMAL, // 异常状态在normal后面
    PS_PAUSE,
    PS_DUMP
};

class Processor;
using ProcessorPtr = Ref<Processor>;
using ProcessorEventHandle = std::function<int32_t(void *ctx, int ctxSize)>;
using ProcessorCmdHandle = std::function<int32_t(void *ctx, int ctxSize)>;
using ProcessorReplyHandle = std::function<int32_t(uint16_t sn, MsgOpCode op)>;
using LogFunc = void (*)(int level, const char *msg);

class Processor : public Referable {
public:
    static ProcessorPtr GetInstance(bool destroy = false);

    // set os status to updating, framework begin to updating the optimizer state data
    TResult BeginUpdating(int64_t backupStep);

    // set os status to updating, framework begin to copying the optimizer state data to local copy
    TResult BeginCopying();

    // set os status finished,  framework already end to update the optimizer state data
    TResult FinishedUpdate(int64_t step);

    TResult ResetLimitStep();

    // set dump result
    TResult SetDumpResult(int32_t result);

    // init
    TResult Initialize(int32_t rank, int32_t worldSize, bool enableLocalCopy, const AccTlsOption &tlsOption,
                       bool enableUce = true, bool enableArf = false, bool enableZit = false);

    // init processor
    TResult Start(std::string &masterIp, int32_t port, std::string localIp = "");  // rank_:ip

    // register event handler
    TResult RegisterEventHandler(ProcessorEvent event, ProcessorEventHandle handle);

    TResult ReportReplicaInfo(std::vector<std::vector<int32_t>> &groups,
        std::vector<int32_t> replicaCnt, std::vector<int32_t> replicaShift);

    TResult ReportDpInfo(std::vector<int32_t> &dpRankList);

    TResult ReportLoadCkptStep(int64_t step);

    TResult WaitNextAction();

    TResult WaitRepairAction();

    TResult ReportStatus(ReportState state);

    int32_t GetRepairId() const
    {
        return repairId_;
    }

    bool GetHotSwitch() const
    {
        return hotSwitch_;
    }

    std::string GetRepairType();

    // destory processor
    void Destroy(bool inInner = false);

    void ReportBeforeDestroy();

    Processor();

protected:
    bool ProcessorIsRunning();

    TResult HeartbeatSend();

    // heartbeat loop
    void HeartbeatThread();

    TResult Connect2Controller();

    TResult Register2Controller();

    TResult ReportInfo2Controller(std::vector<int32_t> replicaCnt, std::vector<int32_t> replicaShift);

    TResult IsMaxRetryTime(uint32_t timeRetried);

    TResult ReportDp2Controller();

    void HandleBackupCtrlList(uint8_t *data);

    void HandleDumpCkpt(uint8_t *data, uint32_t len);

    TResult DumpCkpt(uint8_t *data);

    void HandleRename(uint8_t *data, uint32_t len);

    TResult Rename(uint8_t *data);

    TResult ReStart();

    void StartBackupController(uint8_t *data);

    void RequestHandleRegister();

    TResult InitTlsOption(const AccTlsOption &tlsOption);

    void ExtraReplyHandleRegister();

    void HandleExit(uint8_t *data, uint32_t len);

    void HandleBroadcast(uint8_t *data, uint32_t len);

    void HandleDeviceStop(uint8_t *data, uint32_t len);

    void HandleDeviceClean(uint8_t *data, uint32_t len);

    void HandleRepair(uint8_t *data, uint32_t len);

    void HandleNotifyNormal(uint8_t *data, uint32_t len);

    void HandleControllerReply(uint8_t *data, uint32_t len);

    void HandleCollection(uint8_t *data, uint32_t len);

    void HandlePrelock(uint8_t *data, uint32_t len);

    void HandleRollback(uint8_t *data, uint32_t len);

    void HandlePause(uint8_t *data, uint32_t len);

    void HandleContinue(uint8_t *data, uint32_t len);

    void HandleDowngradeRebuild(uint8_t *data, uint32_t len);

    void HandleUpgradeRebuild(uint8_t *data, uint32_t len);

    void HandleUpgradeRollback(uint8_t *data, uint32_t len);

    void HandleUpgradeRepair(uint8_t *data, uint32_t len);

    void HandlePtComm(uint8_t *data, uint32_t len);

    void HandleLaunchTcpStoreServer(uint8_t *data, uint32_t len);

    void HandleDestroyNofity(uint8_t *data, uint32_t len);

    TResult EventProcess(ProcessorEvent eventCode, void *ctx, int ctxSize);

    TResult CheckMsgSnAndReply(uint16_t sn, MsgOpCode replyOp);

    TResult UpdateSerialNumber(uint16_t sn, MsgOpCode replyOp = TTP_MSG_OP_BUTT);

    TResult GetBackupStatus(uint16_t sn, TTPReplyMsg& msg);

    void SaveBackupStatus(const TTPReplyMsg& msg);

    int32_t ProcessResultAndHBReply(uint16_t sn, MsgOpCode op);

    // waiting for framework enter repair-wait state
    TResult RepairBarrierWithFramework();

    // must be used in statusMutex_
    std::pair<int64_t, int64_t> GetNowStep();

    void WaitForLimitStepRelease(int64_t nowStep);

    bool CheckOptimStateOK(int64_t step);

private:
    inline std::string IpPort() const
    {
        if ((controllerIdx_ < 0) || (static_cast<int32_t>(controllerIps_.size()) <= controllerIdx_)) {
            return "(empty ip)";
        }

        return controllerIps_[controllerIdx_] + ":" + std::to_string(port_);
    }

    void ReplyPrelock(ResultAndHBReplyMsg &reply, uint32_t nowStatus);

    TResult LaunchTcpStoreClient();

    TResult GetTcpStoreUrl(std::string &url);

    void LaunchTcpStoreServerThread(LaunchTcpStoreMsg *serverInfo);

    TResult StopAndCleanBeforeDump();
    TResult DeSerializedUpgradeRepairMsg(RepairMsg *msg, TTPReplyMsg &replyMsg, RepairContext &rc);

private:
    int32_t rank_ = -1;    // current process rank_ number
    int32_t worldSize_ = -1;
    int32_t groupTypeNum_ = 0;
    bool localCopySwitch_ = false;
    bool arfSwitch_ = false;
    bool uceSwitch_ = false;
    bool zitSwitch_ = false;
    bool mindSpore_ = false;
    int32_t startBackup_ = -1;
    std::vector<std::vector<int32_t>> rankList_;  // rank_ number list belong to the same group
    std::vector<int32_t> replicaCnt_;
    std::vector<int32_t> replicaShift_;
    std::vector<int32_t> dpList_;    // dp Group where current rank belongs to; only used in dp zit

    AccTcpClientPtr newClient_ = nullptr;

    std::vector<std::string> controllerIps_;  // controller list
    int32_t controllerIdx_ = 0;                   // controller idx
    int32_t port_ = 0;                           // controller port

    TrainStatus trainStatus_{};
    std::atomic<int64_t> limitStep_ = {INT64_MAX};
    int64_t lockStep_ = -1;
    std::atomic<bool> isPrelockOk_{false};  // prelock结果
    bool hotSwitch_ = false;
    int32_t repairId_ = -1;
    int64_t repairStep_ = -1;
    std::mutex statusMutex_;
    std::mutex initOrDestroyMutex_;
    ProcessorEventHandle eventHandleList_[PROCESSOR_EVENT_BUTT]{};
    ProcessorReplyHandle extraReplyHandleList_[TTP_MSG_OP_BUTT]{};

    // bool dump status -1 is init value
    int32_t dumpRet_ = -1;
    PthreadTimedwait dumpCond_;

    // keep sync with framework
    PthreadTimedwait repairWaitCond_;

    std::thread thread_;
    std::atomic<bool> isStarted_{false}; // 心跳线程是否启动
    std::atomic<bool> isStopped_{true};   // 通知心跳线程结束

    std::atomic<uint32_t> processorStatus_ = {PS_BEGIN};  // processor状态
    bool envClearFlag_ = false;
    sem_t waitSem_;      // wait next action used
    int waitResult_ = 0;

    std::atomic<uint16_t> actionSn_{0};    // Lastest received action Sn from Controller
    TTPReplyMsg replyMsgBackup_;    // heartbeat status will not save for ExtraReplyHandler
    AccTlsOption tlsOption_ {};
    PthreadTimedwait replySem_;
    TResult replyRet_ = TTP_OK;
    std::atomic<bool> readyToExit_{false};
    uint32_t repairType_ = ControllerRepairType::CRT_BUTT;
};

}  // namespace ttp
}  // namespace ock
#endif