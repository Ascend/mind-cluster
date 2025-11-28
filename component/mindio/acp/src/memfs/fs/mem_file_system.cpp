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
#include "memfs_str_util.h"
#include "service_configure.h"
#include "memfs_logger.h"
#include "mem_fs_constants.h"
#include "mem_fs_state.h"
#include "inode_permission.h"
#include "mem_file_system.h"

using namespace ock::common;
using namespace ock::common::config;
using namespace ock::memfs;

static constexpr auto MAX_INODE_COUNT = 4UL * 1024UL * 1024UL;
static constexpr auto MAX_FILE_NAME_LEN = 255;
static constexpr auto DIR_BLOCK_SIZE = 4096;
static constexpr auto DIR_BLOCK_COUNT = 8;
static constexpr auto STAT_FS_BLOCK_SIZE = 512;
static constexpr mode_t ACC_MODE_MASK = 0777;
static constexpr auto DIR_LINK_COUNT = 2;

MemFileSystem *MemFileSystem::mfsInstance{ nullptr };

bool OpenedFile::Initialize(const std::shared_ptr<MemFsInode> &ino, bool readonly) noexcept
{
    if (allocatedFlag) {
        return false;
    }

    readonlyMode = readonly;
    allocatedFlag = true;
    inode = ino;
    offset = 0;
    return true;
}
bool OpenedFile::Release() noexcept
{
    if (!allocatedFlag) {
        return false;
    }
    inode.reset();
    allocatedFlag = false;
    return true;
}

void OpenedFile::Lock() noexcept
{
    lock.lock();
}

void OpenedFile::Unlock() noexcept
{
    lock.unlock();
}

MemFileSystem::MemFileSystem(uint64_t blkSize, uint64_t blkCnt, std::string name) noexcept
    : maxOpenFiles{ ServiceConfigure::GetInstance().GetMemFsConfig().maxOpenFiles },
      blockSize{ blkSize },
      blockCount{ blkCnt },
      fsName{ std::move(name) },
      inodeGenerator{ MemFsConstants::ROOT_INODE_NUMBER + 1UL },
      openedFiles{ new (std::nothrow) OpenedFile[ServiceConfigure::GetInstance().GetMemFsConfig().maxOpenFiles] }
{
    bmm_opt.useDevShm = false;
    bmm_opt.blkCount = blkCnt;
    bmm_opt.blkSize = blkSize;
    injectCheck = [](MemFileSystem *fs, const std::string &path, struct stat &statBuf, InodeAcl &acl) -> int {
        return -ENOENT;
    };
}

MemFileSystem::~MemFileSystem() noexcept
{
    if (openedFiles != nullptr) {
        delete[] openedFiles;
        openedFiles = nullptr;
    }
}

int MemFileSystem::Initialize() noexcept
{
    if (blockSize == 0UL || blockCount == 0UL) {
        MFS_LOG_ERROR("configure invalid, blockSize: " << blockSize << ", blockCount: " << blockCount);
        return -1;
    }

    if (openedFiles == nullptr) {
        MFS_LOG_ERROR("open files allocate failed.");
        return -1;
    }

    auto ret = bmm.Initialize(bmm_opt);
    if (ret != MCode::MFS_OK) {
        MFS_LOG_ERROR("initialize mfs(" << fsName.c_str() << ") bmm failed(" << ret << ")");
        return -1;
    }
    MemfsState::Instance().SetState(MemfsStateCode::STARTING, MemfsStartProgress::FIFTY_PERCENT);

    auto ignInode = MemFsConstants::INODE_INVALID;
    auto rootInode = MemFsConstants::ROOT_INODE_NUMBER;
    auto root = std::make_shared<MemFsInode>(ignInode, rootInode, "/", InodeType::INODE_DIR, 0755, bmm);
    if (!root->Valid()) {
        MFS_LOG_ERROR("create new root inode invalid.");
        bmm.UnInitialize();
        return -1;
    }

    root->uid = getuid();
    root->gid = getgid();
    root->blockSize = DIR_BLOCK_SIZE;
    inodeMapping[rootInode] = root;
    for (int i = 0; i < maxOpenFiles; ++i) {
        freeFds.push_back(i);
    }

    mfsInstance = this;
    return 0;
}

void MemFileSystem::Destroy() noexcept
{
    inodeMapping.clear();
    bmm.UnInitialize();
    mfsInstance = nullptr;
}

int MemFileSystem::Open(const std::string &path, uint64_t &outInode) noexcept
{
    uint64_t ino;
    uint64_t pino;
    std::string lastToken;
    auto errorCode = GetInodeWithPath(path, ino, pino, lastToken);
    if (errorCode != 0) {
        errno = -errorCode;
        return -1;
    }

    auto inode = GetInodeWithPermCheck(ino, PermitType::PERM_READ);
    if (inode == nullptr) {
        return -1;
    }

    auto fd = AllocateFd();
    if (fd < 0) {
        errno = EMFILE;
        return -1;
    }

    bool currInodeWriting = false;
    RwLockGuard lockGuard{ inode->inodeLock, false };
    if (inode->writing) {
        currInodeWriting = true;
    } else {
        inode->openCount++;
    }
    lockGuard.Unlock();

    if (currInodeWriting) {
        LOG_WARN("src inode is link writing open ignore, path: " << path);
        errno = EBUSY;
    }

    auto file = GetNewOpenedFile(fd);
    if (file == nullptr) {
        errno = EIO;
        LOG_ERROR("new opened file(fd-" << fd << ") is null");
        return -1;
    }
    file->Initialize(inode, true);
    outInode = file->inode->inode;
    file->Unlock();

    return fd;
}

