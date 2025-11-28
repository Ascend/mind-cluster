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
#include <cstdio>
#include <cstdint>
#include <cstdlib>
#include <dirent.h>
#include <iostream>
#include <sys/stat.h>
#include <unistd.h>
#include "common.h"
#include "file_utils.h"

namespace {
    constexpr long MIN_MALLOC_SIZE = 1;
    constexpr long DEFAULT_MAX_DATA_SIZE = 1024 * 1024 * 1024;
    constexpr int  PER_PERMISSION_MASK_RWX = 0b111;
    constexpr size_t MAX_TLS_INFO_LEN = 10 * 1024;
}

using namespace ock::ttp;

namespace ock {
namespace ttp {

static long g_defaultMaxDataSize = DEFAULT_MAX_DATA_SIZE;

static size_t GetFileSize(const std::string &filePath)
{
    if (!FileUtils::CheckFileExists(filePath)) {
        TTP_LOG_ERROR("File does not exist!");
        return 0;
    }
    std::string baseDir = "/";
    std::string errMsg{};
    if (!FileUtils::RegularFilePath(filePath, baseDir, errMsg)) {
        TTP_LOG_ERROR("Regular file failed by " << errMsg);
        return 0;
    }

    FILE *fp = fopen(filePath.c_str(), "rb");
    if (fp == nullptr) {
        TTP_LOG_ERROR("File failed to open file.");
        return 0;
    }
    auto ret = fseek(fp, 0, SEEK_END);
    if (ret != 0) {
        TTP_LOG_ERROR("Error seeking to end of file");
        (void)fclose(fp);
        return 0;
    }
    size_t fileSize = static_cast<size_t>(ftell(fp));
    ret = fseek(fp, 0, SEEK_SET);
    if (ret != 0) {
        TTP_LOG_ERROR("Error seeking to set of file");
        (void)fclose(fp);
        return 0;
    }
    (void)fclose(fp);
    return fileSize;
}

static bool CheckDataSize(long size)
{
    if ((size > g_defaultMaxDataSize) || (size < MIN_MALLOC_SIZE)) {
        TTP_LOG_ERROR("Input data size(" << size << ") out of range[" <<
            MIN_MALLOC_SIZE << "," << g_defaultMaxDataSize << "].");
        return false;
    }

    return true;
}

bool FileUtils::CheckFileExists(const std::string &filePath)
{
    struct stat buffer;
    return (stat(filePath.c_str(), &buffer) == 0);
}

bool FileUtils::CheckDirectoryExists(const std::string &dirPath)
{
    struct stat buffer;
    if (stat(dirPath.c_str(), &buffer) != 0) {
        return false;
    }
    return (S_ISDIR(buffer.st_mode) == 1);
}

bool FileUtils::IsSymlink(const std::string &filePath)
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

bool FileUtils::RegularFilePath(const std::string &filePath, const std::string &baseDir, std::string &errMsg)
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
    if (IsSymlink(filePath)) {
        errMsg = "The file is a link.";
        return false;
    }
    char path[PATH_MAX + 1] = { 0x00 };
    const char *ret = realpath(filePath.c_str(), path);
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

bool FileUtils::RegularFilePath(const std::string &filePath, std::string &errMsg)
{
    if (filePath.empty()) {
        errMsg = "The file path is empty.";
        return false;
    }

    if (filePath.size() > PATH_MAX) {
        errMsg = "The file path exceeds the maximum value set by PATH_MAX.";
        return false;
    }
    if (IsSymlink(filePath)) {
        errMsg = "The file is a link.";
        return false;
    }
    char path[PATH_MAX + 1] = { 0x00 };
    const char *ret = realpath(filePath.c_str(), path);
    if (ret == nullptr) {
        errMsg = "The path realpath parsing failed. " + std::to_string(errno);
        return false;
    }
    return true;
}

bool FileUtils::IsFileValid(const std::string &configFile, std::string &errMsg, bool checkPermission,
                            bool onlyCurrentUserOp)
{
    if (!CheckFileExists(configFile)) {
        errMsg = "The input file is not a regular file or not exists";
        return false;
    }
    size_t fileSize = GetFileSize(configFile);
    if (fileSize == 0) {
        errMsg = "The input file is empty";
    } else if (!CheckDataSize(fileSize)) {
        errMsg = "Read input file failed, file is too large";
        return false;
    }
    return true;
}

bool FileUtils::CheckOwner(const std::string &filePath, std::string &errMsg)
{
    struct stat buf;
    int ret = stat(filePath.c_str(), &buf);
    if (ret != 0) {
        errMsg = "Get file stat failed.";
        return false;
    }
    if (buf.st_uid != getuid()) {
        errMsg = "owner id diff: current process user id is " + std::to_string(getuid()) + ", file owner id is " +
            std::to_string(buf.st_uid);
        return false;
    }
    return true;
}

bool FileUtils::CheckPermission(const std::string &filePath, const mode_t &mode, bool onlyCurrentUserOp,
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

static TResult ParseStr2Array(const std::string &token, char splitter, std::set<std::string> &parts)
{
    std::istringstream tokenSteam(token);
    std::string part;
    while (std::getline(tokenSteam, part, splitter)) {
        part = TrimString(part);
        if (!part.empty()) {
            parts.insert(part);
        }
    }

    if (parts.empty()) {
        TTP_LOG_WARN("parse token to array failed");
        return TTP_ERROR;
    }
    return TTP_OK;
}

static TResult ParseStr2KV(const std::string &token, char splitter, std::pair<std::string, std::string> &pair)
{
    std::istringstream stm(token);
    std::string key;
    std::string value;
    if (std::getline(stm, key, splitter) && std::getline(stm, value, splitter)) {
        key = TrimString(key);
        value = TrimString(value);
        if (!key.empty() && !value.empty()) {
            pair.first = key;
            pair.second = value;
            return TTP_OK;
        }
    }

    TTP_LOG_WARN("parse token to kv failed");
    return TTP_ERROR;
}

static bool SetTlsOptionValue(AccTlsOption &tlsOption, const std::string &key, const std::string &value)
{
    if (key == "tlsCaPath") {
        tlsOption.tlsCaPath = value;
    } else if (key == "tlsCert") {
        tlsOption.tlsCert = value;
    } else if (key == "tlsCrlPath") {
        tlsOption.tlsCrlPath = value;
    } else if (key == "tlsPk") {
        tlsOption.tlsPk = value;
    } else if (key == "tlsPkPwd") {
        tlsOption.tlsPkPwd = value;
    } else if (key == "packagePath") {
        tlsOption.packagePath = value;
    } else {
        return false;
    }
    return true;
}

static bool SetTlsOptionValues(AccTlsOption &tlsOption, const std::string &key, std::set<std::string> &values)
{
    if (key == "tlsCrlFile") {
        tlsOption.tlsCrlFile = values;
    } else if (key == "tlsCaFile") {
        tlsOption.tlsCaFile = values;
    } else {
        return false;
    }
    return true;
}

bool FileUtils::ParseTlsInfo(const std::string &tlsInfo, AccTlsOption& tlsOpts)
{
    if (tlsInfo.size() > MAX_TLS_INFO_LEN) {
        TTP_LOG_ERROR("tls info null or len invalid.");
        return false;
    }

    std::istringstream tokenSteam(tlsInfo);
    std::vector<std::string> tokens;
    std::string token;

    while (std::getline(tokenSteam, token, ';')) {
        if (!TrimString(token).empty()) {
            tokens.push_back(token);
        }
    }

    for (std::string &t : tokens) {
        std::pair<std::string, std::string> pair;
        auto ret = ParseStr2KV(t, ':', pair);
        if (ret != TTP_OK) {
            continue;
        }

        bool res = true;
        auto key = pair.first;
        std::set<std::string> paths;
        if (pair.first == "tlsCrlFile" || pair.first == "tlsCaFile") {
            ret = ParseStr2Array(pair.second, ',', paths);
            if (ret != TTP_OK) {
                continue;
            }

            res = SetTlsOptionValues(tlsOpts, pair.first, paths);
        } else {
            res = SetTlsOptionValue(tlsOpts, pair.first, pair.second);
        }

        if (!res) {
            TTP_LOG_WARN("un-match tls info key " << pair.first);
        }
    }

    return true;
}

}  // namespace ttp
}  // namespace ock