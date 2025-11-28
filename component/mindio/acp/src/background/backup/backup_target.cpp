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

#include "memfs_api.h"
#include "background_log.h"
#include "backup_file_manager.h"

using namespace ock::bg::backup;

static constexpr auto STR_TO_NUM_BASE = 10;

std::atomic<uint64_t> BackupTarget::taskIdGen{ 0x1234abUL };

static inline int CompareTime(const struct timespec &t1, const struct timespec t2) noexcept
{
    if (t1.tv_sec != t2.tv_sec) {
        return t1.tv_sec > t2.tv_sec ? 1 : -1;
    }

    if (t1.tv_nsec != t2.tv_nsec) {
        return t1.tv_nsec > t2.tv_nsec ? 1 : -1;
    }

    return 0;
}

bool UnderFsFileView::AddUploadFileToView(const FileTrace &trace, const struct stat &buf, bool force) noexcept
{
    FileMeta oldMeta;
    FileMeta meta(trace.inode, buf.st_mtim);

    if (force || backupFileView->AddFile(trace.path, meta, oldMeta)) {
        return true;
    }

    if (oldMeta.inode == meta.inode) {
        if (CompareTime(oldMeta.lastBackupTime, buf.st_mtim) > 0) {
            BKG_LOG_DEBUG("file(" << trace.path.c_str() << ") last modified(" << buf.st_mtim.tv_sec << "." <<
                buf.st_mtim.tv_nsec << ") less backup time(" << oldMeta.lastBackupTime.tv_sec << "." <<
                oldMeta.lastBackupTime.tv_nsec << "), no need upload.");
            return false;
        }

        if (!backupFileView->RefreshBackupTime(trace.path, trace.inode)) {
            BKG_LOG_INFO("upload file(" << trace.path.c_str() << ") refresh backup time inode not match, ignore.");
        }
        return true;
    }

    if (static_cast<int64_t>(buf.st_ino) != trace.inode) {
        BKG_LOG_INFO("file(" << trace.path.c_str() << ") now inode changed, do not upload.");
        return false;
    }

    return backupFileView->UpdateFile(trace.path, oldMeta.inode, meta);
}

bool UnderFsFileView::DoRemoveFile(uint64_t taskId, const std::string &path, int64_t inode, int64_t ufsInode,
    bool file) noexcept
{
    BKG_LOG_DEBUG("execute delete task(" << taskId << ") path(" << path.c_str() << ") inode(" << inode << ") start");

    ufs::FileMeta fileMeta;
    auto ret = underFs->GetFileMeta(path, fileMeta);
    if (ret < 0 && errno == ENOENT) {
        BKG_LOG_INFO("execute delete task(" << taskId << ") path(" << path.c_str() << ") inode(" << inode <<
            ") already removed");
        return true;
    }

    if (ret < 0) {
        BKG_LOG_INFO("execute delete task(" << taskId << ") stat(" << path.c_str() << ") failed(" << errno << " : " <<
            strerror(errno) << ")");
        return false;
    }

    auto pos = fileMeta.meta.find("st_ino");
    if (pos == fileMeta.meta.end() ||
        (ufsInode > 0 && strtol(pos->second.c_str(), nullptr, STR_TO_NUM_BASE) != ufsInode)) {
        BKG_LOG_INFO("execute delete task(" << taskId << ") path(" << path.c_str() << ") inode(" << inode <<
            ") ufs inode(" << ufsInode << ") not matches removed");
        return true;
    }

    if (file) {
        ret = underFs->RemoveFile(path);
    } else {
        ret = underFs->RemoveDirectory(path);
    }

    if (ret < 0) {
        BKG_LOG_INFO("execute delete task(" << taskId << ") path(" << path.c_str() << ") inode(" << inode <<
            ") failed(" << errno << ":" << strerror(errno) << ")");
        return false;
    }
    return true;
}

int BackupTarget::Initialize(const std::string &srcName, const TaskPool &pool, const MUFS &mufs) noexcept
{
    if (mufs.empty()) {
        BKG_LOG_ERROR("backup target underfs is empty.");
        return -1;
    }

    sourceName = srcName;
    taskPool = pool;
    for (auto &ufs : mufs) {
        underFsFileView.emplace_back(ufs);
    }

    for (auto &view : underFsFileView) {
        if (!view.backupFileView->Valid()) {
            BKG_LOG_ERROR("backup file view invalid.");
            return -1;
        }
    }

    return 0;
}

