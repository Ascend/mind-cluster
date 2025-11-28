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

#include <sys/file.h>
#include "common.h"
#include "rotating_sink.h"

namespace ock {
namespace ttp {

constexpr unsigned int SLEEP_TIME_100MS = 100;

RotatingSink::RotatingSink(
    std::string baseFilename,
    std::size_t maxSize,
    std::size_t maxFiles,
    const spdlog::file_event_handlers &event_handlers)
    : baseFilename_(std::move(baseFilename)),
      maxSize_(maxSize),
      maxFiles_(maxFiles),
      file_helper_{event_handlers},
      lockFile_(-1),
      logFile_(-1)
{
    file_helper_.open(CalcFilename(baseFilename_, 0));
    auto file = baseFilename_.substr(0, baseFilename_.find_last_of('/')) + "/.ttp_log.lock";
    lockFile_ = open(file.c_str(), O_RDWR | O_CREAT, S_IRUSR | S_IWUSR | S_IRGRP);
#ifdef UT_ENABLED
    if (GetEnvValue2Uint32("TEST_LOG_OPEN", 1, 1, 0)) {
        lockFile_ = -1;
    }
#endif
    if (lockFile_ == -1) { // 即使文件锁不可用也避免抛异常中断训练
        spdlog::debug("Failed to open ttp log lock file");
    }
    logFile_ = open(baseFilename_.c_str(), O_RDONLY);
#ifdef UT_ENABLED
    if (GetEnvValue2Uint32("TEST_LOG_OPEN", 1, 1, 0)) {
        logFile_ = -1;
    }
#endif
    if (logFile_ == -1) { // file_helper_.open()保证日志文件可以打开
        spdlog::debug("Failed to open ttp log file");
    }
}

RotatingSink::~RotatingSink()
{
    if (lockFile_ != -1) {
        close(lockFile_);
    }
    if (logFile_ != -1) {
        close(logFile_);
    }
}

// calc filename according to index and file extension if exists.
// e.g. CalcFilename("logs/mylog.txt, 3) => "logs/mylog.3.txt".
std::string RotatingSink::CalcFilename(const std::string &filename, std::size_t index)
{
    if (index == 0u) {
        return filename;
    }

    auto [base, ext] = spdlog::details::file_helper::split_by_extension(filename);
    return spdlog::fmt_lib::format(SPDLOG_FMT_STRING(SPDLOG_FILENAME_T("{}.{}{}")), base, index, ext);
}

bool RotatingSink::SameFile() const
{
    struct stat stFd{};
    struct stat stPt{};
    if (fstat(logFile_, &stFd) == -1 || stat(baseFilename_.c_str(), &stPt) == -1) {
        return false;
    }
    return stFd.st_ino == stPt.st_ino;
}

void RotatingSink::sink_it_(const spdlog::details::log_msg &msg)
{
    spdlog::memory_buf_t formatted;
    base_sink<std::mutex>::formatter_->format(msg, formatted);
    flock(lockFile_, LOCK_EX);
    auto newSize = file_helper_.size() + formatted.size();
    // rotate if the new estimated file size exceeds max size.
    // rotate only if the real size > 0 to better deal with full disk (see issue #2261).
    // we only check the real size when newSize > maxSize_ because it is relatively expensive.
    if (newSize > maxSize_) {
        file_helper_.flush();
        if (file_helper_.size() > 0) {
            if (SameFile()) {
                Rotate();
            } else {
                file_helper_.close();
                file_helper_.reopen(true);
            }
            close(logFile_);
            logFile_ = open(baseFilename_.c_str(), O_RDONLY);
        }
    }
    file_helper_.write(formatted);
    flock(lockFile_, LOCK_UN);
}

void RotatingSink::flush_()
{
    file_helper_.flush();
}

// log.txt -> log.1.txt
// log.1.txt -> log.2.txt
// log.2.txt -> log.3.txt
// log.3.txt -> delete
void RotatingSink::Rotate()
{
    using spdlog::details::os::filename_to_str;
    using spdlog::details::os::path_exists;

    file_helper_.close();
    for (auto i = maxFiles_; i > 0; --i) {
        std::string src = CalcFilename(baseFilename_, i - 1);
        if (!path_exists(src)) {
            continue;
        }

        std::string target = CalcFilename(baseFilename_, i);
        if (!RenameFile(src, target)) {
            // if failed try again after a small delay.
            // this is a workaround to a windows issue, where very high rotation
            // rates can cause the rename to fail with permission denied (because of antivirus?).
            spdlog::details::os::sleep_for_millis(SLEEP_TIME_100MS);
            if (!RenameFile(src, target)) {
                file_helper_.reopen(
                    true);  // truncate the log file anyway to prevent it to grow beyond its limit!
                spdlog::throw_spdlog_ex("rotating_file_sink: failed renaming " +
                    filename_to_str(src) + " to " + filename_to_str(target), errno);
            }
        }
    }
    file_helper_.reopen(true);
}

// delete the target if exists, and rename the src file  to target
// return true on success, false otherwise.
bool RotatingSink::RenameFile(const std::string &srcFilename,
                              const std::string &targetFilename)
{
    // try to delete the target file in case it already exists.
    (void)spdlog::details::os::remove(targetFilename);
    return spdlog::details::os::rename(srcFilename, targetFilename) == 0;
}

}  // namespace ttp
}  // namespace ock