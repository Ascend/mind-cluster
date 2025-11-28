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
#ifndef IOFWD_TIMER_EXECUTOR_H
#define IOFWD_TIMER_EXECUTOR_H

#include <string>
#include <map>
#include <functional>
#include <unordered_map>
#include <atomic>
#include <list>
#include <thread>
#include <mutex>
#include <condition_variable>

namespace ock {
namespace common {
/*
 * A Future represents the result of an asynchronous computation.
 * Methods are provided to check if the computation is complete.
 * Cancellation is performed by the cancel method.
 * Additional methods are provided to determine if the task completed normally or was cancelled.
 */
class Future {
public:
    virtual ~Future() = default;

public:
    /*
     * Returns true if this task completed.
     */
    virtual bool Finished() noexcept = 0;

    /*
     * Attempts to cancel execution of this task.
     */
    virtual bool Cancel() noexcept = 0;

    /*
     * Returns true if this task was cancelled before it completed normally.
     */
    virtual bool IsCancelled() noexcept = 0;
};

struct RunTask {
    const uint64_t taskId;
    const uint64_t timeoutNs;
    std::function<void()> task;
    std::shared_ptr<Future> future;

    RunTask(uint64_t id, uint64_t ns, std::function<void()> t) noexcept
        : taskId{ id }, timeoutNs{ ns }, task{ std::move(t) }
    {}
};

struct TaskComparator {
    bool operator () (const std::pair<uint64_t, uint64_t> &k1, const std::pair<uint64_t, uint64_t> &k2) const
    {
        return (k1.first < k2.first) || (k1.first == k2.first && k1.second < k2.second);
    }
};

class TimerExecutor {
    friend class TaskFuture;

public:
    explicit TimerExecutor(std::string nm) noexcept;
    virtual ~TimerExecutor() noexcept;

    void Shutdown() noexcept;
    void Wait() noexcept;

public:
    /* *
     * @brief Creates and executes an action that becomes enabled after the given time delay.
     * @param task the task to execute
     * @param time the time to delay execution
     * @return a Future representing pending completion of the task.
     */
    template <class R, class P>
    std::shared_ptr<Future> Submit(const std::function<void()> &task, const std::chrono::duration<R, P> &time) noexcept
    {
        auto ns = std::chrono::duration_cast<std::chrono::nanoseconds>(time);
        return Submit(task, ns.count());
    }

    /* *
     * @brief Creates and executes an action that becomes enabled after the given time delay.
     * @param task the task to execute
     * @param ns the time to delay execution, specify count with nanoseconds.
     * @return a Future representing pending completion of the task.
     */
    std::shared_ptr<Future> Submit(const std::function<void()> &task, int64_t ns) noexcept;

private:
    RunTask *RemoveWaitingTask(uint64_t taskId) noexcept;
    void ThreadRunning() noexcept;
    RunTask *GetOutOneTask() noexcept;

private:
    const std::string name;
    std::atomic<bool> destroyed{ false };
    std::atomic<uint64_t> idGen{ 0UL };
    std::map<std::pair<int64_t, uint64_t>, RunTask *, TaskComparator> timeOrderTasks;
    std::unordered_map<uint64_t, RunTask *> hashByIdTasks;
    std::mutex mutex;
    std::condition_variable cond;
    std::shared_ptr<std::thread> thread;
    static std::atomic<uint64_t> leftTaskCnt;
};

/*
 * Implements for Future
 */
class TaskFuture : public Future {
public:
    TaskFuture(TimerExecutor *executor, uint64_t id) noexcept;
    void SetStarted() noexcept;
    void SetFinished() noexcept;
    void SetCancelled() noexcept;

public:
    bool Finished() noexcept override;
    bool Cancel() noexcept override;
    bool IsCancelled() noexcept override;

private:
    TimerExecutor *timerExecutor;
    uint64_t taskId;
    std::atomic<bool> cancelled{ false };
    std::atomic<bool> started{ false };
    std::atomic<bool> finished{ false };
};
}
}

#endif // IOFWD_TIMER_EXECUTOR_H
