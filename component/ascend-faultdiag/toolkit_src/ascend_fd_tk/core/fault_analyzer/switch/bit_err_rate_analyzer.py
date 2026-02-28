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
from ascend_fd_tk.core.common.constants import BIT_ERROR_RATE_LIMIT
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, Domain


@register_analyzer
class BitErrRateAnalyzer(Analyzer):

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)

    def analyse(self) -> List[DiagResult]:
        diag_results = []
        for swi in self.cluster_info.swis_info.values():
            for data in swi.bit_error_rate:
                domain = [
                    Domain(diag_enum.DeviceType.SWITCH.value, f"{swi.name} {swi.swi_id}"),
                    Domain(diag_enum.DeviceType.SWI_PORT.value, f"{data.interface_name}"),
                ]
                diag_results.append(DiagResult(
                    domain,
                    f"BER误码率{data.bit_err_rate}大于阈值{BIT_ERROR_RATE_LIMIT}。",
                    "请检查端口是否脏污")
                )
        return diag_results