bool MemFileSystem::Exist(const std::string &path) noexcept
{
    uint64_t ino;
    uint64_t pino;
    std::string lastToken;
    auto ret = GetInodeWithPath(path, ino, pino, lastToken);
    if (ret == 0) {
        return true;
    }
    return false;
}

int MemFileSystem::GetInodeWithPathTemp(const std::string &path, uint64_t &pino, std::string &lastToken,
    uint64_t &outInode) noexcept
{
    uint64_t ino;
    auto errorCode = GetInodeWithPath(path, ino, pino, lastToken);
    if (errorCode == 0) {
        outInode = ino;
        return OpenTruncate(ino, pino);
    }

    if (errorCode != -ENOENT) {
        errno = -errorCode;
        return -1;
    }

    return errorCode;
}

int MemFileSystem::Create(const std::string &path, mode_t mode, uint64_t &outInode) noexcept
{
    uint64_t pino;
    std::string lastToken;
    int resultCode = GetInodeWithPathTemp(path, pino, lastToken, outInode);
    if (resultCode != -ENOENT) {
        return resultCode;
    }

    auto pInode = GetInodeWithPermCheck(pino, PermitType::PERM_WRITE);
    if (pInode == nullptr) {
        return -1;
    }

    auto fd = AllocateFd();
    if (fd < 0) {
        errno = EMFILE;
        return -1;
    }

    auto newInodeNum = inodeGenerator.fetch_add(1UL);
    std::shared_ptr<MemFsInode> newInode;
    {
        std::unique_lock<std::mutex> lk(inodeMappingLock);
        newInode = std::make_shared<MemFsInode>(pino, newInodeNum, lastToken, InodeType::INODE_REG, mode, bmm);
        if (!newInode->Valid()) {
            ReleaseFd(fd);
            errno = ENOMEM;
            return -1;
        }
        newInode->blockSize = blockSize;
        inodeMapping.insert({ newInodeNum, newInode });

        // insert dentry
        auto errNum = 0;
        if (!pInode->PutDentry(lastToken, newInodeNum, InodeType::INODE_REG, errNum)) {
            // insert fail
            inodeMapping.erase(newInodeNum);
            ReleaseFd(fd);
            errno = errNum;
            return -1;
        }
    }

    auto file = GetNewOpenedFile(fd);
    if (file == nullptr) {
        LOG_ERROR("new opened file(fd-" << fd << ") is null");
        return -1;
    }
    file->Initialize(newInode, false);
    outInode = file->inode->inode;
    file->inode->evictHelper.MoveToTail();
    file->Unlock();
    LOG_INFO("inode-life(ino-" << outInode << ") create parent: " << pino << ", path: " << path << ", fd: " << fd);

    return fd;
}

int MemFileSystem::LinkFile(const std::string &source, const std::string &target, uint64_t &outInode) noexcept
{
    uint64_t sourceInode;
    uint64_t sourcePInode;
    std::string lastToken;
    auto errorCode = GetInodeWithPath(source, sourceInode, sourcePInode, lastToken);
    if (errorCode != 0) {
        errno = -errorCode;
        return -1;
    }

    auto inode = GetInodeWithPermCheck(sourceInode, PermitType::PERM_READ);
    if (inode == nullptr) {
        return -1;
    }

    if (inode->GetType() != INODE_REG) {
        errno = EISDIR;
        return -1;
    }

    uint64_t targetInode;
    uint64_t targetPInode;
    errorCode = GetInodeWithPath(target, targetInode, targetPInode, lastToken);
    if (errorCode == 0) {
        errno = EEXIST;
        return -1;
    }

    auto pInode = GetInodeWithPermCheck(targetPInode, PermitType::PERM_WRITE);
    if (pInode == nullptr) {
        return -1;
    }
    if (errorCode != -ENOENT) {
        errno = -errorCode;
        return -1;
    }

    auto errNum = 0;
    if (!pInode->PutDentry(lastToken, inode->inode, InodeType::INODE_REG, errNum)) {
        errno = errNum;
        return -1;
    }

    inode->AddParent(targetPInode, lastToken);
    outInode = inode->inode;
    LOG_INFO("inode-life(ino-" << outInode << ") link to path: " << target);
    return 0;
}

