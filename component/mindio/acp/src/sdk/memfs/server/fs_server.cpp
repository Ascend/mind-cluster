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
#include <fcntl.h>
#include <utility>
#include <csignal>
#include <pwd.h>

#include "memfs_file_util.h"

#include "auditlog_adapt.h"
#include "file_check_utils.h"
#include "service_configure.h"
#include "mem_fs_state.h"
#include "memfs_logger.h"
#include "bmm_pool.h"
#include "fs_server.h"

using namespace ock::hcom;
using namespace ock::common;

namespace ock {
namespace memfs {

static constexpr int64_t LOG_LEVEL_DEBUG = 0;
static constexpr int64_t LOG_LEVEL_INFO = 1;
static constexpr int64_t LOG_LEVEL_WARN = 2;
static constexpr int64_t LOG_LEVEL_ERROR = 3;
static constexpr auto MB_TO_BYTES_SHIFT = 20;
static constexpr uint8_t THREAD_MAX_WAIT_TIME = 2;
static constexpr int32_t CHECK_SERVER_ACTIVE_PERIOD = 20;
static constexpr int32_t MAX_BACKGROUND_TASK_FINISH_TIME = 30;
static constexpr int32_t CHECK_BACKGROUND_TASK_PERIOD = 1;

static void FsServerLog(int level, const char *msg)
{
    ASSERT_RET_VOID(msg != nullptr);
    switch (level) {
        case LOG_LEVEL_DEBUG:
            HLOG_DEBUG(msg);
            break;
        case LOG_LEVEL_INFO:
            HLOG_INFO(msg);
            break;
        case LOG_LEVEL_WARN:
            HLOG_WARN(msg);
            break;
        case LOG_LEVEL_ERROR:
            HLOG_ERROR(msg);
            break;
        default:
            HLOG_WARN("invalid level " << level << ", " << msg);
            break;
    }
}

MResult ShellFSServer::Start()
{
    std::lock_guard<std::mutex> guard(mMutex);
    if (mStarted) {
        return MFS_OK;
    }

    auto hcomLogger = ock::hcom::NetLogger::Instance();
    auto daggerLogger = OutLogger::Instance();
    if (hcomLogger == nullptr || daggerLogger == nullptr) {
        MFS_LOG_ERROR("get hcom and dagger logger failed.");
        return MFS_ALLOC_FAIL;
    }

    hcomLogger->SetExternalLogFunction(FsServerLog);
    daggerLogger->SetExternalLogFunction(FsServerLog);

    MFS_LOG_INFO("register end, Starting shell fs server");
    if (!PrepareBucketPrefix()) {
        return MFS_INVALID_CONFIG;
    }

    /* init memfs core, need to start fs core firstly */
    MFS_LOG_INFO("Start to create fs core");
    if (MemFsApi::Initialize() != 0) {
        MFS_LOG_ERROR("Failed to create memfs core");
        HLOG_AUDIT("system", "start", "fs server", "failed");
        return MFS_NEW_OBJ_FAIL;
    }
    MFS_LOG_INFO("Fs core created");
    MemfsState::Instance().SetState(MemfsStateCode::STARTING, MemfsStartProgress::EIGHTY_PERCENT);

    if (StartIpcServer() != MFS_OK) {
        HLOG_AUDIT("system", "start", "fs server", "failed");
        StopInner();
        MemFsApi::Destroy();
        return MFS_ERROR;
    }

    mStarted = true;

    RegisterMonitorServerState();
    MFS_LOG_INFO("Shell fs server started");
    HLOG_AUDIT("system", "start", "fs server", "success");
    return MFS_OK;
}

MResult ShellFSServer::StartIpcServer()
{
    /* start ipc server */
    mIpcServer = std::make_shared<IpcServer>();
    if (UNLIKELY(mIpcServer.get() == nullptr)) {
        MFS_LOG_ERROR("Failed to new IpcServer, probably out memory");
        return MFS_NEW_OBJ_FAIL;
    }

    auto result = RegisterHandlers();
    if (UNLIKELY(result != MFS_OK)) {
        HLOG_AUDIT("system", "start", "fs server", "failed");
        MFS_LOG_ERROR("Failed to register handlers");
        return result;
    }

    result = mIpcServer->Start();
    if (UNLIKELY(result != MFS_OK)) {
        MFS_LOG_ERROR("Failed to start IpcServer");
        HLOG_AUDIT("system", "start", "fs server", "failed");
        return result;
    }

    return MFS_OK;
}

void ShellFSServer::Stop()
{
    StopInner();
    MemFsApi::Destroy();
    for (const auto &item : DEFAULT_SERVER_DIR) {
        if (item.first == "/logs") {
            continue;
        }
        FileUtil::RemoveDirRecursive(config::ServiceConfigure::GetInstance().GetWorkPath() + item.first);
    }
    FileUtil::Remove(config::ServiceConfigure::GetInstance().GetWorkPath() + "/.ockiod.pid");

    MFS_LOG_INFO("Shell fs server exited.");
    MemfsState::Instance().SetState(MemfsStateCode::EXITED);
}

void ShellFSServer::RegisterMonitorServerState() noexcept
{
    std::thread([this]() {
        while (true) {
            std::this_thread::sleep_for(std::chrono::seconds(CHECK_SERVER_ACTIVE_PERIOD));
            if (ServerInactive() && ProcessServerExitEvent()) {
                return;
            }
        }
    }).detach();
}

bool ShellFSServer::ProcessServerExitEvent() noexcept
{
    MemfsState::Instance().SetState(MemfsStateCode::PRE_EXITING);
    if (!ServerInactive()) {
        MFS_LOG_INFO("server is active, continue running.");
        MemfsState::Instance().SetState(MemfsStateCode::RUNNING);
        return false;
    }
    MemfsState::Instance().SetState(MemfsStateCode::EXITING);
    raise(SIGINT);
    MFS_LOG_INFO("Shell fs server is exiting...");
    return true;
}

bool ShellFSServer::InputPathValid(const std::string &path, std::string &validPath) noexcept
{
    std::vector<std::string> pathItems;
    StrUtil::Split(path, "/", pathItems);
    validPath.reserve(path.size());
    validPath.clear();
    for (const auto &item : pathItems) {
        if (item.empty()) {
            continue;
        }

        if (item == "." || item == "..") {
            return false;
        }
        validPath.append("/").append(item);
    }

    if (validPath.empty()) {
        validPath.append("/");
    }

    static constexpr auto maxCharValue = 127;
    return std::all_of(path.begin(), path.end(),
                       [](char ch) { return ch < 0 || ch > maxCharValue || isprint(ch) != 0; });
}

bool ShellFSServer::PrepareBucketPrefix()
{
    auto &serviceConfig = config::ServiceConfigure::GetInstance();

    auto &backupConfig = serviceConfig.GetBackgroundConfig().backupServiceConfig;
    if (backupConfig.backups.empty()) {
        MFS_LOG_ERROR("no backup configured");
        return false;
    }

    auto &underFss = serviceConfig.GetUnderFileSystemConfig().instances;
    auto &backup = *backupConfig.backups.begin();
    auto pos = underFss.find(backup.destName);
    if (pos == underFss.end()) {
        MFS_LOG_ERROR("backup under fs name : " << backup.destName.c_str() << " not found.");
        return false;
    }

    auto mountPathPos = pos->second.options.find("mount_path");
    if (mountPathPos == pos->second.options.end()) {
        MFS_LOG_ERROR("backup under fs name :" << backup.destName.c_str() << " no mount path.");
        return false;
    }

    backupPrefix = mountPathPos->second;
    if (backupPrefix.empty() || backupPrefix[0] != '/') {
        MFS_LOG_ERROR("backup under fs name : " << backup.destName.c_str() << " mount path : " << backupPrefix.c_str()
                                                << " invalid");
        return false;
    }

    while (backupPrefix.size() > 1 && backupPrefix[backupPrefix.length() - 1] == '/') {
        backupPrefix.erase(backupPrefix.length() - 1);
    }

    if (backupPrefix[backupPrefix.length() - 1] != '/') {
        backupPrefix.append("/");
    }

    auto dockerPathPos = pos->second.options.find("docker_map_path");
    if (dockerPathPos == pos->second.options.end()) {
        MFS_LOG_ERROR("backup under fs name :" << backup.destName.c_str() << " no docker path.");
        return false;
    }

    dockerPrefix = dockerPathPos->second;
    if (!dockerPrefix.empty() && dockerPrefix[0] != '/') {
        MFS_LOG_ERROR("backup under fs name : " << backup.destName.c_str() << " docker path : " << dockerPrefix.c_str()
                                                << " invalid");
        return false;
    }

    while (dockerPrefix.size() > 1 && dockerPrefix[dockerPrefix.length() - 1] == '/') {
        dockerPrefix.erase(dockerPrefix.length() - 1);
    }

    return true;
}

bool ShellFSServer::FormatFilePath(const std::string &inPath, std::string &path)
{
    if (!dockerPrefix.empty()) {
        if (!StrUtil::StartWith(inPath, dockerPrefix)) {
            return false;
        }

        auto tmpPath = inPath.substr(dockerPrefix.length());
        return InputPathValid(tmpPath, path);
    }

    if (!StrUtil::StartWith(inPath, backupPrefix)) {
        return false;
    }

    auto tmpPath = inPath.substr(backupPrefix.length());
    return InputPathValid(tmpPath, path);
}

int ShellFSServer::OpenFileWithCreateParent(const std::string &path, int flags, mode_t mode)
{
    auto pos = path.find_last_of('/');
    if (pos == std::string::npos) {
        MFS_LOG_ERROR("cannot find parent path");
        return -EINVAL;
    }

    auto parent = path.substr(0, pos);
    if (!parent.empty() && parent != "/") {
        auto ret = MemFsApi::CreateDirectoryWithParents(parent, mode | S_IXUSR);
        if (ret != 0) {
            return -errno;
        }
    }

    auto fd = MemFsApi::OpenFile(path, flags, mode);
    if (fd < 0) {
        return -errno;
    }

    return fd;
}

std::pair<int, int> ShellFSServer::OpenFileRead(OpenFileReq *req, std::vector<uint64_t> &blocks, struct stat &staBuf,
    std::string &realPath)
{
    if (!FormatFilePath(req->FileName(), realPath)) {
        MFS_LOG_ERROR("input file name invalid.");
        return std::make_pair(-1, -EINVAL);
    }

    if (PreloadProgressView::PathExist(realPath)) {
        MFS_LOG_INFO("file path(" << realPath << ") is preloading...");
        PreloadProgressView::Wait(THREAD_MAX_WAIT_TIME, realPath);
        if (PreloadProgressView::PathExist(realPath)) {
            return std::make_pair(-1, -EAGAIN);
        }
    }

    auto fd = MemFsApi::OpenFile(realPath, req->flags, 0);
    if (fd < 0) {
        MFS_LOG_ERROR("Failed to open file @" << realPath << ", flags " << req->flags << ", errno " << errno);
        return std::make_pair(-1, -errno);
    }

    auto result = MemFsApi::GetFileBlocks(fd, blocks);
    if (result != 0) {
        result = -errno;
        MFS_LOG_ERROR("Failed to get blocks @" << realPath << ", flags " << req->flags << ", errno " << errno);
        MemFsApi::CloseFile(fd);
        return std::make_pair(-1, result);
    }

    result = MemFsApi::GetFileMeta(fd, staBuf);
    if (result != MFS_OK) {
        result = -errno;
        MFS_LOG_ERROR("Failed to get blocks @" << realPath << ", flags " << req->flags << ", errno " << errno);
        MemFsApi::CloseFile(fd);
        blocks.clear();
        return std::make_pair(-1, result);
    }

    return std::make_pair(fd, 0);
}

std::pair<int, int> ShellFSServer::ProcessLinkFile(const std::string &source, const std::string &target)
{
    std::string realSourcePath;
    if (!FormatFilePath(source, realSourcePath)) {
        MFS_LOG_ERROR("input source file name invalid.");
        return std::make_pair(-1, EINVAL);
    }
    std::string realTargetPath;
    if (!FormatFilePath(target, realTargetPath)) {
        MFS_LOG_ERROR("input target file name invalid.");
        return std::make_pair(-1, EINVAL);
    }

    auto pos = realTargetPath.find_last_of('/');
    if (pos == std::string::npos) {
        MFS_LOG_ERROR("cannot find parent path for target");
        return std::make_pair(-1, EINVAL);
    }

    auto parent = realTargetPath.substr(0, pos);
    if (!parent.empty() && parent != "/") {
        auto ret = MemFsApi::CreateDirectoryWithParents(parent, S_IRWXU);
        if (ret != 0) {
            return std::make_pair(-1, errno);
        }
    }

    auto result = MemFsApi::Link(realSourcePath, realTargetPath);
    if (result == 0) {
        return std::make_pair(0, 0);
    }

    if (errno == EEXIST) {
        MemFsApi::Unlink(realTargetPath);
        result = MemFsApi::Link(realSourcePath, realTargetPath);
        if (result == 0) {
            return std::make_pair(0, 0);
        }
    }

    return std::make_pair(-1, errno);
}

std::pair<int, int> ShellFSServer::ProcessRenameFile(const std::string &src, const std::string &dst, uint32_t flags)
{
    std::string realSourcePath;
    if (!FormatFilePath(src, realSourcePath)) {
        MFS_LOG_ERROR("input source file name invalid.");
        return std::make_pair(-1, EINVAL);
    }
    std::string realTargetPath;
    if (!FormatFilePath(dst, realTargetPath)) {
        MFS_LOG_ERROR("input target file name invalid.");
        return std::make_pair(-1, EINVAL);
    }

    auto pos = realTargetPath.find_last_of('/');
    if (pos == std::string::npos) {
        MFS_LOG_ERROR("cannot find parent path for target");
        return std::make_pair(-1, EINVAL);
    }

    auto result = MemFsApi::Rename(realSourcePath, realTargetPath, flags);
    if (result == 0) {
        return std::make_pair(0, 0);
    }

    return std::make_pair(-1, errno);
}

void ShellFSServer::StopInner()
{
    if (mIpcServer.get() != nullptr) {
        mIpcServer->Stop();
        mIpcServer = nullptr;
    }
}

MResult ShellFSServer::RegisterHandlers()
{
    mIpcServer->RegisterNewChannelHandler(std::bind(&ShellFSServer::HandleNewConnection, this, std::placeholders::_1));
    mIpcServer->RegisterChannelBrokenHandler(
        std::bind(&ShellFSServer::HandleConnectionBroken, this, std::placeholders::_1));

    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_OPEN_FILE,
        std::bind(&ShellFSServer::HandleOpenFile, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_ALLOCATE_MORE_BLOCK,
        std::bind(&ShellFSServer::HandleAllocMoreBlock, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_TRUNCATE,
        std::bind(&ShellFSServer::HandleTruncateFile, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_FLUSH_SYNC_CLOSE_FILE,
        std::bind(&ShellFSServer::HandleFlushSyncCloseFile, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_MAKE_DIR,
        std::bind(&ShellFSServer::HandleMakeDir, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_GET_SHARED_FILE_INFO,
        std::bind(&ShellFSServer::HandleGetSharedFileInfo, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_ACCESS,
        std::bind(&ShellFSServer::HandleAccessFile, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_OPEN_FILE_FOR_READ,
        std::bind(&ShellFSServer::HandleOpenFile4Read, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_LINK_FILE,
        std::bind(&ShellFSServer::HandleLinkFile, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_RENAME_FILE,
        std::bind(&ShellFSServer::HandleRenameFile, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_PRELOAD_FILE,
        std::bind(&ShellFSServer::HandleMakeCacheAsync, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_GET_SERVER_STATUS,
        std::bind(&ShellFSServer::HandleGetServerStatus, this, std::placeholders::_1));
    mIpcServer->RegisterNewRequestHandler(FileOpCode::IPC_OP_CHECK_BACKGROUND_TASK,
        std::bind(&ShellFSServer::HandleCheckBackgroundTask, this, std::placeholders::_1));
    return MFS_OK;
}

MResult ShellFSServer::HandleNewConnection(const ChannelPtr &ch)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);

    int fds[1]{};
    fds[0] = MemFsApi::GetShareMemoryFd();

    pthread_rwlock_wrlock(&mChannelRWLock);
    mChannelInfo[ch->GetId()] = std::make_shared<ChannelInfo>();
    mChannelInfo[ch->GetId()]->channelId = ch->GetId();
    pthread_rwlock_unlock(&mChannelRWLock);

    MFS_LOG_INFO("Sending shared file descriptor " << fds[0] << " to client");

    int result = ch->SendFds(fds, 1);
    if (UNLIKELY(result != MFS_OK)) {
        MFS_LOG_ERROR("Failed to send fd to client");
        return result;
    }

    return MFS_OK;
}

void ShellFSServer::ConnectionBrokenCleanFd(const std::shared_ptr<ChannelInfo> &info)
{
    pthread_spin_lock(&info->readLock);
    auto readSize = info->readFd.size();
    if (readSize > 0) {
        MFS_LOG_INFO("Handle broken event , close read file count " << readSize);
    }
    for (const auto &it : info->readFd) {
        MFS_LOG_INFO("Close read file, fd " << it.first);
        MemFsApi::CloseFile(it.first);
    }
    pthread_spin_unlock(&info->readLock);

    pthread_spin_lock(&info->writeLock);
    auto writeSize = info->writeFd.size();
    if (writeSize > 0) {
        MFS_LOG_INFO("Handle broken event , close write file count " << writeSize);
    }
    for (const auto &it : info->writeFd) {
        MFS_LOG_INFO("Discard write file, fd " << it.first);
        MemFsApi::DiscardFile(it.second, it.first);
    }
    pthread_spin_unlock(&info->writeLock);
}

void ShellFSServer::HandleConnectionBroken(const ChannelPtr &ch)
{
    MFS_LOG_INFO("Handle connection broken, channel id " << ch->GetId());
    pthread_rwlock_wrlock(&mChannelRWLock);
    auto iter = mChannelInfo.find(ch->GetId());
    if (iter != mChannelInfo.end()) {
        if (iter->second->channelId != ch->GetId()) {
            MFS_LOG_ERROR("Find channel id " << ch->GetId() << " different from record " << iter->second->channelId);
            pthread_rwlock_unlock(&mChannelRWLock);
            return;
        }

        auto info = iter->second;
        std::thread cleanThread(&ShellFSServer::ConnectionBrokenCleanFd, info);
        cleanThread.detach();
        mChannelInfo.erase(iter);
    }
    pthread_rwlock_unlock(&mChannelRWLock);
}

void ShellFSServer::InsertOpenFdInfo(ChannelInfo::OpType type, uint64_t chId, int fd, std::string path)
{
    std::shared_ptr<ChannelInfo> info = nullptr;
    pthread_rwlock_rdlock(&mChannelRWLock);
    auto iter = mChannelInfo.find(chId);
    if (iter != mChannelInfo.end()) {
        info = iter->second;
    }
    pthread_rwlock_unlock(&mChannelRWLock);

    if (info == nullptr) {
        MFS_LOG_ERROR("Failed to insert open fd info by invalid channel id " << chId);
        return;
    }

    if (type == ChannelInfo::OPEN_FOR_WRITE) {
        pthread_spin_lock(&info->writeLock);
        info->writeFd[fd] = std::move(path);
        pthread_spin_unlock(&info->writeLock);
    } else if (type == ChannelInfo::OPEN_FOR_READ) {
        pthread_spin_lock(&info->readLock);
        info->readFd[fd] = std::move(path);
        pthread_spin_unlock(&info->readLock);
    }
}

void ShellFSServer::EraseOpenFdInfo(uint64_t chId, int fd)
{
    std::shared_ptr<ChannelInfo> info = nullptr;
    pthread_rwlock_rdlock(&mChannelRWLock);
    auto iter = mChannelInfo.find(chId);
    if (iter != mChannelInfo.end()) {
        info = iter->second;
    }
    pthread_rwlock_unlock(&mChannelRWLock);

    if (info == nullptr) {
        MFS_LOG_ERROR("Failed to erase open fd info by invalid channel id " << chId);
        return;
    }

    // can not tell type, erase all
    pthread_spin_lock(&info->writeLock);
    auto writeInfo = info->writeFd.find(fd);
    if (writeInfo != info->writeFd.end()) {
        info->writeFd.erase(writeInfo);
    }
    pthread_spin_unlock(&info->writeLock);

    pthread_spin_lock(&info->readLock);
    auto readInfo = info->readFd.find(fd);
    if (readInfo != info->readFd.end()) {
        info->readFd.erase(readInfo);
    }
    pthread_spin_unlock(&info->readLock);
}

MResult ShellFSServer::FetchWritePathByFd(uint64_t chId, int fd, std::string &path)
{
    std::shared_ptr<ChannelInfo> info = nullptr;
    pthread_rwlock_rdlock(&mChannelRWLock);
    auto iter = mChannelInfo.find(chId);
    if (iter != mChannelInfo.end()) {
        info = iter->second;
    }
    pthread_rwlock_unlock(&mChannelRWLock);

    if (info == nullptr) {
        LOG_WARN("Failed to get path by invalid channel id " << chId);
        return MFS_INVALID_PARAM;
    }

    bool success = true;
    pthread_spin_lock(&info->writeLock);
    auto writeInfo = info->writeFd.find(fd);
    if (writeInfo != info->writeFd.end()) {
        path = writeInfo->second;
        info->writeFd.erase(writeInfo);
    } else {
        success = false;
    }
    pthread_spin_unlock(&info->writeLock);

    if (!success) {
        LOG_WARN("Failed to get path channel id " << chId << ", fd " << fd);
        return MFS_INVALID_PARAM;
    }

    return MFS_OK;
}

MResult ShellFSServer::FetchReadPathByFd(uint64_t chId, int fd, std::string &path)
{
    std::shared_ptr<ChannelInfo> info = nullptr;
    pthread_rwlock_rdlock(&mChannelRWLock);
    auto iter = mChannelInfo.find(chId);
    if (iter != mChannelInfo.end()) {
        info = iter->second;
    }
    pthread_rwlock_unlock(&mChannelRWLock);

    if (info == nullptr) {
        LOG_WARN("Failed to get path by invalid channel id " << chId);
        return MFS_INVALID_PARAM;
    }

    bool success = true;
    pthread_spin_lock(&info->readLock);
    auto readInfo = info->readFd.find(fd);
    if (readInfo != info->readFd.end()) {
        path = readInfo->second;
        info->readFd.erase(readInfo);
    } else {
        success = false;
    }
    pthread_spin_unlock(&info->readLock);

    if (!success) {
        LOG_WARN("Failed to get path channel id " << chId << ", fd " << fd);
        return MFS_INVALID_PARAM;
    }

    return MFS_OK;
}

MResult ShellFSServer::HandleGetSharedFileInfo(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);

    auto req = static_cast<ShareFileInfoReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(ShareFileInfoReq) && req != nullptr, MFS_INVALID_PARAM);

    ShareFileInfoResp resp;
    uint64_t blockSize = 0;
    uint64_t blockCnt = 0;
    MemFsApi::GetShareFileCfg(blockSize, blockCnt);
    resp.singleFileSize = blockCnt * blockSize;
    resp.fileCount = 1;
    resp.maxBlkCountInSingleFile = blockCnt;
    resp.flags = req->flags;

    auto &writeConf = config::ServiceConfigure::GetInstance().GetMemFsConfig().writeParallel;
    resp.writeParallelEnabled = writeConf.enabled;
    resp.writeParallelThreadNum = writeConf.threadNum;
    resp.writeParallelSlice = (static_cast<uint64_t>(writeConf.sliceInMB) << MB_TO_BYTES_SHIFT);

    auto &ufsConfig = config::ServiceConfigure::GetInstance().GetUnderFileSystemConfig();

    std::map<std::string, std::string>::const_iterator it;
    std::map<std::string, std::string>::const_iterator dit;
    auto pos = ufsConfig.instances.find(ufsConfig.defaultName);
    if (pos == ufsConfig.instances.end()) {
        MFS_LOG_ERROR("cannot find default UFS with name : " << ufsConfig.defaultName);
        resp.result = -1;
    } else if ((it = pos->second.options.find("mount_path")) == pos->second.options.end()) {
        MFS_LOG_ERROR("cannot find default UFS mount path, ufs name = " << ufsConfig.defaultName);
        resp.result = -1;
    } else if ((dit = pos->second.options.find("docker_map_path")) != pos->second.options.end() &&
               !dit->second.empty()) {
        resp.UfsPath(dit->second);
        resp.result = 0;
    } else {
        resp.UfsPath(it->second);
        resp.result = 0;
    }
    MFS_LOG_INFO("Get shared file info " << resp.ToString());

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&resp), sizeof(resp), 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response to client shared file info result " << result);
        return result;
    }

    return MFS_OK;
}

MResult ShellFSServer::HandleMakeDir(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);

    auto req = static_cast<MakeDirReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(MakeDirReq) && req != nullptr, MFS_INVALID_PARAM);

    MFS_LOG_INFO("MakeDirReq " << req->ToString());
    MakeDirResp resp;
    int result;
    std::string realPath;

    if (!FormatFilePath(req->Path(), realPath)) {
        resp.result = -EINVAL;
    } else if ((result = MemFsApi::CreateDirectory(realPath, req->flags, req->recursive)) != 0) {
        MFS_LOG_ERROR("Failed to mkdir @ " << realPath);
        resp.result = -errno;
    } else {
        MFS_LOG_INFO("Dir @" << realPath << "created");
    }

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&resp), sizeof(resp), 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response to client for mkdir @" << req->Path().c_str() << ", result " << result);
        return result;
    }

    return MFS_OK;
}

void RecordOpenFile(const ServiceContext &ctx, const OpenFileReq *req, const std::string &realPath, const int fd)
{
    MFS_LOG_DEBUG("Opened file " << realPath << ", fd " << fd);
    const auto &channel = ctx.Channel();
    UdsInfo info;
    channel->GetRemoteUdsIdInfo(info);
    if (req->flags == (O_CREAT | O_TRUNC | O_WRONLY)) {
        HLOG_AUDIT(std::string("user: ") + std::to_string(info.uid), "(re)create file", realPath, "success");
    } else {
        HLOG_AUDIT(std::string("user: ") + std::to_string(info.uid), "read file", realPath, "success");
    }
}

MResult ShellFSServer::Reply(ServiceContext &ctx, OpenFileReq *req, OpenFileResp &resp, std::string &path)
{
    if (resp.result == 0) {
        InsertOpenFdInfo(ChannelInfo::OPEN_FOR_WRITE, ctx.Channel()->GetId(), resp.fd, path);
    }

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&resp), sizeof(resp), 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to response for openFile @" << req->FileName().c_str() << ", result " << result);
        EraseOpenFdInfo(ctx.Channel()->GetId(), resp.fd);
        MemFsApi::DiscardFile(path, resp.fd);
        return result;
    }
    return MFS_OK;
}

