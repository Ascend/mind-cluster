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
from typing import List, Dict

from diag_tool.core.common.constants import BIT_ERROR_RATE_LIMIT
from diag_tool.core.config import port_mapping_config
from diag_tool.core.log_parser.base import FindResult
from diag_tool.core.model.switch import OpticalModelBaseInfo, InterfaceBrief, SwiOpticalModel, DeviceInterface, \
    InterfaceMapping, AlarmInfo, InterfaceInfo, BitErrRate, TransceiverInfo, OpticalStateFlagDiagInfo, PortMapping
from diag_tool.utils import logger
from diag_tool.utils.form_parser import FormParser
from diag_tool.utils.helpers import split_str, to_int
from diag_tool.utils.table_parser import TableParser

_DIAG_LOGGER = logger.DIAG_LOGGER
TWO_HUNDRED_GE = "(200GE)"
INTERFACE_STR = "interface"


class SwitchParser:

    @staticmethod
    def parse_op_state_flag_diag_info(cmd_res: str) -> List[OpticalStateFlagDiagInfo]:
        titles_dict = {"items": "Items", "status": "Status"}
        end_sign = "------------"
        parse_data_list = TableParser.parse(cmd_res, titles_dict, separate_title_content_lines_num=1, end_sign=end_sign)
        results = [OpticalStateFlagDiagInfo.from_dict(parse_data) for parse_data in parse_data_list]
        return results

    @staticmethod
    def trans_opt_module_results(cmd_res: str):
        titles_dict = {  # 需与标题顺序一致
            "items": "Items", "value": "Value", "high_alarm": "HighAlarm", "high_warn": "HighWarn",
            "low_alarm": "LowAlarm", "low_warn": "LowWarn", "status": "Status",
        }
        end_sign = "------------"
        parse_data_list = TableParser.parse(cmd_res, titles_dict, separate_title_content_lines_num=1, end_sign=end_sign)
        optical_model_base_info_list = []
        for data in parse_data_list:
            if data.get("items"):
                optical_model_base_info_list.append(OpticalModelBaseInfo.from_dict(data))
        return optical_model_base_info_list

    @staticmethod
    def filter_opt_module_info(opt_module_log_info):
        result_dict = {}
        for log_info in opt_module_log_info:
            info_dict = log_info.info_dict
            if not info_dict:
                continue
            group_key = (f"{info_dict.get('chip_id', '')}{info_dict.get('port_id', '')}{info_dict.get('items', '')}"
                         f"{info_dict.get('lane_id', '')}{info_dict.get('mode', '')}")
            current_time = info_dict.get('time', '')
            if group_key not in result_dict or current_time > result_dict[group_key].get('time', ''):
                result_dict[group_key] = info_dict
        return list(result_dict.values())

    @classmethod
    def parse_bit_err_rate(cls, cmd_res: str, interface_briefs: List[InterfaceBrief]) -> List[BitErrRate]:
        cmd_res_list = split_str(cmd_res, "display interface troubleshooting")
        bit_err_rate_list = []
        for cmd_res_str, interface_brief in zip(cmd_res_list, interface_briefs):
            parse_data_dict = FormParser(multi_key_in_line_separator=["       "]).parse(cmd_res_str)
            bit_err_rate = parse_data_dict.get("Bit error rate")
            if not bit_err_rate:
                continue
            try:
                bit_err_rate_f = float(bit_err_rate)
            except ValueError:
                continue
            if bit_err_rate_f > BIT_ERROR_RATE_LIMIT:
                bit_err_rate_list.append(BitErrRate(interface_brief.interface, bit_err_rate))
        return bit_err_rate_list

    @classmethod
    def parse_lldp_nei_brief(cls, cmd_res: str) -> List[InterfaceMapping]:
        titles_dict = {  # 需与标题顺序一致
            "local_interface": "Local Interface", "exptime": "Exptime(s)", "neighbor_interface": "Neighbor Interface",
            "neighbor_device": "Neighbor Device",
        }
        parse_data_list = TableParser.parse(cmd_res, titles_dict, separate_title_content_lines_num=1)
        device_mapping_list = []
        for data in parse_data_list:
            local_interface = data.get("local_interface")
            neighbor_interface = data.get("neighbor_interface")
            neighbor_device = data.get("neighbor_device")
            if not local_interface or not neighbor_interface or not neighbor_device:
                continue
            neighbor_info = DeviceInterface(neighbor_device, neighbor_interface)
            device_mapping_list.append(InterfaceMapping(local_interface, neighbor_info))
        return device_mapping_list

    @classmethod
    def parse_alarms(cls, cmd_res) -> List[AlarmInfo]:
        titles_dict = {
            "sequence": "Sequence",
            "alarm_id": "AlarmId",
            "severity": "Severity",
            "date_time": "Date Time",
            "description": "Description",
        }
        table = TableParser.parse(cmd_res, titles_dict, {}, 1, end_sign="------", both_strip=False)
        alarms = [AlarmInfo.from_dict(raw) for raw in table]
        cur_alarm = None
        result = []
        for alarm in alarms:
            if alarm.alarm_id:
                cur_alarm = alarm
                result.append(cur_alarm)
            else:
                cur_alarm.date_time += alarm.date_time
                cur_alarm.description += alarm.description
        return result

    @classmethod
    def parse_interface_info(cls, cmd_res) -> List[InterfaceInfo]:
        pattern = r"(.+) current state : [a-zA-Z]+ \(ifindex"
        re_pattern = re.compile(pattern)
        cmd_res_list = split_str(cmd_res, pattern, regex=True)
        result = []
        for cmd_res_str in cmd_res_list:
            form = FormParser().parse(cmd_res_str)
            interface_info = InterfaceInfo.from_dict(form)
            search = re_pattern.search(cmd_res_str)
            if search:
                interface_info.interface_name = search.group(1)
            result.append(interface_info)
        return result

    @classmethod
    def parse_datetime(cls, cmd_res: str) -> str:
        time_str = ""
        for line in cmd_res.splitlines():
            if "-" in line:
                time_str = line.strip()
        return time_str

    @classmethod
    def parse_esn(cls, cmd_res: str) -> str:
        form = FormParser().parse(cmd_res)
        return form.get("ESN", "")

    @classmethod
    def parse_transceiver_info(cls, cmd_res: str) -> List[TransceiverInfo]:
        transceiver_key = "transceiver information"
        cmd_res_list = split_str(cmd_res, transceiver_key)
        result = []
        for part in cmd_res_list:
            form = FormParser(append_multi_line=True, skip_sign="----------------------").parse(part)
            for interface, info in form.items():
                transceiver_info = TransceiverInfo.from_dict(info)
                interface = str(interface).replace(transceiver_key, "").strip()
                transceiver_info.interface = interface
                result.append(transceiver_info)
        return result

    @classmethod
    def parse_alarm_verbose(cls, cmd_res: str) -> List[AlarmInfo]:
        if not cmd_res:
            return []
        parts = re.split(r"[\r\n]{2,}", cmd_res)
        results = []
        for part in parts:
            if not part.strip():
                continue
            form = FormParser(multi_key_in_line_separator=["        "]).parse(part)
            if not form:
                continue
            result = AlarmInfo.from_dict(form)
            results.append(result)
        return results

    @classmethod
    def parse_port_mapping(cls, cmd_res: str) -> Dict[str, str]:
        titles_dict = {
            "interface_name": "Interface",
            "if_index": "IfIndex",
            "tb": "TB",
            "tp": "TP",
            "chip_id": "Chip",
            "port_id": "Port",
            "core": "Core"
        }
        table_data_list = TableParser.parse(cmd_res, titles_dict, separate_title_content_lines_num=1, end_sign="------")
        mapping_list = [PortMapping.from_dict(data) for data in table_data_list]
        interface_chip_port_mapping = {}
        for data in mapping_list:
            if data.interface_name and data.chip_id and data.port_id:
                interface_chip_port_mapping.update({f"{data.chip_id}/{data.port_id}": data.interface_name})
        if interface_chip_port_mapping:
            return interface_chip_port_mapping
        port_mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()
        for interface_name, port_mapping in port_mapping_config_instance.l1_interface_port_map.items():
            if not port_mapping:
                continue
            interface_chip_port_mapping.update({f"{port_mapping.swi_chip_id}/{port_mapping.phy_id}": interface_name})
        return interface_chip_port_mapping

    @classmethod
    def parse_opt_module_info_from_line(
            cls,
            switch_log_info: List[FindResult],
            port_mapping: Dict[str, str]) -> List[SwiOpticalModel]:
        convert_dict = {
            "txpower": "TxPower Lane",
            "rxpower": "RxPower Lane",
            "bias": "Bias Lane",
            "SNR0": "HostSNR Lane",
            "SNR1": "MediaSNR Lane"
        }
        filter_info_dict = cls.filter_opt_module_info(switch_log_info)
        swi_opt_model_info = {}
        for info_dict in filter_info_dict:
            if not info_dict:
                continue
            chip_id = to_int(info_dict.get("chip_id", "")) - 1
            port_id = to_int(info_dict.get("port_id", "")) - 1
            interface_name = port_mapping.get(f"{chip_id}/{port_id}")
            if not interface_name:
                continue
            opt_base_info = OpticalModelBaseInfo.from_dict(info_dict)
            convert_key = opt_base_info.items
            if convert_key == "SNR":
                convert_key = f"{convert_key}{info_dict.get('mode', '')}"
            items = convert_dict.get(convert_key)
            if not items:
                continue
            lane_id = to_int(info_dict.get("lane_id", ""))
            opt_base_info.items = f"{items}{lane_id}"
            swi_opt_model_info.setdefault(interface_name, []).append(opt_base_info)

        return [SwiOpticalModel(interface_name, base_info) for interface_name, base_info in swi_opt_model_info.items()]

    @classmethod
    def parse_interface_brief(cls, cmd_res: str) -> List[InterfaceBrief]:
        titles_dict = {
            "interface": "Interface", "phy": "PHY", "protocol": "Protocol", "in_uti": "InUti",
            "out_uti": "OutUti", "in_errors": "inErrors", "out_errors": "outErrors",
        }
        parse_data_list = TableParser.parse(cmd_res, titles_dict)
        interface_info_list = []
        for data in parse_data_list:
            interface_name = data.get(INTERFACE_STR)
            if not interface_name:
                continue
            if TWO_HUNDRED_GE in interface_name:
                data[INTERFACE_STR] = interface_name.replace(TWO_HUNDRED_GE, "")
            interface_instance = InterfaceBrief.from_dict(data)
            interface_info_list.append(interface_instance)
        return interface_info_list

    @classmethod
    def parse_opt_module_info_from_table(
            cls, cmd_res: str, interface_briefs: List[InterfaceBrief]
    ) -> List[SwiOpticalModel]:
        cmd_res_list = split_str(cmd_res, "dis optical-module interface")
        optical_model_list = []
        for cmd_res_str, interface_brief in zip(cmd_res_list, interface_briefs):
            optical_model_base_info = cls.trans_opt_module_results(cmd_res_str)
            op_state_flag_diag_info = cls.parse_op_state_flag_diag_info(cmd_res_str)
            if not optical_model_base_info:
                continue
            optical_model = SwiOpticalModel(interface_brief.interface, optical_model_base_info, op_state_flag_diag_info)
            optical_model_list.append(optical_model)
        return optical_model_list
