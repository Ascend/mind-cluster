/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2021-2021. All rights reserved.
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
#ifndef OCK_MEMFS_CORE_MEMFS_API_H
#define OCK_MEMFS_CORE_MEMFS_API_H
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/statvfs.h>

#include <cstdint>
#include <string>
#include <map>
#include <vector>
#include <functional>
#include <algorithm>
#include <mutex>
#include <condition_variable>

namespace ock {
namespace memfs {
struct FileOpNotify {
    std::function<int(int fd, const std::string &name, int flags, uint64_t inode)> openNotify;
    std::function<void(int fd, bool abnormal)> closeNotify;
    std::function<int(const std::string &name, uint64_t inode)> newFileNotify;
    std::function<int(const std::string &name, mode_t mode, uid_t owner, gid_t group)> mkdirNotify;
    std::function<void(const std::string &name, uint64_t inode)> unlinkNotify;
    std::function<int(const std::string &name)> preloadFileNotify;
    std::function<bool()> bgTaskEmptyNotify;

    FileOpNotify() noexcept
    {
        openNotify = [](int fd, const std::string &name, int flags, int64_t inode) -> int { return 0; };
        closeNotify = [](int fd, bool abnormal) {};
        newFileNotify = [](const std::string &name, uint64_t inode) { return 0; };
        mkdirNotify = [](const std::string &name, mode_t mode, uid_t owner, gid_t group) -> int { return 0; };
        unlinkNotify = [](const std::string &name, uint64_t inode) {};
        preloadFileNotify = [](const std::string &name) { return 0; };
        bgTaskEmptyNotify = []() { return true; };
    }
};

struct MemfsFileAcl {
    uint16_t ownerPerm{ 0 };
    uint16_t groupPerm{ 0 };
    uint16_t otherPerm{ 0 };
    uint16_t permMask{ 0 };
    std::map<uid_t, uint16_t> usersAcl;
    std::map<gid_t, uint16_t> groupsAcl;

    inline bool Empty() const noexcept
    {
        return usersAcl.empty() && groupsAcl.empty();
    }
};

using ExternalStat = std::function<int(const std::string &, struct stat &, MemfsFileAcl &)>;

struct PreloadProgressView {
    static bool PathExist(const std::string &path)
    {
        std::unique_lock<std::mutex> lock(gViewMutex);
        return std::find(g_loadPathVec.begin(), g_loadPathVec.end(), path) != g_loadPathVec.end();
    }

    static void InsertPath(const std::string &path)
    {
        std::unique_lock<std::mutex> lock(gViewMutex);
        auto pos = std::find(g_loadPathVec.begin(), g_loadPathVec.end(), path);
        if (pos == g_loadPathVec.end()) {
            g_loadPathVec.emplace_back(path);
        }
    }

    static void RemovePath(const std::string &path)
    {
        std::unique_lock<std::mutex> lock(gViewMutex);
        auto pos = std::find(g_loadPathVec.begin(), g_loadPathVec.end(), path);
        if (pos != g_loadPathVec.end()) {
            g_loadPathVec.erase(pos);
        }
        gCond.notify_one();
    }

    static void Wait(uint8_t seconds, const std::string &path)
    {
        std::unique_lock<std::mutex> lock(gViewMutex);
        gCond.wait_for(lock, std::chrono::seconds(seconds), [&] {
            return std::find(g_loadPathVec.begin(), g_loadPathVec.end(), path) == g_loadPathVec.end();
        });
    }

private:
    static std::vector<std::string> g_loadPathVec;
    static std::mutex gViewMutex;
    static std::condition_variable gCond;
};

class MemFsApi {
public:
    /* *
     * 初始化 MemFs
     * @return 成功返回0，失败返回-1
     */
    static int Initialize() noexcept;

    /* *
     * @brief 停止 MemFs
     */
    static void Destroy() noexcept;

    /* *
     * @brief 获取共享内存fd
     * @return 共享内存fd，失败返回-1
     */
    static int GetShareMemoryFd() noexcept;

    /* *
     * @brief 检查后台任务数是否为零
     * @return 任务数为零返回true，否则返回false
     */
    static bool BackgroundTaskEmpty() noexcept;

