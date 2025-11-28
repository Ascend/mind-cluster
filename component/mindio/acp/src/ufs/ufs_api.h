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
#ifndef OCK_DFS_UFS_API_H
#define OCK_DFS_UFS_API_H

#include <sys/time.h>
#include <cstdint>
#include <limits>
#include <list>
#include <string>
#include <map>
#include <memory>
#include <utility>

#include "byte_buffer.h"

namespace ock {
namespace ufs {
class Closable {
public:
    virtual ~Closable() = default;

public:
    virtual int Close() noexcept = 0;
};

class InputStream : public Closable {
public:
    ~InputStream() override = default;

public:
    /* *
     * 此流总字节数
     * @return 返回总字节数，无法确定则返回-1
     */
    virtual int64_t TotalSize() noexcept;

    /* *
     * 从InputStream中读取下一个字节
     *
     * @param byte [out] 读取到的一个字节
     *
     * @return
     * 1 读取成功
     * 0 示读到数据，到流结尾
     * -1 读取失败，error指定错误码
     */
    virtual int Read(uint8_t &byte) noexcept = 0;

    /* *
     * 从InputStream中读取最多count个字节的数据
     * @param buf 保存读取数据的buffer
     * @param count 最多读取个数
     * @return 成功读取个数，失败时返回-1，error指定错误码
     */
    virtual int64_t Read(uint8_t *buf, uint64_t count) noexcept;

    virtual int64_t Read(utils::ByteBuffer &buf) noexcept;
};

class NullInputStream : public InputStream {
public:
    NullInputStream() = default;
    ~NullInputStream() override = default;

public:
    int Read(uint8_t &byte) noexcept override
    {
        return 0;
    }

    int64_t Read(uint8_t *buf, uint64_t count) noexcept override
    {
        return 0L;
    }

    int Close() noexcept override
    {
        return 0;
    }
};

class OutputStream : public Closable {
public:
    ~OutputStream() override = default;

public:
    /* *
     * 向OutputStream中写入一个字节
     * @param byte 要写入的数据
     * @return
     * 1 写入成功
     * 0 未写入数据，到流结尾
     * -1 写入失败，error指定错误码
     */
    virtual int Write(uint8_t byte) noexcept = 0;

    /* *
     * 向OutputStream中写入最多count字节
     * @param buf 要写入的数据
     * @param count 要写入的长度
     * @return 写入成功字节数，失败返回-1，error指定错误原因
     */
    virtual int64_t Write(const uint8_t *buf, uint64_t count) noexcept;

    /* *
     * @brief 将输出流中的数据同步到存储设备，以保证发生故障不丢失数据
     * @return 成功返回0，失败返回-1，error指定错误原因
     */
    virtual int Sync() noexcept;
};

class NullOutputStream : public OutputStream {
public:
    NullOutputStream() = default;
    ~NullOutputStream() override = default;

public:
    int Write(uint8_t byte) noexcept override
    {
        return 0;
    }
    int64_t Write(const uint8_t *buf, uint64_t count) noexcept override
    {
        return 0L;
    }

    int Sync() noexcept override
    {
        return 0;
    }

    int Close() noexcept override
    {
        return 0;
    }
};

/**
 * @brief 列举文件列表时，用于分页显示
 */
class ListFilePageMarker {
public:
    virtual ~ListFilePageMarker() = default;

    /* *
     * @brief 表示是否列举完成
     * @return true表示完成，false表示未完成
     */
    virtual bool Finished() noexcept = 0;
};

class FinishedMarker : public ListFilePageMarker {
public:
    bool Finished() noexcept override
    {
        return true;
    }
};

/**
 * @brief 列举文件列表时，每一项的信息
 */
struct ListFileItem {
    const std::string name;
    const struct timespec mtime;
    const bool commonFile;
    explicit ListFileItem(std::string n, bool isFile = true) noexcept : ListFileItem{ std::move(n), { 0, 0 }, isFile }
    {}

