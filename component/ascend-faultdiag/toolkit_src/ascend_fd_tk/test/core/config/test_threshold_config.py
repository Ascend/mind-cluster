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

"""光模块类型阈值架构单测：{TYPE}_{METRIC} 命名约定 + 类型视图按类型选阈值

覆盖：
- get_optical_metric：类型阈值优先，无类型/未定义类型阈值时回退基础阈值
- OpticalThresholdView：属性访问等价 get_optical_metric
- 诊断方法（OpticalModuleInfo）按 optical_type 选择阈值
- threshold_config.json 覆盖 ODSP_/LPO_ 类型化阈值并经视图生效
"""

import json
import os
import shutil
import tempfile
import unittest

from ascend_fd_tk.core.config.threshold_config import OpticalThreshold800G, BaseThreshold
from ascend_fd_tk.core.config.threshold_loader import ThresholdConfigLoader
from ascend_fd_tk.core.model.optical_module import LanePowerInfo, OpticalModuleInfo
from ascend_fd_tk.core.model.threshold import Threshold, ThresholdStatus
from ascend_fd_tk.core.report.threshold_report import ThresholdConfig

CONFIG_FILE = "threshold_config.json"


class TestGetOpticalMetric(unittest.TestCase):
    """类型化阈值选择规则（{TYPE}_{METRIC} 命名约定）"""

    def test_typed_threshold_preferred(self):
        """指定类型且存在 {TYPE}_{METRIC} 阈值时优先返回"""
        self.assertIs(OpticalThreshold800G.get_optical_metric("TX_BIAS_MA", "LPO"), OpticalThreshold800G.LPO_TX_BIAS_MA)

    def test_odsp_falls_back_to_base(self):
        """ODSP 为默认类型，直接使用基础阈值（未定义 ODSP_ 前缀阈值）"""
        self.assertIs(OpticalThreshold800G.get_optical_metric("TX_BIAS_MA", "ODSP"), OpticalThreshold800G.TX_BIAS_MA)
        self.assertIs(
            OpticalThreshold800G.get_optical_metric("TX_POWER_DBM", "odsp"), OpticalThreshold800G.TX_POWER_DBM
        )

    def test_non_lpo_types_fall_back_to_base(self):
        """规则：optical_type 为空或不为 LPO，一律使用基础阈值（即 ODSP 阈值）"""
        # 未知/意外类型（采集侧脏数据）也回退基础阈值
        for unexpected_type in ("", " ", None, "UNKNOWN", "unknown", "810"):
            self.assertIs(
                OpticalThreshold800G.get_optical_metric("TX_BIAS_MA", unexpected_type),
                OpticalThreshold800G.TX_BIAS_MA,
                msg=f"optical_type={unexpected_type!r} 应回退基础阈值",
            )

    def test_type_case_insensitive(self):
        """类型名大小写不敏感（采集侧可能为小写）"""
        self.assertIs(OpticalThreshold800G.get_optical_metric("TX_BIAS_MA", "lpo"), OpticalThreshold800G.LPO_TX_BIAS_MA)
        self.assertIs(
            OpticalThreshold800G.get_optical_metric("TX_BIAS_MA", " Lpo "), OpticalThreshold800G.LPO_TX_BIAS_MA
        )

    def test_fallback_without_type(self):
        """类型为空（A3/交换机侧模块）回退基础阈值"""
        self.assertIs(OpticalThreshold800G.get_optical_metric("TX_BIAS_MA", ""), OpticalThreshold800G.TX_BIAS_MA)
        self.assertIs(OpticalThreshold800G.get_optical_metric("TX_BIAS_MA", None), OpticalThreshold800G.TX_BIAS_MA)

    def test_fallback_when_typed_not_defined(self):
        """该类型未定义 {TYPE}_{METRIC} 阈值时回退基础阈值（指标扩展无需全类型覆盖）"""
        self.assertIs(BaseThreshold.get_optical_metric("CDR_HOST_SNR_DB", "LPO"), OpticalThreshold800G.CDR_HOST_SNR_DB)

    def test_optical_view_attribute_access(self):
        """get_optical_view 返回的视图属性访问与 get_optical_metric 等价"""
        view = OpticalThreshold800G.get_optical_view("LPO")
        self.assertIs(view.TX_POWER_DBM, OpticalThreshold800G.LPO_TX_POWER_DBM)
        view = BaseThreshold.get_optical_view("")
        self.assertIs(view.TX_POWER_DBM, BaseThreshold.TX_POWER_DBM)