/*
 * renames a file, moving it between directories if required.
 * Any other hard links to the file are unaffected.
 * Open file descriptors for source are also unaffected.
 *
 * If  target  already  exists, it will be atomically replaced, so that there is no point at which another process
 * attempting to access target will find it missing.
 *
 * If source and target are existing hard links referring to the same file, then rename() does nothing, and returns a
 * success status.
 *
 * If target exists but the operation fails for some reason, rename() guarantees to leave an instance of target in
 * place.
 *
 * source can specify a directory.  In this case, target must either not exist, or it must specify an empty directory.
 */
int MemFileSystem::Rename(const std::string &source, const std::string &target, uint32_t flags) noexcept
{
    RenameContext context{ source, target, flags };
    auto ret = PrepareRenameContext(context);
    if (ret != 0) {
        return ret;
    }

    if (context.sourceIno == context.targetIno) {
        return 0;
    }

    ret = CheckRenameFlags(context);
    if (ret != 0) {
        return ret;
    }

    auto result = RealRenameProcess(context);
    PostRenameProcess(context, result);

    return result;
}

int MemFileSystem::SetBackupFinished(int fd) noexcept
{
    auto file = GetAlreadyOpenedFile(fd);
    if (file == nullptr) {
        errno = EBADF;
        return -1;
    }

    __sync_fetch_and_or(&file->inode->backup, static_cast<uint8_t>(1));
    file->Unlock();
    return 0;
}

int MemFileSystem::Close(int fd) noexcept
{
    auto file = GetAlreadyOpenedFile(fd);
    if (file == nullptr) {
        errno = EBADF;
        return -1;
    }

    struct timespec now {
        0, 0
    };
    clock_gettime(CLOCK_REALTIME_COARSE, &now);

    RwLockGuard lockGuard{ file->inode->inodeLock, false };
    if (file->readonlyMode) {
        file->inode->accessTime = now;
        file->inode->openCount--;
    } else {
        file->inode->modifiedTime = now;
        file->inode->writing = false;
    }
    lockGuard.Unlock();

    auto allocated = file->Release();
    file->Unlock();
    if (!allocated) {
        errno = EBADF;
        return -1;
    }

    ReleaseFd(fd);
    return 0;
}

int MemFileSystem::MakeDirectory(const std::string &path, mode_t mode, uint64_t &outInode) noexcept
{
    uint64_t ino;
    uint64_t pino;
    std::string lastToken;
    auto errorCode = GetInodeWithPath(path, ino, pino, lastToken);
    if (errorCode != -ENOENT) {
        errno = errorCode == 0 ? EEXIST : -errorCode;
        return -1;
    }

    auto pInode = GetInode(pino);
    if (pInode == nullptr) {
        errno = ENOENT;
        return -1;
    }

    struct stat parentStat {};
    InodeAcl inodeAcl;
    parentStat.st_mode = mode;
    parentStat.st_uid = getuid();
    parentStat.st_gid = getgid();
    errorCode = PreCheckCreateDir(path, pInode, parentStat, inodeAcl);
    if (errorCode != 0) {
        errno = -errorCode;
        return -1;
    }

    auto createResult = this->CreateNewDirInner(pInode, lastToken, parentStat, inodeAcl);
    if (createResult.first != 0) {
        return -1;
    }

    createResult.second->evictHelper.MoveToTail();
    outInode = createResult.second->inode;
    LOG_INFO("inode-life(ino-" << outInode << ") mkdir parent: " << pino << ", path: " << path);
    return 0;
}

int MemFileSystem::RemoveDirectory(const std::string &path, uint64_t &outInode) noexcept
{
    return RemoveDentry(path, InodeType::INODE_DIR, ENOTDIR, outInode);
}

int MemFileSystem::RemoveFile(const std::string &path, uint64_t &outInode) noexcept
{
    return RemoveDentry(path, InodeType::INODE_REG, EISDIR, outInode);
}

int MemFileSystem::ListDirectory(const std::string &path, std::vector<std::pair<std::string, InodeType>> &dnt) noexcept
{
    dnt.clear();
    uint64_t ino;
    uint64_t pino;
    std::string lastToken;
    auto errorCode = GetInodeWithPath(path, ino, pino, lastToken);
    if (errorCode != 0) {
        errno = -errorCode;
        return -1;
    }

    auto currInode = GetInode(ino);
    if (currInode == nullptr) {
        errno = ENOENT;
        return -1;
    }

    if (currInode->GetType() != InodeType::INODE_DIR) {
        errno = ENOTDIR;
        return -1;
    }

    if (!currInode->GetPermInfo().ContainsPermission(PermitType::PERM_READ)) {
        errno = EPERM;
        return -1;
    }

    RwLockGuard lockGuard{ currInode->inodeLock, true };
    for (auto &x : (*currInode->entries)) {
        dnt.emplace_back(x.first, x.second.type);
    }

    return 0;
}