    explicit ListFileItem(std::string n, const struct timespec mt, bool isFile = true) noexcept
        : name{ std::move(n) }, mtime{ mt.tv_sec, mt.tv_nsec }, commonFile{ isFile }
    {}
};

/**
 * @brief 列举文件列表时的返回信息，除了文件信息列表，还有分页信息，如果未完成可以传入分页信息再次列举
 */
struct ListFileResult {
    std::list<ListFileItem> files;
    std::shared_ptr<ListFilePageMarker> marker;
};

/**
 * @brief 读取文件时，指定范围读取
 */
struct FileRange {
    const uint64_t begin;
    const uint64_t count;
    const uint64_t fileTotalSize{ 0 };
    FileRange() noexcept : FileRange{ 0UL } {}
    explicit FileRange(uint64_t b) noexcept : FileRange{ b, std::numeric_limits<uint64_t>::max(), 0 } {}
    FileRange(uint64_t b, uint64_t c) noexcept : begin{ b }, count{ c }, fileTotalSize{ 0 } {}
    FileRange(uint64_t b, uint64_t c, uint64_t f) noexcept : begin{ b }, count{ c }, fileTotalSize{ f } {}
};

struct FileAcl {
    uint16_t ownerPerm{ 0 };
    uint16_t groupPerm{ 0 };
    uint16_t otherPerm{ 0 };
    uint16_t permMask{ 0 };
    std::map<uid_t, uint16_t> users;
    std::map<gid_t, uint16_t> groups;
};

/**
 * @brief 查询文件元数据时返回的信息
 */
struct FileMeta {
    std::string name;
    uint32_t mode{ 0U };
    uint64_t size{ 0UL };
    struct timespec mtime {
        0, 0
    };
    std::map<std::string, std::string> meta;
    FileAcl acl;
};

struct FileMode {
    uint32_t owner;
    uint32_t group;
    int64_t mode;
    std::string description;

    explicit FileMode(mode_t m) noexcept : owner{ 0U }, group{ 0U }, mode{ m } {}

    explicit FileMode(mode_t m, uint32_t u, uint32_t g) noexcept : owner{ u }, group{ g }, mode{ m } {}

    explicit FileMode(std::string desc) : owner{ 0 }, group{ 0 }, mode{ -1L }, description{ std::move(desc) } {}
};

/**
 * @brief 获取的一个文件锁对象
 */
class FileLock {
public:
    virtual ~FileLock() = default;

public:
    /* *
     * @brief 对一个文件上锁，如果锁被其它进程占用，阻塞直到上锁成功
     * @return 成功返回0，失败返回-1，error指定错误码
     * @details 本操作常见错误码
     * EINTR : 上锁阻塞被信号唤醒
     */
    virtual int Lock() noexcept = 0;

    /* *
     * @brief 尝试对一个文件上锁，如果锁被其它进程占用，立即返回-1，errno设置为 EACCES or EAGAIN
     * @return 成功返回0，失败返回-1，error指定错误码
     * @details 本操作常见错误码
     * EACCES or EAGAIN :　锁被其它进程占用
     */
    virtual int TryLock() noexcept = 0;

    /* *
     * @brief 解锁
     * @return 成功返回0，失败返回-1，error指定错误码
     */
    virtual int Unlock() noexcept = 0;
};

/**
 * @brief 基于的文件服务抽象
 */
class BaseFileService {
public:
    virtual ~BaseFileService() = default;

public:
    /* *
     * @brief 做健康检查
     * @return 健康返回true，否则返回false
     */
    virtual bool HealthyCheck() noexcept = 0;

    /* *
     * @brief 上传一个文件，用于要传的文件数据已完成准备好情况，通过flags指定已存在、不存在的处理
     * @param path 文件路径
     * @param flags 写文件的标志，可填：
     * - O_CREAT : 文件不存在在则创建，文件存在则Truncate
     * - O_CREAT | O_EXCL : 文件不存在在则创建，文件存在则报错
     * - 以上标记可以另带上O_SYNC或O_DSYNC
     * @param mode 文件权限
     * @param dataBuffer 文件数据
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限写入文件
     * EDQUOT : 配置空间不足
     * EINVAL : 参数无效
     * EISDIR : 同名路径存在一个文件夹
     * ENOENT : 父目录不存在
     * ENOENT : 没有带O_CREAT标志上传文件，但对应文件原本不存在
     * EEXIST : 同时O_CREAT|O_EXCL标记上传文件时，文件已存在
     * EOPNOTSUPP : 文件服务不支持此操作
     * EIO : 发生IO错误
     */
    virtual int PutFile(const std::string &path, int flags, const FileMode &mode,
        utils::ByteBuffer &dataBuffer) noexcept = 0;

    /* *
     * @brief 上传一个文件，用于要传的文件数据已完成准备好情况，如果文件已存在则覆盖
     * @param path 文件路径
     * @param mode 文件权限
     * @param dataBuffer 文件数据
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限写入文件
     * EDQUOT : 配置空间不足
     * EINVAL : 参数无效
     * EISDIR : 同名路径存在一个文件夹
     * ENOENT : 父目录不存在
     * EIO : 发生IO错误
     */
    virtual int PutFile(const std::string &path, const FileMode &mode, utils::ByteBuffer &dataBuffer) noexcept = 0;

