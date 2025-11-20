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
import ctypes
import functools
import os

from mindcluster_tools.dcmi.dcmi_cmd import DcmiMainCmdEnum, MAIN_CMD_TO_SUB_CMD
from mindcluster_tools.error.error import DcmiReturnValueError
from mindcluster_tools.utils.const import EID_SIZE

# Mocking ip starts from 0
_roce_ip_start = 0


def validate_process(func):
    """Decorator that automatically adds validation to all methods beginning with _process"""

    @functools.wraps(func)
    def wrapped_method(*args, **kwargs):
        result = func(*args, **kwargs)
        if result != 0:
            raise DcmiReturnValueError(
                f"function [{func.__name__}] return value is not 0]"
            )
        return result

    return wrapped_method


class In4(ctypes.Structure):
    _fields_ = [
        ("reserved", ctypes.c_uint64),
        ("prefix", ctypes.c_uint32),
        ("addr", ctypes.c_uint32),
    ]


class In6(ctypes.Structure):
    _fields_ = [
        ("subnet_prefix", ctypes.c_uint64),
        ("interface_id", ctypes.c_uint64),
    ]


class EidInfo(ctypes.Structure):
    _fields_ = [
        ("eid", ctypes.c_ubyte * EID_SIZE),
        ("eid_index", ctypes.c_uint32),
        ("reserved", ctypes.c_ubyte * 4),
    ]


class CSuperPodInfo(ctypes.Structure):
    _reserve_len = 27
    _fields_ = [
        ("sdid", ctypes.c_uint),
        ("super_pod_size", ctypes.c_uint),
        ("super_pod_id", ctypes.c_uint),
        ("server_index", ctypes.c_uint),
        ("chassis_id", ctypes.c_uint),
        ("super_pod_type", ctypes.c_char),
        ("reserve", ctypes.c_char * _reserve_len),
    ]


class CBoardInfo(ctypes.Structure):
    _fields_ = [
        ("board_id", ctypes.c_uint),
        ("pcb_id", ctypes.c_uint),
        ("bom_id", ctypes.c_uint),
        ("slot_id", ctypes.c_uint),
    ]


@functools.lru_cache()
def _get_dcmi_lib():
    dcmi_lib = ctypes.CDLL("libdcmi.so", mode=os.RTLD_LAZY | os.RTLD_GLOBAL)
    return dcmi_lib


@validate_process
def _process_dcmi_init(dcmi_lib):
    return dcmi_lib.dcmi_init()


def dcmi_init():
    """DCMI initialization"""
    dcmi_lib = _get_dcmi_lib()
    dcmi_lib.dcmi_init.argtypes = None
    dcmi_lib.dcmi_init.restype = ctypes.c_int
    return _process_dcmi_init(dcmi_lib)


@validate_process
def _process_dcmi_get_all_device_count(dcmi_lib, count):
    return dcmi_lib.dcmi_get_all_device_count(count)


def get_all_device_count():
    """Get total number of devices"""
    dcmi_lib = _get_dcmi_lib()
    dcmi_lib.dcmi_get_all_device_count.argtypes = [ctypes.POINTER(ctypes.c_int)]
    dcmi_lib.dcmi_get_all_device_count.restype = ctypes.c_int
    count = ctypes.c_int(0)
    _process_dcmi_get_all_device_count(dcmi_lib, ctypes.byref(count))
    return count.value


@validate_process
def _process_dcmi_get_card_list(dcmi_lib, card_num, card_list, list_len):
    return dcmi_lib.dcmi_get_card_list(card_num, card_list, list_len)


def dcmi_get_card_list():
    """Get NPU ID list"""
    dcmi_lib = _get_dcmi_lib()
    list_len = 16
    dcmi_lib.dcmi_get_card_list.argtypes = [
        ctypes.POINTER(ctypes.c_int),
        ctypes.POINTER(ctypes.c_int * list_len),
        ctypes.c_int,
    ]
    dcmi_lib.dcmi_get_card_list.restype = ctypes.c_int
    card_num, card_list = ctypes.c_int(0), (ctypes.c_int * list_len)()
    _process_dcmi_get_card_list(
        dcmi_lib, ctypes.byref(card_num), ctypes.byref(card_list), list_len
    )
    return card_num.value, card_list[: card_num.value]


