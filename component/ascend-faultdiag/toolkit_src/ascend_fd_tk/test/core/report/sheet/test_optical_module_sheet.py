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

"""HostToSwitchOpticalModuleSheetGenerator 单元测试。"""

import types
import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.config.threshold_config import OpticalThreshold800G
from ascend_fd_tk.core.report.sheet.optical_module_sheet import (
    HostToSwitchOpticalModuleSheetGenerator,
)


def _make_optical(lane_count=2):
    lanes = [
        types.SimpleNamespace(
            tx_power=str(i + 1),
            rx_power=str(i + 2),
            bias=str(i + 3),
            host_snr=str(i + 4),
            media_snr=str(i + 5),
        )
        for i in range(lane_count)
    ]
    return types.SimpleNamespace(sn="SN001", lane_power_infos=lanes)


def _make_chip(optical=None):
    optical = optical if optical is not None else _make_optical()
    return types.SimpleNamespace(
        npu_id="0",
        chip_id="1",
        chip_phy_id="2",
        npu_type="910B",
        speed="200G",
        duplex="full",
        net_health="Success",
        link_status="UP",
        hccn_lldp_info=None,
        hccn_optical_info=None,
        hccn_link_stat_info=None,
        get_optical_module_info=lambda: optical if isinstance(optical, list) else [optical],
    )


def _make_cluster(chips):
    cluster = MagicMock()
    cluster.hosts_info = {
        "host1": types.SimpleNamespace(
            hostname="h1", sn_num="SN1", room_name="r1", cabinet_id="c1", npu_chip_info=chips
        )
    }
    cluster.get_threshold.return_value = OpticalThreshold800G
    return cluster


class TestHostOpticalCollect(unittest.TestCase):
    def test_collect_optical_module_data(self):
        gen = HostToSwitchOpticalModuleSheetGenerator(_make_cluster({"0": _make_chip()}), excel_gen=MagicMock())
        data_list = gen._collect_optical_module_data()
        self.assertEqual(len(data_list), 1)
        data = data_list[0]
        self.assertEqual(data.host_id, "host1")
        self.assertEqual(data.host_name, "h1")
        self.assertEqual(data.npu_id, "0")
        self.assertEqual(data.host_tx_power0, "1")
        self.assertEqual(data.host_tx_power1, "2")
        self.assertEqual(data.host_tx_bias0, "3")

    def test_collect_multiple_optical_one_chip(self):
        chip = _make_chip([_make_optical(), _make_optical()])
        gen = HostToSwitchOpticalModuleSheetGenerator(_make_cluster({"0": chip}), excel_gen=MagicMock())
        data_list = gen._collect_optical_module_data()
        self.assertEqual(len(data_list), 2)

    def test_collect_skips_chip_without_optical(self):
        cluster = _make_cluster({"0": _make_chip()})
        chip = cluster.hosts_info["host1"].npu_chip_info["0"]
        chip.get_optical_module_info = lambda: []
        gen = HostToSwitchOpticalModuleSheetGenerator(cluster, excel_gen=MagicMock())
        self.assertEqual(gen._collect_optical_module_data(), [])

    def test_get_lane_info_missing_lanes_filled(self):
        gen = HostToSwitchOpticalModuleSheetGenerator(_make_cluster({}), excel_gen=MagicMock())
        optical = _make_optical(lane_count=2)
        lane_data = gen._get_lane_info(optical)
        self.assertEqual(lane_data["host_tx_power0"], "1")
        self.assertEqual(lane_data["host_tx_power1"], "2")
        self.assertEqual(lane_data["host_tx_power2"], "")
        self.assertEqual(lane_data["host_tx_power3"], "")
        self.assertEqual(lane_data["host_snr_lane3"], "")


class TestHostOpticalGenerate(unittest.TestCase):
    def test_generate_sheet_no_data_returns(self):
        cluster = _make_cluster({})
        gen = HostToSwitchOpticalModuleSheetGenerator(cluster, excel_gen=MagicMock())
        with patch("ascend_fd_tk.core.report.sheet.optical_module_sheet.generate_threshold_excel") as mock_gen:
            gen.generate_sheet()
            mock_gen.assert_not_called()

    def test_generate_sheet_with_data(self):
        gen = HostToSwitchOpticalModuleSheetGenerator(_make_cluster({"0": _make_chip()}), excel_gen=MagicMock())
        with patch("ascend_fd_tk.core.report.sheet.optical_module_sheet.generate_threshold_excel") as mock_gen:
            gen.generate_sheet()
            sheets = mock_gen.call_args.kwargs["sheets"]
            self.assertEqual(len(sheets), 1)
            self.assertEqual(sheets[0].sheet_name, "主机NPU<->RoCE交换机端口光模块信息")
            self.assertEqual(len(sheets[0].data_list), 1)

    def test_generate_sheet_link_filter(self):
        cluster = _make_cluster({"0": _make_chip(), "1": _make_chip()})
        cluster.hosts_info["host1"].npu_chip_info["1"].npu_id = "1"
        gen = HostToSwitchOpticalModuleSheetGenerator(
            cluster, excel_gen=MagicMock(), link_filter=lambda d: d.npu_id == "0"
        )
        with patch("ascend_fd_tk.core.report.sheet.optical_module_sheet.generate_threshold_excel") as mock_gen:
            gen.generate_sheet()
            sheets = mock_gen.call_args.kwargs["sheets"]
            self.assertEqual(len(sheets[0].data_list), 1)
            self.assertEqual(sheets[0].data_list[0].npu_id, "0")


if __name__ == "__main__":
    unittest.main()
