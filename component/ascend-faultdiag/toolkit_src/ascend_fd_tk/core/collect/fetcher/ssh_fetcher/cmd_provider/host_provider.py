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

from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.cmd_provider.host_base import HostBaseProvider


class HostCmdProvider(HostBaseProvider):
    """A3命令提供"""

    def link_stat_cmd(self, chip_phy_id) -> str:
        return "hccn_tool -i {} -link_stat -g".format(chip_phy_id)

    def stat_cmd(self, chip_phy_id) -> str:
        return "hccn_tool -i {} -stat -g".format(chip_phy_id)

    def lldp_cmd(self, chip_phy_id) -> str:
        return "hccn_tool -i {} -lldp -g".format(chip_phy_id)

    def speed_cmd(self, chip_phy_id) -> str:
        return f"hccn_tool -i {chip_phy_id} -speed -g"

    def duplex_cmd(self, chip_phy_id) -> str:
        return f"hccn_tool -i {chip_phy_id} -duplex -g"

    def net_health_cmd(self, chip_phy_id) -> str:
        return f"hccn_tool -i {chip_phy_id} -net_health -g"

    def hccn_link_cmd(self, chip_phy_id) -> str:
        return f"hccn_tool -i {chip_phy_id} -link -g"

    def cdr_snr_cmd(self, chip_phy_id) -> str:
        return f"hccn_tool -i {chip_phy_id} -scdr -t 5"

    def dfx_cfg_cmd(self, chip_phy_id) -> str:
        return f"hccn_tool -i {chip_phy_id} -optical -g dfx_cfg"

    def optical_cmd(self, chip_phy_id) -> str:
        return f"hccn_tool -i {chip_phy_id} -optical -g"

    def optical_loopback_enable_cmd(self, npu_id, model) -> str:
        return f"hccn_tool -i {npu_id} -optical -t {model}"
