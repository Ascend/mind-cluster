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
import logging
import os
import re

from ascend_fd.model.node_info import FaultFilterTime
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.kg_engine_main import kg_engine_analyze
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.model.package_data import PackageData
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.model.response import RootCause, Response
from ascend_fd.pkg.diag.message import MULTI_FAULT_IN_KNOWLEDGE_GRAPH, SOME_SUBTASKS_FAILED, FAULT_CHAINS_NOTE, \
    FAULT_CHAINS_MAX_NOTE
from ascend_fd.utils.constant.str_const import SUPER_POD_SCENE
from ascend_fd.utils.fault_code import KG_DIAGNOSIS_NORMAL, HCCL_FAULT_LIST, DEVICE_CQE_FAULT, LINK_DOWN_FAULT, \
    LINK_STATUS_CHANGE, PRE_TRACEBACK_FAULT, CANN_ERRCODE_CUSTOM, AISW_CANN_MEMORY_INFO, OOM_CANN_FAULT_LIST, \
    PRE_SWITCH_FAULT, MINDIE_ERRCODE_COMMON
from ascend_fd.utils.load_kg_config import EntityAttribute
from ascend_fd.utils.regular_table import SORT_RULES, LOWEST_PRIORITY_NUM, PRE_COMP_OS_FAULT, PRE_COMP_SWITCH_FAULT, \
    PRE_AMCT_FAULT, MIN_TIME, MAX_TIME, OS_FAULT_PREFIX, MINDIE_FAULT_PREFIX
from ascend_fd.utils.status import FileNotExistError, InfoNotFoundError, InnerError
from ascend_fd.utils.tool import MultiProcessJob, get_version, load_json_data, get_component_version, \
    get_parse_json, collect_parse_results
from ascend_fd.configuration.config import DEFAULT_USER_CONF
from ascend_fd.utils.i18n import LANG

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
MAX_WORKER_CHAIN_NUM = 3


def start_kg_diag_job(cfg):
    """
    Start kg diag job
    :param cfg: diag config
    :return:
    {
        analyze_success: true / false
        fault: [
            {
                code: 故障错误码,
                class: 故障类别,
                ...（故障建议、描述等）
                event_attr: {设备1: {event_attribute: [{清洗返回结果}，...]}}
                fault_source: [设备列表],
                fault_chains: [
                    {"worker": [故障设备1, 故障设备2, ...], "chains": "关键传播链"},
                    {设备n相关故障链信息...},
                    ...
                ]
            },
            {code2相关信息...}
            ...
        ],
        note_msgs: xx
    }
    """
    results = dict()
    failed_details = dict()
    version_info = dict()
    parsed_saver = cfg.parsed_saver
    multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=20, task_id=cfg.task_id)
    fault_filter_time = cfg.fault_filter_time
    if parsed_saver.scene == SUPER_POD_SCENE:
        analyzer_dict = get_super_pod_analyzer_dict(cfg, parsed_saver)
    else:
        analyzer_dict = cfg.root_worker_devices
    for worker_name, device_list in analyzer_dict.items():
        analyzer_file = _get_analyzer_file(worker_name, parsed_saver)
        if analyzer_file:
            pre_results, pre_failed_details = pre_analyze_job(worker_name, device_list, analyzer_file,
                                                              fault_filter_time)
            # save node and its version info
            version_info[worker_name] = pre_results.get('version_info', dict())
            results.update(pre_results)
            failed_details.update(pre_failed_details)
            continue
        job_name = f"KNOWLEDGE_GRAPH_WORKER_{worker_name}"
        multiprocess_job.add_security_job(job_name,
                                          _kg_diag_job, worker_name, device_list, parsed_saver, job_name)
    multiprocess_job.failed_raise = not results
    multi_results, multi_failed_details = multiprocess_job.join_and_get_results()
    results.update(multi_results)
    results['version_info'] = _merge_version_info(version_info)
    failed_details.update(multi_failed_details)
    if len(failed_details) == len(cfg.root_worker_devices):
        first_key, first_value = next(iter(failed_details.items()))
        raise InnerError(f"All subjobs execute failed. The first subjob: {first_key}, error reason is: {first_value}")
    return hand_all_root_cause(results, failed_details, cfg)


