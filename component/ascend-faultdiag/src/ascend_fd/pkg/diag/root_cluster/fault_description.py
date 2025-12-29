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
from ascend_fd.utils.i18n import get_fault_description_by_code
from ascend_fd.utils.json_dict import JsonObj


class FaultDescription(JsonObj):
    def __init__(self, code: int, string: str = ""):
        self.code = code
        self.string = string
        self.set_string_by_code(code)

    def set_string_by_code(self, code):
        self.string = get_fault_description_by_code(code)

    def format(self, *args, **kwargs):
        self.string = self.string.format(*args, **kwargs)
        return self


# Root Cluster Diag Description
PART_ERROR_WITH_NO_TIMEOUT = FaultDescription(101)
ALL_NO_ERROR = FaultDescription(102)
ALL_SOCKET_ERROR_WITH_TIMEOUT = FaultDescription(107)
ALL_SOCKET_ERROR_NOT_TIMEOUT = FaultDescription(108)
AI_CPU_NOTIFY_TIMEOUT = FaultDescription(109)
ALL_NOTIFY_ERROR_WITH_TIMEOUT = FaultDescription(110)
ALL_NOTIFY_ERROR_NOT_TIMEOUT = FaultDescription(111)
PART_NOTIFY_ERROR = FaultDescription(112)
FIRST_ERROR_RANK = FaultDescription(113)
INVALID_DEVICE_ERROR = FaultDescription(114)
NO_PLOG_ERROR = FaultDescription(115)
CQE_ERROR = FaultDescription(116)
ALL_INIT_FAILED_WITH_TIMEOUT = FaultDescription(117)
ALL_INIT_FAILED_NOT_TIMEOUT = FaultDescription(118)
PART_INIT_FAILED = FaultDescription(119)
TLS_SWITCH_DIFFERENT = FaultDescription(120)
INIT_FAILED_WITH_NO_CONN_NO_LOG = FaultDescription(121)
ALL_NOTIFY_ERROR_NOT_TIMEOUT_INDEX_ERR = FaultDescription(122)
ALL_NOTIFY_ERROR_NOT_TIMEOUT_TAG_ERR = FaultDescription(123)
ALL_NOTIFY_ERROR_NOT_TIMEOUT_REMOTE_CYCLE = FaultDescription(124)
ALL_NOTIFY_ERROR_NOT_TIMEOUT_REMOTE_LOCAL = FaultDescription(125)
CLUSTER_EXCEPTION_LOCATION_ERROR = FaultDescription(126)
NO_VALID_PLOG_INFO_ERROR = FaultDescription(127)
LACK_OF_BASE_INFO_AFTER_RESUMING_TRAINING = FaultDescription(128)
MINDIE_PULL_KV_ERROR = FaultDescription(129)
LAGGING_ON_WAITING_TIMEOUT_REPORT = FaultDescription(130)
MINDIE_LINK_ERROR = FaultDescription(131)
TRANSPORT_INIT_ERROR = FaultDescription(132)
TRANSPORT_INIT_ERROR_NO_DEVICE_ID = FaultDescription(133)
