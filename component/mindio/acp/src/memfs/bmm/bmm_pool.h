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
#ifndef OCK_MEMFS_CORE_BMM_POOL_H
#define OCK_MEMFS_CORE_BMM_POOL_H

#include "bmm_common.h"
#include "common_includes.h"
#include "memfs_logger.h"

namespace ock {
namespace memfs {

struct bitmask {
    unsigned long size; /* number of bits in the map */
    unsigned long *maskp;
};

using NumaAvailableType = int (*)(void);
using NumaSetInterleaveMaskFuncType = void (*)(struct bitmask *nodemask);
using NumaGetMemsAllowedFuncType = struct bitmask *(*)(void);
using NumaBitmaskFreeType = void (*)(struct bitmask *nodemask);

/**
 * @brief Linked node for pool
 */
struct BmmLinkedNode {
    BmmLinkedNode *next = nullptr;
};

/**
 * @brief Block Id
 */
union BmmBlkId {
    struct {
        uint64_t blkAddress : 56; /* block index in one pool */
        uint64_t subPoolId : 8;   /* pool id, not used yet */
    };
    uint64_t whole = 0;
};

/**
 * @brief Bmm pool which provides resources and allocating capability
 */
class MemFsBmmPool {
public:
    MemFsBmmPool() = default;
    ~MemFsBmmPool() = default;

    MResult Initialize(const MemFsBMMOptions &opt);
    void UnInitialize();

    MResult AllocateOne(uint64_t &blkId);
    MResult ReleaseOne(uint64_t &blkId);

    uint64_t GetOffset(uint64_t blkId) const
    {
        BmmBlkId id;
        id.whole = blkId;
        return id.blkAddress - mFileBaseAddress;
    }

    inline int32_t GetFD() const
    {
        return mFd;
    }

    inline uint32_t GetBlkCountTotal() const
    {
        return mBlkCountTotal;
    }

    inline uint32_t GetBlkCountRemaining() const
    {
        return mBlkCountRemaining;
    }

    inline uintptr_t GetHeadNext() const
    {
        return reinterpret_cast<uintptr_t>(mBlkHead.next);
    }

    inline uintptr_t GetFileBaseAddress() const
    {
        return mFileBaseAddress;
    }

    inline uintptr_t GetFileEndAddress() const
    {
        return mFileEndAddress;
    }

    const MemFsBMMOptions &Options() const
    {
        return mBmmOptions;
    }

private:
    MResult ValidateOptions();
    MResult MakePools();

private:
    /* hot used variables */
    ock::memfs::SpinLock mAllocLock; /* spinlock for allocating */
    uint32_t mBlkCountTotal = 0;      /* total block count */
    uint32_t mBlkCountRemaining = 0;  /* remaining block count */
    uintptr_t mFileBaseAddress = 0;   /* base address of mapped address for offset calculation and address validation */
    uintptr_t mFileEndAddress = 0;    /* end address of mapped address for address validation */
    BmmLinkedNode mBlkHead;           /* head for block linked list */
    uint64_t mSubPoolId = 0;          /* sub pool id, used in blockId */
    /* non-hot used variables */
    int32_t mFd = 0;             /* shm file descriptor */
    MemFsBMMOptions mBmmOptions; /* options */
};
}
}

#endif // OCK_MEMFS_CORE_BMM_POOL_H