void BackupTarget::Destroy() noexcept
{
    underFsFileView.clear();
}

bool BackupTarget::DoBackupFileWrapper(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view) noexcept
{
    if (!DoBackupFile(taskId, trace, view)) {
        return false;
    }
    view.backupFileView->RemoveFile(trace.path, trace.inode);
    return true;
}

void BackupTarget::UploadFile(const FileTrace &trace, const struct stat &fileStat, bool force) noexcept
{
    for (auto &ufsView : underFsFileView) {
        if (!ufsView.AddUploadFileToView(trace, fileStat, force)) {
            BKG_LOG_INFO("backup file(" << trace.path.c_str() << ") inode(" << trace.inode <<
                ") not add to view, skip");
            continue;
        }

        auto taskId = taskIdGen.fetch_add(1UL);
        BKG_LOG_INFO("submit task(" << taskId << ") file backup file(" << trace.path << ") inode(" << trace.inode <<
            ")");
        auto fd = memfs::MemFsApi::OpenFile(trace.path, O_RDONLY);
        if (fd < 0) {
            BKG_LOG_ERROR("open file(" << trace.path << ") read failed.");
            return;
        }
        taskPool->Submit([taskId, trace, &ufsView, fd, this]() -> bool {
            NotifyProcessMark marker(BackupFileManager::GetInstance().GetInitiator(sourceName).get());
            bool ret = DoBackupFileWrapper(taskId, trace, ufsView);
            memfs::MemFsApi::CloseFile(fd);
            memfs::MemFsApi::Unlink(trace.path);
            return ret;
        });
    }
}

int BackupTarget::CreateFileAndStageSync(const FileTrace &trace, const struct stat &buf) noexcept
{
    for (auto &ufsView : underFsFileView) {
        if (!ufsView.AddUploadFileToView(trace, buf)) {
            BKG_LOG_INFO("backup file(" << trace.path.c_str() << ") inode(" << trace.inode <<
                ") not add to view, skip");
            continue;
        }

        // 1. create stage file
        std::string stagePath = trace.path;
        stagePath.append(".m.stg");
        auto outs =
            ufsView.underFs->PutFile(stagePath, ufs::FileMode(buf.st_mode, buf.st_uid, buf.st_gid));
        if (outs == nullptr) {
            BKG_LOG_ERROR("create stage file(" << stagePath.c_str() << ") on UFS failed(" << errno << " : " <<
                strerror(errno) << ")");
            return -1;
        }
    }
    return 0;
}

int BackupTarget::RemoveFileAndStageSync(const ock::bg::backup::FileTrace &trace) noexcept
{
    std::string stagePath = trace.path;
    stagePath.append(".m.stg");

    for (auto &ufsView : underFsFileView) {
        auto ret = ufsView.underFs->RemoveFile(stagePath);
        if (ret < 0 && errno != ENOENT) {
            BKG_LOG_ERROR("remove file: " << stagePath << " failed: " << errno << strerror(errno));
            return -1;
        }

        ufs::FileMeta fileMeta;
        ret = ufsView.underFs->GetFileMeta(trace.path, fileMeta);
        if (ret < 0) {
            if (errno == ENOENT) {
                continue;
            }

            BKG_LOG_ERROR("stat for file: " << trace.path << " failed: " << errno << strerror(errno));
            return -1;
        }

        if (fileMeta.size > 0) {
            BKG_LOG_WARN("file: " << trace.path << " size: " << fileMeta.size << " do not remove when discard.");
            continue;
        }

        ret = ufsView.underFs->RemoveFile(trace.path);
        if (ret < 0 && errno != ENOENT) {
            BKG_LOG_ERROR("remove file: " << stagePath << " failed: " << errno << strerror(errno));
            return -1;
        }
    }

    return 0;
}

