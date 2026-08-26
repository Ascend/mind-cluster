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

"""ThresholdConfigLoader 单元测试：阈值配置文件的解析校验与按代际应用

覆盖本次新增的配置能力：
- 嵌套 threshold 子对象格式解析
- 字段名校验（_THRESHOLD_FIELDS）
- 空值处理（None/空白串视为未配置，保持默认）
- 字符串字段类型校验（desc/unit/normal_value_* 非字符串忽略）
- desc/unit 长度上限校验
- 数值字段合法性校验（含 JSON 数字类型透传）
- 按代际（A3/A5）应用覆盖与重置
"""

import json
import os
import shutil
import tempfile
import unittest

from ascend_fd_tk.core.config.threshold_config import A5Threshold, BaseThreshold
from ascend_fd_tk.core.config.threshold_loader import (
    ThresholdConfigLoader,
    DESC_MAX_LEN,
    UNIT_MAX_LEN,
)

CONFIG_FILE = "threshold_config.json"


class TestValidateField(unittest.TestCase):
    """_validate_field 单字段校验（逐分支）"""

    def test_empty_value_returns_none(self):
        """None 与空白字符串均视为未配置，返回 None"""
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "desc", None))
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "desc", ""))
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "desc", "   "))

    def test_non_string_string_field_returns_none(self):
        """字符串字段（desc/unit/normal_value_*）传数字/bool/数组/对象时忽略"""
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "desc", 123))
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "unit", ["dB"]))
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "unit", {"x": 1}))
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "normal_value_alarm", True))
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "normal_value_alarm", 0))

    def test_overlong_desc_unit_returns_none(self):
        """desc/unit 超过长度上限时忽略，等于上限时合法"""
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "desc", "x" * (DESC_MAX_LEN + 1)))
        self.assertEqual(ThresholdConfigLoader._validate_field("K", "desc", "x" * DESC_MAX_LEN), "x" * DESC_MAX_LEN)
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "unit", "y" * (UNIT_MAX_LEN + 1)))
        self.assertEqual(ThresholdConfigLoader._validate_field("K", "unit", "y" * UNIT_MAX_LEN), "y" * UNIT_MAX_LEN)

    def test_invalid_numeric_returns_none(self):
        """数值字段无法转数值时忽略"""
        self.assertIsNone(ThresholdConfigLoader._validate_field("K", "low_value_alarm", "abc"))

    def test_valid_values_return_str(self):
        """合法字段返回字符串（数值字段支持 JSON 数字透传，强转字符串）"""
        self.assertEqual(ThresholdConfigLoader._validate_field("K", "desc", "host snr"), "host snr")
        self.assertEqual(ThresholdConfigLoader._validate_field("K", "unit", "dB"), "dB")
        self.assertEqual(ThresholdConfigLoader._validate_field("K", "normal_value_alarm", "UP"), "UP")
        self.assertEqual(ThresholdConfigLoader._validate_field("K", "low_value_alarm", "18"), "18")
        self.assertEqual(ThresholdConfigLoader._validate_field("K", "low_value_alarm", 18.5), "18.5")


class TestThresholdLoaderParse(unittest.TestCase):
    """配置文件解析与覆盖暂存"""

    def setUp(self):
        self._tmp_dir = tempfile.mkdtemp()
        self._loader = ThresholdConfigLoader()

    def tearDown(self):
        # 恢复全部 Profile 为代码默认值，避免影响同进程内其他用例
        ThresholdConfigLoader().apply("A3")
        shutil.rmtree(self._tmp_dir, ignore_errors=True)

    def _write_config(self, config):
        path = os.path.join(self._tmp_dir, CONFIG_FILE)
        with open(path, "w", encoding="utf-8") as f:
            json.dump(config, f, ensure_ascii=False)
        return path

    def test_config_file_not_found(self):
        """配置文件不存在：覆盖为空，应用后保持默认值"""
        empty_dir = tempfile.mkdtemp()
        try:
            self.assertEqual(self._loader.parse(empty_dir), {})
            self._loader.apply("A5")
            self.assertEqual(A5Threshold.HOST_SNR_DB.desc, "host snr")
            self.assertEqual(A5Threshold.HOST_SNR_DB.unit, "dB")
        finally:
            shutil.rmtree(empty_dir, ignore_errors=True)

    def test_invalid_json(self):
        """配置文件内容非法 JSON：覆盖为空，不抛异常"""
        path = os.path.join(self._tmp_dir, CONFIG_FILE)
        with open(path, "w", encoding="utf-8") as f:
            f.write("{not valid json")
        self.assertEqual(self._loader.parse(self._tmp_dir), {})

    def test_non_dict_config(self):
        """配置文件顶层不是对象：覆盖为空"""
        self._write_config(["HOST_SNR_DB"])
        self.assertEqual(self._loader.parse(self._tmp_dir), {})

    def test_config_value_not_object(self):
        """配置项的值不是对象：该项被忽略"""
        self._write_config({"HOST_SNR_DB": "18"})
        self.assertEqual(self._loader.parse(self._tmp_dir), {})

    def test_missing_threshold_object(self):
        """配置项缺少 threshold 子对象：该项被忽略"""
        self._write_config({"HOST_SNR_DB": {"low_value_alarm": "18"}})
        self.assertEqual(self._loader.parse(self._tmp_dir), {})

    def test_invalid_field_ignored(self):
        """未知字段名被忽略，合法字段保留"""
        self._write_config(
            {"HOST_SNR_DB": {"threshold": {"unknown_field": "x", "low_value_alarm": "18", "desc": "ok"}}}
        )
        overrides = self._loader.parse(self._tmp_dir)
        self.assertEqual(overrides, {"HOST_SNR_DB": {"low_value_alarm": "18", "desc": "ok"}})

    def test_empty_value_keeps_default(self):
        """空值（null/空白串）视为未配置，覆盖为空"""
        self._write_config({"HOST_SNR_DB": {"threshold": {"desc": "", "low_value_alarm": None}}})
        self.assertEqual(self._loader.parse(self._tmp_dir), {})

    def test_non_string_desc_unit_ignored(self):
        """非字符串 desc/unit 被忽略，仅保留合法字段"""
        self._write_config({"HOST_SNR_DB": {"threshold": {"desc": 123, "unit": ["dB"], "low_value_alarm": "18"}}})
        overrides = self._loader.parse(self._tmp_dir)
        self.assertEqual(overrides, {"HOST_SNR_DB": {"low_value_alarm": "18"}})

    def test_overlong_desc_unit_ignored(self):
        """超长 desc/unit 被忽略，仅保留合法字段"""
        self._write_config(
            {
                "HOST_SNR_DB": {
                    "threshold": {
                        "desc": "x" * (DESC_MAX_LEN + 1),
                        "unit": "y" * (UNIT_MAX_LEN + 1),
                        "low_value_alarm": "18",
                    }
                }
            }
        )
        overrides = self._loader.parse(self._tmp_dir)
        self.assertEqual(overrides, {"HOST_SNR_DB": {"low_value_alarm": "18"}})

    def test_invalid_numeric_ignored(self):
        """非法数值被忽略，覆盖为空"""
        self._write_config({"NIC_TX_POWER_DBM": {"threshold": {"low_value_alarm": "abc"}}})
        self.assertEqual(self._loader.parse(self._tmp_dir), {})


