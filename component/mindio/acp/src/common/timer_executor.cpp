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
#include "timer_executor.h"

using namespace ock::common;

static constexpr auto TH_NAME_PREFIX_LEN = 12U;
static constexpr auto MAX_WAIT_TIME = 300;

std::atomic<uint64_t> TimerExecutor::leftTaskCnt{ 0UL };

TimerExecutor::TimerExecutor(std::string nm) noexcept : name{ std::move(nm) } {}

TimerExecutor::~TimerExecutor() noexcept
{
    try {
        Shutdown();
    } catch (const std::exception &e) {
        LOG_ERROR("MindIO BG Timer shutdown failed: " << e.what());
    }
}

void TimerExecutor::Shutdown() noexcept
{
    bool expected = false;
    if (!destroyed.compare_exchange_strong(expected, true)) {
        LOG_INFO("timer(" << name.c_str() << ") already shutdown, no need shutdown again.");
        return;
    }

    if (thread == nullptr) {
        return;
    }

    LOG_DEBUG("shutdown for executor(" << name.c_str() << ")");
    {
        std::unique_lock<std::mutex> lockGuard(mutex);
        cond.notify_one();
    }
    thread->join();
    thread.reset();

    for (auto &timeOrderTask : timeOrderTasks) {
        auto task = timeOrderTask.second;
        if (task != nullptr) {
            auto future = dynamic_cast<TaskFuture *>(task->future.get());
            if (future != nullptr) {
                future->SetCancelled();
                leftTaskCnt.fetch_sub(1U);
            }
            delete task;
            task = nullptr;
        }
    }

    timeOrderTasks.clear();
    hashByIdTasks.clear();
    LOG_INFO("timer(" << name.c_str() << ") success shutdown.");
}

void TimerExecutor::Wait() noexcept
{
    auto start = std::chrono::high_resolution_clock::now();
    while (leftTaskCnt.load() != 0) {
        auto cur = std::chrono::high_resolution_clock::now();
        if (std::chrono::duration_cast<std::chrono::seconds>(cur - start).count() > MAX_WAIT_TIME) {
            LOG_WARN("Wait for timerExecutor task exec timeout.");
            return;
        }
    }
}

std::shared_ptr<Future> TimerExecutor::Submit(const std::function<void()> &task, int64_t ns) noexcept
{
    if (destroyed.load()) {
        return nullptr;
    }

    auto timePoint = std::chrono::steady_clock::now() + std::chrono::nanoseconds(ns);
    auto timePointValue = std::chrono::duration_cast<std::chrono::nanoseconds>(timePoint.time_since_epoch()).count();

    auto id = idGen.fetch_add(1UL);
    auto runTask = new (std::nothrow) RunTask(id, timePointValue, task);
    if (runTask == nullptr) {
        LOG_ERROR("timer(" << name.c_str() << ") allocate new task failed.");
        return nullptr;
    }

    auto future = std::make_shared<TaskFuture>(this, id);
    runTask->future = future;

    LOG_DEBUG("executor(" << name.c_str() << ") submit task(" << id << ") with ns(" << ns << ")");
    std::unique_lock<std::mutex> lockGuard(mutex);
    leftTaskCnt.fetch_add(1U);
    timeOrderTasks[std::make_pair(timePointValue, id)] = runTask;
    hashByIdTasks[id] = runTask;
    if (thread == nullptr) {
        thread = std::make_shared<std::thread>(&TimerExecutor::ThreadRunning, this);
    }
    cond.notify_one();

    return future;
}

RunTask *TimerExecutor::RemoveWaitingTask(uint64_t taskId) noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    auto pos = hashByIdTasks.find(taskId);
    if (pos == hashByIdTasks.end()) {
        return nullptr;
    }

    auto task = pos->second;
    hashByIdTasks.erase(pos);
    if (task != nullptr) {
        timeOrderTasks.erase(std::make_pair(task->timeoutNs, task->taskId));
    }

    return task;
}

void TimerExecutor::ThreadRunning() noexcept
{
    std::string thName = name.size() > TH_NAME_PREFIX_LEN ? name.substr(0, TH_NAME_PREFIX_LEN) : name;
    thName.append("_TM");
    pthread_setname_np(pthread_self(), thName.c_str());

    while (!destroyed.load()) {
        auto task = GetOutOneTask();
        if (task == nullptr) {
            auto timeout = std::chrono::steady_clock::now() + std::chrono::hours(1);
            std::unique_lock<std::mutex> lockGuard(mutex);
            if (!destroyed.load()) {
                cond.wait_until(lockGuard, timeout);
            }
            continue;
        }

        auto future = dynamic_cast<TaskFuture *>(task->future.get());
        if (future != nullptr) {
            future->SetStarted();
            task->task();
            future->SetFinished();
            leftTaskCnt.fetch_sub(1U);
        }

        delete task;
        task = nullptr;
    }
}

RunTask *TimerExecutor::GetOutOneTask() noexcept
{
    std::unique_lock<std::mutex> lockGuard(mutex);
    auto pos = timeOrderTasks.end();
    RunTask *task;
    while (!destroyed.load() && !timeOrderTasks.empty()) {
        pos = timeOrderTasks.begin();
        task = pos->second;
        if (task != nullptr) {
            auto timeout = std::chrono::steady_clock::time_point(std::chrono::nanoseconds(task->timeoutNs));
            if (std::chrono::steady_clock::now() >= timeout) {
                break;
            }

            LOG_DEBUG("executor(" << name.c_str() << ") task(" << task->taskId << ") should delay.");
            cond.wait_until(lockGuard, timeout);
            pos = timeOrderTasks.begin();
        }
    }

    if (destroyed.load()) {
        LOG_DEBUG("executor(" << name.c_str() << ") already destroyed.");
        return nullptr;
    }

    if (pos == timeOrderTasks.end()) {
        LOG_DEBUG("executor(" << name.c_str() << ") empty, no task now.");
        return nullptr;
    }

    task = pos->second;
    timeOrderTasks.erase(pos);
    if (task != nullptr) {
        hashByIdTasks.erase(task->taskId);
    }

    return task;
}

TaskFuture::TaskFuture(TimerExecutor *executor, uint64_t id) noexcept : timerExecutor{ executor }, taskId{ id } {}

void TaskFuture::SetStarted() noexcept
{
    started.store(true);
}

void TaskFuture::SetFinished() noexcept
{
    finished.store(true);
    timerExecutor = nullptr;
}

void TaskFuture::SetCancelled() noexcept
{
    cancelled.store(true);
    timerExecutor = nullptr;
}

bool TaskFuture::Finished() noexcept
{
    return finished.load();
}

bool TaskFuture::Cancel() noexcept
{
    if (timerExecutor == nullptr) {
        return false;
    }

    auto task = timerExecutor->RemoveWaitingTask(taskId);
    if (task == nullptr) {
        timerExecutor = nullptr;
        return false;
    }

    timerExecutor = nullptr;
    cancelled.store(true);
    finished.store(true);
    delete task;
    task = nullptr;

    return true;
}

bool TaskFuture::IsCancelled() noexcept
{
    return cancelled.load();
}