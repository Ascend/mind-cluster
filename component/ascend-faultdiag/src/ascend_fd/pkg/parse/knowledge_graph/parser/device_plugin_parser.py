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
import os
import re
import logging
from typing import Union, List

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.parse_info import KGParseFilePath
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.utils.regular_table import DEVICEPLUGIN_SOURCE

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class DevicePluginParser(FileParser):
    SOURCE_FILE = DEVICEPLUGIN_SOURCE
    TARGET_FILE_PATTERNS = "device_plugin_path"
    FILE_TIME_REGEX = re.compile(r"devicePlugin-(\d{4}-\d{2}-\d{2})T(\d{2}-\d{2}-\d{2}.\d{3}).log")
    LOG_TIME_REGEX = re.compile(r"(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}.\d{6})")
    ASSERTION_KEYWORD_PATTERN = re.compile(r"Assertion:0x(\d)")
    DEFAULT_DEVICE_PLUGIN_FILENAME = "devicePlugin.log"

    def __init__(self, params):
        """
        Device plugin log parser
        """
        super().__init__(params)

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse log file
        :param parse_ctx: file paths
        :param task_id: unique task id
        :return: parse descriptor result
        """
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        file_source_list = self._filter_dp_list(parse_ctx.parse_file_path)
        results = self._parse_files(file_source_list)
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return results, {}

    def _filter_dp_list(self, parse_filepath: KGParseFilePath):
        """
        Filter device plugin file list
        :param parse_filepath: file paths dict
        :return a list of dp log path
        """
        dp_list = self.find_log(parse_filepath)
        if len(dp_list) < 2:
            return dp_list
        dp_list.sort(key=lambda x: self._get_filename(x))
        if not self.start_time or not self.end_time:
            return dp_list[-2:]
        if self.start_time > self.end_time:
            return dp_list[-2:]
        filtered_dp_list = list()
        for file_source in dp_list:
            filename = self._get_filename(file_source)
            if self.is_sdk_input and not filename:
                filtered_dp_list.append(file_source)
                continue
            if filename == self.DEFAULT_DEVICE_PLUGIN_FILENAME:
                filtered_dp_list.append(file_source)
                continue
            date_and_time = self.FILE_TIME_REGEX.findall(filename)
            if not date_and_time:
                continue
            # transfer YYYY-MM-DD hh-mm-ss.*** to YYYY-MM-DD hh:mm:ss.***
            date_and_time = date_and_time[0][0] + " " + date_and_time[0][1].replace("-", ":")
            if date_and_time < self.start_time or date_and_time < self.resuming_training_time:
                continue
            filtered_dp_list.append(file_source)
            if date_and_time > self.end_time:
                break
        return filtered_dp_list[-2:]

    def _parse_files(self, file_source_list: List[Union[str, LogInfoSaver]]):
        """
        Parse single device plugin file
        :param file_source_list: device plugin file source list
        :return: fault event list of the file
        """
        feasible_events = dict()
        for file_source in file_source_list:
            for dp_log in self._yield_log(file_source):
                event_dict = self.parse_single_line(dp_log)
                if event_dict:
                    event_dict.update({"source_file": self._get_source_file(file_source)})
                    self._handle_dp_events(event_dict, feasible_events, dp_log, file_source)
        event_storage = EventStorage()
        for event_dict in feasible_events.values():
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()

    def _handle_dp_events(self, event_dict: dict, feasible_events: dict, dp_log: str,
                          file_source: Union[str, LogInfoSaver]):
        """
        Handle dp events based on its fault type, in which recovery applied in switch fault case
        :param event_dict: matched fault dict
        :param feasible_events: storage for dp events
        :param dp_log: a single line of dp log
        """
        occur_time, assertion_flag = self._filter_dp_info(dp_log)
        if not occur_time or occur_time < self.resuming_training_time:
            return
        if self.start_time and occur_time < self.start_time:
            return
        if self.end_time and occur_time > self.end_time:
            return
        self.supplement_common_info(event_dict, file_source, occur_time)
        event_code = event_dict.get("event_code", "unknown_code")
        # in switch fault case, Assertion flag 1 signals fault occurrence for switch, otherwise 0 means fault recovery
        if not assertion_flag or assertion_flag == "1":
            # if not in switch fault case or flag 1 in switch fault case, record event occurrence
            feasible_events[event_code] = event_dict
            return
        if assertion_flag == "0":
            # if flag 0 in switch fault case, recover the fault
            feasible_events.pop(event_code, None)

    def _filter_dp_info(self, dp_log):
        """
        Filter device plugin fault occur time and assertion flag
        :param dp_log: single line of dp log
        :return: time and assertion flag
        """
        ret = self.LOG_TIME_REGEX.findall(dp_log)
        line_time = ret[0].replace("/", "-") if ret else ""
        ret = self.ASSERTION_KEYWORD_PATTERN.findall(dp_log)
        assertion_flag = ret[0] if ret else ""
        return line_time, assertion_flag
