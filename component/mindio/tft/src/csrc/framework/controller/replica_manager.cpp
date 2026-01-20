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

#include "replica_manager.h"
#include "controller.h"

namespace ock {
namespace ttp {

TResult DefaultReplicaManager::ChooseRank(const RankChooseInfo &rankChooseInfo, std::vector<int32_t> &tmpRankVec,
                                          uint32_t repCnt, uint32_t repShift)
{
    TTP_ASSERT_RETURN(repCnt != 0, TTP_ERROR);

    auto rankMask = GenerateRankMask(rankChooseInfo);
    auto rankSize = rankMask.size();
    // 1. 每个rank都有全量数据,选一个好的就行
    if (rankSize == repCnt) {
        for (auto [rank, mask] : rankMask) {
            if (mask == MASK_NORMAL) {
                tmpRankVec.push_back(rank);
                return TTP_OK;
            }
        }
        return TTP_ERROR;
    }

    // 2. 两副本或多副本, 在dp组内要选出全量信息
    auto offset = rankSize / repCnt;
    for (auto i = 0U; i < offset; i++) {
        for (auto idx = i; idx < rankSize; idx += offset) {
            auto [rank, mask] = rankMask[idx];
            if (mask == MASK_NORMAL) {
                tmpRankVec.push_back(rank);
                break;
            }
        }
    }

    return tmpRankVec.size() == offset  ? TTP_OK : TTP_ERROR;
}

RankMask DefaultReplicaManager::GenerateRankMask(const RankChooseInfo &rankChooseInfo)
{
    auto &[step, errorRanks, rankVec] = rankChooseInfo;

    RankMask rankMask;
    for (auto rank : rankVec) {
        rankMask.emplace_back(rank, MASK_NORMAL);
    }

    AutoLock statusMapLock(Controller::GetInstance()->statusMapLock_, TYPE_READ);
    for (auto &[curRank, mask] : rankMask) {
        if (errorRanks.find(curRank) != errorRanks.end()) {
            mask = MASK_ERROR;
            continue;
        }

        auto it = Controller::GetInstance()->statusMap_.find(curRank);
        if (it == Controller::GetInstance()->statusMap_.end()) {
            mask = MASK_ERROR;
            continue;
        }

        if (it->second.data_aval != TTP_STATUS_NORMAL) {
            mask = MASK_ERROR;
            continue;
        }

        bool err = (it->second.data_status != Updated || it->second.step != step);
        mask = err ? MASK_ERROR : MASK_NORMAL;

        if (err) {
            TTP_LOG_WARN("rank mask error, rank:" << curRank << ", expect step:" << step
                                                  << ", actual step:" << it->second.step
                                                  << ", data_status:" << static_cast<int>(it->second.data_status));
        }
    }

    return std::move(rankMask);
}

TResult DefaultReplicaManager::RepairSelectReplica(RankMask &rankMask, std::vector<RepairInfo> &rInfo, uint32_t repCnt,
                                                   uint32_t repShift, int16_t groupIdx)
{
    TTP_ASSERT_RETURN(repCnt != 0, TTP_ERROR);

    auto rankSize = rankMask.size();
    TTP_ASSERT_RETURN(rankSize != 0, TTP_ERROR);

    for (auto i = 0U; i < rankSize; i++) {
        auto [curRank, mask] = rankMask[i];
        if (mask == MASK_NORMAL || mask == MASK_UCE_LOW) {
            continue;
        }

        RepairType rt = mask == MASK_UCE_HIGH ? RepairType::RT_UCE_HIGHLEVEL : RepairType::RT_RECV_REPAIR;

        auto offset = rankSize / repCnt;
        bool find = false;
        for (uint32_t idx = (i + offset) % rankSize; idx != i; idx = (idx + offset) % rankSize) {
            auto [repRank, repMask] = rankMask[idx];
            if (repMask == MASK_NORMAL) {
                find = true;
                rInfo.push_back(RepairInfo{curRank, repRank, curRank, groupIdx, -1, rt});
                rInfo.push_back(RepairInfo{repRank, repRank, curRank, groupIdx, -1, RepairType::RT_SEND});
                break;
            }
        }

        if (!find) {
            TTP_LOG_ERROR("all rank is abnormal! rank:" << curRank);
            return TTP_ERROR;
        }
    }
    return TTP_OK;
}

bool DefaultReplicaManager::GetCanRepair()
{
    return canRepair_.load();
}

void DefaultReplicaManager::SetCanRepair(bool canRepair)
{
    return canRepair_.store(canRepair);
}

TResult X1ReplicaManager::ChooseRank(const RankChooseInfo &rankChooseInfo, std::vector<int32_t> &tmpRankVec,
                                     uint32_t repCnt, uint32_t repShift)
{
    TTP_ASSERT_RETURN(repCnt != 0, TTP_ERROR);

    auto rankMask = GenerateRankMask(rankChooseInfo);
    auto rankSize = rankMask.size();
    // 1. 每个rank都有全量数据,选一个好的就行
    if (rankSize == repCnt) {
        for (auto [rank, mask] : rankMask) {
            if (mask == MASK_NORMAL) {
                tmpRankVec.push_back(rank);
                return TTP_OK;
            }
        }
        return TTP_ERROR;
    }

    // 2. 两副本或多副本, 在副本组内要选出全量信息
    auto repSize = repShift * repCnt;
    for (auto i = 0U; i < repShift; i++) {
        for (auto idx = i; idx < repSize; idx += repShift) {
            auto [rank, mask] = rankMask[idx];
            if (mask == MASK_NORMAL) {
                tmpRankVec.push_back(rank);
                break;
            }
        }
    }

    // 3. 异形副本切分时支持dump条件判断
    if (tmpRankVec.size() != repShift) {
        SetCanRepair(false);
        tmpRankVec.clear();
        for (auto idx = repSize; idx < rankSize; idx++) {
            auto [rank, mask] = rankMask[idx];
            if (mask != MASK_NORMAL) {
                return TTP_ERROR;
            }
            tmpRankVec.push_back(rank);
        }
    }

    return TTP_OK;
}

TResult X1ReplicaManager::RepairSelectReplica(RankMask &rankMask, std::vector<RepairInfo> &rInfo, uint32_t repCnt,
                                              uint32_t repShift, int16_t groupIdx)
{
    TTP_ASSERT_RETURN(repCnt != 0, TTP_ERROR);
    auto repSize = repCnt * repShift;

    for (auto i = 0U; i < repSize; i++) {
        auto [curRank, mask] = rankMask[i];
        if (mask == MASK_NORMAL || mask == MASK_UCE_LOW) {
            continue;
        }

        RepairType rt = mask == MASK_UCE_HIGH ? RepairType::RT_UCE_HIGHLEVEL : RepairType::RT_RECV_REPAIR;

        bool find = false;
        for (uint32_t idx = (i + repShift) % repSize; idx != i; idx = (idx + repShift) % repSize) {
            auto [repRank, repMask] = rankMask[idx];
            if (repMask == MASK_NORMAL) {
                find = true;
                rInfo.push_back(RepairInfo{curRank, repRank, curRank, groupIdx, -1, rt});
                rInfo.push_back(RepairInfo{repRank, repRank, curRank, groupIdx, -1, RepairType::RT_SEND});
                break;
            }
        }

        if (!find) {
            TTP_LOG_ERROR("all rank is abnormal! rank:" << curRank);
            return TTP_ERROR;
        }
    }
    return TTP_OK;
}

}}