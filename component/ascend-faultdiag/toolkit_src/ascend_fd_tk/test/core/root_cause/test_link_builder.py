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

"""root_cause/link_builder 单元测试。"""

import types
import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.config.port_mapping_config import L1InterfacePortMapping
from ascend_fd_tk.core.root_cause.constants import (
    LINK_STATUS_ABNORMAL,
    LINK_STATUS_NORMAL,
    PORT_SPEED_200G,
    PORT_SPEED_400G,
    RULE_L1_200G_HILINK,
)
from ascend_fd_tk.core.root_cause.link_builder import LinkBuilder


def _make_host_info():
    return types.SimpleNamespace(
        host_id="h1",
        hostname="H1",
        room_name="r1",
        cabinet_id="c1",
        npu_chip_info={"0": types.SimpleNamespace(npu_id="0", chip_id="0", chip_phy_id="phy0")},
    )


def _make_l1_swi(ifaces=None):
    return types.SimpleNamespace(
        swi_id="ip1",
        name="L1-1",
        room_name="r1",
        cabinet_id="c1",
        interface_full_infos=ifaces or {},
    )


class TestLinkBuilder(unittest.TestCase):
    @patch("ascend_fd_tk.core.root_cause.link_builder.SnrChecker")
    def test_build(self, MockSnr):
        cluster = MagicMock()
        cluster.get_chassis_mappings.return_value = None
        host_links, l2_links = LinkBuilder(cluster).build()
        self.assertEqual(host_links, [])
        self.assertEqual(l2_links, [])

        cluster.get_chassis_mappings.return_value = types.SimpleNamespace(
            l1_swi_server_mappings=[types.SimpleNamespace(l1_swi_ip="ip1", server_super_pod_id="pod1")]
        )
        cluster.swis_info = {"ip1": _make_l1_swi()}
        port_mapping = L1InterfacePortMapping(
            swi_port="X/1", swi_chip_id="c0", phy_id="p0", xpu="npu", xpu_id="0", chip_id="0"
        )
        MockSnr.get_port_speed.side_effect = lambda name: PORT_SPEED_200G
        mock_snr = MockSnr.return_value
        mock_snr.check_hilink_snr_abnormal.return_value = (False, "")
        mock_snr.format_hilink_snr.return_value = ""

        with patch("ascend_fd_tk.core.root_cause.link_builder.get_port_mapping_config_instance") as mock_cfg:
            mock_cfg.return_value.l1_interface_port_map = {"200G-1": port_mapping}
            cluster.find_host_info_by_server_spod_id.return_value = _make_host_info()
            links = LinkBuilder(cluster)._build_host_to_l1_links()
        self.assertEqual(len(links), 1)
        link = links[0]
        self.assertEqual(link.l1_switch_id, "ip1")
        self.assertEqual(link.l1_switch_chip_id, "c0")
        self.assertEqual(link.l1_interface, "200G-1")
        self.assertEqual(link.host_id, "h1")
        self.assertEqual(link.npu_id, "0")
        self.assertEqual(link.chip_phy_id, "phy0")
        self.assertEqual(link.link_status, LINK_STATUS_NORMAL)

        cluster.find_host_info_by_server_spod_id.return_value = None
        mock_snr.check_hilink_snr_abnormal.return_value = (True, "")
        with patch("ascend_fd_tk.core.root_cause.link_builder.get_port_mapping_config_instance") as mock_cfg:
            mock_cfg.return_value.l1_interface_port_map = {"200G-1": port_mapping}
            links = LinkBuilder(cluster)._build_host_to_l1_links()
        self.assertEqual(links[0].link_status, LINK_STATUS_ABNORMAL)
        self.assertIn(RULE_L1_200G_HILINK, links[0].triggered_rules)

    @patch("ascend_fd_tk.core.root_cause.link_builder.SnrChecker")
    def test_build_l1_to_l2_links(self, MockSnr):
        cluster = MagicMock()
        cluster.get_chassis_mappings.return_value = types.SimpleNamespace(
            l1_swi_server_mappings=[
                types.SimpleNamespace(l1_swi_ip="ip1", l1_swi_name="L1-1", server_super_pod_id="pod1")
            ]
        )
        l1_full = types.SimpleNamespace(
            interface_mapping=types.SimpleNamespace(
                remote_device_interface=types.SimpleNamespace(device_name="ip2", interface="200G-1")
            ),
            get_optical_module_info=lambda: None,
        )
        l2_full = types.SimpleNamespace(interface="200G-1", interface_info=None, get_optical_module_info=lambda: None)
        l1_swi = _make_l1_swi({"400G-1": l1_full})
        l2_swi = types.SimpleNamespace(
            swi_id="ip2",
            name="L2-1",
            room_name="r2",
            cabinet_id="c2",
            interface_full_infos={"200G-1": l2_full},
        )
        cluster.swis_info = {"ip1": l1_swi, "ip2": l2_swi}
        cluster.find_peer_swi.return_value = l2_swi

        mock_snr = MockSnr.return_value
        MockSnr.get_port_speed.side_effect = lambda name: (PORT_SPEED_400G if "400G" in name else PORT_SPEED_200G)
        mock_snr.check_hilink_snr_abnormal.return_value = (False, "")
        mock_snr.format_hilink_snr.return_value = ""

        with patch("ascend_fd_tk.core.root_cause.link_builder.get_port_mapping_config_instance") as mock_cfg:
            mock_cfg.return_value.find_port_mapping_by_name.return_value = None
            links = LinkBuilder(cluster)._build_l1_to_l2_links()

        self.assertEqual(len(links), 1)
        link = links[0]
        self.assertEqual(link.l1_switch_id, "ip1")
        self.assertEqual(link.l1_interface, "400G-1")
        self.assertEqual(link.l1_port_speed, PORT_SPEED_400G)
        self.assertEqual(link.l2_switch_id, "ip2")
        self.assertEqual(link.l2_interface, "200G-1")
        self.assertEqual(link.link_status, LINK_STATUS_NORMAL)


if __name__ == "__main__":
    unittest.main()
