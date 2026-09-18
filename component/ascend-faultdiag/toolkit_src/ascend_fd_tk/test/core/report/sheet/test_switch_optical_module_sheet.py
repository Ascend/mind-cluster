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

"""SwitchOpticalModuleSheetGenerator 单元测试：LPO/ODSP 混插时按各自类型选阈值。"""

import unittest
from types import SimpleNamespace

from ascend_fd_tk.core.config.threshold_config import A5Threshold
from ascend_fd_tk.core.model.switch import (
    CommonInfo,
    DeviceInterface,
    InterfaceFullInfo,
    InterfaceMapping,
    OpticalModelBaseInfo,
    SwiOpticalModel,
    TransceiverInfo,
)
from ascend_fd_tk.core.model.threshold import ThresholdStatus
from ascend_fd_tk.core.report.sheet.switch_optical_module_sheet import SwitchOpticalModuleSheetGenerator


def _make_full_info(interface, transceiver_type, tx_power):
    base_info = [
        OpticalModelBaseInfo(items="TxPower[dBm](Lane0)", value=tx_power),
        OpticalModelBaseInfo(items="RxPower[dBm](Lane0)", value="-3.0"),
    ]
    return InterfaceFullInfo(
        interface=interface,
        transceiver_info=TransceiverInfo(interface=interface, common_information=CommonInfo(transceiver_type)),
        swi_optical_model=SwiOpticalModel(interface_name=interface, base_info=base_info),
    )


def _make_switch(name, iface, peer_name, peer_iface, transceiver_type, tx_power):
    return SimpleNamespace(
        name=name,
        swi_id="10.0.0.1",
        sn="SN1",
        slot_id="61",
        room_name="room1",
        cabinet_id="cab1",
        generation="A5",
        interface_mapping=[
            InterfaceMapping(local_interface_name=iface, remote_device_interface=DeviceInterface(peer_name, peer_iface))
        ],
        interface_full_infos={iface: _make_full_info(iface, transceiver_type, tx_power)},
    )


class TestSwitchOpticalModuleSheetMixedTypes(unittest.TestCase):
    def setUp(self):
        # sw1 端口 LPO 模块，sw2 端口 ODSP 模块，TX 均为 -6.0：
        # LPO_TX_POWER_DBM low_alarm=-5.70 → 本端应 ALARM；基础阈值 low_warn 高于 -6.0 → 对端应 WARN
        sw1 = _make_switch("sw1", "100GE1/0/1", "sw2", "100GE1/0/2", "400G_LPO", "-6.0")
        sw2 = _make_switch("sw2", "100GE1/0/2", "sw1", "100GE1/0/1", "400G_ODSP", "-6.0")
        self.cluster_info = SimpleNamespace(
            swis_info={"sw1": sw1, "sw2": sw2},
            get_threshold=lambda: A5Threshold,
        )
        self.gen = SwitchOpticalModuleSheetGenerator(self.cluster_info)
        self.rows = self.gen._collect_optical_module_data()
        self.configs = {cfg.field_name: cfg for cfg in self.gen._create_threshold_configs()}

    def test_row_optical_type_collected(self):
        """行数据的本端/对端光模块类型正确收集（来自 transceiver_type 后缀推导）。"""
        self.assertEqual(len(self.rows), 1)
        row = self.rows[0]
        self.assertEqual(row.local_optical_type, "LPO")
        self.assertEqual(row.peer_optical_type, "ODSP")

    def test_local_lpo_row_uses_lpo_threshold(self):
        """本端 LPO 列按 LPO 阈值判定（-6.0 < LPO low_alarm -5.70 → ALARM）。"""
        status, th_value = self.configs["local_tx_power0"].value_checker(self.rows[0], "-6.0")
        self.assertEqual(status, ThresholdStatus.LOW_THRESHOLD_ALARM)
        self.assertEqual(th_value, A5Threshold.LPO_TX_POWER_DBM.low_alarm_th)

    def test_peer_odsp_row_uses_base_threshold(self):
        """对端 ODSP 列按基础阈值判定（-6.0 介于 low_alarm 与 low_warn 之间 → WARN）。"""
        status, th_value = self.configs["peer_tx_power0"].value_checker(self.rows[0], "-6.0")
        self.assertEqual(status, ThresholdStatus.LOW_THRESHOLD_WARN)
        self.assertEqual(th_value, A5Threshold.TX_POWER_DBM.low_warn_th)


if __name__ == "__main__":
    unittest.main()
