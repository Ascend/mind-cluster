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
 *
 *
 * This is for testing only
 */
#include <cstdlib>
#include <cstdio>
#include <string>
#include <fcntl.h>
#include "securec.h"
#include "common_includes.h"
#include "memfs_sdk_api.h"

using namespace ock::memfs;

uint32_t g_writeTimes = 1;

int Start()
{
    return MemFsClientInitialize();
}

int MkDir(const std::string &path)
{
    return MemFsMkDir(path.c_str(), 0);
}

int OpenFileAndWrite1GB(const std::string &path)
{
    uint64_t size = 2147483648L * 4;
    char *dataToBeWritten = new (std::nothrow) char[size];
    if (dataToBeWritten == nullptr) {
        LOG_ERROR("alloc size: (" << size << ") memory failed");
        return MFS_ALLOC_FAIL;
    }
    auto ret = memset_s(dataToBeWritten, size, '1', size);
    if (ret != 0) {
        LOG_ERROR("memset failed,ret:" << ret);
        delete[] dataToBeWritten;
        dataToBeWritten = nullptr;
        return ret;
    }

    int fd = MemFsOpenFile(path.c_str(), O_CREAT);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(fd < 0, "Failed to open a file");

    LOG_INFO("created file @" << path << ", fd " << fd);
    for (uint32_t i = 0; i < g_writeTimes; i++) {
        struct timespec ts {};
        clock_gettime(CLOCK_MONOTONIC, &ts);
        MemFsWrite(fd, reinterpret_cast<uintptr_t>(dataToBeWritten), size);
        struct timespec tsEnd {};
        clock_gettime(CLOCK_MONOTONIC, &tsEnd);

        uint64_t start = ts.tv_sec * 1000000000L + ts.tv_nsec;
        uint64_t end = tsEnd.tv_sec * 1000000000L + tsEnd.tv_nsec;
        std::cout << "\n\tWrite " << (size / 1024 / 1024 / 1024) << "GB data to MindIO MemFS took " << // 1024
            ((end - start) / 1000 / 1000) <<                                                           // 1000
            "ms\n" << std::endl;
    }

    size = size / 4; // 4
    for (uint32_t i = 0; i < g_writeTimes; i++) {
        struct timespec ts {};
        clock_gettime(CLOCK_MONOTONIC, &ts);
        MemFsWrite(fd, reinterpret_cast<uintptr_t>(dataToBeWritten), size);
        struct timespec tsEnd {};
        clock_gettime(CLOCK_MONOTONIC, &tsEnd);

        uint64_t start = ts.tv_sec * 1000000000L + ts.tv_nsec;
        uint64_t end = tsEnd.tv_sec * 1000000000L + tsEnd.tv_nsec;
        std::cout << "\n\tWrite " << (size / 1024 / 1024 / 1024) << "GB data to MindIO MemFS took " << // 1024
            ((end - start) / 1000 / 1000) <<                                                           // 1000
            "ms\n" << std::endl;
    }

    LOG_INFO("written file @" << path << ", fd " << fd);

    LOG_ERROR_RETURN_IT_IF_NOT_OK(MemFsFlush(fd), "Failed to flush file " << fd);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(MemFsClose(fd), "Failed to close file " << fd);

    return 0;
}

int main(int argc, char *argv[])
{
    Start();

    LOG_ERROR_RETURN_IT_IF_NOT_OK(MkDir("/aa/"), "-- mkdir failed");

    LOG_ERROR_RETURN_IT_IF_NOT_OK(OpenFileAndWrite1GB("/aa/file1"), "-- open file failed");

    (void)getchar();

    return 0;
}