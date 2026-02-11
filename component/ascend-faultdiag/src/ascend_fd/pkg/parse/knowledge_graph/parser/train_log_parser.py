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
from ascend_fd.utils.fault_code import PYTORCH_ERRCODE_COMMON, PRE_TRACEBACK_FAULT, PRE_SEG_FAULT
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
        TracebackInfoParser(self.occur_time, self._get_source_file(file_source))
        SegInfoParser(self.occur_time, self._get_source_file(file_source))
        self.train_framework = self._get_train_framework_type(file_source)
        self.default_conf = self._filter_parsers_conf_by_train_framework(self.train_framework, self.default_conf)
        self.user_conf = self._filter_parsers_conf_by_train_framework(self.train_framework, self.user_conf)
        if self.is_sdk_input:
            self._parse_sdk_input(event_storage, file_source)
        else:
            self._parse_filestream(event_storage, file_source)
        return event_storage.generate_event_list() + TrainCallFaultParser.get_all_events()

    def _determine_occur_time(self, file_source: Union[str, LogInfoSaver]):
        if isinstance(file_source, str):
            return self.params.get("end_time") or \
                time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(float(os.path.getmtime(file_source))))
        return getattr(file_source, "modification_time", "") or self.params.get("end_time") or KG_MAX_TIME

    def _parse_single_train_log(self, line: str, event_storage: EventStorage, multi_line_locator):
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
            TrainCallFaultParser.parse_all(line)
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

    def _parse_sdk_input(self, event_storage: EventStorage, file_source: LogInfoSaver):
        self.pattern_matcher.log_lines = file_source.log_lines
        for idx, line in enumerate(file_source.log_lines):
            event_dict = self._parse_single_train_log(line, event_storage, multi_line_locator=idx)
            if not event_dict:
                continue
            self._complete_event_dict(event_dict, file_source)
            event_storage.record_event(event_dict)

    def _parse_filestream(self, event_storage: EventStorage, file_path: str):
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
                event_dict = self._parse_single_train_log(line, event_storage, multi_line_locator=file_stream)
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

    def _filter_parsers_conf_by_train_framework(self, train_framework, config):
        """
        Update parsers_conf, remove event code of other train framework
        :param train_framework:
        """
        parsers_conf = {}
        other_framework_code_keys = self.KEY_FOR_FILTER_OTHER_FRAMEWORK_CODE.get(train_framework, [])
        for code, params in config.items():
            if any(key in code.lower() for key in other_framework_code_keys):
                continue
            parsers_conf.update({code: params})
        return parsers_conf

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


