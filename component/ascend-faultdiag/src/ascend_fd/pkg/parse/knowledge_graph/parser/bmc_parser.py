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
import logging

from itertools import chain
from typing import Union

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.utils.regular_table import KG_MAX_TIME
from ascend_fd.utils.tool import MultiProcessJob
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class BMCParser(FileParser):
    TARGET_FILE_PATTERNS = "bmc_log_path"
    SOURCE_FILE = "BMCLog"

    def __init__(self, params):
        super().__init__(params)

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse bmc log file
        :param parse_ctx: parse context
        :param task_id: unique task id
        :return: parse descriptor result
        """
        file_list = self.find_log(parse_ctx.parse_file_path)
        if not file_list:
            return [], {}
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_list), task_id=task_id)
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(file_list):
                results.update({
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_file(file_source)
                })
        else:
            for idx, file_source in enumerate(file_list):
                multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}",
                                                  self._parse_file, file_source)
            results, _ = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return list(chain(*results.values())), {}

    def _parse_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse single bmc log line by line
        :param file_source: log file path
        :return: a list of event dict
        """
        event_storage = EventStorage()
        for log_line in self._yield_log(file_source):
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            # temporarily assign a max time to the event as the log time format is unknown
            self.supplement_common_info(event_dict, file_source, KG_MAX_TIME)
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()


class BMCAppDumpParser(BMCParser):
    TARGET_FILE_PATTERNS = "bmc_app_dump_log_path"
    SOURCE_FILE = "BMCAppDumpLog"


class BMCDeviceDumpParser(BMCParser):
    TARGET_FILE_PATTERNS = "bmc_device_dump_log_path"
    SOURCE_FILE = "BMCDeviceDumpLog"


class BMCLogDumpParser(BMCParser):
    TARGET_FILE_PATTERNS = "bmc_log_dump_log_path"
    SOURCE_FILE = "BMCLogDumpLog"
