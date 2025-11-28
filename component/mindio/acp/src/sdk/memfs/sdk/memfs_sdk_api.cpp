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
#include "memfs_sdk_api.h"

#include "common_includes.h"
#include "fs_operation.h"
#include "memfs_check_dir_event.h"

using namespace ock::memfs;

static constexpr auto EVENT_TIME_INTERVAL_MS = 1000;
static MemFsClientOperation *g_fsOperation = MemFsClientOperation::Instance();

namespace ock {
namespace memfs {
/**
 * @brief Initialize the mem fs client
 */
MResult MemFsClientInitialize(const std::map<std::string, std::string> &serverInfoParam)
{
    ASSERT_RETURN(g_fsOperation != nullptr, MFS_NOT_INITIALIZED);
    IpcClientConfig config;
    return g_fsOperation->Initialize(config, serverInfoParam);
}

int32_t MemFsClientInitialize(const ClientInitParam &param, const std::map<std::string, std::string> &serverInfoParam)
{
    ASSERT_RETURN(g_fsOperation != nullptr, MFS_NOT_INITIALIZED);
    if (!param.ipcTlsEnabled) {
        return MemFsClientInitialize(serverInfoParam);
    }

    IpcClientConfig config(param.ipcTlsCertPath, param.ipcTlsCaPath, param.ipcTlsCrlPath, param.ipcTlsPriKeyPath,
        param.ipcTlsPasswordPath, param.pmtPath);
    return g_fsOperation->Initialize(config, serverInfoParam);
}

/**
 * @brief Un-initialize the mem fs client
 */
void MemFsClientUnInitialize()
{
    ASSERT_RET_VOID(g_fsOperation != nullptr);
    return g_fsOperation->UnInitialize();
}

bool MemFsIsForkedProcess()
{
    ASSERT_RETURN(g_fsOperation != nullptr, false);
    return g_fsOperation->IsForkedProcess();
}

int32_t MemFsMkDir(const char *path, int32_t flags, bool recursive)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);
    ASSERT_RETURN(path != nullptr, -1);
    ASSERT_RETURN(strlen(path) < FS_PATH_MAX, -1);

    return g_fsOperation->MakeDir(path, flags, recursive) == 0 ? 0 : -1;
}

int32_t MemFsLinkFile(const char *source, const char *target)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);
    ASSERT_RETURN(source != nullptr, -1);
    ASSERT_RETURN(target != nullptr, -1);
    ASSERT_RETURN(strlen(source) < FS_PATH_MAX, -1);
    ASSERT_RETURN(strlen(target) < FS_PATH_MAX, -1);

    auto ret = g_fsOperation->LinkFile(source, target);
    if (ret == 0) {
        return 0;
    }

    errno = -ret;
    return -1;
}

int32_t MemFsRenameFile(const char *source, const char *target, uint32_t flags)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);
    ASSERT_RETURN(source != nullptr, -1);
    ASSERT_RETURN(target != nullptr, -1);
    ASSERT_RETURN(strlen(source) < FS_PATH_MAX, -1);
    ASSERT_RETURN(strlen(target) < FS_PATH_MAX, -1);

    auto ret = g_fsOperation->RenameFile(source, target, flags);
    if (ret == 0) {
        return 0;
    }

    errno = -ret;
    return -1;
}

/**
 * @brief Open a file in mem fs
 *
 * @param path             [in] path of the file
 * @param flags            [in] flags of open
 *
 * @return fd if ok, -1 means failed
 */
int32_t MemFsOpenFile(const char *path, int32_t flags, int32_t &retVaule)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);
    ASSERT_RETURN(path != nullptr, -1);
    ASSERT_RETURN(strlen(path) < FS_PATH_MAX, -1);

    int32_t fd = -1;
    auto ret = g_fsOperation->OpenFile(std::string(path), flags, fd);
    if (ret != MFS_OK) {
        retVaule = -ret;
    }
    return fd;
}

