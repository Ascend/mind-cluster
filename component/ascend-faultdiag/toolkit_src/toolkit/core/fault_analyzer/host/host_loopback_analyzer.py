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
from toolkit.core.fault_analyzer.base import Analyzer
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.diag_result import DiagResult, Domain
from toolkit.core.model.host import HostInfo


class HostLoopbackAnalyzer(Analyzer):

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)

    @staticmethod
    def get_npu_chip_domain(ip, npu_id, chip_id):
        return [Domain(diag_enum.DeviceType.SERVER, ip),
                Domain(diag_enum.DeviceType.NPU, npu_id),
                Domain(diag_enum.DeviceType.CHIP, chip_id), ]

    def analyse(self) -> List[DiagResult]:
        hosts_info = self.cluster_info.hosts_info
        diag_results = []
        for _, host_info in hosts_info.items():
            loopback_results = self.host_loopback_diag(host_info)
            diag_results.extend(loopback_results)
        return diag_results

    def host_loopback_diag(self, host_info: HostInfo):
        results = []
        if not host_info or not host_info.loopback_info_list:
            return results
        for loopback_info in host_info.loopback_info_list:
            if not loopback_info.host_input_enable:
                return results
            if not loopback_info.host_input_link_stat.is_first_record_up():
                results.append(DiagResult(
                    self.get_npu_chip_domain(host_info.host_id, loopback_info.npu_id, loopback_info.chip_phy_id),
                    "本端环回类型1后端口down，诊断为本端故障",
                    "建议交叉模组、线缆，如果有CDR板需要做CDR环回"
                ))
            if loopback_info.media_output_enable and loopback_info.media_output_link_stat.is_first_record_up():
                results.append(DiagResult(
                    self.get_npu_chip_domain(host_info.host_id, loopback_info.npu_id, loopback_info.chip_phy_id),
                    "本端环回类型1后端口up，环回类型2后端口down，诊断为本端端口光模块故障/赃污",
                    "建议排查本端端口光模块故障/赃污"
                ))
        return results
