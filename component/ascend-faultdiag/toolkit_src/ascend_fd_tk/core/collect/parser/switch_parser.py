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

from ascend_fd_tk.core.config import port_mapping_config
from ascend_fd_tk.core.log_parser.base import FindResult
from ascend_fd_tk.core.model.switch import (
    OpticalModelBaseInfo,
    InterfaceBrief,
    SwiOpticalModel,
    DeviceInterface,
    InterfaceMapping,
    AlarmInfo,
    InterfaceInfo,
    BitErrRate,
    TransceiverInfo,
    OpticalStateFlagDiagInfo,
    PortMapping,
    QosCreditPortInfo,
    QosCreditInfo,
    QosCreditRow,
)
from ascend_fd_tk.utils import logger
from ascend_fd_tk.utils.form_parser import FormParser
from ascend_fd_tk.utils.helpers import split_str, to_int
from ascend_fd_tk.utils.table_parser import TableParser

_DIAG_LOGGER = logger.DIAG_LOGGER
INTERFACE_STR = "interface"


class SwitchParser:
    _TRANSCEIVER_PATTERN = r"(.{4,20}) transceiver(\d{0,3}) information"
    _TRANSCEIVER_PATTERN_RE = re.compile(_TRANSCEIVER_PATTERN)

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
            "items": "Items",
            "value": "Value",
            "high_alarm": "HighAlarm",
            "high_warn": "HighWarn",
            "low_alarm": "LowAlarm",
            "low_warn": "LowWarn",
            "status": "Status",
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
            group_key = (
                f"{info_dict.get('chip_id', '')}{info_dict.get('port_id', '')}{info_dict.get('items', '')}"
                f"{info_dict.get('lane_id', '')}{info_dict.get('mode', '')}"
            )
            current_time = info_dict.get('time', '')
            if group_key not in result_dict or current_time > result_dict[group_key].get('time', ''):
                result_dict[group_key] = info_dict
        return list(result_dict.values())

    @classmethod
    def parse_bit_err_rate(cls, cmd_res: str) -> List[BitErrRate]:
        pattern = r"(.{1,10}/\d/.{1,15}) :"
        re_pattern = re.compile(pattern)
        cmd_res_list = split_str(cmd_res, pattern, regex=True)
        bit_err_rate_list = []
        for cmd_res_part in cmd_res_list:
            parse_data_dict = FormParser(multi_key_in_line_separator=["        "]).parse(cmd_res_part)
            bit_err_rate = parse_data_dict.get("Bit error rate")
            if not bit_err_rate:
                continue
            search = re_pattern.search(cmd_res_part)
            interface = ""
            if search:
                interface = search.group(1)
            if interface:
                bit_err_rate_list.append(BitErrRate(interface, bit_err_rate))
        return bit_err_rate_list

    @classmethod
    def parse_lldp_nei_brief(cls, cmd_res: str) -> List[InterfaceMapping]:
        titles_dict = {  # 需与标题顺序一致
            "local_interface": "Local Interface",
            "exptime": "Exptime(s)",
            "neighbor_interface": "Neighbor Interface",
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
        cmd_res_list = split_str(cmd_res, cls._TRANSCEIVER_PATTERN, regex=True)
        result = []
        for part in cmd_res_list:
            form = FormParser(append_multi_line=True, skip_sign="----------------------").parse(part)
            for interface, info in form.items():
                transceiver_info = TransceiverInfo.from_dict(info)
                search_res = cls._TRANSCEIVER_PATTERN_RE.search(interface)
                if search_res:
                    transceiver_info.interface = search_res.group(1).strip()
                    transceiver_info.optical_id = search_res.group(2)
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
            "core": "Core",
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
        cls, switch_log_info: List[FindResult], port_mapping: Dict[str, str]
    ) -> List[SwiOpticalModel]:
        convert_dict = {
            "txpower": "TxPower Lane",
            "rxpower": "RxPower Lane",
            "bias": "Bias Lane",
            "SNR0": "HostSNR Lane",
            "SNR1": "MediaSNR Lane",
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
            "interface": "Interface",
            "phy": "PHY",
            "protocol": "Protocol",
            "in_uti": "InUti",
            "out_uti": "OutUti",
            "in_errors": "inErrors",
            "out_errors": "outErrors",
        }
        parse_data_list = TableParser.parse(cmd_res, titles_dict)
        interface_info_list = []
        for data in parse_data_list:
            interface_name = data.get(INTERFACE_STR)
            if not interface_name:
                continue
            # 处理端口回显eg："800GE1/0/128:4(200GE)"
            if '(' in interface_name:
                data[INTERFACE_STR] = interface_name.split('(')[0].strip()
            interface_instance = InterfaceBrief.from_dict(data)
            interface_info_list.append(interface_instance)
        return interface_info_list

    @classmethod
    def parse_opt_module_info_from_table(cls, cmd_res: str) -> List[SwiOpticalModel]:
        cmd_res_list = split_str(cmd_res, "dis optical-module interface")
        optical_model_list = []
        for cmd_res_str in cmd_res_list:
            block_list = split_str(cmd_res_str, "diagnostic information")
            for block_text in block_list:
                optical_id = cls.extract_optical_module_id(block_text)
                first_line = block_text.splitlines()[0].strip()
                interface = first_line.split()[0]
                table_list = split_str(block_text, "==============")
                base_info, diag_flag_info = [], []
                for table_str in table_list:
                    if "Digital Diagnostic Monitoring" in table_str:
                        base_info = cls.trans_opt_module_results(table_str)
                    if "State-flag Diagnostic Monitoring" in table_str:
                        diag_flag_info = cls.parse_op_state_flag_diag_info(table_str)
                if not base_info and not diag_flag_info:
                    continue
                optical_model = SwiOpticalModel(interface, optical_id, base_info, diag_flag_info)
                optical_model_list.append(optical_model)
        return optical_model_list

    @staticmethod
    def extract_optical_module_id(line: str) -> str:
        """提取 optical-module(42) 括号内数字。"""
        key = "optical-module("
        idx1 = line.find(key)
        if idx1 < 0:
            return ""
        idx2 = line.find(")", idx1)
        if idx2 < 0:
            return ""
        num_str = line[idx1 + len(key) : idx2]
        return num_str if num_str.isdigit() else ""

    # QoS credit 表 titles_dict：key 顺序即列作用域顺序（port | vl0...vlN | VNA | total）
    _QOS_CREDIT_TITLES_DICT = {"port": "port", "vl_credits": "vl0", "vna": "VNA", "total": "total"}

    @classmethod
    def parse_qos_credit(cls, cmd_res: str) -> List[QosCreditInfo]:
        """解析 NPU 槽位 switch 芯片 QoS credit 回显（批量命令输出按命令行拆分为各芯片段）。

        每个芯片段包含三张表（列结构一致，列名：port vl0...vl15 VNA total）：
            QOS ALLOC CREDIT TABLE:    分配的 Credit
            QOS USED CREDIT TABLE:     已使用的 Credit
            QOS CURRENT CREDIT TABLE:  当前可用的 Credit
        满足 alloc credit = used credit + current credit。
        """
        results: List[QosCreditInfo] = []
        if not cmd_res:
            return results
        for slot_id, chip_id, segment in cls._split_qos_credit_segments(cmd_res):
            alloc_rows, used_rows, current_rows = cls._split_qos_credit_tables(segment)
            if not alloc_rows:
                continue
            ports = []
            for port_id, alloc in alloc_rows.items():
                used = used_rows.get(port_id, {})
                current = current_rows.get(port_id, {})
                ports.append(
                    QosCreditPortInfo(
                        port_id=port_id,
                        vl_alloc_credits=alloc.get("vl_credits", []),
                        vl_used_credits=used.get("vl_credits", []),
                        vl_current_credits=current.get("vl_credits", []),
                        vna_alloc=alloc.get("vna", ""),
                        vna_used=used.get("vna", ""),
                        vna_current=current.get("vna", ""),
                        total_alloc=alloc.get("total", ""),
                        total_used=used.get("total", ""),
                        total_current=current.get("total", ""),
                    )
                )
            if ports:
                results.append(QosCreditInfo(slot_id=slot_id, chip_id=chip_id, ports=ports))
        return results

    @classmethod
    def _split_qos_credit_segments(cls, cmd_res: str) -> List[Tuple[str, str, str]]:
        """按命令行拆分回显为各芯片段：返回 [(slot_id, chip_id, 段内容)]，无命令行时返回空列表。"""
        segments = []
        SLOT_ID_INDEX, CHIP_ID_INDEX, SPLIT_LEN = 5, 7, 8
        for block in split_str(cmd_res, "dis forward information enp slot "):
            lines = block.splitlines()
            parts = lines[0].strip().split()
            if len(parts) > SPLIT_LEN:
                slot_id, chip_id = parts[SLOT_ID_INDEX], parts[CHIP_ID_INDEX]
                segments.append((slot_id, chip_id, "\n".join(lines[1:])))
        return segments

    @classmethod
    def _split_qos_credit_tables(
        cls, segment: str
    ) -> Tuple[Dict[str, QosCreditRow], Dict[str, QosCreditRow], Dict[str, QosCreditRow]]:
        """按表标题后缀切分三张表并逐表解析：返回 (alloc_rows, used_rows, current_rows)。

        按行内特征（三张表标题共有后缀 " CREDIT TABLE:"）切分，
        每块以标题行开头，再按标题前缀归属到对应表；返回值为 {port_id: {"vl_credits", "vna", "total"}}。
        """
        alloc_rows, used_rows, current_rows = {}, {}, {}
        for block in split_str(segment, " CREDIT TABLE:"):
            if block.startswith("QOS ALLOC CREDIT TABLE:"):
                alloc_rows = cls._parse_qos_credit_rows(block)
            elif block.startswith("QOS USED CREDIT TABLE:"):
                used_rows = cls._parse_qos_credit_rows(block)
            elif block.startswith("QOS CURRENT CREDIT TABLE:"):
                current_rows = cls._parse_qos_credit_rows(block)
        return alloc_rows, used_rows, current_rows

    @classmethod
    def _parse_qos_credit_rows(cls, block: str) -> Dict[str, Dict[str, object]]:
        """解析单张 QoS credit 表块（以标题行开头）：返回 {port_id: {"vl_credits": [...], "vna": str, "total": str}}。

        处理动态 vl 列：以首列 vl0 定位 vl 块起始、VNA 定位终点，
        TableParser 将 vl0 至 VNA 之间作为一个整体字符串取到 vl_credits 字段，再按空白拆分。
        """
        rows: Dict[str, Dict[str, object]] = {}
        for item in TableParser.parse(block, cls._QOS_CREDIT_TITLES_DICT):
            port_id = item.get("port", "")
            vl_credits = item.get("vl_credits", "").split()
            vna = item.get("vna", "")
            total = item.get("total", "")
            # 过滤非数据行（端口非数字）与列值缺失行（如超长异常行、列数不齐行）
            if not port_id.isdigit() or not vl_credits or not vna or not total:
                continue
            rows[port_id] = {"vl_credits": vl_credits, "vna": vna, "total": total}
        return rows
