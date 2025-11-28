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
#ifndef OCK_TTP_LOCKS_H
#define OCK_TTP_LOCKS_H

#include <mutex>
#include <atomic>
#include <pthread.h>

namespace ock {
namespace ttp {
class Lock {
public:
    Lock() = default;
    ~Lock() = default;

    Lock(const Lock &) = delete;
    Lock &operator = (const Lock &) = delete;
    Lock(Lock &&) = delete;
    Lock &operator = (Lock &&) = delete;

    inline void DoLock()
    {
        mLock.lock();
    }

    inline void UnLock()
    {
        mLock.unlock();
    }

private:
    std::mutex mLock;
};

class RecursiveLock {
public:
    RecursiveLock() = default;
    ~RecursiveLock() = default;

    RecursiveLock(const RecursiveLock &) = delete;
    RecursiveLock &operator = (const RecursiveLock &) = delete;
    RecursiveLock(RecursiveLock &&) = delete;
    RecursiveLock &operator = (RecursiveLock &&) = delete;

    inline void DoLock()
    {
        mLock.lock();
    }
    inline void UnLock()
    {
        mLock.unlock();
    }

private:
    std::recursive_mutex mLock;
};

class ReadWriteLock {
public:
    ReadWriteLock()
    {
        pthread_rwlock_init(&mLock, nullptr);
    }
    ~ReadWriteLock()
    {
        pthread_rwlock_destroy(&mLock);
    }

    ReadWriteLock(const ReadWriteLock &) = delete;
    ReadWriteLock &operator = (const ReadWriteLock &) = delete;
    ReadWriteLock(ReadWriteLock &&) = delete;
    ReadWriteLock &operator = (ReadWriteLock &&) = delete;

    inline void LockRead()
    {
        pthread_rwlock_rdlock(&mLock);
    }

    inline void LockWrite()
    {
        pthread_rwlock_wrlock(&mLock);
    }

    inline void UnLock()
    {
        pthread_rwlock_unlock(&mLock);
    }

private:
    pthread_rwlock_t mLock {};
};

class SpinLock {
public:
    SpinLock() = default;
    ~SpinLock() = default;

    SpinLock(const SpinLock &) = delete;
    SpinLock &operator = (const SpinLock &) = delete;
    SpinLock(SpinLock &&) = delete;
    SpinLock &operator = (SpinLock &&) = delete;

    inline void TryLock()
    {
        mFlag.test_and_set(std::memory_order_acquire);
    }

    inline void Lock()
    {
        while (mFlag.test_and_set(std::memory_order_acquire)) {
        }
    }

    inline void UnLock()
    {
        mFlag.clear(std::memory_order_release);
    }

private:
    std::atomic_flag mFlag = ATOMIC_FLAG_INIT;
};

template <class T> class Locker {
public:
    explicit Locker(T *lock) : mLock(lock)
    {
        if (mLock != nullptr) {
            mLock->DoLock();
        }
    }

    ~Locker()
    {
        if (mLock != nullptr) {
            mLock->UnLock();
        }
    }

    Locker(const Locker &) = delete;
    Locker &operator = (const Locker &) = delete;
    Locker(Locker &&) = delete;
    Locker &operator = (Locker &&) = delete;

private:
    T *mLock;
};

template <class T> class ReadLocker {
public:
    explicit ReadLocker(T *lock) : mLock(lock)
    {
        if (mLock != nullptr) {
            mLock->LockRead();
        }
    }

    ~ReadLocker()
    {
        if (mLock != nullptr) {
            mLock->UnLock();
        }
    }

    ReadLocker(const ReadLocker &) = delete;
    ReadLocker &operator = (const ReadLocker &) = delete;
    ReadLocker(ReadLocker &&) noexcept = delete;
    ReadLocker &operator = (ReadLocker &&) noexcept = delete;

private:
    T *mLock;
};

template <class T> class WriteLocker {
public:
    explicit WriteLocker(T *lock) : mLock(lock)
    {
        if (mLock != nullptr) {
            mLock->LockWrite();
        }
    }

    ~WriteLocker()
    {
        if (mLock != nullptr) {
            mLock->UnLock();
        }
    }

    WriteLocker(const WriteLocker &) = delete;
    WriteLocker &operator = (const WriteLocker &) = delete;
    WriteLocker(WriteLocker &&) noexcept = delete;
    WriteLocker &operator = (WriteLocker &&) noexcept = delete;

private:
    T *mLock;
};
#define GUARD(lLock, alias) Locker<Lock> __l##alias(lLock)
#define RECURSIVE_GUARD(lock) Locker<RecursiveLock> __locker##lock(lock)
}
}

#endif
