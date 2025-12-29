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
import re
from typing import Union, List

from ascend_fd.model.context import KGParseCtx
from ascend_fd.utils import regular_table
from ascend_fd.utils.tool import MultiProcessJob, check_and_format_time_str
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.pkg.diag.root_cluster.utils import NEGATIVE_ONE
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage
from ascend_fd.pkg.parse.blacklist.blacklist_op import BlackListManager
from ascend_fd.pkg.parse.root_cluster.parser import filter_single_rank_info
from ascend_fd.utils.fault_code import RUNTIME_AICORE_EXECUTE_FAULT, AISW_CANN_MEMORY_INFO
from ascend_fd.utils.regular_table import CANN_PLOG_SOURCE, CANN_DEVICE_SOURCE

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
DEFAULT_EXEC_TIMEOUT = 1800
TIME_BUFFER = 300


class CANNLogParser(FileParser):
    HEXADECIMAL_TRIPLET_PATTERN = re.compile(r"errcode:(\((?:0x[0-9A-Fa-f]{1,16}|0), (?:0x[0-9A-Fa-f]{1,16}|0), "
                                             r"(?:0x[0-9A-Fa-f]{1,16}|0)\))")

    def __init__(self, params: dict):
        """
        Log info parser
        """
        super().__init__(params)
        self.event_list = []

    @staticmethod
    def get_time(line):
        """
        Get the time info
        :param line: log line
        :return: time info
        """
        # Log format, eg. "[ERROR] RUNTIME(python3):2023-02-08-14:03:57.xxx.xxx Log string xxx"
        # Parsing "time" format, eg. "2023-02-08 14:03:57.xxxxxx"
        occur_time = line[line.find(":") + 1:]
        occur_time = occur_time.split(" ")[0]
        return check_and_format_time_str(occur_time.strip())

    @staticmethod
    def check_repeat_line(line: str, repeat_line_set: set):
        """
        Check whether the key information (excluding the time) in this line has been parsed
        :param line: the log line
        :param repeat_line_set: the maintained repeat line set
        :return: bool, true mean the line's key info has been parsed
        """
        line_key = line.split("[", 2)[-1]
        if line_key in repeat_line_set:
            return True
        repeat_line_set.add(line_key)
        return False

    @staticmethod
    def get_device_id_from_line(line):
        """
        Get device id from log line
        :param line: log line
        :return: logic_device_id, phy_device_id
        """
        if regular_table.ENTRY_ROOT_INFO in line:
            logic_device_id = filter_single_rank_info(line, regular_table.ENTRY_DEVICE_INFO)
            if logic_device_id != NEGATIVE_ONE:
                return logic_device_id, ""
        if regular_table.RANK_NUM_INFO in line and regular_table.RANK_INFO in line:
            logic_device_id = filter_single_rank_info(line, regular_table.OLD_DEVICE_INFO) or \
                              filter_single_rank_info(line, regular_table.LOGIC_DEVICE_INFO)
            phy_device_id = filter_single_rank_info(line, regular_table.PHY_DEVICE_INFO)
            return logic_device_id if logic_device_id != NEGATIVE_ONE else "", \
                phy_device_id if phy_device_id != NEGATIVE_ONE else ""
        if regular_table.TOTAL_RANK_INFO in line and regular_table.SERVER_ID_INFO in line:
            logic_device_id = filter_single_rank_info(line, regular_table.LOGIC_DEVICE_INFO)
            return logic_device_id, ""
        return "", ""

    @staticmethod
    def set_time(line_time, start_time, end_time):
        if not start_time or line_time < start_time:
            start_time = line_time
        if not end_time or line_time > end_time:
            end_time = line_time
        return start_time, end_time

    @staticmethod
    def _verify_device_id(cur_device_id, phy_device_id, logic_device_id, sdk_device_id):
        # priority: sdk -> existed device id -> the physical device id for pid -> the logic device id for pid
        if sdk_device_id and sdk_device_id != "Unknown":
            return sdk_device_id
        device_id = cur_device_id or phy_device_id or logic_device_id
        return device_id

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse log file
        :param parse_ctx: file path
        :param task_id: the task unique id
        :return: parse fault events list
        """
        events_list = []
        pid_file_dict = self.find_log(parse_ctx.parse_file_path)
        if not pid_file_dict:
            return [], {}
        pid_file_item = pid_file_dict.items()
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        if self.is_sdk_input:
            results = dict()
            for pid, file_list in pid_file_item:
                results.update({
                    f"{self.SOURCE_FILE}_{pid}": self._parse_files_of_pid(pid, file_list)
                })
        else:
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(pid_file_item),
                                               task_id=task_id)
            for pid, file_list in pid_file_item:
                multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_{pid}", self._parse_files_of_pid, pid, file_list)
            results, _ = multiprocess_job.join_and_get_results()
        if self.SOURCE_FILE == CANN_PLOG_SOURCE:
            device_pid_map = dict()
            pid_info_cache = dict()
            for event_list, start_end_time, pid_device_id in results.values():
                pid, device_id = pid_device_id
                self.params.setdefault("pid_device_dict", dict()).update({pid: device_id})
                pid_info_cache.update({pid: (event_list, start_end_time)})
                if not device_id or device_id == "Unknown":
                    continue
                store_pid = device_pid_map.get(device_id, "")
                if not store_pid:
                    device_pid_map.update({device_id: pid})
                    continue
                store_pid_info = pid_info_cache.get(store_pid)
                store_start_end_time = store_pid_info[1]
                if store_start_end_time[0] <= start_end_time[0]:
                    pid_info_cache.pop(store_pid)
                    device_pid_map.update({device_id: pid})
                    continue
                pid_info_cache.pop(pid)
            self.params["useful_pid"] = list(pid_info_cache.keys())
            for event_list, start_end_time in pid_info_cache.values():
                events_list.extend(event_list)
                self._update_train_time(start_end_time[0], start_end_time[1])
        else:  # source_file == "CANN_Device"
            usefully_pid = self.params.get("useful_pid", [])
            for event_list, _, pid_device_id in results.values():
                if pid_device_id[0] not in usefully_pid:
                    continue
                events_list.extend(event_list)
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return events_list, {}

    def _update_train_time(self, start_time, end_time):
        """
        Update the train time interval of this train job
        :param start_time: the first log time
        :param end_time: the last log time
        """
        if not self.start_time or start_time and start_time < self.start_time:
            self.start_time = start_time
        if not self.end_time or end_time and end_time > self.end_time:
            self.end_time = end_time
        self.params.update({"start_time": self.start_time, "end_time": self.end_time})

    def _parse_files_of_pid(self, pid, file_list: List[Union[str, LogInfoSaver]]):
        """
        Parses all log files of the same PID
        :param pid: the pid of files
        :param file_list: the file list of one pid
        :return: the fault event list, the log start time and end time in this log file
        """
        event_storage = EventStorage()
        repeat_line_set = set()
        device_id = ""
        logic_device_id, phy_device_id, sdk_device_id = "", "", ""
        start_time, end_time = "", ""
        blacklist_manager = BlackListManager()
        aicore_errcode_record = ""
        memory_info_parser = MemoryInfoParser()
        for file_source in file_list:
            if not self.is_sdk_input and not os.path.isfile(file_source):
                continue
            sdk_device_id = getattr(file_source, "device_id_str", "")
            if self.SOURCE_FILE == CANN_DEVICE_SOURCE:
                device_id = self.params.get("pid_device_dict", {}).get(pid, "")
                kg_logger.warning("Cannot get device id with pid %s from plog", pid)
            for line in self._yield_log(file_source):
                line_time = self.get_time(line)
                # skip if the time is invalid or it is earlier than the resuming training time
                if not line_time or line_time < self.resuming_training_time:
                    continue
                start_time, end_time = self.set_time(line_time, start_time, end_time)
                # skip the repeat line
                if self.check_repeat_line(line, repeat_line_set):
                    continue
                if self.SOURCE_FILE == CANN_PLOG_SOURCE:
                    logic, phy = self.get_device_id_from_line(line)
                    logic_device_id = logic or logic_device_id
                    phy_device_id = phy or phy_device_id
                # if a retry occurs(record this key word), all previously obtained fault events of the PID are invalid
                # 但是需要忽略场景："Process group work %s, seq_num %u dispatch sucess.This error log can be ignored."
                if "error log can be ignored" in line and "Process group work" not in line:
                    repeat_line_set.clear()
                    event_storage.clear_event()
                # if the line startswith "[ERROR]", it may be an error msg and need to check
                if not line.startswith(regular_table.ERROR_ALL):
                    memory_info_parser.parse_line(line, file_source)
                    continue
                # check if it is on the blacklist
                if blacklist_manager.is_log_line_need_ignore(line):
                    continue
                # ensure a valid aicore errcode per file if it exists
                if not aicore_errcode_record:
                    match = self.HEXADECIMAL_TRIPLET_PATTERN.findall(line)
                    aicore_errcode_record = match[0] if match else ""
                # match keywords in the fault library.
                event_dict = self.parse_single_line(line)
                if not event_dict:
                    continue
                self.supplement_common_info(event_dict, file_source, line_time)
                # aicore execute fault errcode would be recorded, reported and explained through the diag report
                if event_dict.get("event_code") == RUNTIME_AICORE_EXECUTE_FAULT and aicore_errcode_record:
                    event_dict.update({"error_code": aicore_errcode_record})
                event_storage.record_event(event_dict)
        device_id = self._verify_device_id(device_id, phy_device_id, logic_device_id, sdk_device_id)
        event_storage.add_device_id(device_id)
        memory_info_parser.device_id = device_id
        event_list = event_storage.generate_event_list() + memory_info_parser.get_memory_event()
        return event_list, (start_time, end_time), (pid, device_id)


class MemoryInfoParser:
    MEMORY_INFO_PATTERN = re.compile(
        r"](?P<memory_attribute>[0-9_A-Z]{1,30}) (dev\d{0,3} )?(mem stats|Mem stats).{1,15}module_name="
        r"(?P<module_name>[A-Z]{1,10})[;,]? module_id=(?P<module_id>\d{1,10})[;,]? current_alloced_size="
        r"(?P<current_alloced_size>\d{1,20})[;,]? alloced_peak_size=(?P<alloced_peak_size>\d{1,20})[;,]? alloc_cnt="
        r"(?P<alloc_cnt>\d{1,5})[;,]? free_cnt=(?P<free_cnt>\d{1,5})"
    )
    MEMORY_INFO_KEY_WORDS = ["module_name=", "module_id=", "current_alloced_size=", "alloced_peak_size=", "alloc_cnt=",
                             "free_cnt="]

    def __init__(self):
        self.source_file = ""
        self.device_id = ""
        self.key_info_dict = dict()

    def parse_line(self, line: str, file_source: Union[str, LogInfoSaver]):
        """
        Parse the plog line, find the memory event info
        :param line: log line
        :param file_source: log file path or item info saver
        """
        if any(key_word not in line for key_word in self.MEMORY_INFO_KEY_WORDS):
            return
        # 若有多组内存信息，只保存最后一组的信息
        pre_line = line.split("module_id=")
        # memory_flag_name包含内存属性和模块名的字符串。e.g: SVM_MEM Mem stats (Bytes). (module_name=RUNTIME;
        memory_flag_name = pre_line[0].split("]")[-1]
        self.key_info_dict.update({memory_flag_name: line})
        self.source_file = os.path.basename(file_source) if isinstance(file_source, str) else file_source.path

    def get_memory_event(self) -> list:
        """
        Get memory events
        :return: memory event list
        """
        memory_events = []
        if not self.key_info_dict:
            return memory_events
        key_info_list = list(self.key_info_dict.values())
        occur_time = ""
        memory_info = []
        for key_info in key_info_list:
            occur_time = occur_time or CANNLogParser.get_time(key_info)
            regex_data = self.MEMORY_INFO_PATTERN.search(key_info)
            if not regex_data:
                continue
            memory_info.append(regex_data.groupdict())
        sorted_memory_info = sorted(memory_info, key=lambda info: (
            int(info["current_alloced_size"]), int(info["alloced_peak_size"])), reverse=True)
        memory_events.append({
            "event_code": AISW_CANN_MEMORY_INFO,
            "key_info": "\n".join(key_info_list),
            "source_device": self.device_id or "Unknown",
            "occur_time": occur_time,
            "source_file": self.source_file,
            "memory_info": sorted_memory_info
        })
        return memory_events


class CANNPlogParser(CANNLogParser):
    TARGET_FILE_PATTERNS = "plog_path"
    SOURCE_FILE = CANN_PLOG_SOURCE


class CANNDeviceLogParser(CANNLogParser):
    TARGET_FILE_PATTERNS = "device_log_path"
    SOURCE_FILE = CANN_DEVICE_SOURCE
