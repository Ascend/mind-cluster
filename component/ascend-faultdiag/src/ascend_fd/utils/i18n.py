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
from dataclasses import dataclass

from ascend_fd.configuration.config import LANGUAGE


@dataclass
class Label:
    version_info: str
    label_type: str
    version: str
    component: str
    root_cluster_analysis: str
    knowledge_graph_analysis: str
    node_anomaly_analysis: str
    net_congestion_analysis: str
    analysis_failed: str
    please_check_the_log: str
    note: str
    root_cause_device: str
    case_description: str
    root_cause_device_chain: str
    remote_link_chain: str
    fault_source: str
    cause: str
    description: str
    fault_description: str
    error_case: str
    fixed_case: str
    suggestion: str
    fault_propagation_chain: str
    plog_log: str
    log_specification: str
    first_error_device: str
    last_error_device: str
    suspected_root_cause_fault: str
    status_code: str
    error_code: str
    result_description: str
    key_info: str
    faulty_workers: str
    faulty_device: str
    fault_process: str
    fault_occurrence_period: str
    fault_probability: str
    complete_fault_details_refer_to_json: str
    module: str
    fault_category: str
    refer_plog_of_corresponding_pid: str
    plog_of_the_first_error_device_shown_below: str
    unrecorded_aicore_fault: str
    suspected_op_script_error: str
    instance_name: str
    node_name: str
    partial: str
    all_: str
    traceback_cause: str
    traceback_description: str
    traceback_former_suggestion: str
    traceback_latter_suggestion: str
    invalid_field_in_log_format: str
    left_bracket: str
    right_bracket: str
    pytorch_custom_event_default_description: str
    mindie_custom_event_default_description: str
    cann_custom_event_default_description: str


label_en = Label(
    version_info="Version Info",
    label_type="Type",
    version="Version",
    root_cluster_analysis="Root Cause Node Analysis",
    knowledge_graph_analysis="Fault Event Analysis",
    node_anomaly_analysis="Device Resource Analysis",
    net_congestion_analysis="Network Congestion Analysis",
    analysis_failed="Analysis Failed",
    please_check_the_log="Please check the log for error messages",
    note="Note",
    root_cause_device="Root Cause Node(s)",
    case_description="Symptom",
    root_cause_device_chain="Root Cause Node Chain",
    remote_link_chain="Inter-Device Waiting Chain",
    fault_source="Faulty Device(s)",
    cause="Fault Name",
    fault_description="Fault Description",
    error_case="Incorrect Example",
    fixed_case="Correct Example",
    suggestion="Handling Suggestion",
    fault_propagation_chain="Fault Propagation Chain",
    description="Description",
    plog_log="PLOG Log",
    log_specification="Log Specification",
    first_error_device="First Error Device",
    last_error_device="Last Error Device",
    suspected_root_cause_fault="Suspected Root Cause Fault",
    status_code="Status Code",
    error_code="Error Code",
    result_description="Symptom Description",
    key_info="Key Info",
    faulty_workers="Faulty Workers",
    faulty_device="Faulty Device",
    fault_process="Fault Process",
    fault_occurrence_period="Fault Occurrence Period",
    fault_probability="Fault Probability",
    component="Component",
    complete_fault_details_refer_to_json="Please refer to the JSON report for whole fault occurrence period records",
    module="Module",
    fault_category="Fault Category",
    refer_plog_of_corresponding_pid="Please check the original PLOG log of the corresponding PID on "
                                    "{device}: plog-{plog_info}_xxx.log",
    plog_of_the_first_error_device_shown_below="The process log of the first error device {device} is shown below:\n",
    instance_name="Instance Name",
    node_name="Node Name",
    unrecorded_aicore_fault="Data input mismatches, out-of-bounds access, calculation overflows, or other exceptions.",
    suspected_op_script_error="It may be an issue with the code of the operator",
    partial="Some",
    all_="All",
    traceback_cause="Python Execution Error",
    traceback_description="Python execution error with traceback information printed.",
    traceback_former_suggestion="1. Please troubleshoot the error based on the Traceback information.",
    traceback_latter_suggestion="2. If there are multiple faulty devices, "
                                "the Traceback information for each error may vary. "
                                "Please check the specific Traceback information of each device in the JSON report.",
    invalid_field_in_log_format="The log format references a field that has not been provided",
    left_bracket="(",
    right_bracket=")",
    pytorch_custom_event_default_description="Error code reported by PyTorch",
    cann_custom_event_default_description="Error code reported by CANN",
    mindie_custom_event_default_description="Error code reported by MindIE"
)

