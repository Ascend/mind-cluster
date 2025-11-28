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
#ifndef OCK_MEMFS_CORE_BMM_COMMON_H
#define OCK_MEMFS_CORE_BMM_COMMON_H

#include "common_includes.h"

namespace ock {
namespace memfs {
struct MemFsBMMOptions {
    bool useDevShm = false;   /* using memfd_create or shm_open to create shared file */
    uint32_t blkCount = 8192; /* total size */
    uint32_t blkSize = 4096;  /* single block size */
    std::string name;         /* name of the pool */

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "name " << name << ", useDevShm " << useDevShm << ", blkCount " << blkCount << ", blockSize " << blkSize;
        return oss.str();
    }
};
}
}

#endif // OCK_MEMFS_CORE_BMM_COMMON_H
