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
#ifndef ACC_LINKS_ACC_DEF_H
#define ACC_LINKS_ACC_DEF_H

#include <atomic>
#include <cstdint>
#include <functional>
#include <set>
#include <string>
#include <sstream>
#include <thread>

#include "common_referable.h"

namespace ock {
namespace acc {
constexpr uint32_t MAX_RECV_BODY_LEN = 10 * 1024 * 1024; /* max receive body len limit */
constexpr uint32_t UNO_1024 = 1024;
constexpr uint32_t UNO_500 = 500;
constexpr uint32_t UNO_256 = 256;
constexpr uint32_t UNO_48 = 48;
constexpr uint32_t UNO_32 = 32;
constexpr uint32_t UNO_16 = 16;
constexpr uint32_t UNO_7 = 7;
constexpr uint32_t UNO_2 = 2;
constexpr uint32_t UNO_1 = 1;

constexpr int16_t MIN_MSG_TYPE = 0;
constexpr int16_t MAX_MSG_TYPE = UNO_48;
constexpr uint32_t ACC_LINK_RECV_TIMEOUT = 1800;

/**
 * @brief Header of connecting to server
 */
struct AccConnReq {
    int16_t magic = 0;
    int16_t version = 0;
    uint64_t rankId = 0;
};

/**
 * @brief Response of connecting
 */
struct AccConnResp {
    int16_t result = 0;
};

/**
 * @brief Result of message sending
 */
enum AccMsgSentResult {
    MSG_SENT = 0,
    MSG_TIMEOUT = 1,
    MSG_LINK_BROKEN = 2,
    /* add error code ahead of this */
    MSG_BUTT,
};

/**
 * @brief Header of message
 */
struct AccMsgHeader {
    int16_t type = 0;     /* data type or opCode */
    int16_t result = 0;   /* result for response */
    uint32_t bodyLen = 0; /* length of data */
    uint32_t seqNo = 0;   /* seqNo */
    uint32_t crc = 0;     /* reserved crc */

    AccMsgHeader() = default;

    AccMsgHeader(int16_t t, uint32_t bLen, uint32_t sno) : type(t), bodyLen(bLen), seqNo(sno)
    {
    }

    AccMsgHeader(int16_t t, int16_t r, uint32_t bLen, uint32_t sno) : type(t), result(r), bodyLen(bLen), seqNo(sno)
    {
    }

    std::string ToString() const
    {
        std::ostringstream oss;
        oss << "type: " << type << ", result: " << result << ", bodyLen: " <<
            bodyLen << ", seqNo: " << seqNo << ", crc: " << crc;
        return oss.str();
    }
};

/**
 * @brief Options of Tcp Server, required when start a tcp server
 */
struct AccTcpServerOptions {
    std::string listenIp;                    /* listen ip */
    uint16_t listenPort = 9966L;             /* listen port */
    uint16_t workerCount = UNO_2;            /* number of worker threads */
    int16_t workerThreadPriority = 0;        /* priority of worker threads */
    int16_t workerPollTimeoutMs = UNO_500;   /* epoll timeout */
    int16_t workerStartCpuId = -1;           /* start cpu id of workers */
    uint16_t linkSendQueueSize = UNO_1024;   /* send queue size */
    uint16_t keepaliveIdleTime = UNO_32;     /* tcp keepalive idle time */
    uint16_t keepaliveProbeTimes = UNO_7;    /* tcp keepalive probe times */
    uint16_t keepaliveProbeInterval = UNO_2; /* tcp keepalive probe interval */
    bool reusePort = true;                   /* reuse listen port */
    bool enableListener = false;             /* start listener or not */
    int16_t magic = 0;                       /* magic number of  */
    int16_t version = 0;                     /* version */
    uint32_t maxWorldSize = UNO_1024;        /* max client number */
};

/**
 * @brief Options of worker
 */
struct AccTcpWorkerOptions {
    uint16_t pollingTimeoutMs = UNO_500; /* poll/epoll timeout */
    uint16_t index = 0;                  /* index of the worker */
    int16_t cpuId = -1;                  /* cpu id for bounding */
    int16_t threadPriority = -1;         /* thread nice */
    std::string name_ = "AccWrk";        /* worker name */

    inline std::string ToString() const
    {
        std::ostringstream oss;
        oss << "name " << name_ << ", index " << index << ", cpu " << cpuId << ", thread-priority " << threadPriority
            << ", poll-timeout-ms " << pollingTimeoutMs;
        return oss.str();
    }

    inline std::string Name() const
    {
        return name_ + ":" + std::to_string(index);
    }
};

/**
 * @brief Callback function of private key password decryptor, see @RegisterDecryptHandler
 *
 * @param cipherText       [in] the encrypted text(private key password)
 * @param plainText        [out] the decrypted text(private key password)
 * @param plaintextLen     [out] the length of plainText
 */
using AccDecryptHandler = std::function<int(const std::string &cipherText, char *plainText, size_t &plainTextLen)>;

/**
 * @brief Tls related option, required if TLS enabled
 */
struct AccTlsOption {
    bool enableTls = false;
    std::string tlsTopPath;           /* root path of certifications */
    std::string tlsCert;              /* certification of server */
    std::string tlsCrlPath;           /* optional, crl file path */
    std::string tlsCaPath;            /* ca file path */
    std::set<std::string> tlsCaFile;  /* paths of ca */
    std::set<std::string> tlsCrlFile; /* path of crl file */
    std::string tlsPk;                /* private key */
    std::string tlsPkPwd;             /* private key password, required, encrypt or plain both allowed */
    std::string packagePath;          /* path of lib file */

    AccTlsOption() : enableTls(false)
    {
    }

    bool ValidateOption(std::string &error) const
    {
        if (!enableTls) {
            return true;
        }

        if (tlsCert.empty()) {
            error = "Failed to validate tlsCert which is empty";
            return false;
        }
        if (tlsCaPath.empty()) {
            error = "Failed to validate tlsCaPath which is empty";
            return false;
        }
        if (tlsCaFile.empty()) {
            error = "Failed to validate tlsCaFile which is empty";
            return false;
        }
        if (tlsPk.empty()) {
            error = "Failed to validate tlsPk which is empty";
            return false;
        }

        return true;
    }
};

/**
 * @brief Result codes
 */
enum AccResult {
    ACC_OK = 0,
    ACC_ERROR = -1,
    ACC_NEW_OBJECT_FAIL = -2,
    ACC_MALLOC_FAIL = -3,
    ACC_INVALID_PARAM = -4,
    ACC_NOT_INITIALIZED = -5,
    ACC_TIMEOUT = -6,
    ACC_CONNECTION_NOT_READY = -7,
    ACC_EPOLL_ERROR = -8,
    ACC_LINK_OPTION_ERROR = -9,
    ACC_QUEUE_IS_FULL = -10,
    ACC_LINK_ERROR = -11,
    ACC_LINK_EAGAIN = -12,
    ACC_LINK_MSG_READY = -13,
    ACC_LINK_MSG_SENT = -14,
    ACC_LINK_MSG_INVALID = -15,
    ACC_LINK_NEED_RECONN = -16,
    ACC_LINK_ADDRESS_IN_USE = -17,
    ACC_RESULT_BUTT = -18,
};

#define ACC_API __attribute__((visibility("default")))
}  // namespace acc
}  // namespace ock

#endif  // ACC_LINKS_ACC_DEF_H
