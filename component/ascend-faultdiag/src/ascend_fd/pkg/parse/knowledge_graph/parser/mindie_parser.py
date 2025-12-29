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

from typing import Union

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.mindie_info import MindIEParseInfo, MindIELinkErrorInfo, MindIEPullKVErrorInfo
from ascend_fd.model.node_info import DeviceInfo
from ascend_fd.model.parse_info import SingleFileParseInfo, FilesParseInfo
from ascend_fd.utils.regular_table import MINDIE_SOURCE, KG_MAX_TIME
from ascend_fd.utils.tool import MultiProcessJob, check_and_format_time_str
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
NEGATIVE_ONE = -1
NUM_TWO = 2


def split_pid_info(file_path):
    """
    Split pid info from file path
    :param file_path: file path
    :return: pid info
    """
    # 文件名格式: mindie-llm_6521_20250512155117411.log
    # 文件名格式: mindie-ms_controller_198_20250613170316138.log
    file_name = os.path.basename(file_path)
    if "_" not in file_name:
        return ""
    temp = file_name.split("_")
    for i in range(1, len(temp) - 1):
        if temp[i].isdigit():
            return temp[i]
    return ""


def find_device_info(log_line: str, device_info: DeviceInfo):
    """
    Find device info from log line
    :param log_line: log line
    :param device_info: device info
    """
    if "local link info" not in log_line:
        return False
    # 日志格式：local link info: global_rank = 48, local_host_ip = 172.26.2.227, local_instance_id = 20347916160048,
    # 参数示例：local_device_ip = 10.0.2.31, local_logic_device_id = 0, local_physical_device_id = 0.
    split_log = log_line.split(",")
    for split_item in split_log:
        detail_info = split_item.split("=")[-1].strip(".").strip(" ")
        if not detail_info:
            continue
        if "local_physical_device_id" in split_item:
            device_info.device_id = detail_info
            device_info.phy_device_id = detail_info
            continue
        if "local_logic_device_id" in split_item:
            device_info.logic_device_id = detail_info
            continue
        if "local_device_ip" in split_item:
            device_info.device_ip = detail_info
    return bool(device_info.device_id and device_info.logic_device_id and device_info.device_ip)


