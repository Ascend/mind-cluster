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

#ifndef HLOG_HLOG_H
#define HLOG_HLOG_H

#include <cstdint>
#include <cstring>
#include <sstream>

#ifndef HLOG_FILENAME
#define HLOG_FILENAME (strrchr(__FILE__, '/') ? strrchr(__FILE__, '/') + 1 : __FILE__)
#endif

#define SPDLOG_LEVEL_TRACE 0
#define SPDLOG_LEVEL_DEBUG 1
#define SPDLOG_LEVEL_INFO 2
#define SPDLOG_LEVEL_WARN 3
#define SPDLOG_LEVEL_ERROR 4
#define SPDLOG_LEVEL_CRITICAL 5
#define SPDLOG_LEVEL_OFF 6

#define HLOG_INTERNAL(level, file, line, msg)                                   \
    do {                                                                        \
        if (ock::hlog::Hlog::gLogger->IsHigherLevel(static_cast<int>(level))) { \
            std::ostringstream oss;                                             \
            oss.str("");                                                        \
            oss.clear();                                                        \
            oss << "[" << (file) << ":" << (line) << "] " << msg;               \
            ock::hlog::Hlog::gLogger->Log(level, oss.str());                    \
        }                                                                       \
    } while (0)

#define HLOG_NO_LOC_INTERNAL(level, msg)                                        \
    do {                                                                        \
        if (ock::hlog::Hlog::gLogger->IsHigherLevel(static_cast<int>(level))) { \
            std::ostringstream oss;                                             \
            oss.str("");                                                        \
            oss.clear();                                                        \
            oss << (msg);                                                       \
            ock::hlog::Hlog::gLogger->Log(level, oss.str());                    \
        }                                                                       \
    } while (0)

#define HLOG_CRITICAL(msg) HLOG_INTERNAL(SPDLOG_LEVEL_CRITICAL, HLOG_FILENAME, __LINE__, msg)
#define HLOG_ERROR(msg) HLOG_INTERNAL(SPDLOG_LEVEL_ERROR, HLOG_FILENAME, __LINE__, msg)
#define HLOG_WARN(msg) HLOG_INTERNAL(SPDLOG_LEVEL_WARN, HLOG_FILENAME, __LINE__, msg)
#define HLOG_INFO(msg) HLOG_INTERNAL(SPDLOG_LEVEL_INFO, HLOG_FILENAME, __LINE__, msg)

#define HLOG_NO_LOC_CRITICAL(msg) HLOG_NO_LOC_INTERNAL(SPDLOG_LEVEL_CRITICAL, msg)
#define HLOG_NO_LOC_ERROR(msg) HLOG_NO_LOC_INTERNAL(SPDLOG_LEVEL_ERROR, msg)
#define HLOG_NO_LOC_WARN(msg) HLOG_NO_LOC_INTERNAL(SPDLOG_LEVEL_WARN, msg)
#define HLOG_NO_LOC_INFO(msg) HLOG_NO_LOC_INTERNAL(SPDLOG_LEVEL_INFO, msg)

#define HLOG_CRITICAL_FL(file, line, msg) HLOG_INTERNAL(SPDLOG_LEVEL_CRITICAL, file, line, msg)
#define HLOG_ERROR_FL(file, line, msg) HLOG_INTERNAL(SPDLOG_LEVEL_ERROR, file, line, msg)
#define HLOG_WARN_FL(file, line, msg) HLOG_INTERNAL(SPDLOG_LEVEL_WARN, file, line, msg)
#define HLOG_INFO_FL(file, line, msg) HLOG_INTERNAL(SPDLOG_LEVEL_INFO, file, line, msg)

#define HLOG_DEBUG_INTERNAL(file, line, msg)                  \
    do {                                                      \
        if (ock::hlog::Hlog::gLogger->IsDebug()) {            \
            std::ostringstream oss;                           \
            oss.str("");                                      \
            oss.clear();                                      \
            oss << "[" << (file) << ":" << (line) << "] " << (msg); \
            ock::hlog::Hlog::gLogger->DebugLog(oss.str());    \
        }                                                     \
    } while (0)