@validate_process
def _process_dcmi_get_device_id_in_card(
    dcmi_lib, card_id, device_id_max, mcu_id, cpu_id
):
    return dcmi_lib.dcmi_get_device_id_in_card(card_id, device_id_max, mcu_id, cpu_id)


def get_device_id_in_card(card_id):
    """Get the (number of chips, MCU_ID, CPU_ID) on the specified NPU management unit"""
    dcmi_lib = _get_dcmi_lib()
    dcmi_lib.dcmi_get_device_id_in_card.argtypes = [
        ctypes.c_int,
        ctypes.POINTER(ctypes.c_int),
        ctypes.POINTER(ctypes.c_int),
        ctypes.POINTER(ctypes.c_int),
    ]
    dcmi_lib.dcmi_get_device_id_in_card.restype = ctypes.c_int
    device_id_max, mcu_id, cpu_id = ctypes.c_int(0), ctypes.c_int(0), ctypes.c_int(0)
    _process_dcmi_get_device_id_in_card(
        dcmi_lib,
        card_id,
        ctypes.byref(device_id_max),
        ctypes.byref(mcu_id),
        ctypes.byref(cpu_id),
    )
    return device_id_max.value, mcu_id.value, cpu_id.value


@validate_process
def _process_get_super_pod_info(
    dcmi_lib, card_id, device_id, main_cmd, sub_cmd, buf, size
):
    return dcmi_lib.dcmi_get_device_info(
        card_id, device_id, main_cmd, sub_cmd, buf, size
    )


def get_super_pod_info():
    """Get superpod information"""
    dcmi_lib = _get_dcmi_lib()
    dcmi_lib.dcmi_get_device_info.argtypes = [
        ctypes.c_int,
        ctypes.c_int,
        ctypes.c_uint,
        ctypes.c_uint,
        ctypes.c_void_p,
        ctypes.POINTER(ctypes.c_uint),
    ]
    dcmi_lib.dcmi_get_device_info.restype = ctypes.c_int
    card_id, device_id = 0, 0
    main_cmd = DcmiMainCmdEnum.DCMI_MAIN_CMD_CHIP_INF.value
    sub_enum = MAIN_CMD_TO_SUB_CMD.get(DcmiMainCmdEnum.DCMI_MAIN_CMD_CHIP_INF, None)
    if sub_enum is None:
        raise KeyError
    sub_cmd = sub_enum.DCMI_CHIP_SUB_CMD_SPOD_INFO.value
    spod_info = CSuperPodInfo()
    _process_get_super_pod_info(
        dcmi_lib,
        card_id,
        device_id,
        main_cmd,
        sub_cmd,
        ctypes.byref(spod_info),
        ctypes.byref(ctypes.c_uint(ctypes.sizeof(CSuperPodInfo))),
    )
    return spod_info


@validate_process
def _process_dcmi_get_urma_device_cnt(dcmi_lib, card_id, device_id, count):
    return dcmi_lib.dcmi_get_urma_device_cnt(card_id, device_id, count)


def get_urma_device_cnt(card_id, device_id):
    """Get the number of URMA for the device"""
    dcmi_lib = _get_dcmi_lib()
    dcmi_lib.dcmi_get_urma_device_cnt.argtypes = [
        ctypes.c_int,
        ctypes.c_int,
        ctypes.POINTER(ctypes.c_int),
    ]
    dcmi_lib.dcmi_get_urma_device_cnt.restype = ctypes.c_int
    count = ctypes.c_int(0)
    _process_dcmi_get_urma_device_cnt(dcmi_lib, card_id, device_id, ctypes.byref(count))
    return count.value


@validate_process
def _process_dcmi_get_eid_list_by_urma_dev_index(
    dcmi_lib, card_id, device_id, dev_index, eid_ptr, eid_cnt
):
    return dcmi_lib.dcmi_get_eid_list_by_urma_dev_index(
        card_id, device_id, dev_index, eid_ptr, eid_cnt
    )


