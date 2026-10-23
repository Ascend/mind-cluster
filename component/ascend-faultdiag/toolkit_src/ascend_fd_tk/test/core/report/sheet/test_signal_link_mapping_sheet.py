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

"""SignalLinkMappingSheetGenerator 单元测试。"""

import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.root_cause.constants import (
    LINK_STATUS_NORMAL,
    RULE_L1_200G_HILINK,
)
from ascend_fd_tk.core.root_cause.model import HostToL1LinkData, L1ToL2LinkData
from ascend_fd_tk.core.report.sheet.signal_link_mapping_sheet import (
    SignalLinkMappingSheetGenerator,
)


def _make_gen(**kwargs):
    return SignalLinkMappingSheetGenerator(MagicMock(), excel_gen=MagicMock(), **kwargs)


class TestSignalLinkKeys(unittest.TestCase):
    def test_downstream_coarse_key(self):
        gen = _make_gen()
        d1 = HostToL1LinkData(host_id="h1", chip_phy_id="phy0", l1_switch_id="s1")
        d2 = HostToL1LinkData(host_id="h1", npu_id="2", l1_switch_id="s1")
        self.assertEqual(gen._downstream_coarse_key(d1), ("h1", "phy0", "s1"))
        self.assertEqual(gen._downstream_coarse_key(d2), ("h1", "cpu_2", "s1"))

    def test_upstream_group_key(self):
        gen = _make_gen()
        data = L1ToL2LinkData(l1_switch_id="s1", l1_switch_chip_id="ch0")
        self.assertEqual(gen._upstream_group_key(data), ("s1", "ch0"))


class TestSignalLinkMergeRanges(unittest.TestCase):
    def test_downstream_merge_ranges_two_level(self):
        gen = _make_gen()
        data_list = [
            HostToL1LinkData(host_id="h1", chip_phy_id="p0", l1_switch_id="s1", l1_switch_chip_id="c0"),
            HostToL1LinkData(host_id="h1", chip_phy_id="p0", l1_switch_id="s1", l1_switch_chip_id="c0"),
            HostToL1LinkData(host_id="h1", chip_phy_id="p0", l1_switch_id="s1", l1_switch_chip_id="c1"),
        ]
        ranges = gen._compute_downstream_merge_ranges(data_list)
        self.assertTrue(any(r[2] > r[0] for r in ranges))
        self.assertIn((2, 10, 3, 10), ranges)

    def test_downstream_merge_ranges_empty(self):
        self.assertEqual(_make_gen()._compute_downstream_merge_ranges([]), [])

    def test_upstream_merge_ranges(self):
        gen = _make_gen()
        data_list = [
            L1ToL2LinkData(l1_switch_id="s1", l1_switch_chip_id="c0", l1_interface="400G-1"),
            L1ToL2LinkData(l1_switch_id="s1", l1_switch_chip_id="c0", l1_interface="400G-2"),
            L1ToL2LinkData(l1_switch_id="s2", l1_switch_chip_id="c0", l1_interface="400G-1"),
        ]
        ranges = gen._compute_upstream_merge_ranges(data_list)
        self.assertTrue(any(r[2] > r[0] for r in ranges))
        self.assertIn((2, 1, 3, 1), ranges)


class TestSignalLinkGenerate(unittest.TestCase):
    def test_no_root_cause_filter_returns(self):
        gen = SignalLinkMappingSheetGenerator(MagicMock(), excel_gen=MagicMock(), root_cause_filter=None)
        with patch("ascend_fd_tk.core.report.sheet.signal_link_mapping_sheet.generate_threshold_excel") as mock_gen:
            gen.generate_sheet()
            mock_gen.assert_not_called()

    def test_generate_both_sheets(self):
        rcf = MagicMock()
        rcf.host_to_l1_links = [HostToL1LinkData(host_id="h1", l1_switch_chip_id="c0")]
        rcf.l1_to_l2_links = [L1ToL2LinkData(l1_switch_id="s1", l1_switch_chip_id="c0")]
        gen = SignalLinkMappingSheetGenerator(MagicMock(), excel_gen=MagicMock(), root_cause_filter=rcf)
        with patch("ascend_fd_tk.core.report.sheet.signal_link_mapping_sheet.generate_threshold_excel") as mock_gen:
            gen.generate_sheet()
            sheets = mock_gen.call_args.kwargs["sheets"]
            self.assertEqual(len(sheets), 2)
            self.assertEqual(sheets[0].sheet_name, "NPU&CPU到L1链路分析")
            self.assertEqual(sheets[1].sheet_name, "L1到L2链路分析")

    def test_link_filter_filters_rows(self):
        rcf = MagicMock()
        rcf.host_to_l1_links = [
            HostToL1LinkData(host_id="h1", l1_switch_chip_id="c0"),
            HostToL1LinkData(host_id="h2", l1_switch_chip_id="c0"),
        ]
        rcf.l1_to_l2_links = []
        gen = SignalLinkMappingSheetGenerator(
            MagicMock(), excel_gen=MagicMock(), root_cause_filter=rcf, link_filter=lambda d: d.host_id == "h1"
        )
        with patch("ascend_fd_tk.core.report.sheet.signal_link_mapping_sheet.generate_threshold_excel") as mock_gen:
            gen.generate_sheet()
            sheets = mock_gen.call_args.kwargs["sheets"]
            self.assertEqual(len(sheets), 1)
            self.assertEqual(len(sheets[0].data_list), 1)
            self.assertEqual(sheets[0].data_list[0].host_id, "h1")

    def test_rule_config_checker(self):
        gen = _make_gen()
        config = gen._rule_config("l1_hilink_snr", RULE_L1_200G_HILINK)
        triggered = HostToL1LinkData(triggered_rules=[RULE_L1_200G_HILINK])
        normal = HostToL1LinkData(triggered_rules=[])
        status_alarm, _ = config.value_checker(triggered, "10")
        status_normal, _ = config.value_checker(normal, "10")
        self.assertEqual(status_alarm.name, "NOT_EQUAL_THRESHOLD_ALARM")
        self.assertEqual(status_normal.name, "NORMAL")

    def test_link_status_threshold_config(self):
        gen = _make_gen()
        configs = gen._create_downstream_threshold_configs()
        link_status_config = next(c for c in configs if c.field_name == "link_status")
        status, _ = link_status_config.threshold.check_value("异常")
        self.assertEqual(status.name, "NOT_EQUAL_THRESHOLD_ALARM")
        status_normal, _ = link_status_config.threshold.check_value(LINK_STATUS_NORMAL)
        self.assertEqual(status_normal.name, "NORMAL")


if __name__ == "__main__":
    unittest.main()
