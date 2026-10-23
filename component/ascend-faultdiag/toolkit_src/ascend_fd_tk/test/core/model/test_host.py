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

"""host 模型单元测试：光模块解析、端口解析、SDID 转换与链路状态统计。"""

import unittest

from ascend_fd_tk.core.config.threshold_config import BaseThreshold
from ascend_fd_tk.core.model.host import (
    CdrSnrInfo,
    HCCNDfxCfgInfo,
    HCCNLinkHistory,
    HCCNLinkStatInfo,
    HCCNOpticalInfo,
    HccnPortHccsInfo,
    SpodInfo,
)
from ascend_fd_tk.core.model.threshold import Threshold


class TestHCCNOpticalInfo(unittest.TestCase):
    def test_extract_numeric_value(self):
        self.assertEqual(HCCNOpticalInfo.extract_numeric_value("0.2239 mW"), "0.2239")
        self.assertEqual(HCCNOpticalInfo.extract_numeric_value("N/A"), "N/A")
        self.assertEqual(HCCNOpticalInfo.extract_numeric_value(12), 12)

    def test_is_low_tx_power(self):
        low = HCCNOpticalInfo(tx_power0="0.5", tx_power_low_thres="1.0")
        self.assertTrue(low.is_low_tx_power())
        normal = HCCNOpticalInfo(
            tx_power0="2.0", tx_power1="2.0", tx_power2="2.0", tx_power3="2.0", tx_power_low_thres="1.0"
        )
        self.assertFalse(normal.is_low_tx_power())

    def test_is_high_power_enable_and_present(self):
        info = HCCNOpticalInfo(high_power_enable_reg="enabled", present="present")
        self.assertTrue(info.is_high_power_enable())
        self.assertTrue(info.is_optical_present(BaseThreshold))


class TestHccnPortHccsInfo(unittest.TestCase):
    def test_trans_to_list(self):
        self.assertEqual(HccnPortHccsInfo.trans_to_list("[2 4 4 4 4 4]"), [2, 4, 4, 4, 4, 4])
        self.assertEqual(HccnPortHccsInfo.trans_to_list("[]"), [])
        self.assertEqual(HccnPortHccsInfo.trans_to_list(123), 123)


class TestHCCNDfxCfgInfo(unittest.TestCase):
    def test_is_tx_disable(self):
        self.assertFalse(HCCNDfxCfgInfo(tx_disable_status="0x0").is_tx_disable())
        self.assertTrue(HCCNDfxCfgInfo(tx_disable_status="0x1").is_tx_disable())
        self.assertTrue(HCCNDfxCfgInfo(tx_disable_status="").is_tx_disable())


class TestSpodInfo(unittest.TestCase):
    def test_sdid_convert_empty(self):
        self.assertEqual(SpodInfo.sdid_convert(""), ("", "", "", "", ""))

    def test_sdid_convert_known_value(self):
        self.assertEqual(SpodInfo.sdid_convert(0x01010101), ("4", "0", "1", "64", "1.1.1.1"))


class TestHCCNLinkStatInfo(unittest.TestCase):
    def test_is_long_time_down(self):
        current = "Fri Aug 29 10:00:00 2026"
        long_down = HCCNLinkStatInfo(
            current, link_history=[HCCNLinkHistory(time="Fri Aug 29 09:00:00 2026", link_status="LINK DOWN")]
        )
        self.assertTrue(long_down.is_long_time_down())
        short_down = HCCNLinkStatInfo(
            current, link_history=[HCCNLinkHistory(time="Fri Aug 29 09:59:50 2026", link_status="LINK DOWN")]
        )
        self.assertFalse(short_down.is_long_time_down())
        self.assertFalse(HCCNLinkStatInfo(current).is_long_time_down())

    def test_get_down_interval(self):
        stat = HCCNLinkStatInfo(
            "Fri Aug 29 10:30:00 2026",
            link_history=[
                HCCNLinkHistory(time="Fri Aug 29 10:00:00 2026", link_status="LINK UP"),
                HCCNLinkHistory(time="Fri Aug 29 10:05:00 2026", link_status="LINK DOWN"),
                HCCNLinkHistory(time="Fri Aug 29 10:10:00 2026", link_status="LINK UP"),
                HCCNLinkHistory(time="Fri Aug 29 10:20:00 2026", link_status="LINK DOWN"),
            ],
        )
        intervals = [abs(x) for x in stat.get_down_interval()]
        self.assertEqual(intervals, [300.0, 600.0])


class TestCdrSnrInfo(unittest.TestCase):
    def test_get_snr_abnormal_desc(self):
        th = Threshold(low_value_warn="4", desc="SNR", unit="dB")
        snr = CdrSnrInfo(cdr_host_snr_lane1="3", cdr_media_snr_lane2="3", cdr_host_snr_lane3="10")
        desc = snr.get_snr_abnormal_desc(th, th)
        self.assertIn("cdr_host_snr_lane1", desc)
        self.assertIn("cdr_media_snr_lane2", desc)
        self.assertNotIn("cdr_host_snr_lane3", desc)


if __name__ == "__main__":
    unittest.main()
