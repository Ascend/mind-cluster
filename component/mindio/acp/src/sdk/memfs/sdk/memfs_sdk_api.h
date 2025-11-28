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
#ifndef OCK_MEMFS_CORE_MEMFS_SDK_API_H
#define OCK_MEMFS_CORE_MEMFS_SDK_API_H

#include <cstdint>
#include <cstdlib>
#include <functional>
#include <list>
#include <map>
#include <string>
#include <unordered_map>
#include <vector>

#include "memfs_sdk_types.h"

namespace ock {
namespace memfs {
struct ClientInitParam {
    bool ipcTlsEnabled{ false };
    std::string ipcTlsCertPath;
    std::string ipcTlsCrlPath;
    std::string ipcTlsCaPath;
    std::string ipcTlsPriKeyPath;
    std::string ipcTlsPasswordPath;
    std::string pmtPath;
};

/**
 * @brief Initialize the mem fs client
 */
int32_t MemFsClientInitialize(const std::map<std::string, std::string> &serverInfoParam);

/**
 * @brief Initialize the mem fs client
 */
int32_t MemFsClientInitialize(const ClientInitParam &param, const std::map<std::string, std::string> &serverInfoParam);

/**
 * @brief Un-initialize the mem fs client
 */
void MemFsClientUnInitialize();

/**
 * @brief mem fs client is forked
 */
bool MemFsIsForkedProcess();

/**
 * @brief Make dir
 *
 * @param path             [in] path of the dir
 * @param flags            [in] flags
 * @param recursive        [in] recursive
 *
 * @return 0 successfully, otherwise, they shall return -1.
 */
int32_t MemFsMkDir(const char *path, int32_t flags, bool recursive = false);

/**
 * @brief make a new name for a file, creates a new link (also known as a hard link) to an existing file
 * @param source [in] old path name
 * @param target [in] new path name
 * @return On success, zero is returned.  On error, -1 is returned, and errno is set appropriately.
 */
int32_t MemFsLinkFile(const char *source, const char *target);

/**
 * @brief change the name or location of a file
 * @param source [in] old path name
 * @param target [in] new path name
 * @param flags [in] flags of rename
 * @return On success, zero is returned.  On error, -1 is returned, and errno is set appropriately.
 */
int32_t MemFsRenameFile(const char *source, const char *target, uint32_t flags);

/**
 * @brief Open a file in mem fs
 *
 * @param path             [in] path of the file
 * @param flags            [in] flags of open
 *
 * @return fd if ok, -1 means failed
 */
int32_t MemFsOpenFile(const char *path, int32_t flags, int32_t &retVaule = errno);

/**
 * @brief Write data into file descriptor in blocking way
 *
 * @param fd               [in] the file descriptor allocated by mem fs
 * @param data             [in] the ptr of data to be written
 * @param size             [in] the size of data
 *
 * @return size written, -1 means error
 */
int64_t MemFsWrite(int32_t fd, uintptr_t data, size_t size);

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
int64_t MemFsWriteV(int32_t fd, const std::vector<ock::memfs::Buffer> &buffers);

/**
 * @brief Read data from file descriptor in blocking way
 *
 * @param fd               [in] the file descriptor allocated by mem fs
 * @param data             [in] the ptr of data to be read
 * @param size             [in] the size of data
 *
 * @return size read, -1 means error
 */
int64_t MemFsRead(int32_t fd, uintptr_t data, uint64_t position, size_t size);

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
int64_t MemFsReadM(int32_t fd, std::vector<ock::memfs::ReadBuffer> &readVector);

/**
 * @brief Flush the data to file
 *
 * @param fd               [in] the fd to be flushed
 *
 * @return 0 means ok
 */
int32_t MemFsFlush(int32_t fd);

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
int32_t MemFsAccess(const char *path, int32_t mode);

/**
 * @brief This function shall set the file-position indicator for the stream pointed to by stream,
 * shall be obtained by adding offset to the position specified by whence
 *
 * @param fd             [in] fd of the file
 * @param offset           [in] The offset value from whence.
 * @param whence           [in] The specified point is the beginning of the file for SEEK_SET,
 * the current value of the file-position indicator for SEEK_CUR,
 * or end-of-file for SEEK_END
 * @return Return 0 if they succeed, otherwise, they shall return -1.
 */
int32_t MemFsSeek(int32_t fd, off_t offset, int32_t whence);

/**
 * @brief This function shall obtain the current value of the file-position indicator for the stream pointed to by
 * stream.
 *
 * @param fd             [in] fd of the file
 *
 * @return Return the current value of the file-position indicator for the stream measured in bytes from the beginning
 * of the file if they succeed, otherwise, return -1.
 */
off_t MemFsTell(int32_t fd);

/**
 * @brief This function shall obtain the current size of this file.
 * stream.
 *
 * * @param fd             [in] fd of the file
 *
 * @return Return the size of the file if they succeed, otherwise, return -1.
 */
ssize_t MemFsGetSize(int32_t fd);

/**
 * @brief This function do the file control operation described by cmd on FD.
 *
 * @param fd             [in] the fd to be operated
 * @param cmd            [in] the command to be operated on the fd
 * @return Return 0 if they succeed, otherwise, they shall return -1.
 */
int32_t MemFsCntl(int32_t fd, int32_t cmd);

/**
 * @brief Close the file
 *
 * @param fd               [in] the fd be closed
 * @return
 */
int32_t MemFsClose(int32_t fd);

/**
 * @brief Close the file with unlink when write failed
 *
 * @param fd               [in] the fd be closed
 * @return
 */
int32_t MemFsCloseWithUnlink(int32_t fd);


/**
 * @brief 注册一个事件，观测目录是否完成要求的备份文件数目
 * @param dirInfo 指定的多个观测目录，每个目录都指定要求的文件数目
 * @param timeoutSec 最长等待结果时间，单位秒
 * @param callback 指定的回调函数，回调有两个参数: uint64
 * eventId参见下一个传出参数，int result表示结果，0为成功，非0失败
 * @param eventId 本次注册事件生成的唯一event id，用于在回调中区分是哪个事件
 * @return 注册结果， 0表示成功，非0表示失败，失败时不会回调
 */
int32_t MemfsRegisterWatchDir(const std::unordered_map<std::string, uint64_t> &dirInfo, uint64_t timeoutSec,
    const std::function<void(uint64_t, int)> &callback, uint64_t &eventId);

/**
 * 预加载文件
 * @param path 文件路径
 * @return 0表示成功，非0表示失败
 */
int32_t MemfsPreloadFile(const char *path);

/**
 * @brief Check the background tasks finish or not.
 *
 * @return Return 0 if all tasks finish, otherwise, they shall return -1.
 */
int32_t MemFsCheckBackgroundTask();

} // namespace memfs
} // namespace ock

#endif // OCK_MEMFS_CORE_MEMFS_SDK_API_H
