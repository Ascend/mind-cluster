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

#include "auditlog_adapt.h"
#include "hlog.h"
#include "service_configure.h"
#include "memfs_file_util.h"
#include "user_group_cache.h"
#include "mem_fs_constants.h"
#include "ipc_server.h"


using namespace ock::common;
using namespace ock::common::config;
using namespace ock::memfs;
using namespace ock::hcom;

static const char *SOCKET_PATH_SUFFIX = "/uds/mindio_memfs_123.s";

IpcServer::IpcServer() : serverRunUid{ getuid() }, serverRunGid{ getgid() }, mAutoCloseTimer{ "AutoClose" } {}

MResult IpcServer::Start()
{
    std::unique_lock<std::mutex> guard(mMutex);
    if (mStarted) {
        MFS_LOG_WARN("IpcServer has been already started");
        return MFS_OK;
    }
    auto result = CreateService();
    ASSERT_RETURN(result == MFS_OK, result);
    mStarted = true;
    ock::common::HLOG_AUDIT("system", "start", "net server", "success");
    guard.unlock();

    auto &config = ServiceConfigure::GetInstance().GetIpcMessageConfig();
    return MFS_OK;
}

void IpcServer::Stop()
{
    using namespace hcom;
    std::unique_lock<std::mutex> guard(mMutex);
    if (!mStarted || mService == nullptr) {
        MFS_LOG_WARN("IpcClient was not started");
        return;
    }
    mService->Destroy("ipc_server");
    HLOG_AUDIT("system", "stop", "net server", "success");
    if (mService != nullptr) {
        mService = nullptr;
    }
    mStarted = false;
    guard.unlock();

    auto &config = ServiceConfigure::GetInstance().GetIpcMessageConfig();
}

int32_t IpcServer::CreateSocketPath(std::string &sockPath)
{
    sockPath = ServiceConfigure::GetInstance().GetWorkPath();

    std::string::size_type position = sockPath.find_last_of('/');
    if (position == std::string::npos) {
        MFS_LOG_WARN("get service install path failed : invalid folder path.");
        return MFS_INVALID_CONFIG;
    }
    sockPath.append(SOCKET_PATH_SUFFIX);
    return MFS_OK;
}

MResult IpcServer::CreateService()
{
    MFS_LOG_INFO("Starting Ipc Server");
    if (mService != nullptr) {
        MFS_LOG_ERROR("service already created");
        return MFS_ERROR;
    }

    if (CreateSocketPath(socketPath) != MFS_OK) {
        MFS_LOG_WARN("failed to create socket path");
        return MFS_ERROR;
    }

    uint16_t sockPerm = 0600;
    std::string udsUrl = std::string("uds://") + socketPath + ":" + std::to_string(sockPerm);
    ock::hcom::UBSHcomServiceOptions options{};
    options.workerGroupMode = ock::hcom::UBSHcomWorkerMode::NET_EVENT_POLLING;
    options.maxSendRecvDataSize = MAX_MESSAGE_SIZE + 2048; // RECEIVE_SEG_SIZE = 2048

    mService = UBSHcomService::Create(UBSHcomServiceProtocol::SHM, "ipc_server", options);
    if (mService == nullptr) {
        MFS_LOG_ERROR("failed to create service already created");
        return MFS_ERROR;
    }

    UBSHcomTlsOptions tlsOpt;
    tlsOpt.enableTls = false;
    mService->SetTlsOptions(tlsOpt);

    mService->SetMaxSendRecvDataCount(1024); // SEG_COUNT = 1024
    mService->SetQueuePrePostSize(32); // RECEIVE_SIZE_PER_QP = 32
    mService->SetSendQueueSize(64); // QUEUE_SIZE = 64
    mService->SetPollingBatchSize(16); // SEND_SIZE = 16
    mService->Bind(udsUrl,
        std::bind(&IpcServer::NewChannel, this, std::placeholders::_1, std::placeholders::_2, std::placeholders::_3));
    mService->RegisterChannelBrokenHandler(std::bind(&IpcServer::ChannelBroken, this, std::placeholders::_1),
        ock::hcom::UBSHcomChannelBrokenPolicy::BROKEN_ALL);
    mService->RegisterRecvHandler(std::bind(&IpcServer::RequestReceived, this, std::placeholders::_1));
    mService->RegisterSendHandler(IpcServer::RequestPosted);
    mService->RegisterOneSideHandler(IpcServer::OneSideDone);

    int32_t result;
    if ((result = mService->Start()) != 0) {
        MFS_LOG_ERROR("failed to start service " << result);
        ock::common::HLOG_AUDIT("system", "create instance", "net server", "fail");
        return MFS_ERROR;
    }
    MFS_LOG_INFO("Ipc server started");

    ock::common::HLOG_AUDIT("system", "create instance", "net server", "success");
    return MFS_OK;
}

bool IpcServer::NewConnectionCheckLimit()
{
    std::unique_lock<std::mutex> lockGuard{ connCntMutex };
    if (totalConnCount >= MemFsConstants::MAX_IPC_CONNECTIONS_COUNT) {
        return false;
    }

    totalConnCount++;
    return true;
}

