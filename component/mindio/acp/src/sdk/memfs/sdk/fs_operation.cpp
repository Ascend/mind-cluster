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
#include <sys/mman.h>
#include <fcntl.h>
#include <csignal>
#include <chrono>
#include "fs_operation.h"

using namespace ock::memfs;

static constexpr auto MAX_VALID_WRITE_PARALLEL_THREAD = 100U;
static constexpr auto MIN_WRITE_PARALLEL_SLICE_COUNT = 2U;
static constexpr auto MAX_LOAD_RETRY_SECOND_TIME = 10U;
static constexpr auto MAX_REPULL_SERVER_TIMES = 3U;
static constexpr auto MAX_CHECK_BACKGROUND_TASK_TIMES = 10U; // 10*30s = 300s

MemFsClientOperation *MemFsClientOperation::gInstance = nullptr;
std::mutex MemFsClientOperation::gLock;

MResult MemFsClientOperation::Initialize(const IpcClientConfig &config,
    const std::map<std::string, std::string> &serverInfoParam)
{
    std::lock_guard<std::mutex> guard(mMutex);
    if (mInited) {
        return MFS_OK;
    }

    auto ret = RegisterAtFork();
    if (ret != MFS_OK) {
        return ret;
    }

    mIpClient = std::make_shared<IpcClient>(
        [this]() {
            LOG_INFO("connection to server restore");
            std::unique_lock<std::mutex> lockGuard(mMutex);
            return ConnectedProcess();
        },
        [this]() {
            LOG_INFO("connection to server shutdown");
            std::unique_lock<std::mutex> lockGuard(mMutex);
            DisconnectedProcess();
        },
        serverInfoParam);

    ASSERT_RETURN(mIpClient != nullptr, MFS_NEW_OBJ_FAIL);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mIpClient->Start(config), "Failed to start ipc client");

    auto result = mIpClient->Connect();
    if (result != MFS_OK) {
        mIpClient->Stop();
        mIpClient = nullptr;
        LOG_ERROR("Failed to connect to ip server");
        return result;
    }

    result = ConnectedProcess();
    if (result != MFS_OK) {
        return result;
    }

    mInited = true;
    return MFS_OK;
}

void MemFsClientOperation::UnInitialize()
{
    std::lock_guard<std::mutex> guard(mMutex);
    if (!mInited) {
        return;
    }
    mIpClient->Stop();
    mIpClient = nullptr;
    mInited = false;
    DisconnectedProcess();

    if (mWriteThreadPool.get() != nullptr) {
        mWriteThreadPool->Stop();
    }
    mWriteThreadPool = nullptr;
}

MResult MemFsClientOperation::GetSharedFileInfo()
{
    ASSERT_RETURN(mIpClient != nullptr, MFS_NEW_OBJ_FAIL);
    ShareFileInfoReq req;
    req.flags = 0;

    ShareFileInfoResp resp;
    auto result = mIpClient->SyncCall<ShareFileInfoReq, ShareFileInfoResp>(IPC_OP_GET_SHARED_FILE_INFO, req, resp);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(result, "Failed to call server to get shared file info, result " << result);

    if (resp.result != MFS_OK) {
        LOG_ERROR("Failed get shared file info, server error " << resp.result);
        return resp.result;
    }

    if (resp.singleFileSize > MemFsFileOpInfo::gMaxSharedFileSize) {
        LOG_ERROR("Exceeded the maximum shared file size, size:" << resp.singleFileSize << ", maxSize" <<
            MemFsFileOpInfo::gMaxSharedFileSize);
        return MFS_INVALID_PARAM;
    }

    if (resp.fileCount != 1U) {
        LOG_ERROR("file count:" << resp.fileCount << " should be 1");
        return MFS_INVALID_PARAM;
    }

    if (resp.maxBlkCountInSingleFile == 0) {
        LOG_ERROR("file max block count is zero");
        return MFS_INVALID_PARAM;
    }

    MResult ret = PrepareWriteParallelConfig(resp);
    if (ret != MFS_OK) {
        LOG_ERROR("PrepareWriteParallelConfig failed : " << ret);
        return ret;
    }

    mSharedFileInfo = resp;

    LOG_INFO("Shared file info " << resp.ToString());

    return MFS_OK;
}

MResult MemFsClientOperation::PrepareWriteParallelConfig(const ock::memfs::ShareFileInfoResp &resp)
{
    if (!resp.writeParallelEnabled) {
        return MFS_OK;
    }

    if (resp.writeParallelThreadNum == 0 || resp.writeParallelThreadNum > MAX_VALID_WRITE_PARALLEL_THREAD) {
        return MFS_INVALID_PARAM;
    }

    if (resp.writeParallelSlice == 0UL || resp.writeParallelSlice > resp.singleFileSize) {
        return MFS_INVALID_PARAM;
    }

    mWriteThreadPool = ExecutorService::Create(resp.writeParallelThreadNum);
    if (mWriteThreadPool.get() == nullptr) {
        LOG_ERROR("create thread pool with count = " << resp.writeParallelThreadNum << " failed.");
        return MFS_ALLOC_FAIL;
    }

    auto success = mWriteThreadPool->Start();
    if (!success) {
        LOG_ERROR("start thread pool with count = " << resp.writeParallelThreadNum << " failed.");
        mWriteThreadPool = nullptr;
        return MFS_ALLOC_FAIL;
    }

    return MFS_OK;
}

int MemFsClientOperation::LinkFile(const std::string &source, const std::string &target)
{
    ASSERT_RETURN(mInited, -ENODATA);
    ASSERT_RETURN(GetServiceStatus(), -ENODATA);
    ASSERT_RETURN(!source.empty(), -EINVAL);
    ASSERT_RETURN(!target.empty(), -EINVAL);
    ASSERT_RETURN(source.size() < FS_PATH_MAX, -EINVAL);
    ASSERT_RETURN(target.size() < FS_PATH_MAX, -EINVAL);

    LinkFileReq req;
    req.SourcePath(source);
    req.TargetPath(target);

    LinkFileRes resp;
    auto result = mIpClient->SyncCall<LinkFileReq, LinkFileRes>(IPC_OP_LINK_FILE, req, resp);
    if (result != 0) {
        LOG_ERROR("send ipc message: " << req.ToString() << " failed: " << result);
        return -ECOMM;
    }

    if (resp.result != 0) {
        LOG_ERROR("link for: " << req.ToString() << " failed: " << resp.errorCode << ": " << strerror(resp.errorCode));
        return -resp.errorCode;
    }

    return 0;
}