def get_eid_list_by_urma_dev_index(card_id, device_id, dev_index):
    """Get EID list information for the specified URMA device"""
    dcmi_lib = _get_dcmi_lib()
    dcmi_lib.dcmi_get_eid_list_by_urma_dev_index.argtypes = [
        ctypes.c_int,
        ctypes.c_int,
        ctypes.c_int,
        ctypes.POINTER(EidInfo),
        ctypes.POINTER(ctypes.c_int),
    ]
    dcmi_lib.dcmi_get_eid_list_by_urma_dev_index.restype = ctypes.c_int
    MAX_LEN = 32
    eid_ptr = (EidInfo * MAX_LEN)()
    eid_cnt = ctypes.c_int(0)
    _process_dcmi_get_eid_list_by_urma_dev_index(
        dcmi_lib, card_id, device_id, dev_index, eid_ptr, ctypes.byref(eid_cnt)
    )
    eid_list = []
    for j in range(eid_cnt.value):
        cur_eid_info = eid_ptr[j]
        eid_list.append(
            "".join([format(c, "02x") for c in cur_eid_info.eid[:EID_SIZE]])
        )
    return eid_list


def _process_get_local_id(dcmi_lib, card_id, local_id):
    return dcmi_lib.dcmi_get_device_phyid_from_logicid(card_id, ctypes.byref(local_id))


def get_local_id(card_id):
    """Get local_id by card_id using dcmi_get_device_phyid_from_logicid. In Ascend950, card_id equals logic_id in a broad."""
    dcmi_lib = _get_dcmi_lib()
    dcmi_lib.dcmi_get_device_phyid_from_logicid.argtypes = [
        ctypes.c_int,
        ctypes.POINTER(ctypes.c_int),
    ]
    dcmi_lib.dcmi_get_device_phyid_from_logicid.restype = ctypes.c_int
    card_id = ctypes.c_int(card_id)
    local_id = ctypes.c_int(0)
    _process_get_local_id(dcmi_lib, card_id, local_id)
    return local_id.value


def _process_get_device_board_info(dcmi_lib, card_id, device_id, board_info):
    dcmi_lib.dcmi_get_device_board_info(card_id, device_id, ctypes.byref(board_info))


def get_device_board_info():
    """Get board information, board_id can confirm whether it is a standard card form factor"""
    dcmi_lib = _get_dcmi_lib()
    dcmi_lib.dcmi_get_device_board_info.argtypes = [
        ctypes.c_int,
        ctypes.c_int,
        ctypes.POINTER(CBoardInfo),
    ]
    dcmi_lib.dcmi_get_device_board_info.restype = ctypes.c_int
    card_id = ctypes.c_int(0)
    device_id = ctypes.c_int(0)
    board_info = CBoardInfo()
    _process_get_device_board_info(dcmi_lib, card_id, device_id, board_info)
    return board_info


def _process_get_mainboard_id(dcmi_lib, card_id, device_id, mainboard_id):
    return dcmi_lib.dcmi_get_mainboard_id(
        card_id, device_id, ctypes.byref(mainboard_id)
    )


def get_mainboard_id():
    """Get mainboard_id to determine the number of P interconnects in the current standard card"""
    dcmi_lib = _get_dcmi_lib()
    dcmi_lib.dcmi_get_mainboard_id.argtypes = [
        ctypes.c_int,
        ctypes.c_int,
        ctypes.POINTER(ctypes.c_uint),
    ]
    dcmi_lib.dcmi_get_mainboard_id.restype = ctypes.c_int
    card_id = ctypes.c_int(0)
    device_id = ctypes.c_int(0)
    mainboard_id = ctypes.c_uint(0)
    _process_get_mainboard_id(dcmi_lib, card_id, device_id, mainboard_id)
    return mainboard_id.value


def get_roce_ip_list():
    """Get L3 layer address, currently returns mock address"""
    global _roce_ip_start
    _roce_ip_start += 1
    return [f"1.2.0.{_roce_ip_start}", f"3.4.0.{_roce_ip_start}"]
