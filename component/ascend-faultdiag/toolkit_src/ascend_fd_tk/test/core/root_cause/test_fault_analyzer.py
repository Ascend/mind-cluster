#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
# Licensed under the Apache License, Version 2.0

import unittest

from ascend_fd_tk.core.root_cause.constants import (
    IS_ROOT_CAUSE,
    NOT_ROOT_CAUSE,
    ROOT_CAUSE_L1_SWITCH_BOARD,
    ROOT_CAUSE_L1_OPTICAL,
    ROOT_CAUSE_L2_OPTICAL,
    LINK_STATUS_ABNORMAL,
    RULE_L1_200G_HILINK,
    RULE_L1_400G_HILINK,
    RULE_L1_OPTICAL_HOST,
    RULE_L1_OPTICAL_MEDIA,
    NPU,
    CPU,
)
from ascend_fd_tk.core.root_cause.model import HostToL1LinkData, L1ToL2LinkData
from ascend_fd_tk.core.root_cause.fault_analyzer import FaultAnalyzer


class TestFaultAnalyzer(unittest.TestCase):
    def test_group_host_to_l1_links(self):
        links = [
            HostToL1LinkData(host_id="h1", npu_id="n1", chip_phy_id="c1", l1_switch_id="s1"),
            HostToL1LinkData(host_id="h1", npu_id="n2", chip_phy_id="c2", l1_switch_id="s1"),
        ]
        self.assertEqual(len(FaultAnalyzer._group_host_to_l1_links(links)), 2)
        # CPU类型
        cpu_links = [HostToL1LinkData(host_id="h1", npu_id="cpu0", npu_type=CPU, l1_switch_id="s1")]
        groups = FaultAnalyzer._group_host_to_l1_links(cpu_links)
        self.assertIn(("h1", "cpu_cpu0", "s1"), groups)

    def test_calc_board_stats(self):
        rows = [HostToL1LinkData(l1_switch_chip_id="c1", link_status=LINK_STATUS_ABNORMAL)]
        self.assertEqual(FaultAnalyzer._calc_board_stats(rows), (1, 1))

    def test_host_to_l1_analysis(self):
        rows = [HostToL1LinkData(npu_type=NPU, npu_id="n1", chip_phy_id="c1")]
        self.assertIn("所有200G端口hilink SNR异常", FaultAnalyzer._host_to_l1_fault_component_analysis(rows, 1, 1))
        self.assertIn("部分200G端口hilink SNR异常", FaultAnalyzer._host_to_l1_fault_component_analysis(rows, 2, 1))
        self.assertEqual(FaultAnalyzer._host_to_l1_fault_component_analysis(rows, 1, 0), "")

    def test_host_to_l1_root_cause(self):
        rows = [HostToL1LinkData(host_id="h1", npu_id="n1", chip_phy_id="c1")]
        self.assertEqual(FaultAnalyzer._create_host_to_l1_root_cause_ports(rows, 1, 1)[0].is_root_cause, IS_ROOT_CAUSE)
        self.assertEqual(FaultAnalyzer._create_host_to_l1_root_cause_ports(rows, 2, 1)[0].is_root_cause, NOT_ROOT_CAUSE)

    def test_analyze_host_to_l1_fault(self):
        links = [
            HostToL1LinkData(
                host_id="h1",
                npu_id="n1",
                chip_phy_id="c1",
                l1_switch_id="s1",
                l1_switch_chip_id="chip1",
                npu_type=NPU,
                link_status=LINK_STATUS_ABNORMAL,
                triggered_rules=[RULE_L1_200G_HILINK],
            )
        ]
        ports = FaultAnalyzer().analyze_host_to_l1_fault(links)
        self.assertTrue(len(ports) >= 1)

    def test_l1_to_l2_component_analysis(self):
        self.assertIn(
            "所有光模块端口host SNR均异常",
            FaultAnalyzer.l1_to_l2_component_analysis(
                [L1ToL2LinkData(l1_switch_chip_phy_id="p1", triggered_rules=[RULE_L1_OPTICAL_HOST])]
            ),
        )
        self.assertIn(
            "L1交换机400G端口hilink SNR异常",
            FaultAnalyzer.l1_to_l2_component_analysis([L1ToL2LinkData(triggered_rules=[RULE_L1_400G_HILINK])]),
        )
        self.assertEqual(FaultAnalyzer.l1_to_l2_component_analysis([L1ToL2LinkData(triggered_rules=[])]), "")

    def test_l1_to_l2_root_cause_ports(self):
        # L1交换板是根因
        rows = [
            L1ToL2LinkData(
                l1_switch_id="s1",
                l1_switch_chip_id="c1",
                l1_switch_chip_phy_id="p1",
                triggered_rules=[RULE_L1_OPTICAL_HOST],
            )
        ]
        ports = FaultAnalyzer._create_l1_to_l2_root_cause_ports(rows)
        self.assertTrue(
            any(p.component == ROOT_CAUSE_L1_SWITCH_BOARD and p.is_root_cause == IS_ROOT_CAUSE for p in ports)
        )
        # L1光模块是根因
        rows2 = [L1ToL2LinkData(l1_switch_id="s1", l1_interface="400G-1", triggered_rules=[RULE_L1_400G_HILINK])]
        ports2 = FaultAnalyzer._create_l1_to_l2_root_cause_ports(rows2)
        self.assertTrue(any(p.component == ROOT_CAUSE_L1_OPTICAL and p.is_root_cause == IS_ROOT_CAUSE for p in ports2))
        # L2光模块是根因
        rows3 = [
            L1ToL2LinkData(
                l1_switch_id="s1",
                l1_interface="400G-1",
                l2_switch_id="s2",
                l2_interface="200G-1",
                triggered_rules=[RULE_L1_OPTICAL_MEDIA],
            )
        ]
        ports3 = FaultAnalyzer._create_l1_to_l2_root_cause_ports(rows3)
        self.assertTrue(any(p.component == ROOT_CAUSE_L2_OPTICAL and p.is_root_cause == IS_ROOT_CAUSE for p in ports3))
        # 无L2信息时跳过L2光模块
        rows4 = [
            L1ToL2LinkData(
                l1_switch_id="s1",
                l1_interface="400G-1",
                l2_switch_id="",
                l2_interface="",
                triggered_rules=[RULE_L1_OPTICAL_MEDIA],
            )
        ]
        ports4 = FaultAnalyzer._create_l1_to_l2_root_cause_ports(rows4)
        self.assertEqual(len([p for p in ports4 if p.component == ROOT_CAUSE_L2_OPTICAL]), 0)

    def test_analyze_l1_to_l2_fault(self):
        links = [
            L1ToL2LinkData(
                l1_switch_id="s1",
                l1_switch_chip_id="c1",
                l1_switch_chip_phy_id="p1",
                l1_interface="400G-1",
                l2_switch_id="s2",
                l2_interface="200G-1",
                triggered_rules=[RULE_L1_400G_HILINK, RULE_L1_OPTICAL_HOST],
            )
        ]
        self.assertTrue(len(FaultAnalyzer().analyze_l1_to_l2_fault(links)) >= 2)


if __name__ == "__main__":
    unittest.main()
