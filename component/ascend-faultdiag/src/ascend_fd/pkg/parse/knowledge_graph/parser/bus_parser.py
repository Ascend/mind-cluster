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
from ascend_fd.utils.tool import MultiProcessJob, TimeBoundExtractor
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser, EventStorage

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class BusParser(FileParser):
    TARGET_FILE_PATTERNS = "bus_log_path"
    SOURCE_FILE = "LCNE_LOG"
    TIME_PATTERN = re.compile(r"^([A-Za-z]{3}\s\d{1,2}\s\d{4}\s\d{2}:\d{2}:\d{2})")
    ZIP_PATTERN = re.compile(r"_(\d{14})\.log\.zip$")
    LOG_PATTERN = re.compile(r"_(\d{14})\.log$")
    # 读取文件时缓冲区块大小
    CHUNK_SIZE = 4096
    # 如果没有训练时间，则所取的日志时间范围，单位分钟
    TIME_RANGE = 30

    def __init__(self, params):
        super().__init__(params)
        self.time_extractor = TimeBoundExtractor(
            line_time_parser=self._parse_log_time,
            line_validator=self._is_valid_log_line
        )

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

    @staticmethod
    def _is_valid_log_line(line: str) -> bool:
        """
        验证是否为有效日志行
        """
        return ' %%' in line

    def extract_time_from_path(self, path: str) -> Optional[datetime]:
        """
        从路径中提取时间戳，支持 .log.zip 和 .log 格式
        """
        # 匹配 .log.zip 文件中的时间
        zip_match = self.ZIP_PATTERN.search(path)
        if zip_match:
            time_str = zip_match.group(1)
        else:
            # 匹配普通 .log 文件中的时间
            log_match = self.LOG_PATTERN.search(path)
            time_str = log_match.group(1) if log_match else None

        if not time_str:
            return None

        try:
            return datetime.strptime(time_str, "%Y%m%d%H%M%S")
        except ValueError:
            return None

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse lcne log file
        :param parse_ctx: file paths
        :param task_id: unique task id
        :return: parse descriptor result
        """
        file_path_dict = self.find_log(parse_ctx.parse_file_path)
        if not file_path_dict:
            return [], {}
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)

        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")
        file_path_list = list(chain(*file_path_dict.values()))
        # 过滤出有效的日志文件
        filtered_file_list = self.filter_files_in_range(file_path_list)

        # 单文件清洗
        multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(filtered_file_list), task_id=task_id)
        for idx, file_path in enumerate(filtered_file_list):
            multiprocess_job.add_security_job(f"{self.SOURCE_FILE}_ID-{idx}_{os.path.basename(file_path)}",
                                              self._parse_file, file_path)
        results, _ = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return list(chain(*results.values())), {}

    def filter_files_in_range(self, file_paths: List[str]) -> List[str]:
        """
        过滤出与[start_time, end_time]有时间重叠的日志文件
        如果不存在训练时间窗口（self.start_time或self.end_time为None），直接返回所有解压后的 .log.zip 文件和非压缩文件
        """
        filtered_files = []
        has_valid_time_window = self.start_time and self.end_time

        for path in file_paths:
            # 处理 .log.zip 文件
            if path.endswith('.log.zip'):
                zip_time = self.extract_time_from_path(path)
                if not zip_time:
                    kg_logger.warning(f"Ignore zip files without timestamp: {path}")
                    continue
                # 如果时间早于 start_time，直接跳过
                if has_valid_time_window and zip_time < self.start_time:
                    continue
                # 解压并获取解压后的实际文件路径
                unzipped_path = self.unzip_file(path)
                if not unzipped_path:
                    continue
                path = unzipped_path

            try:
                if not has_valid_time_window:
                    filtered_files.append(path)
                    continue
                # 获取文件的最早和最晚时间
                earliest = self._get_time_bound(path, mode="earliest")
                latest = self._get_time_bound(path, mode="latest")

                if not earliest or not latest:
                    continue

                # 关键条件：检查两个区间是否重叠
                is_overlapping = (
                        earliest <= self.end_time
                        and
                        latest >= self.start_time
                )
                if is_overlapping:
                    filtered_files.append(path)
            except Exception as e:
                kg_logger.error(f"[Error] Exception occurred while processing {path}: {str(e)}")

        return filtered_files

    def _parse_log_time(self, line: str):
        """
        解析单行日志时间戳
        """
        match = self.TIME_PATTERN.match(line.strip())
        if not match:
            return None

        try:
            return datetime.strptime(match.group(1), "%b %d %Y %H:%M:%S")
        except ValueError:
            return None

    def _get_time_bound(self, file_path: str, mode: str = "latest") -> Optional[datetime]:
        """
        获取日志时间边界（支持正向/逆向搜索）
        :param mode: 'earliest' 从文件开头搜索最早时间 / 'latest' 从文件末尾搜索最新时间
        """
        return self.time_extractor.get_time_bound(file_path, mode)

    def _parse_file(self, file_path):
        """
        Parse single lcne log line by line
        :param file_path: log file path
        :return: a list of event dict
        """
        event_storage = EventStorage()
        for log_line in self._yield_log(file_path):
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                continue
            # 检查事件发生时间是否处于start_time和end_time之间
            occur_time = self._parse_log_time(log_line)
            if occur_time:
                if (self.start_time and self.end_time and
                        not (self.start_time <= occur_time <= self.end_time)):
                    continue
                event_dict["occur_time"] = occur_time.strftime("%Y-%m-%d %H:%M:%S")
            event_dict.update({"source_file": os.path.basename(file_path)})
            if "source_device" not in event_dict:
                event_dict.update({"source_device": "Unknown"})
            event_storage.record_event(event_dict)
        return event_storage.generate_event_list()
