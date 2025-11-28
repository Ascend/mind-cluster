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
#include <csignal>
#include <fcntl.h>
#include <pwd.h>
#include <sys/wait.h>
#include <sys/resource.h>
#include "auditlog_adapt.h"
#include "ipc_client.h"

using namespace ock::memfs;
using namespace ock::hcom;
using namespace ock::common;

static constexpr auto LOCK_FILE_MAX_TIME = 60;
static constexpr auto MAX_PROC_OPEN_FILE = 1024UL;

static void IpcClientLog(int level, const char *msg)
{
    ASSERT_RET_VOID(msg != nullptr);
    switch (level) {
        case 0:
            LOG_DEBUG(msg);
            break;
        case 1: // 1
            LOG_INFO(msg);
            break;
        case 2: // 2
            LOG_WARN(msg);
            break;
        case 3: // 3
            LOG_ERROR(msg);
            break;
        default:
            LOG_WARN("invalid level " << level << ", " << msg);
            break;
    }
}

IpcClient::IpcClient(const std::function<int()> &connectCb, const std::function<void()> &disconnectCb,
    const std::map<std::string, std::string> &serverInfoParam)
    : mConnectCallback{ connectCb }, mDisconnectCallback{ disconnectCb }, mServerInfoParam{ serverInfoParam }
{}

MResult IpcClient::Start()
{
    IpcClientConfig defConfig;
    return Start(defConfig);
}

MResult IpcClient::Start(const IpcClientConfig &config)
{
    std::lock_guard<std::mutex> guard(mMutex);
    if (mStarted) {
        LOG_WARN("IpcClient has been already started");
        return MFS_OK;
    }

    auto result = CreateService(config);
    ASSERT_RETURN(result == MFS_OK, result);

    mStarted = true;
    return MFS_OK;
}

void IpcClient::Stop()
{
    using namespace hcom;
    std::lock_guard<std::mutex> guard(mMutex);
    if (!mStarted || mService == nullptr) {
        LOG_WARN("IpcClient was not started");
        return;
    }
    if (mChannel != nullptr) {
        mService->Disconnect(mChannel);
    }
    mService->Destroy("ipcClient");
    if (mService != nullptr) {
        mService = nullptr;
    }
    mStarted = false;
}

MResult IpcClient::CreateService(const IpcClientConfig &config)
{
    if (mService != nullptr) {
        LOG_ERROR("service already created");
        return MFS_ERROR;
    }

    ock::hcom::UBSHcomServiceOptions options{};
    options.workerGroupMode = ock::hcom::UBSHcomWorkerMode::NET_EVENT_POLLING;
    options.maxSendRecvDataSize = MAX_MESSAGE_SIZE + RECEIVE_SEG_SIZE;

    mService = UBSHcomService::Create(UBSHcomServiceProtocol::SHM, "ipcClient", options);
    if (mService == nullptr) {
        LOG_ERROR("failed to create service already created");
        return MFS_ERROR;
    }
    auto serviceLog = ock::hcom::NetLogger::Instance();
    if (serviceLog == nullptr) {
        LOG_ERROR("get hcom logger failed.");
        return MFS_ERROR;
    }
    serviceLog->SetExternalLogFunction(IpcClientLog);

    UBSHcomTlsOptions tlsOpt;
    tlsOpt.enableTls = false;
    mService->SetTlsOptions(tlsOpt);

    mService->SetMaxSendRecvDataCount(SEG_COUNT);
    mService->SetQueuePrePostSize(RECEIVE_SIZE_PER_QP);
    mService->SetSendQueueSize(QUEUE_SIZE);
    mService->SetRecvQueueSize(SEND_SIZE);
    mService->RegisterChannelBrokenHandler(std::bind(&IpcClient::ChannelBroken, this, std::placeholders::_1),
        ock::hcom::UBSHcomChannelBrokenPolicy::BROKEN_ALL);
    mService->RegisterRecvHandler(std::bind(&IpcClient::RequestReceived, this, std::placeholders::_1));
    mService->RegisterSendHandler(std::bind(&IpcClient::RequestPosted, this, std::placeholders::_1));
    mService->RegisterOneSideHandler(std::bind(&IpcClient::OneSideDone, this, std::placeholders::_1));

    int32_t result;
    if ((result = mService->Start()) != 0) {
        LOG_ERROR("failed to start service " << result);
        return MFS_ERROR;
    }

    LOG_INFO("service started");
    return MFS_OK;
}

