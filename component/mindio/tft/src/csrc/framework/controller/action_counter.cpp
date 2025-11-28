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

#include "action_counter.h"

namespace ock {
namespace ttp {

constexpr uint32_t BYTE_TO_BIT = 8;
constexpr uint64_t LOW_32BIT_MASK = 0xFFFFFFFF;

TResult AtomicStatusVector::Initialize(int32_t worldSize, uint32_t initStatus, uint32_t okStatus)
{
    if (worldSize <= 0) {
        TTP_LOG_ERROR("Illegal input worldSize: " << worldSize);
        return TTP_ERROR;
    }
    if (initStatus > INT32_MAX) {
        TTP_LOG_ERROR("Illegal initStatus: " << initStatus << ", which shall not be larger then INT32_MAX");
        return TTP_ERROR;
    }
    if (okStatus > INT32_MAX) {
        TTP_LOG_ERROR("Illegal okStatus: " << okStatus << ", which shall not be larger then INT32_MAX");
        return TTP_ERROR;
    }
    initStatus_ = initStatus;
    okStatus_ = okStatus;
    worldSize_ = worldSize;
    status_ = new (std::nothrow) std::atomic<uint64_t>[worldSize];
    if (status_ == nullptr) {
        TTP_LOG_ERROR("new std::atomic<uint64_t>[worldSize] failed, worldSize: " << worldSize);
        return TTP_ERROR;
    }
    uint64_t initVal = initStatus;  // ensure initStatus >= 0; default sn = 0
    for (int32_t i = 0; i < worldSize; i++) {
        status_[i].store(initVal);
    }
    return TTP_OK;
}

TResult AtomicStatusVector::GetRankSn(int32_t rank, uint16_t &sn)
{
    if (rank < 0 || rank >= worldSize_) {
        TTP_LOG_ERROR("Illegal rank: " << rank);
        return TTP_ERROR;
    }
    uint64_t val = status_[rank].load() >> (BYTE_TO_BIT * (sizeof(uint64_t) - sizeof(uint16_t)));
    sn = static_cast<uint16_t>(val);
    return TTP_OK;
}

uint64_t AtomicStatusVector::GenerateWholeStatus(uint16_t sn, uint32_t status)
{
    uint64_t val = sn;
    val = val << (BYTE_TO_BIT * (sizeof(uint64_t) - sizeof(uint16_t)));
    val += status;
    return val;
}

TResult AtomicStatusVector::InitRankStatus(const std::vector<int32_t> &ranks, uint16_t sn)
{
    for (uint32_t i = 0; i < ranks.size(); i++) {
        int32_t rank = ranks[i];
        if (rank < 0 || rank >= worldSize_) {
            TTP_LOG_ERROR("Illegal rank: " << rank);
            return TTP_ERROR;
        }
        uint64_t val = GenerateWholeStatus(sn, initStatus_);
        status_[rank].store(val);
    }
    return TTP_OK;
}

TResult AtomicStatusVector::CheckSetRankStatus(int32_t rank, uint16_t sn, uint32_t status, bool &isFistReply)
{
    isFistReply = false;
    if (status > INT32_MAX) {
        TTP_LOG_ERROR("Illegal status: " << status << ", which shall not be larger then INT32_MAX");
        return TTP_ERROR;
    }

    if (rank < 0 || rank >= worldSize_) {
        TTP_LOG_ERROR("Illegal rank: " << rank);
        return TTP_ERROR;
    }

    const uint64_t initVal = GenerateWholeStatus(sn, GetInitStatus());
    uint64_t currVal = GenerateWholeStatus(sn, status);
    bool isAtomic = false;
    while (!isAtomic) {
        uint64_t expectStatus = status_[rank].load();
        isFistReply = (expectStatus == initVal);
        uint16_t savedSn = 0;
        int32_t ret = GetRankSn(rank, savedSn);
        if (ret != TTP_OK) {
            return TTP_ERROR;
        }
        if (savedSn != sn) {
            TTP_LOG_ERROR("Event sn != curSn, status not set. rank: " << rank <<
                          ", Event sn: " << sn << ", curSn: " << savedSn);
            return TTP_ERROR;
        }
        isAtomic = status_[rank].compare_exchange_weak(expectStatus, currVal);
    }
    return TTP_OK;
}

TResult AtomicStatusVector::CheckRankGroupStatus(const std::vector<int32_t> &ranks, uint16_t sn)
{
    TResult checkRet = TTP_OK;
    uint64_t okVal = GenerateWholeStatus(sn, okStatus_);
    uint64_t initVal = GenerateWholeStatus(sn, initStatus_);
    for (uint32_t i = 0; i < ranks.size(); i++) {
        int32_t rank = ranks[i];
        if (rank < 0 || rank >= worldSize_) {
            TTP_LOG_ERROR("Illegal rank: " << rank);
            return TTP_ERROR;
        }
        uint16_t curSn = 0;
        GetRankSn(rank, curSn);
        if (curSn != sn) {
            TTP_LOG_INFO("Event sn is not equal to curSn. rank: " << rank <<
                         ", Event sn: " << sn << ", curSn: " << curSn);
            return TTP_ERROR;
        }
        if (status_[rank].load() == initVal) {
            return TTP_NEED_RETRY;
        }

        if (status_[rank].load() != okVal) {
            checkRet = TTP_ERROR;
        }
    }
    return checkRet;
}

TResult AtomicStatusVector::LoadRank(int32_t rank, uint16_t sn, uint64_t &result)
{
    result = 0;
    if (rank < 0 || rank >= worldSize_) {
        TTP_LOG_ERROR("Illegal rank: " << rank);
        return TTP_ERROR;
    }
    uint16_t curSn = 0;
    GetRankSn(rank, curSn);
    if (curSn != sn) {
        TTP_LOG_INFO("Event sn is not equal to curSn. rank: " << rank << ", Event sn: " << sn << ", curSn: " << curSn);
        return TTP_ERROR;
    }

    result = status_[rank].load();
    return TTP_OK;
}

TResult AtomicCounter::CheckAdd(uint16_t sn, uint32_t adder, uint64_t &addResult)
{
    bool isAtomic = false;
    while (!isAtomic) {
        uint64_t expectCounter = counter_.load();
        if (sn != GetSn()) {
            TTP_LOG_WARN("sn not equal, Event sn: " << sn << ", curSn: " << GetSn());
            addResult = 0;
            return TTP_ERROR;
        }
        if (UINT32_MAX - adder < GetCount()) {
            TTP_LOG_ERROR("integer wrap found, counter: " << GetCount() << ", adder: " << adder);
            addResult = 0;
            return TTP_ERROR;
        }
        addResult = expectCounter + adder;
        isAtomic = counter_.compare_exchange_weak(expectCounter, addResult);
    }
    return TTP_OK;
}

uint16_t AtomicCounter::GetSn()
{
    return static_cast<uint16_t>(counter_.load() >> (BYTE_TO_BIT * (sizeof(uint64_t) - sizeof(uint16_t))));
}

uint32_t AtomicCounter::GetCount()
{
    return static_cast<uint32_t>(counter_.load() & LOW_32BIT_MASK);
}

uint64_t AtomicCounter::Load()
{
    return counter_.load();
}

void AtomicCounter::Store(uint16_t sn, uint32_t count)
{
    uint64_t snTmp = static_cast<uint64_t>(sn) << (BYTE_TO_BIT * (sizeof(uint64_t) - sizeof(uint16_t)));
    uint64_t countTmp = count;
    counter_.store(snTmp | countTmp);
}

}  // namespace ttp
}  // namespace ock