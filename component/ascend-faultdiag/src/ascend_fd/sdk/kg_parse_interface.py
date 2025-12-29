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
import re

from ascend_fd.pkg.customize.custom_entity.custom_update import update_entity
from ascend_fd.utils.status import ParamError, FileNotExistError, BaseError
from ascend_fd.utils.regular_table import SAVER_TO_SOURCE_FILE_MAP
from ascend_fd.pkg.parse.knowledge_graph.kg_parse_job import get_parse_ctx
from ascend_fd.pkg.parse.knowledge_graph.utils.package_parser import PackageParser
from ascend_fd.utils.tool import MAX_PARAM_LEN, validate_list_length, Field, SchemaValidator, init_sdk_task, \
    validate_type
from ascend_fd.pkg.parse.parser_saver import SaverFactory, LogInfoSaver, CustomLogSaver
from ascend_fd.model.cfg import ParseCFG
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import ParserFactory
from ascend_fd.pkg.parse.knowledge_graph.parser.custom_log_parser import CustomLogParser

MAX_SERVER_NUM_LIMIT = 4096
MAX_ITEM_NUM_LIMIT = 10000
MAX_DEVICE_NUM = 16


def parse_knowledge_graph(input_log_list: list, custom_entity: dict = None):
    """
    Parse knowledge graph fault event
    :param input_log_list: the input from user
    :param custom_entity: a temporally used entity only for the current task
    :return: a list of results and a list of accumulated error
    """
    err_msg_list = []
    results = []
    task_id = init_sdk_task()
    try:
        param_list, parse_conf, err_list = filter_input(input_log_list, custom_entity)
    except (ParamError, FileNotExistError) as e:
        err_msg_list.append(f"Input validation failed, the reason is: [{e}]")
        return results, err_msg_list
    err_msg_list.extend(err_list)
    results, err_msg = parse_data(param_list, parse_conf, task_id)
    err_msg_list.extend(err_msg)
    return results, err_msg_list


def parse_data(params_list: list, parser_conf: dict, task_id: str):
    """
    Parse data with multiprocessing, return a sorted results list
    """
    results = []
    err_msg = []
    for param in params_list:
        server, items = param
        kg_parser = KnowledgeGraphParser(server, items, parser_conf)
        try:
            results.append(kg_parser._parse_through_log_saver(task_id))
        except BaseError as err:
            err_msg.append(f"Failed to parse knowledge graph for server {server}, reason is: {err}")
            continue
    return results, err_msg


def filter_input(input_log_list: list, custom_entity: dict):
    """
    Filter the input logs list
    :param input_log_list: the input to be filtered
    :param custom_entity: a temporally used entity only for the current task
    :return: a filtered list
    """
    validate_type(input_log_list, list, "input_log_list")
    custom_entity = custom_entity or {}
    validate_type(custom_entity, dict, "custom_entity")
    parse_conf = {}
    err_list = []
    update_entity(sdk_entity=custom_entity, output_dict=parse_conf)
    if custom_entity:
        diff = list(set(custom_entity.keys()) - (set(parse_conf.get("knowledge-repository", {}).keys())))
        if diff:
            err_list.append(f"Some entities failed to update, please check the input: {diff}")
    validate_list_length(input_log_list, limit=MAX_SERVER_NUM_LIMIT, label="the input list")
    filtered_list = []
    source_file_set = set()
    for idx, log_info in enumerate(input_log_list):
        # Field 'item_type' (SDK input level) → Field 'source_file' (component logic level)
        server, items, single_source_file_set = input_params_validation(log_info, idx)
        filtered_list.append((server, items))
        source_file_set = source_file_set.union(single_source_file_set)
    unsupported_source_file = get_unsupported_source_file(source_file_set, parse_conf)
    if unsupported_source_file:
        err_list.append("The following item types are unsupported: {}, please check the input."
                        .format(", ".join(unsupported_source_file)))
    return filtered_list, parse_conf, err_list


def get_unsupported_source_file(source_file_set: set, parser_conf: dict):
    supported_custom_source_file = {
        entity.get("source_file", "")
        for entity in parser_conf.get("knowledge-repository", {}).values()
    }
    unsupported_source_file = set()
    for source_file in source_file_set:
        if source_file not in supported_custom_source_file \
                and not any(source_file in source_file_list for source_file_list in SAVER_TO_SOURCE_FILE_MAP.values()):
            unsupported_source_file.add(source_file)
    return unsupported_source_file


def custom_time_validator(line, label: str):
    """
    Validate modification time, uncompleted 6 digits of fractional seconds parts will be padded if possible
    """
    pure_line = line.strip()
    time_regex = re.compile(r"(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})(?:\.(\d{1,6}))?")
    match = time_regex.match(pure_line)
    if not match:
        raise ParamError(f"The format of {label} is invalid, the correct format is: YYYY-MM-DD HH:MM:SS[.ssssss] "
                         f"(the fractional seconds part is optional and will be padded to 6 digits if provided)")
    base_time, micro_seconds = match.groups()
    if micro_seconds is None:
        return base_time + ".000000"
    expected_digits = 6
    micro_seconds = micro_seconds.ljust(expected_digits, "0")
    return base_time + "." + micro_seconds


