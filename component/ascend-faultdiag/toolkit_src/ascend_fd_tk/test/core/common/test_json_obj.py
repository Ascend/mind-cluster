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

"""json_obj 单元测试：JsonObj 序列化/反序列化与类型校验。"""

import unittest
from typing import Dict, List

from ascend_fd_tk.core.common.json_obj import JsonObj


class _Child(JsonObj):
    def __init__(self, name: str, value: int = 0):
        self.name = name
        self.value = value


class _Parent(JsonObj):
    def __init__(
        self, title: str, child: _Child = None, children: List[_Child] = None, child_map: Dict[str, _Child] = None
    ):
        self.title = title
        self.child = child
        self.children = children or []
        self.child_map = child_map or {}


class _CamelParent(_Parent):
    @classmethod
    def _parse_to_py_key(cls):
        return True


class TestJsonObj(unittest.TestCase):
    def test_to_dict_roundtrip(self):
        """to_dict/from_dict 应往返无损。"""
        obj = _Parent("t", child=_Child("c", 1), children=[_Child("a")], child_map={"k": _Child("v", 2)})
        d = obj.to_dict()
        self.assertEqual(d["title"], "t")
        self.assertEqual(d["child"], {"name": "c", "value": 1})
        self.assertEqual(d["children"], [{"name": "a", "value": 0}])
        self.assertEqual(d["child_map"], {"k": {"name": "v", "value": 2}})
        restored = _Parent.from_dict(d)
        self.assertEqual(restored.to_dict(), d)

    def test_to_json_and_str(self):
        """to_json 与 __str__ 应输出 JSON 字符串（含空列表/字典字段）。"""
        obj = _Parent("t", child=_Child("c", 1))
        self.assertEqual(
            obj.to_json(), '{"title": "t", "child": {"name": "c", "value": 1}, "children": [], "child_map": {}}'
        )
        self.assertEqual(str(obj), obj.to_json())

    def test_from_json(self):
        """from_json 应从 JSON 字符串构建对象。"""
        obj = _Parent.from_json('{"title": "x", "child": {"name": "c", "value": 3}}')
        self.assertEqual(obj.title, "x")
        self.assertEqual(obj.child.value, 3)

    def test_missing_field_uses_default(self):
        """缺省字段应使用构造参数默认值。"""
        obj = _Parent.from_dict({"title": "t"})
        self.assertIsNone(obj.child)
        self.assertEqual(obj.children, [])
        self.assertEqual(obj.child_map, {})

    def test_check_parameter_type_mismatch(self):
        """check_parameter=True 时字段类型不匹配应抛 TypeError。"""
        with self.assertRaises(TypeError):
            _Parent.from_dict({"title": 123}, check_parameter=True)

    def test_list_item_type_mismatch(self):
        """List[JsonObj] 元素类型不匹配应抛 TypeError。"""
        with self.assertRaises(TypeError):
            _Parent.from_dict({"title": "t", "children": [{"name": 1}]}, check_parameter=True)

    def test_parse_to_py_key(self):
        """_parse_to_py_key 为 True 时应将顶层 camelCase 键转为小写键。"""
        obj = _CamelParent.from_dict({"Title": "t", "Child": {"name": "c", "value": 9}})
        self.assertEqual(obj.title, "t")
        self.assertEqual(obj.child.name, "c")
        self.assertEqual(obj.child.value, 9)

    def test_pickle_via_reduce(self):
        """__reduce__ 应支持 pickle 往返。"""
        import pickle

        obj = _Parent("t", child=_Child("c", 1))
        restored = pickle.loads(pickle.dumps(obj))
        self.assertEqual(restored.to_dict(), obj.to_dict())


if __name__ == "__main__":
    unittest.main()
