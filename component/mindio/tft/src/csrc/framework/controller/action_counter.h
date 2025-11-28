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

#ifndef OCK_TTP_ACTION_COUNTER_H
#define OCK_TTP_ACTION_COUNTER_H

#include "common.h"

namespace ock {
namespace ttp {

class AtomicStatusVector : public Referable {
public:
    AtomicStatusVector() : status_(nullptr) {};
    virtual ~AtomicStatusVector()
    {
        delete[] status_;
        status_ = nullptr;
    }
    TResult Initialize(int32_t worldSize, uint32_t initStatus, uint32_t okStatus);
    TResult GetRankSn(int32_t rank, uint16_t &sn);
    TResult InitRankStatus(const std::vector<int32_t> &ranks, uint16_t sn);
    TResult CheckSetRankStatus(int32_t rank, uint16_t sn, uint32_t status, bool &isFistReply);
    // status set OK if and only if every rank in ranks is OK; return value: if error occur; status: Check result.
    TResult CheckRankGroupStatus(const std::vector<int32_t> &ranks, uint16_t sn);
    uint32_t GetInitStatus() {return initStatus_;}
    uint32_t GetOkStatus() {return okStatus_;}
    TResult LoadRank(int32_t rank, uint16_t sn, uint64_t &result);
    uint64_t GenerateWholeStatus(uint16_t sn, uint32_t status);
private:
    std::atomic<uint64_t> *status_ { nullptr }; // uint64_t: [high 16bit: sn; middle 16bit: padding; low 32bit: status]
    int32_t worldSize_ { 0 };
    uint32_t initStatus_ { TTP_BUTT };
    uint32_t okStatus_ { TTP_OK };
};

using AtomicStatusPtr = Ref<AtomicStatusVector>;

class AtomicCounter : public Referable {
public:
    AtomicCounter() = default;
    ~AtomicCounter() override = default;
    TResult CheckAdd(uint16_t sn, uint32_t adder, uint64_t &addResult);
    uint16_t GetSn();
    uint32_t GetCount();
    uint64_t Load();
    void Store(uint16_t sn, uint32_t count);  // force write, without checking sn

private:
    std::atomic<uint64_t> counter_{ 0 };
};

}  // namespace ttp
}  // namespace ock
#endif