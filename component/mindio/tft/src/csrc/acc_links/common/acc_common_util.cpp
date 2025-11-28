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
#include <unistd.h>

#include "acc_includes.h"
#include "acc_file_validator.h"
#include "acc_common_util.h"

namespace ock {
namespace acc {
bool AccCommonUtil::IsValidIPv4(const std::string &ip)
{
    constexpr size_t maxIpv4Len = 15;
    if (ip.size() > maxIpv4Len) {
        return false;
    }
    std::regex ipv4Regex("^(?:(?:25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)($|(?!\\.$)\\.)){4}$");
    return std::regex_match(ip, ipv4Regex);
}

Result AccCommonUtil::SslShutdownHelper(SSL *ssl)
{
    if (!ssl) {
        LOG_ERROR("ssl ptr is nullptr");
        return ACC_ERROR;
    }

    const int sslShutdownTimes = 5;
    const int sslRetryInterval = 1;  // s
    int ret = OpenSslApiWrapper::SslShutdown(ssl);
    if (ret == 1) {
        return ACC_OK;
    } else if (ret < 0) {
        ret = OpenSslApiWrapper::SslGetError(ssl, ret);
        LOG_ERROR("ssl shutdown failed!, error code is:" << ret);
        return ACC_ERROR;
    } else if (ret != 0) {
        LOG_ERROR("unknown ssl shutdown ret val!");
        return ACC_ERROR;
    }

    for (int i = UNO_1; i <= sslShutdownTimes; ++i) {
        sleep(sslRetryInterval);
        LOG_INFO("ssl showdown retry times:" << i);
        ret = OpenSslApiWrapper::SslShutdown(ssl);
        if (ret == 1) {
            return ACC_OK;
        } else if (ret < 0) {
            LOG_ERROR("ssl shutdown failed!, error code is:" << OpenSslApiWrapper::SslGetError(ssl, ret));
            return ACC_ERROR;
        } else if (ret != 0) {
            LOG_ERROR("unknown ssl shutdown ret val!");
            return ACC_ERROR;
        }
    }
    return ACC_ERROR;
}

uint32_t AccCommonUtil::GetEnvValue2Uint32(const char *envName)
{
    // 0 should be illegal for this env variable
    constexpr uint32_t maxUint32Len = 35;
    const char *tmpEnvValue = std::getenv(envName);
    if (tmpEnvValue != nullptr && strlen(tmpEnvValue) <= maxUint32Len && IsAllDigits(tmpEnvValue)) {
        uint32_t envValue = String2Uint(tmpEnvValue);
        return envValue;
    }
    return 0;
}

uint32_t AccCommonUtil::String2Uint(const char *str)
{ // avoid throwing ex during converting to std::string
    try {
        auto num = std::stoul(str);
        if (num > UINT32_MAX) {
            return 0;
        }
        return static_cast<uint32_t>(num);
    } catch (...) {
        return 0;
    }
}

bool AccCommonUtil::IsAllDigits(const std::string &str)
{
    if (str.empty()) {
        return false;
    }
    return std::all_of(str.begin(), str.end(), [](unsigned char ch) {
        return std::isdigit(ch);
    });
}

std::string AccCommonUtil::TrimString(const std::string &input)
{
    if (input.empty()) {
        return "";
    }
    auto start = input.begin();
    while (start != input.end() && std::isspace(*start)) {
        start++;
    }

    auto end = input.end();
    do {
        end--;
    } while (std::distance(start, end) > 0 && std::isspace(*end));

    return std::string(start, end + 1);
}

#define CHECK_FILE_PATH_TLS(key, path)                                                         \
    do {                                                                                       \
        if (FileValidator::IsSymlink(path) || !FileValidator::Realpath(path)           \
            || !FileValidator::IsFile(path) || !FileValidator::CheckFileSize(path)) {  \
            LOG_ERROR("TLS " #key " check failed");                                            \
            return ACC_ERROR;                                                                  \
        }                                                                                      \
    } while (0)

#define CHECK_FILE_PATH(key, required)                                           \
    do {                                                                         \
        if (!tlsOption.key.empty()) {                                            \
            std::string path = tlsOption.tlsTopPath + "/" + tlsOption.key;       \
            CHECK_FILE_PATH_TLS(key, path);                                      \
        } else if (required) {                                                   \
            LOG_ERROR("TLS check failed, " #key " is required");                 \
            return ACC_ERROR;                                                    \
        }                                                                        \
    } while (0)

#define CHECK_DIR_PATH_TLS(key, path)                                                    \
    do {                                                                                 \
        if (FileValidator::IsSymlink(path) || !FileValidator::Realpath(path)     \
            || !FileValidator::IsDir(path)) {                                        \
            LOG_ERROR("TLS " #key " check failed");                                      \
            return ACC_ERROR;                                                            \
        }                                                                                \
    } while (0)

#define CHECK_DIR_PATH(key, required)                                                                               \
    do {                                                                                                            \
        if (!tlsOption.key.empty()) {                                                                               \
            std::string path = (#key == "tlsTopPath") ? tlsOption.key : tlsOption.tlsTopPath + "/" + tlsOption.key; \
            CHECK_DIR_PATH_TLS(key, path);                                                                          \
        } else if (required) {                                                                                      \
            LOG_ERROR("TLS check failed, " #key " is required");                                                    \
            return ACC_ERROR;                                                                                       \
        }                                                                                                           \
    } while (0)

#define CHECK_FILE_SET_TLS(key, topPath)                                                 \
    do {                                                                                 \
        for (const std::string &file : tlsOption.key) {                                  \
            std::string filePath = (topPath) + "/" + (file);                             \
            CHECK_FILE_PATH_TLS(key, filePath);                                          \
        }                                                                                \
    } while (0)

#define CHECK_FILE_SET(key, topPath, required)                                               \
    do {                                                                                     \
        if (!tlsOption.key.empty()) {                                                        \
            CHECK_FILE_SET_TLS(key, topPath);                                                \
        } else if (required) {                                                               \
            LOG_ERROR("TLS check failed, " #key " is required");                             \
            return ACC_ERROR;                                                                \
        }                                                                                    \
    } while (0)

Result AccCommonUtil::CheckTlsOptions(const AccTlsOption &tlsOption)
{
    if (!tlsOption.enableTls) {
        return ACC_OK;
    }
    CHECK_DIR_PATH(tlsTopPath, false);
    CHECK_DIR_PATH(tlsCaPath, true);
    CHECK_DIR_PATH(tlsCrlPath, false);
    CHECK_FILE_PATH(tlsCert, true);
    CHECK_FILE_SET(tlsCaFile, tlsOption.tlsTopPath + "/" + tlsOption.tlsCaPath, true);
    CHECK_FILE_SET(tlsCrlFile, tlsOption.tlsTopPath + "/" + tlsOption.tlsCrlPath, false);
    return ACC_OK;
}
}  // namespace acc
}  // namespace ock
