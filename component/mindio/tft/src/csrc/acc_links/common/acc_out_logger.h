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
#ifndef ACC_LINKS_ACC_OUT_LOGGER_H
#define ACC_LINKS_ACC_OUT_LOGGER_H

#include <ctime>
#include <iomanip>
#include <cstring>
#include <iostream>
#include <mutex>
#include <unistd.h>
#include <sstream>
#include <sys/time.h>
#include <sys/syscall.h>

#include "common_loggers.h"

#define LOG_DEBUG(ARGS) COMMON_OUT_LOG(ock::ttp::DEBUG_LEVEL, ACC, ARGS)
#define LOG_INFO(ARGS) COMMON_OUT_LOG(ock::ttp::INFO_LEVEL, ACC, ARGS)
#define LOG_WARN(ARGS) COMMON_OUT_LOG(ock::ttp::WARN_LEVEL, ACC, ARGS)
#define LOG_ERROR(ARGS) COMMON_OUT_LOG(ock::ttp::ERROR_LEVEL, ACC, ARGS)

#ifndef ENABLE_TRACE_LOG
#define LOG_TRACE(ARGS)
#elif
#define LOG_TRACE(ARGS) COMMON_OUT_LOG(ock::ttp::DEBUG_LEVEL, AccLinksTrace, ARGS)
#endif

#define ASSERT_RETURN(ARGS, RET)           \
    do {                                   \
        if (UNLIKELY(!(ARGS))) {           \
            LOG_ERROR("Assert " << #ARGS); \
            return RET;                    \
        }                                  \
    } while (0)

#define ASSERT_RET_VOID(ARGS)              \
    do {                                   \
        if (UNLIKELY(!(ARGS))) {           \
            LOG_ERROR("Assert " << #ARGS); \
            return;                        \
        }                                  \
    } while (0)

#define ASSERT(ARGS)                       \
    do {                                   \
        if (UNLIKELY(!(ARGS))) {           \
            LOG_ERROR("Assert " << #ARGS); \
        }                                  \
    } while (0)

#define VALIDATE_RETURN(ARGS, msg, RET)          \
    do {                                         \
        if (__builtin_expect(!(ARGS), 0) != 0) { \
            LOG_ERROR(msg);                      \
            return RET;                          \
        }                                        \
    } while (0)

#define LOG_ERROR_RETURN_IT_IF_NOT_OK(result, msg) \
    do {                                           \
        auto innerResult = (result);               \
        if (UNLIKELY(innerResult != ACC_OK)) {     \
            LOG_ERROR(msg);                        \
            return innerResult;                    \
        }                                          \
    } while (0)

#endif  // ACC_LINKS_ACC_OUT_LOGGER_H