MResult ShellFSServer::HandleOpenFile(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);

    auto req = static_cast<OpenFileReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(OpenFileReq) && req != nullptr, MFS_INVALID_PARAM);

    std::string realPath;
    OpenFileResp resp;

    if (!MemFsApi::Serviceable()) {
        resp.result = MFS_UNSERVICEABLE;
        MFS_LOG_ERROR("Failed to open file by unserviceable now");
        goto REPLY;
    }

    if (!FormatFilePath(req->FileName(), realPath)) {
        MFS_LOG_ERROR("input file name invalid.");
        resp.fd = -1;
        resp.result = -EINVAL;
    } else {
        /* call fs to mkdir */
        auto fd = OpenFileWithCreateParent(realPath, req->flags, S_IRUSR | S_IWUSR);
        if (fd < 0) {
            MFS_LOG_ERROR("Failed to open @" << realPath << ", 0o" << std::oct << req->flags << ":" << errno);
            resp.result = -errno;
            resp.fd = -1;
        } else {
            resp.fd = fd;

            /* allocate one block and append to fd */
            uint64_t blkId = 0;
            uint64_t blockSize;
            resp.result = MemFsApi::AllocDataBlock(fd, blkId, blockSize);
            if (resp.result != 0) {
                MFS_LOG_ERROR("Failed to allocate more block as allocate failed: " << fd);
                MemFsApi::DiscardFile(realPath, fd);
            } else {
                resp.blockSize = blockSize;
                resp.dataBlock.size = resp.blockSize;
                resp.dataBlock.offset = MemFsApi::GetBlockOffset(blkId);

                RecordOpenFile(ctx, req, realPath, fd);
            }
        }
    }