def get_super_pod_analyzer_dict(cfg, parsed_saver):
    analyzer_dict = cfg.root_worker_devices.copy()
    if parsed_saver.infer_task_flag:
        instance_key = parsed_saver.infer_instance
        container_ips = parsed_saver.cluster_info.get(instance_key, [])
        infer_workers = {
            parsed_saver.container_worker_map[ip]
            for ip in container_ips
            if ip in parsed_saver.container_worker_map
        }
        worker_mappings = [
            (parsed_saver.bmc_path_dict, get_host_worker_name_by_bmc_worker_name),
            (parsed_saver.lcne_path_dict, get_host_worker_name_by_lcne_worker_name)
        ]
        for path_dict, name_func in worker_mappings:
            filtered = {
                worker: path
                for worker, path in path_dict.items()
                if name_func(worker, cfg) in infer_workers
            }
            analyzer_dict.update(filtered)
    else:
        analyzer_dict.update(parsed_saver.bmc_path_dict)
        analyzer_dict.update(parsed_saver.lcne_path_dict)
    return analyzer_dict


def single_diag_job(parsed_data, cfg):
    """
    Single diag job
    :param parsed_data: single parsed data
    :param cfg: diag config
    :return: kg_result, the single-diag result
    """
    results = dict()
    root_device_causes = parsed_data.get("response", {})
    response = _get_pre_response(root_device_causes, "SINGLE_DIAG_WORKER")
    if response.root_causes:
        result = {"worker-local": {"worker_name": "worker-local", "root_causes": response.root_causes},
                  "version_info": get_component_version(parsed_data)}
        results.update(result)
    if not response.analyze_success:
        raise InnerError(f"The single-diag analyze failed. The reason is: {response.error}")
    return hand_all_root_cause(results, dict(), cfg)


def hand_all_root_cause(results, failed_details, cfg=None):
    """
    Hand all root cause
    :param results: analysis results
    :param failed_details: failed details
    :param cfg: diag config
    :return: the results of kg_diag in the worker
    """
    tmp_code_dict = dict()
    kg_result = {"analyze_success": True, "version_info": results.pop("version_info", {})}
    for result_dict in results.values():
        worker_name = result_dict.get("worker_name", "")
        # 处理推理任务
        if cfg and cfg.parsed_saver.infer_task_flag:
            container_info = get_parse_json(_get_kg_parser_file(worker_name, cfg.parsed_saver, "server-info.json"))
            container_ip = container_info.get("container_ip", "")
            infer_group = cfg.parsed_saver.ip_infer_group.get(container_ip)
            if (not infer_group or infer_group.infer_group_name != cfg.parsed_saver.infer_instance) and not any(
                    worker_name.startswith(prefix) for prefix in ["BMC", "LCNE"]):
                continue
        for code, root_cause in result_dict.get("root_causes", {}).items():
            handle_root_cause(code, root_cause, tmp_code_dict, worker_name)
    fault_info, fault_chains_flag, max_worker_chains_flag = _filter_and_sort_root_cause(tmp_code_dict, cfg)
    kg_result["note_msgs"] = _get_kg_note_msgs(failed_details, fault_chains_flag, max_worker_chains_flag)
    if len(fault_info) > 1:
        kg_result["note_msgs"].append(MULTI_FAULT_IN_KNOWLEDGE_GRAPH)
    kg_result["fault"] = fault_info
    if failed_details:
        kg_result['failed_jobs'] = failed_details
    return kg_result


def handle_root_cause(code, root_cause, tmp_code_dict, worker_name):
    """
    Handle root cause result
    :param code: fault code string
    :param root_cause: fault root cause
    :param tmp_code_dict: temporary code dict
    :param worker_name: worker name
    """
    for event_attr in root_cause.events_attribute or [None]:
        device = event_attr.get("source_device", "Unknown") if event_attr else "Unknown"
        entities_attr = root_cause.entities_attribute
        if code.startswith((PRE_COMP_OS_FAULT, PRE_COMP_SWITCH_FAULT, PRE_AMCT_FAULT)) or \
                entities_attr.get("component", "") == "AI Server":
            root_device = "{}".format(worker_name)
        else:
            root_device = "{} device-{}".format(worker_name, device)
        if code not in tmp_code_dict:
            tmp_code_dict[code] = {"code": code}
            tmp_code_dict[code].update(entities_attr)
        if event_attr:
            tmp_code_dict[code].setdefault("event_attr", dict()).update({root_device: [event_attr]})
        # merge workers with the same fault chain.
        device_chains = root_cause.chains.get(device, "")
        if device_chains:
            fault_chains = tmp_code_dict[code].setdefault("fault_chains", {device_chains: []})
            fault_chains_worker = fault_chains.setdefault(device_chains, [])
            # don't add duplicate root_device
            if root_device not in fault_chains_worker:
                fault_chains_worker.append(root_device)
        # don't add duplicate root_device, adapt scenario e.g. ['worker-0 device-Unknown', 'worker-0 device-Unknown'...]
        fault_source = tmp_code_dict[code].setdefault("fault_source", list())
        if root_device not in fault_source:
            fault_source.append(root_device)


