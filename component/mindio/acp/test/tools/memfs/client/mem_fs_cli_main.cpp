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
#include <fcntl.h>

#include <map>
#include <list>
#include <string>
#include <iostream>

#include "hlog.h"
#include "common_includes.h"
#include "memfs_sdk_api.h"

using namespace ock::hlog;

using Iter = std::list<std::string>::const_iterator;

struct WriteFileParam {
    int fd = -1;
    std::string sourceFile;
    int64_t size = 0;
    int64_t skip = 0;
};

struct ReadFileParam {
    int fd = -1;
    std::string targetFile;
    int64_t size = 0;
    int64_t seek = 0;
};

static constexpr auto MODE_BASE = 8;
static constexpr auto FD_BASE = 10;
static constexpr auto RW_BUF_SZ = 16 * 1024;
static constexpr int MFS_LOG_FILE_SIZE = 100 * 1024 * 1024; // 100MB
static constexpr int MFS_LOG_FILE_COUNT = 50;

static std::map<int, std::string> openFiles;
static std::map<std::string, std::function<void(const std::list<std::string> &)>> processors;

static std::list<std::string> SplitInputCommands(const std::string &input) noexcept;
static void InitializeCommandProcessor() noexcept;
static void RunCommand() noexcept;
static void PrintUsage(const std::list<std::string> &inputs) noexcept;
static void CreateDirectory(const std::list<std::string> &inputs) noexcept;
static void RemoveDirectory(const std::list<std::string> &inputs) noexcept;
static void OpenFile(const std::list<std::string> &inputs) noexcept;
static void CloseFile(const std::list<std::string> &inputs) noexcept;
static void WriteFile(const std::list<std::string> &inputs) noexcept;
static void ReadFile(const std::list<std::string> &inputs) noexcept;
static void ListFiles(const std::list<std::string> &inputs) noexcept;

int main(int argc, char *argv[])
{
    Hlog::CreateInstance(1, 1, "./log_memfs_cli.log", MFS_LOG_FILE_SIZE, MFS_LOG_FILE_COUNT);

    auto ret = MemFsClientInitialize();
    if (ret != 0) {
        std::cerr << "ERROR: Initialize MemFs Client Failed: " << ret << std::endl;
        return -1;
    }

    MemFsClientUnInitialize();
    RunCommand();
    return 0;
}

static bool ParseOptions(Iter &pos, const Iter &end, const std::string &op, std::string &arg)
{
    if (pos->length() > op.length()) {
        arg = pos->substr(op.length());
        return true;
    }

    ++pos;
    if (pos == end) {
        return false;
    }

    arg = *pos;
    return true;
}


static std::list<std::string> SplitInputCommands(const std::string &input) noexcept
{
    std::vector<std::string> items;
    StrUtil::Split(input, " ", items);

    std::list<std::string> result;
    std::for_each(items.begin(), items.end(), [&result](std::string &item) {
        if (!item.empty()) {
            result.push_back(std::move(item));
        }
    });
    return std::move(result);
}

static void PrintUsage(const std::list<std::string> &inputs) noexcept
{
    std::cout << "help: print this message." << std::endl;
    std::cout << std::endl;

    std::cout << "create directory:" << std::endl;
    std::cout << "\tmkdir [-p] [-m mode=0755] <path>" << std::endl;
    std::cout << "remove directory:" << std::endl;
    std::cout << "\trmdir [-r] <path>" << std::endl;
    std::cout << "open file to read or write:" << std::endl;
    std::cout << "\topen <-w|-r> <path>" << std::endl;
    std::cout << "close one file:" << std::endl;
    std::cout << "\tclose <fd>" << std::endl;
    std::cout << "write file use given source file:" << std::endl;
    std::cout << "\twrite <fd> <source> <length> [<skip>]" << std::endl;
    std::cout << "read file use given target file:" << std::endl;
    std::cout << "\tread <fd> <off> <length> <dest_file>" << std::endl;
    std::cout << "list open files:" << std::endl;
    std::cout << "\tfiles" << std::endl;
    std::cout << "exit:" << std::endl;
    std::cout << "\tquit or exit" << std::endl;
}

