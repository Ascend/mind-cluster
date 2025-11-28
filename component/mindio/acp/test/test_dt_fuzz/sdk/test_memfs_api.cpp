/*
* Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
*/
#include <map>
#include <string>
#include <fcntl.h>

#include "memfs_sdk_api.h"
#include "utils.h"
#include "fs_server.h"
#include "service_configure.h"
#include "Secodefuzz/secodeFuzz.h"
#include "mockcpp/mockcpp.hpp"
#include "under_fs_manager.h"
#include "background_manager.h"
#include "test_memfs_api.h"

using namespace ock::memfs;

TEST_F(MemFsApiDtFuzz, mem_fs_write_fuzz)
{
    DT_Enable_Leak_Check(0, 0);
    DT_Set_Running_Time_Second(FUZZ_TEST_SECONDS);
    MemFsMkDir("/mnt/dpc01/fuzz/fuzz_write_test", 0);
    DT_FUZZ_START(0, FUZZ_TEST_TIMES, "test_write_file", 0)
    {
        int index = 0;
        std::string path = "path1";
        char *fuzzPath =
            DT_SetGetString(&g_Element[index++], path.length() + 1, 3200, const_cast<char *>(path.c_str()));
        std::string finalPath = std::string("/mnt/dpc01/fuzz/fuzz_write_test/") + fuzzPath;

        int fd = MemFsOpenFile(finalPath.c_str(), O_CREAT | O_TRUNC | O_WRONLY);

        uint32_t writeSize = *(uint32_t *)DT_SetGetS32(&g_Element[index++], 1024);
        if (writeSize < MAX_BUFFER_TO_WRITE) {
            char *buffer = new (std::nothrow) char[writeSize];
            if (buffer != nullptr) {
                MemFsWrite(fd, reinterpret_cast<uintptr_t>(buffer), writeSize);
                delete[] buffer;
                buffer = nullptr;
            }
        }

        MemFsClose(fd);
    }
    DT_FUZZ_END()
}

TEST_F(MemFsApiDtFuzz, mem_fs_read_fuzz)
{
    DT_Enable_Leak_Check(0, 0);
    DT_Set_Running_Time_Second(FUZZ_TEST_SECONDS);
    MemFsMkDir("/mnt/dpc01/fuzz/fuzz_read_test", 0);
    DT_FUZZ_START(0, FUZZ_TEST_TIMES, "test_read_file", 0)
    {
        int index = 0;
        std::string path = "path1";
        char *fuzzPath =
            DT_SetGetString(&g_Element[index++], path.length() + 1, 3200, const_cast<char *>(path.c_str()));
        std::string finalPath = std::string("/mnt/dpc01/fuzz/fuzz_read_test/") + fuzzPath;

        int32_t flags = *(int *)DT_SetGetS32(&g_Element[index++], O_CREAT | O_TRUNC | O_WRONLY);
        int fd = MemFsOpenFile(finalPath.c_str(), flags);

        uint32_t readSize = *(uint32_t *)DT_SetGetS32(&g_Element[index++], 1024);
        if (readSize < MAX_BUFFER_TO_WRITE) {
            char *buffer = new (std::nothrow) char[readSize];
            if (buffer != nullptr) {
                uint64_t position = *(uint64_t *)DT_SetGetU64(&g_Element[index++], 0);
                auto ret = MemFsRead(fd, reinterpret_cast<uintptr_t>(buffer), position, readSize);
                delete[] buffer;
                buffer = nullptr;
            }
        }

        MemFsClose(fd);
    }
    DT_FUZZ_END()
}

TEST_F(MemFsApiDtFuzz, mem_fs_preload_test)
{
    DT_Enable_Leak_Check(0, 0);
    DT_Set_Running_Time_Second(FUZZ_TEST_SECONDS);
    MemFsMkDir("/mnt/dpc01/fuzz/fuzz_preload_file_test", 0);
    DT_FUZZ_START(0, FUZZ_TEST_TIMES, "test_preload_file_test", 0)
    {
        int index = 0;
        std::string path = "path1";
        char *fuzzPath =
                DT_SetGetString(&g_Element[index++], path.length() + 1, 3200, const_cast<char *>(path.c_str()));
        std::string finalPath = std::string("/mnt/dpc01/fuzz/fuzz_preload_file_test/") + fuzzPath;

        MemfsPreloadFile(finalPath.c_str());
    }
    DT_FUZZ_END()
}

TEST_F(MemFsApiDtFuzz, mem_fs_rename_and_link_test)
{
    DT_Enable_Leak_Check(0, 0);
    DT_Set_Running_Time_Second(FUZZ_TEST_SECONDS);
    MemFsMkDir("/mnt/dpc01/fuzz/fuzz_rename_file_test", 0);
    DT_FUZZ_START(0, FUZZ_TEST_TIMES, "test_rename_file_test", 0)
    {
        int index = 0;
        std::string path = "path1";
        char *fuzzSrcPath =
                DT_SetGetString(&g_Element[index++], path.length() + 1, 3200, const_cast<char *>(path.c_str()));
        char *fuzzTarPath =
                DT_SetGetString(&g_Element[index++], path.length() + 1, 3200, const_cast<char *>(path.c_str()));
        std::string srcPath = std::string("/mnt/dpc01/fuzz/fuzz_rename_file_test/") + fuzzSrcPath;
        std::string tarPath = std::string("/mnt/dpc01/fuzz/fuzz_rename_file_test/") + fuzzTarPath;

        int fd = MemFsOpenFile(srcPath.c_str(), O_CREAT | O_TRUNC | O_WRONLY);
        auto writeSize = 10U;
        char *buffer = new (std::nothrow) char[writeSize];
        if (buffer != nullptr) {
            MemFsWrite(fd, reinterpret_cast<uintptr_t>(buffer), writeSize);
            delete[] buffer;
            buffer = nullptr;
        }
        MemFsClose(fd);

        MemFsRenameFile(srcPath.c_str(), tarPath.c_str(), 0);
        MemFsLinkFile(tarPath.c_str(), srcPath.c_str());
    }
    DT_FUZZ_END()
}

void MemFsApiDtFuzz::SetUpTestSuite()
{
    const char *cwd = std::getenv("PWD");
    std::string curDir{ cwd };
    std::string rootDir = std::move(curDir.substr(0, curDir.find_last_of('/')));
    std::string serverPath = rootDir + "/output";
    std::string ockiodPath = rootDir + "/output/bin/ockiod";

    std::map<std::string, std::string> serverInfo {
        { "memfs.data_block_pool_capacity_in_gb", "16" },
        { "memfs.data_block_size_in_mb", "64" },
        { "memfs.write.parallel.enabled", "true" },
        { "memfs.write.parallel.thread_num", "16" },
        { "memfs.write.parallel.slice_in_mb", "16" },
        { "background.backup.thread_num", "32" },
        { "background.backup.failed_auto_evict_file", "true"},
        { "background.backup.failed_max_cnt_for_unserviceable", "11"},
        { "server.worker.path", serverPath },
        { "server.ockiod.path", ockiodPath }
    };
    MemFsClientInitialize(serverInfo);
}

void MemFsApiDtFuzz::TearDownTestSuite()
{
    MemFsClientUnInitialize();
}
