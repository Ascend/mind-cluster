#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
# normal rules
from ascend_fd.utils.i18n import get_label_for_language

DEVICE_IP_FILE = r"device_ip_info.json"
SERVER_INFO_FILE = "server-info.json"
# Rc job rule
INFER_FILE = "mindie-cluster-info.json"
CONTAINER_FILE = "server-info.json"
TIMEOUT_SOCKET = "Socket"
TIMEOUT_NOTIFY = "Notify"
TIMEOUT_FFTS = "FFTSPlus"
TIMEOUT_ROOT_INFO = "RootInfo"
TIMEOUT_NORMAL = "normal"
DEFAULT_IDENTIFIER = "identifier-NA"
PLOG_ORIGIN = r"plog-(\d{1,10})_\d{1,20}\.log"
PLOG_DEVICE_ORIGIN = r"device-(\d{1,10})_\d{1,20}\.log"
PLOG_PARSED = r"plog-parser-(\d{1,10})-(\d)\.log"
TIME_OUT = r"timeOut\[(\d{1,10})\]"
RANK_NUM_INFO = "rankNum["
TOTAL_RANK_INFO = "totalRanks["
ENTRY_RANKS_INFO = "ranks["
ENTRY_DEVICE_INFO = "deviceLogicId["
IDENTIFIER_INFO = "identifier["
RANK_INFO = "rank["
SERVER_INFO = "server["
SERVER_ID_INFO = "serverId["
OLD_DEVICE_INFO = "device["
LOGIC_DEVICE_INFO = "logicDevId["
PHY_DEVICE_INFO = "phydevId["
DEVICE_IP_INFO = "deviceIp["
TAG_INFO = "tag["
NOTIFY_INDEX_INFO = "index["
NOTIFY_REMOTE_RANK_INFO = "remote rank:["
NOTIFY_IDENTIFIER_INFO = "group:["
# socket timeout info
SOCKET_VIRTUAL_NIC_IP_INFO = "[Vnic]Listen on ip["
SOCKET_PHY_ID_INFO = "devPhyId["
CLUSTER_EXCEPTION_ROOT_DEVICE = "Cluster Exception Location[IP/ID]:["
CLUSTER_EXCEPTION_ROOT_CAUSE = "ExceptionType:["
NOTIFY_TASK_EXCEPTION = "TaskExceptionHandler"
ERROR_ALL = "[ERROR]"
ERROR_HCCL = "[ERROR] HCCL"
ERROR_CQE = "cqe err status[12]"
ERROR_CQE_NEW = "cqe error status[12]"
ERROR_CQE_SPLIT = "ip:["
ERROR_CQE_NEW_SPLIT = "remoteIP:["
CONNECT_TIMEOUT = "CONNECT_TIMEOUT"
EXEC_TIMEOUT = "EXEC_TIMEOUT"
RDMA_TIMEOUT = "RDMA_TIMEOUT"
RDMA_RETRY_CNT = "RDMA_RETRY_CNT"
CONNECT_TIMEOUT_KEYWORD = "HCCL_CONNECT_TIMEOUT is set"
EXEC_TIMEOUT_KEYWORD = "ExecTimeOut is set"
RDMA_TIMEOUT_KEYWORD = "rdmaTimeOut is set"
RDMA_RETRY_CNT_KEYWORD = "rdmaRetryCnt is set"
DEFAULT_CONNECT_TIMEOUT_SET_KEYWORD = "HCCL_CONNECT_TIMEOUT set by"
DEFAULT_EXEC_TIMEOUT_SET_KEYWORD = "HCCL_EXEC_TIMEOUT set by"
DEFAULT_RDMA_TIMEOUT_SET_KEYWORD = "HCCL_RDMA_TIMEOUT set by"
DEFAULT_RDMA_RETRY_CNT_SET_KEYWORD = "HCCL_RDMA_RETRY_CNT set by"
EXTERNAL_INPUT_KEYWORD = "externalinput.cc"
ENTRY_ROOT_INFO = "Entry-HcclCommInitRootInfo"
GET_ROOT_INFO = "HcclGetRootInfo success"
INIT_ROOT_INFO = "HcclCommInitRootInfo"
HCCL_IP_INFO = r"hccn_tool -i (\d{1,3}) -ip -g"
HCCL_IPADDR = r"ipaddr:(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})"
TLS_SWITCH = r"TLS SWITCH \((\d{1,3})\)"
HOST_SN = r'Serial Number:\s{0,10}([A-Z0-9]{15,25})'
BMC_BOARD_SN = 'Board Serial Number\s{0,10}:\s{0,10}([A-Z0-9]{10,15})'
BMC_COMPLETE_MACHINE_SN = 'Product Serial Number\s{0,2}:\s{0,2}([A-Z0-9]{15,25})'
LCNE_BOARD_SN = r"\[GetCPUTablebar_code\]outbuf=([A-Z0-9]{10,15})"
DEVICE_ID = r'device-(\d{1,3})'
DEV_OS_ID = r'dev-os-(\d{1,3})'
ATTR_INIT_SUCCESS = "attr init success"
N_SECOND_RECOVERY_FINISH = "HcclCommResume:success, take time"
LAGGING_INFO_ON_WAITING = "ReportTimeoutProc: report timeout"

