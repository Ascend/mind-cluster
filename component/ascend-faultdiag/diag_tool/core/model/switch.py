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
import re
from typing import List, Dict, Tuple

from diag_tool.core.common.diag_enum import PowerUnitType
from diag_tool.core.common.json_obj import JsonObj
from diag_tool.core.model.hccs import HccsInfo
from diag_tool.core.model.inspection import InspectionInterfaceInfo
from diag_tool.core.model.optical_module import OpticalModule, OpticalModuleInfo, LanePowerInfo
from diag_tool.utils import helpers, list_tool


class InterfaceBrief(JsonObj):
    def __init__(self, interface: str = "", phy: str = "", protocol: str = "", in_uti: str = "", out_uti: str = "",
                 in_errors: str = "", out_errors: str = ""):
        self.interface = interface
        self.phy = phy
        self.protocol = protocol
        self.in_uti = in_uti
        self.out_uti = out_uti
        self.in_errors = in_errors
        self.out_errors = out_errors


class OpticalModelBaseInfo(JsonObj):
    def __init__(self, items: str = "", value: str = "", high_alarm: str = "", low_alarm: str = "", high_warn: str = "",
                 low_warn: str = "", status: str = ""):
        self.items = items
        self.value = value
        self.high_alarm = high_alarm
        self.low_alarm = low_alarm
        self.high_warn = high_warn
        self.low_warn = low_warn
        self.status = status

    def to_str(self, format_type="simple"):
        if format_type == "alarm":
            return f"{self.items}:{self.value}[{self.low_alarm}:{self.high_alarm}]"
        if format_type == "warn":
            return f"{self.items}:{self.value}[{self.low_warn}:{self.high_warn}]"
        return f"{self.items}:{self.value}"


class OpticalStateFlagDiagInfo(JsonObj):

    def __init__(self, items="", status=""):
        self.items = items
        self.status = status


class SwiOpticalModel(JsonObj):
    _TX_POWER = "TxPower"
    _RX_POWER = "RxPower"
    _BIAS = "Bias"
    _MEDIA_SNR = "MediaSNR"
    _HOST_SNR = "HostSNR"
    _MAX_LANE_SIZE = 8

    def __init__(self, interface_name: str = "", base_info: List[OpticalModelBaseInfo] = None,
                 state_flag_diag_infos: List[OpticalStateFlagDiagInfo] = None):
        self.interface_name = interface_name
        self.base_info: List[OpticalModelBaseInfo] = base_info or []
        self.state_flag_diag_infos: List[OpticalStateFlagDiagInfo] = state_flag_diag_infos or []

    def get_lane_power_infos(self) -> List[LanePowerInfo]:
        if not self.base_info:
            return []
        lane_power_infos = []
        for lane in range(self._MAX_LANE_SIZE):
            lane_str = str(lane)
            lane_power_info = LanePowerInfo(lane_id=lane_str, power_unit_type=PowerUnitType.DBM)
            for opt_module_info in self.base_info:
                if lane_str not in opt_module_info.items:
                    continue
                if self._RX_POWER in opt_module_info.items:
                    lane_power_info.rx_power = opt_module_info.value
                elif self._TX_POWER in opt_module_info.items:
                    lane_power_info.tx_power = opt_module_info.value
                elif self._BIAS in opt_module_info.items:
                    lane_power_info.bias = opt_module_info.value
                elif self._MEDIA_SNR in opt_module_info.items:
                    lane_power_info.media_snr = opt_module_info.value
                elif self._HOST_SNR in opt_module_info.items:
                    lane_power_info.host_snr = opt_module_info.value
            lane_power_infos.append(lane_power_info)
        return lane_power_infos


class BitErrRate(JsonObj):
    def __init__(self, interface_name: str = "", bit_err_rate: str = ""):
        self.interface_name = interface_name
        self.bit_err_rate = bit_err_rate


class DeviceInterface(JsonObj):
    def __init__(self, device_name: str = "", interface: str = ""):
        self.device_name = device_name
        self.interface = interface


class InterfaceMapping(JsonObj):
    def __init__(self, local_interface_name: str = "", remote_device_interface: DeviceInterface = None):
        self.local_interface_name = local_interface_name
        self.remote_device_interface = remote_device_interface


