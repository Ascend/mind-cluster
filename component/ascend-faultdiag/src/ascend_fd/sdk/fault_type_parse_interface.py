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
from itertools import chain

from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage
from ascend_fd.pkg.customize.custom_entity.valid import source_check
from ascend_fd.utils.load_kg_config import ParseRegexMap
from ascend_fd.utils.tool import MultiProcessJob, PatternSingleOrMultiLineMatcher, validate_type, init_sdk_task
from ascend_fd.configuration.config import DEFAULT_USER_CONF
from ascend_fd.utils.load_kg_config import Schema, EntityAttribute
from ascend_fd.utils.status import ParamError

MAX_SERVER_NUM_LIMIT = 256
MAX_ITEM_NUM_LIMIT = 100


def input_params_validation(log_info: dict):
    """
    Validate input parameters
    :param log_info: single log info dict
    :return: fetched info
    """
    domain = log_info.get("log_domain", dict())
    items = log_info.get("log_items", [])
    if not domain or not items:
        raise ParamError("Invalid parameters, both 'log_domain' and 'log_items' should be "
                         "available and valid for a log element.")
    validate_type(domain, dict, "log_domain")
    validate_type(items, list, "log_items")
    items_size = len(items)
    if items_size > MAX_ITEM_NUM_LIMIT:
        raise ParamError("The size of server info 'log_items' exceeds the limit, which is {}, "
                         "whereas the current size is {}".format(MAX_ITEM_NUM_LIMIT, items_size))
    if "device" not in domain:
        raise ParamError("Insufficient parameters, 'device' should be a filed of 'log_domain'.")
    server, device = domain.get("server", ""), domain.get("device", [])
    if not server:
        raise ParamError("Invalid parameters for 'server', 'server' should be available and valid for 'log_domain'.")
    validate_type(server, str, "server")
    validate_type(device, list, "device")
    return server, device, items


def valid_input(input_log_list: list):
    """
    Validate the input logs list
    :param input_log_list: the input to be filtered
    :return:
    """
    validate_type(input_log_list, list, "input_log_list")
    input_list_size = len(input_log_list)
    if input_list_size > MAX_SERVER_NUM_LIMIT:
        raise ParamError("The size of the input list exceeds the limit, which is {}, "
                         "whereas the current size is {}.".format(MAX_SERVER_NUM_LIMIT, input_list_size))
    filtered_list = []
    for log_info in input_log_list:
        valid_params = input_params_validation(log_info)
        filtered_list.append(valid_params)
    return filtered_list


def parse_fault_type(input_log_list: list):
    """
    Parse the input stream, output a structured output stream
    :param input_log_list: input from the user
    :return: an analyzed and structured fault list
    """
    err_msg_list = []
    results = []
    try:
        filtered_input_list = valid_input(input_log_list)
    except ParamError as e:
        err_msg_list.append(f"Parse failed, the reason is: [{e}]")
        return results, err_msg_list
    task_id = init_sdk_task()
    multiprocess_job = MultiProcessJob("PARSE_INTERFACE", pool_size=len(filtered_input_list), task_id=task_id,
                                       daemon=False, failed_raise=False)
    for idx, log_info in enumerate(filtered_input_list):
        server, device, items = log_info
        log_parser = ParseDataPacker(server, device)
        multiprocess_job.add_security_job(f"PARSE_SERVER_{server}_ID_{idx}", log_parser.parse, items, task_id)
    results, failed_details = multiprocess_job.join_and_get_results()
    fault_map = dict()
    for fault in list(chain(*results.values())):
        saved_fault = fault_map.get(fault.error_type, None)
        if saved_fault is None:
            fault_map.update({fault.error_type: fault})
            continue
        saved_fault.update_domain_info(fault.server, fault.device)
    results = format_output(fault_map)
    if failed_details:
        err_msg_list.append(f"Some parsing works failed, the failure can be partly attributed to: "
                            f"[{next(iter(failed_details.values()))}]")
    return results, err_msg_list


def format_output(fault_map: dict):
    """
    Format the result and then output it
    :param fault_map: the fault record
    :return: a structured result
    """
    results = []
    for error_type, fault in fault_map.items():
        results.append({
            "error_type": error_type,
            "fault_domain": fault.attribute.class_,
            "attribute": fault.get_attribute(),
            "device_list": fault.get_device_list()
        })
    return results