int MemFileSystem::GetFileDataBlocks(int fd, std::vector<uint64_t> &blocks) noexcept
{
    auto file = GetAlreadyOpenedFile(fd);
    if (file == nullptr) {
        errno = EBADF;
        return -1;
    }

    auto currInode = file->inode;
    if (currInode->GetType() != InodeType::INODE_REG) {
        file->Unlock();
        errno = EISDIR;
        return -1;
    }

    blocks.clear();
    RwLockGuard lockGuard{ currInode->inodeLock, true };
    blocks.reserve(currInode->blocks->size());
    for (auto &block : (*currInode->blocks)) {
        blocks.push_back(block);
    }
    lockGuard.Unlock();

    file->Unlock();
    return 0;
}

int MemFileSystem::AppendFileDataBlock(int fd, uint64_t blockId) noexcept
{
    auto file = GetAlreadyOpenedFile(fd);
    if (file == nullptr) {
        errno = EBADF;
        return -1;
    }

    auto currInode = file->inode;
    if (currInode->GetType() != InodeType::INODE_REG) {
        file->Unlock();
        errno = EISDIR;
        return -1;
    }

    RwLockGuard lockGuard{ currInode->inodeLock, false };
    currInode->blocks->push_back(blockId);
    lockGuard.Unlock();

    file->Unlock();
    return 0;
}

int MemFileSystem::AppendFileDataBlocks(int fd, const std::vector<uint64_t> &blocks) noexcept
{
    auto file = GetAlreadyOpenedFile(fd);
    if (file == nullptr) {
        errno = EBADF;
        return -1;
    }

    auto currInode = file->inode;
    if (currInode->GetType() != InodeType::INODE_REG) {
        file->Unlock();
        errno = EISDIR;
        return -1;
    }

    RwLockGuard lockGuard{ currInode->inodeLock, false };
    for (unsigned long block : blocks) {
        currInode->blocks->push_back(block);
    }
    lockGuard.Unlock();

    file->Unlock();
    return 0;
}

int MemFileSystem::TruncateFile(int fd, uint64_t length) noexcept
{
    auto file = GetAlreadyOpenedFile(fd);
    if (file == nullptr || file->readonlyMode) {
        errno = EBADF;
        return -1;
    }

    auto currInode = file->inode;
    if (currInode->GetType() != InodeType::INODE_REG) {
        file->Unlock();
        errno = EISDIR;
        return -1;
    }

    RwLockGuard lockGuard{ currInode->inodeLock, false };
    auto truncatedBlocks = (length + blockSize - 1UL) / blockSize;
    auto currentBlockCount = currInode->blocks->size();
    if (currentBlockCount >= truncatedBlocks) {
        for (auto i = truncatedBlocks; i < currentBlockCount; ++i) {
            bmm.ReleaseOne((*currInode->blocks)[i]);
        }

        currInode->blocks->resize(truncatedBlocks);
        currInode->fileSize = length;
    } else {
        currInode->fileSize = currentBlockCount * blockSize;
    }
    lockGuard.Unlock();

    file->Unlock();
    return 0;
}

int MemFileSystem::GetFileMeta(int fd, struct stat &metadata) noexcept
{
    auto file = GetAlreadyOpenedFile(fd);
    if (file == nullptr) {
        errno = EBADF;
        return -1;
    }

    auto currInode = file->inode;
    RwLockGuard lockGuard{ currInode->inodeLock, true };
    GetInodeMetaInLock(currInode, metadata);

    file->Unlock();
    return 0;
}

int MemFileSystem::GetMeta(const std::string &path, struct stat &metadata) noexcept
{
    InodeAcl acl;
    return GetMeta(path, metadata, acl);
}

int MemFileSystem::GetMeta(const std::string &path, struct stat &metadata, InodeAcl &acl) noexcept
{
    uint64_t ino;
    uint64_t pino;
    std::string lastToken;
    auto errorCode = GetInodeWithPath(path, ino, pino, lastToken);
    if (errorCode != 0) {
        errno = -errorCode;
        return -1;
    }

    auto currInode = GetInodeWithPermCheck(ino, PermitType::PERM_READ);
    if (currInode == nullptr) {
        return -1;
    }

    RwLockGuard lockGuard{ currInode->inodeLock, true };
    GetInodeMetaInLock(currInode, metadata);
    if (currInode->acl != nullptr) {
        acl = *currInode->acl;
    }

    return 0;
}

int MemFileSystem::Chmod(const std::string &path, mode_t mode) noexcept
{
    uint64_t ino;
    uint64_t pino;
    std::string lastToken;
    auto errorCode = GetInodeWithPath(path, ino, pino, lastToken);
    if (errorCode != 0) {
        errno = -errorCode;
        return -1;
    }

    auto currInode = GetInode(ino);
    if (currInode == nullptr) {
        errno = ENOENT;
        return -1;
    }

    struct timespec now {
        0, 0
    };
    clock_gettime(CLOCK_REALTIME_COARSE, &now);
    RwLockGuard lockGuard{ currInode->inodeLock, true };
    currInode->accessMode = (mode & ACC_MODE_MASK);
    currInode->changedTime = now;

    return 0;
}

