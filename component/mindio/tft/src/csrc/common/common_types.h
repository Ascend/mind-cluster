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

#ifndef OCK_TTP_TYPES_H
#define OCK_TTP_TYPES_H

#include "common_constants.h"

using namespace ock::ttp;

namespace ock {
namespace ttp {
struct TrainStatus {
    int64_t step = 0;            // current update step
    int8_t run_status = 0;       // status of train process
    int8_t npu_status = 0;       // npu status
    int8_t data_aval = 0;        // if os data can dump from npu when npu is unhealthy
    int8_t data_status = 0;      // optimizer state data statue
    int64_t lastUpdateTime = 0;  // The number of milliseconds since the epoch
    int64_t backup_step = 0;     // backup step when localCopySwitch is on
};

struct RepairMsgUnit {
    int32_t srcRank;    // send rank
    int32_t dstRank;    // recv rank
    int16_t replicaIdx; // local rank's replica idx
    int16_t groupType;
    RepairType type;
};

struct CallbackCtx {
    int32_t rank;
    uint16_t sn;
};

struct RegisterReply {
    int32_t ret;
    int32_t repairId;
    bool hotSwitch = false;
};

struct EnvVarValue {
    uint32_t minVal;
    uint32_t maxVal;
    uint32_t defaultVal;
};

}  // namespace ttp
}  // namespace ock

#endif // OCK_TTP_TYPES_H
