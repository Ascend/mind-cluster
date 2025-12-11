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
#ifndef OCK_TTP_CONTROLLER_H
#define OCK_TTP_CONTROLLER_H

#include <map>
#include <mutex>
#include <set>
#include <string>
#include <thread>
#include <unordered_map>
#include <vector>
#include "acc_tcp_server.h"
#include "action_engine.h"
#include "common.h"
#include "controller_state_machine.h"
#include "mindx_engine.h"

using namespace ock::acc;

namespace ock {
namespace ttp {

using GroupMap = std::set<std::vector<int32_t>>;
using RankMask = std::vector<std::pair<int32_t, uint8_t>>;

struct GroupInfo {
    uint32_t repCnt {};
    GroupMap dpGroups;
};

struct BackupInfo {
    int32_t rank;
    std::string ip;
    std::string port;

    // mock used
    bool operator == (const BackupInfo &b) const
    {
        return (this->rank == b.rank && this->ip == b.ip && this->port == b.port);
    }
};

struct RepairInfo {
    int32_t msgRank;
    int32_t srcRank;
    int32_t dstRank;
    int16_t groupType;
    int16_t replicaIdx;
    RepairType type;
};

struct RankChooseInfo {
    int64_t step;
    std::set<int32_t> errorRanks;
    std::vector<int32_t> rankVec;
};

struct NotifyRankInfo {
    std::map<int32_t, int32_t> rankList;
    uint32_t hcclTime;
};

struct ZitParm {
    std::set<int32_t> isolateRanks;
    std::string strategyParm;
};

using StateMachinePtr = Ref<controllerStateMachine>;

class Controller : public Referable {
public:
    using Ptr = Ref<Controller>;

    static Ptr GetInstance(bool destroy = false);

    // init
    TResult Initialize(int32_t rank, int32_t worldSize, bool enableLocalCopy = false,
                        bool enableARF = false, bool enableZIT = false);

    // init controller
    TResult Start(std::string &masterIp, int32_t port, const AccTlsOption &tlsOpts, uint32_t controllerIdx = 0);

    std::vector<int32_t> GetAllLinkRanks(bool excludeError = false);

    TResult SendMsg(int16_t msgType, const AccDataBufferPtr &d, std::vector<int32_t> &targetRanks,
                    const std::vector<AccDataBufferPtr> &cbCtx);

    TResult Destroy(bool isInner = false);

    TResult ExitLogsHandler();

    Controller();

    // select back up controller
    std::vector<BackupInfo> SelectBackUpController();

    bool GetHighAvailabilitySwitch();

private:
    void ExitNotify();

    TResult MindXNotifyStopTrain(void *ctx, int32_t ctxSize);

    TResult MindXNotifyFaultRanks(void *ctx, int32_t ctxSize);

    TResult MindXNotifyDump(void *ctx, int32_t ctxSize);

    TResult MindXNotifyElegantDump(void *ctx, int32_t ctxSize);

    TResult MindXNotifyPauseTrain(void *ctx, int32_t ctxSize);

    TResult MindXNotifyContinueTrain(void *ctx, int32_t ctxSize);

    TResult MindXNotifyHotSwitch(void *ctx, int32_t ctxSize);

    TResult MindXNotifyStopSwitch(void *ctx, int32_t ctxSize);

    TResult MindXNotifyMigration(void *ctx, int32_t ctxSize);

    TResult MindXNotifyArfRepair(void *ctx, int32_t ctxSize);

    TResult MindXNotifyUceRepair(void *ctx, int32_t ctxSize);

    TResult MindXNotifyDownGradeRepair(void *ctx, int32_t ctxSize);

    TResult MindXNotifyUpGradeRepair(void *ctx, int32_t ctxSize);

    TResult MindXNotifyExit(void *ctx, int32_t ctxSize);

    TResult MindXInvalidNotify(void *ctx, int32_t ctxSize);

    TResult ErrorRankMsgModify(std::map<int32_t, int32_t> &errorRankMap, std::string msg);

    TResult TcpServerInit();

    TResult ActionEngineInit();

    TResult StateMachineInit();

    TResult HandleHeartBeat(const AccTcpRequestContext &context);

    TResult BeginExceptionCkpt(const std::set<int32_t> &errorRanks, bool isTcpStoreOK);

    TResult Rename();

    TResult BroadcastMsgStuff(BroadcastIpMsg *broadcastIpMsg, std::string &ipList);

    TResult BroadcastCrtlIps();

    TResult HandleRegister(const AccTcpRequestContext &context);

    TResult HandleReportInfo(const AccTcpRequestContext &context);

    TResult HandleReportDp(const AccTcpRequestContext &context);

    TResult RegisterStatus(RegisterMsg *registerMsg);

