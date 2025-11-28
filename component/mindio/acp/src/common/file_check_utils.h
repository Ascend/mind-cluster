/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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

#ifndef OCK_FILE_CHECK_UTILS_H
#define OCK_FILE_CHECK_UTILS_H

#include <sys/stat.h>
#include <iostream>
#include <unistd.h>
#include <fcntl.h>
#include <climits>
#include <chrono>
#include <thread>
#include <string>
#include <regex>
#include <vector>
namespace {
constexpr uint64_t MIN_MALLOC_SIZE = 1;
constexpr uint64_t DEFAULT_MAX_DATA_SIZE = 1073741824;
constexpr int PER_PERMISSION_MASK_RWX = 0b111;
}

namespace ock {
namespace common {
static const uint64_t g_defaultMaxDataSize = DEFAULT_MAX_DATA_SIZE;
const std::map<std::string, mode_t> DEFAULT_SERVER_DIR = { { "/logs", 0700 },
                                                           { "/conf", 0700 },
                                                           { "/uds", 0700 },
                                                           { "/ccae", 0750 } };
class FileCheckUtils {
public:
    static const mode_t FILE_MODE_750 = 0b111101000;
    static const mode_t FILE_MODE_550 = 0b101101000;
    static const mode_t FILE_MODE_500 = 0b101000000;
    static const mode_t FILE_MODE_640 = 0b110100000;
    static const mode_t FILE_MODE_600 = 0b110000000;
    static const mode_t FILE_MODE_444 = 0b100100100;
    static const mode_t FILE_MODE_440 = 0b100100000;
    static const mode_t FILE_MODE_400 = 0b100000000;
    /*
     * judge file exists
     * @param path: file full path
     * @param pattern: regex pattern
     */
    static inline bool CheckFileExists(const std::string &filePath)
    {
        struct stat buffer;
        return (stat(filePath.c_str(), &buffer) == 0);
    }

    /* *
     * is directory exists.
     * @param dir directory
     * @return
     */
    static bool CheckDirectoryExists(const std::string &dirPath)
    {
        struct stat buffer;
        if (stat(dirPath.c_str(), &buffer) != 0) {
            return false;
        }
        return (S_ISDIR(buffer.st_mode) == 1);
    }

    /* * Check whether the destination path is a link
     * @param filePath raw file path
     * @return
     */
    static bool IsSymlink(const std::string &filePath)
    {
        std::string cleanPath = filePath;
        while (!cleanPath.empty() && cleanPath.back() == '/') {
            cleanPath.pop_back();
        }
        struct stat buf;
        if (lstat(cleanPath.c_str(), &buf) != 0) {
            return false;
        }
        return S_ISLNK(buf.st_mode);
    }

    /* * Regular the file path using realPath.
     * @param filePath file path
     * @param baseDir file path must in base dir
     * @param errMsg the err msg
     * @return
     */
    static bool RegularFilePath(const std::string &filePath, const std::string &baseDir, std::string &errMsg)
    {
        if (filePath.empty()) {
            errMsg = "The file path: " + GetBaseFileName(filePath) + " is empty.";
            return false;
        }
        if (baseDir.empty()) {
            errMsg = "The file path basedir: " + GetBaseFileName(baseDir) + " is empty.";
            return false;
        }
        if (filePath.size() > PATH_MAX) {
            errMsg = "The file path: " + GetBaseFileName(filePath) + " exceeds the maximum value set by PATH_MAX.";
            return false;
        }
        if (IsSymlink(filePath)) {
            errMsg = "The file: " + GetBaseFileName(filePath) + " is a link.";
            return false;
        }
        if (CheckFileExists(filePath) || CheckDirectoryExists(filePath)) {
            char path[PATH_MAX + 1] = { 0x00 };
            char *ret = realpath(filePath.c_str(), path);
            if (ret == nullptr) {
                errMsg = "The path: " + GetBaseFileName(filePath) + " realpath parsing failed.";
                return false;
            }
            std::string realFilePath(path, path + strlen(path));

            std::string dir = baseDir.back() == '/' ? baseDir : baseDir + "/";
            if (realFilePath.rfind(dir, 0) != 0) {
                errMsg = "The file: " + GetBaseFileName(filePath) + " is invalid, it's not in baseDir directory.";
                return false;
            }
        }

        return true;
    }

    /* * Check the existence of the file and the size of the file.
     * @param filePath the input file path
     * @param errMsg the err msg
     * @param isFileExit if true, then file not exit, return false; else return true
     * @param mode file mode such as 0b111'101'000 means 750
     * @param checkOwner if true, check file owner; else not check
     * @param checkPermission if true, check file mode; else not check
     * @param maxfileSize default 10485760(10M)， maxfileSize must in (1, 1073741824](1G)
     * @return
     */
    static bool IsFileValid(const std::string &filePath, std::string &errMsg, bool isFileExit = true,
        mode_t mode = FILE_MODE_750, bool checkOwner = true, bool checkPermission = true,
        uint64_t maxfileSize = 10485760)
    {
        if (!CheckFileExists(filePath)) {
            errMsg = "The input file: " + GetBaseFileName(filePath) + " is not a regular file or not exists";
            return !isFileExit;
        }

        if (!CheckDirectoryExists(filePath)) {
            size_t fileSize = GetFileSize(filePath);
            if (fileSize == 0) {
                errMsg = "The input file: " + GetBaseFileName(filePath) + " is empty";
            } else if (!CheckDataSize(fileSize, maxfileSize)) {
                errMsg = "Read input file: " + GetBaseFileName(filePath) + " failed, file is too large";
                return false;
            }
        }

        if (checkOwner) {
            if (!ConstrainOwner(filePath, errMsg)) {
                errMsg = "Check path: " + GetBaseFileName(filePath) + " failed, by:" + errMsg;
                return false;
            }
        }

        if (checkPermission) {
            if (!ConstrainPermission(filePath, mode, errMsg)) {
                errMsg = "Check path: " + GetBaseFileName(filePath) + " failed, by:" + errMsg;
                return false;
            }
        }

        return true;
    }

