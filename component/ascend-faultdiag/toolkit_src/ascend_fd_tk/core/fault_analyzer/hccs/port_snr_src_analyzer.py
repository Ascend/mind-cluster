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

from ascend_fd_tk.core.common import diag_enum

from ascend_fd_tk.core.config import port_mapping_config
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, SwitchDomain


@register_analyzer
class PortSnrSrcAnalyzer(Analyzer):
    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self.swis_info = {k: v for k, v in cluster_info.swis_info.items() if v.hccs_info}
        # pylint: disable=duplicate-code
        self.port_mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()
        self.threshold = cluster_info.get_threshold()
        self.xpu_snr_limit_map = {
            diag_enum.XPU.CPU.value: self.threshold.CHIP_CPU_PORT_SNR_LINE,
            diag_enum.XPU.NPU.value: self.threshold.CHIP_NPU_PORT_SNR_LINE,
        }

    def analyse(self) -> List[DiagResult]:
        diag_results = []
        for swi in self.swis_info.values():
            for port_snr in swi.hccs_info.interface_snr_list:
                port_mapping = self.port_mapping_config_instance.find_port_mapping_by_name(port_snr.interface_name)
                if not port_mapping:
                    continue
                th = self.xpu_snr_limit_map.get(port_mapping.xpu, self.threshold.SWITCH_PORT_SNR_LINE)
                peer_port = ""
                if port_mapping.xpu and port_mapping.xpu_id:
                    peer_port = f"（对端{port_mapping.xpu}{port_mapping.xpu_id}）"
                fault_info_list = [f"光模块端口 {port_mapping.swi_port} Hilink SNR 异常{peer_port}："]
                for abnormal_lane_snr in port_snr.abnormal_lane_snr:
                    check_res = th.check_value_str(abnormal_lane_snr.snr_value)
                    if not check_res:
                        continue
                    fault_info_list.append(f"  {abnormal_lane_snr.lane_name} {check_res}")
                if len(fault_info_list) <= 1:
                    continue
                diag_results.append(
                    DiagResult(
                        domain=SwitchDomain(swi_id=swi.swi_id, interface=port_snr.interface_name),
                        fault_info="\n".join(fault_info_list),
                        suggestion="请对链路进行排查，并检查端口是否脏污",
                    )
                )
        return diag_results