class TestTypedThresholdInDiagnosis(unittest.TestCase):
    """诊断方法按光模块类型选择阈值（类型化阈值放宽后判定结果不同）"""

    def test_abnormal_bias_by_optical_type(self):
        """bias 检查按类型选择阈值：LPO 放宽后不再异常，ODSP/无类型沿用基础阈值"""
        original = OpticalThreshold800G.LPO_TX_BIAS_MA
        OpticalThreshold800G.LPO_TX_BIAS_MA = Threshold(low_value_alarm="0.4", desc="lpo tx bias", unit="mA")
        try:
            module_lpo = OpticalModuleInfo(lane_power_infos=[LanePowerInfo("0", bias="0.5")], optical_type="LPO")
            module_odsp = OpticalModuleInfo(lane_power_infos=[LanePowerInfo("0", bias="0.5")], optical_type="ODSP")
            module_none = OpticalModuleInfo(lane_power_infos=[LanePowerInfo("0", bias="0.5")])
            self.assertEqual(module_lpo.get_abnormal_bias_infos(OpticalThreshold800G), "")
            self.assertIn("Lane0", module_odsp.get_abnormal_bias_infos(OpticalThreshold800G))
            self.assertIn("Lane0", module_none.get_abnormal_bias_infos(OpticalThreshold800G))
        finally:
            OpticalThreshold800G.LPO_TX_BIAS_MA = original

    def test_abnormal_snr_by_optical_type(self):
        """host/media SNR 检查：LPO 不支持 SNR（未定义 LPO_ 前缀 SNR 阈值），一律回退基础阈值"""
        self.assertIs(OpticalThreshold800G.get_optical_metric("HOST_SNR_DB", "LPO"), OpticalThreshold800G.HOST_SNR_DB)
        module_lpo = OpticalModuleInfo(lane_power_infos=[LanePowerInfo("0", host_snr="8.0")], optical_type="LPO")
        module_odsp = OpticalModuleInfo(lane_power_infos=[LanePowerInfo("0", host_snr="8.0")], optical_type="ODSP")
        self.assertIn("Lane0", module_lpo.get_abnormal_snr_infos(OpticalThreshold800G))
        self.assertIn("Lane0", module_odsp.get_abnormal_snr_infos(OpticalThreshold800G))


class TestTypedThresholdJsonOverride(unittest.TestCase):
    """threshold_config.json 覆盖类型化阈值（LPO_ 前缀配置键；ODSP 用基础阈值名覆盖）"""

    # pylint: disable=duplicate-code
    def setUp(self):
        self._tmp_dir = tempfile.mkdtemp()
        self._loader = ThresholdConfigLoader()

    def tearDown(self):
        # 恢复全部 Profile 为代码默认值，避免影响同进程内其他用例
        ThresholdConfigLoader().apply()
        shutil.rmtree(self._tmp_dir, ignore_errors=True)

    def _write_config(self, config):
        path = os.path.join(self._tmp_dir, CONFIG_FILE)
        with open(path, "w", encoding="utf-8") as f:
            json.dump(config, f, ensure_ascii=False)
        return path

    def test_json_override_typed_threshold(self):
        """JSON 覆盖 LPO_TX_BIAS_MA 后，LPO 视图生效且不影响 ODSP（基础）阈值"""
        self._write_config({"LPO_TX_BIAS_MA": {"threshold": {"high_value_alarm": "1"}}})
        self._loader.parse(self._tmp_dir)
        self._loader.apply()
        self.assertEqual(OpticalThreshold800G.LPO_TX_BIAS_MA.high_alarm_th, "1")
        view = OpticalThreshold800G.get_optical_view("LPO")
        self.assertEqual(view.TX_BIAS_MA.high_alarm_th, "1")
        # 未配置项保持默认（ODSP 用基础阈值）
        self.assertEqual(OpticalThreshold800G.TX_BIAS_MA.high_alarm_th, "13.00")
        self.assertEqual(BaseThreshold.TX_BIAS_MA.high_alarm_th, "10")

    def test_json_override_applies_to_diagnosis(self):
        """JSON 覆盖 LPO_TX_BIAS_MA 后，诊断判定随类型阈值变化"""
        self._write_config({"LPO_TX_BIAS_MA": {"threshold": {"high_value_alarm": "1"}}})
        self._loader.parse(self._tmp_dir)
        self._loader.apply()
        # LPO 阈值被覆盖为高阈值1，bias=7.5 判定异常；ODSP 沿用基础阈值（4.00~13.00），判定正常
        module_lpo = OpticalModuleInfo(lane_power_infos=[LanePowerInfo("0", bias="7.5")], optical_type="LPO")
        self.assertIn("Lane0", module_lpo.get_abnormal_bias_infos(OpticalThreshold800G))
        module_odsp = OpticalModuleInfo(lane_power_infos=[LanePowerInfo("0", bias="7.5")], optical_type="ODSP")
        self.assertEqual(module_odsp.get_abnormal_bias_infos(OpticalThreshold800G), "")

    def test_reset_restores_typed_default(self):
        """重复 apply 先重置再应用，类型化阈值覆盖不残留"""
        self._write_config({"LPO_TX_BIAS_MA": {"threshold": {"low_value_alarm": "1"}}})
        self._loader.parse(self._tmp_dir)
        self._loader.apply()
        self.assertEqual(OpticalThreshold800G.LPO_TX_BIAS_MA.low_alarm_th, "1")
        # 空配置重新应用后恢复代码默认值
        os.remove(os.path.join(self._tmp_dir, CONFIG_FILE))
        self._loader.parse(self._tmp_dir)
        self._loader.apply()
        self.assertEqual(OpticalThreshold800G.LPO_TX_BIAS_MA.low_alarm_th, "4.00")


