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

#ifndef MINDIO_TTP_CONTROLLERSTATEMACHINE_H
#define MINDIO_TTP_CONTROLLERSTATEMACHINE_H

#include <unordered_map>
#include <thread>
#include "common.h"

namespace ock {
namespace ttp {

using StateActionFunc = std::function<int32_t()>;

enum StateOp : uint16_t {
    STATE_OP_INIT,
    STATE_OP_NORMAL,
    STATE_OP_PAUSE,
    STATE_OP_STEP_FINISH,
    STATE_OP_MIGRATION,
    STATE_OP_ENV_CLEAR,
    STATE_OP_REPAIR,
    STATE_OP_ABNORMAL,
    STATE_OP_DG_REPAIR,
    STATE_OP_DOWNGRADE,
    STATE_OP_UPGRADE,
    STATE_OP_DUMP,
    STATE_OP_EXIT,
    STATE_OP_FINAL,
    STATE_OP_BUTT,
};

inline const char *StrStateOp(uint16_t state)
{
    static std::vector<std::string_view> stateOps = {
        "STATE_OP_INIT",
        "STATE_OP_NORMAL",
        "STATE_OP_PAUSE",
        "STATE_OP_STEP_FINISH",
        "STATE_OP_MIGRATION",
        "STATE_OP_ENV_CLEAR",
        "STATE_OP_REPAIR",
        "STATE_OP_ABNORMAL",
        "STATE_OP_DG_REPAIR",
        "STATE_OP_DOWNGRADE",
        "STATE_OP_UPGRADE",
        "STATE_OP_DUMP",
        "STATE_OP_EXIT",
        "STATE_OP_FINAL",
        "STATE_OP_BUTT",
    };

    if (state >= stateOps.size()) {
        return "";
    }

    return stateOps[state].data();
}

class stateActionEngine : public Referable {
public:
    stateActionEngine() = default;
    ~stateActionEngine() override = default;

    TResult DoStateAction(StateOp opcode);
    void ActionRegister(StateOp opcode, const StateActionFunc &func);
private:
    StateActionFunc actionHandle_[STATE_OP_BUTT]{};
};

using stateActionEnginePtr = Ref<stateActionEngine>;

class controllerStateMachine : public Referable {
public:
    controllerStateMachine() = default;
    ~controllerStateMachine() override = default;

    void StartStateMachine();
    TResult Initialize(int32_t rank);
    StateOp GetCurrentState(){return state_;}
    void ControllerActionRegister(StateOp opcode, const StateActionFunc &func);
    void Stop();
private:
    TResult CheckAndUpdateState();
    std::atomic<bool> stopFlag_ = {false };
    std::atomic<bool> threadExistFlag_ = {false };
    StateOp state_ = STATE_OP_INIT;
    stateActionEnginePtr stateActionEnginePtr_ = nullptr;
    std::unordered_map<StateOp, std::unordered_map<int32_t, StateOp>> stateTransitionRules_;
    std::thread stateChecker_;
};

}  // namespace ttp
}  // namespace ock

#endif  // MINDIO_TTP_CONTROLLERSTATEMACHINE_H
