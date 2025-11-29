#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

from abc import ABC
import copy
import functools
import json
from typing import List
import netifaces
import ctypes

from bitarray.util import hex2ba, ba2int

from mindcluster_tools.dcmi_querier import DCMIQuerier
from mindcluster_tools.interface import ToDict
from mindcluster_tools.topo import TopoSingleFactory
from mindcluster_tools.utils import parse_eid
from mindcluster_tools.dcmi import dcmi
from mindcluster_tools.roce import get_npu_roce_ip
from mindcluster_tools.error.error import ParamError, TopoMissMatchError, GetIpError
from mindcluster_tools.utils.product_type_enum import ProductType
from mindcluster_tools.utils.const import FE0, FE1, FE3, FE8, FE9
from mindcluster_tools.utils.const import (
    URMA_LEVEL_0,
    URMA_LEVEL_1,
    URMA_LEVEL_2,
    URMA_LEVEL_3,
)
from mindcluster_tools.utils.const import (
    ROOTINFO_VERSION,
    CLUSTER_PLANE_ID,
    CLUSTER_NET_INSTANCE,
    CLOS_NET_TYPE,
    TOPO_NET_TYPE
)


# Hierarchy mapping from fe_id to urma_device
URMA_DEVICE_LEVEL_MAP = {
    # UB
    FE0: URMA_LEVEL_1,
    FE1: URMA_LEVEL_0,
    # UBG
    FE3: URMA_LEVEL_2,
    # UBOE
    FE8: URMA_LEVEL_2,
    FE9: URMA_LEVEL_2,
}

SUPER_POD_TYPE_TO_TOPO_FILE = {
    ProductType.SERVER_8P.value: "server_8p.json",
    ProductType.POD_1D.value: "superpod_1d.json",
    ProductType.POD_2D.value: "superpod_2d.json",
    ProductType.SERVER_16P.value: "server_16p.json",
    ProductType.STANDARD_1P.value: "card_1p.json",
    ProductType.STANDARD_4P.value: "card_4p_mesh.json",
}


def exclude_fields(*fields):
    """Decorator to exclude specified field names when converting objects to Dict"""

    def decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            data = func(*args, **kwargs)
            for field in fields:
                data.pop(field, None)
            return data

        return wrapper

    return decorator


class RootInfoEncoder(json.JSONEncoder):
    """Serializer: Convert objects to dictionaries via to_dict method for serialization"""

    def default(self, obj):
        if isinstance(obj, ToDict):
            return obj.to_dict()
        return json.JSONEncoder.default(self, obj)


class Address(ToDict, ABC):
    """Base class for EID/IP addresses"""

    def __init__(self, addr, port_ids=None, plane_id="", addr_type="common type"):
        self.addr_type = addr_type
        self.addr = addr
        self.port_ids = [] if not port_ids else port_ids
        self.plane_id = plane_id


class EID(Address):
    def __init__(self, die_id, eid_type, addr, port_ids, plane_id):
        super().__init__(addr, port_ids, plane_id, "EID")
        self.eid_type = eid_type
        self.die_id = die_id

    @exclude_fields("port_ids", "eid_type", "die_id")
    def to_dict(self):
        ret = copy.deepcopy(self.__dict__)
        ret["ports"] = [f"{self.die_id}/{i}" for i in self.port_ids]
        return ret


class IP(Address):
    def __init__(self, addr, plane_id="", port_ids=None, die_id=None):
        super().__init__(addr, port_ids, plane_id, "IPV4")
        self.die_id = die_id

    @exclude_fields("port_ids", "die_id")
    def to_dict(self):
        ret = copy.deepcopy(self.__dict__)
        ret["ports"] = [f"{self.die_id}/{i}" for i in self.port_ids] if self.die_id is not None else []
        return ret


class UrmaDevice(ToDict):
    def __init__(
        self,
        net_layer,
        net_instance_id,
        net_type,
        net_attr="",
        rank_addr_list: List[Address] = None,
    ):
        self.net_layer = net_layer
        self.net_instance_id = net_instance_id
        self.net_type = net_type
        self.net_attr = net_attr
        self.rank_addr_list = [] if not rank_addr_list else rank_addr_list

    @exclude_fields()
    def to_dict(self):
        return copy.deepcopy(self.__dict__)


