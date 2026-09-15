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

from ascend_fd_tk.core.common.diag_enum import NpuType, OpticalModuleType
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.diag_result import DiagResult, HostDomain
from ascend_fd_tk.core.model.host_a5 import HCCNOpticalInfoA5, NpuChipInfoA5


@register_analyzer(generation=[NpuType.A5])
class HostOpticalLosLoLAnalyzerA5(Analyzer):
    @staticmethod
    def _check_optical_indicator(
        host_id: str, npu_chip_info: NpuChipInfoA5, optical_info_list: List[HCCNOpticalInfoA5]
    ) -> List[DiagResult]:
        results = []
        for optical_info in optical_info_list:
            if not optical_info:
                continue
            # LPO 光模块不支持 Lol 标志，仅检查其余标志，避免无效值误报
            if optical_info.optical_type == OpticalModuleType.LPO.value:
                items = ("TxFault Flag", "TxLos Flag", "RxLos Flag")
            else:
                items = ("TxFault Flag", "TxLos Flag", "RxLos Flag", "TxLol Flag", "RxLol Flag")
            for state in optical_info.state_flag:
                if state.items not in items:
                    continue
                abnormal_lanes = [
                    f"Lane{lane_idx}：{lane_val}"
                    for lane_idx, lane_val in enumerate(state.lanes)
                    if lane_val and lane_val != "Normal"
                ]
                if not abnormal_lanes:
                    continue
                results.append(
                    DiagResult(
                        domain=HostDomain(
                            host_id=host_id,
                            npu_id=npu_chip_info.npu_id,
                            udie_id=optical_info.udie_id,
                            npu_port_id=optical_info.port_id,
                        ),
                        fault_info=f"光模块 {state.items} 指标状态异常：\n" + "\n".join(abnormal_lanes),
                        suggestion="请检查光模块相关指标",
                    )
                )
        return results

    def analyse(self) -> List[DiagResult]:
        results = []
        for host_info in self.cluster_info.hosts_info.values():
            for npu_chip_info in host_info.npu_chip_info.values():
                optical_info_list = npu_chip_info.hccn_optical_info
                if not optical_info_list:
                    continue
                results.extend(
                    self._check_optical_indicator(
                        host_id=host_info.host_id,
                        npu_chip_info=npu_chip_info,
                        optical_info_list=optical_info_list,
                    )
                )
        return results