class TestReportTypedThresholdConfig(unittest.TestCase):
    """report 侧 ThresholdConfig.for_optical_metric：按行数据 optical_type 动态选阈值"""

    class _Row:
        def __init__(self, optical_type):
            self.optical_type = optical_type
            self.host_tx_bias0 = "1.5"

    def test_lpo_row_uses_lpo_threshold(self):
        config = ThresholdConfig.for_optical_metric(
            OpticalThreshold800G, "TX_BIAS_MA", "host_tx_bias0", "主机侧TX Bias Lane 0"
        )
        original = OpticalThreshold800G.LPO_TX_BIAS_MA
        OpticalThreshold800G.LPO_TX_BIAS_MA = Threshold(low_value_alarm="1.8", desc="lpo tx bias", unit="mA")
        try:
            # LPO 行：bias=1.5 低于 LPO 低阈值1.8 判定异常
            status, th_value = config.value_checker(self._Row("LPO"), "1.5")
            self.assertEqual(status, ThresholdStatus.LOW_THRESHOLD_ALARM)
            self.assertEqual(th_value, "1.8")
        finally:
            OpticalThreshold800G.LPO_TX_BIAS_MA = original

    def test_non_lpo_row_falls_back_to_base(self):
        config = ThresholdConfig.for_optical_metric(
            OpticalThreshold800G, "TX_BIAS_MA", "host_tx_bias0", "主机侧TX Bias Lane 0"
        )
        # ODSP/空类型行回退基础阈值（4.00~13.00），bias=7.5 判定正常
        for optical_type in ("ODSP", ""):
            status, _ = config.value_checker(self._Row(optical_type), "7.5")
            self.assertEqual(status, ThresholdStatus.NORMAL, msg=f"optical_type={optical_type!r}")

    def test_base_threshold_kept_for_unit_display(self):
        """threshold 保留基础阈值，用于列名单位显示与构造校验"""
        config = ThresholdConfig.for_optical_metric(
            OpticalThreshold800G, "TX_BIAS_MA", "host_tx_bias0", "主机侧TX Bias Lane 0"
        )
        self.assertIs(config.threshold, OpticalThreshold800G.TX_BIAS_MA)
        self.assertEqual(config.threshold.unit, "mA")

    def test_peer_type_field(self):
        """type_field 支持对端字段（如对端交换机光模块类型）"""
        row = type("PeerRow", (), {"peer_optical_type": "LPO", "peer_tx_bias0": "1.5"})()
        config = ThresholdConfig.for_optical_metric(
            OpticalThreshold800G, "TX_BIAS_MA", "peer_tx_bias0", "对端TX Bias Lane 0", type_field="peer_optical_type"
        )
        original = OpticalThreshold800G.LPO_TX_BIAS_MA
        OpticalThreshold800G.LPO_TX_BIAS_MA = Threshold(low_value_alarm="1.8", desc="lpo tx bias", unit="mA")
        try:
            status, _ = config.value_checker(row, "1.5")
            self.assertEqual(status, ThresholdStatus.LOW_THRESHOLD_ALARM)
        finally:
            OpticalThreshold800G.LPO_TX_BIAS_MA = original


if __name__ == "__main__":
    unittest.main()
