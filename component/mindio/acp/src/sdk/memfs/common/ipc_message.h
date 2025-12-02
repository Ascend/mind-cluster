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
#ifndef OCK_MEMFS_CORE_MESSAGE_H
#define OCK_MEMFS_CORE_MESSAGE_H

#include <cstdint>
#include <functional>
#include <iomanip>
#include "mem_fs_state.h"
#include "file_check_utils.h"
#include "common_includes.h"
#include "hcom_service.h"
#include "hcom_service_context.h"
#include "securec.h"

using namespace ock::common;

namespace ock {
namespace memfs {
using ChannelPtr = ock::hcom::UBSHcomChannelPtr;
using ServiceContext = ock::hcom::UBSHcomServiceContext;
using UdsInfo = ock::hcom::UBSHcomNetUdsIdInfo;
using NewRequestHandler = std::function<int32_t(ServiceContext &)>;
using NewChannelHandler = std::function<int32_t(const ChannelPtr &)>;
using ChannelBrokenHandler = std::function<void(const ChannelPtr &)>;

constexpr uint32_t MAX_MESSAGE_SIZE = 1048576;
constexpr uint32_t PRINTABLE_WIDTH = 2;

enum FileOpCode {
    IPC_OP_OPEN_FILE = 0,
    IPC_OP_ALLOCATE_MORE_BLOCK = 1,
    IPC_OP_TRUNCATE = 2,
    IPC_OP_FLUSH_SYNC_CLOSE_FILE = 3,
    IPC_OP_MAKE_DIR = 4,
    IPC_OP_GET_SHARED_FILE_INFO = 5,
    IPC_OP_ACCESS = 6,
    IPC_OP_OPEN_FILE_FOR_READ = 7,
    IPC_OP_LINK_FILE = 8,
    IPC_OP_RENAME_FILE = 9,
    IPC_OP_PRELOAD_FILE = 10,
    IPC_OP_GET_SERVER_STATUS = 11,
    IPC_OP_CHECK_BACKGROUND_TASK = 12,
};

inline std::string PrintableString(const std::string &str) noexcept
{
    std::ostringstream oss;
    for (auto &ch : str) {
        if (isprint(ch)) {
            oss << ch;
        } else {
            oss << std::hex << "\\0x" << std::setw(PRINTABLE_WIDTH) << std::setfill('0') <<
                static_cast<uint32_t>(static_cast<uint8_t>(ch)) << std::oct;
        }
    }
    return oss.str();
}

/*
 * Open file request to server via ipc/rpc
 */
struct OpenFileReq {
    char fileName[FS_PATH_MAX]{}; /* file name */
    int flags = 0;                /* flags */

    DECLARE_CHAR_ARRAY_SET_FUNC(FileName, fileName);
    DECLARE_CHAR_ARRAY_GET_FUNC(FileName, fileName);

    std::string ToString() const
    {
        std::ostringstream oss;
        std::string reqFileName = FileCheckUtils::RemovePrefixPath(PrintableString(FileName()));
        oss << "fn " << reqFileName << ", flags=0o" << std::oct << flags << std::dec;
        return oss.str();
    }
};

/*
 * link file request to server via ipc/rpc
 */
struct LinkFileReq {
    uint32_t flags{ 0 };
    char sourcePath[FS_PATH_MAX]{}; /* file name */
    char targetPath[FS_PATH_MAX]{}; /* file name */

    DECLARE_CHAR_ARRAY_SET_FUNC(SourcePath, sourcePath);
    DECLARE_CHAR_ARRAY_SET_FUNC(TargetPath, targetPath);
    DECLARE_CHAR_ARRAY_GET_FUNC(SourcePath, sourcePath);
    DECLARE_CHAR_ARRAY_GET_FUNC(TargetPath, targetPath);

    std::string ToString() const
    {
        std::ostringstream oss;
        std::string reqSourcePath = FileCheckUtils::RemovePrefixPath(PrintableString(SourcePath()));
        std::string reqTargetPath = FileCheckUtils::RemovePrefixPath(PrintableString(TargetPath()));
        oss << "source: " << reqSourcePath << ", target: " << reqTargetPath << ", flags=" << flags;
        return oss.str();
    }
};

/**
 * @brief link file response from server via ipc/rpc
 */
struct LinkFileRes {
    int result;
    int errorCode;

    LinkFileRes() : LinkFileRes{ -1, EINVAL } {}
    LinkFileRes(int res, int err) : result{ res }, errorCode{ err } {}
};

using RenameFileReq = LinkFileReq;
using RenameFileRes = LinkFileRes;

/*
 * Shared data block info allocated from server,
 * which could write data directly
 */
struct SharedDataBlock {
    uint64_t offset = 0; /* offset to base address of shared file */
    uint32_t size = 0;   /* size of the block */
};

/*
 * Open file response from server
 */
struct OpenFileResp {
    int32_t result = 0;          /* result of creating file */
    int32_t fd = -1;             /* fd assigned by server */
    uint32_t blockSize = 0;      /* block size */
    SharedDataBlock dataBlock{}; /* first allocated data block */

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "result " << result << ", fd " << fd << ", blkSize " << blockSize << ", dBlk {" << dataBlock.size <<
            "Bytes," << dataBlock.offset << "}";
        return oss.str();
    }
};

/*
 * Open file response from server
 */
struct OpenFileWithBlockResp {
    int32_t result = 0;      /* result of creating file */
    int32_t fd = -1;         /* fd assigned by server */
    uint32_t blockSize = 0;  /* block size */
    uint32_t blockCount = 0; /* block count */
    uint64_t fileSize = 0;   /* file size */
    uint64_t dataBlock[];    /* first allocated data block */

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "result " << result << ", fd " << fd << ", blkSize " << blockSize << ", blkCount " << blockCount;
        return oss.str();
    }
};

