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

"""bmc 模型单元测试：微瓦转毫瓦、光模块信息构建与 lane 数分析。"""

import unittest

from ascend_fd_tk.core.common.diag_enum import PowerUnitType
from ascend_fd_tk.core.model.bmc import BmcInfo, BmcNpuInfo, LinkDownOpticalModuleHistoryLog


class TestBmcNpuInfo(unittest.TestCase):
    def test_u_to_m(self):
        self.assertEqual(BmcNpuInfo._u_to_m("1234"), "1.23")
        self.assertEqual(BmcNpuInfo._u_to_m("1234", radio=0.1), "0.12")
        self.assertEqual(BmcNpuInfo._u_to_m("abc"), "")

    def test_get_optical_module_info(self):
        history_log = LinkDownOpticalModuleHistoryLog(
            log_time="2026-08-29 10:00:00",
            location="NPU3",
            optical_module_id="op-1",
            tx_power_current="1000 2000 3000 4000",
            rx_power_current="100 200 300 400",
            tx_bias_current="5 6 7 8",
            tx_los="0",
            rx_los="1",
            host_snr="30 29 28 27",
            media_snr="25 24 23 22",
        )
        npu_info = BmcNpuInfo(lane_len=4, link_down_optical_module_history_log=history_log)
        module_info = npu_info.get_optical_module_info()
        self.assertIsNotNone(module_info)
        self.assertEqual(len(module_info.lane_power_infos), 4)
        lane0 = module_info.lane_power_infos[0]
        self.assertEqual(lane0.tx_power, "0.1")
        self.assertEqual(lane0.rx_power, "0.01")
        self.assertEqual(lane0.bias, "0.01")
        self.assertEqual(lane0.power_unit_type, PowerUnitType.MW)
        self.assertEqual(module_info.op_id, "op-1")

    def test_get_optical_module_info_none(self):
        self.assertIsNone(BmcNpuInfo().get_optical_module_info())


class TestBmcInfo(unittest.TestCase):
    def test_analyse_lane_len(self):
        bmc_info = BmcInfo(
            "bmc-1",
            "sn-1",
            link_down_optical_module_history_logs=[
                LinkDownOpticalModuleHistoryLog(tx_power_current="1000 2000 3000", rx_power_current="100 200 0")
            ],
        )
        self.assertEqual(bmc_info._analyse_lane_len(), 2)

    def test_get_bmc_npu_infos(self):
        bmc_info = BmcInfo(
            "bmc-1",
            "sn-1",
            link_down_optical_module_history_logs=[
                LinkDownOpticalModuleHistoryLog(
                    log_time="2026-08-29 10:00:00", location="NPU1", tx_power_current="1000 2000"
                ),
                LinkDownOpticalModuleHistoryLog(
                    log_time="2026-08-29 09:00:00", location="NPU1", tx_power_current="100 200"
                ),
            ],
        )
        npu_infos = bmc_info.get_bmc_npu_infos()
        self.assertEqual(len(npu_infos), 1)
        # location NPU1 -> chip_phy_id 0，取最近一条记录
        self.assertEqual(npu_infos[0].chip_phy_id, "0")
        self.assertEqual(npu_infos[0].link_down_optical_module_history_log.log_time, "2026-08-29 10:00:00")


if __name__ == "__main__":
    unittest.main()