void IpcServer::CloseConnectionForLimit()
{
    std::unique_lock<std::mutex> lockGuard{ connCntMutex };
    if (totalConnCount > 0L) {
        totalConnCount--;
    }
}

void IpcServer::RefreshTimeoutForChannel(const ock::hcom::UBSHcomChannelPtr &channel) noexcept
{
    auto newFuture = mAutoCloseTimer.Submit(
        [this, channel]() {
            MFS_LOG_INFO("channel: " << channel->GetId() << " timeout close.");
            mService->Disconnect(channel);
            ChannelBroken(channel);
        },
        std::chrono::hours(24));
    if (newFuture == nullptr) {
        MFS_LOG_ERROR("submit timer for channel:" << channel->GetId() << " auto close failed.");
        return;
    }

    std::unique_lock<std::mutex> lockGuard{ mAutoCloseMutex };
    auto pos = mChannelCloseFutures.find(channel->GetId());
    if (pos != mChannelCloseFutures.end()) {
        pos->second->Cancel();
        mChannelCloseFutures.erase(pos);
    }

    mChannelCloseFutures.emplace(channel->GetId(), newFuture);
}

void IpcServer::CancelTimeoutForChannel(uint64_t channelId) noexcept
{
    std::unique_lock<std::mutex> lockGuard{ mAutoCloseMutex };
    auto pos = mChannelCloseFutures.find(channelId);
    if (pos != mChannelCloseFutures.end()) {
        pos->second->Cancel();
        mChannelCloseFutures.erase(pos);
    }
}

int32_t IpcServer::NewChannel(const std::string &ipPort, const ChannelPtr &newChannel, const std::string &payload)
{
    MFS_LOG_INFO("a new channel from " << ipPort << " payload " << payload);

    auto &config = ServiceConfigure::GetInstance().GetIpcMessageConfig();
    MFS_LOG_INFO("Client authorEnabled:" << config.authorEnabled << ", authorEncrypted:" << config.authorEncrypted);
    HLOG_AUDIT("system", "connect from client", "", "success");

    if (!NewConnectionCheckLimit()) {
        MFS_LOG_ERROR("Too many connections.");
        ock::common::HLOG_AUDIT("net server", "connectionCount", "", "fail");
        return MFS_ERROR;
    }

    if (mHandleNewChannel != nullptr) {
        mHandleNewChannel(newChannel);
    }

    RefreshTimeoutForChannel(newChannel);
    return MFS_OK;
}

int32_t IpcServer::CheckConnectUser(pid_t pid, uid_t user, gid_t group) const
{
    auto &config = ServiceConfigure::GetInstance();
    if (!config.GetIpcMessageConfig().permitSuperUser && user == 0) {
        MFS_LOG_WARN("a new channel from pid(" << pid << ") super user not allowed.");
        return MFS_ERROR;
    }

    if (!config.GetMemFsConfig().multiGroupEnabled) {
        if (user == serverRunUid || user == 0 || group == serverRunGid) {
            return MFS_OK;
        }

        auto userGroupCache = UserGroupCache::GetSystemInstance();
        if (userGroupCache != nullptr && userGroupCache->UserInGroup(user, serverRunGid)) {
            return MFS_OK;
        }

        MFS_LOG_WARN("a new channel from pid(" << pid << ") user(" << user << ") group(" << group << ") not allowed.");
        return MFS_ERROR;
    }

    return MFS_OK;
}

void IpcServer::ChannelBroken(const ChannelPtr &ch)
{
    MFS_LOG_INFO("channel " << ch->GetId() << " broken.");
    UdsInfo info;
    ch->GetRemoteUdsIdInfo(info);
    ock::common::HLOG_AUDIT("system", "disconnect to client", "", "success");
    if (mHandleBrokenChannel != nullptr) {
        mHandleBrokenChannel(ch);
    }
    CloseConnectionForLimit();
    CancelTimeoutForChannel(ch->GetId());
}

int32_t IpcServer::RequestReceived(ServiceContext &ctx)
{
    if (LIKELY(ctx.OpCode() >= MAX_NEW_REQ_HANDLER)) {
        MFS_LOG_ERROR("Invalid opcode " << ctx.OpCode());
        return MFS_ERROR;
    }

    auto &handler = mHandlers[ctx.OpCode()];
    if (UNLIKELY(handler == nullptr)) {
        MFS_LOG_ERROR("Invalid opcode " << ctx.OpCode() << ", no handle registered");
        return MFS_ERROR;
    }

    UdsInfo info;
    ctx.Channel()->GetRemoteUdsIdInfo(info);

    RefreshTimeoutForChannel(ctx.Channel());
    return handler(ctx);
}

int32_t IpcServer::RequestPosted(const ServiceContext &ctx)
{
    return MFS_OK;
}

int32_t IpcServer::OneSideDone(const ServiceContext &ctx)
{
    return MFS_OK;
}