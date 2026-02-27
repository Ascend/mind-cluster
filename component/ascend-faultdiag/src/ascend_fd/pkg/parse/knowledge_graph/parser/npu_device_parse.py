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
import bisect
import re
from pathlib import Path
from typing import Union, Optional

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage
from ascend_fd.utils import regular_table
from ascend_fd.utils.fault_code import FIBER_OR_COPPER_LINK_FAULT
from ascend_fd.utils.regular_table import NPU_OS_SOURCE, NPU_DEVICE_SOURCE, NPU_HISTORY_SOURCE
from ascend_fd.utils.tool import MultiProcessJob, check_and_format_time_str
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
DEVICE_ID_PATTERN = re.compile(f"{regular_table.DEVICE_ID}|{regular_table.DEV_OS_ID}")
LOCAL_FAULT_FLAG = "1"
PCS_NORMAL_TH = "0"
NEGATIVE_ONE = "-1"
TS_PATTERN = re.compile(r'^(\d{14})-\d{9}$')


class BaseNpuLogParser(FileParser):

    def __init__(self, params):
        """
        NPU Device Log info parser
        """
        super().__init__(params)

    @staticmethod
    def _get_occur_time(line, dir_name) -> str:
        pass

    @staticmethod
    def _filter_log_list(log_list, start_time, end_time, time_format):
        """
        Filter log list by time, only the files in the training period are retained.
        :param log_list: log list
        :param start_time: start time
        :param end_time: end time
        :param time_format: time format
        :return: new log list
        """
        if start_time >= end_time:
            return log_list
        start_file = time_format.format(start_time.replace("-", "").replace(":", "").replace(" ", ""))
        end_file = time_format.format(end_time.replace("-", "").replace(":", "").replace(" ", ""))
        # use bisect to get the start time file and end time file index, obtaining logs in the Training Time Range
        start_idx = max(bisect.bisect(log_list, start_file) - 1, 0)
        end_idx = bisect.bisect(log_list, end_file)
        return log_list[start_idx: end_idx]

    def parse(self, file_dict: dict, task_id: str):
        pass

    def process_parse_file_list(self, file_source_list: list, task_id: str):
        """
        Parse multiple files, multiple processes them in cmd input scene
        :param file_source_list: the log file list
        :param task_id: the task unique id
        :return:
        """
        events_list = []
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(file_source_list):
                results.update({
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}":
                        self._parse_single_file(file_source)
                })
        else:
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_source_list),
                                               task_id=task_id)
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
        Parse single slog file
        :param file_source: slog file path
        :return: parse result
        """
        if not self.is_sdk_input and not os.path.isfile(file_source):
            return []
        device_id = self._determine_device_id(file_source)
        dir_name = os.path.dirname(file_source.path) if self.is_sdk_input else os.path.dirname(file_source)
        event_storage = EventStorage()
        for line in self._yield_log(file_source):
            occur_time = self._get_occur_time(line, dir_name)
            if not occur_time or occur_time < self.resuming_training_time:
                continue  # when cannot get time or this line is before resuming training, ignore this line
            # not in the Training Time Range, ignore it
            if self.start_time and occur_time < self.start_time:
                continue
            if self.end_time and occur_time > self.end_time:
                continue
            event_dict = self.parse_single_line(line)
            if not event_dict:
                continue
            # if rf_lf does not exist or when rf_lf is 1, pcs_err_cnt is equal to 0, the fault condition is not met
            rf_lf = event_dict.get("rf_lf", "")
            pcs_err_cnt = event_dict.get("pcs_err_cnt", PCS_NORMAL_TH)
            is_custom_fault = event_dict.get("event_code", "") == FIBER_OR_COPPER_LINK_FAULT
            local_fault_not_met = (rf_lf == LOCAL_FAULT_FLAG) and (pcs_err_cnt == PCS_NORMAL_TH)
            if is_custom_fault and (not rf_lf or local_fault_not_met):
                continue
            event_dict.update({"source_device": device_id})
            self.supplement_common_info(event_dict, file_source, occur_time)
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()

    def _determine_device_id(self, file_source):
        dir_name = os.path.dirname(file_source.path) if self.is_sdk_input else os.path.dirname(file_source)
        device_file_name = ""
        if self.SOURCE_FILE == NPU_HISTORY_SOURCE or self.SOURCE_FILE == NPU_DEVICE_SOURCE:
            device_file_name = os.path.basename(dir_name)
        if self.SOURCE_FILE == NPU_OS_SOURCE:
            device_file_name = os.path.basename(os.path.dirname(os.path.dirname(dir_name)))
        device_id_re = DEVICE_ID_PATTERN.findall(device_file_name)
        if not device_id_re:
            device_id = getattr(file_source, "device_id_str", "Unknown")
            if device_id == "Unknown":
                kg_logger.warning("The %s may not be a regular file, please check.", self._get_filename(file_source))
        else:
            device_id = next((match for match in device_id_re[0] if match), "Unknown")
        return device_id


class NpuHistoryLogParser(BaseNpuLogParser):
    TARGET_FILE_PATTERNS = "hisi_logs_path"
    SOURCE_FILE = NPU_HISTORY_SOURCE
    DIR_ERR_NUM = 100

    @staticmethod
    def _get_occur_time(line, dir_name):
        """
        Get the time info.
        history.log e.g: [yyyy-mm-dd-hh:mm:ss.******] *****
        :param line: log line
        :return: time info
        """
        time_str = line[line.find("[") + 1: line.find("]")]
        occur_time = check_and_format_time_str(time_str.strip())
        if not occur_time:
            time_str = NpuHistoryLogParser._extract_timestamp_from_dir(dir_name)
            occur_time = check_and_format_time_str(time_str.strip())
        return occur_time

    @staticmethod
    def _extract_timestamp_from_dir(path: str) -> Optional[str]:
        """
        Extract the timestamp from log dir path.
        path e.g: /hisi_logs/device-0/yyyymmddhhmmss-xxxxxxxxx/log/kernel.log
        :param path: dir path of log
        :return: timestamp string
        """
        p = Path(path)
        parts = p.parts
        try:
            idx = parts.index("hisi_logs")
        except ValueError:
            return ""
        for segment in parts[idx + 1:]:
            match = TS_PATTERN.match(segment)
            if match:
                return match.group(1)
        return ""

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse hisi logs file, contain history.log
        :param parse_ctx: file path
        :param task_id: the task unique id
        :return: hisi logs parse result
        """
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        return self.process_parse_file_list(self.find_log(parse_ctx.parse_file_path), task_id)


