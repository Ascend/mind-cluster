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

"""ClusterInfoCache 单元测试：排序、合并、代际阈值与查找。"""

import unittest

from ascend_fd_tk.core.model.bmc import BmcInfo
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.host import HostInfo
from ascend_fd_tk.core.config.threshold_config import OpticalThreshold800G, BaseThreshold
from ascend_fd_tk.core.model.switch import SwitchInfo


class TestClusterInfoCache(unittest.TestCase):
    def _build_cache(self, generation="A3"):
        hosts = {
            "h2": HostInfo("2", "sn-2", server_index="idx-2"),
            "h1": HostInfo("1", "sn-1", server_index="idx-1"),
        }
        bmcs = {"b2": BmcInfo("bmc-2", "sn-2"), "b1": BmcInfo("bmc-1", "sn-1")}
        swis = {
            "sw2": SwitchInfo("sw-2", "10.0.0.2"),
            "sw1": SwitchInfo("sw-1", "10.0.0.1"),
        }
        return ClusterInfoCache(hosts_info=hosts, bmcs_info=bmcs, swis_info=swis, generation=generation)

    def test_sort_info(self):
        cache = self._build_cache()
        cache.sort_info()
        self.assertEqual(list(cache.hosts_info), ["h1", "h2"])
        self.assertEqual(list(cache.bmcs_info), ["b1", "b2"])
        self.assertEqual(list(cache.swis_info), ["sw1", "sw2"])

    def test_update_merges_and_overrides_generation(self):
        cache = self._build_cache("A3")
        other = ClusterInfoCache(
            hosts_info={"h3": HostInfo("3", "sn-3")},
            generation="A5",
        )
        cache.update(other)
        self.assertIn("h3", cache.hosts_info)
        self.assertEqual(cache.generation, "A5")

    def test_update_keeps_generation_when_none(self):
        cache = self._build_cache("A3")
        other = ClusterInfoCache(hosts_info={})
        cache.update(other)
        self.assertEqual(cache.generation, "A3")

    def test_get_threshold_by_generation(self):
        self.assertEqual(self._build_cache("A3").get_threshold(), BaseThreshold)
        self.assertEqual(self._build_cache("A5").get_threshold(), OpticalThreshold800G)
        self.assertEqual(self._build_cache(None).get_threshold(), BaseThreshold)

    def test_find_host_info_by_server_spod_id(self):
        cache = self._build_cache()
        host_info = cache.find_host_info_by_server_spod_id("idx-2")
        self.assertEqual(host_info.host_id, "2")
        self.assertIsNone(cache.find_host_info_by_server_spod_id("missing"))

    def test_find_bmc_info_by_sn_num(self):
        cache = self._build_cache()
        bmc_info = cache.find_bmc_info_by_sn_num("sn-1")
        self.assertEqual(bmc_info.bmc_id, "bmc-1")

    def test_find_host_info_by_sn_num(self):
        cache = self._build_cache()
        host_info = cache.find_host_info_by_sn_num("sn-1")
        self.assertEqual(host_info.host_id, "1")

    def test_init_diag_data_builds_swi_name_map(self):
        cache = self._build_cache()
        cache.init_diag_data()
        self.assertEqual(cache.find_peer_swi("sw-1").swi_id, "10.0.0.1")
        self.assertIsNone(cache.find_peer_swi("not-exist"))


if __name__ == "__main__":
    unittest.main()