/**
 * @brief Write data into file descriptor in blocking way
 *
 * @param fd               [in] the file descriptor allocated by mem fs
 * @param data             [in] the ptr of data to be written
 * @param size             [in] the size of data
 *
 * @return size written, -1 means error
 */
int64_t MemFsWrite(int32_t fd, uintptr_t data, size_t size)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    return g_fsOperation->Write(fd, data, size) != 0 ? -1 : static_cast<int64_t>(size);
}

/**
 * @brief writes buffers of data described by iov to the file associated with the file descriptor fd ("gather output"),
 * it works just like `ssize_t MemFsWrite(int32_t fd, uintptr_t data, size_t size)` except that multiple buffers are
 * written out.
 *
 * @param fd               [in] the file descriptor allocated by mem fs
 * @param buffers          [in] buffers of data to be written
 *
 * @return size written, -1 means error
 */
int64_t MemFsWriteV(int32_t fd, const std::vector<ock::memfs::Buffer> &buffers)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);
    uint64_t writeSize = 0UL;
    auto ret = g_fsOperation->Write(fd, buffers, writeSize);
    return ret == 0 ? static_cast<int64_t>(writeSize) : -1L;
}

/**
 * @brief read buffers of data described by iov to the file associated with the
 * file descriptor fd ("gather output"), it works just like `ssize_t
 * MemFsRead(int32_t fd, uintptr_t data, uint64_t position, size_t size)` except
 * that multiple buffers are read.
 *
 * @param fd               [in] the file descriptor allocated by mem fs
 * @param readVector       [in] buffers of data to be read
 *
 * @return size read, -1 means error
 */
int64_t MemFsReadM(int32_t fd, std::vector<ock::memfs::ReadBuffer> &readVector)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);
    uint64_t readSize = 0UL;
    auto ret = g_fsOperation->Read(fd, readVector, readSize);
    return ret == 0 ? static_cast<int64_t>(readSize) : -1L;
}

/**
 * @brief Read data from file descriptor in blocking way
 *
 * @param fd               [in] the file descriptor allocated by mem fs
 * @param data             [in] the ptr of data to be read
 * @param size             [in] the size of data
 *
 * @return size read, -1 means error
 */
int64_t MemFsRead(int32_t fd, uintptr_t data, uint64_t position, size_t size)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    return g_fsOperation->Read(fd, data, position, size) != 0 ? -1 : static_cast<int64_t>(size);
}

/**
 * @brief Flush the data to file
 *
 * @param fd               [in] the fd to be flushed
 *
 * @return 0 means ok
 */
MResult MemFsFlush(int32_t fd)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    return g_fsOperation->Flush(fd) == 0 ? 0 : -1;
}

/**
 * @brief Checks whether the calling process can access the file pathname.
 *
 * @param path             [in] path of the file
 * @param mode             [in] The mode specifies the accessibility check(s) to be performed,
 *                              and is either the value F_OK, or a mask consisting of the bitwise
 *                              OR of one or more of R_OK, W_OK, and X_OK.  F_OK tests for the
 *                              existence of the file.  R_OK, W_OK, and X_OK test whether the
 *                              file exists and grants read, write, and execute permissions,
 *                              respectively.
 *
 * @return On success (all requested permissions granted, or mode is F_OK
       and the file exists), zero is returned.  On error (at least one
       bit in mode asked for a permission that is denied, or mode is
       F_OK and the file does not exist, or some other error occurred),
       -1 is returned.
 */
int32_t MemFsAccess(const char *path, int32_t mode)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);
    ASSERT_RETURN(path != nullptr, -1);
    ASSERT_RETURN(strlen(path) < FS_PATH_MAX, -1);

    return g_fsOperation->Access(path, mode) == 0 ? 0 : -1;
}

