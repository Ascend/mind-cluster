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

#ifndef OCKIO_MEMFS_CHECK_DIR_EVENT_H
#define OCKIO_MEMFS_CHECK_DIR_EVENT_H

#include <list>
#include <unordered_set>
#include <unordered_map>
#include <sstream>
#include "memfs_event_manager.h"

namespace ock {
namespace memfs {
struct DirectoryInfo {
    std::string pathName;
    uint64_t fileCount;
    DirectoryInfo() noexcept : fileCount{ 0UL } {}
    DirectoryInfo(std::string path, uint64_t count) noexcept : pathName{ std::move(path) }, fileCount{ count } {}
};

using DirectoriesInfo = std::list<DirectoryInfo>;

class MemfsCheckDirEvent : public MemfsEvent {
public:
    MemfsCheckDirEvent(std::string nm, uint64_t timeoutSec, uint64_t intervalMs, DirectoriesInfo info,
        const std::function<void(uint64_t, int)> &cb) noexcept;

public:
    bool PreCheckEvent() noexcept override;
    EventResult Process() noexcept override;
    void Callback(uint64_t eventId, int result) noexcept override;

private:
    static bool PreCheckOneDirectory(const std::string &ufsPath, const std::string &path,
        std::string &resolvedPath) noexcept;
    EventResult CheckOneDirectory(const DirectoryInfo &directoryInfo) noexcept;
    int64_t ScanFinishedFileCount(const std::string &fromPath, std::stringstream &message) noexcept;
    int64_t ScanFinishedFileCount(const std::string &fromPath, std::list<std::string> &nextDirs,
        std::stringstream &message) noexcept;
    static const std::string &GetTypeName(unsigned char key) noexcept;

private:
    DirectoriesInfo directoriesInfo;
    std::unordered_set<std::string> invalidTypes;
    std::function<void(uint64_t, int)> callback;
    static const std::string stageFileSuffix;
    static const std::unordered_map<unsigned char, std::string> typeNames;
};
} // ock
} // memfs

#endif // OCKIO_MEMFS_CHECK_DIR_EVENT_H
