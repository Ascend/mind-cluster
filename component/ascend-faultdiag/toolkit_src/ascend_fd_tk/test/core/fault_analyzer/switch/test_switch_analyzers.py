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
from types import SimpleNamespace
from unittest.mock import MagicMock

from ascend_fd_tk.core.common.constants import BIT_ERROR_RATE_LIMIT
from ascend_fd_tk.core.fault_analyzer.switch.bit_err_rate_analyzer import BitErrRateAnalyzer
from ascend_fd_tk.core.fault_analyzer.switch.crc_err_rising_alarm_analyzer import CrcRisingCheckItem
from ascend_fd_tk.core.fault_analyzer.switch.lane_reduction_analyzer import LaneReductionAnalyzer
from ascend_fd_tk.core.fault_analyzer.switch.op_los_alarm_analyzer import OpticalInvalidAnalyzer
from ascend_fd_tk.core.fault_analyzer.switch.op_state_flag_diag_info_analyzer import OpStateFlagDiagInfoAnalyzer
from ascend_fd_tk.core.fault_analyzer.switch.switch_analyzer import SwitchAnalyzer


def _cluster_with_switch(switch_info):
    return SimpleNamespace(swis_info={switch_info.swi_id: switch_info})


class TestSwitchAnalyzers(unittest.TestCase):
    def test_bit_error_rate_reports_only_values_above_threshold(self):
        high = str(BIT_ERROR_RATE_LIMIT * 2)
        switch_info = SimpleNamespace(
            swi_id="10.0.0.1",
            slot_id="61",
            bit_error_rate=[
                SimpleNamespace(bit_err_rate=high, interface_name="100GE1/0/1"),
                SimpleNamespace(bit_err_rate="invalid", interface_name="100GE1/0/2"),
                SimpleNamespace(bit_err_rate=str(BIT_ERROR_RATE_LIMIT), interface_name="100GE1/0/3"),
            ],
        )

        results = BitErrRateAnalyzer(_cluster_with_switch(switch_info)).analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertEqual(results[0].domain.interface, "100GE1/0/1")
        self.assertIn(high, results[0].fault_info)

    def test_lane_reduction_parses_alarm_and_includes_peer(self):
        if_info = MagicMock()
        peer_info = MagicMock()
        peer_info.get_inspection_interface_info.return_value = "switch-02/100GE1/0/2"
        alarm = SimpleNamespace(
            alarm_id_int=LaneReductionAnalyzer._ERR_CODE,
            alarm_id="0xF10509",
            description="EntPhysicalName=100GE1/0/1,Reason=lane reduction",
        )
        switch_info = SimpleNamespace(
            swi_id="10.0.0.1",
            slot_id="61",
            interface_full_infos={"100GE1/0/1": if_info},
            active_alarm_info=[alarm],
        )
        cluster = _cluster_with_switch(switch_info)
        cluster.find_peer_swi_interface_info_by_if_info = MagicMock(return_value=(None, peer_info))

        results = LaneReductionAnalyzer(cluster).analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertIn("发生降lane", results[0].fault_info)
        self.assertIn("switch-02/100GE1/0/2", results[0].fault_info)

    def test_crc_rising_alarm_parses_statistics_and_domain(self):
        local_info = SimpleNamespace(device_id="10.0.0.1", interface="100GE1/0/1")
        if_info = MagicMock()
        if_info.get_inspection_interface_info.return_value = local_info
        alarm = SimpleNamespace(
            alarm_id_int=CrcRisingCheckItem._ERR_CODE,
            alarm_id="0x081300BC",
            description=(
                "hwIfMonitorCrcErrorStatistics=9, hwIfMonitorCrcErrorThreshold=5, "
                "hwIfMonitorCrcErrorInterval=60, InterfaceName=100GE1/0/1"
            ),
        )
        switch_info = SimpleNamespace(
            swi_id="10.0.0.1",
            slot_id="61",
            interface_full_infos={"100GE1/0/1": if_info},
            active_alarm_info=[alarm],
        )
        cluster = _cluster_with_switch(switch_info)
        cluster.find_peer_swi_interface_info_by_if_info = MagicMock(return_value=(None, None))

        results = CrcRisingCheckItem(cluster).analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertEqual(results[0].domain.interface, "100GE1/0/1")
        self.assertIn("统计次数9，阈值5", results[0].fault_info)

    def test_optical_los_alarm_reports_reason_and_skips_other_codes_or_bad_descriptions(self):
        alarm = SimpleNamespace(
            alarm_id_int=OpticalInvalidAnalyzer._LOS_ALARM_ERR_CODE,
            alarm_id="0x8130059",
            description="EntPhysicalName=100GE1/0/1,Detail=x,Reason=No signal)",
        )
        other = SimpleNamespace(alarm_id_int=0, alarm_id="0", description=alarm.description)
        malformed = SimpleNamespace(
            alarm_id_int=OpticalInvalidAnalyzer._LOS_ALARM_ERR_CODE,
            alarm_id="0x8130059",
            description="EntPhysicalName=100GE1/0/1,Detail=x",
        )
        switch_info = SimpleNamespace(swi_id="10.0.0.1", slot_id="61", active_alarm_info=[other, malformed, alarm])

        results = OpticalInvalidAnalyzer(_cluster_with_switch(switch_info)).analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertIn("No signal", results[0].fault_info)

    def test_state_flag_reports_abnormal_lane_and_module_state(self):
        op_model = SimpleNamespace(
            interface_name="100GE1/0/1",
            optical_id="opt-1",
            state_flag_diag_infos=[
                SimpleNamespace(items="Rx Flag", status="Normal|Alarm|-|Normal"),
                SimpleNamespace(items="Module State", status="LowPwr"),
                SimpleNamespace(items="Unrelated", status="Alarm"),
            ],
        )
        switch_info = SimpleNamespace(swi_id="10.0.0.1", slot_id="61", optical_models=[op_model])

        results = OpStateFlagDiagInfoAnalyzer(_cluster_with_switch(switch_info)).analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertIn("lane1", results[0].fault_info)
        self.assertIn("LowPwr", results[0].fault_info)

    def test_switch_analyzer_delegates_single_ended_checks_once(self):
        optical_info = SimpleNamespace(optical_id="opt-1")
        full_info = MagicMock()
        full_info.get_optical_module_info.return_value = optical_info
        switch_info = SimpleNamespace(
            swi_id="10.0.0.1",
            slot_id="61",
            name="switch-01",
            interface_mapping=[],
            interface_full_infos={"100GE1/0/1": full_info},
        )
        cluster = _cluster_with_switch(switch_info)
        cluster.get_threshold = MagicMock(return_value=SimpleNamespace(SNR_LANE_DIFF_DB=object()))
        analyzer = SwitchAnalyzer(cluster)
        analyzer.fault_check = MagicMock()
        analyzer.fault_check.power_analyze_single_ended.return_value = ["power"]
        analyzer.fault_check.snr_analyze_single_ended.return_value = ["snr"]
        analyzer.fault_check.bias_analyze_single_ended.return_value = ["bias"]

        results = analyzer.inter_switch_fault_analyze(switch_info)

        self.assertEqual(results, ["power", "snr", "bias"])
        analyzer.fault_check.power_analyze_single_ended.assert_called_once()
        analyzer.fault_check.snr_analyze_single_ended.assert_called_once()
        analyzer.fault_check.bias_analyze_single_ended.assert_called_once()
        self.assertEqual(analyzer.fault_check.power_analyze_single_ended.call_args.args[0].slot_id, "61")


if __name__ == "__main__":
    unittest.main()