int MemFsClientOperation::RenameFile(const std::string &source, const std::string &target, uint32_t flags)
{
    ASSERT_RETURN(mInited, -ENODATA);
    ASSERT_RETURN(GetServiceStatus(), -ENODATA);
    ASSERT_RETURN(!source.empty(), -EINVAL);
    ASSERT_RETURN(!target.empty(), -EINVAL);
    ASSERT_RETURN(source.size() < FS_PATH_MAX, -EINVAL);
    ASSERT_RETURN(target.size() < FS_PATH_MAX, -EINVAL);

    RenameFileReq req;
    req.flags = flags;
    req.SourcePath(source);
    req.TargetPath(target);

    LinkFileRes resp;
    auto result = mIpClient->SyncCall<RenameFileReq, RenameFileRes>(IPC_OP_RENAME_FILE, req, resp);
    if (result != 0) {
        LOG_ERROR("send ipc message: rename-" << req.ToString() << " failed: " << result);
        return -ECOMM;
    }

    if (resp.result != 0) {
        LOG_ERROR("rename for:" << req.ToString() << " failed: " << resp.errorCode << ":" << strerror(resp.errorCode));
        return -resp.errorCode;
    }

    return 0;
}

MResult MemFsClientOperation::MakeDir(const std::string &path, int32_t flags, bool recursive)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(!path.empty(), MFS_INVALID_PARAM);
    ASSERT_RETURN(path.size() < FS_PATH_MAX, MFS_INVALID_PARAM);

    MakeDirReq req;
    req.Path(path);
    req.flags = flags;
    req.recursive = recursive;

    MakeDirResp resp;
    auto result = mIpClient->SyncCall<MakeDirReq, MakeDirResp>(IPC_OP_MAKE_DIR, req, resp);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(result,
        "Failed to call server to mkdir @" << FileCheckUtils::RemovePrefixPath(path) << ", result " << result);

    if (resp.result != MFS_OK) {
        LOG_ERROR("Failed to mkdir @" << FileCheckUtils::RemovePrefixPath(path) << ", server error " << resp.result);
        return resp.result;
    }

    LOG_TRACE("Created dir @" << FileCheckUtils::RemovePrefixPath(path) << " in memfs");
    return MFS_OK;
}

MResult MemFsClientOperation::CheckOpenFile4ReadResponse(const std::string &path, OpenFileWithBlockResp *resp,
    uint32_t len)
{
    if (resp->result != MFS_OK) {
        LOG_WARN("Failed to open file @" << FileCheckUtils::RemovePrefixPath(path) << ", server error " <<
            resp->result << ", fd = " << resp->fd);
        auto result = resp->result;
        free(resp);
        return result;
    }

    if (resp->blockSize == 0U || resp->fd < 0) {
        LOG_ERROR("Failed to open file @" << FileCheckUtils::RemovePrefixPath(path) << ", server block size is zero.");
        free(resp);
        return -EBADE;
    }

    auto dataShouldSize = sizeof(OpenFileWithBlockResp) + resp->blockCount * sizeof(uint64_t);
    if (dataShouldSize > len) {
        LOG_ERROR("Failed to open file @" << FileCheckUtils::RemovePrefixPath(path) << ", server response length: " <<
            len << ", resp(" << resp->ToString() << "), data should size = " << dataShouldSize);
        free(resp);
        return -ECOMM;
    }

    auto ret = CheckBlocksValid(resp->dataBlock, resp->blockSize, resp->blockCount);
    if (ret != MFS_OK) {
        LOG_ERROR("response data blocks invalid, size: " << resp->blockSize << ", count: " << resp->blockCount);
        free(resp);
        return -ECOMM;
    }

    return MFS_OK;
}

MResult MemFsClientOperation::OpenFile4ReadRequest(const std::string &path, int32_t flags, OpenFileWithBlockResp *&resp,
    uint32_t &rspLen)
{
    /* create request */
    OpenFileReq req;
    req.FileName(path);
    req.flags = flags;

    /* record request time */
    auto startTime = std::chrono::high_resolution_clock::now();
    uint32_t execTime = 0;

    /* real call server */
    int32_t result = 0;
    do {
        result = mIpClient->SyncCall<OpenFileReq, OpenFileWithBlockResp>(IPC_OP_OPEN_FILE_FOR_READ, req, &resp, rspLen);
        LOG_ERROR_RETURN_IT_IF_NOT_OK(result,
            "Failed to call server to open file @" << FileCheckUtils::RemovePrefixPath(path) << ", result " << result);
        if (resp == nullptr) {
            LOG_ERROR("Failed to open file @" << FileCheckUtils::RemovePrefixPath(path) <<
                ", server response nullptr.");
            return -ECOMM;
        }
        auto curTime = std::chrono::high_resolution_clock::now();
        execTime = std::chrono::duration_cast<std::chrono::seconds>(curTime - startTime).count();
    } while (resp->result == -EAGAIN && execTime < MAX_LOAD_RETRY_SECOND_TIME);

    return result;
}

