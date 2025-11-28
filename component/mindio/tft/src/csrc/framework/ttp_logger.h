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

#ifndef TTP_LOGGER_H
#define TTP_LOGGER_H

namespace ock {
namespace ttp {

constexpr int MFS_LOG_FILE_SIZE = 10 * 1024 * 1024; // 10MB
constexpr int MFS_LOG_FILE_COUNT = 5;

class TTPLogger {
public:
    TTPLogger() = default;
    ~TTPLogger() = default;

    static void Init();

    static int CreateLog();

    static void Log(int level, const std::string &message);

private:
    static int GetLogLevel();

    static int ValidateParams(const std::string &path);

    static int CreateLogImpl(int minLogLevel, std::string path, int rotationFileSize = MFS_LOG_FILE_SIZE,
        int rotationFileCount = MFS_LOG_FILE_COUNT);

    static constexpr int GAP = 1;
};

}  // namespace ttp
}  // namespace ock

#endif