#define HLOG_DEBUG_NO_LOC_INTERNAL(msg)                    \
    do {                                                   \
        if (ock::hlog::Hlog::gLogger->IsDebug()) {         \
            std::ostringstream oss;                        \
            oss.str("");                                   \
            oss.clear();                                   \
            oss << (msg);                                  \
            ock::hlog::Hlog::gLogger->DebugLog(oss.str()); \
        }                                                  \
    } while (0)

#define HLOG_NO_LOC_DEBUG(msg) HLOG_DEBUG_NO_LOC_INTERNAL(msg)
#define HLOG_DEBUG(msg) HLOG_DEBUG_INTERNAL(HLOG_FILENAME, __LINE__, msg)
#define HLOG_DEBUG_FL(file, line, msg) HLOG_DEBUG_INTERNAL(file, line, msg)

void BdmHlogFmtLog(int level, const char *file, int line, const char *fmt, ...);

namespace ock {
namespace hlog {
enum HlogErrNo {
    HLOG_OK = 0,
    HLOG_ERR_CREATE_FAILED = 1,
    HLOG_ERR_INVALID_PARAM = 2,
    HLOG_ERR_NOT_INITIALIZED = 3,
    HLOG_ERR_SET_OPTION_FAILED = 4,
    HLOG_ERR_INVALID_LEVEL = 5,
};
class Hlog {
public:
    Hlog(int logType, int minLogLevel, const std::string &path, int rotationFileSize, int rotationFileCount)
        : mLogType(logType),
          mMinLogLevel(minLogLevel),
          mFilePath(path),
          mRotationFileSize(rotationFileSize),
          mRotationFileCount(rotationFileCount)
    {}

    ~Hlog()
    {
        try {
            if (gLogger != nullptr) {
                gLogger->Flush();
            }
            if (gAuditLogger != nullptr) {
                gAuditLogger->Flush();
            }
        } catch (...) {
            // 捕获异常，处于退出阶段，无需处理
        }
    };

    int Init();
    int InitAudit();
    int Exit()
    {
        DestroyLogInstances();
        return 0;
    }
    int SetPattern(const std::string &pattern);

    inline bool IsDebug() const
    {
        return mDebugEnabled;
    }

    inline bool IsHigherLevel(int nowLevel) const
    {
        return nowLevel >= mMinLogLevel;
    }

    int Log(int level, const std::string &message) const;
    int Log(int level, const char *prefix, const char *message) const;
    int DebugLog(const std::string &message) const;
    int AuditLog(const char *message) const;
    void Flush() const;
    void FlushAudit() const;

    /* *
     * @brief This is not thread safe, need to be called at the first place of main thread
     */
    static int CreateInstance(int logType, int minLogLevel, const char *path, int rotationFileSize,
        int rotationFileCount);
    static int LogMessage(int level, const char *prefix, const char *message);
    static int SetInstancePattern(const char *pattern);
    static Hlog *gLogger;
    static Hlog *gAuditLogger;

    /* *
     * @brief This is not thread safe, need to be called at the first place of main thread
     */
    static int CreateInstanceAudit(const char *path, int rotationFileSize, int rotationFileCount);
    static int LogAuditMessage(int level, const char *prefix, const char *message);
    static void DestroyLogInstances() noexcept;

    static const std::string GetLastErrorMessage();

private:
    static int ValidateParams(int logType, int minLogLevel, const char *path, int rotationFileSize,
        int rotationFileCount);

private:
    // log options
    int mLogType;
    int mMinLogLevel;
    std::string mFilePath;
    int mRotationFileSize;
    int mRotationFileCount;
    bool mDebugEnabled = false;

    static thread_local std::string gLastErrorMessage;
};
}
}

#endif // HLOG_HLOG_H
