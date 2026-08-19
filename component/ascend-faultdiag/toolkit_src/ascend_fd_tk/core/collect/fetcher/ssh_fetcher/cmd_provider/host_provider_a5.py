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


class HostCmdProviderA5(HostBaseProvider):
    def optical_top_headline(self, npu_id) -> str:
        # 查询光模块的头版头条信息
        return f"hccn_tool -g -optical -i {npu_id}"

    def optical_info_cmd(self, npu_id, optical_id) -> str:
        # 查询光模块信息
        return f"hccn_tool -g -optical -i {npu_id} -optical_id {optical_id}"

    def nic_info_cmd(self) -> str:
        # 查询所有 hinic 网卡列表
        return "hinicadm5 info"

    def nic_port_num_cmd(self, card_name) -> str:
        # 查询指定网卡的端口数
        return f"hinicadm5 info -i {card_name}"

    def nic_sfp_cmd(self, card_name, port_id) -> str:
        # 查询指定网卡指定端口的 SFP 信息
        return f"hinicadm5 sfp -i {card_name} -p {port_id}"
