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
#include <cstdlib>
#include <string_view>
#include "action_engine.h"

namespace ock {
namespace ttp {

constexpr int16_t MIN_MSG_TYPE = 0;
constexpr int16_t MAX_MSG_TYPE = UNO_48;
constexpr uint32_t TTP_SENT_ERROR_RETRY_INTERVAL = 2000;
constexpr uint32_t TTP_SENT_ERROR_RETRY_TIMES = 2;
constexpr uint32_t TTP_CB_WAIT_RETRY_INTERVAL = 2000;
constexpr uint32_t TTP_CB_WAIT_RETRY_TIMES = 3;
constexpr uint32_t TTP_COLLECTION_TIME_LIMIT = 5;

constexpr EnvVarValue NormalActionTime = {.minVal = 30, .maxVal = 1800, .defaultVal = 180};

TResult ActionEngine::Process(ActionOp opcode, std::vector<ActionInfo> &info, bool waitReply, uint16_t sn,
                              uint32_t retryTimes)
{
    if (opcode >= ACTION_OP_BUTT) {
        TTP_LOG_ERROR("input opcode error! opcode: " << StrActionOp(opcode));
        return TTP_ERROR;
    }

    if (!inited_.load()) {
        TTP_LOG_ERROR("action engine not initialize! opcode: " << StrActionOp(opcode));
        return TTP_ERROR;
    }

    if (inAction_.fetch_add(1) > 0) {
        inAction_.fetch_sub(1);
        TTP_LOG_ERROR("performing an action now! opcode: " << StrActionOp(opcode));
        return TTP_ERROR;
    }

    std::vector<int32_t> allTargetRanks;
    uint32_t rankNum = 0;
    for (uint32_t i = 0; i < info.size(); i++) {
        rankNum += info[i].targetRanks.size();
        for (auto rk : info[i].targetRanks) {
            allTargetRanks.push_back(rk);
        }
    }

    replyStatusLock_.LockWrite();
    for (uint32_t i = 0; i < info.size(); i++) {
        replyStatus_->InitRankStatus(info.at(i).targetRanks, sn);
    }
    expectCount_.Store(sn, rankNum);
    realCount_.Store(sn, 0);
    replyStatusLock_.UnLock();

    TResult ret = ProcessInner(opcode, info, waitReply, retryTimes);
    if (ret == TTP_OK) {
        TTP_LOG_INFO("do action success. opcode: " << StrActionOp(opcode));
    } else {
        // set time-out ranks' data_aval to ABNORMAL
        std::vector<int32_t> ranksNoResponse;
        if (waitReply) {    // use reply status to judge if no response
            FindNoResponseRanks(replyStatus_, sn, allTargetRanks, ranksNoResponse);
        } else {    // use cb status to judge if no response
            FindNoResponseRanks(cbStatus_, sn, allTargetRanks, ranksNoResponse);
        }
        markStatusHandle_(ranksNoResponse);
        TTP_LOG_INFO("do action failed. opcode: " << StrActionOp(opcode));
    }
    inAction_.fetch_sub(1);
    return ret;
}

TResult ActionEngine::ProcessInner(ActionOp opcode, std::vector<ActionInfo> &info, bool waitReply, uint32_t retryTimes)
{
    uint32_t infoSize = info.size();
    uint32_t rankNum = 0;
    TResult ret = TTP_OK;
    std::vector<int32_t> ranks;

    for (uint32_t i = 0; i < infoSize; i++) {
        rankNum += info[i].targetRanks.size();
        for (int32_t curRank : info[i].targetRanks) {
            ranks.push_back(curRank);
        }
    }

    if (rankNum == 0) {
        TTP_LOG_INFO("no rank need to send actionOp. opcode: " << StrActionOp(opcode));
        return TTP_OK;
    }

    if (!waitReply) {
        return ActionSend(info, expectCount_.GetSn());
    }

    int32_t replyReSendTime = 0;
    std::vector<int32_t> reSendRanks;
    const uint16_t sn = expectCount_.GetSn();
    std::vector<ActionInfo> reSendInfo = info;

    while (replyReSendTime <= static_cast<int32_t>(retryTimes)) {
        // Re-Send, need to get rank whose cb is not received
        ret = ActionSend(reSendInfo, sn);
        if (ret != TTP_OK) {
            TTP_LOG_ERROR("ActionSend failed, sn: " << sn);
            return ret;
        }

        int32_t timeRet = replyTimedwait_.PthreadTimedwaitSecs(replyRetryIntervalMilliSec_.at(opcode));
        TTP_LOG_DEBUG("wait action reply end. opcode: " << StrActionOp(opcode) << ", timedwaitRet: " << timeRet);
        if (timeRet != ETIMEDOUT) {
            // reply error not need retry,only timeout retry
            ret = replyStatus_->CheckRankGroupStatus(ranks, sn);
            if (ret != TTP_NEED_RETRY) {
                return ret;
            }
        }
        reSendRanks.clear();
        reSendInfo.clear();
        InitReplyStatus(info, reSendRanks, reSendInfo);
        replyReSendTime++;
    }

    TTP_LOG_ERROR("ProcessInner failed after reply error retry. opCode: " << StrActionOp(opcode) <<
        ", sn: " << expectCount_.GetSn() << ", not reply ranks: " << IntVec2String(reSendRanks) <<
        ". Maybe environment variable TTP_NORMAL_ACTION_TIME_LIMIT should be set a larger value.");
    return TTP_ERROR;
}

void ActionEngine::InitReplyStatus(const std::vector<ActionInfo> &info, std::vector<int32_t> &reSendRanks,
                                   std::vector<ActionInfo> &reSendInfo)
{
    uint32_t count = 0;
    replyStatusLock_.LockWrite();
    uint16_t sn = expectCount_.GetSn();
    FindUnackInfo(info, replyStatus_, sn, reSendInfo);

    for (ActionInfo &actInfo : reSendInfo) {
        replyStatus_->InitRankStatus(actInfo.targetRanks, sn);
        count += actInfo.targetRanks.size();
        for (auto rk : actInfo.targetRanks) {
            reSendRanks.push_back(rk);
        }
    }
    TTP_LOG_DEBUG("ProcessInner: expectCount, sn: " << sn << ", count: " << count);
    expectCount_.Store(sn, count);
    realCount_.Store(sn, 0);
    replyStatusLock_.UnLock();
}

TResult ActionEngine::ActionSend(std::vector<ActionInfo> &info, uint16_t sn)
{
    std::vector<int32_t> ranks;
    for (ActionInfo &actInfo : info) {
        for (int32_t curRank : actInfo.targetRanks) {
            ranks.push_back(curRank);
        }
    }
    if (ranks.empty()) {
        return TTP_OK;
    }

    // clear cbStatus and cbCounter for reply error re-send
    cbStatusLock_.LockWrite();
    cbStatus_->InitRankStatus(ranks, sn);
    sendCount_.Store(sn, ranks.size());
    cbCount_.Store(sn, 0);
    cbStatusLock_.UnLock();

    int32_t cbRetryTime = 0;
    int32_t timedwaitRet = 0;
    std::vector<int32_t> reSendRanks;
    std::vector<ActionInfo> reSendInfo = info;
    while (cbRetryTime <= TTP_CB_WAIT_RETRY_TIMES) {
        // Re-Send, need to get rank whose cb is not received
        TResult reSendRet = SendInfo(reSendInfo);
        if (reSendRet != TTP_OK) {
            return reSendRet;
        }

        timedwaitRet = cbTimedwait_.PthreadTimedwaitSecs(TTP_CB_WAIT_RETRY_INTERVAL);
        if (timedwaitRet != ETIMEDOUT) {
            // send errro and no cb all need retry
            reSendRet = cbStatus_->CheckRankGroupStatus(ranks, sendCount_.GetSn());
            TTP_LOG_DEBUG("timedwaitRet != ETIMEDOUT " << ETIMEDOUT << " curSn: " << sendCount_.GetSn() <<
                ", checkResult: " << reSendRet);
            if (reSendRet == TTP_OK) {
                return TTP_OK;
            }
        }
        reSendRanks.clear();
        reSendInfo.clear();
        InitCbStatus(info, reSendRanks, reSendInfo);

        // Time out, or cbStatus not OK
        cbRetryTime++;
        TTP_LOG_INFO("cb outdated and retry: cbRetryTime: " << cbRetryTime
            << ", timedwaitRet: " << timedwaitRet << ", ETIMEDOUT=" << ETIMEDOUT);
    }
    TTP_LOG_ERROR("ActionSend failed after cb error retry. MsgType: " << info.front().msgType <<
        ", sn: " << sendCount_.GetSn() << ", cb not reply ranks: " << IntVec2String(reSendRanks));
    return TTP_ERROR;
}

void ActionEngine::InitCbStatus(const std::vector<ActionInfo> &info, std::vector<int32_t> &reSendRanks,
                                std::vector<ActionInfo> &reSendInfo)
{
    uint32_t count = 0;
    cbStatusLock_.LockWrite();
    uint16_t sn = sendCount_.GetSn();
    FindUnackInfo(info, cbStatus_, sn, reSendInfo, true);

    for (ActionInfo &actInfo : reSendInfo) {
        cbStatus_->InitRankStatus(actInfo.targetRanks, sn);
        count += actInfo.targetRanks.size();
        for (auto rk : actInfo.targetRanks) {
            reSendRanks.push_back(rk);
        }
    }
    sendCount_.Store(sn, count);
    cbCount_.Store(sn, 0);
    cbStatusLock_.UnLock();
}

void ActionEngine::ReplyRegister(int16_t msgType, const AccExtraNewReqHandler &h)
{
    TTP_ASSERT_RET_VOID(msgType >= MIN_MSG_TYPE && msgType < MAX_MSG_TYPE);
    TTP_ASSERT_RET_VOID(h != nullptr);
    TTP_ASSERT_RET_VOID(extraReplyHandle_[msgType] == nullptr);
    extraReplyHandle_[msgType] = h;
}

void ActionEngine::ReplyParseRegister(int16_t msgType, const AccExtraNewReqReplyParseHandler &h)
{
    TTP_ASSERT_RET_VOID(msgType >= MIN_MSG_TYPE && msgType < MAX_MSG_TYPE);
    TTP_ASSERT_RET_VOID(h != nullptr);
    TTP_ASSERT_RET_VOID(extraReplyParseHandler_[msgType] == nullptr);
    extraReplyParseHandler_[msgType] = h;
}

TResult ActionEngine::MsgCallBack(AccMsgSentResult result,
    const AccMsgHeader & /* header */, const AccDataBufferPtr &cbCtx)
{
    if (cbCtx->DataLen() != sizeof(CallbackCtx)) {
        TTP_LOG_ERROR("ActionEngine: MsgCallBack receive illegal-length cbCtx");
        return TTP_ERROR;
    }

    CallbackCtx *msg = static_cast<CallbackCtx *>(cbCtx->DataPtrVoid());
    TTP_ASSERT_RETURN(msg != nullptr, TTP_ERROR);
    bool counterAdd = false;
    if (cbStatus_->CheckSetRankStatus(msg->rank, msg->sn, result, counterAdd) != TTP_OK) {
        TTP_LOG_DEBUG("receive rank: "<< msg->rank << "outdated callback msg, received msg sn: " <<
            msg->sn << ", ActionEngine sn: " << cbCount_.GetSn() << ", drop callback msg.");
        return TTP_ERROR;
    }

    uint64_t addResult = 0;
    if (counterAdd) {
        if (cbCount_.CheckAdd(msg->sn, 1, addResult) != TTP_OK) {
            TTP_LOG_INFO("rank: " << msg->rank << "callback sn check failed, received msg sn: " <<
                msg->sn << ", ActionEngine sn: " << cbCount_.GetSn() << ", drop callback msg.");
            return TTP_ERROR;
        }
    }

    if (addResult == sendCount_.Load()) {
        cbStatusLock_.LockRead();
        if (sendCount_.GetSn() == msg->sn) {
            int32_t ret = cbTimedwait_.PthreadSignal();
            if (ret == 0) {
                TTP_LOG_DEBUG("callback cbCount OK, send cbTimedwait.PthreadSignal sn: " << msg->sn);
            } else {
                TTP_LOG_ERROR("cbTimedwait.PthreadSignal failed! sn:" << msg->sn << " ret:" << ret);
                cbStatusLock_.UnLock();
                return TTP_ERROR;
            }
        }
        cbStatusLock_.UnLock();
    }
    return TTP_OK;
}

TResult ActionEngine::Initialize(AccTcpServerPtr mServer, const ActionMsgSend &sendFunc, int32_t worldSize)
{
    TTP_ASSERT_RETURN(sendFunc != nullptr, TTP_ERROR);
    actionSend_ = sendFunc;

    uint32_t normalActionTime = GetEnvValue2Uint32("TTP_NORMAL_ACTION_TIME_LIMIT",
                                                   NormalActionTime.minVal,
                                                   NormalActionTime.maxVal,
                                                   NormalActionTime.defaultVal);
    TTP_LOG_INFO("[env] TTP_NORMAL_ACTION_TIME_LIMIT:" << normalActionTime);

    for (uint32_t op = ACTION_OP_BROADCAST_IP; op < ACTION_OP_BUTT; op++) {
        replyRetryIntervalMilliSec_[op] = normalActionTime;
    }

    replyRetryIntervalMilliSec_[ACTION_OP_COLLECTION] = TTP_COLLECTION_TIME_LIMIT;

    TcpHandleRegister(mServer);

    TResult ret = InitSendStatus(worldSize);
    if (ret != TTP_OK) {
        return TTP_ERROR;
    }

    inited_.store(true);
    return TTP_OK;
}

void ActionEngine::TcpHandleRegister(AccTcpServerPtr mServer)
{
    auto cbMethod = [this](AccMsgSentResult result, const AccMsgHeader &header, const AccDataBufferPtr &cbCtx) {
        return MsgCallBack(result, header, cbCtx);
    };
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_CKPT_SEND, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_CTRL_NOTIFY, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_RENAME, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_EXIT, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_DESTROY_NOTIFY, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_DEVICE_STOP, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_DEVICE_CLEAN, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_REPAIR, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_NOTIFY_NORMAL, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_COLLECTION, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_PRELOCK, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_PAUSE, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_CONTINUE, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_ROLLBACK, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_DOWNGRADE_REBUILD, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_UPGRADE_REBUILD, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_UPGRADE_REPAIR, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_UPGRADE_ROLLBACK, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_PT_COMM, cbMethod);
    mServer->RegisterRequestSentHandler(TTP_MSG_OP_LAUNCH_STORE_SERVER, cbMethod);

