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

from typing import List, Type, Dict

import re

from ascend_fd_tk.core.collect.parser.host_parser import HostParser
from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.core.model.host_a5 import (
    HCCNOpticalInfoA5,
    NICPortLaneInfoA5,
    OpticalModuleHardwareAttr,
    OpticalModuleMonitorItem,
    OpticalModuleSerialInfo,
    OpticalStateFlag,
    OpticalTopHeadline,
)
from ascend_fd_tk.utils.helpers import split_multiline
from ascend_fd_tk.utils.form_parser import FormParser
from ascend_fd_tk.utils.table_parser import TableParser


class HostParserA5(HostParser):
    """A5 回显清洗器，继承 A3 实现，按需重写存在格式差异的方法。"""

    # 网卡 SFP lane 解析正则（按 lane 号动态收集，不硬编码 lane 数量）
    _NIC_BIAS_RE = re.compile(r"Internally measured Tx(\d{1,2}) bias current monitor:\s{0,2}([\d.]{1,10})\s{0,10}mA")
    _NIC_TX_POWER_RE = re.compile(
        r"Internally measured Tx(\d{1,2}) output optical power:\s{0,2}[\d.]{1,10}\s{0,10}mW\s{0,2}/\s{0,}(\S{1,10})\s{0,10}dBm"
    )
    _NIC_RX_POWER_RE = re.compile(
        r"Internally measured Rx(\d{1,2}) input optical power:\s{0,2}[\d.]{1,10}\s{0,10}mW\s{0,2}/\s{0,}(\S{1,10})\s{0,10}dBm"
    )
    _NIC_TX_LOS_RE = re.compile(r"Latched Tx LOS flag\(([^)]{1,2})\),\s{1,10}lane\s{1,2}(\d{1,2})$")
    _NIC_RX_LOS_RE = re.compile(r"Latched Rx LOS flag\(([^)]{1,2})\),\s{1,10}media lane\s{1,2}(\d{1,2})$")
    _NIC_TX_CDR_LOL_RE = re.compile(r"Latched Tx CDR LOL flag\(([^)]{1,2})\),\s{1,10}lane\s{1,2}(\d{1,2})$")
    _NIC_RX_CDR_LOL_RE = re.compile(r"Latched Rx CDR LOL flag\(([^)]{1,2})\),\s{1,10}media lane\s{1,2}(\d{1,2})$")
    _NIC_HOST_SNR_RE = re.compile(r"Host side SNR lane\s{1,2}(\d{1,2}):\s{0,2}([\d.]{1,10})\s{0,10}dB")
    _NIC_MEDIA_SNR_RE = re.compile(r"Media side SNR lane\s{1,2}(\d{1,2}):\s{0,2}([\d.]{1,10})\s{0,10}dB")
    _NIC_CARD_NAME_RE = re.compile(r"\|----(\w{1,10})\(")

    @classmethod
    def parse_optical_top_headline(cls, cmd_res: str) -> List[OpticalTopHeadline]:
        titles_dict = {
            "npu_id": "NPU",
            "optical_silk_screen_num": "Optical silk screen num",
            "optical_id": "Optical id",
            "speed": "Speed",
            "used_for": "Used for",
        }
        end_sign = "+-----"
        parse_data_list = TableParser.parse(
            cmd_res, titles_dict, separate_title_content_lines_num=1, end_sign=end_sign, col_separator="|"
        )
        results = [OpticalTopHeadline.from_dict(parse_data) for parse_data in parse_data_list]
        return results

    @classmethod
    def parse_optical_info_a5(cls, cmd_res: str, npu_id, optical_id) -> HCCNOpticalInfoA5:
        """解析 A5 光模块多表信息。"""
        if not cmd_res or not cmd_res.strip():
            return None
        cmd_res_list = split_multiline(cmd_res, ["+-", "+-"])
        cls_list = [
            OpticalModuleHardwareAttr,
            OpticalModuleSerialInfo,
            OpticalStateFlag,
            OpticalModuleMonitorItem,
        ]
        parsed: dict = {}
        for segment, model_cls in zip(cmd_res_list, cls_list):
            if model_cls is OpticalStateFlag:
                parsed[model_cls] = cls._parse_state_flag_table(segment, model_cls.parse_title_dict())
            else:
                parsed[model_cls] = cls._parse_table(segment, model_cls.parse_title_dict(), model_cls)

        return HCCNOpticalInfoA5(
            npu_id=npu_id,
            optical_id=optical_id,
            hardware_attr=parsed.get(OpticalModuleHardwareAttr, []),
            serial_info=parsed.get(OpticalModuleSerialInfo, []),
            state_flag=parsed.get(OpticalStateFlag, []),
            monitor_item=parsed.get(OpticalModuleMonitorItem, []),
        )

    @classmethod
    def _parse_table(cls, segment: str, titles_dict: Dict[str, str], model_cls: Type[JsonObj]) -> List[JsonObj]:
        """通用表解析：TableParser 解析 + 模型类构造。"""
        if not segment.strip():
            return []
        parse_list = TableParser.parse(
            segment, titles_dict, separate_title_content_lines_num=1, end_sign="+--", col_separator="|"
        )
        return [model_cls.from_dict(item) for item in parse_list]

    @classmethod
    def _parse_state_flag_table(cls, segment: str, titles_dict: dict) -> List[OpticalStateFlag]:
        """解析 state_flag 表：lane 列数不固定，用 "Lane" 关键字匹配首个 Lane0 列，
        让 TableParser 把所有 Lane 列作为一个整体字符串取到 `lanes` 字段，再 split("|") 拆分。

        回显标题行示例（lane 数可能为 2/4/8 等）：
            | NPU | Optical | Items        | Lane0 | Lane1 | Lane2 | Lane3 |
        """
        if not segment.strip():
            return []
        parse_list = TableParser.parse(
            segment, titles_dict, separate_title_content_lines_num=1, end_sign="+--", col_separator="|"
        )
        results = []
        for item in parse_list:
            lanes_raw = item.get("lanes", "")
            if not lanes_raw:
                continue
            # split("|") 后每段含空格（TableParser 按列起始位置切片），需逐项 strip
            item["lanes"] = [lane.strip() for lane in lanes_raw.split("|")]
            results.append(OpticalStateFlag.from_dict(item))
        return results

    @classmethod
    def parse_nic_card_names(cls, cmd_res: str) -> List[str]:
        """解析 `hinicadm5 info` 回显，提取所有网卡名（如 hinic0、hinic1）。

        回显示例：
            |----hinic0(CAL_2X400G_UB_EXP)
            |----hinic1(CAL_2X400G_UB_EXP)
        """
        if not cmd_res:
            return []
        return cls._NIC_CARD_NAME_RE.findall(cmd_res)

    @classmethod
    def parse_nic_port_num(cls, cmd_res: str) -> str:
        """解析 `hinicadm5 info -i <card>` 回显，提取 port num 字段值。

        回显为 `key : value` 形式，使用 FormParser 解析后取 port num 字段。
        回显示例：
            port num            : 2
        """
        if not cmd_res:
            return ""
        parsed = FormParser(key_separator=":").parse(cmd_res)
        return parsed.get("port num", "")

    @classmethod
    def parse_nic_sfp_info(cls, cmd_res: str, port_id: str) -> NICPortLaneInfoA5:
        """解析 `hinicadm5 sfp -i <card> -p <port>` 回显，提取各 lane 信息。"""
        port_lane_info = NICPortLaneInfoA5(port_id=port_id)
        if not cmd_res:
            return port_lane_info

        # 用字典收集 lane 数据（key=lane号, value=值），不再硬编码 lane 数量
        bias_lanes: Dict[int, str] = {}
        tx_power_lanes: Dict[int, str] = {}
        rx_power_lanes: Dict[int, str] = {}
        tx_los_lanes: Dict[int, str] = {}
        rx_los_lanes: Dict[int, str] = {}
        tx_cdr_lol_lanes: Dict[int, str] = {}
        rx_cdr_lol_lanes: Dict[int, str] = {}
        host_snr_lanes: Dict[int, str] = {}
        media_snr_lanes: Dict[int, str] = {}

        for line in cmd_res.splitlines():
            line = line.strip()
            if not line:
                continue
            # bias/power/snr 类：group(1)=lane, group(2)=value
            cls._fill_lane(line, cls._NIC_BIAS_RE, bias_lanes, lane_group=1, value_group=2)
            cls._fill_lane(line, cls._NIC_TX_POWER_RE, tx_power_lanes, lane_group=1, value_group=2)
            cls._fill_lane(line, cls._NIC_RX_POWER_RE, rx_power_lanes, lane_group=1, value_group=2)
            cls._fill_lane(line, cls._NIC_HOST_SNR_RE, host_snr_lanes, lane_group=1, value_group=2)
            cls._fill_lane(line, cls._NIC_MEDIA_SNR_RE, media_snr_lanes, lane_group=1, value_group=2)
            # flag 类：group(1)=flag, group(2)=lane
            cls._fill_lane(line, cls._NIC_TX_LOS_RE, tx_los_lanes, lane_group=2, value_group=1)
            cls._fill_lane(line, cls._NIC_RX_LOS_RE, rx_los_lanes, lane_group=2, value_group=1)
            cls._fill_lane(line, cls._NIC_TX_CDR_LOL_RE, tx_cdr_lol_lanes, lane_group=2, value_group=1)
            cls._fill_lane(line, cls._NIC_RX_CDR_LOL_RE, rx_cdr_lol_lanes, lane_group=2, value_group=1)

        # 汇总所有出现过的 lane 号，按升序生成连续列表（缺失的填空串）
        all_lane_ids = sorted(
            set(bias_lanes)
            | set(tx_power_lanes)
            | set(rx_power_lanes)
            | set(tx_los_lanes)
            | set(rx_los_lanes)
            | set(tx_cdr_lol_lanes)
            | set(rx_cdr_lol_lanes)
            | set(host_snr_lanes)
            | set(media_snr_lanes)
        )
        if not all_lane_ids:
            return port_lane_info
        # lane 号从 1 连续到 max，缺失的填空串
        max_lane_id = all_lane_ids[-1]

        def to_list(lane_dict: Dict[int, str]) -> List[str]:
            return [lane_dict.get(i, "") for i in range(1, max_lane_id + 1)]

        port_lane_info.bias_lanes = to_list(bias_lanes)
        port_lane_info.tx_power_lanes = to_list(tx_power_lanes)
        port_lane_info.rx_power_lanes = to_list(rx_power_lanes)
        port_lane_info.tx_los_lanes = to_list(tx_los_lanes)
        port_lane_info.rx_los_lanes = to_list(rx_los_lanes)
        port_lane_info.tx_cdr_lol_lanes = to_list(tx_cdr_lol_lanes)
        port_lane_info.rx_cdr_lol_lanes = to_list(rx_cdr_lol_lanes)
        port_lane_info.host_snr_lanes = to_list(host_snr_lanes)
        port_lane_info.media_snr_lanes = to_list(media_snr_lanes)
        return port_lane_info

    @classmethod
    def _fill_lane(
        cls,
        line: str,
        pattern: re.Pattern,
        lanes: Dict[int, str],
        lane_group: int,
        value_group: int,
    ) -> None:
        """按正则匹配 line，将 value 写入 lanes[lane_id]。

        Args:
            lane_group: lane 编号在捕获组中的索引。
            value_group: 值在捕获组中的索引。
                bias/power/snr 类正则：group(1)=lane, group(2)=value。
                flag 类正则：group(1)=flag, group(2)=lane。
        """
        match = pattern.search(line)
        if not match:
            return
        lane_id = int(match.group(lane_group))
        if lane_id >= 1:
            lanes[lane_id] = match.group(value_group)
