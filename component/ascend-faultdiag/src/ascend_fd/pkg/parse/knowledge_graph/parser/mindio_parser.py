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
from typing import Union

from ascend_fd.pkg.parse.parser_saver import LogInfoSaver

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage
from ascend_fd.utils.regular_table import MINDIO_SOURCE
from ascend_fd.utils.tool import MultiProcessJob, check_and_format_time_str

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class MindIOLogParser(FileParser):
    SOURCE_FILE = MINDIO_SOURCE
    TARGET_FILE_PATTERNS = "mindio_log_path"
    LOG_TIME_REGEX = re.compile(r"(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3})")

    def __init__(self, params):
        super().__init__(params)

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse log file
        :param parse_ctx: file paths
        :param task_id: unique task id
        :return: parse descriptor result
        """
        events_list = []
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        file_list = self.find_log(parse_ctx.parse_file_path)
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(sorted(file_list)):
                results.update({
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_file(file_source)
                })
        else:
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_list), task_id=task_id)
            for idx, file_source in enumerate(file_list):
                multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}",
                                                  self._parse_file, file_source)
            results, _ = multiprocess_job.join_and_get_results()
        for event_list in results.values():
            events_list.extend(event_list)
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return events_list, {}

    def _parse_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse single ttp_log log file
        :param file_source: log file path or LogInfoSaver instance
        :return: fault event list of the file
        """
        event_storage = EventStorage()
        file_path = self._get_source_file(file_source)
        for log_line in self._yield_log(file_source):
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            occur_time = self._get_occur_time(log_line)
            if not occur_time:
                continue
            if self.start_time and occur_time < self.start_time:
                continue
            if self.end_time and occur_time > self.end_time:
                continue
            if occur_time < self.resuming_training_time:
                continue
            event_dict.update({"source_file": os.path.basename(file_path)})
            if "source_device" not in event_dict:
                event_dict.update({"source_device": "Unknown"})
            event_dict.update({"occur_time": occur_time})
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()

    def _get_occur_time(self, log_line: str):
        """
        Get occur time
        :param log_line: the log line that contains time info
        """
        ret = self.LOG_TIME_REGEX.findall(log_line)
        if not ret:
            return ""
        occur_time = ret[0].replace(",", ".")
        return check_and_format_time_str(occur_time)
