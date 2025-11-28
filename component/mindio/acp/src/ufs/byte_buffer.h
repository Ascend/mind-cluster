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
#ifndef UFS_BYTE_BUFFER_H
#define UFS_BYTE_BUFFER_H

#include <utility>
#include "securec.h"

namespace ock {
namespace ufs {
namespace utils {
class ByteBuffer {
public:
    explicit ByteBuffer() noexcept : data{ nullptr }, capacity(0), offset{ 0 } {}
    explicit ByteBuffer(uint64_t cap) noexcept : data{ new (std::nothrow) uint8_t[cap] }, capacity{ cap }, offset{ 0 }
    {}
    explicit ByteBuffer(uint8_t *p, uint64_t len) noexcept : data{ p }, capacity(len), offset{ 0 }, wrapped{ true } {}

    virtual ~ByteBuffer() noexcept
    {
        if (!wrapped) {
            delete[] data;
            data = nullptr;
        }

        data = nullptr;
        capacity = offset = 0UL;
    }

    explicit ByteBuffer(const ByteBuffer &buf) = delete;

    ByteBuffer(ByteBuffer &&buf) noexcept
    {
        data = buf.data;
        capacity = buf.capacity;
        offset = buf.offset;
        wrapped = buf.wrapped;

        buf.data = nullptr;
        buf.capacity = buf.offset = 0UL;
    }

    ByteBuffer &operator = (const ByteBuffer &buf) = delete;

    ByteBuffer &operator = (ByteBuffer &&buf) noexcept
    {
        if (!wrapped) {
            delete[] data;
            data = nullptr;
        }
        data = buf.data;
        capacity = buf.capacity;
        offset = buf.offset;
        wrapped = buf.wrapped;

        buf.data = nullptr;
        buf.capacity = buf.offset = 0UL;
        return *this;
    }

public:
    bool Valid() const noexcept
    {
        return data != nullptr && capacity != 0;
    }

    int Write(const uint8_t *buf, uint64_t len) noexcept
    {
        if (offset + len > capacity) {
            return -1;
        }

        if (len == 0) {
            return 0;
        }

        auto ret = memcpy_s(data + offset, capacity - offset, buf, len);
        if (ret != EOK) {
            return -1;
        }

        offset += len;
        return 0;
    }

    int WriteAt(const uint8_t *buf, uint64_t len, uint64_t pos) noexcept
    {
        if (pos + len > capacity) {
            return -1;
        }

        if (len == 0UL) {
            return 0;
        }

        auto ret = memcpy_s(data + pos, capacity - pos, buf, len);
        if (ret != EOK) {
            return -1;
        }

        return 0;
    }

    int Read(uint8_t *buf, uint64_t len) noexcept
    {
        if (offset + len > capacity) {
            return -1;
        }

        if (len == 0) {
            return 0;
        }

        auto ret = memcpy_s(buf, capacity - offset, data + offset, len);
        if (ret != EOK) {
            return -1;
        }

        offset += len;
        return 0;
    }

    int ReadAt(uint8_t *buf, uint64_t len, uint64_t pos) noexcept
    {
        if (pos + len > capacity) {
            return -1;
        }

        if (len == 0) {
            return 0;
        }

        auto ret = memcpy_s(buf, capacity - pos, data + pos, len);
        if (ret != EOK) {
            return -1;
        }

        return 0;
    }

public:
    const uint8_t *Data() const noexcept
    {
        return data;
    }
    uint8_t *Data() noexcept
    {
        return data;
    }
    uint64_t Capacity() const noexcept
    {
        return capacity;
    }
    uint64_t Offset() const noexcept
    {
        return offset;
    }
    void Offset(uint64_t off) noexcept
    {
        if (off > capacity) {
            offset = capacity;
        } else {
            offset = off;
        }
    }
    void AddOffset(uint64_t delta) noexcept
    {
        if (offset + delta > capacity) {
            offset = capacity;
        } else {
            offset += delta;
        }
    }
    void Reset() noexcept
    {
        offset = 0;
    }

private:
    uint8_t *data;
    uint64_t capacity;
    uint64_t offset;
    bool wrapped = false;
};
}
}
}

#endif // UFS_BYTE_BUFFER_H