static void InitializeCommandProcessor() noexcept
{
    processors["help"] = PrintUsage;
    processors["mkdir"] = CreateDirectory;
    processors["rmdir"] = RemoveDirectory;
    processors["open"] = OpenFile;
    processors["close"] = CloseFile;
    processors["write"] = WriteFile;
    processors["read"] = ReadFile;
    processors["files"] = ListFiles;
}

static void RunCommand() noexcept
{
    InitializeCommandProcessor();

    std::string input;
    std::cout << "admin> ";
    std::flush(std::cout);
    while (std::getline(std::cin, input)) {
        auto commands = SplitInputCommands(input);
        if (commands.empty()) {
            std::cout << "admin> ";
            std::flush(std::cout);
            continue;
        }

        if (commands.front() == "quit" || commands.front() == "exit") {
            break;
        }

        auto command = commands.front();
        auto pos = processors.find(commands.front());
        if (pos == processors.end()) {
            std::cout << "ERROR invalid command : " << commands.front() << std::endl;
            std::cout << "admin> ";
            std::flush(std::cout);
            continue;
        }

        commands.pop_front();
        auto &processor = pos->second;
        processor(commands);
        std::cout << "admin> ";
        std::flush(std::cout);
    }
}

static void CreateDirectory(const std::list<std::string> &inputs) noexcept
{
    std::string path;
    mode_t mode = 0755;
    bool recursive = false;

    auto it = inputs.begin();
    while (it != inputs.end()) {
        if (*it == "-p") {
            recursive = true;
            ++it;
            continue;
        }

        if (StrUtil::StartWith(*it, "-m")) {
            std::string modeStr;
            if (!ParseOptions(it, inputs.end(), "-m", modeStr)) {
                std::cout << "Error: missing argument for '-m'" << std::endl;
                return;
            }

            mode = std::strtol(modeStr.c_str(), nullptr, MODE_BASE);
            continue;
        }

        path = *it;
        ++it;
    }

    if (path.empty()) {
        std::cout << "-- Error : no path" << std::endl;
        return;
    }

    std::cout << "-- create directory(" << path << "), mode=0" << std::oct << mode << std::dec << ", recursive=";
    std::cout << recursive << std::endl;
    auto ret = MemFsMkDir(path.c_str(), static_cast<int>(mode), recursive);
    if (ret != 0) {
        std::cout << "-- Error : failed : " << errno << ", " << strerror(errno) << std::endl;
        return;
    }
    std::cout << "-- success" << std::endl;
}

static void RemoveDirectory(const std::list<std::string> &inputs) noexcept
{
    std::string path;
    bool recursive = false;

    for (auto &in : inputs) {
        if (in == "-p") {
            recursive = true;
        } else {
            path = in;
        }
    }

    if (path.empty()) {
        std::cout << "-- Error : no path" << std::endl;
        return;
    }

    std::cout << "-- remove directory(" << path << "), recursive=" << recursive << std::endl;
    auto ret = MemFsRmDir(path.c_str(), recursive);
    if (ret != 0) {
        std::cout << "-- Error : failed : " << errno << ", " << strerror(errno) << std::endl;
        return;
    }
    std::cout << "-- success" << std::endl;
}