REPLY:
    auto ret = Reply(ctx, req, resp, realPath);
    if (ret != 0) {
        LOG_ERROR("Reply failed ret " << ret);
        return ret;
    }

    return MFS_OK;
}

MResult ShellFSServer::HandleAllocMoreBlock(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);

    auto req = static_cast<AllocateMoreBlockReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(AllocateMoreBlockReq) && req != nullptr, MFS_INVALID_PARAM);

    int result = 0;
    uint64_t blockSize = 0;
    std::vector<uint64_t> blockIds;
    if (!MemFsApi::Serviceable()) {
        result = MFS_UNSERVICEABLE;
        MFS_LOG_ERROR("Failed to alloc block by unserviceable now");
    } else if (CheckFdOpenedByChannel(ctx.Channel(), req->fd, ChannelInfo::OpType::OPEN_FOR_WRITE) != MCode::MFS_OK) {
        MFS_LOG_ERROR("Failed to allocate more block as fd is invalid " << req->fd);
        result = -EBADF;
    } else {
        result = MemFsApi::AllocDataBlocks(req->fd, req->size, blockIds, blockSize);
    }

    return ReplyForAllocBlocks(ctx, result, blockSize, blockIds);
}

MResult ShellFSServer::ReplyForAllocBlocks(ServiceContext &ctx, int result, uint64_t blkSz,
                                           const std::vector<uint64_t> &blocks)
{
    auto req = static_cast<const AllocateMoreBlockReq *>(ctx.MessageData());
    uint32_t respLen = sizeof(AllocateMoreBlockResp) + sizeof(uint64_t) * blocks.size();
    auto buffer = new (std::nothrow) uint8_t[respLen];
    if (buffer == nullptr) {
        MFS_LOG_ERROR("allocate response size " << respLen << " failed.");
        return MCode::MFS_ALLOC_FAIL;
    }

    std::unique_ptr<uint8_t[]> bufRelease(buffer);
    auto *resp = reinterpret_cast<OpenFileWithBlockResp *>(buffer);
    resp->result = result;
    resp->fd = req->fd;
    resp->fileSize = req->size;
    resp->blockSize = blkSz;
    resp->blockCount = static_cast<uint32_t>(blocks.size());
    for (auto i = 0U; i < resp->blockCount; i++) {
        resp->dataBlock[i] = MemFsApi::GetBlockOffset(blocks[i]);
    }

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(resp), respLen, 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response to client for allocate block for " << req->fd << ", result is "
                                                                                  << result);
        return result;
    }

    return MCode::MFS_OK;
}

