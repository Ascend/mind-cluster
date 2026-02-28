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
import re
from datetime import datetime
from itertools import chain

from ascend_fd.model.context import KGParseCtx
from ascend_fd.utils.tool import MultiProcessJob, check_and_format_time_str
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage
from ascend_fd.utils.regular_table import COMPOSITE_SWITCH_CHIP_SOURCE

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class LCNEParser(FileParser):
    _type = "lcne"
    TARGET_FILE_PATTERNS = "lcne_log_path"
    SOURCE_FILE = "LCNELog"
    TIME_FORMATS = [
        # 格式1: 'May 27 2025 11:25:00+08:00'
        (re.compile(r'\b[A-Z][a-z]{2}\s{0,3}\d{1,2}\s{0,3}\d{4}\s{0,3}\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}\b'),
         "%b %d %Y %H:%M:%S%z"),
        # 格式2: 'May 27 2025 11:25:00'
        (re.compile(r'\b[A-Z][a-z]{2}\s{0,3}\d{1,2}\s{0,3}\d{4}\s{0,3}\d{2}:\d{2}:\d{2}\b'),
         '%b %d %Y %H:%M:%S')
    ]
    TIME_FORMAT = "%Y-%m-%dT%H:%M:%S%z"
    MAX_TIME = "9999-12-31 23:59:59.999999"

    def __init__(self, params):
        super().__init__(params)
        self.default_conf.update(self.regex_conf.get(COMPOSITE_SWITCH_CHIP_SOURCE, {}))
        self.user_conf.update(self.regex_conf.get(COMPOSITE_SWITCH_CHIP_SOURCE, {}))
        self.timezone_trans_flag = self.get_timezone_trans_flag()

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse lcne log file
        :param parse_ctx: file paths
        :param task_id: unique task id
        :return: parse descriptor result
        """
        file_list = self.find_log(parse_ctx.parse_file_path)
        if not file_list:
            return [], {}
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        self.is_sdk_input = parse_ctx.is_sdk_input
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(file_list):
                results.update({
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_file(file_source)
                })
        else:
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_list), task_id=task_id)
            for idx, file_source in enumerate(file_list):
                multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}",
                                                  self._parse_file, file_source)
            results, _ = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return list(chain(*results.values())), {}

    def _filter_lcne_time(self, context: str):
        """
        Filter lcne time
        :param context: singe line in raw log
        :return: filtered time from context
        """
        for regex, time_format in self.TIME_FORMATS:
            ret = regex.findall(context)
            if not ret:
                continue
            try:
                time_obj = datetime.strptime(ret[0], time_format).strftime(self.TIME_FORMAT)
                return check_and_format_time_str(str(time_obj), self.timezone_trans_flag)
            except ValueError:
                pass
        return ""

    def _parse_file(self, file_source):
        """
        Parse single lcne log line by line
        :param file_source: log file source
        :return: a list of event dict
        """
        event_storage = EventStorage()
        for log_line in self._yield_log(file_source):
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            occur_time = self._filter_lcne_time(log_line) or self.MAX_TIME
            self.supplement_common_info(event_dict, file_source, occur_time)
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()
