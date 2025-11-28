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

#include "controller_state_machine.h"

namespace ock {
namespace ttp {

struct StateTransitionItem {
    StateOp curState;
    int32_t retCode;
    StateOp nextState;
};

StateTransitionItem g_stateTransItemTable[]{
        {STATE_OP_INIT,        TTP_OK,           STATE_OP_NORMAL},
        {STATE_OP_INIT,        TTP_ERROR,        STATE_OP_FINAL},
        {STATE_OP_NORMAL,      TTP_OK,           STATE_OP_NORMAL},
        {STATE_OP_NORMAL,      TTP_ERROR,        STATE_OP_ABNORMAL},
        {STATE_OP_NORMAL,      TTP_PAUSE,        STATE_OP_PAUSE},
        {STATE_OP_NORMAL,      TTP_SWITCH,       STATE_OP_STEP_FINISH},
        {STATE_OP_NORMAL,      TTP_STOP_SERVICE, STATE_OP_FINAL},
        {STATE_OP_PAUSE,       TTP_OK,           STATE_OP_NORMAL},
        {STATE_OP_PAUSE,       TTP_ERROR,        STATE_OP_ABNORMAL},
        {STATE_OP_PAUSE,       TTP_STOP_SERVICE, STATE_OP_EXIT},
        {STATE_OP_STEP_FINISH, TTP_OK,           STATE_OP_MIGRATION},
        {STATE_OP_STEP_FINISH, TTP_STOP_SERVICE, STATE_OP_EXIT},
        {STATE_OP_MIGRATION,   TTP_OK,           STATE_OP_NORMAL},
        {STATE_OP_MIGRATION,   TTP_ERROR,        STATE_OP_DUMP},
        {STATE_OP_MIGRATION,   TTP_STOP_SERVICE, STATE_OP_EXIT},
        {STATE_OP_ABNORMAL,    TTP_OK,           STATE_OP_ENV_CLEAR},
        {STATE_OP_ABNORMAL,    TTP_ERROR,        STATE_OP_DUMP},
        {STATE_OP_ABNORMAL,    TTP_STOP_SERVICE, STATE_OP_EXIT},
        {STATE_OP_ENV_CLEAR,   TTP_OK,           STATE_OP_REPAIR},
        {STATE_OP_ENV_CLEAR,   TTP_ERROR,        STATE_OP_DUMP},
        {STATE_OP_ENV_CLEAR,   TTP_STOP_SERVICE, STATE_OP_EXIT},
        {STATE_OP_ENV_CLEAR,   TTP_DOWNGRADE,    STATE_OP_DG_REPAIR},
        {STATE_OP_DG_REPAIR,   TTP_OK,           STATE_OP_DOWNGRADE},
        {STATE_OP_DG_REPAIR,   TTP_ERROR,        STATE_OP_DUMP},
        {STATE_OP_DG_REPAIR,   TTP_STOP_SERVICE, STATE_OP_EXIT},
        {STATE_OP_REPAIR,      TTP_OK,           STATE_OP_NORMAL},
        {STATE_OP_REPAIR,      TTP_DOWNGRADE,    STATE_OP_DG_REPAIR},
        {STATE_OP_REPAIR,      TTP_ERROR,        STATE_OP_DUMP},
        {STATE_OP_REPAIR,      TTP_STOP_SERVICE, STATE_OP_EXIT},
        {STATE_OP_DOWNGRADE,   TTP_OK,           STATE_OP_DOWNGRADE},
        {STATE_OP_DOWNGRADE,   TTP_ERROR,        STATE_OP_DUMP},
        {STATE_OP_DOWNGRADE,   TTP_NEED_RETRY,   STATE_OP_UPGRADE},
        {STATE_OP_DOWNGRADE,   TTP_STOP_SERVICE, STATE_OP_EXIT},
        {STATE_OP_UPGRADE,     TTP_OK,           STATE_OP_ENV_CLEAR},
        {STATE_OP_UPGRADE,     TTP_ERROR,        STATE_OP_DUMP},
        {STATE_OP_UPGRADE,     TTP_STOP_SERVICE, STATE_OP_EXIT},
        {STATE_OP_DUMP,        TTP_OK,           STATE_OP_EXIT},
        {STATE_OP_DUMP,        TTP_ERROR,        STATE_OP_EXIT},
        {STATE_OP_EXIT,        TTP_STOP_SERVICE, STATE_OP_FINAL},
};

TResult controllerStateMachine::Initialize(int32_t rank)
{
    state_ = STATE_OP_INIT;
    stopFlag_.store(false);
    stateActionEnginePtr_ = MakeRef<stateActionEngine>();
    if (stateActionEnginePtr_ == nullptr) {
        TTP_LOG_ERROR("rank:" << rank << " create stateActionEngine failed");
        return TTP_ERROR;
    }
    // Initialize state Transition Rules to speedup looking up;
    for (uint32_t i = 0; i < sizeof(g_stateTransItemTable) / sizeof(StateTransitionItem); i++) {
        stateTransitionRules_[g_stateTransItemTable[i].curState][g_stateTransItemTable[i].retCode] =
                g_stateTransItemTable[i].nextState;
    }

    TTP_LOG_INFO("init controller StateMachine success, rank: " << rank << " state: " << StrStateOp(state_));
    return TTP_OK;
}

void controllerStateMachine::StartStateMachine()
{
    bool expectVal = false;
    if (!threadExistFlag_.compare_exchange_strong(expectVal, true)) {
        TTP_LOG_WARN("has state machine thread running..");
        return;
    }
    std::thread tmpThread(&controllerStateMachine::CheckAndUpdateState, this);
    stateChecker_ = std::move(tmpThread);
    stateChecker_.detach();
}

TResult controllerStateMachine::CheckAndUpdateState()
{
    TResult ret = TTP_OK;
    while (state_ != STATE_OP_FINAL && !stopFlag_.load()) {
        ret = stateActionEnginePtr_->DoStateAction(state_);
        auto it = stateTransitionRules_.at(state_).find(ret);
        if (it == stateTransitionRules_.at(state_).end()) {
            TTP_LOG_ERROR("No action defined for currentState and retCode! currentState: " <<
                StrStateOp(state_) << ", retCode: " << ret);
            ret = TTP_ERROR;
            break;
        }
        state_ = it->second;
        TTP_LOG_INFO("nextState: " << StrStateOp(state_));
    }
    TTP_LOG_INFO("controller StateMachine thread exit...");

    threadExistFlag_.store(false);
    return ret;
}

void controllerStateMachine::ControllerActionRegister(StateOp opcode, const StateActionFunc &func)
{
    stateActionEnginePtr_->ActionRegister(opcode, func);
}

void stateActionEngine::ActionRegister(StateOp opcode, const StateActionFunc &func)
{
    TTP_ASSERT_RET_VOID(opcode >= 0 && opcode < STATE_OP_BUTT);
    TTP_ASSERT_RET_VOID(func != nullptr);
    TTP_ASSERT_RET_VOID(actionHandle_[opcode] == nullptr);
    actionHandle_[opcode] = func;
}
TResult stateActionEngine::DoStateAction(StateOp opcode)
{
    if (opcode >= STATE_OP_BUTT) {
        TTP_LOG_ERROR("stateActionEngine: input opcode error! opcode: " << opcode);
        return TTP_ERROR;
    }

    TResult ret = TTP_OK;
    if (actionHandle_[opcode] != nullptr) {
        ret = static_cast<TResult>(actionHandle_[opcode]());
    } else {
        TTP_LOG_ERROR("stateActionEngine: no handle func! opcode: " << opcode);
        ret = TTP_ERROR;
    }
    return ret;
}

void controllerStateMachine::Stop()
{
    stopFlag_.store(true);
    while (threadExistFlag_.load()) {
        usleep(TTP_WAIT_TIME_1MS); // 1ms
    }
    TTP_LOG_INFO("controller StateMachine exit done... exit called by python when training done.");
}

}  // namespace ttp
}  // namespace ock