    std::string BuildStr4BackupCtrl();

    void SelectErrorRanks();

    TResult CheckTrainStatus();

    bool IsBackupToMaster();

    TResult HandleNewConnection(const AccConnReq &req, const AccTcpLinkComplexPtr &link);

    TResult HandleLinkBroken(const AccTcpLinkComplexPtr &link);

    TResult ChooseRank(const RankChooseInfo &rankChooseInfo, std::vector<int32_t> &tmpRankVec, uint32_t repCnt);

    RankMask GenerateRankMask(const RankChooseInfo &rankChooseInfo);

    TResult ProcessRepairFlow(bool isPreLocked);

    TResult MindXConfirmStrategy(bool isPreLocked = true);

    TResult HandleRecoverStrategy();

    TResult MindXReConfirmStrategy();

    TResult MindXInnerInteraction();

    TResult PauseTrain();

    TResult ContinueTrain();

    void ReportMindXRepairResult(int64_t &step);

    // state machine callback
    TResult InitCallback();

    TResult NormalCallback();

    TResult PauseCallback();

    TResult StepFinishCallback();

    TResult MigrationCallback();

    TResult AbnormalCallback();

    TResult DumpCallback();

    TResult ExitCallback();

    TResult EnvClearCallback();

    TResult RepairCallback();

    TResult DowngradeRepairCallback();

    TResult DowngradeRunning();

    TResult UpgradeRunning();

    void SwapHotSwitchRankInfo();

    bool CheckHotSwitchRegister();

    bool CheckCanRepair(bool isZit = false);

    bool CheckDpGroup();

    void InitializeVariables();

    TResult PauseWait();

    ControllerRepairType ConfirmRepairType();

    std::pair<bool, bool> ConfirmRepairFlag();

    TResult HandleDumpStatus(const std::set<int32_t> &errorRanks);

    TResult Prelock();

    TResult DoPause();

    TResult UpdateStatus(HeartBeatMsg *originHeartBeatMsg);

    TResult HandleCollectionReply(const AccTcpRequestContext &context);

    TResult HandlePrelockReply(const AccTcpRequestContext &context);

    TResult HandleNotifyNormalReply(const AccTcpRequestContext &context);

    TResult HandleCleanReply(const AccTcpRequestContext &context);

    TResult ResultAndHbReplyParse(const AccTcpRequestContext &context, TTPReplyMsg &msg);

    TResult PrelockResultAndHbReplyParse(const AccTcpRequestContext &context, TTPReplyMsg &msg);

    TResult UCERepair();

    TResult UpGradeRepair();

    TResult ARFRepair();

    TResult WaitDumpOrExitStrategy();

    TResult ARFWait(bool isFirst);

    TResult RepairProcess(const std::vector<RepairInfo> &rInfo);

    std::pair<int64_t, std::vector<int32_t>> SelectLockStep();

    RankMask RepairCheckStatus(const std::vector<int32_t> &rankVec);

    void GetAllRepairInfo(RankMask &rankMask, std::vector<RepairInfo> &rInfo, int16_t groupIdx);

    TResult RepairSelectReplica(RankMask &rankMask, std::vector<RepairInfo> &rInfo,
                                uint32_t repCnt, int16_t groupIdx);

    TResult GenerateRepairMsg(const std::vector<RepairInfo> &rInfo, std::vector<ActionInfo> &info);

    TResult PrepareRepairMsg(std::vector<RepairInfo> &rInfo, const std::set<int32_t> &errRanks, bool isZit = false);

    TResult InitDpGroupMap(const ReplicaMsg *replicaMsg);

    void ChooseIsolateRanks();

    TResult NotifyIsolateRanks(std::set<int32_t> &isolateRanks, std::set<int32_t> &errRanks);

    TResult DowngradeNotifyNormalRanks();

    TResult GenerateDownGradeMsgInner(std::vector<ActionInfo> &info, const std::vector<int32_t> &msgRanks,
                                      const std::vector<std::vector<int32_t>> &dpGroups,
                                      const std::vector<int32_t> &normalRanks, uint16_t sn);

    TResult GenerateDownGradeMsg(std::vector<ActionInfo> &info, std::vector<int32_t> &normalRanks, uint16_t sn);

    TResult IsolateRanksSetStatus(std::set<int32_t> &isolateRanks, std::set<int32_t> &errRanks);

    void RecordRankIp(int32_t rankIn);

    TResult HcclCommGroupRepair(const std::set<int32_t> &isolateRanks);  // repair DP group and global group hccl

    TResult UpGradeCommGroupRepair(const std::set<int32_t> &isolateRanks);

    TResult NotifyRankRollback(const std::vector<int32_t> &targetRanks, RepairType type);

