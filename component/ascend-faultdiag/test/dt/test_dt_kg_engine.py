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

import unittest
from unittest.mock import patch, MagicMock

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_compiler import ExprCompiler
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.vertex import Vertex
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.model.package_data import PackageData
from ascend_fd.utils.load_kg_config import EntityAttribute
from ascend_fd.utils.status import InfoIncorrectError, FileNotExistError


class MockWith:

    def __init__(self, return_value):
        self.return_value = return_value

    def __enter__(self):
        return self.return_value

    def __exit__(self, exc_type, exc_val, exc_tb):
        pass


class TestBuildPackageData(unittest.TestCase):

    def setUp(self) -> None:
        self.pkg_data_file_str = """
        {
            "20033": [
                {
                    "event_code": "20033",
                    "key_info": "[ERROR] RUNTIME(396379,python):2023-11-18-09:11:00.889.606",
                    "occur_time": "2023-11-18 09:11:00",
                    "count": "33",
                    "event_id": "key1"
                }
            ]
        }
        """

    @patch("ascend_fd.utils.tool.os")
    @patch("ascend_fd.utils.tool.safe_read_open")
    def test_build(self, safe_read_open, path):
        mock_file_stream = MagicMock()
        mock_file_stream.read.return_value = self.pkg_data_file_str
        safe_read_open.return_value = MockWith(mock_file_stream)
        path.exists.return_value = True
        pkg_data = PackageData([], "any_path")
        event_map = pkg_data.event_map
        self.assertIn("key1", event_map)

    def test_file_path_not_exists(self):
        file_path = "Not found"
        with self.assertRaises(FileNotExistError):
            PackageData([], file_path)

    @patch("ascend_fd.utils.tool.os")
    @patch("ascend_fd.utils.tool.safe_read_open")
    def test_empty_file_error(self, safe_read_open, path):
        mock_file_stream = MagicMock()
        mock_file_stream.read.return_value = "{}"
        safe_read_open.return_value = MockWith(mock_file_stream)
        path.exists.return_value = True
        with self.assertRaises(InfoIncorrectError) as error:
            PackageData([], "any_path")
        self.assertEqual(str(error.exception), "InfoIncorrectError(504): Failed to load data in JSON format.")


class TestExpr(unittest.TestCase):

    def setUp(self) -> None:
        self.compiler = ExprCompiler()
        src_event = Vertex("", {
            "event_id": "",
            "count": 1,
            "source": "nothing",
            "num": 2,
            "source_device": "2"
        }, EntityAttribute({}))
        dest_event = Vertex("", {
            "event_id": "",
            "count": 1,
            "source": "every thing",
            "num": 3,
            "source_device": "2"
        }, EntityAttribute({}))
        self.param = {
            "src": src_event,
            "dest": dest_event
        }

    def test_equals(self):
        res = self.compiler.compile('aaa == aaa').eval({})
        self.assertTrue(res)
        res = self.compiler.compile('src.count == dest.count').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('src.source == nothing').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('dest.source == "every thing"').eval(self.param)
        self.assertTrue(res)

    def test_not_equals(self):
        res = self.compiler.compile('aaa == 1').eval({})
        self.assertFalse(res)
        res = self.compiler.compile('src.count == dest.source').eval(self.param)
        self.assertFalse(res)
        res = self.compiler.compile('dest.source == nothing').eval(self.param)
        self.assertFalse(res)
        res = self.compiler.compile('dest.source == "everything"').eval(self.param)
        self.assertFalse(res)

    def test_gt(self):
        res = self.compiler.compile('1 > 0.1').eval({})
        self.assertTrue(res)
        res = self.compiler.compile('dest.num > src.num').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('dest.count > 0').eval(self.param)
        self.assertTrue(res)

    def test_gte(self):
        res = self.compiler.compile('1 >= 0.1').eval({})
        self.assertTrue(res)
        res = self.compiler.compile('1 >= 1').eval({})
        self.assertTrue(res)
        res = self.compiler.compile('src.count >= src.count').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('dest.count >= src.count').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('dest.num >= src.num').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('dest.num >= 2').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('dest.count >= 1').eval(self.param)
        self.assertTrue(res)

    def test_lt(self):
        res = self.compiler.compile('1 < 2.3').eval({})
        self.assertTrue(res)
        res = self.compiler.compile('src.num < dest.num').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('dest.count < 3').eval(self.param)
        self.assertTrue(res)

    def test_lte(self):
        res = self.compiler.compile('1 <= 2.3').eval({})
        self.assertTrue(res)
        res = self.compiler.compile('1 <= 1').eval({})
        self.assertTrue(res)
        res = self.compiler.compile('src.count <= src.count').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('src.count <= dest.count').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('2 <= dest.num').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('1 <= dest.count').eval(self.param)
        self.assertTrue(res)

    def test_and(self):
        res = self.compiler.compile('1 <= 2.3 and 2 > 1').eval({})
        self.assertTrue(res)
        res = self.compiler.compile('src.count <= src.count and src.count <= dest.count').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('2 <= source.num && 2 <= dest.num').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('1 <= 2.3 && 1 <= dest.count').eval(self.param)
        self.assertTrue(res)

    def test_or(self):
        res = self.compiler.compile('1 <= 2.3 or 2 < 1').eval({})
        self.assertTrue(res)
        res = self.compiler.compile('src.count > src.num or src.count <= dest.count').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('2 > dest.count || 5 <= dest.num').eval(self.param)
        self.assertTrue(res)
        res = self.compiler.compile('1 <= 2.3 || 2 > dest.count').eval(self.param)
        self.assertTrue(res)

    def test_bracket(self):
        res = self.compiler.compile('1 > 2.3 or (2 > 1 and 1 == 1)').eval({})
        self.assertTrue(res)
        res = self.compiler.compile(
            'src.count < dest.count or (src.num <= dest.num and 2 <= dest.num)').eval(self.param)
        self.assertTrue(res)
