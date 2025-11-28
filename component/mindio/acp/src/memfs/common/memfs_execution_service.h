/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
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
#ifndef OCK_MEMFS_EXECUTION_SERVICE_H
#define OCK_MEMFS_EXECUTION_SERVICE_H
#include <atomic>
#include <mutex>
#include <thread>
#include <unistd.h>
#include <vector>
#include <functional>

#include "memfs_common.h"
#include "memfs_ring_buffer.h"
#include "memfs_out_logger.h"

namespace ock {
namespace memfs {
enum RunnableType {
    NORMAL = 0,
    STOP = 1,
};

class Runnable {
public:
    Runnable() : mTask{ nullptr } {}

    explicit Runnable(const std::function<void()> &task) : mTask{ task } {}
    virtual ~Runnable() = default;

    virtual void Run()
    {
        if (mTask != nullptr) {
            mTask();
        }
    }
private:
    inline void Type(RunnableType type)
    {
        mType = type;
    }

    inline RunnableType Type() const
    {
        return mType;
    }

private:
    RunnableType mType = RunnableType::NORMAL;
    std::function<void()> mTask;

    friend class ExecutorService;
};
using RunnablePtr = std::shared_ptr<Runnable>;

constexpr uint32_t ES_MAX_THR_NUM = 256;

class ExecutorService;
using ExecutorServicePtr = std::shared_ptr<ExecutorService>;

class ExecutorService {
public:
    static ExecutorServicePtr Create(uint16_t threadNum, uint32_t queueCapacity = 10000)
    {
        if (threadNum > ES_MAX_THR_NUM || threadNum == 0) {
            DAGGER_LOG_ERROR(ExecSer, "The num of thread must 1-" << ES_MAX_THR_NUM);
            return nullptr;
        }

        ExecutorService* rawPtr = new (std::nothrow) ExecutorService(threadNum, queueCapacity);
        if (!rawPtr) {
            DAGGER_LOG_ERROR(ExecSer, "ExecutorServicePtr create fault");
            return nullptr;
        }
        return std::shared_ptr<ExecutorService>(rawPtr);
    }

public:
    ~ExecutorService()
    {
        if (!mStopped) {
            Stop();
        }

        while (!mThreads.empty()) {
            delete (mThreads.back());
            mThreads.pop_back();
        }
    }

    bool Start();

    void Stop();

    inline bool Execute(const RunnablePtr &runnable)
    {
        if (UNLIKELY(runnable.get() == nullptr)) {
            return false;
        }

        return mRunnableQueue.Enqueue(runnable);
    }

    inline bool Execute(const std::function<void()> &task)
    {
        return Execute(MakeRef<Runnable>(task));
    }

private:
    ExecutorService(uint16_t threadNum, uint32_t queueCapacity)
        : mRunnableQueue(queueCapacity),
          mThreadNum(threadNum),
          mThreads(0),
          mStarted(false),
          mStopped(false),
          mStartedThreadNum(0)
    {}

    void RunInThread(int16_t cpuId);
    void DoRunnable(bool &flag);

private:
    RingBufferBlockingQueue<RunnablePtr> mRunnableQueue;
    uint16_t mThreadNum = 0;
    int16_t mCpuSetStartIdx = -1;
    std::vector<std::thread *> mThreads;

    std::atomic<bool> mStarted;
    std::atomic<bool> mStopped;
    std::atomic<uint16_t> mStartedThreadNum;

    std::string mThreadName;
};

inline bool ExecutorService::Start()
{
    if (mStarted) {
        return true;
    }

    /* init ring buffer blocking queue */
    auto result = mRunnableQueue.Initialize();
    if (result != 0) {
        DAGGER_LOG_ERROR(ExecSer, "Failed to initialize queue, result " << result);
        return false;
    }

    for (uint16_t i = 0; i < mThreadNum; i++) {
        auto cpuId = mCpuSetStartIdx < 0 ? -1 : mCpuSetStartIdx + i;
        auto *thr = new (std::nothrow) std::thread(&ExecutorService::RunInThread, this, cpuId);
        if (thr == nullptr) {
            DAGGER_LOG_ERROR(ExecSer, "Failed to create executor thread " << i);
            return false;
        }

        mThreads.push_back(thr);
    }

    while (mStartedThreadNum < mThreadNum) {
        usleep(1);
    }

    mStarted = true;
    return true;
}

inline void ExecutorService::Stop()
{
    if (!mStarted || mStopped) {
        return;
    }

    for (uint32_t i = 0; i < mThreads.size(); ++i) {
        RunnablePtr stopTask = std::make_shared<Runnable>();
        if (stopTask == nullptr) {
            DAGGER_LOG_ERROR(ExecSer, "Failed to new stop task, probably out of memory");
            break;
        }
        stopTask->Type(RunnableType::STOP);

        if (!mRunnableQueue.EnqueueFirst(stopTask)) {
            continue;
        }
    }

    for (auto &thr : mThreads) {
        if (thr != nullptr) {
            thr->join();
        }
    }

    mStopped = true;
    mRunnableQueue.UnInitialize();
}

inline void ExecutorService::DoRunnable(bool &flag)
{
    try {
        RunnablePtr runnable = nullptr;
        mRunnableQueue.Dequeue(runnable);
        if (runnable.get() != nullptr) {
            if (runnable->Type() == RunnableType::NORMAL) {
                runnable->Run();
            } else if (runnable->Type() == RunnableType::STOP) {
                flag = false;
            } else {
                DAGGER_LOG_ERROR(ExecSer, "Un-reachable path");
            }
        } else {
            DAGGER_LOG_ERROR(ExecSer, "Task is null");
        }
    } catch (std::runtime_error &ex) {
        DAGGER_LOG_ERROR(ExecSer, "Caught error " << ex.what() << " when execute a task, continue");
    } catch (...) {
        DAGGER_LOG_ERROR(ExecSer, "Caught unknown error when execute a task, continue");
    }
}

inline void ExecutorService::RunInThread(int16_t cpuId)
{
    bool runFlag = true;
    uint16_t threadIndex = mStartedThreadNum++;

    auto threadName = mThreadName.empty() ? "executor" : mThreadName;
    threadName += std::to_string(threadIndex);
    if (cpuId != -1) {
        cpu_set_t cpuSet;
        CPU_ZERO(&cpuSet);
        CPU_SET(cpuId, &cpuSet);
        if (pthread_setaffinity_np(pthread_self(), sizeof(cpuSet), &cpuSet) != 0) {
            DAGGER_LOG_WARN(ExecSer, "Failed to bind executor thread" << threadName << " << to cpu " << cpuId);
        }
    }

    pthread_setname_np(pthread_self(), threadName.c_str());
    DAGGER_LOG_INFO(ExecSer, "Thread is started for executor service <" << threadName << "> cpuId " << cpuId);

    while (runFlag) {
        DoRunnable(runFlag);
    }
    DAGGER_LOG_INFO(ExecSer, "Thread for executor service <" << threadName << "> cpuId " << cpuId << " exiting");
}

}
}

#endif // OCK_MEMFS_EXECUTION_SERVICE_H