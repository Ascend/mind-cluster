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
#ifndef OCK_MEMFS_CORE_FS_OPERATION_H
#define OCK_MEMFS_CORE_FS_OPERATION_H

#include <csignal>
#include <unordered_map>
#include <utility>
#include "memfs_execution_service.h"
#include "ipc_client.h"
#include "memfs_sdk_types.h"

namespace ock {
namespace memfs {
/**
 * @brief File open info
 */
class MemFsFileOpInfo {
public:
    MemFsFileOpInfo() = default;
    ~MemFsFileOpInfo()
    {
        if (mBlkMappedAddress != nullptr) {
            delete[] mBlkMappedAddress;
            mBlkMappedAddress = nullptr;
        }
    }

    MResult Initialize(int32_t sharedFd, uint32_t blkSize, uint32_t maxBlkCount, uint64_t fileSize = 0)
    {
        ASSERT_RETURN(maxBlkCount != 0, MFS_INVALID_PARAM);
        ASSERT_RETURN(blkSize > 0 && blkSize <= gMaxBlockSize, MFS_INVALID_PARAM);
        ASSERT_RETURN((uint64_t)maxBlkCount * blkSize <= gMaxSharedFileSize, MFS_INVALID_PARAM);

        mBlkMappedAddress = new (std::nothrow) uintptr_t[maxBlkCount];
        if (UNLIKELY(mBlkMappedAddress == nullptr)) {
            LOG_ERROR("Failed to new mapped address space of file op info");
            return MFS_NEW_OBJ_FAIL;
        }

        mBlkBucketCount = maxBlkCount;
        mSharedFd = sharedFd;
        mBlockSize = blkSize;
        mPosition = 0;
        mFileSize = fileSize;
        return MFS_OK;
    }

    MResult SetBlkLocalAddress(uint64_t position, uintptr_t address)
    {
        ASSERT_RETURN(mBlkMappedAddress != nullptr, MFS_NOT_INITIALIZED);
        uint64_t index = position / mBlockSize;
        ASSERT_RETURN(index < mBlkBucketCount, MFS_INVALID_PARAM);
        mBlkMappedAddress[index] = address;
        return MFS_OK;
    }

    MResult GetBlkLocalAddress(uintptr_t &address)
    {
        ASSERT_RETURN(mBlkMappedAddress != nullptr, MFS_NOT_INITIALIZED);
        uint64_t index = mPosition / mBlockSize;
        ASSERT_RETURN(index < mBlkBucketCount, MFS_INVALID_PARAM);
        address = mBlkMappedAddress[index] + mPosition % mBlockSize;
        return MFS_OK;
    }

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "sFD " << mSharedFd << ", blkSize " << mBlockSize << ", pos " << mPosition << ", bucketCount " <<
            mBlkBucketCount << ", blockCount " << mCurrentBlockCount;
        return oss.str();
    }

    inline uint64_t Position() const
    {
        return mPosition;
    }

private:
    int32_t mSharedFd = -1;                 /* file descriptor */
    uint32_t mBlockSize = 0;                /* file block size */
    uint64_t mPosition = 0;                 /* current operation */
    uintptr_t *mBlkMappedAddress = nullptr; /* mapped address */
    uint32_t mBlkBucketCount = 0;
    int32_t mOpenFlag = 0;
    uint64_t mFileSize = 0;
    uint64_t mCurrentBlockCount = 0;

private:
    static const uint64_t gMaxBlockSize = 1073741824L;         /* 1GB */
    static const uint64_t gMaxSharedFileSize = 1099511627776L; /* 1TB */

    friend class MemFsClientOperation;
};
using MemFsFileOpInfoPtr = std::shared_ptr<MemFsFileOpInfo>;

/*
 * @brief File operation info map
 */
class MemFsFileMap {
public:
    MResult Put(int32_t fd, MemFsFileOpInfoPtr &fi)
    {
        std::lock_guard<std::mutex> guard(mMutex);
        mFd2InfoMap[fd] = fi;
        return MFS_OK;
    }

    MResult Get(int32_t fd, MemFsFileOpInfoPtr &file)
    {
        std::lock_guard<std::mutex> guard(mMutex);
        auto iter = mFd2InfoMap.find(fd);
        if (iter == mFd2InfoMap.end()) {
            return MFS_ERROR;
        }

        file = iter->second;
        if (file == nullptr) {
            LOG_ERROR("FileMapInfo is null after Get operation.");
            return MFS_ERROR;
        }
        return MFS_OK;
    }

    MResult Remove(int32_t fd)
    {
        std::lock_guard<std::mutex> guard(mMutex);
        mFd2InfoMap.erase(fd);
        return MFS_OK;
    }