MResult ShellFSServer::CheckTruncateFileReq(const UBSHcomChannelPtr &channel, const ock::memfs::TruncateFileReq *req)
{
    if (CheckFdOpenedByChannel(channel, req->fd, ChannelInfo::OpType::OPEN_FOR_WRITE) != MCode::MFS_OK) {
        MFS_LOG_ERROR("Failed to truncate fd is invalid " << req->fd);
        return MCode::MFS_INVALID_PARAM;
    }

    if (req->size > MAX_TRUNCATE_FILE_SIZE) {
        MFS_LOG_ERROR("Failed to truncate fd:" << req->fd << ", file size too large: " << req->size);
        return MCode::MFS_INVALID_PARAM;
    }

    return MCode::MFS_OK;
}

MResult ShellFSServer::HandleTruncateFile(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);

    auto req = static_cast<TruncateFileReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(TruncateFileReq) && req != nullptr, MFS_INVALID_PARAM);

    TruncateFileResp resp;
    resp.fd = req->fd;

    if (CheckTruncateFileReq(ctx.Channel(), req) != MCode::MFS_OK) {
        MFS_LOG_ERROR("Failed to truncate fd is invalid " << req->fd);
        resp.result = -EINVAL;
    } else {
        /* truncate fd */
        resp.result = MemFsApi::TruncateFile(req->fd, req->size);
        resp.fd = req->fd;
    }

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&resp), sizeof(resp), 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response to client for allocate block for " << req->fd << ", result is "
                                                                                  << result);
        return result;
    }

    MFS_LOG_DEBUG("Truncate fd " << req->fd << ", size " << req->size);
    return MFS_OK;
}

