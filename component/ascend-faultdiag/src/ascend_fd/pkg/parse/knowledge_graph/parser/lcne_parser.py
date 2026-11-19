#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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
from itertools import chain
import logging
import os
import re

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import EventStorage, FileParser
from ascend_fd.pkg.parse.knowledge_graph.tools.utils import GroupSelectMixin
from ascend_fd.utils.constant.str_const import UNKNOWN_DEVICE_ID
from ascend_fd.utils.constant.ub_const import (
    AGE_PERIOD_KEYWORD,
    PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA,
    PRECHECK_UBMEM_TIMEOUT,
    SIDE_AFTER,
    SIDE_BEFORE,
    SNAP_DIE_ID,
    SNAP_DROP_METRICS,
    SNAP_PORT_HEADER,
    SNAPSHOOT_TAG,
    SNAP_START_END,
    UBMEM_TIMEOUT_KEYWORD,
)
from ascend_fd.utils.regular_table import COMPOSITE_SWITCH_CHIP_SOURCE
from ascend_fd.utils.tool import MultiProcessJob, check_and_format_time_str

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


class LCNEParser(GroupSelectMixin, FileParser):
    _type = "lcne"
    TARGET_FILE_PATTERNS = "lcne_log_path"
    SOURCE_FILE = "LCNELog"
    SNAP_SHOT_LOG = "nsdk_hccl_shapshot.log"

    # 行示例: "... set ubmem timeout xxxxxxxx age_period = 40000, ..."，单位微秒
    AGE_PERIOD_PATTERN = re.compile(rf"{AGE_PERIOD_KEYWORD} = (?P<{AGE_PERIOD_KEYWORD}>\d+)")
    TIME_FORMATS = [
        # 格式1: 'May 27 2025 11:25:00+08:00'
        (
            re.compile(r'\b[A-Z][a-z]{2}\s{0,3}\d{1,2}\s{0,3}\d{4}\s{0,3}\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}\b'),
            "%b %d %Y %H:%M:%S%z",
        ),
        # 格式2: 'May 27 2025 11:25:00'
        (re.compile(r'\b[A-Z][a-z]{2}\s{0,3}\d{1,2}\s{0,3}\d{4}\s{0,3}\d{2}:\d{2}:\d{2}\b'), '%b %d %Y %H:%M:%S'),
        # 格式3: '2025-05-26 17:25:42.348'
        (re.compile(r'\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3}'), '%Y-%m-%d %H:%M:%S.%f'),
    ]
    TIME_FORMAT = "%Y-%m-%dT%H:%M:%S%z"
    MAX_TIME = "9999-12-31 23:59:59.999999"

    TIME_PATTERN = re.compile(r"(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(?:\.\d+)?)")
    # 组头行同时捕获 devid：同一文件可能打印多台设备的快照，devid 作为分组/去重/事件归属的键
    HEADER_PATTERN = re.compile(
        rf"{re.escape(SNAPSHOOT_TAG)}\s*devid\s*=\s*(\d+)\s*,\s*{SNAP_DIE_ID}\s*=\s*(\d+)\s*,\s*"
        rf"{SNAP_START_END}\s*=\s*(\d+)"
    )

    def __init__(self, params):
        super().__init__(params)
        self.default_conf.update(self.regex_conf.get(COMPOSITE_SWITCH_CHIP_SOURCE, {}))
        self.user_conf.update(self.regex_conf.get(COMPOSITE_SWITCH_CHIP_SOURCE, {}))
        self.timezone_trans_flag = self.get_timezone_trans_flag()
        self.start_time = self.params.get("start_time")
        self.end_time = self.params.get("end_time")

    def parse(self, parse_ctx: KGParseCtx, task_id: str):
        """
        Parse lcne log file
        :param parse_ctx: knowledge graph parser context
        :param task_id: unique task id
        :return: parse descriptor result
        """
        file_list = self.find_log(parse_ctx.parse_file_path)
        if not file_list:
            return [], {}
        kg_logger.info("%s files parse job started.", self.SOURCE_FILE)
        self.is_sdk_input = parse_ctx.is_sdk_input
        # 断点续训时间只在 parse_ctx 中
        self.resuming_training_time = getattr(parse_ctx, "resuming_training_time", "") or ""
        if self.is_sdk_input:
            results = dict()
            for idx, file_source in enumerate(file_list):
                results.update(
                    {f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}": self._parse_file(file_source)}
                )
        else:
            multiprocess_job = MultiProcessJob("KNOWLEDGE_GRAPH", pool_size=len(file_list), task_id=task_id)
            for idx, file_source in enumerate(file_list):
                multiprocess_job.add_security_job(
                    f"{self.SOURCE_FILE}_ID-{idx}_{self._get_filename(file_source)}", self._parse_file, file_source
                )
            results, _ = multiprocess_job.join_and_get_results()
        kg_logger.info("%s files parse job is complete.", self.SOURCE_FILE)
        return list(chain(*results.values())), {}

    @classmethod
    def _is_snapshoot_file(cls, file_source) -> bool:
        """按文件名区分快照日志。"""
        return os.path.basename(cls._get_filename(file_source)) == cls.SNAP_SHOT_LOG

    def _filter_lcne_time(self, context: str):
        """
        Filter lcne time
        :param context: singe line in raw log
        :return: filtered time from context
        """
        for regex, time_format in self.TIME_FORMATS:
            ret = regex.findall(context)
            if not ret:
                continue
            try:
                time_obj = datetime.strptime(ret[0], time_format).strftime(self.TIME_FORMAT)
                return check_and_format_time_str(str(time_obj), self.timezone_trans_flag)
            except ValueError:
                pass
        return ""

    def _parse_file(self, file_source):
        """
        Parse single lcne log line by line
        :param file_source: log file source
        :return: event list
        """
        is_snapshot = self._is_snapshoot_file(file_source)
        event_storage = EventStorage()
        snapshot_groups = {SIDE_BEFORE: [], SIDE_AFTER: []}
        current_group = {}
        for log_line in self._yield_log(file_source):
            if is_snapshot:
                current_group = self._parse_snapshoot_line(log_line, current_group, snapshot_groups, file_source)
                continue
            event_dict = self.parse_single_line(log_line)
            if not event_dict:
                # 配置库未命中时，尝试匹配 PRECHECK 行
                event_dict = self._match_precheck_line(log_line)
            if not event_dict:
                continue
            occur_time = self._filter_lcne_time(log_line) or self.MAX_TIME
            self.supplement_common_info(event_dict, file_source, occur_time)
            event_storage.record_event(event_dict)
        if is_snapshot:
            # 快照数据合成 PRECHECK 事件后，与普通 lcne 事件一样走 EventStorage
            self._finish_group(current_group, snapshot_groups)
            for event_dict in self._assemble_snapshoot_events(snapshot_groups, file_source):
                event_storage.record_event(event_dict, with_occurrence=False)
        return event_storage.generate_event_list()

    def _match_precheck_line(self, log_line: str):
        """
        Match the ubmem timeout line in diaglog and generate a PRECHECK event.
        It does not depend on the fault repository in kg-config.json.
        Line sample: "... set ubmem timeout xxxxxxxx age_period = 40000, ..."
        :param log_line: single line in raw log
        :return: PRECHECK event dict or {}
        """
        if UBMEM_TIMEOUT_KEYWORD not in log_line or AGE_PERIOD_KEYWORD not in log_line:
            return {}
        match = self.AGE_PERIOD_PATTERN.search(log_line)
        if not match:
            return {}
        return {
            "event_code": PRECHECK_UBMEM_TIMEOUT,
            "key_info": log_line,
            "attribute": {AGE_PERIOD_KEYWORD: match.group(AGE_PERIOD_KEYWORD)},
            "is_custom_event": False,
        }

    def _assemble_snapshoot_events(self, snapshot_groups: dict, file_source):
        """
        Aggregate snapshot groups by device and generate one PRECHECK event per device (devid).

        仅保留同一 die 的建链（start_end=0）/建链销毁（start_end=1）两组数据，
        并把指标值规整为 {"udieX": {port: {指标: "0x.."}}}，
        使规则层可以复用 ubctl 的 before/after 对比逻辑。
        :param snapshot_groups: {side: [group]}，按 before/after 归类的数据块
        :param file_source: current snapshot log file
        :return: event dict list
        """
        groups_by_dev = {}
        for side_groups in snapshot_groups.values():
            for group in side_groups:
                groups_by_dev.setdefault(group.get("devid", UNKNOWN_DEVICE_ID), []).append(group)
        # 快照组缺失 source_file 时，兜底使用当前快照文件
        fallback_source = self._get_file_path(file_source)
        events = []
        for devid, dev_groups in groups_by_dev.items():
            before_groups, after_groups = self._select_groups(dev_groups)
            if not before_groups or not after_groups:
                continue
            before = {udie: group["ports_metrics"] for udie, group in before_groups.items()}
            after = {udie: group["ports_metrics"] for udie, group in after_groups.items()}
            # occur_time 取该设备选中 after 组的组头时间
            occur_time = max((group.get("time", "") for group in after_groups.values()), default="")
            # 与旧 UB 需求保持一致：before/after 的源文件用换行拼接
            before_source = next(
                (group.get("source_file", "") for group in before_groups.values() if group.get("source_file")), ""
            )
            after_source = next(
                (group.get("source_file", "") for group in after_groups.values() if group.get("source_file")), ""
            )
            source_file = "\n".join(filter(None, [before_source, after_source]))
            event_dict = {
                "event_code": PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA,
                "key_info": f"{self.SNAP_SHOT_LOG} before/after raw data for dev {devid}",
                "attribute": {
                    "before": before,
                    "after": after,
                    "source_lines": {"before": [], "after": []},
                },
                "is_custom_event": False,
                "occurrence": [],
                "source_device": devid,
                "source_file": source_file or fallback_source,
                "occur_time": occur_time,
                "type": self.SOURCE_FILE,
            }
            events.append(event_dict)
        return events

    def _select_groups(self, groups: list):
        """
        按任务时间窗逐 die 选择 before/after 数据块（选组逻辑统一由 GroupSelectMixin 实现）。
        :param groups: 全部数据块（可跨文件）
        :return: {udie: before 组}、{udie: after 组}
        """
        return self._select_before_after_groups(
            groups,
            getattr(self, "start_time", ""),
            getattr(self, "end_time", ""),
            getattr(self, "resuming_training_time", ""),
        )

    def _parse_snapshoot_line(self, log_line: str, current: dict, groups: dict, file_source) -> dict:
        """
        Parse a single snapshoot log line and return the current group state.
        :param log_line: single log line
        :param current: current group being accumulated
        :param groups: {side: [group]}，完成的数据块按 before/after 归类
        :param file_source: current log file
        :return: updated current group
        """
        if not log_line:
            return current
        header = self.HEADER_PATTERN.search(log_line)
        if header:
            self._finish_group(current, groups)
            return {
                "devid": header.group(1),
                "side": SIDE_BEFORE if header.group(3) == "0" else SIDE_AFTER,
                "udie": f"udie{header.group(2)}",
                "ports": [],
                "ports_metrics": {},
                "table": "",
                "time": self._header_time(log_line),
                "source_file": self._get_file_path(file_source),
            }
        if not current or SNAPSHOOT_TAG not in log_line:
            return current
        content = log_line.split(SNAPSHOOT_TAG, 1)[1].strip()
        if content:
            self._parse_tag_content(log_line, content, current)
        return current

    def _header_time(self, log_line: str) -> str:
        """
        取组头行行首绝对时间戳并规范化为 "%Y-%m-%d %H:%M:%S.%f"（无微秒补 .000000）
        :param log_line: 组头行
        :return: 规范化的组时间；无时间戳或格式非法返回 ""
        """
        time_match = self.TIME_PATTERN.search(log_line)
        group_dt = self._parse_group_time(time_match.group(1)) if time_match else None
        return group_dt.strftime("%Y-%m-%d %H:%M:%S.%f") if group_dt else ""

    def _parse_tag_content(self, log_line: str, content: str, current: dict):
        parts = content.split()
        if not parts:
            return
        key, values = parts[0], parts[1:]
        if key == SNAP_PORT_HEADER:
            current["table"] = SNAP_PORT_HEADER
            current["ports"] = self._to_ports(values)
            return
        if key not in SNAP_DROP_METRICS:
            # 非目标表头（如 nl）只用于结束 port 表，不记录行内容
            if key == "nl":
                current["table"] = ""
            return
        if current.get("table") != SNAP_PORT_HEADER:
            return
        port_ids = current.get("ports", [])
        if len(values) > len(port_ids):
            kg_logger.warning(
                "snapshoot metric %s has %d values but only %d ports, skip the line.",
                key,
                len(values),
                len(port_ids),
            )
            return
        for idx, value in enumerate(values):
            current["ports_metrics"].setdefault(port_ids[idx], {})[key] = value

    @staticmethod
    def _to_ports(values: list) -> list:
        ports = []
        for value in values:
            if not value.isdigit():
                break
            ports.append(int(value, 10))
        return ports

    @staticmethod
    def _finish_group(group, groups: dict):
        if not group:
            return
        group.setdefault("ports_metrics", {})
        if group["ports_metrics"]:
            groups[group["side"]].append(group)
