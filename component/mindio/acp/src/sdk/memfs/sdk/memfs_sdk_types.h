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

#ifndef OCKIO_MEMFS_SDK_TYPES_H
#define OCKIO_MEMFS_SDK_TYPES_H

#include <cstdint>

namespace ock {
namespace memfs {
struct Buffer {
    void *buffer;
    uint64_t size;

    Buffer() : buffer{ nullptr }, size{ 0 } {}
    Buffer(void *buf, uint64_t sz) : buffer{ buf }, size{ sz } {}
};
struct ReadBuffer {
    void *buffer;
    uint64_t start;
    uint64_t size;

    ReadBuffer() : buffer{ nullptr }, start{ 0 }, size{ 0 } {}
    ReadBuffer(void *buf, uint64_t st, uint64_t sz) : buffer{ buf }, start{ st }, size{ sz } {}
};
}
}

#endif // OCKIO_MEMFS_SDK_TYPES_H