label_zh = Label(
    version_info="版本信息",
    label_type="类型",
    version="版本",
    module="模块",
    component="组件",
    note="说明",
    analysis_failed="分析失败",
    instance_name="实例名",
    node_name="节点名",
    root_cluster_analysis="根因节点分析",
    knowledge_graph_analysis="故障事件分析",
    node_anomaly_analysis="设备资源分析",
    net_congestion_analysis="网络拥塞分析",
    root_cause_device="根因节点",
    case_description="现象描述",
    root_cause_device_chain="根因节点链",
    remote_link_chain="卡间等待链",
    fault_source="故障设备",
    cause="故障名称",
    fault_description="故障描述",
    error_case="错误示例",
    fixed_case="正确示例",
    suggestion="建议方案",
    fault_propagation_chain="关键传播链",
    description="描述",
    plog_log="PLOG日志",
    log_specification="日志说明",
    first_error_device="首错节点",
    last_error_device="尾错节点",
    suspected_root_cause_fault="疑似根因故障",
    status_code="状态码",
    error_code="错误码",
    result_description="结果描述",
    key_info="关键日志",
    faulty_workers="故障设备",
    faulty_device="故障节点",
    fault_process="故障进程",
    fault_probability="故障概率",
    fault_occurrence_period="故障区间",
    complete_fault_details_refer_to_json="全部故障区间请查阅Json报告获取",
    fault_category="故障分类",
    please_check_the_log="请查看日志报错信息",
    refer_plog_of_corresponding_pid="该节点的原始Plog日志请查看{device}上对应pid的Plog日志：plog-{plog_info}_xxx.log",
    plog_of_the_first_error_device_shown_below="首错节点{device}的Plog日志如下：\n",
    unrecorded_aicore_fault="数据输入不匹配、访问越界、计算溢出等异常。",
    suspected_op_script_error="可能算子本身代码问题",
    partial="部分",
    all_="全部",
    traceback_cause="Python执行报错",
    traceback_description="Python执行报错，并打印了相关Traceback信息。",
    traceback_former_suggestion="1. 请根据Python Traceback信息排查错误；",
    traceback_latter_suggestion="2. 若存在多个故障设备，每个报错的Traceback信息可能不一致，请在json报告中查看每个设备的具体Traceback信息；",
    invalid_field_in_log_format="日志格式中引用了未提供的字段",
    left_bracket="（",
    right_bracket="）",
    pytorch_custom_event_default_description="PyTorch上报故障码",
    mindie_custom_event_default_description="MindIE上报错误码",
    cann_custom_event_default_description="CANN上报ERROR CODE"
)