class MindieParser(FileParser):
    _type = "mindie"
    TARGET_FILE_PATTERNS = "mindie_log_path"
    SOURCE_FILE = MINDIE_SOURCE
    FILENAME_REGEX = re.compile(r"mindie-[a-z-_]{1,15}_\d{1,7}_(\d{12,18})(?:\.\d{2})?.log")
    IP_REGEX = re.compile(r'\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b')
    mindie_parse_info = MindIEParseInfo()

    def __init__(self, params):
        super().__init__(params)
        self.container_ip = ""
        self.timezone_trans_flag = self.get_timezone_trans_flag()

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse mindie log file
        :param parse_ctx: file paths
        :param task_id: unique task id
        :return: parse descriptor result
        """
        file_list = self.find_log(parse_ctx.parse_file_path)
        files_parse_info = FilesParseInfo([], [])
        if not file_list:
            return files_parse_info, {}
        self.is_sdk_input = parse_ctx.is_sdk_input
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(sorted(file_list)):
                results.update({
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_file(file_source)
                })
        else:
            filtered_file_list = sorted((file for file in file_list if self._extract_log_time(file)),
                                        key=self._extract_log_time)
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_list), task_id=task_id)
            for idx, file_source in enumerate(filtered_file_list):
                multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}",
                                                  self._parse_file, file_source)
            results, _ = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)

        # 多文件并行清洗结过返回数据结构是{文件名：单文件清洗结果}
        for result in results.values():
            if not files_parse_info.container_ip:
                files_parse_info.container_ip = result.container_ip
            files_parse_info.event_list.extend(result.event_list)
            if result.device_info:
                files_parse_info.device_info_list.append(result.device_info)
            for key, value in result.link_error_info.items():
                files_parse_info.link_error_info.setdefault(key, set()).update(value)
            for key, value in result.pull_kv_error_info.items():
                files_parse_info.pull_kv_error_info.setdefault(key, set()).update(value)
        # 通过pid更新event的device_id
        pid_to_device_id_map = {}
        for device_info in files_parse_info.device_info_list:
            if device_info.pid and device_info.device_id:
                pid_to_device_id_map.update({device_info.pid: device_info.device_id})
        for event in files_parse_info.event_list:
            temp_pid = event.setdefault("pid", "")
            if not temp_pid:
                continue
            temp_device_id = pid_to_device_id_map.get(temp_pid)
            if not temp_device_id:
                continue
            event.update({"source_device": temp_device_id})
        return files_parse_info, {}

    def find_container_ip(self, log_line):
        """
        Find server ip from log line
        :param log_line: log line
        """
        if not self.container_ip and "localIp is" in log_line:
            # 日志格式:[endpoint] useTls is 0, localIp is 192.168.208.84, port is 10003
            self.container_ip = log_line.split("localIp is ")[-1].strip().split(",")[0]

    def add_link_error_info(self, link_error_info, log_line):
        """
        Add link error info to event dict
        :param link_error_info: link_error_info
        :param log_line: log line
        """
        ips = self.IP_REGEX.findall(log_line)
        if len(ips) < NUM_TWO:
            return
        link_error_info.setdefault(ips[0], set()).add(ips[1])

    def filter_error_info(self, log_line):
        """
        Filter link error info from log line
        :param log_line: log line
        """
        # 日志格式：Link from xxx to xxx failed, error code is MIE05E01001B

        if not log_line:
            return
        link_error_happened_flag = "MIE05E01001B" in log_line
        pull_kv_error_happened_flag = "MIE05E01001A" in log_line
        if not (link_error_happened_flag or pull_kv_error_happened_flag):
            return
        ips = self.IP_REGEX.findall(log_line)
        if len(ips) < NUM_TWO:
            return
        if link_error_happened_flag:
            self.mindie_parse_info.link_error_list.append(MindIELinkErrorInfo(ips[0], ips[1]))
        if pull_kv_error_happened_flag:
            self.mindie_parse_info.pull_kv_error_list.append(MindIEPullKVErrorInfo(ips[0], ips[1]))

    def get_log_time(self, line):
        """
        mindie现网日志时间格式包含如下：
        [2025-11-18 23:01:07.144+08:00]
        [2025-11-18 23:00:59.940]
        [2025-11-19 00:42:30,094]
        [2025-06-30 20:41:24.643119+08:00]
        """
        start_index = line.find("[")
        end_index = line.find("]")
        occur_time = ""
        if 0 <= start_index < end_index:
            occur_time = line[start_index + 1: end_index]
        if occur_time:
            occur_time = occur_time.replace(",", ".")
            return check_and_format_time_str(occur_time, self.timezone_trans_flag)
        return ""

    def _extract_log_time(self, file_source):
        """
        Extract time from filename, validate it and rearrange its format
        :param file_source: base filename
        :return: matched time info or empty str
        """
        ret = self.FILENAME_REGEX.findall(self._get_filename(file_source))
        return ret[0] if ret else ""

    def _parse_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse single mindie log line by line
        :param file_source: log file path or LogInfoSaver instance
        :return: a list of event dict
        """
        event_storage = EventStorage()
        single_file_parse_info = SingleFileParseInfo("", [], DeviceInfo(), {})
        device_info = single_file_parse_info.device_info
        link_error_info = single_file_parse_info.link_error_info
        pull_kv_error_info = single_file_parse_info.pull_kv_error_info
        find_stop_flag = False  # 用于判定device_info里面的信息是否补齐

        file_path = self._get_file_path(file_source)
        pid = split_pid_info(file_path)
        for log_line in self._yield_log(file_source):
            if not find_stop_flag:
                find_stop_flag = find_device_info(log_line, device_info)
            if "mindie-server" in file_path:
                self.find_container_ip(log_line)

            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            event_dict.update({"pid": pid})
            occur_time = self.get_log_time(log_line)
            # mindie是长稳服务，故障时间暂不通过时间过滤，后续添加用户输入时处理
            if not occur_time:
                # temporarily assign a max time to the event as the log time format is unknown
                modification_time = getattr(file_source, "modification_time", "")
                occur_time = modification_time or KG_MAX_TIME
            # component: Coordinator or Controller, specified by sdk input
            component = getattr(file_source, "component", "")
            specified_type = self.SOURCE_FILE if not component else f"{self.SOURCE_FILE}_{component}"
            self.supplement_common_info(event_dict, file_source, occur_time, specified_type=specified_type)
            event_storage.record_event(event_dict)
            # 过滤建链失败信息，用于后续MindIE根因设备诊断
            if event_dict.get("event_code") == "AISW_MindIE_LLM_TEXTGENERATOR_19":
                self.add_link_error_info(link_error_info, log_line)
            if event_dict.get("event_code") == "AISW_MindIE_LLM_TEXTGENERATOR_23":
                self.add_link_error_info(pull_kv_error_info, log_line)
        device_info.pid = pid

        single_file_parse_info.device_info = device_info if find_stop_flag else None
        single_file_parse_info.container_ip = self.container_ip
        single_file_parse_info.event_list = event_storage.generate_event_list()
        return single_file_parse_info