class NPU(ToDict):
    def __init__(self, device_id, local_id, level_list: List[UrmaDevice] = None):
        self.device_id = device_id
        self.local_id = local_id
        self.level_list = [] if not level_list else level_list

    @exclude_fields()
    def to_dict(self):
        return copy.deepcopy(self.__dict__)


class RootInfo(ToDict):
    version = ROOTINFO_VERSION

    def __init__(self, rank_list: List[NPU] = None):
        self.version = RootInfo.version
        self.rank_list = [] if not rank_list else rank_list

    @exclude_fields()
    def to_dict(self):
        return copy.deepcopy(self.__dict__)


def construct_urma_device_with_generate_eid(npu_info):
    npu_id, layer_id, die_count, die_port_count, superpod_id, chassis_id = npu_info
    urma_device_level_info_map = {
        URMA_LEVEL_0: (f"{superpod_id}_{chassis_id}", TOPO_NET_TYPE),
        URMA_LEVEL_1: (f"{superpod_id}", CLOS_NET_TYPE),
        URMA_LEVEL_2: ("cluster", CLOS_NET_TYPE),
        URMA_LEVEL_3: ("cluster", CLOS_NET_TYPE),
    }
    querier = DCMIQuerier()
    topo = TopoSingleFactory.get_topo()
    ud = UrmaDevice(layer_id, *urma_device_level_info_map[layer_id], "", [])
    if layer_id == URMA_LEVEL_0:
        construct_l0(
            (chassis_id, die_count, die_port_count, layer_id, npu_id), querier, ud
        )
        addr = querier.query(
            npu_id,
            0,
            1,
            URMA_DEVICE_LEVEL_MAP[layer_id],
            chassis_id,
            parse_eid.EID_TYPE_LOGIC,
        )
        ud.rank_addr_list.append(
            EID(
                0,
                parse_eid.EID_TYPE_LOGIC,
                addr,
                topo.get_ports_by_level_and_die(npu_id, layer_id, 0),
                "0",
            )
        )
        addr = querier.query(
            npu_id,
            1,
            1,
            URMA_DEVICE_LEVEL_MAP[layer_id],
            chassis_id,
            parse_eid.EID_TYPE_LOGIC,
        )
        ud.rank_addr_list.append(
            EID(
                1,
                parse_eid.EID_TYPE_LOGIC,
                addr,
                topo.get_ports_by_level_and_die(npu_id, layer_id, 1),
                "1",
            )
        )
    elif layer_id == 1:
        addr = querier.query(
            npu_id,
            0,
            2,
            URMA_DEVICE_LEVEL_MAP[layer_id],
            chassis_id,
            parse_eid.EID_TYPE_LOGIC,
        )
        ud.rank_addr_list.append(
            EID(
                0,
                parse_eid.EID_TYPE_LOGIC,
                addr,
                topo.get_ports_by_level_and_die(npu_id, layer_id, 0),
                "0",
            )
        )
        addr = querier.query(
            npu_id,
            1,
            2,
            URMA_DEVICE_LEVEL_MAP[layer_id],
            chassis_id,
            parse_eid.EID_TYPE_LOGIC,
        )
        ud.rank_addr_list.append(
            EID(
                1,
                parse_eid.EID_TYPE_LOGIC,
                addr,
                topo.get_ports_by_level_and_die(npu_id, layer_id, 1),
                "1",
            )
        )
    elif layer_id == 3:
        ip_map = get_npu_roce_ip()
        ports = topo.get_ports_by_level(npu_id, layer_id)
        ud.rank_addr_list.append(IP(ip_map[npu_id], CLUSTER_PLANE_ID, ports))
    return ud


