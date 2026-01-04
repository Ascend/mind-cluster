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
#include "mindx_engine.h"

namespace ock {
namespace ttp {

MindXEnginePtr MindXEngine::GetInstance(bool destroy)
{
    static std::mutex gMutex;
    static MindXEnginePtr gInstance;

    if (gInstance == nullptr) {
        std::lock_guard<std::mutex> guard(gMutex);
        if (gInstance.Get() == nullptr) {
            gInstance = MakeRef<MindXEngine>();
            if (gInstance == nullptr) {
                TTP_LOG_ERROR("Create MindxEngine failed,out of memory");
                throw std::bad_alloc();
            }
        }
    } else if (destroy) {
        std::lock_guard<std::mutex> guard(gMutex);
        gInstance = nullptr;
    }

    return gInstance;
}

void MindXEngine::Register2MindX()
{
    auto ret = EventProcess(MindXEvent::MINDX_EVENT_REGISTER, nullptr, 0);
    TTP_RET_LOG(ret, "check mindx callback all register, ret:" << ret);
    isRegistered_.store(ret == TTP_OK);

    ret = waiter_.Initialize();
    TTP_RET_LOG(ret, "MindXEngine init waiter ret:" << ret);
}

TResult MindXEngine::ReportFaultRanks(std::map<int32_t, int32_t> &errors, std::map<int32_t, std::string> &errorCodes,
                                      ReadWriteLock &lock)
{
    if (!isRegistered_.load()) {
        TTP_LOG_WARN("mindx no register, skip report fault ranks!");
        return TTP_ERROR;
    }

    lock.LockRead();
    ProcessFaultContext nrsc {errors, errorCodes};
    lock.UnLock();

    auto ret = EventProcess(MindXEvent::MINDX_EVENT_REPORT_FAULT_RANKS, &nrsc, sizeof(ProcessFaultContext));
    TTP_RET_LOG(ret, "report fault ranks action ret:" << ret);

    return ret;
}

TResult MindXEngine::ReportStopComplete(RepairResult code, const std::string &msg,
                                        std::map<int32_t, int32_t> &errors, ReadWriteLock &lock)
{
    if (!isRegistered_.load()) {
        TTP_LOG_WARN("mindx no register, skip report stop complete!");
        return TTP_OK;
    }

    lock.LockRead();
    StopCompleteContext nrsc {static_cast<int32_t>(code), msg, errors};
    lock.UnLock();

    waiter_.SignalClean();

    auto ret = EventProcess(MindXEvent::MINDX_EVENT_REPORT_STOP_COMPLETE, &nrsc, sizeof(StopCompleteContext));
    TTP_RET_LOG(ret, "report stop complete action ret:" << ret);

    return ret;
}

TResult MindXEngine::ReportStrategies(std::vector<std::string> &strategies,
    std::map<int32_t, int32_t> &errors, ReadWriteLock &lock)
{
    if (!isRegistered_.load()) {
        TTP_LOG_WARN("mindx no register, skip report strategies!");
        return TTP_OK;
    }

    repairStrategies_ = strategies;

    lock.LockRead();
    RecoverStrategyContext nrsc {errors, repairStrategies_};
    lock.UnLock();

    waiter_.SignalClean();

    auto ret = EventProcess(MindXEvent::MINDX_EVENT_REPORT_STRATEGIES, &nrsc, sizeof(RecoverStrategyContext));
    TTP_RET_LOG(ret, "report strategies action ret:" << ret);

    return ret;
}

TResult MindXEngine::ReportResult(RepairResult code, const std::string& msg,
    std::map<int32_t, int32_t> &errors, ReadWriteLock &lock)
{
    if (!isRegistered_.load()) {
        TTP_LOG_WARN("mindx no register, skip report result!");
        return TTP_OK;
    }

    lock.LockRead();
    RecoverStatusContext nrsc {static_cast<int32_t>(code), msg, lastStrategy_, errors};
    lock.UnLock();

    waiter_.SignalClean();

    auto ret = EventProcess(MindXEvent::MINDX_EVENT_REPORT_RESULT, &nrsc, sizeof(RecoverStatusContext));
    TTP_RET_LOG(ret, "report result action ret:" << ret);

    if (code == RepairResult::REPAIR_SUCCESS) {
        lastStrategy_.clear();
        repairStrategies_.clear();
    }

    return ret;
}

TResult MindXEngine::WakeUp()
{
    if (!isRegistered_.load()) {
        return TTP_OK;
    }

    auto ret = waiter_.PthreadSignal();

    return ret == TTP_OK ? TTP_OK : TTP_ERROR;
}

TResult MindXEngine::Wait(long time)
{
    if (!isRegistered_.load()) {
        return TTP_OK;
    }

    auto ret = waiter_.PthreadTimedwaitSecs(time);
    TTP_RET_LOG(ret, "wait mindx wake up action ret:" << ret);

    return ret == TTP_OK ? TTP_OK : TTP_ERROR;
}

TResult MindXEngine::RegisterEventHandler(MindXEvent event, MindXEventHandle handle)
{
    TTP_ASSERT_RETURN(event < MindXEvent::MINDX_EVENT_BUTT, TTP_ERROR);
    TTP_ASSERT_RETURN(handle != nullptr, TTP_ERROR);
    TTP_ASSERT_RETURN(eventHandleList_[static_cast<int>(event)] == nullptr, TTP_ERROR);
    eventHandleList_[static_cast<int>(event)] = handle;
    return TTP_OK;
}

TResult MindXEngine::EventProcess(MindXEvent eventCode, void *ctx, int ctxSize)
{
    TTP_ASSERT_RETURN(eventCode < MindXEvent::MINDX_EVENT_BUTT, TTP_ERROR);
    auto code = static_cast<uint32_t>(eventCode);
    if (eventHandleList_[code] == nullptr) {
        TTP_LOG_WARN("event handle is null! event: " << code);
        return TTP_ERROR;
    }
    return static_cast<TResult>(eventHandleList_[code](ctx, ctxSize));
}

TResult MindXEngine::PrepareAction(const std::string& action, std::map<int32_t, int32_t> &faultRanks)
{
    if (action == ACTION_HOT_SWITCH) {
        TTP_LOG_INFO("Mindx calling notify do HOT switch...");
        return EventProcess(MindXEvent::MINDX_EVENT_HOT_SWITCH, &faultRanks, faultRanks.size());
    } else if (action == ACTION_STOP_SWITCH) {
        TTP_LOG_INFO("Mindx calling notify do STOP switch...");
        return EventProcess(MindXEvent::MINDX_EVENT_STOP_SWITCH, nullptr, 0);
    } else {
        TTP_LOG_WARN("prepare action:" << action << " not supported!");
        return TTP_ERROR;
    }
}

TResult MindXEngine::ChangeStrategy(const std::string& strategy, std::string& param)
{
    TResult ret = TTP_ERROR;
#ifndef UT_ENABLED  // UT Test skip invalid strategy verify
    if (!repairStrategies_.empty()) {
        auto it = std::find(repairStrategies_.begin(), repairStrategies_.end(), strategy);
        if (it == repairStrategies_.end()) {
            TTP_LOG_ERROR("receive strategy: " << strategy << " is not valid!");
            return EventProcess(MindXEvent::MINDX_EVENT_INVALID, nullptr, 0);
        }
    }
#endif
    lastStrategy_ = strategy;
    if (strategy == STRATEGY_RETRY) {
        TTP_LOG_INFO("Mindx calling notify do retry repair...");
        return EventProcess(MindXEvent::MINDX_EVENT_RETRY, nullptr, 0);
    } else if (strategy == STRATEGY_ARF) {
        TTP_LOG_INFO("Mindx calling notify do ARF repair...");
        return EventProcess(MindXEvent::MINDX_EVENT_ARF, nullptr, 0);
    } else if (strategy == STRATEGY_DOWNGRADE) {
        TTP_LOG_INFO("Mindx calling notify do Downgrade repair...");
        return EventProcess(MindXEvent::MINDX_EVENT_DOWNGRADE, &param, param.length());
    } else if (strategy == STRATEGY_UPGRADE) {
        TTP_LOG_INFO("Mindx calling notify do upgrade repair...");
        return EventProcess(MindXEvent::MINDX_EVENT_UPGRADE, &param, param.length());
    } else if (strategy == STRATEGY_DUMP) {
        TTP_LOG_INFO("Mindx calling notify do DUMP ...");
        return EventProcess(MindXEvent::MINDX_EVENT_DUMP, nullptr, 0);
    } else if (strategy == STRATEGY_EXIT) {
        TTP_LOG_INFO("Mindx calling notify do EXIT ...");
        return EventProcess(MindXEvent::MINDX_EVENT_EXIT, nullptr, 0);
    } else if (strategy == STRATEGY_MIGRATION) {
        TTP_LOG_INFO("Mindx calling notify do Migration ...");
        return EventProcess(MindXEvent::MINDX_EVENT_MIGRATION, nullptr, 0);
    } else if (strategy == STRATEGY_CONTINUE) {
        TTP_LOG_INFO("Mindx calling notify do Continue ...");
        return EventProcess(MindXEvent::MINDX_EVENT_CONTINUE_TRAIN, nullptr, 0);
    } else {
        TTP_LOG_WARN("repair strategy:" << strategy << " not supported!");
        return ret;
    }
}

TResult MindXEngine::MindXNotifyDump()
{
    lastStrategy_ = STRATEGY_DUMP;
    return EventProcess(MindXEvent::MINDX_EVENT_ELEGANT_DUMP, nullptr, 0);
}

void MindXEngine::Destroy()
{
    std::lock_guard<std::mutex> guard(destroyMutex_);
    isRegistered_.store(false);
    std::fill(std::begin(eventHandleList_), std::end(eventHandleList_), nullptr);
}

}
}