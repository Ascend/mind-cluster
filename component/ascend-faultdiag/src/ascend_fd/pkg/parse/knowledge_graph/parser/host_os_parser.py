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
from collections import defaultdict
from datetime import datetime
from itertools import chain
from pathlib import Path
from typing import Union

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.parse_info import KGParseFilePath
from ascend_fd.utils.regular_table import OS_SOURCE, OS_DEMESG_SOURCE, OS_VMCORE_DMESG_SOURCE, OS_SYSMON_SOURCE, \
    KG_MAX_TIME
from ascend_fd.utils.status import FileOpenError
from ascend_fd.utils.tool import MultiProcessJob, check_and_format_time_str, MAX_SIZE, MB_SHIFT
from ascend_fd.utils.constant.str_const import DEFAULT_YEAR
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
CHUNK_SIZE = 50 * 1024 * 1024


class HostOsParser(FileParser):
    TIME_REGEX = None
    TIME_FORMAT = None

    def __init__(self, params):
        """
        Host OS Log parser
        """
        super().__init__(params)

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse log file
        :param parse_ctx: dict contains all file paths
        :param task_id: the task unique id
        :return: parse descriptor result
        """

        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        file_source_list = self._data_preprocessing(parse_ctx.parse_file_path)
        err_msg = dict()
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(file_source_list):
                results.update({
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_file(file_source)
                })
        else:
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_source_list),
                                               task_id=task_id, daemon=False)
            for idx, file_source in enumerate(file_source_list):
                multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}",
                                                  self._parse_file, file_source, task_id)
            results, err_msg = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        err_msg_dict = {self.__class__.__name__: list(chain(err_msg.values()))} if err_msg else {}
        return list(chain(*results.values())), err_msg_dict

    def _data_preprocessing(self, parse_filepath: KGParseFilePath):
        """
        Filter host log
        :param parse_filepath: all file paths
        :return: host log list
        """
        return self.find_log(parse_filepath)

    def _parse_file(self, file_source, task_id: str = ""):
        """
        Parse the single host os file
        :param file_source: host os file path
        :return: fault events list of the file
        """
        event_storage = EventStorage()
        if self.is_sdk_input:
            results_list = self._parse_from_sdk(file_source)
            for result in results_list:
                event_storage.record_event(result)
        else:
            results_dict = self._parse_from_filestream(file_source, task_id)
            for result in chain(*results_dict.values()):
                event_storage.record_event(result)
        return event_storage.generate_event_list()

    def _parse_from_filestream(self, file_path, task_id):
        file_size = os.stat(file_path).st_size
        if file_size > MAX_SIZE:
            raise FileOpenError(
                f"The size of {os.path.basename(file_path)} should be less than {MAX_SIZE >> MB_SHIFT} MB.")
        multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=10, task_id=task_id)
        for idx, start_pos in enumerate(range(0, file_size, CHUNK_SIZE)):
            end_pos = min(start_pos + CHUNK_SIZE, file_size)
            multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_CHUNK-ID-{idx}_{os.path.basename(file_path)}",
                                              self._parse_chunk, file_path, start_pos, end_pos)
        results, _ = multiprocess_job.join_and_get_results()
        return results

    def _parse_from_sdk(self, file_source: LogInfoSaver):
        events_list = []
        for line in file_source.log_lines:
            self._parser_single_os_log(file_source, line, events_list)
        return events_list

    def _parse_chunk(self, file_path, start_pos, end_pos):
        """
        Parse the chunk of the file
        :param file_path: host os file path
        :param start_pos: position where the file pointer start
        :param end_pos: position where the file pointer end
        :return: fault events list of the file
        """
        events_list = []
        for line in self._from_chunk_yield_log(file_path, start_pos, end_pos):
            self._parser_single_os_log(file_path, line, events_list)
        return events_list

    def _parser_single_os_log(self, file_source: Union[str, LogInfoSaver], line, events_list: list):
        event_dict = self.parse_single_line(line)
        if not event_dict:
            return
        event_dict.update({"source_file": self._get_source_file(file_source)})
        # Host file time use regex to get. Therefore, the time is obtained after the fault is obtained.
        try:
            occur_time = self._filter_host_time(line)
        except ValueError:
            return  # when cannot get time, ignore this line
        if not occur_time:
            return
        if self.is_sdk_input and file_source.modification_time and occur_time.startswith(DEFAULT_YEAR):
            year_len = 4  # the year info: yyyy
            occur_time = file_source.modification_time[:year_len] + occur_time[year_len:]
        if not self._check_time(occur_time):
            return
        self.supplement_common_info(event_dict, file_source, occur_time)
        events_list.append(event_dict)

    def _filter_host_time(self, context: str):
        """
        Filter host time
        :param context: singe line in raw log
        :return: filtered time from context
        """
        if self.TIME_REGEX is not None:
            ret = self.TIME_REGEX.findall(context)  # list e.g: ['Aug 15 11:04:11 2024'] or ['Aug 15 11:04:11']
            if ret:
                try:
                    # if this time don't have year, will use 1900, 1900-08-15 11:04:11
                    time_obj = datetime.strptime(ret[0], self.TIME_FORMAT)  # datetime e.g: 2024-08-15 11:04:11
                except ValueError:
                    return ""
                return str(time_obj) + ".000000"  # str e.g: 2024-08-15 11:04:11.000000
        # other format: e.g: ****-**-**T**:**:**.******+**:**xxxxxxxxxxxxxxxxxx UTC time format
        time_str = context.split("+")[0].replace("T", "-")
        return check_and_format_time_str(time_str.strip())

    def _check_time(self, occur_time):
        """
        Determine whether the fault occurrence time is within the training period
        :param occur_time: fault occurrence time
        :return: bool
        """
        if not self.start_time or not self.end_time:
            return True
        # check whether the year of the start time is the same as that of the end time.
        year_len = 4  # the year info: yyyy
        if occur_time[:year_len] != DEFAULT_YEAR:
            return self.start_time <= occur_time <= self.end_time and occur_time >= self.resuming_training_time
        cross_year_flag = self.start_time[:year_len] != self.end_time[:year_len]
        if cross_year_flag:
            return occur_time[year_len:] >= self.start_time[year_len:] or \
                occur_time[year_len:] <= self.end_time[year_len:]
        return self.start_time[year_len:] <= occur_time[year_len:] <= self.end_time[year_len:]


class HostMsgParser(HostOsParser):
    SOURCE_FILE = OS_SOURCE
    TARGET_FILE_PATTERNS = "host_log_path"
    TIME_REGEX = re.compile(r"\w{3} .?\d \d{2}:\d{2}:\d{2}")
    TIME_FORMAT = "%b %d %H:%M:%S"

    def _data_preprocessing(self, file_dict):
        """
        Filter path list directly or filter LogInfoSaver group by the common path
        :param file_dict: dict contains all file paths
        :return: filtered messages list
        """
        msg_list = self.find_log(file_dict)
        if self.is_sdk_input:
            grouped_item = defaultdict(list)
            for msg in msg_list:
                grouped_item[msg.path].append(msg)
            filtered_paths = self._sort_and_filter_msg_list(list(grouped_item.keys()))
            grouped_item = {path: grouped_item[path] for path in filtered_paths}
            return list(chain(*grouped_item.values()))
        return self._sort_and_filter_msg_list(msg_list)

    def _sort_and_filter_msg_list(self, msg_list: list):
        """
        Sort and filter msg list by path
        :param msg_list: the msg list to filter
        :return: the filtered messages list
        """
        result_msg = list()
        for file_path in msg_list:
            if "messages-during" in os.path.basename(file_path):
                result_msg.append(file_path)
        if result_msg:
            return result_msg

        if len(msg_list) <= 2:
            return msg_list
        msg_list.sort()
        if msg_list[0].endswith("messages"):
            msg_list = msg_list[1:] + [msg_list[0]]
        if not self.end_time and not self.start_time:
            return msg_list[-2:]
        start_time = self.start_time.split(" ")[0].replace("-", "")
        end_time = self.end_time.split(" ")[0].replace("-", "")
        if start_time > end_time:
            return msg_list[-2:]
        left, right = -1, -1
        for idx, file_path in enumerate(msg_list):
            sfx_time = os.path.basename(file_path).split("-")[-1]
            if sfx_time == "messages":
                continue
            if left == -1 and sfx_time >= start_time:
                left = idx
            if sfx_time > end_time:
                right = idx + 1
                break
        if left == -1:
            return [msg_list[-1]]
        if right == -1:
            result_msg = msg_list[left:]
        else:
            result_msg = msg_list[left:right]
        return result_msg[-2:]


class HostDMesgParser(HostOsParser):
    SOURCE_FILE = OS_DEMESG_SOURCE
    TARGET_FILE_PATTERNS = "host_dmesg_path"
    TIME_REGEX = re.compile(r"\w{3} .?\d \d{2}:\d{2}:\d{2} \d{4}")
    TIME_FORMAT = "%b %d %H:%M:%S %Y"


class HostSysMonParser(HostOsParser):
    SOURCE_FILE = OS_SYSMON_SOURCE
    TARGET_FILE_PATTERNS = "host_sysmon_path"
    TIME_FORMAT = "%b %d %H:%M:%S"


class HostVmCoreParser(HostOsParser):
    SOURCE_FILE = OS_VMCORE_DMESG_SOURCE
    TARGET_FILE_PATTERNS = "host_vmcore_dmesg_path"
    TIME_FORMAT = "%Y-%m-%d-%H:%M:%S"

    def _parse_chunk(self, file_path, start_pos, end_pos):
        """
        Parse the chunk of the file
        :param file_path: vmcore-dmesg.txt file path
        :param start_pos: position where the file pointer start
        :param end_pos: position where the file pointer end
        :return: fault events list of the file
        """
        events_list = []
        occur_time = self._get_occur_time(file_path)
        if not occur_time or not self._check_time(occur_time):
            return events_list
        for line in self._from_chunk_yield_log(file_path, start_pos, end_pos):
            event_dict = self.parse_single_line(line)
            if not event_dict:
                continue
            event_dict.update({"source_file": os.path.basename(file_path)})
            if "source_device" not in event_dict:
                event_dict.update({"source_device": "Unknown"})
            event_dict.update({"occur_time": occur_time})
            events_list.append(event_dict)
        return events_list

    def _get_occur_time(self, file_path):
        """
        Get the fault occurrence time
        :param file_path: vmcore-dmesg.txt file path. e.g: /crash/127.0.0.1-2024-09-23-11:25:29/vmcore-dmesg.txt
        :return: fault occurrence time
        """
        occur_time = ""
        if not file_path:
            return occur_time
        file_path = Path(file_path)
        file_name = file_path.name
        ip_time_dir = file_path.parent.name
        crash_dir = file_path.parent.parent.name
        if file_name != "vmcore-dmesg.txt" or crash_dir != "crash":
            return occur_time
        time_str = ip_time_dir.split('-', 1)[-1]
        try:
            datetime_time = datetime.strptime(time_str, self.TIME_FORMAT)
        except ValueError:
            # 如果解析失败，则不是有效的时间格式
            return occur_time
        return str(datetime_time) + ".000000"  # e.g: 2024-09-23 11:25:29.000000

    def _parse_from_sdk(self, file_source: LogInfoSaver):
        events_list = []
        occur_time = self._get_occur_time(file_source.path) or file_source.modification_time or KG_MAX_TIME
        if not occur_time or not self._check_time(occur_time):
            return events_list
        for line in file_source.log_lines:
            event_dict = self.parse_single_line(line)
            if not event_dict:
                continue
            self.supplement_common_info(event_dict, file_source, occur_time)
            events_list.append(event_dict)
        return events_list