def construct_l0(param_info, querier, ud):
    topo = TopoSingleFactory.get_topo()
    chassis_id, die_count, die_port_count, layer_id, npu_id = param_info
    for k1 in range(die_count):
        for k2 in range(die_port_count):
            if not topo.is_p2p_edge(npu_id, f"{k1}/{k2}"):
                # not p2p edge, pass
                continue
            addr = querier.query(
                npu_id,
                k1,
                k2,
                URMA_DEVICE_LEVEL_MAP[layer_id],
                chassis_id,
                parse_eid.EID_TYPE_PHY,
            )
            eid = EID(k1, parse_eid.EID_TYPE_PHY, addr, [k2], f"{k1}")
            ud.rank_addr_list.append(eid)


def parse_dcmi_param(board_id, mainboard_id, rank_count, params, spod_info):
    super_pod_id = ctypes.c_int(spod_info.super_pod_id).value
    chassis_id = spod_info.chassis_id
    server_index = spod_info.server_index
    super_pod_type = int.from_bytes(spod_info.super_pod_type, byteorder="big")
    STANDARD_BOARD_ID = [26, 27]
    # Standard card form factor uses mainboard_id to distinguish different interconnect topologies
    if board_id in STANDARD_BOARD_ID:
        super_pod_type = mainboard_id
        local_id_list = list(range(rank_count))
        super_pod_id = -1
        chassis_id = -1
    # 1D/2D pods use physical ID as local_id
    elif super_pod_type == 1 or super_pod_type == 2:
        local_id_list = [dcmi.get_local_id(i) for i in range(rank_count)]
    else:
        # server_16p form factor requires local_id calculation
        if super_pod_type == 3:
            local_id_list = [
                (server_index % 2) * rank_count + i for i in range(rank_count)
            ]
        else:
            local_id_list = list(range(rank_count))
    if params["topo_path"] is not None:
        topo_path = params["topo_path"]
    else:
        topo_path = SUPER_POD_TYPE_TO_TOPO_FILE.get(super_pod_type, None)
        if topo_path is None:
            raise TopoMissMatchError(
                "Can not found topo file for superpod type {}".format(super_pod_type)
            )
    dcmi_param = (
        topo_path,
        local_id_list,
        super_pod_id,
        chassis_id,
        super_pod_type,
        server_index,
    )
    return dcmi_param


def parse_param(params):
    # query dcmi
    if dcmi.is_dcmi_available():
        rank_count, _ = dcmi.dcmi_get_card_list()
        spod_info = dcmi.get_super_pod_info()
        board_id, mainboard_id = (
            dcmi.get_device_board_info().board_id,
            dcmi.get_mainboard_id(),
        )
        dcmi_param = parse_dcmi_param(
            board_id, mainboard_id, rank_count, params, spod_info
        )
        (
            topo_path,
            local_id_list,
            super_pod_id,
            chassis_id,
            super_pod_type,
            server_index,
        ) = dcmi_param
    # not query dcmi
    else:
        if any(
            [
                params["rank_count"] is None,
                params["super_pod_id"] is None,
                params["chassis_id"] is None,
                params["topo_path"] is None,
            ]
        ):
            raise ParamError("Some parameters are missing")
        local_id_list = [i for i in range(params["rank_count"])]
        super_pod_id = params["super_pod_id"]
        chassis_id = params["chassis_id"]
        topo_path = params["topo_path"]
        super_pod_type = ProductType.POD_2D.value
        server_index = 0

    parsed_param = (
        local_id_list,
        super_pod_id,
        chassis_id,
        topo_path,
        super_pod_type,
        server_index,
    )
    return parsed_param


def get_urma_device_level_by_eid(hex_eid_str):

    eid = hex2ba(hex_eid_str)
    fe_id = ba2int(eid[parse_eid.FE_ID_RANGE_START : parse_eid.FE_ID_RANGE_END])
    return URMA_DEVICE_LEVEL_MAP[fe_id]


