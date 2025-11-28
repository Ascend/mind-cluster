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
#include <algorithm>
#include <string>

#include "background_log.h"
#include "thread_pool.h"

using namespace ock::bg::util;

static constexpr auto TH_NAME_PREFIX_LEN = 10U;

ThreadPool::ThreadPool(int thCount, std::string name) noexcept
    : running(false), threadNum{ thCount }, poolName{ std::move(name) }
{}

ThreadPool::~ThreadPool() noexcept
{
    if (running) {
        Stop();
        running.store(false);
    }
}

int ThreadPool::Start() noexcept
{
    int failedCount = 0;
    running.store(true);
    for (int i = 0; i < threadNum; ++i) {
        try {
            threads.push_back(std::make_shared<std::thread>(&ThreadPool::ThreadWork, this, i));
        } catch (const std::exception &ex) {
            failedCount++;
            BKG_LOG_ERROR("Exception: (" << ex.what() << ")");
        } catch (...) {
            failedCount++;
            BKG_LOG_ERROR("Uncatch Exception");
        }
    }

    if (failedCount > 0) {
        Stop();
        return -1;
    }

    return 0;
}

void ThreadPool::Push(const std::function<void()> &task) noexcept
{
    std::unique_lock<std::mutex> lock(mtx);
    tasks.push_back(task);
    lock.unlock();
    unCompleteTaskCnt.fetch_add(1);
    cond.notify_one();
}

void ThreadPool::Stop() noexcept
{
    running.store(false);
    cond.notify_all();
    for (auto &th : threads) {
        if (th != nullptr) {
            th->join();
        }
    }
}

void ThreadPool::ThreadWork(int threadId) noexcept
{
    std::string name = poolName.size() > TH_NAME_PREFIX_LEN ? poolName.substr(0, TH_NAME_PREFIX_LEN) : poolName;
    name.append(std::to_string(threadId));
    pthread_setname_np(pthread_self(), name.c_str());

    while (running.load()) {
        std::function<void()> task;
        std::unique_lock<std::mutex> lock(mtx);
        while (tasks.empty() && running) {
            cond.wait(lock);
        }

        if (!tasks.empty()) {
            task = tasks.front();
            tasks.pop_front();
        }
        lock.unlock();

        if (task) {
            task();
        }
        unCompleteTaskCnt.fetch_sub(1);
    }
}