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
import json

from ascend_fd.model.mindie_info import MindIEParseResult
from ascend_fd.pkg.parse.knowledge_graph.parser.mindie_parser import MindieParser
from ascend_fd.utils.status import ParamError
from ascend_fd.utils.tool import MultiProcessJob, validate_list_length, MAX_PARAM_LEN, Field, SchemaValidator, \
    sort_results_by_id, init_sdk_task, validate_type
from ascend_fd.pkg.parse.root_cluster.parser import PidFileParser

MAX_SERVER_NUM_LIMIT = 256
DEVICE_NUM_LIMIT = 100
NEGATIVE_ONE = -1
MAX_DEVICE_NUM = 16


def parse_root_cluster(input_log_list: list):
    """
    Parse root cluster, analyze node base info
    :param input_log_list: the input from user
    :return: a list of results and a list of accumulated error
    """
    err_msg_list = []
    results = []
    try:
        param_list = filter_input(input_log_list)
    except ParamError as err:
        err_msg_list.append(f"Input validation failed, the reason is: [{err}]")
        return results, err_msg_list
    results, failed_details = parse_data(param_list)
    if failed_details:
        first_key, first_value = next(iter(failed_details.items()))
        err_msg_list.append(f"Some Root Cluster parse job failed. The first failed job: {first_key}, "
                            f"error reason is: {first_value}")
    return results, err_msg_list


def filter_input(input_log_list: list):
    """
    Validate the input logs list
    :param input_log_list: the input to be filtered
    :return: a filtered input
    """
    validate_type(input_log_list, list, "input_log_list")
    validate_list_length(input_log_list, limit=MAX_SERVER_NUM_LIMIT, label="the input list")
    filtered_list = []
    unique_server_instance = set()
    for idx, log_info in enumerate(input_log_list):
        server, instance, items = input_params_validation(log_info, idx)
        unique_key = (server, instance)
        if unique_key in unique_server_instance:
            continue
        unique_server_instance.add(unique_key)
        filtered_list.append((server, instance, items))
    return filtered_list


def input_params_validation(log_info: dict, idx: int):
    """
    Validate atomic data structure through validater
    :param log_info: atomic data structure to be validated
    :param idx: index of the current element in the input log list
    :return: extracted data for those valid ones
    """
    rc_parse_input_schema = {
        "log_domain": Field(
            type=dict,
            sub_schema={
                "server": Field(type=str, size_limit=MAX_PARAM_LEN),
                "instance_id": Field(type=str, size_limit=MAX_PARAM_LEN)
            }
        ),
        "log_items": Field(
            type=list,
            size_limit=DEVICE_NUM_LIMIT,
            sub_schema={
                "item_type": Field(
                    type=str, choices=["plog", "mindie", "MINDIE", "PLOG", "MindIE"], size_limit=MAX_PARAM_LEN
                ),
                "pid": Field(type=int),
                "device_id": Field(type=int, mandatory=False, allow_empty=True, choices=range(MAX_DEVICE_NUM)),
                "rank_id": Field(type=int, mandatory=False, allow_empty=True),
                "log_lines": Field(type=list, allow_empty=True, sub_element_type=str)
            }
        )
    }
    validator = SchemaValidator(rc_parse_input_schema)
    validator.validate(log_info, root=f"input_log_list[{idx}]")
    items = log_info.get("log_items", [])
    domain = log_info.get("log_domain", dict())
    server = domain.get("server", "")
    instance_id = domain.get("instance_id", "")
    return server, instance_id, items


def parse_data(params_list: list):
    """
    Parse data with multiprocessing, return a sorted results list
    """
    mindie_parser = MindIESDKParser()
    # these two lists are going to keep in the same size, which corresponds with each other one by one
    pid_data_list, rc_parser_list, instance_pid_map = data_initialization(params_list, mindie_parser)

    # pid parser list format: [{pid: PidFileParser}]
    pid_parser_list, failed_details = multiprocess_log_lines(pid_data_list, rc_parser_list)

    results = format_server_result(instance_pid_map, pid_data_list, pid_parser_list, rc_parser_list)

    mindie_parse_data = mindie_parser.get_parsed_result()
    if mindie_parse_data:
        results.append(mindie_parse_data)
    return results, failed_details


