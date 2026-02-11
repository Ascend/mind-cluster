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
import time
from datetime import datetime
from itertools import chain

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage
from ascend_fd.pkg.parse.parser_saver import MatchedCustomInfo, LogInfoSaver
from ascend_fd.utils.regular_table import KG_MAX_TIME
from ascend_fd.utils.tool import MultiProcessJob

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class CustomLogParser(FileParser):
    SOURCE_FILE = ""
    TARGET_FILE_PATTERNS = "custom_log_list"

    def __init__(self, params):
        super().__init__(params)

    @staticmethod
    def _get_occur_time(line, log_time_format):
        """
        Get the time info.
        :param line: log line
        :param log_time_format: log time format
        :return: time info
        """
        log_time_regex = re.sub(r"%(Y|f|[mdHMS])", lambda match: {
            "Y": r"\d{4}",
            "f": r"\d{3,6}",
            "m": r"\d{2}",
            "d": r"\d{2}",
            "H": r"\d{2}",
            "M": r"\d{2}",
            "S": r"\d{2}"
        }[match.group(1)], log_time_format)
        time_regex = re.compile(log_time_regex)
        ret = time_regex.findall(line)
        if not ret:
            return ""
        try:
            # if this time don't have year, will use 1900, 1900-08-15 11:04:11
            time_obj = datetime.strptime(ret[0], log_time_format)
            time_obj = time_obj.strftime("%Y-%m-%d-%H:%M:%S.%f")
        except ValueError:
            return ""
        return time_obj

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse the custom log
        :param parse_ctx: file path
        :param task_id: the task unique id
        :return: events list
        """
        kg_logger.info("%s files parse job started.", self.TARGET_FILE_PATTERNS)
        file_info_list = self.find_log(parse_ctx.parse_file_path)
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        if self.is_sdk_input:
            results = dict()
            for idx, each_custom_info in enumerate(parse_ctx.custom_info_list):
                results.update({
                    f"{self.TARGET_FILE_PATTERNS}_ID-{idx}": self._parse_each_custom_info(each_custom_info)
                })
        else:
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_info_list), task_id=task_id,
                                               daemon=False)
            for idx, each_custom_info in enumerate(file_info_list):
                multiprocess_job.add_security_job(f"{self.TARGET_FILE_PATTERNS}_ID-{idx}",
                                                  self._parse_each_custom_info, each_custom_info, task_id)
            results, _ = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.TARGET_FILE_PATTERNS)
        return list(chain(*results.values())), {}

    def _parse_each_custom_info(self, each_custom_info: MatchedCustomInfo, task_id: str = ""):
        """
        Parse each custom parsing file list
        :param each_custom_info: list of each custom parsing info
        :param task_id: the task unique id
        :return: event dict list
        """
        if self.is_sdk_input:
            return self._parse_custom_info_from_sdk(each_custom_info)
        return self._parse_custom_info_from_filestream(each_custom_info, task_id)

    def _parse_custom_info_from_sdk(self, each_custom_info: MatchedCustomInfo):
        event_storage = EventStorage()
        log_map = each_custom_info.sdk_custom_log_map
        results = []
        for source_file, log_item_list in log_map.items():
            for log_item in log_item_list:
                self.user_conf = self.regex_user.get(source_file, {})
                if not self.user_conf:
                    continue
                results.extend(self._parse_log_item(log_item, fault_type=source_file))
        for event in results:
            event_storage.record_event(event)
        return event_storage.generate_event_list()

    def _parse_log_item(self, log_item: LogInfoSaver, fault_type: str):
        event_list = []
        for line in self._yield_log(log_item):
            event_dict = self.parse_from_user_repository(line)
            if not event_dict:
                continue
            occur_time = log_item.modification_time or KG_MAX_TIME
            self.supplement_common_info(event_dict, log_item, occur_time, specified_type=fault_type)
            event_list.append(event_dict)
        return event_list

    def _parse_custom_info_from_filestream(self, each_custom_info: MatchedCustomInfo, task_id: str):
        each_custom_log_list = each_custom_info.custom_log_list
        if not each_custom_log_list:
            return []
        event_storage = EventStorage()
        multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(each_custom_log_list), task_id=task_id)
        for idx, file_path in enumerate(each_custom_log_list):
            multiprocess_job.add_security_job(f"{self.TARGET_FILE_PATTERNS}_ID-{idx}_{os.path.basename(file_path)}",
                                              self._parse_file, file_path, each_custom_info.custom_file_info)
        results, _ = multiprocess_job.join_and_get_results()
        for result in chain(*results.values()):
            event_storage.record_event(result)
        return event_storage.generate_event_list()

    def _parse_file(self, file_path, file_info):
        """
        Parse each file
        :param file_path: each custom parsing file
        :param file_info: ConfigInfo, file info
        """
        if not os.path.isfile(file_path) or not file_info.source_file:
            return []
        occur_time = self.params.get("end_time") or \
                     time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(float(os.path.getmtime(file_path))))

        for source in file_info.source_file:
            self.user_conf.update(self.regex_user.get(source, {}))

        events_list = []
        for line in self._yield_log(file_path):
            if file_info.log_time_format:
                occur_time = self._get_occur_time(line, file_info.log_time_format)
                if not occur_time:
                    kg_logger.warning("Custom parsing failed, time format: [%s], actual log: [%s].",
                                      file_info.log_time_format, line)
                    continue
                if occur_time < self.resuming_training_time:
                    continue
                if self.start_time and occur_time < self.start_time:
                    continue
                if self.end_time and occur_time > self.end_time:
                    continue
            event_dict = self.parse_from_user_repository(line)
            if not event_dict:
                continue
            event_dict.update({
                "occur_time": occur_time,
                "source_file": os.path.basename(file_path),
                "type": file_info.source_file
            })
            events_list.append(event_dict)
        return events_list
