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
#ifndef OCK_DFS_BACKUP_FILE_TRACER_H
#define OCK_DFS_BACKUP_FILE_TRACER_H

#include <cstdint>
#include <string>
#include <unordered_map>

namespace ock {
namespace bg {
namespace backup {
struct FileTrace {
    std::string path;
    int64_t inode;

    FileTrace() noexcept : inode{ -1L } {}

    FileTrace(std::string p, int64_t i) noexcept : path{ std::move(p) }, inode{ i } {}
};

struct FileTraceCount {
    FileTrace trace;
    int64_t count;

    FileTraceCount(std::string p, int64_t i) noexcept : trace{ std::move(p), i }, count{ 1L } {}

    explicit FileTraceCount(FileTrace ft) noexcept : trace{ std::move(ft) }, count{ 1L } {}
};

class BackupFileTracer {
public:
    BackupFileTracer()
    {
        pthread_spin_init(&lock, 0);
    }

    ~BackupFileTracer()
    {
        pthread_spin_destroy(&lock);
    }
    void TraceOpen(int fd, const std::string &name, int64_t inode) noexcept;
    bool CloseFind(int fd, FileTrace &tracer) noexcept;

private:
    std::unordered_map<int, FileTraceCount> tracers;
    pthread_spinlock_t lock{};
};
}
}
}

#endif // OCK_DFS_BACKUP_FILE_TRACER_H
