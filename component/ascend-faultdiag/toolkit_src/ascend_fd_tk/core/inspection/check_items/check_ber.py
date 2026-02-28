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

from ascend_fd_tk.core.common.diag_enum import Customer
from ascend_fd_tk.core.context.register import register_inspection_check_item
from ascend_fd_tk.core.inspection.base import InspectionCheckItem
from ascend_fd_tk.core.inspection.config.config_factory import InspectionConfigFactory
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.inspection import InspectionErrorItem


@register_inspection_check_item
class BerCheckItem(InspectionCheckItem):
    _OPERATIONS_PHASE_ERR = "运维阶段故障, 误码率(BER): {}异常, 高于阶段阈值: {}"
    _OPERATIONS_PHASE_WARN = "运维阶段告警, 误码率(BER): {}异常, 高于阶段阈值: {}"
    _DELIVERY_PHASE_ERR = "交付阶段故障, 误码率(BER): {}异常, 高于阶段阈值: {}"

    def __init__(self, cluster_info: ClusterInfoCache, customer: Customer):
        super().__init__(cluster_info, customer)
        self.inspection_config = InspectionConfigFactory.get_inspection_config(customer)

    def check(self) -> List[InspectionErrorItem]:
        if not self.inspection_config or not self.inspection_config.ber_threshold:
            return []
        ber_threshold = self.inspection_config.ber_threshold
        results = []
        for swi_info in self.cluster_info.swis_info.values():
            interface_full_infos = swi_info.interface_full_infos
            for if_err_rate in swi_info.bit_error_rate:
                if_info = interface_full_infos.get(if_err_rate.interface_name)
                if not if_info:
                    continue
                f_bit_err_rate = float(if_err_rate.bit_err_rate)
                fault_desc = ""
                if f_bit_err_rate >= ber_threshold.operations_phase_err:
                    fault_desc = self._OPERATIONS_PHASE_ERR.format(if_err_rate.bit_err_rate,
                                                                   ber_threshold.operations_phase_err)
                elif f_bit_err_rate >= ber_threshold.operations_phase_warn:
                    fault_desc = self._OPERATIONS_PHASE_WARN.format(if_err_rate.bit_err_rate,
                                                                    ber_threshold.operations_phase_warn)
                elif f_bit_err_rate >= ber_threshold.delivery_phase_err:
                    fault_desc = self._DELIVERY_PHASE_ERR.format(if_err_rate.bit_err_rate,
                                                                 ber_threshold.delivery_phase_err)
                if not fault_desc:
                    continue
                result = InspectionErrorItem(if_info.get_inspection_interface_info(), fault_desc=fault_desc)
                _, peer_interface_info = self.cluster_info.find_peer_swi_interface_info_by_if_info(if_info)
                if peer_interface_info:
                    result.peer_interface = peer_interface_info.get_inspection_interface_info()
                results.append(result)
        return results