MResult MemFsClientOperation::OpenFile4Read(const std::string &path, int32_t flags, int32_t &fd)
{
    /* call server */
    OpenFileWithBlockResp *resp = nullptr;
    uint32_t rspLen = 0;
    auto result = OpenFile4ReadRequest(path, flags, resp, rspLen);
    if (resp == nullptr) {
        LOG_ERROR("Failed to open file @" << FileCheckUtils::RemovePrefixPath(path) << ", server response nullptr.");
        return -ECOMM;
    }
    if ((result = CheckOpenFile4ReadResponse(path, resp, rspLen)) != MFS_OK) {
        return result;
    }

    /* create file op info */
    MemFsFileOpInfoPtr fi = std::make_shared<MemFsFileOpInfo>();
    if (fi.get() == nullptr) {
        LOG_ERROR("Failed to malloc @" << FileCheckUtils::RemovePrefixPath(path));
        Close(resp->fd);
        free(resp);
        return MFS_NEW_OBJ_FAIL;
    }

    /* initialize to allocate buckets for blk addresses */
    result = fi->Initialize(mSharedFd, resp->blockSize, resp->blockCount, resp->fileSize);
    if (result != MFS_OK) {
        LOG_ERROR("Failed to init file operation info @" << FileCheckUtils::RemovePrefixPath(path));
        Close(resp->fd);
        free(resp);
        return result;
    }

    /* set the first block address */
    for (uint64_t blkIndex = 0; blkIndex < resp->blockCount; ++blkIndex) {
        result = fi->SetBlkLocalAddress(blkIndex * resp->blockSize, mSharedFileAddress + resp->dataBlock[blkIndex]);
        if (result != MFS_OK) {
            LOG_ERROR("Failed to set address in file operation info @" << FileCheckUtils::RemovePrefixPath(path));
            Close(resp->fd);
            free(resp);
            return result;
        }
    }
    fi->mCurrentBlockCount = resp->blockCount;

    /* put the fd into map */
    result = mFileMaps.Put(resp->fd, fi);
    if (result != MFS_OK) {
        LOG_ERROR("Failed to add file operation info in @" << FileCheckUtils::RemovePrefixPath(path));
        Close(resp->fd);
        free(resp);
        return result;
    }

    fd = resp->fd;
    fi->mOpenFlag = flags;

    free(resp);
    return MFS_OK;
}

MResult MemFsClientOperation::OpenFile4Write(const std::string &path, int32_t flags, int32_t &fd)
{
    /* create request */
    OpenFileReq req;
    if (!req.FileName(path)) {
        LOG_ERROR("file path: " << FileCheckUtils::RemovePrefixPath(path) << " invalid.");
        return MFS_INVALID_PARAM;
    }

    req.flags = flags;

    /* call server */
    OpenFileResp resp;
    auto result = mIpClient->SyncCall<OpenFileReq, OpenFileResp>(IPC_OP_OPEN_FILE, req, resp);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(result,
        "Failed to call server to open file @" << FileCheckUtils::RemovePrefixPath(path) << ", result " << result);

    LOG_TRACE("Open file response info " << resp.ToString());

    if (resp.result != MFS_OK || resp.fd < 0 || resp.blockSize != resp.dataBlock.size) {
        LOG_ERROR("Failed to open file @" << FileCheckUtils::RemovePrefixPath(path) << ", server error " <<
            resp.result << ", fd:" << resp.fd);
        return resp.result;
    }

    auto ret = CheckBlockValid(resp.dataBlock.offset, resp.blockSize);
    if (ret != MFS_OK) {
        LOG_ERROR("open file @" << FileCheckUtils::RemovePrefixPath(path) << " to read response block invalid.");
        return ret;
    }

    /* create file op info */
    MemFsFileOpInfoPtr fi = std::make_shared<MemFsFileOpInfo>();
    ASSERT_RETURN(fi.get() != nullptr, MFS_NEW_OBJ_FAIL);

    /* initialize to allocate buckets for blk addresses */
    LOG_ERROR_RETURN_IT_IF_NOT_OK(fi->Initialize(mSharedFd, resp.blockSize, mSharedFileInfo.maxBlkCountInSingleFile),
        "Failed to init file operation info");

    /* set the first block address */
    LOG_ERROR_RETURN_IT_IF_NOT_OK(fi->SetBlkLocalAddress(0, mSharedFileAddress + resp.dataBlock.offset),
        "Failed to set address in file operation info");

    fi->mCurrentBlockCount = 1U;
    LOG_TRACE("fi for fd " << resp.fd << ", " << fi->ToString());

    /* put the fd into map */
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Put(resp.fd, fi), "Failed to add file operation info in");

    fd = resp.fd;
    fi->mOpenFlag = flags;

    LOG_TRACE("Open file @" << FileCheckUtils::RemovePrefixPath(path) << " in memfs, fd " << resp.fd);
    return MFS_OK;
}

MResult MemFsClientOperation::OpenFile(const std::string &path, int32_t flags, int32_t &fd)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(!path.empty(), MFS_INVALID_PARAM);
    ASSERT_RETURN(path.size() < FS_PATH_MAX, MFS_INVALID_PARAM);
    ASSERT_RETURN(flags == O_RDONLY || flags == (O_CREAT | O_TRUNC | O_WRONLY), MFS_INVALID_PARAM);

    if ((flags & O_ACCMODE) == O_RDONLY) {
        return OpenFile4Read(path, flags, fd);
    } else {
        return OpenFile4Write(path, flags, fd);
    }
}

MResult MemFsClientOperation::Write(int32_t fd, uintptr_t data, uint64_t size)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(fd >= 0, MFS_INVALID_PARAM);
    ASSERT_RETURN(data != 0, MFS_INVALID_PARAM);

    MemFsFileOpInfoPtr fi;
    if (CheckWriteParam(fd, size, fi) != MFS_OK) {
        errno = EINVAL;
        return MFS_INVALID_PARAM;
    }

    auto ret = AllocateManyBlocksForWrite(fd, fi, size);
    if (ret != 0) {
        errno = -ret;
        return MFS_ERROR;
    }

    std::vector<Buffer> buffers;
    buffers.emplace_back(reinterpret_cast<void *>(data), size);

    auto multiThreadMinSize = MIN_WRITE_PARALLEL_SLICE_COUNT * mSharedFileInfo.writeParallelSlice;
    if (!mSharedFileInfo.writeParallelEnabled || size < multiThreadMinSize) {
        return WriteInSingleThread(fi, buffers, size);
    }

    return WriteInMultiThreads(fi, buffers, size);
}