MResult ShellFSServer::HandleCloseFileInner(uint64_t chId, FlushSyncCloseFileReq &req)
{
    /* op file */
    switch (req.op) {
        case FOF_FLUSH: {
            return MFS_OK;
        }
        case FOF_CLOSE: {
            MResult result = MFS_OK;
            if ((req.flags & O_ACCMODE) != O_RDONLY) {
                result = MemFsApi::TruncateFile(req.fd, req.fileSize);
            }
            if (result == MFS_OK) {
                result = MemFsApi::CloseFile(req.fd);
            }
            if (result != MFS_OK) {
                MFS_LOG_ERROR("Failed to close file as result " << result);
                return result;
            }
            if ((req.flags & O_ACCMODE) == O_RDONLY) {
                std::string path;
                result = FetchReadPathByFd(chId, req.fd, path);
                if (result == MFS_OK) {
                    result = MemFsApi::Unlink(path);
                }
                if (result != MFS_OK) {
                    MFS_LOG_WARN("Failed to remove file(" << path << ") as result " << strerror(errno));
                    return result;
                }
                MFS_LOG_INFO("load from memfs file finished, close and remove memfs file success.");
            }
            return result;
        }
        case FOF_CLOSE_WITH_UNLINK: {
            std::string path;
            MResult result = FetchWritePathByFd(chId, req.fd, path);
            if (result == MFS_OK) {
                result = MemFsApi::DiscardFile(path, req.fd);
            }
            if (result != MFS_OK) {
                MFS_LOG_WARN("Discard file as result " << result);
            }
            return MFS_OK;
        }
        default: {
            MFS_LOG_ERROR("Invalid operator " << req.op << ", fd " << req.fd);
            return MFS_INVALID_PARAM;
        }
    }
}

