/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */
#include <gtest/gtest.h>
#include <mockcpp/mockcpp.hpp>
#include "common.h"
#include "action_engine.h"

#include "acc_tcp_server_default.h"

using namespace ock::ttp;

#ifndef MOCKER_CPP
#define MOCKER_CPP(api, TT) MOCKCPP_NS::mockAPI(#api, reinterpret_cast<TT>(api))
#endif
#ifndef MOCKCPP_RESET
#define MOCKCPP_RESET GlobalMockObject::reset()
#endif

namesapce {
AccReqSentHandler g_requestSentHandle[UNO_48]{};
AccNewReqHandler g_newRequestHandle[UNO_48]{};
AccTcpServerPtr g_testServer;
ActionEnginePtr g_testEngine;
int32_t g_worldSize = 8;
int32_t g_sn = 1;
int32_t g_replyStatus = TTP_OK;
AccMsgSentResult g_msgResult = MSG_SENT;
std::atomic<uint32_t> g_sendCount = { 0 };
std::atomic<uint32_t> g_replyCount = { 0 };
std::atomic<bool> g_hasReply = { false };

class TestActionEngine : public testing::Test {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();
    void SetUp() override;
    void TearDown() override;

public:
    void InitSource();
};

void RegisterRequestSentStub(AccTcpServer *server, int16_t msgType, const AccReqSentHandler &h)
{
    ASSERT_RET_VOID(msgType >= MIN_MSG_TYPE && msgType < MAX_MSG_TYPE);
    ASSERT_RET_VOID(h != nullptr);
    g_requestSentHandle[msgType] = h;
}

void RegisterNewRequestStub(AccTcpServer *server, int16_t msgType, const AccNewReqHandler &h)
{
    ASSERT_RET_VOID(msgType >= MIN_MSG_TYPE && msgType < MAX_MSG_TYPE);
    ASSERT_RET_VOID(h != nullptr);
    g_newRequestHandle[msgType] = h;
}

void GetAllLinkRanks(std::vector<int32_t> &ranks)
{
    ranks.clear();
    for (int32_t i = 0; i < g_worldSize; i++) {
        ranks.push_back(i);
    }
}

TResult SendMsg(int16_t msgType, const AccDataBufferPtr &d, std::vector<int32_t> &targetRanks,
                const std::vector<AccDataBufferPtr> &cbCtx)
{
    uint32_t rankNum = targetRanks.size();
    for (uint32_t i = 0; i < rankNum; i++) {
        int32_t curRank = targetRanks.at(i);
        g_sendCount.fetch_add(1);

        if (curRank >= g_worldSize || g_requestSentHandle[msgType] == nullptr) {
            return TTP_ERROR;
        }

        AccMsgHeader header = { msgType, d->DataLen(), 0 };
        g_requestSentHandle[msgType](g_msgResult, header, cbCtx.at(i));
    }

    if (g_hasReply.load()) {
        int16_t replyType = msgType + 1;
        for (uint32_t i = 0; i < rankNum; i++) {
            int32_t curRank = targetRanks[i];
            g_replyCount.fetch_add(1);

            if (curRank >= g_worldSize || g_newRequestHandle[replyType] == nullptr) {
                return TTP_ERROR;
            }

            AccMsgHeader header = { replyType, sizeof(TTPReplyMsg), 0 };
            AccTcpLinkComplexPtr ptr;
            AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(TTPReplyMsg));
            TTPReplyMsg *replyMsg = static_cast<TTPReplyMsg *>(buffer->DataPtrVoid());
            replyMsg->status = g_replyStatus;
            replyMsg->sn = g_sn;
            replyMsg->rank = curRank;
            AccTcpRequestContext ctx(header, buffer, ptr);
            g_newRequestHandle[replyType](ctx);
        }
    }

    return TTP_OK;
}

