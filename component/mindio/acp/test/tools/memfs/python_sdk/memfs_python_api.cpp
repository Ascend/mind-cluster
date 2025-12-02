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
#include <cstdio>
#include <cstdint>
#include <cerrno>
#include <map>
#include <Python.h>
#include "hcom_log.h"
#include "memfs_out_logger.h"
#include "common_includes.h"
#include "memfs_sdk_api.h"

namespace ock {
namespace memfs {

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

static PyObject *SetLogger(PyObject *self, PyObject *args)
{
    s_loggerLevels[0] = "DEBUG";
    s_loggerLevels[1] = "INFO";
    s_loggerLevels[2] = "WARN";
    s_loggerLevels[3] = "ERROR";

    const char *path = nullptr;
    if (!PyArg_ParseTuple(args, "s", &path)) {
        LOG_ERROR("MemFs get logger path failed.");
        return Py_None;
    }
    auto file = fopen(path, "a");
    if (file == nullptr) {
        LOG_ERROR("open logger file failed.");
        return Py_None;
    }

    g_sLogger = file;
    auto serviceLog = ock::hcom::UBSHcomNetOutLogger::Instance();
    serviceLog->SetExternalLogFunction(ClientLog);
    ock::memfs::OutLogger::Instance()->SetExternalLogFunction(ClientLog);
    return Py_None;
}

static PyObject *CreateFile(PyObject *self, PyObject *args)
{
    const char *path = nullptr;
    if (!PyArg_ParseTuple(args, "s", &path)) {
        LOG_ERROR("MemFs get open args failed.");
        return Py_BuildValue("i", -EINVAL);
    }

    auto fd = MemFsOpenFile(path, O_CREAT | O_WRONLY | O_TRUNC);
    if (fd < 0) {
        LOG_ERROR("MemFs open file for write failed: " << errno << " : " << strerror(errno));
        return Py_BuildValue("i", fd);
    }

    return Py_BuildValue("i", fd);
}

static PyObject *LinkFile(PyObject *self, PyObject *args)
{
    const char *source = nullptr;
    const char *target = nullptr;
    if (!PyArg_ParseTuple(args, "ss", &source, &target)) {
        LOG_ERROR("MemFs get link args failed.");
        errno = EINVAL;
        return Py_BuildValue("i", -1);
    }

    auto result = MemFsLinkFile(source, target);
    if (result != 0) {
        LOG_ERROR("MemFs link file failed: " << result << " : " << strerror(errno));
        return Py_BuildValue("i", -1);
    }

    return Py_BuildValue("i", 0);
}

static PyObject *OpenFile(PyObject *self, PyObject *args)
{
    const char *path = nullptr;
    if (!PyArg_ParseTuple(args, "s", &path)) {
        LOG_ERROR("MemFs get open args failed.");
        return Py_BuildValue("i", -EINVAL);
    }

    auto fd = MemFsOpenFile(path, O_RDONLY);
    if (fd < 0) {
        LOG_ERROR("MemFs open file for read failed: " << errno << " : " << strerror(errno));
        return Py_BuildValue("i", fd);
    }

    return Py_BuildValue("i", fd);
}

static PyObject *WriteFile(PyObject *self, PyObject *args)
{
    int fd = -1;
    PyBytesObject *data = nullptr;
    if (!PyArg_ParseTuple(args, "iS", &fd, &data)) {
        LOG_ERROR("MemFs get write args failed.");
        return Py_BuildValue("i", -EINVAL);
    }

    auto size = PyBytes_Size(&data->ob_base.ob_base);
    auto ret = MemFsWrite(fd, (uintptr_t)data->ob_sval, size);
    if (ret < 0) {
        LOG_ERROR("MemFs write file failed: " << errno << " : " << strerror(errno));
        return Py_BuildValue("l", ret);
    }

    return Py_BuildValue("l", ret);
}

static PyObject *ReadFile(PyObject *self, PyObject *args)
{
    int fd = -1;
    int64_t offset = 0;
    int64_t length = 0;
    if (!PyArg_ParseTuple(args, "iLL", &fd, &offset, &length)) {
        LOG_ERROR("MemFs get read args failed.");
        return Py_BuildValue("(iS)", -1, "");
    }

    auto data = new (std::nothrow) char[length];
    if (data == nullptr) {
        LOG_ERROR("MemFs allocate memory failed.");
        return Py_BuildValue("(iS)", -1, "");
    }

    auto fileSize = MemFsGetSize(fd);
    if (fileSize < 0) {
        LOG_ERROR("get file size failed for fd(" << fd << ") return : " << fileSize);
        return Py_BuildValue("(iS)", -1, "");
    }

    if (offset >= fileSize) {
        return Py_BuildValue("(iS)", 0, "");
    }

    length = std::min(length, fileSize - offset);
    auto ret = MemFsRead(fd, (uintptr_t)data, offset, length);
    if (ret < 0) {
        LOG_ERROR("MemFs read file failed: " << errno << " : " << strerror(errno));
        delete[] data;
        data = nullptr;
        return Py_BuildValue("(iS)", -1, "");
    }

    auto res = PyBytes_FromStringAndSize(data, length);
    delete[] data;
    data = nullptr;
    return Py_BuildValue("(iS)", -1, res);
}

static PyObject *CloseFile(PyObject *self, PyObject *args)
{
    int fd = -1;
    if (!PyArg_ParseTuple(args, "i", &fd)) {
        LOG_ERROR("MemFs get close args failed.");
        return Py_BuildValue("i", -EINVAL);
    }

    auto ret = MemFsClose(fd);
    if (ret < 0) {
        LOG_ERROR("MemFs close file failed: " << errno << " : " << strerror(errno));
        return Py_BuildValue("i", ret);
    }

    return Py_BuildValue("i", 0);
}

static PyObject *RenameFile(PyObject *self, PyObject *args)
{
    const char *oldPath = nullptr;
    const char *newPath = nullptr;
    const char *flags = nullptr;
    static std::map<std::string, uint32_t> flagsConvert = { { "zero", 0 },
                                                            { "no_replace", RENAME_FLAG_NOREPLACE },
                                                            { "exchange", RENAME_FLAG_EXCHANGE },
                                                            { "force", RENAME_FLAG_FORCE } };
    if (!PyArg_ParseTuple(args, "sss", &oldPath, &newPath, &flags)) {
        LOG_ERROR("MemFs get rename args failed.");
        return Py_BuildValue("i", -EINVAL);
    }

    auto pos = flagsConvert.find(flags);
    if (pos == flagsConvert.end()) {
        LOG_ERROR("Rename flags(" << flags << ") invalid.");
        return Py_BuildValue("i", EINVAL);
    }

    auto ret = MemFsRenameFile(oldPath, newPath, pos->second);
    if (ret < 0) {
        LOG_ERROR("MemFs rename file failed: " << errno << " : " << strerror(errno));
        return Py_BuildValue("i", -errno);
    }

    return Py_BuildValue("i", 0);
}

static PyObject *WatchDirectory(PyObject *self, PyObject *args)
{
    const char *path = nullptr;
    int count = 0;
    int timeout = 0;
    PyObject *func;
    if (!PyArg_ParseTuple(args, "siiO", &path, &count, &timeout, &func)) {
        LOG_ERROR("MemFs get open args failed.");
        return Py_BuildValue("i", -EINVAL);
    }

    if (!PyCallable_Check(func)) {
        LOG_ERROR("MemFs input function cannot be called.");
        return Py_BuildValue("i", -EINVAL);
    }

    uint64_t eventId;
    std::list<std::pair<std::string, uint64_t>> info;
    info.emplace_back(path, static_cast<uint64_t>(count));

    Py_IncRef(func);
    auto ret = MemfsRegisterWatchDir(
        info, timeout,
        [func](uint64_t id, int result) {
            LOG_INFO("event id: " << id << " result :" << result);
            auto gilState = PyGILState_Ensure();
            PyObject_CallObject(func, Py_BuildValue("(i)", result));
            Py_DecRef(func);
            PyGILState_Release(gilState);
        },
        eventId);
    if (ret != 0) {
        Py_DecRef(func);
        return Py_BuildValue("i", ret);
    }

    LOG_INFO("register for path: " << path << ", id = " << eventId);
    return Py_BuildValue("i", 0);
}

static PyObject *Initialize(PyObject *self, PyObject *args)
{
    Py_InitializeEx(1);
    PyEval_InitThreads();
    auto ret = MemFsClientInitialize();
    if (ret != 0) {
        return Py_BuildValue("i", ret);
    }

    return Py_BuildValue("i", 0);
}

static PyObject *Destroy(PyObject *self, PyObject *args)
{
    MemFsClientUnInitialize();
    return Py_None;
}

static PyObject *Access(PyObject *self, PyObject *args)
{
    long position;
    if (!PyArg_ParseTuple(args, "l", &position)) {
        LOG_ERROR("MemFs get access args failed.");
        return Py_BuildValue("i", -EINVAL);
    }

    auto positionAsUint64 = static_cast<uint64_t>(position);
    auto address = &positionAsUint64;
    LOG_INFO("before to access to address: " << address);

    *address = 0;
    return Py_BuildValue("i", 0);
}

static PyMethodDef memfs_client_methods[] = {
    {
        "create", CreateFile, METH_VARARGS, "create new file to write: fd:int = create(path:str)"
    },
    {
        "open", OpenFile, METH_VARARGS, "open exist file to read: fd:int = open(path:str)"
    },
    {
        "link", LinkFile, METH_VARARGS,
        "creates a new link (also known as a hard link) to an existing file: ret:int = link(oldpath:str, newpath:str)"
    },
    {
        "write", WriteFile, METH_VARARGS, "write a file: ret:int = write(fd:int, data:bytes)"
    },
    {
        "read", ReadFile, METH_VARARGS, "read a file: (res:int, data: bytes) = read(fd:int, off:long, len:long)"
    },
    {
        "close", CloseFile, METH_VARARGS, "close a file: ret:int = close(fd:int)"
    },
    {
        "init", Initialize, METH_VARARGS, "initialize"
    },
    {
        "destroy", Destroy, METH_VARARGS, "destroy"
    },
    {
        "watch", WatchDirectory, METH_VARARGS, "watch : ret:int = watch(path:str, file_count:int, timeout:int, func)"
    },
    {
        "rename", RenameFile, METH_VARARGS, "rename a file: ret:int = rename(oldpath: str, newpath: str)"
    },
    {
        "logger", SetLogger, METH_VARARGS, "logger(path: str)"
    },
    {
        "access", Access, METH_VARARGS, "access(address: long)"
    },
    {
        nullptr, nullptr, 0, nullptr
    }
};

static struct PyModuleDef memfs_definition = { PyModuleDef_HEAD_INIT, "memfs_client", "Python/C API for memfs", -1,
                                               memfs_client_methods };

PyMODINIT_FUNC PyInit_memfs(void)
{
    Py_Initialize();
    return PyModule_Create(&memfs_definition);
}

}
}