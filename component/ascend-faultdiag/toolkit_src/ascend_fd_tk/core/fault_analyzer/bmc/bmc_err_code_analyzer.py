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

from ascend_fd_tk.core.common.diag_enum import DeviceType
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.fault_analyzer.bmc.bmc_err_code_config import _ERR_CODE_EVENT_LIST
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, Domain
from ascend_fd_tk.utils import helpers


@register_analyzer
class BmcErrCodeAnalyzer(Analyzer):

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self._err_code_event_map = dict()
        for event in _ERR_CODE_EVENT_LIST:
            self._err_code_event_map.setdefault(event.err_code, []).append(event)

    @staticmethod
    def _get_domain_list(bmc_id, err_code_event, bmc_sel_event_description):
        domain = [Domain(DeviceType.BMC.value, bmc_id)]
        if not err_code_event.err_info_pattern:
            return domain
        search_data = err_code_event.err_info_pattern.search(bmc_sel_event_description)
        if not search_data:
            return domain
        hardware_info = search_data.groupdict()
        npu = hardware_info.get("npu") or hardware_info.get("npu1") or hardware_info.get("npu2")
        chip = hardware_info.get("chip")
        if npu:
            domain.append(Domain(DeviceType.NPU.value, npu))
        if chip:
            domain.append(Domain(DeviceType.CHIP.value, chip))
        return domain

    def analyse(self) -> List[DiagResult]:
        results = []
        for bmc_info in self.cluster_info.bmcs_info.values():
            for bmc_sel in bmc_info.bmc_sel_list:
                err_code_event = self._get_event_by_keywords(bmc_sel)
                if not err_code_event:
                    continue
                fault_info = self._get_fault_info(bmc_info.sn_num, err_code_event, bmc_sel)
                domain_list = self._get_domain_list(bmc_info.bmc_id, err_code_event, bmc_sel.event_description)
                results.append(
                    DiagResult(domain_list, fault_info, err_code_event.handle_suggestion, bmc_sel.event_code)
                )
        return results

    def _get_event_by_keywords(self, bmc_sel):
        event_code_int = helpers.parse_hex(bmc_sel.event_code)
        err_code_event_list = self._err_code_event_map.get(event_code_int, [])
        for err_code_event in err_code_event_list:
            if err_code_event.keywords in bmc_sel.event_description:
                return err_code_event
        return None

    def _get_fault_info(self, sn_num, err_code_event, bmc_sel):
        fault_info_list = []
        host_info = self.cluster_info.find_host_info_by_sn_num(sn_num)
        if host_info:
            fault_info_list.append(f"{DeviceType.SERVER.value}{host_info.host_id}:")
        fault_info_list.append(f"{err_code_event.err_desc},\n原始故障描述: {bmc_sel.event_description},"
                               f"\n故障发生时间: {bmc_sel.generation_time}")
        return "".join(fault_info_list)
