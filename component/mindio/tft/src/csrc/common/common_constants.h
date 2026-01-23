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

#ifndef OCK_TTP_CONSTANTS_H
#define OCK_TTP_CONSTANTS_H

#include <cstdint>
#include <vector>
#include <string_view>

namespace ock {
namespace ttp {

constexpr uint32_t TTP_LINK_SEND_QUEUE_SIZE = 100;
#ifdef UT_ENABLED  // ut测试使用少量worker,提升运行效率
constexpr uint32_t TTP_SERVER_WORKER_COUNT = 2;
#else
constexpr uint32_t TTP_SERVER_WORKER_COUNT = 10;
#endif
constexpr uint32_t TTP_DEFAULT_START_VERSION = 1;
constexpr uint32_t TTP_SLEEP_TIME = 2;
constexpr uint32_t LOG_PRINT_INTERVAL = 5000;
constexpr uint32_t TTP_HEARTBEAT_CONNECT_THRESHOLD = 3;
constexpr uint32_t TTP_WAIT_TIME_1MS = 1000;
constexpr int32_t INVALID_RANK_ID = INT32_MAX;
constexpr uint32_t TTP_MAX_WORLD_SIZE = 100000;
constexpr int32_t TTP_MAX_OPTIM_NUM = 1000;  // 最大支持优化器套数
constexpr int32_t TTP_MAX_COMM_GROUP_NUM = 2000;  // 通信组重建时最大支持重建组数
constexpr int32_t TTP_MAX_ZIT_PARAM_LEN = 1024;  // zit参数string最大长度, char*加1
constexpr size_t MAX_CIPHER_LEN = 10 * 1024 * 1024;
constexpr int32_t TTP_ERROR_CODE_DATA_LEN = 16;  // 故障码数据长度

enum OptimStatus : int8_t {
    Updated = 0,
    Updating,
    Copying,
};

enum TResult : int {
    TTP_OK = 0,
    TTP_ERROR = 1,
    TTP_WAIT_CHECK = 2,
    TTP_STOP_SERVICE = 3,
    TTP_TIMEOUT = 4,
    TTP_NEED_RETRY = 5,
    TTP_DOWNGRADE = 6,
    TTP_REPAIR = 7,
    TTP_PAUSE = 8,
    TTP_SWITCH = 9,
    TTP_BUTT,
};

enum MsgOpCode : int16_t {
    TTP_MSG_OP_REGISTER = 0,    // register
    TTP_MSG_OP_INIT_REPORT,     // processor report replica info
    TTP_MSG_OP_DP_REPORT,       // processor report origin dp group
    TTP_MSG_OP_HEARTBEAT_SEND,  // processor send heart beat to controller
    TTP_MSG_OP_CKPT_SEND,       // controller send ckpt action to processor
    TTP_MSG_OP_CKPT_REPLY,      // processor reply result of ckpt to controller
    TTP_MSG_OP_CTRL_NOTIFY,     // controller send back up controller to processor
    TTP_MSG_OP_RENAME,          // controller send rename requet to processor
    TTP_MSG_OP_RENAME_REPLY,    // processor reply result of rename to controller
    TTP_MSG_OP_COLLECTION,      // controller send to processor to collect latest status
    TTP_MSG_OP_COLLECTION_REPLY,
    TTP_MSG_OP_PRELOCK,
    TTP_MSG_OP_PRELOCK_REPLY,
    TTP_MSG_OP_PAUSE,
    TTP_MSG_OP_PAUSE_REPLY,
    TTP_MSG_OP_CONTINUE,
    TTP_MSG_OP_CONTINUE_REPLY,
    TTP_MSG_OP_DEVICE_STOP,
    TTP_MSG_OP_STOP_REPLY,
    TTP_MSG_OP_DEVICE_CLEAN,
    TTP_MSG_OP_CLEAN_REPLY,
    TTP_MSG_OP_REPAIR,
    TTP_MSG_OP_REPAIR_REPLY,
    TTP_MSG_OP_ROLLBACK,
    TTP_MSG_OP_ROLLBACK_REPLY,
    TTP_MSG_OP_NOTIFY_NORMAL,
    TTP_MSG_OP_NORMAL_REPLY,
    TTP_MSG_OP_EXIT,
    TTP_MSG_OP_DESTROY_NOTIFY,
    TTP_MSG_OP_DOWNGRADE_REBUILD,
    TTP_MSG_OP_DOWNGRADE_REBUILD_REPLY,
    TTP_MSG_OP_PT_COMM,
    TTP_MSG_OP_PT_COMM_REPLY,
    TTP_MSG_OP_UPGRADE_REBUILD,
    TTP_MSG_OP_UPGRADE_REBUILD_REPLY,
    TTP_MSG_OP_UPGRADE_REPAIR,
    TTP_MSG_OP_UPGRADE_REPAIR_REPLY,
    TTP_MSG_OP_UPGRADE_ROLLBACK,
    TTP_MSG_OP_UPGRADE_ROLLBACK_REPLY,
    TTP_MSG_OP_LAUNCH_STORE_SERVER,
    TTP_MSG_OP_LAUNCH_STORE_SERVER_REPLY,
    TTP_MSG_OP_BUTT,
};

inline const char *StrMsgOpCode(uint16_t type)
{
    static std::vector<std::string_view> msgTypes = {
        "TTP_MSG_OP_REGISTER",
        "TTP_MSG_OP_INIT_REPORT",
        "TTP_MSG_OP_DP_REPORT",
        "TTP_MSG_OP_HEARTBEAT_SEND",
        "TTP_MSG_OP_CKPT_SEND",
        "TTP_MSG_OP_CKPT_REPLY",
        "TTP_MSG_OP_CTRL_NOTIFY",
        "TTP_MSG_OP_RENAME",
        "TTP_MSG_OP_RENAME_REPLY",
        "TTP_MSG_OP_COLLECTION",
        "TTP_MSG_OP_COLLECTION_REPLY",
        "TTP_MSG_OP_PRELOCK",
        "TTP_MSG_OP_PRELOCK_REPLY",
        "TTP_MSG_OP_PAUSE",
        "TTP_MSG_OP_PAUSE_REPLY",
        "TTP_MSG_OP_CONTINUE",
        "TTP_MSG_OP_CONTINUE_REPLY",
        "TTP_MSG_OP_DEVICE_STOP",
        "TTP_MSG_OP_STOP_REPLY",
        "TTP_MSG_OP_DEVICE_CLEAN",
        "TTP_MSG_OP_CLEAN_REPLY",
        "TTP_MSG_OP_REPAIR",
        "TTP_MSG_OP_REPAIR_REPLY",
        "TTP_MSG_OP_ROLLBACK",
        "TTP_MSG_OP_ROLLBACK_REPLY",
        "TTP_MSG_OP_NOTIFY_NORMAL",
        "TTP_MSG_OP_NORMAL_REPLY",
        "TTP_MSG_OP_EXIT",
        "TTP_MSG_OP_DESTROY_NOTIFY",
        "TTP_MSG_OP_DOWNGRADE_REBUILD",
        "TTP_MSG_OP_DOWNGRADE_REBUILD_REPLY",
        "TTP_MSG_OP_PT_COMM",
        "TTP_MSG_OP_PT_COMM_REPLY",
        "TTP_MSG_OP_UPGRADE_REBUILD",
        "TTP_MSG_OP_UPGRADE_REBUILD_REPLY",
        "TTP_MSG_OP_UPGRADE_REPAIR",
        "TTP_MSG_OP_UPGRADE_REPAIR_REPLY",
        "TTP_MSG_OP_UPGRADE_ROLLBACK",
        "TTP_MSG_OP_UPGRADE_ROLLBACK_REPLY",
        "TTP_MSG_OP_LAUNCH_STORE_SERVER",
        "TTP_MSG_OP_LAUNCH_STORE_SERVER_REPLY",
        "TTP_MSG_OP_BUTT",
    };

    if (type >= msgTypes.size()) {
        return "";
    }

    return msgTypes[type].data();
}

enum ControllerRepairType : uint32_t {
    CRT_DUMP = 0,
    CRT_RETRY,
    CRT_ARF,
    CRT_UPGRADE,
    CRT_DOWNGRADE,
    CRT_BUTT
};

enum TrainStateEnum : int8_t {
    TTP_STATUS_NORMAL = 0,
    TTP_STATUS_ABNORMAL,
    TTP_STATUS_UCE_HIGH,
    TTP_STATUS_UCE_LOW,
    TTP_STATUS_UCE_CORRUPTED,
    TTP_STATUS_OFFLINE,
    TTP_STATUS_ISOLATE,
    TTP_STATUS_INIT_FINISH,
    TTP_STATUS_PREREPAIR_FINISH,
    TTP_STATUS_PRECISION_ERROR,
    TTP_STATUS_STEP_FINISH,
    TTP_STATUS_EXIT,
    TTP_STATUS_HCCL_FAILED,
    TTP_STATUS_BUTT,
};

enum class RepairType : int16_t {
    RT_SEND,
    RT_UCE_HIGHLEVEL,
    RT_UCE_LOWLEVEL,
    RT_RECV_REPAIR,
    RT_ROLLBACK,
    RT_LOAD_CKPT,
    RT_LOAD_REBUILD,
    RT_BUTT,
};

enum MaskStatusEnum : uint8_t {
    MASK_NORMAL = 0,
    MASK_ERROR,
    MASK_UCE_HIGH,
    MASK_UCE_LOW,
};

enum FrameworkTypeEnum : uint32_t {
    TYPE_DEFAULT = 0,
    TYPE_X1,
};

} // namespace ttp
} // namespace ock

#endif  // OCK_TTP_CONSTANTS_H