def format_server_result(instance_pid_map: dict, pid_data_list: list, pid_parser_list: list, rc_parser_list: list):
    """
    Format and return server results, three lists hold the same and order
    """
    results = []
    rank_id_map = dict()
    for idx, rc_parser in enumerate(rc_parser_list):
        server_result = dict()
        instance = instance_pid_map.get(rc_parser.instance_id, None)
        unique_instance_pid = getattr(instance, "pid_input_with_dev_id", 0)
        has_device_id_input = unique_instance_pid > 0
        unique_pid_count = unique_instance_pid or getattr(instance, "all_pid_count", 0)
        for pid, parser in pid_parser_list[idx].items():
            log_item = pid_data_list[idx].get(pid, dict())
            # if all device id absence, fabricate a rank map for every log_item,
            # otherwise for those log_item with device_id inputs
            if log_item.get("device_id", NEGATIVE_ONE) == NEGATIVE_ONE and has_device_id_input:
                server_result.update(rc_parser.get_result_for_pid(pid, parser))
                continue
            original_rank_id = log_item.get("rank_id", NEGATIVE_ONE)
            allocated_rank_id = rc_parser.allocate_rank_id(unique_pid_count, rank_id_map, original_rank_id)
            rc_parser.fabricate_rank_map(allocated_rank_id, unique_pid_count, parser)
            server_result.update(rc_parser.get_result_for_pid(pid, parser))
        if server_result:
            results.append(server_result)
    return results


def multiprocess_log_lines(pid_data_list: list, rc_parser_list: list):
    """
    Multiprocess log lines and return
    :param pid_data_list: a pid-formatted dict
    :param rc_parser_list: list or RootClusterParser
    :return: a list of pid parser for each server and failed details
    """
    task_id = init_sdk_task()
    multiprocess_job = MultiProcessJob("ROOT_CLUSTER_PARSE_INTERFACE", pool_size=len(rc_parser_list), task_id=task_id,
                                       failed_raise=False)
    # parse server
    for idx, rc_parser in enumerate(rc_parser_list):
        multiprocess_job.add_security_job(f"RC_PARSE_PROCESS_RESULT_SERVER_{rc_parser}_ID_{idx}",
                                          rc_parser.parse_server, pid_data_list[idx])
    results_with_id, failed_details = multiprocess_job.join_and_get_results()
    # this list still has the same order with rc_parser_list and pid_data_list
    pid_parser_list = sort_results_by_id(results_with_id)
    return pid_parser_list, failed_details


def data_initialization(params_list: list, mindie_parser):
    """
    Initialize pid_data_list and rc_parser_list
    pid_data_list: a reconstructed dict formatted by pid
    format example: {pid: {"item_type": "xxx", "rank_id": 0, "device_id": 0, "log_lines": "xxx"}}
    rc_parser_list: a list of RootClusterParser
    """
    pid_data_list = []
    rc_parser_list = []
    instance_pid_map = dict()
    for idx, param in enumerate(params_list):
        server, instance_id, items = param
        root_cluster_parser = RootClusterParser(server, instance_id, items)
        # idx assignment is used for further sorting operation
        root_cluster_parser.idx = idx
        rc_parser_list.append(root_cluster_parser)
        pid_data, pid_item_with_dev_id = root_cluster_parser.get_pid_structured_data()
        instance_pid_map.setdefault(instance_id, Instance(instance_id)).count_pid(len(pid_data), pid_item_with_dev_id)
        pid_data_list.append(pid_data)
        # 过滤mindie日志
        mindie_parser.get_mindie_log_lines(items)
    return pid_data_list, rc_parser_list, instance_pid_map


class Instance:
    def __init__(self, instance_id):
        self.instance_id = instance_id
        self.all_pid_count = 0
        self.pid_input_with_dev_id = 0

    def count_pid(self, all_pid, pid_with_dev_id):
        self.all_pid_count += all_pid
        self.pid_input_with_dev_id += pid_with_dev_id


