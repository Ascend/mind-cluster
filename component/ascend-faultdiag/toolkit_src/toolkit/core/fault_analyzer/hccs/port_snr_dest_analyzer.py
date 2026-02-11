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

from toolkit.core.common import diag_enum, constants
from toolkit.core.common.diag_enum import DeviceType
from toolkit.core.config import port_mapping_config
from toolkit.core.context.register import register_analyzer
from toolkit.core.fault_analyzer.base import Analyzer
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.diag_result import DiagResult, Domain


@register_analyzer
class PortSnrDestAnalyzer(Analyzer):
    _XPU_SNR_LIMIT_MAP = {
        diag_enum.XPU.CPU.value: constants.CHIP_CPU_PORT_SNR_LIMIT,
        diag_enum.XPU.NPU.value: constants.CHIP_NPU_PORT_SNR_LIMIT,
    }

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self.swis_info = {k: v for k, v in cluster_info.swis_info.items() if v.hccs_info}

    def analyse(self) -> List[DiagResult]:
        diag_results = []
        port_mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()
        for swi in self.swis_info.values():
            for port_snr in swi.hccs_info.hccs_chip_port_snr_list:
                port_mapping = port_mapping_config_instance.find_swi_port(port_snr.swi_chip_id, port_snr.port_id)
                if not port_mapping:
                    continue
                domain = [
                    Domain(DeviceType.SWITCH.value, f"{swi.name} {swi.swi_id}"),
                    Domain(DeviceType.SWI_PORT.value, f"{port_mapping.swi_port}"),
                ]
                # TODO 确认SNR类型
                check_res = self.cluster_info.get_threshold().CDR_HOST_SNR_LINE.check_value_str(port_snr.snr)
                if not check_res:
                    continue
                peer_port = ""
                if port_snr.xpu and port_mapping.xpu_id:
                    peer_port = f"对端{port_snr.xpu}{port_mapping.xpu_id}端口, "

                diag_res = DiagResult(domain,
                                      f"{peer_port}lane {port_snr.lane_id} {check_res}",
                                      "请检查端口是否脏污")
                diag_results.append(diag_res)
        return diag_results