class TestThresholdLoaderApply(unittest.TestCase):
    """覆盖应用与按代际行为"""

    def setUp(self):
        self._tmp_dir = tempfile.mkdtemp()
        self._loader = ThresholdConfigLoader()

    def tearDown(self):
        ThresholdConfigLoader().apply("A3")
        shutil.rmtree(self._tmp_dir, ignore_errors=True)

    def _write_config(self, config):
        path = os.path.join(self._tmp_dir, CONFIG_FILE)
        with open(path, "w", encoding="utf-8") as f:
            json.dump(config, f, ensure_ascii=False)
        return path

    def test_valid_overrides_applied(self):
        """合法 desc/unit/数值覆盖生效"""
        self._write_config(
            {"HOST_SNR_DB": {"threshold": {"low_value_alarm": "18", "desc": "host snr custom", "unit": "dB"}}}
        )
        self._loader.parse(self._tmp_dir)
        self._loader.apply("A5")
        self.assertEqual(A5Threshold.HOST_SNR_DB.low_alarm_th, "18")
        self.assertEqual(A5Threshold.HOST_SNR_DB.desc, "host snr custom")
        self.assertEqual(A5Threshold.HOST_SNR_DB.unit, "dB")
        # 未配置项保持默认
        self.assertEqual(A5Threshold.CDR_HOST_SNR_DB.desc, "cdr host snr")

    def test_apply_to_both_generations(self):
        """同一份配置分别覆盖 A3/A5 代际 Profile"""
        self._write_config({"TX_BIAS_MA": {"threshold": {"low_value_alarm": "6", "unit": "mA"}}})
        self._loader.parse(self._tmp_dir)
        self._loader.apply("A3")
        self.assertEqual(BaseThreshold.TX_BIAS_MA.low_alarm_th, "6")
        self._loader.apply("A5")
        self.assertEqual(A5Threshold.TX_BIAS_MA.low_alarm_th, "6")

    def test_numeric_json_number_accepted(self):
        """数值字段以 JSON 数字配置时正常应用"""
        self._write_config({"RX_POWER_DBM": {"threshold": {"low_value_alarm": -10.0}}})
        self._loader.parse(self._tmp_dir)
        self._loader.apply("A5")
        self.assertEqual(A5Threshold.RX_POWER_DBM.low_alarm_th, "-10.0")
        self.assertEqual(A5Threshold.RX_POWER_DBM._low_alarm_th_f, -10.0)

    def test_invalid_key_ignored_at_apply(self):
        """配置键不是合法阈值名时应用阶段忽略"""
        default = A5Threshold.HOST_SNR_DB.low_alarm_th
        self._write_config({"NOT_A_THRESHOLD": {"threshold": {"low_value_alarm": "1"}}})
        self._loader.parse(self._tmp_dir)
        self._loader.apply("A5")  # 不抛异常，阈值保持不变
        self.assertEqual(A5Threshold.HOST_SNR_DB.low_alarm_th, default)

    def test_apply_resets_previous_overrides(self):
        """重复 apply 先重置再应用，旧覆盖不残留"""
        self._write_config({"HOST_SNR_DB": {"threshold": {"low_value_alarm": "18"}}})
        self._loader.parse(self._tmp_dir)
        self._loader.apply("A5")
        self.assertEqual(A5Threshold.HOST_SNR_DB.low_alarm_th, "18")
        self._write_config({"HOST_SNR_DB": {"threshold": {"low_value_alarm": "17"}}})
        self._loader.parse(self._tmp_dir)
        self._loader.apply("A5")
        self.assertEqual(A5Threshold.HOST_SNR_DB.low_alarm_th, "17")
        # 未再配置的字段恢复默认
        self.assertEqual(A5Threshold.HOST_SNR_DB.desc, "host snr")


if __name__ == "__main__":
    unittest.main()
