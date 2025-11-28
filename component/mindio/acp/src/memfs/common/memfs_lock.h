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
#ifndef OCK_MEMFS_LOCK_H
#define OCK_MEMFS_LOCK_H

#include <atomic>

namespace ock {
namespace memfs {
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
}
}
#endif // OCK_MEMFS_LOCK_H