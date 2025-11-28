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
#include "ufs_api.h"

using namespace ock::ufs;

int64_t InputStream::TotalSize() noexcept
{
    return -1L;
}

int64_t InputStream::Read(uint8_t *buf, uint64_t count) noexcept
{
    auto total = static_cast<int64_t>(count);
    for (auto i = 0L; i < total; i++) {
        auto ret = Read(buf[i]);
        if (ret == 0) {
            return i;
        }

        if (ret < 0) {
            return -1;
        }
    }

    return total;
}

int64_t InputStream::Read(utils::ByteBuffer &buf) noexcept
{
    auto count = Read(buf.Data() + buf.Offset(), buf.Capacity() - buf.Offset());
    if (count > 0) {
        buf.AddOffset(static_cast<uint32_t>(count));
    }
    return count;
}

int64_t OutputStream::Write(const uint8_t *buf, uint64_t count) noexcept
{
    auto total = static_cast<int64_t>(count);
    for (auto i = 0L; i < total; i++) {
        auto ret = Write(buf[i]);
        if (ret <= 0L) {
            return (i == 0L && ret < 0L) ? -1L : i;
        }
    }

    return total;
}

int OutputStream::Sync() noexcept
{
    return 0;
}

std::shared_ptr<InputStream> BaseFileService::GetFile(const std::string &path) noexcept
{
    return GetFile(path, FileRange{});
}

int BaseFileService::GetFile(const std::string &path, OutputStream &outputStream) noexcept
{
    return GetFile(path, FileRange{}, outputStream);
}

int BaseFileService::GetFile(const std::string &path, utils::ByteBuffer &dataBuffer) noexcept
{
    return GetFile(path, dataBuffer, FileRange{});
}