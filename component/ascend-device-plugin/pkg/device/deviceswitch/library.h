#ifndef LINGQU_LIBRARY_H
#define LINGQU_LIBRARY_H

#define MAX_EVENT_RESV_LENGTH 32
#define DCMIDLLEXPORT static

#define LQ_DCMI_EVENT_FILTER_FLAG_EVENT_TYPE_ID (1UL << 0)
#define LQ_DCMI_EVENT_FILTER_FLAG_EVENT_ID (1UL << 1)
#define LQ_DCMI_EVENT_FILTER_FLAG_SERVERITY (1UL << 2)
#define LQ_DCMI_EVENT_FILTER_FLAG_CHIP_ID (1UL << 3)

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

typedef struct LqDcmiEvent{
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
    EventTypeId = 1UL << 0,
    EventId = 1UL << 1,
    Severity = 1 << 2,
    ChipId = 1 << 3
} LqDcmiEventFilterFlag;

typedef struct lq_dcmi_event_filter {
    LqDcmiEventFilterFlag filterFlag;
    HalReportFaultType eventTypeId;
    unsigned int eventId;
    unsigned char severity;
    unsigned int chipId;
} LqDcmiEventFilter;


typedef void (*lq_dcmi_fault_event_callback)(struct LqDcmiEvent *event);

DCMIDLLEXPORT int lq_dcmi_init();
DCMIDLLEXPORT int lq_dcmi_subscribe_fault_event(struct lq_dcmi_event_filter filter);
DCMIDLLEXPORT int lq_dcmi_get_fault_info(unsigned int list_len, unsigned int *event_list_len, struct LqDcmiEvent *event_list);

#endif// LINGQU_LIBRARY_H