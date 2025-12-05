/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

#include <cstdarg>
#include <cstdio>
#include <atomic>
#include <iostream>
#include "spdlog/sinks/stdout_sinks.h"
#include "spdlog/sinks/rotating_file_sink.h"

#include "common.h"
#include "spdlog/common.h"
#include "spdlog/spdlog.h"
#include "rotating_sink.h"
#include "ttp_logger.h"

namespace ock {
namespace ttp {

static bool g_perProc = false;
static std::shared_ptr<spdlog::logger> gSPDLogger;

constexpr int MIN_LOG_LEVEL = 0;
constexpr int MAX_LOG_LEVEL = 5;
constexpr int ROTATION_FILE_SIZE_MAX = 10 * 1024 * 1024; // 10MB
constexpr int ROTATION_FILE_SIZE_MIN = 2 * 1024 * 1024;  //  2MB
constexpr int ROTATION_FILE_COUNT_MAX = 5;

struct FilePermissionHandler {
    // 日志文件打开后设置权限为640(rw-r-----)
    void AfterOpen(const std::string &filename)
    {
        chmod(filename.c_str(), S_IRUSR | S_IWUSR | S_IRGRP);
        spdlog::debug("Set permissions 640 for: {}", filename);
    }

    // 日志文件关闭后设置权限为440(r--r-----)
    void AfterClose(const std::string &filename)
    {
        chmod(filename.c_str(), S_IRUSR | S_IRGRP);
        spdlog::debug("Set permissions 440 for: {}", filename);
    }
};

void TTPLogger::Init()
{
    bool print = GetEnvValue2Uint32("TTP_LOG_STDOUT", 1, 1, 0);
    OutLogger::Instance()->SetLogLevel(static_cast<LogLevel>(GetLogLevel()));
    if (print) { return; }
    spdlog::set_level(static_cast<spdlog::level::level_enum>(GetLogLevel() + GAP));

    auto ret = TTPLogger::CreateLog();
    if (ret != TTP_OK) { return; }
    OutLogger::Instance()->SetExternalLogFunction([](int level, const char *msg) {
        std::string output = "[" + std::to_string(syscall(SYS_gettid)) + "]" + msg;
        TTPLogger::Log(level + GAP, output);
    });
}

int TTPLogger::GetLogLevel()
{
    std::map<std::string, int> logLevelMap = {
        {"DEBUG",   0},
        {"INFO",    1},
        {"WARNING", 2},
        {"ERROR",   3},
    };

    auto level = std::getenv("TTP_LOG_LEVEL");
    if (level == nullptr) {
        return logLevelMap["INFO"];
    }

    auto it = logLevelMap.find(std::string(level));
    auto logLevel = it != logLevelMap.end() ? it->second : logLevelMap["INFO"];
    return logLevel;
}

int TTPLogger::CreateLog()
{
    static std::atomic<bool> initState {false};
    bool inited = false;
    if (initState.compare_exchange_strong(inited, true)) {
        const char* logPath = std::getenv("TTP_LOG_PATH");
        if (logPath == nullptr || std::string(logPath).empty()) {
            logPath = "logs";
        }
        auto path = std::string(logPath) + "/ttp_log.log";
        auto ret = CreateLogImpl(GetLogLevel() + GAP, path);
        if (ret != TTP_OK) {
            std::cerr << "Create TTPLogger instance failed: " << ret << std::endl;
        }
        return ret;
    }
}

int TTPLogger::ValidateParams(const std::string &path)
{
    if (path.empty()) {
        std::cerr << "Invalid path, which is empty" << std::endl;
        return TTP_ERROR;
    }
    if (path.size() >= PATH_MAX) { // PATH_MAX时Linux就已报错
        std::cerr << "Invalid path, which exceeds the maximum value set by PATH_MAX" << std::endl;
        return TTP_ERROR;
    }
    if (FileUtils::IsSymlink(path)) {
        std::cerr << "Invalid path, which is a link" << std::endl;
        return TTP_ERROR;
    }
    return TTP_OK;
}

int TTPLogger::CreateLogImpl(int minLogLevel, std::string path, int rotationFileSize, int rotationFileCount)
{
    const char* mode = std::getenv("TTP_LOG_MODE");
    if (mode == nullptr || std::string(mode) != "ONLY_ONE") { // 默认PER_PROC
        path += "." + std::to_string(syscall(SYS_getpid));
        g_perProc = true;
    }
    auto ret = ValidateParams(path);
    if (ret != TTP_OK) {
        return ret;
    }

    size_t pos = 0;
    while ((pos = path.find_first_of('/', pos + 1)) != std::string::npos) {
        auto subdir = path.substr(0, pos);
        ret = mkdir(subdir.c_str(), S_IRUSR | S_IWUSR | S_IXUSR | S_IRGRP | S_IXGRP);
        if (ret != 0 && (errno != EEXIST || FileUtils::IsRegularFile(subdir.c_str()))) {
            std::cerr << "Create logger path failed: " << strerror(errno) << std::endl;
            return TTP_ERROR;
        }
    }

#ifdef UT_ENABLED
    rotationFileSize = GetEnvValue2Uint32("TTP_LOG_SIZE", 1, MFS_LOG_FILE_SIZE, rotationFileSize);
#endif

    std::string logName = "log:" + path;
    spdlog::file_event_handlers event_handlers{};
    event_handlers.after_open = [](const std::string &filename, std::FILE *) {
        FilePermissionHandler().AfterOpen(filename);
    };
    event_handlers.after_close = [](const std::string &filename) {
        FilePermissionHandler().AfterClose(filename);
    };

    if (g_perProc) { // PER_PROC模式 多线程间安全
        gSPDLogger = spdlog::rotating_logger_mt(logName, path, rotationFileSize,
            rotationFileCount, false, event_handlers);
    } else { // ONLY_ONE模式 多线程间/多进程间安全
        auto sink = std::make_shared<RotatingSink>(path, rotationFileSize,
            rotationFileCount, event_handlers);
        gSPDLogger = std::make_shared<spdlog::logger>(logName, sink);
        spdlog::register_logger(gSPDLogger);
    }

    if (gSPDLogger == nullptr) {
        return TTP_ERROR;
    }

    gSPDLogger->set_pattern("%Y-%m-%d %H:%M:%S.%f %t %l %v");
    spdlog::flush_every(std::chrono::seconds(1));
    gSPDLogger->set_level(static_cast<spdlog::level::level_enum>(minLogLevel));
    gSPDLogger->flush_on(spdlog::level::err);

    return TTP_OK;
}

void TTPLogger::Log(int level, const std::string &message)
{
    if (gSPDLogger == nullptr) {
        std::cerr << "spdlog logger is nullptr" << std::endl;
        return;
    }

    if (level < MIN_LOG_LEVEL || level > MAX_LOG_LEVEL) {
        std::cerr << "Invalid log level, which should be within 0 to 5" << std::endl;
        return;
    }

    gSPDLogger->log(static_cast<spdlog::level::level_enum>(level), "{}", message);
}

}  // namespace ttp
}  // namespace ock