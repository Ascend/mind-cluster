#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
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
"""掉卡类跨侧联合判定（RFC 模式2/3/6/7/10/11/12/13/14）。

从 kg_diag_job 拆分出的独立功能模块：负责超节点（super_pod）场景下
BMC（sel.txt/fdm_log）掉卡锚点与 host（kernel/bbox/os_info/device_info）
证据的跨 worker 关联、按板聚合与联合判定，避免 kg_diag_job 因行数超限。
"""

import copy
import logging
import os

from ascend_fd.configuration.config import DEFAULT_USER_CONF, KNOWLEDGE_GRAPH_CONF
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.model.response import RootCause
from ascend_fd.utils.constant.dev_log_const import (
    LOST_DEV_TIME_WINDOW_SEC,
    PRECHECK_BBOX_OS_OOM_EXCEPTION,
    PRECHECK_DEVICE_INFO_HEARTBEAT_LOST,
    PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC,
    PRECHECK_FDM_MATA_UNCORRECTED_ERROR,
    PRECHECK_KERNEL_CORE_LOCKUP,
    PRECHECK_KERNEL_ERR_TYPE3_MULTI_BIT_ECC,
    PRECHECK_KERNEL_EXECUTE_NON_EXECUTABLE,
    PRECHECK_KERNEL_HARD_LOCKUP,
    PRECHECK_KERNEL_OOM_HUGE_PAGE_EXHAUSTED,
    PRECHECK_KERNEL_PANIC,
    PRECHECK_KERNEL_PANIC_SOFTLOCKUP,
    PRECHECK_KERNEL_WATCHDOG_TIMEOUT,
    PRECHECK_OS_INFO_OS_OOM_EXCEPTION,
    PRECHECK_SEL_LOG_LOST_NPU,
    PRECHECK_SEL_MULTIBIT_ECC,
    PRECHECK_SMMU_TIMEOUT,
)
from ascend_fd.utils.constant.ub_const import PRECHECK_PREFIX
from ascend_fd.utils.fault_code import (
    COMP_LOST_DEV_CORELOCKUP_CUSTOM,
    COMP_LOST_DEV_HARDLOCKUP_CUSTOM,
    COMP_LOST_DEV_HBM_THERMAL_OSOOM_CUSTOM,
    COMP_LOST_DEV_HBM_OOM_CUSTOM,
    COMP_LOST_DEV_ILLEGAL_ADDR_CUSTOM,
    COMP_LOST_DEV_MATA_ECC_CUSTOM,
    COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM,
    COMP_LOST_DEV_SOFTLOCKUP_CUSTOM,
    COMP_LOST_DEV_UCE_PANIC_CUSTOM,
)
from ascend_fd.utils.i18n import LANG
from ascend_fd.utils.load_kg_config import EntityAttribute, Schema
from ascend_fd.utils.status import FileNotExistError
from ascend_fd.utils.tool import (
    collect_parse_results,
    get_event_occur_times,
    load_json_data,
    times_within_window,
)

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
# 跨侧掉卡判定的事件来源侧标识（与 _CROSS_SIDE_LOST_DEV_PATTERNS 的元组第二项对应）
BMC_SIDE = "bmc"
HOST_SIDE = "host"


def get_super_pod_analyzer_dict(cfg, parsed_saver):
    analyzer_dict = cfg.root_worker_devices.copy()
    if parsed_saver.infer_task_flag:
        instance_key = parsed_saver.infer_instance
        container_ips = parsed_saver.cluster_info.get(instance_key, [])
        infer_workers = {
            parsed_saver.container_worker_map[ip] for ip in container_ips if ip in parsed_saver.container_worker_map
        }
        worker_mappings = [
            (parsed_saver.bmc_path_dict, get_host_worker_name_by_bmc_worker_name),
            (parsed_saver.lcne_path_dict, get_host_worker_name_by_lcne_worker_name),
        ]
        for path_dict, name_func in worker_mappings:
            filtered = {worker: path for worker, path in path_dict.items() if name_func(worker, cfg) in infer_workers}
            analyzer_dict.update(filtered)
    else:
        analyzer_dict.update(parsed_saver.bmc_path_dict)
        analyzer_dict.update(parsed_saver.lcne_path_dict)
    return analyzer_dict


