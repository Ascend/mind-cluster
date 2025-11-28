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
#include <fcntl.h>
#include <pybind11/pybind11.h>
#include <pybind11/stl.h>
#include <pybind11/functional.h>

#include <cerrno>
#include <map>
#include <string>
#include <iostream>

#include "hcom_log.h"
#include "common_includes.h"
#include "memfs_sdk_api.h"

namespace py = pybind11;

using namespace ock::memfs;

static FILE *g_sLogger = stderr;
static std::map<int, std::string> s_loggerLevels;
static constexpr uint32_t RENAME_FLAG_NOREPLACE = (1U << 0);
static constexpr uint32_t RENAME_FLAG_EXCHANGE = (1U << 1);
static constexpr uint32_t RENAME_FLAG_FORCE = (1U << 10);

static void ClientLog(int level, const char *msg)
{
    auto pos = s_loggerLevels.find(level);
    if (pos == s_loggerLevels.end()) {
        return;
    }

    (void)fprintf(g_sLogger, "[%s] %s\n", pos->second.c_str(), msg);
    (void)fflush(g_sLogger);
}

static int Initialize(bool tls, const std::string &cert, const std::string &ca, const std::string &pk,
    const std::string &pw, const std::string &pmt, const std::string &crl)
{
    if (!tls) {
        return MemFsClientInitialize();
    }

    ClientInitParam param;
    param.ipcTlsEnabled = true;
    param.ipcTlsCertPath = cert;
    param.ipcTlsCrlPath = crl;
    param.ipcTlsCaPath = ca;
    param.ipcTlsPriKeyPath = pk;
    param.ipcTlsPasswordPath = pw;
    param.pmtPath = pmt;
    return MemFsClientInitialize(param);
}

static void Destroy()
{
    MemFsClientUnInitialize();
}

static void SetLogger(const std::string &path)
{
    s_loggerLevels[0] = "DEBUG";
    s_loggerLevels[1] = "INFO";
    s_loggerLevels[2] = "WARN";
    s_loggerLevels[3] = "ERROR";

    auto file = fopen(path.c_str(), "a");
    if (file == nullptr) {
        LOG_ERROR("open logger file failed.");
        return;
    }

    g_sLogger = file;
    auto serviceLog = ock::hcom::UBSHcomNetOutLogger::Instance();
    serviceLog->SetExternalLogFunction(ClientLog);
    ock::memfs::OutLogger::Instance()->SetExternalLogFunction(ClientLog);
}

static int CreateFile(const std::string &path)
{
    auto fd = MemFsOpenFile(path.c_str(), O_CREAT | O_WRONLY | O_TRUNC);
    if (fd < 0) {
        LOG_ERROR("MemFs create file for write failed: " << errno << " : " << strerror(errno));
        return -1;
    }

    return fd;
}

static int OpenFile(const std::string &path)
{
    auto fd = MemFsOpenFile(path.c_str(), O_RDONLY);
    if (fd < 0) {
        LOG_ERROR("MemFs open file for read failed: " << errno << " : " << strerror(errno));
        return -1;
    }

    return fd;
}

static ssize_t WriteFile(int fd, const py::bytes &data)
{
    char *buffer;
    ssize_t length;
    if (PYBIND11_BYTES_AS_STRING_AND_SIZE(data.ptr(), &buffer, &length)) {
        errno = EINVAL;
        return -1L;
    }

    auto ret = MemFsWrite(fd, reinterpret_cast<uintptr_t>(buffer), static_cast<uint64_t>(length));
    if (ret < 0) {
        LOG_ERROR("MemFs write file failed: " << errno << " : " << strerror(errno));
        return -1;
    }

    return ret;
}

static ssize_t WriteFileV(int fd, const std::vector<py::bytes> &buffers)
{
    std::vector<Buffer> realBuffers;
    realBuffers.reserve(buffers.size());
    for (auto &buf : buffers) {
        char *buffer;
        ssize_t length;
        if (PYBIND11_BYTES_AS_STRING_AND_SIZE(buf.ptr(), &buffer, &length)) {
            errno = EINVAL;
            return -1L;
        }

        realBuffers.emplace_back(buffer, static_cast<uint64_t>(length));
    }

    auto ret = MemFsWriteV(fd, realBuffers);
    if (ret < 0) {
        LOG_ERROR("MemFs write file failed: " << errno << " : " << strerror(errno));
        return -1;
    }

    return ret;
}

