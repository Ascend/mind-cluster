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
import zipfile
from datetime import datetime
from typing import Optional, List

from itertools import chain

from ascend_fd.model.context import KGParseCtx
from ascend_fd.utils.tool import MultiProcessJob
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class BusParser(FileParser):
    TARGET_FILE_PATTERNS = "bus_log_path"
    SOURCE_FILE = "LCNE_LOG"
    TIME_PATTERN = re.compile(r"^([A-Za-z]{3}\s\d{1,2}\s\d{4}\s\d{2}:\d{2}:\d{2})")
    LOG_PATTERN = re.compile(r"_(\d{14})\.log$")
    # 读取文件时缓冲区块大小
    CHUNK_SIZE = 4096
    # 如果没有训练时间，则所取的日志时间范围，单位分钟
    TIME_RANGE = 30

    def __init__(self, params):
        super().__init__(params)

    @staticmethod
    def unzip_file(zip_path: str) -> Optional[str]:
        """
        解压 .log.zip 文件，返回解压后的路径
        """
        try:
            # 解压到原目录，文件名去掉.zip 后缀
            unzip_dir = os.path.dirname(zip_path)
            with zipfile.ZipFile(zip_path, 'r') as z:
                files = z.namelist()
                if not files:
                    return None
                # 解压第一个文件
                z.extract(files[0], unzip_dir)
                return os.path.join(unzip_dir, files[0])
        except Exception as e:
            kg_logger.error(f"unzip failed: {zip_path}, error: {str(e)}")
            return None

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        pass

    def collect(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Collect raw events without time filtering.
        """
        file_path_dict = self.find_log(parse_ctx.parse_file_path)
        if not file_path_dict:
            return [], {}, {}
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        file_path_list = list(chain(*file_path_dict.values()))
        all_file_list = self._get_all_log_files(file_path_list)
        multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(all_file_list), task_id=task_id)
        for idx, file_path in enumerate(all_file_list):
            multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_ID-{idx}_{os.path.basename(file_path)}",
                                              self._parse_file_without_filter, file_path)
        results, _ = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return list(chain(*results.values())), {}, {}

    def _get_all_log_files(self, file_paths: List[str]) -> List[str]:
        """
        Get all log files without time filtering.
        """
        all_files = set()
        for path in file_paths:
            if path.endswith('log.zip'):
                unzipped_path = self.unzip_file(path)
                if unzipped_path:
                    all_files.add(unzipped_path)
            elif path.endswith('.log'):
                all_files.add(path)
        return list(all_files)

    def _parse_file_without_filter(self, file_path: str):
        """
        Parse file without time filtering.
        """
        event_storage = EventStorage()
        for line in self._yield_log(file_path):
            event_dict = self.parse_single_line(line)
            if not event_dict:
                continue
            occur_time = self._parse_log_time(line)
            if not occur_time:
                occur_time = ""
            self.supplement_common_info(event_dict, file_path, occur_time)
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()

    def filter_events(self, events_list: list, collect_result: dict):
        """
        Filter events by start_time and end_time from params.
        """
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        if not self.start_time and not self.end_time:
            return events_list
        filtered_list = []
        for event in events_list:
            occur_time = event.get("occur_time", "")
            if not occur_time:
                filtered_list.append(event)
                continue
            if self.start_time and occur_time < self.start_time:
                continue
            if self.end_time and occur_time > self.end_time:
                continue
            filtered_list.append(event)
        return filtered_list

    def _parse_log_time(self, line: str) -> str:
        """
        解析单行日志时间戳
        """
        match = self.TIME_PATTERN.match(line.strip())
        if not match:
            return ""
        try:
            time_obj = datetime.strptime(match.group(1), "%b %d %Y %H:%M:%S")
            return str(time_obj) + ".000000"
        except ValueError:
            return ""