    auto replyMethod = [this](const AccTcpRequestContext &context) { return HandleReplyCommon(context); };
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_CKPT_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_RENAME_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_STOP_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_CLEAN_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_REPAIR_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_DOWNGRADE_REBUILD_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_NORMAL_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_COLLECTION_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_PRELOCK_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_PAUSE_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_CONTINUE_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_ROLLBACK_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_UPGRADE_REBUILD_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_UPGRADE_REPAIR_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_UPGRADE_ROLLBACK_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_PT_COMM_REPLY, replyMethod);
    mServer->RegisterNewRequestHandler(TTP_MSG_OP_LAUNCH_STORE_SERVER_REPLY, replyMethod);
}

TResult ActionEngine::InitSendStatus(int32_t worldSize)
{
    cbStatus_ = MakeRef<AtomicStatusVector>();
    if (cbStatus_ == nullptr) {
        TTP_LOG_ERROR("ActionEngine: create cbStatus failed");
        return TTP_ERROR;
    }
    if (cbStatus_->Initialize(worldSize, MSG_BUTT, MSG_SENT) == TTP_ERROR) {
        TTP_LOG_ERROR("ActionEngine: cbStatus initialize failed");
        return TTP_ERROR;
    }
    if (cbTimedwait_.Initialize() == TTP_ERROR) {
        TTP_LOG_ERROR("ActionEngine: cbTimedwait initialize failed");
        return TTP_ERROR;
    }

    replyStatus_ = MakeRef<AtomicStatusVector>();
    if (replyStatus_ == nullptr) {
        TTP_LOG_ERROR("ActionEngine: create replyStatus failed");
        return TTP_ERROR;
    }
    if (replyStatus_->Initialize(worldSize, TTP_BUTT, TTP_OK) == TTP_ERROR) {
        TTP_LOG_ERROR("replyStatus Initialize failed");
        return TTP_ERROR;
    }
    if (replyTimedwait_.Initialize() == TTP_ERROR) {
        TTP_LOG_ERROR("ActionENgine: replyTimedwait initialize failed");
        return TTP_ERROR;
    }
    return TTP_OK;
}

