#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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
from ascend_fd.utils.i18n import get_fault_entity_details_by_code


class FaultEntity:
    def __init__(self, code: str):
        self.code = code
        cause, description, suggestion = get_fault_entity_details_by_code(code)
        self.attribute = {
            "code": code,
            "cause_zh": cause,
            "description_zh": description,
        }
        if suggestion:
            self.update_attribute({"suggestion_zh": suggestion})

    def update_attribute(self, attr_info: dict):
        """
        Update attribute
        :param attr_info: new attribute dict
        """
        if attr_info:
            self.attribute.update(attr_info)


# kg normal code and attribute
KG_DIAG_NORMAL_ENTITY = FaultEntity("NORMAL_OR_UNSUPPORTED")

# node anomaly normal code and attribute
NODE_DIAG_NORMAL_ENTITY = FaultEntity("NODE_RES_NORMAL")
# npu overload frequency reduction code and attribute
NPU_STATUS_ABNORMAL_ENTITY = FaultEntity("NODE_RES_ABNORMAL_01")
NPU_OVER_TEMPERATURE_ENTITY = FaultEntity("NODE_RES_ABNORMAL_02")
# cpu resource preemption code and attribute
ALL_PROCESS_PREEMPT_FAULT_ENTITY = FaultEntity("NODE_RES_ABNORMAL_03")
SINGLE_PROCESS_PREEMPT_FAULT_ENTITY = FaultEntity("NODE_RES_ABNORMAL_04")
PART_PROCESS_PREEMPT_FAULT_ENTITY = FaultEntity("NODE_RES_ABNORMAL_05")

# network congestion code and attribute
NET_DIAG_NORMAL_ENTITY = FaultEntity("NET_CONGESTION_NORMAL")
NET_LINK_CONGESTION_FAULT_ENTITY = FaultEntity("NET_CONGESTION_ABNORMAL_01")
