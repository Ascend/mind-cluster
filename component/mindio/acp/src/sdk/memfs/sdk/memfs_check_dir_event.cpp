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
#include <sys/types.h>
#include <sys/stat.h>
#include <dirent.h>
#include <unistd.h>

#include <climits>
#include <cstdlib>
#include <cerrno>
#include <cstring>
#include "common_includes.h"
#include "fs_operation.h"
#include "memfs_check_dir_event.h"


namespace ock {
namespace memfs {
const std::string MemfsCheckDirEvent::stageFileSuffix{ ".m.stg" };
const std::unordered_map<unsigned char, std::string> MemfsCheckDirEvent::typeNames = {
    { static_cast<unsigned char>(DT_FIFO), "FIFO" },   { static_cast<unsigned char>(DT_CHR), "CHAR" },
    { static_cast<unsigned char>(DT_DIR), "DIR" },     { static_cast<unsigned char>(DT_BLK), "BLOCK" },
    { static_cast<unsigned char>(DT_REG), "REGULAR" }, { static_cast<unsigned char>(DT_SOCK), "SOCKET" }
};

MemfsCheckDirEvent::MemfsCheckDirEvent(std::string nm, uint64_t timeoutSec, uint64_t intervalMs,
    ock::memfs::DirectoriesInfo info, const std::function<void(uint64_t, int)> &cb) noexcept
    : MemfsEvent(std::move(nm), timeoutSec, intervalMs), directoriesInfo{ std::move(info) }, callback{ cb }
{}

bool MemfsCheckDirEvent::PreCheckEvent() noexcept
{
    if (directoriesInfo.empty()) {
        LOG_ERROR("input directories is empty.");
        return false;
    }

    if (callback == nullptr) {
        LOG_ERROR("input callback is null");
        return false;
    }

    auto clientInstance = MemFsClientOperation::Instance();
    if (clientInstance == nullptr) {
        LOG_ERROR("Failed to get MemFsClientOperation instance");
        return false;
    }
    auto ufsMountPath = clientInstance->GetUfsMountPath();
    for (auto &info : directoriesInfo) {
        if (info.fileCount == 0U) {
            LOG_ERROR("file count for path(" << info.pathName << ") is zero");
            return false;
        }

        if (!PreCheckOneDirectory(ufsMountPath, info.pathName, info.pathName)) {
            return false;
        }
    }

    return true;
}

EventResult MemfsCheckDirEvent::Process() noexcept
{
    for (auto &info : directoriesInfo) {
        auto result = CheckOneDirectory(info);
        if (!result.finished) {
            return result;
        }

        if (result.result != 0) {
            return result;
        }
    }

    return EventResult{ true, 0 };
}

void MemfsCheckDirEvent::Callback(uint64_t eventId, int result) noexcept
{
    callback(eventId, result);
}

bool MemfsCheckDirEvent::PreCheckOneDirectory(const std::string &ufsPath, const std::string &path,
    std::string &resolvedPath) noexcept
{
    char pathBuf[PATH_MAX + 1];
    pathBuf[PATH_MAX] = '\0';
    auto realPath = realpath(path.c_str(), pathBuf);
    if (realPath == nullptr) {
        LOG_ERROR("resolved input path(" << path << ") failed: " << errno << ": " << strerror(errno));
        return false;
    }

    if (!StrUtil::StartWith(realPath, ufsPath)) {
        LOG_ERROR("input path(" << path << ") not start with ufs mount path");
        return false;
    }

    struct stat statBuf {};
    auto ret = lstat(realPath, &statBuf);
    if (ret != 0) {
        LOG_ERROR("stat for path(" << realPath << ") failed : " << errno << ":" << strerror(errno));
        return false;
    }

    if ((statBuf.st_mode & S_IFMT) != S_IFDIR) {
        LOG_ERROR("path(" << realPath << ") is not directory but: " << (statBuf.st_mode & S_IFMT));
        return false;
    }

    resolvedPath = std::move(std::string(realPath));
    return true;
}

EventResult MemfsCheckDirEvent::CheckOneDirectory(const ock::memfs::DirectoryInfo &directoryInfo) noexcept
{
    std::stringstream message;
    message << "check_dir(path=" << directoryInfo.pathName << ", count=" << directoryInfo.fileCount << ")";

    auto realFileCount = ScanFinishedFileCount(directoryInfo.pathName, message);
    if (realFileCount < 0) {
        return EventResult{ false, static_cast<int>(realFileCount), message.str() };
    }

    if (static_cast<uint64_t>(realFileCount) < directoryInfo.fileCount) {
        message << "expect: " << directoryInfo.fileCount << ", real: " << realFileCount << ", not finished.";
        return EventResult{ false, 0, message.str() };
    }

    if (static_cast<uint64_t>(realFileCount) > directoryInfo.fileCount) {
        message << "expect: " << directoryInfo.fileCount << ", real: " << realFileCount << ", too many.";
        return EventResult{ true, -EFBIG, message.str() };
    }

    return EventResult{ true, 0 };
}

int64_t MemfsCheckDirEvent::ScanFinishedFileCount(const std::string &fromPath, std::stringstream &message) noexcept
{
    int64_t currentCount;
    int64_t totalCount = 0;
    std::string currentPath(fromPath);
    std::list<std::string> nextDirs;

    while ((currentCount = ScanFinishedFileCount(currentPath, nextDirs, message)) >= 0) {
        totalCount += currentCount;
        if (nextDirs.empty()) {
            break;
        }

        currentPath = std::move(nextDirs.front());
        nextDirs.pop_front();
    }

    if (currentCount < 0) {
        return -1L;
    }

    return totalCount;
}

int64_t MemfsCheckDirEvent::ScanFinishedFileCount(const std::string &fromPath, std::list<std::string> &nextDirs,
    std::stringstream &message) noexcept
{
    auto dir = opendir(fromPath.c_str());
    if (dir == nullptr) {
        int errnoNum = errno;
        if (errnoNum != GetErrorNumber()) {
            SetErrorNumber(errnoNum);
            LOG_ERROR("opendir for(" << fromPath << ") failed : " << errnoNum << " : " << strerror(errnoNum));
        }

        message << "opendir(" << fromPath << ") failed: " << errnoNum << ":" << strerror(errnoNum);
        return -errnoNum;
    }

    bool containsStage = false;
    int64_t fileCount = 0L;
    struct dirent *entry;
    while ((entry = readdir(dir)) != nullptr) {
        std::string name(entry->d_name);
        if (name == "." || name == "..") {
            continue;
        }

        if (entry->d_type == DT_DIR) {
            nextDirs.push_back(std::string(fromPath).append("/").append(name));
            continue;
        }

        if (entry->d_type != DT_REG) {
            std::string fullPath = std::string(fromPath).append("/").append(name);
            if (invalidTypes.find(fullPath) == invalidTypes.end()) {
                LOG_WARN("scan dir(" << fromPath << ") find entry(" << name << ") not regular file but: " <<
                    GetTypeName(entry->d_type) << "(" << static_cast<uint32_t>(entry->d_type) << ")");
                invalidTypes.insert(fullPath);
            }
            continue;
        }

        if (name.length() > stageFileSuffix.length() && StrUtil::EndWith(name, stageFileSuffix)) {
            LOG_DEBUG("stage file exist in path(" << fromPath << ") with name : " << name);
            message << "dir(" << fromPath << ")contains stage file: " << name;
            containsStage = true;
            break;
        }

        fileCount++;
    }
    closedir(dir);

    if (containsStage) {
        return -EAGAIN;
    }

    return fileCount;
}

const std::string &MemfsCheckDirEvent::GetTypeName(unsigned char key) noexcept
{
    static const std::string UNKNOWN_NAME = "UNKNOWN_NAME";
    auto pos = typeNames.find(key);
    if (pos != typeNames.end()) {
        return pos->second;
    }
    return UNKNOWN_NAME;
}
}
}