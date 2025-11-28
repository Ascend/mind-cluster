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
#ifndef OCK_TTP_ACTION_ENGINE_H
#define OCK_TTP_ACTION_ENGINE_H

#include "acc_tcp_server.h"
#include "acc_tcp_worker.h"
#include "action_counter.h"

using namespace ock::acc;
namespace ock {
namespace ttp {

enum ActionOp : uint32_t {
    ACTION_OP_BROADCAST_IP = 0,
    ACTION_OP_COLLECTION,
    ACTION_OP_PRELOCK,
    ACTION_OP_PAUSE,
    ACTION_OP_CONTINUE,
    ACTION_OP_DEVICE_STOP,
    ACTION_OP_DEVICE_CLEAN,
    ACTION_OP_REPAIR,
    ACTION_OP_ROLLBACK,
    ACTION_OP_NOTIFY_NORMAL,
    ACTION_OP_SAVECKPT,
    ACTION_OP_RENAME,
    ACTION_OP_EXIT,
    ACTION_OP_DESTROY_NOTIFY,
    ACTION_OP_DOWNGRADE_REBUILD,
    ACTION_OP_UPGRADE_REBUILD,
    ACTION_OP_PT_COMM,
    ACTION_OP_UPGRADE_REPAIR,
    ACTION_OP_UPGRADE_ROLLBACK,
    ACTION_OP_LAUNCH_STORE_SERVER,
    ACTION_OP_BUTT,
};

inline const char *StrActionOp(uint32_t op)
{
    static std::vector<std::string_view> actionOps = {
        "ACTION_OP_BROADCAST_IP",
        "ACTION_OP_COLLECTION",
        "ACTION_OP_PRELOCK",
        "ACTION_OP_PAUSE",
        "ACTION_OP_CONTINUE",
        "ACTION_OP_DEVICE_STOP",
        "ACTION_OP_DEVICE_CLEAN",
        "ACTION_OP_REPAIR",
        "ACTION_OP_ROLLBACK",
        "ACTION_OP_NOTIFY_NORMAL",
        "ACTION_OP_SAVECKPT",
        "ACTION_OP_RENAME",
        "ACTION_OP_EXIT",
        "ACTION_OP_DESTROY_NOTIFY",
        "ACTION_OP_DOWNGRADE_REBUILD",
        "ACTION_OP_UPGRADE_REBUILD",
        "ACTION_OP_PT_COMM",
        "ACTION_OP_UPGRADE_REPAIR",
        "ACTION_OP_UPGRADE_ROLLBACK",
        "ACTION_OP_LAUNCH_STORE_SERVER",
        "ACTION_OP_BUTT",
    };

    if (op >= actionOps.size()) {
        return "";
    }

    return actionOps[op].data();
}

struct ActionInfo {
    int16_t msgType;
    AccDataBufferPtr d;
    std::vector<int32_t> targetRanks;
};

using ActionMsgSend = std::function<TResult(int16_t msgType,
    const AccDataBufferPtr &d, std::vector<int32_t> &targetRanks, const std::vector<AccDataBufferPtr> &cbCtx)>;
using AccExtraNewReqHandler = std::function<TResult(const AccTcpRequestContext &context)>;
using AccExtraNewReqReplyParseHandler = std::function<TResult(const AccTcpRequestContext &context, TTPReplyMsg &msg)>;
using MarkRankStatusMethod = std::function<TResult(const std::vector<int32_t> &ranks)>;

class ActionEngine : public Referable {
public:
    ActionEngine() = default;
    ~ActionEngine() override = default;

    TResult Initialize(AccTcpServerPtr mServer, const ActionMsgSend &sendFunc, int32_t worldSize);

    void ReplyRegister(int16_t msgType, const AccExtraNewReqHandler &h);

    void ReplyParseRegister(int16_t msgType, const AccExtraNewReqReplyParseHandler &h);

    void RankStatusRegister(const MarkRankStatusMethod &h);

    TResult Process(ActionOp opcode, std::vector<ActionInfo> &info, bool waitReply, uint16_t sn,
                    uint32_t retryTimes = 0);    // if not need retry, retryTime = 0

private:
    void TcpHandleRegister(AccTcpServerPtr mServer);

    TResult ProcessInner(ActionOp opcode, std::vector<ActionInfo> &info, bool waitReply, uint32_t retryTimes);

    TResult ActionSend(std::vector<ActionInfo> &info, uint16_t sn);

    TResult MsgCallBack(AccMsgSentResult result, const AccMsgHeader &header, const AccDataBufferPtr &cbCtx);

    TResult HandleReplyCommon(const AccTcpRequestContext &context);

    TResult FindUnackInfo(const std::vector<ActionInfo> &info, AtomicStatusPtr statusPtr, uint16_t sn,
                          std::vector<ActionInfo> &unackInfo, bool isErrorRetry = false);

    TResult FindNoResponseRanks(AtomicStatusPtr statusPtr, uint16_t sn, const std::vector<int32_t> &ranks,
                                std::vector<int32_t> &ranksNoResponse);

    TResult SendInfo(std::vector<ActionInfo> &info);

    void InitReplyStatus(const std::vector<ActionInfo> &info, std::vector<int32_t> &reSendRanks,
                         std::vector<ActionInfo> &reSendInfo);

    void InitCbStatus(const std::vector<ActionInfo> &info, std::vector<int32_t> &reSendRanks,
                      std::vector<ActionInfo> &reSendInfo);

    TResult InitSendStatus(int32_t worldSize);

    ActionMsgSend actionSend_;
    std::atomic<bool> inited_{false};
    std::atomic<uint32_t> inAction_{0}; // 同一时刻仅支持发起一个动作

    // callback相关
    AtomicCounter sendCount_;
    AtomicCounter cbCount_;
    AtomicStatusPtr cbStatus_;
    PthreadTimedwait cbTimedwait_;

    // reply相关
    AccExtraNewReqHandler extraReplyHandle_[UNO_48]{};
    AccExtraNewReqReplyParseHandler extraReplyParseHandler_[UNO_48]{};
    AtomicCounter expectCount_;
    AtomicCounter realCount_;
    AtomicStatusPtr replyStatus_;
    PthreadTimedwait replyTimedwait_;

    // sn related Lock
    ReadWriteLock cbStatusLock_;
    ReadWriteLock replyStatusLock_;

    // handler: mark no-response ranks as abnormal
    MarkRankStatusMethod markStatusHandle_;

    // reply error retry interval, read from env, in ms
    std::unordered_map<uint32_t, uint32_t> replyRetryIntervalMilliSec_;
};
using ActionEnginePtr = Ref<ActionEngine>;

}  // namespace ttp
}  // namespace ock
#endif
