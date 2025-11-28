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
#include <cstdarg>
#include <cstdio>
#include <iostream>
#include "spdlog/sinks/stdout_sinks.h"
#include "spdlog/sinks/rotating_file_sink.h"

#include "securec.h"
#include "spdlog/common.h"
#include "spdlog/spdlog.h"
#include "hlog.h"

namespace ock {
namespace hlog {
Hlog *Hlog::gLogger = nullptr;
thread_local std::string Hlog::gLastErrorMessage("");
Hlog *Hlog::gAuditLogger = nullptr;
static std::shared_ptr<spdlog::logger> gSPDLogger;      // spd logger
static std::shared_ptr<spdlog::logger> gAuditSPDLogger; // spd logger

const int STDOUT_TYPE = 0;
const int FILE_TYPE = 1;
const int STDERR_TYPE = 2;

constexpr int MIN_LOG_LEVEL_MAX = 5;
constexpr int ROTATION_FILE_SIZE_MAX = 100 * 1024 * 1024; // 100MB
constexpr int ROTATION_FILE_SIZE_MIN = 2 * 1024 * 1024;   // 2MB
constexpr int ROTATION_FILE_COUNT_MAX = 50;

int Hlog::ValidateParams(int logType, int minLogLevel, const char *path, int rotationFileSize, int rotationFileCount)
{
    if (logType != 0 && logType != 1 && logType != 2) { // 2
        gLastErrorMessage = "Invalid log type, which should be 0,1,2";
        return HLOG_ERR_INVALID_PARAM;
    } else if (minLogLevel < 0 || minLogLevel > MIN_LOG_LEVEL_MAX) {
        gLastErrorMessage = "Invalid min log level, which should be 0,1,2,3,4,5";
        return HLOG_ERR_INVALID_PARAM;
    }

    // for stdout
    if (logType == 0) {
        return HLOG_OK;
    }

    // for file
    if (path == nullptr) {
        gLastErrorMessage = "Invalid path, which is empty";
        return HLOG_ERR_INVALID_PARAM;
    } else if (rotationFileSize > ROTATION_FILE_SIZE_MAX || rotationFileSize < ROTATION_FILE_SIZE_MIN) {
        gLastErrorMessage = "Invalid max file size, which should be between 2MB to 100MB";
        return HLOG_ERR_INVALID_PARAM;
    } else if (rotationFileCount > ROTATION_FILE_COUNT_MAX || rotationFileCount < 1) {
        gLastErrorMessage = "Invalid max file count, which should be less than 50";
        return HLOG_ERR_INVALID_PARAM;
    }
    return HLOG_OK;
}

int Hlog::CreateInstance(int logType, int minLogLevel, const char *path, int rotationFileSize, int rotationFileCount)
{
    if (gLogger != nullptr) {
        return HLOG_OK;
    }

    int hr = ValidateParams(logType, minLogLevel, path, rotationFileSize, rotationFileCount);
    if (hr != 0) {
        return hr;
    }

    std::string realPath = std::string(path);
    Hlog *tmpLogger = new (std::nothrow) Hlog(logType, minLogLevel, realPath, rotationFileSize, rotationFileCount);
    if (tmpLogger == nullptr) {
        return HLOG_ERR_CREATE_FAILED;
    }
    hr = tmpLogger->Init();
    if (hr != 0) {
        delete tmpLogger;
        tmpLogger = nullptr;
        return hr;
    }

    gLogger = tmpLogger;
    return HLOG_OK;
}

int Hlog::LogMessage(int level, const char *prefix, const char *message)
{
    if (prefix == nullptr || message == nullptr) {
        gLastErrorMessage = "Invalid param, fmt or message is null";
        return HLOG_ERR_INVALID_PARAM;
    } else if (gLogger == nullptr) {
        gLastErrorMessage = "No logger created";
        return HLOG_ERR_NOT_INITIALIZED;
    }

    return gLogger->Log(level, prefix, message);
}

int Hlog::SetInstancePattern(const char *pattern)
{
    if (pattern == nullptr) {
        gLastErrorMessage = "Invalid param, pattern is empty";
        return HLOG_ERR_INVALID_PARAM;
    }

    if (gLogger == nullptr) {
        gLastErrorMessage = "No logger created";
        return HLOG_ERR_NOT_INITIALIZED;
    }

    return gLogger->SetPattern(std::string(pattern));
}

int Hlog::CreateInstanceAudit(const char *path, int rotationFileSize, int rotationFileCount)
{
    if (gAuditLogger != nullptr) {
        return HLOG_OK;
    }

    int hr = ValidateParams(1, static_cast<int>(spdlog::level::info), path, rotationFileSize, rotationFileCount);
    if (hr != 0) {
        return hr;
    }

    std::string realPath = std::string(path);
    Hlog *tmpLogger = new (std::nothrow)
        Hlog(1, static_cast<int>(spdlog::level::info), realPath, rotationFileSize, rotationFileCount);
    if (tmpLogger == nullptr) {
        return HLOG_ERR_CREATE_FAILED;
    }
    hr = tmpLogger->InitAudit();
    if (hr != 0) {
        delete tmpLogger;
        tmpLogger = nullptr;
        return hr;
    }
    gAuditSPDLogger->set_pattern("[%Y-%m-%d %H:%M:%S.%f]%v");
    gAuditSPDLogger->flush();

    gAuditLogger = tmpLogger;
    return HLOG_OK;
}

int Hlog::LogAuditMessage(int level, const char *prefix, const char *message)
{
    (void)level;
    if (prefix == nullptr || message == nullptr) {
        gLastErrorMessage = "Invalid param, fmt or message is null";
        return HLOG_ERR_INVALID_PARAM;
    } else if (gAuditLogger == nullptr) {
        gLastErrorMessage = "No logger created";
        return HLOG_ERR_NOT_INITIALIZED;
    }

    return gAuditLogger->AuditLog(message);
}

void Hlog::DestroyLogInstances() noexcept
{
    if (gLogger != nullptr) {
        gLogger->Flush();
        delete gLogger;
        gLogger = nullptr;
    }

    if (gAuditLogger != nullptr) {
        gAuditLogger->Flush();
        delete gAuditLogger;
        gAuditLogger = nullptr;
    }
}

const std::string Hlog::GetLastErrorMessage()
{
    return gLastErrorMessage;
}

int Hlog::Init()
{
    try {
        // create logger according to the log type
        // stdout mainly use for
        if (this->mLogType == STDOUT_TYPE) {
            gSPDLogger = spdlog::stdout_logger_mt("console");
            gSPDLogger->set_pattern("%Y-%m-%d %H:%M:%S.%f %t %l %v");
        } else if (this->mLogType == FILE_TYPE) {
            std::string logName = "log:" + this->mFilePath;
            gSPDLogger = spdlog::rotating_logger_mt(logName.c_str(), this->mFilePath, this->mRotationFileSize,
                this->mRotationFileCount);
            gSPDLogger->set_pattern("%v");
            gSPDLogger->info("", "");
            gSPDLogger->set_pattern("%Y-%m-%d %H:%M:%S.%f %t %v");
            gSPDLogger->info("Log started at [{}] level",
                spdlog::level::to_string_view(static_cast<spdlog::level::level_enum>(this->mMinLogLevel)).data());
            gSPDLogger->info("Log default format: yyyy-mm-dd hh:mm:ss.uuuuuu threadid loglevel msg");
            gSPDLogger->set_pattern("%Y-%m-%d %H:%M:%S.%f %t %l %v");
            spdlog::flush_every(std::chrono::seconds(1));
        } else if (this->mLogType == STDERR_TYPE) {
            gSPDLogger = spdlog::stderr_logger_mt("console");
            gSPDLogger->set_pattern("%C/%m/%d %H:%M:%S.%f %t %l %v");
        }
        gSPDLogger->set_level(static_cast<spdlog::level::level_enum>(this->mMinLogLevel));
        gSPDLogger->flush_on(spdlog::level::err);

        if (this->mMinLogLevel < static_cast<int>(spdlog::level::info)) {
            this->mDebugEnabled = true;
        }
    } catch (const spdlog::spdlog_ex &ex) {
        gLastErrorMessage = "Failed to create log: ";
        gLastErrorMessage += ex.what();
        return HLOG_ERR_CREATE_FAILED;
    }
    return HLOG_OK;
}

int Hlog::InitAudit()
{
    try {
        // create logger according to the log type
        // stdout mainly use for
        if (this->mLogType == STDOUT_TYPE) {
            gAuditSPDLogger = spdlog::stdout_logger_mt("console");
            gAuditSPDLogger->set_pattern("%Y-%m-%d %H:%M:%S.%f %t %l %v");
        } else if (this->mLogType == FILE_TYPE) {
            std::string logName = "log:" + this->mFilePath;
            gAuditSPDLogger = spdlog::rotating_logger_mt(logName.c_str(), this->mFilePath, this->mRotationFileSize,
                this->mRotationFileCount);
            gAuditSPDLogger->set_pattern("%v");
            gAuditSPDLogger->info("", "");
            gAuditSPDLogger->set_pattern("%Y-%m-%d %H:%M:%S.%f %t %v");
            gAuditSPDLogger->info("Log started at [{}] level",
                spdlog::level::to_string_view(static_cast<spdlog::level::level_enum>(this->mMinLogLevel)).data());
            gAuditSPDLogger->info("Log default format: yyyy-mm-dd hh:mm:ss.uuuuuu threadid loglevel msg");
            gAuditSPDLogger->set_pattern("%Y-%m-%d %H:%M:%S.%f %t %l %v");
            spdlog::flush_every(std::chrono::seconds(1));
        } else if (this->mLogType == STDERR_TYPE) {
            gAuditSPDLogger = spdlog::stderr_logger_mt("console");
            gAuditSPDLogger->set_pattern("%C/%m/%d %H:%M:%S.%f %t %l %v");
        }
        gAuditSPDLogger->set_level(static_cast<spdlog::level::level_enum>(this->mMinLogLevel));
        gAuditSPDLogger->flush_on(spdlog::level::err);

        if (this->mMinLogLevel < static_cast<int>(spdlog::level::info)) {
            this->mDebugEnabled = true;
        }
    } catch (const spdlog::spdlog_ex &ex) {
        gLastErrorMessage = "Failed to create log: ";
        gLastErrorMessage += ex.what();
        return HLOG_ERR_CREATE_FAILED;
    }
    return HLOG_OK;
}

int Hlog::SetPattern(const std::string &pattern)
{
    if (gSPDLogger.get() == nullptr) {
        gLastErrorMessage = "spdlog logger is not created yet";
        return HLOG_ERR_NOT_INITIALIZED;
    }

    try {
        gSPDLogger->set_pattern(pattern);
    } catch (const spdlog::spdlog_ex &ex) {
        gLastErrorMessage = "Failed to set pattern to spd logger: ";
        gLastErrorMessage += ex.what();
        return HLOG_ERR_SET_OPTION_FAILED;
    }
    return HLOG_OK;
}

int Hlog::Log(int level, const std::string &message) const
{
    if (level < 0 || level > 5) { // 5
        gLastErrorMessage = "Invalid log level, which should be 0~5";
        return HLOG_ERR_INVALID_LEVEL;
    }
    gSPDLogger->log(static_cast<spdlog::level::level_enum>(level), "{}", message);
    return HLOG_OK;
}

int Hlog::Log(int level, const char *prefix, const char *message) const
{
    if (level < 0 || level > 5) { // 5
        gLastErrorMessage = "Invalid log level, which should be 0~5";
        return HLOG_ERR_INVALID_LEVEL;
    }
    gSPDLogger->log(static_cast<spdlog::level::level_enum>(level), "{}] {}", prefix, message);
    return HLOG_OK;
}

int Hlog::DebugLog(const std::string &message) const
{
    gSPDLogger->debug("{}", message);
    return HLOG_OK;
}

int Hlog::AuditLog(const char *message) const
{
    gAuditSPDLogger->info("{}", message);
    return HLOG_OK;
}

void Hlog::Flush() const
{
    gSPDLogger->flush();
}

void Hlog::FlushAudit() const
{
    gAuditSPDLogger->flush();
}
}
}