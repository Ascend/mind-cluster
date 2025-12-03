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
#ifndef OCK_MEMFS_CONFIGURATION_H
#define OCK_MEMFS_CONFIGURATION_H

#include <map>
#include <string>
#include <iostream>
#include <sstream>

#include "memfs_common.h"
#include "memfs_file_util.h"
#include "memfs_kv_reader.h"
#include "memfs_validator.h"

namespace ock {
namespace memfs {

enum ConfValueType {
    VINT = 0,
    VFLOAT = 1,
    VSTRING = 2,
    VBOOL = 3,
    VLONG = 4,
    VUINT = 5,
    VULONG = 6,
};

using ValidatorTag = uint16_t;

class Configuration {
public:
    static bool ReadConf(const std::string &path, KVReader &kv)
    {
        /* check path exist or not */
        if (!FileUtil::Exist(path)) {
            std::cout << "Load conf failed as file " << path << " doesn't exists" << std::endl;
            return false;
        }

        /* check with realpath */
        std::string realPath = path;
        if (!FileUtil::CanonicalPath(realPath)) {
            std::cout << "Load conf failed as file " << path << " is not real path" << std::endl;
            return false;
        }

        /* read kv from file */
        if (!kv.FromFile(realPath)) {
            std::cout << "Load conf failed as load file " << path << " failed" << std::endl;
            return false;
        }

        return true;
    }

    template <class T> static bool ReadConf(const std::string &path, bool stopIfInvalid = true)
    {
        KVReader kv;
        bool ret = ReadConf(path, kv);
        if (!ret) {
            std::cout << "Failed to create key/value reader." << std::endl;
            return false;
        }

        auto conf = GetInstance<T>();
        if (conf.get() == nullptr) {
            std::cout << "Failed to get instance for config." << std::endl;
            return false;
        }

        /* set kv into conf */
        uint32_t size = kv.Size();
        for (uint32_t i = 0; i < size; i++) {
            std::string key;
            std::string value;
            kv.GetI(i, key, value);
            if (!conf->SetWithTypeAutoConvert(key, value) && stopIfInvalid) {
                std::cout << "Failed to set a key/value pair for key<" << key << "> value<" << value << ">" <<
                    std::endl;
                return false;
            }
        }

        return true;
    }

    using Ptr = std::shared_ptr<Configuration>;

    template <class T> static Ptr GetInstance()
    {
        static Ptr gInstance = nullptr;
        static std::mutex gLock;

        if (gInstance == nullptr) {
            std::lock_guard<std::mutex> guard(gLock);
            if (gInstance != nullptr) {
                return gInstance;
            }

            gInstance = std::make_shared<T>();
            if (gInstance == nullptr) {
                std::cout << "Failed to new configuration object, probably out of memory" << std::endl;
                return nullptr;
            }

            /* load default conf */
            gInstance->LoadDefaultConf();
        }

        return gInstance;
    }

public:
    Configuration() = default;
    virtual ~Configuration() = default;

    int32_t GetInt(const std::string &key, int32_t defaultValue = 0);
    const std::string &GetStr(const std::string &key, const std::string &defaultValue = "");
    bool GetBool(const std::string &key, bool defaultValue = false);

    void AddIntConf(const std::pair<std::string, int> &, const ValidatorPtr &validator = nullptr, ValidatorTag tag = 0);

    void AddStrConf(const std::pair<std::string, std::string> &, const ValidatorPtr &validator = nullptr,
        ValidatorTag tag = 0);
    void AddBoolConf(const std::pair<std::string, bool> &);

    void AddValidator(const std::string &key, const ValidatorPtr &validator, ValidatorTag tag);

    std::vector<std::string> Validate(ValidatorTag tag = 0);

protected:
    virtual void LoadDefaultConf() {};

protected:
    bool SetWithTypeAutoConvert(const std::string &key, const std::string &value, bool skipIfLack = false);

    void ValidateOneType(const std::string &, const ValidatorPtr &, const ConfValueType &, std::vector<std::string> &);

    void ValidateUintType(const std::string &, const ValidatorPtr &, std::vector<std::string> &);

    void ValidateULongType(const std::string &, const ValidatorPtr &, std::vector<std::string> &);

protected:
    std::map<std::string, int32_t> mIntItems;     /* int32_t value config */
    std::map<std::string, uint32_t> mUintItems;   /* uint32_t value config */
    std::map<std::string, uint64_t> mULongItems;  /* uint64_t value config */
    std::map<std::string, float> mFloatItems;     /* float value config */
    std::map<std::string, std::string> mStrItems; /* string value config */
    std::map<std::string, bool> mBoolItems;       /* bool value config */
    std::map<std::string, long> mLongItems;       /* long value config */