# Kg job rule
DEVICE_LOG_ORIGIN = r"device-(\d{1,10})_\d{1,20}\.log"
DEV_OS_INFO = "dev-os-"
DEV_NPU_HISI_HISTORY_ORIGIN = "history.log"
DEV_NPU_INFO = "device-"
DATETIME_REGEX = r"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{6}"
PRE_COMP_OS_FAULT = "Comp_OS_"
PRE_COMP_SWITCH_FAULT = "Comp_Switch_L1"
PRE_AMCT_FAULT = "AISW_CANN_AMCT"

# Node job rule
NPU_SMI_DETAILS_CSV = r"npu_smi_(\d{1,3})_details.csv"
NAD_OUT_FILENAME = "nad_clean.csv"
PROCESS_CORE_CSV = r"process_(\d{1,10}).csv"
PROCESS_FILE = 'process_'
MEM_USED_FILE = 'mem_used.csv'
PROCESS_VALID_COLUMNS = ['time', 'pid', 'rss', 'cpu', "cpu_affinity"]

# Net job rule
NPU_DETAILS_CSV = r"npu_(\d{1,3})_details.csv"
NIC_OUT_FILENAME = "nic_clean.csv"

# Version information label
# VERSION_INFO_LABEL_LIST and SHOW_LABEL_LIST must correspond to each other.
VERSION_INFO_LABEL_LIST = ["driver_version", "firm_version", "nnae_version", "cann_version", "pytorch_version",
                           "torch_npu_version", "mindspore_version"]
SHOW_LABEL_LIST = ["Driver", "Firmware", "NNAE", "Toolkit", "PyTorch", "Torch-npu", "MindSpore"]

# component sort rules
SORT_RULES = {"BMC": 0, "LCNE": 1, "NPU": 2, "Network": 3, "HostOS": 4, "CANN": 5, "AI Framework": 6}
LOWEST_PRIORITY_NUM = 9999
# Error code and component relationship dictionary
ERROR_CODE_COMPONENT = {
    "1": "GE",
    "2": "FE",
    "3": "AI CPU",
    "4": "TEFusion",
    "7": "Vector Operator Plugin",
    "8": "Vector Operator",
    "B": "TBE Pass Compilation Tool (Back-End)",
    "C": "Auto Tune",
    "D": "RLTune",
    "E": "RTS",
    "F": "LxFusion & AutoDeploy",
    "G": "AOE",
    "H": "ACL",
    "I": "HCCL",
    "J": "HCCP",
    "K": "Profiling",
    "L": "Driver",
    "M": "Queue Schedule",
    "N": "DVPP",
    "O": "AMCT",
    "Z": "Public Operator or AclNN",
}

zh_lb = get_label_for_language(specified_language="zh")
en_lb = get_label_for_language(specified_language="en")
TRACEBACK_FAULT_ENTITY_ATTR = {
    "class": "User",
    "component": "AI Framework",
    "module": "Python",
    "cause_zh": zh_lb.traceback_cause,
    "description_zh": zh_lb.traceback_description,
    "suggestion_zh": f"{zh_lb.traceback_former_suggestion}\n{zh_lb.traceback_latter_suggestion}",
    "cause_en": en_lb.traceback_cause,
    "description_en": en_lb.traceback_description,
    "suggestion_en": f"{en_lb.traceback_former_suggestion}\n{en_lb.traceback_latter_suggestion}"
}

# default max time and min time
MAX_TIME = "9999-12-31-23:59:59.999999"
MIN_TIME = "0000-01-01-00:00:00.000000"
KG_MAX_TIME = "9999-12-31 23:59:59.999999"
KG_MIN_TIME = "0000-01-01 00:00:00.000000"