class NpuSlogParser(BaseNpuLogParser):

    @staticmethod
    def _get_occur_time(line, dir_name):
        """
        Get the time info.
        e.g:
        [****] ****:yyyy-mm-dd-hh:mm:ss.***.*** *******
        [****] ****: yyyy-mm-dd-hh:mm:ss.***.*** ******* (There's an extra space after the colon.)
        :param line: log line
        :return: time info
        """
        time_str = line[line.find(":") + 1:].strip()
        time_str = time_str.split(" ")[0]
        return check_and_format_time_str(time_str.strip())

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse slog file
        :param parse_ctx: file path
        :param task_id: the task unique id
        :return: slog parse result
        """
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        file_source_list = []
        slog_dict = self.find_log(parse_ctx.parse_file_path)
        if not self.is_sdk_input:
            slog_dict = self._filter_slog_by_time(slog_dict)
        for log_dir, log_list in slog_dict.items():
            for file_source in log_list:
                if isinstance(file_source, str):
                    file_source_list.append(os.path.join(log_dir, file_source))
                elif isinstance(file_source, LogInfoSaver):
                    file_source_list.append(file_source)
        return self.process_parse_file_list(file_source_list, task_id)

    def _filter_slog_by_time(self, slog_dict: dict):
        """
        Filter the slog dict by time
        Two log format:
        1) slog/dev-os-*/(debug/run)/device-os/device-os_***.log
        2) slog/dev-os-*/device-*/device-*_***.log
        :param slog_dict: slog dict
        :return: new slog dict
        """
        new_slog_dict = dict()
        if not self.start_time or not self.end_time:  # not get the train time interval
            for log_dir, log_list in slog_dict.items():
                new_slog_dict.update({log_dir: log_list[-2:]})  # use latest 2 files
            return new_slog_dict
        for log_dir, log_list in slog_dict.items():
            # file name timestamp use 17 numbers, start time and end time have 20 numbers, so need to cut
            log_format = '{}{}'.format(os.path.basename(log_dir), "_{:17}.log")
            new_log_list = self._filter_log_list(log_list, self.start_time, self.end_time, log_format)
            new_slog_dict.update({log_dir: new_log_list[-2:]})  # use latest 2 files
        return new_slog_dict


class NpuOsLogParser(NpuSlogParser):
    TARGET_FILE_PATTERNS = "slog_path"
    SOURCE_FILE = NPU_OS_SOURCE


class NpuDeviceLogParser(NpuSlogParser):
    TARGET_FILE_PATTERNS = "slog_path"
    SOURCE_FILE = NPU_DEVICE_SOURCE