MResult MemFsClientOperation::Write(int32_t fd, const std::vector<Buffer> &buffers, uint64_t &length)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(fd >= 0, MFS_INVALID_PARAM);
    ASSERT_RETURN(!buffers.empty(), MFS_INVALID_PARAM);

    auto totalSize = 0UL;
    std::for_each(buffers.begin(), buffers.end(), [&totalSize](const Buffer &buf) { totalSize += buf.size; });

    MemFsFileOpInfoPtr fi;
    if (CheckWriteParam(fd, buffers, totalSize, fi) != MFS_OK) {
        errno = EINVAL;
        return MFS_INVALID_PARAM;
    }

    auto ret = AllocateManyBlocksForWrite(fd, fi, totalSize);
    if (ret != 0) {
        errno = -ret;
        return MFS_ERROR;
    }

    auto multiThreadMinSize = MIN_WRITE_PARALLEL_SLICE_COUNT * mSharedFileInfo.writeParallelSlice;
    MResult writeResult;
    if (!mSharedFileInfo.writeParallelEnabled || totalSize < multiThreadMinSize) {
        writeResult = WriteInSingleThread(fi, buffers, totalSize);
    } else {
        writeResult = WriteInMultiThreads(fi, buffers, totalSize);
    }

    if (writeResult == MCode::MFS_OK) {
        length = totalSize;
    }

    return writeResult;
}

MResult MemFsClientOperation::Flush(int32_t fd)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(fd >= 0, MFS_INVALID_PARAM);

    LOG_TRACE("Start to flush");

    MemFsFileOpInfoPtr fi;
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");

    /* create request */
    TruncateFileReq req;
    req.fd = fd;
    req.flags = 0;
    req.size = fi->mPosition;
    ASSERT_RETURN(fi->mBlockSize != 0, MFS_ERROR);
    req.offsetInLastBlock = fi->mPosition % fi->mBlockSize;

    TruncateFileResp resp;
    auto result = mIpClient->SyncCall<TruncateFileReq, TruncateFileResp>(IPC_OP_TRUNCATE, req, resp);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(result, "Failed to flush for fd " << fd << ", result " << result);

    if (resp.result != MFS_OK) {
        LOG_ERROR("Failed to flush for fd " << fd << ", server error " << resp.result);
        return resp.result;
    }

    LOG_TRACE("Flush file " << fd << " in memfs, resp fd "
                            << ", size = " << req.size << ", fd" << resp.fd);
    return MFS_OK;
}

MResult MemFsClientOperation::ReadInSingleThread(int32_t fd, ock::memfs::MemFsFileOpInfoPtr &fi,
    std::vector<ReadBuffer> &buffers, uint64_t size)
{
    fi->mPosition = buffers[0].start;
    FileReadParam readParam{ buffers, 0, fi->mPosition, 0, 0, size };
    auto ret = ReadDataFromPosition(fi, readParam);
    if (ret != MFS_OK) {
        LOG_ERROR("single thread read data failed, the fd is:" << fd);
        return ret;
    }
    fi->mPosition += size;
    return MFS_OK;
}

MResult MemFsClientOperation::ReadInMultiThreads(int32_t fd, ock::memfs::MemFsFileOpInfoPtr &fi,
    std::vector<ReadBuffer> &buffers, uint64_t &length)
{
    auto sliceSize = mSharedFileInfo.writeParallelSlice;
    auto sliceCount = (length + mSharedFileInfo.writeParallelSlice - 1UL) / mSharedFileInfo.writeParallelSlice;
    if (sliceCount > mSharedFileInfo.writeParallelThreadNum) {
        sliceSize = length / mSharedFileInfo.writeParallelThreadNum;
        sliceCount = mSharedFileInfo.writeParallelThreadNum;
    }
    fi->mPosition = buffers[0].start;
    ParallelRwContext context{ fi, static_cast<int>(sliceCount) };
    const std::vector<FileReadParam> param{};
    auto taskParams = context.GenerateRwParameters(param, buffers, length, sliceSize);
    for (auto &taskParam : taskParams) {
        auto success = mWriteThreadPool->Execute([this, &context, &taskParam]() {
            auto ret = ReadDataFromPosition(context.fileInfo, taskParam);
            if (ret != 0) {
                LOG_ERROR("failed to copy:" << ret << " for task " << taskParam.index << " in " << context.totalCount);
                __sync_add_and_fetch(&context.failedCount, 1);
            }

            if (__sync_add_and_fetch(&context.finishedCount, 1) >= context.totalCount) {
                context.cond.notify_one();
            }
        });
        if (!success) {
            __sync_add_and_fetch(&context.failedCount, 1);
            __sync_add_and_fetch(&context.finishedCount, 1);
            LOG_ERROR("submit task(" << taskParam.index << ") failed");
        }
    }

    std::unique_lock<std::mutex> lockGuard(context.mutex);
    while (__sync_fetch_and_add(&context.finishedCount, 0) < context.totalCount) {
        context.cond.wait_for(lockGuard, std::chrono::milliseconds(1));
    }

    if (__sync_fetch_and_add(&context.failedCount, 0) > 0) {
        errno = ENOMEM;
        return MFS_ERROR;
    }

    fi->mPosition += length;
    return MFS_OK;
}

MResult MemFsClientOperation::Read(int32_t fd, uintptr_t data, uint64_t position, uint64_t size)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(fd >= 0, MFS_INVALID_PARAM);
    ASSERT_RETURN(data != 0, MFS_INVALID_PARAM);

    MemFsFileOpInfoPtr fi;
    if (CheckReadParam(fd, size, fi) != MFS_OK) {
        errno = EINVAL;
        return MFS_INVALID_PARAM;
    }

    std::vector<ReadBuffer> buffers;
    buffers.emplace_back(reinterpret_cast<void *>(data), position, size);

    auto multiThreadMinSize = MIN_WRITE_PARALLEL_SLICE_COUNT * mSharedFileInfo.writeParallelSlice;
    MResult ReadResult;
    if (!mSharedFileInfo.writeParallelEnabled || size < multiThreadMinSize) {
        ReadResult = ReadInSingleThread(fd, fi, buffers, size);
    } else {
        ReadResult = ReadInMultiThreads(fd, fi, buffers, size);
    }

    return ReadResult;
}

