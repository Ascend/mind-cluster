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
import logging
from typing import Union

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.utils.regular_table import VOLCANO_SCHEDULER_SOURCE, VOLCANO_CONTROLLER_SOURCE

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class VolcanoParser(FileParser):
    LOG_TIME_REGEX = re.compile(r"(\d{4} \d{2}:\d{2}:\d{2}.\d{6})")

    def __init__(self, params):
        """
        Volcano parser
        """
        super().__init__(params)
        self.train_across_year = False

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
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
        find_log = self.find_log(parse_ctx.parse_file_path)
        file_source = find_log[0] if find_log else ""
        if not file_source:
            return [], {}
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        self._check_and_process_train_time()
        results = self._parse_file(file_source)
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return results, {}

    def _check_and_process_train_time(self):
        """
        Drop year info of start/end time, mark if it is a cross-year training process
        format of start time and end time: YYYY-MM-DD hh-mm-ss.******
        """
        digits_of_year = 4
        if self.start_time and self.end_time and self.start_time[:digits_of_year] != self.end_time[:digits_of_year]:
            self.train_across_year = True
        if self.start_time:
            self.start_time = self.start_time[digits_of_year + 1:]
        if self.end_time:
            self.end_time = self.end_time[digits_of_year + 1:]

    def _parse_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse single volcano log file
        :param file_source: volcano log file path
        :return: fault event list of the file
        """
        event_storage = EventStorage()
        for log_line in self._yield_log(file_source):
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            occur_time = self._get_occur_time(log_line)
            if not occur_time:
                continue
            if self.start_time and occur_time < self.start_time and not self.train_across_year:
                continue
            if self.end_time and occur_time > self.end_time and not self.train_across_year:
                continue
            if occur_time < self.resuming_training_time and not self.train_across_year:
                continue
            self.supplement_common_info(event_dict, file_source, occur_time)
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()

    def _get_occur_time(self, log_line):
        """
        Check if the event is occurred in the train
        :param log_line: the log line that is going to be checked
        """
        ret = self.LOG_TIME_REGEX.findall(log_line)
        if not ret:
            return ""
        digits_of_month = 2
        # transfer MMDD hh:mm:ss.****** to MM-DD hh:mm:ss.******
        return ret[0][:digits_of_month] + '-' + ret[0][digits_of_month:]


class VolcanoSchedulerParser(VolcanoParser):
    SOURCE_FILE = VOLCANO_SCHEDULER_SOURCE
    TARGET_FILE_PATTERNS = "volcano_scheduler_path"


class VolcanoControllerParser(VolcanoParser):
    SOURCE_FILE = VOLCANO_CONTROLLER_SOURCE
    TARGET_FILE_PATTERNS = "volcano_controller_path"