void BackupTarget::RemoveFile(const std::string &path, int64_t inode) noexcept
{
    ufs::FileMeta fileMeta;
    for (auto &ufsView : underFsFileView) {
        if (ufsView.underFs->GetFileMeta(path, fileMeta) != 0) {
            BKG_LOG_WARN("get file(" << path.c_str() << ") metadata from UFS failed(" << errno << " : " <<
                strerror(errno) << "), skip");
            continue;
        }

        auto ufsInode = -1L;
        bool commonFile = true;
        auto pos = fileMeta.meta.find("st_ino");
        if (pos != fileMeta.meta.end()) {
            ufsInode = static_cast<int64_t>(std::stoul(pos->second));
        }

        pos = fileMeta.meta.find("st_mode");
        if (pos != fileMeta.meta.end()) {
            auto mode = static_cast<mode_t>(std::strtol(pos->second.c_str(), nullptr, 10));
            commonFile = !S_ISDIR(mode);
        }

        if (!ufsView.backupFileView->RemoveFile(path, inode)) {
            BKG_LOG_INFO("remove path(" << path.c_str() << ") with inode(" << inode << ") not match, ignore.");
        }

        auto taskId = taskIdGen.fetch_add(1UL);
        BKG_LOG_DEBUG("submit task(" << taskId << ") file remove file(" << path.c_str() << ") inode(" << inode << ")");
        taskPool->Submit([taskId, path, inode, ufsInode, commonFile, &ufsView, this]() -> bool {
            NotifyProcessMark marker(BackupFileManager::GetInstance().GetInitiator(sourceName).get());
            return ufsView.DoRemoveFile(taskId, path, inode, ufsInode, commonFile);
        });
    }
}

int BackupTarget::CreateDir(const std::string &name, mode_t mode, uid_t owner, gid_t group) noexcept
{
    ufs::FileMode fileMode{ mode, owner, group };
    ufs::FileMeta fileMeta;
    for (auto &ufsView : underFsFileView) {
        auto ret = ufsView.underFs->CreateDirectory(name, fileMode);
        if (ret == 0) {
            continue;
        }

        if (errno != EEXIST) {
            BKG_LOG_ERROR("create dir(" << name << ") failed : " << errno << " : " << strerror(errno));
            return -1;
        }

        ret = ufsView.underFs->GetFileMeta(name, fileMeta);
        if (ret != 0) {
            BKG_LOG_ERROR("stat path(" << name << ") failed : " << errno << " : " << strerror(errno));
            return -1;
        }

        if ((fileMeta.mode & S_IFMT) != S_IFDIR) {
            errno = ENOTDIR;
            return -1;
        }
    }

    return 0;
}

int BackupTarget::StatFile(const std::string &path, struct stat &buf) noexcept
{
    struct ufs::FileAcl acl;
    return StatFile(path, buf, acl);
}

int BackupTarget::StatFile(const std::string &path, struct stat &buf, struct ufs::FileAcl &acl) noexcept
{
    if (underFsFileView.empty()) {
        return -EINVAL;
    }

    auto &targetUfs = underFsFileView.front().underFs;
    ock::ufs::FileMeta fileMeta;
    auto ret = targetUfs->GetFileMeta(path, fileMeta);
    if (ret != 0) {
        return -1;
    }

    std::map<std::string, std::string>::const_iterator pos;
    if ((pos = fileMeta.meta.find("st_uid")) != fileMeta.meta.end()) {
        buf.st_uid = std::strtoul(pos->second.c_str(), nullptr, STR_TO_NUM_BASE);
    }

    if ((pos = fileMeta.meta.find("st_gid")) != fileMeta.meta.end()) {
        buf.st_gid = std::strtoul(pos->second.c_str(), nullptr, STR_TO_NUM_BASE);
    }

    if ((pos = fileMeta.meta.find("st_mode")) != fileMeta.meta.end()) {
        buf.st_mode = std::strtoul(pos->second.c_str(), nullptr, STR_TO_NUM_BASE);
    }

    if ((pos = fileMeta.meta.find("st_size")) != fileMeta.meta.end()) {
        buf.st_size = std::strtol(pos->second.c_str(), nullptr, STR_TO_NUM_BASE);
    }

    if ((pos = fileMeta.meta.find("st_ino")) != fileMeta.meta.end()) {
        buf.st_ino = std::strtoul(pos->second.c_str(), nullptr, STR_TO_NUM_BASE);
    }

    acl = fileMeta.acl;
    return 0;
}

int BackupTarget::CheckStgMtime(UnderFsFileView &view, const std::string &stgPath) noexcept
{
    int checkTimes = 0;
    int timeOut = 10; // If mtime remains unchanged for more than 10 seconds, the synchronization times out.
    struct timespec oldMtime {
        0, 0
    };
    while (checkTimes < timeOut) {
        ufs::FileMeta stgFileMeta;
        auto ret = view.underFs->GetFileMeta(stgPath, stgFileMeta);
        if (ret != 0 && errno == ENOENT) {
            BKG_LOG_INFO("stg file(" << stgPath.c_str() <<
                ") has been removed, "
                "the original file has been backed up.");
            return FILE_BEEN_REMOVED;
        } else if (ret != 0) {
            BKG_LOG_ERROR("try to get stg file(" << stgPath.c_str() << ") stat failed(" << errno << " : " <<
                strerror(errno) << ")");
            checkTimes++;
            std::this_thread::sleep_for(std::chrono::seconds(1));
            continue;
        }

        struct timespec newMtime = stgFileMeta.mtime;
        if (oldMtime.tv_sec == newMtime.tv_sec && oldMtime.tv_nsec == newMtime.tv_nsec) {
            checkTimes++;
        } else {
            oldMtime = newMtime;
            checkTimes = 0;
        }
        std::this_thread::sleep_for(std::chrono::seconds(1));
    }
    return MTIME_NO_CHANGE_TIMEOUT;
}

