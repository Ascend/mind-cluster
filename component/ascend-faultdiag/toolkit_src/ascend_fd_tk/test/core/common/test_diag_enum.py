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

"""diag_enum 单元测试：get_enum 枚举查找工具函数。"""

import unittest

from ascend_fd_tk.core.common.diag_enum import DeviceStatus, DeviceType, NpuType, get_enum


class TestGetEnum(unittest.TestCase):
    def test_get_by_name(self):
        """按枚举成员名称查找（区分大小写）。"""
        self.assertIs(get_enum(NpuType, name="A3"), NpuType.A3)

    def test_get_by_value(self):
        """按枚举成员值查找。"""
        self.assertIs(get_enum(NpuType, value="A5"), NpuType.A5)
        self.assertIs(get_enum(DeviceType, value="交换机"), DeviceType.SWITCH)

    def test_name_precedence_over_value(self):
        """name 非空时优先按 name 查找，value 被忽略。"""
        self.assertIs(get_enum(NpuType, name="A2", value="A5"), NpuType.A2)

    def test_not_found_returns_none(self):
        """未找到时返回 None。"""
        self.assertIsNone(get_enum(NpuType, name="A9"))
        self.assertIsNone(get_enum(NpuType, value="A9"))
        self.assertIsNone(get_enum(NpuType, name="", value=""))

    def test_case_insensitive(self):
        """不区分大小写时可按小写 name/value 命中。"""
        self.assertIs(get_enum(NpuType, name="a3", case_sensitive=False), NpuType.A3)
        self.assertIs(get_enum(DeviceType, value="交换机", case_sensitive=False), DeviceType.SWITCH)

    def test_case_sensitive_default(self):
        """区分大小写时小写 name 无法命中。"""
        self.assertIsNone(get_enum(NpuType, name="a3"))

    def test_invalid_enum_cls_returns_none(self):
        """非枚举类入参返回 None。"""
        self.assertIsNone(get_enum(dict, name="x"))
        self.assertIsNone(get_enum("NpuType", name="A3"))


class TestEnumStr(unittest.TestCase):
    def test_device_status_str(self):
        """DeviceStatus 的 __str__ 输出小写名称。"""
        self.assertEqual(str(DeviceStatus.NORMAL), "normal")

    def test_device_type_str(self):
        """DeviceType 的 __str__ 输出中文值。"""
        self.assertEqual(str(DeviceType.SWITCH), "交换机")


if __name__ == "__main__":
    unittest.main()
