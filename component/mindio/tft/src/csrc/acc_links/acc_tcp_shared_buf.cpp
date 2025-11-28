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
#include "acc_common_util.h"
#include "acc_tcp_shared_buf.h"

namespace ock {
namespace acc {
AccDataBuffer::AccDataBuffer(uint32_t memSize)
    : dataSize_(memSize), memSize_{ memSize },
      data_{ new (std::nothrow) uint8_t[memSize] }
{
}

AccDataBuffer::AccDataBuffer(const void *data, uint32_t size)
    : AccDataBuffer{ size }
{
    if (data_ != nullptr) {
        const uint8_t* src_ptr = static_cast<const uint8_t*>(data);
        std::copy_n(src_ptr, size, data_);
        dataSize_ = size;
    }
}

AccDataBuffer::~AccDataBuffer()
{
    delete[] data_;
    data_ = nullptr;
    memSize_ = 0;
    dataSize_ = 0;
}

bool AccDataBuffer::AllocIfNeed(uint32_t newSize) noexcept
{
    if (newSize > MAX_RECV_BODY_LEN) {
        return false;
    }

    if (data_ == nullptr) {
        memSize_ = std::max(memSize_, newSize);
        data_ = static_cast<uint8_t *>(malloc(memSize_));
        return data_ != nullptr;
    }

    if (newSize > memSize_) {
        /* free old and malloc new one */
        free(data_);
        data_ = nullptr;

        memSize_ = std::max(memSize_, newSize);
        data_ = static_cast<uint8_t *>(malloc(memSize_));
        return data_ != nullptr;
    }

    return true;
}

AccDataBufferPtr AccDataBuffer::Create(const void *data, uint32_t size)
{
    auto buffer = AccMakeRef<AccDataBuffer>(data, size);
    if (buffer.Get() == nullptr || buffer->data_ == nullptr) {
        return nullptr;
    }

    return buffer;
}

AccDataBufferPtr AccDataBuffer::Create(uint32_t memSize)
{
    auto buffer = AccMakeRef<AccDataBuffer>(memSize);
    if (buffer.Get() == nullptr || buffer->data_ == nullptr) {
        return nullptr;
    }

    return buffer;
}
} // namespace acc
} // namespace ock