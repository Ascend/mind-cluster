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
#ifndef OCK_DFS_BACKGROUND_CONSTANTS_H
#define OCK_DFS_BACKGROUND_CONSTANTS_H

#include <cstdint>

namespace ock {
namespace bg {
class BackgroundConstants {
public:
    static constexpr int TASK_MAX_RETRY_TIMES = 10;
    static constexpr long TASK_RETRY_FIRST_WAIT_MILL_SECONDS = 1L * 1000L;
    static constexpr uint32_t TASK_MIN_THREAD_NUM = 1U;
    static constexpr uint32_t TASK_MAX_THREAD_NUM = 256U;
};
}
}
#endif // OCK_DFS_BACKGROUND_CONSTANTS_H
