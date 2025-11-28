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
#include <cerrno>

#include "dfs_adapter.h"

using namespace ock;
using namespace ock::bg::backup;

bool DfsAdapter::HealthyCheck() noexcept
{
    return true;
}

int DfsAdapter::PutFile(const std::string &path, int flags, const ufs::FileMode &mode,
    ufs::utils::ByteBuffer &dataBuffer) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::PutFile(const std::string &path, const ufs::FileMode &mode, ufs::utils::ByteBuffer &dataBuffer) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::PutFile(const std::string &path, int flags, const ufs::FileMode &mode,
    ufs::InputStream &inputStream) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::PutFile(const std::string &path, const ufs::FileMode &mode, ufs::InputStream &inputStream) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

std::shared_ptr<ufs::OutputStream> DfsAdapter::PutFile(const std::string &path, int flags,
    const ufs::FileMode &mode, const FileRange &range) noexcept
{
    errno = EOPNOTSUPP;
    return nullptr;
}

std::shared_ptr<ufs::OutputStream> DfsAdapter::PutFile(const std::string &path, const ufs::FileMode &mode,
    const FileRange &range) noexcept
{
    errno = EOPNOTSUPP;
    return nullptr;
}

std::shared_ptr<ufs::OutputStream> DfsAdapter::PutFile(const std::string &path, const FileMode &mode) noexcept
{
    errno = EOPNOTSUPP;
    return nullptr;
}

int DfsAdapter::GetFile(const std::string &path, ufs::utils::ByteBuffer &dataBuffer,
    const ufs::FileRange &range) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

std::shared_ptr<ufs::InputStream> DfsAdapter::GetFile(const std::string &path, const ufs::FileRange &range) noexcept
{
    errno = EOPNOTSUPP;
    return nullptr;
}

int DfsAdapter::GetFile(const std::string &path, const ufs::FileRange &range, ufs::OutputStream &outputStream) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::MoveFile(const std::string &source, const std::string &destination) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::CopyFile(const std::string &source, const std::string &destination) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::RemoveFile(const std::string &path) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::CreateDirectory(const std::string &path, const ufs::FileMode &mode) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::RemoveDirectory(const std::string &path) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::ListFiles(const std::string &path, ufs::ListFileResult &result) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::ListFiles(const std::string &path, ufs::ListFileResult &result,
    std::shared_ptr<ufs::ListFilePageMarker> marker) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::GetFileMeta(const std::string &path, ufs::FileMeta &meta) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

int DfsAdapter::SetFileMeta(const std::string &path, std::map<std::string, std::string> &meta) noexcept
{
    errno = EOPNOTSUPP;
    return -1;
}

std::shared_ptr<ufs::FileLock> DfsAdapter::GetFileLock(const std::string &path) noexcept
{
    errno = EOPNOTSUPP;
    return nullptr;
}