def get_ports_info_by_eid(
    hex_eid_str, local_id, urma_device_level, topo, super_pod_type
):
    eid = hex2ba(hex_eid_str)
    port_id = ba2int(eid[parse_eid.PORT_ID_RANGE_START : parse_eid.PORT_ID_RANGE_END])
    # logic EID
    if port_id > parse_eid.LOGIC_PORT_FLAG:
        die_id = (
            (port_id - parse_eid.LOGIC_PORT_FLAG - 1) // parse_eid.DIE_COUNT_IN_A_NPU
        ) % parse_eid.LOGIC_PORT_COUNT_IN_A_DIE
        ports = topo.get_ports_by_level_and_die(local_id, urma_device_level, die_id)
        eid_type = parse_eid.EID_TYPE_LOGIC
    # physic EID
    else:
        # L2 layer in server form factor, directly connected to UBOE
        if port_id == 0 and super_pod_type in [
            ProductType.SERVER_8P.value,
            ProductType.SERVER_16P.value,
        ]:
            L2_UBOE_PORTS = [8]
            ports = L2_UBOE_PORTS
            die_id = 0
            eid_type = None
        else:
            ports = [(port_id - 1) % parse_eid.PHY_PORT_COUNT_IN_A_DIE]
            die_id = (
                (port_id - 1) // parse_eid.PHY_PORT_COUNT_IN_A_DIE
            ) % parse_eid.DIE_COUNT_IN_A_NPU
            eid_type = parse_eid.EID_TYPE_PHY
    return ports, die_id, eid_type


def process_urma_device_for_npu(npu_id: int):
    device_id_max, _, _ = dcmi.get_device_id_in_card(npu_id)
    for i in range(device_id_max):
        urma_device_cnt = dcmi.get_urma_device_cnt(npu_id, i)
        # Currently in simulation, urma_device_cnt is fixed at 1, query all EID information
        if urma_device_cnt == 1:
            return dcmi.get_eid_list_by_urma_dev_index(npu_id, i, urma_device_cnt - 1)
        # Mock interface returns data based on urma_device_level
        eid_list = []
        for j in range(urma_device_cnt):
            eid_list.extend(dcmi.get_eid_list_by_urma_dev_index(npu_id, i, j))
        return eid_list


def eid_filter(eid_list):
    res = []
    for eid_str in eid_list:
        # Mock shared library contains all-zero useless EIDs
        if all(char == "0" for char in eid_str):
            continue
        eid = hex2ba(eid_str)
        fe_id = ba2int(eid[parse_eid.FE_ID_RANGE_START : parse_eid.FE_ID_RANGE_END])
        if fe_id not in URMA_DEVICE_LEVEL_MAP:
            continue
        res.append(eid_str)
    return res


def get_host_ip():
    try:
        interfaces = netifaces.interfaces()
        # exclude 127.0.0.1
        for interface in interfaces:
            if interface == "lo":
                continue
            addrs = netifaces.ifaddresses(interface)
            # find IPV4 address
            if netifaces.AF_INET in addrs:
                ipv4_info = addrs[netifaces.AF_INET][0]
                return ipv4_info["addr"]
    except Exception as e:
        raise GetIpError(
            "Failed to get local IP address, please check network settings"
        ) from e


def get_level0_info(chassis_id, super_pod_id, super_pod_type, server_index, npu_id):
    # 1d/2d pod
    if super_pod_type in (ProductType.POD_1D.value, ProductType.POD_2D.value):
        net_instance_id = f"{super_pod_id}_{chassis_id}"
    # server 8p superpod
    elif super_pod_type == ProductType.SERVER_8P.value and super_pod_id != -1:
        net_instance_id = f"{super_pod_id}_{server_index}"
    # All others are IP addresses
    else:
        net_instance_id = get_host_ip()
        # For standard card 2p/4p, add group identifier
        if super_pod_type in (
            ProductType.STANDARD_2P.value,
            ProductType.STANDARD_4P.value,
        ):
            net_instance_id += (
                f"_{npu_id // (super_pod_type - ProductType.STANDARD_1P.value)}"
            )
    return net_instance_id, TOPO_NET_TYPE


def cut_ip_from_eid(eid_str):
    IP_HEX_LENGTH = 8
    ip_int = int(eid_str[len(eid_str) - IP_HEX_LENGTH :], 16)

    # Extract four 8-bit bytes
    octets = [
        (ip_int >> 24) & 0xFF,
        (ip_int >> 16) & 0xFF,
        (ip_int >> 8) & 0xFF,
        ip_int & 0xFF,
    ]
    ip_address = ".".join(str(octet) for octet in octets)
    return ip_address


