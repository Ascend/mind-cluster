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
#ifndef OCK_DFS_RETRY_TASK_POOL_H
#define OCK_DFS_RETRY_TASK_POOL_H

#include "thread_pool.h"
#include "timer_executor.h"

namespace ock {
namespace bg {
namespace util {
class RetryTaskPool {
public:
    struct RetryTaskConfig {
        std::string name;
        bool autoEvictFile { false };
        uint32_t thCnt = 1;
        uint32_t maxFailCntForUnserviceable { 0 };
        uint32_t retryTimes = 1;
        uint32_t retryIntervalSec = 1;
        uint64_t firstWaitMs = 0;
    };

    explicit RetryTaskPool(RetryTaskConfig &config) noexcept;
    virtual ~RetryTaskPool() noexcept;

public:
    int Start() noexcept;
    void Submit(const std::function<bool()> &task) noexcept;
    void ReportCCAE(bool serviceable) noexcept;
    inline bool IsTaskPoolEmpty() noexcept
    {
        return threadPool.IsTasksEmpty();
    }

private:
    std::string reportPath;
    ThreadPool threadPool;
    common::TimerExecutor timerExecutor;
    bool autoEvictFile { false };
    std::atomic<uint64_t> failCnt { 0 };
    uint32_t maxFailCntForUnserviceable { 0 };
    const uint32_t maxRetryTimes;
    std::chrono::milliseconds maxRetryIntervalMs;
    std::chrono::milliseconds firstWaitTime;
    friend class RetryTask;
};

class RetryTask {
public:
    explicit RetryTask(std::function<bool()> task, RetryTaskPool &pool) noexcept;

    static void Process(const std::shared_ptr<RetryTask> &task) noexcept;
    static void Waiting(const std::shared_ptr<RetryTask> &task) noexcept;

private:
    uint32_t retryTimes;
    std::function<bool()> realTask;
    RetryTaskPool &retryTaskPool;
};
}
}
}


#endif // OCK_DFS_RETRY_TASK_POOL_H