    TResult UpgradePreCheck(const std::set<int32_t> &isolateRanks);

    TResult IsActionSnOK();

    TResult MarkNoReponseRanks(const std::vector<int32_t> &noResponseRanks);

    TResult DoNoDataAction(const std::vector<int32_t> &ranks, MsgOpCode mOp, ActionOp aOp,
                           bool reply, bool sendRepairType = false);

    bool CheckTcpStoreServerAvailable();

    bool CheckIpPortAccessible(const std::string &ip, uint16_t port);

    TResult GetTcpStoreUrl(std::set<std::string> &ipList, uint16_t &port);

    TResult TransforHostNameToIp(const char *hostName, std::string &ip);

    TResult HandleTcpStoreError(bool &tcpStoreOK);

    inline bool IsMsgLenValid(uint32_t len);

    std::set<int32_t> GetErrorRanks();

    std::set<int32_t> GetIsolateRanks();

    TResult ZitHandleNewFault();

    TResult ZitHandleStrategy();

    // cluster world size
    int32_t worldSize_ = -1;
    int32_t rank_ = -1;
    // backup info reportred rank count
    std::atomic<int32_t> reportedCnt_ { 0 };
    uint32_t controllerIdx_ = 0; // master_controller is 0

    AccTcpServerPtr mServer_ = nullptr;
    ActionEnginePtr engine_ = nullptr;
    std::atomic<uint16_t> actionSn_{1};
    int32_t rankToRename_ = -1;

    // buckup ip list
    std::vector<BackupInfo> backupInfoList_;
    std::string controllerIp_ = "";
    int32_t controllerPort_ = 0;

    // rank_ list string  -> dp group rank_ list
    ReadWriteLock dpGroupMapLock_;
    std::vector<GroupInfo> dpGroupListMap_;
    GroupMap originDpSet_;
    ReadWriteLock originDpLock_;
    // rank_ -> heart beat status
    ReadWriteLock statusMapLock_;
    std::unordered_map<int32_t, TrainStatus> statusMap_;
    std::unordered_map<int32_t, TrainStatus> statusMapTmp_;

    // rank_ -> error type
    ReadWriteLock errorRankLock_;
    std::map<int32_t, int32_t> errorRankMsg_;
    std::set<int32_t> hotSwitchRanks_;

    ZitParm zitParam_;
    // node ip  -> node rank list
    ReadWriteLock ipMapLock_;
    std::unordered_map<int32_t, std::string> rankIpMap_;

    // rank_ -> link
    ReadWriteLock linkMapLock_;
    std::unordered_map<int32_t, AccTcpLinkComplexPtr> rankLinkMap_;
    std::unordered_map<uint32_t, int32_t> linkIdMap_;
    std::unordered_map<int32_t, AccTcpLinkComplexPtr> rankLinkMapTmp_;
    std::unordered_map<uint32_t, int32_t> linkIdMapTmp_;
    // init lock
    std::mutex initOrDestroyMutex_;

    std::atomic<bool> isInited_{false};
    std::atomic<bool> isStarted_{false};
    std::atomic<bool> isStopped_{false};
    std::atomic<bool> isMasterCtrl_{false};
    std::atomic<bool> isAlreadyBrod_{false};
    std::atomic<bool> isBackupToMaster_{false};
    std::atomic<bool> isSupportBackupToMaster_{true};
    std::atomic<bool> isNeedToReportResult_{false};

    // State Machine
    PthreadTimedwait pthreadTimeChecker_ ;
    StateMachinePtr stateMachine_ = nullptr;

    // Mindx engine ptr
    MindXEnginePtr mindXEngine_ = nullptr;

    // repair
    int64_t repairStep_ = -1;
    std::atomic<int64_t> loadCkptRepairStep_{-1};
    std::atomic<int32_t> repairId_{0};
    std::atomic<int32_t> prelockRet_{0};
    std::atomic<bool> canRetryCleanFlag_{false};

    MindXEvent repairEvent_ = MindXEvent::MINDX_EVENT_BUTT;
    uint32_t waitMindxTimes_ = {0};
    uint32_t waitPauseTimes_ = {0};
    uint32_t waitHcclTime_ = {30};
    uint32_t repairType_ = ControllerRepairType::CRT_BUTT;
    bool unableRepair_ = false;
    uint32_t hcclFlag_ = {0};

    // feature switch
    bool localCopySwitch_ = false;
    bool arfSwitch_ = false; // air refueling
    bool zitSwitch_ = false; // zero interruption
    bool uceSwitch_ = false;
    // enable pytorch or mindSpore
    bool mindSpore_ = false;
    bool isPorcessorExit_ = false;
};

using ControllerPtr = Controller::Ptr;

}  // namespace ttp
}  // namespace ock

#endif
