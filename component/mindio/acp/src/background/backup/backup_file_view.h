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
#ifndef OCK_DFS_BACKUP_FILE_VIEW_H
#define OCK_DFS_BACKUP_FILE_VIEW_H

#include <pthread.h>
#include <vector>
#include <string>
#include <unordered_map>

namespace ock {
namespace bg {
namespace backup {
struct FileMeta {
    int64_t inode;
    struct timespec mtime;
    struct timespec lastBackupTime;

    FileMeta() noexcept : inode(-1L), mtime{ 0, 0 }, lastBackupTime{ 0, 0 } {}

    FileMeta(int64_t i, const struct timespec &mt) noexcept;
};

class BackupFileViewBucket {
public:
    BackupFileViewBucket() noexcept;
    virtual ~BackupFileViewBucket() noexcept;

private:
    pthread_spinlock_t lock{};
    std::unordered_map<std::string, FileMeta> fileView;
    friend class BackupFileView;
};

class BackupFileView {
public:
    BackupFileView() noexcept;
    virtual ~BackupFileView() noexcept;
    inline bool Valid() const noexcept
    {
        return viewBuckets != nullptr;
    }

public:
    bool AddFile(const std::string &path, const FileMeta &meta, FileMeta &old) const noexcept;
    bool RemoveFile(const std::string &path, int64_t inode) const noexcept;
    bool GetFile(const std::string &path, FileMeta &meta) const noexcept;
    bool UpdateFile(const std::string &path, int64_t expectInode, const FileMeta &meta) const noexcept;
    bool RefreshBackupTime(const std::string &path, int64_t expectInode) const noexcept;

private:
    BackupFileViewBucket *viewBuckets;
};
}
}
}


#endif // OCK_DFS_BACKUP_FILE_VIEW_H
