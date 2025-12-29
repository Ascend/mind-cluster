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

from ascend_fd.pkg.customize.custom_config import ConfigInfo
from ascend_fd.pkg.customize.custom_config.config_info import CustomFileInfo
from ascend_fd.pkg.customize.custom_entity.valid import code_check, paragraph_check, in_check, source_check, \
    rule_check, line_check_func_factor, check_missing_attribute_when_add


class ValidCase(unittest.TestCase):
    def test_code_check(self):
        self.assertTrue(code_check("AbCd123_-"))
        self.assertTrue(code_check("A" * 50))
        self.assertFalse(code_check("A" * 51))

    def test_source_check(self):
        self.assertTrue(source_check("TrainLog"))
        self.assertTrue(source_check("NPU_OS"))
        self.assertTrue(source_check("ABCD123"))
        self.assertFalse(source_check("1|2|3|4|5|6|7|8|9|10|11"))

    def test_rule_check(self):
        self.assertTrue(rule_check("", [{"dst_code": "12345"}], {"12345"}))
        self.assertFalse(rule_check("12345", [{"dst_code": "12345"}], {"12345"}))
        self.assertFalse(rule_check("", [{"dst_code": "12345"}], {"12346"}))
        self.assertTrue(rule_check("", [{"dst_code": "AbCd123_"}], {"AbCd123_"}))
        self.assertFalse(rule_check("", [{"dst_code": "A" * 51}], {"A" * 50}))
        self.assertFalse(rule_check("", [{}], {"12345"}))

    def test_paragraph_check(self):
        self.assertTrue(paragraph_check("这是一个测试的graph字段。"))
        self.assertTrue(paragraph_check("这是一个测试的graph字段。\n这是一个测试的graph字段。\n这是一个测试的graph字段。"))
        self.assertTrue(paragraph_check(["这是一个测试的graph字段。", "这是一个测试的graph字段。"]))
        self.assertFalse(paragraph_check(["这是一个测试的\ngraph字段。", "这是一个测试的graph字段。"]))
        self.assertTrue(paragraph_check("A" * 2000))
        self.assertFalse(paragraph_check("A" * 2001))

    def test_in_check(self):
        self.assertTrue(in_check(["AbCd123测试"]))
        self.assertTrue(in_check(["AbCd123测试", "AbCd123测试"]))
        self.assertTrue(in_check([["AbCd123测试"], ["AbCd123测试"]]))
        self.assertFalse(in_check([["AbCd123\n测试"], ["AbCd123测试"]]))
        self.assertFalse(in_check(["A" * 201]))

    def test_line_check_func(self):
        self.assertTrue(line_check_func_factor(length_range=(1, 5))("AbCd1"))
        self.assertFalse(line_check_func_factor(length_range=(1, 4))("AbCd1"))
        self.assertFalse(line_check_func_factor(length_range=(1, 4))("测试"))
        self.assertTrue(line_check_func_factor(length_range=(1, 4), allow_zh=True)("测试"))

    def test_check_missing_attribute(self):
        contain_set = {
            "attribute.class", "attribute.component", "attribute.module", "attribute.cause_zh",
            "attribute.description_zh", "not_right.wrong", "anything.else"
        }
        self.assertSetEqual(check_missing_attribute_when_add(contain_set),
                            {"attribute.suggestion_zh", "source_file", "regex.in"})
        all_contain_set = {
            "attribute.class", "attribute.component", "attribute.module", "attribute.cause_zh",
            "attribute.description_zh", "attribute.suggestion_zh", "source_file", "regex.in"
        }
        self.assertSetEqual(check_missing_attribute_when_add(all_contain_set), set())

    def test_check_custom_file_primary_params(self):
        data = {
            "enable_model_asrt": 1,
            "train_log_size": 1048576,
            "custom_parse_file": []
        }
        with self.assertRaises(TypeError) as cm:
            ConfigInfo.from_dict(data, True)
        exception = cm.exception
        self.assertEqual(str(exception), "Field 'enable_model_asrt' type mismatch. expected: <class 'bool'>, actual: "
                                         "<class 'int'>, value: 1")
        data["enable_model_asrt"] = True
        self.assertEqual(ConfigInfo.from_dict(data, True), ConfigInfo(enable_model_asrt=True))

    def test_check_custom_file_second_params(self):
        data = {
            "custom_parse_file": [
                {
                    "file_path_glob": "test_custom/*.log",
                    "log_time_format": "%H:%M:%S,%f",
                    "source_file": [1]
                }
            ]
        }
        with self.assertRaises(TypeError) as cm:
            ConfigInfo.from_dict(data, True)
        exception = cm.exception
        self.assertEqual(str(exception), "Field 'source_file' type mismatch. expected: typing.List[str], actual: "
                                         "<class 'list'>, value: [1]")
        data["custom_parse_file"][0]["source_file"] = ["1"]
        actl_val = CustomFileInfo(file_path_glob="test_custom/*.log", log_time_format="%H:%M:%S,%f", source_file=["1"])
        self.assertEqual(ConfigInfo.from_dict(data, True), ConfigInfo(custom_parse_file=[actl_val]))

    def test_timezone_config(self):
        data = {
            "timezone_config" : {
                "lcne" : True,
                "mindie": True
            }
        }
        config_info = ConfigInfo.from_dict(data, False)
        self.assertTrue(config_info.timezone_config.lcne)
        self.assertTrue(config_info.timezone_config.mindie)

    def test_No_timezone_config(self):
        data = {
            "timezone_config": {
            }
        }
        config_info = ConfigInfo.from_dict(data, False)
        self.assertFalse(config_info.timezone_config.lcne)
        self.assertFalse(config_info.timezone_config.mindie)