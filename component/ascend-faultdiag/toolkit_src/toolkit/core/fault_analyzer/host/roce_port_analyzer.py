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

from toolkit.core.common import diag_enum
from toolkit.core.context.register import register_analyzer
from toolkit.core.fault_analyzer.base import Analyzer
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.diag_result import DiagResult, Domain
from toolkit.core.model.switch import SwitchInfo
from toolkit.utils import logger

LOGGER = logger.DIAG_LOGGER


@register_analyzer
class RocePortAnalyzer(Analyzer):

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)

    @staticmethod
    def _get_npu_chip_domain(ip, npu_id, chip_id):
        return [Domain(diag_enum.DeviceType.SERVER, ip),
                Domain(diag_enum.DeviceType.NPU, npu_id),
                Domain(diag_enum.DeviceType.CHIP, chip_id), ]

    def analyse(self) -> List[DiagResult]:
        result = []
        for host_info in self.cluster_info.hosts_info.values():
            for npu_chip_info in host_info.npu_chip_info.values():
                peer_roce_swi = npu_chip_info.hccn_lldp_info
                if not peer_roce_swi or not peer_roce_swi.system_name_tlv:
                    diag_result = DiagResult(
                        self._get_npu_chip_domain(host_info.host_id, npu_chip_info.npu_id, npu_chip_info.chip_phy_id),
                        f"未采集到NPU光模块对端lldp信息",
                        f"请检查是否插上光模块或未连接交换机")
                    LOGGER.warning(f"{diag_result.get_domain_desc()} {diag_result.fault_info}")
                    continue
                peer_roce_swi_info = self._find_peer_roce_swi_info(peer_roce_swi.system_name_tlv)
                if not peer_roce_swi_info:
                    LOGGER.warning(
                        f"未采集到[服务器{host_info.host_id}]->[NPU {npu_chip_info.npu_id}]->[chip {npu_chip_info.chip_phy_id}]"
                        f"对端[交换机 {peer_roce_swi.system_name_tlv}]信息")
                    continue
                for peer_roce_port_info in peer_roce_swi_info.interface_info:
                    if peer_roce_port_info.interface_name == peer_roce_swi.port_id_tlv:
                        continue
                    if "auto" in (peer_roce_port_info.speed.lower(), npu_chip_info.speed.lower()):
                        continue
                    if "auto" in (peer_roce_port_info.duplex.lower(), npu_chip_info.duplex.lower()):
                        continue
                    if peer_roce_port_info.speed != npu_chip_info.speed or \
                            peer_roce_port_info.duplex.lower() != npu_chip_info.duplex.lower():
                        diag_result = DiagResult(
                            self._get_npu_chip_domain(host_info.host_id, npu_chip_info.npu_id,
                                                      npu_chip_info.chip_phy_id),
                            f"NPU端口与对端交换机: {peer_roce_swi.system_name_tlv}, ip: {peer_roce_swi_info.swi_id}, 端口{peer_roce_swi.port_id_tlv}连接信息不相同,"
                            f"本端Speed: {npu_chip_info.speed}, Duplex: {npu_chip_info.duplex}. 对端Speed: {peer_roce_port_info.speed}, Duplex: {peer_roce_port_info.duplex}",
                            f"请保持两端设置一致"
                        )
                        result.append(diag_result)
                    break
        return result

    def _find_peer_roce_swi_info(self, peer_roce_swi_name) -> SwitchInfo:
        for swi_info in self.cluster_info.swis_info.values():
            if swi_info.name == peer_roce_swi_name:
                return swi_info
        return None
