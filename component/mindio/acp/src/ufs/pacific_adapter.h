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
#ifndef OCK_DFS_PACIFIC_ADAPTER_H
#define OCK_DFS_PACIFIC_ADAPTER_H

#include <dirent.h>
#include <mutex>
#include <sys/types.h>
#include "ufs_api.h"

namespace ock {
namespace ufs {
struct FileAclEntry {
    uint16_t tag;
    uint16_t perm;
    uint32_t id;
};

struct FileAclHeader {
    uint32_t version;
    FileAclEntry entries[0];
};

class PacificAdapter : public BaseFileService {
public:
    explicit PacificAdapter(std::string root) noexcept;
    ~PacificAdapter() noexcept override;

public:
    bool HealthyCheck() noexcept override;
    int PutFile(const std::string &path, int flags, const FileMode &mode,
        utils::ByteBuffer &dataBuffer) noexcept override;
    int PutFile(const std::string &path, const FileMode &mode, utils::ByteBuffer &dataBuffer) noexcept override;
    int PutFile(const std::string &path, int flags, const FileMode &mode, InputStream &inputStream) noexcept override;
    int PutFile(const std::string &path, const FileMode &mode, InputStream &inputStream) noexcept override;
    std::shared_ptr<OutputStream> PutFile(const std::string &path, int flags, const FileMode &mode,
        const FileRange &range) noexcept override;
    std::shared_ptr<OutputStream> PutFile(const std::string &path, const FileMode &mode,
        const FileRange &range) noexcept override;
    std::shared_ptr<OutputStream> PutFile(const std::string &path, const FileMode &mode) noexcept override;
    int GetFile(const std::string &path, utils::ByteBuffer &dataBuffer, const FileRange &range) noexcept override;
    std::shared_ptr<InputStream> GetFile(const std::string &path, const FileRange &range) noexcept override;
    int GetFile(const std::string &path, const FileRange &range, OutputStream &outputStream) noexcept override;

    int MoveFile(const std::string &source, const std::string &destination) noexcept override;
    int CopyFile(const std::string &source, const std::string &destination) noexcept override;
    int RemoveFile(const std::string &path) noexcept override;

    int CreateDirectory(const std::string &path, const FileMode &mode) noexcept override;
    int RemoveDirectory(const std::string &path) noexcept override;

    int ListFiles(const std::string &path, ListFileResult &result) noexcept override;
    int ListFiles(const std::string &path, ListFileResult &result,
        std::shared_ptr<ListFilePageMarker> marker) noexcept override;
    int GetFileMeta(const std::string &path, FileMeta &meta) noexcept override;
    int SetFileMeta(const std::string &path, std::map<std::string, std::string> &meta) noexcept override;
    std::shared_ptr<FileLock> GetFileLock(const std::string &path) noexcept override;

private:
    std::string GetRealPath(const std::string &path) noexcept;

private:
    static int64_t MovePosForRange(const std::string &path, int fd, const FileRange &range) noexcept;
    static int64_t MovePosForRange2Write(const std::string &path, int fd, const FileRange &range) noexcept;
    static bool IsDefaultOwner(const FileMode &mode) noexcept;
    static int CorrectFileMetadata(int fd, const FileMode &mode) noexcept;
    static int WriteSmallFile(int fd, const std::string &fullPath, const FileMode &mode,
        utils::ByteBuffer &dataBuffer) noexcept;
    static int SupportedFlags(int flags) noexcept;
    static int GetFileAcl(const std::string &path, FileAcl &acl) noexcept;
    static void ParseFileAcl(const FileAclHeader *header, uint32_t entryCount, FileAcl &acl) noexcept;

private:
    std::string mountPath;
};

class FileInputStream : public InputStream {
public:
    explicit FileInputStream(std::string path, int fd, int64_t limit) noexcept;
    ~FileInputStream() noexcept override;

public:
    int Read(uint8_t &byte) noexcept override;
    int64_t Read(uint8_t *buf, uint64_t count) noexcept override;
    int Close() noexcept override;

private:
    std::string filePath;
    int fileDesc;
    int64_t maxLeftBytes;
};

class FileOutputStream : public OutputStream {
public:
    explicit FileOutputStream(std::string path, int fd, int64_t count) noexcept;
    ~FileOutputStream() noexcept override;

public:
    int Write(uint8_t byte) noexcept override;
    int64_t Write(const uint8_t *buf, uint64_t count) noexcept override;
    int Sync() noexcept override;
    int Close() noexcept override;

private:
    int fileDesc;
    std::string filePath;
    int64_t maxLeftBytes;
};

class ReadDirMarker : public ListFilePageMarker {
public:
    ReadDirMarker(std::string p, DIR *d, off_t off) noexcept
        : path{ std::move(p) }, dir{ d }, position{ off }, finished{ false }
    {}

    ~ReadDirMarker() noexcept override
    {
        path.clear();
        dir = nullptr;
    }

    bool Finished() noexcept override
    {
        return finished;
    }

private:
    std::string path;
    DIR *dir;
    off_t position;
    bool finished;
    friend class PacificAdapter;
};

class PacificFileLock : public FileLock {
public:
    PacificFileLock(std::string path, int fd) noexcept;
    ~PacificFileLock() noexcept override;

public:
    int Lock() noexcept override;
    int TryLock() noexcept override;
    int Unlock() noexcept override;

private:
    std::string filePath;
    int fileDesc;
};
}
}

#endif // OCK_DFS_PACIFIC_ADAPTER_H
