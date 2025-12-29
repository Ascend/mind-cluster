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
import time
import logging
from typing import Union

from ascend_fd.model.context import KGParseCtx
from ascend_fd.utils.fault_code import PYTORCH_ERRCODE_COMMON, PRE_TRACEBACK_FAULT
from ascend_fd.utils.regular_table import TRAIN_LOG_SOURCE, KG_MAX_TIME
from ascend_fd.utils.tool import safe_read_open_with_size, PatternSingleOrMultiLineMatcher, MultiProcessJob
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class TrainLogParser(FileParser):
    TYPE_OF_MA_LOG = "Platform=ModelArts-Service"
    TARGET_FILE_PATTERNS = "train_log_path"
    SOURCE_FILE = TRAIN_LOG_SOURCE
    MINDSPORE_KEY = "mindspore"
    TORCH_KEY = "torch"
    TF_KEY = "tensorflow"
    MINDIO_FINISH_REPAIR_KEYWORD = "Mindio do repair operation ok"
    MINDIO_REPAIR_STATUS_LIST = ["recover", "retry"]
    KEY_FOR_FILTER_OTHER_FRAMEWORK_CODE = {
        MINDSPORE_KEY: [TORCH_KEY, TF_KEY],
        TORCH_KEY: [MINDSPORE_KEY, TF_KEY],
        TF_KEY: [TORCH_KEY, MINDSPORE_KEY]
    }
    READ_LINES = 100

    def __init__(self, params):
        """
        The Train Log parser
        """
        super().__init__(params)
        self.pattern_matcher = PatternSingleOrMultiLineMatcher()
        self.train_framework = ""
        self.occur_time = ""

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse the train log
        :param parse_ctx: the parse context object containing file path
        :param task_id: the task unique id
        :return: traceback error
        """
        events_list = []
        self.is_sdk_input = parse_ctx.is_sdk_input
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        file_source_list = self.find_log(parse_ctx.parse_file_path)
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
                multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}",
                                                  self._parse_single_file, file_source)
            results, _ = multiprocess_job.join_and_get_results()
        for event_list in results.values():
            events_list.extend(event_list)
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return events_list, {}

    def _parse_single_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse the single train log file
        :param file_source: the train log file source
        :return: event dict list
        """
        self.occur_time = self._determine_occur_time(file_source)
        event_storage = EventStorage()
        traceback_parser = TracebackInfoParser(self.occur_time, self._get_source_file(file_source))
        self.train_framework = self._get_train_framework_type(file_source)
        self._filter_parsers_conf_by_train_framework(self.train_framework)
        if self.is_sdk_input:
            self._parse_sdk_input(event_storage, file_source, traceback_parser)
        else:
            self._parse_filestream(event_storage, file_source, traceback_parser)
        return event_storage.generate_event_list() + traceback_parser.get_event_list()

    def _determine_occur_time(self, file_source: [str, LogInfoSaver]):
        if isinstance(file_source, str):
            return self.params.get("end_time") or \
                time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(float(os.path.getmtime(file_source))))
        return getattr(file_source, "modification_time", "") or self.params.get("end_time") or KG_MAX_TIME

    def _parse_single_train_log(self, line: str, event_storage: EventStorage, traceback_parser, multi_line_locator):
        """
        Process single line of train log, return a non-empty event dict if matched
        """
        if self.MINDIO_FINISH_REPAIR_KEYWORD in line and \
                any(status in line for status in self.MINDIO_REPAIR_STATUS_LIST):
            event_storage.clear_event()
            return {}
        if self.TYPE_OF_MA_LOG in line:  # MA logs support only user-defined faults
            event_dict = self.parse_from_user_repository(line) if self.custom_config.enable_model_asrt else {}
        else:
            traceback_parser.parse_line(line)
            # update matcher index for multiline parsing
            self._update_matcher_by_locator(multi_line_locator)
            event_dict = self.parse_single_line(line, framework_name=self.train_framework)
        if not event_dict:
            return {}
        # 如果PTA通用故障事件未解析出相关属性，则不记录该故障事件
        if event_dict.get("event_code", "").startswith(PYTORCH_ERRCODE_COMMON) \
                and not event_dict.get("complement"):
            return {}
        return event_dict

    def _update_matcher_by_locator(self, multi_line_locator):
        """
        Update matcher index for multiline parsing, filestream for cmd input and index for sdk input
        """
        if self.is_sdk_input:
            self.pattern_matcher.update_line_index(multi_line_locator)
            return
        self.pattern_matcher.update_stream(multi_line_locator)

    def _parse_sdk_input(self, event_storage: EventStorage, file_source: LogInfoSaver, traceback_parser):
        self.pattern_matcher.log_lines = file_source.log_lines
        for idx, line in enumerate(file_source.log_lines):
            event_dict = self._parse_single_train_log(line, event_storage, traceback_parser, multi_line_locator=idx)
            if not event_dict:
                continue
            self._complete_event_dict(event_dict, file_source)
            event_storage.record_event(event_dict)

    def _parse_filestream(self, event_storage: EventStorage, file_path: str, traceback_parser):
        """
        For cmd input, parse from filestream for a limited file size
        """
        if not os.path.isfile(file_path):
            return
        with safe_read_open_with_size(file_path, size=self.custom_config.train_log_size) as file_stream:
            while True:
                line = file_stream.readline()
                if not line:
                    break
                event_dict = self._parse_single_train_log(line, event_storage, traceback_parser,
                                                          multi_line_locator=file_stream)
                if not event_dict:
                    continue
                self._complete_event_dict(event_dict, file_path)
                event_storage.record_event(event_dict)

    def _complete_event_dict(self, event_dict, file_source: Union[str, LogInfoSaver]):
        """
        Fill info to complete an event dict
        """
        if not event_dict.get("event_code").startswith(PYTORCH_ERRCODE_COMMON):
            event_dict.update({"source_device": "Unknown"})
        if getattr(file_source, "device_id", ""):
            event_dict.update({"source_device": file_source.device_id_str})
        self.supplement_common_info(event_dict, file_source, self.occur_time)

    def _filter_parsers_conf_by_train_framework(self, train_framework):
        """
        Update parsers_conf, remove event code of other train framework
        :param train_framework:
        """
        parsers_conf = {}
        other_framework_code_keys = self.KEY_FOR_FILTER_OTHER_FRAMEWORK_CODE.get(train_framework, [])
        for code, params in self.parsers_conf.items():
            if any(key in code.lower() for key in other_framework_code_keys):
                continue
            parsers_conf.update({code: params})
        self.parsers_conf = parsers_conf

    def _get_train_framework_type(self, file_source: Union[str, LogInfoSaver]):
        """
        Get train framework type from log
        :param file_source: file source
        :return: train framework type
        """
        if self.is_sdk_input:
            for line in file_source.log_lines[:self.READ_LINES]:
                framework_cate = self._check_framework_key(line)
                if framework_cate:
                    return framework_cate
            return ""
        with safe_read_open_with_size(file_source, size=self.custom_config.train_log_size) as file_stream:
            while True:
                line = file_stream.readline()
                if not line:
                    break
                framework_cate = self._check_framework_key(line)
                if framework_cate:
                    return framework_cate
        return ""

    def _check_framework_key(self, line):
        """
        Check the framework keyword in train log. Return the framework cate
        :param line: log line
        :return: framework keyword
        """
        for key in [self.TORCH_KEY, self.MINDSPORE_KEY, self.TF_KEY]:
            if key in line.lower():
                return key
        return ""