fault_descriptions_zh = {
    101: "所有节点的Plog都没有记录超时类错误日志。日志中有报错的节点为疑似根因节点，请排查。",
    102: "所有有效节点的Plog都没有错误日志信息，无法定位根因节点。同时请确认是否为正常的任务？",
    107: "通信域内所有节点Plog报错算子下发建链超时，且最早报错节点与最晚报错节点间报错间隔超过设置的超时时间（{}s）。"
         "1、请优先调高超时阈值（HCCL_CONNECT_TIMEOUT）；2、若阈值合理，最晚报错节点疑似为根因节点，请排查。",
    108: "训练/推理任务所使用的{}节点Plog报错算子下发建链超时，且最早报错节点与最晚报错节点间报错间隔未超过设置的超时时间（{}s）。"
         "请优先排查相互等待或者等待关系末端的设备。",
    109: "训练/推理任务所使用的节点报错{} Notify Wait 超时。请优先排查相互等待或者等待关系末端的设备。",
    110: "通信域内所有节点Plog报错Notify Wait超时（FFTS+ run failed），且最早报错节点与最晚报错节点间报错间隔超过设置的超时时间（{}s）。"
         "1、请优先调高超时阈值（HCCL_EXEC_TIMEOUT）；2、若阈值合理，最晚报错节点疑似为根因节点，请排查。",
    111: "通信域内所有节点Plog报错Notify Wait超时（FFTS+ run failed），且最早报错节点与最晚报错节点间报错间隔未超过设置的超时时间（{}s）。无法定位根因节点，请排查全部节点。",
    112: "通信域内部分节点Plog报错Notify Wait超时（FFTS+ run failed），其余节点未报此错误。未报此错误的节点为疑似根因节点，请排查。",
    113: "最早报错节点所报错误非超时类错误，所有未报超时类错误的报错节点为疑似根因节点，请排查。",
    114: "此任务可能是单卡任务，或者是在HCCL初始化前出现错误。无法获取有效的卡信息。请直接排查任务所使用的对应卡。",
    115: "未查找到有效的Plog文件，无法定位根因节点。请确认是否存在Plog文件？",
    116: "部分节点发生RoCE重传超次(ERROR CQE)，此类节点为疑似根因节点，请排查。",
    117: "通信域内所有节点Plog报错root info方式初始化超时，且最早报错节点与最晚报错节点间报错间隔超过设置的超时时间（{}s）。"
         "1、请优先调高超时阈值（HCCL_CONNECT_TIMEOUT）；2、若阈值合理，最晚报错节点疑似为根因节点，请排查。",
    118: "通信域内所有节点Plog报错root info方式初始化超时，且最早报错节点与最晚报错节点间报错间隔未超过设置的超时时间（{}s）。"
         "初步怀疑报错通信域的root节点为根因节点。请按以下步骤排查：1、首先排查集群host网络是否互通；"
         "2、根因节点的所在服务器未能及时处理全量连接请求；"
         "3、根因节点的所在服务器系统配置的open file max num或tcp的backlog太小，二者都推荐设置为65535。",
    119: "通信域内部分节点Plog报错root info方式初始化超时，其余节点未报此错误。未报此错误的节点为疑似根因节点，请排查。",
    120: "检测到节点存在TLS SWITCH状态不一致现象，上述根因节点TLS SWITCH状态为{}，其他节点的状态为{}。少量状态不一致的节点为疑似根因节点，请排查。",
    121: "检测到部分节点初始化时与Root Rank连接失败，且这些节点无对应的Plog日志记录，怀疑这些节点初始化时异常退出。"
         "此类故障无法定位到具体的server与device，请联系工程师，基于报错通信域的rank_id进行排查。",
    122: "通信域内所有节点Plog报错Notify Wait超时，且最早报错节点与最晚报错节点间报错间隔未超过设置的超时时间（{}s）。"
         "当前通信域报错算子index包括{}, 最小index[{}]对应的节点为疑似根因节点，请排查。",
    123: "通信域内所有节点Plog报错Notify Wait超时，且最早报错节点与最晚报错节点间报错间隔未超过设置的超时时间（{}s）。"
         "当前通信域报错算子tag包括{}, 少量tag[{}]对应的节点为疑似根因节点，请排查。",
    124: "通信域内所有节点Plog报错Notify Wait超时（FFTS+ run failed），存在节点间相互死锁等待。疑似通信算子编排问题，无法定位根因节点。",
    125: "通信域内所有节点Plog报错Notify Wait超时（FFTS+ run failed），存在部分节点通信主流等待自身从流超时，其他节点等待该节点超时。"
         "该类节点疑似为根因节点，请排查。",
    126: "HCCL异常扩散机制检测出本次任务的疑似根因节点，对应异常原因为：'{}'，请排查。",
    127: "Plog未查找到有效的日志信息，无法定位根因节点。请保证Plog文件中包含集群Rank信息、Error日志等有效内容。",
    128: "此任务断点续训复训失败，部分节点无法获取有效卡信息，疑似在HCCL初始化前出现错误。",
    129: "此推理实例发生MindIE pull kv失败，优先排查发生pull kv失败的节点。",
    130: "所有Plog中无超时信息，疑似存在进程异常退出或卡死的节点。",
    131: "此推理实例发生MindIE建链失败，请排查发生建链失败的节点。",
    132: "任务所使用的节点报错Transport init error，请优先排查等待关系末端的设备或者相互等待的设备。",
    133: "任务所使用的节点报错Transport init error，未解析到卡间等待关系信息，请优先排查未打印Transport init error日志的设备。"
}

