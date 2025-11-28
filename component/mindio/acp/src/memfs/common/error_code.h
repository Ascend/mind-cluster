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
#ifndef OCK_MEMFS_CORE_ERROR_CODE_H
#define OCK_MEMFS_CORE_ERROR_CODE_H

#include <stdint.h>
#include <map>

namespace ock {
namespace memfs {
using MResult = int32_t;

enum MCode : int32_t {
    MFS_OK = 0,
    MFS_ERROR = 1,
    MFS_INVALID_PARAM = 2,
    MFS_ALLOC_FAIL = 3,
    MFS_NEW_OBJ_FAIL = 4,
    MFS_NOT_INITIALIZED = 5,
    MFS_INVALID_CONFIG = 6,
    MFS_ALREADY_DONE = 7,
    MFS_UNSERVICEABLE = 8,
    /* add error code ahead of this */
    MFS_MAX,
};
}
}

#endif // OCK_MEMFS_CORE_ERROR_CODE_H
