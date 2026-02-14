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
import os.path
from abc import ABC, abstractmethod
from typing import Union, Dict

from ascend_fd.configuration.config import CUSTOM_CONFIG_PATH
from ascend_fd.model.parse_info import KGParseFilePath
from ascend_fd.pkg.customize.custom_config.config_info import get_config_info, ConfigInfo
from ascend_fd.utils.status import ParamError
from ascend_fd.utils.tool import PatternMatcher, PatternSingleOrMultiLineMatcher, LINE_NUM_OF_MULTILINE_MATCH, \
    safe_read_open, merge_occurrence
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.utils.regular_table import MIN_TIME, DEVICEPLUGIN_SOURCE, COMPOSITE_SWITCH_CHIP_SOURCE, LCNE_SOURCE

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
EVENT_CODE = "event_code"
OCCURRENCE = "occurrence"
CUSTOM_EVENT = "is_custom_event"
NEGATIVE_ONE = -1


class FileParser(ABC):
    TARGET_FILE_PATTERNS = None
    SOURCE_FILE = None
    _type = ""

    def __init__(self, params: dict):
        """
        Log file parser base class
        """
        super(FileParser, self).__init__()
        self.params = params
        self.regex_conf = params.get("default_conf", {})
        self.default_conf = self.regex_conf.get(self.SOURCE_FILE, {})
        self.regex_user = params.get("user_conf", {})
        self.user_conf = self.regex_user.get(self.SOURCE_FILE, {})
        self.pattern_matcher = PatternMatcher()
        self.start_time = None
        self.end_time = None
        self.resuming_training_time = MIN_TIME
        self.is_sdk_input = False
        self.custom_config = get_config_info() if os.path.exists(CUSTOM_CONFIG_PATH) else ConfigInfo()

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        if getattr(cls, "SOURCE_FILE", None) and not cls.__subclasses__():
            if cls.SOURCE_FILE:
                ParserFactory.register_parser(cls.SOURCE_FILE, cls)

    @staticmethod
    def _yield_log(file_source):
        """
        Yield the log line by line
        :param file_source: the log file
        :return: log line
        """
        if isinstance(file_source, str):
            try:
                yield from FileParser._yield_log_by_file(file_source)
            except UnicodeDecodeError:
                kg_logger.warning("UTF-8 decode failed, try to decode by latin-1")
                yield from FileParser._yield_log_by_file(file_source, encoding="latin-1")
        elif isinstance(file_source, LogInfoSaver):
            for line in file_source.log_lines:
                yield line

    @staticmethod
    def _yield_log_by_file(file_source, encoding="UTF-8"):
        with safe_read_open(file_source, "r", encoding=encoding) as file_stream:
            while True:
                line = file_stream.readline()
                if not line:
                    break
                # txt log file may have \00, it will make the loop impossible to exit
                line = line.replace('\00', '')
                yield line.strip()

    @staticmethod
    def _get_file_path(file_source: Union[str, LogInfoSaver]):
        """
        Get a file path through file_source
        :param file_source: file source, file path or LogInfoSaver
        :return: file_path
        """
        if isinstance(file_source, str):
            return file_source
        return file_source.path

    @staticmethod
    def _get_filename(file_source: Union[str, LogInfoSaver]):
        """
        Get a filename through file_source
        :param file_source: file source, file path or LogInfoSaver
        :return: filename
        """
        if isinstance(file_source, str):
            return os.path.basename(file_source)
        if not file_source.filename:
            return "unknown_filename"
        return file_source.filename

    @staticmethod
    def _get_source_file(file_source: Union[str, LogInfoSaver]):
        """
        Get the source file of where the log is
        :param file_source: file source, file path or LogInfoSaver
        :return: filename or file path in sdk scene as required
        """
        if isinstance(file_source, str):
            return os.path.basename(file_source)
        if not file_source.filename:
            return "Unknown"
        return file_source.path

    @staticmethod
    def _from_chunk_yield_log(file_path: str, start_pos: int, end_pos: int):
        """
        Yield the log line by line
        :param file_path: the log file
        :param start_pos: position where the file pointer start
        :param end_pos: position where the file pointer end
        :return: log line
        """
        with safe_read_open(file_path, "r", encoding="UTF-8") as file:
            file.seek(start_pos)  # 移动到当前块的起始位置
            # 如果不是从文件头开始，跳过第一行，避免分块时行被截断
            if start_pos > 0:
                file.readline()
            # 按行读取文件，直到到达块的末尾
            while file.tell() < end_pos:
                line = file.readline()
                if not line:
                    break
                # txt log file may have \00, it will make the loop impossible to exit
                line = line.replace('\00', '')
                yield line.strip()

    @staticmethod
    def _get_last_line_log(file_path: str):
        """
        Get the last line log
        :param file_path: the log file
        :return: the last line log
        """
        with safe_read_open(file_path, "r", encoding="UTF-8") as file_stream:
            file_stream.seek(0, 2)  # 将文件指针移到文件末尾
            current_position = file_stream.tell()  # 获取当前文件指针位置
            while current_position > 0:
                file_stream.seek(current_position)
                last_char = file_stream.read(1)
                if last_char != "\n":
                    current_position -= 1  # 如果读取到空行，则将指针位置继续前移
                    continue
                last_line = file_stream.readline().replace('\00', '').strip()  # 读取一行并去除空白字符
                if last_line:
                    return last_line
                current_position -= 1  # 将指针位置继续前移
            return file_stream.readline().replace('\00', '').strip()

    @abstractmethod
    def parse(self, file_dict: dict, task_id: str):
        """
        Parse log file
        :param file_dict: file path dict
        :param task_id: the task unique id
        :return:
        """
        pass

    def get_timezone_trans_flag(self):
        """
        Get timezone tans flag
        :return: the timezone tans flag
        """
        if not self.custom_config:
            return False
        timezone_config = self.custom_config.timezone_config
        if not timezone_config:
            return False
        return timezone_config.get_trans_flag_by_type(self._type)

    def supplement_common_info(self, event_dict: dict, file_source: Union[str, LogInfoSaver], occur_time: str,
                               specified_type: str = ""):
        """
        Supplement absent common info
        :param event_dict: event dict for fault
        :param file_source: file source, file path or LogInfoSaver
        :param occur_time: fault occur time
        :param specified_type: given explicitly if the type need to be specified
        """
        if "source_device" not in event_dict:
            if self.is_sdk_input:
                event_dict.update({"source_device": file_source.device_id_str})
            else:
                event_dict.update({"source_device": "Unknown"})
        event_dict.update({
            "source_file": self._get_source_file(file_source),
            "occur_time": occur_time,
            "type": specified_type or self.SOURCE_FILE
        })

    def find_log(self, parse_filepath: KGParseFilePath):
        """
        Find the log path which need to parse
        :param parse_filepath: input log file path
        :return: log path list
        """
        log_list = parse_filepath.to_dict().get(self.TARGET_FILE_PATTERNS, [])
        if not log_list:
            kg_logger.warning("No %s files found in the directory", self.TARGET_FILE_PATTERNS)
        return log_list

    def single_line_parse(self, line: str, config):
        for code, params in config.items():
            if self.pattern_matcher.compare(params, line):
                key_info = self.pattern_matcher.key_info or line
                event_dict = {EVENT_CODE: code, "key_info": key_info, CUSTOM_EVENT: True}
                attr_result = self.pattern_matcher.match_attr(params.get("attr_regex", ""), key_info)
                # PTA故障通过ACL接口获取设备号，若获取失败则日志为缺省值-1，将-1转换为Unknown
                if attr_result.get("source_device") == "-1":
                    attr_result.update({"source_device": "Unknown"})
                event_dict.update(attr_result)
                event_dict.update(self.pattern_matcher.get_attr_info(attr_result, line, event_dict[EVENT_CODE]))
                return event_dict
        return {}

    def parse_from_user_repository(self, line: str) -> Dict[str, str]:
        return self.single_line_parse(line, self.user_conf)

    def parse_from_default_repository(self, line):
        result = self.single_line_parse(line, self.default_conf)
        if result:
            return result

        log_data = line
        if isinstance(self.pattern_matcher, PatternSingleOrMultiLineMatcher):
            log_data = self.pattern_matcher.read_multi_line(LINE_NUM_OF_MULTILINE_MATCH, line)
        event_dict = self.match_event_code(log_data, self.default_conf)
        if not event_dict and (self.SOURCE_FILE == DEVICEPLUGIN_SOURCE or self.SOURCE_FILE == LCNE_SOURCE):
            event_dict = self.match_event_code(log_data, self.regex_conf.get(COMPOSITE_SWITCH_CHIP_SOURCE, {}))
        return event_dict

    def parse_single_line(self, line: str, framework_name=""):
        """
        Parse singe line of log file, the log file type can be train log, NPU log, CANN log and OS log
        :param line: the single line to be parsed
        :param framework_name: the job AI framework_name, only use in trainLog
        :return: event dict
        """
        return self.parse_from_user_repository(line) or self.parse_from_default_repository(line)

    def match_event_code(self, line, parser_conf):
        log_data = line
        if isinstance(self.pattern_matcher, PatternSingleOrMultiLineMatcher):
            log_data = self.pattern_matcher.read_multi_line(LINE_NUM_OF_MULTILINE_MATCH, line)
        for code, params in parser_conf.items():
            if not self.pattern_matcher.compare(params, log_data):
                continue
            key_info = self.pattern_matcher.key_info or log_data
            event_dict = {EVENT_CODE: code, "key_info": key_info, CUSTOM_EVENT: False}
            attr_result = self.pattern_matcher.match_attr(params.get("attr_regex", ""), key_info)
            if not attr_result:
                return event_dict
            # PTA故障通过ACL接口获取设备号，若获取失败则日志为缺省值-1，将-1转换为Unknown
            if attr_result.get("source_device") == "-1":
                attr_result.update({"source_device": "Unknown"})
            event_dict.update(attr_result)
            event_dict.update(self.pattern_matcher.get_attr_info(attr_result, line, event_dict[EVENT_CODE]))
            return event_dict
        return {}


