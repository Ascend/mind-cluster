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
from unittest.mock import ANY, MagicMock, call

from ascend_fd_tk.core.fault_analyzer.host.host_analyzer import HostAnalyzer
from ascend_fd_tk.core.fault_analyzer.host.host_analyzer_a5 import HostAnalyzerA5
from ascend_fd_tk.core.fault_analyzer.host.host_credit_analyzer_a5 import HostCreditAnalyzerA5
from ascend_fd_tk.core.fault_analyzer.host.host_loopback_analyzer import HostLoopbackAnalyzer
from ascend_fd_tk.core.fault_analyzer.host.host_nic_lane_analyzer_a5 import HostNicLaneAnalyzerA5
from ascend_fd_tk.core.fault_analyzer.host.host_op_los_lol_analyzer import HostOpticalLosLoLAnalyzer
from ascend_fd_tk.core.fault_analyzer.host.host_op_los_lol_analyzer_a5 import HostOpticalLosLoLAnalyzerA5
from ascend_fd_tk.core.fault_analyzer.host.host_op_status_analyzer import HostOpticalStatusAnalyzer
from ascend_fd_tk.core.fault_analyzer.host.inter_host_analyzer import InterHostFaultAnalyzer
from ascend_fd_tk.core.fault_analyzer.host.roce_port_analyzer import RocePortAnalyzer
from ascend_fd_tk.core.model.diag_result import HostDomain


def _cluster(host_info=None, switch_info=None, threshold=None):
    cluster = SimpleNamespace(
        hosts_info={host_info.host_id: host_info} if host_info else {},
        swis_info={switch_info.swi_id: switch_info} if switch_info else {},
    )
    cluster.get_threshold = MagicMock(return_value=threshold or SimpleNamespace())
    return cluster