MResult IpcClient::Connect()
{
    using namespace hcom;
    ASSERT_RETURN(mService != nullptr, MFS_ERROR);
    if (mChannel.Get() != nullptr) {
        return MFS_OK;
    }

    if (PrepareServerWorkerDir() != 0) {
        return -1;
    }
    ock::hcom::UBSHcomConnectOptions options{};
    options.linkCount = EP_SIZE;

    if (ClientConnectServerProcess() != 0) {
        return -1;
    }

    auto startTime = std::chrono::high_resolution_clock::now();
    std::string udsUrl = std::string("uds://") + mSocketFullPath;

    while (mServerStatus.status != RUNNING) {
        LOG_INFO("client check server status...");
        std::this_thread::sleep_for(std::chrono::seconds(1));
        auto curTime = std::chrono::high_resolution_clock::now();
        if (std::chrono::duration_cast<std::chrono::seconds>(curTime - startTime).count() > DEFAULT_TIMEOUT) {
            LOG_ERROR("server process init failed or timeout.");
            return MFS_ERROR;
        }
        if (!connectFlag && mService->Connect(udsUrl, mChannel, options) != 0) {
            LOG_WARN("client try connect failed, wait server process ready...");
            continue;
        }
        if (flockFlag) {
            UnlockFile();
        }

        connectFlag = true;
        mChannelConnected.store(true);
        mTimeout = CHANNEL_DEFAULT_TIMEOUT;
        mChannel->SetChannelTimeOut(mTimeout, mTimeout);
        LOG_INFO("connect to server success, channelId " << mChannel->GetId() << ", set timeout(s)" << mTimeout);

        auto result = GetServerStatus();
        if (result != MFS_OK) {
            return result;
        }
    }
    mServiceable = true;
    return MFS_OK;
}

void IpcClient::ShutDownConnection()
{
    std::unique_lock<std::mutex> lockGuard{ mConnectionMutex };
    if (mChannel == nullptr) {
        return;
    }

    LOG_INFO("channel " << mChannel->GetId() << " shutdown");
    mDisconnectCallback();
    mChannel = nullptr;
    mServiceable = false;
}

void IpcClient::RestoreConnection()
{
    constexpr uint32_t attempt = 3;
    auto result = 0;
    ock::hcom::UBSHcomConnectOptions options{};
    options.linkCount = EP_SIZE;
    std::string udsUrl = std::string("uds://") + mSocketFullPath;

    std::unique_lock<std::mutex> lockGuard{ mConnectionMutex };
    for (uint32_t i = 0; i < attempt; ++i) {
        if ((result = mService->Connect(udsUrl, mChannel, options)) == 0) {
            mChannelConnected.store(true);
            LOG_WARN("connect success: netClient channel " << mChannel.Get());
            break;
        }
    }

    if (result != 0) {
        LOG_ERROR("failed to connect to server, result " << result);
        return;
    }

    result = mConnectCallback();
    if (result != MFS_OK) {
        LOG_ERROR("connected callback invoke failed: " << result);
        if (mChannel != nullptr) {
            mService->Disconnect(mChannel);
        }
    }
}

void IpcClient::ChannelBroken(const ChannelPtr &ch)
{
    mChannelConnected.store(false);
    LOG_INFO("channel " << ch->GetId() << " broken.");
    ShutDownConnection();
}

int32_t IpcClient::RequestReceived(const ServiceContext &ctx)
{
    return 0;
}

