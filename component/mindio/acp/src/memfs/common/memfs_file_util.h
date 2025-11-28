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
#ifndef OCK_MEMFS_FILE_UTIL_H
#define OCK_MEMFS_FILE_UTIL_H

#include <cstring>
#include <dirent.h>
#include <string>
#include <sys/stat.h>
#include <unistd.h>

namespace ock {
namespace memfs {

class FileUtil {
public:
    /*
     * @brief Check if file or dir exists
     */
    static bool Exist(const std::string &path);

    /*
     * @brief Remove the dir without sub dirs
     */
    static bool Remove(const std::string &path, bool canonicalPath = true);

    /*
     * @brief Remove the dir recursively, its sub dir will be removed
     */
    static bool RemoveDirRecursive(const std::string &path);

    /*
     * @brief Get the realpath for security consideration
     */
    static bool CanonicalPath(std::string &path);
};

inline bool FileUtil::Exist(const std::string &path)
{
    return access(reinterpret_cast<const char *>(path.c_str()), 0) != -1;
}

inline bool FileUtil::Remove(const std::string &path, bool canonicalPath)
{
    if (path.empty() || path.size() > 4096L) {
        return false;
    }

    std::string realPath = path;
    if (canonicalPath && !CanonicalPath(realPath)) {
        return false;
    }

    return ::remove(realPath.c_str()) == 0;
}

inline bool FileUtil::RemoveDirRecursive(const std::string &path)
{
    if (path.empty() || path.size() > 4096L) {
        return false;
    }

    std::string realPath = path;
    if (!CanonicalPath(realPath)) {
        return false;
    }

    DIR *dir = opendir(realPath.c_str());
    if (dir == nullptr) {
        return false;
    }

    struct dirent *entry = nullptr;
    while ((entry = readdir(dir))) {
        if (strcmp(entry->d_name, ".") == 0 || strcmp(entry->d_name, "..") == 0) {
            continue;
        }

        struct stat statBuf {};
        std::string absPath = realPath + "/" + entry->d_name;
        if (!stat(absPath.c_str(), &statBuf) && S_ISDIR(statBuf.st_mode)) {
            RemoveDirRecursive(absPath);
        }

        ::remove(absPath.c_str());
    }

    ::closedir(dir);

    ::remove(realPath.c_str());
    return true;
}

inline bool FileUtil::CanonicalPath(std::string &path)
{
    if (path.empty() || path.size() > 4096L) {
        return false;
    }

    /* It will allocate memory to store path */
    char *realPath = realpath(path.c_str(), nullptr);
    if (realPath == nullptr) {
        return false;
    }

    path = realPath;
    free(realPath);
    return true;
}

}
}

#endif // OCK_MEMFS_FILE_UTIL_H