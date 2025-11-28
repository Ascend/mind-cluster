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
#include <unistd.h>
#include <climits>
#include <fstream>
#include <algorithm>

#include "hlog.h"
#include "auditlog_adapt.h"
#include "mem_fs_constants.h"
#include "service_configure.h"

using namespace ock::memfs;
using namespace ock::common::config;

static constexpr auto MAX_MEM_DATA_CAP = 1024;  // GB
static constexpr auto MAX_DATA_BLOCK_SZ = 1024; // MB
static constexpr auto MIN_OPEN_FS = 1024;
static constexpr auto MAX_OPEN_FS = 65536;
static constexpr auto MAX_UFS_CNT = 32;
static constexpr auto MAX_BG_BAK_TH_NUM = 256;

static constexpr auto MIN_WRITE_PARALLEL_THREAD = 2;
static constexpr auto MAX_WRITE_PARALLEL_THREAD = 96;

static constexpr auto MIN_WRITE_PARALLEL_SLICE_MB = 1;
static constexpr auto MAX_WRITE_PARALLEL_SLICE_MB = 1024;

static constexpr auto MIN_IPC_MSG_SZ = 4;
static constexpr auto MAX_IPC_MSG_SZ = 1024;

static constexpr auto MIN_BG_OP_FAILED_CNT = 10;
static constexpr auto MAX_BG_OP_FAILED_CNT = 60;

static constexpr auto MIN_BG_FAIL_RETRY_TIMES = 10;
static constexpr auto MAX_BG_FAIL_RETRY_TIMES = 60;

static constexpr auto MIN_BG_FAIL_RETRY_INTERVAL_SEC = 10;
static constexpr auto MAX_BG_FAIL_RETRY_INTERVAL_SEC = 60;

static constexpr auto GB_TO_BYTES_SHIFT = 30;
static constexpr auto MB_TO_BYTES_SHIFT = 20;
static constexpr auto KB_TO_BYTES_SHIFT = 10;

static constexpr auto FREE_MEM_DIVISOR_FOUR = 4UL;

static auto mfsMntPath = std::make_pair("memfs.mount_path", "/mnt/memfs");
static auto mfsMemCap = std::make_pair("memfs.data_block_pool_capacity_in_gb", 128);
static auto mfsBlkSz = std::make_pair("memfs.data_block_size_in_mb", 128);
static auto mfsMaxOpenFiles = std::make_pair("memfs.max_open_files", 4096);
static auto mfsWriteParallelEnabled = std::make_pair("memfs.write.parallel.enabled", true);
static auto mfsWriteParallelThreadNum = std::make_pair("memfs.write.parallel.thread_num", 16);
static auto mfsWriteParallelSliceInMb = std::make_pair("memfs.write.parallel.slice_in_mb", 16);
static auto mfsMultiGroupEnabled = std::make_pair("memfs.multi_group.enabled", false);

static auto ufsCnt = std::make_pair("underfs.count", 1);
static auto ufsDft = std::make_pair("underfs.default", 1);

static auto backupEnabled = std::make_pair("background.backup.enabled", true);
static auto backupThNum = std::make_pair("background.backup.thread_num", 32);
static auto backupUfs = std::make_pair("background.backup.ufs_name", "pacific1");
static auto backupFailRetryTimes = std::make_pair("background.backup.failed_retry_times", 10);
static auto backupFailRetryMaxIntervalSec = std::make_pair("background.backup.failed_retry_max_interval_sec", 10);
static auto backupFailMaxCnt = std::make_pair("background.backup.failed_max_cnt_for_unserviceable", 10);
static auto backupAutoEvictFile = std::make_pair("background.backup.failed_auto_evict_file", false);