def pre_analyze_job(worker_name, root_device_list, kg_analyzer_source, fault_filter_time):
    """
    Pre-analyzed job
    :param worker_name: worker name
    :param root_device_list: root devices
    :param kg_analyzer_source: the input path of the kg_analyzer file or a same dict form of sdk input
    :param fault_filter_time: time to filter fault before resuming training or the first plog log time.
    :return: inference result, failed job list and os error exception
    """
    kg_logger.info("Start pre-analyzed task for %s.", worker_name)
    job_name = f"PRE_ANALYZE_WORKER_{worker_name}"
    response = Response()
    failed_details = dict()
    results = dict()
    try:
        response = _check_and_get_results(worker_name, root_device_list, kg_analyzer_source, job_name,
                                          fault_filter_time)
    except Exception as err:
        failed_details.update({job_name: str(err)})
        kg_logger.error(str(err))
    if response.root_causes:
        results = {job_name: {"worker_name": worker_name, "root_causes": response.root_causes}}
        if isinstance(kg_analyzer_source, str):
            parse_json = get_parse_json(kg_analyzer_source)
            version_info = get_component_version(parse_json)
            if version_info:
                results.update({"version_info": version_info})
    return results, failed_details


def _check_and_get_results(worker_name, root_device_list, parse_results, job_name, fault_filter_time):
    """
    Check and get the analysis results
    :param worker_name: worker name
    :param root_device_list: root devices
    :param parse_results: the input file paths or a dict format of kg_analyzer from sdk
    :param job_name: job name
    :return: the response analyzed in advance
    """
    _version_check(worker_name, parse_results)
    root_device_causes = _get_root_device_causes(parse_results, worker_name, root_device_list)
    response = _get_pre_response(root_device_causes, job_name, fault_filter_time)
    _resp_check(response, worker_name)
    return response


def extract_last_brackets(text):
    """
    栈结构提取最后一个括号块
    """
    stack = []
    last_start = last_end = None
    for i, char in enumerate(text):
        # 遇到左括号：将当前索引压入栈
        if char == '(':
            stack.append(i)
            continue
        # 遇到右括号：尝试匹配栈顶的左括号
        if char == ')':
            if not stack:
                continue
            # 弹出栈顶的左括号索引（与该右括号匹配）
            start = stack.pop()
            # 若弹出后栈为空，说明这是最后一层括号对
            # 记录区间为 (start+1, i)，排除左括号和右括号本身
            if not stack:
                last_start, last_end = start + 1, i
    return text[last_start:last_end] if last_start is not None else None


def _get_detail_from_key_info(code_result: dict):
    """
    处理lcne日志中description缺失的部分
    """
    description_filed = f"description_{LANG}"
    entities_attr = code_result.get("entities_attribute", {})
    if entities_attr.get(description_filed) == '${detail}':
        events_attr = code_result.get("events_attribute", [{}])
        key_info = events_attr[0].get("key_info", "") if events_attr else ""
        if code_result.get('code') == 'Comp_Bus_Custom_01' and key_info:
            # 最后一个冒号和最后一个句号之间之间的内容
            pattern = r":(?!.*:)(.*?)\.(?=\.*$)"
            match = re.search(pattern, key_info)
            if match:
                entities_attr[description_filed] = match.group(1).strip()
            return
        # 栈结构提取最后一个括号块
        match = extract_last_brackets(key_info)
        if match:
            entities_attr[description_filed] = match


