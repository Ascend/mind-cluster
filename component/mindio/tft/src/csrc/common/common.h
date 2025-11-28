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
#ifndef OCK_TTP_COMMON_H
#define OCK_TTP_COMMON_H

#include <cstdint>
#include <algorithm>
#include <cstdlib>
#include <unistd.h>
#include <semaphore.h>
#include "../framework/file_utils.h"
#include "common_utils.h"
#include "common_messages.h"
#include "common_locks.h"
#include "common_loggers.h"
#include "common_referable.h"
#include "pthread_timedwait.h"
#include "read_write_lock_guard.h"
#include "securec.h"

#endif  // OCK_TTP_COMMON_H