MResult MemFsClientOperation::Read(int32_t fd, std::vector<ReadBuffer> &buffers, uint64_t &length)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(fd >= 0, MFS_INVALID_PARAM);
    ASSERT_RETURN(!buffers.empty(), MFS_INVALID_PARAM);

    auto totalSize = 0UL;
    std::for_each(buffers.begin(), buffers.end(), [&totalSize](const ReadBuffer &buf) { totalSize += buf.size; });
    // after sort, tensors start is from small to large, which can be read sequentially
    std::sort(buffers.begin(), buffers.end(),
        [](const ReadBuffer &bufA, const ReadBuffer &bufB) { return bufA.start < bufB.start; });

    MemFsFileOpInfoPtr fi;
    if (CheckReadParam(fd, buffers, totalSize, fi) != MFS_OK) {
        errno = EINVAL;
        return MFS_INVALID_PARAM;
    }

    auto multiThreadMinSize = MIN_WRITE_PARALLEL_SLICE_COUNT * mSharedFileInfo.writeParallelSlice;
    MResult ReadResult;
    if (!mSharedFileInfo.writeParallelEnabled || totalSize < multiThreadMinSize) {
        ReadResult = ReadInSingleThread(fd, fi, buffers, totalSize);
    } else {
        ReadResult = ReadInMultiThreads(fd, fi, buffers, totalSize);
    }

    if (ReadResult == MCode::MFS_OK) {
        length = totalSize;
    }

    return ReadResult;
}

MResult MemFsClientOperation::Access(const std::string &path, int32_t mode)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    /* create request */
    AccessFileReq req;
    req.Path(path);
    req.mode = mode;

    AccessFileResp resp;
    auto result = mIpClient->SyncCall<AccessFileReq, AccessFileResp>(IPC_OP_ACCESS, req, resp);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(result,
        "Failed to access for path " << FileCheckUtils::RemovePrefixPath(path) << ", result " << result);

    if (resp.result != MFS_OK) {
        LOG_ERROR("Failed to access for path " << FileCheckUtils::RemovePrefixPath(path) << ", server error " <<
            resp.result);
        return resp.result;
    }

    LOG_TRACE("Access file " << FileCheckUtils::RemovePrefixPath(path) << " in memfs, result  " << resp.result);
    return MFS_OK;
}

MResult MemFsClientOperation::Seek(int32_t fd, int64_t offset, int32_t whence)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(fd >= 0, MFS_INVALID_PARAM);

    MemFsFileOpInfoPtr fi;
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");

    int64_t newPosition;
    switch (whence) {
        case SEEK_SET:
            newPosition = offset;
            break;
        case SEEK_CUR:
            newPosition = static_cast<int64_t>(fi->mPosition) + offset;
            break;
        case SEEK_END:
            newPosition = static_cast<int64_t>(fi->mFileSize) + offset;
            break;
        default:
            LOG_ERROR("Invalid param, whence " << whence);
            return MFS_INVALID_PARAM;
    }
    if (newPosition < 0 || newPosition > fi->mFileSize) {
        LOG_ERROR("Failed to seek for path " << fd << ", position " << fi->mPosition << ", file size" <<
            fi->mFileSize << ", offset " << offset << ", whence " << whence);
        return MFS_INVALID_PARAM;
    }

    fi->mPosition = newPosition;

    LOG_TRACE("Seed file " << fd << " in memfs, position " << newPosition);
    return MFS_OK;
}

MResult MemFsClientOperation::Tell(int32_t fd, int64_t &tell)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(fd >= 0, MFS_INVALID_PARAM);

    MemFsFileOpInfoPtr fi;
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");

    tell = static_cast<int64_t>(fi->mPosition);

    LOG_TRACE("Tell file " << fd << " in memfs, tell  " << tell);
    return MFS_OK;
}

MResult MemFsClientOperation::GetSize(int32_t fd, int64_t &size)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    MemFsFileOpInfoPtr fi;
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");

    size = static_cast<int64_t>(fi->mFileSize);

    LOG_TRACE("Get file size " << fd << " in memfs, size  " << size);
    return MFS_OK;
}

MResult MemFsClientOperation::Cntl(int32_t fd, int32_t cmd)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(fd >= 0, MFS_INVALID_PARAM);
    ASSERT_RETURN(cmd == F_GETFL, MFS_INVALID_PARAM);

    MemFsFileOpInfoPtr fi;
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");

    LOG_TRACE("Get file status " << fd << " in memfs");
    return MFS_OK;
}

MResult MemFsClientOperation::Close(int32_t fd, FileOpFile op)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(fd >= 0, MFS_INVALID_PARAM);

    MemFsFileOpInfoPtr fi;
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");

    /* create request */
    FlushSyncCloseFileReq req;
    req.fd = fd;
    req.op = op;
    req.flags = fi->mOpenFlag;
    req.fileSize = fi->mPosition;

    FlushSyncCloseFileResp resp;
    auto result =
        mIpClient->SyncCall<FlushSyncCloseFileReq, FlushSyncCloseFileResp>(IPC_OP_FLUSH_SYNC_CLOSE_FILE, req, resp);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(result, "Failed to close for fd " << fd << ", result " << result);

    if (resp.result != MFS_OK) {
        LOG_ERROR("Failed to close for fd " << fd << ", server error " << resp.result);
        return resp.result;
    }

    mFileMaps.Remove(fd);

    LOG_TRACE("Close file " << fd << " in memfs, resp fd "
                            << ", size = " << req.fileSize << ", fd" << resp.fd);
    return MFS_OK;
}

MResult MemFsClientOperation::Close(int32_t fd)
{
    return Close(fd, FOF_CLOSE);
}

MResult MemFsClientOperation::CloseWithUnlink(int32_t fd)
{
    return Close(fd, FOF_CLOSE_WITH_UNLINK);
}


MResult MemFsClientOperation::CheckBlocksValid(const uint64_t *blocks, uint32_t blockSize, uint32_t blockCount) const
{
    auto shareFileTotalSize = mSharedFileEndAddress - mSharedFileAddress;
    for (auto i = 0U; i < blockCount; i++) {
        if (blocks[i] >= shareFileTotalSize || blocks[i] + blockSize > shareFileTotalSize) {
            return MFS_ERROR;
        }
    }

    return MFS_OK;
}

