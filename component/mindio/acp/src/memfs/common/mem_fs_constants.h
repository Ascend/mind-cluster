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
#ifndef OCK_DFS_MEM_FS_CONSTANTS_H
#define OCK_DFS_MEM_FS_CONSTANTS_H

#include <cstdint>
#include <limits>

namespace ock {
namespace memfs {
class MemFsConstants {
public:
    static constexpr uint64_t INODE_INVALID = std::numeric_limits<uint64_t>::max();
    static constexpr uint64_t ROOT_INODE_NUMBER = 0x1000UL;
    static constexpr uint64_t DATA_BLOCK_SIZE = 128UL << 20;
    static constexpr int OPEN_FILE_FD_START = 1000000;
    static constexpr int64_t MEMFS_CONF_MAX_FILE_SIZE = 4L * 1024L * 1024L; // 4MB
    static constexpr int64_t WHITE_LIST_MAX_FILE_SIZE = 4L * 1024L * 1024L; // 4MB
    static constexpr int MAX_IPC_CONNECTIONS_COUNT = 32;

    /**
     * @brief rename的flags
     */
    static constexpr uint32_t RENAME_FLAG_NOREPLACE = (1U << 0);
    static constexpr uint32_t RENAME_FLAG_EXCHANGE = (1U << 1);
    static constexpr uint32_t RENAME_FLAG_FORCE = (1U << 10);

    /**
     * ACL used
     */
    static constexpr uint16_t ACL_TAG_USER_OBJ = 0x01U;
    static constexpr uint16_t ACL_TAG_USER = 0x02U;
    static constexpr uint16_t ACL_TAG_GROUP_OBJ = 0x04U;
    static constexpr uint16_t ACL_TAG_GROUP = 0x08U;
    static constexpr uint16_t ACL_TAG_MASK = 0x10U;
    static constexpr uint16_t ACL_TAG_OTHER = 0x20U;
    static constexpr uint16_t ACL_DEFAULT_MASK = 0x7U;
    static constexpr auto ACL_XATTR_KEY = "system.posix_acl_access";
};
}
}

#endif // OCK_DFS_MEM_FS_CONSTANTS_H
