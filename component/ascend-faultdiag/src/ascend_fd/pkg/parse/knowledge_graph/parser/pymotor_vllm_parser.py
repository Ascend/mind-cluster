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
from ascend_fd.model.parse_info import SingleFileParseInfo
from ascend_fd.utils.regular_table import (
    PYMOTOR_VLLM_SOURCE,
    KG_MAX_TIME,
    VLLM_FILENAME_PATTERN,
    PYMOTOR_FILENAME_PATTERN,
    DEVICE_ID_PATTERN,
    IPADDR_PATTERN,
    NETMASK_PATTERN,
    HOST_IP_PATTERN,
    POD_IP_PATTERN,
)
from ascend_fd.utils.tool import check_and_format_time_str, PatternSingleOrMultiLineMatcher
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
    """

    _type = "pymotor_vllm"
    TARGET_FILE_PATTERNS = "pymotor_vllm_log_path"
    SOURCE_FILE = PYMOTOR_VLLM_SOURCE

    # Accept both file patterns from shared log directory (pre-compiled for performance)
    ACCEPTED_PATTERNS = [re.compile(VLLM_FILENAME_PATTERN), re.compile(PYMOTOR_FILENAME_PATTERN)]

    SOURCE_DEVICE_KEY = "source_device"
    DEFAULT_DEVICE_ID = "Unknown"
    TIME_ZERO_MICROSECOND = ".000000"

    # Time format regexes (ordered by specificity):
    # 1. vLLM with milliseconds:   2026-05-26 15:15:48,667 - INFO -
    # 2. PyMotor style (no ms):     2026-05-26 15:16:20  [INFO][root]
    # 3. vLLM short (no year):     INFO 05-26 15:16:42 [vllm.py:788]
    DATETIME_MS_REGEX = re.compile(r"(\d{4}-\d{2}-\d{2}\s{1,}\d{2}:\d{2}:\d{2}[,.]\d{3})")
    DATETIME_NO_MS_REGEX = re.compile(r"(\d{4}-\d{2}-\d{2}\s{1,}\d{2}:\d{2}:\d{2})\s{1,}\[")
    DATETIME_SHORT_REGEX = re.compile(r"(?:INFO|WARNING|ERROR|DEBUG)\s{1,}(\d{2}-\d{2}\s{1,}\d{2}:\d{2}:\d{2})\s{1,}\[")

    def __init__(self, params):
        super().__init__(params)
        self.pattern_matcher = PatternSingleOrMultiLineMatcher()
        self.host_ip = ""
        self.pod_ip = ""
        self.device_info_list = []
        self.timezone_trans_flag = self.get_timezone_trans_flag()

    @staticmethod
    def _is_accepted_file(filename):
        """Check if filename matches any accepted pattern."""
        return any(p.match(filename) for p in PyMotorVLLMParser.ACCEPTED_PATTERNS)

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse vLLM/PyMotor log file(s) from shared log folder.
        Accepts both vllm-*.log and mindie-motor-*.log files.

        :param parse_ctx: knowledge graph parser context
        :param task_id: unique task id
        :return: parse descriptor result
        """
        self.is_sdk_input = parse_ctx.is_sdk_input
        self.resuming_training_time = parse_ctx.resuming_training_time
        plog_start_time = self.params.get("start_time")
        plog_end_time = self.params.get("end_time")

        file_list = self.find_log(parse_ctx.parse_file_path)
        file_list = [f for f in file_list if self._is_accepted_file(os.path.basename(f))]

        if not file_list:
            return [], {}

        results = dict()
        pv_start_time, pv_end_time = "", ""
        for idx, file_source in enumerate(sorted(file_list)):
            result = self._parse_file(file_source)
            results.update({f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": result})
            if result.start_time:
                pv_start_time = result.start_time if not pv_start_time else min(pv_start_time, result.start_time)
            if result.end_time:
                pv_end_time = result.end_time if not pv_end_time else max(pv_end_time, result.end_time)

        intersect_start_time = (
            max(plog_start_time, pv_start_time) if plog_start_time and pv_start_time else pv_start_time
        )
        intersect_end_time = min(plog_end_time, pv_end_time) if plog_end_time and pv_end_time else pv_end_time

        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)

        event_list = []
        for result in results.values():
            for event in result.event_list:
                event_time = event.get("occur_time", "")
                if not event_time:
                    continue
                if intersect_start_time and event_time < intersect_start_time:
                    continue
                if intersect_end_time and event_time > intersect_end_time:
                    continue
                event_list.append(event)

        return event_list, {}

    def get_log_time(self, line):
        """
        Extract time from log line.
        Supports all formats found across vLLM and PyMotor logs:
        1. 2026-05-26 15:15:48,667 - INFO - (vLLM with milliseconds)
        2. 2026-05-26 15:16:20  [INFO][root]... (PyMotor / vLLM no ms)
        3. INFO 05-26 15:16:42 [vllm.py:788] (vLLM short, no year)
        """
        match = self.DATETIME_MS_REGEX.search(line)
        if match:
            occur_time = match.group(1).replace(",", ".")
            return check_and_format_time_str(occur_time, self.timezone_trans_flag)

        match = self.DATETIME_NO_MS_REGEX.search(line)
        if match:
            occur_time = match.group(1) + self.TIME_ZERO_MICROSECOND
            return check_and_format_time_str(occur_time, self.timezone_trans_flag)

        match = self.DATETIME_SHORT_REGEX.search(line)
        if match:
            short_time = match.group(1)
            current_year = datetime.now(timezone(timedelta(hours=8))).year
            occur_time = f"{current_year}-{short_time}{self.TIME_ZERO_MICROSECOND}"
            return check_and_format_time_str(occur_time, self.timezone_trans_flag)

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

    def extract_host_pod_ip(self, log_line):
        """
        Extract host_ip and pod_ip from log line
        Format: host_ip: X.X.X.X / pod_ip: X.X.X.X
        """
        if not self.host_ip:
            host_match = re.search(HOST_IP_PATTERN, log_line)
            if host_match and IPAddress.is_valid_ip(host_match.group(1)):
                self.host_ip = host_match.group(1)

        if not self.pod_ip:
            pod_match = re.search(POD_IP_PATTERN, log_line)
            if pod_match and IPAddress.is_valid_ip(pod_match.group(1)):
                self.pod_ip = pod_match.group(1)

    def _parse_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse single log file line by line.
        Handles both vLLM and PyMotor log content transparently.

        :param file_source: log file path or LogInfoSaver instance
        :return: SingleFileParseInfo with events and device info
        """
        event_storage = EventStorage()
        device_info_collector = {}
        start_time, end_time = "", ""

        log_lines = list(self._yield_log(file_source))
        self.pattern_matcher.log_lines = log_lines
        for line_num, log_line in enumerate(log_lines, start=1):
            self.pattern_matcher.update_line_index(line_num - 1)
            occur_time, start_time, end_time = self._process_line(
                log_line,
                line_num,
                file_source,
                event_storage,
                device_info_collector,
                start_time,
                end_time,
            )

        return self._build_result(event_storage, start_time, end_time, device_info_collector)

    def _process_line(
        self,
        log_line,
        line_num,
        file_source,
        event_storage,
        device_info_collector,
        start_time,
        end_time,
    ):
        """Process a single log line: extract info, record event."""
        self.extract_host_pod_ip(log_line)

        current_device_info = self.find_device_info(log_line)
        if current_device_info:
            device_info_collector[current_device_info.device_id] = current_device_info

        occur_time = self.get_log_time(log_line)
        if not occur_time:
            occur_time = start_time or end_time or getattr(file_source, "modification_time", "") or KG_MAX_TIME

        if occur_time and occur_time >= self.resuming_training_time:
            if not start_time or occur_time < start_time:
                start_time = occur_time
            if not end_time or occur_time > end_time:
                end_time = occur_time

        event_dict = self.parse_single_line(log_line)
        if not event_dict:
            return occur_time, start_time, end_time

        self.supplement_common_info(event_dict, file_source, occur_time)

        # Use real device_id extracted from log content instead of default
        if current_device_info and event_dict.get(self.SOURCE_DEVICE_KEY) == self.DEFAULT_DEVICE_ID:
            event_dict[self.SOURCE_DEVICE_KEY] = current_device_info.device_id

        device_id_from_event = event_dict.get(self.SOURCE_DEVICE_KEY, "")
        if device_id_from_event and device_id_from_event in device_info_collector:
            event_dict["device_ip"] = device_info_collector[device_id_from_event].device_ip

        event_storage.record_event(event_dict)
        return occur_time, start_time, end_time

    def _build_result(self, event_storage, start_time, end_time, device_info_collector):
        """Assemble SingleFileParseInfo from collected data."""
        result = SingleFileParseInfo("", [], DeviceInfo(), {})
        result.container_ip = self.pod_ip or self.host_ip
        result.event_list = event_storage.generate_event_list()
        result.start_time = start_time
        result.end_time = end_time

        if device_info_collector:
            devices = list(device_info_collector.values())
            result.device_info = devices[0]
            result.device_info_list = devices
        else:
            result.device_info_list = []

        return result