int BackupTarget::TryLockStg(uint64_t taskId, UnderFsFileView &view, std::shared_ptr<ock::ufs::FileLock> &fileLock,
    const std::string &stgPath, const struct stat &stgLockBuf) noexcept
{
    while (true) {
        fileLock = view.underFs->GetFileLock(stgPath);
        if (fileLock == nullptr) {
            BKG_LOG_WARN("failed to acquire file lock for task(" << taskId << ") path(" << stgPath.c_str() << ")");
            return LOCK_ERROR;
        } else {
            BKG_LOG_INFO("acquire file lock for task(" << taskId << ") path(" << stgPath.c_str() << ") success");
        }

        if (fileLock->TryLock() == 0) {
            break;
        }
        BKG_LOG_INFO("lock file : " << stgPath.c_str() << " failed");
        auto ret = CheckStgMtime(view, stgPath);
        if (ret == FILE_BEEN_REMOVED) {
            return ret;
        }

        if (fileLock->TryLock() == 0) {
            ufs::FileMeta stgFileMeta;
            ret = view.underFs->GetFileMeta(stgPath, stgFileMeta);
            if (ret != 0 && errno == ENOENT) {
                BKG_LOG_INFO("get lock success, but stg file(" << stgPath.c_str() <<
                    ") has been removed, "
                    "the original file has been backed up.");
                fileLock->Unlock();
                return FILE_BEEN_REMOVED;
            } else {
                break;
            }
        } else {
            BKG_LOG_INFO("primary node synchronization timeout. Try to obtain the lock.");
        }

        ret = view.underFs->RemoveFile(stgPath);
        if (ret != 0 && errno != ENOENT) {
            BKG_LOG_ERROR("remove file: " << stgPath.c_str() << " failed: " << errno << strerror(errno) <<
                ", obtain the lock failed.");
        } else {
            BKG_LOG_DEBUG("remove file: " << stgPath.c_str() << " success.");
        }
        auto outs = view.underFs->PutFile(stgPath, ufs::FileMode(stgLockBuf.st_mode,
                                                                 stgLockBuf.st_uid, stgLockBuf.st_gid));
        if (outs == nullptr) {
            BKG_LOG_ERROR("create stg file(" << stgPath.c_str() << ") on UFS failed(" << errno << " : " <<
                        strerror(errno) << ")");
            return LOCK_ERROR;
        }
    }
    return TRYLOCK_SUCCESS;
}

int BackupTarget::UnLockStg(std::shared_ptr<ock::ufs::FileLock> &fileLock, const std::string &stgPath) noexcept
{
    auto ret = fileLock->Unlock();
    if (ret != 0) {
        BKG_LOG_ERROR("unlock file : " << stgPath.c_str() << " failed(" << errno << " : " << strerror(errno) << ")");
    } else {
        BKG_LOG_INFO("unlock file : " << stgPath.c_str() << " success");
    }
    return ret;
}

bool BackupTarget::DoRealBackupFile(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view,
    struct stat &stgLockBuf) noexcept
{
    auto realFilePath = trace.path;
    auto stagePath = trace.path;
    stagePath.append(".m.stg");

    std::shared_ptr<ock::ufs::FileLock> fileLock;
    auto ret = TryLockStg(taskId, view, fileLock, stagePath, stgLockBuf);
    if (ret == FILE_BEEN_REMOVED) {
        return true;
    } else if (ret == LOCK_ERROR) {
        return false;
    }

    BKG_LOG_INFO("lock file : " << stagePath.c_str() << " success, create real file.");

    auto outs = view.underFs->PutFile(realFilePath, ufs::FileMode(stgLockBuf.st_mode,
                                                                  stgLockBuf.st_uid, stgLockBuf.st_gid));
    if (outs == nullptr) {
        BKG_LOG_ERROR("create real file(" << realFilePath.c_str() << ") on UFS failed(" << errno << " : " <<
            strerror(errno) << ")");
        fileLock->Unlock();
        return false;
    }

