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
#include <vector>

#include "background_log.h"
#include "backup_file_tracer.h"
#include "backup_initiator.h"

using namespace ock::bg::backup;

int BackupInitiator::Initialize(const TaskPool &pool, const std::shared_ptr<BackupTarget> &target) noexcept
{
    taskPool = pool;
    backupTarget = target;

    auto ret = Prepare();
    if (ret != 0) {
        taskPool.reset();
        backupTarget.reset();
        return -1;
    }

    return 0;
}

void BackupInitiator::Destroy() noexcept
{
    taskPool.reset();
    backupTarget.reset();
}

CompareFileResult BackupInitiator::CompareFile(const std::string &path, int64_t inode, struct stat &buf) noexcept
{
    auto ret = GetAttribute(0, path, buf);
    if (ret == -ENOENT) {
        return FILE_NOT_EXIST;
    }

    if (ret != 0) {
        BKG_LOG_ERROR("get file(" << path.c_str() << ") attribute failed : " << ret << " : " << strerror(-ret) << ".");
        return IO_FAILED;
    }

    if (buf.st_ino != static_cast<uint64_t>(inode)) {
        return INODE_NOT_MATCH;
    }

    return FILE_SAME;
}

void BackupInitiator::SubmitRemoveFileTask(const std::string &path, int64_t inode) noexcept
{
    BKG_LOG_DEBUG("remove path(" << path.c_str() << "), inode(" << inode << ")");
    struct stat fileStat {};
    auto result = CompareFile(path, inode, fileStat);
    if (result == IO_FAILED) {
        BKG_LOG_ERROR("stat for file(" << path.c_str() << ") failed, skip.");
        return;
    }

    if (result == INODE_NOT_MATCH) {
        BKG_LOG_INFO("remove file, new created, inode from(" << inode << ") to (" << fileStat.st_ino << "), skip");
        return;
    }

    if (result == FILE_SAME) {
        BKG_LOG_ERROR("file(" << path.c_str() << ") inode(" << inode << ") removed, but still exist skip");
        return;
    }

    backupTarget->RemoveFile(path, inode);
}

int BackupInitiator::CompareTraceFile(const FileTrace &trace, struct stat &fileStat) noexcept
{
    auto result = CompareFile(trace.path, trace.inode, fileStat);
    if (result == FILE_NOT_EXIST) {
        BKG_LOG_INFO("file(" << trace.path.c_str() << ") not exist now, skip.");
        return -1;
    }

    if (result == IO_FAILED) {
        BKG_LOG_ERROR("stat for file(" << trace.path.c_str() << ") failed, skip.");
        return -1;
    }

    if (result == INODE_NOT_MATCH) {
        BKG_LOG_INFO("file changed, inode from(" << trace.inode << ") to (" << fileStat.st_ino << "), skip");
        return -1;
    }
    return 0;
}

NotifyProcessMark::NotifyProcessMark(BackupInitiator *init) noexcept : initiator{ init }
{
    initiator->SetProcessingMark();
}

NotifyProcessMark::~NotifyProcessMark() noexcept
{
    initiator->ClearProcessingMark();
}
