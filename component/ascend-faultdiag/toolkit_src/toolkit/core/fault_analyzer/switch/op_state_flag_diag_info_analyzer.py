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

from toolkit.core.common.diag_enum import DeviceType
from toolkit.core.common.json_obj import JsonObj
from toolkit.core.context.register import register_analyzer
from toolkit.core.fault_analyzer.base import Analyzer
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.diag_result import DiagResult, Domain


class CheckItemInfo(JsonObj):

    def __init__(self, subfix: str, normal_value: str, is_multi_lane: bool, desc: str):
        self.subfix = subfix
        self.normal_value = normal_value
        self.is_multi_lane = is_multi_lane
        self.desc = desc


@register_analyzer
class OpStateFlagDiagInfoAnalyzer(Analyzer):
    """
    检查display optical-module interface xxx 的State Flag信息
    """

    _SEPARATOR = "|"
    _CHECK_ITEMS = [
        CheckItemInfo("Flag", "Normal", True, "收发光指标"),
        CheckItemInfo("Datapath State", "Active", True, "通道状态"),
        CheckItemInfo("Module State", "Ready", False, "功率模式")
    ]

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)

    def analyse(self) -> List[DiagResult]:
        results = []
        for swi_info in self.cluster_info.swis_info.values():
            for op_model in swi_info.optical_models:
                fault_info_list = []
                for diag_info in op_model.state_flag_diag_infos:
                    for check_item in self._CHECK_ITEMS:
                        if diag_info.items.endswith(check_item.subfix):
                            check_states = diag_info.status.split(self._SEPARATOR) if check_item.is_multi_lane else [
                                diag_info.status]
                            cur_fault_info = []
                            for idx, check_state in enumerate(check_states):
                                check_state = str(check_state).strip()
                                if check_state != check_item.normal_value:
                                    lane_info = f"{('lane' + str(idx) + ' ') if check_item.is_multi_lane else ''}"
                                    lane_fault_info = f"{lane_info}{check_item.desc} {diag_info.items}值异常: {check_state}, 预期为: {check_item.normal_value}"
                                    cur_fault_info.append(lane_fault_info)
                            fault_info_list.extend(cur_fault_info)
                            break
                if not fault_info_list:
                    continue
                domain = [Domain(DeviceType.SWITCH, swi_info.swi_id),
                          Domain(DeviceType.SWI_PORT, op_model.interface_name)]
                lane_fault_info = '\n'.join(fault_info_list)
                fault_info = f"端口{op_model.interface_name} flag信息异常：\n{lane_fault_info}"
                result = DiagResult(domain, fault_info, "请检查端口")
                results.append(result)
        return results
