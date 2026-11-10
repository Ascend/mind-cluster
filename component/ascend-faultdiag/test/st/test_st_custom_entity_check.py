#!/usr/bin/env python3
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

from ascend_fd.pkg.customize.custom_entity.custom_check import EntityChecker

ENTITY_CODE = "Test_Code_1"
DEFAULT_ENTITY_CODE = "Default_Code_1"


def build_entity(attribute=None, rule=None, source_file="TrainLog", regex=None):
    entity = {
        "attribute": {
            "class": "Software",
            "component": "NPU",
            "module": "NPU",
            "cause_zh": "原因",
            "description_zh": "描述",
            "suggestion_zh": "建议",
        },
        "source_file": source_file,
        "rule": [] if rule is None else rule,
        "regex": {"in": [["Machine Check"]]} if regex is None else regex,
    }
    if attribute:
        entity["attribute"].update(attribute)
    return entity


class STCustomEntityCheck(unittest.TestCase):
    """
    ST for the custom fault entity value rules.
    Each test case maps to the excel case id in its docstring.
    """

    @staticmethod
    def check(entity, origin_entities=None):
        origin = {} if origin_entities is None else origin_entities
        return EntityChecker(origin, {ENTITY_CODE: entity}).check()

    def test_cause_zh_with_six_symbol_classes(self):
        """Ascend_Fd_Entity_Function_0009"""
        entity = build_entity(attribute={"cause_zh": "中文，。abc123!@# 空格"})
        self.assertTrue(self.check(entity))

    def test_cause_zh_longer_than_200(self):
        """Ascend_Fd_Entity_Function_0008"""
        entity = build_entity(attribute={"cause_zh": "中" * 201})
        self.assertFalse(self.check(entity))

    def test_class_is_only_space(self):
        """Ascend_Fd_Entity_Function_0005"""
        entity = build_entity(attribute={"class": " "})
        self.assertTrue(self.check(entity))

    def test_class_contains_chinese(self):
        """Ascend_Fd_Entity_Function_0004"""
        entity = build_entity(attribute={"class": "中文"})
        self.assertFalse(self.check(entity))

    def test_component_is_only_punctuation(self):
        """Ascend_Fd_Entity_Function_0006"""
        entity = build_entity(attribute={"component": "!@#$"})
        self.assertTrue(self.check(entity))

    def test_paragraph_fields_longer_than_2000(self):
        """Ascend_Fd_Entity_Function_0010"""
        long_paragraph = "A" * 2001
        for field in ("description_zh", "suggestion_zh", "error_case", "fixed_case"):
            with self.subTest(field=field):
                self.assertFalse(self.check(build_entity(attribute={field: long_paragraph})))

    def test_description_zh_is_only_newline(self):
        """Ascend_Fd_Entity_Function_0011"""
        entity = build_entity(attribute={"description_zh": "\n"})
        self.assertTrue(self.check(entity))

    def test_error_case_element_longer_than_200(self):
        """Ascend_Fd_Entity_Function_0013"""
        entity = build_entity(attribute={"error_case": ["A" * 201]})
        self.assertFalse(self.check(entity))

    def test_fixed_case_list_contains_rn(self):
        """Ascend_Fd_Entity_Function_0014"""
        entity = build_entity(attribute={"fixed_case": ["rn"]})
        self.assertTrue(self.check(entity))

    def test_module_with_four_symbol_classes(self):
        """Ascend_Fd_Entity_Function_0007"""
        entity = build_entity(attribute={"module": "abc123!@# "})
        self.assertTrue(self.check(entity))

    def test_suggestion_zh_with_seven_symbol_classes(self):
        """Ascend_Fd_Entity_Function_0012"""
        entity = build_entity(attribute={"suggestion_zh": "中文，。abc123!@# rn"})
        self.assertTrue(self.check(entity))

    def test_code_and_attribute_longer_than_50(self):
        """Ascend_Fd_Entity_Function_0001"""
        over_limit = "A" * 51
        for key in ("class", "component", "module"):
            with self.subTest(key=key):
                self.assertFalse(self.check(build_entity(attribute={key: over_limit})))
        for rule in ({"dst_code": over_limit}, {"dst_code": DEFAULT_ENTITY_CODE, "expression": "A" * 201}):
            with self.subTest(rule=rule):
                origin = {DEFAULT_ENTITY_CODE: {}}
                self.assertFalse(self.check(build_entity(rule=[rule]), origin))
        # The entity code itself is checked before the attribute validation.
        origin = {}
        self.assertFalse(EntityChecker(origin, {over_limit: build_entity()}).check())

    def test_code_equal_50(self):
        """Ascend_Fd_Entity_Function_0002"""
        code = "A" * 50
        self.assertTrue(EntityChecker({}, {code: build_entity()}).check())

    def test_dst_code_not_exist(self):
        """Ascend_Fd_Entity_Function_0016"""
        entity = build_entity(rule=[{"dst_code": "Not_Exist_Code"}])
        self.assertFalse(self.check(entity))

    def test_regex_in_mixed_list_and_string(self):
        """Ascend_Fd_Entity_Function_0024"""
        entity = build_entity(regex={"in": [["Machine Check"], "Another Check"]})
        self.assertFalse(self.check(entity))

    def test_regex_in_element_longer_than_200(self):
        """Ascend_Fd_Entity_Function_0025"""
        entity = build_entity(regex={"in": [["A" * 201]]})
        self.assertFalse(self.check(entity))

    def test_regex_in_with_six_symbol_classes(self):
        """Ascend_Fd_Entity_Function_0026"""
        entity = build_entity(regex={"in": [["中文，。abc123!@# "]]})
        self.assertTrue(self.check(entity))

    def test_dst_code_length_one(self):
        """Ascend_Fd_Entity_Function_0003"""
        entity = build_entity(rule=[{"dst_code": "A"}])
        self.assertTrue(self.check(entity, {"A": {}}))

    def test_rule_expression_longer_than_200(self):
        """Ascend_Fd_Entity_Function_0015"""
        entity = build_entity(rule=[{"dst_code": DEFAULT_ENTITY_CODE, "expression": "A" * 201}])
        self.assertFalse(self.check(entity, {DEFAULT_ENTITY_CODE: {}}))

    def test_source_file_not_in_builtin_list(self):
        """Ascend_Fd_Entity_Function_0017"""
        # A custom source file name is only checked by the character rule, so it is accepted here.
        entity = build_entity(source_file="Custom_Log_Name")
        self.assertTrue(self.check(entity))


if __name__ == "__main__":
    unittest.main()