int MemFileSystem::Chown(const std::string &path, uid_t uid, gid_t gid) noexcept
{
    uint64_t ino;
    uint64_t pino;
    std::string lastToken;
    auto errorCode = GetInodeWithPath(path, ino, pino, lastToken);
    if (errorCode != 0) {
        errno = -errorCode;
        return -1;
    }

    auto currInode = GetInode(ino);
    if (currInode == nullptr) {
        errno = ENOENT;
        return -1;
    }

    struct timespec now {
        0, 0
    };
    clock_gettime(CLOCK_REALTIME_COARSE, &now);
    RwLockGuard lockGuard{ currInode->inodeLock, true };
    currInode->uid = uid;
    currInode->gid = gid;
    currInode->changedTime = now;

    return 0;
}

int MemFileSystem::GetFileSystemStat(struct statvfs &statBuf) noexcept
{
    if (getuid() != 0) {
        errno = EPERM;
        return -1;
    }

    statBuf.f_bsize = blockSize;
    statBuf.f_frsize = blockSize;
    statBuf.f_blocks = blockCount;
    statBuf.f_bfree = bmm.GetBmmPool().GetBlkCountRemaining();
    statBuf.f_bavail = statBuf.f_bfree;

    {
        std::unique_lock<std::mutex> lk(inodeMappingLock);
        statBuf.f_ffree = MAX_INODE_COUNT - inodeMapping.size();
    }

    statBuf.f_files = MAX_INODE_COUNT;
    statBuf.f_favail = statBuf.f_ffree;
    statBuf.f_fsid = MemFsConstants::ROOT_INODE_NUMBER;
    statBuf.f_flag = ST_NODEV | ST_NODIRATIME | ST_NOEXEC | ST_RELATIME | ST_SYNCHRONOUS;
    statBuf.f_namemax = MAX_FILE_NAME_LEN;

    return 0;
}

int MemFileSystem::OpenTruncate(uint64_t ino, uint64_t pino) noexcept
{
    auto parentInode = GetInode(pino);
    if (parentInode == nullptr) {
        errno = ENOENT;
        return -1;
    }

    auto currInode = GetInode(ino);
    if (currInode == nullptr) {
        errno = ENOENT;
        return -1;
    }

    if (currInode->GetType() == InodeType::INODE_DIR) {
        errno = EISDIR;
        return -1;
    }

    if (!currInode->GetPermInfo().ContainsPermission(PermitType::PERM_WRITE)) {
        errno = EPERM;
        return -1;
    }

    auto fd = AllocateFd();
    if (fd < 0) {
        errno = EMFILE;
        return -1;
    }

    RwLockGuard lockGuard{ currInode->inodeLock, false };
    if (currInode->writing || currInode->openCount > 0) {
        lockGuard.Unlock();
        errno = EBUSY;
        ReleaseFd(fd);
        return -1;
    }

    currInode->writing = true;
    for (auto &block : *currInode->blocks) {
        bmm.ReleaseOne(block);
    }
    currInode->blockSize = blockSize;
    currInode->blocks->clear();
    currInode->fileSize = 0;
    __sync_fetch_and_and(&currInode->backup, static_cast<uint8_t>(0));
    lockGuard.Unlock();

    auto file = GetNewOpenedFile(fd);
    if (file == nullptr) {
        LOG_ERROR("new opened file(fd-" << fd << ") is null");
        return -1;
    }
    file->Initialize(currInode, false);
    file->Unlock();
    LOG_INFO("inode-life(ino-" << ino << ") truncate open with fd: " << fd);

    return fd;
}

int MemFileSystem::AllocateFd() noexcept
{
    std::unique_lock<std::mutex> lk(freeFdLock);
    if (freeFds.empty()) {
        MFS_LOG_ERROR("too many open files.");
        return -1;
    }

    auto fd = freeFds.back();
    freeFds.pop_back();
    return fd + MemFsConstants::OPEN_FILE_FD_START;
}

void MemFileSystem::ReleaseFd(int fd) noexcept
{
    if (!ValidFd(fd)) {
        MFS_LOG_ERROR("release invalid fd(" << fd << ").");
        return;
    }

    std::unique_lock<std::mutex> lk(freeFdLock);
    freeFds.push_back(fd - MemFsConstants::OPEN_FILE_FD_START);
}

OpenedFile *MemFileSystem::GetNewOpenedFile(int fd) noexcept
{
    if (!ValidFd(fd)) {
        return nullptr;
    }
    auto file = &openedFiles[fd - MemFsConstants::OPEN_FILE_FD_START];
    file->Lock();

    return file;
}

OpenedFile *MemFileSystem::GetAlreadyOpenedFile(int fd) noexcept
{
    auto file = GetNewOpenedFile(fd);
    if (file == nullptr) {
        return nullptr;
    }

    if (!file->allocatedFlag) {
        file->Unlock();
        return nullptr;
    }

    return file;
}

std::shared_ptr<MemFsInode> MemFileSystem::GetInode(uint64_t inode) noexcept
{
    if (inode == MemFsConstants::INODE_INVALID) {
        return nullptr;
    }

    std::unique_lock<std::mutex> lk(inodeMappingLock);
    auto pos = inodeMapping.find(inode);
    if (pos == inodeMapping.end()) {
        return nullptr;
    }

    return pos->second;
}

