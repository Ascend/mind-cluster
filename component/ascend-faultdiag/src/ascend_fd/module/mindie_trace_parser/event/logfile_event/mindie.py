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
import abc
import re

from ascend_fd.module.mindie_trace_parser.common.enum import LogFileType
from ascend_fd.module.mindie_trace_parser.event.base import EventBaseInfo
from ascend_fd.module.mindie_trace_parser.event.logfile_event.base import LogFileEvent


class MindIeLogFileEvent(LogFileEvent, metaclass=abc.ABCMeta):

    def __init__(self, log_filename: str):
        super().__init__(log_filename, EventBaseInfo(self.log_type()))

    @classmethod
    def log_type(cls):
        return LogFileType.MINDIE_SERVER


class MindIeServerLogFileEvent(MindIeLogFileEvent):
    _FILENAME_PATTERN = re.compile(r"(mindie-server-)(\w)(.{1,100})")

    @classmethod
    def is_file_matching(cls, filename):
        return "mindie-server" in filename

    def find_component_info(self) -> [str, str]:
        search = self._FILENAME_PATTERN.search(self.log_filename)
        if not search:
            return "", ""
        component_sign = search.groups()[0] + search.groups()[1]
        instance_id = search.groups()[1] + search.groups()[2]
        return component_sign, instance_id


class MindIeMsLogFileEvent(MindIeLogFileEvent):
    _FILENAME_PATTERN = re.compile(r"^([\w\-]{1,200})(controller|coordinator)-(.{1,100})")

    @classmethod
    def is_file_matching(cls, filename):
        return "controller" in filename or "coordinator" in filename

    def find_component_info(self) -> [str, str]:
        search = self._FILENAME_PATTERN.search(self.log_filename)
        if not search:
            raise RuntimeError(
                f"Pattern: {self._FILENAME_PATTERN.pattern} search failed in filename: {self.log_filename}")
        component_sign = search.groups()[1]
        instance_id = search.groups()[2]
        return component_sign, instance_id
