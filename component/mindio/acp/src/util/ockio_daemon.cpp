/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
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

#include <iostream>
#include <csignal>

#include "memfs_file_util.h"

#include "hlog.h"
#include "auditlog_adapt.h"
#include "fs_server.h"
#include "mem_fs_state.h"
#include "service_configure.h"
#include "under_fs_manager.h"
#include "background_manager.h"
#include "memfs_api.h"

using namespace ock::common::config;
using namespace ock::hlog;
using namespace ock::memfs;

static constexpr int MFS_LOG_TYPE = 1; // file
static constexpr int MFS_LOG_LEVEL = SPDLOG_LEVEL_INFO;
static constexpr int MFS_LOG_FILE_SIZE = 100 * 1024 * 1024; // 100MB
static constexpr int MFS_LOG_FILE_COUNT = 50;
static constexpr int MFS_AUDIT_LOG_FILE_SIZE = 100 * 1024 * 1024; // 100MB
static constexpr int MFS_AUDIT_LOG_FILE_COUNT = 50;

static sem_t g_mainSem;

static int LoggerInitialize()
{
    auto path = ServiceConfigure::GetInstance().GetWorkPath();

    std::string::size_type position = path.find_last_of('/');
    if (position == std::string::npos) {
        std::cerr << "create logger path failed : invalid folder path." << std::endl;
        return -1;
    }
    path.append("/logs");
    auto ret = mkdir(path.c_str(), S_IRUSR | S_IWUSR | S_IXUSR);
    if (ret != 0 && errno != EEXIST) {
        std::cerr << "create logger path failed : " << strerror(errno) << std::endl;
        return -1;
    }

    auto auditLogPath = path;
    auditLogPath.append("/ockiod_audit.log");
    ret = Hlog::CreateInstanceAudit(auditLogPath.c_str(), MFS_AUDIT_LOG_FILE_SIZE, MFS_AUDIT_LOG_FILE_COUNT);
    if (ret != 0) {
        std::cerr << "create HLOG AuditLog instance failed : " << ret << std::endl;
        return -1;
    }

    auto logPath = path;
    logPath.append("/ockiod.log");
    ret = Hlog::CreateInstance(MFS_LOG_TYPE, MFS_LOG_LEVEL, logPath.c_str(), MFS_LOG_FILE_SIZE, MFS_LOG_FILE_COUNT);
    if (ret != 0) {
        std::cerr << "create HLOG instance failed : " << ret << std::endl;
        return -1;
    }

    return 0;
}

static int DaemonModuleInitialize()
{
    auto ret = ock::ufs::UnderFsManager::GetInstance().Initialize();
    if (ret != 0) {
        std::cerr << "failed to initialize under fs : " << ret << std::endl;
        return -1;
    }

    ret = ock::bg::BackgroundManager::GetInstance().Initialize();
    if (ret != 0) {
        std::cerr << "failed to initialize background : " << ret << std::endl;
        return -1;
    }

    return 0;
}

static void WaitingSignalHandler(int signal)
{
    sem_post(&g_mainSem);
}

static void RegisterWaitingSignal()
{
    sighandler_t oldIntHandler = signal(SIGINT, WaitingSignalHandler);
    if (oldIntHandler == SIG_ERR) {
        MFS_LOG_ERROR("register SIGINT handler failed");
    }
}

static int SetMainWaiting()
{
    RegisterWaitingSignal();

    int ret = sem_init(&g_mainSem, 0, 0);
    if (ret != 0) {
        MFS_LOG_ERROR("init start main sem failed:[" << ret << "][" << errno << "]");
        return -1;
    }

    MFS_LOG_INFO("begin set main waiting");
    while (true) {
        struct timespec ts {};
        clock_gettime(CLOCK_REALTIME, &ts);
        ts.tv_sec += 1; // wait 1 second base on current time
        int wait = sem_timedwait(&g_mainSem, &ts);
        if (wait == 0) {
            MFS_LOG_WARN("catch sem, received exit signal.");
            break;
        } else if (errno != ETIMEDOUT) { // no necessary deal default errno ETIMEDOUT
            MFS_LOG_WARN("received other signal errno[" << errno << "].");
        }
    }
    sem_destroy(&g_mainSem);
    return 0;
}


static void PageFaultSignalHandler(int signal)
{
    std::cerr << "received exit signal[" << signal << "]" << std::endl;
    _exit(EXIT_FAILURE);
}

static void RegisterPageFaultSignal()
{
    sighandler_t oldTermHandler = signal(SIGSEGV, PageFaultSignalHandler);
    if (oldTermHandler == SIG_ERR) {
        std::cerr << "register SIGSEGV handler failed" << std::endl;
    }

    sighandler_t oldHupHandler = signal(SIGBUS, PageFaultSignalHandler);
    if (oldHupHandler == SIG_ERR) {
        std::cerr << "register SIGBUS handler failed" << std::endl;
    }
}

int main(int argc, char *argv[])
{
    RegisterPageFaultSignal();
    if (argc <= 1 || argv == nullptr) {
        return -EINVAL;
    }

    if (argv[1] == nullptr) {
        std::cerr << "Failed , the workspace path is null. " << std::endl;
        return -EINVAL;
    }

    if (std::string(argv[1]).empty()) {
        std::cerr << "Failed to obtain the workspace path. " << std::endl;
        return -EINVAL;
    }

    auto result = ::setenv("HCOM_FILE_PATH_PREFIX", argv[1], 1);
    if (result) {
        std::cerr << "set hcom path failed , error :" << strerror(errno) << std::endl;
        return -EINVAL;
    }

    ServiceConfigure::GetInstance().SetWorkPath(std::string(argv[1]));

    if (LoggerInitialize() != 0) {
        return -1;
    }

    auto confRet = ServiceConfigure::GetInstance().Initialize();
    if (confRet != 0) {
        std::cerr << "failed to initialize service configure : " << confRet << std::endl;
        return -1;
    }

    auto memFs = ShellFSServer::Instance();
    MResult ret = memFs->Start();
    if (ret != MFS_OK) {
        std::cerr << "Memory file system start failed : " << ret << std::endl;
        HLOG_AUDIT("system", "start", "MemFS", "failed");
        return -1;
    }
    MemfsState::Instance().SetState(MemfsStateCode::STARTING, MemfsStartProgress::HUNDRED_PERCENT);
    MFS_LOG_INFO("Memory file system STARTED.");

    if (DaemonModuleInitialize() != 0) {
        HLOG_AUDIT("system", "start", "MemFS", "failed");
        return -1;
    }
    MemfsState::Instance().SetState(MemfsStateCode::RUNNING);
    SetMainWaiting();
    memFs->Stop();
    return 0;
}
