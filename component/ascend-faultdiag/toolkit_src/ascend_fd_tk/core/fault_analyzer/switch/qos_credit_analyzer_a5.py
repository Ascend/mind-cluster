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
from ascend_fd_tk.core.model.diag_result import DiagResult, SwitchDomain
from ascend_fd_tk.core.model.switch import QosCreditPortInfo
from ascend_fd_tk.utils.helpers import to_int


@register_analyzer(generation=[NpuType.A5])
class QosCreditAnalyzerA5(Analyzer):
    """A5 NPU 槽位芯片 QoS Credit 故障分析器。

    QoS Credit 诊断规则：当前可用 Credit（current credit）为 0 且分配的 Credit
    （alloc credit）非 0，表示该维度 Credit 不足（已耗尽），存在链路拥塞风险。
    满足 alloc credit = used credit + current credit；分配与使用均为 0 表示未启用，
    不视为异常。
    """

    def _check_port(self, port_info: QosCreditPortInfo) -> List[str]:
        """检查单端口各维度 Credit 是否不足，返回异常描述列表。"""
        abnormal_infos = []
        for dim_name, (alloc_v, cur_v) in enumerate(zip(port_info.vl_alloc_credits, port_info.vl_current_credits)):
            alloc_int = to_int(alloc_v)
            cur_int = to_int(cur_v)
            # current credit 为 0 且 alloc credit 非 0 → 该维度 Credit 不足
            if cur_int == 0 and alloc_int != 0:
                abnormal_infos.append(f"vl{dim_name} 分配数量={alloc_int} 当前可用数量={cur_int}")
        return abnormal_infos

    def analyse(self) -> List[DiagResult]:
        results = []
        for switch_info in self.cluster_info.swis_info.values():
            for credit_info in switch_info.qos_credit_infos:
                for port_info in credit_info.ports:
                    abnormal_infos = self._check_port(port_info)
                    if not abnormal_infos:
                        continue
                    domain = SwitchDomain(
                        swi_id=switch_info.swi_id,
                        slot_id=switch_info.slot_id,
                    )
                    fault_info = f"chip{credit_info.chip_id}, port{port_info.port_id} QoS Credit不足：\n" + "\n".join(
                        abnormal_infos
                    )
                    suggestion = "端口QoS Credit额度不足，建议排查对应端口业务流量与链路拥塞情况"
                    results.append(DiagResult(domain=domain, fault_info=fault_info, suggestion=suggestion))
        return results
