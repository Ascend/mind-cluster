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

#ifndef OCK_TTP_MACROS_H
#define OCK_TTP_MACROS_H
#include "common_loggers.h"
namespace ock {
namespace ttp {

#define SET_HCCL_BIT(ARG) ((ARG) ? 0x01 : 0x00)
#define SET_UCE_BIT(ARG) ((ARG) ? 0x02 : 0x00)
#define ONLY_HCCL_BIT(ARG) ((ARG) == 0x01)

#define TTP_LOG_DEBUG(ARGS) COMMON_LOG_DEBUG(TTP, ARGS)
#define TTP_LOG_INFO(ARGS) COMMON_LOG_INFO(TTP, ARGS)
#define TTP_LOG_WARN(ARGS) COMMON_LOG_WARN(TTP, ARGS)
#define TTP_LOG_ERROR(ARGS) COMMON_LOG_ERROR(TTP, ARGS)

#define TTP_RET_LOG(RET, ARGS)                   \
    do {                                         \
        if (__builtin_expect((RET) == 0, 1)) {   \
            TTP_LOG_INFO(ARGS);                  \
        } else {                                 \
            TTP_LOG_WARN(ARGS);                  \
        }                                        \
    } while (0)

#define TTP_LOG_LIMIT_INFO(limit, ARGS) \
    do {                                \
        static uint32_t printCnt = 0;   \
        if (printCnt++ == (limit)) {    \
            TTP_LOG_INFO(ARGS);         \
            printCnt -= limit;          \
        }                               \
    } while (0)

#define TTP_LOG_LIMIT_WARN(limit, ARGS) \
    do {                                \
        static uint32_t printCnt = 0;   \
        if (printCnt++ == (limit)) {    \
            TTP_LOG_WARN(ARGS);         \
            printCnt -= limit;          \
        }                               \
    } while (0)

}

#define TTP_ASSERT_RETURN(ARGS, RET)             \
    do {                                         \
        if (__builtin_expect(!(ARGS), 0) != 0) { \
            TTP_LOG_ERROR("Assert " << #ARGS);   \
            return RET;                          \
        }                                        \
    } while (0)

#define TTP_ASSERT_RET_VOID(ARGS)                \
    do {                                         \
        if (__builtin_expect(!(ARGS), 0) != 0) { \
            TTP_LOG_ERROR("Assert " << #ARGS);   \
            return;                              \
        }                                        \
    } while (0)

#define TTP_ASSERT_RET_VOID_NO_LOG(ARGS)         \
    do {                                         \
        if (__builtin_expect(!(ARGS), 0) != 0) { \
            return;                              \
        }                                        \
    } while (0)
}

#endif // OCK_TTP_MACROS_H
