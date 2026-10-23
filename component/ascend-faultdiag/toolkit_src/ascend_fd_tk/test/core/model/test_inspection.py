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

"""巡检数据模型单元测试：接口描述与 CSV 行转换。"""

import unittest

from ascend_fd_tk.core.model.inspection import InspectionErrorItem, InspectionInterfaceInfo


class TestInspectionInterfaceInfo(unittest.TestCase):
    def test_str(self):
        info = InspectionInterfaceInfo(
            device_name="node0", device_id="1.1.1.1", device_sn="sn01", interface="eth0", interface_sn="mod01"
        )
        text = str(info)
        self.assertIn("node0", text)
        self.assertIn("1.1.1.1", text)
        self.assertIn("eth0", text)
        self.assertIn("mod01", text)


class TestInspectionErrorItem(unittest.TestCase):
    def test_to_csv_dict(self):
        local = InspectionInterfaceInfo(
            device_name="node0", device_id="1.1.1.1", device_sn="sn01", interface="eth0", interface_sn="mod01"
        )
        peer = InspectionInterfaceInfo(
            device_name="sw0", device_id="2.2.2.2", device_sn="sn02", interface="100GE1/1", interface_sn="mod02"
        )
        item = InspectionErrorItem(local_interface=local, peer_interface=peer, fault_desc="los")
        row = item.to_csv_dict()
        self.assertEqual(row["A端设备名称"], "node0")
        self.assertEqual(row["A端IP"], "1.1.1.1")
        self.assertEqual(row["A端接口"], "eth0")
        self.assertEqual(row["A端光模块SN"], "mod01")
        self.assertEqual(row["B端设备名称"], "sw0")
        self.assertEqual(row["B端接口"], "100GE1/1")
        self.assertEqual(row["B端光模块SN"], "mod02")
        self.assertEqual(row["问题现象"], "los")
        self.assertEqual(len(row), 11)


if __name__ == "__main__":
    unittest.main()
