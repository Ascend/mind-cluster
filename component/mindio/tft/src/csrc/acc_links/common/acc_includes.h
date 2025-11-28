/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#ifndef ACC_LINKS_ACC_INCLUDES_H
#define ACC_LINKS_ACC_INCLUDES_H

#include <arpa/inet.h>
#include <atomic>
#include <cstdint>
#include <fcntl.h>
#include <functional>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <set>
#include <string>
#include <sys/epoll.h>
#include <sys/poll.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <sys/un.h>
#include <thread>
#include <unordered_map>
#include <utility>

#include "acc_def.h"
#include "acc_tcp_link.h"
#include "acc_tcp_request_context.h"
#include "acc_tcp_server.h"
#include "acc_tcp_shared_buf.h"

#include "acc_out_logger.h"

namespace ock {
namespace acc {
using Result = int32_t;

/**
 * @brief New an object return with ref object
 *
 * @param args             [in] args of object
 * @return Ref object, if new failed internal, an empty Ref object will be returned
 */
template <typename C, typename... ARGS>
inline ock::ttp::Ref<C> AccMakeRef(ARGS... args)
{
    return new (std::nothrow) C(args...);
}

#ifndef LIKELY
#define LIKELY(x) (__builtin_expect(!!(x), 1) != 0)
#endif

#ifndef UNLIKELY
#define UNLIKELY(x) (__builtin_expect(!!(x), 0) != 0)
#endif

/**
 * @brief Close fd in safe way, to avoid double close
 *
 * @param fd               [in] fd to be closed
 */
inline void SafeCloseFd(int &fd, bool needShutdown = true)
{
    if (UNLIKELY(fd < 0)) {
        return;
    }

    auto tmpFd = fd;
    if (__sync_bool_compare_and_swap(&fd, tmpFd, -1)) {
        if (needShutdown) {
            shutdown(tmpFd, SHUT_RDWR);
        }
        close(tmpFd);
    }
}
}  // namespace acc
}  // namespace ock

#endif  // ACC_LINKS_ACC_INCLUDES_H