TResult ActionEngine::HandleReplyCommon(const AccTcpRequestContext &context)
{
    int16_t type = context.MsgType();
    TTP_ASSERT_RETURN(type >= 0 && type < TTP_MSG_OP_BUTT, TTP_ERROR);
    TTPReplyMsg recvMsg = {TTP_ERROR, 0, -1};
    if (extraReplyParseHandler_[type] == nullptr) {
        if (context.DataLen() != sizeof(TTPReplyMsg)) {
            TTP_LOG_ERROR("invalid reply data len: " << context.DataLen() << " msg_type: " << StrMsgOpCode(type));
            return TTP_ERROR;    // directly return, avoid segmentation fault; will retry after reply time-out
        }
        TTPReplyMsg *replyMsg = static_cast<TTPReplyMsg *>(context.DataPtr());
        TTP_ASSERT_RETURN(replyMsg != nullptr, TTP_ERROR);
        recvMsg = *replyMsg;
        if (replyMsg->status != TTP_OK) {
            TTP_LOG_ERROR("handle reply failed, rank: " << replyMsg->rank <<" msg_type: " << StrMsgOpCode(type));
        }
    } else {
        int32_t ret = extraReplyParseHandler_[type](context, recvMsg);
        if (ret != TTP_OK) {
            TTP_LOG_ERROR("handle extra reply failed, msg.sn: " << recvMsg.sn << " msg_type: " << StrMsgOpCode(type));
        }
    }

    bool counterAdd = false;
    if (replyStatus_->CheckSetRankStatus(recvMsg.rank, recvMsg.sn, recvMsg.status, counterAdd) != TTP_OK) {
        TTP_LOG_INFO(type << ": receive outdated reply msg, rank:" << recvMsg.rank << "received msg sn: " <<
            recvMsg.sn << ", ActionEngine sn: " << expectCount_.GetSn() << ", drop reply msg.");
        return TTP_ERROR;
    }

    uint64_t addResult = 0;
    if (counterAdd) {
        if (realCount_.CheckAdd(recvMsg.sn, 1, addResult) != TTP_OK) {
            TTP_LOG_INFO("msg type: " << type << ", reply sn check failed, received msg sn: " << recvMsg.sn <<
                ", ActionEngine sn: " << expectCount_.GetSn() << ", rank: " << recvMsg.rank << ", drop reply msg.");
            return TTP_ERROR;
        }
        if (extraReplyHandle_[type] != nullptr) {
            extraReplyHandle_[type](context);
        }
    }
    TTP_LOG_DEBUG("msg type:" << type << ", realCount.sn: " << realCount_.GetSn() << ", realCount.count:" <<
        realCount_.GetCount() << ", expectCount.sn: " << expectCount_.GetSn() << ", expectCount.count: " <<
        expectCount_.GetCount() << ", addResult: " << addResult << ", expectCount.Load(): "<< expectCount_.Load());

    if (addResult == expectCount_.Load()) {
        replyStatusLock_.LockRead();
        if (realCount_.GetSn() == recvMsg.sn) {
            replyTimedwait_.PthreadSignal();
        }
        replyStatusLock_.UnLock();
    }
    return TTP_OK;
}