/*
 * Allocate more block request when data block is fully to write
 */
struct AllocateMoreBlockReq {
    int32_t fd = 0;    /* fd assigned by server when open a file */
    int32_t flags = 0; /* flags of allocating */
    uint64_t size = 0; /* size requested */

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "fd " << fd << ", size " << size << ", flags " << flags;
        return oss.str();
    }
};

/*
 * Response of allocating more data block
 */
using AllocateMoreBlockResp = OpenFileWithBlockResp;

/*
 * File truncate request
 */
struct TruncateFileReq {
    uint64_t size = 0;              /* total position */
    uint64_t offsetInLastBlock = 0; /* offset in last block */
    int32_t fd = -1;                /* fd returned when open file */
    int32_t flags = 0;              /*  flags for future extension */

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "size " << size << ", offsetInLastBlock " << offsetInLastBlock << ", fd " << fd << ", flags " << flags;
        return oss.str();
    }
};

/*
 * File truncate response
 */
struct TruncateFileResp {
    int32_t result = 0; /* truncate result */
    int32_t fd = -1;    /* fd returned when open file */
    int32_t flags = 0;  /*  flags for future extension */
};

enum FileOpFile {
    FOF_BEGIN = 0,
    FOF_FLUSH = FOF_BEGIN,
    FOF_SYNC,
    FOF_CLOSE,
    FOF_CLOSE_WITH_UNLINK,
    FOF_TOTAL
};

/*
 * Request of flush or sync file
 */
struct FlushSyncCloseFileReq {
    int32_t fd = -1;           /* fd returned when open file */
    FileOpFile op = FOF_BEGIN; /* operation */
    int32_t flags = 0;         /* flags */
    uint64_t fileSize = 0;     /* file size */

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "fd " << fd << ", op " << op << ", flags " << flags << ", file size " << fileSize;
        return oss.str();
    }
};

struct FlushSyncCloseFileResp {
    int32_t result = 0;        /* truncate result */
    int32_t fd = -1;           /* fd returned when open file */
    FileOpFile op = FOF_BEGIN; /* operation */
    int32_t flags = 0;         /*  flags for future extension */
};

struct MakeDirReq {
    char path[FS_PATH_MAX]{}; /* file name */
    int32_t flags = 0;        /* flags */
    bool recursive = false;   /* recursive make directory */

    DECLARE_CHAR_ARRAY_SET_FUNC(Path, path);
    DECLARE_CHAR_ARRAY_GET_FUNC(Path, path);

    std::string ToString() const
    {
        std::ostringstream oss;
        std::string reqPath = FileCheckUtils::RemovePrefixPath(PrintableString(Path()));
        oss << "path " << reqPath << ", flags " << flags << ", recursive " << recursive;
        return oss.str();
    }
};

struct MakeDirResp {
    int32_t result = 0; /* result */
    int32_t flags = 0;  /* flags */

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "mkdir result " << result << ", flags " << flags;
        return oss.str();
    }
};


struct ShareFileInfoReq {
    int32_t flags = 0;
};

struct ShareFileInfoResp {
    int32_t result = 0;
    int32_t flags = 0;
    uint32_t fileCount = 0;
    uint32_t maxBlkCountInSingleFile = 0;
    uint64_t singleFileSize = 0;
    bool writeParallelEnabled{ false };
    uint32_t writeParallelThreadNum{ 1 };
    uint64_t writeParallelSlice{ 0UL };
    char ufsPath[FS_PATH_MAX]{};

    DECLARE_CHAR_ARRAY_SET_FUNC(UfsPath, ufsPath);
    DECLARE_CHAR_ARRAY_GET_FUNC(UfsPath, ufsPath);

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "result " << result << ", fileCount " << fileCount << ", flags " << flags << ", singleFileSize " <<
            singleFileSize << ", maxBlkCountInSingleFile " << maxBlkCountInSingleFile << "writeParallel(enabled=" <<
            writeParallelEnabled << ", thread=" << writeParallelThreadNum << ", slice=" << writeParallelSlice <<
            ") ufs(" << FileCheckUtils::RemovePrefixPath(UfsPath()) << ")";
        return oss.str();
    }
};

struct AccessFileReq {
    char path[FS_PATH_MAX]{}; /* file name */
    // R_OK	4	Test for read permission.
    // W_OK	2	Test for write permission.
    // X_OK	1	Test for execute permission.
    // F_OK	0	Test for existence.
    int32_t mode = 0;

    DECLARE_CHAR_ARRAY_SET_FUNC(Path, path);
    DECLARE_CHAR_ARRAY_GET_FUNC(Path, path);
};

struct AccessFileResp {
    int32_t result = 0; /* access result */
};

struct PreloadFileReq {
    char fileName[FS_PATH_MAX]{}; /* file name */

    DECLARE_CHAR_ARRAY_SET_FUNC(FileName, fileName);
    DECLARE_CHAR_ARRAY_GET_FUNC(FileName, fileName);

    std::string ToString() const
    {
        std::ostringstream oss;
        std::string reqPath = FileCheckUtils::RemovePrefixPath(PrintableString(FileName()));
        oss << "fn " << PrintableString(reqPath);
        return oss.str();
    }
};

struct PreloadFileResp {
    int32_t result = 0; /* access result */
};

struct ServerStatusReq {
    int32_t flags = 0;
};

struct ServerStatusResp {
    MemfsStartProgress progress; /* server start progress */
    MemfsStateCode status;
    int32_t result;
};

struct CheckBackgroundTaskReq {
};

struct CheckBackgroundTaskResp {
    int32_t result = 0; /* result */
};
}
}
#endif // OCK_MEMFS_CORE_MESSAGE_H
