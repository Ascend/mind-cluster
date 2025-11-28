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

#ifndef OCK_TTP_UTILS_H
#define OCK_TTP_UTILS_H

#include <map>
#include <set>
#include <vector>
#include <regex>
#include <unordered_map>

#include "common_constants.h"
#include "common_types.h"

namespace ock {
namespace ttp {
template<class T>
inline std::string IntVec2String(const std::vector<T> &vec)
{
    std::string ret = "[";
    uint32_t vecSize = vec.size();
    if (vecSize > 0) {
        for (auto val : vec) {
            ret += std::to_string(val) + ", ";
        }
    } else  {
        ret += ", ";
    }

    ret = ret.substr(0, ret.length() - 2);    // 2
    ret += "]";
    return ret;
}

template<class T>
inline std::string IntSet2String(const std::set<T> &set)
{
    std::string ret = "[";
    uint32_t setSize = set.size();
    if (setSize > 0) {
        for (auto val : set) {
            ret += std::to_string(val) + ", ";
        }
    } else  {
        ret += ", ";
    }

    ret = ret.substr(0, ret.length() - 2);    // 2
    ret += "]";
    return ret;
}

inline bool IsValidIpV4(const std::string& address)
{
    // 校验输入长度，防止正则表达式栈溢出
    constexpr size_t maxIpLen = 15;
    if (address.size() > maxIpLen) {
        return false;
    }
    std::regex ipV4Pattern("^(?:(?:25[0-5]|2[0-4]\\d|1\\d\\d|[1-9]?\\d)($|(?!\\.$)\\.)){4}$");
    std::regex zeroPattern("^0+\\.0+\\.0+\\.0+$");
#ifndef UT_ENABLED  // ut测试跳过ip全0校验
    if (std::regex_match(address, zeroPattern)) {
        return false;
    }
#endif
    if (!std::regex_match(address, ipV4Pattern)) {
        return false;
    }
    return true;
}

inline uint32_t String2Uint(const char *str)    // avoid throwing ex during converting to std::string
{
    try {
        auto num = std::stoul(str);
        if (num > UINT32_MAX) {
            return 0;
        }
        return static_cast<uint32_t>(num);
    } catch (...) {
        return 0;
    }
}

inline bool IsAllDigits(const std::string& str)
{
    if (str.empty()) {
        return false;
    }
    return std::all_of(str.begin(), str.end(), [](unsigned char ch) {
        return std::isdigit(ch);
    });
}

inline uint32_t GetEnvValue2Uint32(const char* envName, uint32_t minVal, uint32_t maxVal, uint32_t defaultVal)
{
    constexpr uint32_t maxUint32Len = 35;
    const char *tmpEnvValue = std::getenv(envName);
    uint32_t envValue = defaultVal;
    if (tmpEnvValue != nullptr && strlen(tmpEnvValue) <= maxUint32Len && IsAllDigits(tmpEnvValue)) {
        envValue = String2Uint(tmpEnvValue);
    }

    return (envValue < minVal || envValue > maxVal) ? defaultVal : envValue;
}

inline bool GetMsEnv(const char *envName)
{
    constexpr uint32_t maxMsEnvLen = 10;
    const char *envVal = std::getenv(envName);
    if (envVal == nullptr || strlen(envVal) > maxMsEnvLen) {
        return false;
    }
    std::string envStr(envVal);
    std::transform(envStr.begin(), envStr.end(), envStr.begin(), ::tolower);
    return envStr == "true" || envStr == "1";
}

inline bool IsOverflow(uint32_t a, uint32_t b)
{
    uint64_t ret = static_cast<uint64_t>(a) * static_cast<uint64_t>(b);
    return ret > UINT32_MAX;
}

inline std::string TrimString(const std::string &input)
{
    if (input.empty()) {
        return "";
    }
    auto start = input.begin();
    while (start != input.end() && std::isspace(*start)) {
        start++;
    }

    auto end = input.end();
    do {
        end--;
    } while (std::distance(start, end) > 0 && std::isspace(*end));

    return std::string(start, end + 1);
}

template <typename K, typename V>
inline std::set<K> GetMapKeysToSet(const std::map<K, V>& map)
{
    std::set<K> keys;
    std::transform(map.begin(), map.end(), std::inserter(keys, keys.begin()),
                   [](const std::pair<K, V>& p) { return p.first; });
    return std::move(keys);
}

template <typename K, typename V>
inline std::vector<K> GetMapKeysToVector(const std::map<K, V>& map)
{
    std::vector<K> keys;
    std::transform(map.begin(), map.end(), std::back_inserter(keys),
                   [](const std::pair<K, V>& p) { return p.first; });
    return std::move(keys);
}

template <typename K, typename V>
inline void SwapMapWithKeys(std::unordered_map<K, V>& map, std::unordered_map<K, V>& mapTmp)
{
    if (map.empty() || mapTmp.empty()) {
        return;
    }
    for (auto &[key, val] : mapTmp) {
        auto it = map.find(key);
        if (it != map.end()) {
            std::swap(val, it->second);
        }
    }
}

template <typename K, typename V>
inline void SwapMapWithVals(std::unordered_map<K, V>& map, std::unordered_map<K, V>& mapTmp)
{
    std::unordered_map<K, V> m;
    for (auto [key, val] : mapTmp) {
        for (auto [key_, val_] : map) {
            if (val == val_) {
                m[key_] = val_;
            }
        }
    }
    for (auto [key, val] : m) {
        map.erase(key);
    }
    if (!m.empty()) {
        map.insert(mapTmp.begin(), mapTmp.end());
        mapTmp.swap(m);
    }
}

}  // namespace ttp
}  // namespace ock
#endif  // OCK_TTP_UTILS_H
