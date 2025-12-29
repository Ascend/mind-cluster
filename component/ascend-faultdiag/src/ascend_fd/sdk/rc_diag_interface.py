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
from ascend_fd.model.mindie_info import MindIEParseResult
from ascend_fd.pkg.diag.root_cluster.mindie_diag_job import MindIEDiagWorker
from ascend_fd.utils.status import ParamError
from ascend_fd.utils.tool import validate_type, init_sdk_task
from ascend_fd.pkg.parse.parser_saver import ParsedDataSaver
from ascend_fd.model.cfg import DiagCFG
from ascend_fd.pkg.diag.root_cluster.rc_diag_job import RCDiagWorker
from ascend_fd.utils.status import BaseError
from ascend_fd.wrapper.json_wrapper import JsonWrapper


def diag_root_cluster(input_log_list: list):
    """
    Diagnose root cluster through a list of rc-parser
    :param input_log_list: the input from user
    :return: a list of results and a list of accumulated error
    """
    results = dict()
    err_msg_list = []
    try:
        rc_parser_dict, validation_err_msgs, mindie_parse_result = filter_and_reconstruct_input(input_log_list)
    except ParamError as err:
        err_msg_list.append(f"All validation failed, the reason is: {err}")
        return results, err_msg_list
    if validation_err_msgs:
        err_msg_list.append(f"Validation failed for some rc-parser inputs, the reasons are [{validation_err_msgs}]")
    if len(validation_err_msgs) == len(rc_parser_dict):
        err_msg_list.append("Root Cluster diagnosis job failed since all input validation failed.")
    diag_cfg = DiagCFG("", "", "", ParsedDataSaver("", {}))
    rc_diagnosis = RCDiagWorker(diag_cfg, sdk_input=rc_parser_dict)
    rc_diagnosis.cfg.parsed_saver.mindie_parse_result = mindie_parse_result
    task_id = init_sdk_task()
    try:
        MindIEDiagWorker(rc_diagnosis.cfg).start_job()
        diag_result = rc_diagnosis.start_job()
    except (BaseError, TypeError) as err:
        err_msg = f"Root Cluster diagnosis job failed. The reason is: {err}"
        err_msg_list.append(err_msg)
        return results, err_msg_list
    if not diag_result.detect_workers_devices:
        err_msg_list.append("The list of workers to be checked is empty. Please check the root cluster diag result.")
    result_formatter = JsonWrapper(result={"Rc": diag_result.to_dict()}, failed_details=dict(), performance_flag=False,
                                   task_id=task_id, single_diag_flag=False)
    results = result_formatter.export_rc_sdk_results()
    return results, err_msg_list


def filter_and_reconstruct_input(input_log_list: list):
    """
    Filter the input and assign server name to them if there is no name
    """

    validate_type(input_log_list, list, "input_log_list")
    if not input_log_list:
        raise ParamError("The input list is empty.")
    server_structured_data = dict()
    err_msg = []
    unknown_server_prefix = "unknown_server"
    unknown_server_suffix = 0
    mindie_parse_result = MindIEParseResult()
    for idx, rc_parser in enumerate(input_log_list):
        try:
            validate_type(rc_parser, dict, f"rc_parser_[{idx}]")
        except ParamError as err:
            err_msg.append(err)
            continue
        if rc_parser.get("mindie", False):
            mindie_parse_result.link_error_info_map = rc_parser.get("link_error_info_map", {})
            mindie_parse_result.pull_kv_error_map = rc_parser.get("pull_kv_error_map", {})
            continue
        server = get_server_id(rc_parser, unknown_server_prefix, unknown_server_suffix)
        if server.startswith(unknown_server_prefix):
            unknown_server_suffix += 1
        server_structured_data[server] = rc_parser
    return server_structured_data, err_msg, mindie_parse_result


def get_server_id(rc_parser: dict, server_prefix: str, server_suffix: int):
    """
    Try to extract a server name, if no result, assemble a server name.
    """
    for pid_parse_result in rc_parser.values():
        server = pid_parse_result.get("base", {}).get("server_name", "")
        if server and isinstance(server, str):
            return server
    return f"{server_prefix}_{server_suffix}"
