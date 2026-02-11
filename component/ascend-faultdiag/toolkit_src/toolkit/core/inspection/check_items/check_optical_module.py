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

from toolkit.core.common.diag_enum import Customer, DeviceType
from toolkit.core.context.register import register_inspection_check_item
from toolkit.core.inspection.base import InspectionCheckItem
from toolkit.core.inspection.config.base import OpticalThreshold, FaultDescTemplate
from toolkit.core.inspection.config.config_factory import InspectionConfigFactory
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.inspection import InspectionErrorItem
from toolkit.core.model.optical_module import LanePowerInfo, OpticalModuleInfo
from toolkit.core.model.switch import SwitchInfo
from toolkit.utils import helpers


@register_inspection_check_item
class OpticalModuleCheckItem(InspectionCheckItem):

    def __init__(self, cluster_info: ClusterInfoCache, customer: Customer):
        super().__init__(cluster_info, customer)
        self.swi_info_name_map = {swi_info.name: swi_info for swi_id, swi_info in cluster_info.swis_info.items()}
        self.inspection_config = InspectionConfigFactory.get_inspection_config(customer)

    @staticmethod
    def _check_txrx_power(power_s: str, power_threshold: float, lane_id: str, power_type: str):
        is_float, power_f = helpers.to_float(power_s)
        if is_float and power_f < power_threshold:
            return FaultDescTemplate.LOW_POWER_WARN.value.format(lane_id, power_type, power_s, power_threshold)
        return ""

    @staticmethod
    def _check_lane_power_diff(lane_power_infos: List[LanePowerInfo], threshold_config: OpticalThreshold,
                               attr: str, port_type: str) -> str:
        max_power = max(lane_power_infos, key=lambda x: helpers.to_float(getattr(x, attr))[1])
        min_power = min(lane_power_infos, key=lambda x: helpers.to_float(getattr(x, attr))[1])
        max_power_str = getattr(max_power, attr)
        min_power_str = getattr(min_power, attr)
        is_max_power_float, max_power_float = helpers.to_float(max_power_str)
        is_min_power_float, min_power_float = helpers.to_float(min_power_str)
        if not is_max_power_float or not is_min_power_float:
            return ""
        if max_power_float - min_power_float <= threshold_config.txrx_power_diff_threshold:
            return ""
        msg = FaultDescTemplate.LANE_POWER_DIFF_FAULT.value.format(port_type,
                                                                   threshold_config.txrx_power_diff_threshold,
                                                                   max_power.lane_id, getattr(max_power, attr),
                                                                   min_power.lane_id, getattr(min_power, attr))
        return msg

    def check(self) -> List[InspectionErrorItem]:
        error_items = []
        roce_swis, hccs_swis = [], []
        for swi_info in self.swi_info_name_map.values():
            if swi_info.hccs_info:
                hccs_swis.append(swi_info)
            else:
                roce_swis.append(swi_info)
        error_items.extend(self._check_swi_op(hccs_swis, self.inspection_config.hccs_swi_optical_threshold))
        error_items.extend(self._check_swi_op(roce_swis, self.inspection_config.roce_swi_optical_threshold))
        return error_items

    def _check_swi_op(self, swi_infos: List[SwitchInfo],
                      threshold_config: OpticalThreshold) -> List[InspectionErrorItem]:
        results = []
        for swi_info in swi_infos:
            for interface_full_info in swi_info.interface_full_infos.values():
                optical_module_info = interface_full_info.get_optical_module_info()
                if not optical_module_info or not optical_module_info.lane_power_infos:
                    continue
                check_result_str = self._build_check_result_str(optical_module_info, threshold_config)
                if not check_result_str:
                    continue
                local_inspection_info = interface_full_info.get_inspection_interface_info()
                result = InspectionErrorItem(local_interface=local_inspection_info, fault_desc=check_result_str)
                _, peer_interface_info = self.cluster_info.find_peer_swi_interface_info_by_if_info(interface_full_info)
                if peer_interface_info:
                    result.peer_interface = peer_interface_info.get_inspection_interface_info()
                results.append(result)
        return results

    # 构建检查结果字符串
    def _build_check_result_str(self, optical_module_info: OpticalModuleInfo,
                                threshold_config: OpticalThreshold) -> str:
        check_results = [self._check_lane_power_diff(optical_module_info.lane_power_infos, threshold_config,
                                                     "tx_power", DeviceType.TX_PORT.value),
                         self._check_lane_power_diff(optical_module_info.lane_power_infos, threshold_config,
                                                     "rx_power", DeviceType.RX_PORT.value)]
        for lane_power_info in optical_module_info.lane_power_infos:
            check_results.extend(self._check_lane(lane_power_info, threshold_config))
        check_result_str = "\n".join(check_result for check_result in check_results if check_result)
        return check_result_str

    def _check_lane(self, lane_power_info: LanePowerInfo, threshold_config: OpticalThreshold):
        check_result = [self._check_txrx_power(lane_power_info.rx_power, threshold_config.txrx_power_threshold,
                                               lane_power_info.lane_id, DeviceType.TX_PORT.value),
                        self._check_txrx_power(lane_power_info.tx_power, threshold_config.txrx_power_threshold,
                                               lane_power_info.lane_id, DeviceType.RX_PORT.value)]
        is_float, media_snr_f = helpers.to_float(lane_power_info.media_snr)
        if is_float and media_snr_f < threshold_config.media_snr_error_threshold:
            media_snr_msg = FaultDescTemplate.MEDIA_SNR_WARN.value.format(lane_power_info.lane_id,
                                                                          lane_power_info.media_snr,
                                                                          threshold_config.media_snr_error_threshold)
            check_result.append(media_snr_msg)
        return check_result
