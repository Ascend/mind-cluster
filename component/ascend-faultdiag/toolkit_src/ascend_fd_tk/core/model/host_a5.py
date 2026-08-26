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

from typing import List, Dict

from ascend_fd_tk.core.common.constants import OP_PRESENT, HIGH_POWER_ENABLE_A5
from ascend_fd_tk.core.common.diag_enum import PowerUnitType
from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.core.log_parser.base import FindResult
from ascend_fd_tk.core.context.host_registry import register_host_info
from ascend_fd_tk.core.model.optical_module import OpticalModuleInfo, LanePowerInfo


class OpticalModuleHardwareAttr(JsonObj):
    """光模块硬件属性实体"""

    def __init__(
        self,
        npu_id: str = "",
        optical_id: str = "",
        present: str = "",
        high_power: str = "",
        identifier: str = "",
        wave_len: str = "",
        phy_code: str = "",
        revision: str = "",
    ):
        self.npu_id = npu_id
        self.optical_id = optical_id
        self.present = present
        self.high_power = high_power
        self.identifier = identifier
        self.wave_len = wave_len
        self.phy_code = phy_code
        self.revision = revision

    @classmethod
    def parse_title_dict(cls) -> Dict[str, str]:
        return {
            "npu_id": "NPU",
            "optical_id": "Optical",
            "present": "Present",
            "high_power": "High_power",
            "identifier": "Identifier",
            "wave_len": "Wave_len",
            "phy_code": "Phy_code",
            "revision": "Revision",
        }

    def is_optical_present(self) -> bool:
        return self.present == OP_PRESENT

    def is_high_power(self) -> bool:
        return self.high_power == HIGH_POWER_ENABLE_A5


class OpticalModuleSerialInfo(JsonObj):
    """光模块生产序列号信息实体"""

    def __init__(
        self,
        npu_id: str = "",
        optical_id: str = "",
        manufacture_name: str = "",
        part_number: str = "",
        serial_number: str = "",
        org_unique_id: str = "",
        manufacture_date: str = "",
    ):
        self.npu_id = npu_id
        self.optical_id = optical_id
        self.manufacture_name = manufacture_name
        self.part_number = part_number
        self.serial_number = serial_number
        self.org_unique_id = org_unique_id
        self.manufacture_date = manufacture_date

    @classmethod
    def parse_title_dict(cls) -> Dict[str, str]:
        return {
            "npu_id": "NPU",
            "optical_id": "Optical",
            "manufacture_name": "Name",
            "part_number": "Part_number",
            "serial_number": "Serial_number",
            "org_unique_id": "Org_unique_id",
            "manufacture_date": "Manufact_date",
        }


class OpticalStateFlag(JsonObj):
    """光模块 State Flag 信息实体。

    lane 数量不固定（取决于光模块规格，可能 2/4/8 等），
    `lanes` 索引 0 对应回显 Lane0，依次类推。
    """

    def __init__(
        self,
        npu_id: str = "",
        optical_id: str = "",
        items: str = "",
        lanes: List[str] = None,
    ):
        self.npu_id = npu_id
        self.optical_id = optical_id
        self.items = items
        self.lanes = lanes or []

    @classmethod
    def parse_title_dict(cls) -> Dict[str, str]:
        """返回固定核心字段标题；Lane 列由 parser 扫描回显动态补全。"""
        return {
            "npu_id": "NPU",
            "optical_id": "Optical",
            "items": "Items",
            "lanes": "Lane",
        }


class OpticalModuleMonitorItem(JsonObj):
    """光模块阈值告警监测项"""

    def __init__(
        self,
        npu_id: str = "",
        optical_id: str = "",
        items: str = "",
        value: str = "",
        high_alarm: str = "",
        high_warn: str = "",
        low_warn: str = "",
        low_alarm: str = "",
        status: str = "",
        used_by: str = "",
    ):
        self.npu_id = npu_id
        self.optical_id = optical_id
        self.items = items
        self.value = value
        self.high_alarm = high_alarm
        self.high_warn = high_warn
        self.low_warn = low_warn
        self.low_alarm = low_alarm
        self.status = status
        self.used_by = used_by

    @classmethod
    def parse_title_dict(cls) -> Dict[str, str]:
        return {
            "npu_id": "NPU",
            "optical_id": "Optical",
            "items": "Items",
            "value": "Value",
            "high_alarm": "HighAlarm",
            "high_warn": "HighWarn",
            "low_warn": "LowWarn",
            "low_alarm": "LowAlarm",
            "status": "Status",
            "used_by": "Used_by",
        }