MResult MemFsClientOperation::CheckWriteParam(int32_t fd, uint64_t size, MemFsFileOpInfoPtr &fi)
{
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");

    if ((fi->mOpenFlag & O_ACCMODE) == O_RDONLY) {
        LOG_ERROR("write fd(" << fd << ") is opened with readonly mode.");
        return MFS_INVALID_PARAM;
    }

    if (size > mSharedFileInfo.singleFileSize || fi->mPosition + size > mSharedFileInfo.singleFileSize) {
        LOG_ERROR("write fd(" << fd << ") with size(" << size << ") now pos(" << fi->mPosition << ") too large");
        return MFS_INVALID_PARAM;
    }

    return MFS_OK;
}

MResult MemFsClientOperation::CheckWriteParam(int32_t fd, const std::vector<Buffer> &buffers, uint64_t total,
    MemFsFileOpInfoPtr &fi)
{
    for (const auto &buf : buffers) {
        ASSERT_RETURN(buf.buffer != nullptr, MFS_INVALID_PARAM);
        ASSERT_RETURN(buf.size > 0, MFS_INVALID_PARAM);
    }

    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");
    if (((uint32_t)fi->mOpenFlag & O_ACCMODE) == O_RDONLY) {
        LOG_ERROR("writev fd(" << fd << ") is opened with readonly mode.");
        return MFS_INVALID_PARAM;
    }

    if (total > mSharedFileInfo.singleFileSize || fi->mPosition + total > mSharedFileInfo.singleFileSize) {
        LOG_ERROR("writev fd(" << fd << ") with size(" << total << ") now pos(" << fi->mPosition << ") too large");
        return MFS_INVALID_PARAM;
    }

    return MFS_OK;
}

uint64_t MemFsClientOperation::NeedAllocateBytesForWrite(const MemFsFileOpInfoPtr &fi, uint64_t size)
{
    auto currentSupportBytes = fi->mCurrentBlockCount * fi->mBlockSize;
    auto afterWritePosition = fi->mPosition + size;
    if (afterWritePosition <= currentSupportBytes) {
        return 0UL;
    }

    return afterWritePosition - currentSupportBytes;
}

MResult MemFsClientOperation::CheckReadParam(int32_t fd, uint64_t size, MemFsFileOpInfoPtr &fi)
{
    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");

    if ((fi->mOpenFlag & O_ACCMODE) != O_RDONLY) {
        LOG_ERROR("write fd(" << fd << ") is opened with readonly mode.");
        return MFS_INVALID_PARAM;
    }

    if (size > mSharedFileInfo.singleFileSize || fi->mPosition + size > mSharedFileInfo.singleFileSize) {
        LOG_ERROR("read fd(" << fd << ") with size(" << size << ") now pos(" << fi->mPosition << ") too large");
        return MFS_INVALID_PARAM;
    }

    return MFS_OK;
}

MResult MemFsClientOperation::CheckReadParam(int32_t fd, const std::vector<ReadBuffer> &buffers, uint64_t total,
    MemFsFileOpInfoPtr &fi)
{
    for (const auto &buf : buffers) {
        ASSERT_RETURN(buf.buffer != nullptr, MFS_INVALID_PARAM);
        ASSERT_RETURN(buf.size > 0, MFS_INVALID_PARAM);
        ASSERT_RETURN(buf.start > 0, MFS_INVALID_PARAM);
    }

    LOG_ERROR_RETURN_IT_IF_NOT_OK(mFileMaps.Get(fd, fi), "No fd found with " << fd << ", probably has not been opened");
    if (((uint32_t)fi->mOpenFlag & O_ACCMODE) != O_RDONLY) {
        LOG_ERROR("readMutil fd(" << fd << ") is opened with write only mode.");
        return MFS_INVALID_PARAM;
    }

    if (total > mSharedFileInfo.singleFileSize || fi->mPosition + total > mSharedFileInfo.singleFileSize) {
        LOG_ERROR("readv fd(" << fd << ") with size(" << total << ") now pos(" << fi->mPosition << ") too large");
        return MFS_INVALID_PARAM;
    }

    return MFS_OK;
}

int MemFsClientOperation::AllocateManyBlocksForWrite(int fd, ock::memfs::MemFsFileOpInfoPtr &fi, uint64_t size)
{
    auto needAllocateBytes = NeedAllocateBytesForWrite(fi, size);
    if (needAllocateBytes == 0UL) {
        return 0;
    }

    /* create request */
    AllocateMoreBlockReq req;
    AllocateMoreBlockResp *resp = nullptr;
    uint32_t rspLen = 0;
    req.fd = fd;
    req.flags = 0;
    req.size = needAllocateBytes;
    auto ret = mIpClient->SyncCall<AllocateMoreBlockReq, AllocateMoreBlockResp>(IPC_OP_ALLOCATE_MORE_BLOCK, req, &resp,
        rspLen);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(ret, "Failed to call alloc block for fd " << fd << ", result " << ret);
    if (resp == nullptr) {
        LOG_ERROR("Failed to open file @(allocate-hide), server response nullptr.");
        return -ECOMM;
    }
    ret = CheckOpenFile4ReadResponse("(allocate-hide)", resp, rspLen);
    if (ret != MFS_OK) {
        LOG_ERROR("Failed to call alloc block for fd " << fd);
        return ret;
    }

    for (auto i = 0U; i < resp->blockCount; i++) {
        fi->mBlkMappedAddress[fi->mCurrentBlockCount++] = mSharedFileAddress + resp->dataBlock[i];
    }

    free(resp);
    return 0;
}

