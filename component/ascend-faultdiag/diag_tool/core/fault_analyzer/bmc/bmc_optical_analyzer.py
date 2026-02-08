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

from diag_tool.core.common.diag_enum import DeviceType
from diag_tool.core.context.register import register_analyzer
from diag_tool.core.fault_analyzer.base import Analyzer
from diag_tool.core.model.cluster_info_cache import ClusterInfoCache
from diag_tool.core.model.diag_result import DiagResult, Domain
from diag_tool.core.model.optical_module import LanePowerInfo
from diag_tool.utils import helpers


@register_analyzer
class BmcOpticalAnalyzer(Analyzer):

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self.threshold = self.cluster_info.get_threshold()

    @staticmethod
    def _check_lox(lox: str, transceiver_type: str, lox_type: str) -> str:
        if not lox:
            return ""
        if helpers.parse_hex(lox) > 0:
            return f"{transceiver_type} {lox_type}值{lox}大于0"
        return ""

    def analyse(self) -> List[DiagResult]:
        results = []
        for bmc_info in self.cluster_info.bmcs_info.values():
            for bmc_npu_info in bmc_info.get_bmc_npu_infos():
                optical_module_info = bmc_npu_info.get_optical_module_info()
                if not optical_module_info or not optical_module_info.lane_power_infos:
                    continue
                domain = [Domain(DeviceType.BMC, bmc_info.bmc_id),
                          Domain(DeviceType.NPU, bmc_npu_info.npu_id)]
                if bmc_npu_info.chip_id:
                    domain.append(Domain(DeviceType.CHIP, bmc_npu_info.chip_id))
                # 此处仅记录linkdown数据, 所以有光模块信息即可认为存在故障
                res_list = [f"NPU存在linkdown, 记录时间{optical_module_info.log_time}, 可能为闪断或硬件故障"]
                for lane_power_info in optical_module_info.lane_power_infos:
                    res_list.extend(self._check_lane_power_info(lane_power_info))
                res_list.append(self._check_lox(optical_module_info.tx_los, "Tx", "los"))
                res_list.append(self._check_lox(optical_module_info.rx_los, "Rx", "los"))
                res_list = [res for res in res_list if res]
                fault_info = "\n".join(res_list)
                results.append(DiagResult(domain, fault_info, "请检查端口是否存在脏污"))
        return results

    def _check_lane_power_info(self, lane_power_info: LanePowerInfo) -> List[str]:
        res_list = [self.threshold.TX_POWER_THRESHOLD_CONFIG_MW.check_value_str(lane_power_info.tx_power),
                    self.threshold.RX_POWER_THRESHOLD_CONFIG_MW.check_value_str(lane_power_info.rx_power),
                    self.threshold.TX_BIAS_MA.check_value_str(lane_power_info.bias),
                    self.threshold.HOST_SNR_DB.check_value_str(lane_power_info.host_snr),
                    self.threshold.MEDIA_SNR_DB.check_value_str(lane_power_info.media_snr)]
        res_list = [f"lane{lane_power_info.lane_id}: {res}" for res in res_list if res]
        return res_list
