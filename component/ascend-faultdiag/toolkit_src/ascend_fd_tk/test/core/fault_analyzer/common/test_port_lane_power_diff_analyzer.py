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

from ascend_fd_tk.core.fault_analyzer.common.port_lane_power_diff_analyzer import (
    PortLanePowerDiffAnalyzer,
)
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.host import HCCNOpticalInfo, HostInfo, NpuChipInfo


class TestPortLanePowerDiffAnalyzer(unittest.TestCase):
    @staticmethod
    def _build_cluster_info(chip_phy_id: str, npu_chip_info: NpuChipInfo):
        host_info = HostInfo(
            host_id="host-01",
            sn_num="sn-01",
            npu_chip_info={chip_phy_id: npu_chip_info},
        )
        return ClusterInfoCache(hosts_info={host_info.host_id: host_info})

    @staticmethod
    def _build_optical_info():
        return HCCNOpticalInfo(
            tx_power0="10",
            tx_power1="1",
            tx_power2="1",
            tx_power3="1",
            rx_power0="10",
            rx_power1="1",
            rx_power2="1",
            rx_power3="1",
        )

    def test_host_chip_domain_uses_chip_phy_id_when_npu_id_differs(self):
        npu_chip_info = NpuChipInfo(
            hccn_optical_info=self._build_optical_info(),
            npu_id="2",
            chip_id="0",
            chip_phy_id="5",
        )

        results = PortLanePowerDiffAnalyzer(self._build_cluster_info("5", npu_chip_info)).analyse()

        self.assertEqual(len(results), 1)
        self.assertEqual(results[0].domain.npu_id, "2")
        self.assertEqual(results[0].domain.chip_phy_id, "5")
        self.assertIn("chip:5", results[0].get_domain_desc())


if __name__ == "__main__":
    unittest.main()