    void Clear()
    {
        std::lock_guard<std::mutex> guard(mMutex);
        mFd2InfoMap.clear();
    }

private:
    std::mutex mMutex;
    std::unordered_map<int32_t, MemFsFileOpInfoPtr> mFd2InfoMap;
};

/**
 * @brief
 * 一次写文件的参数信息，如果启动多线程并行写入，这个表示一路并行的写入参数
 */
struct FileWriteParam {
    const std::vector<Buffer> *buffers;
    const int index;
    const uint64_t fileOffset;
    const uint64_t pageIndex;
    const uint64_t pageOffset;
    const uint64_t writeSize;

    FileWriteParam() : buffers{ nullptr }, index{ 0 }, fileOffset{ 0 }, pageIndex{ 0 }, pageOffset{ 0 }, writeSize{ 0 }
    {}
    FileWriteParam(const std::vector<Buffer> &buf, int idx, uint64_t fo, uint64_t pi, uint64_t po, uint64_t ws)
        : buffers{ &buf }, index{ idx }, fileOffset{ fo }, pageIndex{ pi }, pageOffset{ po }, writeSize{ ws }
    {}
};

/**
 * @brief
 * 一次读文件的参数信息，如果启动多线程并行写入，这个表示一路并行的读入参数
 */
struct FileReadParam {
    std::vector<ReadBuffer> *buffers;
    const int index;
    const uint64_t fileOffset;
    const uint64_t pageIndex;
    const uint64_t pageOffset;
    const uint64_t readSize;
    FileReadParam() : buffers{ nullptr }, index{ 0 }, fileOffset{ 0 }, pageOffset{ 0 }, pageIndex{ 0 }, readSize{ 0 } {}
    FileReadParam(std::vector<ReadBuffer> &buf, int idx, uint64_t fo, uint64_t pi, uint64_t po, uint64_t ws)
        : buffers{ &buf }, index{ idx }, fileOffset{ fo }, pageIndex{ pi }, pageOffset{ po }, readSize{ ws }
    {}
};

/**
 * @brief 当启用多线程并行读写时，用于控制多线程的上下文对象。
 */
struct ParallelRwContext {
    MemFsFileOpInfoPtr fileInfo;
    std::mutex mutex;
    std::condition_variable cond;
    int totalCount;
    int finishedCount;
    int failedCount;

    ParallelRwContext(MemFsFileOpInfoPtr fi, int count)
        : fileInfo{ std::move(fi) }, totalCount{ count }, finishedCount{ 0 }, failedCount{ 0 }
    {}

    template <typename T, typename U>
    U GenerateRwParameters(const U &param, T &buffer, uint64_t totalSize, uint64_t sliceSize) const
    {
        U result;
        auto taskIndex = 0;
        auto totalOffset = 0UL;
        auto pageIndex = 0UL;
        auto pageOffset = 0UL;
        while (true) {
            auto needCopyBytes = (taskIndex == totalCount - 1) ? totalSize - totalOffset : sliceSize;
            result.emplace_back(buffer, taskIndex, fileInfo->Position() + totalOffset, pageIndex, pageOffset,
                needCopyBytes);
            if (++taskIndex >= totalCount) {
                break;
            }
            totalOffset += needCopyBytes;

            // use bytes of need copy in iov
            auto needSkipBytes = needCopyBytes;
            for (; needSkipBytes > 0UL; pageIndex++, pageOffset = 0) {
                if (pageOffset + needSkipBytes < buffer[pageIndex].size) {
                    pageOffset += needSkipBytes;
                    break;
                }

                needSkipBytes -= (buffer[pageIndex].size - pageOffset);
            }
        }
        return std::move(result);
    }
};

/**
 * @brief Client operation
 */
class MemFsClientOperation {
public:
    static inline MemFsClientOperation *Instance() noexcept
    {
        if (gInstance == nullptr) {
            std::lock_guard<std::mutex> guard(gLock);
            if (gInstance == nullptr) {
                gInstance = new (std::nothrow) MemFsClientOperation();
                if (gInstance == nullptr) {
                    LOG_ERROR("Failed to new MemFsClient object, probably out of memory");
                    return nullptr;
                }
            }
        }
        return gInstance;
    }

public:
    ~MemFsClientOperation() = default;

    MResult Initialize(const IpcClientConfig &config, const std::map<std::string, std::string> &serverInfoParam);
    void UnInitialize();

    inline bool IsForkedProcess() const
    {
        return mForkChild;
    }

