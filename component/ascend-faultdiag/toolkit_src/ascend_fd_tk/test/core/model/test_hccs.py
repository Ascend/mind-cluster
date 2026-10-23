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

"""hccs 数据模型单元测试：路由 miss、超时统计、链路状态与 lane 告警。"""

import unittest

from ascend_fd_tk.core.model.hccs import (
    HccsInfo,
    HccsPortStatistic,
    HccsRouteMiss,
    InterfaceProxyResponseDetail,
    InterfaceSnr,
    LCNEInfo,
    LaneSnr,
    PortLinkStatusRecord,
    ProxyTimeoutStatis,
)


class TestHccsSmallModels(unittest.TestCase):
    def test_route_miss_and_timeout_flags(self):
        self.assertTrue(HccsRouteMiss(lp_direction_miss="3").is_lp_direction_miss())
        self.assertFalse(HccsRouteMiss(lp_direction_miss="0").is_lp_direction_miss())
        st = ProxyTimeoutStatis(rp_rx="1", rp_tx="2", lp_tx="3")
        self.assertTrue(st.is_rx_timeout_happend())
        self.assertTrue(st.is_rp_tx_timeout_happend())
        self.assertTrue(st.is_lp_tx_timeout_happend())
        st_default = ProxyTimeoutStatis()
        self.assertFalse(st_default.is_rx_timeout_happend())
        self.assertFalse(st_default.is_rp_tx_timeout_happend())
        self.assertFalse(st_default.is_lp_tx_timeout_happend())

    def test_port_record_time(self):
        record = PortLinkStatusRecord(record="[10:30:15] link up")
        self.assertEqual(record.record_time, "[10:30:15]")
        self.assertEqual(PortLinkStatusRecord(record="link up without time").record_time, "")

    def test_lane_snr_to_str(self):
        self.assertEqual(str(LaneSnr("lane0", "30")), "lane0:30")
        snr = InterfaceSnr(interface_name="eth0", abnormal_lane_snr=[LaneSnr("lane0", "30"), LaneSnr("lane1", "25")])
        self.assertEqual(snr.lane_snr_to_str(), "['lane0:30', 'lane1:25']")


class TestLCNEInfo(unittest.TestCase):
    def test_str(self):
        info = LCNEInfo(server_id="1", xpu="A3", npu_id="2", die_id="3", port_id="4", interface="eth0")
        text = str(info)
        self.assertIn("server-1", text)
        self.assertIn("A3-2", text)
        self.assertIn("port-4", text)

    def test_link_down_time_and_lane_width(self):
        info = LCNEInfo(link_status=[PortLinkStatusRecord(record="[10:30:15] link down")])
        self.assertEqual(info.get_first_link_down_time(), "10:30:15")
        self.assertEqual(LCNEInfo().get_first_link_down_time(), "")
        self.assertEqual(
            LCNEInfo(link_status=[PortLinkStatusRecord(record="[10:30:15] link up")]).get_first_link_down_time(), ""
        )
        self.assertEqual(
            LCNEInfo(link_status=[PortLinkStatusRecord(record="link down no bracket")]).get_first_link_down_time(), ""
        )
        info_reduced = LCNEInfo(link_status=[PortLinkStatusRecord(record="[10:00:00] link up,width:2")])
        self.assertEqual(info_reduced.get_lane_width(), 2)
        self.assertTrue(info_reduced.is_lane_error())
        info_full = LCNEInfo(link_status=[PortLinkStatusRecord(record="[10:00:00] link up,width:4")])
        self.assertEqual(info_full.get_lane_width(), 4)
        self.assertFalse(info_full.is_lane_error())
        self.assertEqual(LCNEInfo().get_lane_width(), 4)

    def test_route_miss_and_pack_block(self):
        info = LCNEInfo(lp_direction_miss=1, rp_id_using_cnt=1, lp_id_using_cnt=1, cur_pkt=1)
        self.assertTrue(info.is_lp_route_miss())
        self.assertTrue(info.is_rp_pack_block())
        self.assertTrue(info.is_lp_pack_block())
        self.assertTrue(info.is_voq_pack_block())
        self.assertFalse(LCNEInfo().is_lp_route_miss())
        self.assertFalse(LCNEInfo().is_voq_pack_block())


class TestHccsInfo(unittest.TestCase):
    def test_route_miss_and_link_status(self):
        info = HccsInfo(hccs_route_miss=[HccsRouteMiss(interface="eth0", lp_direction_miss="5")])
        self.assertEqual(info.get_route_miss_by_interface("eth0").lp_direction_miss, 5)
        self.assertEqual(info.get_route_miss_by_interface("missing").interface, "")
        info = HccsInfo(
            link_status_records=[
                PortLinkStatusRecord(chip="0", port="1", record="a"),
                PortLinkStatusRecord(chip="0", port="1", record="b"),
                PortLinkStatusRecord(chip="0", port="2", record="c"),
            ]
        )
        self.assertEqual(len(info.get_link_status_by_chip_port("0", "1")), 2)
        self.assertEqual(info.get_link_status_by_chip_port("9", "9"), [])

    def test_package_block_and_timeout_detail(self):
        info = HccsInfo(
            hccs_port_statistic=[
                HccsPortStatistic(chip="0", port="1", proxy_module="m", dfx_state_name="s", dfx_result="blocked")
            ]
        )
        self.assertEqual(info.get_package_block_by_condition("0", "1", "m", "s"), "blocked")
        self.assertEqual(info.get_package_block_by_condition("0", "1", "m", "x"), "")
        info = HccsInfo(
            proxy_response_detail_interfaces=[
                InterfaceProxyResponseDetail(
                    interface="eth0", proxy_type="rx", response_type="timeout", address="0x10"
                ),
                InterfaceProxyResponseDetail(
                    interface="eth0", proxy_type="tx", response_type="timeout", address="0x20"
                ),
                InterfaceProxyResponseDetail(
                    interface="eth1", proxy_type="rx", response_type="timeout", address="0x30"
                ),
            ]
        )
        self.assertEqual(info.get_timeout_detail_by_condition("eth0", "rx", "timeout"), ["0x10"])
        self.assertEqual(info.get_timeout_detail_by_condition("eth0", "tx", "timeout"), ["0x20"])
        self.assertEqual(info.get_timeout_detail_by_condition("missing", "rx", "timeout"), [])


if __name__ == "__main__":
    unittest.main()