void TestActionEngine::InitSource()
{
    MOCKER_CPP(&AccTcpServerDefault::RegisterRequestSentHandler,
        void(*)(AccTcpServer *server, int16_t, const AccReqSentHandler&)).stubs().will(invoke(RegisterRequestSentStub));
    MOCKER_CPP(&AccTcpServerDefault::RegisterNewRequestHandler,
        void(*)(AccTcpServer *server, int16_t, const AccNewReqHandler&)).stubs().will(invoke(RegisterNewRequestStub));

    g_testServer = AccTcpServer::Create();
    ASSERT_NE(g_testServer, nullptr);

    g_testEngine = MakeRef<ActionEngine>();
    ASSERT_NE(g_testEngine, nullptr);

    ActionMsgSend send = std::bind(&SendMsg, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3,
                                   std::placeholders::_4);
    int32_t ret = g_testEngine->Initialize(g_testServer, send, g_worldSize);
    ASSERT_EQ(ret, TTP_OK);

    auto statusMakrMethod = [this](const std::vector<int32_t> &ranks) { return TTP_OK; };
    g_testEngine->RankStatusRegister(statusMakrMethod);

    MOCKCPP_RESET;
}

void TestActionEngine::SetUpTestCase() {}

void TestActionEngine::TearDownTestCase()
{
    for (int i = MIN_MSG_TYPE; i < MAX_MSG_TYPE; i++) {
        g_requestSentHandle[i] = nullptr;
    }
}

void TestActionEngine::SetUp() {}

void TestActionEngine::TearDown()
{
    GlobalMockObject::reset();
}

TEST_F(TestActionEngine, engine_init)
{
    TestActionEngine::InitSource();

    ASSERT_NE(g_requestSentHandle[TTP_MSG_OP_CKPT_SEND], nullptr);
    ASSERT_NE(g_requestSentHandle[TTP_MSG_OP_CTRL_NOTIFY], nullptr);
    ASSERT_NE(g_requestSentHandle[TTP_MSG_OP_RENAME], nullptr);
    ASSERT_NE(g_requestSentHandle[TTP_MSG_OP_EXIT], nullptr);
}

TEST_F(TestActionEngine, invalid_actionOp)
{
    TestActionEngine::InitSource();

    std::vector<ActionInfo> info;
    int32_t ret = g_testEngine->Process(ACTION_OP_BUTT, info, false, g_sn);
    ASSERT_EQ(ret, TTP_ERROR);
}

TEST_F(TestActionEngine, engine_uninit)
{
    g_testEngine = MakeRef<ActionEngine>();
    ASSERT_NE(g_testEngine, nullptr);

    std::vector<ActionInfo> info;
    int32_t ret = g_testEngine->Process(ACTION_OP_EXIT, info, false, g_sn);
    ASSERT_EQ(ret, TTP_ERROR);
}

TEST_F(TestActionEngine, cb_failed)
{
    constexpr uint32_t TTP_CB_WAIT_RETRY_TIMES = 3;
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(0);
    info.push_back({TTP_MSG_OP_EXIT, buffer, ranks});

    g_msgResult = MSG_TIMEOUT;
    g_sendCount.store(0);
    int32_t ret = g_testEngine->Process(ACTION_OP_EXIT, info, false, g_sn);
    ASSERT_EQ(ret, TTP_ERROR);
    ASSERT_EQ(g_sendCount.load(), (TTP_CB_WAIT_RETRY_TIMES + 1) * g_worldSize);

    g_sendCount.store(0);
    g_msgResult = MSG_SENT;
    ret = g_testEngine->Process(ACTION_OP_EXIT, info, false, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), g_worldSize);
    g_sendCount.store(0);
}

TEST_F(TestActionEngine, reply_check)
{
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(uint16_t));
    uint16_t *sn = static_cast<uint16_t *>(buffer->DataPtrVoid());
    *sn = g_sn;
    info.push_back({TTP_MSG_OP_DEVICE_STOP, buffer, ranks});

    int32_t ret;
    g_hasReply.store(true);
    g_sendCount.store(0);
    g_replyCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_DEVICE_STOP, info, true, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
    ASSERT_EQ(g_replyCount.load(), ranks.size());

    g_replyStatus = TTP_ERROR;
    g_sendCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_DEVICE_STOP, info, true, g_sn);
    g_hasReply.store(false);
    g_replyStatus = TTP_OK;
    ASSERT_EQ(ret, TTP_ERROR);
    ASSERT_EQ(g_sendCount.load(), g_worldSize); // reply error not retry
    g_sendCount.store(0);
}

