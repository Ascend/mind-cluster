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
#ifndef OCK_TTP_REPLICA_MANAGER_H
#define OCK_TTP_REPLICA_MANAGER_H

#include <vector>
#include "common.h"

namespace ock {
namespace ttp {

using GroupMap = std::set<std::vector<int32_t>>;
using RankMask = std::vector<std::pair<int32_t, uint8_t>>;

class DefaultReplicaManager {
public:
    virtual TResult ChooseRank(const RankChooseInfo &rankChooseInfo, std::vector<int32_t> &tmpRankVec, uint32_t repCnt,
                       uint32_t repShift);

    RankMask GenerateRankMask(const RankChooseInfo &rankChooseInfo);

    virtual TResult RepairSelectReplica(RankMask &rankMask, std::vector<RepairInfo> &rInfo,
                                uint32_t repCnt, uint32_t repShift, int16_t groupIdx);

    bool GetCanRepair();

protected:
    void SetCanRepair(bool canRepair);

protected:
    std::atomic<bool> canRepair_{true};
};

class X1ReplicaManager : public DefaultReplicaManager {
public:

    TResult ChooseRank(const RankChooseInfo &rankChooseInfo, std::vector<int32_t> &tmpRankVec, uint32_t repCnt,
                       uint32_t repShift) override;

    TResult RepairSelectReplica(RankMask &rankMask, std::vector<RepairInfo> &rInfo,
                                uint32_t repCnt, uint32_t repShift, int16_t groupIdx) override;
};

inline uint32_t GetFrameworkType()
{
    uint32_t typeValue = 0;
    const char *type = std::getenv("TTP_FRAMEWORK_TYPE");
    if (type != nullptr) {
        typeValue = String2Uint(type);
    }
    return typeValue;
}

inline DefaultReplicaManager& CreateReplicaManager()
{
    uint32_t typeValue = GetFrameworkType();

    switch (typeValue) {
        case FrameworkTypeEnum::TYPE_X1:
            static X1ReplicaManager x1manager;
            return x1manager;
        default:
            static DefaultReplicaManager manager;
            return manager;
    }
}

}}
#endif // OCK_TTP_REPLICA_MANAGER_H
