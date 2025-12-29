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

from itertools import chain
from typing import List, Union

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.parse_info import KGParseFilePath
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.utils.regular_table import DOCKER_RUNTIME_SOURCE, NPU_EXPORTER_SOURCE
from ascend_fd.utils.tool import MultiProcessJob
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class CommonDlParser(FileParser):
    DEFAULT_LOG_FILENAME = None
    FILE_TIME_REGEX = None
    LOG_TIME_REGEX = re.compile(r"(\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}.\d{6})")

    def __init__(self, params):
        super().__init__(params)

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse log file
        :param parse_ctx: file paths
        :param task_id: unique task id
        :return: parse descriptor result
        """
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        self.resuming_training_time = parse_ctx.resuming_training_time
        self.is_sdk_input = parse_ctx.is_sdk_input
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        file_source_list = self._path_preprocessing(parse_ctx.parse_file_path)
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(file_source_list):
                results.update({
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_file(file_source)
                })
        else:
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_source_list), task_id=task_id)
            for idx, file_source in enumerate(file_source_list):
                multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}",
                                                  self._parse_file, file_source)
            results, _ = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return list(chain(*results.values())), {}

    def _path_preprocessing(self, parse_filepath: KGParseFilePath):
        """
        Preprocessing may be needed for certain parser
        :param parse_filepath: file paths dict
        :return: preprocessed file path list
        """
        return self._filter_dl_log_list(self.find_log(parse_filepath), self.FILE_TIME_REGEX, self.DEFAULT_LOG_FILENAME)

    def _filter_dl_log_list(self, dl_list: List[Union[str, LogInfoSaver]], file_time_regex, default_file_name):
        """
        Filter dl log list
        :param file_time_regex: file time regex used for matching the corresponding log filename
        :param default_file_name: default file name
        :return: a list of dl log path
        """
        if len(dl_list) < 2:
            return dl_list
        dl_list.sort(key=lambda x: self._get_filename(x))
        if not self.start_time or not self.end_time:
            return dl_list[-2:]
        if self.start_time > self.end_time:
            return dl_list[-2:]
        filtered_dl_list = list()
        for file_source in dl_list:
            filename = self._get_filename(file_source)
            if self.is_sdk_input and not filename:
                filtered_dl_list.append(file_source)
                continue
            if filename == default_file_name:
                filtered_dl_list.append(file_source)
                continue
            date_and_time = file_time_regex.findall(filename)
            if not date_and_time:
                continue
            # transfer YYYY-MM-DD hh-mm-ss.*** to YYYY-MM-DD hh:mm:ss.***
            date_and_time = date_and_time[0][0] + " " + date_and_time[0][1].replace("-", ":")
            if date_and_time < self.start_time or date_and_time < self.resuming_training_time:
                continue
            filtered_dl_list.append(file_source)
            if date_and_time > self.end_time:
                break
        return filtered_dl_list

    def _parse_file(self, file_source: Union[str, LogInfoSaver]):
        """
        Parse single file line by line
        :param file_source: log file source, path or log object
        :return: a list of event dict
        """
        event_storage = EventStorage()
        for log_line in self._yield_log(file_source):
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            ret = self.LOG_TIME_REGEX.findall(log_line)
            # transfer YYYY/MM/DD hh:mm:ss.****** to YYYY-MM-DD hh:mm:ss.****** for further comparison with train time
            occur_time = ret[0].replace("/", "-") if ret else ""
            if not occur_time or occur_time < self.resuming_training_time:
                continue
            if self.start_time and occur_time < self.start_time:
                continue
            if self.end_time and occur_time > self.end_time:
                continue
            self.supplement_common_info(event_dict, file_source, occur_time)
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()


class NpuExporterParser(CommonDlParser):
    SOURCE_FILE = NPU_EXPORTER_SOURCE
    TARGET_FILE_PATTERNS = "npu_exporter_path"
    DEFAULT_LOG_FILENAME = "npu-exporter.log"
    FILE_TIME_REGEX = re.compile(r"npu-exporter-(\d{4}-\d{2}-\d{2})T(\d{2}-\d{2}-\d{2}.\d{3}).log")


class DockerRuntimeParser(CommonDlParser):
    SOURCE_FILE = DOCKER_RUNTIME_SOURCE
    TARGET_FILE_PATTERNS = "docker_runtime_path"
    DEFAULT_LOG_FILENAME_LIST = ["runtime-run.log", "hook-run.log"]
    RUNTIME_RUN_FILE_TIME_REGEX = re.compile(r"runtime-run-(\d{4}-\d{2}-\d{2})T(\d{2}-\d{2}-\d{2}.\d{3}).log")
    HOOK_RUN_FILE_TIME_REGEX = re.compile(r"hook-run-(\d{4}-\d{2}-\d{2})T(\d{2}-\d{2}-\d{2}.\d{3}).log")
    FILE_TIME_REGEX_LIST = [RUNTIME_RUN_FILE_TIME_REGEX, HOOK_RUN_FILE_TIME_REGEX]

    def _path_preprocessing(self, file_dict):
        """
        Filter two types of log file lists separately, then compose them as one
        :param file_dict: file paths dict
        :return: preprocessed file path list
        """
        path_lists = [[], []]
        for file_source in self.find_log(file_dict):
            filename = self._get_filename(file_source)
            # currently only one fault for docker runtime, which comes from runtime-run.log
            if filename.startswith("runtime") or self.is_sdk_input:
                path_lists[0].append(file_source)
                continue
            if filename.startswith("hook"):
                path_lists[1].append(file_source)
        composed_path_list = []
        for path_list, time_regex, default_name in zip(path_lists, self.FILE_TIME_REGEX_LIST,
                                                       self.DEFAULT_LOG_FILENAME_LIST):
            composed_path_list.extend(self._filter_dl_log_list(path_list, time_regex, default_name))
        return composed_path_list