def _get_root_device_causes(kg_analyzer_source, worker_name, root_device_list):
    """
    Get the analysis result of the root cause device
    :param kg_analyzer_source: input path of the kg_analyzer file or its dict format input from sdk
    :param worker_name: worker name
    :param root_device_list: root devices
    :return: the analysis result of the root cause device
    """
    root_device_causes = {}
    kg_analyzer = load_json_data(kg_analyzer_source) if isinstance(kg_analyzer_source, str) else kg_analyzer_source
    resp_json = kg_analyzer.get("response", {})
    # 处理lcne日志中description缺失的部分
    for worker_result in resp_json.values():
        for code_result in worker_result.get("root_causes", {}).values():
            if "Bus" in code_result.get("code", ""):
                _get_detail_from_key_info(code_result)
    if not root_device_list:
        # 当提前分析的结果中有具体的device时，过滤掉device为Unknown的
        device_causes = {device: resp for device, resp in resp_json.items() if device != "Unknown"}
        return device_causes or resp_json
    for device_id in root_device_list:
        device_cause = resp_json.get(device_id) or resp_json.get("Unknown")
        if device_cause:
            root_device_causes.update({device_id: device_cause})
            continue
        kg_logger.info("[%s device-%s] didn't identify any fault events.", worker_name, device_id)
    return root_device_causes


def _get_pre_response(root_device_causes, job_name, fault_filter_time=FaultFilterTime(MIN_TIME, MAX_TIME)):
    """
    Get the response of the advance diagnosis analysis
    :param root_device_causes: the analysis result of the root cause device
    :param job_name: job name
    :return: the response analyzed in advance
    """
    response = Response()
    if not root_device_causes:
        response.analyze_success = True
        response.root_causes = response.NORMAL_ROOT_CAUSES
        return response

    analyze_flags = []
    root_causes = dict()
    for device_id, device_cause in root_device_causes.items():
        analyze_success = device_cause.get("analyze_success", False)
        analyze_flags.append(analyze_success)
        if not analyze_success:
            device_job_name = f"{job_name}_device-{device_id}"
            error_info = device_cause.get("error", "")
            kg_logger.warning("The %s job is executed failed. The reason is %s.", device_job_name, error_info)
            continue
        all_events_dict = device_cause.get("root_causes", {})
        oom_cann_faults = set(all_events_dict.keys()) & set(OOM_CANN_FAULT_LIST)
        for code, event in all_events_dict.items():
            # if the fault of the current device don't contain any of OOM_CANN_FAULT_LIST, skip AISW_CANN_MEMORY_INFO
            if code == AISW_CANN_MEMORY_INFO and not oom_cann_faults:
                continue
            root_cause = root_causes.get(code)
            events_attr_device = _filter_by_resuming_train_time(event.get("events_attribute", []),
                                                                fault_filter_time)
            if not events_attr_device:
                continue
            chains = event.get("chains", {})
            if root_cause:
                root_cause.events_attribute.extend(events_attr_device)
                root_cause.chains.update(chains)
            else:
                entities_attr = EntityAttribute(event.get("entities_attribute"))
                root_cause = RootCause(event.get("code", ""), entities_attr, events_attr_device, chains)
            root_causes.update({code: root_cause})
        root_causes = root_causes or response.NORMAL_ROOT_CAUSES
    return _get_analyze_response(root_causes, any(analyze_flags), job_name)


def _filter_by_resuming_train_time(events_attr_device, fault_filter_time):
    """
    Filter event by the last resuming starting time after the breakpoint
    :param events_attr_device: event list to filter
    :param fault_filter_time: the resuming training time or the first plog time to compare with
    :return: a filtered event list
    """
    filtered_event_list = []
    for event in events_attr_device:
        # faults from npu_info_before/after has no occurrence info, they occur in the end train time
        # faults from OS would not be filtered as well since the year info usually lost
        if "occurrence" not in event or event.get("event_code", "").startswith((OS_FAULT_PREFIX, MINDIE_FAULT_PREFIX)):
            filtered_event_list.append(event)
            continue
        for occurrence_tuple in sorted(event.get("occurrence", [])):
            # occurrence_tuple format pattern: (occur_time, key_info)
            occur_time = _process_occur_time(occurrence_tuple[0], fault_filter_time.start_train_time)
            if fault_filter_time.start_train_time <= occur_time <= fault_filter_time.end_train_time:
                event["occur_time"] = occurrence_tuple[0]
                event["key_info"] = occurrence_tuple[1]
                filtered_event_list.append(event)
                break
    return filtered_event_list