fault_descriptions_en = {
    101: "The Plogs of all nodes do not record timeout error logs. "
         "The node that reports an error in the log is a possible root cause node. Check the node.",
    102: "The Plogs of all valid nodes do not contain error log information. "
         "As a result, the root cause node cannot be located. Check whether the task is normal.",
    107: "Plogs of all nodes in the communication domain report socket link setup timeout errors, "
         "and the interval between the earliest and "
         "latest nodes that report the error exceeds the configured timeout interval ({}s). "
         "1. Increase the timeout threshold (HCCL_CONNECT_TIMEOUT). "
         "2. If the threshold is proper, the node that reports the latest error may be the root node. "
         "In this case, check the node.",
    108: "{} plogs of the training/inference task report errors in establishing connections. "
         "The interval between the earliest and latest nodes that reports the error did not exceed the configured "
         "timeout interval ({} seconds). Please prioritize troubleshooting devices that are in "
         "mutual waiting or at the end of a waiting relationship.",
    109: "the training/inference task report errors in {} Notify Wait timeout. "
         "Please prioritize troubleshooting devices that are in mutual waiting or the end of a waiting relationship.",
    110: "Plogs of all nodes in the communication domain report the Notify Wait timeout (FFTS+ run failed) error, "
         "and the interval between the earliest and latest nodes that report the error exceeds "
         "the configured timeout interval ({}s). "
         "1. Increase the timeout threshold (HCCL_EXEC_TIMEOUT). "
         "2. If the threshold is proper, the node that reports the latest error may be the root node. "
         "In this case, check the node.",
    111: "Plogs of all nodes in the communication domain report the Notify Wait timeout (FFTS+ run failed) error, "
         "and the interval between the earliest and latest nodes that report the error does not exceed "
         "the configured timeout interval ({}s). The root node cannot be located. Check all nodes.",
    112: "Plogs of some nodes in the communication domain report the Notify Wait timeout (FFTS+ run failed) error. "
         "Other nodes do not report this error. The nodes that do not report this error may be root cause nodes. "
         "Check them.",
    113: "The error reported earliest by the node is not a timeout error. "
         "All nodes that do not report timeout errors are suspected root cause nodes.",
    114: "The task may be a single-card task or an error occurs before HCCL initialization. "
         "The valid card information cannot be obtained. Check the card used by the job.",
    115: "The valid Plog file is not found, and the root cause node cannot be located. "
         "Check whether the Plog file exists.",
    116: "Excessive RoCE retransmissions (ERROR CQE) occur on some nodes. "
         "These nodes may be the root cause nodes. Check them.",
    117: "Plogs of all nodes in the communication domain report an error indicating "
         "that the initialization in root info mode times out, and the interval between the earliest and "
         "latest nodes that report the error exceeds the configured timeout interval ({}s). "
         "1. Increase the timeout threshold (HCCL_CONNECT_TIMEOUT). "
         "2. If the threshold is proper, the node that reports the latest error may be the root node. "
         "In this case, check the node.",
    118: "Plogs of all nodes in the communication domain report an error indicating that "
         "the initialization in root info mode times out, and the interval between the earliest and "
         "latest nodes that report the error does not exceed the configured timeout interval ({}s). "
         "It is preliminarily suspected that the root node in the communication domain that reports "
         "the error is the root cause node. "
         "1. Check whether the host network of the cluster is connected. "
         "2. Check that the server where the root cause node is located does not process all connection requests "
         "in a timely manner. "
         "3. Check that the value of open file max num or tcp backlog configured on the server "
         "where the root cause node is located is too small. The recommended value is 65535.",
    119: "Plogs of some nodes in the communication domain report an error indicating "
         "that the initialization in root info mode times out. "
         "Other nodes do not report this error. "
         "The nodes that do not report this error may be root cause nodes. Check them.",
    120: "The TLS switch status of the preceding root node is {}, and the TLS switch status of other nodes is {}. "
         "A few nodes with inconsistent status are suspected root cause nodes. Check them.",
    121: "Some nodes fail to connect to the root rank during initialization, "
         "and these nodes do not have corresponding Plogs. It is suspected that these nodes exit abnormally "
         "during initialization. For such faults, the server and device cannot be located. "
         "In this case, contact technical support to locate the fault based on the rank ID of the communication domain.",
    122: "Plogs of all nodes in the communication domain report Notify Wait timeout errors, "
         "and the interval between the earliest and latest nodes that report the error does not exceed "
         "the configured timeout interval ({}s). The indexes of the error operators in the current communicator "
         "include {}. The node corresponding to the minimum index [{}] may be the root cause node. Check the node.",
    123: "Plogs of all nodes in the communication domain report Notify Wait timeout errors, "
         "and the interval between the earliest and latest nodes that report the error does not exceed the "
         "configured timeout interval ({}s). The tags of the error operators in the current communicator include {}. "
         "The node corresponding to some tags [{}] may be the root cause nodes. Check them.",
    124: "Plogs of all nodes in the communication domain report the Notify Wait timeout (FFTS+ run failed) error. "
         "Deadlocks occur between nodes. It is suspected that the communication operator orchestration is incorrect "
         "and the root cause node cannot be located.",
    125: "Plogs of all nodes in the communication domain report the Notify Wait timeout error (FFTS+ run failed). "
         "The primary stream of some nodes times out when waiting for its secondary stream, "
         "and other nodes also time out when waiting for these nodes. "
         "This type of nodes may be the root cause nodes. Check them.",
    126: "The HCCL exception diffusion mechanism detects the possible root cause node of the task. "
         "The corresponding exception cause is '{}'. Please check.",
    127: "No valid log information is found in the Plog. The root cause node cannot be located. "
         "Ensure that the Plog file contains valid information such as cluster rank information and error logs.",
    128: "Resumable training fails. Effective card information cannot be obtained for some nodes. "
         "It is suspected that an error occurs before HCCL initialization.",
    129: "The inference instance encountered a failure in MindIE pulling kv process. "
         "Prioritize troubleshooting the node where the pull kv operation failed.",
    130: "The Plogs of all nodes do not record timeout error logs, "
         "indicating potential abnormal process termination or lagging node(s).",
    131: "Linking failed for this inference instance. "
         "Please troubleshoot the node(s) where the linking failure occurred.",
    132: "The node(s) report transport init error, "
         "Please check the device at the end of waiting relationship or the devices that are waiting for each other.",
    133: "The node(s) report transport init error, but no waiting relationship information was resolved. "
         "Please first check the devices that did not print the Transport init error log."
}

