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
#ifndef OCK_MEMFS_CORE_FS_SERVER_H
#define OCK_MEMFS_CORE_FS_SERVER_H
#include <cstdint>

#include "ipc_server.h"
#include "memfs_api.h"

namespace ock {
namespace memfs {
struct ChannelInfo {
    enum OpType {
        OPEN_FOR_WRITE,
        OPEN_FOR_READ,
    };

    ChannelInfo()
    {
        pthread_spin_init(&readLock, 0);
        pthread_spin_init(&writeLock, 0);
    }

    ~ChannelInfo()
    {
        pthread_spin_destroy(&readLock);
        pthread_spin_destroy(&writeLock);
    }

    pthread_spinlock_t readLock;
    std::unordered_map<int, std::string> readFd; // store <fd, path>

    pthread_spinlock_t writeLock;
    std::unordered_map<int, std::string> writeFd; // store <fd, path>

    uint64_t channelId = 0;
};

class ShellFSServer {
public:
    MResult Start();
    void Stop();

    static std::shared_ptr<ShellFSServer> Instance()
    {
        static auto instance = std::make_shared<ShellFSServer>();
        return instance;
    }

    static bool InputPathValid(const std::string &path, std::string &validPath) noexcept;

private:
    MResult StartIpcServer();
    bool PrepareBucketPrefix();
    bool FormatFilePath(const std::string &inPath, std::string &path);
    static int OpenFileWithCreateParent(const std::string &path, int flags, mode_t mode);
    std::pair<int, int> OpenFileRead(OpenFileReq *req, std::vector<uint64_t> &blocks, struct stat &staBuf,
        std::string &realPath);
    std::pair<int, int> ProcessLinkFile(const std::string &source, const std::string &target);
    std::pair<int, int> ProcessRenameFile(const std::string &source, const std::string &target, uint32_t flags);
    void StopInner();
    MResult RegisterHandlers();
    /* ipc handler functions */
    MResult HandleNewConnection(const ChannelPtr &ch);
    void HandleConnectionBroken(const ChannelPtr &ch);
    MResult HandleGetSharedFileInfo(ServiceContext &ctx);
    MResult HandleMakeDir(ServiceContext &ctx);
    MResult Reply(ServiceContext &ctx, OpenFileReq *req, OpenFileResp &resp, std::string &path);
    MResult HandleOpenFile(ServiceContext &ctx);
    MResult HandleAllocMoreBlock(ServiceContext &ctx);
    MResult ReplyForAllocBlocks(ServiceContext &ctx, int result, uint64_t blkSz, const std::vector<uint64_t> &blocks);
    MResult CheckTruncateFileReq(const ock::hcom::UBSHcomChannelPtr &channel, const TruncateFileReq *req);
    MResult HandleTruncateFile(ServiceContext &ctx);
    MResult CheckCloseFileReq(const ock::hcom::UBSHcomChannelPtr &channel, const FlushSyncCloseFileReq *req);
    MResult HandleFlushSyncCloseFile(ServiceContext &ctx);
    MResult HandleAccessFile(ServiceContext &ctx);
    MResult HandleOpenFile4Read(ServiceContext &ctx);
    MResult HandleLinkFile(ServiceContext &ctx);
    MResult HandleRenameFile(ServiceContext &ctx);
    MResult HandleUnlinkFile(ServiceContext &ctx);
    MResult HandleCheckBackgroundTask(ServiceContext &ctx);
    MResult HandleMakeCacheAsync(ServiceContext &ctx);
    MResult HandleGetServerStatus(ServiceContext &ctx);
    static void ConnectionBrokenCleanFd(const std::shared_ptr<ChannelInfo> &info);
    void InsertOpenFdInfo(ChannelInfo::OpType type, uint64_t chId, int fd, std::string path);
    void EraseOpenFdInfo(uint64_t chId, int fd);
    MResult FetchReadPathByFd(uint64_t chId, int fd, std::string &path);
    MResult FetchWritePathByFd(uint64_t chId, int fd, std::string &path);
    MResult HandleCloseFileInner(uint64_t chId, FlushSyncCloseFileReq &req);
    MResult CheckFdOpenedByChannel(const ChannelPtr &ch, int fd, ChannelInfo::OpType opType);
    void RegisterMonitorServerState() noexcept;
    bool ProcessServerExitEvent() noexcept;
    inline bool ServerInactive() noexcept
    {
        return mIpcServer->GetConnCnt() == 0L && MemFsApi::BackgroundTaskEmpty();
    }

private:
    std::string backupPrefix;
    std::string dockerPrefix;
    pthread_rwlock_t mChannelRWLock{};
    std::unordered_map<uint64_t, std::shared_ptr<ChannelInfo>> mChannelInfo;
    IpcServerPtr mIpcServer = nullptr;
    std::mutex mMutex;
    bool mStarted = false;
    static constexpr uint64_t MAX_TRUNCATE_FILE_SIZE = 1UL * 1024UL * 1024UL * 1024UL * 1024UL; // 1TB
};
}
}


#endif // OCK_MEMFS_CORE_FS_SERVER_H
