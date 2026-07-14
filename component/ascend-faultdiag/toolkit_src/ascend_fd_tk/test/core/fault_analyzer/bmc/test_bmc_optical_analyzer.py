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

import unittest

from ascend_fd_tk.core.fault_analyzer.bmc.bmc_optical_analyzer import BmcOpticalAnalyzer

from .helpers import BMC_ID, build_bmc_info, build_cluster_info, build_linkdown_log


class TestBmcOpticalAnalyzer(unittest.TestCase):
    def test_analyse_reports_latest_linkdown_record_and_lane_thresholds(self):
        older_log = build_linkdown_log(
            log_time="2026-06-01 08:00:00",
            optical_module_id="op-old",
            tx_power_current="5000 5000 0 0 0 0 0 0",
            rx_power_current="5000 5000 0 0 0 0 0 0",
            tx_bias_current="7000 7000 0 0 0 0 0 0",
            tx_los="0x0",
            rx_los="0x0",
            host_snr="25 25 0 0 0 0 0 0",
            media_snr="25 25 0 0 0 0 0 0",
        )
        latest_log = build_linkdown_log(
            log_time="2026-06-01 09:00:00",
            optical_module_id="op-new",
            tx_power_current="1000 30000 0 0 0 0 0 0",
            rx_power_current="1000 30000 0 0 0 0 0 0",
            tx_bias_current="5000 11000 0 0 0 0 0 0",
            tx_los="0x1",
            rx_los="0x2",
            host_snr="17 25 0 0 0 0 0 0",
            media_snr="17 25 0 0 0 0 0 0",
        )
        cluster_info = build_cluster_info(build_bmc_info(linkdown_logs=[older_log, latest_log]))

        results = BmcOpticalAnalyzer(cluster_info).analyse()

        self.assertEqual(len(results), 1)
        result = results[0]
        self.assertEqual(result.domain.bmc_id, BMC_ID)
        self.assertEqual(result.domain.npu_id, "0")
        self.assertEqual(result.domain.chip_phy_id, "0")
        self.assertEqual(result.suggestion, "请检查端口是否存在脏污")
        self.assertIn("记录时间2026-06-01 09:00:00", result.fault_info)
        self.assertNotIn("2026-06-01 08:00:00", result.fault_info)
        self.assertIn("lane0: tx power实际值：0.1，低于故障阈值：0.2，单位：mW", result.fault_info)
        self.assertIn("lane0: rx power实际值：0.1，低于故障阈值：0.1445，单位：mW", result.fault_info)
        self.assertIn("lane0: tx bias实际值：5.0，低于故障阈值：6，单位：mA", result.fault_info)
        self.assertIn("lane0: host snr实际值：17，低于故障阈值：18，单位：dB", result.fault_info)
        self.assertIn("lane0: media snr实际值：17，低于故障阈值：18，单位：dB", result.fault_info)
        self.assertIn("lane1: tx power实际值：3.0，高于故障阈值：2.5，单位：mW", result.fault_info)
        self.assertIn("lane1: rx power实际值：3.0，高于故障阈值：2.3，单位：mW", result.fault_info)
        self.assertIn("lane1: tx bias实际值：11.0，高于故障阈值：10，单位：mA", result.fault_info)
        self.assertIn("Tx los值0x1大于0", result.fault_info)
        self.assertIn("Rx los值0x2大于0", result.fault_info)

    def test_analyse_skips_bmc_without_linkdown_history(self):
        cluster_info = build_cluster_info(build_bmc_info())

        self.assertEqual(BmcOpticalAnalyzer(cluster_info).analyse(), [])

    def test_analyse_uses_detected_four_lane_linkdown_record(self):
        four_lane_log = build_linkdown_log(
            log_time="2026-06-01 09:30:00",
            optical_module_id="op-four-lane",
            tx_power_current="1000 5000 5000 5000 0 0 0 0",
            rx_power_current="1000 5000 5000 5000 0 0 0 0",
            tx_bias_current="5000 7000 7000 7000 0 0 0 0",
            tx_los="0x0",
            rx_los="0x0",
            host_snr="17 25 25 25 0 0 0 0",
            media_snr="17 25 25 25 0 0 0 0",
        )
        cluster_info = build_cluster_info(build_bmc_info(linkdown_logs=[four_lane_log]))

        results = BmcOpticalAnalyzer(cluster_info).analyse()

        self.assertEqual(len(results), 1)
        result = results[0]
        self.assertEqual(result.domain.npu_id, "0")
        self.assertEqual(result.domain.chip_phy_id, "0")
        self.assertIn("记录时间2026-06-01 09:30:00", result.fault_info)
        self.assertIn("lane0: tx power实际值：0.1，低于故障阈值：0.2，单位：mW", result.fault_info)
        self.assertIn("lane0: rx power实际值：0.1，低于故障阈值：0.1445，单位：mW", result.fault_info)
        self.assertIn("lane0: host snr实际值：17，低于故障阈值：18，单位：dB", result.fault_info)
        self.assertNotIn("lane4:", result.fault_info)
        self.assertNotIn("Tx los值", result.fault_info)
        self.assertNotIn("Rx los值", result.fault_info)


if __name__ == "__main__":
    unittest.main()
