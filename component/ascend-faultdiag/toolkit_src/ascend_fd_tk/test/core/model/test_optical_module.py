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

"""光模块数据模型单元测试：功率单位转换与异常项统计。"""

import unittest

from ascend_fd_tk.core.common.diag_enum import PowerUnitType
from ascend_fd_tk.core.config.threshold_config import BaseThreshold
from ascend_fd_tk.core.model.optical_module import LanePowerInfo, OpticalModuleInfo


class TestLanePowerInfo(unittest.TestCase):
    def test_power_dbm_conversion(self):
        lane_mw = LanePowerInfo(tx_power="1", rx_power="2", power_unit_type=PowerUnitType.MW)
        self.assertEqual(lane_mw.tx_power_dbm, "0.0")
        self.assertNotEqual(lane_mw.rx_power_dbm, "2")
        lane_dbm = LanePowerInfo(tx_power="-5.5", rx_power="-6.5", power_unit_type=PowerUnitType.DBM)
        self.assertEqual(lane_dbm.tx_power_dbm, "-5.5")
        self.assertEqual(lane_dbm.rx_power_dbm, "-6.5")


class TestOpticalModuleInfo(unittest.TestCase):
    @staticmethod
    def _make_module(lanes):
        return OpticalModuleInfo(
            lane_power_infos=[
                LanePowerInfo(
                    lane_id=lane_id,
                    tx_power=tx,
                    rx_power=rx,
                    host_snr=host_snr,
                    media_snr=media_snr,
                    bias=bias,
                    power_unit_type=unit,
                )
                for lane_id, tx, rx, host_snr, media_snr, bias, unit in lanes
            ]
        )

    def test_get_lane_diff_desc(self):
        module = self._make_module(
            [
                ("0", "1", "1", "30", "", "", PowerUnitType.MW),
                ("1", "1", "1", "20", "", "", PowerUnitType.MW),
            ]
        )
        result = module.get_lane_diff_desc(BaseThreshold)
        self.assertIn("光模块SNR Lane间差值异常", result)
        self.assertIn("Lane0", result)
        module_normal = self._make_module(
            [
                ("0", "1", "1", "30", "", "", PowerUnitType.MW),
                ("1", "1", "1", "30", "", "", PowerUnitType.MW),
            ]
        )
        self.assertEqual(module_normal.get_lane_diff_desc(BaseThreshold), "")

    def test_get_abnormal_snr_infos(self):
        module = self._make_module(
            [
                ("0", "1", "1", "10", "", "", PowerUnitType.MW),
                ("1", "1", "1", "30", "25", "", PowerUnitType.MW),
            ]
        )
        result = module.get_abnormal_snr_infos(BaseThreshold)
        self.assertIn("Lane0", result)
        self.assertNotIn("Lane1", result)
        module_normal = self._make_module(
            [
                ("0", "1", "1", "30", "25", "", PowerUnitType.MW),
            ]
        )
        self.assertEqual(module_normal.get_abnormal_snr_infos(BaseThreshold), "")

    def test_get_abnormal_bias_infos(self):
        module = self._make_module(
            [
                ("0", "1", "1", "", "", "15", PowerUnitType.MW),
                ("1", "1", "1", "", "", "8", PowerUnitType.MW),
            ]
        )
        result = module.get_abnormal_bias_infos(BaseThreshold)
        self.assertIn("Lane0", result)
        self.assertNotIn("Lane1", result)

    def test_get_abnormal_power_infos(self):
        module = self._make_module(
            [
                ("0", "5", "0.01", "", "", "", PowerUnitType.MW),
                ("1", "1", "1", "", "", "", PowerUnitType.MW),
            ]
        )
        rx_list, tx_list = module.get_abnormal_power_infos(BaseThreshold)
        self.assertIn("Lane0", rx_list[0])
        self.assertIn("Lane0", tx_list[0])
        self.assertEqual(len(rx_list), 1)
        self.assertEqual(len(tx_list), 1)
        module_dbm = self._make_module(
            [
                ("0", "-20", "30", "", "", "", PowerUnitType.DBM),
            ]
        )
        rx_list, tx_list = module_dbm.get_abnormal_power_infos(BaseThreshold())
        self.assertEqual(len(rx_list), 1)
        self.assertEqual(len(tx_list), 1)

    def test_get_abnormal_txrx_infos(self):
        self.assertIsNone(OpticalModuleInfo(tx_los="1").get_abnormal_txrx_infos())
        self.assertIsNone(OpticalModuleInfo(tx_los="0").get_abnormal_txrx_infos())
        self.assertIsNone(OpticalModuleInfo().get_abnormal_txrx_infos())


if __name__ == "__main__":
    unittest.main()
