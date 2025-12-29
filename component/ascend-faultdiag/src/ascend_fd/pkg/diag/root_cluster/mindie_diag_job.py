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
from ascend_fd.model.cfg import DiagCFG


def diag_link_error(link_error_info_map, local_to_remote_link, remote_to_local_link):
    """
    Init diag_link_error
    :param
    """
    link_error_ip_set = set()
    for ip, ip_list in link_error_info_map.items():
        for sub_ip in ip_list:
            if sub_ip not in link_error_info_map.keys():
                link_error_ip_set.add(sub_ip)
                remote_to_local_link.setdefault(sub_ip, set()).add(ip)
                continue
            if ip in link_error_info_map.get(sub_ip):
                link_error_ip_set.add(ip)
                link_error_ip_set.add(sub_ip)
                local_to_remote_link.setdefault(ip, set()).add(sub_ip)
                local_to_remote_link.setdefault(sub_ip, set()).add(ip)
    return link_error_ip_set


class MindIEDiagWorker:

    def __init__(self, cfg: DiagCFG = None):
        """
        Init mindie diag job
        :param cfg: Diag Config
        :param
        """
        self.cfg = cfg

    def start_job(self):
        """
        Start mindie diag job
        """
        mindie_parse_result = self.cfg.parsed_saver.mindie_parse_result
        mindie_diag_result = self.cfg.parsed_saver.mindie_diag_result

        self.cfg.parsed_saver.mindie_diag_result.link_error_ip_list = list(
            diag_link_error(mindie_parse_result.link_error_info_map, mindie_diag_result.local_to_remote,
                            mindie_diag_result.remote_to_local))
        self.diag_pull_kv_error()

    def diag_pull_kv_error(self):
        """
        Init diag_pull_kv_error
        :param
        """
        pull_kv_error_ip_set = set()
        for ip, sub_ip_list in self.cfg.parsed_saver.mindie_parse_result.pull_kv_error_map.items():
            pull_kv_error_ip_set.add(ip)
            for sub_ip in sub_ip_list:
                pull_kv_error_ip_set.add(sub_ip)
        self.cfg.parsed_saver.mindie_diag_result.pull_kv_error_ip_list = list(pull_kv_error_ip_set)