int32_t IpcClient::RequestPosted(const ServiceContext &ctx)
{
    return MFS_OK;
}

int32_t IpcClient::OneSideDone(const ServiceContext &ctx)
{
    return MFS_OK;
}

bool IpcClient::LockFile()
{
    std::string lockfile = mServerWorkerPath + "/.lockfile";
    fileDesc = open(lockfile.c_str(), O_CREAT | O_RDWR, S_IRUSR | S_IWUSR);
    if (fileDesc == -1) {
        LOG_ERROR("create or open lock file failed, errno: " << strerror(errno));
        return false;
    }
    struct flock fl {};
    fl.l_type = F_WRLCK;
    fl.l_whence = SEEK_SET;
    fl.l_start = 0;
    fl.l_len = 0;

    if (fcntl(fileDesc, F_SETLK, &fl) != 0) {
        close(fileDesc);
        fileDesc = -1;
        return false;
    }
    flockFlag = true;
    return true;
}

void IpcClient::UnlockFile()
{
    std::string lockfile = mServerWorkerPath + "/.lockfile";
    struct flock fl {};
    fl.l_type = F_UNLCK;
    fl.l_whence = SEEK_SET;
    fl.l_start = 0;
    fl.l_len = 0;

    if (fcntl(fileDesc, F_SETLK, &fl) != 0) {
        return;
    }
    close(fileDesc);
    unlink(lockfile.c_str());
}

int IpcClient::CreateDirectory(const std::string &path, mode_t mode)
{
    size_t pre = 0;
    size_t pos;
    std::string dir;

    if (path[0] == '/') {
        dir += '/';
    }
    while ((pos = path.find_first_of('/', pre)) != std::string::npos) {
        dir = path.substr(0, pos++);
        pre = pos;
        if (dir.empty()) {
            continue;
        }
        if (mkdir(dir.c_str(), mode) && errno != EEXIST) {
            LOG_ERROR("create dir failed, dir: " << dir << ", error: " << strerror(errno));
            return -1;
        }
    }
    if (mkdir(path.c_str(), mode) && errno != EEXIST) {
        LOG_ERROR("create dir failed, path: " << path << ", error: " << strerror(errno));
        return -1;
    }
    return 0;
}

int IpcClient::PrepareServerWorkerDir()
{
    for (const auto &item : mServerInfoParam) {
        if (item.first == "server.worker.path") {
            if (!item.second.empty() && item.second.back() == '/') {
                mServerWorkerPath = item.second.substr(0, item.second.size() - 1);
            } else {
                mServerWorkerPath = item.second;
            }
        } else if (item.first == "server.ockiod.path") {
            mServerOckiodPath = item.second;
        }
    }
    if (CreateDirectory(mServerWorkerPath) != 0) {
        LOG_ERROR("create dir failed, cur dir path :" << mServerWorkerPath);
        return -1;
    }
    mSocketFullPath = mServerWorkerPath + "/uds/mindio_memfs_123.s";
    return 0;
}

int IpcClient::CreateServerWorkDir()
{
    if (FileUtil::Exist(mSocketFullPath)) {
        if (!FileUtil::Remove(mSocketFullPath)) {
            LOG_ERROR("remove file failed, errno: " << strerror(errno));
            return -1;
        }
    }

    for (const auto &dir : DEFAULT_SERVER_DIR) {
        std::string dirPath = mServerWorkerPath;
        dirPath.append(dir.first);
        auto ret = CreateDirectory(dirPath, dir.second);
        if (ret != 0) {
            LOG_ERROR("create dir failed, cur dir path :" << dirPath);
            return -1;
        }
    }
    return 0;
}