int MemFileSystem::GetInodeWithPath(const std::string &path, uint64_t &ino, uint64_t &pino, std::string &lastToken)
{
    // get token
    auto tokens = GetPathNameTokens(path);
    if (tokens.empty()) {
        lastToken = "";
        ino = MemFsConstants::ROOT_INODE_NUMBER;
        pino = MemFsConstants::INODE_INVALID;
        return 0;
    }

    lastToken = tokens.back();
    auto tokenCount = tokens.size();
    auto scanNode = GetInode(MemFsConstants::ROOT_INODE_NUMBER);

    int ret = -ENOENT;
    ino = MemFsConstants::INODE_INVALID;
    pino = MemFsConstants::INODE_INVALID;
    for (size_t i = 0; i < tokenCount; ++i) {
        if (scanNode == nullptr) {
            ret = -ENOENT;
            break;
        }
        if (scanNode->GetType() != InodeType::INODE_DIR) {
            ret = -ENOTDIR;
            break;
        }

        auto perm = scanNode->GetPermInfo();
        if (!perm.ContainsPermission(PermitType::PERM_EXECUTE)) {
            ret = -EPERM;
            break;
        }

        Dentry d;
        if (!scanNode->GetDentry(tokens[i], d)) {
            pino = (i + 1 == tokenCount) ? scanNode->GetInodeNumber() : MemFsConstants::INODE_INVALID;
            ret = -ENOENT;
            break;
        }

        if (i + 1 == tokenCount) {
            pino = scanNode->GetInodeNumber();
            ino = d.inode;
            ret = 0;
            break;
        }
        scanNode = GetInode(d.inode);
    }
    return ret;
}

int MemFileSystem::RemoveDentry(const std::string &path, InodeType type, int nonMatchError, uint64_t &oInode) noexcept
{
    uint64_t ino;
    uint64_t pino;
    std::string lastToken;
    auto errorCode = GetInodeWithPath(path, ino, pino, lastToken);
    if (errorCode != 0) {
        errno = -errorCode;
        return -1;
    }

    auto currInode = GetInode(ino);
    if (currInode == nullptr) {
        errno = ENOENT;
        return -1;
    }

    auto pInode = GetInodeWithPermCheck(pino, PermitType::PERM_WRITE);
    if (pInode == nullptr) {
        return -1;
    }

    if (currInode->GetType() != type) {
        errno = nonMatchError;
        return -1;
    }

    if (currInode->ActiveFile()) {
        LOG_WARN("file is active, cannot be removed, path: " << path);
        errno = EBUSY;
        return 0;
    }

    if (!currInode->TryRemove()) {
        errno = ENOTEMPTY;
        return -1;
    }

    // remove dentry
    Dentry d;
    if (!pInode->DeleteDentry(lastToken, d) || d.inode != ino) {
        errno = ENOENT;
        return -1;
    }

    uint64_t parentCount = 0UL;
    currInode->RemoveParent(pino, lastToken, parentCount);
    if (parentCount > 0UL) {
        return 0;
    }

    oInode = ino;
    std::unique_lock<std::mutex> lk(inodeMappingLock);
    inodeMapping.erase(ino);
    LOG_INFO("inode-life(ino-" << ino << ") remove parent: " << pino << ", path: " << path);
    return 0;
}

void MemFileSystem::GetInodeMetaInLock(const std::shared_ptr<MemFsInode> &inode, struct stat &metadata) noexcept
{
    metadata.st_uid = inode->uid;
    metadata.st_gid = inode->gid;
    metadata.st_mode = inode->accessMode;
    metadata.st_blksize = static_cast<int>(inode->blockSize);
    if (inode->GetType() == INODE_REG) {
        metadata.st_mode |= __S_IFREG;
        metadata.st_blocks = static_cast<int64_t>(inode->blocks->size());
        metadata.st_size = static_cast<int64_t>(inode->fileSize);
        metadata.st_blocks = metadata.st_blocks * metadata.st_blksize / STAT_FS_BLOCK_SIZE;
        metadata.st_nlink = inode->GetLinkCountInLock();
    } else {
        metadata.st_mode |= __S_IFDIR;
        metadata.st_size = DIR_BLOCK_SIZE;
        metadata.st_blocks = DIR_BLOCK_COUNT;
        metadata.st_nlink = DIR_LINK_COUNT;
    }
    metadata.st_mtim = inode->modifiedTime;
    metadata.st_atim = inode->accessTime;
    metadata.st_ctim = inode->changedTime;
    metadata.st_ino = inode->inode;
    metadata.st_dev = 0;
    metadata.st_rdev = 0;
}

int MemFileSystem::PreCheckCreateDir(const std::string &path, const std::shared_ptr<MemFsInode> &parentInode,
    struct stat &statBuf, InodeAcl &inodeAcl) noexcept
{
    auto errNum = injectCheck(this, path, statBuf, inodeAcl);
    if (errNum == -ENOENT) {
        auto perm = parentInode->GetPermInfo();
        if (!perm.ContainsPermission(PermitType::PERM_WRITE)) {
            return -EPERM;
        }
        return 0;
    }