# supported source_file config
CANN_PLOG_SOURCE = "CANN_Plog"
CANN_DEVICE_SOURCE = "CANN_Device"
TRAIN_LOG_SOURCE = "TrainLog"
NPU_OS_SOURCE = "NPU_OS"
NPU_DEVICE_SOURCE = "NPU_Device"
NPU_HISTORY_SOURCE = "NPU_History"
OS_SOURCE = "OS"
OS_DEMESG_SOURCE = "OS-dmesg"
OS_VMCORE_DMESG_SOURCE = "OS-vmcore-dmesg"
OS_SYSMON_SOURCE = "OS-sysmon"
NODEDLOG_SOURCE = "NodeDLog"
DEVICEPLUGIN_SOURCE = "DL_DevicePlugin"
VOLCANO_SCHEDULER_SOURCE = "DL_Volcano_Scheduler"
VOLCANO_CONTROLLER_SOURCE = "DL_Volcano_Controller"
DOCKER_RUNTIME_SOURCE = "DL_Docker_Runtime"
NPU_EXPORTER_SOURCE = "DL_Npu_Exporter"
MINDIO_SOURCE = "MindIO"
MINDIE_SOURCE = "MindIE"
AMCT_SOURCE = "CANN_Amct"
BMC_SOURCE = "BMCLog"
BMC_APP_DUMP_SOURCE = "BMCAppDumpLog"
BMC_DEVICE_DUMP_SOURCE = "BMCDeviceDumpLog"
BMC_LOG_DUMP_SOURCE = "BMCLogDumpLog"
LCNE_SOURCE = "LCNELog"

# actually not supported, but used in certain cases
NPU_INFO_SOURCE = "NPU_INFO"
DMI_DECODE_SOURCE = "DMI_DECODE"
MINDIE_CLUSTER_SOURCE = "MINDIE_CLUSTER"
CUSTOM_LOG_SOURCE = "CustomLog"

# composite switch chip info source for devicePlugin and LCNE
COMPOSITE_SWITCH_CHIP_SOURCE = "DL_DevicePlugin | LCNELog"

# supported source file list
SUPPORTED_SOURCE_FILE_LIST = [CANN_PLOG_SOURCE, CANN_DEVICE_SOURCE, TRAIN_LOG_SOURCE, NPU_OS_SOURCE, NPU_DEVICE_SOURCE,
                              NPU_HISTORY_SOURCE, OS_SOURCE, OS_DEMESG_SOURCE, OS_VMCORE_DMESG_SOURCE, OS_SYSMON_SOURCE,
                              NODEDLOG_SOURCE, DEVICEPLUGIN_SOURCE, VOLCANO_SCHEDULER_SOURCE, VOLCANO_CONTROLLER_SOURCE,
                              DOCKER_RUNTIME_SOURCE, NPU_EXPORTER_SOURCE, MINDIE_SOURCE, AMCT_SOURCE, BMC_SOURCE,
                              BMC_APP_DUMP_SOURCE, BMC_DEVICE_DUMP_SOURCE, BMC_LOG_DUMP_SOURCE, LCNE_SOURCE]

# saver to source file map
SAVER_TO_SOURCE_FILE_MAP = {
        "ProcessLogSaver": [CANN_PLOG_SOURCE, CANN_DEVICE_SOURCE],
        "EnvInfoSaver": [NPU_INFO_SOURCE],
        "TrainLogSaver": [TRAIN_LOG_SOURCE],
        "DevLogSaver": [NPU_OS_SOURCE, NPU_DEVICE_SOURCE, NPU_HISTORY_SOURCE],
        "HostLogSaver": [OS_SOURCE, OS_DEMESG_SOURCE, OS_VMCORE_DMESG_SOURCE,
                         OS_SYSMON_SOURCE],
        "DlLogSaver": [NODEDLOG_SOURCE, DEVICEPLUGIN_SOURCE, VOLCANO_SCHEDULER_SOURCE, VOLCANO_CONTROLLER_SOURCE,
                       DOCKER_RUNTIME_SOURCE, NPU_EXPORTER_SOURCE, MINDIO_SOURCE],
        "MindieLogSaver": [MINDIE_SOURCE, MINDIE_CLUSTER_SOURCE],
        "BMCLogSaver": [BMC_SOURCE, BMC_APP_DUMP_SOURCE, BMC_DEVICE_DUMP_SOURCE, BMC_LOG_DUMP_SOURCE],
        "LCNELogSaver": [LCNE_SOURCE],
        "AMCTLogSaver": [AMCT_SOURCE]
    }

OS_FAULT_PREFIX = "Comp_OS"
MINDIE_FAULT_PREFIX = "AISW_MindIE"