TResult ActionEngine::FindUnackInfo(const std::vector<ActionInfo> &info, AtomicStatusPtr statusPtr, uint16_t sn,
                                    std::vector<ActionInfo> &unackInfo, bool isErrorRetry)
{
    unackInfo.clear();
    for (const ActionInfo &actInfo : info) {
        std::vector<int32_t> unAckRanks;
        for (int32_t curRank : actInfo.targetRanks) {
            TResult ret = statusPtr->CheckRankGroupStatus(std::vector<int32_t>(1, curRank), sn);
            if (ret == TTP_NEED_RETRY || (ret == TTP_ERROR && isErrorRetry)) {
                unAckRanks.push_back(curRank);
            }
        }
        // re-send unAckRanks
        if (!unAckRanks.empty()) {
            unackInfo.push_back(actInfo);
            unackInfo.back().targetRanks = unAckRanks;
        }
    }
    return TTP_OK;
}

TResult ActionEngine::FindNoResponseRanks(AtomicStatusPtr statusPtr, uint16_t sn, const std::vector<int32_t> &ranks,
                                          std::vector<int32_t> &ranksNoResponse)
{
    ranksNoResponse.clear();
    uint64_t initVal = statusPtr->GenerateWholeStatus(sn, statusPtr->GetInitStatus());
    for (auto rank : ranks) {
        uint64_t status = 0;
        if (statusPtr->LoadRank(rank, sn, status) != TTP_OK) {
            continue;
        }
        if (initVal == status) {    // init status, no cb or reply
            ranksNoResponse.push_back(rank);
        }
    }
    return TTP_OK;
}

