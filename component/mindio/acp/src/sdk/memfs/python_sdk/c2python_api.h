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
#ifndef OCKIO_C2PYTHON_API_H
#define OCKIO_C2PYTHON_API_H

#include <cstdint>
#include <vector>
#include <string>

#include <Python.h>

#include "memfs_sdk_types.h"
#include "nds_file_driver.h"

namespace ock {
namespace memfs {
namespace sdk {
class CloseableFile {
public:
    explicit CloseableFile(std::string path) noexcept;
    virtual ~CloseableFile() noexcept;
    virtual int Close() noexcept = 0;
    virtual int Initialize() noexcept = 0;

protected:
    const std::string filePath;
    int fd;
};

class Readable : public CloseableFile {
public:
    using CloseableFile::CloseableFile;
    ~Readable() noexcept override;
    virtual ssize_t Read(void *buffer, size_t count, off_t offset) noexcept = 0;
    virtual ssize_t ReadV(std::vector<ock::memfs::ReadBuffer> readVector) noexcept = 0;
    virtual ssize_t Length() const noexcept = 0;

protected:
    ssize_t fileSize{ -1 };
};

class Writeable : public CloseableFile {
public:
    explicit Writeable(std::string path, mode_t mode) noexcept;
    ~Writeable() noexcept override;
    virtual ssize_t Write(const void *buffer, size_t count) noexcept = 0;
    virtual ssize_t WriteV(std::vector<ock::memfs::Buffer> buffers) noexcept = 0;
    virtual int Drop() noexcept = 0;
    virtual int Flush() noexcept = 0;
protected:
    const mode_t fileMode;
};

class ReadableFile : public Readable {
public:
    using Readable::Readable;
    int Initialize() noexcept override;
    ssize_t Read(void *buffer, size_t count, off_t offset) noexcept override;
    ssize_t ReadV(std::vector<ock::memfs::ReadBuffer> readVector) noexcept override;
    int Close() noexcept override;
    ssize_t Length() const noexcept override
    {
        return fileSize;
    }
};

class FReadableFile : public Readable {
public:
    using Readable::Readable;
    int Initialize() noexcept override;
    ssize_t Read(void *buffer, size_t count, off_t offset) noexcept override;
    ssize_t ReadV(std::vector<ock::memfs::ReadBuffer> readVector) noexcept override;
    int Close() noexcept override;
    ssize_t Length() const noexcept override
    {
        return fileSize;
    }

protected:
    FILE* fp = nullptr;
};

class NdsReadableFile : public Readable {
public:
    using Readable::Readable;
    int Initialize() noexcept override;
    ssize_t Read(void *buffer, size_t count, off_t offset) noexcept override;
    ssize_t ReadV(std::vector<ock::memfs::ReadBuffer> readVector) noexcept override;
    int Close() noexcept override;
    ssize_t Length() const noexcept override
    {
        return fileSize;
    }

private:
    static NdsFileDriver *g_ndsDriver;
};

class WriteableFile : public Writeable {
public:
    using Writeable::Writeable;
    int Initialize() noexcept override;
    ssize_t Write(const void *buffer, size_t count) noexcept override;
    ssize_t WriteV(std::vector<ock::memfs::Buffer> buffers) noexcept override;
    int Close() noexcept override;
    int Drop() noexcept override;
    int Flush() noexcept override;
};

class FWriteableFile : public Writeable {
public:
    using Writeable::Writeable;
    int Initialize() noexcept override;
    ssize_t Write(const void *buffer, size_t count) noexcept override;
    ssize_t WriteV(std::vector<ock::memfs::Buffer> buffers) noexcept override;
    int Close() noexcept override;
    int Drop() noexcept override;
    int Flush() noexcept override;
    bool RemoveStgFile() noexcept;

protected:
    FILE* fp = nullptr;
};

int PreloadFile(const std::vector<std::string> &paths) noexcept;
int LinkFile(const std::string &source, const std::string &target) noexcept;
int Initialize(const std::map<std::string, std::string> &serverInfoParam) noexcept;
void UnInitialize() noexcept;
int CheckBackgroundTask() noexcept;
using CheckDirCallback = std::function<void(int32_t, PyObject*)>;

struct CallbackReceiver {
    virtual ~CallbackReceiver() {};
    virtual void CheckDirCallback(int32_t result, PyObject *param) = 0;
};

void InitCallbackHandler(CallbackReceiver *handler);
void HandleCheckDirCallback(int32_t result, PyObject *callbackParam);
int RegisterChecker(const std::unordered_map<std::string, uint64_t> &dirInfo,
                    PyObject *callbackParam, uint64_t timeoutSec);

}
}
}

#endif // OCKIO_C2PYTHON_API_H
