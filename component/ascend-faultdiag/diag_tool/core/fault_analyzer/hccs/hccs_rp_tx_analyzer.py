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
import collections
from typing import List, Dict

from diag_tool.core.common.diag_enum import HCCSProxyModule, HccsPackErrorCnt, DeviceType
from diag_tool.core.config import port_mapping_config
from diag_tool.core.context.register import register_analyzer
from diag_tool.core.fault_analyzer.base import Analyzer
from diag_tool.core.model.cluster_info_cache import ClusterInfoCache
from diag_tool.core.model.cluster_mapping import ChassisMapping
from diag_tool.core.model.diag_result import DiagResult, Domain
from diag_tool.core.model.hccs import LCNEInfo
from diag_tool.core.model.switch import SwitchInfo
from diag_tool.utils.date_tool import DateObj

GlobalInterface = collections.namedtuple("GlobalInterface", ["server_id", "interface_name"])


class HCCSCommonAnalyzer(Analyzer):
    LONG_LINK_THRESHOLD = 20

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self.chassis_mappings: ChassisMapping = cluster_info.get_chassis_mappings()
        self.swis_info = {k: v for k, v in cluster_info.swis_info.items() if v.hccs_info}
        self.port_mapping_config = port_mapping_config.get_port_mapping_config_instance()
        # 采集数据处理
        self.lcne_infos: Dict[str, Dict[str, LCNEInfo]] = {}

    @staticmethod
    def check_long_link_down(collect_time: str, lcne_info: LCNEInfo) -> bool:
        link_down_time = lcne_info.get_first_link_down_time()
        if not link_down_time:
            return False
        collect_time_obj = DateObj(collect_time, "%Y-%m-%d %H:%M:%S%z")
        link_down_time_obj = DateObj(link_down_time, "%Y-%m-%d %H:%M:%S:%f")
        if collect_time_obj.diff_seconds(link_down_time_obj) > HCCSCommonAnalyzer.LONG_LINK_THRESHOLD:
            return True
        return False

    @staticmethod
    def check_link_up_down():
        return False

    def analyse(self) -> List[DiagResult]:
        pass

    def data_init(self):
        for single_swi_info in self.swis_info.values():
            switch_ip = single_swi_info.swi_id
            swi_server_info = self.chassis_mappings.find_mapping_by_l1_swi_ip(switch_ip)
            if not swi_server_info or not swi_server_info.server_super_pod_id:
                continue
            self.build_lcne_infos(single_swi_info, swi_server_info.server_super_pod_id)

    def build_lcne_infos(self, swi_info: SwitchInfo, server_id: str):
        lcne_info_list = []
        for interface in self.port_mapping_config.l1_interface_port_map.keys():
            lcne_info = LCNEInfo()
            lcne_info.interface = interface
            interface_port_info = self.port_mapping_config.l1_interface_port_map.get(interface)
            lcne_info.port_id = interface_port_info.phy_id
            lcne_info.die_id = interface_port_info.chip_id
            lcne_info.npu_id = interface_port_info.xpu_id
            lcne_info.chip_id = interface_port_info.swi_chip_id
            lcne_info.server_id = server_id

            hccs_route_miss = swi_info.hccs_info.get_route_miss_by_interface(interface)
            lcne_info.rp_direction_miss = hccs_route_miss.rp_direction_miss
            lcne_info.lp_direction_miss = hccs_route_miss.lp_direction_miss
            lcne_info.nc_direction_miss = hccs_route_miss.nc_direction_miss

            lcne_info.link_status = swi_info.hccs_info.get_link_status_by_chip_port(lcne_info.chip_id,
                                                                                    lcne_info.port_id)

            lp_id_using_cnt = swi_info.hccs_info.get_package_block_by_condition(
                lcne_info.chip_id, lcne_info.port_id, HCCSProxyModule.LP.value, HccsPackErrorCnt.LP_PACK_STUACK.value)
            rp_id_using_cnt = swi_info.hccs_info.get_package_block_by_condition(
                lcne_info.chip_id, lcne_info.port_id, HCCSProxyModule.RP.value, HccsPackErrorCnt.RP_PACK_STUACK.value)
            lcne_info.lp_id_using_cnt = lp_id_using_cnt if lp_id_using_cnt else 0
            lcne_info.rp_id_using_cnt = rp_id_using_cnt if rp_id_using_cnt else 0

            lcne_info_list.append(lcne_info)
        self.lcne_infos.update({
            server_id: {item.interface: item for item in lcne_info_list}
        })

    def get_lcne_info(self, server_id: str, interface: str):
        single_server_lcne_infos = self.lcne_infos.get(server_id, {})
        if not single_server_lcne_infos:
            return None
        return single_server_lcne_infos.get(interface, None)


