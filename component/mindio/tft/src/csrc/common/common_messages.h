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

#ifndef OCK_TTP_MESSAGES_H
#define OCK_TTP_MESSAGES_H

#include "common_constants.h"
#include "common_macros.h"

using MResult = int32_t;

namespace ock {
namespace ttp {

struct TTPReplyMsg {
    MResult status;
    uint16_t sn;
    int32_t rank;
};

struct CkptMsg {
    int32_t repairId;   // build group with same pg_name
    uint32_t num;       // group number
    int64_t step;      // save step
    uint16_t sn;       // action serial number
    bool isTcpStoreOK; // if need starting tcp store client
    int32_t ranks[0];  // group rank list : (group_idx group_size group_ranklist) * group_num
    bool JudgeVariableValid(uint32_t msgLen, int32_t worldSize) const
    {
        int32_t pIdx = 0;
        TTP_ASSERT_RETURN(msgLen >= sizeof(CkptMsg), false);
        TTP_ASSERT_RETURN(num <= TTP_MAX_OPTIM_NUM, false);
        TTP_ASSERT_RETURN(step > 0 && step < INT64_MAX, false);
        uint32_t offset = sizeof(CkptMsg);
        for (uint32_t idx = 0; idx < num; idx++) {
            offset += sizeof(int32_t) + sizeof(int32_t);
            TTP_ASSERT_RETURN(msgLen > offset, false);
            TTP_ASSERT_RETURN(ranks[pIdx] >= 0 && ranks[pIdx] < TTP_MAX_OPTIM_NUM, false);
            int32_t rankSize = ranks[++pIdx];
            TTP_ASSERT_RETURN(rankSize > 0 && rankSize <= worldSize, false); // rank size
            offset += rankSize * sizeof(int32_t);
            TTP_ASSERT_RETURN(msgLen >= offset, false);
            for (int32_t i = 1; i <= rankSize; i++) {
                TTP_ASSERT_RETURN(ranks[pIdx + i] >= 0 && ranks[pIdx + i] < worldSize, false); // rank
            }
            pIdx += ranks[pIdx] + 1;
        }
        return true;
    }
};

struct RegisterMsg {
    int32_t rank;      // my rank number
};

struct ReplicaMsg {
    int32_t rank;
    int32_t num;
    bool enableArf;
    bool enableRetry;
    bool enableZit;
    int32_t ranks[0];   // replicaCnt + replicaShift + (rankSize + ranks) * groupNum
    bool JudgeVariableValid(uint32_t msgLen, int32_t worldSize) const
    {
        TTP_ASSERT_RETURN(msgLen >= sizeof(ReplicaMsg), false);
        TTP_ASSERT_RETURN(rank < worldSize, false);
        TTP_ASSERT_RETURN(num <= TTP_MAX_OPTIM_NUM && num > 0, false);
        int32_t pIdx = num + num;
        uint32_t offset = (pIdx) * sizeof(int32_t) + sizeof(ReplicaMsg);
        TTP_ASSERT_RETURN(msgLen > offset, false);
        const int allReplicaNum = -1;
        for (int32_t groupIdx = 0; groupIdx < num; groupIdx++) {
            int32_t replicaCnt = ranks[groupIdx];
            TTP_ASSERT_RETURN(replicaCnt > 0 || replicaCnt == allReplicaNum, false); // replica cnt
            TTP_ASSERT_RETURN(ranks[groupIdx + num] >= 0, false); // replica offset

            offset += sizeof(int32_t);
            TTP_ASSERT_RETURN(msgLen >= offset, false);
            int32_t rankSize = ranks[pIdx];
            TTP_ASSERT_RETURN(rankSize > 0 && replicaCnt <= rankSize && rankSize <= TTP_MAX_WORLD_SIZE, false);

            offset += rankSize * sizeof(int32_t);
            TTP_ASSERT_RETURN(msgLen >= offset, false);
            for (int32_t i = 1; i <= rankSize; i++) {
                TTP_ASSERT_RETURN(ranks[pIdx + i] >= 0 && ranks[pIdx + i] < worldSize, false); // rank
            }
            pIdx += rankSize + 1;
        }
        return true;
    }
};

struct DowngradeRunMsg {
    uint32_t num;      // group num(world group、dp_cp dp_ep)
    int32_t repairId;
    uint16_t sn;       // action serial number
    int32_t ranks[0];  // group rank list : (group_idx group_size group_ranklist) * group_num
    bool JudgeVariableValid(uint32_t msgLen, int32_t worldSize) const
    {
        int32_t pIdx = 0;
        TTP_ASSERT_RETURN(msgLen >= sizeof(DowngradeRunMsg), false);
        TTP_ASSERT_RETURN(num <= TTP_MAX_COMM_GROUP_NUM, false);
        uint32_t offset = sizeof(DowngradeRunMsg);
        for (uint32_t idx = 0; idx < num; idx++) {
            offset += sizeof(int32_t) + sizeof(int32_t);
            TTP_ASSERT_RETURN(msgLen > offset, false);
            int32_t rankSize = ranks[++pIdx];
            TTP_ASSERT_RETURN(rankSize > 0 && rankSize <= worldSize, false); // rank size
            offset += rankSize * sizeof(int32_t);
            TTP_ASSERT_RETURN(msgLen >= offset, false);
            for (int32_t i = 1; i <= rankSize; i++) {
                TTP_ASSERT_RETURN(ranks[pIdx + i] >= 0 && ranks[pIdx + i] < worldSize, false); // rank
            }
            pIdx += ranks[pIdx] + 1;
        }
        return true;
    }
};

struct HeartBeatMsg {
    int32_t rank;
    int32_t repairId;
    struct TrainStatus status;
    bool JudgeVariableValid() const
    {
        TTP_ASSERT_RETURN(rank >= 0 && rank < TTP_MAX_WORLD_SIZE, false);
        TTP_ASSERT_RETURN(status.step >= 0 && status.step < INT64_MAX, false);
        TTP_ASSERT_RETURN(status.backup_step >= -1 && status.backup_step < INT64_MAX, false);
        TTP_ASSERT_RETURN(status.run_status >= TTP_STATUS_NORMAL && status.run_status < TTP_STATUS_BUTT, false);
        TTP_ASSERT_RETURN(status.npu_status >= TTP_STATUS_NORMAL && status.npu_status < TTP_STATUS_BUTT, false);
        TTP_ASSERT_RETURN(status.data_aval >= TTP_STATUS_NORMAL && status.data_aval < TTP_STATUS_BUTT, false);
        TTP_ASSERT_RETURN(status.data_status >= TTP_STATUS_NORMAL && status.data_status < TTP_STATUS_BUTT, false);
        return true;
    }
};

struct ResultAndHBReplyMsg {
    TResult ret;
    uint16_t sn;
    int64_t repairStep = -1;
    struct HeartBeatMsg hb;
};

struct RepairMsg {
    int32_t repairId;       // The number of times that repair operations were performed
    uint32_t repairType;
    uint32_t repairNum;      // RepairMsgUnit number
    int64_t step;
    uint16_t sn;
    uint32_t rankNum;       // rank list number
    int32_t arr[0];         // rank list + RepairMsgUnit list + zit param
    bool JudgeVariableValid(uint32_t msgLen, int32_t worldSize)
    {
        TTP_ASSERT_RETURN(msgLen >= sizeof(RepairMsg), false);
        TTP_ASSERT_RETURN(repairType <= ControllerRepairType::CRT_BUTT, false);
        TTP_ASSERT_RETURN(repairNum <= worldSize, false);
        TTP_ASSERT_RETURN(rankNum <= worldSize, false);
        TTP_ASSERT_RETURN(step > 0 && step < INT64_MAX, false);
        TTP_ASSERT_RETURN(msgLen >=
                          sizeof(RepairMsg) + sizeof(int32_t) * rankNum + sizeof(RepairMsgUnit) * repairNum, false);
        for (uint32_t i = 0; i < rankNum; i++) {
            TTP_ASSERT_RETURN(arr[i] >= 0 && arr[i] < worldSize, false); // rank
        }

        RepairMsgUnit *uptr = reinterpret_cast<RepairMsgUnit *>(&arr[rankNum]);
        for (uint32_t j = 0; j < repairNum; j++) {
            TTP_ASSERT_RETURN(uptr[j].srcRank >= 0 && uptr[j].srcRank < worldSize, false);
            TTP_ASSERT_RETURN(uptr[j].dstRank >= 0 && uptr[j].dstRank < worldSize, false);
            TTP_ASSERT_RETURN(uptr[j].groupType >= 0 && uptr[j].groupType < TTP_MAX_OPTIM_NUM, false);
            TTP_ASSERT_RETURN(uptr[j].type >= RepairType::RT_SEND && uptr[j].type < RepairType::RT_BUTT, false);
        }
        return true;
    }
};

struct DpMsg {
    int32_t rank;
    uint32_t dpNum;
    int32_t dpList[0];
    bool JudgeVariableValid(uint32_t msgLen, int32_t worldSize) const
    {
        TTP_ASSERT_RETURN(msgLen >= sizeof(DpMsg), false);
        TTP_ASSERT_RETURN(rank < worldSize, false);
        TTP_ASSERT_RETURN(dpNum <= worldSize, false);
        TTP_ASSERT_RETURN(msgLen == sizeof(DpMsg) + sizeof(int32_t) * dpNum, false);
        for (int32_t i = 0; i < dpNum; i++) {
            TTP_ASSERT_RETURN(dpList[i] >= 0 && dpList[i] < worldSize, false);
        }
        return true;
    }
};

struct RebuildGroupMsg {
    uint32_t rankNum;    // dp rank list number
    int32_t repairId;
    uint16_t sn;
    int32_t ranks[0];
    bool JudgeVariableValid(uint32_t msgLen, int32_t worldSize)
    {
        TTP_ASSERT_RETURN(msgLen >= sizeof(RebuildGroupMsg), false);
        TTP_ASSERT_RETURN(rankNum <= worldSize, false);
        TTP_ASSERT_RETURN(msgLen >= sizeof(RebuildGroupMsg) + sizeof(int32_t) * rankNum, false);
        for (uint32_t i = 0; i < rankNum; i++) {
            TTP_ASSERT_RETURN(ranks[i] >= 0 && ranks[i] < worldSize, false); // rank
        }
        return true;
    }
};

struct RollbackMsg {
    int64_t step;
    int32_t repairId;   // build group with same pg_name
    RepairType type;
    uint16_t sn;
    uint32_t dataLen;   // string length + 1, including '\0'
    char data[0];
};

struct CommonMsg {
    int32_t rank;
    uint16_t sn;
};

struct BroadcastIpMsg {
    uint8_t enableARF;
    uint8_t enableZIT;
    uint16_t sn;
    uint32_t ipLen;       // char[] length, including '\0'
    char arr[0];        // ip:port

    bool JudgeVariableValid(uint32_t msgLen) const
    {
        TTP_ASSERT_RETURN(msgLen >= sizeof(BroadcastIpMsg), false);
        TTP_ASSERT_RETURN(enableARF <= static_cast<uint8_t>(true), false);
        TTP_ASSERT_RETURN(enableZIT <= static_cast<uint8_t>(true), false);
        TTP_ASSERT_RETURN(ipLen > 0, false);
        uint32_t broadcastIpMsgLen = sizeof(BroadcastIpMsg) + ipLen;
        TTP_ASSERT_RETURN(broadcastIpMsgLen == msgLen, false);
        return true;
    }
};

struct PrelockMsg {
    uint16_t sn;
    int64_t step;
};

struct PauseMsg {
    uint16_t sn;
    int64_t step;
    bool hotSwitch;
};

}  // namespace ttp
}  // namespace ock

#endif // OCK_TTP_MESSAGES_H