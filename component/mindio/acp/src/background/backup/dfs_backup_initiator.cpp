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
#include <sys/stat.h>
#include <fcntl.h>
#include "../background_log.h"
#include "ufs_api.h"
#include "memfs_api.h"
#include "dfs_backup_initiator.h"
using namespace ock::bg::backup;
using namespace ock::memfs;
static constexpr auto IO_BUF_SZ = 16U * 1024U;

__thread bool DfsBackupInitiator::marked = false;

int DfsBackupInitiator::GetAttribute(uint64_t taskId, const std::string &path, struct stat &buf) noexcept
{
    return 0;
}

int DfsBackupInitiator::MultiCopyFileToUfs(uint64_t taskId, const std::string &path, UFS &ufs) noexcept
{
    auto fd = MDogOpenFile(MDOG_INVALID_INODENUM, static_cast<short>(path.size()), path.c_str());
    if (fd < 0) {
        BKG_LOG_ERROR("task(" << taskId << ") open ockiod file(" << path.c_str() << ") failed: " << fd);
        return -1;
    }

    auto memfsFd = MemFsApi::OpenFile(path, O_RDONLY, 0);
    if (memfsFd < 0) {
        BKG_LOG_ERROR("open file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        return -1;
    }

    struct stat statBuf {};
    auto ret = MemFsApi::GetFileMeta(memfsFd, statBuf);
    if (ret != 0) {
        BKG_LOG_ERROR("stat file(" << path.c_str() << ") failed(" << errno << " : " << strerror(errno) << ")");
        MemFsApi::CloseFile(memfsFd);
        MDogCloseFile(fd, false);
        return -1;
    }

    auto os =
        ufs->PutFile(path, ufs::FileMode{ statBuf.st_mode & 0777, statBuf.st_uid, statBuf.st_gid });
    if (os == nullptr) {
        BKG_LOG_ERROR("task(" << taskId << ") create file(" << path.c_str() << ") failed: " << strerror(errno));
        MDogCloseFile(fd, false);
        MemFsApi::CloseFile(memfsFd);
        return -1;
    }

    ufs::utils::ByteBuffer buffer(IO_BUF_SZ);
    ssize_t readSize;
    off_t offset = 0;
    while ((readSize = MDogReadWriteData(fd, (char *)buffer.Data(), buffer.Capacity(), offset, false)) > 0) {
        auto writeSize = os->Write(buffer.Data(), static_cast<uint32_t>(readSize));
        if (writeSize != static_cast<int>(readSize)) {
            BKG_LOG_ERROR("task(" << taskId << ") write file(" << path.c_str() << ") size(" << readSize << " vs " <<
                writeSize << ")");
            MDogCloseFile(fd, false);
            MemFsApi::CloseFile(memfsFd);
            return -1;
        }

        offset += readSize;
    }
    MemFsApi::CloseFile(memfsFd);
    MDogCloseFile(fd, false);
    return 0;
}

void DfsBackupInitiator::SetProcessingMark() noexcept
{
    marked = true;
}

void DfsBackupInitiator::ClearProcessingMark() noexcept
{
    marked = false;
}

bool DfsBackupInitiator::CheckProcessingMark() noexcept
{
    return marked;
}

int DfsBackupInitiator::Prepare() noexcept
{
    FileOpNotify notify;

    notify.openNotify = [this](int fd, const std::string &name, int64_t inode, bool dir) -> int {
        return OpenFileNotify(fd, name, inode, dir);
    };

    notify.closeNotify = [this](int fd, bool dir) { CloseFileNotify(fd, dir); };

    notify.mkdirNotify = [this](const std::string &name, int64_t inode) { CreateDirectoryNotify(name, inode); };

    notify.unlinkNotify = [this](const std::string &name, int64_t inode) { RemoveFileNotify(name, inode); };

    auto ret = MDogFsRegisterFileOpNotify(notify);
    if (ret != 0) {
        BKG_LOG_ERROR("register file operate notify filed(" << ret << ")");
        return -1;
    }

    BKG_LOG_INFO("register file operate notify success.");
    return 0;
}

int DfsBackupInitiator::OpenFileNotify(int fd, const std::string &path, int64_t inode, bool dir) noexcept
{
    if (CheckProcessingMark()) {
        return 0;
    }

    BKG_LOG_DEBUG("open file(" << path.c_str() << ") fd(" << fd << ") inode(" << inode << ") is_dir(" <<
        (dir ? "true" : "false") << ")");
    fileTracer.TraceOpen(fd, path, inode, dir);
    return 0;
}

void DfsBackupInitiator::CloseFileNotify(int fd, bool dir) noexcept
{
    if (CheckProcessingMark()) {
        return;
    }

    FileTrace trace;
    BKG_LOG_DEBUG("close file(" << fd << ")");
    if (!fileTracer.CloseFind(fd, trace)) {
        BKG_LOG_ERROR("close file(" << fd << ") is_dir(" << (dir ? "true" : "false") << ") not opened.");
        return;
    }

    if (trace.path == "/") {
        return;
    }

    NotifyProcessMark marker(this);
    SplitUploadFileTask();
}

void DfsBackupInitiator::CreateDirectoryNotify(const std::string &path, int64_t inode) noexcept
{
    if (CheckProcessingMark()) {
        return;
    }

    BKG_LOG_DEBUG("create directory path(" << path.c_str() << "), inode(" << inode << ")");
    FileTrace trace{ path, inode };

    NotifyProcessMark marker(this);
}

void DfsBackupInitiator::RemoveFileNotify(const std::string &path, int64_t inode) noexcept
{
    if (CheckProcessingMark()) {
        return;
    }

    NotifyProcessMark marker(this);
    SubmitRemoveFileTask(path, inode);
}
int DfsBackupInitiator::CopyFileToMemfs(uint64_t taskId, const std::string &path, BackupInitiator::UFS &ufs,
    const TaskInfo &taskInfo) noexcept
{
    return 0;
}

int DfsBackupInitiator::RecordToMemfsTaskResult(uint64_t taskId, const std::string &path, int taskResult,
    const TaskInfo &taskInfo) noexcept
{
    return 0;
}