fault_entities_zh = {
    "NORMAL_OR_UNSUPPORTED": {
        "cause": "故障事件分析模块无结果",
        "description": "故障事件分析模块无结果，可能为正常训练或推理作业，无故障发生。如果训练或推理任务异常中断，存在问题无法解决，请联系工程师处理。",
        "suggestion": ["1. 若存在问题无法解决，请联系工程师定位排查"]
    },
    "NODE_RES_NORMAL": {
        "cause": "设备资源诊断无异常",
        "description": "未发生NPU过载降频和资源抢占。"
    },
    "NODE_RES_ABNORMAL_01": {
        "cause": "NPU发生过载降频",
        "description": "NPU状态异常，发生降频。",
        "suggestion": ["1. 请检查故障详细信息中的NPU节点状态；"]
    },
    "NODE_RES_ABNORMAL_02": {
        "cause": "NPU发生过载降频",
        "description": "NPU温度过高，发生降频。",
        "suggestion": ["1. 请检查故障详细信息中的NPU节点的风扇状况或其他温度问题；"]
    },
    "NODE_RES_ABNORMAL_03": {
        "cause": "CPU抢占（全进程抢占）",
        "description": "设备资源产生异常，所有训练进程发生CPU资源抢占。",
        "suggestion": ["1. 请查看设备是否有非法进程抢占CPU资源；"]
    },
    "NODE_RES_ABNORMAL_04": {
        "cause": "CPU抢占（单进程抢占）",
        "description": "设备资源产生异常，单个训练进程发生CPU资源抢占。",
        "suggestion": ["1. 请查看设备是否有非法进程抢占CPU资源；"]
    },
    "NODE_RES_ABNORMAL_05": {
        "cause": "CPU抢占（部分进程抢占）",
        "description": "设备资源产生异常，部分训练进程发生CPU资源抢占。",
        "suggestion": ["1. 请查看设备是否有非法进程抢占CPU资源；"]
    },
    "NET_CONGESTION_NORMAL": {
        "cause": "链路无拥塞异常",
        "description": "通信链路无拥塞现象。"
    },
    "NET_CONGESTION_ABNORMAL_01": {
        "cause": "链路拥塞异常",
        "description": "部分通信链路发生冲突拥塞。",
        "suggestion": ["1. 建议检查交换机路由策略；"]
    }
}

