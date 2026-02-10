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

from typing import List

from diag_tool.core.common.diag_enum import DeviceType
from diag_tool.core.context.register import register_analyzer
from diag_tool.core.fault_analyzer.base import Analyzer
from diag_tool.core.model.cluster_info_cache import ClusterInfoCache
from diag_tool.core.model.diag_result import DiagResult, Domain
from diag_tool.utils import logger

_DIAG_LOGGER = logger.DIAG_LOGGER


@register_analyzer
class HostOpticalStatusAnalyzer(Analyzer):
    _NET_HEALTH_STATUS = "Success"
    _UP_STATUS = "UP"
    _FAULT_DESC = "端口光模块状态异常, 网络健康状态: {}, 连接状态: {}."
    _PEER_PORT_INFO = " 对端交换机: {}, 对端端口: {}."
    _SUGGESTION = "请检查NPU端口, 或对端端口状态"

    def analyse(self) -> List[DiagResult]:
        result = []
        for host_info in self.cluster_info.hosts_info.values():
            for chip_info in host_info.npu_chip_info.values():
                is_healthy = chip_info.net_health and chip_info.net_health != self._NET_HEALTH_STATUS
                is_up = chip_info.link_status and chip_info.link_status != self._UP_STATUS
                if is_healthy or is_up:
                    domain = [
                        Domain(DeviceType.SERVER, host_info.host_id),
                        Domain(DeviceType.NPU, chip_info.npu_id),
                        Domain(DeviceType.CHIP, chip_info.chip_phy_id),
                    ]
                    fault_desc = self._FAULT_DESC.format(chip_info.net_health, chip_info.link_status or 'NA')
                    hccn_lldp_info = chip_info.hccn_lldp_info
                    if hccn_lldp_info and hccn_lldp_info.system_name_tlv:
                        fault_desc += self._PEER_PORT_INFO.format(hccn_lldp_info.system_name_tlv,
                                                                  hccn_lldp_info.port_id_tlv)
                    diag_result = DiagResult(domain, fault_info=fault_desc, suggestion=self._SUGGESTION)
                    result.append(diag_result)
        return result
