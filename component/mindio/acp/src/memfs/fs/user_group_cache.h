/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
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

#ifndef OCKIO_USER_GROUP_CACHE_H
#define OCKIO_USER_GROUP_CACHE_H
#include <sys/types.h>
#include <pwd.h>
#include <grp.h>

#include <chrono>
#include <memory>
#include <mutex>
#include <map>
#include <list>
#include <unordered_map>
#include <unordered_set>
#include <functional>
#include <utility>

#include "memfs_common.h"
#include "common_locker.h"

namespace ock {
namespace memfs {
/**
 * @brief 一个用户的信息，包含一个user id和对应的group id
 */
class UserInfo {
public:
    UserInfo(uid_t uid, gid_t gid) noexcept;
    uid_t GetUserId() const noexcept;
    gid_t GetGroupId() const noexcept;

private:
    const uid_t userId;
    const gid_t groupId;
};

/**
 * @brief 一个组的信息，包含一个 group_id和加入此组的一批用户
 */
class GroupInfo {
public:
    explicit GroupInfo(gid_t gid, std::unordered_set<uid_t> users = std::unordered_set<uid_t>()) noexcept;
    gid_t GetGroupId() const noexcept;
    bool Contains(uid_t uid) const noexcept;

private:
    const gid_t groupId;
    const std::unordered_set<uid_t> usersInGroup;
};

/**
 * @brief 缓存的数据类型，可能是用户信息或组信息，同时含有一个时间时间戳信息，单位到微秒
 * @tparam T 泛型类型，可以是 UserInfo 或 GroupInfo
 */
template <class T> class CachedItem {
public:
    explicit CachedItem(const std::shared_ptr<T> &item) noexcept
        : timestamp{ GetCurrentTimestamp() }, cachedItem{ item }
    {}

    inline int64_t GetTimeStamp() const noexcept
    {
        return timestamp;
    }

    inline const std::shared_ptr<T> &GetItem() const noexcept
    {
        return cachedItem;
    }

    inline void RefreshItem(const std::shared_ptr<T> &item) noexcept
    {
        cachedItem = item;
        timestamp = GetCurrentTimestamp();
    }

    inline bool Expired(int64_t macroSeconds) const noexcept
    {
        return GetCurrentTimestamp() - timestamp > macroSeconds;
    }

    inline bool TimeStampBefore(int64_t expiredTimestamp) const noexcept
    {
        return timestamp < expiredTimestamp;
    }

    static int64_t GetCurrentTimestamp() noexcept
    {
        auto now = std::chrono::steady_clock::now();
        auto nowMacroSeconds = std::chrono::time_point_cast<std::chrono::microseconds>(now);
        return nowMacroSeconds.time_since_epoch().count();
    }

private:
    int64_t timestamp;
    std::shared_ptr<T> cachedItem;
};

using UserInfoPtr = std::shared_ptr<UserInfo>;
using GroupInfoPtr = std::shared_ptr<GroupInfo>;

/**
 * @brief 抽象类型，抽象如果读取用户和组信息
 */
class UserGroupLoader {
public:
    virtual ~UserGroupLoader() = default;
    virtual UserInfoPtr LoadUser(uid_t uid) const noexcept = 0;
    virtual UserInfoPtr LoadUser(const std::string &userName) const noexcept = 0;
    virtual GroupInfoPtr LoadGroup(gid_t gid) const noexcept = 0;
    virtual GroupInfoPtr LoadGroup(const std::string &groupName) const noexcept = 0;
};

/**
 * @brief 一个缓存信息管理
 * @tparam K 缓存的key类型
 * @tparam V 缓存的value类型
 */
template <class K, class V> class ItemCacheManager {
public:
    explicit ItemCacheManager(const std::function<std::shared_ptr<V>(K)> &ld) noexcept : loader{ ld } {}

public:
    std::shared_ptr<V> Get(K key, int64_t timeout) noexcept
    {
        // read from cache first
        auto expiredTimestamp = CachedItem<V>::GetCurrentTimestamp() - timeout;
        {
            RwLockGuard lockGuard{ lock, true };
            auto pos = items.find(key);
            if (pos != items.end() && !pos->second.Expired(timeout)) {
                return pos->second.GetItem();
            }
        }

        // need reload
        RwLockGuard lockGuard{ lock, false };
        auto pos = items.find(key);
        if (pos != items.end()) {
            if (!pos->second.TimeStampBefore(expiredTimestamp)) { // valid
                return pos->second.GetItem();
            }

            // expired
            RemoveExpiresInLock(expiredTimestamp);
        }

        // add item to cache
        auto cachedValue = loader(key);
        if (cachedValue == nullptr) {
            return nullptr;
        }

        CachedItem<V> cachedItem{ cachedValue };
        items.emplace(key, cachedItem);
        sortedKeys.emplace_back(key);
        RemoveExpiresInLock(expiredTimestamp);

        return cachedValue;
    }

private:
    std::unordered_set<K> RemoveExpiresInLock(int64_t expiredTimestamp) noexcept
    {
        std::unordered_set<K> result;
        while (!sortedKeys.empty()) {
            auto kp = sortedKeys.begin();

            auto vp = items.find(*kp);
            if (vp == items.end()) {
                sortedKeys.erase(kp);
                continue;
            }

            if (!vp->second.TimeStampBefore(expiredTimestamp)) {
                break;
            }

            result.emplace(*kp);
            sortedKeys.erase(kp);
            items.erase(vp);
        }

        return result;
    }

private:
    std::function<std::shared_ptr<V>(K)> loader;
    std::unordered_map<K, CachedItem<V>> items;
    std::list<K> sortedKeys;
    ReadWriteLock lock;
};

/**
 * @brief User与Group的cache，用于判断一个用户是否属于一个组
 */
class UserGroupCache {
public:
    static std::shared_ptr<UserGroupCache> GetSystemInstance() noexcept;

public:
    explicit UserGroupCache(const std::shared_ptr<UserGroupLoader> &loader, int64_t expiresUs = -1L) noexcept;
    virtual ~UserGroupCache() noexcept;

public:
    bool UserInGroup(uid_t user, gid_t group) noexcept;

private:
    UserInfoPtr GetUsersWithCache(uid_t user) noexcept;
    GroupInfoPtr GetGroupWithCache(gid_t group) noexcept;

private:
    std::shared_ptr<UserGroupLoader> userGroupLoader;
    const int64_t expiresMacroSeconds;
    ItemCacheManager<uid_t, UserInfo> users;
    ItemCacheManager<gid_t, GroupInfo> groups;
};

/**
 * @brief 系统的用户加载器
 */
class SystemUserGroupLoader : public UserGroupLoader {
public:
    UserInfoPtr LoadUser(uid_t uid) const noexcept override;
    UserInfoPtr LoadUser(const std::string &userName) const noexcept override;
    GroupInfoPtr LoadGroup(gid_t gid) const noexcept override;
    GroupInfoPtr LoadGroup(const std::string &groupName) const noexcept override;
    static int GetUidByName(const std::string &name, uid_t &userId) noexcept;

private:
    using LowGetUserFun = std::function<int(struct passwd &, char *, size_t, struct passwd *&)>;
    using LowGetGroupFun = std::function<int(struct group &, char *, size_t, struct group *&)>;

    static GroupInfoPtr ParseGroupInfo(struct group *group) noexcept;
    static GroupInfoPtr GetGroup(const LowGetGroupFun &low, const std::string &info) noexcept;
    static UserInfoPtr GetUser(const LowGetUserFun &low, const std::string &info) noexcept;
};

class SystemUserGroupWrapper {
public:
    static int GetPwNam(const std::string &name, struct passwd *pwd, char *buf, size_t bufLen, struct passwd **result);
    static int GetPwUid(uid_t uid, struct passwd *pwd, char *buf, size_t bufLen, struct passwd **result);
    static int GetGrNam(const std::string &name, struct group *grp, char *buf, size_t bufLen, struct group **result);
    static int GetGrGid(gid_t gid, struct group *grp, char *buf, size_t bufLen, struct group **result);
};
}
}

#endif // OCKIO_USER_GROUP_CACHE_H
