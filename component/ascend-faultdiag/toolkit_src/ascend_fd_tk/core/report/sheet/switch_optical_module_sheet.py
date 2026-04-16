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

"""
交换机间端口连接光模块信息报告Sheet生成器
"""

from dataclasses import dataclass
from typing import List, Dict, Tuple

from ascend_fd_tk.core.report.sheet.base import BaseSheetGenerator
from ascend_fd_tk.core.report.threshold_report import ThresholdConfig, create_threshold_report, generate_threshold_excel


@dataclass
class SwitchOpticalModuleData:
    """交换机光模块数据类，用于存储交换机间端口连接光模块信息"""
    # 本端交换机信息
    local_switch_name: str
    local_switch_id: str
    local_switch_sn: str
    local_interface: str

    # 本端光模块信息
    local_optical_vendor: str = ""
    local_optical_model: str = ""
    local_optical_sn: str = ""
    local_optical_temp: str = ""

    # 本端Lane信息
    local_tx_power0: str = ""
    local_rx_power0: str = ""
    local_snr_lane0: str = ""
    local_tx_power1: str = ""
    local_rx_power1: str = ""
    local_snr_lane1: str = ""
    local_tx_power2: str = ""
    local_rx_power2: str = ""
    local_snr_lane2: str = ""
    local_tx_power3: str = ""
    local_rx_power3: str = ""
    local_snr_lane3: str = ""

    # 对端交换机信息
    peer_switch_name: str = ""
    peer_switch_id: str = ""
    peer_switch_sn: str = ""
    peer_interface: str = ""

    # 对端光模块信息
    peer_optical_vendor: str = ""
    peer_optical_model: str = ""
    peer_optical_sn: str = ""
    peer_optical_temp: str = ""

    # 对端Lane信息
    peer_tx_power0: str = ""
    peer_rx_power0: str = ""
    peer_snr_lane0: str = ""
    peer_tx_power1: str = ""
    peer_rx_power1: str = ""
    peer_snr_lane1: str = ""
    peer_tx_power2: str = ""
    peer_rx_power2: str = ""
    peer_snr_lane2: str = ""
    peer_tx_power3: str = ""
    peer_rx_power3: str = ""
    peer_snr_lane3: str = ""


