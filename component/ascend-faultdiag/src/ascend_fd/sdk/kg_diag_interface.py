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
from ascend_fd.model.node_info import FaultFilterTime
from ascend_fd.utils.status import ParamError
from ascend_fd.utils.tool import MAX_PARAM_LEN, Field, SchemaValidator, validate_list_length, MultiProcessJob, \
    sort_results_by_id, validate_type, init_sdk_task
from ascend_fd.pkg.parse.knowledge_graph.utils.data_descriptor import DataDescriptor
from ascend_fd.pkg.diag.knowledge_graph.kg_diag_job import pre_analyze_job, hand_all_root_cause
from ascend_fd.utils.regular_table import MIN_TIME, MAX_TIME
from ascend_fd.wrapper.json_wrapper import JsonWrapper

MAX_SERVER_NUM_LIMIT = 4096
SDK_STR_PARAM_LIMIT = 1000


def diag_knowledge_graph(input_log_list: list):
    """
    Diagnose knowledge graph through a certain data structure, which contains a dict-format kg-analyzer
    :param input_log_list: the input from user
    :return: a list of results and a list of accumulated error
    """
    err_msg_list = []
    results = []
    try:
        filtered_input_list, input_validation_err = filter_input(input_log_list)
    except ParamError as err:
        err_msg_list.append(f"All validation failed, the reason is: {err}")
        return results, err_msg_list
    err_msg_list.extend(input_validation_err)
    task_id = init_sdk_task()
    multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH_PARSE_INTERFACE", pool_size=len(filtered_input_list),
                                       task_id=task_id, failed_raise=False)
    for idx, param in enumerate(filtered_input_list):
        server, kg_analyzer = param
        multiprocess_job.add_security_job(f"KG_DIAG_SERVER_{server}_ID_{idx}", diagnose_server, server,
                                          kg_analyzer, task_id, idx)
    multiprocess_results, _ = multiprocess_job.join_and_get_results()
    results = sort_results_by_id(multiprocess_results)
    return results, err_msg_list


def transform_kg_parser_to_kg_analyzer(fault_list):
    """
    Try to build chains through fault, transfer distributed fault into to an integrated format
    :param fault_list: a list of fault info, namely kg-parser
    :return: an integrated format, namely kg-analyzer
    """
    data_descriptor = DataDescriptor()
    data_descriptor.update_events(fault_list)
    data_descriptor.deal_event_data()
    single_worker_result = data_descriptor.get_single_worker_fault_analysis()
    return single_worker_result


def diagnose_server(server: str, kg_analyzer: dict, task_id: str, idx: int):
    """
    Diagnose a single server, return a formatted results
    """
    device_list = []
    for device in kg_analyzer.get("response", dict()).keys():
        device_list.append(device)
    pre_results, pre_failed_details = pre_analyze_job(server, device_list, kg_analyzer,
                                                      FaultFilterTime(MIN_TIME, MAX_TIME))
    kg_result = hand_all_root_cause(pre_results, pre_failed_details)
    result_formatter = JsonWrapper(result=kg_result, failed_details=dict(), performance_flag=False,
                                   task_id=task_id, single_diag_flag=False)
    return result_formatter.export_kg_sdk_results(), idx


def filter_input(input_log_list: list):
    """
    Filter input log list
    :param input_log_list: input log list
    :return: a filtered input list and a err msg list
    """
    validate_type(input_log_list, list, "input_log_list")
    validate_list_length(input_log_list, MAX_SERVER_NUM_LIMIT, "the input list")
    filtered_list = []
    err_msg_list = []
    for idx, log_info in enumerate(input_log_list):
        try:
            valid_params = process_input(log_info, idx)
        except ParamError as err:
            err_msg_list.append(f"Validation for the input list[{idx}] failed, the reason is: {err}]")
            continue
        filtered_list.append(valid_params)
    if not filtered_list:
        err_msg_list.append("All input invalid or empty input.")
    return filtered_list, err_msg_list


def process_input(input_log: dict, idx: int):
    """
    Process the atomic info element
    """
    source = input_log.get("source", "")
    validate_type(source, str, f"input_log_list[{idx}].source")
    if source.lower() == "ccae":
        input_schema = {
            "server": Field(type=str),
            "source": Field(type=str, allow_empty=True),
            "fault": Field(
                type=list,
                sub_schema={
                    "event_code": Field(type=str, size_limit=MAX_PARAM_LEN),
                    "key_info": Field(type=str, allow_empty=True),
                    "source_file": Field(type=str, allow_empty=True, size_limit=SDK_STR_PARAM_LIMIT),
                    "source_device": Field(type=str, size_limit=MAX_PARAM_LEN),
                    "occur_time": Field(type=str, size_limit=MAX_PARAM_LEN)
                }
            )
        }
        validator = SchemaValidator(input_schema)
        validator.validate(input_log, root=f"input_log_list[{idx}]")
        kg_analyzer = transform_kg_parser_to_kg_analyzer(input_log.get("fault"))
    else:
        input_schema = {
            "server": Field(type=str),
            "source": Field(type=str, allow_empty=True, mandatory=False),
            "fault": Field(type=list, size_limit=1, sub_element_type=dict)
        }
        validator = SchemaValidator(input_schema)
        validator.validate(input_log, root=f"input_log_list[{idx}]")
        kg_analyzer = next(iter(input_log.get("fault", [])), {})
    server = input_log.get("server", "")
    return server, kg_analyzer


def extract_device_info(kg_analyzer: dict, server: str):
    """
    Extract device info from kg-analyzer
    """
    device_info = dict()
    response = kg_analyzer.get("response", dict())
    for device in response.keys():
        device_info.setdefault(server, []).append(device)
    return device_info
