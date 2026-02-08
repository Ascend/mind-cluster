from typing import Dict, Tuple, Type

from diag_tool.core.common.json_obj import JsonObj
from diag_tool.core.config import port_mapping_config
from diag_tool.core.config.threshold import OpticalModuleThreshold
from diag_tool.core.model.bmc import BmcInfo
from diag_tool.core.model.cluster_mapping import ChassisMapping, L1SwiServerMapping
from diag_tool.core.model.host import HostInfo
from diag_tool.core.model.switch import SwitchInfo, InterfaceFullInfo


class ClusterInfoCache(JsonObj):

    def __init__(self, hosts_info: Dict[str, HostInfo] = None, bmcs_info: Dict[str, BmcInfo] = None,
                 swis_info: Dict[str, SwitchInfo] = None):
        self.hosts_info: Dict[str, HostInfo] = hosts_info or {}
        self.bmcs_info: Dict[str, BmcInfo] = bmcs_info or {}
        self.swis_info: Dict[str, SwitchInfo] = swis_info or {}
        self._swi_info_name_map = {swi_info.name: swi_info for swi_id, swi_info in self.swis_info.items()}
        self._chassis_mappings: ChassisMapping = None
        # 后续通过客户类型修改阈值
        self._threshold: Type[OpticalModuleThreshold] = OpticalModuleThreshold

    def update(self, cache: "ClusterInfoCache"):
        self.hosts_info.update(cache.hosts_info)
        self.bmcs_info.update(cache.bmcs_info)
        self.swis_info.update(cache.swis_info)

    def init_diag_data(self):
        l1_swi_server_mappings = self._build_l1_swi_server_mappings()
        self._chassis_mappings = ChassisMapping(l1_swi_server_mappings)
        self._swi_info_name_map = {swi_info.name: swi_info for swi_id, swi_info in self.swis_info.items()}

    def get_threshold(self):
        return self._threshold

    def get_chassis_mappings(self):
        return self._chassis_mappings

    # 根据服务器superpod id找到服务器信息
    def find_host_info_by_server_spod_id(self, server_id) -> HostInfo:
        for host_info in self.hosts_info.values():
            if server_id == host_info.server_index:
                return host_info
        return None

    def find_bmc_info_by_sn_num(self, sn_num) -> BmcInfo:
        for bmc_info in self.bmcs_info.values():
            if bmc_info.sn_num == sn_num:
                return bmc_info
        return None

    def find_host_info_by_sn_num(self, sn_num) -> HostInfo:
        for host_info in self.hosts_info.values():
            if host_info.sn_num == sn_num:
                return host_info
        return None

    # 找对端交换机
    def find_peer_swi(self, peer_device: str) -> SwitchInfo:
        return self._swi_info_name_map.get(peer_device)

    # 找对端端口信息
    def find_peer_swi_interface_info(self, peer_device: str, peer_interface: str) -> Tuple[
        SwitchInfo, InterfaceFullInfo]:
        peer_device_info = self.find_peer_swi(peer_device)
        if not peer_device_info:
            return None, None
        return peer_device_info, peer_device_info.interface_full_infos.get(peer_interface)

    # 根据接口全量信息找对端端口
    def find_peer_swi_interface_info_by_if_info(self, interface_full_info: InterfaceFullInfo) -> Tuple[
        SwitchInfo, InterfaceFullInfo]:
        if not interface_full_info.interface_mapping:
            return None, None
        remote_device_interface = interface_full_info.interface_mapping.remote_device_interface
        if not remote_device_interface:
            return None, None
        swi_info, peer_interface_info = self.find_peer_swi_interface_info(
            remote_device_interface.device_name, remote_device_interface.interface
        )
        if peer_interface_info:
            return swi_info, peer_interface_info
        return None, None

    # 构建L1 1520交换板与服务器连接关系表
    def _build_l1_swi_server_mappings(self):
        l1_swi_server_mappings = []
        port_mapping_config_instance = port_mapping_config.get_port_mapping_config_instance()
        for switch_info in self.swis_info.values():
            if not switch_info.hccs_info:
                continue
            if not switch_info.hccs_info.hccs_map_table_list:
                continue
            for hccs_map_table in switch_info.hccs_info.hccs_map_table_list:
                if port_mapping_config_instance.is_local_addr(hccs_map_table.start_addr):
                    continue
                cur_port_mapping = port_mapping_config_instance.find_global_addr_port_mapping(hccs_map_table.start_addr)
                if not cur_port_mapping:
                    continue
                host_info = self.find_host_info_by_server_spod_id(cur_port_mapping.server_id)
                if not host_info:
                    continue
                bmc_info = self.find_bmc_info_by_sn_num(host_info.sn_num)
                mapping = L1SwiServerMapping(switch_info.swi_id, switch_info.name,
                                             host_info and host_info.server_index,
                                             host_info and host_info.host_id, bmc_info and bmc_info.bmc_id)
                l1_swi_server_mappings.append(mapping)
                break
        return l1_swi_server_mappings
