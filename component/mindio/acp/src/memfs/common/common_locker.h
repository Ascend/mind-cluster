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
#ifndef OCK_DFS_COMMON_LOCKER_H
#define OCK_DFS_COMMON_LOCKER_H

#include <pthread.h>
#include <map>

namespace ock {
namespace memfs {
class ReadWriteLock {
public:
    ReadWriteLock() = default;
    ~ReadWriteLock() noexcept
    {
        pthread_rwlock_destroy(&locker);
    }
    void ReadLock()
    {
        pthread_rwlock_rdlock(&locker);
    }
    void WriteLock()
    {
        pthread_rwlock_wrlock(&locker);
    }
    void Unlock()
    {
        pthread_rwlock_unlock(&locker);
    }

private:
    pthread_rwlock_t locker = PTHREAD_RWLOCK_INITIALIZER;
};

class RwLockGuard {
public:
    RwLockGuard(ReadWriteLock &lk, bool rd) noexcept : locker{ &lk }, readLock{ rd }
    {
        if (readLock) {
            locker->ReadLock();
        } else {
            locker->WriteLock();
        }
    }

    ~RwLockGuard() noexcept
    {
        if (!unlocked) {
            locker->Unlock();
        }
    }

    void Unlock() noexcept
    {
        locker->Unlock();
        unlocked = true;
    }

private:
    bool unlocked{ false };
    bool readLock;
    ReadWriteLock *locker;
};

/**
 * @brief 同时操作多个锁时使用。
 * 为防止死锁，先锁序号小的，再锁序号大的
 */
class MultiLockGuard {
public:
    explicit MultiLockGuard(std::map<uint64_t, ReadWriteLock *> locks, bool rd = true) noexcept
        : multiLocks{ std::move(locks) }, readLock{ rd }
    {
        if (readLock) {
            for (auto &locker : multiLocks) {
                locker.second->ReadLock();
            }
        } else {
            for (auto &locker : multiLocks) {
                locker.second->WriteLock();
            }
        }
    }

    ~MultiLockGuard() noexcept
    {
        Unlock();
    }

    void Unlock() noexcept
    {
        if (!unlocked) {
            for (auto &locker : multiLocks) {
                locker.second->Unlock();
            }
            unlocked = true;
        }
    }

private:
    bool unlocked{ false };
    bool readLock;
    std::map<uint64_t, ReadWriteLock *> multiLocks;
};
}
}
#endif // OCK_DFS_COMMON_LOCKER_H