def input_params_validation(log_info: dict, idx: int):
    """
    Validate atomic data structure through validator
    :param log_info: atomic data structure to be validated
    :param idx: index of the current element in the input log list
    :return: extracted data for those valid ones
    """
    item_type_set = set()

    def count_item_type(value):
        item_type_set.add(value)

    kg_parse_input_schema = {
        "log_domain": Field(
            type=dict,
            sub_schema={
                "server": Field(type=str, size_limit=MAX_PARAM_LEN)
            }
        ),
        "log_items": Field(
            type=list,
            size_limit=MAX_ITEM_NUM_LIMIT,
            sub_schema={
                "item_type": Field(type=str, size_limit=MAX_PARAM_LEN, statistic_callback=count_item_type),
                "path": Field(type=str, mandatory=False, allow_empty=True),
                "device_id": Field(type=int, mandatory=False, choices=range(MAX_DEVICE_NUM)),
                "log_lines": Field(type=list, sub_element_type=str),
                "modification_time": Field(type=str, mandatory=False, size_limit=MAX_PARAM_LEN,
                                           custom_validator=custom_time_validator),
                # only used for MindIE fault currently
                "component": Field(type=str, mandatory=False, choices=["Controller", "Coordinator"])
            }
        )
    }
    validator = SchemaValidator(kg_parse_input_schema)
    validator.validate(log_info, root=f"input_log_list[{idx}]")
    items = log_info.get("log_items", [])
    domain = log_info.get("log_domain", dict())
    server = domain.get("server", "")
    return server, items, item_type_set


class KnowledgeGraphParser:
    def __init__(self, server: str, log_items: list, parser_conf: dict):
        self.server = server
        self.log_items = log_items
        self.type_log_saver = dict()
        self.saver_list = list()
        self.source_file_category = set()
        self.parse_conf = parser_conf
        self._init_savers()

    def __str__(self):
        return self.server

    @property
    def saver_class_types(self):
        return [saver.__class__ for saver in self.saver_list]

    @staticmethod
    def _get_reverse_saver_map():
        """
        Construct a reverse map from source file to saver
        """
        source_file_to_saver_map = dict()
        for saver_name, source_file_list in SAVER_TO_SOURCE_FILE_MAP.items():
            for source_file in source_file_list:
                source_file_to_saver_map[source_file] = saver_name
        return source_file_to_saver_map

    @staticmethod
    def _get_saver_instances(used_saver_types) -> list:
        return SaverFactory.batch_create_savers(used_saver_types)

    def _config_saver_type(self):
        source_file_to_saver_map = self._get_reverse_saver_map()
        used_saver_types = set()
        for source_file in self.type_log_saver.keys():
            saver_type = source_file_to_saver_map.get(source_file, "")
            if saver_type:
                used_saver_types.add(saver_type)
            else:
                used_saver_types.add(CustomLogSaver.__name__)
        return used_saver_types

    def _parse_through_log_saver(self, task_id):
        """
        Arrange log saver and parse through them.
        """
        cfg = ParseCFG.sdk_config(task_id, self.saver_list)
        for saver in self.saver_list:
            if isinstance(saver, CustomLogSaver):
                self._add_log_to_custom_saver(saver)
            else:
                self._add_log_to_saver(saver)
        parse_ctx = get_parse_ctx(cfg)
        required_parsers = self._get_unique_required_parser()
        package_parser = PackageParser.init_sdk_package_parser(parse_ctx, required_parsers, self.parse_conf)
        package_parser.parse(task_id)
        final_data = package_parser.desc.get_single_worker_fault_analysis(self.parse_conf)
        return self._format_output(final_data)

    def _get_unique_required_parser(self):
        required_parsers = set()
        for source_file in self.type_log_saver.keys():
            parser_cls = ParserFactory.get_parser_class(source_file)
            if parser_cls:
                required_parsers.add(parser_cls)
            else:
                required_parsers.add(CustomLogParser)
        return required_parsers

    def _init_savers(self):
        self._config_log_item()
        used_saver_types = self._config_saver_type()
        self.saver_list = self._get_saver_instances(used_saver_types)

    def _format_output(self, final_data: dict):
        """
        Format the output
        :param final_data: a dict format kg-analyzer
        :return: a formatted result
        """
        formatted_result = {
            "server": self.server,
            "fault": [final_data]
        }
        return formatted_result

    def _add_log_to_saver(self, saver):
        """
        Add log to saver directly, instead of reading the actual file
        """
        source_file_list = SAVER_TO_SOURCE_FILE_MAP.get(saver.__class__.__name__, [])
        for source_file in source_file_list:
            item_list = self.type_log_saver.get(source_file, [])
            if not item_list:
                continue
            saver.update_log({source_file: item_list})

    def _add_log_to_custom_saver(self, custom_saver: CustomLogSaver):
        """
        Add log to saver directly, instead of reading the actual file
        """
        for source_file in self.type_log_saver.keys():
            if not any(source_file in source_file_list for source_file_list in SAVER_TO_SOURCE_FILE_MAP.values()):
                item_list = self.type_log_saver.get(source_file, [])
                if not item_list:
                    continue
                custom_saver.update_log({source_file: item_list})

    def _config_log_item(self):
        """
        Organize fault structure into a dict, which has keys of source file and values of LogInfoItem
        """
        for log_item in self.log_items:
            source_file = log_item.get("item_type", "")
            item_info = LogInfoSaver(
                source_file=source_file,
                path=log_item.get("path", ""),
                device_id=log_item.get("device_id", -1),
                log_lines=log_item.get("log_lines", []),
                modification_time=log_item.get("modification_time", ""),
                component=log_item.get("component", "")
            )
            self.type_log_saver.setdefault(source_file, []).append(item_info)
