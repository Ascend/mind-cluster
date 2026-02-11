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

from toolkit.core.common import diag_enum
from toolkit.core.config import port_mapping_config
from toolkit.core.context.register import register_analyzer
from toolkit.core.fault_analyzer.base import Analyzer
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.diag_result import DiagResult, Domain


@register_analyzer
class PortSnrSrcAnalyzer(Analyzer):

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self.swis_info = {k: v for k, v in cluster_info.swis_info.items() if v.hccs_info}
        self.port_mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()

    def analyse(self) -> List[DiagResult]:
        diag_results = []
        for swi in self.swis_info.values():
            for port_snr in swi.hccs_info.interface_snr_list:
                domain = [
                    Domain(diag_enum.DeviceType.SWITCH.value, f"{swi.name} {swi.swi_id}"),
                    Domain(diag_enum.DeviceType.SWI_PORT.value, f"{port_snr.interface_name}"),
                ]
                for abnormal_lane_snr in port_snr.abnormal_lane_snr:
                    # TODO 确认SNR类型
                    check_res = self.cluster_info.get_threshold().CDR_HOST_SNR_LINE.check_value_str(
                        abnormal_lane_snr.snr_value)
                    if not check_res:
                        continue
                    diag_results.append(DiagResult(
                        domain,
                        f"{abnormal_lane_snr.lane_name} {check_res}",
                        "请检查端口是否脏污")
                    )
        return diag_results
