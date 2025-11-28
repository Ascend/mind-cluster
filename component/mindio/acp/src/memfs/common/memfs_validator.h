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
#ifndef OCK_MEMFS_VALIDATOR_H
#define OCK_MEMFS_VALIDATOR_H

#include <climits>
#include <string>

#include "memfs_common.h"

namespace ock {
namespace memfs {
class Validator;
using ValidatorPtr = std::shared_ptr<Validator>;

class Validator {
public:
    virtual ~Validator() = default;

    virtual bool Initialize() = 0;

    virtual bool Validate(const std::string &)
    {
        return true;
    }

    virtual bool Validate(int)
    {
        return true;
    }

    virtual bool Validate(uint32_t)
    {
        return true;
    }

    virtual bool Validate(uint64_t)
    {
        return true;
    }

    virtual bool Validate(float)
    {
        return true;
    }

    virtual bool Validate(long)
    {
        return true;
    }

    const std::string &ErrorMessage()
    {
        return mErrMsg;
    }

protected:
    explicit Validator(const std::string &name) : mName(name) {}

protected:
    std::string mName;
    std::string mErrMsg;
};

class VStrNotNull : public Validator {
public:
    static ValidatorPtr Create(const std::string &name)
    {
        return ValidatorPtr(new (std::nothrow) VStrNotNull(name));
    }

    explicit VStrNotNull(const std::string &name) : Validator(name) {};

    ~VStrNotNull() override = default;

    bool Initialize() override
    {
        return true;
    }

    bool Validate(const std::string &value) override
    {
        if (value.empty()) {
            mErrMsg = "Invalid value for <" + mName + ">, it should not be empty";
            return false;
        }
        return true;
    }
};

class VIntRange : public Validator {
public:
    static ValidatorPtr Create(const std::string &name, const int &start, const int &end)
    {
        return ValidatorPtr(new (std::nothrow) VIntRange(name, start, end));
    }
    VIntRange(const std::string &name, const int &start, const int &end) : Validator(name), mStart(start), mEnd(end) {};

    ~VIntRange() override = default;

    bool Initialize() override
    {
        if (mStart >= mEnd) {
            mErrMsg = "Failed to initialize validator for <" + mName + ">, because end should be bigger than start";
            return false;
        }
        return true;
    }

    bool Validate(int value) override
    {
        if (value < mStart || value > mEnd) {
            if (mEnd == INT32_MAX) {
                mErrMsg = "Invalid value for <" + mName + ">, it should be >= " + std::to_string(mStart);
            } else {
                mErrMsg = "Invalid value for <" + mName + ">, it should be between " + std::to_string(mStart) + "~" +
                    std::to_string(mEnd);
            }
            return false;
        }

        return true;
    }

private:
    int mStart;
    int mEnd;
};

}
}

#endif // OCK_MEMFS_VALIDATOR_H