TResult ActionEngine::SendInfo(std::vector<ActionInfo> &info)
{
    uint32_t infoSize = info.size();
    if (infoSize == 0) {
        return TTP_OK;
    }

    uint32_t sendSuccess = 0;
    for (uint32_t i = 0; i < infoSize; i++) {
        std::vector<AccDataBufferPtr> cbCtx;
        for (uint32_t j = 0; j < info[i].targetRanks.size(); j++) {
            AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(CallbackCtx));
            if (buffer == nullptr) {
                TTP_LOG_ERROR("malloc buffer failed.");
                return TTP_ERROR;
            }
            CallbackCtx *cbCtxPtr = static_cast<CallbackCtx *>(buffer->DataPtrVoid());
            cbCtxPtr->sn = sendCount_.GetSn();
            cbCtxPtr->rank = info[i].targetRanks.at(j);
            cbCtx.push_back(buffer);
        }

        int32_t ret = TTP_ERROR;
        uint32_t retryTime = 0;
        while (retryTime <= TTP_SENT_ERROR_RETRY_TIMES) {
            ret = actionSend_(info[i].msgType, info[i].d, info[i].targetRanks, cbCtx); // will change targetRanks
            retryTime++;
            if (ret == TTP_OK) {
                sendSuccess += info[i].targetRanks.size();
                break;
            }
            TTP_LOG_DEBUG("actionSend failed, current send rankNum: " << info[i].targetRanks.size()
                << ", sn: " << sendCount_.GetSn() << ", actionSend count: " << retryTime);
            std::this_thread::sleep_for(std::chrono::milliseconds(TTP_SENT_ERROR_RETRY_INTERVAL));
        }
        if (ret != TTP_OK) {
            TTP_LOG_INFO("send msg success, msg_type:" << StrMsgOpCode(info.front().msgType) <<
                         " rank_num:" << sendSuccess << ", then send msg failed");
            return TTP_ERROR;
        }
    }
    TTP_LOG_INFO("send msg success, msg_type:" << StrMsgOpCode(info.front().msgType) << " rank_num:" << sendSuccess);
    return TTP_OK;
}

void ActionEngine::RankStatusRegister(const MarkRankStatusMethod &h)
{
    TTP_ASSERT_RET_VOID(h != nullptr);
    markStatusHandle_ = h;
}

}  // namespace ttp
}  // namespace ock