class OpticalLaneInfo(JsonObj):
    """光模块阈值告警监测项"""

    def __init__(
        self,
        lane_id: str = "",
    ):
        self.npu_id = lane_id


class HCCNOpticalInfoA5(JsonObj):
    """A5 光模块信息"""

    def __init__(
        self,
        npu_id: str = "",
        optical_id: str = "",
        hardware_attr: List[OpticalModuleHardwareAttr] = None,
        serial_info: List[OpticalModuleSerialInfo] = None,
        state_flag: List[OpticalStateFlag] = None,
        monitor_item: List[OpticalModuleMonitorItem] = None,
    ):
        self.npu_id = npu_id
        self.optical_id = optical_id
        self.hardware_attr = hardware_attr or []
        self.serial_info = serial_info or []
        self.state_flag = state_flag or []
        self.monitor_item = monitor_item or []


class NpuChipInfoA5(JsonObj):
    def __init__(
        self,
        npu_id="",
        chip_id="",
        chip_phy_id="",
        npu_type="",
        hccn_optical_info: Dict[str, HCCNOpticalInfoA5] = None,
    ):
        # pylint: disable=R0801
        self.hccn_optical_info = hccn_optical_info
        self._optical_module_info: OpticalModuleInfo = None
        # 关系属性
        self.npu_type = npu_type
        self.npu_id = npu_id  # 0-7
        self.chip_id = chip_id  # 0-1
        self.chip_phy_id = chip_phy_id  # 0-15

    def get_optical_module_info(self) -> OpticalModuleInfo:
        """返回第一个光模块信息（向后兼容 A3 接口）。"""
        if self._optical_module_info:
            return self._optical_module_info
        infos = self.get_optical_module_infos()
        if not infos:
            return None
        self._optical_module_info = infos[0]
        return self._optical_module_info

    def get_optical_module_infos(self) -> List[OpticalModuleInfo]:
        """返回所有光模块信息（A5 单 NPU 可能有多光模块）。

        从每个光模块的 monitor_item 中按 lane 号解析
        TxPower/RxPower/Bias/HostSNR/MediaSNR，汇总为 LanePowerInfo 列表；
        未采集到的 lane 不填，由报告层按需补空。
        """
        results: List[OpticalModuleInfo] = []
        if not self.hccn_optical_info:
            return results
        for optical_info in self.hccn_optical_info.values():
            if not optical_info or not optical_info.monitor_item:
                continue
            module_info = self._parse_optical_info_to_module_info(optical_info)
            if module_info:
                results.append(module_info)
        return results

    @staticmethod
    def _parse_optical_info_to_module_info(optical_info: "HCCNOpticalInfoA5") -> OpticalModuleInfo:
        """将单个 HCCNOpticalInfoA5 解析为 OpticalModuleInfo。"""
        lane_map: Dict[str, LanePowerInfo] = {}
        for item in optical_info.monitor_item:
            items_name = item.items or ""
            value = item.value or ""
            lane_id = ""
            field = ""
            for prefix, key in (
                ("TxPower Lane", "tx_power"),
                ("RxPower Lane", "rx_power"),
                ("Bias Lane", "bias"),
                ("HostSNR Lane", "host_snr"),
                ("MediaSNR Lane", "media_snr"),
            ):
                if items_name.startswith(prefix):
                    lane_id = items_name.replace(prefix, "").split("(")[0].strip()
                    field = key
                    break
            if not lane_id or not field:
                continue
            lane_info = lane_map.setdefault(lane_id, LanePowerInfo(lane_id=lane_id, power_unit_type=PowerUnitType.DBM))
            setattr(lane_info, field, value)
        if not lane_map:
            return None
        # 按 lane 号升序排序
        lane_power_infos = [lane_map[k] for k in sorted(lane_map.keys(), key=lambda x: int(x) if x.isdigit() else 0)]
        # 光模块 SN 取 serial_info 第一条
        sn = ""
        if optical_info.serial_info:
            sn = optical_info.serial_info[0].serial_number or ""
        return OpticalModuleInfo(
            lane_power_infos=lane_power_infos,
            sn=sn,
            optical_id=optical_info.optical_id or "",
        )


