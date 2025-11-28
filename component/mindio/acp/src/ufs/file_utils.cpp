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
#include <unistd.h>
#include "file_utils.h"

using namespace ock::ufs::utils;

int64_t FileUtils::ReadFull(int fd, uint8_t *buf, uint64_t len) noexcept
{
    auto left = static_cast<int64_t>(len);
    auto finished = 0L;
    while (left > 0L) {
        auto count = read(fd, buf + finished, left);
        if (count < 0) {
            return -1L;
        }

        if (count == 0L) {
            break;
        }

        left -= count;
        finished += count;
    }
    return finished;
}

int64_t FileUtils::WriteFull(int fd, const uint8_t *buf, uint64_t len) noexcept
{
    auto left = static_cast<int64_t>(len);
    auto finished = 0L;
    while (left > 0L) {
        auto count = write(fd, buf + finished, left);
        if (count < 0) {
            return -1L;
        }

        if (count == 0L) {
            break;
        }

        left -= count;
        finished += count;
    }
    return finished;
}