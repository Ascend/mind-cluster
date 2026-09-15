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
from unittest.mock import MagicMock

from ascend_fd_tk.core.fault_analyzer.optical_fault_check import OpticalFaultChecker
from ascend_fd_tk.core.model.diag_result import HostDomain


class TestOpticalFaultChecker(unittest.TestCase):
    def setUp(self):
        self.threshold = SimpleNamespace(HOST_SNR_DB=object(), MEDIA_SNR_DB=object(), TX_BIAS_MA=object())
        self.checker = OpticalFaultChecker(self.threshold)
        self.domain = HostDomain(
            host_id="host-01",
            npu_id="0",
            chip_phy_id="4",
            peer_switch_id="switch-01",
            peer_interface="100GE1/0/1",
        )

    @staticmethod
    def _optical_info(power=([], []), snr="", bias=""):
        info = MagicMock()
        info.get_abnormal_power_infos.return_value = power
        info.get_abnormal_snr_infos.return_value = snr
        info.get_abnormal_bias_infos.return_value = bias
        return info

    def test_single_ended_checks_report_power_snr_and_bias(self):
        info = self._optical_info(power=(["rx lane0 low"], []), snr="snr lane0 low", bias="bias lane0 high")

        power = self.checker.power_analyze_single_ended(self.domain, info)
        snr = self.checker.snr_analyze_single_ended(self.domain, info)
        bias = self.checker.bias_analyze_single_ended(self.domain, info)

        self.assertEqual(len(power), 1)
        self.assertIn("本端收光异常", power[0].fault_info)
        self.assertIn("对端[交换机:switch-01->交换机端口:100GE1/0/1]：NA", power[0].fault_info)
        self.assertEqual(len(snr), 1)
        self.assertIn("snr lane0 low", snr[0].fault_info)
        self.assertEqual(len(bias), 1)
        self.assertIn("bias lane0 high", bias[0].fault_info)

    def test_double_ended_checks_choose_rule_from_both_ends(self):
        local = self._optical_info(power=(["local rx low"], []), snr="local snr low", bias="")
        remote = self._optical_info(power=([], ["remote tx low"]), snr="", bias="remote bias high")

        power = self.checker.power_analyze(self.domain, local, remote)
        snr = self.checker.snr_analyze(self.domain, local, remote)
        bias = self.checker.bias_analyze(self.domain, local, remote)

        self.assertEqual(len(power), 1)
        self.assertIn("本端收光异常，对端发光异常", power[0].fault_info)
        self.assertEqual(len(snr), 1)
        self.assertIn("本端光模块信噪比异常", snr[0].fault_info)
        self.assertEqual(len(bias), 1)
        self.assertIn("对端光模块电流异常", bias[0].fault_info)

    def test_all_checks_skip_normal_values(self):
        local = self._optical_info()
        remote = self._optical_info()

        self.assertEqual(self.checker.power_analyze_single_ended(self.domain, local), [])
        self.assertEqual(self.checker.snr_analyze_single_ended(self.domain, local), [])
        self.assertEqual(self.checker.bias_analyze_single_ended(self.domain, local), [])
        self.assertEqual(self.checker.power_analyze(self.domain, local, remote), [])
        self.assertEqual(self.checker.snr_analyze(self.domain, local, remote), [])
        self.assertEqual(self.checker.bias_analyze(self.domain, local, remote), [])

    def test_rule_maps_cover_every_nonzero_boolean_combination(self):
        for bits in range(1, 16):
            values = tuple(bool(bits & (1 << shift)) for shift in (3, 2, 1, 0))
            self.assertTrue(all(self.checker._check_power_value(*values)))
        for bits in range(1, 4):
            values = tuple(bool(bits & (1 << shift)) for shift in (1, 0))
            self.assertTrue(all(self.checker._check_single_power_value(*values)))
            self.assertTrue(all(self.checker._check_snr_value(*values)))
            self.assertTrue(all(self.checker._check_bias_value(*values)))


if __name__ == "__main__":
    unittest.main()
