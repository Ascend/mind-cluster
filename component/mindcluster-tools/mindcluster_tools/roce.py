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
import os
import glob
import socket
import fcntl
import struct
from typing import Dict, List


def get_ip_address(interface: str) -> str:
    """
    get ip address from network interface
    """
    SIOCGIFADDR = (
        0x8915  # defined in linux/sockios.h which is used to get address of interface
    )
    SIZEOF_SOCKADDR = 256  # sizeof(sockaddr)
    IP_START = 20  # start of ip address
    IP_END = 24  # end of ip address
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        packed_iface = struct.pack(f"{SIZEOF_SOCKADDR}s", interface.encode("utf8"))
        packed_addr = fcntl.ioctl(sock.fileno(), SIOCGIFADDR, packed_iface)[
            IP_START:IP_END
        ]
        return socket.inet_ntoa(packed_addr)
    except (OSError, AttributeError):
        return ""
    finally:
        sock.close()


def get_pcibusid_by_devid(device_id: int) -> str:
    """
    get npu pci bus id by npu_phy_id which is known as device_id
    """
    for base, devices, _ in os.walk("/sys/bus/pci/devices"):
        for bus_id in devices:
            try:
                with open(os.path.join(base, bus_id, "dev_id")) as f:
                    buf = f.read()
                    if buf.strip().isdigit() and int(buf) == device_id:
                        return bus_id
            except FileNotFoundError as e:
                continue
    return ""


def get_network_interfaces() -> List[str]:
    """
    get all network card in current node
    """
    for _, if_list, _ in os.walk("/sys/class/net"):
        return if_list


def get_interface_by_npu(bus_id, interface_list) -> str:
    """
    get network cards by npu bus id
    for example, there is a pcie tree like this
    ----NPU_1
     |- NPU_2
     |- network_card_1
     `- network_card_2
    Then NPU_1 should use network_card_1 and NPU_2 use network_card_2
    return the network card interface name
    """
    abs_path = os.path.realpath(os.path.join("/sys/bus/pci/devices", bus_id))
    pcie_base_path = "/".join(abs_path.split("/")[0:4])
    for interface in interface_list:
        bus_path = os.path.realpath(f"/sys/class/net/{interface}")
        if bus_path.startswith(pcie_base_path):
            interface_list.remove(interface)
            return interface
    return ""


def get_npu_roce_ip() -> Dict[int, str]:
    """
    get {npu_id, ip address} pair.  Each npu use the correct network card
    to get best roce performance. NPU use the network card in same pcie bus switch
    so the NPU can use NDR(NPU Direct RDMA)
    """
    interface_list = get_network_interfaces()
    npus = sorted(
        [int(i[len("/dev/davinci"):]) for i in glob.glob("/dev/davinci[0-9]*")]
    )
    npu_ip_map = {}
    for i in npus:
        bus_id = get_pcibusid_by_devid(i)
        interface = get_interface_by_npu(bus_id, interface_list)
        npu_ip_map[i] = get_ip_address(interface)

    return npu_ip_map
