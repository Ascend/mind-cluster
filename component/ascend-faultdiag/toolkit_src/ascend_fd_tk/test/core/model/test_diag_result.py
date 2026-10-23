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

"""DiagResult 单元测试：各域描述生成与故障类型判定。"""

import unittest

from ascend_fd_tk.core.model.diag_result import BmcDomain, DiagResult, HostDomain, SwitchDomain


class TestHostDomain(unittest.TestCase):
    def test_get_desc(self):
        domain = HostDomain(
            host_id="1",
            npu_id="3",
            chip_phy_id="5",
            udie_id="0",
            npu_port_id="5",
            nic_id="10",
            nic_port_id="25",
            peer_switch_id="sw0",
            peer_interface="100GE1/1",
        )
        self.assertEqual(
            domain.get_desc(),
            "服务器:1->NPU:3->chip:5->UDie:0->NPU端口:5->网卡:10->网卡端口:25->交换机:sw0->交换机端口:100GE1/1",
        )

    def test_get_local_domain_partial(self):
        domain = HostDomain(host_id="1", npu_id="3")
        self.assertEqual(domain.get_local_domain(), ["服务器:1", "NPU:3"])


class TestBmcDomain(unittest.TestCase):
    def test_get_desc(self):
        domain = BmcDomain(bmc_id="bmc0", npu_id="2", chip_phy_id="4")
        self.assertEqual(domain.get_desc(), "BMC:bmc0->NPU:2->chip:4")


class TestSwitchDomain(unittest.TestCase):
    def test_get_desc(self):
        domain = SwitchDomain(
            swi_id="sw0",
            interface="100GE1/1",
            optical_id="3",
            peer_switch_id="sw1",
            peer_switch_interface="100GE1/2",
        )
        self.assertEqual(
            domain.get_desc(),
            "交换机:sw0->交换机端口:100GE1/1->光模块ID:3->交换机:sw1->交换机端口:100GE1/2",
        )


class TestDiagResult(unittest.TestCase):
    def test_fault_type_host(self):
        result = DiagResult(domain=HostDomain(host_id="1"))
        self.assertEqual(result.fault_type, "host")

    def test_fault_type_bmc(self):
        result = DiagResult(domain=BmcDomain(bmc_id="b"))
        self.assertEqual(result.fault_type, "bmc")

    def test_fault_type_switch(self):
        result = DiagResult(domain=SwitchDomain(swi_id="s"))
        self.assertEqual(result.fault_type, "switch")

    def test_to_dict_contains_domain_desc(self):
        result = DiagResult(domain=HostDomain(host_id="1"), fault_info="掉线", suggestion="换线")
        data = result.to_dict()
        self.assertIn("服务器:1", data["故障域"])
        self.assertEqual(data["故障信息"], "掉线")
        self.assertEqual(data["处理建议"], "换线")
        self.assertEqual(data["故障类型"], "host")


if __name__ == "__main__":
    unittest.main()
