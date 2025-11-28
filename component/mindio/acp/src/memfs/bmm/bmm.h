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
#ifndef OCK_MEMFS_CORE_BMM_H
#define OCK_MEMFS_CORE_BMM_H

#include "bmm_common.h"
#include "bmm_pool.h"

namespace ock {
namespace memfs {
class MemFsBMM {
public:
    MemFsBMM() = default;
    ~MemFsBMM() = default;

    /* *
     * @brief Initialization the block memory manager
     *
     * @param opt          [in] option of bmm
     *
     * @return 0 if successfully
     */
    MResult Initialize(const MemFsBMMOptions &opt);

    /* *
     * @brief Un-Initialize the block memory manager
     */
    void UnInitialize();

    /* *
     * @brief Allocate one block
     *
     * @param blkId        [out] allocated blkId
     *
     * @return 0 if successfully, otherwise failed
     */
    inline MResult AllocateOne(uint64_t &blkId)
    {
        ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
        return mBmmPool.AllocateOne(blkId);
    }

    /* *
     * @brief Release one block
     *
     * @param blkId        [in] the block id to be released
     *
     * @return 0 if successfully, otherwise failed
     */
    inline MResult ReleaseOne(uint64_t blkId)
    {
        ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
        return mBmmPool.ReleaseOne(blkId);
    }

    /* *
     * @brief Get file descriptor
     *
     * @return file descriptor number of shared file
     */
    inline int32_t GetFD() const
    {
        ASSERT_RETURN(mInited, -1);
        return mBmmPool.GetFD();
    }

    /* *
     * @brief Get offset of the block based on file base address
     *
     * @param blkId        [in] block id
     *
     * @return offset if successfully, UINT64_MAX if failed
     */
    inline uint64_t GetBlockOffset(uint64_t blkId)
    {
        ASSERT_RETURN(mInited, UINT64_MAX);
        return mBmmPool.GetOffset(blkId);
    }

    const MemFsBmmPool &GetBmmPool() const
    {
        return mBmmPool;
    }
private:
    /* hot used variables */
    MemFsBmmPool mBmmPool; /* memory pool which providing resource and allocating blocks */

    /* non-hot used variables */
    std::mutex mMutex;    /* lock for bmm */
    bool mInited = false; /* initialized or not */
};
using MemFsBMMPtr = std::shared_ptr<MemFsBMM>;
}
}

#endif // OCK_MEMFS_CORE_BMM_H