fault_entities_en = {
    "NORMAL_OR_UNSUPPORTED": {
        "cause": "Fault event analysis yielded no results.",
        "description": "Fault event analysis yielded no results, "
                       "which may indicate normal training/inference task with no faults occurring. "
                       "If the training or inference tasks are abnormally interrupted "
                       "and the issues cannot be resolved, please contact technical support for assistance.",
        "suggestion": ["1. If there are issues that cannot be resolved, "
                       "please contact the technical support for troubleshooting."]
    },
    "NODE_RES_NORMAL": {
        "cause": "Normal device resource diagnosis",
        "description": "Frequency reduction and resource preemption do not occur due to NPU overload."
    },
    "NODE_RES_ABNORMAL_01": {
        "cause": "NPU frequency reduction due to overload",
        "description": "The NPU status is abnormal, and frequency reduction occurs.",
        "suggestion": ["1. Please check the NPU node status in the fault details."]
    },
    "NODE_RES_ABNORMAL_02": {
        "cause": "NPU frequency reduction due to overload",
        "description": "The NPU temperature is too high, and frequency reduction occurs.",
        "suggestion": ["1. Please check the fan status of the NPU node in the fault details or "
                       "other temperature-related issues."]
    },
    "NODE_RES_ABNORMAL_03": {
        "cause": "CPU preemption (full-process preemption)",
        "description": "Device resources are abnormal, and CPU resource preemption occurs on all training processes.",
        "suggestion": ["1. Please check if there are any unauthorized processes seizing CPU resources on the device."]
    },
    "NODE_RES_ABNORMAL_04": {
        "cause": "CPU preemption (single-process preemption)",
        "description": "Device resources are abnormal, and CPU resource preemption occurs on a single training process.",
        "suggestion": ["1. Please check if there are any unauthorized processes seizing CPU resources on the device."]
    },
    "NODE_RES_ABNORMAL_05": {
        "cause": "CPU preemption (preemption by some processes)",
        "description": "Device resources are abnormal, and CPU resource preemption occurs on some training processes.",
        "suggestion": ["1. Please check if there are any unauthorized processes seizing CPU resources on the device."]
    },
    "NET_CONGESTION_NORMAL": {
        "cause": "No link congestion exception",
        "description": "The communication link is not congested."
    },
    "NET_CONGESTION_ABNORMAL_01": {
        "cause": "Link congestion exception",
        "description": "Some communication links are congested due to conflicts.",
        "suggestion": ["1. It is recommended to check the switch routing policies."]
    }
}

note_msg_zh = {
    "MULTI_RANK_NOTE_MSG": "根因节点分析检测出了多个的疑似故障根因节点，将优先排查这几个节点",
    "MAX_RANK_NOTE_MSG": "注：根因节点过多，仅展示16条，已按报错时间排序。所有根因节点可在diag_report.json中查询。",
    "MAX_DEVICE_NOTE_MSG": "注：部分故障的故障设备过多，仅展示16条。所有故障设备可在diag_report.json中查询。",
    "MAX_WORKER_CHAINS_NOTE_MSG": "注：拥有传播链的故障设备过多，仅展示前16个设备。完整关键传播链的故障设备信息可在diag_report.json中查询。",
    "NET_SINGLE_WORKER_MSG": "单机环境不存在网络拥塞情况",
    "UNKNOWN_ROOT_ERROR_RANK": "未诊断出根因节点，故障事件分析将尝试检测全部设备",
    "SOME_SUBTASKS_FAILED": "本分析模块下部分分析子项执行失败，诊断结果可能会受到影响从而不准确。失败信息可在diag_report.json中查询",
    "SOME_DEVICE_FAILED": "重传超次故障链中对端节点无法确认具体的Worker ID和Device ID，请通过IP检查对应Device设备",
    "FAULT_CHAINS_NOTE": "关键传播链只展示每个故障设备最长的一条链路",
    "FAULT_CHAINS_MAX_NOTE": "部分故障设备拥有关键传播链的最大链路组数超过{}，请在json格式诊断结果中查询完整关键传播链信息",
    "NO_GROUP_RANK_INFO_NOTE": "未解析到Notify超时报错对应的通信域信息，根因节点结果可能不准确",
    "REMOTE_LINKS_NOTE": "部分卡间存在等待关系，在“卡间等待链”中展示一条示例",
    "REMOTE_LINKS_MAX_NOTE": "部分卡间等待关系链过长，仅展示16条，请在json格式诊断结果中查询完整卡间等待关系链",
    "MULTI_FAULT_IN_KNOWLEDGE_GRAPH": "诊断出多个故障，已按照优先级排序，请重点排查靠前的故障"
}