int MemFsClientOperation::WriteDataFromPosition(const MemFsFileOpInfoPtr &fi, const FileWriteParam &param)
{
    auto currentPosition = param.fileOffset;
    auto leftBytes = param.writeSize;
    auto pageIndex = param.pageIndex;
    auto pageOffset = param.pageOffset;
    while (leftBytes > 0UL) {
        auto blockIndex = currentPosition / fi->mBlockSize;
        auto blockOffset = currentPosition % fi->mBlockSize;
        auto blockLeft = fi->mBlockSize - blockOffset;

        auto blockCopySize = std::min(blockLeft, leftBytes);
        auto blockCopyLeftSize = blockCopySize;
        auto copyDest = reinterpret_cast<void *>(fi->mBlkMappedAddress[blockIndex] + blockOffset);
        while (blockCopyLeftSize > 0UL) {
            auto vecCopySize = std::min(blockCopyLeftSize, (*param.buffers)[pageIndex].size - pageOffset);
            auto copySource = static_cast<const uint8_t *>((*param.buffers)[pageIndex].buffer) + pageOffset;
            auto err = memcpy_s(copyDest, blockCopyLeftSize, copySource, vecCopySize);
            if (err != EOK) {
                errno = ENOMEM;
                return -1;
            }

            copyDest = reinterpret_cast<void *>(reinterpret_cast<uintptr_t>(copyDest) + vecCopySize);
            blockCopyLeftSize -= vecCopySize;
            if (vecCopySize >= (*param.buffers)[pageIndex].size - pageOffset) {
                pageIndex++;
                pageOffset = 0;
            } else {
                pageOffset += vecCopySize;
            }
        }

        leftBytes -= blockCopySize;
        currentPosition += blockCopySize;
    }

    return 0;
}

int MemFsClientOperation::ReadDataFromPosition(const MemFsFileOpInfoPtr &fi, const FileReadParam &param)
{
    auto currentPosition = param.fileOffset;
    auto leftBytes = param.readSize;
    auto pageIndex = param.pageIndex;
    auto pageOffset = param.pageOffset;
    while (leftBytes > 0UL) {
        auto blockIndex = currentPosition / fi->mBlockSize;
        auto blockOffset = currentPosition % fi->mBlockSize;
        auto blockLeft = fi->mBlockSize - blockOffset;

        auto blockCopySize = std::min(blockLeft, leftBytes);
        auto blockCopyLeftSize = blockCopySize;
        auto copySource = reinterpret_cast<const void *>(fi->mBlkMappedAddress[blockIndex] + blockOffset);
        while (blockCopyLeftSize > 0UL) {
            auto curBufferLeftSize = (*param.buffers)[pageIndex].size - pageOffset;
            auto vecCopySize = std::min(blockCopyLeftSize, curBufferLeftSize);
            auto copyDest = static_cast<uint8_t *>((*param.buffers)[pageIndex].buffer) + pageOffset;
            std::copy_n(static_cast<const uint8_t *>(copySource), vecCopySize, copyDest);
            copySource = reinterpret_cast<const void *>(reinterpret_cast<uintptr_t>(copySource) + vecCopySize);
            blockCopyLeftSize -= vecCopySize;
            if (vecCopySize >= curBufferLeftSize) {
                pageIndex++;
                pageOffset = 0;
            } else {
                pageOffset += vecCopySize;
            }
        }

        leftBytes -= blockCopySize;
        currentPosition += blockCopySize;
    }

    return 0;
}

int MemFsClientOperation::WriteInSingleThread(ock::memfs::MemFsFileOpInfoPtr &fi, const std::vector<Buffer> &buffers,
    uint64_t length)
{
    FileWriteParam writeParam{ buffers, 0, fi->mPosition, 0, 0, length };
    auto ret = WriteDataFromPosition(fi, writeParam);
    if (ret != 0) {
        LOG_ERROR("failed to copy data : " << ret);
        return ret;
    }

    fi->mPosition += length;
    return 0;
}

int MemFsClientOperation::WriteInMultiThreads(ock::memfs::MemFsFileOpInfoPtr &fi, const std::vector<Buffer> &buffers,
    uint64_t length)
{
    auto sliceSize = mSharedFileInfo.writeParallelSlice;
    auto sliceCount = (length + mSharedFileInfo.writeParallelSlice - 1UL) / mSharedFileInfo.writeParallelSlice;
    if (sliceCount > mSharedFileInfo.writeParallelThreadNum) {
        sliceSize = length / mSharedFileInfo.writeParallelThreadNum;
        sliceCount = mSharedFileInfo.writeParallelThreadNum;
    }

    ParallelRwContext context{ fi, static_cast<int>(sliceCount) };
    const std::vector<FileWriteParam> param{};
    auto taskParams = context.GenerateRwParameters(param, buffers, length, sliceSize);
    for (auto &taskParam : taskParams) {
        auto success = mWriteThreadPool->Execute([this, &context, &taskParam]() {
            auto ret = WriteDataFromPosition(context.fileInfo, taskParam);
            if (ret != 0) {
                LOG_ERROR("failed to copy:" << ret << " for task " << taskParam.index << " in " << context.totalCount);
                __sync_add_and_fetch(&context.failedCount, 1);
            }

            if (__sync_add_and_fetch(&context.finishedCount, 1) >= context.totalCount) {
                context.cond.notify_one();
            }
        });
        if (!success) {
            __sync_add_and_fetch(&context.failedCount, 1);
            __sync_add_and_fetch(&context.finishedCount, 1);
            LOG_ERROR("submit task(" << taskParam.index << ") failed");
        }
    }

    std::unique_lock<std::mutex> lockGuard(context.mutex);
    while (__sync_fetch_and_add(&context.finishedCount, 0) < context.totalCount) {
        context.cond.wait_for(lockGuard, std::chrono::milliseconds(1));
    }

    if (__sync_fetch_and_add(&context.failedCount, 0) > 0) {
        errno = ENOMEM;
        return MFS_ERROR;
    }

    fi->mPosition += length;
    return MFS_OK;
}

