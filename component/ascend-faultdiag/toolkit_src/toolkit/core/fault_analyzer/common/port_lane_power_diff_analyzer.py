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

from toolkit.core.common import constants
from toolkit.core.common.diag_enum import DeviceType
from toolkit.core.context.register import register_analyzer
from toolkit.core.fault_analyzer.base import Analyzer
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.diag_result import DiagResult, Domain
from toolkit.core.model.optical_module import LanePowerInfo
from toolkit.utils import helpers


@register_analyzer
class PortLanePowerDiffAnalyzer(Analyzer):
    """
    端口lane间power差检查, 不得大于3db
    """
    _LANE_POWER_DIFF_FAULT = "{}端口Lane最大值和最小值差值大于{}db, 实际最大值lane{}: {}{}, 最小值lane{}: {}{}"
    _NA_POWER = "NA"

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)

    def analyse(self) -> List[DiagResult]:
        swi_results = self._analyse_swi_lane_power_diff()
        host_results = self._analyse_host_lane_power_diff()
        bmc_results = self._analyse_bmc_lane_power_diff()
        return swi_results + host_results + bmc_results

    def _analyse_host_lane_power_diff(self) -> List[DiagResult]:
        results = []
        for host_info in self.cluster_info.hosts_info.values():
            for npu_chip_info in host_info.npu_chip_info.values():
                optical_module_info = npu_chip_info.get_optical_module_info()
                if not optical_module_info:
                    continue
                domain = [Domain(DeviceType.SERVER, host_info.host_id),
                          Domain(DeviceType.NPU, npu_chip_info.npu_id),
                          Domain(DeviceType.CHIP, npu_chip_info.chip_id)]
                res = self._generate_diag_result(domain, optical_module_info.lane_power_infos)
                if not res:
                    continue
                results.append(res)
        return results

    def _analyse_swi_lane_power_diff(self) -> List[DiagResult]:
        results = []
        for swi_info in self.cluster_info.swis_info.values():
            for interface_full_info in swi_info.interface_full_infos.values():
                optical_module_info = interface_full_info.get_optical_module_info()
                if not optical_module_info:
                    continue
                domain = [Domain(DeviceType.SWITCH, swi_info.swi_id),
                          Domain(DeviceType.SWI_PORT, interface_full_info.interface)]
                res = self._generate_diag_result(domain, optical_module_info.lane_power_infos)
                if not res:
                    continue
                results.append(res)
        return results

    def _analyse_bmc_lane_power_diff(self) -> List[DiagResult]:
        results = []
        for bmc_info in self.cluster_info.bmcs_info.values():
            for bmc_npu_info in bmc_info.get_bmc_npu_infos():
                optical_module_info = bmc_npu_info.get_optical_module_info()
                if not optical_module_info:
                    continue
                domain = [Domain(DeviceType.BMC, bmc_info.bmc_id),
                          Domain(DeviceType.NPU, bmc_npu_info.npu_id), ]
                if bmc_npu_info.chip_id:
                    domain.append(Domain(DeviceType.CHIP, bmc_npu_info.chip_id))
                res = self._generate_diag_result(domain, optical_module_info.lane_power_infos)
                if not res:
                    continue
                results.append(res)
        return results

    def _generate_diag_result(self, domain: list[Domain], lane_power_infos: List[LanePowerInfo]) -> DiagResult:
        check_results = [self._check_lane_power_diff(lane_power_infos, "tx_power_dbm",
                                                     DeviceType.TX_PORT.value),
                         self._check_lane_power_diff(lane_power_infos, "rx_power_dbm",
                                                     DeviceType.RX_PORT.value)]
        if not any(check_results):
            return None
        fault_info = "\n".join(check_results)
        res = DiagResult(domain, fault_info, "请检查端口")
        return res

    def _check_lane_power_diff(self, lane_power_infos: List[LanePowerInfo],
                               attr: str, port_type: str) -> str:
        origin_attr = attr.replace("_dbm", "")
        lane_power_infos = [lane_power_info for lane_power_info in lane_power_infos if
                            getattr(lane_power_info, origin_attr) != self._NA_POWER]
        if not lane_power_infos:
            return ""
        max_power_info = max(lane_power_infos, key=lambda x: helpers.to_float(getattr(x, attr))[1])
        min_power_info = min(lane_power_infos, key=lambda x: helpers.to_float(getattr(x, attr))[1])
        max_power_str = getattr(max_power_info, attr)
        min_power_str = getattr(min_power_info, attr)
        is_max_power_float, max_power_float = helpers.to_float(max_power_str)
        is_min_power_float, min_power_float = helpers.to_float(min_power_str)
        if not is_max_power_float or not is_min_power_float:
            return ""
        if max_power_float - min_power_float <= constants.POWER_LANE_DIFF_THRESHOLD:
            return ""
        msg = self._LANE_POWER_DIFF_FAULT.format(port_type,
                                                 constants.POWER_LANE_DIFF_THRESHOLD,
                                                 max_power_info.lane_id, getattr(max_power_info, origin_attr),
                                                 max_power_info.power_unit_type.value,
                                                 min_power_info.lane_id, getattr(min_power_info, origin_attr),
                                                 min_power_info.power_unit_type.value)
        return msg