class SwitchOpticalModuleSheetGenerator(BaseSheetGenerator):
    """交换机间端口连接光模块信息报告Sheet生成器"""

    @staticmethod
    def _get_optical_module_data(interface_full_info) -> Dict[str, str]:
        """
        从InterfaceFullInfo获取光模块数据

        :param interface_full_info: 接口完整信息对象
        :return: 光模块数据字典
        """
        data = {}

        # 获取光模块基本信息
        if hasattr(interface_full_info, 'transceiver_info') and interface_full_info.transceiver_info:
            transceiver_info = interface_full_info.transceiver_info
            if hasattr(transceiver_info, 'manufacture_information'):
                manu_info = transceiver_info.manufacture_information
                data["optical_vendor"] = getattr(manu_info, 'manu_name', "")
                data["optical_model"] = getattr(manu_info, 'manu_part_num', "")
                data["optical_sn"] = getattr(manu_info, 'manu_serial_number', "")

        # 获取光模块温度
        if hasattr(interface_full_info, 'swi_optical_model') and interface_full_info.swi_optical_model:
            swi_optical_model = interface_full_info.swi_optical_model
            data["optical_temp"] = getattr(swi_optical_model, 'temperature', "")

        # 获取光模块Lane信息
        optical_module_info = interface_full_info.get_optical_module_info()
        if optical_module_info:
            lane_infos = optical_module_info.lane_power_infos or []
            for i in range(8):
                if i < len(lane_infos):
                    lane = lane_infos[i]
                    data[f"tx_power{i}"] = lane.tx_power or ""
                    data[f"rx_power{i}"] = lane.rx_power or ""
                    data[f"snr_lane{i}"] = lane.media_snr or lane.host_snr or ""

        return data

    @staticmethod
    def _create_header_config() -> Tuple[Dict[str, str], List[str]]:
        """
        创建header映射和顺序

        :return: (header_mapping, header_order)
        """
        header_mapping = {
            # 本端交换机信息
            "local_switch_name": "本端交换机名称",
            "local_switch_id": "本端交换机ID",
            "local_switch_sn": "本端交换机SN",
            "local_interface": "本端接口",

            # 本端光模块信息
            "local_optical_vendor": "本端光模块厂商",
            "local_optical_model": "本端光模块型号",
            "local_optical_sn": "本端光模块SN",
            "local_optical_temp": "本端光模块温度",

            # 本端Lane信息
            "local_tx_power0": "本端TX Power Lane 0",
            "local_rx_power0": "本端RX Power Lane 0",
            "local_snr_lane0": "本端SNR Lane 0",
            "local_tx_power1": "本端TX Power Lane 1",
            "local_rx_power1": "本端RX Power Lane 1",
            "local_snr_lane1": "本端SNR Lane 1",
            "local_tx_power2": "本端TX Power Lane 2",
            "local_rx_power2": "本端RX Power Lane 2",
            "local_snr_lane2": "本端SNR Lane 2",
            "local_tx_power3": "本端TX Power Lane 3",
            "local_rx_power3": "本端RX Power Lane 3",
            "local_snr_lane3": "本端SNR Lane 3",

            # 对端交换机信息
            "peer_switch_name": "对端交换机名称",
            "peer_switch_id": "对端交换机ID",
            "peer_switch_sn": "对端交换机SN",
            "peer_interface": "对端接口",

            # 对端光模块信息
            "peer_optical_vendor": "对端光模块厂商",
            "peer_optical_model": "对端光模块型号",
            "peer_optical_sn": "对端光模块SN",
            "peer_optical_temp": "对端光模块温度",

            # 对端Lane信息
            "peer_tx_power0": "对端TX Power Lane 0",
            "peer_rx_power0": "对端RX Power Lane 0",
            "peer_snr_lane0": "对端SNR Lane 0",
            "peer_tx_power1": "对端TX Power Lane 1",
            "peer_rx_power1": "对端RX Power Lane 1",
            "peer_snr_lane1": "对端SNR Lane 1",
            "peer_tx_power2": "对端TX Power Lane 2",
            "peer_rx_power2": "对端RX Power Lane 2",
            "peer_snr_lane2": "对端SNR Lane 2",
            "peer_tx_power3": "对端TX Power Lane 3",
            "peer_rx_power3": "对端RX Power Lane 3",
            "peer_snr_lane3": "对端SNR Lane 3"
        }

        # 定义header顺序，将相关信息分组显示
        header_order = [
            # 本端交换机和接口信息
            "本端交换机名称", "本端交换机ID", "本端交换机SN", "本端接口",

            # 本端光模块基本信息
            "本端光模块厂商", "本端光模块型号", "本端光模块SN", "本端光模块温度",

            # 本端光模块Lane信息
            "本端TX Power Lane 0", "本端RX Power Lane 0", "本端SNR Lane 0",
            "本端TX Power Lane 1", "本端RX Power Lane 1", "本端SNR Lane 1",
            "本端TX Power Lane 2", "本端RX Power Lane 2", "本端SNR Lane 2",
            "本端TX Power Lane 3", "本端RX Power Lane 3", "本端SNR Lane 3",

            # 对端交换机和接口信息
            "对端交换机名称", "对端交换机ID", "对端交换机SN", "对端接口",

            # 对端光模块基本信息
            "对端光模块厂商", "对端光模块型号", "对端光模块SN", "对端光模块温度",

            # 对端光模块Lane信息
            "对端TX Power Lane 0", "对端RX Power Lane 0", "对端SNR Lane 0",
            "对端TX Power Lane 1", "对端RX Power Lane 1", "对端SNR Lane 1",
            "对端TX Power Lane 2", "对端RX Power Lane 2", "对端SNR Lane 2",
            "对端TX Power Lane 3", "对端RX Power Lane 3", "对端SNR Lane 3"
        ]

        return header_mapping, header_order

    def generate_sheet(self) -> None:
        """
        生成交换机间端口连接光模块信息Excel Sheet
        """
        # 收集光模块数据
        optical_module_data_list = self._collect_optical_module_data()

        # 如果没有数据，跳过生成Sheet
        if not optical_module_data_list:
            return

        # 创建阈值配置
        threshold_configs = self._create_threshold_configs()

        # 创建header映射和顺序
        header_mapping, header_order = self._create_header_config()

        # 创建报告Sheet
        sheet = create_threshold_report(
            sheet_name="交换机间端口连接光模块信息",
            data_list=optical_module_data_list,
            header_mapping=header_mapping,
            header_order=header_order,
            threshold_configs=threshold_configs,
            na_rep="-"
        )

        # 生成Excel
        generate_threshold_excel(
            excel_gen=self.excel_gen,
            sheets=[sheet]
        )

    def _collect_optical_module_data(self) -> List[SwitchOpticalModuleData]:
        """
        收集交换机间端口连接光模块数据

        :return: 光模块数据列表
        """
        data_list = []

        # 用于跟踪已处理的连接，避免重复
        processed_connections = set()

        # 遍历所有交换机
        for switch_info in self.cluster_info.swis_info.values():
            # 遍历交换机的接口映射
            for interface_mapping in switch_info.interface_mapping:
                if not interface_mapping.remote_device_interface:
                    continue

                local_interface = interface_mapping.local_interface_name
                peer_switch_name = interface_mapping.remote_device_interface.device_name
                peer_interface = interface_mapping.remote_device_interface.interface

                # 创建连接唯一标识，避免重复处理
                connection_key1 = f"{switch_info.name}@{local_interface}->{peer_switch_name}@{peer_interface}"
                connection_key2 = f"{peer_switch_name}@{peer_interface}->{switch_info.name}@{local_interface}"

                if connection_key1 in processed_connections or connection_key2 in processed_connections:
                    continue

                # 标记此连接为已处理
                processed_connections.add(connection_key1)
                processed_connections.add(connection_key2)

                # 获取本端交换机的接口信息
                local_interface_full_info = switch_info.interface_full_infos.get(local_interface)
                if not local_interface_full_info:
                    continue

                # 获取对端交换机信息
                peer_switch = self.cluster_info.swis_info.get(peer_switch_name)
                if not peer_switch:
                    # 尝试通过交换机名称查找
                    for swi_info in self.cluster_info.swis_info.values():
                        if swi_info.name == peer_switch_name:
                            peer_switch = swi_info
                            break

                # 收集本端光模块信息
                local_optical_data = self._get_optical_module_data(local_interface_full_info)

                # 收集对端光模块信息（即使没有找到对端交换机或接口，也要继续展示本端信息）
                peer_optical_data = {}
                if peer_switch:
                    # 获取对端接口信息
                    peer_interface_full_info = peer_switch.interface_full_infos.get(peer_interface)
                    if peer_interface_full_info:
                        peer_optical_data = self._get_optical_module_data(peer_interface_full_info)

                # 创建交换机光模块数据对象
                data = SwitchOpticalModuleData(
                    # 本端交换机信息
                    local_switch_name=switch_info.name,
                    local_switch_id=switch_info.swi_id,
                    local_switch_sn=switch_info.sn,
                    local_interface=local_interface,

                    # 本端光模块信息
                    local_optical_vendor=local_optical_data.get("optical_vendor", ""),
                    local_optical_model=local_optical_data.get("optical_model", ""),
                    local_optical_sn=local_optical_data.get("optical_sn", ""),
                    local_optical_temp=local_optical_data.get("optical_temp", ""),

                    # 本端Lane信息
                    local_tx_power0=local_optical_data.get("tx_power0", ""),
                    local_rx_power0=local_optical_data.get("rx_power0", ""),
                    local_snr_lane0=local_optical_data.get("snr_lane0", ""),
                    local_tx_power1=local_optical_data.get("tx_power1", ""),
                    local_rx_power1=local_optical_data.get("rx_power1", ""),
                    local_snr_lane1=local_optical_data.get("snr_lane1", ""),
                    local_tx_power2=local_optical_data.get("tx_power2", ""),
                    local_rx_power2=local_optical_data.get("rx_power2", ""),
                    local_snr_lane2=local_optical_data.get("snr_lane2", ""),
                    local_tx_power3=local_optical_data.get("tx_power3", ""),
                    local_rx_power3=local_optical_data.get("rx_power3", ""),
                    local_snr_lane3=local_optical_data.get("snr_lane3", ""),

                    # 对端交换机信息
                    peer_switch_name=peer_switch.name if peer_switch else peer_switch_name,
                    peer_switch_id=peer_switch.swi_id if peer_switch else "",
                    peer_switch_sn=peer_switch.sn if peer_switch else "",
                    peer_interface=peer_interface,

                    # 对端光模块信息
                    peer_optical_vendor=peer_optical_data.get("optical_vendor", ""),
                    peer_optical_model=peer_optical_data.get("optical_model", ""),
                    peer_optical_sn=peer_optical_data.get("optical_sn", ""),
                    peer_optical_temp=peer_optical_data.get("optical_temp", ""),

                    # 对端Lane信息
                    peer_tx_power0=peer_optical_data.get("tx_power0", ""),
                    peer_rx_power0=peer_optical_data.get("rx_power0", ""),
                    peer_snr_lane0=peer_optical_data.get("snr_lane0", ""),
                    peer_tx_power1=peer_optical_data.get("tx_power1", ""),
                    peer_rx_power1=peer_optical_data.get("rx_power1", ""),
                    peer_snr_lane1=peer_optical_data.get("snr_lane1", ""),
                    peer_tx_power2=peer_optical_data.get("tx_power2", ""),
                    peer_rx_power2=peer_optical_data.get("rx_power2", ""),
                    peer_snr_lane2=peer_optical_data.get("snr_lane2", ""),
                    peer_tx_power3=peer_optical_data.get("tx_power3", ""),
                    peer_rx_power3=peer_optical_data.get("rx_power3", ""),
                    peer_snr_lane3=peer_optical_data.get("snr_lane3", "")
                )

                data_list.append(data)

        return data_list

    def _create_threshold_configs(self) -> List[ThresholdConfig]:
        """
        创建阈值配置

        :return: 阈值配置列表
        """
        threshold_cls = self.cluster_info.get_threshold()

        return [
            # 本端TX Power阈值（dBm）
            ThresholdConfig(
                field_name="local_tx_power0",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="本端TX Power Lane 0"
            ),
            ThresholdConfig(
                field_name="local_tx_power1",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="本端TX Power Lane 1"
            ),
            ThresholdConfig(
                field_name="local_tx_power2",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="本端TX Power Lane 2"
            ),
            ThresholdConfig(
                field_name="local_tx_power3",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="本端TX Power Lane 3"
            ),

            # 本端RX Power阈值（dBm）
            ThresholdConfig(
                field_name="local_rx_power0",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="本端RX Power Lane 0"
            ),
            ThresholdConfig(
                field_name="local_rx_power1",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="本端RX Power Lane 1"
            ),
            ThresholdConfig(
                field_name="local_rx_power2",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="本端RX Power Lane 2"
            ),
            ThresholdConfig(
                field_name="local_rx_power3",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="本端RX Power Lane 3"
            ),

            # 本端SNR阈值（dB）
            ThresholdConfig(
                field_name="local_snr_lane0",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="本端SNR Lane 0"
            ),
            ThresholdConfig(
                field_name="local_snr_lane1",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="本端SNR Lane 1"
            ),
            ThresholdConfig(
                field_name="local_snr_lane2",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="本端SNR Lane 2"
            ),
            ThresholdConfig(
                field_name="local_snr_lane3",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="本端SNR Lane 3"
            ),

            # 对端TX Power阈值（dBm）
            ThresholdConfig(
                field_name="peer_tx_power0",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端TX Power Lane 0"
            ),
            ThresholdConfig(
                field_name="peer_tx_power1",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端TX Power Lane 1"
            ),
            ThresholdConfig(
                field_name="peer_tx_power2",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端TX Power Lane 2"
            ),
            ThresholdConfig(
                field_name="peer_tx_power3",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端TX Power Lane 3"
            ),

            # 对端RX Power阈值（dBm）
            ThresholdConfig(
                field_name="peer_rx_power0",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端RX Power Lane 0"
            ),
            ThresholdConfig(
                field_name="peer_rx_power1",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端RX Power Lane 1"
            ),
            ThresholdConfig(
                field_name="peer_rx_power2",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端RX Power Lane 2"
            ),
            ThresholdConfig(
                field_name="peer_rx_power3",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端RX Power Lane 3"
            ),

            # 对端SNR阈值（dB）
            ThresholdConfig(
                field_name="peer_snr_lane0",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="对端SNR Lane 0"
            ),
            ThresholdConfig(
                field_name="peer_snr_lane1",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="对端SNR Lane 1"
            ),
            ThresholdConfig(
                field_name="peer_snr_lane2",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="对端SNR Lane 2"
            ),
            ThresholdConfig(
                field_name="peer_snr_lane3",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="对端SNR Lane 3"
            )
        ]
