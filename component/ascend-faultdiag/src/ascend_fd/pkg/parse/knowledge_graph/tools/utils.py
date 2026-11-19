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
from datetime import datetime

from ascend_fd.utils.constant.str_const import UNKNOWN_DEVICE_ID
from ascend_fd.utils.constant.ub_const import PRECHECK_PREFIX, SIDE_AFTER, SIDE_BEFORE
from ascend_fd.utils.tool import check_and_format_time_str


def get_device_precheck_event(precheck_info, source_device: str):
    precheck = precheck_info.get(source_device, [])
    events = {}
    for event in precheck:
        if event.get("event_code", "").startswith(PRECHECK_PREFIX) or source_device == UNKNOWN_DEVICE_ID:
            events[event.get("event_code")] = event

    # 多个event事件以各自的event_code为key，全部放入root_causes
    # 如果不是unknown设备，只取PRECHECK_开头的故障事件
    return events


class GroupSelectMixin:
    """
    kernel.log 与 nsdk_hccl_shapshot.log 共用的数据块选组逻辑。

    两个数据源都按 (side, udie, 组时间) 组织数据块，差异只在组结构和指标字段，
    因此这里统一实现时间解析、去重、最后一份兜底和任务时间窗选组。
    """

    UDIES = ("udie0", "udie1")

    @staticmethod
    def _parse_iso_time(time_str):
        """
        解析任务时间窗字符串（yyyy-mm-dd hh:mm:ss[.ffffff]），仅取秒级精度。
        :return: datetime；为空或格式非法返回 None
        """
        formatted = check_and_format_time_str(time_str.strip()) if time_str else ""
        if not formatted:
            return None
        return datetime.strptime(formatted[:19], "%Y-%m-%d %H:%M:%S")

    @staticmethod
    def _parse_group_time(time_str):
        """
        组时间字符串转 datetime，复用统一的日志时间格式化逻辑。
        :return: datetime；为空或格式非法返回 None
        """
        formatted = check_and_format_time_str(time_str.strip()) if time_str else ""
        if not formatted:
            return None
        return datetime.strptime(formatted, "%Y-%m-%d %H:%M:%S.%f")

    @staticmethod
    def _dedup_groups(groups: list) -> list:
        """
        跨文件去重：devid+side+udie+time 相同的组只保留第一份，避免多份 dump 重复组干扰选组。
        组时间缺失时无法判定是否重复，原样保留。
        """
        seen = set()
        unique_groups = []
        for group in groups:
            group_time = group.get("time", "")
            if group_time:
                group_key = (group.get("devid"), group["side"], group["udie"], group_time)
                if group_key in seen:
                    continue
                seen.add(group_key)
            unique_groups.append(group)
        return unique_groups

    @classmethod
    def _latest_group(cls, candidates: list):
        """
        取组时间最晚的一组；组时间均不可用时按列表顺序取最后一个。
        """
        if not candidates:
            return None
        timed_groups = []
        for group in candidates:
            group_dt = cls._parse_group_time(group.get("time", ""))
            if group_dt is not None:
                timed_groups.append((group_dt, group))
        return max(timed_groups, key=lambda item: item[0])[1] if timed_groups else candidates[-1]

    def _select_before_after_groups(self, groups: list, start_time: str, end_time: str, resuming_time: str):
        """
        按时间窗逐 die 选择 before/after 数据块：
          before：窗口起点（有 resuming 取 resuming，否则 start_time）之后的第一组建链组
          after：窗口 [起点, end_time] 内最晚的销毁组
        无完整时间窗时退化为每个 die 取最后一份 before 组和最后一份 after 组。
        :return: ({udie: before 组}, {udie: after 组})
        """
        grouped = {}
        for group in self._dedup_groups(groups):
            grouped.setdefault(group["udie"], []).append(group)

        start_dt = self._parse_iso_time(start_time)
        end_dt = self._parse_iso_time(end_time)
        resuming_dt = self._parse_iso_time(resuming_time)

        before_groups, after_groups = {}, {}
        for udie, die_groups in grouped.items():
            if start_dt and end_dt:
                before_dt, before_group = self._pick_window_group(
                    die_groups, SIDE_BEFORE, start_dt, end_dt, resuming_dt
                )
                after_dt, after_group = self._pick_window_group(die_groups, SIDE_AFTER, start_dt, end_dt, resuming_dt)
                # 两侧都有数据且前后顺序正常才收录，否则丢弃该 die
                if before_group is None or after_group is None or before_dt > after_dt:
                    continue
            else:
                # 无完整时间窗（SDK/无 plog）：每个 die 各取最后一份 before/after 组
                before_group = self._latest_group([group for group in die_groups if group["side"] == SIDE_BEFORE])
                after_group = self._latest_group([group for group in die_groups if group["side"] == SIDE_AFTER])
                if before_group is None or after_group is None:
                    continue
                # 组时间可解析时仍需保证前后顺序正常，避免选中新任务的建链组与上一任务的销毁组
                before_dt = self._parse_group_time(before_group.get("time", ""))
                after_dt = self._parse_group_time(after_group.get("time", ""))
                if before_dt is not None and after_dt is not None and before_dt > after_dt:
                    continue
            before_groups[udie] = before_group
            after_groups[udie] = after_group
        return before_groups, after_groups

    def _pick_window_group(self, die_groups: list, side: str, start_dt, end_dt, resuming_dt):
        """
        时间窗内选单个 die 中 side 侧的目标组，返回 (组时间, 组)，无候选返回 (None, None)：
          before：>= 窗口起点 且 <= end_time 中最早，即窗口起点后的首次建链组
          after：同区间中最晚，即任务结束时的销毁组
        """
        # 断点续训时以续训时刻为窗口起点：before 必须取续训后的建链基线，不能取断点前的组
        lower_dt = resuming_dt or start_dt
        candidates = []
        for group in die_groups:
            if group["side"] != side:
                continue
            group_dt = self._parse_group_time(group.get("time", ""))
            if group_dt is None or group_dt < lower_dt or group_dt > end_dt:
                continue
            candidates.append((group_dt, group))
        if not candidates:
            return None, None
        # before 取首次建链作为计数基线，after 取最晚销毁组
        if side == SIDE_BEFORE:
            return min(candidates, key=lambda item: item[0])
        return max(candidates, key=lambda item: item[0])