    /* * Check the file owner, file must be owner current user
     * @param filePath the input file path
     * @param errMsg error msg
     * @return
     */
    static inline bool ConstrainOwner(const std::string &filePath, std::string &errMsg)
    {
        struct stat buf;
        int ret = stat(filePath.c_str(), &buf);
        if (ret != 0) {
            errMsg = "Get file stat failed.";
            return false;
        }
        if (buf.st_uid != getuid()) {
            errMsg = "owner id diff: current process effective user id is " + std::to_string(getuid()) +
                ", file owner id is " + std::to_string(buf.st_uid);
            return false;
        }
        return true;
    }

    /* * Check the file mode, file must be no greater than mode
     * @param filePath the input file path
     * @param mode file mode
     * @param errMsg error msg
     * @return
     */
    static bool ConstrainPermission(const std::string &filePath, const mode_t &mode, std::string &errMsg)
    {
        struct stat buf;
        int ret = stat(filePath.c_str(), &buf);
        if (ret != 0) {
            errMsg = "Get file stat failed.";
            return false;
        }

        mode_t mask = PER_PERMISSION_MASK_RWX;
        const int perPermWidth = 3;
        std::vector<std::string> permMsg = { "Other group permission", "Owner group permission", "Owner permission" };
        for (int i = perPermWidth; i > 0; i--) {
            uint32_t curPerm = (buf.st_mode & (mask << ((i - 1) * perPermWidth))) >> ((i - 1) * perPermWidth);
            uint32_t maxPerm = (mode & (mask << ((i - 1) * perPermWidth))) >> ((i - 1) * perPermWidth);
            if ((curPerm | maxPerm) != maxPerm) {
                errMsg = " Check " + permMsg[i - 1] + " failed: Current permission is " + std::to_string(curPerm) +
                    ", but required no greater than " + std::to_string(maxPerm) + ".";

                return false;
            }
        }
        return true;
    }

    static size_t GetFileSize(const std::string &filePath)
    {
        if (!FileCheckUtils::CheckFileExists(filePath)) {
            std::cerr << "File does not exist!" << std::endl;
            return 0;
        }
        std::string baseDir = "/";
        std::string errMsg{};
        if (!FileCheckUtils::RegularFilePath(filePath, baseDir, errMsg)) {
            std::cerr << "Regular file failed by " << errMsg << std::endl;
            return 0;
        }

        FILE *fp = fopen(filePath.c_str(), "rb");
        if (fp == nullptr) {
            std::cerr << "File failed to open file." << std::endl;
            return 0;
        }
        auto ret = fseek(fp, 0, SEEK_END);
        if (ret != 0) {
            std::cerr << "Error seeking to end of file" << std::endl;
            fclose(fp);
            return 0;
        }
        size_t fileSize = static_cast<size_t>(ftell(fp));
        ret = fseek(fp, 0, SEEK_SET);
        if (ret != 0) {
            std::cerr << "Error seeking to set of file" << std::endl;
            fclose(fp);
            return 0;
        }
        (void)fclose(fp);
        return fileSize;
    }

    static std::string GetBaseFileName(const std::string &path)
    {
        std::string tempPath = path;
        if (!tempPath.empty() && (tempPath.back() == '/' || tempPath.back() == '\\')) {
            tempPath.pop_back();
        }
        size_t lastSlashPos = tempPath.find_last_of("/\\");
        if (lastSlashPos == std::string::npos) {
            return tempPath;
        }
        return tempPath.substr(lastSlashPos + 1);
    }

    static bool CheckDataSize(uint64_t size, uint64_t maxFileSize)
    {
        if (maxFileSize <= MIN_MALLOC_SIZE || maxFileSize > g_defaultMaxDataSize) {
            return false;
        }
        if ((size > maxFileSize) || (size < MIN_MALLOC_SIZE)) {
            std::cerr << "Input data size(" << size << ") out of range[" << MIN_MALLOC_SIZE << "," << maxFileSize <<
                "]." << std::endl;
            return false;
        }

        return true;
    }

    static std::string RemovePrefixPath(const std::string &fullPath) noexcept
    {
        size_t firstSlashPos = fullPath.find('/');
        if (firstSlashPos == std::string::npos) {
            return fullPath;
        }
        size_t secondSlashPos = fullPath.find('/', firstSlashPos + 1);
        if (secondSlashPos == std::string::npos) {
            return fullPath.substr(firstSlashPos + 1);
        }
        return fullPath.substr(secondSlashPos + 1);
    }
};
} // namespace common
} // namespace ock

#endif // OCK_FILE_CHECK_UTILS_H