static auto ipcMsg = std::make_pair("ipc.max_message_size_in_kb", 16);
static auto ipcPermSuperUser = std::make_pair("ipc.permit_super_user_access", true);
static auto ipcMsgAuthorEnabled = std::make_pair("ipc.author.enabled", false);
static auto ipcMsgAuthorEncrypted = std::make_pair("ipc.author.encrypted", false);
static auto ipcMsgAuthorFilePath = std::make_pair("ipc.author.file_path", "");
static auto ipcMsgkmcKsfMaster = std::make_pair("ipc.kmc.ksf.master", "");
static auto ipcMsgkmcKsfStandby = std::make_pair("ipc.kmc.ksf.standby", "");
static auto ipcTlsEnabled = std::make_pair("ipc.tls.enabled", false);
static auto ipcTlsCert = std::make_pair("ipc.tls.cert", "");
static auto ipcTlsCrl = std::make_pair("ipc.tls.crl", "");
static auto ipcTlsCA = std::make_pair("ipc.tls.ca", "");
static auto ipcTlsPK = std::make_pair("ipc.tls.pk", "");
static auto ipcTlsPkPwd = std::make_pair("ipc.tls.pk.pwd", "");

static bool ParseMemFsConfig(const ConfigurationPtr &conf, MemFsConfig &memFs) noexcept;
static bool ParseUnderFileSystemsConfig(const ConfigurationPtr &conf, UnderFsConfig &ufs) noexcept;
static void ParseBackgroundConfig(const ConfigurationPtr &conf, BackgroundConfig &config) noexcept;
static void ParseIpcMessageConfig(const ConfigurationPtr &conf, IpcMessageConfig &config) noexcept;

void MemFsConfigure::LoadDefaultConf()
{
    LoadMemFsDefaultConfig();
    LoadUnderFsDefaultConfig();
    LoadBackgroundDefaultConfig();
    LoadIpcConfig();
}

void MemFsConfigure::LoadMemFsDefaultConfig() noexcept
{
    AddStrConf(mfsMntPath, VStrNotNull::Create(mfsMntPath.first));
    AddIntConf(mfsMemCap, VIntRange::Create(mfsMemCap.first, 1, MAX_MEM_DATA_CAP));
    AddIntConf(mfsBlkSz, VIntRange::Create(mfsBlkSz.first, 1, MAX_DATA_BLOCK_SZ));
    AddIntConf(mfsMaxOpenFiles, VIntRange::Create(mfsMaxOpenFiles.first, MIN_OPEN_FS, MAX_OPEN_FS));

    AddBoolConf(mfsWriteParallelEnabled);
    AddIntConf(mfsWriteParallelThreadNum,
        VIntRange::Create(mfsWriteParallelThreadNum.first, MIN_WRITE_PARALLEL_THREAD, MAX_WRITE_PARALLEL_THREAD));
    AddIntConf(mfsWriteParallelSliceInMb,
        VIntRange::Create(mfsWriteParallelSliceInMb.first, MIN_WRITE_PARALLEL_SLICE_MB, MAX_WRITE_PARALLEL_SLICE_MB));

    AddBoolConf(mfsMultiGroupEnabled);
}

void MemFsConfigure::LoadUnderFsDefaultConfig() noexcept
{
    AddIntConf(ufsCnt, VIntRange::Create(ufsCnt.first, 1, MAX_UFS_CNT));
    AddIntConf(ufsDft, VIntRange::Create(ufsDft.first, 1, MAX_UFS_CNT));
    for (auto i = 1; i <= MAX_UFS_CNT; ++i) {
        // underfs.1.type = pacific
        auto key = "underfs." + std::to_string(i) + ".type";
        AddStrConf(std::make_pair(key, "pacific"), VStrNotNull::Create(key));

        // underfs.1.name = pacific01
        key = "underfs." + std::to_string(i) + ".name";
        auto value = "pacific" + std::to_string(i);
        AddStrConf(std::make_pair(key, value), VStrNotNull::Create(key));

        // underfs.1.mount_path = /
        key = "underfs." + std::to_string(i) + ".mount_path";
        value = "/";
        AddStrConf(std::make_pair(key, value), VStrNotNull::Create(key));

        // underfs.1.docker_map_path =
        key = "underfs." + std::to_string(i) + ".docker_map_path";
        value = "";
        AddStrConf(std::make_pair(key, value));
    }
}