def _get_analyzer_file(worker_name, parsed_saver):
    """
    Get the kg_analyzer file path
    :param worker_name: worker name
    :param parsed_saver: to obtain worker directory path
    :return: path of the JSON file
    """
    input_path = ""
    try:
        input_path = _get_kg_parser_file(worker_name, parsed_saver, "kg-analyzer.json", fuzzy_match=True)
    except FileNotExistError as err:
        kg_logger.error(str(err))
    return input_path


def _get_kg_parser_file(worker_name, parsed_saver, file_name, fuzzy_match=False):
    """
    Get the parsed JSON file in the parse folder
    :param worker_name: worker name
    :param parsed_saver: to obtain worker directory path
    :param file_name: file name
    :return: file path for single worker
    """
    worker_dir = parsed_saver.get_worker_dir_path(worker_name)
    if not worker_dir:
        kg_logger.error('The %s dir is not exist.', worker_name)
        raise FileNotExistError(f'The {worker_name} dir is not exist.')
    path_list = []
    if fuzzy_match:
        path_list = collect_parse_results(worker_dir, file_name)
    file_path = path_list[0] if path_list else os.path.join(worker_dir, file_name)
    if not os.path.exists(file_path):
        kg_logger.error('The %s is not exist in %s.', file_name, worker_dir)
        raise FileNotExistError(f'The {file_name} is not exist in {worker_dir}.')
    return file_path


def get_host_worker_name_by_bmc_worker_name(worker_name, cfg):
    host_info_instance = cfg.parsed_saver.super_pod_info_saver.find_host_by_bmc_worker_name(
        worker_name.split()[0].split(":")[-1]
    )
    return host_info_instance and host_info_instance.log_dir.split("/")[1]


def get_host_worker_name_by_lcne_worker_name(worker_name, cfg):
    host_info_instance = cfg.parsed_saver.super_pod_info_saver.find_host_by_lcne_worker_name(
        worker_name.split()[0].split(":")[-1]
    )
    return host_info_instance and host_info_instance.log_dir.split("/")[1]


def _build_lost_dev_link(code, cause, anchor_event):
    """构建掉卡类根因的链文本，锚点事件 key_info 太长时截断。"""
    anchor_info = anchor_event.get("key_info", "") if isinstance(anchor_event, dict) else ""
    return "{}（{}）-> 掉卡（{}）".format(code, cause, (anchor_info[:30] if anchor_info else "no key info"))


def cross_side_lost_device_analyze(cfg, parsed_saver):
    """
    掉卡类跨侧联合判定（模式2/3/6/7/10/11/12/13/14）：
    - sel.txt/fdm_log（BMC worker）掉卡锚点与 kernel/bbox/os_info（host worker）证据跨 worker 关联
    - worker 级关联：BMC/LCNE worker 通过 super_pod_info_saver（dmidecode/fruinfo 的 SN、A5 pod 的 SlotId）映射到 host worker
    - 设备级关联：按 source_device，清洗后 BMC 与 host 事件均为统一设备号格式，直接归并判定
    - pre_analyze 已过滤 PRECHECK_ 码，此处直接读取各 worker 的 kg-analyzer.json
    :return: {job_name: {"worker_name": host_worker, "root_causes": {code: RootCause}}}
    """
    schema = Schema([DEFAULT_USER_CONF, KNOWLEDGE_GRAPH_CONF])
    # 仅超节点场景调用（见 start_kg_diag_job），BMC/LCNE worker 与 host worker 都在 analyzer_dict 中
    analyzer_dict = get_super_pod_analyzer_dict(cfg, parsed_saver)
    worker_precheck = _collect_worker_precheck(parsed_saver, analyzer_dict)
    host_workers = _aggregate_worker_precheck_by_slot(worker_precheck, cfg)
    return _build_cross_side_results(host_workers, schema)


