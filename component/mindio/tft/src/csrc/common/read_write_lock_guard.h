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
#ifndef OCK_TTP_READ_WRITE_LOCK_GUARD_H
#define OCK_TTP_READ_WRITE_LOCK_GUARD_H

#include <mutex>
#include "common_locks.h"

namespace ock {
namespace ttp {

constexpr uint32_t TYPE_READ = 0;
constexpr uint32_t TYPE_WRITE = 1;

template <typename T>
class ReadWriteLockGuard {
public:
    explicit ReadWriteLockGuard(T& rdLock, uint32_t lockType) : rdLock_(rdLock), lockType_(lockType)
    {
        if (lockType_ == TYPE_READ) {
            rdLock_.LockRead();
        } else {
            rdLock_.LockWrite();
        }
    }

    ~ReadWriteLockGuard()
    {
        rdLock_.UnLock();
    }

    ReadWriteLockGuard(const ReadWriteLockGuard&) = delete;
    ReadWriteLockGuard& operator=(const ReadWriteLockGuard&) = delete;

private:
    T& rdLock_;
    uint32_t lockType_;
};

using AutoLock = ReadWriteLockGuard<ReadWriteLock>;

}  // namespace ttp
}  // namespace ock

#endif // OCK_TTP_READ_WRITE_LOCK_GUARD_H