class AlarmInfo(JsonObj):

    def __init__(self, alarm_id: str = "", severity: str = "", date_time: str = "", description: str = "",
                 start_time="", clear_time=""):
        self.alarm_id = alarm_id
        self.alarm_id_int = helpers.parse_hex(alarm_id)
        self.severity = severity
        self.date_time = date_time or start_time
        self.clear_time = clear_time
        self.description = description

    @classmethod
    def _parse_to_py_key(cls):
        return True


class ManufactureInfo(JsonObj):

    def __init__(self, manu_serial_number=""):
        self.manu_serial_number = manu_serial_number

    @classmethod
    def _parse_to_py_key(cls):
        return True


class TransceiverDiagInfo(JsonObj):

    def __init__(self, bias_current_m_a="", current_rx_power_d_bm="", current_tx_power_d_bm=""):
        self.bias_current_m_a = bias_current_m_a
        self.current_rx_power_d_bm = current_rx_power_d_bm
        self.current_tx_power_d_bm = current_tx_power_d_bm

    @classmethod
    def _parse_to_py_key(cls):
        return True


class DiagEnhancedInfo(JsonObj):

    def __init__(self, odsp_junction_temperature_celsius="", optical_snr=""):
        self.odsp_junction_temperature_celsius = odsp_junction_temperature_celsius
        self.optical_snr = optical_snr

    @classmethod
    def _parse_to_py_key(cls):
        return True


class TransceiverInfo(JsonObj):
    _LANE_PATTERN = re.compile(r"([\d.-]{1,7})\|([\d.-]{1,7}) {1,7}\(Lane(\d{1,2})\|Lane(\d{1,2})\)")

    def __init__(self, interface="", manufacture_information: ManufactureInfo = ManufactureInfo(),
                 diagnostic_information: TransceiverDiagInfo = None,
                 diagnostic_enhanced_information: DiagEnhancedInfo = None):
        self.interface = interface
        self.manufacture_information = manufacture_information
        self.diagnostic_information = diagnostic_information
        self.diagnostic_enhanced_information = diagnostic_enhanced_information

    @classmethod
    def _parse_to_py_key(cls):
        return True

    def get_lane_power_infos(self) -> List[LanePowerInfo]:
        diag_info = self.diagnostic_information
        if not diag_info:
            return []
        bais_lanes = self._parse_lanes(diag_info.bias_current_m_a)
        cur_rx_powers = self._parse_lanes(diag_info.current_rx_power_d_bm)
        cur_tx_powers = self._parse_lanes(diag_info.current_tx_power_d_bm)
        media_snrs = []
        if self.diagnostic_enhanced_information:
            media_snrs = self._parse_lanes(self.diagnostic_enhanced_information.optical_snr)

        lanes_map = {"bais_lanes": bais_lanes, "cur_tx_powers": cur_tx_powers, "cur_rx_powers": cur_rx_powers,
                     "media_snrs": media_snrs}
        max_len = len(max(lanes_map.values(), key=len))
        max_len_lanes = {k: v for k, v in lanes_map.items() if len(v) == max_len}
        if not max_len_lanes:
            return []
        results = []
        for i in range(max_len):
            result = LanePowerInfo(power_unit_type=PowerUnitType.DBM)
            for k, v in max_len_lanes.items():
                if not result.lane_id:
                    result.lane_id = v[i][0]
                if k == "bais_lanes":
                    result.bias = v[i][1]
                elif k == "cur_tx_powers":
                    result.tx_power = v[i][1]
                elif k == "cur_rx_powers":
                    result.rx_power = v[i][1]
                elif k == "media_snrs":
                    result.media_snr = v[i][1]
            results.append(result)
        return results

    def _parse_lanes(self, lane_str: str) -> List[Tuple[str, str]]:
        """
        解析形如以下格式的数据返回Lane序号, lane值
        7.35|7.35  (Lane0|Lane1)
        7.35|7.35  (Lane2|Lane3)
        """
        results = []
        for part in self._LANE_PATTERN.findall(lane_str):
            results.append((part[2], part[0]))
            results.append((part[3], part[1]))
        return results


class InterfaceInfo(JsonObj):

    def __init__(self, interface_name="", speed="", duplex=""):
        self.interface_name = interface_name
        self.speed = speed
        self.duplex = duplex

    @classmethod
    def _parse_to_py_key(cls):
        return True


