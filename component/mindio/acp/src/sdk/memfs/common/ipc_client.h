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
#ifndef OCK_MEMFS_CORE_IPC_CLIENT_H
#define OCK_MEMFS_CORE_IPC_CLIENT_H

#include <cstdlib>
#include "common_includes.h"
#include "ipc_message.h"

namespace ock {
namespace memfs {
constexpr uint32_t EP_SIZE = 4;
constexpr uint32_t SEND_SIZE = 16;
constexpr uint32_t RECEIVE_SIZE_PER_QP = 32;
constexpr uint32_t QUEUE_SIZE = 64;
constexpr uint32_t SEG_COUNT = 1024;
constexpr uint32_t RECEIVE_SEG_SIZE = 2048;

constexpr uint32_t PROCESS_PLACEHOLDER = 0;
constexpr uint32_t WORKER_PATH_PLACEHOLDER = 1;
constexpr uint32_t ARGS_TOTAL_PLACEHOLDER = 2;
constexpr uint32_t DEFAULT_TIMEOUT = 120;
constexpr uint32_t CHANNEL_DEFAULT_TIMEOUT = 60;

struct IpcClientConfig {
    const bool tlsEnabled;
    const std::string certPath;
    const std::string caPath;
    const std::string crlPath;
    const std::string priKeyPath;
    const std::string passwordPath;
    const std::string pmtPath;

    IpcClientConfig() : tlsEnabled{ false } {}
    IpcClientConfig(std::string cert, std::string ca, std::string pk, std::string pw, std::string pmt)
        : IpcClientConfig{ std::move(cert), std::move(ca), "", std::move(pk), std::move(pw), std::move(pmt) }
    {}
    IpcClientConfig(std::string cert, std::string ca, std::string crl, std::string pk, std::string pw, std::string pmt)
        : tlsEnabled{ true },
          certPath{ std::move(cert) },
          caPath{ std::move(ca) },
          crlPath{ std::move(crl) },
          priKeyPath{ std::move(pk) },
          passwordPath{ std::move(pw) },
          pmtPath{ std::move(pmt) }
    {}
};

/**
 * @brief Currently only support one thread one client yet,
 * multiple thread need to be done later
 */
class IpcClient {
public:
    IpcClient(const std::function<int()> &connectCb, const std::function<void()> &disconnectCb,
        const std::map<std::string, std::string> &serverInfoParam);
    MResult Start();
    MResult Start(const IpcClientConfig &config);
    void Stop();

    MResult Connect();
    void ShutDownConnection();
    void RestoreConnection();
    MResult GetServerStatus();
    template <typename TReq, typename TResp> MResult SyncCall(FileOpCode opCode, TReq &req, TResp &resp)
    {
        using namespace hcom;
        if (mChannel == nullptr) {
            RestoreConnection();
        }

        ock::hcom::UBSHcomRequest reqMsg(static_cast<void *>(&req), sizeof(TReq), opCode);
        ock::hcom::UBSHcomResponse respMsg(&resp, sizeof(resp));
        static constexpr int maxRetryTimes = 3;
        for (auto i = 0; i < maxRetryTimes; i++) {
            if (mChannel == nullptr) {
                RestoreConnection();
                continue;
            }
            auto result = mChannel->Call(reqMsg, respMsg);
            if (UNLIKELY(result == ock::hcom::SerCode::SER_NOT_ESTABLISHED)) {
                RestoreConnection();
                continue;
            }

            if (UNLIKELY(result != MFS_OK || respMsg.errorCode != MFS_OK)) {
                LOG_ERROR("Failed to call server with op " << opCode << ", result " << UBSHcomNetErrStr(result) <<
                    ", error code " << respMsg.errorCode);
                return result;
            }

            break;
        }

        return MFS_OK;
    }

    template <typename TReq, typename TResp> MResult SyncCall(FileOpCode opCode, TReq &req, TResp **resp, uint32_t &len)
    {
        using namespace hcom;
        ASSERT_RETURN(resp != nullptr, MFS_INVALID_PARAM);
        if (mChannel == nullptr) {
            RestoreConnection();
        }

        ock::hcom::UBSHcomRequest reqMsg(static_cast<void *>(&req), sizeof(TReq), opCode);
        ock::hcom::UBSHcomResponse respMsg;
        static constexpr int maxRetryTimes = 3;
        for (auto i = 0; i < maxRetryTimes; i++) {
            if (mChannel == nullptr) {
                RestoreConnection();
                continue;
            }
            auto result = mChannel->Call(reqMsg, respMsg);
            if (UNLIKELY(result == ock::hcom::SerCode::SER_NOT_ESTABLISHED)) {
                RestoreConnection();
                continue;
            }

            if (UNLIKELY(result != MFS_OK || respMsg.errorCode != MFS_OK)) {
                LOG_ERROR("Failed to call server with op " << opCode << ", result " << UBSHcomNetErrStr(result) <<
                    ", error code " << respMsg.errorCode);
                return result;
            }

            break;
        }

        ASSERT_RETURN(respMsg.size >= sizeof(TResp), MFS_ERROR);
        ASSERT_RETURN(respMsg.address != nullptr, MFS_ERROR);

        *resp = reinterpret_cast<TResp *>(respMsg.address);
        len = respMsg.size;
        return MFS_OK;
    }

    MResult ReceiveFD(int32_t &fd)
    {
        ASSERT_RETURN(mChannel.Get() != nullptr, MFS_NOT_INITIALIZED);
        int fds[1]{};
        auto result = mChannel->ReceiveFds(fds, 1, 10000);
        LOG_ERROR_RETURN_IT_IF_NOT_OK(result, "Failed to receive shared fd from server, result " << result);
        fd = fds[0];
        return MFS_OK;
    }

    inline bool GetIpcServiceStatus()
    {
        return mServiceable;
    }

#ifdef UT_ENABLED
    inline const ChannelPtr &GetChannel() const
    {
        return mChannel;
    }
#endif

private:
    bool LockFile();
    void UnlockFile();
    int CreateServerWorkDir();
    int PrepareServerWorkerDir();
    static int CreateDirectory(const std::string &path, mode_t mode = 0755);
    int CreateDefaultMemfsConf();
    int CleanProcessBeforeRePull();
    int ClientForkSubProcess();
    int DaemonInit();
    int ClientConnectServerProcess();
    int32_t CreateService(const IpcClientConfig &config);
    void ChannelBroken(const ChannelPtr &ch);
    int32_t RequestReceived(const ServiceContext &ctx);
    int RequestPosted(const ServiceContext &ctx);
    int OneSideDone(const ServiceContext &ctx);

private:
    hcom::UBSHcomService *mService = nullptr;
    hcom::UBSHcomChannelPtr mChannel = nullptr;
    int16_t mTimeout = -1;

    std::mutex mMutex;
    int fileDesc{ -1 };
    bool mStarted = false;
    bool connectFlag = false;
    bool flockFlag = false;
    std::mutex mConnectionMutex;
    std::atomic<bool> mChannelConnected{ false };
    const std::map<std::string, std::string> mServerInfoParam;
    std::string mServerWorkerPath;
    std::string mServerOckiodPath;
    std::string mSocketFullPath;
    ServerStatusResp mServerStatus{};
    const std::function<int()> mConnectCallback;
    const std::function<void()> mDisconnectCallback;
    bool mServiceable = false;
};
using IpcClientPtr = std::shared_ptr<IpcClient>;
}
}

#endif // OCK_MEMFS_CORE_IPC_CLIENT_H
