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
#ifndef OCK_TTP_MINDX_ENGINE_H
#define OCK_TTP_MINDX_ENGINE_H

#include <unordered_map>
#include <mutex>
#include <algorithm>
#include <map>
#include "common.h"

namespace ock {
namespace ttp {

enum class MindXEvent : uint32_t {
    MINDX_EVENT_REGISTER,
    MINDX_EVENT_REPORT_FAULT_RANKS,
    MINDX_EVENT_REPORT_STOP_COMPLETE,
    MINDX_EVENT_REPORT_STRATEGIES,
    MINDX_EVENT_REPORT_RESULT,
    MINDX_EVENT_STOP_TRAIN,
    MINDX_EVENT_PAUSE_TRAIN,
    MINDX_EVENT_CONTINUE_TRAIN,
    MINDX_EVENT_NOTIFY_FAULT_RANKS,
    MINDX_EVENT_HOT_SWITCH,
    MINDX_EVENT_STOP_SWITCH,
    MINDX_EVENT_MIGRATION,
    MINDX_EVENT_ARF,
    MINDX_EVENT_RETRY,
    MINDX_EVENT_DUMP,
    MINDX_EVENT_EXIT,
    MINDX_EVENT_INVALID,
    MINDX_EVENT_ELEGANT_DUMP,
    MINDX_EVENT_DOWNGRADE,
    MINDX_EVENT_UPGRADE,
    MINDX_EVENT_BUTT,
};

enum class RepairResult : int32_t {
    REPAIR_SUCCESS = 0,            ///< repair success
    CHOOSE_ZIT_RANK_SUCCESS = 201, ///< select zit rank success
    REPAIR_COMMON_ERROR = 499,     ///< only support exit strategy
    NO_PROCESSORS = 404,           ///< training processors has not been initialized
    FAILED_ALLOW_REOCVER = 405,    ///< switching from retry to recover strategy is supported
    FAILED_ALLOW_DUMP = 406,       ///< switching to the dump or exit strategy is supported
};

struct StopCompleteContext {
    int32_t code;
    std::string msg;
    std::map<int32_t, int32_t> errorInfoMap;
};

struct RecoverStrategyContext {
    std::map<int32_t, int32_t> errorInfoMap;
    std::vector<std::string> strategies;
};

struct RecoverStatusContext {
    int32_t code;
    std::string msg;
    std::string strategy;
    std::map<int32_t, int32_t> errorInfoMap;
};

struct ProcessFaultContext {
    std::map<int32_t, int32_t> errorInfoMap;
    std::map<int32_t, std::string> errorCodeMap;
};

using MindXEventHandle = std::function<int32_t(void *ctx, int ctxSize)>;

const std::string STRATEGY_DUMP = "dump";
const std::string STRATEGY_ARF  = "recover";
const std::string STRATEGY_RETRY  = "retry";
const std::string STRATEGY_EXIT  = "exit";
const std::string STRATEGY_PAUSE = "pause";
const std::string STRATEGY_CONTINUE = "continue";
const std::string STRATEGY_MIGRATION = "migration";
const std::string STRATEGY_DOWNGRADE = "downgrade";
const std::string STRATEGY_UPGRADE = "upgrade";

const std::string ACTION_HOT_SWITCH = "hot switch";
const std::string ACTION_STOP_SWITCH = "stop switch";

class MindXEngine : public Referable {
public:
    using Ptr = Ref<MindXEngine>;

    static Ptr GetInstance(bool destroy = false);

    // register event handler
    TResult RegisterEventHandler(MindXEvent event, MindXEventHandle handle);

    TResult ReportFaultRanks(std::map<int32_t, int32_t> &errors,  std::map<int32_t, std::string> &errorCodes,
                             ReadWriteLock &lock);

    TResult ReportStopComplete(RepairResult code, const std::string& msg,
                               std::map<int32_t, int32_t> &errors, ReadWriteLock &lock);

    TResult ReportStrategies(std::vector<std::string> &strategies,
                             std::map<int32_t, int32_t> &errors, ReadWriteLock &lock);

    TResult ReportResult(RepairResult code, const std::string& msg,
                         std::map<int32_t, int32_t> &errors, ReadWriteLock &lock);

    TResult WakeUp();

    TResult Wait(long time);

    void Register2MindX();

    void Destroy();

    TResult EventProcess(MindXEvent eventCode, void *ctx, int ctxSize);

    TResult PrepareAction(const std::string& action, std::map<int32_t, int32_t> &faultRanks);

    TResult ChangeStrategy(const std::string& strategy, std::string& param);

    TResult MindXNotifyDump();

    bool IsRegistered()
    {
        return isRegistered_.load();
    }

private:
    std::atomic<bool> isRegistered_{false};
    MindXEventHandle eventHandleList_[static_cast<int>(MindXEvent::MINDX_EVENT_BUTT)]{};
    std::vector<std::string> repairStrategies_;
    std::string lastStrategy_;
    std::mutex destroyMutex_;
    PthreadTimedwait waiter_;
};

using MindXEnginePtr = MindXEngine::Ptr;

}
}
#endif // OCK_TTP_MINDX_ENGINE_H
