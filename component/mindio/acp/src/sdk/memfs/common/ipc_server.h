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
#ifndef OCK_MEMFS_CORE_IPC_SERVER_H
#define OCK_MEMFS_CORE_IPC_SERVER_H

#include "common_includes.h"
#include "ipc_message.h"
#include "memfs_logger.h"
#include "timer_executor.h"

namespace ock {
namespace memfs {
class IpcServer {
public:
    IpcServer();

    MResult Start();

    void Stop();

    MResult RegisterNewRequestHandler(uint32_t opCode, const NewRequestHandler &h)
    {
        std::lock_guard<std::mutex> guard(mMutex);
        if (opCode >= MAX_NEW_REQ_HANDLER) {
            MFS_LOG_ERROR("Invalid opCode " << opCode << " which should be less than " << MAX_NEW_REQ_HANDLER);
            return MFS_INVALID_PARAM;
        }

        if (mHandlers[opCode] != nullptr) {
            MFS_LOG_ERROR("Handler for opCode " << opCode << " already registered");
            return MFS_ALREADY_DONE;
        }

        mHandlers[opCode] = h;

        return MFS_OK;
    }

    MResult RegisterNewChannelHandler(const NewChannelHandler &h)
    {
        std::lock_guard<std::mutex> guard(mMutex);
        if (mHandleNewChannel != nullptr) {
            MFS_LOG_ERROR("Failed to register new channel handler");
            return MFS_ERROR;
        }

        mHandleNewChannel = h;
        return MFS_OK;
    }

    MResult RegisterChannelBrokenHandler(const ChannelBrokenHandler &h)
    {
        std::lock_guard<std::mutex> guard(mMutex);
        if (mHandleBrokenChannel != nullptr) {
            MFS_LOG_ERROR("Failed to register channel broken handler");
            return MFS_ERROR;
        }

        mHandleBrokenChannel = h;
        return MFS_OK;
    }

    inline bool IsStarted() const
    {
        return mStarted;
    }

    inline int64_t GetConnCnt()
    {
        std::lock_guard<std::mutex> guard(connCntMutex);
        return totalConnCount;
    }

private:
    int32_t CreateService();

    static int32_t CreateSocketPath(std::string &sockPath);

    int32_t NewChannel(const std::string &ipPort, const ChannelPtr &newChannel, const std::string &payload);

    int32_t CheckConnectUser(pid_t pid, uid_t user, gid_t group) const;

    void ChannelBroken(const ChannelPtr &ch);

    int32_t RequestReceived(ServiceContext &ctx);

    static int RequestPosted(const ServiceContext &ctx);

    static int OneSideDone(const ServiceContext &ctx);

    bool NewConnectionCheckLimit();

    void CloseConnectionForLimit();

    void RefreshTimeoutForChannel(const ock::hcom::UBSHcomChannelPtr &channel) noexcept;

    void CancelTimeoutForChannel(uint64_t channelId) noexcept;

private:
    static constexpr uint32_t MAX_NEW_REQ_HANDLER = 16;

private:
    hcom::UBSHcomService *mService = nullptr;
    NewRequestHandler mHandlers[MAX_NEW_REQ_HANDLER]{};
    NewChannelHandler mHandleNewChannel = nullptr;
    ChannelBrokenHandler mHandleBrokenChannel = nullptr;

    std::string socketPath;
    std::mutex mMutex;
    std::condition_variable mCond;
    bool mStarted = false;

    const uid_t serverRunUid;
    const gid_t serverRunGid;
    std::mutex connCntMutex;
    int64_t totalConnCount{ 0L };
    ock::common::TimerExecutor mAutoCloseTimer;
    std::unordered_map<uint64_t, std::shared_ptr<ock::common::Future>> mChannelCloseFutures;
    std::mutex mAutoCloseMutex;
};
using IpcServerPtr = std::shared_ptr<IpcServer>;
}
}

#endif // OCK_MEMFS_CORE_IPC_SERVER_H
