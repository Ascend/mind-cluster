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

import re
from typing import List

from toolkit.core.common.diag_enum import DeviceType
from toolkit.core.context.register import register_analyzer
from toolkit.core.fault_analyzer.base import Analyzer
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.diag_result import DiagResult, Domain


@register_analyzer
class LaneReductionAnalyzer(Analyzer):
    _ERR_CODE = 0xf10509

    _IF_PATTERN = re.compile(r"EntPhysicalName=([^,]+)")
    _LANE_REDUCTION_DESC_TEMPLATE = "端口{}发生降lane"

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)

    def analyse(self) -> List[DiagResult]:
        result = []
        for swi_info in self.cluster_info.swis_info.values():
            interface_full_infos = swi_info.interface_full_infos
            for alarm_info in swi_info.active_alarm_info:
                if alarm_info.alarm_id_int != self._ERR_CODE:
                    continue
                search = self._IF_PATTERN.search(alarm_info.description)
                if not search:
                    continue
                if_name = search.group(1)
                if_info = interface_full_infos.get(if_name)
                if not if_info:
                    continue
                fault_desc = self._LANE_REDUCTION_DESC_TEMPLATE.format(if_name)
                _, peer_info = self.cluster_info.find_peer_swi_interface_info_by_if_info(if_info)
                if peer_info:
                    fault_desc += f", 对端端口信息: {peer_info.get_inspection_interface_info()}"
                res = DiagResult([Domain(DeviceType.SWITCH, swi_info.swi_id), Domain(DeviceType.SWI_PORT, if_name)],
                                 fault_info=fault_desc, suggestion=f"请检查端口", err_code=alarm_info.alarm_id)
                result.append(res)
        return result
