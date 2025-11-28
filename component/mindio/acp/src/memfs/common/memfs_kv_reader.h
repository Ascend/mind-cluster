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
#ifndef OCK_MEMFS_KV_READER_H
#define OCK_MEMFS_KV_READER_H

#include <cstring>
#include <fstream>
#include <map>
#include <mutex>
#include <string>
#include <vector>
#include <unordered_map>

#include "memfs_str_util.h"

#ifndef UNLIKELY
#define UNLIKELY(x) (__builtin_expect(!!(x), 0) != 0)
#endif

namespace ock {
namespace memfs {
struct KvPair {
    std::string name;
    std::string value;
};

class KVReader {
public:
    KVReader() = default;
    ~KVReader();

    KVReader(const KVReader &) = delete;
    KVReader &operator = (const KVReader &) = delete;
    KVReader(const KVReader &&) = delete;
    KVReader &operator = (const KVReader &&) = delete;

    bool FromFile(const std::string &filePath);

    bool SetItem(const std::string &key, const std::string &value);

    uint32_t Size();
    void GetI(uint32_t index, std::string &outKey, std::string &outValue);

private:
    std::map<std::string, uint32_t> mItemsIndex;
    std::vector<KvPair *> mItems;
    std::mutex mLock;
};

inline KVReader::~KVReader()
{
    {
        std::lock_guard<std::mutex> guard(mLock);
        for (auto &val : mItems) {
            KvPair *p = val;
            delete (p);
        }

        mItems.clear();
        mItemsIndex.clear();
    }
}

inline bool KVReader::FromFile(const std::string &filePath)
{
    char path[PATH_MAX + 1] = {0x00};
    if (strlen(filePath.c_str()) > PATH_MAX || realpath(filePath.c_str(), path) == nullptr) {
        return false;
    }

    /* open file to read */
    std::ifstream inConfFile(path);
    if (!inConfFile) {
        return false;
    }

    bool result = true;
    std::string strLine;
    while (getline(inConfFile, strLine)) {
        StrUtil::StrTrim(strLine);
        /* skip the line start with # */
        if (strLine.empty() || strLine[0] == '#') {
            continue;
        }

        /* skip the line without = */
        std::string::size_type equalDivPos = 0;
        if (std::string::npos == (equalDivPos = strLine.find('='))) {
            continue;
        }

        /* extract the line the value before = is the key, the value after = is the value, after trim */
        std::string strKey = strLine.substr(0, equalDivPos);
        std::string strValue = strLine.substr(equalDivPos + 1, strLine.size() - 1);
        StrUtil::StrTrim(strKey);
        StrUtil::StrTrim(strValue);

        /* skip the empty key */
        if (strKey.empty()) {
            continue;
        }

        /* set key value */
        if (!SetItem(strKey, strValue)) {
            result = false;
            break;
        }
    }

    inConfFile.close();
    inConfFile.clear();

    return result;
}

inline bool KVReader::SetItem(const std::string &key, const std::string &value)
{
    std::lock_guard<std::mutex> guard(mLock);
    auto iter = mItemsIndex.find(key);
    if (iter != mItemsIndex.end()) {
        mItems.at(iter->second)->value = value;
    } else {
        // check nullptr
        auto *kv = new (std::nothrow) KvPair();
        if (UNLIKELY(kv == nullptr)) {
            return false;
        }
        kv->name = key;
        kv->value = value;
        mItems.push_back(kv);
        mItemsIndex[key] = mItems.size() - 1;
    }
    return true;
}

inline uint32_t KVReader::Size()
{
    std::lock_guard<std::mutex> guard(mLock);
    return mItems.size();
}

inline void KVReader::GetI(uint32_t index, std::string &outKey, std::string &outValue)
{
    std::lock_guard<std::mutex> guard(mLock);
    if (index >= mItems.size()) {
        return;
    }

    outKey = mItems.at(index)->name;
    outValue = mItems.at(index)->value;
}
}
}

#endif // OCK_MEMFS_KV_READER_H