class TracebackInfoParser:
    TRACEBACK_HEADER = "Traceback (most recent call last):"
    TRACEBACK_ERROR_PATTERN = re.compile("(^[\w\.]{1,100}Error|^[\w\.]{0,100}Exception)")
    MAX_LINE = 200
    UNKNOWN_TRACEBACK = f"{PRE_TRACEBACK_FAULT}_UnknownError"

    def __init__(self, default_time, source_file):
        self.traceback_info = []
        self.traceback_events = {}
        self.default_time = default_time
        self.source_file = source_file

    def parse_line(self, line):
        """
        Parse the train log line, find the traceback info block
        :param line: the origin log line
        """
        line = line.rstrip()
        if line.startswith(self.TRACEBACK_HEADER) or line.endswith(self.TRACEBACK_HEADER):
            self.traceback_info.append(line)
            return
        if not self.traceback_info:
            return
        self.traceback_info.append(line)
        if len(self.traceback_info) == self.MAX_LINE and self.UNKNOWN_TRACEBACK not in self.traceback_events:
            self.traceback_events.update({
                self.UNKNOWN_TRACEBACK:
                    {
                        "event_code": self.UNKNOWN_TRACEBACK,
                        "key_info": "\n".join(self.traceback_info),
                        "source_device": "Unknown",
                        "occur_time": self.default_time,
                        "source_file": self.source_file
                    }
            })
            self.traceback_info.clear()
            return
        error_re = self.TRACEBACK_ERROR_PATTERN.match(line)
        if error_re:
            error_name = error_re.group(1)
            event_code = f"{PRE_TRACEBACK_FAULT}_{error_name}"
            if event_code not in self.traceback_events:
                self.traceback_events.update({
                    event_code:
                        {
                            "event_code": event_code,
                            "key_info": "\n".join(self.traceback_info),
                            "source_device": "Unknown",
                            "occur_time": self.default_time,
                            "source_file": self.source_file
                        }
                })
            self.traceback_info.clear()
            return

    def get_event_list(self):
        """
        Get the traceback events list
        :return: all traceback events list
        """
        if self.traceback_info and self.UNKNOWN_TRACEBACK not in self.traceback_events:
            self.traceback_events.update({
                self.UNKNOWN_TRACEBACK:
                    {
                        "event_code": self.UNKNOWN_TRACEBACK,
                        "key_info": "\n".join(self.traceback_info),
                        "source_device": "Unknown",
                        "occur_time": self.default_time,
                        "source_file": self.source_file
                    }
            })
            self.traceback_info.clear()
        return list(self.traceback_events.values())
