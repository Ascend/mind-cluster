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
#ifndef OCK_TTP_REF_H
#define OCK_TTP_REF_H

#include <cstdint>
#include <utility>

namespace ock {
namespace ttp {
class Referable {
public:
    Referable() = default;
    virtual ~Referable() = default;

    inline void IncreaseRef()
    {
        __sync_fetch_and_add(&mRefCount, 1);
    }

    inline void DecreaseRef()
    {
        // delete itself if reference count equal to 0
        if (__sync_sub_and_fetch(&mRefCount, 1) == 0) {
            delete this;
        }
    }

    inline int32_t GetRef()
    {
        return __sync_fetch_and_add(&mRefCount, 0);
    }

protected:
    int32_t mRefCount = 0;
};

#if __GNUC__ == 4 && __GNUC_MINOR__ == 8 && __GNUC_PATCHLEVEL__ == 5
template <class T, class U = T> T exchangeHdagger(T &obj, U &&new_value)
{
    T old_value = std::move(obj);
    obj = std::forward<U>(new_value);
    return old_value;
}
#endif

template <typename T> class Ref {
public:
    // constructor
    Ref() noexcept = default;

    // fix: can't be explicit
    Ref(T *newObj) noexcept
    {
        // if new obj is not null, increase reference count and assign to mObj
        // else nothing need to do as mObj is nullptr by default
        if (newObj != nullptr) {
            newObj->IncreaseRef();
            mObj = newObj;
        }
    }

    Ref(const Ref<T> &other) noexcept
    {
        // if other's obj is not null, increase reference count and assign to mObj
        // else nothing need to do as mObj is nullptr by default
        if (other.mObj != nullptr) {
            other.mObj->IncreaseRef();
            mObj = other.mObj;
        }
    }

#if __GNUC__ == 4 && __GNUC_MINOR__ == 8 && __GNUC_PATCHLEVEL__ == 5
    Ref(Ref<T> &&other) noexcept : mObj(exchangeHdagger(other.mObj, nullptr))
#else
    Ref(Ref<T> &&other) noexcept : mObj(std::__exchange(other.mObj, nullptr))
#endif
    {
        // move constructor
        // since this mObj is null, just exchange
    }

    // de-constructor
    ~Ref()
    {
        if (mObj != nullptr) {
            mObj->DecreaseRef();
        }
    }

    // operator =
    inline Ref<T> &operator = (T *newObj)
    {
        this->Set(newObj);
        return *this;
    }

    inline Ref<T> &operator = (const Ref<T> &other)
    {
        if (this != &other) {
            this->Set(other.mObj);
        }
        return *this;
    }

    Ref<T> &operator = (Ref<T> &&other) noexcept
    {
        if (this != &other) {
            auto tmp = mObj;
#if __GNUC__ == 4 && __GNUC_MINOR__ == 8 && __GNUC_PATCHLEVEL__ == 5
            mObj = exchangeHdagger(other.mObj, nullptr);
#else
            mObj = std::__exchange(other.mObj, nullptr);
#endif
            if (tmp != nullptr) {
                tmp->DecreaseRef();
            }
        }
        return *this;
    }

    // equal operator
    inline bool operator == (const Ref<T> &other) const
    {
        return mObj == other.mObj;
    }

    inline bool operator == (T *other) const
    {
        return mObj == other;
    }

    inline bool operator != (const Ref<T> &other) const
    {
        return mObj != other.mObj;
    }

    inline bool operator != (T *other) const
    {
        return mObj != other;
    }

    // get operator and set
    inline T *operator->() const
    {
        return mObj;
    }

    inline T *Get() const
    {
        return mObj;
    }

    inline void Set(T *newObj)
    {
        if (newObj == mObj) {
            return;
        }

        if (newObj != nullptr) {
            newObj->IncreaseRef();
        }

        if (mObj != nullptr) {
            mObj->DecreaseRef();
        }

        mObj = newObj;
    }

private:
    T *mObj = nullptr;
};

/*
 * @brief New an object return with ref object
 *
 * @param args             [in] args of object
 *
 * @return Ref object, if new failed internal, an empty Ref object will be returned
 */
template <typename C, typename... ARGS> static inline Ref<C> MakeRef(ARGS... args)
{
    return new (std::nothrow) C(args...);
}
}
}
#endif // OCK_TTP_REF_H
