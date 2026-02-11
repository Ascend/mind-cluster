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

import re

from toolkit.core.common.constants import CMD_EXEC_SUCCESS
from toolkit.core.model.host import HCCNOpticalInfo, HCCNLinkStatInfo, HCCNLinkHistory, HCCNStatInfo, HCCNLLDPInfo, \
    HccnPortHccsInfo, SpodInfo, CdrSnrInfo, HCCNDfxCfgInfo
from toolkit.utils.form_parser import FormParser


class HostParser:
    _NET_HEALTH_PATTERN = re.compile(r"net health status: (.+?)$")
    _LINK_STATUS_PATTERN = re.compile(r"link status: (.+?)$")
    _RECORD_PATTERN = re.compile(r'\[devid \d+\]\s+(\w+\s+\w+\s+\d+\s+\d+:\d+:\d+\s+\d+)\s+(LINK UP|LINK DOWN)')

    @classmethod
    def parse_optical_info(cls, cmd_res: str) -> HCCNOpticalInfo:
        if "Control link unreachable" in cmd_res:
            return HCCNOpticalInfo(control_link_unreachable=True)
        form = FormParser().parse(cmd_res)
        if not form:
            return None
        return HCCNOpticalInfo.from_dict(form)

    @classmethod
    def parse_link_stat_info(cls, cmd_res: str) -> HCCNLinkStatInfo:
        link_up_count = 0
        link_down_count = 0
        link_records = []
        current_time = ""
        current_time_match = re.search(r'current time\s+:\s+(\w+\s+\w+\s+\d+\s+\d+:\d+:\d+\s+\d+)', cmd_res)
        if current_time_match:
            current_time = current_time_match.group(1)
        up_match = re.search(r'link up count\s+:\s+(\d+)', cmd_res)
        if up_match:
            link_up_count = int(up_match.group(1))
        down_match = re.search(r'link down count\s+:\s+(\d+)', cmd_res)
        if down_match:
            link_down_count = int(down_match.group(1))
        matches = cls._RECORD_PATTERN.findall(cmd_res)
        for time_str, status in matches:
            link_records.append(HCCNLinkHistory(time_str, status))
        return HCCNLinkStatInfo(current_time, link_up_count, link_down_count, link_records)

    @classmethod
    def parse_stat_info(cls, cmd_res: str) -> HCCNStatInfo:
        return HCCNStatInfo.from_dict(FormParser().parse(cmd_res))

    @classmethod
    def parse_lldp_info(cls, cmd_res: str) -> HCCNLLDPInfo:
        if not cmd_res:
            return HCCNLLDPInfo(None, None)
        port_id_tlv, system_name_tlv = None, None
        lines = cmd_res.strip().split('\n')
        # 返回第一行非空行作为序列号
        for i, line in enumerate(lines):  # 跳过第一行标题行
            line = line.strip()
            if 'Port ID TLV' in line:
                for j in range(i + 1, min(i + 3, len(lines))):
                    if 'Ifname:' in lines[j]:
                        port_id_tlv = lines[j].split('Ifname:')[1].strip()
                        break
            elif 'System Name TLV' in line:
                if i + 1 < len(lines):
                    system_name_tlv = lines[i + 1].strip()
        return HCCNLLDPInfo(port_id_tlv, system_name_tlv)

    @classmethod
    def parse_hccs_info(cls, cmd_res: str) -> HccnPortHccsInfo:
        return HccnPortHccsInfo.from_dict(FormParser().parse(cmd_res))

    @classmethod
    def parse_roce_speed(cls, cmd_res: str) -> str:
        lines = cmd_res.strip().splitlines()
        if not lines:
            return ""
        parts = lines[-1].split()
        if len(parts) >= 3 and "Speed" in parts[0]:
            return parts[1].strip()
        return ""

    @classmethod
    def parse_spod_info(cls, cmd_res: str) -> SpodInfo:
        if not cmd_res.strip():
            return None
        return SpodInfo.from_dict(FormParser().parse(cmd_res))

    @classmethod
    def parse_npu_type(cls, cmd_res) -> str:
        match = re.search(r'Device (d80\d+)', cmd_res)
        if match:
            return match.group(1)
        return ""

    @classmethod
    def parse_optical_loopback_enable(cls, cmd_res) -> bool:
        if cmd_res.stdout and CMD_EXEC_SUCCESS in cmd_res.stdout:
            return True
        return False

    @classmethod
    def parse_hccn_tool_net_health(cls, cmd_res: str) -> str:
        search = cls._NET_HEALTH_PATTERN.search(cmd_res)
        if search:
            return search.group(1)
        return ""

    @classmethod
    def parse_hccn_tool_link_status(cls, cmd_res: str) -> str:
        search = cls._LINK_STATUS_PATTERN.search(cmd_res)
        if search:
            return search.group(1)
        return ""

    @classmethod
    def parse_hccn_tool_cdr(cls, cmd_res: str) -> CdrSnrInfo:
        if not cmd_res or not cmd_res.strip():
            return None
        return CdrSnrInfo.from_dict(FormParser().parse(cmd_res))

    @classmethod
    def parse_hccn_dfx_cfgr(cls, cmd_res: str) -> HCCNDfxCfgInfo:
        if not cmd_res or not cmd_res.strip():
            return None
        return HCCNDfxCfgInfo.from_dict(FormParser().parse(cmd_res))