@register_analyzer
class HCCSAnalyzer(HCCSCommonAnalyzer):
    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self.data_init()

    def filter_timeout_interface(self, swi_info: SwitchInfo, server_id: str):
        rp_tx_timeout_interfaces = {}
        for proxy_timeout in swi_info.hccs_info.proxy_timeout_statis:
            if proxy_timeout.is_rp_tx_timeout_happend():
                rp_tx_timeout_interfaces.update(
                    self.filter_rp_tx_timeout_interface(swi_info, server_id, proxy_timeout.interface)
                )
        return rp_tx_timeout_interfaces

    def filter_rp_tx_timeout_interface(self, swi_info: SwitchInfo, server_id: str, proxy_timeout_interface: str):
        timeout_addr_list = swi_info.hccs_info.get_timeout_detail_by_condition(
            proxy_timeout_interface, "REMOTE_PROXY", "TX_TIMEOUT")
        if not timeout_addr_list:
            return {}
        timeout_interfaces = []
        for timeout_addr in timeout_addr_list:
            if self.port_mapping_config.is_local_addr(timeout_addr):
                local_addr_mapping = self.port_mapping_config.find_local_addr_port_mapping(timeout_addr)
                if not local_addr_mapping:
                    continue
                interface = self.port_mapping_config.find_interface(
                    proxy_timeout_interface, local_addr_mapping.xpu_id, local_addr_mapping.chip_id)
                timeout_interfaces.append(GlobalInterface(server_id, interface))
            else:
                global_addr_mapping = self.port_mapping_config.find_global_addr_port_mapping(timeout_addr)
                if not global_addr_mapping:
                    continue
                interface = self.port_mapping_config.find_interface(
                    proxy_timeout_interface, global_addr_mapping.npu_id, global_addr_mapping.chip_id)
                timeout_interfaces.append(GlobalInterface(global_addr_mapping.server_id, interface))
        rp_tx_timeout_interfaces = {}
        rp_tx_timeout_interfaces.update({
            GlobalInterface(server_id, proxy_timeout_interface): timeout_interfaces
        })
        return rp_tx_timeout_interfaces

    def check_local_interface(self, swi_info: SwitchInfo, lcne_info: LCNEInfo):
        diag_results = []
        if not lcne_info:
            return diag_results
        local_domain = [
            Domain(DeviceType.SWITCH.value, swi_info.swi_id),
            Domain(DeviceType.SWI_PORT.value, str(lcne_info))
        ]
        if self.check_long_link_down(swi_info.date_time, lcne_info):
            diag_results.append(DiagResult(local_domain, "交换机端口长期down", "排查交换机端口link状态信息"))
        if self.check_link_up_down():
            diag_results.append(DiagResult(local_domain, "交换机端口闪断", "排查交换机端口link状态信息"))
        if lcne_info.is_rp_pack_block():
            diag_results.append(DiagResult(local_domain, f"[{local_domain.__str__()}]rp窝包",
                                           f"排查否存在信仰证反压异常"))
        elif lcne_info.is_voq_pack_block():
            diag_results.append(DiagResult(local_domain, f"[{local_domain.__str__()}]voq窝包",
                                           f"排查否存在信仰证反压异常"))
        return diag_results

    def check_remote_interfaces(self, swi_info: SwitchInfo, local_lcne_info: LCNEInfo,
                                remote_interfaces: List[GlobalInterface]):
        diag_results = []
        for remote_interface in remote_interfaces:
            remote_lcne_info = self.get_lcne_info(remote_interface.server_id, remote_interface.interface_name)
            if not remote_lcne_info:
                continue
            remote_domain = [
                Domain(DeviceType.SWITCH.value, swi_info.swi_id),
                Domain(DeviceType.SWI_PORT.value, remote_lcne_info.__str__())
            ]
            if remote_lcne_info.is_lp_route_miss():
                diag_results.append(DiagResult(remote_domain,
                                               f"[{remote_lcne_info.__str__()}] lp方向路由miss",
                                               f"[{remote_lcne_info.__str__()}] -> [{local_lcne_info.__str__()}]"))
            if remote_lcne_info.is_lp_pack_block():
                diag_results.append(DiagResult(remote_domain,
                                               f"[{remote_lcne_info.__str__()}]lp窝包",
                                               f"[{remote_lcne_info.__str__()}] -> [{local_lcne_info.__str__()}]"))
            elif remote_lcne_info.is_voq_pack_block():
                diag_results.append(DiagResult(remote_domain,
                                               f"[{remote_lcne_info.__str__()}]voq窝包",
                                               f"[{remote_lcne_info.__str__()}] -> [{local_lcne_info.__str__()}]"))
            # 排查L2
            if local_lcne_info.server_id != remote_interface.server_id:
                diag_results.extend(self.l2_diag(local_lcne_info, remote_lcne_info))
        return diag_results

    def rp_tx_timeout_diag(self, rp_tx_timeout_interfaces: Dict[GlobalInterface, List[GlobalInterface]],
                           swi_info: SwitchInfo):
        diag_results = []
        for interface, remote_interfaces in rp_tx_timeout_interfaces.items():
            # rp端口问题
            lcne_info = self.get_lcne_info(interface.server_id, interface.interface_name)
            diag_results.extend(self.check_local_interface(swi_info, lcne_info))
            # 对端问题
            self.check_remote_interfaces(swi_info, lcne_info, remote_interfaces)
        return diag_results

    def l2_diag(self, local_lcne_info: LCNEInfo, remote_lcne_info: LCNEInfo):
        return []

    def analyse(self):
        if not self.swis_info:
            return []
        diag_results = []
        for swi_info in self.swis_info.values():
            for proxy_timeout in swi_info.hccs_info.proxy_timeout_statis:
                if proxy_timeout.is_rp_tx_timeout_happend():
                    domain = [
                        Domain(DeviceType.SWITCH.value, swi_info.swi_id),
                        Domain(DeviceType.SWI_PORT.value, proxy_timeout.interface)
                    ]
                    fault_info = f"HCCS RP TX超时，超时次数：{proxy_timeout.rp_tx}"
                    suggestion = "交换机端口长期down、端口闪断、窝包或者路由miss"
                    diag_results.append(DiagResult(domain, fault_info, suggestion))

            swi_server_info = self.chassis_mappings.find_mapping_by_l1_swi_ip(swi_info.swi_id)
            if not swi_server_info or not swi_server_info.server_super_pod_id:
                continue
            rp_tx_timeout_interfaces = self.filter_timeout_interface(swi_info, swi_server_info.server_super_pod_id)
            diag_results.extend(self.rp_tx_timeout_diag(rp_tx_timeout_interfaces, swi_info))
        return diag_results
