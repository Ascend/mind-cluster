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

from ascend_fd_tk.core.common.diag_enum import DeviceType
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, Domain
from ascend_fd_tk.core.model.switch import SwitchInfo


@register_analyzer
class OpticalInvalidAnalyzer(Analyzer):
    _LOS_ALARM_ERR_CODE = 0x8130059
    _LOS_ALARM_ERR_PATTERN = re.compile(r"EntPhysicalName=([^,]+).*Reason=([^)]+)")

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)

    def analyse(self) -> List[DiagResult]:
        results = []
        for switch_info in self.cluster_info.swis_info.values():
            results.extend(self._los_alarm_check(switch_info))

        return results

    def _los_alarm_check(self, switch_info: SwitchInfo):
        results = []
        for alarm_info in switch_info.active_alarm_info:
            if alarm_info.alarm_id_int != self._LOS_ALARM_ERR_CODE:
                continue
            search = self._LOS_ALARM_ERR_PATTERN.search(alarm_info.description)
            if not search:
                continue
            ifname = search.group(1)
            reason = search.group(2)
            domain = [
                Domain(DeviceType.SWITCH, switch_info.swi_id),
                Domain(DeviceType.SWI_PORT, ifname),
            ]
            res = DiagResult(domain, f"光模块链路Los告警, 原因: {reason}", "请检查光模块",
                             err_code=alarm_info.alarm_id)
            results.append(res)
        return results
