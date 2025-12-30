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

import sys
import ast

from bitarray.util import hex2ba, ba2int


# EID physic type flag
EID_TYPE_PHY = 0
# EID logic type flag
EID_TYPE_LOGIC = 1
FE_INTER_CHESSIS = 0
FE_INNER_CHESSIS = 1

EID_LENGTH = 128
NPU_COUNT_IN_A_BOARD = 8
DIE_COUNT_IN_A_NPU = 2
PHY_PORT_COUNT_IN_A_DIE = 9
LOGIC_PORT_COUNT_IN_A_DIE = 2
# When the port number is greater than 180, it is identified as a logical port
LOGIC_PORT_FLAG = 180
# Port field range
PORT_ID_RANGE_START, PORT_ID_RANGE_END, PORT_ID_LENGTH = 120, 128, 8
# Board id field range
BOARD_ID_RANGE_START, BOARD_ID_RANGE_END, BOARD_ID_LENGTH = 128 - 12, 128 - 8, 4
# Chassis id field range
CHESSIS_ID_RANGE_START, CHESSIS_ID_RANGE_END, CHESSIS_ID_LENGTH = 128 - 16, 128 - 12, 4
# Const 233 id field range
UBC_223_RANGE_START, UBC_223_RANGE_END, UBC_223_LENGTH = 128 - 32, 128 - 16, 16
# hex value 0xdfdf is 223 + 223 * 256
UBC_VALUE = 223 + 223 * 256
# Fe id 233 id field range
FE_ID_RANGE_START, FE_ID_RANGE_END, FE_ID_LENGTH = 128 - 74, 128 - 69, 5
# Super pod id 233 id field range
SUPER_POD_ID_RANGE_START, SUPER_POD_ID_RANGE_END, SUPER_POD_ID_LENGTH = (
    128 - 92,
    128 - 76,
    16,
)
BIT_53TH_INDEX, BIT_53TH_VALUE = 128 - 53, 1


def parse_eid(hex_eid_str):
    res = {}
    print(f"----------EID: {hex_eid_str}----------")
    eid = hex2ba(hex_eid_str)
    if len(eid) != EID_LENGTH:
        print("EID length error")
        return {}
    port_id = ba2int(eid[PORT_ID_RANGE_START:PORT_ID_RANGE_END])
    board_id = ba2int(eid[BOARD_ID_RANGE_START:BOARD_ID_RANGE_END])
    chassis_id = ba2int(eid[CHESSIS_ID_RANGE_START:CHESSIS_ID_RANGE_END])
    fe_id = ba2int(eid[FE_ID_RANGE_START:FE_ID_RANGE_END])
    super_pod_id = ba2int(eid[SUPER_POD_ID_RANGE_START:SUPER_POD_ID_RANGE_END])
    print(f"Port ID: {port_id}")
    print(f"Board ID: {board_id}")
    print(f"Chassis ID: {chassis_id}")
    print(f"FE ID: {fe_id}")
    print(f"Super Pod ID: {super_pod_id}")
    res["port_id"] = port_id
    res["board_id"] = board_id
    res["chassis_id"] = chassis_id
    res["fe_id"] = fe_id
    res["super_pod_id"] = super_pod_id
    return res


def get_hex_eid_str(input_str):
    input_str = input_str.strip()
    if input_str.startswith("["):
        return ast.literal_eval(input_str)
    else:
        return input_str


def main(arg_list=None):
    res = []
    if arg_list is None:
        hex_eid_str = get_hex_eid_str(sys.argv[1])
    else:
        hex_eid_str = arg_list
    if isinstance(hex_eid_str, str):
        res.append(hex_eid_str)
    elif isinstance(hex_eid_str, list):
        for i in hex_eid_str:
            res.append(parse_eid(i))
    return res
