/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
#include <climits>
#include <regex>
#include <cstdint>
#include <iostream>
#include <sys/stat.h>
#include <unistd.h>

#include "acc_includes.h"
#include "acc_file_validator.h"

namespace {
constexpr long MIN_MALLOC_SIZE = 1;
constexpr long DEFAULT_MAX_DATA_SIZE = 1024 * 1024 * 1024;
constexpr mode_t PER_PERMISSION_MASK_RWX = 0b111;
}  // namespace

namespace ock {
namespace acc {
static long g_defaultMaxDataSize = DEFAULT_MAX_DATA_SIZE;
static const mode_t FILE_MODE = 0740;

static size_t GetFileSize(const std::string &filePath)
{
    if (!FileValidator::CheckFileExists(filePath)) {
        std::cerr << "File does not exist!" << std::endl;
        return 0;
    }
    std::string baseDir = "/";
    std::string errMsg;
    if (!FileValidator::RegularFilePath(filePath, baseDir, errMsg)) {
        std::cerr << "Regular file failed by " << errMsg << std::endl;
        return 0;
    }

    FILE *fp = fopen(filePath.c_str(), "rb");
    if (fp == nullptr) {
        std::cerr << "File: failed to open file." << std::endl;
        return 0;
    }
    (void)fseek(fp, 0, SEEK_END);
    size_t fileSize = static_cast<size_t>(ftell(fp));
    (void)fseek(fp, 0, SEEK_SET);
    (void)fclose(fp);
    return fileSize;
}

static bool CheckDataSize(long size)
{
    if ((size > g_defaultMaxDataSize) || (size < MIN_MALLOC_SIZE)) {
        std::cerr << "Input data size(" << size << ") out of range[" << MIN_MALLOC_SIZE << "," << g_defaultMaxDataSize
                  << "]." << std::endl;
        return false;
    }

    return true;
}

bool FileValidator::Realpath(std::string &path)
{
    if (path.empty() || path.size() > PATH_MAX) {
        return false;
    }

    /* It will allocate memory to store path */
    char* tmp = new char[PATH_MAX + 1];
    char* realPath = realpath(path.c_str(), tmp);
    if (realPath == nullptr) {
        delete[] tmp;
        return false;
    }

    path = realPath;
    realPath = nullptr;
    delete[] tmp;
    return true;
}

bool FileValidator::IsFile(const std::string &path)
{
    struct stat buf;
    if (lstat(path.c_str(), &buf) != 0) {
        return false;
    }
    return S_ISREG(buf.st_mode);
}

bool FileValidator::IsDir(const std::string &path)
{
    struct stat buf;
    if (lstat(path.c_str(), &buf) != 0) {
        return false;
    }
    return S_ISDIR(buf.st_mode);
}

bool FileValidator::CheckFileSize(const std::string &path, uint32_t maxSize)
{
    if (!CheckFileExists(path)) {
        return false;
    }

    return GetFileSize(path) <= static_cast<size_t>(maxSize);
}

bool FileValidator::CheckFileExists(const std::string &filePath)
{
    struct stat buffer;
    return (stat(filePath.c_str(), &buffer) == 0);
}

bool FileValidator::IsSymlink(const std::string &filePath)
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

bool FileValidator::RegularFilePath(const std::string &filePath, const std::string &baseDir, std::string &errMsg)
{
    if (filePath.empty()) {
        errMsg = "The file path is empty.";
        return false;
    }
    if (baseDir.empty()) {
        errMsg = "The file path basedir is empty.";
        return false;
    }
    if (filePath.size() > PATH_MAX) {
        errMsg = "The file path exceeds the maximum value set by PATH_MAX.";
        return false;
    }
    if (baseDir.size() > PATH_MAX) {
        errMsg = "The file path basedir exceeds the maximum value set by PATH_MAX.";
        return false;
    }
    if (IsSymlink(filePath)) {
        errMsg = "The file is a link.";
        return false;
    }
    char path[PATH_MAX + 1] = { 0x00 }; // Auto-managed stack array

    char *ret = realpath(filePath.c_str(), path);
    if (ret == nullptr) {
        errMsg = "The path realpath parsing failed.";
        return false;
    }

    std::string realFilePath(path, path + strlen(path));

    std::string dir = baseDir.back() == '/' ? baseDir : baseDir + "/";
    if (realFilePath.rfind(dir, 0) != 0) {
        errMsg = "The file is invalid, it's not in baseDir directory.";
        return false;
    }

    return true;
}

bool FileValidator::IsFileValid(const std::string &configFile, std::string &errMsg)
{
    if (!CheckFileExists(configFile)) {
        errMsg = "The input file is not a regular file or not exists";
        return false;
    }

    size_t fileSize = GetFileSize(configFile);
    if (fileSize == 0) {
        errMsg = "The input file is empty";
    } else if (!CheckDataSize(fileSize)) {
        errMsg = "Read input file failed, file is too large.";
        return false;
    }
    return true;
}

bool FileValidator::CheckPermission(const std::string &filePath, const mode_t &mode, bool onlyCurrentUserOp,
                                    std::string &errMsg)
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
        const uint32_t readPerm = 4;
        const uint32_t noPerm = 0;
        if (onlyCurrentUserOp && i != perPermWidth && curPerm != noPerm && curPerm != readPerm) {
            errMsg = " Check " + permMsg[i - 1] + " failed: Current permission is " + std::to_string(curPerm) +
                     ", but required no write or execute permission.";
            return false;
        }
    }
    return true;
}
}  // namespace acc
}  // namespace ock