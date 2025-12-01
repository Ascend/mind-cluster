/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
 */
#include <gtest/gtest.h>
#include <mockcpp/mokc.h>
#include <mockcpp/mockcpp.hpp>
#include <gmock/gmock.h>
#include "acc_tcp_ssl_helper.h"
#include "common.h"
#include "controller_state_machine.h"

using namespace ock::ttp;

#define MOCKER_CPP(api, TT) MOCKCPP_NS::mockAPI(#api, reinterpret_cast<TT>(api))
#ifndef MOCKCPP_RESET
#define MOCKCPP_RESET GlobalMockObject::reset()
#endif

namesapce {
using controllerStateMachinePtr = Ref<controllerStateMachine>;
controllerStateMachinePtr stateMachinePtr = nullptr;
std::atomic<uint32_t> g_callCount;
int32_t g_initSn = 0;
int32_t g_normalSn = 0;
int32_t g_pauseSn = 0;
int32_t g_abnormalSn = 0;
int32_t g_dumpSn = 0;
int32_t g_exitSn = 0;
int32_t g_envclearSn = 0;
int32_t g_repairSn = 0;
int32_t g_downgradeSn = 0;
int32_t g_upgradeSn = 0;
int32_t g_dgRepairSn = 0;

constexpr uint32_t CHECK_COUNT_ONE = 1;
constexpr uint32_t CHECK_COUNT_TWO = 2;
constexpr uint32_t CHECK_COUNT_THREE = 3;
constexpr uint32_t CHECK_COUNT_FOUR = 4;
constexpr uint32_t CHECK_COUNT_FIVE = 5;
constexpr uint32_t CHECK_COUNT_SIX = 6;
constexpr uint32_t CHECK_COUNT_SEVEN = 7;
constexpr uint32_t CHECK_COUNT_EIGHT = 8;
constexpr uint32_t CHECK_COUNT_NINE = 9;

bool    g_dump = false;
bool    g_pause = false;
bool    g_continue = false;
bool    g_downgrade = false;


class TestControllerStateMachine : public testing::Test {
public:
    static void SetUpTestCase();
    static void TearDownTestCase();
    void SetUp() override;
    void TearDown() override;
    void InitSource();
private:
    int32_t InitCallback();
    int32_t NormalCallback();
    int32_t PauseCallback();
    int32_t AbnormalCallback();
    int32_t DumpCallback();
    int32_t ExitCallback();
    int32_t EnvclearCallback();
    int32_t RepairCallback();
    int32_t DowngradeCallback();
    int32_t UpgradeCallback();
    int32_t DgRepairCallback();
};

void TestControllerStateMachine::SetUpTestCase() {}

void TestControllerStateMachine::TearDownTestCase() {}

void TestControllerStateMachine::SetUp() {}

void TestControllerStateMachine::TearDown() {}

void TestControllerStateMachine::InitSource()
{
    g_callCount.store(0);
    stateMachinePtr = MakeRef<controllerStateMachine>();
    int32_t ret = stateMachinePtr->Initialize(0);
    ASSERT_EQ(ret, TTP_OK);

    g_initSn = 0;
    g_normalSn = 0;
    g_pauseSn = 0;
    g_abnormalSn = 0;
    g_dumpSn = 0;
    g_exitSn = 0;
    g_envclearSn = 0;
    g_repairSn = 0;
    g_downgradeSn = 0;
    g_upgradeSn = 0;
    g_dgRepairSn = 0;

    auto initMethod = [this]() { return InitCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_INIT, initMethod);
    auto normalMethod = [this]() { return NormalCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_NORMAL, normalMethod);
    auto abnormalMethod = [this]() { return AbnormalCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_ABNORMAL, abnormalMethod);
    auto dumpMethod = [this]() { return DumpCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_DUMP, dumpMethod);
    auto exitMethod = [this]() { return ExitCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_EXIT, exitMethod);
    auto envclearMethod = [this]() { return EnvclearCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_ENV_CLEAR, envclearMethod);
    auto repairMethod = [this]() { return RepairCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_REPAIR, repairMethod);
    auto downgradeMethod = [this]() { return DowngradeCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_DOWNGRADE, downgradeMethod);
    auto upgradeMethod = [this]() { return UpgradeCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_UPGRADE, upgradeMethod);
    auto dpRepairMethod = [this]() { return DgRepairCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_DG_REPAIR, dpRepairMethod);
    auto pauseMethod = [this]() { return PauseCallback(); };
    stateMachinePtr->ControllerActionRegister(STATE_OP_PAUSE, pauseMethod);
}
int32_t TestControllerStateMachine::InitCallback()
{
    g_callCount.fetch_add(1);
    g_initSn = g_callCount.load();
    TTP_LOG_INFO("STATE_OP_INIT --> STATE_OP_NORMAL");
    return TTP_OK;
}

int32_t TestControllerStateMachine::NormalCallback()
{
    g_callCount.fetch_add(1);
    g_normalSn = g_callCount.load();
    TTP_LOG_INFO("STATE_OP_NORMAL --> STATE_OP_ABNORMAL");
    if (g_pause) {
        TTP_LOG_INFO("STATE_OP_ABNORMAL --> STATE_OP_PAUSE");
        return TTP_PAUSE;
    }
    return TTP_ERROR;
}
int32_t TestControllerStateMachine::AbnormalCallback()
{
    g_callCount.fetch_add(1);
    g_abnormalSn = g_callCount.load();
    if (g_dump) {
        TTP_LOG_INFO("STATE_OP_ABNORMAL --> STATE_OP_DUMP");
        return TTP_ERROR;
    } else {
        TTP_LOG_INFO("STATE_OP_ABNORMAL --> STATE_OP_ENV_CLEAR");
        return TTP_OK;
    }
}
int32_t TestControllerStateMachine::PauseCallback()
{
    g_callCount.fetch_add(1);
    g_pauseSn = g_callCount.load();
    if (g_continue) {
        TTP_LOG_INFO("STATE_OP_PAUSE --> STATE_OP_NORMAL");
        return TTP_OK;
    } else {
        TTP_LOG_INFO("STATE_OP_PAUSE --> STATE_OP_ABNORMAL");
        return TTP_ERROR;
    }
}
int32_t TestControllerStateMachine::DumpCallback()
{
    g_callCount.fetch_add(1);
    g_dumpSn = g_callCount.load();
    TTP_LOG_INFO("STATE_OP_DUMP --> STATE_OP_EXIT");
    return TTP_OK;
}
int32_t TestControllerStateMachine::ExitCallback()
{
    g_callCount.fetch_add(1);
    g_exitSn = g_callCount.load();
    TTP_LOG_INFO("STATE_OP_EXIT --> STATE_OP_FINAL");
    return TTP_STOP_SERVICE;
}
int32_t TestControllerStateMachine::EnvclearCallback()
{
    g_callCount.fetch_add(1);
    g_envclearSn = g_callCount.load();
    if (g_downgrade) {
        TTP_LOG_INFO("STATE_OP_ENV_CLEAR --> STATE_OP_DG_REPAIR");
        return TTP_DOWNGRADE;
    } else {
        TTP_LOG_INFO("STATE_OP_ENV_CLEAR --> STATE_OP_REPAIR");
        return TTP_OK;
    }
}
int32_t TestControllerStateMachine::RepairCallback()
{
    g_callCount.fetch_add(1);
    g_repairSn = g_callCount.load();
    TTP_LOG_INFO("STATE_OP_REPAIR --> STATE_OP_DUMP");
    return TTP_ERROR;
}
int32_t TestControllerStateMachine::DowngradeCallback()
{
    g_callCount.fetch_add(1);
    g_downgradeSn = g_callCount.load();
    TTP_LOG_INFO("STATE_OP_DOWNGRADE --> STATE_OP_UPGRADE");
    return TTP_NEED_RETRY;
}
int32_t TestControllerStateMachine::UpgradeCallback()
{
    g_callCount.fetch_add(1);
    g_upgradeSn = g_callCount.load();
    TTP_LOG_INFO("STATE_OP_UPGRADE --> STATE_OP_DUMP");
    return TTP_ERROR;
}
int32_t TestControllerStateMachine::DgRepairCallback()
{
    g_callCount.fetch_add(1);
    g_dgRepairSn = g_callCount.load();
    TTP_LOG_INFO("STATE_OP_DG_REPAIR --> STATE_OP_DOWNGRADE");
    return TTP_OK;
}


TEST_F(TestControllerStateMachine, dump_state)
{
    g_dump = true;
    InitSource();
    stateMachinePtr->StartStateMachine();
    sleep(1); // wait state transition

    ASSERT_EQ(g_initSn, 1);
    ASSERT_EQ(g_normalSn, 2);
    ASSERT_EQ(g_abnormalSn, 3);
    ASSERT_EQ(g_dumpSn, 4);
    ASSERT_EQ(g_exitSn, 5);

    g_dump = false;
}

TEST_F(TestControllerStateMachine, read_tls_option)
{
    AccTlsOption tlsOpts;
    std::string tlsInfo = R"(
tlsCert: test/llt/openssl_cert/certs/cert.pem;
tlsCrlPath: test/llt/openssl_cert/crl/;
tlsCaPath: test/llt/openssl_cert/ca/;
tlsCaFile: ca_cert_1.pem, ca_cert_2.pem;
tlsCrlFile: crl.pem;
tlsPk: test private key;
tlsPkPwd: test private key pwd;
packagePath: /etc/lib/
)";
    bool ret = FileUtils::ParseTlsInfo(tlsInfo, tlsOpts);
    tlsOpts.enableTls = true;
    ASSERT_EQ(ret, true);
    ASSERT_EQ(tlsOpts.enableTls, true);
    ASSERT_EQ(tlsOpts.tlsPkPwd, "test private key pwd");
    ASSERT_EQ(tlsOpts.tlsPk, "test private key");
    ASSERT_EQ(tlsOpts.tlsCrlPath, "test/llt/openssl_cert/crl/");
    ASSERT_EQ(tlsOpts.tlsCrlFile.size(), 1);
    ASSERT_EQ(tlsOpts.tlsCaPath, "test/llt/openssl_cert/ca/");
    ASSERT_EQ(tlsOpts.tlsCaFile.size(), 2);
    ASSERT_EQ(tlsOpts.tlsCert, "test/llt/openssl_cert/certs/cert.pem");
    for (auto file : tlsOpts.tlsCaFile) {
        std::cout << file << std::endl;
    }
    ASSERT_EQ(tlsOpts.packagePath, "/etc/lib/");
}

