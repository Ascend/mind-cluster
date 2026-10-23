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

"""root_cause/model 单元测试。"""

import unittest

from ascend_fd_tk.core.model.diag_result import BmcDomain, HostDomain, SwitchDomain
from ascend_fd_tk.core.root_cause.model import AnalyzedRootCausePort, HostToL1LinkData, L1ToL2LinkData


class TestRootCauseModel(unittest.TestCase):
    def test_host_domain_match(self):
        port = AnalyzedRootCausePort(host_id="h1", npu_id="0", chip_phy_id="phy0")
        self.assertTrue(port.matches_domain(HostDomain(host_id="h1", npu_id="0", chip_phy_id="phy0")))

    def test_host_domain_no_match(self):
        port = AnalyzedRootCausePort(host_id="h1")
        self.assertFalse(port.matches_domain(HostDomain(host_id="h2")))

    def test_bmc_domain_match(self):
        port = AnalyzedRootCausePort(npu_id="0", chip_phy_id="phy0")
        self.assertTrue(port.matches_domain(BmcDomain(npu_id="0", chip_phy_id="phy0")))

    def test_bmc_domain_no_match(self):
        port = AnalyzedRootCausePort(npu_id="0")
        self.assertFalse(port.matches_domain(BmcDomain(npu_id="1")))

    def test_switch_domain_optical_match(self):
        port = AnalyzedRootCausePort(switch_id="s1", port="400G-1")
        self.assertTrue(port.matches_domain(SwitchDomain(swi_id="s1", interface="400G-1")))

    def test_switch_domain_board_match(self):
        port = AnalyzedRootCausePort(switch_id="s1", switch_chip_id="c0")
        self.assertTrue(port.matches_domain(SwitchDomain(swi_id="s1", interface="c0")))

    def test_switch_domain_no_match(self):
        port = AnalyzedRootCausePort(switch_id="s1", port="400G-1")
        self.assertFalse(port.matches_domain(SwitchDomain(swi_id="s2", interface="400G-1")))

    def test_none_domain_returns_false(self):
        port = AnalyzedRootCausePort(host_id="h1")
        self.assertFalse(port.matches_domain(None))

    def test_link_data_defaults(self):
        host_link = HostToL1LinkData(host_id="h1")
        l1_l2_link = L1ToL2LinkData(l1_switch_id="s1")
        self.assertEqual(host_link.triggered_rules, [])
        self.assertEqual(host_link.link_status, "")
        self.assertEqual(l1_l2_link.l2_switch_id, "")


if __name__ == "__main__":
    unittest.main()
