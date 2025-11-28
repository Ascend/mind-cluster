/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
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
 *
 * Author: bao
 */
#ifndef OCK_MEMFS_RING_BUFFER_H
#define OCK_MEMFS_RING_BUFFER_H

#include <semaphore.h>
#include <sstream>

#include "memfs_lock.h"

namespace ock {
namespace memfs {

template <typename T> class RingBuffer {
public:
    explicit RingBuffer(uint32_t capacity) : mCapacity(capacity) {}

    RingBuffer() = delete;

    ~RingBuffer()
    {
        UnInitialize();
    }

    inline int32_t Initialize()
    {
        if (mCapacity == 0) {
            return -1;
        }

        mCount = 0;
        mHead = 0;
        mTail = 0;

        if (mRingBuf != nullptr) {
            return 0;
        }

        mRingBuf = new (std::nothrow) T[mCapacity];
        if (mRingBuf == nullptr) {
            return -1;
        }

        return 0;
    }

    inline void UnInitialize()
    {
        if (mRingBuf == nullptr) {
            return;
        }

        delete[] mRingBuf;
        mRingBuf = nullptr;
    }

    inline bool PushBack(const T &item)
    {
        mLock.Lock();
        if (mCapacity <= mCount) {
            mLock.UnLock();
            return false;
        }

        mRingBuf[mTail] = item;
        if (mTail != mCapacity - 1) {
            ++mTail;
        } else {
            mTail = 0;
        }
        ++mCount;
        mLock.UnLock();
        return true;
    }

    inline bool PushFront(const T &item)
    {
        mLock.Lock();
        if (mCapacity <= mCount) {
            mLock.UnLock();
            return false;
        }

        /* move to tail */
        if (mHead == 0) {
            mHead = mCapacity - 1;
        } else {
            mHead--;
        }

        mRingBuf[mHead] = item;
        ++mCount;

        mLock.UnLock();
        return true;
    }

    inline bool PopFront(T &item)
    {
        mLock.Lock();
        if (mCount == 0) {
            mLock.UnLock();
            return false;
        }

        item = mRingBuf[mHead];
        if (mHead != mCapacity - 1) {
            ++mHead;
        } else {
            mHead = 0;
        }
        --mCount;
        mLock.UnLock();
        return true;
    }

    RingBuffer(const RingBuffer &) = delete;
    RingBuffer(RingBuffer &&) = delete;
    RingBuffer &operator = (const RingBuffer &) = delete;
    RingBuffer &operator = (RingBuffer &&) = delete;

private:
    T *mRingBuf = nullptr;
    SpinLock mLock;
    uint32_t mCapacity = 0;
    uint32_t mCount = 0;
    uint32_t mHead = 0;
    uint32_t mTail = 0;
};

template <typename T> class RingBufferBlockingQueue {
public:
    explicit RingBufferBlockingQueue(uint32_t capacity) : mRingBuffer(capacity) {}

    ~RingBufferBlockingQueue()
    {
        UnInitialize();
    }

    inline int Initialize()
    {
        if (sem_init(&mSem, 0, 0) != 0) {
            return -1;
        }

        return mRingBuffer.Initialize();
    }

    inline void UnInitialize()
    {
        mRingBuffer.UnInitialize();
        sem_destroy(&mSem);
    }

    inline bool Enqueue(const T &item)
    {
        auto result = mRingBuffer.PushBack(item);
        if (result) {
            sem_post(&mSem);
        }
        return result;
    }

    inline bool Enqueue(T &item)
    {
        auto result = mRingBuffer.PushBack(item);
        if (result) {
            sem_post(&mSem);
        }
        return result;
    }

    inline bool EnqueueFirst(T &item)
    {
        auto result = mRingBuffer.PushFront(item);
        if (result) {
            sem_post(&mSem);
        }
        return result;
    }

    inline bool EnqueueFirst(const T &item)
    {
        auto result = mRingBuffer.PushFront(item);
        if (result) {
            sem_post(&mSem);
        }
        return result;
    }

    inline bool Dequeue(T &item)
    {
        while (true) {
            auto result = mRingBuffer.PopFront(item);
            if (!result) {
                sem_wait(&mSem);
            } else {
                return result;
            }
        }
    }

private:
    RingBuffer<T> mRingBuffer; /* ring buffer to data store */
    sem_t mSem {};             /* semaphore to wait and notify */
};
}
}

#endif // OCK_MEMFS_RING_BUFFER_H
