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

#ifndef MIES_FILE_UTIL_H
#define MIES_FILE_UTIL_H

#include <string>
#include "acc_def.h"

using namespace ock::acc;

namespace ock {
namespace ttp {

class FileUtils {
public:
    /**
     * judge file exists
     * @param path: file full path
     * @param pattern: regex pattern
     */
    static bool CheckFileExists(const std::string& filePath);

    /**
     * is directory exists.
     * @param dir directory
     * @return
     */
    static bool CheckDirectoryExists(const std::string& dirPath);

    /** Check whether the destination path is a link
     * @param filePath raw file path
     * @return
     */
    static bool IsSymlink(const std::string& filePath);

    /** Regular the file path using realPath.
     * @param filePath raw file path
     * @param baseDir file path must in base dir
     * @param errMsg the err msg
     * @return
     */
    static bool RegularFilePath(const std::string& filePath, const std::string& baseDir, std::string &errMsg);

    /** Regular the file path using realPath.
     * @param filePath raw file path
     * @param errMsg the err msg
     * @return
     */
    static bool RegularFilePath(const std::string& filePath, std::string &errMsg);

    /** Check the existence of the file and the size of the file.
     * @param configFile the input file path
     * @param errMsg the err msg
     * @param checkPermission check perm
     * @param onlyCurrentUserOp strict check, only current user can write or execute
     * @return
     */
    static bool IsFileValid(const std::string& configFile, std::string &errMsg, bool checkPermission = true,
        bool onlyCurrentUserOp = true);

    static bool IsRegularFile(const char* path);

    static bool CheckOwner(const std::string &filePath, std::string &errMsg);

    static bool CheckPermission(const std::string &filePath, const mode_t &mode, bool onlyCurrentUserOp,
        std::string &errMsg);

    static bool ParseTlsInfo(const std::string &tlsInfo, AccTlsOption& tlsOpts);
};

}  // namespace ttp
}  // namespace ock

#endif // MIES_FILE_UTIL_H