    /* *
     * @brief 注册操作回调
     * @param notify 回调集
     * @return 成功返回0，失败返回-1
     */
    static int RegisterFileOpNotify(const FileOpNotify &notify) noexcept;

    /* *
     * 设置外部查询元数据接口，在权限检查时用到
     * @param externalStat 查询元数据接口
     */
    static void SetExternalStat(const ExternalStat &externalStat) noexcept;

    /* *
     * @brief 打开一个文件获取fd，仅支持两种打开
     * (1) 写方式打开，不存在则创建，存在则长度截断为0，flags要传入O_CREAT | O_TRUNC | O_WRONLY，mode传入文件权限
     * (2) 文件已存在，只读打开，flags要传入O_RDONLY，mode不必传
     * @param path 文件路径
     * @param flags 要么传入(O_CREAT | O_TRUNC | O_WRONLY)，要么以传入(O_RDONLY)
     * @param mode 只有在新创建时有效
     * @return 成功时得到句柄 >= 0，失败返回-1，errno设置为错误码
     */
    static int OpenFile(const std::string &path, int flags, mode_t mode = 0644) noexcept;

    /* *
     * @brief 创建并以写的方式打开文件
     * @param path 文件路径
     * @param mode 只有在新创建时有效
     * @param inodeNum 文件 inode 信息
     * @return 成功时得到句柄 >= 0，失败返回-1，errno设置为错误码
     */
    static int CreateAndOpenFile(const std::string &path, uint64_t &inodeNum, mode_t mode = 0644) noexcept;

    /* *
     * 关闭文件
     * @param fd 文件描述符
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int CloseFile(int fd) noexcept;

    /* *
     * 丢弃写过程中操作失败的内存文件
     * @param path 文件路径
     * @param fd 文件描述符
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int DiscardFile(const std::string &path, int fd) noexcept;

    /* *
     * 设置已备份完成标记
     * @param fd 文件描述符
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int SetBackupFinished(int fd) noexcept;

    /* *
     * @brief 为文件追加申请数据块，写文件时使用
     * @param fd 文件描述符
     * @param blockId 追加得到的数据块
     * @param blockSize 数据块的size，单位bytes
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int AllocDataBlock(int fd, uint64_t &blockId, uint64_t &blockSize) noexcept;

    /* *
     * 向文件追加申请多个数据块，写文件时使用
     * @param fd 文件描述符
     * @param bytes 希望增加的数据块的总size，单位bytes
     * @param blocks 追加得到的数据块
     * @param blockSize 数据块的size，单位bytes
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int AllocDataBlocks(int fd, uint64_t bytes, std::vector<uint64_t> &blocks, uint64_t &blockSize) noexcept;

    /* *
     * 设置文件长度，在写完文件时调用
     * @param fd 文件描述符
     * @param length 文件长度，不能超过现有总块长
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int TruncateFile(int fd, uint64_t length) noexcept;

    /* *
     * @brief 根据fd获取文件元数据，参见fstat
     * @param fd 文件描述符
     * @param statBuf 文件元数据
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int GetFileMeta(int fd, struct stat &statBuf) noexcept;

    /* *
     * @brief 读取文件的全部数据块信息
     * @param fd 文件描述符
     * @param blocks 文件数据块
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int GetFileBlocks(int fd, std::vector<uint64_t> &blocks) noexcept;

    /* *
     * 根据数据块的id转化为地址，在读文件的时候有用
     * @param blockId 数据块
     * @return 返回数据块地址，无效数据块返回nullptr
     */
    static void *BlockToAddress(uint64_t blockId) noexcept;

    /* *
     * @brief 根据 block id查询在共享内存里的偏移
     * @param blockId block id
     * @return 偏移
     */
    static uint64_t GetBlockOffset(uint64_t blockId) noexcept;

    /* *
     * @brief 创建一个目录
     * @param path 目录的全路径
     * @param mode 目录的权限
     * @param recursive 是否递归创建
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int CreateDirectory(const std::string &path, mode_t mode, bool recursive = false) noexcept;

    /* *
     * @brief 创建目录，如果父目录不存在，一起创建；如果要创建的目录已存在，也不报错
     * @param path 要创建的目录路径
     * @param mode 权限
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     * @note
     * 如果父目录某一层级正在删除，会失败 errno = EAGAIN
     * 如果父目录某一层级存在一个同名文件，会失败 errno = ENOTDIR
     */
    static int CreateDirectoryWithParents(const std::string &path, mode_t mode) noexcept;