    std::map<std::string, ConfValueType> mValueTypes; /* value type of keys */

    std::map<ValidatorTag, std::map<std::string, ValidatorPtr>> mTagValueValidator; /* validator */

    std::mutex mMutex;
};

using ConfigurationPtr = Configuration::Ptr;

inline bool Configuration::SetWithTypeAutoConvert(const std::string &key, const std::string &value, bool skipIfLack)
{
    ConfValueType valueType = ConfValueType::VSTRING;
    {
        auto iter = mValueTypes.find(key);
        if (iter != mValueTypes.end()) {
            valueType = iter->second;
        } else if (skipIfLack) {
            return true;
        } else {
            std::cout << "<" << key << ">, it is an unknown key, skip it." << std::endl;
            return false;
        }

        if (valueType == ConfValueType::VINT) {
            long tmp = 0;
            if (!StrUtil::StrToLong(value, tmp)) {
                std::cout << "<" << key << ">, it was empty or in wrong type, it should be a int number." << std::endl;
                return false;
            }
            mIntItems[key] = static_cast<int32_t>(tmp);
        } else if (valueType == ConfValueType::VFLOAT) {
            if (!StrUtil::StrToFloat(value, mFloatItems[key])) {
                std::cout << "<" << key << ">, it was empty or in wrong type, it should be a float number." <<
                    std::endl;
                return false;
            }
        } else if (valueType == ConfValueType::VSTRING) {
            mStrItems[key] = value;
        } else if (valueType == ConfValueType::VBOOL) {
            bool b = false;
            std::istringstream(value) >> std::boolalpha >> b;
            mBoolItems[key] = b;
        } else if (valueType == ConfValueType::VLONG) {
            if (!StrUtil::StrToLong(value, mLongItems[key])) {
                std::cout << "<" << key << ">, it was empty or in wrong type, it should be a long number." << std::endl;
                return false;
            }
        } else if (valueType == ConfValueType::VUINT) {
            if (!StrUtil::StrToUint(value, mUintItems[key])) {
                std::cout << "<" << key << ">, it was empty or in wrong type, it should be a uint number." << std::endl;
                return false;
            }
        } else if (valueType == ConfValueType::VULONG) {
            if (!StrUtil::StrToULong(value, mULongItems[key])) {
                std::cout << "<" << key << ">, it was empty or in wrong type, it should be a ulong number" << std::endl;
                return false;
            }
        }

        return true;
    }
}

inline int32_t Configuration::GetInt(const std::string &key, int32_t defaultValue)
{
    std::lock_guard<std::mutex> guard(mMutex);
    auto iter = mIntItems.find(key);
    if (iter != mIntItems.end()) {
        return iter->second;
    }

    return defaultValue;
}

inline const std::string &Configuration::GetStr(const std::string &key, const std::string &defaultValue)
{
    std::lock_guard<std::mutex> guard(mMutex);
    auto iter = mStrItems.find(key);
    if (iter != mStrItems.end()) {
        return iter->second;
    }
    return defaultValue;
}

inline bool Configuration::GetBool(const std::string &key, bool defaultValue)
{
    std::lock_guard<std::mutex> guard(mMutex);
    auto iter = mBoolItems.find(key);
    if (iter != mBoolItems.end()) {
        return iter->second;
    }
    return defaultValue;
}

inline void Configuration::AddValidator(const std::string &key, const ValidatorPtr &validator, ValidatorTag tag)
{
    if (UNLIKELY(key.empty())) {
        std::cout << "Failed to added validator as key is empty" << std::endl;
        return;
    } else if (validator == nullptr) {
        return;
    }

    auto tagIter = mTagValueValidator.find(tag);
    if (tagIter == mTagValueValidator.end()) {
        /* no tag added yet, add new map */
        std::map<std::string, ValidatorPtr> tmpMap;
        tmpMap[key] = validator;
        mTagValueValidator.emplace(tag, tmpMap);
    } else {
        /* already exist */
        tagIter->second[key] = validator;
    }
}

inline void Configuration::AddIntConf(const std::pair<std::string, int> &pair, const ValidatorPtr &validator,
    ValidatorTag tag)
{
    mIntItems[pair.first] = pair.second;
    mValueTypes[pair.first] = ConfValueType::VINT;
    AddValidator(pair.first, validator, tag);
}

inline void Configuration::AddStrConf(const std::pair<std::string, std::string> &pair, const ValidatorPtr &validator,
    ValidatorTag tag)
{
    // check nullptr
    mStrItems[pair.first] = pair.second;
    mValueTypes[pair.first] = ConfValueType::VSTRING;
    AddValidator(pair.first, validator, tag);
}

inline void Configuration::AddBoolConf(const std::pair<std::string, bool> &pair)
{
    mBoolItems[pair.first] = pair.second;
    mValueTypes[pair.first] = ConfValueType::VBOOL;
}

inline void Configuration::ValidateUintType(const std::string &key, const ValidatorPtr &validator,
    std::vector<std::string> &errors)
{
    auto valueIter = mUintItems.find(key);
    if (valueIter == mUintItems.end()) {
        errors.push_back("Failed to find <" + key + "> in uint value map, which should not happen.");
        return;
    }

    // validate the value
    if (!(validator->Validate(valueIter->second))) {
        errors.push_back(validator->ErrorMessage());
    }
    return;
}

inline void Configuration::ValidateULongType(const std::string &key, const ValidatorPtr &validator,
    std::vector<std::string> &errors)
{
    auto valueIter = mULongItems.find(key);
    if (valueIter == mULongItems.end()) {
        errors.push_back("Failed to find <" + key + "> in unsigned long value map, which should not happen.");
        return;
    }

    // validate the value
    if (!(validator->Validate(valueIter->second))) {
        errors.push_back(validator->ErrorMessage());
    }
    return;
}

inline void Configuration::ValidateOneType(const std::string &key, const ValidatorPtr &validator,
    const ConfValueType &vType, std::vector<std::string> &errors)
{
    // find the configured value according to the type
    if (vType == ConfValueType::VSTRING) {
        auto valueIter = mStrItems.find(key);
        if (valueIter == mStrItems.end()) {
            errors.push_back("Failed to find <" + key + "> in string value map, which should not happen.");
            return;
        }

        // validate the value
        if (!(validator->Validate(valueIter->second))) {
            errors.push_back(validator->ErrorMessage());
            return;
        }
    } else if (vType == ConfValueType::VFLOAT) {
        auto valueIter = mFloatItems.find(key);
        if (valueIter == mFloatItems.end()) {
            errors.push_back("Failed to find <" + key + "> in float value map, which should not happen.");
            return;
        }

        // validate the value
        if (!(validator->Validate(valueIter->second))) {
            errors.push_back(validator->ErrorMessage());
            return;
        }
    } else if (vType == ConfValueType::VINT) {
        auto valueIter = mIntItems.find(key);
        if (valueIter == mIntItems.end()) {
            errors.push_back("Failed to find <" + key + "> in int value map, which should not happen.");
            return;
        }

        // validate the value
        if (!(validator->Validate(valueIter->second))) {
            errors.push_back(validator->ErrorMessage());
            return;
        }
    } else if (vType == ConfValueType::VUINT) {
        return ValidateUintType(key, validator, errors);
    } else if (vType == ConfValueType::VULONG) {
        return ValidateULongType(key, validator, errors);
    } else if (vType == ConfValueType::VLONG) {
        auto valueIter = mLongItems.find(key);
        if (valueIter == mLongItems.end()) {
            errors.push_back("Failed to find <" + key + "> in long value map, which should not happen.");
            return;
        }

        // validate the value
        if (!(validator->Validate(valueIter->second))) {
            errors.push_back(validator->ErrorMessage());
            return;
        }
    }
}

inline std::vector<std::string> Configuration::Validate(ValidatorTag tag)
{
    std::vector<std::string> errors;
    auto tagValidatorIter = mTagValueValidator.find(tag);
    if (tagValidatorIter == mTagValueValidator.end()) {
        errors.emplace_back("No validator found for tag found");
        return errors;
    }

    /* validate one tag */
    for (auto &item : tagValidatorIter->second) {
        if (item.second.get() == nullptr) {
            errors.push_back("The validator of <" + item.first + "> is null, skip.");
            continue;
        } else {
            /* initialize, if failed then skip it */
            if (!(item.second->Initialize())) {
                errors.push_back(item.second->ErrorMessage());
                continue;
            }
        }

        /* firstly find the value type */
        auto typeIter = mValueTypes.find(item.first);
        if (typeIter == mValueTypes.end()) {
            errors.push_back("Failed to find <" + item.first + "> in type map, which should not happen.");
            continue;
        }

        ValidateOneType(item.first, item.second, typeIter->second, errors);
    }

    return errors;
}
}
}

#endif // OCK_MEMFS_CONFIGURATION_H
