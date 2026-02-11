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

import collections
import re
from datetime import datetime, timedelta
from typing import List, Tuple, Dict

from toolkit.core.common import diag_enum
from toolkit.core.common.diag_enum import Customer
from toolkit.core.context.register import register_inspection_check_item
from toolkit.core.inspection.base import InspectionCheckItem
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.inspection import InspectionErrorItem
from toolkit.core.model.switch import SwitchInfo, AlarmInfo, InterfaceFullInfo


def _parse_date(date_str: str) -> datetime:
    date_fmt = diag_enum.TimeFormat.BMC_DATE_FMT.value
    if "+" in date_str:
        date_fmt = diag_enum.TimeFormat.TYPE_CLOCK.value
    try:
        return datetime.strptime(date_str, date_fmt)
    except ValueError:
        return None


def find_all_valid_intervals(time_list: List[datetime], target_time: str,
                             min_records: int = 5) -> Tuple[datetime, datetime, int]:
    try:
        target_dt = _parse_date(target_time)
        start_time = target_dt - timedelta(hours=24)

        # 筛选在时间范围内的记录
        filtered = []
        for time_dt in time_list:
            if start_time <= time_dt <= target_dt:
                filtered.append(time_dt)

        # 如果不超过5条，返回None
        if len(filtered) <= min_records:
            return None

        # 超过5条，返回最早、最晚时间和计数
        filtered.sort()
        return (
            filtered[0],
            filtered[-1],
            len(filtered)
        )

    except Exception:
        return None


@register_inspection_check_item
class LinkStatCheckItem(InspectionCheckItem):
    _LINK_DOWN_ERR_CODE = 0x8520003
    _LINK_DOWN_PATTERN = re.compile(r"ifName=([^,]+)")
    _LINK_DOWN_DESC_TEMPLATE = "端口{} down, 事件时间{}"
    _LINK_FLAPPING_DESC_TEMPLATE = "端口{}闪断, 周期{}-{}内闪断{}次"

    def __init__(self, cluster_info: ClusterInfoCache, customer: Customer):
        super().__init__(cluster_info, customer)

    def check(self) -> List[InspectionErrorItem]:
        result = []
        for swi_info in self.cluster_info.swis_info.values():
            result.extend(self._check_link_down(swi_info))
            result.extend(self._check_link_flapping(swi_info))
        return result

    def _check_link_down(self, swi_info: SwitchInfo):
        result = []
        for alarm_info in swi_info.active_alarm_info:
            if alarm_info.alarm_id_int == self._LINK_DOWN_ERR_CODE:
                search = self._LINK_DOWN_PATTERN.search(alarm_info.description)
                if not search:
                    continue
                if_name = search.group(1)
                interface_full_info = swi_info.interface_full_infos.get(if_name)
                if not interface_full_info:
                    continue
                check_result_str = self._LINK_DOWN_DESC_TEMPLATE.format(if_name, alarm_info.date_time)
                inspection_error_item = self._build_err_item(check_result_str, interface_full_info)
                result.append(inspection_error_item)
        return result

    def _check_link_flapping(self, swi_info: SwitchInfo) -> List[InspectionErrorItem]:
        result = []
        link_down_info = [alarm_info for alarm_info in swi_info.history_alarm_info if
                          alarm_info.alarm_id_int == self._LINK_DOWN_ERR_CODE]
        if not link_down_info:
            return result
        port_link_down_map: Dict[str, List[AlarmInfo]] = collections.defaultdict(list)
        for alarm_info in link_down_info:
            search = self._LINK_DOWN_PATTERN.search(alarm_info.description)
            if not search:
                continue
            if_name = search.group(1)
            port_link_down_map[if_name].append(alarm_info)
        for if_name, alarm_infos in port_link_down_map.items():
            interface_full_info = swi_info.interface_full_infos.get(if_name)
            if not interface_full_info:
                continue
            alarm_date_list = [_parse_date(alarm_info.date_time) for alarm_info in alarm_infos]
            flapping_date_range = find_all_valid_intervals(list(filter(bool, alarm_date_list)), swi_info.date_time)
            if not flapping_date_range:
                continue
            check_result_str = self._LINK_FLAPPING_DESC_TEMPLATE.format(
                if_name, flapping_date_range[0], flapping_date_range[1], flapping_date_range[2])
            inspection_error_item = self._build_err_item(check_result_str, interface_full_info)
            result.append(inspection_error_item)
        return result

    def _build_err_item(self, check_result_str: str, interface_full_info: InterfaceFullInfo) -> InspectionErrorItem:
        local_inspection_info = interface_full_info.get_inspection_interface_info()
        inspection_error_item = InspectionErrorItem(local_interface=local_inspection_info,
                                                    fault_desc=check_result_str)
        _, peer_interface_info = self.cluster_info.find_peer_swi_interface_info_by_if_info(interface_full_info)
        if peer_interface_info:
            inspection_error_item.peer_interface = peer_interface_info.get_inspection_interface_info()
        return inspection_error_item
