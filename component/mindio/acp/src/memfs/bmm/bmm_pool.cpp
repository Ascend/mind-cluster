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
#include <sys/file.h>
#include <sys/mman.h>
#include <cstdlib>
#include <dlfcn.h>
#include <unistd.h>
#include <numa.h>
#include <linux/version.h>
#include "mem_fs_constants.h"
#include "securec.h"
#include "bmm_pool.h"

namespace ock {
namespace memfs {

static constexpr int8_t DL_SPLIT_LENGTH = 4;

static void SafeCloseFd(int32_t &fd)
{
    auto tmpFd = fd;
    if (UNLIKELY(tmpFd < 0)) {
        return;
    }
    if (__sync_bool_compare_and_swap(&fd, tmpFd, -1)) {
        close(tmpFd);
    }
}

static bool CheckNumaPathLdConfig(char *realPath)
{
    FILE *fp;
    char path[PATH_MAX + 1];
    path[PATH_MAX] = '\0';
    char numaPath[PATH_MAX + 1];
    numaPath[PATH_MAX] = '\0';
    fp = popen("ldconfig -p | grep 'libnuma\\.so\\.1'", "r");
    if (fp == nullptr) {
        MFS_LOG_WARN("failed to run numa ldconfig command .");
        return false;
    }

    while (fgets(path, sizeof(path) - 1, fp) != nullptr) {
        char *arrow = strstr(path, " => ");
        if (arrow != nullptr) {
            arrow += DL_SPLIT_LENGTH;
            int ret = sscanf_s(arrow, "%s", numaPath, sizeof(numaPath) - 1);
            if (ret == -1) {
                MFS_LOG_WARN("format numaPath failed." << strerror(errno));
                pclose(fp);
                return false;
            }
            if (realpath(numaPath, realPath) == nullptr) {
                MFS_LOG_WARN("read numa real path failed." << strerror(errno));
                pclose(fp);
                return false;
            }
        }
    }
    pclose(fp);
    return true;
}

static void SetNumaInterleave()
{
    char realPath[PATH_MAX + 1];
    realPath[PATH_MAX] = '\0';
    void *numaHandle;
    if (!CheckNumaPathLdConfig(realPath)) {
        return;
    }
    std::string numaPath(realPath);
    numaHandle = dlopen(numaPath.c_str(), RTLD_LAZY);
    if (numaHandle == nullptr) {
        MFS_LOG_WARN("open numa path dl fail, follow the default configuration.");
        return;
    }
    NumaAvailableType numaAvailable = reinterpret_cast<NumaAvailableType>(dlsym(numaHandle, "numa_available"));
    NumaSetInterleaveMaskFuncType numaSetInterleaveMask =
        reinterpret_cast<NumaSetInterleaveMaskFuncType>(dlsym(numaHandle, "numa_set_interleave_mask"));
    NumaGetMemsAllowedFuncType numaGetMemsAllowed =
        reinterpret_cast<NumaGetMemsAllowedFuncType>(dlsym(numaHandle, "numa_get_mems_allowed"));
    NumaBitmaskFreeType numaBitmaskFree =
        reinterpret_cast<NumaBitmaskFreeType>(dlsym(numaHandle, "numa_bitmask_free"));
    if (numaAvailable == nullptr || numaSetInterleaveMask == nullptr ||
        numaGetMemsAllowed == nullptr || numaBitmaskFree == nullptr) {
        MFS_LOG_WARN("dlsym fail from numa, follow the default configuration.");
        dlclose(numaHandle);
        numaAvailable = nullptr;
        numaSetInterleaveMask = nullptr;
        numaGetMemsAllowed = nullptr;
        numaBitmaskFree = nullptr;
        return;
    }

    if (numaAvailable() == 0) {
        MFS_LOG_INFO("NUMA available, set interleave mask on all nodes.");
        struct bitmask *allNodes = numaGetMemsAllowed();
        numaSetInterleaveMask(allNodes);
        numaBitmaskFree(allNodes);
    }

    dlclose(numaHandle);
    numaAvailable = nullptr;
    numaSetInterleaveMask = nullptr;
    numaGetMemsAllowed = nullptr;
    numaBitmaskFree = nullptr;
}

MResult MemFsBmmPool::ValidateOptions()
{
    /* do later */

    /* must validate the filename carefully, otherwise there could be security hole */
    return MFS_OK;
}

MResult MemFsBmmPool::Initialize(const MemFsBMMOptions &opt)
{
    MFS_LOG_INFO("Start to initialize bmm pool with options [" << opt.ToString() << "], this could take some time");

    /* validate options */
    RETURN_IT_IF_NOT_OK(ValidateOptions());

    mBmmOptions = opt;

    int fd = -1;
    std::string path = "ock_mfs." + mBmmOptions.name;
    if (mBmmOptions.useDevShm) {
        fd = shm_open(path.c_str(), O_CREAT | O_RDWR, 0600); // 0600
    } else {
#if LINUX_VERSION_CODE < KERNEL_VERSION(3, 17, 0) // memfd_create was introduced in v3.17-rc1
        fd = shm_open(path.c_str(), O_CREAT | O_RDWR, 0600);  // 0600
#else
        fd = syscall(SYS_memfd_create, path.c_str(), 0);
#endif
    }

    if (fd < 0) {
        MFS_LOG_ERROR("Failed to create shm file " << path << " for bmm" << mBmmOptions.name << ", error " <<
            strerror(errno) << ", please check if fd is out of limit");
        return MFS_ERROR;
    }

    /* truncate */
    int64_t size = mBmmOptions.blkCount;
    size = size * mBmmOptions.blkSize;
    if (ftruncate(fd, size) != 0) {
        SafeCloseFd(fd);
        MFS_LOG_ERROR("Failed to truncate file " << path << " for bmm " << mBmmOptions.name << ", error " <<
            strerror(errno));
        return MFS_ERROR;
    }

    /* set numa */
    SetNumaInterleave();

    /* mmap */
    auto mappedAddress = mmap(nullptr, size, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    if (mappedAddress == MAP_FAILED) {
        SafeCloseFd(fd);
        MFS_LOG_ERROR("Failed to mmap file " << path << " for bmm " <<
            mBmmOptions.name << ", error " << strerror(errno));
        return MFS_ERROR;
    }
    mFileBaseAddress = reinterpret_cast<uintptr_t>(mappedAddress);
    mFileEndAddress = mFileBaseAddress + size;

    auto result = MakePools();
    if (result != MFS_OK) {
        SafeCloseFd(fd);
        munmap(mappedAddress, size);
        MFS_LOG_ERROR("Failed to make pools");
        return MFS_ERROR;
    }

    mFd = fd;

    MFS_LOG_INFO("Bmm pool " << opt.name << " is initialized successfully");
    return MFS_OK;
}

void MemFsBmmPool::UnInitialize()
{
    if (mFileBaseAddress == 0) {
        MFS_LOG_WARN("Bmm pool " << mBmmOptions.name << " has not been initialized");
        return;
    }

    void *address = reinterpret_cast<void *>(mFileBaseAddress);
    mFileBaseAddress = 0;
    mFileEndAddress = 0;

    int64_t size = mBmmOptions.blkCount;
    size = size * mBmmOptions.blkSize;
    munmap(address, size);
    SafeCloseFd(mFd);

    mBlkCountTotal = 0;
    mBlkCountRemaining = 0;
    mBlkHead.next = nullptr;
}

MResult MemFsBmmPool::MakePools()
{
    ASSERT_RETURN(mFileBaseAddress != 0, MFS_NOT_INITIALIZED);

    /* linked list */
    uintptr_t currentAddress = mFileBaseAddress;

    /* set head blk */
    auto curBlk = reinterpret_cast<BmmLinkedNode *>(currentAddress);
    mBlkHead.next = curBlk;

    auto stepSize = static_cast<uint64_t>(mBmmOptions.blkSize);

    /* loop n-1 times */
    for (uint32_t i = 0; i < mBmmOptions.blkCount - 1; i++) {
        /* move to next block */
        currentAddress += stepSize;
        /* set current next */
        curBlk->next = reinterpret_cast<BmmLinkedNode *>(currentAddress);
        /* move curBlk */
        curBlk = reinterpret_cast<BmmLinkedNode *>(currentAddress);
    }

    /* set the last to nullptr */
    curBlk->next = nullptr;

    /* set counters */
    mBlkCountTotal = mBmmOptions.blkCount;
    mBlkCountRemaining = mBlkCountTotal;

    return MFS_OK;
}

MResult MemFsBmmPool::AllocateOne(uint64_t &blkId)
{
    ASSERT_RETURN(mBlkCountTotal != 0, MFS_NOT_INITIALIZED);

    mAllocLock.Lock();
    if (mBlkCountRemaining == 0) {
        mAllocLock.UnLock();
        return MFS_ERROR;
    }

    if (UNLIKELY(mBlkHead.next == nullptr)) {
        mAllocLock.UnLock();
        return MFS_ERROR;
    }

    auto tmpBlk = mBlkHead.next;
    mBlkHead.next = tmpBlk->next;
    --mBlkCountRemaining;
    mAllocLock.UnLock();

    BmmBlkId id;
    id.blkAddress = reinterpret_cast<uintptr_t>(tmpBlk);
    id.subPoolId = mSubPoolId;

    blkId = id.whole;

    return MFS_OK;
}

MResult MemFsBmmPool::ReleaseOne(uint64_t &blkId)
{
    BmmBlkId id;
    id.whole = blkId;

    /* validate the blkId */
    ASSERT_RETURN(id.subPoolId == mSubPoolId, MFS_ERROR);
    auto address = reinterpret_cast<uintptr_t>(id.blkAddress);
    if (UNLIKELY(address < mFileBaseAddress || address > mFileEndAddress)) {
        MFS_LOG_ERROR("Invalid blkId " << id.subPoolId << "," << id.blkAddress);
        return MFS_INVALID_PARAM;
    }

    /* insert the head of the linked list */
    mAllocLock.Lock();
    auto tmpNext = mBlkHead.next;
    mBlkHead.next = reinterpret_cast<BmmLinkedNode *>(address);
    mBlkHead.next->next = tmpNext;
    ++mBlkCountRemaining;
    mAllocLock.UnLock();

    return MFS_OK;
}
}
}