class InterfaceFullInfo(JsonObj, OpticalModule):

    def __init__(self, interface: str,
                 device_sn: str = "",
                 device_id: str = "",
                 device_name: str = "",
                 interface_info: InterfaceInfo = None,
                 transceiver_info: TransceiverInfo = None,
                 interface_mapping: InterfaceMapping = None,
                 bit_err_rate: BitErrRate = None,
                 swi_optical_model: SwiOpticalModel = None,
                 interface_brief: InterfaceBrief = None):
        self.interface = interface
        self.device_name = device_name
        self.device_id = device_id
        self.device_sn = device_sn
        self.interface_info = interface_info
        self.transceiver_info = transceiver_info
        self.interface_mapping = interface_mapping
        self.bit_err_rate = bit_err_rate
        self.swi_optical_model = swi_optical_model
        self.interface_brief = interface_brief
        self._optical_module_info: OpticalModuleInfo = None

    def get_optical_module_info(self) -> OpticalModuleInfo:
        if self._optical_module_info:
            return self._optical_module_info
        if not self.swi_optical_model and not self.transceiver_info:
            return None
        lane_power_infos = self.swi_optical_model and self.swi_optical_model.get_lane_power_infos()
        if not lane_power_infos:
            lane_power_infos = self.transceiver_info.get_lane_power_infos()
        self._optical_module_info = OpticalModuleInfo(lane_power_infos, self.transceiver_info and
                                                      self.transceiver_info.manufacture_information.manu_serial_number)
        return self._optical_module_info

    def get_inspection_interface_info(self) -> InspectionInterfaceInfo:
        return InspectionInterfaceInfo(
            device_name=self.device_name,
            device_id=self.device_id,
            device_sn=self.device_sn,
            interface=self.interface,
            interface_sn=self._get_sn_num()
        )

    def _get_sn_num(self):
        if not self.transceiver_info or not self.transceiver_info.manufacture_information:
            return None
        return self.transceiver_info.manufacture_information.manu_serial_number


class PortMapping(JsonObj):
    def __init__(self, interface_name: str = "", chip_id: str = "", port_id: str = ""):
        self.interface_name = interface_name
        self.chip_id = chip_id
        self.port_id = port_id


class SwitchInfo(JsonObj):
    def __init__(self, name: str, swi_id: str, sn="", optical_models: List[SwiOpticalModel] = None,
                 interface_briefs: List[InterfaceBrief] = None,
                 interface_mapping: List[InterfaceMapping] = None,
                 active_alarm_info: List[AlarmInfo] = None,
                 history_alarm_info: List[AlarmInfo] = None,
                 interface_info: List[InterfaceInfo] = None,
                 hccs_info: HccsInfo = None,
                 date_time: str = "",
                 bit_error_rate: List[BitErrRate] = None,
                 transceiver_infos: List[TransceiverInfo] = None):
        self.sn = sn
        self.name = name
        self.swi_id = swi_id  # 设备的唯一标志, 可以是IP, 名称, SN号
        self.optical_models = optical_models or []
        self.interface_briefs = interface_briefs or []
        self.interface_mapping = interface_mapping or []
        self.active_alarm_info = active_alarm_info or []
        self.history_alarm_info = history_alarm_info or []
        self.interface_info = interface_info
        self.hccs_info = hccs_info
        self.date_time = date_time
        self.bit_error_rate = bit_error_rate or []
        self.transceiver_infos = transceiver_infos or []
        self._interface_full_infos: Dict[str, InterfaceFullInfo] = {}

    @property
    def interface_full_infos(self) -> Dict[str, InterfaceFullInfo]:
        if self._interface_full_infos:
            return self._interface_full_infos
        interface_full_info = {}
        for interface_brief in self.interface_briefs:
            interface = interface_brief.interface
            full_info = InterfaceFullInfo(
                interface=interface,
                device_id=self.swi_id,
                device_sn=self.sn,
                device_name=self.name,
                interface_info=list_tool.find_first(self.interface_info,
                                                    lambda info: info.interface_name == interface),
                transceiver_info=list_tool.find_first(self.transceiver_infos,
                                                      lambda info: info.interface == interface),
                interface_mapping=list_tool.find_first(self.interface_mapping,
                                                       lambda info: info.local_interface_name == interface),
                bit_err_rate=list_tool.find_first(self.bit_error_rate,
                                                  lambda rate: rate.interface_name == interface),
                swi_optical_model=list_tool.find_first(self.optical_models,
                                                       lambda rate: rate.interface_name == interface),
                interface_brief=interface_brief
            )
            interface_full_info[interface] = full_info
        self._interface_full_infos = interface_full_info
        return interface_full_info
