/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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
#ifndef OCKIO_MEM_FS_STATE_H
#define OCKIO_MEM_FS_STATE_H
#include <cstdint>
#include <string>
#include <memory>

#include "auditlog_adapt.h"
#include "non_copyable.h"

namespace ock {
namespace memfs {
enum MemfsStateCode : int32_t {
    PREPARING = 0,
    STARTING = 1,
    RUNNING = 2,
    PRE_EXITING = 3,
    EXITING = 4,
    EXITED = 5
};

enum MemfsStartProgress : int32_t {
    ZERO_PERCENT = 0,       // 0%
    FIFTY_PERCENT = 50,     // 50%
    EIGHTY_PERCENT = 80,    // 80%
    HUNDRED_PERCENT = 100   // 100%
};

class MemfsState {
public:
    static MemfsState &Instance() noexcept
    {
        static MemfsState instance;
        return instance;
    }

    void SetState(const MemfsStateCode &state) noexcept
    {
        memfsState = state;
        ock::common::HLOG_AUDIT("system", GetMemfsState(state), "fs server", "success");
    }

    void SetState(const MemfsStateCode &state, const MemfsStartProgress &progress) noexcept
    {
        memfsState = state;
        memfsProgress = progress;
        ock::common::HLOG_AUDIT("system", GetMemfsState(state) + "(" + std::to_string(progress) + "%)",
                                "fs server", "success");
    }

    std::pair<MemfsStateCode, MemfsStartProgress> GetState() noexcept
    {
        return { memfsState, memfsProgress };
    }

private:
    static std::string GetMemfsState(const MemfsStateCode &state) noexcept
    {
        switch (state) {
            case PREPARING:
                return "PREPARING";
            case STARTING:
                return "STARTING";
            case RUNNING:
                return "RUNNING";
            case PRE_EXITING:
                return "PRE_EXITING";
            case EXITING:
                return "EXITING";
            case EXITED:
                return "EXITED";
        }
        return "UNKNOWN";
    }

private:
    MemfsStateCode memfsState = MemfsStateCode::EXITED;
    MemfsStartProgress memfsProgress = MemfsStartProgress::ZERO_PERCENT;
};
}
}
#endif // OCKIO_MEM_FS_STATE_H