def get_urma_device_map(device_info):
    chassis_id, npu_id, super_pod_id, local_id, super_pod_type, server_index = device_info
    urma_device_level_info_map = {
        URMA_LEVEL_0: get_level0_info(
            chassis_id, super_pod_id, super_pod_type, server_index, npu_id
        ),
        # Non-super nodes don't have L1 and won't match
        URMA_LEVEL_1: (f"{super_pod_id}", CLOS_NET_TYPE),
        URMA_LEVEL_2: (CLUSTER_NET_INSTANCE, CLOS_NET_TYPE),
        URMA_LEVEL_3: (CLUSTER_NET_INSTANCE, CLOS_NET_TYPE),
    }
    eid_list = process_urma_device_for_npu(npu_id)
    eid_list = eid_filter(eid_list)
    topo = TopoSingleFactory.get_topo()
    urma_device_map = {}
    for eid_str in eid_list:
        urma_device_level = get_urma_device_level_by_eid(eid_str)
        if urma_device_level not in urma_device_map:
            urma_device_map[urma_device_level] = UrmaDevice(
                urma_device_level,
                *urma_device_level_info_map[urma_device_level],
                "",
                [],
            )
        ports, die_id, eid_type = get_ports_info_by_eid(
            eid_str, local_id, urma_device_level, topo, super_pod_type
        )
        # Filter out empty ports
        if ports:
            if eid_type is not None:
                urma_device_map[urma_device_level].rank_addr_list.append(
                    EID(
                        die_id,
                        eid_type,
                        eid_str,
                        ports,
                        str(die_id) if urma_device_level <= 1 else CLUSTER_PLANE_ID,
                    )
                )
            else:
                uboe_ip = cut_ip_from_eid(eid_str)
                urma_device_map[urma_device_level].rank_addr_list.append(
                    IP(uboe_ip, CLUSTER_PLANE_ID, ports, die_id)
                )
    urma_device_map[URMA_LEVEL_3] = UrmaDevice(
        URMA_LEVEL_3, *urma_device_level_info_map[URMA_LEVEL_3], "", []
    )
    ip_map = get_npu_roce_ip()
    if npu_id in ip_map:
        ports = topo.get_ports_by_level(npu_id, URMA_LEVEL_3)
        urma_device_map[URMA_LEVEL_3].rank_addr_list.append(IP(ip_map[npu_id], CLUSTER_PLANE_ID, ports))
    if (
        super_pod_type == ProductType.SERVER_16P.value or super_pod_id == -1
    ) and URMA_LEVEL_1 in urma_device_map:
        del urma_device_map[URMA_LEVEL_1]
    return urma_device_map


def construct_rootinfo(params):
    (
        local_id_list,
        super_pod_id,
        chassis_id,
        topo_path,
        super_pod_type,
        server_index,
    ) = parse_param(params)
    die_count, die_port_count = params["die_count"], params["die_port_count"]
    TopoSingleFactory.set_topo_path(topo_path)
    rootinfo = RootInfo([])
    levels = TopoSingleFactory.get_topo().get_level_list()
    for device_id, local_id in enumerate(local_id_list):
        # Direct rootinfo generation requires this parameter, otherwise it will be read via DCMI
        if dcmi.is_dcmi_available():
            npu = NPU(device_id, local_id, [])
            device_info = (
                chassis_id,
                device_id,
                super_pod_id,
                local_id,
                super_pod_type,
                server_index,
            )
            urma_device_map = get_urma_device_map(device_info)
            for key in sorted(urma_device_map.keys()):
                if urma_device_map[key].rank_addr_list:
                    npu.level_list.append(urma_device_map[key])
        else:
            npu = NPU(device_id, device_id, [])
            for j in levels:
                npu_info = (device_id, j, die_count, die_port_count, super_pod_id, chassis_id)
                ud = construct_urma_device_with_generate_eid(npu_info)
                npu.level_list.append(ud)
        # Query DCMI to get URMA device information
        rootinfo.rank_list.append(npu)

    ret = json.dumps(rootinfo, indent=2, cls=RootInfoEncoder)
    return ret