MResult ShellFSServer::CheckFdOpenedByChannel(const ChannelPtr &ch, int fd, ChannelInfo::OpType opType)
{
    if (ch == nullptr || fd < 0) {
        return MCode::MFS_INVALID_PARAM;
    }

    pthread_rwlock_rdlock(&mChannelRWLock);
    auto iter = mChannelInfo.find(ch->GetId());
    if (iter == mChannelInfo.end()) {
        pthread_rwlock_unlock(&mChannelRWLock);
        MFS_LOG_ERROR("cannot found channel id " << ch->GetId() << " in records");
        return MCode::MFS_INVALID_PARAM;
    }

    auto info = iter->second;
    pthread_rwlock_unlock(&mChannelRWLock);

    if (info == nullptr) {
        MFS_LOG_ERROR("record info for " << ch->GetId() << " invalid");
        return MCode::MFS_INVALID_PARAM;
    }

    bool exist = false;
    if (opType == ChannelInfo::OpType::OPEN_FOR_READ) {
        pthread_spin_lock(&info->readLock);
        auto fdrPos = info->readFd.find(fd);
        exist = (fdrPos != info->readFd.end());
        pthread_spin_unlock(&info->readLock);
    } else if (opType == ChannelInfo::OpType::OPEN_FOR_WRITE) {
        pthread_spin_lock(&info->writeLock);
        auto fdwPos = info->writeFd.find(fd);
        exist = (fdwPos != info->writeFd.end());
        pthread_spin_unlock(&info->writeLock);
    } else {
        return MCode::MFS_INVALID_PARAM;
    }

    if (!exist) {
        MFS_LOG_ERROR("fd " << fd << ", opType " << opType << " not open for channel " << ch->GetId());
        return MCode::MFS_INVALID_PARAM;
    }

    return MCode::MFS_OK;
}