static void OpenFile(const std::list<std::string> &inputs) noexcept
{
    std::string path;
    bool readMode = false;
    bool writeMode = false;

    for (auto &in : inputs) {
        if (in == "-r") {
            readMode = true;
        } else if (in == "-w") {
            writeMode = true;
        } else {
            path = in;
        }
    }

    if (path.empty() || readMode == writeMode) {
        std::cout << "-- Error : parameter wrong" << std::endl;
        return;
    }

    int flags;
    if (readMode) {
        std::cout << "-- open file(" << path << ") for read" << std::endl;
        flags = O_RDONLY;
    } else {
        std::cout << "-- open file(" << path << ") for write" << std::endl;
        flags = O_CREAT | O_TRUNC | O_WRONLY;
    }

    auto fd = MemFsOpenFile(path.c_str(), flags);
    if (fd < 0) {
        std::cout << "-- Error : failed : " << errno << ", " << strerror(errno) << std::endl;
        return;
    }
    std::cout << "-- success fd = " << fd << std::endl;
    openFiles[fd] = path;
}

static void CloseFile(const std::list<std::string> &inputs) noexcept
{
    if (inputs.empty()) {
        std::cout << "-- Error :missing input fd" << std::endl;
        return;
    }

    auto fd = static_cast<int>(std::strtol(inputs.front().c_str(), nullptr, FD_BASE));
    auto pos = openFiles.find(fd);
    if (pos == openFiles.end()) {
        std::cout << "-- Error : input fd(" << fd << ") invalid." << std::endl;
        return;
    }

    std::cout << "-- close file(" << fd << " -> " << pos->second << ") for read" << std::endl;
    auto ret = MemFsClose(fd);
    if (ret != 0) {
        std::cout << "-- Error : failed : " << errno << ", " << strerror(errno) << std::endl;
        return;
    }
    std::cout << "-- success" << std::endl;
    openFiles.erase(pos);
}

static bool ParseWriteParam(const std::list<std::string> &inputs, WriteFileParam &writeParam) noexcept
{
    auto backup = inputs;
    if (backup.empty()) {
        std::cout << "-- Error : missing parameter fd" << std::endl;
        return false;
    }
    writeParam.fd = static_cast<int>(std::strtol(backup.front().c_str(), nullptr, FD_BASE));
    backup.pop_front();

    if (backup.empty()) {
        std::cout << "-- Error : missing parameter source file path" << std::endl;
        return false;
    }
    writeParam.sourceFile = backup.front();
    backup.pop_front();

    if (backup.empty()) {
        std::cout << "-- Error : missing parameter write length" << std::endl;
        return false;
    }
    writeParam.size = std::strtol(backup.front().c_str(), nullptr, FD_BASE);
    backup.pop_front();

    if (!backup.empty()) {
        writeParam.skip = std::strtol(backup.front().c_str(), nullptr, FD_BASE);
    }

    if (writeParam.skip < 0) {
        writeParam.skip = 0;
    }

    auto pos = openFiles.find(writeParam.fd);
    if (pos == openFiles.end()) {
        std::cout << "-- Error : input fd(" << writeParam.fd << ") invalid." << std::endl;
        return false;
    }

    return true;
}

static void WriteFile(const std::list<std::string> &inputs) noexcept
{
    WriteFileParam writeParam;

    if (!ParseWriteParam(inputs, writeParam)) {
        return;
    }

    auto &file = openFiles[writeParam.fd];
    std::cout << "-- write file(" << writeParam.fd << " -> " << file << ") from source file : ";
    std::cout << writeParam.sourceFile << ", length = " << writeParam.size;
    if (writeParam.skip > 0) {
        std::cout << ", skip source file bytes = " << writeParam.skip;
    }
    std::cout << std::endl;

    auto inFd = open(writeParam.sourceFile.c_str(), O_RDONLY);
    if (inFd < 0) {
        std::cout << "-- open file(" << writeParam.sourceFile << ") to read failed: " << strerror(errno) << std::endl;
        return;
    }

    auto offset = lseek(inFd, writeParam.skip, SEEK_SET);
    if (offset != writeParam.skip) {
        std::cout << "-- seek file(" << writeParam.sourceFile << ") to failed: " << strerror(errno) << std::endl;
        close(inFd);
        return;
    }

    auto left = writeParam.size;
    char buffer[RW_BUF_SZ];
    while (left > 0) {
        auto needRead = std::min(static_cast<int64_t>(RW_BUF_SZ), left);
        auto bytes = read(inFd, buffer, needRead);
        if (bytes <= 0) {
            std::cout << "-- read file(" << writeParam.sourceFile << ") failed: " << strerror(errno) << std::endl;
            close(inFd);
            return;
        }

        auto realBytes = MemFsWrite(writeParam.fd, (uintptr_t)buffer, bytes);
        if (realBytes < bytes) {
            std::cout << "-- write file(" << file << ") failed: " << strerror(errno) << std::endl;
            close(inFd);
            return;
        }

        left -= bytes;
    }

    close(inFd);
    std::cout << "-- success" << std::endl;
}