note_msg_en = {
    "MULTI_RANK_NOTE_MSG": "Root Cause Node Analysis has detected multiple suspected root cause nodes, "
                           "and these nodes are prioritized for investigation.",
    "MAX_RANK_NOTE_MSG": "Note: Due to the excessive number of root cause nodes, only 16 are displayed, "
                         "sorted by error reporting time. All root cause nodes can be queried in diag_report.json.",
    "MAX_DEVICE_NOTE_MSG": "Note: The quantity of faults for some faulty devices is too huge, "
                           "so only 16 entries are displayed. All faulty devices can be queried in diag_report.json. ",
    "MAX_WORKER_CHAINS_NOTE_MSG": "Note: There are too many faulty devices with fault propagation chains, "
                                  "so only the first 16 devices are displayed. "
                                  "For details, please refer to diag_report.json. ",
    "NET_SINGLE_WORKER_MSG": "There is no network congestion in a single-worker environment.",
    "UNKNOWN_ROOT_ERROR_RANK": "The root cause node has not been identified, "
                               "so Fault Event Analysis attempts to detect all devices. ",
    "SOME_SUBTASKS_FAILED": "Some analysis sub-items under this analysis module have failed to execute, "
                            "which may affect the diagnostic results and lead to inaccuracies. "
                            "The failure information can be queried in diag_report.json.",
    "SOME_DEVICE_FAILED": "In the retransmission overload fault chain, "
                          "the peer node cannot confirm the specific Worker ID and Device ID. "
                          "Please check the corresponding Device equipment via IP. ",
    "FAULT_CHAINS_NOTE": "The fault propagation chain only displays the longest link for each faulty device. ",
    "FAULT_CHAINS_MAX_NOTE": "Some faulty devices have a maximum number of fault propagation chains exceeding {}; "
                             "please refer to the complete fault propagation chains information "
                             "in diag_report.json.",
    "NO_GROUP_RANK_INFO_NOTE": "The communication domain information corresponding to the Notify timeout error "
                               "was not parsed, and the root cause node result may be inaccurate. ",
    "REMOTE_LINKS_NOTE": "There are waiting relationships between some devices, "
                         "and an example is demonstrated in the \"Inter-Device Waiting Chain\".",
    "REMOTE_LINKS_MAX_NOTE": "Some inter-device waiting chains are too long, so only 16 of them are displayed, "
                             "please refer to the complete inter-device chains information in diag_report.json.",
    "MULTI_FAULT_IN_KNOWLEDGE_GRAPH": "Multiple faults have been diagnosed and sorted by priority. "
                                      "Please focus on troubleshooting the top-ranked ones."
}


def get_label_for_language(specified_language: str = ""):
    lang = specified_language or LANG
    if lang == "en":
        return label_en
    else:
        return label_zh


def get_note_msg_by_code(code: str):
    if LANG == "en":
        return note_msg_en.get(code, "")
    else:
        return note_msg_zh.get(code, "")


def get_fault_description_by_code(code: int):
    if LANG == "en":
        return fault_descriptions_en.get(code, "")
    else:
        return fault_descriptions_zh.get(code, "")


def get_fault_entity_details_by_code(code: str):
    if LANG == "en":
        entity = fault_entities_en.get(code, "")
    else:
        entity = fault_entities_zh.get(code, "")
    return entity.get("cause", ""), entity.get("description"), entity.get("suggestion", "")


def get_language():
    lang = LANGUAGE.lower()
    return lang if lang in ["zh", "en"] else "zh"


LANG = get_language()
