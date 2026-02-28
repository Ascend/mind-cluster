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
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, Domain


@register_analyzer
class HccsSerdesAnalyzer(Analyzer):
    _POWER_ERR_CODE_PREFIX = "0x380"

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)

    def analyse(self) -> List[DiagResult]:
        results = []
        for swi_info in self.cluster_info.swis_info.values():
            if not swi_info.hccs_info:
                continue
            for info in swi_info.hccs_info.serdes_dump_info_list:
                fault_desc_list = []
                if info.cdr_los == "1":
                    fault_desc_list.append("存在CDR失锁")
                if info.csr119_data.startswith(self._POWER_ERR_CODE_PREFIX):
                    fault_desc_list.append(f"存在电源故障, 故障码: {info.csr119_data}")
                if not fault_desc_list:
                    continue
                fault_desc = f"交换芯片: {info.chip_id}, 端口: {info.port_id}" + ",".join(fault_desc_list)
                domains = [Domain(diag_enum.DeviceType.SWITCH.value, swi_info.swi_id),
                           Domain(diag_enum.DeviceType.SWI_PORT.value, info.swi_port_id)]
                res = DiagResult(domains, fault_desc, "请检查端口故障.")
                results.append(res)
        return results
