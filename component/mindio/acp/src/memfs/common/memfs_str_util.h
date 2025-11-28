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
#ifndef OCK_MEMFS_STR_UTIL_H
#define OCK_MEMFS_STR_UTIL_H

#include <cmath>
#include <cstring>
#include <string>
#include <vector>

namespace ock {
namespace memfs {

class StrUtil {
public:
    static bool StartWith(const std::string &src, const std::string &start);
    static bool EndWith(const std::string &src, const std::string &end);

    static bool StrToLong(const std::string &src, long &value);
    static bool StrToUint(const std::string &src, uint32_t &value);
    static bool StrToULong(const std::string &src, uint64_t &value);
    static bool StrToFloat(const std::string &src, float &value);

    static void Split(const std::string &src, const std::string &sep, std::vector<std::string> &out);

    static void StrTrim(std::string &src);
};

inline bool StrUtil::StartWith(const std::string &src, const std::string &start)
{
    return src.compare(0, start.size(), start) == 0;
}

inline bool StrUtil::EndWith(const std::string &src, const std::string &end)
{
    return src.size() >= end.length() && src.compare(src.size() - end.size(), end.size(), end) == 0;
}

inline bool StrUtil::StrToLong(const std::string &src, long &value)
{
    char *remain = nullptr;
    errno = 0;
    value = std::strtol(src.c_str(), &remain, 10L); // 10 is decimal digits
    if ((value == 0 && src != "0") || remain == nullptr || strlen(remain) > 0 || errno == ERANGE) {
        return false;
    }
    return true;
}

inline bool StrUtil::StrToUint(const std::string &src, uint32_t &value)
{
    char *remain = nullptr;
    errno = 0;
    value = std::strtoul(src.c_str(), &remain, 10L); // 10 is decimal digits
    if ((value == 0 && src != "0") || remain == nullptr || strlen(remain) > 0 || errno == ERANGE) {
        return false;
    }
    return true;
}

inline bool StrUtil::StrToULong(const std::string &src, uint64_t &value)
{
    char *remain = nullptr;
    errno = 0;
    value = std::strtoul(src.c_str(), &remain, 10L); // 10 is decimal digits
    if ((value == 0 && src != "0") || remain == nullptr || strlen(remain) > 0 || errno == ERANGE) {
        return false;
    }
    return true;
}

inline bool StrUtil::StrToFloat(const std::string &src, float &value)
{
    constexpr float EPSINON = 0.000001;

    char *remain = nullptr;
    errno = 0;
    value = std::strtof(src.c_str(), &remain);
    if (remain == nullptr || strlen(remain) > 0 ||
        ((value - HUGE_VALF) >= -EPSINON && (value - HUGE_VALF) <= EPSINON && errno == ERANGE)) {
        return false;
    } else if ((value >= -EPSINON && value <= EPSINON) && (src != "0.0")) {
        return false;
    }
    return true;
}

inline void StrUtil::Split(const std::string &src, const std::string &sep, std::vector<std::string> &out)
{
    std::string::size_type pos1 = 0;
    std::string::size_type pos2 = src.find(sep);

    std::string tmpStr;
    while (std::string::npos != pos2) {
        tmpStr = src.substr(pos1, pos2 - pos1);
        out.emplace_back(tmpStr);
        pos1 = pos2 + sep.size();
        pos2 = src.find(sep, pos1);
    }

    if (pos1 != src.length()) {
        tmpStr = src.substr(pos1);
        out.emplace_back(tmpStr);
    }
}

inline void StrUtil::StrTrim(std::string &src)
{
    if (src.empty()) {
        return;
    }

    src.erase(0, src.find_first_not_of(' '));
    src.erase(src.find_last_not_of(' ') + 1);
}

}
}

#endif // OCK_MEMFS_STR_UTIL_H