void MemFsConfigure::LoadBackgroundDefaultConfig() noexcept
{
    AddBoolConf(backupEnabled);
    AddIntConf(backupThNum, VIntRange::Create(backupThNum.first, 1, MAX_BG_BAK_TH_NUM));
    AddStrConf(backupUfs, VStrNotNull::Create(backupUfs.first));
    AddIntConf(backupFailMaxCnt, VIntRange::Create(backupFailMaxCnt.first, MIN_BG_OP_FAILED_CNT, MAX_BG_OP_FAILED_CNT));
    AddBoolConf(backupAutoEvictFile);
    AddIntConf(backupFailRetryTimes,
        VIntRange::Create(backupFailRetryTimes.first, MIN_BG_FAIL_RETRY_TIMES, MAX_BG_FAIL_RETRY_TIMES));
    AddIntConf(backupFailRetryMaxIntervalSec, VIntRange::Create(backupFailRetryMaxIntervalSec.first,
        MIN_BG_FAIL_RETRY_INTERVAL_SEC, MAX_BG_FAIL_RETRY_INTERVAL_SEC));
}

void MemFsConfigure::LoadIpcConfig() noexcept
{
    AddIntConf(ipcMsg, VIntRange::Create(backupEnabled.first, MIN_IPC_MSG_SZ, MAX_IPC_MSG_SZ));
    AddBoolConf(ipcPermSuperUser);
    AddBoolConf(ipcMsgAuthorEnabled);
    AddBoolConf(ipcMsgAuthorEncrypted);
    AddStrConf(ipcMsgAuthorFilePath);
    AddStrConf(ipcMsgkmcKsfMaster);
    AddStrConf(ipcMsgkmcKsfStandby);
    AddBoolConf(ipcTlsEnabled);
    AddStrConf(ipcTlsCert);
    AddStrConf(ipcTlsCrl);
    AddStrConf(ipcTlsCA);
    AddStrConf(ipcTlsPK);
    AddStrConf(ipcTlsPkPwd);
}


ServiceConfigure &ServiceConfigure::GetInstance() noexcept
{
    static ServiceConfigure instance;
    return instance;
}

int ServiceConfigure::Initialize() noexcept
{
    std::string configPath = GetWorkPath();
    if (configPath.empty()) {
        return -1;
    }
    configPath.append("/conf/memfs.conf");
    HLOG_INFO("start to read config file.");

    ConfigurationPtr conf = Configuration::GetInstance<MemFsConfigure>();
    if (conf.get() == nullptr) {
        HLOG_ERROR("create config object failed");
        return -1;
    }

    if (!CheckConfigFile(configPath)) {
        return -1;
    }

    if (!conf->ReadConf<MemFsConfigure>(configPath)) {
        HLOG_ERROR("read config file failed");
        return -1;
    }

    auto errors = conf->Validate();
    if (!errors.empty()) {
        for (auto &error : errors) {
            std::cerr << "config(" << error << ")" << std::endl;
            HLOG_ERROR("config(" << error << ") failed");
        }
        return -1;
    }

    if (!ParseMemFsConfig(conf, memFsConfig)) {
        return -1;
    }

    if (!ParseUnderFileSystemsConfig(conf, underFileSystems)) {
        return -1;
    }

    ParseBackgroundConfig(conf, backgroundConfig);
    ParseIpcMessageConfig(conf, ipcMessageConfig);

    HLOG_INFO("finished to read config file.");
    return 0;
}

void ServiceConfigure::Destroy() noexcept {}

const MemFsConfig &ServiceConfigure::GetMemFsConfig() const noexcept
{
    return memFsConfig;
}

const UnderFsConfig &ServiceConfigure::GetUnderFileSystemConfig() const noexcept
{
    return underFileSystems;
}

const BackgroundConfig &ServiceConfigure::GetBackgroundConfig() const noexcept
{
    return backgroundConfig;
}

const IpcMessageConfig &ServiceConfigure::GetIpcMessageConfig() const noexcept
{
    return ipcMessageConfig;
}

