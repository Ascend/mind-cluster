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
#ifndef IOFWD_THREAD_POOL_H
#define IOFWD_THREAD_POOL_H

#include <atomic>
#include <condition_variable>
#include <functional>
#include <list>
#include <memory>
#include <mutex>
#include <vector>
#include <thread>

#include "non_copyable.h"

namespace ock {
namespace bg {
namespace util {
class ThreadPool : public common::NonCopyable {
public:
    ThreadPool(int thCount, std::string name) noexcept;
    ~ThreadPool() noexcept override;

    int Start() noexcept;
    void Push(const std::function<void()> &task) noexcept;
    void Stop() noexcept;
    inline bool IsTasksEmpty()
    {
        std::unique_lock<std::mutex> lock(mtx);
        return tasks.empty() && unCompleteTaskCnt.load() == 0;
    }

private:
    void ThreadWork(int threadId) noexcept;

private:
    std::atomic<bool> running;
    std::atomic<uint64_t> unCompleteTaskCnt { 0UL };
    int threadNum;
    std::vector<std::shared_ptr<std::thread>> threads;
    std::list<std::function<void()>> tasks;
    std::string poolName;
    std::mutex mtx;
    std::condition_variable cond;
};
}
}
}

#endif // IOFWD_THREAD_POOL_H