class TestHostAnalyzers(unittest.TestCase):
    def test_host_optical_status_covers_absent_unreachable_low_power_and_tx_disable(self):
        threshold = SimpleNamespace()
        analyzer = HostAnalyzer(_cluster(threshold=threshold))
        domain = HostDomain(host_id="host-01", npu_id="0", chip_phy_id="4")
        optical = MagicMock(present="0", control_link_unreachable=False)
        optical.is_optical_present.return_value = False
        chip = SimpleNamespace(hccn_optical_info=optical, hccn_dfx_cfg=None)
        self.assertIn("未在位", analyzer._analyze_optical_status(domain, chip)[0].fault_info)

        optical.is_optical_present.return_value = True
        optical.control_link_unreachable = True
        self.assertIn("unreachable", analyzer._analyze_optical_status(domain, chip)[0].fault_info)

        optical.control_link_unreachable = False
        optical.is_high_power_enable.return_value = False
        optical.high_power_enable_reg = "0x0"
        self.assertIn("低功率", analyzer._analyze_optical_status(domain, chip)[0].fault_info)

        optical.is_high_power_enable.return_value = True
        dfx = MagicMock(tx_disable_status="1")
        dfx.is_tx_disable.return_value = True
        chip.hccn_dfx_cfg = dfx
        self.assertIn("关光", analyzer._analyze_optical_status(domain, chip)[0].fault_info)
        self.assertEqual(optical.is_optical_present.call_args_list, [call(threshold)] * 4)

    def test_host_analyzer_reports_power_snr_cdr_and_link_thresholds(self):
        threshold = SimpleNamespace(
            HOST_SNR_DB=object(),
            MEDIA_SNR_DB=object(),
            SNR_LANE_DIFF_DB=object(),
            CDR_HOST_SNR_DB=object(),
            CDR_MEDIA_SNR_DB=object(),
            HCCN_LINK_DOWN_CNT=SimpleNamespace(high_alarm_th="4", high_warn_th="2", desc="24小时down次数"),
        )
        analyzer = HostAnalyzer.__new__(HostAnalyzer)
        analyzer._threshold = threshold
        domain = HostDomain(host_id="host-01", npu_id="0", chip_phy_id="4")
        optical = MagicMock()
        optical.get_abnormal_power_infos.return_value = (["rx low"], ["tx high"])
        optical.get_abnormal_snr_infos.return_value = "snr low"
        optical.get_lane_diff_desc.return_value = "lane diff"
        cdr = MagicMock()
        cdr.get_snr_abnormal_desc.return_value = "cdr snr low"
        cdr.get_lane_diff_desc.return_value = "cdr lane diff"
        link_stat = MagicMock(link_history=["up", "down"])
        link_stat.link_down_within_24h.return_value = (4, ["t1", "t2"])
        chip = SimpleNamespace(cdr_snr_info=cdr, hccn_link_stat_info=link_stat)

        self.assertEqual(len(analyzer._analyze_power(domain, optical)), 1)
        self.assertEqual(len(analyzer._analyze_optical_snr(domain, optical)), 2)
        self.assertEqual(len(analyzer._analyze_cdr(domain, chip)), 2)
        self.assertIn("链路异常", analyzer._analyze_link_stat(domain, chip)[0].fault_info)

    def test_loopback_reports_local_port_and_optical_faults(self):
        first_down = MagicMock()
        first_down.is_first_record_up.return_value = False
        media_up = MagicMock()
        media_up.is_first_record_up.return_value = True
        loopback = SimpleNamespace(
            host_input_enable=True,
            host_input_link_stat=first_down,
            media_output_enable=True,
            media_output_link_stat=media_up,
            npu_id="0",
            chip_phy_id="4",
        )
        host = SimpleNamespace(host_id="host-01", loopback_info_list=[loopback])

        results = HostLoopbackAnalyzer.host_loopback_diag(host)

        self.assertEqual(len(results), 2)
        self.assertIn("本端故障", results[0].fault_info)
        self.assertIn("光模块故障", results[1].fault_info)

    def test_host_los_lol_reports_flags_and_missing_optical_info(self):
        optical = SimpleNamespace(rx_los_flag="0x1", tx_los_flag="0x0", rx_lo_l_flag="0x2", tx_lo_l_flag="")
        chip_fault = SimpleNamespace(npu_id="0", chip_phy_id="4", hccn_optical_info=optical)
        chip_missing = SimpleNamespace(npu_id="1", chip_phy_id="5", hccn_optical_info=None)
        host = SimpleNamespace(host_id="host-01", npu_chip_info={"4": chip_fault, "5": chip_missing})

        results = HostOpticalLosLoLAnalyzer(_cluster(host_info=host)).analyse()

        self.assertEqual(len(results), 3)
        self.assertTrue(any("Rx Los" in result.fault_info for result in results))
        self.assertTrue(any("未查询到" in result.fault_info for result in results))

    def test_host_los_lol_a5_reports_only_abnormal_supported_flags(self):
        optical = SimpleNamespace(
            udie_id="1",
            port_id="0",
            optical_type="",
            state_flag=[
                SimpleNamespace(items="RxLos Flag", lanes=["Normal", "Alarm", "", "Normal"]),
                SimpleNamespace(items="Temperature", lanes=["Alarm"]),
            ],
        )
        chip = SimpleNamespace(npu_id="0", hccn_optical_info=[optical])

        results = HostOpticalLosLoLAnalyzerA5._check_optical_indicator("host-01", chip, chip.hccn_optical_info)

        self.assertEqual(len(results), 1)
        self.assertIn("Lane1：Alarm", results[0].fault_info)
        self.assertEqual(results[0].domain.udie_id, "1")
        self.assertEqual(results[0].domain.npu_port_id, "0")

    def test_host_los_lol_a5_skips_lol_flags_for_lpo_module(self):
        optical = SimpleNamespace(
            udie_id="1",
            port_id="0",
            optical_type="LPO",
            state_flag=[
                # LPO 不支持 Lol 标志，即使值异常也不应上报
                SimpleNamespace(items="TxLol Flag", lanes=["Alarm", "Alarm"]),
                SimpleNamespace(items="RxLos Flag", lanes=["Normal", "Alarm"]),
            ],
        )
        chip = SimpleNamespace(npu_id="0", hccn_optical_info=[optical])

        results = HostOpticalLosLoLAnalyzerA5._check_optical_indicator("host-01", chip, chip.hccn_optical_info)

        self.assertEqual(len(results), 1)
        self.assertIn("RxLos Flag", results[0].fault_info)
        self.assertNotIn("Lol", results[0].fault_info)

    def test_host_optical_status_reports_peer_details(self):
        threshold = SimpleNamespace(
            NET_HEALTH_THRESHOLD=SimpleNamespace(normal_alarm_th="Healthy"),
            LINK_STATUS_THRESHOLD=SimpleNamespace(normal_alarm_th="UP"),
        )
        lldp = SimpleNamespace(system_name_tlv="switch-01", port_id_tlv="100GE1/0/1")
        chip = SimpleNamespace(net_health="Fault", link_status="DOWN", hccn_lldp_info=lldp, npu_id="0", chip_phy_id="4")
        host = SimpleNamespace(host_id="host-01", npu_chip_info={"4": chip})

        results = HostOpticalStatusAnalyzer(_cluster(host_info=host, threshold=threshold)).analyse()

        self.assertEqual(len(results), 1)
        self.assertIn("switch-01", results[0].fault_info)
        self.assertIn("100GE1/0/1", results[0].fault_info)

    def test_roce_port_reports_speed_or_duplex_mismatch(self):
        lldp = SimpleNamespace(system_name_tlv="switch-01", port_id_tlv="100GE1/0/1")
        chip = SimpleNamespace(npu_id="0", chip_phy_id="4", hccn_lldp_info=lldp, speed="100000", duplex="full")
        host = SimpleNamespace(host_id="host-01", npu_chip_info={"4": chip})
        peer_port = SimpleNamespace(interface_name="100GE1/0/1", speed="200000", duplex="full")
        switch_info = SimpleNamespace(swi_id="10.0.0.1", name="switch-01", interface_info=[peer_port])

        results = RocePortAnalyzer(_cluster(host_info=host, switch_info=switch_info)).analyse()

        self.assertEqual(len(results), 1)
        self.assertIn("连接信息不相同", results[0].fault_info)

    def test_inter_host_delegates_single_ended_when_peer_is_missing(self):
        optical = object()
        chip = MagicMock(
            hccn_optical_info=object(),
            hccn_lldp_info=None,
            npu_id="0",
            chip_phy_id="4",
            chip_id="0",
        )
        # 本分支 get_optical_module_info 返回列表（A5 单 NPU 多光模块）
        chip.get_optical_module_info.return_value = [optical]
        host = SimpleNamespace(host_id="host-01", npu_chip_info={"4": chip})
        cluster = _cluster(host_info=host)
        analyzer = InterHostFaultAnalyzer(cluster)
        analyzer.fault_check = MagicMock()
        analyzer.fault_check.power_analyze_single_ended.return_value = ["power"]
        analyzer.fault_check.snr_analyze_single_ended.return_value = ["snr"]
        analyzer.fault_check.bias_analyze_single_ended.return_value = ["bias"]

        results = analyzer.analyse()

        self.assertEqual(results, ["power", "snr", "bias"])
        analyzer.fault_check.power_analyze_single_ended.assert_called_once_with(ANY, optical)
        analyzer.fault_check.snr_analyze_single_ended.assert_called_once_with(ANY, optical)
        analyzer.fault_check.bias_analyze_single_ended.assert_called_once_with(ANY, optical)

    def test_host_analyzer_a5_reports_bias_power_snr_and_lane_diff(self):
        abnormal = MagicMock()
        abnormal.check_value_str.return_value = "异常"
        lane_diff = MagicMock()
        lane_diff.check_lane_diff_desc.return_value = ["lane0-lane1差值异常"]
        # 本分支按光模块类型取阈值视图（LPO 优先，其他回退 ODSP）
        threshold_view = SimpleNamespace(
            TX_BIAS_MA=abnormal,
            TX_POWER_DBM=abnormal,
            RX_POWER_DBM=abnormal,
            HOST_SNR_DB=abnormal,
            MEDIA_SNR_DB=abnormal,
            SNR_LANE_DIFF_DB=lane_diff,
        )
        threshold = MagicMock()
        threshold.get_optical_view.return_value = threshold_view
        analyzer = HostAnalyzerA5.__new__(HostAnalyzerA5)
        analyzer._threshold = threshold
        domain = HostDomain(host_id="host-01", npu_id="0", udie_id="1", npu_port_id="0")
        optical = SimpleNamespace(
            optical_type="",
            monitor_item=[
                SimpleNamespace(items="Bias Lane0(mA)", value="20"),
                SimpleNamespace(items="TxPower Lane0(dBm)", value="3"),
                SimpleNamespace(items="RxPower Lane0(dBm)", value="-20"),
                SimpleNamespace(items="HostSNR Lane0(dB)", value="10"),
                SimpleNamespace(items="MediaSNR Lane0(dB)", value="10"),
            ],
        )

        self.assertEqual(len(analyzer._analyze_bias(domain, optical)), 1)
        self.assertEqual(len(analyzer._analyze_power(domain, optical)), 1)
        self.assertEqual(len(analyzer._analyze_optical_snr(domain, optical)), 3)
        threshold.get_optical_view.assert_called_with("")

    def test_host_nic_lane_a5_skips_placeholders_and_reports_flags_and_numeric_values(self):
        flag_threshold = MagicMock()
        flag_threshold.check_value_str.side_effect = lambda value: "异常" if value not in ("0", "normal") else ""
        numeric_threshold = MagicMock()
        numeric_threshold.check_value_str.side_effect = lambda value: "异常" if value not in ("0", "normal") else ""
        threshold = SimpleNamespace(
            NIC_LANE_FLAG=flag_threshold,
            NIC_TX_BIAS_MA=numeric_threshold,
            NIC_TX_POWER_DBM=numeric_threshold,
            NIC_RX_POWER_DBM=numeric_threshold,
            NIC_HOST_SNR_DB=numeric_threshold,
            NIC_MEDIA_SNR_DB=numeric_threshold,
        )
        analyzer = HostNicLaneAnalyzerA5.__new__(HostNicLaneAnalyzerA5)
        analyzer._threshold = threshold
        port = SimpleNamespace(
            port_id="1",
            tx_los_lanes=["0", "1"],
            rx_los_lanes=[],
            tx_cdr_lol_lanes=[],
            rx_cdr_lol_lanes=[],
            bias_lanes=["-inf", "20"],
            tx_power_lanes=[],
            rx_power_lanes=[],
            host_snr_lanes=[],
            media_snr_lanes=[],
        )
        nic = SimpleNamespace(card_name="nic0")

        results = analyzer._check_nic_port_lanes("host-01", nic, port)

        self.assertEqual(len(results), 2)
        self.assertTrue(any("TxLos" in result.fault_info for result in results))
        self.assertTrue(any("bias" in result.fault_info for result in results))
        self.assertNotIn(call("-inf"), numeric_threshold.check_value_str.call_args_list)

    def test_host_credit_a5_reports_exhausted_pri_credit(self):
        """alloc == used 且非 0 的优先级应上报信用不足；alloc==used==0、alloc!=used、空值不报。"""
        chip = SimpleNamespace(
            npu_id="0",
            credit_info_list=[
                SimpleNamespace(
                    udie_id="0",
                    port_id="5",
                    link_alloc_port_share_credit="100",
                    link_cur_used_port_share_credit="36",
                    link_alloc_vl_pri_credits=["0", "8", "16", "", "4"],
                    link_cur_used_pri_credits=["0", "8", "10", "", "0"],
                )
            ],
        )
        host = SimpleNamespace(host_id="host-01", npu_chip_info={"0": chip})
        results = HostCreditAnalyzerA5(_cluster(host_info=host)).analyse()
        self.assertEqual(len(results), 1)
        self.assertIn("vl1 分配数量=8 使用数量=8", results[0].fault_info)
        self.assertNotIn("vl2", results[0].fault_info)
        self.assertEqual(results[0].domain.udie_id, "0")
        self.assertEqual(results[0].domain.npu_port_id, "5")

    def test_host_credit_a5_no_result_when_all_normal(self):
        """所有优先级均正常（未耗尽）时不应产生诊断结果。"""
        chip = SimpleNamespace(
            npu_id="0",
            credit_info_list=[
                SimpleNamespace(
                    udie_id="0",
                    port_id="5",
                    link_alloc_port_share_credit="100",
                    link_cur_used_port_share_credit="36",
                    link_alloc_vl_pri_credits=["0", "8"],
                    link_cur_used_pri_credits=["0", "4"],
                ),
                SimpleNamespace(udie_id="1", port_id="6", link_alloc_vl_pri_credits=[], link_cur_used_pri_credits=[]),
            ],
        )
        host = SimpleNamespace(host_id="host-01", npu_chip_info={"0": chip})
        results = HostCreditAnalyzerA5(_cluster(host_info=host)).analyse()
        self.assertEqual(results, [])


if __name__ == "__main__":
    unittest.main()
