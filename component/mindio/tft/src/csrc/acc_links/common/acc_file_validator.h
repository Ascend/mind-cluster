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
#ifndef ACC_LINKS_ACC_FILE_VALIDATOR_H
#define ACC_LINKS_ACC_FILE_VALIDATOR_H

#include <string>

namespace ock {
namespace acc {
class FileValidator {
    static constexpr uint32_t MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
public:
    /**
     * @brief Get the realpath for security consideration
     */
    static bool Realpath(std::string &path);

    /**
     * @brief Find whether the path is a file or not
     *
     * @param path         [in] input path
     * @return true if it is a file
     */
    static bool IsFile(const std::string &path);

    /**
     * @brief Find whether the path is a directory or not
     *
     * @param path         [in] input path
     * @return true if it is a directory
     */
    static bool IsDir(const std::string &path);

    /**
     * @brief Find whether the path exceed the max size or not
     *
     * @param path         [in] input path
     * @param maxSize      [in] the max size allowed
     * @return true if the file size is less or equals to maxSize
     */
    static bool CheckFileSize(const std::string &path, uint32_t maxSize = MAX_FILE_SIZE);

    /**
     * judge file exists
     * @param filePath: file full path
     */
    static bool CheckFileExists(const std::string &filePath);

    /**
     * is directory symlink.
     * @param filePath directory
     * @return
     */
    static bool IsSymlink(const std::string &filePath);

    /** Regular the file path using realPath.
     * @param filePath raw file path
     * @param baseDir file path must in base dir
     * @param errMsg the err msg
     * @return
     */
    static bool RegularFilePath(const std::string &filePath, const std::string &baseDir, std::string &errMsg);

    /** Check the existence of the file and the size of the file.
     * @param configFile the input file path
     * @param errMsg the err msg
     * @return
     */
    static bool IsFileValid(const std::string &configFile, std::string &errMsg);

    /** Check the permission of the file.
     * @param filePath the input file path
     * @param mode the permission allowed
     * @param onlyCurrentUserOp strict check, only current user can write or execute
     * @param errMsg the err msg
     * @return
     */
    static bool CheckPermission(const std::string &filePath, const mode_t &mode, bool onlyCurrentUserOp,
                                std::string &errMsg);
};
}  // namespace acc
}  // namespace ock

#endif
