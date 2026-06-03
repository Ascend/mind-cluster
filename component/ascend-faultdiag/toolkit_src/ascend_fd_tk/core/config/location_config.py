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

from dataclasses import dataclass, field
import os
from typing import Dict, Optional

from openpyxl import load_workbook

from ascend_fd_tk.utils.logger import DIAG_LOGGER


@dataclass
class HostLocationInfo:
    host_name: str = ""
    room_name: str = ""
    cabinet_id: str = ""
    l1_switch_name: str = ""
    l1_switch_ip: str = ""


@dataclass
class SwitchLocationInfo:
    room_name: str = ""
    cabinet_id: str = ""
    l2_switch_name: str = ""
    l2_switch_ip: str = ""


@dataclass
class LocationConfig:
    host_location_map: Dict[str, HostLocationInfo] = field(default_factory=dict)
    switch_location_map: Dict[str, SwitchLocationInfo] = field(default_factory=dict)

    def get_host_room_name(self, host_name: str) -> str:
        info = self.host_location_map.get(host_name)
        return info.room_name if info else ""

    def get_host_cabinet_id(self, host_name: str) -> str:
        info = self.host_location_map.get(host_name)
        return info.cabinet_id if info else ""

    def get_switch_room_name(self, switch_identifier: str) -> str:
        info = self._find_switch_location(switch_identifier)
        return info.room_name if info else ""

    def get_switch_cabinet_id(self, switch_identifier: str) -> str:
        info = self._find_switch_location(switch_identifier)
        return info.cabinet_id if info else ""

    def enrich_host_info(self, host_info):
        if host_info.hostname:
            host_info.room_name = self.get_host_room_name(host_info.hostname)
            host_info.cabinet_id = self.get_host_cabinet_id(host_info.hostname)

    def enrich_switch_info(self, switch_info):
        if switch_info.name:
            switch_info.room_name = self.get_switch_room_name(switch_info.name)
            switch_info.cabinet_id = self.get_switch_cabinet_id(switch_info.name)

    def _find_switch_location(self, switch_identifier: str) -> Optional[SwitchLocationInfo]:
        info = self.switch_location_map.get(switch_identifier)
        if info:
            return info
        for key, loc_info in self.switch_location_map.items():
            if loc_info.l2_switch_name == switch_identifier:
                return loc_info
        return None

    def to_dict(self):
        return {
            "host_location_map": {k: v.__dict__ for k, v in self.host_location_map.items()},
            "switch_location_map": {k: v.__dict__ for k, v in self.switch_location_map.items()},
        }

    @classmethod
    def from_dict(cls, data: Dict):
        config = cls()
        for k, v in data.get("host_location_map", {}).items():
            config.host_location_map[k] = HostLocationInfo(**v)
        for k, v in data.get("switch_location_map", {}).items():
            config.switch_location_map[k] = SwitchLocationInfo(**v)
        return config


class LldConfigReader:
    LLD_FILE_NAME = "LLD.xlsx"
    L1_SHEET_NAME = "灵衢L1网络对应关系"
    L2_SHEET_NAME = "灵衢L2网络对应关系"

    @staticmethod
    def _build_col_index(header: list, expected_cols: list) -> Dict[str, int]:
        col_idx = {}
        for col_name in expected_cols:
            for i, h in enumerate(header):
                if h == col_name:
                    col_idx[col_name] = i
                    break
        return col_idx

    @staticmethod
    def _find_sheet_by_keyword(wb, keyword: str):
        for name in wb.sheetnames:
            if keyword in name:
                return wb[name]
        return None

    def read_from_dir(self, dir_path: str) -> Optional[LocationConfig]:
        file_path = os.path.join(dir_path, self.LLD_FILE_NAME)
        if not os.path.exists(file_path):
            DIAG_LOGGER.warning("目录%s中未找到%s配置文件", dir_path, self.LLD_FILE_NAME)
            return None
        DIAG_LOGGER.info("读取配置文件: %s", file_path)
        try:
            wb = load_workbook(filename=file_path, read_only=True, data_only=True)
        except Exception as e:
            DIAG_LOGGER.error("读取配置文件%s失败: %s", file_path, e)
            return None
        try:
            return self._read(wb)
        finally:
            wb.close()

    def _read(self, wb) -> LocationConfig:
        config = LocationConfig()
        config.host_location_map = self._read_l1_sheet(wb)
        config.switch_location_map = self._read_l2_sheet(wb)
        host_count = len(config.host_location_map)
        switch_count = len(config.switch_location_map)
        DIAG_LOGGER.info(
            "加载 %d 条%s信息，%d 条%s信息", host_count, self.L1_SHEET_NAME, switch_count, self.L2_SHEET_NAME
        )
        return config

    def _iter_sheet_rows(self, wb, sheet_name: str, expected_cols: list):
        """迭代 Excel sheet 的数据行，yield (col_idx, values)"""
        ws = self._find_sheet_by_keyword(wb, sheet_name)
        if not ws:
            DIAG_LOGGER.warning("未找到包含%s的Sheet", sheet_name)
            return
        rows = list(ws.iter_rows(values_only=True))
        if len(rows) <= 1:
            DIAG_LOGGER.warning("工作表【%s】内容行数不足，有效数据为空或仅一行", sheet_name)
            return
        header = [str(cell or "").strip() for cell in rows[0]]
        col_idx = self._build_col_index(header, expected_cols)
        for row in rows[1:]:
            if not row:
                continue
            values = [str(cell or "").strip() if cell else "" for cell in row]
            yield col_idx, values

    def _read_l1_sheet(self, wb) -> Dict[str, HostLocationInfo]:
        result = {}
        for col_idx, values in self._iter_sheet_rows(
            wb, self.L1_SHEET_NAME, ["服务器", "机房名称", "机柜编号", "L1名称", "L1_IP"]
        ):
            host_name = values[col_idx["服务器"]] if "服务器" in col_idx else ""
            if not host_name:
                continue
            result[host_name] = HostLocationInfo(
                host_name=host_name,
                room_name=values[col_idx["机房名称"]] if "机房名称" in col_idx else "",
                cabinet_id=values[col_idx["机柜编号"]] if "机柜编号" in col_idx else "",
                l1_switch_name=values[col_idx["L1名称"]] if "L1名称" in col_idx else "",
                l1_switch_ip=values[col_idx["L1_IP"]] if "L1_IP" in col_idx else "",
            )
        return result

    def _read_l2_sheet(self, wb) -> Dict[str, SwitchLocationInfo]:
        result = {}
        for col_idx, values in self._iter_sheet_rows(
            wb, self.L2_SHEET_NAME, ["设备名", "机房名称", "机柜编号", "管理IP配置"]
        ):
            switch_ip = values[col_idx["管理IP配置"]] if "管理IP配置" in col_idx else ""
            switch_name = values[col_idx["设备名"]] if "设备名" in col_idx else ""
            if not switch_ip and not switch_name:
                continue
            loc_info = SwitchLocationInfo(
                room_name=values[col_idx["机房名称"]] if "机房名称" in col_idx else "",
                cabinet_id=values[col_idx["机柜编号"]] if "机柜编号" in col_idx else "",
                l2_switch_name=switch_name,
                l2_switch_ip=switch_ip,
            )
            key = switch_ip if switch_ip else switch_name
            result[key] = loc_info
        return result
