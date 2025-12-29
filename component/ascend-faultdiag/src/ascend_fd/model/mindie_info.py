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
from typing import List, Dict

from ascend_fd.utils.json_dict import JsonObj


class IPBaseInfo(JsonObj):
    def __init__(self, local_ip, remote_ip):
        self.local_ip = local_ip
        self.remote_ip = remote_ip


class MindIELinkErrorInfo(IPBaseInfo, JsonObj):
    def __init__(self, local_ip, remote_ip):
        super().__init__(local_ip, remote_ip)


class MindIEPullKVErrorInfo(IPBaseInfo, JsonObj):
    def __init__(self, local_ip, remote_ip):
        super().__init__(local_ip, remote_ip)


class MindIEParseInfo(JsonObj):
    # 未加工、未去重的原始数据
    def __init__(self):
        self.link_error_list: List[MindIELinkErrorInfo] = []
        self.pull_kv_error_list: List[MindIEPullKVErrorInfo] = []


class MindIEParseResult(JsonObj):
    # 原始数据处理后的数据
    def __init__(self):
        self.mindie = True
        self.link_error_info_map: Dict[str, List[str]] = {}
        self.pull_kv_error_map: Dict[str, List[str]] = {}

    def reconstruct_result(self, mindie_parse_info):
        for link_error_info in mindie_parse_info.link_error_list:
            self.link_error_info_map.setdefault(link_error_info.local_ip, []).append(link_error_info.remote_ip)
        for pull_kv_error_info in mindie_parse_info.pull_kv_error_list:
            self.pull_kv_error_map.setdefault(pull_kv_error_info.local_ip, []).append(pull_kv_error_info.remote_ip)


class MindIEDiagResult(JsonObj):
    # 诊断后的数据
    def __init__(self):
        self.link_error_ip_list = []
        self.local_to_remote = {}
        self.remote_to_local = {}
        self.pull_kv_error_ip_list = []