MResult ShellFSServer::CheckCloseFileReq(const UBSHcomChannelPtr &channel, const ock::memfs::FlushSyncCloseFileReq *req)
{
    auto opType = ((req->flags & O_ACCMODE) == O_RDONLY) ? ChannelInfo::OPEN_FOR_READ : ChannelInfo::OPEN_FOR_WRITE;
    if (CheckFdOpenedByChannel(channel, req->fd, opType) != MCode::MFS_OK) {
        MFS_LOG_ERROR("Failed to close file as fd (" << req->fd << ") is invalid.");
        return MCode::MFS_INVALID_PARAM;
    }

    if (req->op < FOF_BEGIN || req->op >= FOF_TOTAL) {
        MFS_LOG_ERROR("Failed to close file as fd (" << req->fd << ") op(" << req->op << ") invalid");
        return MCode::MFS_INVALID_PARAM;
    }

    if (req->fileSize > MAX_TRUNCATE_FILE_SIZE) {
        MFS_LOG_ERROR("Failed to close truncate fd:" << req->fd << ", file size too large: " << req->fileSize);
        return MCode::MFS_INVALID_PARAM;
    }

    return MCode::MFS_OK;
}

MResult ShellFSServer::HandleFlushSyncCloseFile(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);
    auto req = static_cast<FlushSyncCloseFileReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(FlushSyncCloseFileReq) && req != nullptr, MFS_INVALID_PARAM);

    FlushSyncCloseFileResp resp;
    resp.result = 0;

    if (CheckCloseFileReq(ctx.Channel(), req) != MCode::MFS_OK) {
        resp.result = -EINVAL;
    } else {
        resp.fd = req->fd;
        resp.flags = req->flags;
        resp.result = HandleCloseFileInner(ctx.Channel()->GetId(), *req);
    }

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&resp), sizeof(resp), 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response to client for close fd " << req->fd << ", result is " << result);
    }

    if (resp.result == 0) {
        EraseOpenFdInfo(ctx.Channel()->GetId(), req->fd);
    }
    return result;
}

MResult ShellFSServer::HandleAccessFile(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(AccessFileReq), MFS_INVALID_PARAM);
    AccessFileResp resp;

    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&resp), sizeof(resp), 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response to client for close, result is " << result);
        return result;
    }
    return MFS_OK;
}

MResult ShellFSServer::HandleOpenFile4Read(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);
    auto req = static_cast<OpenFileReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(OpenFileReq) && req != nullptr, MFS_INVALID_PARAM);

    MFS_LOG_DEBUG("OpenFileReq " << req->ToString().c_str());
    std::vector<uint64_t> blocks;
    struct stat stat {};
    std::string realPath;
    auto pairResult = OpenFileRead(req, blocks, stat, realPath);
    uint32_t respLen = sizeof(OpenFileWithBlockResp) + sizeof(uint64_t) * blocks.size();
    auto buffer = new (std::nothrow) uint8_t[respLen];
    if (buffer == nullptr) {
        MFS_LOG_ERROR("failed to allocate response buffer");
        MemFsApi::CloseFile(pairResult.first);
        return MFS_ALLOC_FAIL;
    }

    std::unique_ptr<uint8_t[]> bufRelease(buffer);

    /* fill response */
    auto *resp = reinterpret_cast<OpenFileWithBlockResp *>(buffer);
    resp->fd = pairResult.first;
    resp->result = pairResult.second;
    resp->fileSize = stat.st_size;
    resp->blockSize = stat.st_blksize;
    resp->blockCount = blocks.size();
    for (uint32_t i = 0; i < resp->blockCount; ++i) {
        resp->dataBlock[i] = MemFsApi::GetBlockOffset(blocks[i]);
    }

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(resp), respLen, 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MemFsApi::CloseFile(resp->fd);
        MFS_LOG_ERROR("Failed to send response to client for openFile @" << realPath << ", result " << result);
        return result;
    }

    InsertOpenFdInfo(ChannelInfo::OPEN_FOR_READ, ctx.Channel()->GetId(), resp->fd, realPath);
    return MFS_OK;
}

