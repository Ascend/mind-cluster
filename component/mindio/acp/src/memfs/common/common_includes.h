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
#ifndef OCK_MEMFS_CORE_COMMON_INCLUDES_H
#define OCK_MEMFS_CORE_COMMON_INCLUDES_H

#include <cstdint>
#include <climits>
#include <condition_variable>
#include <mutex>
#include <string>
#include <vector>
#include <set>
#include <map>
#include <unordered_map>
#include <thread>
#include <algorithm>
#include <fstream>
#include <functional>
#include <memory>

#include "memfs_out_logger.h"
#include "hlog.h"

#include "error_code.h"
#include "memfs_lock.h"
#include "memfs_common.h"
#include "memfs_str_util.h"
#include "memfs_file_util.h"

using namespace ock::memfs;

#define LOG_DEBUG(ARGS) DAGGER_LOG_DEBUG(MemFS, ARGS)
#define LOG_INFO(ARGS) DAGGER_LOG_INFO(MemFS, ARGS)
#define LOG_WARN(ARGS) DAGGER_LOG_WARN(MemFS, ARGS)
#define LOG_ERROR(ARGS) DAGGER_LOG_ERROR(MemFS, ARGS)

#ifdef LOG_TRACE_INFO_ENABLED
#define LOG_TRACE(ARGS) DAGGER_LOG_INFO(MemFSTrace, ARGS)
#else
#define LOG_TRACE(ARGS)
#endif // LOG_TRACE_INFO_ENABLED

#define ASSERT_RETURN(ARGS, RET) DAGGER_ASSERT_RETURN(MemFS, ARGS, RET)
#define ASSERT_RET_VOID(ARGS) DAGGER_ASSERT_RET_VOID(MemFS, ARGS)
#define ASSERT(ARGS) DAGGER_ASSERT(MemFS, ARGS)

#define LOG_ERROR_RETURN_IT_IF_NOT_OK(result, msg) \
    do {                                           \
        auto innerResult = (result);               \
        if (UNLIKELY(innerResult != MFS_OK)) {     \
            LOG_ERROR(msg);                        \
            return innerResult;                    \
        }                                          \
    } while (0)

#define RETURN_IT_IF_NOT_OK(result)            \
    do {                                       \
        auto innerResult = (result);           \
        if (UNLIKELY(innerResult != MFS_OK)) { \
            return innerResult;                \
        }                                      \
    } while (0)

#define DECLARE_CHAR_ARRAY_SET_FUNC(func, CHAR_ARRAY)    \
    bool func(const std::string &other)                  \
    {                                                    \
        if (other.length() > (sizeof(CHAR_ARRAY) - 1)) { \
            return false;                                \
        }                                                \
                                                         \
        for (uint32_t i = 0; i < other.length(); i++) {  \
            (CHAR_ARRAY)[i] = other.at(i);                 \
        }                                                \
                                                         \
        (CHAR_ARRAY)[other.length()] = '\0';             \
        return true;                                     \
    }

#define DECLARE_CHAR_ARRAY_GET_FUNC(func, CHAR_ARRAY)                   \
    std::string func() const                                            \
    {                                                                   \
        return { CHAR_ARRAY, strnlen(CHAR_ARRAY, sizeof(CHAR_ARRAY)) }; \
    }

/* constant defines */
constexpr uint32_t FS_PATH_MAX = 2048;
constexpr int32_t FS_KB_UNIT = 1024;

#endif // OCK_MEMFS_CORE_COMMON_INCLUDES_H
