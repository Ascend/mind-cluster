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

import bisect
import json
import re
from typing import List, Dict

from diag_tool.core.common.json_obj import JsonObj
from diag_tool.core.common.path import ConfigPath

_UTF8_ENCODING = "utf-8"

_SEPARATOR = "-"
_SWI_PORT_SEPARATOR = "/"


class L1GlobalAddrMapping(JsonObj):

    def __init__(self, server_id="", npu_id="", chip_id="", addr_start="", addr_end=""):
        self.server_id = server_id
        self.npu_id = npu_id
        self.chip_id = chip_id
        self.addr_start = addr_start
        self.addr_end = addr_end

    @classmethod
    def from_config(cls, server_npu_chip_id: str, addr_range: str):
        server_id, npu_id, chip_id = server_npu_chip_id.split(_SEPARATOR)
        addr_start, addr_end = addr_range.split(_SEPARATOR)
        return cls(server_id, npu_id, chip_id, addr_start, addr_end)


class L1LocalAddrMapping(JsonObj):

    def __init__(self, xpu="", xpu_id="", chip_id="", addr_start="", addr_end=""):
        self.xpu = xpu
        self.xpu_id = xpu_id
        self.chip_id = chip_id
        self.addr_start = addr_start
        self.addr_end = addr_end

    @classmethod
    def from_config(cls, xpu_chip_id: str, addr_range: str):
        xpu, xpu_id, chip_id = xpu_chip_id.split(_SEPARATOR)
        addr_start, addr_end = addr_range.split(_SEPARATOR)
        return cls(xpu, xpu_id, chip_id, addr_start, addr_end)


class L1InterfacePortMapping(JsonObj):
    _XPU_PATTERN = re.compile(r"(.PU)(\d)_\w(\w)")
    _NPU = "NPU"

    def __init__(self, swi_port="", swi_chip_id="", phy_id="", logic_id="", xpu="", xpu_id="", chip_id="", macro=""):
        self.swi_port = swi_port
        self.swi_chip_id = swi_chip_id
        self.phy_id = phy_id
        self.logic_id = logic_id
        self.xpu = xpu
        self.xpu_id = xpu_id
        self.chip_id = chip_id
        self.macro = macro

    @classmethod
    def from_config(cls, port: str, id_group: str):
        phy_id, logic_id, xpu, macro = id_group.split(_SEPARATOR)
        search = cls._XPU_PATTERN.search(xpu)
        if search:
            xpu, xpu_id, chip_id = search.groups()
            if xpu == cls._NPU and xpu_id.isdigit():
                xpu_id = str(int(xpu_id) - 1)
        else:
            xpu, xpu_id, chip_id = "", "", ""
        swi_chip_id = str(int(port.split(_SWI_PORT_SEPARATOR)[1]) - 1)
        return cls(port, swi_chip_id, phy_id, logic_id, xpu.lower(), xpu_id, chip_id, macro)