def _process_occur_time(occur_time, resuming_training_time):
    """
    Check whether the occurred time has a year digit
    :param occur_time: the fault occurring time
    :param resuming_training_time: the resuming training time to extract the year info
    :return: processed time format
    """
    # normal occur time format: YYYY-MM-DD hh:mm:ss.****** volcano fault time format MM-DD hh:mm:ss.******
    if occur_time and " " in occur_time:
        date = occur_time.split(" ")[0]
        valid_date_len = 10
        # volcano related faults have no year in the occurred time
        if len(date) < valid_date_len:
            digits_of_year = 5
            occur_time = resuming_training_time[:digits_of_year] + occur_time
        occur_time = occur_time.replace(" ", "-")
    return occur_time


def _kg_diag_job(worker_name, root_device_list, parsed_saver, job_name):
    """
    Knowledge graph diagnosis job
    :param worker_name: worker name
    :param root_device_list: root devices in the worker
    :param parsed_saver: to obtain file path
    :param job_name: job name
    :return: the results of kg_diag in the worker
    """
    kg_logger.info("Start knowledge graph diagnosis task for %s.", worker_name)
    input_file = _get_kg_parser_file(worker_name, parsed_saver, "kg-parser.json", fuzzy_match=True)
    _version_check(worker_name, input_file)
    package_data = PackageData([], input_file)
    root_device_list = root_device_list or list(package_data.fault_devices)
    if root_device_list:
        response = _single_device_kg_diag(root_device_list, package_data.event_map, job_name)
    else:
        response = kg_engine_analyze([DEFAULT_USER_CONF], package_data)
    _resp_check(response, worker_name)
    return {"worker_name": worker_name, "root_causes": response.root_causes}


def _single_device_kg_diag(root_device_list, all_event_map, job_name):
    """
    Knowledge graph diagnosis on each device in the current worker
    :param root_device_list: root devices in the worker
    :param all_event_map: all events map
    :param job_name: job name
    :return: inference result
    """
    analyze_flags = []
    root_causes = dict()
    for root_device in root_device_list:
        package_data = PackageData([root_device])
        package_data.load_single_device_events(all_event_map, root_device)
        resp = kg_engine_analyze([DEFAULT_USER_CONF], package_data)
        analyze_success = resp.analyze_success
        analyze_flags.append(analyze_success)
        if not analyze_success:
            device_job_name = f"{job_name}_device-{root_device}"
            kg_logger.warning("The %s job is executed failed. The reason is %s.", device_job_name, resp.error)
            continue
        for code, event in resp.root_causes.items():
            root_cause = root_causes.get(code)
            if not root_cause:
                root_causes.update({code: event})
                continue
            events_attribute_dev = event.events_attribute or []
            root_cause.events_attribute.extend(events_attribute_dev)
            root_cause.chains.update(event.chains)
            root_causes.update({code: root_cause})
    return _get_analyze_response(root_causes, any(analyze_flags), job_name)


def _get_analyze_response(root_causes, worker_analyze_flag, job_name):
    """
    Get analyze response
    :param root_causes: root causes
    :param worker_analyze_flag: worker analyze flag
    :param job_name: job name
    :return: inference result
    """
    resp = Response()
    _filter_in_device(root_causes)
    resp.analyze_success = worker_analyze_flag
    if not worker_analyze_flag:
        resp.error = (f"The {job_name} job is executed failed. "
                      f"Check the subtasks whose names start with {job_name}_device.")
    resp.root_causes = root_causes
    return resp


def _filter_in_device(root_causes):
    """
    Filtering between devices
    :param root_causes: root causes
    """
    # 当同一个故障事件诊断出具体的device时，过滤掉device为Unknown的
    for root_cause in root_causes.values():
        events_attr = root_cause.events_attribute
        events_attr_device = list(filter(lambda attr: attr.get("source_device", "Unknown") != "Unknown", events_attr))
        root_cause.events_attribute = events_attr_device or events_attr


