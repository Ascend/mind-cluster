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

#ifndef ROTATING_SINK_H
#define ROTATING_SINK_H

#include <spdlog/spdlog.h>
#include <spdlog/sinks/rotating_file_sink.h>
#include <spdlog/sinks/base_sink.h>
#include <mutex>

namespace ock {
namespace ttp {

class RotatingSink : public spdlog::sinks::base_sink<std::mutex> {
public:
    RotatingSink(std::string baseFilename,
                 std::size_t maxSize,
                 std::size_t maxFiles,
                 const spdlog::file_event_handlers &event_handlers = {});
    static std::string CalcFilename(const std::string &filename, std::size_t index);
    ~RotatingSink() override;

protected:
    void sink_it_(const spdlog::details::log_msg &msg) override;
    void flush_() override;

private:
    bool SameFile() const;

    // log.txt -> log.1.txt
    // log.1.txt -> log.2.txt
    // log.2.txt -> log.3.txt
    // log.3.txt -> delete
    void Rotate();

    // delete the target if exists, and rename the src file  to target
    // return true on success, false otherwise.
    bool RenameFile(const std::string &srcFilename, const std::string &targetFilename);

    std::string baseFilename_;
    std::size_t maxSize_;
    std::size_t maxFiles_;
    spdlog::details::file_helper file_helper_;
    int lockFile_;
    int logFile_;
};

}  // namespace ttp
}  // namespace ock

#endif