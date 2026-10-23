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

"""主机在线采集命令提供者单元测试：代际无关命令与 A3/A5 专属命令。"""

import unittest

from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.cmd_provider.host_base import HostBaseProvider
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.cmd_provider.host_provider import HostCmdProvider
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.cmd_provider.host_provider_a5 import HostCmdProviderA5


class TestHostBaseProvider(unittest.TestCase):
    def test_generation_independent_commands(self):
        provider = HostBaseProvider()
        self.assertEqual(provider.npu_mapping_cmd(), "npu-smi info -m")
        self.assertEqual(provider.sn_num_cmd(), "dmidecode -s system-serial-number")
        self.assertEqual(provider.hostname_cmd(), "hostname")
        self.assertEqual(provider.hccs_cmd(0, 1), "npu-smi info -t hccs -i 0 -c 1")
        self.assertEqual(provider.spod_cmd(2, 3), "npu-smi info -t spod-info -i 2 -c 3")


class TestHostCmdProvider(unittest.TestCase):
    def test_a3_commands(self):
        provider = HostCmdProvider()
        self.assertEqual(provider.lldp_cmd(0), "hccn_tool -i 0 -lldp -g")
        self.assertEqual(provider.speed_cmd(1), "hccn_tool -i 1 -speed -g")
        self.assertEqual(provider.optical_cmd(2), "hccn_tool -i 2 -optical -g")
        self.assertEqual(provider.link_stat_cmd(3), "hccn_tool -i 3 -link_stat -g")
        self.assertEqual(provider.cdr_snr_cmd(4), "hccn_tool -i 4 -scdr -t 5")
        self.assertEqual(provider.optical_loopback_enable_cmd(5, "1"), "hccn_tool -i 5 -optical -t 1")


class TestHostCmdProviderA5(unittest.TestCase):
    def test_a5_commands(self):
        provider = HostCmdProviderA5()
        self.assertEqual(provider.optical_top_headline(0), "hccn_tool -g -optical -i 0")
        self.assertEqual(provider.optical_port_info_cmd(0, 2, 3), "hccn_tool -g -optical -i 0 -u 2 -p 3")
        self.assertEqual(provider.nic_info_cmd(), "hinicadm5 info")
        self.assertEqual(provider.nic_port_num_cmd("h0"), "hinicadm5 info -i h0")
        self.assertEqual(provider.nic_sfp_cmd("h0", 1), "hinicadm5 sfp -i h0 -p 1")


if __name__ == "__main__":
    unittest.main()
