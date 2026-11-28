#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
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
import os
from datetime import datetime, timezone, timedelta

from typing import Union

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.node_info import DeviceInfo
from ascend_fd.utils.regular_table import (
    PYMOTOR_VLLM_SOURCE,
    KG_MAX_TIME,
    VLLM_FILENAME_PATTERN,
    PYMOTOR_FILENAME_PATTERN,
    MMC_FILENAME_PATTERN,
    DEVICE_ID_PATTERN,
    IPADDR_PATTERN,
    NETMASK_PATTERN,
)
from ascend_fd.utils.tool import (
    MultiProcessJob,
    PatternSingleOrMultiLineMatcher,
    check_and_format_time_str,
    safe_read_open,
)
from ascend_fd.utils.net_tools import IPAddress
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import (
    FileParser,
    EventStorage,
)

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")

# pylint: disable=R0801


class PyMotorVLLMParser(FileParser):
    """
    Unified parser for vLLM and PyMotor logs.

    - vLLM can run standalone: only vllm-*.log files present.
    - PyMotor always runs with vLLM: mindie-motor-*.log files (controller, coordinator)
      and vllm-*.log files (mixed vLLM + NodeManager/EngineServer) are all in the same folder.

    File naming:
      - vllm-*.log          -> vLLM logs + PyMotor internal components (NodeManager, EngineServer)
      - mindie-motor-*.log  -> PyMotor standalone components (controller, coordinator)
      - mmc*.log            -> Memcache logs
    """

    _type = "pymotor_vllm"
    TARGET_FILE_PATTERNS = "pymotor_vllm_log_path"
    SOURCE_FILE = PYMOTOR_VLLM_SOURCE

    # Accept all file patterns from shared log directory (pre-compiled for performance)
    MMC_FILENAME_REGEX = re.compile(MMC_FILENAME_PATTERN)
    ACCEPTED_PATTERNS = [
        re.compile(VLLM_FILENAME_PATTERN),
        re.compile(PYMOTOR_FILENAME_PATTERN),
        MMC_FILENAME_REGEX,
    ]
    # Memcache logs are verbose, their DEBUG lines are skipped before the fault rule matching.
    LOG_LEVEL_REGEX = re.compile(r"\b(DEBUG|INFO|WARNING|WARN|ERROR|FATAL)\b", re.IGNORECASE)
    # Only the log header is inspected so that a DEBUG word inside the message body is ignored.
    LOG_LEVEL_SEARCH_LIMIT = 80

    SOURCE_DEVICE_KEY = "source_device"
    DEFAULT_DEVICE_ID = "Unknown"
    TIME_ZERO_MICROSECOND = ".000000"

    # Time format regexes (ordered by specificity):
    # 1. memcache/memfabric:       2026-11-13 18:56:55.164739 DEBUG
    # 2. vLLM with milliseconds:   2026-05-26 15:15:48,667 - INFO -
    # 3. PyMotor style (no ms):    2026-05-26 15:16:20  [INFO][root]
    # 4. vLLM short (no year):     INFO 05-26 15:16:42 [vllm.py:788]
    DATETIME_MICROSECOND_REGEX = re.compile(r"(\d{4}-\d{2}-\d{2}\s{1,}\d{2}:\d{2}:\d{2}\.\d{6})")
    DATETIME_MS_REGEX = re.compile(r"(\d{4}-\d{2}-\d{2}\s{1,}\d{2}:\d{2}:\d{2}[,.]\d{3})")
    DATETIME_NO_MS_REGEX = re.compile(r"(\d{4}-\d{2}-\d{2}\s{1,}\d{2}:\d{2}:\d{2})\s{1,}\[")
    DATETIME_SHORT_REGEX = re.compile(r"(?:INFO|WARNING|ERROR|DEBUG)\s{1,}(\d{2}-\d{2}\s{1,}\d{2}:\d{2}:\d{2})\s{1,}\[")

    def __init__(self, params):
        super().__init__(params)
        self.pattern_matcher = PatternSingleOrMultiLineMatcher()
        self.timezone_trans_flag = self.get_timezone_trans_flag()
        self._last_line_time = ""
        self._plog_start_time = ""
        self._plog_end_time = ""
        self._is_mmc_file = False

    @staticmethod
    def _is_accepted_file(filename):
        """Check if filename matches any accepted pattern."""
        return any(p.match(filename) for p in PyMotorVLLMParser.ACCEPTED_PATTERNS)

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse vLLM/PyMotor log file(s) from shared log folder.
        Accepts vllm-*.log, mindie-motor-*.log and mmc-*.log files.

        :param parse_ctx: knowledge graph parser context
        :param task_id: unique task id
        :return: parse descriptor result
        """
        self.is_sdk_input = parse_ctx.is_sdk_input
        self.resuming_training_time = parse_ctx.resuming_training_time

        file_list = self.find_log(parse_ctx.parse_file_path)
        err_msg = {}

        if self.is_sdk_input:
            # SDK input: file_list contains LogInfoSaver objects, skip filename filtering
            results = {
                f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_file(file_source)
                for idx, file_source in enumerate(file_list)
            }
        else:
            file_list = [f for f in file_list if self._is_accepted_file(os.path.basename(f))]
            if not file_list:
                kg_logger.info("No %s files matched, skip the %s log parse job.", self.SOURCE_FILE, self.SOURCE_FILE)
                return [], {}
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_list), task_id=task_id)
            for idx, file_source in enumerate(file_list):
                multiprocess_job.add_security_job(
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}",
                    self._parse_file,
                    file_source,
                )
            results, err_msg = multiprocess_job.join_and_get_results()

        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)

        event_list = []
        for file_event_list in results.values():
            event_list.extend(file_event_list)

        err_msg_dict = {self.__class__.__name__: list(err_msg.values())} if err_msg else {}
        return event_list, err_msg_dict

    def get_log_time(self, line):
        """
        Extract time from log line.
        Supports all formats found across vLLM, PyMotor and memcache logs:
        1. 2026-11-13 18:56:55.164739 DEBUG (memcache / memfabric, microseconds)
        2. 2026-05-26 15:15:48,667 - INFO - (vLLM with milliseconds)
        3. 2026-05-26 15:16:20  [INFO][root]... (PyMotor / vLLM no ms)
        4. INFO 05-26 15:16:42 [vllm.py:788] (vLLM short, no year)
        """
        match = self.DATETIME_MICROSECOND_REGEX.search(line)
        if match:
            return check_and_format_time_str(match.group(1), self.timezone_trans_flag)

        match = self.DATETIME_MS_REGEX.search(line)
        if match:
            line_time = match.group(1).replace(",", ".")
            return check_and_format_time_str(line_time, self.timezone_trans_flag)

        match = self.DATETIME_NO_MS_REGEX.search(line)
        if match:
            line_time = match.group(1) + self.TIME_ZERO_MICROSECOND
            return check_and_format_time_str(line_time, self.timezone_trans_flag)

        match = self.DATETIME_SHORT_REGEX.search(line)
        if match:
            short_time = match.group(1)
            current_year = datetime.now(timezone(timedelta(hours=8))).year
            line_time = f"{current_year}-{short_time}{self.TIME_ZERO_MICROSECOND}"
            return check_and_format_time_str(line_time, self.timezone_trans_flag)

        return ""

    def find_device_info(self, log_line):
        """
        Find device info from log line
        Format: device_id: X, device_ip_info: ['ipaddr:X.X.X.X\n', 'netmask:X.X.X.X\n']
        """
        device_match = re.search(DEVICE_ID_PATTERN, log_line)
        if not device_match:
            return None

        device_id = device_match.group(1)
        ipaddr_match = re.search(IPADDR_PATTERN, log_line)
        netmask_match = re.search(NETMASK_PATTERN, log_line)

        if not ipaddr_match or not netmask_match:
            return None

        ip_addr = ipaddr_match.group(1).strip().replace("\\n", "").replace("\\r", "")

        if not IPAddress.is_valid_ip(ip_addr):
            return None

        device_info = DeviceInfo()
        device_info.device_id = device_id
        device_info.device_ip = ip_addr

        return device_info

    def _parse_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse single log file line by line.
        File input is streamed, so the cross-line matcher reads the following lines on demand
        instead of the whole file being cached in memory. SDK input keeps its in-memory line list.

        :param file_source: log file path or LogInfoSaver instance
        :return: a list of event dict
        """
        event_storage = EventStorage()
        device_info_collector = {}
        self._last_line_time = ""
        self._plog_start_time = self.params.get("start_time")
        self._plog_end_time = self.params.get("end_time")
        self._is_mmc_file = self._is_mmc_log(file_source)
        if self.is_sdk_input:
            self._parse_sdk_input(event_storage, file_source, device_info_collector)
        else:
            self._parse_filestream(event_storage, file_source, device_info_collector)
        return event_storage.generate_event_list()

    def _parse_sdk_input(self, event_storage, file_source, device_info_collector):
        """
        Parse the in-memory lines of the SDK input, the line index is used for multiline matching.

        :param event_storage: event storage
        :param file_source: LogInfoSaver instance
        :param device_info_collector: device info collector
        """
        self.pattern_matcher.log_lines = file_source.log_lines
        for index, log_line in enumerate(file_source.log_lines):
            event_dict, stop = self._parse_single_log(log_line, file_source, device_info_collector, index)
            if stop:
                break
            if not event_dict:
                continue
            event_storage.record_event(event_dict)

    def _parse_filestream(self, event_storage, file_path: str, device_info_collector):
        """
        Parse the log file from the file stream, the file stream is used for multiline matching.

        :param event_storage: event storage
        :param file_path: log file path
        :param device_info_collector: device info collector
        """
        self.pattern_matcher.log_lines = None
        with safe_read_open(file_path, "r", encoding="UTF-8") as file_stream:
            while True:
                raw_line = file_stream.readline()
                if not raw_line:
                    break
                log_line = raw_line.replace("\00", "").strip()
                event_dict, stop = self._parse_single_log(log_line, file_path, device_info_collector, file_stream)
                if stop:
                    break
                if not event_dict:
                    continue
                event_storage.record_event(event_dict)

    def _update_matcher_by_locator(self, multi_line_locator):
        """
        Update the matcher location for multiline parsing: line index for SDK input,
        file stream for file input
        """
        if self.is_sdk_input:
            self.pattern_matcher.update_line_index(multi_line_locator)
            return
        self.pattern_matcher.update_stream(multi_line_locator)

    def _parse_single_log(self, log_line, file_source, device_info_collector, multi_line_locator):
        """
        Process a single log line, return the matched event dict and whether the remaining
        lines should be skipped.

        :return: (event dict, True if the plog end time is exceeded)
        """
        # Memcache DEBUG lines are noisy and never carry a fault, skip them before parsing.
        if self._is_mmc_file and self._is_debug_log_line(log_line):
            return {}, False
        log_time = self.get_log_time(log_line)
        if log_time:
            self._last_line_time = log_time
        # Lines without a timestamp inherit the time of the previous line.
        line_time = log_time or self._last_line_time
        if line_time and self._plog_end_time and line_time > self._plog_end_time:
            # The plog end time is exceeded, all the remaining lines are out of the window.
            return {}, True
        if line_time and self._plog_start_time and line_time < self._plog_start_time:
            # Earlier than the plog start time, skip the current line only.
            return {}, False

        self._update_matcher_by_locator(multi_line_locator)
        current_device_info = self.find_device_info(log_line)
        if current_device_info:
            device_info_collector[current_device_info.device_id] = current_device_info

        event_time = line_time or getattr(file_source, "modification_time", "") or KG_MAX_TIME
        # The fallback time of untimestamped lines must also stay in the plog window.
        if self._plog_start_time and event_time < self._plog_start_time:
            return {}, False
        if self._plog_end_time and event_time > self._plog_end_time:
            return {}, False

        event_dict = self.parse_single_line(log_line)
        if not event_dict:
            return {}, False

        self.supplement_common_info(event_dict, file_source, event_time)

        # Use real device_id extracted from log content instead of default
        if current_device_info and event_dict.get(self.SOURCE_DEVICE_KEY) == self.DEFAULT_DEVICE_ID:
            event_dict[self.SOURCE_DEVICE_KEY] = current_device_info.device_id

        device_id_from_event = event_dict.get(self.SOURCE_DEVICE_KEY, "")
        if device_id_from_event and device_id_from_event in device_info_collector:
            event_dict["device_ip"] = device_info_collector[device_id_from_event].device_ip

        return event_dict, False

    @classmethod
    def _is_mmc_log(cls, file_source):
        """Check whether the file is a Memcache (mmc*.log) log file."""
        return bool(cls.MMC_FILENAME_REGEX.match(cls._get_filename(file_source)))

    @classmethod
    def _is_debug_log_line(cls, line):
        """
        Check whether the log level of a line is DEBUG.
        Only the log header is inspected so that a DEBUG word inside the message body is ignored.
        """
        match = cls.LOG_LEVEL_REGEX.search(line, 0, cls.LOG_LEVEL_SEARCH_LIMIT)
        return bool(match) and match.group(1).upper() == "DEBUG"