TEST_F(TestActionEngine, action_rename)
{
    TestActionEngine::InitSource();

    int32_t rank = 0;
    AccDataBufferPtr buffer = AccDataBuffer::Create(sizeof(int32_t));
    int32_t *rankPtr = static_cast<int32_t *>(buffer->DataPtrVoid());
    *rankPtr = rank;

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    ranks.push_back(rank);
    info.push_back({TTP_MSG_OP_RENAME, buffer, ranks});

    int32_t ret;
    g_sendCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_RENAME, info, false, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
}

TEST_F(TestActionEngine, action_exit)
{
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(0);
    info.push_back({TTP_MSG_OP_EXIT, buffer, ranks});

    int32_t ret;
    g_sendCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_EXIT, info, false, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
}

TEST_F(TestActionEngine, action_stop)
{
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(0);
    info.push_back({TTP_MSG_OP_DEVICE_STOP, buffer, ranks});

    int32_t ret;
    g_sendCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_DEVICE_STOP, info, false, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
}

TEST_F(TestActionEngine, action_clean)
{
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(0);
    info.push_back({TTP_MSG_OP_DEVICE_CLEAN, buffer, ranks});

    int32_t ret;
    g_hasReply.store(true);
    g_sendCount.store(0);
    g_replyCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_DEVICE_CLEAN, info, true, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
    ASSERT_EQ(g_replyCount.load(), ranks.size());
    g_hasReply.store(false);
}

TEST_F(TestActionEngine, action_save_ckpt)
{
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(0);
    info.push_back({TTP_MSG_OP_CKPT_SEND, buffer, ranks});

    int32_t ret;
    g_hasReply.store(true);
    g_sendCount.store(0);
    g_replyCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_SAVECKPT, info, true, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
    ASSERT_EQ(g_replyCount.load(), ranks.size());
    g_hasReply.store(false);
}

TEST_F(TestActionEngine, action_prelock)
{
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(0);
    info.push_back({TTP_MSG_OP_PRELOCK, buffer, ranks});

    int32_t ret;
    g_hasReply.store(true);
    g_sendCount.store(0);
    g_replyCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_PRELOCK, info, true, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
    ASSERT_EQ(g_replyCount.load(), ranks.size());
    g_hasReply.store(false);
}

TEST_F(TestActionEngine, action_notify_normal)
{
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(0);
    info.push_back({TTP_MSG_OP_NOTIFY_NORMAL, buffer, ranks});

    int32_t ret;
    g_hasReply.store(true);
    g_sendCount.store(0);
    g_replyCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_NOTIFY_NORMAL, info, true, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
    ASSERT_EQ(g_replyCount.load(), ranks.size());
    g_hasReply.store(false);
}

TEST_F(TestActionEngine, action_collection)
{
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(0);
    info.push_back({TTP_MSG_OP_COLLECTION, buffer, ranks});

    int32_t ret;
    g_hasReply.store(true);
    g_sendCount.store(0);
    g_replyCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_COLLECTION, info, true, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
    ASSERT_EQ(g_replyCount.load(), ranks.size());
    g_hasReply.store(false);
}

TEST_F(TestActionEngine, action_repair)
{
    TestActionEngine::InitSource();

    std::vector<int32_t> ranks;
    std::vector<ActionInfo> info;
    GetAllLinkRanks(ranks);
    AccDataBufferPtr buffer = AccDataBuffer::Create(0);
    info.push_back({TTP_MSG_OP_REPAIR, buffer, ranks});

    int32_t ret;
    g_hasReply.store(true);
    g_sendCount.store(0);
    g_replyCount.store(0);
    ret = g_testEngine->Process(ACTION_OP_REPAIR, info, true, g_sn);
    ASSERT_EQ(ret, TTP_OK);
    ASSERT_EQ(g_sendCount.load(), ranks.size());
    ASSERT_EQ(g_replyCount.load(), ranks.size());
    g_hasReply.store(false);
}
}