class MindIESDKParser:
    def __init__(self):
        self.mindie_log_lines = []
        self.mindie_parser = MindieParser({})

    def get_mindie_log_lines(self, log_items):
        """
        Get mindie log line
        items: items
        """
        for log_item in log_items:
            log_lines = log_item.get("log_lines", [])
            if not log_lines:
                continue
            item_type = log_item.get("item_type", "")
            if item_type.lower() != "mindie":
                continue
            self.mindie_log_lines.extend(log_lines)

    def get_parsed_result(self):
        """
        Get parsed result
        """
        if not self.mindie_log_lines:
            return {}
        for log_line in self.mindie_log_lines:
            self.mindie_parser.filter_error_info(log_line)
        result = MindIEParseResult()
        result.reconstruct_result(self.mindie_parser.mindie_parse_info)
        return result.to_dict()


class RootClusterParser:
    def __init__(self, server: str, instance_id: str, log_items: list):
        self.server = server
        self.instance_id = instance_id
        self.log_items = log_items
        self.rank_id_set = set()
        self.idx = 0

    def __str__(self):
        return self.server

    @staticmethod
    def get_result_for_pid(pid: int, pid_file_parser):
        this_pid_results = pid_file_parser.get_result()
        if not this_pid_results:
            return {}
        return {str(pid): json.loads(json.dumps(this_pid_results.__dict__, default=lambda obj: obj.__dict__))}

    def parse_server(self, pid_structured_data: dict):
        """
        Parse server, return parsed pid parser map
        """
        pid_parser_map = dict()
        for pid, data in pid_structured_data.items():
            pid_file_parser = PidFileParser(str(pid), device_ip_map={})
            pid_file_parser.parse_log(data.get("log_lines"))
            device_id = data.get("device_id", NEGATIVE_ONE)
            pid_file_parser.base_info_parser.supplement_base_info(device_id, self.server)
            pid_parser_map.setdefault(pid, pid_file_parser)
        return pid_parser_map, self.idx

    def fabricate_rank_map(self, rank_id, unique_pid_count, pid_file_parser):
        instance_rank_map = {
            self.instance_id: {
                "rank_num": unique_pid_count, "rank_id": str(rank_id)
            }
        }
        pid_file_parser.base_info_parser.supplement_rank_info(self.instance_id, instance_rank_map, self.server)

    def get_pid_structured_data(self):
        """
        Group log line by pid, return a pid-structured dict
        """
        pid_structured_data = dict()
        pid_item_with_device_id = 0
        for log_item in self.log_items:
            log_lines = log_item.get("log_lines", [])
            if not log_lines:
                continue
            item_type = log_item.get("item_type", "")
            if item_type.lower() != "plog":
                continue
            pid = log_item.get("pid")
            value = {k: v for k, v in log_item.items() if k != "pid"}
            if pid not in pid_structured_data:
                pid_structured_data[pid] = value
            else:
                for key, value in value.items():
                    if key == "log_lines":
                        pid_structured_data[pid]["log_lines"].extend(log_lines)
                        continue
                    pid_structured_data[pid][key] = value
            # record the num of input that has a device id
            if log_item.get("device_id", NEGATIVE_ONE) != NEGATIVE_ONE:
                pid_item_with_device_id += 1
        return pid_structured_data, pid_item_with_device_id

    def allocate_rank_id(self, unique_pid_count: int, instance_rank_id_map: dict, original_rank_id: int):
        """
        Return the original if it is valid, otherwise allocate a valid one
        Be compatible with certain situations which have no rank id input

        :param unique_pid_count: The total of all unique rank ids
        :param instance_rank_id_map: a rank id map with respect to instance id
        :param original_rank_id: the original rank_id
        :return: Original rank id or allocated rank id
        """
        if original_rank_id != NEGATIVE_ONE and 0 <= original_rank_id < unique_pid_count:
            instance_rank_id_map.setdefault(self.instance_id, set()).add(original_rank_id)
            return original_rank_id
        for pre_assigned_id in range(unique_pid_count):
            if pre_assigned_id not in instance_rank_id_map.get(self.instance_id, set()):
                instance_rank_id_map.setdefault(self.instance_id, set()).add(pre_assigned_id)
                return pre_assigned_id
        default_rank_id = 0
        return default_rank_id
