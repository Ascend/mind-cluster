from enum import Enum


class BmcDumpLogDataType(Enum):
    BMC_IP = "bmc_ip"
    SN_NUM = "sn_num"
    SEL_INFO = "sel_info"
    SENSOR_INFO = "sensor_info"
    HEALTH_EVENTS = "health_events"
    OP_HISTORY_INFO_LOG = "optical_history_info_log"