    int LinkFile(const std::string &source, const std::string &target);
    int RenameFile(const std::string &source, const std::string &target, uint32_t flags);
    MResult MakeDir(const std::string &dir, int32_t flags, bool recursive = false);
    MResult OpenFile(const std::string &path, int32_t flags, int32_t &fd);
    MResult Write(int32_t fd, uintptr_t data, uint64_t size);
    MResult Write(int32_t fd, const std::vector<Buffer> &buffers, uint64_t &length);
    MResult Read(int32_t fd, uintptr_t data, uint64_t position, uint64_t size);
    MResult Read(int32_t fd, std::vector<ReadBuffer> &buffers, uint64_t &length);
    MResult Flush(int32_t fd);
    MResult Access(const std::string &path, int32_t mode);
    MResult Seek(int32_t fd, int64_t offset, int32_t whence);
    MResult Tell(int32_t fd, int64_t &tell);
    MResult GetSize(int32_t fd, int64_t &size);
    MResult Cntl(int32_t fd, int32_t cmd);
    MResult Close(int32_t fd);
    MResult CloseWithUnlink(int32_t fd);
    MResult PreloadFile(const std::string &path);
    MResult CheckBackgroundTask();
    inline std::string GetUfsMountPath() const
    {
        return mSharedFileInfo.UfsPath();
    }
    inline bool GetServiceStatus()
    {
        return mIpClient->GetIpcServiceStatus();
    }

private:
    /* hidden constructor, use instance() directly */
    MemFsClientOperation() = default;

    MResult GetSharedFileInfo();
    MResult PrepareWriteParallelConfig(const ShareFileInfoResp &resp);
    MResult CheckOpenFile4ReadResponse(const std::string &path, OpenFileWithBlockResp *resp, uint32_t len);
    MResult OpenFile4ReadRequest(const std::string &path, int32_t flags, OpenFileWithBlockResp *&resp,
        uint32_t &rspLen);
    MResult OpenFile4Read(const std::string &path, int32_t flags, int32_t &fd);
    MResult OpenFile4Write(const std::string &path, int32_t flags, int32_t &fd);
    MResult Close(int32_t fd, FileOpFile op);
    MResult CheckBlockValid(uint64_t block, uint32_t blockSize) const
    {
        return CheckBlocksValid(&block, blockSize, 1U);
    }
    MResult CheckBlocksValid(const uint64_t blocks[], uint32_t blockSize, uint32_t blockCount) const;
    MResult CheckWriteParam(int32_t fd, uint64_t size, MemFsFileOpInfoPtr &fi);
    MResult CheckWriteParam(int32_t fd, const std::vector<Buffer> &buffers, uint64_t total, MemFsFileOpInfoPtr &fi);
    MResult CheckReadParam(int32_t fd, uint64_t size, MemFsFileOpInfoPtr &fi);
    MResult CheckReadParam(int32_t fd, const std::vector<ReadBuffer> &buffers, uint64_t total, MemFsFileOpInfoPtr &fi);
    static uint64_t NeedAllocateBytesForWrite(const MemFsFileOpInfoPtr &fi, uint64_t size);
    int AllocateManyBlocksForWrite(int fd, MemFsFileOpInfoPtr &fi, uint64_t size);
    static int WriteDataFromPosition(const MemFsFileOpInfoPtr &fi, const FileWriteParam &param);
    static int ReadDataFromPosition(const MemFsFileOpInfoPtr &fi, const FileReadParam &param);
    static int WriteInSingleThread(MemFsFileOpInfoPtr &fi, const std::vector<Buffer> &buffers, uint64_t length);
    int WriteInMultiThreads(MemFsFileOpInfoPtr &fi, const std::vector<Buffer> &buffers, uint64_t length);
    MResult ReadInSingleThread(int32_t fd, ock::memfs::MemFsFileOpInfoPtr &fi, std::vector<ReadBuffer> &buffers,
        uint64_t size);
    MResult ReadInMultiThreads(int32_t fd, ock::memfs::MemFsFileOpInfoPtr &fi, std::vector<ReadBuffer> &buffers,
        uint64_t &length);
    int ConnectedProcess();
    void DisconnectedProcess();
    int RegisterAtFork();
    static void PrepareForked();
    static void ParentProcessForked();
    static void ChildProcessForked();
    static int RegisterSignalHandler();
    static void SignalHandler(int sig, siginfo_t *info, void *param);
    MResult RealPreloadFile(const std::string &path);

private:
    IpcClientPtr mIpClient = nullptr;
    MemFsFileMap mFileMaps;
    int mSharedFd = -1;
    uintptr_t mSharedFileAddress = 0;
    uintptr_t mSharedFileEndAddress = 0;

    std::mutex mMutex;
    bool mInited = false;
    ShareFileInfoResp mSharedFileInfo;
    ExecutorServicePtr mWriteThreadPool;
    bool mAtForkRegister{ false };
    bool mForkChild{ false };

private:
    static std::mutex gLock;
    static MemFsClientOperation *gInstance;
};
}
}

#endif // OCK_MEMFS_CORE_FS_OPERATION_H