MResult ShellFSServer::HandleLinkFile(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);
    auto req = static_cast<LinkFileReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(LinkFileReq) && req != nullptr, MFS_INVALID_PARAM);

    MFS_LOG_INFO("LinkFileReq " << req->ToString());

    auto resultPair = ProcessLinkFile(req->SourcePath(), req->TargetPath());
    LinkFileRes response{resultPair.first, resultPair.second};

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&response), sizeof(LinkFileRes), 0},
                                       mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response for LinkFileReq @" << req->ToString() << ", result " << result);
        return result;
    }

    return MFS_OK;
}

MResult ShellFSServer::HandleRenameFile(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);
    auto req = static_cast<RenameFileReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(RenameFileReq) && req != nullptr, MFS_INVALID_PARAM);

    MFS_LOG_INFO("RenameFileReq " << req->ToString());

    auto resultPair = ProcessRenameFile(req->SourcePath(), req->TargetPath(), req->flags);
    RenameFileRes response{resultPair.first, resultPair.second};

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&response), sizeof(RenameFileRes), 0},
                                       mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response for RenameFileReq @" << req->ToString() << ", result " << result);
        return result;
    }

    return MFS_OK;
}

MResult ShellFSServer::HandleMakeCacheAsync(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);
    auto req = static_cast<PreloadFileReq *>(ctx.MessageData());

    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(PreloadFileReq) && req != nullptr, MFS_INVALID_PARAM);
    MFS_LOG_INFO("PreloadFileReq " << req->ToString());

    std::string realPath;
    PreloadFileResp resp;

    if (!MemFsApi::Serviceable()) {
        resp.result = MFS_UNSERVICEABLE;
        MFS_LOG_ERROR("Failed to open file by unserviceable now");
    } else if (!FormatFilePath(req->FileName(), realPath)) {
        MFS_LOG_ERROR("input file name invalid.");
        resp.result = -EINVAL;
    } else {
        resp.result = MemFsApi::PreloadFile(realPath);
    }

    if (resp.result != MFS_OK) {
        resp.result = -errno;
        MFS_LOG_ERROR("Failed to preload file @" << req->FileName() << ". errno : " << errno);
    }

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&resp), sizeof(resp), 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response for PreloadFileReq @" << req->ToString() << ", result " << result);
        return result;
    }

    return MFS_OK;
}

MResult ShellFSServer::HandleGetServerStatus(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);
    auto req = static_cast<ServerStatusReq *>(ctx.MessageData());
    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(ServerStatusReq) && req != nullptr, MFS_INVALID_PARAM);

    ServerStatusResp resp{};
    auto state = MemfsState::Instance().GetState();
    resp.status = state.first;
    resp.progress = state.second;
    resp.result = 0;

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&resp), sizeof(resp), 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response for ServerStatusReq.");
        return result;
    }

    return MFS_OK;
}

MResult ShellFSServer::HandleCheckBackgroundTask(ServiceContext &ctx)
{
    ASSERT_RETURN(mStarted, MFS_NOT_INITIALIZED);
    auto req = static_cast<CheckBackgroundTaskReq *>(ctx.MessageData());

    ASSERT_RETURN(ctx.MessageDataLen() == sizeof(CheckBackgroundTaskReq) && req != nullptr, MFS_INVALID_PARAM);
    MFS_LOG_INFO("Try to check background task");

    CheckBackgroundTaskResp resp;
    resp.result = MFS_OK;

    if (!MemFsApi::Serviceable()) {
        resp.result = MFS_UNSERVICEABLE;
        MFS_LOG_ERROR("Failed to check background task by unserviceable now");
    } else {
        auto start = std::chrono::high_resolution_clock::now();
        while (!MemFsApi::BackgroundTaskEmpty()) {
            auto current = std::chrono::high_resolution_clock::now();
            if (std::chrono::duration_cast<std::chrono::seconds>(current - start).count() >=
                MAX_BACKGROUND_TASK_FINISH_TIME) {
                resp.result = -EAGAIN;
                break;
            }
            std::this_thread::sleep_for(std::chrono::seconds(CHECK_BACKGROUND_TASK_PERIOD));
        }
        if (resp.result == MFS_OK) {
            MFS_LOG_INFO("All background tasks are finished.");
        }
    }

    /* response to client */
    ock::hcom::UBSHcomReplyContext replyCtx(ctx.RspCtx(), 0);
    auto *mEmptyCallback = ock::hcom::UBSHcomNewCallback([](ServiceContext &context) {}, std::placeholders::_1);
    if (mEmptyCallback == nullptr) {
        MFS_LOG_ERROR("allocate empty call back failed.");
        return MFS_ALLOC_FAIL;
    }
    auto result = ctx.Channel()->Reply(replyCtx, {static_cast<void *>(&resp), sizeof(resp), 0}, mEmptyCallback);
    if (result != MFS_OK) {
        MFS_LOG_ERROR("Failed to send response for CheckBackgroundTaskReq, result " << result);
        return result;
    }

    return MFS_OK;
}

}
}