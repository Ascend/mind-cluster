/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

#include <sys/utsname.h>
#include <unistd.h>
#include "hlog.h"
#include "service_configure.h"
#include "auditlog_adapt.h"
namespace ock {
namespace common {
const int LOG_BUFF_SIZE = 4096;
// log level
const int LOG_LEVEL = 2;

int32_t GetPid()
{
    pid_t pid = getpid();
    return static_cast<int32_t>(pid);
}


std::string GetHostName()
{
    struct utsname buf;
    if (uname(&buf) != 0) {
        // ensure null termination on failure
        if (buf.nodename == nullptr) {
            return "";
        }
        *buf.nodename = '\0';
    }

    return buf.nodename;
}

void AuditLogTruncate(std::string &message)
{
    using namespace ock::hlog;
    message.resize(LOG_BUFF_SIZE);
    message += "-- the length of audit log exceeds the limit.";

    Hlog::LogAuditMessage(LOG_LEVEL, "", message.c_str());
}

void HLOG_AUDIT(const std::string userId, const std::string &eventType, const std::string &visitResource,
    const std::string &result)
{
    using namespace ock::hlog;
    std::string message = "[" + std::to_string(GetPid()) + "]" + "[" + (GetHostName()) + "]";

    message += "[" + userId + "]";
    message += "[" + eventType + "]";
    if (message.size() > LOG_BUFF_SIZE) {
        AuditLogTruncate(message);
        return;
    }

    message += "[" + visitResource + "]";
    if (message.size() > LOG_BUFF_SIZE) {
        AuditLogTruncate(message);
        return;
    }

    message += "[" + result + "]";
    if (message.size() > LOG_BUFF_SIZE) {
        AuditLogTruncate(message);
        return;
    }
    Hlog::LogAuditMessage(LOG_LEVEL, "", message.c_str());
}
}
}
