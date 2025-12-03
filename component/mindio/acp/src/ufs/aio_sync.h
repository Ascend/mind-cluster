/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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

#ifndef OCK_DFS_AIO_SYNC_H
#define OCK_DFS_AIO_SYNC_H

#include <aio.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <cstdint>
#include <string>
#include <thread>
#include <atomic>
#include "securec.h"
#include "ufs_log.h"
namespace ock {
namespace ufs {
static constexpr auto FLUSH_DISK_CHECK_PERIOD = 10;

void UpdateStgMtime(const std::string &filePath) noexcept
{
    std::atomic<int8_t> timeToChange {0};
    struct timespec times[2];
    std::string stagePath = filePath;
    stagePath.append(".m.stg");

    timeToChange.fetch_add(1);
    // Async I/O status will be checked every 10 ms, here modify mtime every 100 ms.
    if (timeToChange.load()== 10) {
        times[0].tv_sec = UTIME_OMIT;
        times[0].tv_nsec = UTIME_OMIT;

        times[1].tv_sec = UTIME_NOW;
        times[1].tv_nsec = UTIME_NOW;

        if (utimensat(AT_FDCWD, stagePath.c_str(), times, 0) == -1) {
            UFS_LOG_WARN("update file(" << stagePath.c_str() << ") mtime failed(" << errno << " : " <<
                strerror(errno) << ")");
        }
        timeToChange.store(0);
        UFS_LOG_DEBUG("time to UpdateStgMtime");
    }
}

int AioSync(int fileDesc, const std::string &path) noexcept
{
    struct aiocb cb {};

    auto err = memset_s(&cb, sizeof(struct aiocb), 0, sizeof(struct aiocb));
    if (err != EOK) {
        UFS_LOG_ERROR("memset aiocb failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    cb.aio_fildes = fileDesc;
    cb.aio_offset = 0;
    while (aio_fsync(O_SYNC, &cb) == -1) {
        if (errno == EAGAIN) {
            UFS_LOG_DEBUG("errno == EAGAIN, aio_fsync try again.");
            // wait for 10ms
            std::this_thread::sleep_for(std::chrono::milliseconds(FLUSH_DISK_CHECK_PERIOD));
            continue;
        }
        UFS_LOG_ERROR("aio_fsync failed(" << errno << " : " << strerror(errno) << ")");
        close(fileDesc);
        _exit(EXIT_FAILURE);
    }

    int ret;
    // Check the async I/O status every 10 ms.
    while ((ret = aio_error(&cb)) == EINPROGRESS) {
        UpdateStgMtime(path);
        std::this_thread::sleep_for(std::chrono::milliseconds(FLUSH_DISK_CHECK_PERIOD));
    }

    UFS_LOG_INFO("aio_fsync finished, ret = (" << ret << ")");
    return ret;
}
}
}

#endif // OCK_DFS_AIO_SYNC_H