def _version_check(worker_name, parse_result):
    """
    Check if the cleaning version and the diagnostic version of the diagnostic tool are consistent
    :param worker_name: worker name
    :param parse_result: the input file paths or a dict format of kg_analyzer from sdk
    """
    kg_parse_source = get_parse_json(parse_result) if isinstance(parse_result, str) else parse_result
    parse_version = kg_parse_source.get("parse_version", "")
    current_version = get_version()
    if parse_version != current_version:
        kg_logger.warning("The worker %s parse version %s is inconsistent with the current version %s",
                          worker_name, parse_version, current_version)


def _resp_check(response, worker_name):
    """
    Check the knowledge graph diagnostic analysis results
    :param response: the results of kg_diag in the worker
    :param worker_name: worker name
    """
    if not response.analyze_success:
        kg_logger.error("The kg-engine analyze %s failed. The reason is: %s", worker_name, response.error)
        raise InfoNotFoundError(f"The kg-engine analyze {worker_name} failed. The reason is: {response.error}")
    if not response.root_causes:
        kg_logger.warning("Knowledge graph diagnosis normally, "
                          "maybe 1. No related faults have occurred, 2. Unknown faults exist")


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


def _get_kg_note_msgs(failed_details: dict, fault_chains_flag: bool, max_worker_chains_flag: bool) -> list:
    """
    Get kg result note messages
    :param failed_details: failed detail dict
    :param fault_chains_flag: flag indicating whether fault chains exists
    :param max_worker_chains_flag: flag indicating whether fault worker chains num exceed 3
    :return: note message list
    """
    note_msgs_list = []
    # some analysis subtasks (not all) failed.
    if failed_details:
        note_msgs_list.append(SOME_SUBTASKS_FAILED)
    # if fault chains exist, add a note msg.
    if fault_chains_flag:
        note_msgs_list.append(FAULT_CHAINS_NOTE)
    # if the max number of workers that own the faulty chain is more than MAX_WORKER_CHAIN_NUM, add a note msg.
    if max_worker_chains_flag:
        note_msgs_list.append(FAULT_CHAINS_MAX_NOTE.format(MAX_WORKER_CHAIN_NUM))
    return note_msgs_list


def _filter_and_sort_root_cause(code_dict: dict, cfg):
    """
    Filter non-root cause root cause and sort
    :param code_dict: kg diagnostic result for all worker, the key is code
    :param cfg: diag config
    :return: root causes after filter and sort and max worker chains flag
    """
    fault_chains_flag, max_worker_chains_flag = False, False
    # filter non-root cause code
    _remove_non_root_cause(code_dict)
    if cfg and cfg.parsed_saver.scene == "super_pod":
        filter_super_pod_dict(code_dict, cfg)
    root_cause_list = list(code_dict.values())
    for val in root_cause_list:
        # sort worker_name list
        val.setdefault("fault_source", []).sort(key=_root_device_sort_func_in_single_cause(val))
        # sort fault chains by worker num of every chain
        fault_chains = val.setdefault("fault_chains", {})
        fault_chains_flag = fault_chains_flag or bool(fault_chains)
        max_worker_chains_flag = max_worker_chains_flag or len(fault_chains.keys()) > MAX_WORKER_CHAIN_NUM
        val["fault_chains"] = _format_and_sort_fault_chains(fault_chains)

    return _categorize_and_sort_root_cause_(root_cause_list), fault_chains_flag, max_worker_chains_flag


def _categorize_and_sort_root_cause_(root_cause_list: list):
    """
    Categorize root cause for sort
    :param root_cause_list: kg diagnostic result for all worker, the key is code
    :return: result_list
    """
    traceback_list = []
    cann_custom_list = []
    normal_and_pta_list = []
    amct_list = []
    mindie_custom_list = []
    for root_cause_item in root_cause_list:
        if str(root_cause_item.get('code', '')).startswith(PRE_TRACEBACK_FAULT):
            traceback_list.append(root_cause_item)
            continue
        if str(root_cause_item.get('code', '')).startswith(CANN_ERRCODE_CUSTOM):
            cann_custom_list.append(root_cause_item)
            continue
        if str(root_cause_item.get('code', '')).startswith(PRE_AMCT_FAULT):
            amct_list.append(root_cause_item)
            continue
        if str(root_cause_item.get('code', '')).startswith(MINDIE_ERRCODE_COMMON):
            mindie_custom_list.append(root_cause_item)
            continue
        normal_and_pta_list.append(root_cause_item)

    normal_and_pta_list.sort(key=_root_cause_sort_func)
    cann_custom_list.sort(key=_root_cause_sort_func)
    traceback_list.sort(key=_root_cause_sort_func)
    amct_list.sort(key=_root_cause_sort_func)

    return normal_and_pta_list + mindie_custom_list + cann_custom_list + traceback_list + amct_list