void ServiceConfigure::SetWorkPath(const std::string &workPath) noexcept
{
    serverWorkPath = workPath;
}

std::string ServiceConfigure::GetWorkPath() noexcept
{
    return serverWorkPath;
}

bool ServiceConfigure::CheckConfigFile(const std::string &configPath) noexcept
{
    char confRealPath[PATH_MAX + 1] = {'\0'};
    if (realpath(configPath.c_str(), confRealPath) == nullptr) {
        HLOG_ERROR("Get realpath for configure file failed : " << errno << " : " << strerror(errno));
        return false;
    }

    struct stat statBuf {};
    auto ret = stat(confRealPath, &statBuf);
    if (ret != 0) {
        HLOG_ERROR("stat for configure file failed : " << errno << " : " << strerror(errno));
        return false;
    }

    if (statBuf.st_size > ock::memfs::MemFsConstants::MEMFS_CONF_MAX_FILE_SIZE) {
        HLOG_ERROR("stat for configure file size too large : " << statBuf.st_size);
        return false;
    }

    return true;
}

static bool ParseMemFsConfigCheck(uint64_t blockSize, uint64_t &capacity) noexcept
{
    if (blockSize <= 0 || capacity <= 0) {
        HLOG_ERROR("dataBlockSize(" << blockSize << ") capacityInBytes (" << capacity << ")invalid.");
        return false;
    }
    std::ifstream memInfo("/proc/meminfo");
    std::string line;
    uint64_t memAvailable = 0;
    while (std::getline(memInfo, line)) {
        std::istringstream iss(line);
        std::string key;
        uint64_t value;
        iss >> key >> value;
        if (key == "MemAvailable:") {
            memAvailable = value << KB_TO_BYTES_SHIFT;
            break;
        }
    }
    memInfo.close();
    if (capacity > memAvailable) {
        std::string expectMemory = "expect memory(" + std::to_string(capacity >> GB_TO_BYTES_SHIFT) + "GB)";
        ock::common::HLOG_AUDIT("system", "init", expectMemory, "out available memory");
        HLOG_WARN("capacity bytes (" << capacity << ") exceeds available memory (" << memAvailable << ")");
        capacity = memAvailable / FREE_MEM_DIVISOR_FOUR / (1 << GB_TO_BYTES_SHIFT) * (1 << GB_TO_BYTES_SHIFT);
        if (capacity == 0UL) {
            HLOG_ERROR("no free memory is available in the environment.");
            return false;
        }
        std::string realMemory = "real memory(" + std::to_string(capacity >> GB_TO_BYTES_SHIFT) + "GB)";
        ock::common::HLOG_AUDIT("system", "init", realMemory, "success");
    }
    HLOG_INFO("reset valid memory capacity: " << capacity << "bytes, " << (capacity >> GB_TO_BYTES_SHIFT) << "GB");
    return true;
}

static bool ParseMemFsConfig(const ConfigurationPtr &conf, MemFsConfig &memFs) noexcept
{
    memFs.mountPath = conf->GetStr(mfsMntPath.first);

    auto dataBlockSizeInMb = conf->GetInt(mfsBlkSz.first);
    memFs.dataBlockSize = (static_cast<uint64_t>(dataBlockSizeInMb) << MB_TO_BYTES_SHIFT);

    auto capacityInGb = conf->GetInt(mfsMemCap.first);
    auto capacityInBytes = (static_cast<uint64_t>(capacityInGb) << GB_TO_BYTES_SHIFT);

    auto ret = ParseMemFsConfigCheck(memFs.dataBlockSize, capacityInBytes);
    if (!ret) {
        return false;
    }
    memFs.dataBlockCount = capacityInBytes / memFs.dataBlockSize;
    memFs.maxOpenFiles = conf->GetInt(mfsMaxOpenFiles.first);

    memFs.writeParallel.enabled = conf->GetBool(mfsWriteParallelEnabled.first);
    memFs.writeParallel.threadNum = conf->GetInt(mfsWriteParallelThreadNum.first);
    memFs.writeParallel.sliceInMB = conf->GetInt(mfsWriteParallelSliceInMb.first);
    memFs.multiGroupEnabled = conf->GetBool(mfsMultiGroupEnabled.first);
    return true;
}