static bool ParseReadParam(const std::list<std::string> &inputs, ReadFileParam &readParam) noexcept
{
    // read <fd> <off> <length> <dest_file>
    auto backup = inputs;
    if (backup.empty()) {
        std::cout << "-- Error : missing parameter fd" << std::endl;
        return false;
    }
    readParam.fd = static_cast<int>(std::strtol(backup.front().c_str(), nullptr, FD_BASE));
    backup.pop_front();

    if (backup.empty()) {
        std::cout << "-- Error : missing parameter offset" << std::endl;
        return false;
    }
    readParam.seek = std::strtol(backup.front().c_str(), nullptr, FD_BASE);
    backup.pop_front();

    if (backup.empty()) {
        std::cout << "-- Error : missing parameter length" << std::endl;
        return false;
    }
    readParam.size = std::strtol(backup.front().c_str(), nullptr, FD_BASE);
    backup.pop_front();

    if (backup.empty()) {
        std::cout << "-- Error : missing parameter target file path" << std::endl;
        return false;
    }
    readParam.targetFile = backup.front();
    backup.pop_front();

    auto pos = openFiles.find(readParam.fd);
    if (pos == openFiles.end()) {
        std::cout << "-- Error : input fd(" << readParam.fd << ") invalid." << std::endl;
        return false;
    }

    return true;
}

static void ReadFile(const std::list<std::string> &inputs) noexcept
{
    ReadFileParam readParam;

    if (!ParseReadParam(inputs, readParam)) {
        return;
    }

    auto &file = openFiles[readParam.fd];
    std::cout << "-- read file(" << readParam.fd << " -> " << file << ") to target file : " << readParam.targetFile;
    std::cout << ", length = " << readParam.size << ", offset = " << readParam.seek << std::endl;

    auto outFd = open(readParam.targetFile.c_str(), O_CREAT | O_TRUNC | O_WRONLY, 0600);
    if (outFd < 0) {
        std::cout << "-- open file(" << readParam.targetFile << ") to write failed: " << strerror(errno) << std::endl;
        return;
    }

    auto left = readParam.size;
    char buffer[RW_BUF_SZ];
    auto position = static_cast<uint64_t>(readParam.seek);
    while (left > 0) {
        auto needWrite = std::min(static_cast<int64_t>(RW_BUF_SZ), left);
        auto bytes = MemFsRead(readParam.fd, (uintptr_t)buffer, position, needWrite);
        if (bytes <= 0) {
            std::cout << "-- read file(" << file << ") to failed: " << strerror(errno) << std::endl;
            close(outFd);
            return;
        }

        auto readBytes = write(outFd, buffer, bytes);
        if (readBytes < bytes) {
            std::cout << "-- write file(" << readParam.targetFile << ") failed: " << strerror(errno) << std::endl;
            close(outFd);
            return;
        }

        left -= bytes;
        position += bytes;
    }

    close(outFd);
    std::cout << "-- success" << std::endl;
}

static void ListFiles(const std::list<std::string> &inputs) noexcept
{
    if (openFiles.empty()) {
        std::cout << "-- no open files!" << std::endl;
    }

    for (auto &e : openFiles) {
        std::cout << "-- all open files:" << std::endl;
        std::cout << "\t" << e.first << " -> " << e.second << std::endl;
    }
}