TEST_F(TestControllerStateMachine, uce_state)
{
    InitSource();
    stateMachinePtr->StartStateMachine();
    sleep(1); // wait state transition

    ASSERT_EQ(g_initSn, 1);
    ASSERT_EQ(g_normalSn, 2);
    ASSERT_EQ(g_abnormalSn, 3);
    ASSERT_EQ(g_envclearSn, 4);
    ASSERT_EQ(g_repairSn, 5);
    ASSERT_EQ(g_dumpSn, 6);
    ASSERT_EQ(g_exitSn, 7);
}

TEST_F(TestControllerStateMachine, zit_state)
{
    g_downgrade = true;
    InitSource();
    stateMachinePtr->StartStateMachine();
    sleep(1); // wait state transition

    ASSERT_EQ(g_initSn, 1);
    ASSERT_EQ(g_normalSn, 2);
    ASSERT_EQ(g_abnormalSn, 3);
    ASSERT_EQ(g_envclearSn, 4);
    ASSERT_EQ(g_dgRepairSn, 5);
    ASSERT_EQ(g_downgradeSn, 6);
    ASSERT_EQ(g_upgradeSn, 7);
    ASSERT_EQ(g_dumpSn, 8);
    ASSERT_EQ(g_exitSn, 9);

    g_downgrade = false;
}

TEST_F(TestControllerStateMachine, pause_state)
{
    g_pause = true;
    InitSource();
    stateMachinePtr->StartStateMachine();
    sleep(1); // wait state transition

    ASSERT_EQ(g_initSn, CHECK_COUNT_ONE);
    ASSERT_EQ(g_normalSn, CHECK_COUNT_TWO);
    ASSERT_EQ(g_pauseSn, CHECK_COUNT_THREE);
    ASSERT_EQ(g_abnormalSn, CHECK_COUNT_FOUR);
    ASSERT_EQ(g_envclearSn, CHECK_COUNT_FIVE);
    ASSERT_EQ(g_repairSn, CHECK_COUNT_SIX);
    ASSERT_EQ(g_dumpSn, CHECK_COUNT_SEVEN);
    ASSERT_EQ(g_exitSn, CHECK_COUNT_EIGHT);

    g_pause = false;
}
}