def _collect_worker_precheck(parsed_saver, analyzer_dict):
    """收集各 worker 的 PRECHECK 事件：worker -> device -> {code: precheck event}。"""
    worker_precheck = {}
    for worker_name in analyzer_dict:
        analyzer_file = _get_analyzer_file(worker_name, parsed_saver)
        if not analyzer_file:
            continue
        kg_analyzer = load_json_data(analyzer_file)
        resp = kg_analyzer.get("response", {}) if isinstance(kg_analyzer, dict) else {}
        per_device = {}
        for device_id, device_cause in resp.items():
            if not isinstance(device_cause, dict):
                continue
            root_causes = (
                device_cause.get("root_causes", {}) if isinstance(device_cause.get("root_causes"), dict) else {}
            )
            precheck = {
                code: event
                for code, event in root_causes.items()
                if code.startswith(PRECHECK_PREFIX) and isinstance(event, dict)
            }
            if precheck:
                per_device[device_id] = precheck
        if per_device:
            worker_precheck[worker_name] = per_device
    return worker_precheck


def _aggregate_worker_precheck_by_slot(worker_precheck, cfg):
    """按板（slot）聚合：host_worker -> device(slot) -> {side: {code: precheck event}}。

    BMC 事件映射到 host worker 并按板归并；同板下 host 侧事件并入同一 slot 的 HOST_SIDE。
    LCNE worker 无掉卡证据，直接跳过。
    """
    host_workers = {}
    for worker_name, devices in worker_precheck.items():
        if worker_name.startswith("LCNE"):
            continue
        if worker_name.startswith("BMC"):
            host_worker = get_host_worker_name_by_bmc_worker_name(worker_name, cfg)
            side = BMC_SIDE
        else:
            host_worker = worker_name
            side = HOST_SIDE
        if not host_worker:
            continue
        worker_devices = host_workers.setdefault(host_worker, {})
        for device, precheck in devices.items():
            worker_devices.setdefault(device, {}).setdefault(side, {}).update(precheck)
    return host_workers


def _build_cross_side_results(host_workers, schema):
    """逐板做跨侧掉卡判定并构造根因：{job_name: {"worker_name", "root_causes"}}。"""
    extra_results = {}
    for host_worker, worker_devices in host_workers.items():
        causes = {}
        for device, sides in worker_devices.items():
            bmc_events = sides.get(BMC_SIDE, {})
            host_events = sides.get(HOST_SIDE, {})
            matches = _judge_cross_side_lost_device(bmc_events, host_events)
            for code, match in matches.items():
                if code in causes:
                    continue
                causes[code] = _build_cross_side_root_cause(code, match, device, schema)
        if causes:
            extra_results[f"CROSS_SIDE_LOST_DEVICE_{host_worker}"] = {
                "worker_name": host_worker,
                "root_causes": causes,
            }
    return extra_results


def _build_cross_side_root_cause(code, match, device, schema):
    """构造单个跨侧掉卡根因；match 首元素为掉卡锚点，作为链起点。"""
    entity = schema.get_schema_entity(code)
    entities_attr = entity.attribute if entity else EntityAttribute(dict())
    cause = entities_attr.to_json().get(f"cause_{LANG}", "") or entities_attr.to_json().get(f"cause_{LANG.lower()}", "")
    # 清洗后 source_device 已是统一格式，直接透传即可，root_device 与 chains 保持一致
    events_attr = [copy.deepcopy(event) for event in match]
    chains = {device: _build_lost_dev_link(code, cause, match[0])}
    return RootCause(code, entities_attr, events_attr, chains)