def _format_and_sort_fault_chains(fault_chains_dict: dict) -> list:
    """
    Format and sort fault chains by 1. worker num; 2. chains len
    :param fault_chains_dict: fault chains dict
    :return: fault chains list after format and sort
    """
    if not fault_chains_dict:
        return []
    new_chain_list = []
    for chains, workers in fault_chains_dict.items():
        new_chain_list.append({"worker": sorted(workers), "chains": chains})
    new_chain_list.sort(key=lambda x: (len(x.get("worker", [])), len(x.get("chains", "").split("-> "))), reverse=True)
    return new_chain_list


def get_host_worker_name_by_bmc_worker_name(worker_name, cfg):
    host_info_instance, _ = cfg.parsed_saver.super_pod_info_saver.find_from_bmc_worker_name(
        worker_name.split()[0].split(":")[-1])
    return host_info_instance and host_info_instance.log_dir.split("/")[1]


def get_host_worker_name_by_lcne_worker_name(worker_name, cfg):
    host_info_instance, _ = cfg.parsed_saver.super_pod_info_saver.find_from_lcne_worker_name(
        worker_name.split()[0].split(":")[-1])
    return host_info_instance and host_info_instance.log_dir.split("/")[1]


def filter_super_pod_dict(code_dict: dict, cfg):
    """
    Filter super pod cause fault in some scenes
    :param code_dict: kg diagnostic result for all worker, the key is code
    :param cfg: diag config
    :return: root cause list
    """
    all_host_root_worker = []
    for code_value in code_dict.values():
        for fault_source in code_value.get("fault_source", []):
            worker_name = fault_source.split()[0]
            if worker_name not in all_host_root_worker and not worker_name.startswith(
                    "BMC") and not worker_name.startswith("LCNE"):
                all_host_root_worker.append(worker_name)
    all_host_root_worker.extend([root_device for root_device in cfg.root_worker_devices.keys()])
    keys_to_move = generate_keys_to_move(all_host_root_worker, cfg, code_dict)
    for key in keys_to_move:
        code_dict.pop(key, None)


def generate_keys_to_move(all_host_root_worker, cfg, code_dict):
    keys_to_move = []
    for code_key, code_value in code_dict.items():
        filter_worker_list = []
        for fault_source in code_value.get("fault_source", []):
            worker_name = fault_source.split()[0]
            deal_fault_source_by_worker_name(worker_name, filter_worker_list, cfg, all_host_root_worker, fault_source)
        if len(filter_worker_list) == len(code_value.get("fault_source", [])):
            keys_to_move.append(code_key)
            continue
        if filter_worker_list:
            code_dict[code_key]["fault_source"] = [
                worker
                for worker in code_dict[code_key]["fault_source"]
                if worker not in filter_worker_list
            ]
    return keys_to_move


def deal_fault_source_by_worker_name(worker_name, filter_worker_list, cfg, all_host_root_worker, fault_source):
    if worker_name.startswith("BMC"):
        host_worker_name = get_host_worker_name_by_bmc_worker_name(worker_name, cfg)
        if host_worker_name not in all_host_root_worker:
            filter_worker_list.append(fault_source)
    if worker_name.startswith("LCNE"):
        host_worker_name = get_host_worker_name_by_lcne_worker_name(worker_name, cfg)
        if host_worker_name not in all_host_root_worker:
            filter_worker_list.append(fault_source)


