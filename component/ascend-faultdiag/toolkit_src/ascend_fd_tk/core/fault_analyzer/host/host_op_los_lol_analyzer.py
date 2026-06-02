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

from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.diag_result import DiagResult, HostDomain
from ascend_fd_tk.core.model.host import HCCNOpticalInfo
from ascend_fd_tk.utils import helpers


@register_analyzer
class HostOpticalLosLoLAnalyzer(Analyzer):
    @staticmethod
    def _check_optical_indicator(domain: HostDomain, optical_info: HCCNOpticalInfo) -> List[DiagResult]:
        results = []
        flag_map = {
            "Rx Los": optical_info.rx_los_flag,
            "Tx Los": optical_info.tx_los_flag,
            "Rx LoL": optical_info.rx_lo_l_flag,
            "Tx LoL": optical_info.tx_lo_l_flag,
        }
        for k, v in flag_map.items():
            if v and helpers.parse_hex(v) > 0:
                results.append(
                    DiagResult(
                        domain=domain, fault_info=f"光模块{k}指标异常，状态：{v}", suggestion="请检查光模块相关指标"
                    )
                )
        return results

    def analyse(self) -> List[DiagResult]:
        results = []
        for host_info in self.cluster_info.hosts_info.values():
            for npu_chip_info in host_info.npu_chip_info.values():
                optical_info = npu_chip_info.hccn_optical_info
                domain = HostDomain(
                    host_id=host_info.host_id,
                    npu_id=npu_chip_info.npu_id,
                    chip_phy_id=npu_chip_info.chip_phy_id,
                )
                if optical_info:
                    results.extend(self._check_optical_indicator(domain, optical_info))
                    continue
                results.append(
                    DiagResult(domain=domain, fault_info="未查询到光模块信息", suggestion="请检查光模块连接状态")
                )
        return results
