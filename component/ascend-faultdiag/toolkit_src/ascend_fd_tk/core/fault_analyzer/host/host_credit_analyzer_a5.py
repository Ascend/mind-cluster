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

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.diag_result import DiagResult, HostDomain
from ascend_fd_tk.core.model.host_a5 import NpuChipInfoA5, CreditInfo
from ascend_fd_tk.utils.helpers import to_int


@register_analyzer(generation=[NpuType.A5])
class HostCreditAnalyzerA5(Analyzer):
    """A5 端口 Credit 故障分析器。

    虚拟链路（VL）优先级 Credit 诊断规则：某优先级的 Credit 分配数量
    （link_alloc_vl_pri_credit）与使用数量（link_cur_used_pri_credit）相等且非 0，
    表示该优先级 Credit 已耗尽（无剩余额度），存在链路拥塞风险。
    分配与使用均为 0 表示未启用该优先级，不视为异常。
    """

    @staticmethod
    def _check_credit_info(host_id: str, npu_chip_info: NpuChipInfoA5, credit_info: CreditInfo) -> List[DiagResult]:
        """检查单个端口的虚拟链路（VL）优先级 Credit 是否耗尽。"""
        if not credit_info:
            return []
        alloc_list = credit_info.link_alloc_vl_pri_credits or []
        used_list = credit_info.link_cur_used_pri_credits or []
        abnormal_infos: List[str] = []
        for pri_idx in range(min(len(alloc_list), len(used_list))):
            alloc_v = to_int(alloc_list[pri_idx])
            used_v = to_int(used_list[pri_idx])
            # Credit 分配数量与使用数量相等且非 0 → 该优先级 Credit 已耗尽
            if alloc_v == used_v != 0:
                abnormal_infos.append(f"vl{pri_idx} 分配数量={alloc_v} 使用数量={used_v}")
        if not abnormal_infos:
            return []
        domain = HostDomain(
            host_id=host_id,
            npu_id=npu_chip_info.npu_id,
            udie_id=credit_info.udie_id,
            npu_port_id=credit_info.port_id,
        )
        fault_info = "端口虚拟链路（VL）优先级Credit不足：\n" + "\n".join(abnormal_infos)
        suggestion = "端口优先级Credit额度已耗尽，建议排查该端口业务流量与链路拥塞情况"
        return [DiagResult(domain=domain, fault_info=fault_info, suggestion=suggestion)]

    def analyse(self) -> List[DiagResult]:
        results = []
        for host_info in self.cluster_info.hosts_info.values():
            for npu_chip_info in host_info.npu_chip_info.values():
                for credit_info in npu_chip_info.credit_info_list:
                    results.extend(self._check_credit_info(host_info.host_id, npu_chip_info, credit_info))
        return results
