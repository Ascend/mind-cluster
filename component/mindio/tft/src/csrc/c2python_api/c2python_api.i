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
%module(docstring="TTP c to python api.", directors=1, threads=1, naturalvar=1) ttp_c2python_api

%{
#include "c2python_api.cpp"
#include "../common/common.h"
#include "../framework/processor/processor.h"

using namespace ock::ttp;
%}

%include stl.i
%include stdint.i
%include std_map.i
%include std_string.i

%feature("director") CallbackReceiver;

namespace std {
%template() std::vector<int32_t>;
%template() std::vector<std::string>;
%template() std::vector<std::vector<int32_t>>;
%template() std::map<int32_t, int32_t>;
%template() std::map<std::string, bool>;
}

namespace ock {
namespace ttp {

enum class ReportState : int32_t {
    RS_NORMAL,
    RS_UCE,
    RS_UCE_CORRUPTED,
    RS_HCCL_FAILED,
    RS_INIT_FINISH,
    RS_PREREPAIR_FINISH,
    RS_STEP_FINISH,
    RS_UNKNOWN,
};

enum class RepairType : int16_t {
    RT_SEND,
    RT_UCE_HIGHLEVEL,
    RT_UCE_LOWLEVEL,
    RT_RECV_REPAIR,
    RT_ROLLBACK,
    RT_LOAD_CKPT,
    RT_LOAD_REBUILD,
};

struct RepairContext {
    RepairType type;
    std::vector<int32_t> srcRank;
    std::vector<int32_t> dstRank;
    std::vector<int32_t> replicaIdx;
    std::vector<int32_t> groupIdx;
    int32_t repairId;
    int64_t step;
    std::vector<int32_t> ranks;
    std::string zitParam;
};

}  // namespace ttp
}  // namespace ock

%exception {
    try {
        $action
    } catch (const std::runtime_error& e) {
        SWIG_exception(SWIG_RuntimeError, e.what());
    }
}

%rename("%(undercase)s") "";
%rename(start_copying) SetCopying;
%rename(start_updating) SetUpdating;
%rename(end_updating) SetFinished;
%rename(set_dump_status) SetDumpResult;
%rename(set_stop_device_callback) SetDeviceStopCallback;
%rename(set_clean_device_callback) SetDeviceCleanCallback;
%rename(set_register_check_callback) SetRegisterCallback;
%rename(mindx_notify_fault_callback) MindxNotifyFaultRanksCallback;
%rename(mindx_query_high_availability_switch) QueryHighAvailabilitySwitch;

%include "c2python_api.cpp"