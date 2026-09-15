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
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.fault_analyzer.bmc.hccs_link_degraded_analyzer import HccsLinkDegradedAnalyzer
from ascend_fd_tk.core.fault_analyzer.hccs.hccs_rp_tx_analyzer import HCCSAnalyzer as HccsRpTxAnalyzer
from ascend_fd_tk.core.fault_analyzer.hccs.hccs_rx_analyzer import HCCSAnalyzer as HccsRxAnalyzer
from ascend_fd_tk.core.fault_analyzer.hccs.hccs_serdes_analyzer import HccsSerdesAnalyzer
from ascend_fd_tk.core.fault_analyzer.hccs.port_snr_dest_analyzer import PortSnrDestAnalyzer
from ascend_fd_tk.core.fault_analyzer.hccs.port_snr_src_analyzer import PortSnrSrcAnalyzer


class TestHccsAnalyzers(unittest.TestCase):
    def test_serdes_reports_cdr_and_power_faults(self):
        serdes = SimpleNamespace(
            chip_id="0", port_id="1", swi_port_id="100GE1/0/1", cdr_los="1", csr119_data="0x38000001"
        )
        switch_info = SimpleNamespace(
            swi_id="10.0.0.1", slot_id="61", hccs_info=SimpleNamespace(serdes_dump_info_list=[serdes])
        )
        cluster = SimpleNamespace(swis_info={switch_info.swi_id: switch_info})

        results = HccsSerdesAnalyzer(cluster).analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertIn("CDR失锁", results[0].fault_info)
        self.assertIn("0x38000001", results[0].fault_info)

    def test_serdes_skips_switch_without_hccs_and_normal_records(self):
        normal = SimpleNamespace(chip_id="0", port_id="1", swi_port_id="1", cdr_los="0", csr119_data="0x0")
        cluster = SimpleNamespace(
            swis_info={
                "a": SimpleNamespace(swi_id="a", slot_id="61", hccs_info=None),
                "b": SimpleNamespace(
                    swi_id="b", slot_id="61", hccs_info=SimpleNamespace(serdes_dump_info_list=[normal])
                ),
            }
        )
        self.assertEqual(HccsSerdesAnalyzer(cluster).analyse(), [])

    def test_source_port_snr_uses_xpu_threshold_and_reports_abnormal_lane(self):
        threshold = MagicMock()
        threshold.check_value_str.side_effect = lambda value: "低于阈值" if value == "10" else ""
        analyzer = PortSnrSrcAnalyzer.__new__(PortSnrSrcAnalyzer)
        analyzer.threshold = SimpleNamespace(SWITCH_PORT_SNR_LINE=threshold)
        analyzer.xpu_snr_limit_map = {"NPU": threshold}
        mapping = SimpleNamespace(swi_port="100GE1/0/1", xpu="NPU", xpu_id="3")
        analyzer.port_mapping_config_instance = MagicMock()
        analyzer.port_mapping_config_instance.find_port_mapping_by_name.return_value = mapping
        port_snr = SimpleNamespace(
            interface_name="hccs-1",
            abnormal_lane_snr=[
                SimpleNamespace(lane_name="lane0", snr_value="10"),
                SimpleNamespace(lane_name="lane1", snr_value="20"),
            ],
        )
        analyzer.swis_info = {
            "10.0.0.1": SimpleNamespace(
                swi_id="10.0.0.1", slot_id="61", hccs_info=SimpleNamespace(interface_snr_list=[port_snr])
            )
        }

        results = analyzer.analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertIn("lane0 低于阈值", results[0].fault_info)
        self.assertIn("对端NPU3", results[0].fault_info)

    def test_destination_port_snr_maps_chip_port_and_reports_threshold(self):
        threshold = MagicMock()
        threshold.check_value_str.return_value = "低于阈值"
        analyzer = PortSnrDestAnalyzer.__new__(PortSnrDestAnalyzer)
        analyzer.threshold = SimpleNamespace(SWITCH_PORT_SNR_LINE=threshold)
        analyzer.xpu_snr_limit_map = {"CPU": threshold}
        analyzer.port_mapping_config_instance = MagicMock()
        analyzer.port_mapping_config_instance.find_swi_port.return_value = SimpleNamespace(
            swi_port="100GE1/0/1", xpu_id="2"
        )
        port_snr = SimpleNamespace(swi_chip_id="0", port_id="1", lane_id="0", snr="10", xpu="CPU")
        analyzer.swis_info = {
            "10.0.0.1": SimpleNamespace(
                swi_id="10.0.0.1", slot_id="61", hccs_info=SimpleNamespace(hccs_chip_port_snr_list=[port_snr])
            )
        }

        results = analyzer.analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertEqual(results[0].domain.interface, "100GE1/0/1")
        self.assertIn("对端CPU2", results[0].fault_info)

    def test_rp_tx_reports_timeout_even_when_chassis_mapping_is_missing(self):
        timeout = MagicMock(interface="hccs-1", rp_tx=7)
        timeout.is_rp_tx_timeout_happend.return_value = True
        switch_info = SimpleNamespace(
            swi_id="10.0.0.1", slot_id="61", hccs_info=SimpleNamespace(proxy_timeout_statis=[timeout])
        )
        analyzer = HccsRpTxAnalyzer.__new__(HccsRpTxAnalyzer)
        analyzer.swis_info = {switch_info.swi_id: switch_info}
        analyzer.chassis_mappings = MagicMock()
        analyzer.chassis_mappings.find_mapping_by_l1_swi_ip.return_value = None

        results = analyzer.analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertIn("RP TX超时", results[0].fault_info)
        self.assertIn("7", results[0].fault_info)

    def test_rx_reports_timeout_and_xpu_fallback(self):
        timeout = MagicMock(interface="hccs-1", rp_rx=3, lp_tx=2)
        timeout.is_rx_timeout_happend.return_value = True
        switch_info = SimpleNamespace(
            swi_id="10.0.0.1",
            slot_id="61",
            date_time="2026-09-01 00:00:00+0800",
            hccs_info=SimpleNamespace(proxy_timeout_statis=[timeout]),
        )
        analyzer = HccsRxAnalyzer.__new__(HccsRxAnalyzer)
        analyzer.swis_info = {switch_info.swi_id: switch_info}
        analyzer.chassis_mappings = MagicMock()
        analyzer.chassis_mappings.find_mapping_by_l1_swi_ip.return_value = SimpleNamespace(
            server_super_pod_id="server-01", server_ip="192.0.2.1"
        )
        lcne = MagicMock(interface="hccs-1")
        analyzer.lcne_infos = {"server-01": {"hccs-1": lcne}}
        analyzer.check_long_link_down = MagicMock(return_value=False)
        analyzer.check_link_up_down = MagicMock(return_value=False)
        lcne.is_lane_error.return_value = False

        results = analyzer.analyse()

        self.assertEqual(len(results), 2)
        self.assertEqual(results[0].domain.slot_id, "61")
        self.assertIn("RX超时", results[0].fault_info)
        self.assertIn("xpu设备异常", results[1].fault_info)

    def test_bmc_link_degraded_parses_event_and_builds_port_result(self):
        event = SimpleNamespace(
            event_code="0X28000049",
            event_description="CPU1 UBC2 macro3 on CPU board 4 link degraded",
            event_time="2026-09-01 10:00:00",
        )
        bmc = SimpleNamespace(health_events=[event])
        analyzer = HccsLinkDegradedAnalyzer(SimpleNamespace())

        events = analyzer._find_hccs_link_degrade_events(bmc)

        self.assertEqual(len(events), 1)
        self.assertEqual(events[0].cpu_id, "1")
        mapping = SimpleNamespace(swi_port="100GE1/0/1", swi_chip_id="0", phy_id="1")
        chassis = SimpleNamespace(l1_swi_ip="10.0.0.1")
        with patch.object(HccsLinkDegradedAnalyzer, "_find_cpu_peer_swi_port", return_value=mapping):
            results = analyzer._port_fault_analyse(chassis, events)
        self.assertGreaterEqual(len(results), 1)
        self.assertIn("Cpu1 UBC2 macro3", results[0].fault_info)
        self.assertEqual(results[0].fault_time, "2026-09-01 10:00:00")


if __name__ == "__main__":
    unittest.main()