# 跨侧掉卡联合判定模式表：(out_code, ((事件来源, PRECHECK码, 是否参与锚点时间窗), ...))
# - 事件来源：BMC_SIDE（sel.txt/fdm_log）或 HOST_SIDE（kernel/bbox/os_info/device_info）
# - 参与时间窗的码：须在掉卡锚点前推 1min 内命中（按故障码分别解析 occurrence 时间）
# - 不参与时间窗的码：仅判事件存在（如 heartbeat lost 无时间戳、只有状态）
# events_attribute 顺序即元组声明顺序，与 RFC 判定流程一致
_CROSS_SIDE_LOST_DEV_PATTERNS = (
    (
        COMP_LOST_DEV_UCE_PANIC_CUSTOM,
        (
            (BMC_SIDE, PRECHECK_FDM_MATA_UNCORRECTED_ERROR, True),
            (HOST_SIDE, PRECHECK_KERNEL_PANIC, True),
        ),
    ),
    (
        COMP_LOST_DEV_HBM_THERMAL_OSOOM_CUSTOM,
        (
            (HOST_SIDE, PRECHECK_BBOX_OS_OOM_EXCEPTION, True),
            (BMC_SIDE, PRECHECK_SEL_MULTIBIT_ECC, True),
        ),
    ),
    (
        COMP_LOST_DEV_SMMU_CMD_TIMEOUT_CUSTOM,
        ((HOST_SIDE, PRECHECK_SMMU_TIMEOUT, True),),
    ),
    (
        COMP_LOST_DEV_MATA_ECC_CUSTOM,
        (
            (BMC_SIDE, PRECHECK_FDM_MATA_DDRC_MULTI_BIT_ECC, True),
            (HOST_SIDE, PRECHECK_KERNEL_ERR_TYPE3_MULTI_BIT_ECC, True),
        ),
    ),
    (
        COMP_LOST_DEV_HBM_OOM_CUSTOM,
        (
            (HOST_SIDE, PRECHECK_OS_INFO_OS_OOM_EXCEPTION, True),
            (HOST_SIDE, PRECHECK_KERNEL_OOM_HUGE_PAGE_EXHAUSTED, True),
        ),
    ),
    (
        COMP_LOST_DEV_ILLEGAL_ADDR_CUSTOM,
        ((HOST_SIDE, PRECHECK_KERNEL_EXECUTE_NON_EXECUTABLE, True),),
    ),
    (
        COMP_LOST_DEV_HARDLOCKUP_CUSTOM,
        ((HOST_SIDE, PRECHECK_KERNEL_HARD_LOCKUP, True),),
    ),
    (
        COMP_LOST_DEV_CORELOCKUP_CUSTOM,
        ((HOST_SIDE, PRECHECK_KERNEL_CORE_LOCKUP, True),),
    ),
    (
        COMP_LOST_DEV_SOFTLOCKUP_CUSTOM,
        (
            (HOST_SIDE, PRECHECK_DEVICE_INFO_HEARTBEAT_LOST, False),
            (HOST_SIDE, PRECHECK_KERNEL_PANIC_SOFTLOCKUP, True),
            (HOST_SIDE, PRECHECK_KERNEL_WATCHDOG_TIMEOUT, True),
        ),
    ),
)


def _judge_cross_side_lost_device(bmc_events, host_events):
    """
    跨侧联合判定：返回 {out_code: events_attribute 列表}。
    依据 RFC：各模式先经 BMC sel.txt 确认掉卡告警（时间锚点，各模式均需锚点），
    再在掉卡事件时间点往前推 1min 内查对应证据日志，条件均满足即命中。
    - 模式2 (UCE_PANIC): kernel panic + fdm MATA UCE
    - 模式3 (THERMAL_OSOOM): sel掉卡锚点前1min内 bbox 0xA4040001 + sel 0x56000003+0x80e01801
    - 模式6 (SMMU): kernel svm_smmu CMD_SYNC timeout
    - 模式7 (MATA_ECC): kernel error_type:3 + fdm MATA DDRC
    - 模式10 (HBM_OOM): os_info 0xA4040001 + kernel oom huge page
    - 模式11 (ILLEGAL_ADDR): kernel non-executable memory
    - 模式12 (HARDLOCKUP): kernel watchdog hard LOCKUP
    - 模式13 (CORELOCKUP): kernel panic Core LOCKUP
    - 模式14 (SOFTLOCKUP): device_info heartbeat（仅判状态）+ kernel panic softlockup + watchdog
    """
    results = {}
    sel_anchor = bmc_events.get(PRECHECK_SEL_LOG_LOST_NPU)
    if sel_anchor is None:
        return results
    anchor_times = get_event_occur_times(sel_anchor)
    side_source = {BMC_SIDE: bmc_events, HOST_SIDE: host_events}
    for out_code, evidence_specs in _CROSS_SIDE_LOST_DEV_PATTERNS:
        events = [sel_anchor]
        window_evidence = {}
        matched = True
        for idx, (side, code, in_window) in enumerate(evidence_specs):
            event = side_source[side].get(code)
            if event is None:
                matched = False
                break
            events.append(event)
            # 参与时间窗的证据按故障码分组 occurrence 时间，锚点为中心判定
            if in_window:
                window_evidence[event.get("event_code", f"evidence_{idx}")] = get_event_occur_times(event)
        if not matched:
            continue
        if not times_within_window(anchor_times, window_evidence, LOST_DEV_TIME_WINDOW_SEC):
            continue
        results[out_code] = events
    return results