class LogItemParser(FileParser):

    def __init__(self, item_type: str, log_lines: list):
        self.params = {"regex_conf": ParseRegexMap([DEFAULT_USER_CONF]).get_parse_regex()}
        self.SOURCE_FILE = item_type
        super().__init__(self.params)
        self.log_lines = log_lines
        if self.SOURCE_FILE == "TrainLog":
            self.pattern_matcher = PatternSingleOrMultiLineMatcher(log_lines=self.log_lines)

    def parse(self, file_dict: dict = None, task_id: str = ""):
        """
        Parse the single log item
        :return: the result event list
        """
        event_storage = EventStorage()
        for cur_line_idx, line in enumerate(self.log_lines):
            if self.SOURCE_FILE == "TrainLog":
                self.pattern_matcher.update_line_index(cur_line_idx)
            event_dict = self.parse_single_line(line)
            if not event_dict:
                continue
            # the occur_time is set to the line num as there is no resource to fetch it
            event_dict.update({"source_device": "Unknown", "occur_time": str(cur_line_idx)})
            event_storage.record_event(event_dict, with_occurrence=False)
        return event_storage.generate_event_list()


class ParseDataPacker:

    def __init__(self, server: str, device: list):
        self.server = server
        self.device = device
        self.data = []
        self.schema = Schema([DEFAULT_USER_CONF])

    @staticmethod
    def _filter_parameters(items: list):
        """
        Filter valid parameters
        :param items: items list
        :return: a dict of params
        """
        filtered_params_list = []
        log_lines_key = "log_lines"
        for item in items:
            if log_lines_key not in item:
                raise ParamError("Insufficient parameters, '{}' need to exist for a 'log_items' element."
                                 .format(log_lines_key))
            item_type = item.get("item_type", "")
            if not item_type:
                raise ParamError("Invalid parameters, both 'item_type' and 'log_lines' should be "
                                 "available and valid for a log element.")
            log_lines = item.get(log_lines_key, [])
            validate_type(log_lines, list, log_lines_key)
            validate_type(item_type, str, "item_type")
            if not source_check(item_type):
                raise ParamError("Invalid item_type: {}".format(item_type))
            filtered_params_list.append({"item_type": item_type, "log_lines": log_lines})
        return filtered_params_list

    def parse(self, items: list, task_id: str):
        """
        Parse the log items element item by item
        :param items: all items for a server
        :param task_id: task id for multiprocessing
        """
        valid_params = self._filter_parameters(items)
        max_pool_size = 5
        pool_size = min(max_pool_size, len(valid_params))
        multiprocess_job = MultiProcessJob("PARSE_INTERFACE", pool_size=pool_size, task_id=task_id)
        for idx, item in enumerate(items):
            item_parser = LogItemParser(**item)
            multiprocess_job.add_security_job(f"PARSE_LOG_ITEM{idx}", item_parser.parse)
        results, _ = multiprocess_job.join_and_get_results()
        self.data = list(chain(*results.values()))
        return self._format_results()

    def _format_results(self):
        """
        Return the fault results list
        :return: a list of fault
        """
        results = []
        for event in self.data:
            code = event.get("event_code", "")
            key_info = event.get("key_info", "")
            schema_entity = self.schema.get_schema_entity(code)
            if not schema_entity:
                continue
            results.append(Fault(code, self.server, self.device, key_info, schema_entity.attribute))
        return results


class Fault:
    def __init__(self, code: str, server: str, device: list, key_info: str, attribute: EntityAttribute):
        self.error_type = code
        self.server = server
        self.device = device
        self.key_info = key_info
        self.domain = dict()
        self.update_domain_info(server, device)
        self.attribute = attribute

    def update_domain_info(self, server: str, device: list):
        """
        Update the domain record, merge the difference and record the new domain
        :param server: input server ip
        :param device: input device list
        """
        if server not in self.domain:
            self.domain[server] = device
            return
        self.domain.update({server: list(set(self.domain.get(server) + device))})

    def get_attribute(self):
        """
        Return essential attributes of the fault
        """
        return {
            "key_info": self.key_info,
            "component": self.attribute.component,
            "module": self.attribute.module,
            "cause": self.attribute.cause_zh or self.attribute.cause_en,
            "description": self.attribute.description_zh or self.attribute.description_en,
            "suggestion": self.attribute.suggestion_zh or self.attribute.suggestion_en,
        }

    def get_device_list(self):
        """
        Return the formatted device list
        """
        device_list = []
        for server, device in self.domain.items():
            device_list.append({"server": server, "device": device})
        return device_list
