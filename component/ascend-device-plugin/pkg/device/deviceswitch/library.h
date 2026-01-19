/* Copyright(C) 2024. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

#ifndef LINGQU_DCMI_H
#define LINGQU_DCMI_H

#define DCMIDLLEXPORT static

typedef enum {
    HAL_REPORT_FAULT_BLOCK = 0,
    HAL_REPORT_FAULT_MEMORY,
    HAL_REPORT_FAULT_DISCARD,
    HAL_REPORT_FAULT_MEMORY_ALARM,
    HAL_REPORT_FAULT_MODULE_RESET,
    HAL_REPORT_HEART_LASTWORD,
    HAL_REPORT_FAULT_HEART,
    HAL_REPORT_PORT_FAULT_INVALID_PKG,
    HAL_REPORT_PORT_FAULT_UNSTABLE,
    HAL_REPORT_PORT_FAULT_FAIL,
    HAL_REPORT_FAULT_BY_DEVICE,
    HAL_REPORT_FAULT_CONFIG,
    HAL_REPORT_FAULT_MEM_SINGLE,
    HAL_REPORT_FAULT_M7,
    HAL_REPORT_FAULT_BLOCK_C,
    HAL_REPORT_FAULT_MEM_MULTI,
    HAL_REPORT_FAULT_PCIE,
    HAL_REPORT_FAULT_FATAL,
    HAL_REPORT_PORT_FAULT_TIMEOUT_RP,
    HAL_REPORT_PORT_FAULT_TIMEOUT_LP,
    HAL_REPORT_FAULT_MAX
} HalReportFaultType;

typedef struct LqDcmiEvent {
    HalReportFaultType eventType;
    unsigned int subType;
    unsigned short peerportDevice;
    unsigned short peerportId;
    unsigned short switchChipid;
    unsigned short switchPortid;

    unsigned char severity;
    unsigned char assertion;
    char res[6];
    unsigned int eventSerialNum;
    unsigned int notifySerialNum;
    unsigned long alarmRaisedTime;

    unsigned char additionalParam
    [40];
    char additionalInfo[32];
}LqDcmiEvent;

typedef enum {
    EVENT_TYPE_ID = 1UL << 0,
    EVENT_ID = 1UL << 1,
    SEVERITY = 1UL << 2,
    CHIP_ID = 1UL << 3,
} LqDcmiEventFilterFlag;

typedef struct lq_dcmi_event_filter {
    LqDcmiEventFilterFlag filterFlag;
    HalReportFaultType eventTypeId;
    unsigned int eventId;
    unsigned char severity;
    unsigned int chipId;
} LqDcmiEventFilter;


typedef void (*LqDcmiFaultEventCallback)(struct LqDcmiEvent *event);

DCMIDLLEXPORT int lq_dcmi_init();
DCMIDLLEXPORT int lq_dcmi_subscribe_fault_event(struct lq_dcmi_event_filter filter);
DCMIDLLEXPORT int lq_dcmi_get_fault_info(unsigned int listLen, unsigned int *eventListLen,
    struct LqDcmiEvent *eventList);

#endif // LINGQU_DCMI_H