    if (errNum != 0) {
        MFS_LOG_WARN("check stat for dir(" << path << ") failed(" << -errNum << " : " << strerror(-errNum) << ")");
        return errNum;
    }

    if (!S_ISDIR(statBuf.st_mode)) {
        MFS_LOG_WARN("creating dir(" << path << "), exist not a directory, mode(0" << statBuf.st_mode << ").");
        return -ENOTDIR;
    }

    auto localUid = getuid();
    auto localGid = getgid();
    if (statBuf.st_uid == localUid && statBuf.st_gid == localGid) {
        return 0;
    }

    MFS_LOG_INFO("creating dir(" << path << "), exist uid(" << statBuf.st_uid << ") gid(" << statBuf.st_gid <<
        "), but local uid(" << localUid << ") gid(" << localGid << ")");
    return 0;
}

std::shared_ptr<MemFsInode> MemFileSystem::GetInodeWithPermCheck(uint64_t ino, ock::memfs::PermitType perm) noexcept
{
    auto inode = GetInode(ino);
    if (inode == nullptr) {
        errno = ENOENT;
        return nullptr;
    }

    if (!inode->GetPermInfo().ContainsPermission(perm)) {
        errno = EPERM;
        return nullptr;
    }

    return inode;
}

int MemFileSystem::PrepareRenameContext(ock::memfs::RenameContext &context) noexcept
{
    /*
     * 获取source信息，source必须严格存在，父目录和自身都要存在
     */
    auto errorNum = GetInodeWithPath(context.source, context.sourceIno, context.sourcePIno, context.sourceLastName);
    if (errorNum != 0) {
        errno = -errorNum;
        LOG_WARN("get rename source: " << context.source << " failed: " << errno << ":" << strerror(errno));
        return -1;
    }

    /*
     * 获取source父目录，需要检查写权限
     */
    if ((context.sourceParent = GetInodeWithPermCheck(context.sourcePIno, PermitType::PERM_WRITE)) == nullptr) {
        LOG_WARN("get rename source: " << context.source << " parent failed: " << errno << ":" << strerror(errno));
        return -1;
    }

    if ((context.sourceInode = GetInode(context.sourceIno)) == nullptr) {
        LOG_WARN("rename source: " << context.source << " not exist.");
        errno = ENOENT;
        return -1;
    }

    /*
     * 获取target信息，target父目录必须存在，target目录自身可以不存在
     */
    errorNum = GetInodeWithPath(context.target, context.targetIno, context.targetPIno, context.targetLastName);
    if (errorNum != 0 && errorNum != -ENOENT) {
        errno = -errorNum;
        LOG_WARN("get rename target: " << context.target << " failed: " << errno << ":" << strerror(errno));
        return -1;
    }

    /*
     * 获取target父目录，需要检查写权限
     */
    if (context.targetPIno == context.sourcePIno) {
        context.targetParent = context.sourceParent;
    } else if ((context.targetParent = GetInodeWithPermCheck(context.targetPIno, PermitType::PERM_WRITE)) == nullptr) {
        LOG_WARN("get rename target: " << context.target << " parent failed: " << errno << ":" << strerror(errno));
        return -1;
    }

    context.targetInode = GetInode(context.targetIno);
    return 0;
}

int MemFileSystem::CheckRenameFlags(const ock::memfs::RenameContext &context) noexcept
{
    if (context.flags == MemFsConstants::RENAME_FLAG_EXCHANGE) {
        return CheckRenameFlagsExchange(context);
    }

    if (context.flags == MemFsConstants::RENAME_FLAG_NOREPLACE) {
        return CheckRenameFlagsNoReplace(context);
    }

    if (context.flags == MemFsConstants::RENAME_FLAG_FORCE) {
        return 0;
    }

    return CheckRenameFlagsNone(context);
}

int MemFileSystem::CheckRenameFlagsNone(const ock::memfs::RenameContext &context) noexcept
{
    if (context.targetInode == nullptr) {
        return 0;
    }

    if (context.sourceInode->type != context.targetInode->type) {
        errno = context.targetInode->type == InodeType::INODE_DIR ? EISDIR : ENOTDIR;
        return -1;
    }

    if (context.sourceInode->type == InodeType::INODE_REG) {
        return 0;
    }

    if (!context.targetInode->TryRemove()) {
        errno = ENOTEMPTY;
        return -1;
    }

    return 0;
}

int MemFileSystem::CheckRenameFlagsExchange(const ock::memfs::RenameContext &context) noexcept
{
    if (context.targetInode == nullptr) {
        errno = ENOENT;
        return -1;
    }

    return 0;
}

int MemFileSystem::CheckRenameFlagsNoReplace(const ock::memfs::RenameContext &context) noexcept
{
    if (context.targetInode != nullptr) {
        errno = EEXIST;
        return -1;
    }

    return 0;
}

