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
#ifndef OCK_DFS_MEM_FILE_SYSTEM_H
#define OCK_DFS_MEM_FILE_SYSTEM_H

#include <sys/stat.h>
#include <sys/statvfs.h>
#include <atomic>
#include <memory>
#include <mutex>
#include <functional>
#include <unordered_map>

#include "mem_fs_constants.h"
#include "mem_fs_inode.h"
#include "inode_evictor.h"

namespace ock {
namespace memfs {
struct OpenedFile {
    std::mutex lock;
    // protect by lock;
    std::shared_ptr<MemFsInode> inode;
    bool allocatedFlag = false;
    bool readonlyMode = true;
    uint64_t offset;

    bool Initialize(const std::shared_ptr<MemFsInode> &ino, bool readonly) noexcept;
    bool Release() noexcept;
    void Lock() noexcept;
    void Unlock() noexcept;
};

struct RenameContext {
    const std::string source;
    const std::string target;
    const uint32_t flags;

    std::string sourceLastName;
    std::string targetLastName;

    uint64_t sourcePIno{ 0UL };
    uint64_t sourceIno{ 0UL };
    uint64_t targetPIno{ 0UL };
    uint64_t targetIno{ 0UL };

    std::shared_ptr<MemFsInode> sourceParent;
    std::shared_ptr<MemFsInode> sourceInode;
    std::shared_ptr<MemFsInode> targetParent;
    std::shared_ptr<MemFsInode> targetInode;

    RenameContext(std::string src, std::string tgt, uint32_t flg) noexcept
        : source{ std::move(src) }, target{ std::move(tgt) }, flags{ flg }
    {}
};

class MemFileSystem {
public:
    using InjectMkdirCheck =
        std::function<int(MemFileSystem *fs, const std::string &path, struct stat &statBuf, InodeAcl &acl)>;

    MemFileSystem(uint64_t blkSize, uint64_t blkCnt, std::string name) noexcept;
    ~MemFileSystem() noexcept;

    int Initialize() noexcept;
    void Destroy() noexcept;

    MemFsBMM &GetMemFsBMM() noexcept
    {
        return bmm;
    }

    uint64_t GetBlockSize() const noexcept
    {
        return blockSize;
    }

    uint64_t GetBlockCount() const noexcept
    {
        return blockCount;
    }

    void SetInjectMkdirCheck(const InjectMkdirCheck &checker) noexcept
    {
        if (checker != nullptr) {
            injectCheck = checker;
        }
    }

    bool Serviceable() const noexcept
    {
        return serviceable;
    }

    void Serviceable(bool state) noexcept
    {
        serviceable = state;
    }

public:
    int Open(const std::string &path, uint64_t &outInode) noexcept;
    bool Exist(const std::string &path) noexcept;
    int Create(const std::string &path, mode_t mode, uint64_t &outInode) noexcept;
    int LinkFile(const std::string &source, const std::string &target, uint64_t &outInode) noexcept;
    int Rename(const std::string &source, const std::string &target, uint32_t flags) noexcept;
    int SetBackupFinished(int fd) noexcept;
    int Close(int fd) noexcept;
    int MakeDirectory(const std::string &path, mode_t mode, uint64_t &outInode) noexcept;
    int RemoveDirectory(const std::string &path, uint64_t &outInode) noexcept;
    int RemoveFile(const std::string &path, uint64_t &outInode) noexcept;
    int ListDirectory(const std::string &path, std::vector<std::pair<std::string, InodeType>> &dnt) noexcept;
    int GetFileDataBlocks(int fd, std::vector<uint64_t> &blocks) noexcept;
    int AppendFileDataBlock(int fd, uint64_t blockId) noexcept;
    int AppendFileDataBlocks(int fd, const std::vector<uint64_t> &blocks) noexcept;
    int TruncateFile(int fd, uint64_t length) noexcept;

    int GetFileMeta(int fd, struct stat &metadata) noexcept;
    int GetMeta(const std::string &path, struct stat &metadata) noexcept;
    int GetMeta(const std::string &path, struct stat &metadata, InodeAcl &acl) noexcept;

    int Chmod(const std::string &path, mode_t mode) noexcept;
    int Chown(const std::string &path, uid_t uid, gid_t gid) noexcept;

    int GetFileSystemStat(struct statvfs &statBuf) noexcept;

public:
    static std::vector<std::string> GetPathNameTokens(const std::string &path) noexcept;

private:
    int GetInodeWithPathTemp(const std::string &path, uint64_t &pino, std::string &lastToken,
        uint64_t &outInode) noexcept;
    int OpenTruncate(uint64_t ino, uint64_t pino) noexcept;
    int AllocateFd() noexcept;
    void ReleaseFd(int fd) noexcept;
    OpenedFile *GetNewOpenedFile(int fd) noexcept;
    OpenedFile *GetAlreadyOpenedFile(int fd) noexcept;
    std::shared_ptr<MemFsInode> GetInode(uint64_t inode) noexcept;
    int GetInodeWithPath(const std::string &path, uint64_t &ino, uint64_t &pino, std::string &lastToken);
    int RemoveDentry(const std::string &path, InodeType type, int nonMatchError, uint64_t &oInode) noexcept;
    void GetInodeMetaInLock(const std::shared_ptr<MemFsInode> &inode, struct stat &metadata) noexcept;
    int PreCheckCreateDir(const std::string &path, const std::shared_ptr<MemFsInode> &parentInode, struct stat &statBuf,
        InodeAcl &inodeAcl) noexcept;
    std::shared_ptr<MemFsInode> GetInodeWithPermCheck(uint64_t ino, PermitType perm) noexcept;
    int PrepareRenameContext(RenameContext &context) noexcept;
    static int CheckRenameFlags(const RenameContext &context) noexcept;
    static int CheckRenameFlagsNone(const RenameContext &context) noexcept;
    static int CheckRenameFlagsExchange(const RenameContext &context) noexcept;
    static int CheckRenameFlagsNoReplace(const RenameContext &context) noexcept;
    static int RealRenameProcess(const RenameContext &ctx) noexcept;
    void PostRenameProcess(const RenameContext &ctx, int processResult) noexcept;
    std::pair<int, std::shared_ptr<MemFsInode>> CreateNewDirInner(const std::shared_ptr<MemFsInode> &parentInode,
        const std::string &name, const struct stat &statBuf, const InodeAcl &acl) noexcept;

private:
    bool ValidFd(int fd) const noexcept;

private:
    bool serviceable = true;
    int maxOpenFiles;
    uint64_t blockSize;
    uint64_t blockCount;
    const std::string fsName;
    std::atomic<uint64_t> inodeGenerator;
    std::mutex inodeMappingLock;
    std::unordered_map<uint64_t, std::shared_ptr<MemFsInode>> inodeMapping;
    OpenedFile *openedFiles;
    MemFsBMMOptions bmm_opt;
    MemFsBMM bmm;
    std::mutex freeFdLock;
    std::vector<int> freeFds;
    InjectMkdirCheck injectCheck;
    static MemFileSystem *mfsInstance;
    friend class InodeEvictor;
};
}
}


#endif // OCK_DFS_MEM_FILE_SYSTEM_H
