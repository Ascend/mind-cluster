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

"""交换机侧光模块类型推导测试：transceiver_type 后缀为 _LPO 即 LPO，否则默认 ODSP"""

import unittest

from ascend_fd_tk.core.model.switch import CommonInfo, InterfaceFullInfo, TransceiverInfo


class TestCommonInfoGetOpticalType(unittest.TestCase):
    """CommonInfo.get_optical_type：按 transceiver_type 后缀推导光模块类型"""

    def test_lpo_suffix(self):
        self.assertEqual(CommonInfo("400G_LPO").get_optical_type(), "LPO")
        self.assertEqual(CommonInfo("ODSP_400G_LPO").get_optical_type(), "LPO")

    def test_default_odsp(self):
        """无 _LPO 后缀一律默认 ODSP"""
        self.assertEqual(CommonInfo("400G_ODSP").get_optical_type(), "ODSP")
        self.assertEqual(CommonInfo("QSFP-DD").get_optical_type(), "ODSP")
        self.assertEqual(CommonInfo("").get_optical_type(), "ODSP")
        self.assertEqual(CommonInfo().get_optical_type(), "ODSP")

    def test_case_insensitive_and_blank(self):
        """大小写不敏感、忽略首尾空白"""
        self.assertEqual(CommonInfo("400G_lpo").get_optical_type(), "LPO")
        self.assertEqual(CommonInfo(" 400G_LPo ").get_optical_type(), "LPO")


class TestInterfaceFullInfoOpticalType(unittest.TestCase):
    """InterfaceFullInfo.get_optical_module_info：光模块类型随 transceiver_info.common_info 填充"""

    def _build(self, transceiver_type: str) -> InterfaceFullInfo:
        transceiver = TransceiverInfo(interface="Eth1/1", common_information=CommonInfo(transceiver_type))
        return InterfaceFullInfo(interface="Eth1/1", transceiver_info=transceiver)

    def test_optical_type_from_common_info(self):
        info = self._build("400G_LPO").get_optical_module_info()
        self.assertEqual(info.optical_type, "LPO")
        info = self._build("400G_ODSP").get_optical_module_info()
        self.assertEqual(info.optical_type, "ODSP")

    def test_no_transceiver_info(self):
        """无 transceiver_info 时类型为空（回退基础阈值，语义等同 ODSP）"""
        info = InterfaceFullInfo(interface="Eth1/1").get_optical_module_info()
        self.assertIsNone(info)


if __name__ == "__main__":
    unittest.main()