int MemFileSystem::RealRenameProcess(const ock::memfs::RenameContext &ctx) noexcept
{
    auto errorCode = 0;
    if (ctx.flags == MemFsConstants::RENAME_FLAG_EXCHANGE) {
        if (!ctx.sourceParent->ExchangeDentry(ctx.sourceLastName, ctx.targetParent, ctx.targetLastName, errorCode)) {
            MFS_LOG_WARN("exchange(" << ctx.source << ") to(" << ctx.target << ") at last failed.");
            errno = errorCode;
            return -1;
        }

        ctx.targetInode->RemoveParent(ctx.targetPIno, ctx.targetLastName);
        ctx.targetInode->AddParent(ctx.sourcePIno, ctx.sourceLastName);
    } else {
        if (!ctx.sourceParent->RenameDentry(ctx.sourceLastName, ctx.targetParent, ctx.targetLastName, errorCode)) {
            MFS_LOG_WARN("rename (" << ctx.source << ") to(" << ctx.target << ") at last failed.");
            errno = errorCode;
            return -1;
        }
    }

    ctx.sourceInode->RemoveParent(ctx.sourcePIno, ctx.sourceLastName);
    ctx.sourceInode->AddParent(ctx.targetPIno, ctx.targetLastName);
    return 0;
}

void MemFileSystem::PostRenameProcess(const ock::memfs::RenameContext &ctx, int processResult) noexcept
{
    if (processResult != 0) {
        return;
    }

    LOG_INFO("inode-life(ino-" << ctx.sourceInode << ") rename to: " << ctx.target);
    if (ctx.flags == MemFsConstants::RENAME_FLAG_EXCHANGE) {
        return;
    }

    if (ctx.targetInode == nullptr) {
        return;
    }

    std::list<std::shared_ptr<MemFsInode>> indirectInodes;
    std::list<uint64_t> directInodes;
    if (ctx.targetInode->GetType() == InodeType::INODE_DIR) {
        indirectInodes.push_back(ctx.targetInode);
    } else {
        directInodes.push_back(ctx.targetIno);
    }

    while (!indirectInodes.empty()) {
        auto inode = indirectInodes.front();
        indirectInodes.pop_front();
        for (const auto &child : *inode->entries) {
            if (child.second.type != InodeType::INODE_DIR) {
                directInodes.push_back(child.second.inode);
                continue;
            }

            auto childNode = GetInode(child.second.inode);
            if (childNode == nullptr) {
                continue;
            }

            indirectInodes.push_back(childNode);
        }
        directInodes.push_back(inode->inode);
    }

    for (auto ino : directInodes) {
        LOG_INFO("inode-life(ino-" << ino << ") removed by rename replaced path: " << ctx.target);
    }

    std::unique_lock<std::mutex> lockGuard{ inodeMappingLock };
    for (auto ino : directInodes) {
        inodeMapping.erase(ino);
    }
}

bool MemFileSystem::ValidFd(int fd) const noexcept
{
    if (fd < MemFsConstants::OPEN_FILE_FD_START) {
        return false;
    }

    if (fd >= MemFsConstants::OPEN_FILE_FD_START + maxOpenFiles) {
        return false;
    }

    return true;
}

std::vector<std::string> MemFileSystem::GetPathNameTokens(const std::string &path) noexcept
{
    std::vector<std::string> items;
    StrUtil::Split(path, "/", items);

    std::vector<std::string> result;
    std::for_each(items.begin(), items.end(), [&result](std::string &item) {
        if (!item.empty()) {
            result.push_back(std::move(item));
        }
    });
    return std::move(result);
}

std::pair<int, std::shared_ptr<MemFsInode>> MemFileSystem::CreateNewDirInner(
    const std::shared_ptr<MemFsInode> &parentInode, const std::string &name, const struct stat &statBuf,
    const ock::memfs::InodeAcl &acl) noexcept
{
    std::shared_ptr<MemFsInode> newInode;
    auto pino = parentInode->inode;
    auto newInodeNum = inodeGenerator.fetch_add(1UL);

    std::unique_lock<std::mutex> lk(inodeMappingLock);
    newInode = std::make_shared<MemFsInode>(pino, newInodeNum, name, InodeType::INODE_DIR, statBuf.st_mode, bmm);
    if (newInode == nullptr || !newInode->Valid()) {
        errno = ENOMEM;
        return std::make_pair(-1, nullptr);
    }

    newInode->blockSize = DIR_BLOCK_SIZE;
    if (!acl.Empty()) {
        newInode->acl = new (std::nothrow) InodeAcl(acl);
        if (newInode->acl == nullptr) {
            errno = ENOMEM;
            return std::make_pair(-1, nullptr);
        }
    }

    inodeMapping.insert({ newInodeNum, newInode });

    // insert dentry
    auto errNum = 0;
    if (!parentInode->PutDentry(name, newInodeNum, InodeType::INODE_DIR, errNum)) {
        // insert fail
        inodeMapping.erase(newInodeNum);
        errno = errNum;
        return std::make_pair(-1, nullptr);
    }

    return std::make_pair(0, newInode);
}