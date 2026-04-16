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
主机NPU到交换机端口光模块信息报告Sheet生成器
用于生成主机NPU与交换机之间连接的光模块信息报告
"""

from dataclasses import dataclass
from typing import List, Dict, Tuple

from ascend_fd_tk.core.report.sheet.base import BaseSheetGenerator
from ascend_fd_tk.core.report.threshold_report import ThresholdConfig, create_threshold_report, generate_threshold_excel


@dataclass
class HostToSwitchOpticalModuleData:
    """主机到交换机光模块数据类，用于存储主机NPU与交换机之间连接的光模块信息"""
    # 主机信息
    host_id: str  # 主机ID
    host_name: str  # 主机名称
    host_sn: str  # 主机序列号

    # NPU芯片信息
    npu_id: str  # NPU ID
    chip_id: str  # 芯片ID
    chip_phy_id: str  # 芯片物理ID
    npu_type: str  # NPU类型

    # 主机侧光模块基本信息
    host_optical_present: str  # 主机侧光模块状态（是否存在）
    host_optical_vendor: str  # 主机侧光模块厂商名称
    host_optical_model: str  # 主机侧光模块厂商部件号
    host_optical_sn: str  # 主机侧光模块厂商序列号

    # 主机侧光模块性能指标
    host_optical_temp: str  # 主机侧光模块温度
    host_optical_vcc: str  # 主机侧光模块供电电压
    host_link_speed: str  # 主机侧链路速度
    host_link_duplex: str  # 主机侧链路双工模式
    host_net_health: str  # 主机侧网络健康状态
    host_link_status: str  # 主机侧链路状态

    # 主机侧链路统计信息
    host_link_up_count: str  # 主机侧链路UP次数
    host_link_down_count: str  # 主机侧链路DOWN次数

    # 主机侧Lane 0 信息
    host_tx_power0: str  # 主机侧发射功率 Lane 0
    host_rx_power0: str  # 主机侧接收功率 Lane 0
    host_tx_bias0: str  # 主机侧发射偏置电流 Lane 0
    host_snr_lane0: str  # 主机侧SNR Lane 0
    host_media_snr_lane0: str  # 主机侧Media SNR Lane 0

    # 主机侧Lane 1 信息
    host_tx_power1: str  # 主机侧发射功率 Lane 1
    host_rx_power1: str  # 主机侧接收功率 Lane 1
    host_tx_bias1: str  # 主机侧发射偏置电流 Lane 1
    host_snr_lane1: str  # 主机侧SNR Lane 1
    host_media_snr_lane1: str  # 主机侧Media SNR Lane 1

    # 主机侧Lane 2 信息
    host_tx_power2: str  # 主机侧发射功率 Lane 2
    host_rx_power2: str  # 主机侧接收功率 Lane 2
    host_tx_bias2: str  # 主机侧发射偏置电流 Lane 2
    host_snr_lane2: str  # 主机侧SNR Lane 2
    host_media_snr_lane2: str  # 主机侧Media SNR Lane 2

    # 主机侧Lane 3 信息
    host_tx_power3: str  # 主机侧发射功率 Lane 3
    host_rx_power3: str  # 主机侧接收功率 Lane 3
    host_tx_bias3: str  # 主机侧发射偏置电流 Lane 3
    host_snr_lane3: str  # 主机侧SNR Lane 3
    host_media_snr_lane3: str  # 主机侧Media SNR Lane 3

    # 主机侧端口信息（有默认值的字段开始）
    host_port: str = ""  # 主机侧端口

    # 对端交换机信息
    peer_switch_name: str = ""  # 对端交换机名称
    peer_switch_id: str = ""  # 对端交换机ID
    peer_switch_sn: str = ""  # 对端交换机序列号

    # 对端交换机端口信息
    peer_switch_port: str = ""  # 对端交换机端口
    peer_switch_port_speed: str = ""  # 对端交换机端口速度
    peer_switch_port_duplex: str = ""  # 对端交换机端口双工模式
    peer_switch_port_status: str = ""  # 对端交换机端口状态

    # 对端交换机光模块信息
    peer_switch_optical_vendor: str = ""  # 对端交换机光模块厂商
    peer_switch_optical_model: str = ""  # 对端交换机光模块型号
    peer_switch_optical_sn: str = ""  # 对端交换机光模块SN
    peer_switch_optical_temp: str = ""  # 对端交换机光模块温度

    # 对端交换机光模块Lane信息
    peer_switch_tx_power0: str = ""  # 对端交换机发射功率 Lane 0
    peer_switch_rx_power0: str = ""  # 对端交换机接收功率 Lane 0
    peer_switch_tx_power1: str = ""  # 对端交换机发射功率 Lane 1
    peer_switch_rx_power1: str = ""  # 对端交换机接收功率 Lane 1
    peer_switch_tx_power2: str = ""  # 对端交换机发射功率 Lane 2
    peer_switch_rx_power2: str = ""  # 对端交换机接收功率 Lane 2
    peer_switch_tx_power3: str = ""  # 对端交换机发射功率 Lane 3
    peer_switch_rx_power3: str = ""  # 对端交换机接收功率 Lane 3
    peer_switch_snr_lane0: str = ""  # 对端交换机SNR Lane 0
    peer_switch_snr_lane1: str = ""  # 对端交换机SNR Lane 1
    peer_switch_snr_lane2: str = ""  # 对端交换机SNR Lane 2
    peer_switch_snr_lane3: str = ""  # 对端交换机SNR Lane 3


class HostToSwitchOpticalModuleSheetGenerator(BaseSheetGenerator):
    """主机到交换机光模块信息报告Sheet生成器"""
    LANE_NUM = 4

    @staticmethod
    def _create_header_config() -> Tuple[Dict[str, str], List[str]]:
        """
        创建header映射和顺序

        :return: (header_mapping, header_order)
        """
        header_mapping = {
            # 主机信息
            "host_id": "主机ID",
            "host_name": "主机名",
            "host_sn": "主机SN",

            # NPU芯片信息
            "npu_id": "NPU ID",
            "chip_id": "芯片ID",
            "chip_phy_id": "物理芯片ID",
            "npu_type": "NPU类型",
            "host_port": "主机侧端口",

            # 主机侧光模块基本信息
            "host_optical_present": "主机侧光模块状态",
            "host_optical_vendor": "主机侧光模块厂商",
            "host_optical_model": "主机侧光模块型号",
            "host_optical_sn": "主机侧光模块SN",

            # 主机侧光模块性能指标
            "host_optical_temp": "主机侧光模块温度",
            "host_optical_vcc": "主机侧光模块供电电压",
            "host_link_speed": "主机侧链路速度",
            "host_link_duplex": "主机侧双工模式",
            "host_net_health": "主机侧网络健康",
            "host_link_status": "主机侧链路状态",

            # 主机侧链路统计信息
            "host_link_up_count": "主机侧链路UP次数",
            "host_link_down_count": "主机侧链路DOWN次数",

            # 主机侧Lane 0 信息
            "host_tx_power0": "主机侧TX Power Lane 0",
            "host_rx_power0": "主机侧RX Power Lane 0",
            "host_tx_bias0": "主机侧TX Bias Lane 0",
            "host_snr_lane0": "主机侧Host SNR Lane 0",
            "host_media_snr_lane0": "主机侧Media SNR Lane 0",

            # 主机侧Lane 1 信息
            "host_tx_power1": "主机侧TX Power Lane 1",
            "host_rx_power1": "主机侧RX Power Lane 1",
            "host_tx_bias1": "主机侧TX Bias Lane 1",
            "host_snr_lane1": "主机侧Host SNR Lane 1",
            "host_media_snr_lane1": "主机侧Media SNR Lane 1",

            # 主机侧Lane 2 信息
            "host_tx_power2": "主机侧TX Power Lane 2",
            "host_rx_power2": "主机侧RX Power Lane 2",
            "host_tx_bias2": "主机侧TX Bias Lane 2",
            "host_snr_lane2": "主机侧Host SNR Lane 2",
            "host_media_snr_lane2": "主机侧Media SNR Lane 2",

            # 主机侧Lane 3 信息
            "host_tx_power3": "主机侧TX Power Lane 3",
            "host_rx_power3": "主机侧RX Power Lane 3",
            "host_tx_bias3": "主机侧TX Bias Lane 3",
            "host_snr_lane3": "主机侧Host SNR Lane 3",
            "host_media_snr_lane3": "主机侧Media SNR Lane 3",

            # 对端交换机信息
            "peer_switch_name": "对端交换机名称",
            "peer_switch_id": "对端交换机ID",
            "peer_switch_sn": "对端交换机SN",

            # 对端交换机端口信息
            "peer_switch_port": "对端交换机端口",
            "peer_switch_port_speed": "对端交换机端口速度",
            "peer_switch_port_duplex": "对端交换机端口双工模式",
            "peer_switch_port_status": "对端交换机端口状态",

            # 对端交换机光模块信息
            "peer_switch_optical_vendor": "对端交换机光模块厂商",
            "peer_switch_optical_model": "对端交换机光模块型号",
            "peer_switch_optical_sn": "对端交换机光模块SN",
            "peer_switch_optical_temp": "对端交换机光模块温度",

            # 对端交换机光模块Lane信息
            "peer_switch_tx_power0": "对端交换机TX Power Lane 0",
            "peer_switch_rx_power0": "对端交换机RX Power Lane 0",
            "peer_switch_snr_lane0": "对端交换机SNR Lane 0",
            "peer_switch_tx_power1": "对端交换机TX Power Lane 1",
            "peer_switch_rx_power1": "对端交换机RX Power Lane 1",
            "peer_switch_snr_lane1": "对端交换机SNR Lane 1",
            "peer_switch_tx_power2": "对端交换机TX Power Lane 2",
            "peer_switch_rx_power2": "对端交换机RX Power Lane 2",
            "peer_switch_snr_lane2": "对端交换机SNR Lane 2",
            "peer_switch_tx_power3": "对端交换机TX Power Lane 3",
            "peer_switch_rx_power3": "对端交换机RX Power Lane 3",
            "peer_switch_snr_lane3": "对端交换机SNR Lane 3"
        }

        # 定义header顺序，将相关信息分组显示
        header_order = [
            # 主机和NPU信息
            "主机ID", "主机名", "主机SN",
            "NPU ID", "芯片ID", "物理芯片ID", "NPU类型", "主机侧端口",

            # 主机侧光模块基本信息
            "主机侧光模块状态", "主机侧光模块厂商", "主机侧光模块型号", "主机侧光模块SN",

            # 主机侧光模块性能指标
            "主机侧光模块温度", "主机侧光模块供电电压", "主机侧链路速度", "主机侧双工模式", "主机侧网络健康",
            "主机侧链路状态",
            "主机侧链路UP次数", "主机侧链路DOWN次数",

            # 主机侧光模块Lane信息
            "主机侧TX Power Lane 0", "主机侧RX Power Lane 0", "主机侧TX Bias Lane 0", "主机侧Host SNR Lane 0",
            "主机侧Media SNR Lane 0",
            "主机侧TX Power Lane 1", "主机侧RX Power Lane 1", "主机侧TX Bias Lane 1", "主机侧Host SNR Lane 1",
            "主机侧Media SNR Lane 1",
            "主机侧TX Power Lane 2", "主机侧RX Power Lane 2", "主机侧TX Bias Lane 2", "主机侧Host SNR Lane 2",
            "主机侧Media SNR Lane 2",
            "主机侧TX Power Lane 3", "主机侧RX Power Lane 3", "主机侧TX Bias Lane 3", "主机侧Host SNR Lane 3",
            "主机侧Media SNR Lane 3",

            # 对端交换机信息
            "对端交换机名称", "对端交换机ID", "对端交换机SN",

            # 对端交换机端口信息
            "对端交换机端口", "对端交换机端口速度", "对端交换机端口双工模式", "对端交换机端口状态",

            # 对端交换机光模块信息
            "对端交换机光模块厂商", "对端交换机光模块型号", "对端交换机光模块SN", "对端交换机光模块温度",

            # 对端交换机光模块Lane信息
            "对端交换机TX Power Lane 0", "对端交换机RX Power Lane 0", "对端交换机SNR Lane 0",
            "对端交换机TX Power Lane 1", "对端交换机RX Power Lane 1", "对端交换机SNR Lane 1",
            "对端交换机TX Power Lane 2", "对端交换机RX Power Lane 2", "对端交换机SNR Lane 2",
            "对端交换机TX Power Lane 3", "对端交换机RX Power Lane 3", "对端交换机SNR Lane 3"
        ]

        return header_mapping, header_order

    def generate_sheet(self) -> None:
        """
        生成主机NPU与交换机端口信息Excel Sheet
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
            sheet_name="主机NPU<->交换机端口光模块信息",
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

    def _collect_optical_module_data(self) -> List[HostToSwitchOpticalModuleData]:
        """
        收集光模块数据，包括主机NPU光模块和对应的交换机端口信息

        :return: 光模块数据列表
        """
        data_list = []

        # 遍历所有主机
        for host_id, host_info in self.cluster_info.hosts_info.items():
            if not host_info.npu_chip_info:
                continue

            # 遍历所有NPU芯片
            for chip_id, npu_chip_info in host_info.npu_chip_info.items():
                # 使用get_optical_module_info()获取光模块信息
                optical_module_info = npu_chip_info.get_optical_module_info()
                if not optical_module_info:
                    continue

                # 获取原始光模块信息（用于获取基本信息）
                optical_info = getattr(npu_chip_info, 'hccn_optical_info', type('obj', (object,), {}))
                link_stat_info = getattr(npu_chip_info, 'hccn_link_stat_info', type('obj', (object,), {}))
                lldp_info = getattr(npu_chip_info, 'hccn_lldp_info', None)

                # 获取lane信息，确保有4个lane的信息
                lane_infos = optical_module_info.lane_power_infos or []
                lane_data = {}
                for i in range(self.LANE_NUM):
                    if i < len(lane_infos):
                        lane = lane_infos[i]
                        lane_data[f"tx_power{i}"] = lane.tx_power or ""
                        lane_data[f"rx_power{i}"] = lane.rx_power or ""
                        lane_data[f"tx_bias{i}"] = lane.bias or ""
                        lane_data[f"host_snr_lane{i}"] = lane.host_snr or ""
                        lane_data[f"media_snr_lane{i}"] = lane.media_snr or ""
                    else:
                        # 如果lane信息不足，使用空字符串填充
                        lane_data[f"tx_power{i}"] = ""
                        lane_data[f"rx_power{i}"] = ""
                        lane_data[f"tx_bias{i}"] = ""
                        lane_data[f"host_snr_lane{i}"] = ""
                        lane_data[f"media_snr_lane{i}"] = ""

                # 获取对端交换机和端口信息
                peer_switch_name = ""
                peer_port = ""
                peer_switch = None
                peer_interface_info = None

                if lldp_info:
                    peer_switch_name = getattr(lldp_info, 'system_name_tlv', "")
                    peer_port = getattr(lldp_info, 'port_id_tlv', "")

                    # 查找对端交换机和端口信息
                    if peer_switch_name and peer_port:
                        peer_switch, peer_interface_info = self.cluster_info.find_peer_swi_interface_info(
                            peer_switch_name, peer_port
                        )

                # 收集对端交换机信息
                peer_switch_id = ""
                peer_switch_sn = ""
                if peer_switch:
                    peer_switch_id = getattr(peer_switch, 'swi_id', "")
                    peer_switch_sn = getattr(peer_switch, 'sn', "")

                # 收集对端端口信息
                peer_port_speed = ""
                peer_port_duplex = ""
                peer_port_status = ""
                peer_optical_vendor = ""
                peer_optical_model = ""
                peer_optical_sn = ""
                peer_optical_temp = ""
                peer_lane_data = {}

                # 初始化对端lane数据
                for i in range(self.LANE_NUM):
                    peer_lane_data[f"peer_tx_power{i}"] = ""
                    peer_lane_data[f"peer_rx_power{i}"] = ""
                    peer_lane_data[f"peer_snr_lane{i}"] = ""

                if peer_interface_info:
                    # 端口基本信息
                    if hasattr(peer_interface_info, 'interface_info') and peer_interface_info.interface_info:
                        peer_port_speed = getattr(peer_interface_info.interface_info, 'speed', "")
                        peer_port_duplex = getattr(peer_interface_info.interface_info, 'duplex', "")
                        peer_port_status = getattr(peer_interface_info.interface_info, 'status', "")

                    # 对端光模块信息
                    if hasattr(peer_interface_info, 'transceiver_info') and peer_interface_info.transceiver_info:
                        transceiver_info = peer_interface_info.transceiver_info
                        if hasattr(transceiver_info, 'manufacture_information'):
                            manu_info = transceiver_info.manufacture_information
                            peer_optical_vendor = getattr(manu_info, 'manu_name', "")
                            peer_optical_model = getattr(manu_info, 'manu_part_num', "")
                            peer_optical_sn = getattr(manu_info, 'manu_serial_number', "")

                    # 对端光模块性能信息
                    if hasattr(peer_interface_info, 'swi_optical_model') and peer_interface_info.swi_optical_model:
                        swi_optical_model = peer_interface_info.swi_optical_model
                        peer_optical_temp = getattr(swi_optical_model, 'temperature', "")

                    # 获取对端光模块Lane信息
                    peer_optical_info = peer_interface_info.get_optical_module_info()
                    if peer_optical_info:
                        peer_lane_infos = peer_optical_info.lane_power_infos or []
                        for i in range(self.LANE_NUM):
                            if i < len(peer_lane_infos):
                                peer_lane = peer_lane_infos[i]
                                peer_lane_data[f"peer_tx_power{i}"] = peer_lane.tx_power or ""
                                peer_lane_data[f"peer_rx_power{i}"] = peer_lane.rx_power or ""
                                peer_lane_data[f"peer_snr_lane{i}"] = peer_lane.media_snr or peer_lane.host_snr or ""

                # 获取本地端口信息
                local_port = ""
                if hasattr(npu_chip_info, 'hccn_lldp_info') and npu_chip_info.hccn_lldp_info:
                    # 这里假设本地端口信息可能在hccn_lldp_info中，或者需要其他方式获取
                    # 如果无法直接获取，可以留空或使用chip_id作为端口标识
                    local_port = f"PORT-{chip_id}" if chip_id else ""

                # 创建主机到交换机光模块数据对象
                data = HostToSwitchOpticalModuleData(
                    # 主机信息
                    host_id=host_id,
                    host_name=host_info.hostname or "",
                    host_sn=host_info.sn_num or "",

                    # NPU芯片信息
                    npu_id=npu_chip_info.npu_id or "",
                    chip_id=npu_chip_info.chip_id or "",
                    chip_phy_id=npu_chip_info.chip_phy_id or "",
                    npu_type=npu_chip_info.npu_type or "",

                    # 主机侧端口信息
                    host_port=local_port,

                    # 主机侧光模块基本信息
                    host_optical_present=getattr(optical_info, 'present', ""),
                    host_optical_vendor=getattr(optical_info, 'vendor_name', ""),
                    host_optical_model=getattr(optical_info, 'vendor_part_number', ""),
                    host_optical_sn=optical_module_info.sn or getattr(optical_info, 'vendor_serial_number', ""),

                    # 主机侧光模块性能指标
                    host_optical_temp=getattr(optical_info, 'temperature', ""),
                    host_optical_vcc=getattr(optical_info, 'vcc', ""),
                    host_link_speed=npu_chip_info.speed or "",
                    host_link_duplex=npu_chip_info.duplex or "",
                    host_net_health=npu_chip_info.net_health or "",
                    host_link_status=npu_chip_info.link_status or "",

                    # 主机侧链路统计信息
                    host_link_up_count=getattr(link_stat_info, 'link_up_count', ""),
                    host_link_down_count=getattr(link_stat_info, 'link_down_count', ""),

                    # 主机侧Lane 信息
                    host_tx_power0=lane_data.get("tx_power0", ""),
                    host_rx_power0=lane_data.get("rx_power0", ""),
                    host_tx_bias0=lane_data.get("tx_bias0", ""),
                    host_snr_lane0=lane_data.get("host_snr_lane0", ""),
                    host_media_snr_lane0=lane_data.get("media_snr_lane0", ""),

                    host_tx_power1=lane_data.get("tx_power1", ""),
                    host_rx_power1=lane_data.get("rx_power1", ""),
                    host_tx_bias1=lane_data.get("tx_bias1", ""),
                    host_snr_lane1=lane_data.get("host_snr_lane1", ""),
                    host_media_snr_lane1=lane_data.get("media_snr_lane1", ""),

                    host_tx_power2=lane_data.get("tx_power2", ""),
                    host_rx_power2=lane_data.get("rx_power2", ""),
                    host_tx_bias2=lane_data.get("tx_bias2", ""),
                    host_snr_lane2=lane_data.get("host_snr_lane2", ""),
                    host_media_snr_lane2=lane_data.get("media_snr_lane2", ""),

                    host_tx_power3=lane_data.get("tx_power3", ""),
                    host_rx_power3=lane_data.get("rx_power3", ""),
                    host_tx_bias3=lane_data.get("tx_bias3", ""),
                    host_snr_lane3=lane_data.get("host_snr_lane3", ""),
                    host_media_snr_lane3=lane_data.get("media_snr_lane3", ""),

                    # 对端交换机信息
                    peer_switch_name=peer_switch_name,
                    peer_switch_id=peer_switch_id,
                    peer_switch_sn=peer_switch_sn,

                    # 对端交换机端口信息
                    peer_switch_port=peer_port,
                    peer_switch_port_speed=peer_port_speed,
                    peer_switch_port_duplex=peer_port_duplex,
                    peer_switch_port_status=peer_port_status,

                    # 对端交换机光模块信息
                    peer_switch_optical_vendor=peer_optical_vendor,
                    peer_switch_optical_model=peer_optical_model,
                    peer_switch_optical_sn=peer_optical_sn,
                    peer_switch_optical_temp=peer_optical_temp,

                    # 对端交换机光模块Lane信息
                    peer_switch_tx_power0=peer_lane_data.get("peer_tx_power0", ""),
                    peer_switch_rx_power0=peer_lane_data.get("peer_rx_power0", ""),
                    peer_switch_tx_power1=peer_lane_data.get("peer_tx_power1", ""),
                    peer_switch_rx_power1=peer_lane_data.get("peer_rx_power1", ""),
                    peer_switch_tx_power2=peer_lane_data.get("peer_tx_power2", ""),
                    peer_switch_rx_power2=peer_lane_data.get("peer_rx_power2", ""),
                    peer_switch_tx_power3=peer_lane_data.get("peer_tx_power3", ""),
                    peer_switch_rx_power3=peer_lane_data.get("peer_rx_power3", ""),
                    peer_switch_snr_lane0=peer_lane_data.get("peer_snr_lane0", ""),
                    peer_switch_snr_lane1=peer_lane_data.get("peer_snr_lane1", ""),
                    peer_switch_snr_lane2=peer_lane_data.get("peer_snr_lane2", ""),
                    peer_switch_snr_lane3=peer_lane_data.get("peer_snr_lane3", "")
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
            # 主机侧发射功率阈值（mW）
            ThresholdConfig(
                field_name="host_tx_power0",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_MW,
                display_name="主机侧TX Power Lane 0"
            ),
            ThresholdConfig(
                field_name="host_tx_power1",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_MW,
                display_name="主机侧TX Power Lane 1"
            ),
            ThresholdConfig(
                field_name="host_tx_power2",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_MW,
                display_name="主机侧TX Power Lane 2"
            ),
            ThresholdConfig(
                field_name="host_tx_power3",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_MW,
                display_name="主机侧TX Power Lane 3"
            ),

            # 主机侧接收功率阈值（mW）
            ThresholdConfig(
                field_name="host_rx_power0",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_MW,
                display_name="主机侧RX Power Lane 0"
            ),
            ThresholdConfig(
                field_name="host_rx_power1",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_MW,
                display_name="主机侧RX Power Lane 1"
            ),
            ThresholdConfig(
                field_name="host_rx_power2",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_MW,
                display_name="主机侧RX Power Lane 2"
            ),
            ThresholdConfig(
                field_name="host_rx_power3",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_MW,
                display_name="主机侧RX Power Lane 3"
            ),

            # 主机侧电流阈值（mA）
            ThresholdConfig(
                field_name="host_tx_bias0",
                threshold=threshold_cls.TX_BIAS_MA,
                display_name="主机侧TX Bias Lane 0"
            ),
            ThresholdConfig(
                field_name="host_tx_bias1",
                threshold=threshold_cls.TX_BIAS_MA,
                display_name="主机侧TX Bias Lane 1"
            ),
            ThresholdConfig(
                field_name="host_tx_bias2",
                threshold=threshold_cls.TX_BIAS_MA,
                display_name="主机侧TX Bias Lane 2"
            ),
            ThresholdConfig(
                field_name="host_tx_bias3",
                threshold=threshold_cls.TX_BIAS_MA,
                display_name="主机侧TX Bias Lane 3"
            ),

            # 主机侧Host SNR阈值（dB）
            ThresholdConfig(
                field_name="host_snr_lane0",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="主机侧Host SNR Lane 0"
            ),
            ThresholdConfig(
                field_name="host_snr_lane1",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="主机侧Host SNR Lane 1"
            ),
            ThresholdConfig(
                field_name="host_snr_lane2",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="主机侧Host SNR Lane 2"
            ),
            ThresholdConfig(
                field_name="host_snr_lane3",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="主机侧Host SNR Lane 3"
            ),

            # 主机侧Media SNR阈值（dB）
            ThresholdConfig(
                field_name="host_media_snr_lane0",
                threshold=threshold_cls.MEDIA_SNR_DB,
                display_name="主机侧Media SNR Lane 0"
            ),
            ThresholdConfig(
                field_name="host_media_snr_lane1",
                threshold=threshold_cls.MEDIA_SNR_DB,
                display_name="主机侧Media SNR Lane 1"
            ),
            ThresholdConfig(
                field_name="host_media_snr_lane2",
                threshold=threshold_cls.MEDIA_SNR_DB,
                display_name="主机侧Media SNR Lane 2"
            ),
            ThresholdConfig(
                field_name="host_media_snr_lane3",
                threshold=threshold_cls.MEDIA_SNR_DB,
                display_name="主机侧Media SNR Lane 3"
            ),

            # 主机侧网络状态阈值（字符串相等判断）
            ThresholdConfig(
                field_name="host_link_duplex",
                threshold=threshold_cls.DUPLEX_THRESHOLD,
                display_name="主机侧双工模式"
            ),
            ThresholdConfig(
                field_name="host_net_health",
                threshold=threshold_cls.NET_HEALTH_THRESHOLD,
                display_name="主机侧网络健康"
            ),
            ThresholdConfig(
                field_name="host_link_status",
                threshold=threshold_cls.LINK_STATUS_THRESHOLD,
                display_name="主机侧链路状态"
            ),
            ThresholdConfig(
                field_name="host_optical_present",
                threshold=threshold_cls.OPTICAL_PRESENT_THRESHOLD,
                display_name="主机侧光模块状态"
            ),

            # 对端交换机光模块Lane功率阈值（dBm）- 复用主机端配置
            ThresholdConfig(
                field_name="peer_switch_tx_power0",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端TX Power Lane 0"
            ),
            ThresholdConfig(
                field_name="peer_switch_tx_power1",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端TX Power Lane 1"
            ),
            ThresholdConfig(
                field_name="peer_switch_tx_power2",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端TX Power Lane 2"
            ),
            ThresholdConfig(
                field_name="peer_switch_tx_power3",
                threshold=threshold_cls.TX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端TX Power Lane 3"
            ),
            ThresholdConfig(
                field_name="peer_switch_rx_power0",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端RX Power Lane 0"
            ),
            ThresholdConfig(
                field_name="peer_switch_rx_power1",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端RX Power Lane 1"
            ),
            ThresholdConfig(
                field_name="peer_switch_rx_power2",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端RX Power Lane 2"
            ),
            ThresholdConfig(
                field_name="peer_switch_rx_power3",
                threshold=threshold_cls.RX_POWER_THRESHOLD_CONFIG_DBM,
                display_name="对端RX Power Lane 3"
            ),

            # 对端交换机光模块Lane SNR阈值（dB）- 复用主机端配置
            ThresholdConfig(
                field_name="peer_switch_snr_lane0",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="对端SNR Lane 0"
            ),
            ThresholdConfig(
                field_name="peer_switch_snr_lane1",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="对端SNR Lane 1"
            ),
            ThresholdConfig(
                field_name="peer_switch_snr_lane2",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="对端SNR Lane 2"
            ),
            ThresholdConfig(
                field_name="peer_switch_snr_lane3",
                threshold=threshold_cls.HOST_SNR_DB,
                display_name="对端SNR Lane 3"
            )
        ]
