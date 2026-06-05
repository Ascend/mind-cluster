#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
# Licensed under the Apache License, Version 2.0

import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.model.diag_result import HostDomain, BmcDomain, SwitchDomain
from ascend_fd_tk.core.root_cause.constants import (
    IS_ROOT_CAUSE,
    NOT_ROOT_CAUSE,
    UNKNOWN_ROOT_CAUSE,
    ROOT_CAUSE_NPU_CPU,
    ROOT_CAUSE_L1_OPTICAL,
    RULE_L1_200G_HILINK,
    RULE_L1_400G_HILINK,
)
from ascend_fd_tk.core.root_cause.model import AnalyzedRootCausePort, HostToL1LinkData, L1ToL2LinkData
from ascend_fd_tk.core.root_cause.filter import RootCauseFilter


class TestRootCauseFilterMergePorts(unittest.TestCase):
    def _f(self):
        return RootCauseFilter(MagicMock())

    def test_merge_and_dedup(self):
        f = self._f()
        f._merge_analyzed_ports([AnalyzedRootCausePort(component=ROOT_CAUSE_NPU_CPU, host_id="h1")])
        self.assertEqual(len(f.analyzed_root_cause_ports), 1)
        # 不是根因→是根因升级
        f._merge_analyzed_ports(
            [AnalyzedRootCausePort(is_root_cause=NOT_ROOT_CAUSE, component=ROOT_CAUSE_NPU_CPU, host_id="h2")]
        )
        f._merge_analyzed_ports(
            [AnalyzedRootCausePort(is_root_cause=IS_ROOT_CAUSE, component=ROOT_CAUSE_NPU_CPU, host_id="h2")]
        )
        self.assertEqual(len(f.analyzed_root_cause_ports), 2)
        # 是根因不被覆盖
        f._merge_analyzed_ports(
            [AnalyzedRootCausePort(is_root_cause=NOT_ROOT_CAUSE, component=ROOT_CAUSE_NPU_CPU, host_id="h2")]
        )
        self.assertTrue(all(p.is_root_cause == IS_ROOT_CAUSE for p in f.analyzed_root_cause_ports if p.host_id == "h2"))


class TestRootCauseFilterGetStatus(unittest.TestCase):
    def _f(self, ports):
        f = RootCauseFilter(MagicMock())
        f._merge_analyzed_ports(ports)
        return f

    def test_host_domain(self):
        f = self._f([AnalyzedRootCausePort(is_root_cause=IS_ROOT_CAUSE, component=ROOT_CAUSE_NPU_CPU, host_id="h1")])
        self.assertEqual(f.get_root_cause_status(HostDomain(host_id="h1")), IS_ROOT_CAUSE)
        self.assertEqual(f.get_root_cause_status(HostDomain(host_id="h2")), UNKNOWN_ROOT_CAUSE)

    def test_bmc_domain(self):
        f = self._f([AnalyzedRootCausePort(is_root_cause=IS_ROOT_CAUSE, component=ROOT_CAUSE_NPU_CPU, npu_id="n1")])
        self.assertEqual(f.get_root_cause_status(BmcDomain(npu_id="n1")), IS_ROOT_CAUSE)

    def test_switch_domain(self):
        f = self._f(
            [
                AnalyzedRootCausePort(
                    is_root_cause=IS_ROOT_CAUSE, component=ROOT_CAUSE_L1_OPTICAL, switch_id="swi1", port="400G-1"
                )
            ]
        )
        self.assertEqual(f.get_root_cause_status(SwitchDomain(swi_id="swi1", interface="400G-1")), IS_ROOT_CAUSE)


class TestRootCauseFilterBuildAndAnalyze(unittest.TestCase):
    @patch("ascend_fd_tk.core.root_cause.filter.LinkBuilder")
    @patch("ascend_fd_tk.core.root_cause.filter.FaultAnalyzer")
    def test_build_and_analyze(self, MockAnalyzer, MockBuilder):
        MockBuilder.return_value.build.return_value = (
            [HostToL1LinkData(host_id="h1", triggered_rules=[RULE_L1_200G_HILINK])],
            [L1ToL2LinkData(l1_switch_id="s1", triggered_rules=[RULE_L1_400G_HILINK])],
        )
        MockAnalyzer.return_value.analyze_host_to_l1_fault.return_value = [
            AnalyzedRootCausePort(is_root_cause=IS_ROOT_CAUSE, component=ROOT_CAUSE_NPU_CPU, host_id="h1")
        ]
        MockAnalyzer.return_value.analyze_l1_to_l2_fault.return_value = [
            AnalyzedRootCausePort(is_root_cause=IS_ROOT_CAUSE, component=ROOT_CAUSE_L1_OPTICAL, switch_id="s1")
        ]
        f = RootCauseFilter(MagicMock())
        f.build_and_analyze()
        self.assertEqual(len(f.analyzed_root_cause_ports), 2)


if __name__ == "__main__":
    unittest.main()