    /* *
     * @brief 上传一个文件，用于要传的文件数据不能立即读取完成，需要边上传边准备的情况
     * @param path 文件路径
     * @param flags 写文件的标志，可填
     * - O_CREAT : 文件不存在在则创建，文件存在则Truncate
     * - O_CREAT | O_EXCL : 文件不存在在则创建，文件存在则报错
     * - 以上标记可以另带上O_SYNC或O_DSYNC
     * @param mode 文件权限
     * @param inputStream 文件数据流
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限写入文件
     * EDQUOT : 配置空间不足
     * EINVAL : 参数无效
     * EISDIR : 同名路径存在一个文件夹
     * ENOENT : 父目录不存在
     * EIO : 发生IO错误
     */
    virtual int PutFile(const std::string &path, int flags, const FileMode &mode,
        InputStream &inputStream) noexcept = 0;

    /* *
     * @brief 上传一个文件，用于要传的文件数据不能立即读取完成，需要边上传边准备的情况，如果文件已存在则覆盖
     * @param path 文件路径
     * @param mode 文件权限
     * @param inputStream 文件数据流
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限写入文件
     * EDQUOT : 配置空间不足
     * EINVAL : 参数无效
     * EISDIR : 同名路径存在一个文件夹
     * ENOENT : 父目录不存在
     * EIO : 发生IO错误
     */
    virtual int PutFile(const std::string &path, const FileMode &mode, InputStream &inputStream) noexcept = 0;

    /* *
     * @brief 上传一个文件，得到一个指定范围输出流用于写入文件指定范围内容
     * @param path 文件路径
     * @param flags 写文件的标志，可填
     * - O_CREAT : 文件不存在在则创建，文件存在则Truncate
     * - O_CREAT | O_EXCL : 文件不存在在则创建，文件存在则报错
     * - 以上标记可以另带上O_SYNC或O_DSYNC
     * @param mode 文件权限
     * @param range 读取范围
     * @return 输出流，失败返回nullptr，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限写入文件
     * EDQUOT : 配置空间不足
     * EINVAL : 参数无效
     * EISDIR : 同名路径存在一个文件夹
     * ENOENT : 父目录不存在
     */
    virtual std::shared_ptr<OutputStream> PutFile(const std::string &path, int flags, const FileMode &mode,
        const FileRange &range) noexcept = 0;

    /* *
     * @brief 上传一个文件，得到一个指定范围输出流用于写入文件指定范围内容
     * @param path 文件路径
     * @param mode 文件权限
     * @param range 读取范围
     * @return 输出流，失败返回nullptr，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限写入文件
     * EDQUOT : 配置空间不足
     * EINVAL : 参数无效
     * EISDIR : 同名路径存在一个文件夹
     * ENOENT : 父目录不存在
     */
    virtual std::shared_ptr<OutputStream> PutFile(const std::string &path, const FileMode &mode,
        const FileRange &range) noexcept = 0;
    /* *
     * @brief 上传一个文件，得到一个默认范围输出流用于写入文件指定范围内容
     * @param path 文件路径
     * @param mode 文件权限
     * @return 输出流，失败返回nullptr，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限写入文件
     * EDQUOT : 配置空间不足
     * EINVAL : 参数无效
     * EISDIR : 同名路径存在一个文件夹
     * ENOENT : 父目录不存在
     */
    virtual std::shared_ptr<OutputStream> PutFile(const std::string &path, const FileMode &mode) noexcept = 0;

    /* *
     * @brief 读取一个小文件全部内容到buffer中
     * @param path 文件路径
     * @param dataBuffer 读取到的数据
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限读取文件
     * EINVAL : 参数无效
     * ENOENT : 文件不存在
     * EIO : 发生IO错误
     */
    virtual int GetFile(const std::string &path, utils::ByteBuffer &dataBuffer) noexcept;

    /* *
     * @brief 打开文件读取得到一个输入流，从流中读取
     * @param path 文件路径
     * @return 成功返回流，失败返回nullptr，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限读取文件
     * EINVAL : 参数无效
     * ENOENT : 文件不存在
     */
    virtual std::shared_ptr<InputStream> GetFile(const std::string &path) noexcept;

    /* *
     * @brief 读取文件，自动将文件内容写到提供的输出流中
     * @param path 文件路径
     * @param outputStream 文件内容输出到此流
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限读取文件
     * EINVAL : 参数无效
     * ENOENT : 文件不存在
     * EIO : 发生IO错误
     */
    virtual int GetFile(const std::string &path, OutputStream &outputStream) noexcept;

    /* *
     * @brief 读取一个文件小部分内容到buffer中
     * @param path 文件路径
     * @param dataBuffer 读取到的数据
     * @param range 读取范围
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限读取文件
     * EINVAL : 参数无效
     * ENOENT : 文件不存在
     * EIO : 发生IO错误
     */
    virtual int GetFile(const std::string &path, utils::ByteBuffer &dataBuffer, const FileRange &range) noexcept = 0;