def _remove_non_root_cause(code_dict: dict):
    """
    Remove non-root cause fault in some scenes
    :param code_dict: kg diagnostic result for all worker, the key is code
    :return: root cause list
    """
    # if result contains other fault, remove KG_DIAGNOSIS_NORMAL code. 事件优先级: TraceBack > NORMAL_OR_UNSUPPORTED
    if KG_DIAGNOSIS_NORMAL in code_dict and len(code_dict) > 1:
        code_dict.pop(KG_DIAGNOSIS_NORMAL, None)

    code_keys_without_traceback = set()
    for code_key in code_dict.keys():
        if not code_key.startswith((PRE_TRACEBACK_FAULT, PRE_SWITCH_FAULT)):
            code_keys_without_traceback.add(code_key)

    # if result contains non-HCCL fault, remove HCCL_FAULT_LIST code.
    if code_keys_without_traceback - set(HCCL_FAULT_LIST):
        [code_dict.pop(code, None) for code in HCCL_FAULT_LIST]

    # if result contains LINK_DOWN_FAULT, remove DEVICE_CQE_FAULT code.
    if LINK_DOWN_FAULT in code_dict or LINK_STATUS_CHANGE in code_dict:
        code_dict.pop(DEVICE_CQE_FAULT, None)


def _root_cause_sort_func(single_cause_dict):
    """
    Generate the sort rules for root cause list
    :param single_cause_dict: a cause dict
    :return: sort key tuple
    """
    component = SORT_RULES.get(single_cause_dict.get("component"), LOWEST_PRIORITY_NUM)
    code = single_cause_dict.get('code', "UNKNOWN")
    occur_time = _get_root_cause_first_occur_time(single_cause_dict)
    return occur_time, component, code


def _root_device_sort_func_in_single_cause(single_cause_dict):
    """
    Generate the sort rules func for root device list in each cause
    :param single_cause_dict: a cause dict
    :return: sort_func
    """
    event_attr = single_cause_dict.get("event_attr", {})  # attr format: {"dev_name": [{occur_time: xxx, ...},...],...}

    def sort_func(fault_device_name):
        """
        Generate the sort rules
        :param fault_device_name:
        :return: sort key
        """
        return event_attr.get(fault_device_name, [{}])[0].get("occur_time", "9999-12-31-23:59:59")

    return sort_func


def _get_root_cause_first_occur_time(single_cause_dict):
    """
    Get the first occur time of one cause
    :param single_cause_dict: a cause dict
    :return: first occur time
    """
    event_attr = single_cause_dict.get("event_attr", {})  # attr format: {"dev_name": [{occur_time: xxx, ...},...],...}
    event_time = ""
    for device_attr_list in event_attr.values():
        for device_attr in device_attr_list:
            if "occur_time" not in device_attr:
                continue
            attr_time = device_attr.get("occur_time")
            if not event_time or event_time > attr_time:
                event_time = attr_time
    return event_time


def _merge_version_info(version_info: dict):
    """
    Merge version list into one dict
    :param version_info:
        {'worker-0': {'cann_version': '7.0.0',  # different version
              'driver_version': '23.0.6',  # different version
              'firm_version': '7.1.0.11.220',
              'mindspore_version': '2.3.0',
              'nnae_version': '8.0.RC3',
              'pytorch_version': '1.11.0',
              'torch_npu_version': '2.1.0.post8.dev20241009'},
         'worker-1': {'cann_version': '7.0.T10',  # different version
              'driver_version': '23.0.7',  # different version
              'firm_version': '7.1.0.11.220',
              'mindspore_version': '2.3.0',
              'nnae_version': '8.0.RC3',
              'pytorch_version': '1.11.0',
              'torch_npu_version': '2.1.0.post8.dev20241009'}}
    :return:
        {'driver_version': '23.0.6, 23.0.7',  # merge result
         'cann_version': '7.0.0, 7.0.T10 (Non-commercial version)',  # merge result
         'firm_version': '7.1.0.11.220',
         'nnae_version': '8.0.RC3',
         'pytorch_version': '1.11.0',
         'torch_npu_version': '2.1.0.post8.dev20241009',
         'mindspore_version': '2.3.0'}
    """
    merged_version_info = dict()
    for _version_info in version_info.values():
        for version_name, version in _version_info.items():
            merged_version_info[version_name] = merged_version_info.get(version_name, set())
            # add non-commercial mark
            version += " (Non-commercial version)" if "T" in version else ""
            merged_version_info[version_name].add(version)

    # Only keep first 4 versions
    for version_name, versions in merged_version_info.items():
        versions = sorted(list(versions))
        merged_version_info[version_name] = ', '.join(versions[:4])
        if len(versions) > 4:
            merged_version_info[version_name] += '...'

    return merged_version_info
