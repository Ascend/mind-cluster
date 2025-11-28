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
#include "common_includes.h"
#include "memfs_event_manager.h"

using namespace ock::memfs;

MemfsEvent::MemfsEvent(std::string nm, uint64_t timeoutSec, uint64_t intervalMs) noexcept
    : name{ std::move(nm) },
      retryInterval{ std::chrono::milliseconds(intervalMs) },
      endTime{ std::chrono::steady_clock::now() + std::chrono::seconds(timeoutSec) }
{}

MemfsEventManager::MemfsEventManager() noexcept : timerExecutor{ "memfs_event" } {}

MemfsEventManager &MemfsEventManager::GetInstance() noexcept
{
    static MemfsEventManager instance;
    return instance;
}

int MemfsEventManager::SubmitEvent(const std::shared_ptr<MemfsEvent> &event, uint64_t &eventId) noexcept
{
    if (!event->PreCheckEvent()) {
        return -EINVAL;
    }

    event->eventId = eventIdGenerator.fetch_add(1UL);
    auto future = timerExecutor.Submit([this, event]() { ProcessEvent(event); }, event->retryInterval);
    if (future == nullptr) {
        LOG_ERROR("event(" << event->name << ":" << event->eventId << ") submit failed.");
        return -ENOMEM;
    }

    eventId = event->eventId;
    LOG_INFO("submit event(" << event->name << ") success with id:" << event->eventId);
    return 0;
}

void MemfsEventManager::WaitAllEventsFinished() noexcept
{
    timerExecutor.Wait();
}

void MemfsEventManager::ProcessEvent(const std::shared_ptr<MemfsEvent> &event) noexcept
{
    auto result = event->Process();
    if (result.finished) {
        LOG_INFO("event(" << event->name << ":" << event->eventId << ") times:" << event->retryCount <<
            " finished with " << result.message);
        event->Callback(event->eventId, result.result);
        return;
    }

    if (std::chrono::steady_clock::now() > event->endTime) {
        LOG_ERROR("event(" << event->name << ":" << event->eventId << ") times:" << event->retryCount <<
            " failed timeout");
        event->Callback(event->eventId, -ETIMEDOUT);
        return;
    }

    ++event->retryCount;
    auto future = timerExecutor.Submit([this, event]() { ProcessEvent(event); }, event->retryInterval);
    if (future == nullptr) {
        LOG_ERROR("event(" << event->name << ":" << event->eventId << ") times:" << event->retryCount <<
            " submit failed.");
        event->Callback(event->eventId, -ENOMEM);
    }
}