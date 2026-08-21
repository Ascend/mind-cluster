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
import unittest
from unittest.mock import patch

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.parse_info import KGParseFilePath
from ascend_fd.pkg.parse.knowledge_graph.parser.custom_log_parser import CustomLogParser
from ascend_fd.pkg.parse.parser_saver import MatchedCustomInfo


class FakeMultiProcessJob:
    """多进程任务替身：记录实例化信息，不做真实进程调度，避免 UT 启动子进程"""

    instances = []

    def __init__(self, module_name, pool_size, task_id, daemon=True, failed_raise=True):
        self.pool_size = pool_size
        self.instances.append(self)

    def add_security_job(self, *args, **kwargs):
        pass

    def join_and_get_results(self):
        return {}, {}


class TestCustomLogParser(unittest.TestCase):
    """Test cases for CustomLogParser"""

    @staticmethod
    def _create_parser():
        return CustomLogParser({})

    def test_parse_no_matched_custom_log_returns_empty_without_multiprocess(self):
        """自定义配置未匹配到任何文件时，parse 直接返回空结果，不创建 MultiProcessJob（不会触发超时）"""
        FakeMultiProcessJob.instances = []
        parser = self._create_parser()
        parse_file_path = KGParseFilePath(custom_log_list=[])
        # custom_info_list 非空：配置中存在 custom_parse_file 条目，但每个条目都未匹配到文件
        parse_ctx = KGParseCtx(
            parse_file_path=parse_file_path,
            custom_info_list=[MatchedCustomInfo(custom_file_info=None, custom_log_list=[])],
        )
        with patch("ascend_fd.pkg.parse.knowledge_graph.parser.custom_log_parser.MultiProcessJob", FakeMultiProcessJob):
            result, err_dict = parser.parse(parse_ctx, "test_task_no_match")

        self.assertEqual(result, [])
        self.assertEqual(err_dict, {})
        # 未匹配到文件时不应创建空任务池的多进程任务，否则会阻塞直至 MultiProcessJob 超时（默认 600s）
        self.assertEqual(FakeMultiProcessJob.instances, [])

    def test_parse_matched_custom_log_uses_multiprocess(self):
        """存在匹配文件时，parse 正常走多进程解析（确认空匹配提前返回不影响正常路径）"""
        FakeMultiProcessJob.instances = []
        parser = self._create_parser()
        fake_log_path = "/tmp/fake_obs_sdk.log"
        parse_file_path = KGParseFilePath(custom_log_list=[fake_log_path])
        parse_ctx = KGParseCtx(
            parse_file_path=parse_file_path,
            custom_info_list=[MatchedCustomInfo(custom_file_info=None, custom_log_list=[fake_log_path])],
        )
        with patch("ascend_fd.pkg.parse.knowledge_graph.parser.custom_log_parser.MultiProcessJob", FakeMultiProcessJob):
            result, err_dict = parser.parse(parse_ctx, "test_task_match")

        self.assertEqual(result, [])
        self.assertEqual(err_dict, {})
        self.assertTrue(FakeMultiProcessJob.instances, "匹配到文件时应当创建多进程任务")
        self.assertGreaterEqual(FakeMultiProcessJob.instances[0].pool_size, 1)


if __name__ == "__main__":
    unittest.main()
