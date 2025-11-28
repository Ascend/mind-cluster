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
#include "backup_file_view.h"

using namespace ock::bg::backup;

static constexpr auto BUCKET_COUNT_SHIFT = 8;
static constexpr auto BUCKET_COUNT = 1U << BUCKET_COUNT_SHIFT;
static constexpr auto BUCKET_COUNT_MASK = BUCKET_COUNT - 1U;

FileMeta::FileMeta(int64_t i, const struct timespec &mt) noexcept
    : inode{ i }, mtime{ mt.tv_sec, mt.tv_nsec }, lastBackupTime{ 0, 0 }
{
    clock_gettime(CLOCK_REALTIME_COARSE, &lastBackupTime);
}

BackupFileViewBucket::BackupFileViewBucket() noexcept
{
    pthread_spin_init(&lock, 0);
}

BackupFileViewBucket::~BackupFileViewBucket() noexcept
{
    pthread_spin_destroy(&lock);
}

BackupFileView::BackupFileView() noexcept : viewBuckets{ new (std::nothrow) BackupFileViewBucket[BUCKET_COUNT] } {}

BackupFileView::~BackupFileView() noexcept
{
    delete[] viewBuckets;
    viewBuckets = nullptr;
}

bool BackupFileView::AddFile(const std::string &path, const FileMeta &meta, FileMeta &old) const noexcept
{
    auto hashCode = std::hash<std::string>{}(path);
    auto bucketIndex = (static_cast<uint32_t>(hashCode) & BUCKET_COUNT_MASK);
    auto &bucket = viewBuckets[bucketIndex];

    bool success = false;
    pthread_spin_lock(&bucket.lock);
    auto pos = bucket.fileView.find(path);
    if (pos == bucket.fileView.end()) {
        bucket.fileView.emplace(path, meta);
        success = true;
    } else {
        old = pos->second;
    }
    pthread_spin_unlock(&bucket.lock);

    return success;
}

bool BackupFileView::RemoveFile(const std::string &path, int64_t inode) const noexcept
{
    auto hashCode = std::hash<std::string>{}(path);
    auto bucketIndex = (static_cast<uint32_t>(hashCode) & BUCKET_COUNT_MASK);
    auto &bucket = viewBuckets[bucketIndex];

    bool success = false;
    pthread_spin_lock(&bucket.lock);
    auto pos = bucket.fileView.find(path);
    if (pos != bucket.fileView.end() && inode == pos->second.inode) {
        bucket.fileView.erase(pos);
        success = true;
    }
    pthread_spin_unlock(&bucket.lock);

    return success;
}

bool BackupFileView::GetFile(const std::string &path, FileMeta &meta) const noexcept
{
    auto hashCode = std::hash<std::string>{}(path);
    auto bucketIndex = (static_cast<uint32_t>(hashCode) & BUCKET_COUNT_MASK);
    auto &bucket = viewBuckets[bucketIndex];

    bool found = false;
    pthread_spin_lock(&bucket.lock);
    auto pos = bucket.fileView.find(path);
    if (pos != bucket.fileView.end()) {
        meta = pos->second;
        found = true;
    }
    pthread_spin_unlock(&bucket.lock);

    return found;
}

bool BackupFileView::UpdateFile(const std::string &path, int64_t expectInode, const FileMeta &meta) const noexcept
{
    auto hashCode = std::hash<std::string>{}(path);
    auto bucketIndex = (static_cast<uint32_t>(hashCode) & BUCKET_COUNT_MASK);
    auto &bucket = viewBuckets[bucketIndex];

    bool success = false;
    pthread_spin_lock(&bucket.lock);
    auto pos = bucket.fileView.find(path);
    if (pos != bucket.fileView.end() && pos->second.inode == expectInode) {
        pos->second = meta;
        success = true;
    }
    pthread_spin_unlock(&bucket.lock);

    return success;
}

bool BackupFileView::RefreshBackupTime(const std::string &path, int64_t expectInode) const noexcept
{
    auto hashCode = std::hash<std::string>{}(path);
    auto bucketIndex = (static_cast<uint32_t>(hashCode) & BUCKET_COUNT_MASK);
    auto &bucket = viewBuckets[bucketIndex];

    struct timespec now {
        0, 0
    };
    clock_gettime(CLOCK_REALTIME_COARSE, &now);

    bool success = false;
    pthread_spin_lock(&bucket.lock);
    auto pos = bucket.fileView.find(path);
    if (pos != bucket.fileView.end() && pos->second.inode == expectInode) {
        pos->second.lastBackupTime = now;
        success = true;
    }
    pthread_spin_unlock(&bucket.lock);

    return success;
}