    /* *
     * @brief 删除一个目录
     * @param path 目录的全路径
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int RemoveDirectory(const std::string &path) noexcept;

    /* *
     * 列举一个目录中的全部目录项，自动过滤掉"."和".."
     * @param path 要列举的目录路径
     * @param entries 目录项数据，包含名称以及是否为文件（true表示文件, false表示目录)
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int ReadDirectory(const std::string &path, std::vector<std::pair<std::string, bool>> &entries) noexcept;

    /* *
     * @brief 根据路径查询元数据，可以查文件或目录，参见stat
     * @param path 文件或目录的路径
     * @param statBuf 文件或目录元数据
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int GetMeta(const std::string &path, struct stat &statBuf) noexcept;

    /* *
     * @brief 根据路径查询ACL，可以查文件或目录
     * @param path 文件或目录的路径
     * @param statBuf 文件或目录元数据
     * @param acl 文件或目录ACL
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int GetMetaAcl(const std::string &path, struct stat &statBuf, MemfsFileAcl &acl) noexcept;

    /* *
     * @brief 创建一个硬链接
     * @param oldPath 旧名称
     * @param newPath 新名称
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int Link(const std::string &oldPath, const std::string &newPath) noexcept;

    /* *
     * @brief 修改名称(move)
     * @param oldPath 旧名称
     * @param newPath 新名称
     * @param flags 控制标识，默认为0，其它值参见具体描述
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     * @details
     * (1) 修改文件或目录名称，如有必要，会在目录间移动;
     * (2) 对于文件，如果有其它硬连接指向oldPath，不受影响; 如果已打开oldPath的句柄，不受影响;
     * (3) 如果newPath已存在：
     * - flags 为 0 时，如果是文件直接替换；如果是目录，newPath不存在或空目录，则成功，否则失败，错误码ENOTEMPTY
     * - flags 为 RENAME_EXCHANGE 时，原子交换双方
     * - flags 为 RENAME_NOREPLACE 返回错误，错误码是 EEXIST
     * - flags 为 RENAME_FORCE 时，直接替换（无论文件或目录）
     */
    static int Rename(const std::string &oldPath, const std::string &newPath, uint32_t flags = 0U) noexcept;

    /* *
     * 根据路径删除一个文件
     * @param path 文件路径
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int Unlink(const std::string &path) noexcept;

    /* *
     * 修改文件或目录权限
     * @param path 文件或目录路径
     * @param mode 新权限
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int Chmod(const std::string &path, mode_t mode) noexcept;

    /* *
     * 修改文件或目录owner和group
     * @param path 文件或目录路径
     * @param uid 新的owner
     * @param gid 新的group
     * @return 成功时返回0，失败返回-1，errno设置为错误码
     */
    static int Chown(const std::string &path, uid_t uid, gid_t gid) noexcept;

    /* *
     * @brief get file system statistics
     * The 'f_frsize', 'f_favail', 'f_fsid' and 'f_flag' fields are ignored
     * Replaced 'struct statfs' parameter with 'struct statvfs' in
     */
    static int GetFileSystemStat(struct statvfs &statBuf) noexcept;

    /* *
     * @brief 设置服务状态
     * @param state 新的服务状态
     */
    static void Serviceable(bool state) noexcept;

    /* *
     * @brief 获取服务状态
     * @return 可服务时返回true，否则返回false
     */
    static bool Serviceable() noexcept;

    /* *
     * @brief 获取共享文件的块大小与块个数
     * 使用时，先判断Serviceable
     */
    static void GetShareFileCfg(uint64_t &blockSize, uint64_t &blockCnt) noexcept;

    /* *
     * @brief 从底层存储预加载文件到内存
     */
    static int PreloadFile(const std::string &path) noexcept;

private:
    /*
     * 已存在不报错
     */
    static int CreateOneLevelDirectory(const std::string &path, mode_t mode) noexcept;
};
}
}

#endif // OCK_MEMFS_CORE_MEMFS_API_H