    /* *
     * @brief 打开文件读取其中一段数据得到一个输入流，从流中读取
     * @param path 文件路径
     * @param range 读取范围
     * @return 成功返回流，失败返回nullptr，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限读取文件
     * EINVAL : 参数无效
     * ENOENT : 文件不存在
     */
    virtual std::shared_ptr<InputStream> GetFile(const std::string &path, const FileRange &range) noexcept = 0;

    /* *
     * @brief 读取文件一部分内容，自动将文件内容写到提供的输出流中
     * @param path 文件路径
     * @param range 读取范围
     * @param outputStream 文件内容输出到此流
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限读取文件
     * EINVAL : 参数无效
     * ENOENT : 文件不存在
     * EIO : 发生IO错误
     */
    virtual int GetFile(const std::string &path, const FileRange &range, OutputStream &outputStream) noexcept = 0;

    /* *
     * @brief 移动（改名）一个文件或目录，仅可以在同一个under fs实例内部做移动
     * @param source 旧的路径
     * @param destination 新的路径
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限做此操作
     * ENOENT : 旧名称不存在
     * EEXIST : 新名称已存在
     */
    virtual int MoveFile(const std::string &source, const std::string &destination) noexcept = 0;

    /* *
     * @brief Link一个文件，不拷贝数据，仅可以在同一个under fs实例内部做移动
     * @param source 旧的路径
     * @param destination 新的路径
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限做此操作
     * ENOENT : 旧名称不存在
     * EEXIST : 新名称已存在
     */
    virtual int CopyFile(const std::string &source, const std::string &destination) noexcept = 0;

    /* *
     * @brief 删除一个文件
     * @param path 文件路径
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限读取文件
     * ENOENT : 文件不存在
     * EIO : 发生IO错误
     */
    virtual int RemoveFile(const std::string &path) noexcept = 0;

    /* *
     * @brief 创建一个目录
     * @param path 目录路径
     * @param mode 权限信息
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限在当前路径创建目录
     * EEXIST : 目录已存在
     * EINVAL : 参数无效
     */
    virtual int CreateDirectory(const std::string &path, const FileMode &mode) noexcept = 0;

    /* *
     * @brief 删除一个目录
     * @param path 目录路径
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限删除此目录
     * ENOENT : 目录不存在
     * EINVAL : 参数无效
     */
    virtual int RemoveDirectory(const std::string &path) noexcept = 0;

    /* *
     * @brief 从头列举一个目录
     * @param path 目录路径
     * @param result 返回结果，包含目录项和marker，用于下次接着本次结果列举
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限打开此目录
     * ENOENT : 目录不存在
     * EINVAL : 参数无效
     */
    virtual int ListFiles(const std::string &path, ListFileResult &result) noexcept = 0;

    /* *
     * @brief 根据返回的maker往下列举一个目录
     * @param path 目录路径
     * @param result 返回结果，包含目录项和marker，用于下次接着本次结果列举
     * @param marker 上次返回结果中的标记
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限打开此目录
     * ENOENT : 目录不存在
     * EINVAL : 参数无效
     */
    virtual int ListFiles(const std::string &path, ListFileResult &result,
        std::shared_ptr<ListFilePageMarker> marker) noexcept = 0;

    /* *
     * @brief 读取文件元数据
     * @param path 文件路径
     * @param meta [out] 文件元数据
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限访问此文件或目录
     * ENOENT : 文件或目录不存在
     * EINVAL : 参数无效
     */
    virtual int GetFileMeta(const std::string &path, FileMeta &meta) noexcept = 0;

    /* *
     * @brief 设置文件元数据
     * @param path 文件路径
     * @param meta 文件元数据
     * @return 成功返回0，失败返回-1，errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限操作此文件或目录的元数据
     * ENOENT : 文件或目录不存在
     * EINVAL : 参数无效
     */
    virtual int SetFileMeta(const std::string &path, std::map<std::string, std::string> &meta) noexcept = 0;

    /* *
     * @brief 获取文件锁对象
     * @param path 文件路径
     * @return 成功返回锁对象，失败返回nullptr, errno指定错误码
     * @details 本操作常见错误码
     * EACCES : 没有权限获取此文件的锁
     * EINVAL : 参数无效
     * ENOENT : 文件不存在
     * EOPNOTSUPP : 文件服务不支持此操作
     */
    virtual std::shared_ptr<FileLock> GetFileLock(const std::string &path) noexcept = 0;
};
}
}

#endif // OCK_DFS_UFS_API_H
