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

#ifndef OCK_MEMFS_CORE_NDS_SDK_API_H
#define OCK_MEMFS_CORE_NDS_SDK_API_H

#include <dlfcn.h>
#include <mutex>
#include <deque>
#include <unordered_map>

#include "common_includes.h"

#include "nds_api.h"

namespace ock {
namespace memfs {
class NdsFileDriver {
public:
    static NdsFileDriver *Instance()
    {
        if (gInstance != nullptr) {
            return gInstance;
        }
        std::lock_guard<std::mutex> guard(gLock);
        if (gInstance != nullptr) {
            return gInstance;
        }
        gInstance = new (std::nothrow) NdsFileDriver();
        if (gInstance == nullptr) {
            LOG_ERROR("Failed to new NdsFileDriver object, probably out of memory");
            return nullptr;
        }
        return gInstance;
    }

public:
    ~NdsFileDriver() = default;

    void Initialize();
    void UnInitialize();

    bool NdsAvailable();
    NdsFileError_t RegisterHandler(int fd);
    void NdsFileHandleDeregister(int fd);
    int NdsRead(int inFd, void *dataPtr, uint64_t fileOffset, uint64_t leftReadSize);
    ssize_t FileRead(int fd, void *ptr_base, size_t size, off_t file_offset, off_t ptr_offset);

private:
    /* hidden constructor, use Instance() directly */
    NdsFileDriver() = default;
    int InitNdsFileDriver() noexcept;

private:
    std::mutex ndsMutex;
    bool ndsInited = false;
    bool ndsAvailable = false;

    std::unordered_map<int, NdsFileHandle_t> mNdsHandle;

    NdsFileError_t(*ndsRegisterFunc)(NdsFileHandle_t *fh, NdsFileDescr_t *descr) {};
    void(*ndsDeregisterFunc)(NdsFileHandle_t fh) {};
    ssize_t(*ndsReadFunc)(NdsFileHandle_t fh, void *ptr_base, size_t size, off_t file_offset, off_t ptr_offset) {};

private:
    static std::mutex gLock;
    static NdsFileDriver *gInstance;

    static NdsFileError_t (*NdsOpenFunc)();
    static NdsFileError_t (*NdsCloseFunc)();
};
} // namespace memfs
} // namespace ock

#endif // OCK_MEMFS_CORE_NDS_SDK_API_H
