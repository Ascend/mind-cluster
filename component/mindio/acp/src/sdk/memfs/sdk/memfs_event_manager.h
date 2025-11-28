/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#ifndef OCKIO_MEMFS_EVENT_MANAGER_H
#define OCKIO_MEMFS_EVENT_MANAGER_H

#include <cstdint>
#include <atomic>
#include <memory>
#include <chrono>
#include "non_copyable.h"
#include "timer_executor.h"

namespace ock {
namespace memfs {
struct EventResult {
    bool finished;
    int result;
    std::string message;
    EventResult(bool fini, int res, std::string msg) noexcept
        : finished{ fini }, result{ res }, message{ std::move(msg) }
    {}
    EventResult(bool fini, int res) noexcept : EventResult{ fini, res, "" } {}
};
class MemfsEvent {
public:
    MemfsEvent(std::string nm, uint64_t timeoutSec, uint64_t intervalMs) noexcept;
    virtual ~MemfsEvent() = default;
    virtual bool PreCheckEvent() noexcept
    {
        return true;
    }
    virtual EventResult Process() noexcept = 0;
    virtual void Callback(uint64_t eventId, int result) noexcept = 0;

    inline int GetErrorNumber() const noexcept
    {
        return errorNumber;
    }

    inline void SetErrorNumber(int error) noexcept
    {
        errorNumber = error;
    }

private:
    const std::string name;
    const std::chrono::milliseconds retryInterval;
    const std::chrono::steady_clock::time_point endTime;
    int errorNumber{ 0 };
    uint64_t retryCount{ 0UL };
    uint64_t eventId{ 0UL };
    friend class MemfsEventManager;
};

class MemfsEventManager : public common::NonCopyable {
public:
    static MemfsEventManager &GetInstance() noexcept;

public:
    int SubmitEvent(const std::shared_ptr<MemfsEvent> &event, uint64_t &eventId) noexcept;
    void WaitAllEventsFinished() noexcept;

private:
    MemfsEventManager() noexcept;
    void ProcessEvent(const std::shared_ptr<MemfsEvent> &event) noexcept;

private:
    common::TimerExecutor timerExecutor;
    std::atomic<uint64_t> eventIdGenerator{ 0UL };
};
}
}


#endif // OCKIO_MEMFS_EVENT_MANAGER_H