/**
 * @brief This function shall set the file-position indicator for the stream pointed to by stream,
 * shall be obtained by adding offset to the position specified by whence
 *
 * @param path             [in] path of the file
 * @param offset           [in] The offset value from whence.
 * @param whence           [in] The specified point is the beginning of the file for SEEK_SET,
 * the current value of the file-position indicator for SEEK_CUR,
 * or end-of-file for SEEK_END
 * @return Return 0 if they succeed, otherwise, they shall return -1.
 */
int32_t MemFsSeek(int32_t fd, off_t offset, int32_t whence)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    return g_fsOperation->Seek(fd, offset, whence) == 0 ? 0 : -1;
}

/**
 * @brief This function shall obtain the current value of the file-position indicator for the stream pointed to by
 * stream.
 *
 * @return Return the current value of the file-position indicator for the stream measured in bytes from the beginning
 * of the file if they succeed, otherwise, return -1.
 */
off_t MemFsTell(int32_t fd)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    off_t tell;
    return g_fsOperation->Tell(fd, tell) == 0 ? tell : -1;
}

/**
 * @brief This function shall obtain the current size of this file.
 * stream.
 *
 * @return Return the size of the file if they succeed, otherwise, return -1.
 */
ssize_t MemFsGetSize(int32_t fd)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    ssize_t size;
    return g_fsOperation->GetSize(fd, size) == 0 ? size : -1;
}

/**
 * @brief This function do the file control operation described by cmd on FD.
 *
 * @param fd             [in] the fd to be operated
 * @param cmd            [in] the command to be operated on the fd
 * @return Return 0 if they succeed, otherwise, they shall return -1.
 */
int32_t MemFsCntl(int32_t fd, int32_t cmd)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    return g_fsOperation->Cntl(fd, cmd) == 0 ? 0 : -1;
}

/**
 * @brief Close the file
 *
 * @param fd               [in] the fd be closed
 * @return
 */
int32_t MemFsClose(int32_t fd)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    auto ret = g_fsOperation->Close(fd);
    if (ret != MFS_OK) {
        errno = -ret;
        return -1;
    }

    return 0;
}

/**
 * @brief Close the file with unlink when write failed
 *
 * @param fd               [in] the fd be closed
 * @return
 */
int32_t MemFsCloseWithUnlink(int32_t fd)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    auto ret = g_fsOperation->CloseWithUnlink(fd);
    if (ret != MFS_OK) {
        errno = -ret;
        return -1;
    }

    return 0;
}


int32_t MemfsRegisterWatchDir(const std::unordered_map<std::string, uint64_t> &dirInfo, uint64_t timeoutSec,
    const std::function<void(uint64_t, int)> &callback, uint64_t &eventId)
{
    DirectoriesInfo directoriesInfo;
    for (auto &info : dirInfo) {
        directoriesInfo.emplace_back(info.first, info.second);
    }

    auto event = std::make_shared<MemfsCheckDirEvent>("watch-backup-state", timeoutSec, EVENT_TIME_INTERVAL_MS,
        directoriesInfo, callback);
    return MemfsEventManager::GetInstance().SubmitEvent(event, eventId);
}

int32_t MemfsPreloadFile(const char *path)
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);
    ASSERT_RETURN(path != nullptr, -1);
    ASSERT_RETURN(strlen(path) < FS_PATH_MAX, -1);

    auto ret = g_fsOperation->PreloadFile(path);
    if (ret != MFS_OK) {
        errno = -ret;
        return -1;
    }

    return 0;
}

int32_t MemFsCheckBackgroundTask()
{
    ASSERT_RETURN(g_fsOperation != nullptr, -1);

    auto ret = g_fsOperation->CheckBackgroundTask();
    if (ret != MFS_OK) {
        errno = -ret;
        return -1;
    }

    MemfsEventManager::GetInstance().WaitAllEventsFinished();
    return 0;
}
} // namespace memfs
} // namespace ock