class ParserFactory:
    __registry = {}

    @classmethod
    def register_parser(cls, saver_name: str, parser_class: type):
        if not issubclass(parser_class, FileParser):
            raise ParamError(f"{parser_class} must be a subclass of FileParser")
        cls.__registry[saver_name] = parser_class

    @classmethod
    def get_parser_class(cls, source_file: str):
        return cls.__registry.get(source_file, None)


class EventStorage:
    """
    Used to store all events of the same group
    """
    TRAIN_FRAMEWORK = {"PT": "AISW_PyTorch", "MS": "AISW_MindSpore"}
    OCCUR_TIME = "occur_time"

    def __init__(self):
        self.all_events_dict = dict()

    def add_device_id(self, device_id):
        for single_event_dict in self.all_events_dict.values():
            single_event_dict.update({"source_device": device_id or "Unknown"})

    def record_event(self, event_dict: dict, with_occurrence: bool = True):
        """
        Record any occurrence of the fault
        Whereas the first occurrence in the same group would represent the fault (one file or one pid files)
        :param event_dict: fault event attribute
        :param with_occurrence: whether to use occurrence mechanism or not
        """
        event_code = event_dict.get("event_code")
        if not event_code:
            return
        event_flag_name = f"{event_code}_{event_dict.get('source_device', 'Unknown')}"
        line_time = event_dict.get(self.OCCUR_TIME, "")
        if not line_time:
            return
        if with_occurrence:
            self._arrange_event_by_occurrence(event_dict, line_time, event_flag_name)
            return
        self._arrange_event_by_time(event_dict, line_time, event_flag_name)

    def generate_event_list(self):
        """
        Generate all recorded fault events list
        :return: event list
        """
        return list(self.all_events_dict.values())

    def clear_event(self):
        """
        Clear the recorded events
        """
        self.all_events_dict.clear()

    def _arrange_event_by_occurrence(self, event_dict: dict, line_time: str, event_flag_name: str):
        """
        Arrange event by occurrence mechanism, the earliest occurrence would represent the event
        whereas other occurrences will be recorded for further use
        :param event_dict: fault event attribute
        :param line_time: log line time
        :param event_flag_name: event name with the flag
        :return:
        """
        line_key_info = event_dict.get("key_info", "")
        event_dict.setdefault(OCCURRENCE, []).append((line_time, line_key_info))
        store_event = self.all_events_dict.get(event_flag_name, {})
        if not store_event:
            self.all_events_dict.update({event_flag_name: event_dict})
            return
        store_event_time = store_event.get(self.OCCUR_TIME)
        if line_time < store_event_time:
            merge_occurrence(event_dict, store_event)
            self.all_events_dict.update({event_flag_name: event_dict})
            return
        if line_time > store_event_time:
            merge_occurrence(store_event, event_dict)

    def _arrange_event_by_time(self, event_dict: dict, line_time: str, event_flag_name: str):
        """
        Arrange event by the occurring time, the earliest occurrence would represent the event
        :param event_dict: fault event attribute
        :param line_time: log line time
        :param event_flag_name: event name with the flag
        :return:
        """
        store_event_time = self.all_events_dict.get(event_flag_name, {}).get(self.OCCUR_TIME)
        if not store_event_time or line_time < store_event_time:
            self.all_events_dict.update({event_flag_name: event_dict})
