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
from datetime import datetime, timedelta
from operator import itemgetter

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.parse_info import KGParseFilePath
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser
from ascend_fd.utils.regular_table import NODEDLOG_SOURCE
from ascend_fd.utils.tool import MultiProcessJob

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class NodeDLogParser(FileParser):
    SOURCE_FILE = NODEDLOG_SOURCE
    TARGET_FILE_PATTERNS = "noded_log_path"
    KEY_INFO = "key_info"
    EVENT_CODE = "event_code"
    FORMATTER = "%Y-%m-%d %H:%M:%S.%f"
    FILE_TIME_REGEX = re.compile(r"noded-(\d{4}-\d{2}-\d{2})T(\d{2}-\d{2}-\d{2}.\d{3}).log")
    LOG_TIME_REGEX = re.compile(r"\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}.\d{6}")

    def __init__(self, params):
        """
        NodeD Log parser
        """
        super().__init__(params)

    @staticmethod
    def _group_single_file_event(results):
        """
        Group the single file event
        :param results: single file event
        :return: group event
        """
        all_event_dict = {}
        for events_dict in results.values():
            for node_key, event in events_dict.items():
                origin_event = all_event_dict.get(node_key)
                if not origin_event:
                    all_event_dict.update({node_key: event})
                    continue
                if event.get("occur_time", "") > origin_event.get("occur_time", ""):
                    all_event_dict.update({node_key: event})
        return list(all_event_dict.values())

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse log file
        :param parse_ctx: all file paths
        :param task_id: the task unique id
        :return: parse descriptor result
        """
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        file_source_list = self._filter_noded_file(parse_ctx.parse_file_path)
        if not file_source_list:
            return [], {}
        self._check_start_end_time(file_source_list)
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
                job_name = f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}"
                multiprocess_job.add_security_job(job_name, self._parse_single_file, file_source)
            results, _ = multiprocess_job.join_and_get_results()

        all_event_dict = self._group_single_file_event(results)
        return self._deduplicate_event_by_event_code(all_event_dict), {}

    def _filter_noded_file(self, parse_filepath: KGParseFilePath):
        """
        Filter noded log file
        :param parse_filepath: dict contains all file paths
        :return: list contains all file paths
        """
        noded_file_list = self.find_log(parse_filepath)
        if not noded_file_list or len(noded_file_list) == 1:
            return noded_file_list
        noded_file_list.sort(key=lambda x: self._get_filename(x))
        # noded取结束时间前半个小时的日志，只判断结束时间
        if not self.end_time:
            return noded_file_list[-2:]
        filtered_noded_list = []
        end_flag = False
        for file_source in noded_file_list:
            if self.is_sdk_input and not file_source.path:
                filtered_noded_list.append(file_source)
                continue
            file_time = self.FILE_TIME_REGEX.findall(self._get_filename(file_source))
            if not file_time and not self.is_sdk_input:
                continue
            # transfer YYYY-MM-DD hh-mm-ss.*** to YYYY-MM-DD hh:mm:ss.***
            file_time = file_time[0][0] + " " + file_time[0][1].replace("-", ":")
            if file_time < self.start_time:
                continue
            filtered_noded_list.append(file_source)
            if file_time > self.end_time:
                end_flag = True
                break
        # 如果没找到文件时间大于结束时间的，则结束时间在noded.log中
        if not end_flag and self._get_filename(noded_file_list[-1]) == "noded.log":
            filtered_noded_list.append(noded_file_list[-1])
        return filtered_noded_list

    def _check_start_end_time(self, file_source_list):
        """
        Check start and end time
        :param file_source_list: list contains all file paths
        :return: None
        """
        if self.end_time and self.start_time:
            time = (datetime.strptime(self.end_time, self.FORMATTER) - timedelta(minutes=30)).strftime(self.FORMATTER)
            if self.start_time < time:
                self.start_time = time
            return
        # 如果没有plog的结束时间，就获取最后一个文件的修改时间
        last_line_log = self._get_last_line_log(file_source_list[-1]) if not self.is_sdk_input \
            else getattr(file_source_list[-1], "log_lines", [""])[-1]
        # 日志时间格式: [INFO]     2024/09/01 20:39:20.032447 207     ipmimonitor/ipmi_monitor.go:206 XXX
        log_time = self.LOG_TIME_REGEX.findall(last_line_log)
        if log_time:
            self.end_time = log_time[0].replace("/", "-")
        if self.end_time and not self.start_time:
            self.start_time = (datetime.strptime(self.end_time, self.FORMATTER) - timedelta(minutes=30)).strftime(
                self.FORMATTER)

    def _parse_single_file(self, file_source):
        """
        Parse the single noded log file
        :param file_source: noded log file source
        :return: fault events list of the file
        错误日志格式：
        XXX get fault event, [device type: NPU]-[device id: 1]-[error code: 56000005]
        XXX get fault event, [device type: NPU]-[device id: 2]-[error code: 56000005]
        """
        events_dict = {}
        for log_line in self._yield_log(file_source):
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            time_match = self.LOG_TIME_REGEX.findall(log_line)
            if not time_match:
                continue
            occur_time = time_match[0].replace("/", "-")
            if self.start_time and occur_time < self.start_time:
                continue
            if self.end_time and occur_time > self.end_time:
                break
            if occur_time < self.resuming_training_time:
                continue
            self.supplement_common_info(event_dict, file_source, occur_time)
            # device type为错误日志的关键字
            split_result = log_line.split("device type", 1)
            node_log_key = split_result[1] if len(split_result) == 2 else event_dict.get(self.EVENT_CODE)
            events_dict.update({node_log_key: event_dict})
        return events_dict

    def _deduplicate_event_by_event_code(self, all_event_list):
        """
        Deduplicate the event by event code
        :param all_event_list: all events list
        :return: deduplicate events list
        """
        all_event_list.sort(key=itemgetter("occur_time"))
        events_dict = {}
        for event in all_event_list:
            if not event.get(self.EVENT_CODE):
                continue
            origin_event = events_dict.get(event.get(self.EVENT_CODE))
            if origin_event:
                event[self.KEY_INFO] = origin_event[self.KEY_INFO] + os.linesep + event[self.KEY_INFO]
            events_dict.update({event.get(self.EVENT_CODE): event})
        return list(events_dict.values())