class OpticalTopHeadline(JsonObj):
    def __init__(
        self,
        npu_id: str,
        optical_silk_screen_num: str,
        optical_id="",
        speed="",
        used_for="",
    ):
        self.npu_id = npu_id
        self.optical_silk_screen_num = optical_silk_screen_num
        self.optical_id = optical_id
        self.speed = speed
        self.used_for = used_for


class NICPortLaneInfoA5(JsonObj):
    """A5 网卡单端口 SFP 采集的 lane 信息（8 lane）。

    lane 列表索引 0~7 对应回显中的 lane 1~8；未采集到的 lane 为空串。
    flag 类字段（tx_los/rx_los/tx_cdr_lol/rx_cdr_lol）存原始标志值（如 "0"/"0x1"），
    "0" 表示正常，非 "0" 表示异常。
    """

    def __init__(
        self,
        port_id: str = "",
        bias_lanes: List[str] = None,
        tx_power_lanes: List[str] = None,
        rx_power_lanes: List[str] = None,
        tx_los_lanes: List[str] = None,
        rx_los_lanes: List[str] = None,
        tx_cdr_lol_lanes: List[str] = None,
        rx_cdr_lol_lanes: List[str] = None,
        host_snr_lanes: List[str] = None,
        media_snr_lanes: List[str] = None,
    ):
        self.port_id = port_id
        self.bias_lanes = bias_lanes or []
        self.tx_power_lanes = tx_power_lanes or []
        self.rx_power_lanes = rx_power_lanes or []
        self.tx_los_lanes = tx_los_lanes or []
        self.rx_los_lanes = rx_los_lanes or []
        self.tx_cdr_lol_lanes = tx_cdr_lol_lanes or []
        self.rx_cdr_lol_lanes = rx_cdr_lol_lanes or []
        self.host_snr_lanes = host_snr_lanes or []
        self.media_snr_lanes = media_snr_lanes or []


class NICInfoA5(JsonObj):
    """A5 网卡信息"""

    def __init__(
        self,
        card_name: str = "",
        port_num: str = "",
        port_lane_info_list: List[NICPortLaneInfoA5] = None,
    ):
        self.card_name = card_name
        self.port_num = port_num
        self.port_lane_info_list = port_lane_info_list or []


@register_host_info("A5")
class HostInfoA5(JsonObj):
    """A5 代际主机信息（与 HostInfo 完全独立，不继承）。

    与 A3 的差异：npu_chip_info 的 value 类型为 NpuChipInfoA5（字段集与 NpuChipInfo 不同）。
    公共字段（host_id/sn_num/hostname/room_name/...）与 HostInfo 保持一致，便于共用诊断框架。
    """

    def __init__(
        self,
        host_id: str,
        sn_num: str,
        hostname="",
        room_name="",
        cabinet_id="",
        server_superpod_id="",
        server_index="",
        chip_generation: str = "A5",
        msnpureport_log: List[FindResult] = None,
        npu_chip_info: Dict[str, NpuChipInfoA5] = None,
        nic_info_list: List[NICInfoA5] = None,
    ):
        # pylint: disable=R0801
        self.host_id = host_id
        self.sn_num = sn_num
        self.hostname = hostname
        self.room_name = room_name
        self.cabinet_id = cabinet_id
        self.server_superpod_id = server_superpod_id
        self.server_index = server_index
        self.chip_generation = chip_generation
        self.msnpureport_log = msnpureport_log or []
        self.npu_chip_info = npu_chip_info or {}
        self.nic_info_list = nic_info_list or []

    def get_msn_logs_by_type(self, type_key: str) -> List[FindResult]:
        if not type_key:
            return []
        return [find_result for find_result in self.msnpureport_log if find_result.pattern_key == type_key]


class OpticalModuleInfoA5(JsonObj):
    def __init__(self, lane_power_infos: List[LanePowerInfo] = None):
        self.lane_power_infos: List[LanePowerInfo] = lane_power_infos or []
