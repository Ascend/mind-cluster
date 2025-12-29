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

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.utils.regular_table import AMCT_SOURCE
from ascend_fd.utils.tool import MultiProcessJob

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class AMCTLogParser(FileParser):
    SOURCE_FILE = AMCT_SOURCE
    TARGET_FILE_PATTERNS = "amct_path"
    LOG_TIME_REGEX = re.compile(r"(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2},\d{3})")

    def __init__(self, params):
        """
        AMCT Log parser
        """
        super().__init__(params)

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse log file
        :param parse_ctx: all file paths
        :param task_id: the task unique id
        :return: parse descriptor result
        """

        file_source_list = self.find_log(parse_ctx.parse_file_path)
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        if not file_source_list:
            return [], {}
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(file_source_list):
                results.update({
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}":
                        self._parse_single_file(file_source)
                })
        else:
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_source_list), task_id=task_id)
            for idx, file_source in enumerate(file_source_list):
                job_name = f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}"
                multiprocess_job.add_security_job(job_name, self._parse_single_file, file_source)
            results, _ = multiprocess_job.join_and_get_results()
        events_list = []
        for event_list in list(results.values()):
            events_list.extend(event_list)
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return events_list, {}

    def _parse_single_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse the single amct log file
        :return: fault events list of the file
        :param file_source: amct log file path
        eg:
        2024-10-23 10:47:16,612 - WARNING - [AMCT]:[AMCT]: fakequant_precision_mode does not exist in the record file
        """
        event_storage = EventStorage()
        for log_line in self._yield_log(file_source):
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            ret = self.LOG_TIME_REGEX.findall(log_line)
            occur_time = ret[0].replace(",", ".") if ret else ""
            if not occur_time or occur_time < self.resuming_training_time:
                continue
            self.supplement_common_info(event_dict, file_source, occur_time)
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()