int IpcClient::CreateDefaultMemfsConf()
{
    std::string memfsFile = mServerWorkerPath;
    memfsFile.append("/conf/memfs.conf");
    std::ofstream file(memfsFile);
    if (!file.is_open()) {
        LOG_ERROR("open memfs conf file failed. errno: " << errno << ", error: " << strerror(errno));
        return -1;
    }
    file << "[memfs]\n";
    for (const auto &item : mServerInfoParam) {
        if (item.first == "server.worker.path" || item.first == "server.ockiod.path") {
            continue;
        }
        file << item.first << " = " << item.second << "\n";
    }

    file.close();
    return 0;
}

int IpcClient::DaemonInit()
{
    pid_t pid;
    struct rlimit r1 {};
    struct sigaction sa {};

    if (getrlimit(RLIMIT_NOFILE, &r1) < 0) {
        LOG_ERROR("get file limit failed.");
        return MFS_ERROR;
    }
    if ((pid = fork()) < 0) {
        return MFS_ERROR;
    } else if (pid > 0) {
        _exit(EXIT_SUCCESS);
    }

    if (setsid() < 0) {
        return MFS_ERROR;
    }

    sa.sa_handler = SIG_IGN;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = 0;
    if (sigaction(SIGHUP, &sa, nullptr) < 0) {
        LOG_ERROR("ignore SIGHUP failed.");
        return MFS_ERROR;
    }

    if ((pid = fork()) < 0) {
        return MFS_ERROR;
    } else if (pid > 0) {
        _exit(EXIT_SUCCESS);
    }

    if (chdir("/") < 0) {
        return MFS_ERROR;
    }

    if (r1.rlim_max > MAX_PROC_OPEN_FILE) {
        r1.rlim_max = MAX_PROC_OPEN_FILE;
    }
    for (auto i = 0; i < r1.rlim_max; i++) {
        close(i);
    }
    std::string daemonLog = mServerWorkerPath + "/logs/ockiod_daemon.log";
    int fd = open(daemonLog.c_str(), O_CREAT | O_RDWR, S_IRUSR | S_IWUSR | S_IRGRP);
    if (fd < 0) {
        LOG_ERROR("open file failed, error: " << strerror(errno));
        return MFS_ERROR;
    }
    dup2(fd, STDIN_FILENO);
    dup2(fd, STDOUT_FILENO);
    dup2(fd, STDERR_FILENO);
    close(fd);

    return MFS_OK;
}

int IpcClient::CleanProcessBeforeRePull()
{
    std::string pidFile = mServerWorkerPath + "/.ockiod.pid";
    std::ifstream file(pidFile);
    if (!file.is_open()) {
        if (errno != ENOENT) {
            LOG_ERROR("open .ockiod.pid failed, error: " << strerror(errno));
            return MFS_ERROR;
        }
        return MFS_OK;
    }
    std::string dPidStr;
    if (std::getline(file, dPidStr)) {
        uint32_t dPidInt = -1;
        if (dPidStr.empty() || !StrUtil::StrToUint(dPidStr, dPidInt)) {
            LOG_ERROR("failed get daemon pid, .ockiod.pid file is empty or content invalid.");
            return MFS_ERROR;
        }
        auto dPid = static_cast<pid_t>(dPidInt);
        file.close();
        int status;
        auto result = waitpid(dPid, &status, WNOHANG);
        // recorded pid is not child process id.
        if (result == -1 && errno == ECHILD) {
            return MFS_OK;
        }
        if (kill(dPid, 0) == 0) {
            if (kill(dPid, SIGABRT) != 0) {
                LOG_ERROR("failed to send sig to pid " << dPid << ", errno: " << strerror(errno));
                return MFS_ERROR;
            }
            result = waitpid(dPid, &status, 0);
            if (result == -1 && errno != ECHILD) {
                LOG_ERROR("wait pid exit failed, errno: " << strerror(errno));
                return MFS_ERROR;
            }
        }
        return MFS_OK;
    }
    return MFS_ERROR;
}

