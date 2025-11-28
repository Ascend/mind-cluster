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
#ifndef OCK_DFS_SERVICE_CONFIGURE_H
#define OCK_DFS_SERVICE_CONFIGURE_H

#include <cstdint>
#include <string>
#include <map>
#include <list>
#include <unordered_map>

#include "memfs_configuration.h"
#include "non_copyable.h"

namespace ock {
namespace common {
namespace config {

struct WriteParallelConfig {
    bool enabled;
    uint32_t threadNum;
    uint32_t sliceInMB;
};

struct MemFsConfig {
    std::string mountPath;
    uint64_t dataBlockSize;
    uint64_t dataBlockCount;
    bool multiGroupEnabled{ false };
    int maxOpenFiles;
    WriteParallelConfig writeParallel;
};

struct UnderFsInstance {
    std::string name;
    std::string type;
    std::map<std::string, std::string> options;
};

struct UnderFsConfig {
    std::map<std::string, UnderFsInstance> instances;
    std::string defaultName;
};

struct BackupInstance {
    std::string source;
    std::string destType;
    std::string destName;
    bool opened { false };
};

struct BackupServiceConfig {
    bool enabled { false };
    bool autoEvictFile { false };
    uint32_t threadNum { 0 };
    uint32_t maxFailCntForUnserviceable { 0 };
    uint32_t retryTimes { 0 };
    uint32_t retryIntervalSec { 0 };
    std::list<BackupInstance> backups;
};

struct BackgroundConfig {
    BackupServiceConfig backupServiceConfig;
};

struct IpcMessageConfig {
    uint64_t ipcMessageSize{ 0 };
    bool permitSuperUser{ false };
    bool authorEnabled{ false };
    bool authorEncrypted{ false };
    bool tlsEnabled{ false };
    std::string authorFilePath;
    std::string kmcKsfMaster;
    std::string kmcKsfStandby;
    std::string tlsCaPath;
    std::string tlsCertPath;
    std::string tlsCrlPath;
    std::string tlsPriKeyPath;
    std::string tslPriKeyPwdPath;
};

class MemFsConfigure : public ock::memfs::Configuration {
public:
    void LoadDefaultConf() override;

private:
    void LoadMemFsDefaultConfig() noexcept;
    void LoadUnderFsDefaultConfig() noexcept;
    void LoadBackgroundDefaultConfig() noexcept;
    void LoadIpcConfig() noexcept;
};

class ServiceConfigure : public NonCopyable {
public:
    static ServiceConfigure &GetInstance() noexcept;

public:
    int Initialize() noexcept;
    void Destroy() noexcept;
    const MemFsConfig &GetMemFsConfig() const noexcept;
    const UnderFsConfig &GetUnderFileSystemConfig() const noexcept;
    const BackgroundConfig &GetBackgroundConfig() const noexcept;
    const IpcMessageConfig &GetIpcMessageConfig() const noexcept;

public:
    void SetWorkPath(const std::string &workPath) noexcept;
    std::string GetWorkPath() noexcept;
    bool CheckConfigFile(const std::string &configPath) noexcept;

private:
    ServiceConfigure() = default;
    ~ServiceConfigure() override = default;

private:
    std::string serverWorkPath;
    MemFsConfig memFsConfig;
    UnderFsConfig underFileSystems;
    BackgroundConfig backgroundConfig;
    IpcMessageConfig ipcMessageConfig;
};
}
}
}


#endif // OCK_DFS_SERVICE_CONFIGURE_H