class PortMappingConfig(JsonObj):
    _MAX_LOCAL_ADDR = 0x2fffffffffff

    def __init__(self, l1_global_addr_mappings: List[L1GlobalAddrMapping] = None,
                 l1_local_addr_mappings: List[L1LocalAddrMapping] = None,
                 l1_interface_port_map: Dict[str, L1InterfacePortMapping] = None):
        self.l1_global_addr_mappings = sorted(l1_global_addr_mappings or [], key=lambda m: m.addr_start)
        self.l1_local_addr_mappings = sorted(l1_local_addr_mappings or [], key=lambda m: m.addr_start)
        self.l1_interface_port_map = l1_interface_port_map
        self._global_addr_start_list = [int(item.addr_start, 16) for item in self.l1_global_addr_mappings]
        self._local_addr_start_list = [int(item.addr_start, 16) for item in self.l1_local_addr_mappings]

    @classmethod
    def parse(cls):
        with open(ConfigPath.L1_GLOBAL_ADDR_CONFIG_PATH, "r", encoding=_UTF8_ENCODING) as f:
            l1_global_addr_mapping_config = json.loads(f.read()) or {}
            l1_global_addr_mappings = []
            for k, v in l1_global_addr_mapping_config.items():
                l1_global_addr_mappings.append(L1GlobalAddrMapping.from_config(k, v))
        with open(ConfigPath.L1_LOCAL_ADDR_CONFIG, "r", encoding=_UTF8_ENCODING) as f:
            l1_local_addr_mapping_config = json.loads(f.read()) or {}
            l1_local_addr_mappings = []
            for k, v in l1_local_addr_mapping_config.items():
                l1_local_addr_mappings.append(L1LocalAddrMapping.from_config(k, v))
        with open(ConfigPath.L1_INTERFACE_PORT_CONFIG, "r", encoding=_UTF8_ENCODING) as f:
            l1_interface_port_mapping_config = json.loads(f.read()) or {}
            l1_interface_port_map = {}
            for k, v in l1_interface_port_mapping_config.items():
                l1_interface_port_map[k] = L1InterfacePortMapping.from_config(k, v)
        return cls(l1_global_addr_mappings, l1_local_addr_mappings, l1_interface_port_map)

    def is_local_addr(self, addr):
        try:
            addr = int(addr, 16)
        except ValueError:
            addr = 0
        return addr <= self._MAX_LOCAL_ADDR

    def find_global_addr_port_mapping(self, addr: str) -> L1GlobalAddrMapping:
        idx = bisect.bisect_right(self._global_addr_start_list, int(addr, 16))
        if idx < 1 or idx > len(self.l1_global_addr_mappings):
            return None
        return self.l1_global_addr_mappings[idx - 1]

    def find_local_addr_port_mapping(self, addr: str) -> L1LocalAddrMapping:
        idx = bisect.bisect_right(self._local_addr_start_list, int(addr, 16))
        if idx < 1 or idx > len(self.l1_local_addr_mappings):
            return None
        return self.l1_local_addr_mappings[idx - 1]

    def find_swi_port(self, swi_chip_id: str, phy_id="", logic_id="") -> L1InterfacePortMapping:
        swi_port_chip_idx = str(int(swi_chip_id) + 1)
        for port, port_mapping in self.l1_interface_port_map.items():
            port_parts = port.split(_SWI_PORT_SEPARATOR)
            if len(port_parts) != 3 or port_parts[1] != swi_port_chip_idx:
                continue
            if port_mapping.phy_id == phy_id or port_mapping.logic_id == logic_id:
                return port_mapping
        return None

    def find_interface(self, local_interface: str, npu_id: str, die_id: str) -> str:
        local_interface_parts = local_interface.split(_SWI_PORT_SEPARATOR)
        if len(local_interface_parts) != 3:
            return ""
        swi_port_chip_idx = local_interface_parts[1]
        for port, port_mapping in self.l1_interface_port_map.items():
            port_parts = port.split(_SWI_PORT_SEPARATOR)
            if len(port_parts) != 3 or port_parts[1] != swi_port_chip_idx:
                continue
            if port_mapping.xpu_id == npu_id or port_mapping.chip_id == die_id:
                return port
        return ""

    def find_swi_port_by_cpu_macro(self, cpu_id: str, chip_id: str, macro: str) -> str:
        for port, port_mapping in self.l1_interface_port_map.items():
            if port_mapping.xpu_id == cpu_id and port_mapping.chip_id == chip_id and port_mapping.macro == macro:
                return port
        return ""

    def find_port_mapping_by_name(self, interface_name: str) -> L1InterfacePortMapping:
        for port, port_mapping in self.l1_interface_port_map.items():
            if port == interface_name:
                return port_mapping
        return None


_PORT_MAPPING_CONFIG = None


def get_port_mapping_config_instance() -> PortMappingConfig:
    global _PORT_MAPPING_CONFIG
    if not _PORT_MAPPING_CONFIG:
        _PORT_MAPPING_CONFIG = PortMappingConfig.parse()
    return _PORT_MAPPING_CONFIG