static bool ParseUnderFileSystemsConfig(const ConfigurationPtr &conf, UnderFsConfig &ufs) noexcept
{
    auto count = conf->GetInt(ufsCnt.first);
    auto dftFs = conf->GetInt(ufsDft.first);
    if (dftFs > count || dftFs == 0) {
        HLOG_ERROR("default under fs(" << dftFs << ") invalid.");
        return false;
    }

    std::vector<UnderFsInstance> instances;
    for (auto i = 1; i <= count; i++) {
        UnderFsInstance instance;
        auto typeKey = "underfs." + std::to_string(i) + ".type";
        instance.type = conf->GetStr(typeKey);

        auto nameKey = "underfs." + std::to_string(i) + ".name";
        instance.name = conf->GetStr(nameKey);

        auto mountPathKey = "underfs." + std::to_string(i) + ".mount_path";
        instance.options["mount_path"] = conf->GetStr(mountPathKey);

        auto dockerMapPathKey = "underfs." + std::to_string(i) + ".docker_map_path";
        instance.options["docker_map_path"] = conf->GetStr(dockerMapPathKey);
        instances.push_back(instance);
    }

    ufs.defaultName = instances[dftFs - 1].name;
    std::for_each(instances.begin(), instances.end(),
        [&ufs](const UnderFsInstance &inst) { ufs.instances[inst.name] = inst; });

    return true;
}

static void ParseBackgroundConfig(const ConfigurationPtr &conf, BackgroundConfig &config) noexcept
{
    auto &backup = config.backupServiceConfig;
    backup.enabled = conf->GetBool(backupEnabled.first);
    backup.threadNum = conf->GetInt(backupThNum.first);
    backup.autoEvictFile = conf->GetBool(backupAutoEvictFile.first);
    backup.maxFailCntForUnserviceable = static_cast<uint32_t>(conf->GetInt(backupFailMaxCnt.first));
    backup.retryTimes = static_cast<uint32_t>(conf->GetInt(backupFailRetryTimes.first));
    backup.retryIntervalSec = static_cast<uint32_t>(conf->GetInt(backupFailRetryMaxIntervalSec.first));

    BackupInstance instance;
    instance.source = "mfs";
    instance.destType = "under_fs";
    instance.destName = conf->GetStr(backupUfs.first);
    instance.opened = true;
    backup.backups.push_back(instance);
}

static void ParseIpcMessageConfig(const ConfigurationPtr &conf, IpcMessageConfig &config) noexcept
{
    auto maxSizeInKb = conf->GetInt(ipcMsg.first);
    config.ipcMessageSize = (static_cast<uint64_t>(maxSizeInKb) << KB_TO_BYTES_SHIFT);

    config.permitSuperUser = conf->GetBool(ipcPermSuperUser.first);
    config.authorEnabled = conf->GetBool(ipcMsgAuthorEnabled.first);
    config.authorEncrypted = conf->GetBool(ipcMsgAuthorEncrypted.first);
    config.authorFilePath = conf->GetStr(ipcMsgAuthorFilePath.first);
    config.kmcKsfMaster = conf->GetStr(ipcMsgkmcKsfMaster.first);
    config.kmcKsfStandby = conf->GetStr(ipcMsgkmcKsfStandby.first);
    config.tlsEnabled = conf->GetBool(ipcTlsEnabled.first);
    config.tlsCertPath = conf->GetStr(ipcTlsCert.first);
    config.tlsCrlPath = conf->GetStr(ipcTlsCrl.first);
    config.tlsCaPath = conf->GetStr(ipcTlsCA.first);
    config.tlsPriKeyPath = conf->GetStr(ipcTlsPK.first);
    config.tslPriKeyPwdPath = conf->GetStr(ipcTlsPkPwd.first);
}