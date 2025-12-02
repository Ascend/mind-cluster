/*
* Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
*/

#include <unistd.h>
#include "Secodefuzz/secodeFuzz.h"
#include "mockcpp/mockcpp.hpp"
#include "common.h"
#include "controller.h"
#include "processor.h"
#include "test_c2python_common.h"
#include "test_c2python_api.h"

constexpr int32_t NUM_TWO = 2;
constexpr int32_t NUM_127 = 127;
constexpr int32_t NUM_EIGHT = 8;
constexpr int32_t NUM_THREE = 3;

using namespace ock::ttp;

namespace {
ProcessorPtr g_processor0 = nullptr;
ProcessorPtr g_processor1 = nullptr;
ControllerPtr g_controller1 = nullptr;

std::string g_masterIp = "127.0.0.1";
uint32_t g_masterPort = 8000;
AccTlsOption g_tlsOption;
bool g_enableLocalCopy = false;

void C2PythonDtFuzz::Init(bool enableARF = false, bool enableZIT = false)
{
    C2PythonDtFuzz::InitController(g_controller1);
    g_controller1->Initialize(0, NUM_TWO, g_enableLocalCopy, enableARF, enableZIT);

    std::string ip = g_masterIp;
    int32_t port1 = g_masterPort;
    g_controller1->Start(ip, port1, g_tlsOption);

    std::vector<int32_t> ranks = {0, 1};
    std::vector<std::vector<int32_t>> groups = { ranks };
    C2PythonDtFuzz::InitProcessor(g_processor0);
    C2PythonDtFuzz::InitProcessor(g_processor1);

    g_processor0->Initialize(0, NUM_TWO, g_enableLocalCopy, g_tlsOption);
    g_processor1->Initialize(1, NUM_TWO, g_enableLocalCopy, g_tlsOption);

    g_processor0->Start(ip, port1);
    g_processor1->Start(ip, port1);
}

void C2PythonDtFuzz::Destroy()
{
    // connectedLinks_释放AccTcpLinkComplexPtr引用bug，先destroy processor
    if (g_processor0 != nullptr) {
        g_processor0->Destroy();
    }
    if (g_processor1 != nullptr) {
        g_processor1->Destroy();
    }
    if (g_controller1 != nullptr) {
        g_controller1->Destroy();
    }
}

TEST_F(C2PythonDtFuzz, report_status_fuzz)
{
    DT_Enable_Leak_Check(0, 0);
    DT_Set_Running_Time_Second(FUZZ_TEST_SECONDS);

    std::vector<int32_t> ranks = {0, 1};
    std::vector<std::vector<int32_t>> groups = { ranks };
    C2PythonDtFuzz::Init();

    DT_FUZZ_START(0, FUZZ_TEST_TIMES, "test_report_replica_info", 0)
    {
        int32_t index = 0;
        int32_t size1 = *(int32_t *)DT_SetGetNumberRange(&g_Element[index++], 1, 0, NUM_THREE);

        std::vector<int32_t> replicaCnt;
        for (int32_t i = 0; i < size1; i++) {
            int32_t cnt = *(int32_t *)DT_SetGetNumberRange(&g_Element[index++], 1, 0, NUM_THREE);
            replicaCnt.push_back(cnt);
        }

        int32_t size2 = *(int32_t *)DT_SetGetNumberRange(&g_Element[index++], 1, 0, NUM_THREE);
        std::vector<int32_t> replicaShift;
        for (int32_t i = 0; i < size2; i++) {
            replicaShift.push_back(0);
        }

        g_processor0->ReportReplicaInfo(groups, { NUM_TWO }, { 0 });
        g_processor1->ReportReplicaInfo(groups, replicaCnt, replicaShift);
    }
    DT_FUZZ_END()

    C2PythonDtFuzz::Destroy();
}

TEST_F(C2PythonDtFuzz, set_status_fuzz)
{
    DT_Enable_Leak_Check(0, 0);
    DT_Set_Running_Time_Second(FUZZ_TEST_SECONDS);

    std::vector<int32_t> ranks = {0, 1};
    std::vector<std::vector<int32_t>> groups = { ranks };

    DT_FUZZ_START(0, FUZZ_TEST_TIMES, "test_set_step_report_state", 0)
    {
        C2PythonDtFuzz::Init();
        g_processor0->ReportReplicaInfo(groups, { NUM_TWO }, { 0 });
        g_processor1->ReportReplicaInfo(groups, { NUM_TWO }, { 0 });
        int64_t backupStep = *(int64_t *)DT_SetGetS64(&g_Element[0], 1);
        if (backupStep < -1) {
            backupStep = -1;
        }
        int64_t step = *(int64_t *)DT_SetGetS64(&g_Element[1], 1);

        g_processor0->BeginUpdating(backupStep);
        g_processor0->FinishedUpdate(step);

        g_processor1->BeginUpdating(backupStep);
        g_processor1->FinishedUpdate(step);

        int32_t state = *(int32_t *)DT_SetGetS32(&g_Element[2], 0);
        TTP_LOG_INFO("FUZZ ===== backupStep: " << backupStep << ", step: " << step << " ==== state: " << state);

        ReportState rs = ReportState::RS_NORMAL;
        g_processor1->ReportStatus(rs);

        usleep(50); // 50
        C2PythonDtFuzz::Destroy();
    }
    DT_FUZZ_END()
}

TEST_F(C2PythonDtFuzz, controller_service_fuzz)
{
    DT_Enable_Leak_Check(0, 0);
    DT_Set_Running_Time_Second(FUZZ_TEST_SECONDS);

    DT_FUZZ_START(0, FUZZ_TEST_TIMES, "test_controller_severice", 0)
    {
        int32_t index = 0;
        bool arf = false;
        bool zit = false;
        bool localCopy = false;
        int32_t rank = *(int32_t *)DT_SetGetS32(&g_Element[index++], 0);
        int32_t worldSize = *(int32_t *)DT_SetGetS32(&g_Element[index++], NUM_EIGHT);
        int32_t port2 = *(int32_t *)DT_SetGetS32(&g_Element[index++], g_masterPort);
        int32_t num = *(int32_t *)DT_SetGetS32(&g_Element[index++], 0);
        if (num == 0) {
            localCopy = true;
        } else if (num == 1) {
            arf = true;
        } else if (num == NUM_TWO) {
            zit = true;
        }

        AccTlsOption tlsOpts;
        uint8_t subIp1 = *(uint8_t *)DT_SetGetU8(&g_Element[index++], NUM_127);
        uint8_t subIp2 = *(uint8_t *)DT_SetGetU8(&g_Element[index++], 0);
        uint8_t subIp3 = *(uint8_t *)DT_SetGetU8(&g_Element[index++], 0);
        uint8_t subIp4 = *(uint8_t *)DT_SetGetU8(&g_Element[index++], 1);
        std::string ip = std::to_string(subIp1) + "." + std::to_string(subIp2) + "." + std::to_string(subIp3) + "." +
            std::to_string(subIp4);

        TResult ret = Controller::GetInstance()->Initialize(rank, worldSize, localCopy, arf, zit);
        Controller::GetInstance()->Start(ip, port2, tlsOpts);
        if (ret == TTP_OK) {
            Controller::GetInstance()->Destroy();
        }
        usleep(100);    // 100
    }
    DT_FUZZ_END()
}


TEST_F(C2PythonDtFuzz, processor_service_fuzz)
{
    DT_Enable_Leak_Check(0, 0);
    DT_Set_Running_Time_Second(FUZZ_TEST_SECONDS);

    DT_FUZZ_START(0, FUZZ_TEST_TIMES, "test_processor_severice", 0)
    {
        int32_t index = 0;
        AccTlsOption tlsOpts;
        bool localCopy;
        int32_t num = *(int32_t *)DT_SetGetS32(&g_Element[index++], 0);
        if (num == 1) {
            localCopy = true;
        } else {
            localCopy = false;
        }
        int32_t rank = *(int32_t *)DT_SetGetS32(&g_Element[index++], 0);
        int32_t worldSize = *(int32_t *)DT_SetGetS32(&g_Element[index++], NUM_EIGHT);
        int32_t port3 = *(int32_t *)DT_SetGetS32(&g_Element[index++], g_masterPort);
        uint8_t subIp1 = *(uint8_t *)DT_SetGetU8(&g_Element[index++], NUM_127);
        uint8_t subIp2 = *(uint8_t *)DT_SetGetU8(&g_Element[index++], 0);
        uint8_t subIp3 = *(uint8_t *)DT_SetGetU8(&g_Element[index++], 0);
        uint8_t subIp4 = *(uint8_t *)DT_SetGetU8(&g_Element[index++], 1);
        std::string ip = std::to_string(subIp1) + "." + std::to_string(subIp2) + "." + std::to_string(subIp3) + "." +
            std::to_string(subIp4);
        TResult ret = Processor::GetInstance()->Initialize(rank, worldSize, localCopy, tlsOpts);
        Processor::GetInstance()->Start(ip, port3);
        if (ret == TTP_OK) {
            Processor::GetInstance()->Destroy();
        }
        usleep(100);    // 100
    }
    DT_FUZZ_END()
}

void C2PythonDtFuzz::SetUpTestSuite() {}

void C2PythonDtFuzz::TearDownTestSuite() {}
}