static std::pair<int, py::bytes> ReadFile(int fd, uint64_t offset, uint64_t length)
{
    auto data = new (std::nothrow) char[length];
    if (data == nullptr) {
        LOG_ERROR("MemFs allocate memory failed.");
        return std::make_pair(-1, py::bytes());
    }

    auto fileSize = MemFsGetSize(fd);
    if (fileSize < 0) {
        LOG_ERROR("get file size failed for fd(" << fd << ") return : " << fileSize);
        return std::make_pair(-1, py::bytes());
    }

    if (offset >= fileSize) {
        return std::make_pair(0, py::bytes());
    }

    auto readSize = std::min(length, fileSize - offset);
    auto ret = MemFsRead(fd, (uintptr_t)data, offset, readSize);
    if (ret < 0) {
        LOG_ERROR("MemFs read file failed: " << errno << " : " << strerror(errno));
        return std::make_pair(-1, py::bytes());
    }

    auto result = std::string(data, readSize);
    delete[] data;
    data = nullptr;
    return std::make_pair(0, py::bytes(result));
}

static int CloseFile(int fd)
{
    auto ret = MemFsClose(fd);
    if (ret < 0) {
        LOG_ERROR("MemFs close file failed: " << errno << " : " << strerror(errno));
        return ret;
    }

    return 0;
}

static int RenameFile(const std::string &oldPath, const std::string &newPath, uint32_t flags)
{
    auto ret = MemFsRenameFile(oldPath.c_str(), newPath.c_str(), flags);
    if (ret < 0) {
        LOG_ERROR("MemFs rename file failed: " << errno << " : " << strerror(errno));
        return ret;
    }

    return 0;
}

static int WatchDirectories(const std::map<std::string, uint64_t> &info, uint64_t timeout,
    const std::function<void(int)> &callback)
{
    uint64_t eventId;
    std::list<std::pair<std::string, uint64_t>> inputInfo;
    for (const auto &it : info) {
        inputInfo.emplace_back(it.first, it.second);
    }
    auto ret = MemfsRegisterWatchDir(
        inputInfo, timeout,
        [callback](uint64_t id, int result) {
            LOG_INFO("event id: " << id << " result :" << result);
            callback(result);
        },
        eventId);
    if (ret != 0) {
        return -1;
    }

    return 0;
}

static int WatchDirectory(const std::string &path, int count, uint64_t timeout,
                          const std::function<void(int)> &callback)
{
    uint64_t eventId;
    std::list<std::pair<std::string, uint64_t>> inputInfo;
    inputInfo.emplace_back(path, count);
    auto ret = MemfsRegisterWatchDir(
        inputInfo,
        timeout,
        [callback](uint64_t id, int result) {
            LOG_INFO("event id: " << id << " result :" << result);
            callback(result);
        },
        eventId);
    if (ret != 0) {
        return -1;
    }

    return 0;
}

PYBIND11_MODULE(memfs, m)
{
    m.doc() = "memfs python API";
    m.def("logger", &SetLogger, py::arg("path"), "set logger path before initialize instead of standard output.");
    m.def("init", &Initialize, "initialize client for memfs.", py::arg("tls") = false, py::arg("cert") = "",
        py::arg("ca") = "", py::arg("pk") = "", py::arg("pw") = "", py::arg("pmt") = "", py::arg("crl") = "");
    m.def("destroy", &Destroy, "destroy client for memfs.");
    m.def("create", &CreateFile, "create new file for write on memfs.", py::arg("path"));
    m.def("open", &OpenFile, "open exist file for read on memfs.", py::arg("path"));
    m.def("write", &WriteFile, "write file on memfs.", py::arg("fd"), py::arg("data"));
    m.def("writev", &WriteFileV, "write file on memfs.", py::arg("fd"), py::arg("data"));
    m.def("read", &ReadFile, "read file on memfs.", py::arg("fd"), py::arg("off"), py::arg("len"));
    m.def("close", &CloseFile, "close file on memfs.", py::arg("fd"));
    m.def("watch", &WatchDirectories, "watch files for backup finished.", py::arg("info"), py::arg("timeout"),
        py::arg("callback"));
    m.def("watch", &WatchDirectory, "watch files for backup finished.", py::arg("path"), py::arg("count"),
        py::arg("timeout"), py::arg("callback"));
    m.def("rename", &RenameFile, "change the name or location of a file.", py::arg("source"), py::arg("target"),
        py::arg("flags") = 0);

    m.attr("RF_NO_REPLACE") = RENAME_FLAG_NOREPLACE;
    m.attr("RF_EXCHANGE") = RENAME_FLAG_EXCHANGE;
    m.attr("RF_FORCE") = RENAME_FLAG_FORCE;
}