    BKG_LOG_INFO("execute task(" << taskId << ") path(" << trace.path.c_str() << ") inode(" << trace.inode <<
        ") start");

    auto ok = RealBackupFile(taskId, trace, view);
    BKG_LOG_INFO("execute task(" << taskId << ") path(" << trace.path.c_str() << ") inode(" << trace.inode <<
        ") finished(" << ok << ")");

    UnLockStg(fileLock, stagePath);

    return ok;
}

bool BackupTarget::DoBackupFile(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view) noexcept
{
    auto realFilePath = trace.path;
    auto stagePath = trace.path;
    stagePath.append(".m.stg");

    struct stat stgLockBuf {};

    auto initiator = BackupFileManager::GetInstance().GetInitiator(sourceName);
    if (initiator == nullptr) {
        BKG_LOG_ERROR("Error: Initiator for source " << sourceName << " get failed");
        return false;
    }
    if (initiator->GetAttribute(taskId, realFilePath, stgLockBuf) != 0) {
        BKG_LOG_WARN("get file meta for path(" << realFilePath.c_str() << ") in memfs failed(" << errno << " : " <<
            strerror(errno) << ")");
        return false;
    }

    auto ret = StatFile(stagePath, stgLockBuf);
    if (ret != 0 && errno == ENOENT) {
        BKG_LOG_INFO("stg file(" << stagePath.c_str() << ") has been removed(" << errno << " : " << strerror(errno) <<
            ")the original file has been backed up.");
        return true;
    } else if (ret != 0) {
        BKG_LOG_WARN("get file meta for path(" << stagePath << ") failed(" << errno << " : " << strerror(errno) << ")");
        return false;
    }

    return DoRealBackupFile(taskId, trace, view, stgLockBuf);
}

bool BackupTarget::RealBackupAllParentDirectory(uint64_t taskId, const std::string &path,
    UnderFsFileView &view) noexcept
{
    struct stat buf {};
    std::vector<std::string> upNames;
    for (auto pos = path.find('/', 1); pos != std::string::npos; pos = path.find('/', pos + 1)) {
        upNames.emplace_back(path.substr(0, pos));
    }

    if (!path.empty() && path[path.size() - 1] != '/') {
        upNames.emplace_back(path);
    }

    for (auto &current : upNames) {
        if (current == "/") {
            continue;
        }

        auto initiator = BackupFileManager::GetInstance().GetInitiator(sourceName);
        if (initiator == nullptr) {
            BKG_LOG_ERROR("Error: Initiator for source " << sourceName << " get failed");
            return false;
        }
        auto ret = initiator->GetAttribute(taskId, current, buf);
        if (ret != 0) {
            BKG_LOG_ERROR("task(" << taskId << ") try to create parent(" << current.c_str() << ") stat failed(" <<
                errno << " : " << strerror(errno) << ")");
            return errno == ENOENT;
        }

        ufs::FileMeta fileMeta;
        ret = view.underFs->GetFileMeta(current, fileMeta);
        if (ret != 0 && errno != ENOENT) {
            BKG_LOG_ERROR("task(" << taskId << ") try to create parent(" << current.c_str() <<
                ") stat for UFS failed(" << errno << " : " << strerror(errno) << ")");
            return false;
        }

        bool success;
        if (ret != 0) { // directory not exist, create it
            success =
                CreateOneParent(taskId, current, ufs::FileMode{ buf.st_mode & 0777, buf.st_uid, buf.st_gid }, view);
        } else { // directory exist, correct metadata: owner, group, mode
            success = CorrectOneParent(taskId, current, buf, fileMeta, view);
        }

        if (!success) {
            return false;
        }
    }

    return true;
}

bool BackupTarget::RealBackupFile(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view) noexcept
{
    auto pos = trace.path.rfind('/');
    if (pos == std::string::npos) {
        BKG_LOG_ERROR("task(" << taskId << ") path(" << trace.path.c_str() << ") not right");
        return true;
    }

    if (pos > 0 && !RealBackupAllParentDirectory(taskId, trace.path.substr(0, pos), view)) {
        return false;
    }

    auto initiator = BackupFileManager::GetInstance().GetInitiator(sourceName);
    if (initiator == nullptr) {
        BKG_LOG_ERROR("Error: Initiator for source " << sourceName << " get failed");
        return false;
    }
    auto ret = initiator->MultiCopyFileToUfs(taskId, trace.path, view.underFs);
    if (ret < 0) {
        BKG_LOG_WARN("execute task(" << taskId << ") path(" << trace.path.c_str() << ") write ufs failed, retry.");
    }
    return ret == 0;
}

