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

#include <cerrno>
 
#include "securec.h"

#include "file_check_utils.h"
#include "nds_file_driver.h"

using namespace ock::common;
using namespace ock::memfs;

NdsFileDriver *NdsFileDriver::gInstance = nullptr;
std::mutex NdsFileDriver::gLock;

NdsFileError_t (*NdsFileDriver::NdsOpenFunc)();
NdsFileError_t (*NdsFileDriver::NdsCloseFunc)();

static constexpr uint64_t MAX_LIB_FILE_SIZE = 1024 * 1024 * 1024;
static constexpr const char *DPC_NDS_FILE = "/opt/oceanstor/dataturbo/sdk/lib/libdpc_nds.so";
static constexpr const char *NDS_DIR_PATH = "/opt/oceanstor/dataturbo/sdk/lib";
static constexpr uint64_t NDS_SINGLE_MAX_READ_SIZE = 128UL << 20;

void NdsFileDriver::Initialize()
{
    std::lock_guard<std::mutex> guard(ndsMutex);
    if (ndsInited) {
        return;
    }
    int result = this->InitNdsFileDriver();
    ndsInited = true;
    if (result != 0) {
        LOG_WARN("Failed to initial nds file driver, use system call");
        return;
    }
    auto status = NdsOpenFunc();
    if (status.err != NDS_FILE_SUCCESS) {
        LOG_WARN("Nds file driver failed to open, use system call, error:" << status.err);
        return;
    }
    LOG_INFO("initial and open nds file driver success");
    ndsAvailable = true;
}

void NdsFileDriver::UnInitialize()
{
    std::lock_guard<std::mutex> guard(ndsMutex);
    if (!ndsInited || !ndsAvailable) {
        return;
    }
    NdsCloseFunc();
    ndsInited = false;
    ndsAvailable = false;
}

int NdsFileDriver::InitNdsFileDriver() noexcept
{
    std::string errMsg{};
    if (!FileCheckUtils::RegularFilePath(DPC_NDS_FILE, NDS_DIR_PATH, errMsg) || !FileCheckUtils::IsFileValid(
        DPC_NDS_FILE, errMsg, true, FileCheckUtils::FILE_MODE_444, false, false, MAX_LIB_FILE_SIZE)) {
        LOG_WARN("Nds library is invalid:" << errMsg);
        return -1;
    }

    auto ndsDriverDyLib = dlopen(DPC_NDS_FILE, RTLD_LAZY | RTLD_LOCAL);
    if (ndsDriverDyLib == nullptr) {
        LOG_WARN("dlopen occurs error, nds lib not found or glibc version is too low.");
        return -1;
    }
    do {
        NdsOpenFunc = reinterpret_cast<NdsFileError_t (*)()>(dlsym(ndsDriverDyLib, "NdsFileDriverOpen"));
        if (NdsOpenFunc == nullptr) {
            break;
        }
        NdsCloseFunc = reinterpret_cast<NdsFileError_t (*)()>(dlsym(ndsDriverDyLib, "NdsFileDriverClose"));
        if (NdsCloseFunc == nullptr) {
            break;
        }
        ndsRegisterFunc = reinterpret_cast<NdsFileError_t (*)(NdsFileHandle_t * fh, NdsFileDescr_t * descr)>(
            dlsym(ndsDriverDyLib, "NdsFileHandleRegister"));
        if (ndsRegisterFunc == nullptr) {
            break;
        }
        ndsDeregisterFunc =
            reinterpret_cast<void (*)(NdsFileHandle_t fh)>(dlsym(ndsDriverDyLib, "NdsFileHandleDeregister"));
        if (ndsDeregisterFunc == nullptr) {
            break;
        }
        ndsReadFunc = reinterpret_cast<ssize_t (*)(NdsFileHandle_t fh, void *ptr_base, size_t size, off_t file_offset,
            off_t ptr_offset)>(dlsym(ndsDriverDyLib, "NdsFileRead"));
        if (ndsReadFunc == nullptr) {
            break;
        }
    } while (false);
    if (ndsReadFunc == nullptr) {
        dlclose(ndsDriverDyLib);
        NdsOpenFunc = nullptr;
        NdsCloseFunc = nullptr;
        ndsRegisterFunc = nullptr;
        ndsDeregisterFunc = nullptr;
        LOG_WARN("ndsDriverDyLib dlopen symbol invalid.");
        return -1;
    }
    return 0;
}

bool NdsFileDriver::NdsAvailable()
{
    Initialize();
    return ndsAvailable;
}

NdsFileError_t NdsFileDriver::RegisterHandler(int fd)
{
    NdsFileHandle_t fh;
    NdsFileDescr_t descr;
    NdsFileError_t error_code;
    auto err = memset_s((void *)&descr, sizeof(NdsFileDescr_t), 0, sizeof(NdsFileDescr_t));
    if (err != EOK) {
        error_code.err = NDS_FILE_DRIVER_ERROR;
        return error_code;
    }
    descr.fd = fd;
    NdsFileError_t ret = ndsRegisterFunc(&fh, &descr);
    mNdsHandle.insert({ fd, fh });
    return ret;
}

void NdsFileDriver::NdsFileHandleDeregister(int fd)
{
    NdsFileHandle_t fh = mNdsHandle[fd];
    mNdsHandle.erase(fd);
    ndsDeregisterFunc(fh);
}

ssize_t NdsFileDriver::FileRead(int fd, void *ptr_base, size_t size, off_t file_offset, off_t ptr_offset)
{
    NdsFileHandle_t fh = mNdsHandle[fd];
    return ndsReadFunc(fh, ptr_base, size, file_offset, ptr_offset);
}

int NdsFileDriver::NdsRead(int inFd, void *dataPtr, uint64_t fileOffset, uint64_t leftReadSize)
{
    auto pos = static_cast<char *>(dataPtr);
    while (leftReadSize > 0) {
        auto curReadSize = std::min(leftReadSize, NDS_SINGLE_MAX_READ_SIZE);
        ssize_t readSize = FileRead(inFd, pos, curReadSize, (off_t)fileOffset, 0);
        if (readSize < 0) {
            return -1;
        }

        leftReadSize -= static_cast<uint64_t>(readSize);
        fileOffset += static_cast<uint64_t>(readSize);
        pos += static_cast<uint64_t>(readSize);
    }

    return 0;
}