int MemFsClientOperation::ConnectedProcess()
{
    /* get shared file from server */
    auto result = GetSharedFileInfo();
    if (result != MFS_OK) {
        mIpClient->Stop();
        mIpClient = nullptr;
        LOG_ERROR("Failed to receive");
        return result;
    }

    /* receive fd */
    result = mIpClient->ReceiveFD(mSharedFd);
    if (result != MFS_OK) {
        mIpClient->Stop();
        mIpClient = nullptr;
        LOG_ERROR("Failed to receive");
        return result;
    }

    /* check size */
    auto size = mSharedFileInfo.singleFileSize;

    /* mmap */
    auto mappedAddress = mmap(nullptr, size, PROT_READ | PROT_WRITE, MAP_SHARED, mSharedFd, 0);
    if (mappedAddress == MAP_FAILED) {
        close(mSharedFd);
        mSharedFd = -1;
        mIpClient->Stop();
        mIpClient = nullptr;
        LOG_ERROR("Failed to mmap file, error " << strerror(errno));
        return MFS_ERROR;
    }

    mSharedFileAddress = reinterpret_cast<uintptr_t>(mappedAddress);
    mSharedFileEndAddress = mSharedFileAddress + size;
    return MFS_OK;
}

void MemFsClientOperation::DisconnectedProcess()
{
    mFileMaps.Clear();
    if (mSharedFd >= 0) {
        munmap((void *)mSharedFileAddress, mSharedFileEndAddress - mSharedFileAddress);
        close(mSharedFd);
        mSharedFd = -1;
    }
    mWriteThreadPool = nullptr;
}


int MemFsClientOperation::RegisterAtFork()
{
    if (!mAtForkRegister) {
        auto ret = pthread_atfork(PrepareForked, ParentProcessForked, ChildProcessForked);
        if (ret != 0) {
            LOG_INFO("register atfork failed: " << ret << ":" << strerror(ret));
            return MFS_ERROR;
        }

        mAtForkRegister = true;
        LOG_INFO("register at fork success.");

#ifndef NDEBUG
        ret = RegisterSignalHandler();
        LOG_INFO("register signal handler return : " << ret);
#endif
    }

    return MFS_OK;
}

void MemFsClientOperation::PrepareForked()
{
    LOG_WARN("MindIO Client SDK fork now, pid = " << getpid());
}

void MemFsClientOperation::ParentProcessForked()
{
    LOG_WARN("MindIO Client SDK fork finished at parent, pid = " << getpid());
}

void MemFsClientOperation::ChildProcessForked()
{
    LOG_WARN("MindIO Client SDK fork finished at child, pid = " << getpid());
    auto instance = MemFsClientOperation::Instance();
    if (instance == nullptr) {
        return;
    }

    instance->mForkChild = true;
}

int MemFsClientOperation::RegisterSignalHandler()
{
    struct sigaction sa = { { nullptr } };

    sigemptyset(&(sa.sa_mask));
    sa.sa_flags = SA_NODEFER | SA_ONSTACK | SA_RESETHAND | SA_SIGINFO;
    sa.sa_sigaction = SignalHandler;
    if (sigaction(SIGSEGV, &sa, nullptr) < 0) {
        LOG_ERROR("register SIGSEGV signal handler failed : " << errno << " : " << strerror(errno));
        return -1;
    }

    if (sigaction(SIGBUS, &sa, nullptr) < 0) {
        LOG_ERROR("register SIGBUS signal handler failed : " << errno << " : " << strerror(errno));
        return -1;
    }

    LOG_INFO("register signal handler success.");
    return 0;
}

void MemFsClientOperation::SignalHandler(int sig, siginfo_t *info, void *param)
{
    auto tid = pthread_self();
    LOG_ERROR("segmentation fault @pthread " << tid << " signal caught. signo=" << info->si_signo << ", error=" <<
        info->si_errno << ", code=" << info->si_code << ", pid=" << info->si_pid << ", uid=" << info->si_uid <<
        ", status=" << info->si_status << ", addr=" << std::hex << info->si_addr);
}

MResult MemFsClientOperation::RealPreloadFile(const std::string &path)
{
    /* create request */
    PreloadFileReq req;
    if (!req.FileName(path)) {
        LOG_ERROR("file path: " << FileCheckUtils::RemovePrefixPath(path) << " invalid.");
        return MFS_INVALID_PARAM;
    }
    /* call server */
    PreloadFileResp resp;
    auto result = mIpClient->SyncCall<PreloadFileReq, PreloadFileResp>(IPC_OP_PRELOAD_FILE, req, resp);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(result,
        "Failed to call server to preload file @" << FileCheckUtils::RemovePrefixPath(path) << ", result " << result);

    LOG_TRACE("Preload file response result = " << resp.result);

    if (resp.result != MFS_OK) {
        LOG_ERROR("Failed to preload file @" << FileCheckUtils::RemovePrefixPath(path) << ", server error " <<
            resp.result);
        return resp.result;
    }

    return MFS_OK;
}

MResult MemFsClientOperation::PreloadFile(const std::string &path)
{
    ASSERT_RETURN(mInited, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(GetServiceStatus(), MFS_NOT_INITIALIZED);
    ASSERT_RETURN(!path.empty(), MFS_INVALID_PARAM);
    ASSERT_RETURN(path.size() < FS_PATH_MAX, MFS_INVALID_PARAM);

    return RealPreloadFile(path);
}

MResult MemFsClientOperation::CheckBackgroundTask()
{
    if (!mInited || !GetServiceStatus()) {
        return MFS_OK;
    }

    /* create request */
    CheckBackgroundTaskReq req;

    /* call server */
    CheckBackgroundTaskResp resp;

    uint32_t retryTimes = 0;
    while (true) {
        // check for 30s timeout
        auto result = mIpClient->SyncCall<CheckBackgroundTaskReq, CheckBackgroundTaskResp>(IPC_OP_CHECK_BACKGROUND_TASK,
            req, resp);
        LOG_ERROR_RETURN_IT_IF_NOT_OK(result, "Failed to check background task, result " << result);
        if (resp.result == MFS_OK) {
            return MFS_OK;
        }

        if (resp.result == -EAGAIN) {
            retryTimes++;
            if (retryTimes >= MAX_CHECK_BACKGROUND_TASK_TIMES) {
                LOG_ERROR("Check background task timeout.");
                return -ETIMEDOUT;
            }
            continue;
        }

        return resp.result;
    }
}