class TrainCallFaultParser:
    TRAIN_CALL_HEADER = ""
    PRE_TRAIN_CALL_FAULT = ""
    UNKNOWN_TRAIN_CALL_FAULT = ""
    ERROR_PATTERN = re.compile(r'[\w.]{1,100}')
    MAX_LINE = 200
    ERROR_STR = "Error"
    EXCEPTION_STR = "Exception"
    _SUB_INSTANCES = []

    def __init__(self, default_time, source_file):
        self.train_call_info = []
        self.train_call_events = {}
        self.default_time = default_time
        self.source_file = source_file
        self.prefix = ""
        self.has_same_prefix = False
        self._register_instance(self)

    @classmethod
    def parse_all(cls, line):
        """
        Parse all train log
        :param line: the origin log line
        """
        for ins in cls._SUB_INSTANCES:
            ins.parse_line(line)

    @classmethod
    def get_all_events(cls):
        """
        Get all events
        :return: all fault events
        """
        all_events = []
        for ins in cls._SUB_INSTANCES:
            all_events.extend(ins.get_event_list())
        cls.clear_sub_instance()
        return all_events

    @classmethod
    def clear_sub_instance(cls):
        """
        Clean sub instance
        """
        cls._SUB_INSTANCES = []

    @classmethod
    def _register_instance(cls, instance):
        """
        Registering subclass instances
        :param instance: subclass instance
        """
        cls._SUB_INSTANCES.append(instance)

    def parse_line(self, line):
        """
        Parse the train log line, find the train call info block
        :param line: the origin log line
        """
        if not self.TRAIN_CALL_HEADER or not self.PRE_TRAIN_CALL_FAULT or not self.UNKNOWN_TRAIN_CALL_FAULT:
            return
        line = line.rstrip()
        if line.startswith(self.TRAIN_CALL_HEADER):
            self.train_call_info.append(line)
            return
        if line.endswith(self.TRAIN_CALL_HEADER):
            self.train_call_info.append(line)
            self.prefix = line.split(self.TRAIN_CALL_HEADER)[0]
            return

        if not self.train_call_info:
            return
        if len(self.train_call_info) == 1 and line.startswith(self.prefix.strip()):
            self.has_same_prefix = True

        self.train_call_info.append(line)
        if len(self.train_call_info) == self.MAX_LINE and self.UNKNOWN_TRAIN_CALL_FAULT not in self.train_call_events:
            self._update_events()
            self._clear_cache_info()
            return

        if self.has_same_prefix:
            temp_line = line[len(self.prefix):] if line.startswith(self.prefix.strip()) else line
            self._deal_same_prefix(temp_line)
            return
        self._match_specific_err(line)

    def get_event_list(self):
        """
        Get the train call events list
        :return: all train call events list
        """
        if self.train_call_info and self.UNKNOWN_TRAIN_CALL_FAULT not in self.train_call_events:
            self._update_events()
            self._clear_cache_info()
        return list(self.train_call_events.values())

    def _match_specific_err(self, line):
        """
        Match specific error pattern
        :param line: the origin log line
        """
        error_name = self._get_error_name(line)
        if not error_name:
            return
        event_code = f"{self.PRE_TRAIN_CALL_FAULT}_{error_name}"
        if event_code not in self.train_call_events:
            self._update_events(event_code)
        self._clear_cache_info()

    def _get_error_name(self, line):
        """
        Get error name
        :param line: the origin log line
        :return: error name
        """
        if self.ERROR_STR in line:
            prefix = line.split(self.ERROR_STR, 1)[0]
            find_datas = self.ERROR_PATTERN.findall(prefix)
            if find_datas:
                return f"{find_datas[-1]}{self.ERROR_STR}"
        if self.EXCEPTION_STR in line:
            prefix = line.split(self.EXCEPTION_STR, 1)[0]
            if not prefix or prefix.endswith(' '):
                return self.EXCEPTION_STR
            find_datas = self.ERROR_PATTERN.findall(prefix)
            if find_datas:
                return f"{find_datas[-1]}{self.EXCEPTION_STR}"
        return ""

    def _deal_same_prefix(self, line):
        """
        Deal the prefix of train call events
        :param line: the log line without a prefix
        """
        # processing blank line and rows containing critical logs
        keyword_line = ("  ", "\t", "Extension modules", "Segmentation fault")
        if not line or line.startswith(keyword_line) or "(most recent call first):" in line:
            return
        self._match_specific_err(line)
        if self.train_call_info and self.UNKNOWN_TRAIN_CALL_FAULT not in self.train_call_events:
            self.train_call_info.pop(-1)
            self._update_events()
        self._clear_cache_info()

    def _clear_cache_info(self):
        """
        Clear cache info
        """
        self.train_call_info.clear()
        self.prefix = ""
        self.has_same_prefix = False

    def _update_events(self, event_code=""):
        """
        Update train call events
        :param event_code: event code
        """
        event_code = event_code or self.UNKNOWN_TRAIN_CALL_FAULT
        self.train_call_events.update({
            event_code: {
                "event_code": event_code,
                "key_info": "\n".join(self.train_call_info),
                "source_device": "Unknown",
                "occur_time": self.default_time,
                "source_file": self.source_file
            }
        })


class TracebackInfoParser(TrainCallFaultParser):
    TRAIN_CALL_HEADER = "Traceback (most recent call last):"
    PRE_TRAIN_CALL_FAULT = PRE_TRACEBACK_FAULT
    UNKNOWN_TRAIN_CALL_FAULT = f"{PRE_TRAIN_CALL_FAULT}_UnknownError"

    def __init__(self, default_time, source_file):
        super().__init__(default_time, source_file)


class SegInfoParser(TrainCallFaultParser):
    TRAIN_CALL_HEADER = "Fatal Python error: Segmentation fault"
    PRE_TRAIN_CALL_FAULT = PRE_SEG_FAULT
    UNKNOWN_TRAIN_CALL_FAULT = f"{PRE_TRAIN_CALL_FAULT}_UnknownError"

    def __init__(self, default_time, source_file):
        super().__init__(default_time, source_file)
