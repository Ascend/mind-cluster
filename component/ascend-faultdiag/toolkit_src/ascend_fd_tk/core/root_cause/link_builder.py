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
from typing import Dict, List, Optional, Tuple

from ascend_fd_tk.core.config.port_mapping_config import (
    L1InterfacePortMapping,
    PortMappingConfig,
    get_port_mapping_config_instance,
)
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.host import HostInfo, NpuChipInfo
from ascend_fd_tk.core.model.switch import InterfaceFullInfo, SwitchInfo

from ascend_fd_tk.core.root_cause.constants import (
    PORT_SPEED_200G,
    PORT_SPEED_400G,
    LINK_STATUS_NORMAL,
    LINK_STATUS_ABNORMAL,
    RULE_L1_200G_HILINK,
    RULE_L1_400G_HILINK,
    RULE_L2_200G_HILINK,
    CPU,
    NPU,
    TYPE_HOST_SNR,
    TYPE_MEDIA_SNR,
)
from ascend_fd_tk.core.root_cause.model import (
    HostToL1LinkData,
    L1ToL2LinkData,
)
from ascend_fd_tk.core.root_cause.snr_checker import SnrChecker


class LinkBuilder:
    """信号链路构建器，负责构建Host->L1和L1->L2的链路数据

    构建流程:
        1. 遍历机箱映射配置，为每个L1交换机的200G端口构建Host→L1链路
        2. 遍历所有L1交换机的400G端口，查找对端L2交换机构建L1→L2链路
        3. 对每条链路执行SNR规则检测，标记链路状态和触发规则
    """

    def __init__(self, cluster_info: ClusterInfoCache):
        self._cluster_info = cluster_info
        self._snr_checker = SnrChecker(cluster_info.get_threshold())

    def build(self) -> Tuple[List[HostToL1LinkData], List[L1ToL2LinkData]]:
        """构建所有链路数据，返回 (Host→L1链路列表, L1→L2链路列表)"""
        return self._build_host_to_l1_links(), self._build_l1_to_l2_links()

    # ================================================================
    # Host → L1 链路构建
    # ================================================================

    def _build_host_to_l1_links(self) -> List[HostToL1LinkData]:
        """构建 Host→L1 链路数据

        遍历机箱映射中每个L1交换机的200G端口，结合主机信息和端口映射，
        生成 HostToL1LinkData 并执行 SNR 规则检测
        """
        chassis_mappings = self._cluster_info.get_chassis_mappings()
        if not chassis_mappings:
            return []

        port_mapping_cfg = get_port_mapping_config_instance()
        link_map: Dict[str, HostToL1LinkData] = {}

        for mapping in chassis_mappings.l1_swi_server_mappings:
            l1_swi_info = self._cluster_info.swis_info.get(mapping.l1_swi_ip)
            if not l1_swi_info:
                continue
            host_info = self._cluster_info.find_host_info_by_server_spod_id(mapping.server_super_pod_id)
            self._process_l1_200g_ports(l1_swi_info, host_info, port_mapping_cfg, link_map)

        return list(link_map.values())

    def _process_l1_200g_ports(
        self,
        l1_swi_info: SwitchInfo,
        host_info: Optional[HostInfo],
        port_mapping_cfg: PortMappingConfig,
        link_map: Dict[str, HostToL1LinkData],
    ) -> None:
        """处理L1交换机的所有200G端口，创建或更新链路数据"""
        for interface_name, port_mapping in port_mapping_cfg.l1_interface_port_map.items():
            port_speed = SnrChecker.get_port_speed(interface_name)
            if port_speed != PORT_SPEED_200G:
                continue

            link_key = f"{l1_swi_info.name or l1_swi_info.swi_id}:{interface_name}"
            if link_key not in link_map:
                link_map[link_key] = self._create_host_to_l1_link(l1_swi_info, interface_name, port_speed, port_mapping)

            self._fill_host_info(link_map[link_key], host_info, port_mapping)

    def _create_host_to_l1_link(
        self,
        l1_swi_info: SwitchInfo,
        interface_name: str,
        port_speed: str,
        port_mapping: Optional[L1InterfacePortMapping],
    ) -> HostToL1LinkData:
        """创建一条 Host→L1 链路数据，执行SNR规则检测"""

        triggered_rules: List[str] = []
        link_status = LINK_STATUS_NORMAL

        l1_200g_abnormal, _ = self._snr_checker.check_hilink_snr_abnormal(l1_swi_info, interface_name)
        if l1_200g_abnormal:
            triggered_rules.append(RULE_L1_200G_HILINK)
            link_status = LINK_STATUS_ABNORMAL

        return HostToL1LinkData(
            l1_switch_id=l1_swi_info.swi_id,
            l1_switch_name=l1_swi_info.name or "",
            l1_switch_chip_id=port_mapping.swi_chip_id if port_mapping else "",
            l1_interface=interface_name,
            l1_port_speed=port_speed,
            l1_switch_chip_phy_id=port_mapping.phy_id if port_mapping else "",
            l1_hilink_snr=self._snr_checker.format_hilink_snr(l1_swi_info, interface_name),
            link_status=link_status,
            triggered_rules=triggered_rules,
            room_name=l1_swi_info.room_name,
            cabinet_id=l1_swi_info.cabinet_id,
        )

    def _fill_host_info(
        self,
        data: HostToL1LinkData,
        host_info: Optional[HostInfo],
        port_mapping: Optional[L1InterfacePortMapping],
    ) -> None:
        """填充链路数据中的主机和NPU/CPU信息"""
        if host_info and not data.host_id:
            data.host_id = host_info.host_id
            data.host_name = host_info.hostname or ""
            data.room_name = data.room_name or host_info.room_name or ""
            data.cabinet_id = data.cabinet_id or host_info.cabinet_id or ""

        if port_mapping and port_mapping.xpu and not data.npu_type:
            data.npu_type = port_mapping.xpu.upper()

        npu_chip_info = self._find_npu_chip_info(host_info, port_mapping)
        if npu_chip_info and not data.npu_id:
            data.npu_id = npu_chip_info.npu_id or ""
            if data.npu_type != CPU:
                data.chip_phy_id = npu_chip_info.chip_phy_id or ""
        elif not data.npu_id:
            if port_mapping and port_mapping.xpu and port_mapping.xpu.upper() == CPU:
                data.npu_id = port_mapping.xpu_id or ""

    # ================================================================
    # L1 → L2 链路构建
    # ================================================================

    def _build_l1_to_l2_links(self) -> List[L1ToL2LinkData]:
        """构建 L1→L2 链路数据

        遍历所有L1交换机的400G端口，查找对端L2交换机，
        生成 L1ToL2LinkData 并执行 SNR 规则检测。
        无L2对端信息的400G端口也会被包含（L2字段留空）
        """
        l1_to_l2_links: List[L1ToL2LinkData] = []

        for swi_info in self._cluster_info.swis_info.values():
            if not self._is_l1_switch(swi_info):
                continue
            for interface, full_info in swi_info.interface_full_infos.items():
                if SnrChecker.get_port_speed(interface) != PORT_SPEED_400G:
                    continue
                l2_swi_info, l2_interface_full_info = self._find_peer_l2(full_info)
                data = self._create_l1_to_l2_link(swi_info, interface, full_info, l2_swi_info, l2_interface_full_info)
                l1_to_l2_links.append(data)

        return l1_to_l2_links

    def _find_peer_l2(
        self,
        full_info: InterfaceFullInfo,
    ) -> Tuple[Optional[SwitchInfo], Optional[InterfaceFullInfo]]:
        """根据接口映射信息查找对端L2交换机和接口

        Returns: (l2_swi_info, l2_interface_full_info)，未找到时均为None
        """
        if not full_info.interface_mapping or not full_info.interface_mapping.remote_device_interface:
            return None, None

        remote_device = full_info.interface_mapping.remote_device_interface
        l2_swi_info = self._cluster_info.find_peer_swi(remote_device.device_name)
        if not l2_swi_info:
            return None, None

        l2_interface_full_info = l2_swi_info.interface_full_infos.get(remote_device.interface)
        return l2_swi_info, l2_interface_full_info

    def _create_l1_to_l2_link(
        self,
        swi_info: SwitchInfo,
        interface: str,
        full_info: InterfaceFullInfo,
        l2_swi_info: Optional[SwitchInfo],
        l2_interface_full_info: Optional[InterfaceFullInfo],
    ) -> L1ToL2LinkData:
        """创建一条 L1→L2 链路数据，执行SNR规则检测并填充L1/L2字段"""
        port_mapping_cfg = get_port_mapping_config_instance()
        port_mapping = port_mapping_cfg.find_port_mapping_by_name(interface)
        triggered_rules = self._check_l1_to_l2_rules(
            swi_info, interface, full_info, l2_swi_info, l2_interface_full_info
        )
        link_status = LINK_STATUS_ABNORMAL if triggered_rules else LINK_STATUS_NORMAL

        data = L1ToL2LinkData(
            l1_switch_id=swi_info.swi_id,
            l1_switch_name=swi_info.name or "",
            l1_room_name=swi_info.room_name or "",
            l1_cabinet_id=swi_info.cabinet_id or "",
            l1_switch_chip_id=port_mapping.swi_chip_id if port_mapping else "",
            l1_switch_chip_phy_id=port_mapping.phy_id if port_mapping else "",
            l1_interface=interface,
            l1_port_speed=PORT_SPEED_400G,
            l1_hilink_snr=self._snr_checker.format_hilink_snr(swi_info, interface),
            link_status=link_status,
            triggered_rules=triggered_rules,
        )

        self._fill_l1_optical_snr(data, full_info)
        self._fill_l2_info(data, l2_swi_info, l2_interface_full_info)

        return data

    def _fill_l1_optical_snr(self, data: L1ToL2LinkData, full_info: InterfaceFullInfo) -> None:
        """填充L1光模块的host SNR和media SNR"""
        l1_optical = full_info.get_optical_module_info()
        if l1_optical:
            data.l1_host_snr = self._snr_checker.format_lane_snr(l1_optical, TYPE_HOST_SNR)
            data.l1_media_snr = self._snr_checker.format_lane_snr(l1_optical, TYPE_MEDIA_SNR)

    def _fill_l2_info(
        self,
        data: L1ToL2LinkData,
        l2_swi_info: Optional[SwitchInfo],
        l2_interface_full_info: Optional[InterfaceFullInfo],
    ) -> None:
        """填充L2交换机、端口和光模块信息（无L2信息时字段保持默认空值）"""
        if l2_swi_info:
            data.l2_switch_id = l2_swi_info.swi_id
            data.l2_switch_name = l2_swi_info.name or ""
            data.l2_room_name = l2_swi_info.room_name or ""
            data.l2_cabinet_id = l2_swi_info.cabinet_id or ""

        if l2_interface_full_info:
            data.l2_interface = l2_interface_full_info.interface
            if l2_interface_full_info.interface_info:
                data.l2_port_speed = l2_interface_full_info.interface_info.speed or ""

        l2_interface_name = l2_interface_full_info.interface if l2_interface_full_info else ""
        if l2_swi_info and l2_interface_full_info:
            data.l2_hilink_snr = self._snr_checker.format_hilink_snr(l2_swi_info, l2_interface_name)

        l2_optical = l2_interface_full_info.get_optical_module_info() if l2_interface_full_info else None
        if l2_optical:
            data.l2_host_snr = self._snr_checker.format_lane_snr(l2_optical, TYPE_HOST_SNR)
            data.l2_media_snr = self._snr_checker.format_lane_snr(l2_optical, TYPE_MEDIA_SNR)

    # ================================================================
    # SNR 规则检测
    # ================================================================

    def _check_l1_to_l2_rules(
        self,
        l1_swi_info: SwitchInfo,
        l1_interface_name: str,
        l1_interface: InterfaceFullInfo,
        l2_swi_info: Optional[SwitchInfo],
        l2_interface_full_info: Optional[InterfaceFullInfo],
    ) -> List[str]:
        """检测 L1→L2 链路的SNR异常规则

        规则:
        - l1_400g_hilink: L1 400G端口hilink SNR异常 (规则R2)
        - l1_optical_host: L1光模块host SNR异常 (规则R3)
        - l1_optical_media: L1光模块media SNR异常 (规则R4)
        - l2_optical_host: L2光模块host SNR异常 (规则R5)
        - l2_optical_media: L2光模块media SNR异常 (规则R6)
        - l2_200g_hilink: L2 200G端口hilink SNR异常 (规则R7)
        """
        triggered_rules: List[str] = []

        _, l1_400g_abnormal = self._snr_checker.check_hilink_snr_abnormal(l1_swi_info, l1_interface_name)
        if l1_400g_abnormal:
            triggered_rules.append(RULE_L1_400G_HILINK)

        self._check_optical_rules(l1_interface, "l1", triggered_rules)
        self._check_optical_rules(l2_interface_full_info, "l2", triggered_rules)

        if l2_swi_info:
            l2_interface_name = l2_interface_full_info.interface if l2_interface_full_info else ""
            l2_200g_abnormal, _ = self._snr_checker.check_hilink_snr_abnormal(l2_swi_info, l2_interface_name)
            if l2_200g_abnormal:
                triggered_rules.append(RULE_L2_200G_HILINK)

        return triggered_rules

    def _check_optical_rules(
        self,
        interface: Optional[InterfaceFullInfo],
        prefix: str,
        triggered_rules: List[str],
    ) -> None:
        """检测光模块的host/media SNR异常规则

        Args:
            interface: 接口信息，可为None
            prefix: 规则前缀，"l1"或"l2"
            triggered_rules: 触发规则列表，结果追加到此列表
        """
        optical = interface.get_optical_module_info() if interface else None
        if not optical:
            return
        if self._snr_checker.check_optical_host_snr_abnormal(optical):
            triggered_rules.append(f"{prefix}_optical_host")
        if self._snr_checker.check_optical_media_snr_abnormal(optical):
            triggered_rules.append(f"{prefix}_optical_media")

    # ================================================================
    # 辅助方法
    # ================================================================

    @staticmethod
    def _find_npu_chip_info(
        host_info: Optional[HostInfo],
        port_mapping: Optional[L1InterfacePortMapping],
    ) -> Optional[NpuChipInfo]:
        """根据端口映射查找主机上对应的NPU/CPU芯片信息"""
        if not host_info or not port_mapping or not port_mapping.xpu or port_mapping.xpu_id is None:
            return None
        for _, chip in host_info.npu_chip_info.items():
            if port_mapping.xpu.upper() == NPU:
                if chip.npu_id == port_mapping.xpu_id and chip.chip_id == port_mapping.chip_id:
                    return chip
            elif port_mapping.xpu.upper() == CPU:
                if chip.npu_id == port_mapping.xpu_id:
                    return chip
        return None

    def _is_l1_switch(self, swi_info: SwitchInfo) -> bool:
        """判断交换机是否为L1交换机（通过机箱映射配置匹配）"""
        chassis_mappings = self._cluster_info.get_chassis_mappings()
        if not chassis_mappings:
            return False
        return any(
            mapping.l1_swi_ip == swi_info.swi_id or mapping.l1_swi_name == swi_info.name
            for mapping in chassis_mappings.l1_swi_server_mappings
        )
