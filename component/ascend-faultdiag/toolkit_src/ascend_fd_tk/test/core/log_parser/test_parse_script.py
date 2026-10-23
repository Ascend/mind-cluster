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

"""日志关键字清洗脚本与本地日志解析器单元测试。"""

import asyncio
import json
import os
import tempfile
import unittest
from unittest import mock

from ascend_fd_tk.core.log_parser.base import LogParsePattern
from ascend_fd_tk.core.log_parser.local_log_parser import LocalLogParser
from ascend_fd_tk.core.log_parser.parse_script import (
    convert_path_regex,
    find_matching_files,
    load_patterns_from_config,
    process_pattern_group,
    search_keywords_in_file,
)


class TestConvertPathRegex(unittest.TestCase):
    def test_convert_path_regex(self):
        if os.sep == "\\":
            sep = "\\\\"
            self.assertEqual(convert_path_regex(r"log\/dir/.*\.txt"), "log" + sep + "dir" + sep + r".*\.txt")
        else:
            self.assertEqual(convert_path_regex("log/dir/.*.txt"), "log/dir/.*.txt")


class TestFindMatchingFiles(unittest.TestCase):
    def test_find_matching_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            sub = os.path.join(tmp, "a", "b")
            os.makedirs(sub)
            with open(os.path.join(sub, "info.log"), "w", encoding="utf-8") as f:
                f.write("x")
            with open(os.path.join(tmp, "other.log"), "w", encoding="utf-8") as f:
                f.write("y")
            found = list(find_matching_files(tmp, r"a[/\\]b[/\\]info\.log"))
            self.assertEqual(len(found), 1)
            self.assertTrue(found[0].endswith("info.log"))


class TestSearchKeywords(unittest.TestCase):
    def test_search_keywords_in_file(self):
        with tempfile.TemporaryDirectory() as tmp:
            file_path = os.path.join(tmp, "x.log")
            with open(file_path, "w", encoding="utf-8") as f:
                f.write("line ok\nother\nline again ok\n")
            matches = search_keywords_in_file(file_path, [("k", r"ok")])
            self.assertEqual([m[0] for m in matches], ["k", "k"])
            self.assertEqual(search_keywords_in_file(os.path.join(tmp, "no.log"), [("k", r"ok")]), [])


class TestProcessPatternGroup(unittest.TestCase):
    def test_process_pattern_group(self):
        with tempfile.TemporaryDirectory() as tmp:
            log_file = os.path.join(tmp, "node", "diag.log")
            os.makedirs(os.path.dirname(log_file))
            with open(log_file, "w", encoding="utf-8") as f:
                f.write("level: error\nsome text\nlevel: warn\n")
            results = process_pattern_group(
                tmp,
                [
                    ("error", r"level: error", r"node[/\\]diag\.log"),
                    ("warn", r"level: warn", r"node[/\\]diag\.log"),
                ],
            )
            self.assertEqual(len(results), 2)
            self.assertEqual({r["pattern_key"] for r in results}, {"error", "warn"})
            self.assertTrue(all(r["log_path"] for r in results))


class TestLoadPatternsFromConfig(unittest.TestCase):
    def test_load_patterns_from_config(self):
        with tempfile.TemporaryDirectory() as tmp:
            config_path = os.path.join(tmp, "c.json")
            with open(config_path, "w", encoding="utf-8") as f:
                json.dump([{"pattern_key": "k", "keyword_pattern": "re", "filepath_pattern": "p"}], f)
            self.assertEqual(load_patterns_from_config(config_path), [("k", "re", "p")])
            with mock.patch("ascend_fd_tk.core.log_parser.parse_script.sys.exit", side_effect=SystemExit(1)):
                with self.assertRaises(SystemExit):
                    load_patterns_from_config(os.path.join(tmp, "missing.json"))


class TestLocalLogParser(unittest.TestCase):
    def test_find(self):
        with tempfile.TemporaryDirectory() as tmp:
            with open(os.path.join(tmp, "diag.log"), "w", encoding="utf-8") as f:
                f.write("name=abc\nother\n")
            pattern = LogParsePattern.build("name", r"name=", r"diag\.log", r"name=(?P<name>\w+)")
            results = asyncio.run(LocalLogParser().find(tmp, {"name": pattern}))
            self.assertEqual(len(results), 1)
            self.assertEqual(results[0].info_dict, {"name": "abc"})

    def test_find_empty(self):
        self.assertEqual(asyncio.run(LocalLogParser().find("", None)), [])
        self.assertEqual(asyncio.run(LocalLogParser().find("x", {})), [])


if __name__ == "__main__":
    unittest.main()
