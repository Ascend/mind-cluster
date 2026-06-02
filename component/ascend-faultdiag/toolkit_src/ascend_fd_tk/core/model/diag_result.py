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

from typing import Union, List

from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.core.common.constants import FAULT_TYPE_BMC, FAULT_TYPE_HOST, FAULT_TYPE_SWITCH


class HostDomain(JsonObj):
    def __init__(
        self,
        host_id: str = "",
        npu_id: str = "",
        chip_phy_id: str = "",
        peer_switch_id: str = "",
        peer_interface: str = "",
    ):
        self.host_id = host_id
        self.npu_id = npu_id
        self.chip_phy_id = chip_phy_id
        self.peer_switch_id = peer_switch_id
        self.peer_interface = peer_interface

    def get_desc(self) -> str:
        return "->".join(self.get_local_domain() + self.get_peer_domain())

    def get_local_domain(self) -> List[str]:
        parts = []
        if self.host_id:
            parts.append(f"服务器:{self.host_id}")
        if self.npu_id:
            parts.append(f"NPU:{self.npu_id}")
        if self.chip_phy_id:
            parts.append(f"chip:{self.chip_phy_id}")
        return parts

    def get_peer_domain(self) -> List[str]:
        parts = []
        if self.peer_switch_id:
            parts.append(f"交换机:{self.peer_switch_id}")
        if self.peer_interface:
            parts.append(f"交换机端口:{self.peer_interface}")
        return parts


class BmcDomain(JsonObj):
    def __init__(self, bmc_id: str = "", npu_id: str = "", chip_phy_id: str = ""):
        self.bmc_id = bmc_id
        self.npu_id = npu_id
        self.chip_phy_id = chip_phy_id

    def get_desc(self) -> str:
        parts = []
        if self.bmc_id:
            parts.append(f"BMC:{self.bmc_id}")
        if self.npu_id:
            parts.append(f"NPU:{self.npu_id}")
        if self.chip_phy_id:
            parts.append(f"chip:{self.chip_phy_id}")
        return "->".join(parts)


class SwitchDomain(JsonObj):
    def __init__(
        self,
        swi_id: str = "",
        interface: str = "",
        peer_switch_id: str = "",
        peer_switch_interface: str = "",
    ):
        self.swi_id = swi_id
        self.interface = interface
        self.peer_switch_id = peer_switch_id
        self.peer_switch_interface = peer_switch_interface

    def get_desc(self) -> str:
        return "->".join(self.get_local_domain() + self.get_peer_domain())

    def get_local_domain(self) -> List[str]:
        parts = []
        if self.swi_id:
            parts.append(f"交换机:{self.swi_id}")
        if self.interface:
            parts.append(f"交换机端口:{self.interface}")
        return parts

    def get_peer_domain(self) -> List[str]:
        parts = []
        if self.peer_switch_id:
            parts.append(f"交换机:{self.peer_switch_id}")
        if self.peer_switch_interface:
            parts.append(f"交换机端口:{self.peer_switch_interface}")
        return parts


class DiagResult(JsonObj):
    def __init__(
        self,
        domain: Union[HostDomain, BmcDomain, SwitchDomain],
        fault_info: str = "",
        suggestion: str = "",
        err_code: str = "",
    ):
        self.domain = domain
        self.fault_info = fault_info
        self.suggestion = suggestion
        self.err_code = err_code

    @property
    def fault_type(self) -> str:
        if isinstance(self.domain, HostDomain):
            return FAULT_TYPE_HOST
        elif isinstance(self.domain, BmcDomain):
            return FAULT_TYPE_BMC
        elif isinstance(self.domain, SwitchDomain):
            return FAULT_TYPE_SWITCH
        return ""

    def get_domain_desc(self) -> str:
        return self.domain.get_desc()

    def to_dict(self):
        return {
            "故障域": str(self.get_domain_desc()),
            "故障码": self.err_code,
            "故障信息": self.fault_info,
            "处理建议": self.suggestion,
            "故障类型": self.fault_type,
        }
