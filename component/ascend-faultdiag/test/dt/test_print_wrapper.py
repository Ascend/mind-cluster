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
from unittest.mock import patch

from ascend_fd.wrapper.print_wrapper import PrintWrapper
from ascend_fd.pkg.diag.message import NoteMsg
from ascend_fd.utils.i18n import LANG


class TestPrintWrapper(unittest.TestCase):
    """Test PrintWrapper class functionality"""

    def setUp(self) -> None:
        # Mock test data
        self.test_result = {
            "Kg": {
                "analyze_success": True,
                "version_info": {
                    "ascend_version": "23.0.0",
                    "cann_version": "6.0.RC1"
                },
                "fault": [
                    {
                        "code": "NORMAL_OR_UNSUPPORTED",
                        "description_zh": "故障事件分析模块无结果",
                        "fault_source": ["worker-0"]
                    },
                    {
                        "code": "TEST_FAULT_CODE",
                        "class": "Software",
                        "component": "Test",
                        "module": "Test",
                        "fault_source": ["worker-0", "worker-1"],
                        "cause_zh": "测试故障原因",
                        "description_zh": "测试故障描述",
                        "suggestion_zh": ["测试故障建议"]
                    }
                ]
            },
            "Rc": {
                "analyze_success": True,
                "root_cause_device": ["worker-0", "worker-1"],
                "remote_link": "worker-0 -> worker-1 -> worker-2",
                "fault_description": {
                    "string": "测试故障描述"
                }
            }
        }
        self.test_failed_details = {
            "ROOT_CLUSTER": "根因分析失败",
            "KNOWLEDGE_GRAPH": "故障事件分析失败"
        }

    def tearDown(self) -> None:
        pass

    def test_init(self):
        """Test PrintWrapper initialization"""
        wrapper = PrintWrapper(self.test_result, self.test_failed_details, False, False)
        self.assertIsNotNone(wrapper)
        self.assertIsNotNone(wrapper.table)

    def test_format_fault_attr_normal(self):
        """Test formatting of normal fault attributes"""
        wrapper = PrintWrapper({}, {}, False, False)
        
        # Test NORMAL_OR_UNSUPPORTED fault
        normal_fault = {
            "code": "NORMAL_OR_UNSUPPORTED",
            "description_zh": "故障事件分析模块无结果"
        }
        
        result_rows = wrapper._format_fault_attr(normal_fault, 0)
        self.assertEqual(len(result_rows), 2)
        self.assertIn("NORMAL_OR_UNSUPPORTED", str(result_rows[0]))
        self.assertIn("故障事件分析模块无结果", str(result_rows[1]))

    def test_format_fault_attr_abnormal(self):
        """Test formatting of abnormal fault attributes"""
        wrapper = PrintWrapper({}, {}, False, False)

        # Test abnormal fault
        abnormal_fault = {
            "code": "TEST_FAULT_CODE",
            "class": "Software",
            "component": "Test",
            "module": "Test",
            "fault_source": ["worker-0"],
            "cause_zh": "测试故障原因",
            "description_zh": "测试故障描述",
            "suggestion_zh": ["测试故障建议"]
        }

        result_rows = wrapper._format_fault_attr(abnormal_fault, 0)
        # Verify that there are at least status code row and fault classification row
        self.assertGreater(len(result_rows), 2)
        self.assertIn("TEST_FAULT_CODE", str(result_rows[0]))

    def test_parse_remote_link(self):
        """Test parsing remote links"""
        wrapper = PrintWrapper({}, {}, False, False)
        
        # Test short link
        short_link = "worker-0 -> worker-1"
        parsed_link, notes = wrapper._parse_remote_link(short_link)
        self.assertEqual(parsed_link, short_link)
        self.assertEqual(len(notes), 1)
        
        # Test long link
        long_link = " -> ".join([f"worker-{i}" for i in range(20)])
        parsed_link, notes = wrapper._parse_remote_link(long_link)
        self.assertIn("...", parsed_link)
        self.assertEqual(len(notes), 2)
        
        # Test empty link
        empty_link = ""
        parsed_link, notes = wrapper._parse_remote_link(empty_link)
        self.assertEqual(parsed_link, "")
        self.assertEqual(len(notes), 0)

    def test_long_str_format(self):
        """Test formatting long strings"""
        wrapper = PrintWrapper({}, {}, False, False)
        
        test_str = "测试长字符串\t包含制表符"
        formatted_str = wrapper._long_str_format(test_str)
        self.assertIn("测试长字符串", formatted_str)
        self.assertIn("包含制表符", formatted_str)

    def test_add_paragraph(self):
        """Test adding paragraphs"""
        wrapper = PrintWrapper({}, {}, False, False)
        
        # Test adding empty paragraph
        wrapper._add_paragraph([])
        # Test adding non-empty paragraph
        test_rows = [["", "测试", "测试内容"]]
        wrapper._add_paragraph(test_rows)
        # Verify table is not empty
        table_str = wrapper.get_format_table()
        self.assertIsNotNone(table_str)

    def test_add_fault_details(self):
        """Test adding fault details"""
        wrapper = PrintWrapper({}, {}, False, False)
        
        test_fault_details = [
            {"worker": "worker-0", "device": ["worker-0", "0"]},
            {"worker": "worker-1", "device": ["worker-1", "1"]}
        ]
        
        result_rows = wrapper._add_fault_details(test_fault_details)
        self.assertGreater(len(result_rows), 0)

    def test_format_rows(self):
        """Test formatting rows"""
        wrapper = PrintWrapper({}, {}, False, False)
        
        # Test formatting regular row
        test_row = wrapper._format_rows("标题", "指标", "描述")
        self.assertEqual(len(test_row), 3)
        self.assertEqual(test_row[0], "标题")
        self.assertEqual(test_row[1], "指标")
        
        # Test formatting list row
        test_list_row = wrapper._format_rows("标题", "指标", ["项1", "项2"])
        self.assertEqual(len(test_list_row), 3)

    def test_get_format_table(self):
        """Test getting formatted table"""
        wrapper = PrintWrapper(self.test_result, self.test_failed_details, False, False)
        table_str = wrapper.get_format_table()
        self.assertIsInstance(table_str, str)
        self.assertGreater(len(table_str), 0)

    def test_add_rc_rows(self):
        """Test adding root cause analysis rows"""
        wrapper = PrintWrapper(self.test_result, self.test_failed_details, False, False)
        # Verify method execution does not throw exceptions
        wrapper.add_rc_rows(self.test_result["Rc"])

    def test_add_result_rows_with_failure(self):
        """Test adding failure result rows"""
        # Modify test result to failure
        failed_result = {
            "Kg": {
                "analyze_success": False
            },
            "Rc": {
                "analyze_success": False
            }
        }
        wrapper = PrintWrapper(failed_result, self.test_failed_details, False, False)
        table_str = wrapper.get_format_table()
        self.assertIn("分析失败", table_str)


if __name__ == '__main__':
    unittest.main()