int IpcClient::ClientForkSubProcess()
{
    int result = CleanProcessBeforeRePull();
    if (result != MFS_OK) {
        return result;
    }
    pid_t pid = fork();
    if (pid < 0) {
        LOG_ERROR("client fork sub process failed.");
        return MFS_ERROR;
    } else if (pid == 0) {
        LOG_INFO("client launch one sub process success, pid is:" << getpid() << ", process group:" <<
            getpgid(getpid()));
        auto ret = DaemonInit();
        if (ret != 0) {
            return ret;
        }
        pid_t detachPid = getpid();
        LOG_INFO("daemon process pull success, pid is:" << detachPid << ", process group:" << getpgid(detachPid));
        std::string pidFile = mServerWorkerPath + "/.ockiod.pid";
        int writeFd = open(pidFile.c_str(), O_CREAT | O_TRUNC | O_RDWR, S_IRUSR | S_IWUSR);
        if (writeFd == -1) {
            LOG_ERROR("create or open .ockiod.pid failed, error: " << strerror(errno));
            return MFS_ERROR;
        }
        std::string pidStr = std::to_string(detachPid);
        if (write(writeFd, pidStr.c_str(), pidStr.size()) != pidStr.size()) {
            LOG_ERROR("write pid to .ockiod.pid failed, error: " << strerror(errno));
            close(writeFd);
            return MFS_ERROR;
        }
        close(writeFd);
        char *setupArgs[ARGS_TOTAL_PLACEHOLDER + 1];
        std::string targetPath = mServerOckiodPath;
        setupArgs[PROCESS_PLACEHOLDER] = const_cast<char *>(targetPath.c_str());
        setupArgs[WORKER_PATH_PLACEHOLDER] = const_cast<char *>(mServerWorkerPath.c_str());
        setupArgs[ARGS_TOTAL_PLACEHOLDER] = nullptr;
        if (execvp(setupArgs[0], setupArgs) == -1) {
            LOG_ERROR("client exec sub process failed, errno@" << errno << " : " << strerror(errno));
            return MFS_ERROR;
        };
        LOG_INFO("client launch server process exec success.");
    } else {
        sleep(1);
        LOG_INFO("Parent process continues....");
    }
    return MFS_OK;
}


MResult IpcClient::GetServerStatus()
{
    ServerStatusReq req;
    req.flags = 0;

    ServerStatusResp resp{};
    auto result = SyncCall<ServerStatusReq, ServerStatusResp>(IPC_OP_GET_SERVER_STATUS, req, resp);
    LOG_ERROR_RETURN_IT_IF_NOT_OK(result, "Failed to call server to get status messages, result " << result);

    if (resp.result != MFS_OK) {
        LOG_ERROR("Failed get server status message, server error " << resp.result);
        return resp.result;
    }
    mServerStatus = resp;

    return MFS_OK;
}

MResult IpcClient::ClientConnectServerProcess()
{
    int32_t result = 0;
    ock::hcom::UBSHcomConnectOptions options{};
    options.linkCount = EP_SIZE;
    std::string udsUrl = std::string("uds://") + mSocketFullPath;
    if (mService->Connect(udsUrl, mChannel, options) != 0) {
        LOG_WARN("try connect times 0 failed.");
        if (!LockFile()) {
            return MFS_OK;
        }
        if (mService->Connect(udsUrl, mChannel, options) == 0) {
            connectFlag = true;
            UnlockFile();
            return MFS_OK;
        }
        LOG_WARN("lock file and try connect times 1 failed.");

        if (CreateServerWorkDir() != 0 || CreateDefaultMemfsConf() != 0) {
            UnlockFile();
            return MFS_ERROR;
        }
        LOG_INFO("server worker dir and memfs conf prepare finished.");
        result = ClientForkSubProcess();
        if (result != 0) {
            UnlockFile();
            return MFS_ERROR;
        }
        if (mService->Connect(udsUrl, mChannel, options) == 0) {
            connectFlag = true;
            return MFS_OK;
        }
        LOG_WARN("lock file and try connect times 2 failed.");
    } else {
        connectFlag = true;
    }
    return MFS_OK;
}
