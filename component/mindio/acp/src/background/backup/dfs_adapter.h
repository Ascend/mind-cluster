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
#ifndef OCK_DFS_DFS_ADAPTER_H
#define OCK_DFS_DFS_ADAPTER_H

#include "ufs_api.h"
using namespace::ock::ufs;

namespace ock {
namespace bg {
namespace backup {
class DfsAdapter : public ufs::BaseFileService {
public:
    DfsAdapter() = default;
    ~DfsAdapter() override = default;

public:
    bool HealthyCheck() noexcept override;
    int PutFile(const std::string &path, int flags, const ufs::FileMode &mode,
        ufs::utils::ByteBuffer &dataBuffer) noexcept override;
    int PutFile(const std::string &path, const ufs::FileMode &mode,
        ufs::utils::ByteBuffer &dataBuffer) noexcept override;
    int PutFile(const std::string &path, int flags, const ufs::FileMode &mode,
        ufs::InputStream &inputStream) noexcept override;
    int PutFile(const std::string &path, const ufs::FileMode &mode, ufs::InputStream &inputStream) noexcept override;

    std::shared_ptr<ufs::OutputStream> PutFile(const std::string &path, int flags,
        const ufs::FileMode &mode, const FileRange &range) noexcept override;
    std::shared_ptr<ufs::OutputStream> PutFile(const std::string &path, const ufs::FileMode &mode,
        const FileRange &range) noexcept override;
    std::shared_ptr<OutputStream> PutFile(const std::string &path, const FileMode &mode) noexcept override;
    int GetFile(const std::string &path, ufs::utils::ByteBuffer &dataBuffer,
        const ufs::FileRange &range) noexcept override;
    std::shared_ptr<ufs::InputStream> GetFile(const std::string &path, const ufs::FileRange &range) noexcept override;
    int GetFile(const std::string &path, const ufs::FileRange &range,
        ufs::OutputStream &outputStream) noexcept override;

    int MoveFile(const std::string &source, const std::string &destination) noexcept override;
    int CopyFile(const std::string &source, const std::string &destination) noexcept override;

    int RemoveFile(const std::string &path) noexcept override;
    int CreateDirectory(const std::string &path, const ufs::FileMode &mode) noexcept override;
    int RemoveDirectory(const std::string &path) noexcept override;
    int ListFiles(const std::string &path, ufs::ListFileResult &result) noexcept override;
    int ListFiles(const std::string &path, ufs::ListFileResult &result,
        std::shared_ptr<ufs::ListFilePageMarker> marker) noexcept override;

    int GetFileMeta(const std::string &path, ufs::FileMeta &meta) noexcept override;
    int SetFileMeta(const std::string &path, std::map<std::string, std::string> &meta) noexcept override;

    std::shared_ptr<ufs::FileLock> GetFileLock(const std::string &path) noexcept override;
};
}
}
}


#endif // OCK_DFS_DFS_ADAPTER_H
