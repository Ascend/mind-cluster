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
#ifndef OCK_TTP_LOGGERS_H
#define OCK_TTP_LOGGERS_H

#include <ctime>
#include <cstring>
#include <iostream>
#include <mutex>
#include <unistd.h>
#include <sstream>
#include <sys/time.h>
#include <sys/syscall.h>

#ifndef COMMON_LIKELY
#define COMMON_LIKELY(x) (__builtin_expect(!!(x), 1) != 0)
#endif

#ifndef COMMON_UNLIKELY
#define COMMON_UNLIKELY(x) (__builtin_expect(!!(x), 0) != 0)
#endif

namespace ock {
namespace ttp {
#ifndef COMMON_OUT_LOGGER
typedef void (*ExternalLog)(int level, const char *msg);
#endif

enum LogLevel : int {
    DEBUG_LEVEL = 0,
    INFO_LEVEL,
    WARN_LEVEL,
    ERROR_LEVEL,
    BUTT_LEVEL
};

class OutLogger {
public:
    static OutLogger *Instance()
    {
        static OutLogger *gLogger = nullptr;
        static std::mutex gMutex;

        if (__builtin_expect(gLogger == nullptr, 0) != 0) {
            gMutex.lock();
            if (gLogger == nullptr) {
                gLogger = new (std::nothrow) OutLogger();

                if (gLogger == nullptr) {
                    printf("Failed to new OutLogger, probably out of memory");
                }
            }
            gMutex.unlock();
        }

        return gLogger;
    }

    inline void SetLogLevel(LogLevel level)
    {
        mLogLevel = level;
    }

    inline void SetAuditLogLevel(LogLevel level)
    {
        mAuditLogLevel = level;
    }

    inline void SetExternalLogFunction(ExternalLog func)
    {
        mLogFunc = func;
    }

    inline void SetExternalAuditLogFunction(ExternalLog func)
    {
        mLogFunc = func;
    }

    inline void Log(int level, const std::ostringstream &oss)
    {
        if (mLogFunc != nullptr) {
            mLogFunc(level, oss.str().c_str());
            return;
        }

        if (level < mLogLevel) {
            return;
        }

        struct timeval tv {};
        char strTime[24];

        gettimeofday(&tv, nullptr);
        time_t timeStamp = tv.tv_sec;
        struct tm localTime {};
        if (strftime(strTime, sizeof strTime, "%Y-%m-%d %H:%M:%S.", localtime_r(&timeStamp, &localTime)) != 0) {
            std::cout << strTime << tv.tv_usec << " " << LogLevelDesc(level) << " " << syscall(SYS_gettid) << " " <<
                oss.str() << std::endl;
        } else {
            std::cout << " Invalid time " << LogLevelDesc(level) << " " << syscall(SYS_gettid) << " " << oss.str() <<
                std::endl;
        }
    }

    inline void AuditLog(int level, const std::ostringstream &oss)
    {
        if (mAuditLogFunc != nullptr) {
            mAuditLogFunc(level, oss.str().c_str());
            return;
        }

        if (level < mAuditLogLevel) {
            return;
        }

        struct timeval tv {};
        char strTime[24];
        gettimeofday(&tv, nullptr);
        time_t timeStamp = tv.tv_sec;
        struct tm localTime {};
        if (strftime(strTime, sizeof strTime, "%Y-%m-%d %H:%M:%S.", localtime_r(&timeStamp, &localTime)) != 0) {
            std::cout << strTime << tv.tv_usec << " " << LogLevelDesc(level) << " " << syscall(SYS_gettid) << " " <<
                oss.str() << std::endl;
        } else {
            std::cout << " Invalid time " << LogLevelDesc(level) << " " << syscall(SYS_gettid) << " " << oss.str() <<
                std::endl;
        }
    }

    OutLogger(const OutLogger &) = delete;
    OutLogger(OutLogger &&) = delete;

    ~OutLogger()
    {
        mLogFunc = nullptr;
        mAuditLogFunc = nullptr;
    }

private:
    OutLogger() = default;

    inline const std::string &LogLevelDesc(int level)
    {
        static std::string invalid = "invalid";
        if (COMMON_UNLIKELY(level < LogLevel::DEBUG_LEVEL || level >= LogLevel::BUTT_LEVEL)) {
            return invalid;
        }
        return mLogLevelDesc[level];
    }

private:
    const std::string mLogLevelDesc[LogLevel::BUTT_LEVEL] = {"debug", "info", "warn", "error"};

    LogLevel mLogLevel = LogLevel::INFO_LEVEL;
    LogLevel mAuditLogLevel = LogLevel::INFO_LEVEL;
    ExternalLog mLogFunc = nullptr;
    ExternalLog mAuditLogFunc = nullptr;
};

// macro for log
#ifndef COMMON_FILENAME_SHORT
#define COMMON_FILENAME_SHORT (strrchr(__FILE__, '/') ? strrchr(__FILE__, '/') + 1 : __FILE__)
#endif
#define COMMON_OUT_LOG(LEVEL, MODULE, ARGS)                                                       \
    do {                                                                                          \
        std::ostringstream oss;                                                                   \
        oss << "[" << #MODULE << " " << COMMON_FILENAME_SHORT << ":" << __LINE__ << "] " << ARGS; \
        ock::ttp::OutLogger::Instance()->Log(LEVEL, oss);                                                   \
    } while (0)

#define COMMON_LOG_DEBUG(MODULE, ARGS) COMMON_OUT_LOG(LogLevel::DEBUG_LEVEL, MODULE, ARGS)
#define COMMON_LOG_INFO(MODULE, ARGS) COMMON_OUT_LOG(LogLevel::INFO_LEVEL, MODULE, ARGS)
#define COMMON_LOG_WARN(MODULE, ARGS) COMMON_OUT_LOG(LogLevel::WARN_LEVEL, MODULE, ARGS)
#define COMMON_LOG_ERROR(MODULE, ARGS) COMMON_OUT_LOG(LogLevel::ERROR_LEVEL, MODULE, ARGS)

#define COMMON_ASSERT_RETURN(MODULE, ARGS, RET)           \
    do {                                                  \
        if (__builtin_expect(!(ARGS), 0) != 0) {          \
            COMMON_LOG_ERROR(MODULE, "Assert " << #ARGS); \
            return RET;                                   \
        }                                                 \
    } while (0)

#define COMMON_ASSERT_RET_VOID(MODULE, ARGS)              \
    do {                                                  \
        if (__builtin_expect(!(ARGS), 0) != 0) {          \
            COMMON_LOG_ERROR(MODULE, "Assert " << #ARGS); \
            return;                                       \
        }                                                 \
    } while (0)

#define COMMON_ASSERT(MODULE, ARGS)                       \
    do {                                                  \
        if (__builtin_expect(!(ARGS), 0) != 0) {          \
            COMMON_LOG_ERROR(MODULE, "Assert " << #ARGS); \
        }                                                 \
    } while (0)
}
}

#endif // OCK_TTP_LOGGERS_H