bool BackupTarget::CreateOneParent(uint64_t taskId, const std::string &path, const ufs::FileMode &mode,
    UnderFsFileView &view) noexcept
{
    auto ret = view.underFs->CreateDirectory(path, mode);
    if (ret != 0) {
        BKG_LOG_ERROR("task(" << taskId << ") try to create parent(" << path.c_str() << ") mkdir failed(" << errno <<
            " :" << strerror(errno) << ")");
        return false;
    }
    return true;
}

bool BackupTarget::CorrectOneParent(uint64_t taskId, const std::string &path, const struct stat &buf,
    const ufs::FileMeta &meta, UnderFsFileView &view) noexcept
{
    std::map<std::string, std::string> changeMeta;
    auto checker = [&changeMeta](const struct stat &buf, const ufs::FileMeta &meta) {
        auto pos = meta.meta.find("st_mode");
        if (pos == meta.meta.end() || pos->second != std::to_string(buf.st_mode)) {
            changeMeta["st_mode"] = std::to_string(buf.st_mode);
        }
    };

    checker(buf, meta);
    if (changeMeta.empty()) {
        return true;
    }

    auto ret = view.underFs->SetFileMeta(path, changeMeta);
    if (ret < 0) {
        BKG_LOG_ERROR("task(" << taskId << ") modify(" << path.c_str() << ") meta failed(" << errno << " : " <<
            strerror(errno) << ")");
        return false;
    }

    return true;
}

bool BackupTarget::DoMakeFileCache(uint64_t taskId, const FileTrace &trace, UnderFsFileView &view,
    const TaskInfo &taskInfo) noexcept
{
    struct stat buf {};

    BKG_LOG_INFO("execute task(" << taskId << ") path(" << trace.path.c_str() << ") inode(" << trace.inode <<
        ") start");
    auto initiator = BackupFileManager::GetInstance().GetInitiator(sourceName);
    if (initiator == nullptr) {
        BKG_LOG_ERROR("Error: Initiator for source " << sourceName << " get failed");
        return false;
    }
    auto ret = initiator->CompareFile(trace.path, trace.inode, buf);
    if (ret == FILE_NOT_EXIST) {
        BKG_LOG_INFO("task(" << taskId << ") path(" << trace.path.c_str() << ") inode(" << trace.inode <<
            ") file not exit, finished.");
        return true;
    }

    if (ret == IO_FAILED) {
        BKG_LOG_ERROR("task(" << taskId << ") path(" << trace.path.c_str() << ") inode(" << trace.inode <<
            ") file check failed!");
        return true;
    }
    if (ret == INODE_NOT_MATCH) {
        BKG_LOG_INFO("task(" << taskId << ") path(" << trace.path.c_str() << ") inode(" << trace.inode <<
            ") file changed!");
        return true;
    }

    auto taskResult = initiator->CopyFileToMemfs(taskId, trace.path, view.underFs, taskInfo);
    if (taskResult < 0) {
        BKG_LOG_WARN("execute task(" << taskId << ") path(" << trace.path.c_str() << ") write blocks failed, retry.");
    }

    taskResult = initiator->RecordToMemfsTaskResult(taskId, trace.path, taskResult, taskInfo);
    if (taskResult < 0) {
        BKG_LOG_WARN("execute task(" << taskId << ") path(" << trace.path.c_str() << ") inode(" << trace.inode <<
            ") failed.");
        return false;
    }

    return true;
}

void BackupTarget::MakeFileCache(const FileTrace &trace, const TaskInfo &taskInfo) noexcept
{
    ufs::FileMeta fileMeta;
    for (auto &ufsView : underFsFileView) {
        auto taskId = taskIdGen.fetch_add(1UL);
        BKG_LOG_INFO("submit task(" << taskId << ") file read file(" << trace.path.c_str() << ") inode(" <<
            trace.inode << ")");

        taskPool->Submit([taskId, trace, &ufsView, taskInfo, this]() -> bool {
            NotifyProcessMark marker(BackupFileManager::GetInstance().GetInitiator(sourceName).get());
            return DoMakeFileCache(taskId, trace, ufsView, taskInfo);
        });
    }
}
