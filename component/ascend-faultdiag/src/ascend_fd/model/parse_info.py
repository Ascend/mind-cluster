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
from dataclasses import dataclass
from itertools import chain
from typing import List, Dict

from ascend_fd.utils import regular_table
from ascend_fd.model.node_info import DeviceInfo
from ascend_fd.utils.json_dict import JsonObj
from ascend_fd.utils.tool import convert_sets_to_lists


class SingleFileParseInfo(JsonObj):
    def __init__(self, container_ip, event_list: List, device_info: DeviceInfo, link_error_info: dict):
        self.container_ip = container_ip
        self.event_list = event_list
        self.device_info = device_info
        self.link_error_info = link_error_info
        self.pull_kv_error_info = {}


class FilesParseInfo(JsonObj):
    def __init__(self, event_list: List, device_info_list: List[DeviceInfo]):
        self.container_ip = ""
        self.event_list = event_list
        self.device_info_list = device_info_list
        self.link_error_info = {}
        self.pull_kv_error_info = {}

    def trans_parse_info(self):
        device_info_dict = {}
        for device_info in self.device_info_list:
            device_info_dict.update({device_info.pid: vars(device_info)})
        return {
            "container_ip": self.container_ip,
            "device_info": device_info_dict,
            "link_error_info": convert_sets_to_lists(self.link_error_info),
            "pull_kv_error_info": convert_sets_to_lists(self.pull_kv_error_info),
        }


class RankInfo(JsonObj):  # 通信域、卡号信息
    def __init__(self, rank_id: str = "", rank_num: int = -1, identifier: str = ""):
        self.rank_id = rank_id
        self.rank_num = rank_num
        self.identifier = identifier


class PlogBaseInfo(JsonObj):  # plog解析出设备基本信息
    def __init__(self, device_ip: str = "",
                 vNic_ip: str = "",
                 logic_device_id: str = "",
                 phy_device_id: str = "",
                 server_id: str = "",
                 root_list: List[str] = None,
                 time_out_param: Dict = None,
                 rank_info_list: List[RankInfo] = None,
                 rank_map: Dict[str, RankInfo] = None,
                 server_name: str = ""):
        self.device_ip = device_ip
        self.vNic_ip = vNic_ip
        self.logic_device_id = logic_device_id
        self.phy_device_id = phy_device_id
        self.server_id = server_id
        self.root_list = root_list or []
        self.timeout_param = time_out_param or {}
        self.rank_info_list = rank_info_list or []
        self.rank_map: Dict[str, RankInfo] = rank_map or {}
        self.server_name = server_name


class RemoteInfo(JsonObj):
    def __init__(self, device_ip: str = "", phy_device_id: str = "", server_ip: str = ""):
        self.device_ip = device_ip
        self.phy_device_id = phy_device_id
        self.server_ip = server_ip


class TimeoutEvent(JsonObj):
    def __init__(self, error_type: str = "",
                 error_time: str = "",
                 key_info: str = "",
                 root_flag: bool = False,
                 connected_ranks: List[str] = None,
                 identifier: str = "",
                 tag: str = "",
                 index: str = "",
                 remote_rank: str = "",
                 remote_info: List[RemoteInfo] = None):
        self.error_type = error_type
        self.error_time = error_time
        self.key_info = key_info
        self.root_flag = root_flag
        self.connected_ranks = connected_ranks or []
        self.identifier = identifier
        self.tag = tag
        self.index = index
        self.remote_rank = remote_rank
        self.remote_info = remote_info or []


class PlogErrorInfo(JsonObj):
    def __init__(self, first_error_module: str = "",
                 first_error_time: str = "",
                 cqe_links: List[str] = None,
                 timeout_error_events_list: List[TimeoutEvent] = None,
                 cluster_exception: Dict = None,
                 transport_error_remote: RemoteInfo = None,
                 transport_init_error_happened: bool = False):
        self.first_error_module = first_error_module
        self.first_error_time = first_error_time
        self.cqe_links = cqe_links or []
        self.timeout_error_events_list = timeout_error_events_list or []
        self.cluster_exception = cluster_exception or {}
        self.transport_error_remote = transport_error_remote
        self.transport_init_error_happened = transport_init_error_happened


class PlogShowLogs(JsonObj):
    def __init__(self, error: List[str] = None, normal: List[str] = None):
        self.error = error or []
        self.normal = normal or []


class PlogPidParseInfo(JsonObj):
    def __init__(self, pid: str = "",
                 base: PlogBaseInfo = None,
                 error: PlogErrorInfo = None,
                 tls_status: str = "",
                 start_train_time: str = regular_table.MAX_TIME,
                 end_train_time: str = regular_table.MIN_TIME,
                 lagging_time: str = regular_table.MIN_TIME,
                 recovery_success_time: str = "",
                 start_resumable_training_time: str = "",
                 plog_parsed_name: str = "",
                 show_logs: PlogShowLogs = None,
                 aicpu_notify_wait_remote: str = ""):
        self.pid = pid
        self.base = base
        self.error = error
        self.tls_status = tls_status
        self.start_train_time = start_train_time
        self.end_train_time = end_train_time
        self.lagging_time = lagging_time
        self.recovery_success_time = recovery_success_time
        self.start_resumable_training_time = start_resumable_training_time
        self.plog_parsed_name = plog_parsed_name
        self.show_logs = show_logs
        self.aicpu_notify_wait_remote = aicpu_notify_wait_remote


class KGParseFilePath(JsonObj):
    def __init__(
            self,
            plog_path: Dict = None,
            device_log_path: Dict = None,
            npu_info_path: List = None,
            train_log_path: List = None,
            host_log_path: List = None,
            host_dmesg_path: List = None,
            host_sysmon_path: List = None,
            host_vmcore_dmesg_path: List = None,
            hisi_logs_path: List = None,
            slog_path: Dict = None,
            noded_log_path: List = None,
            device_plugin_path: List = None,
            volcano_scheduler_path: List = None,
            volcano_controller_path: List = None,
            mindio_log_path: List = None,
            docker_runtime_path: List = None,
            npu_exporter_path: List = None,
            amct_path: List = None,
            mindie_log_path: List = None,
            mindie_cluster_log_path: List = None,
            bmc_app_dump_log_path: List = None,
            bmc_device_dump_log_path: List = None,
            bmc_log_dump_log_path: List = None,
            bmc_log_path: List = None,
            lcne_log_path: List = None,
            bus_log_path: List = None,
            custom_log_list: List = None
    ):
        self.plog_path = plog_path or {}
        self.device_log_path = device_log_path or {}
        self.npu_info_path = npu_info_path or []
        self.train_log_path = train_log_path or []
        self.host_log_path = host_log_path or []
        self.host_dmesg_path = host_dmesg_path or []
        self.host_sysmon_path = host_sysmon_path or []
        self.host_vmcore_dmesg_path = host_vmcore_dmesg_path or []
        self.hisi_logs_path = hisi_logs_path or []
        self.slog_path = slog_path or {}
        self.noded_log_path = noded_log_path or []
        self.device_plugin_path = device_plugin_path or []
        self.volcano_scheduler_path = volcano_scheduler_path or []
        self.volcano_controller_path = volcano_controller_path or []
        self.mindio_log_path = mindio_log_path or []
        self.docker_runtime_path = docker_runtime_path or []
        self.npu_exporter_path = npu_exporter_path or []
        self.amct_path = amct_path or []
        self.mindie_log_path = mindie_log_path or []
        self.mindie_cluster_log_path = mindie_cluster_log_path or []
        self.bmc_app_dump_log_path = bmc_app_dump_log_path or []
        self.bmc_device_dump_log_path = bmc_device_dump_log_path or []
        self.bmc_log_dump_log_path = bmc_log_dump_log_path or []
        self.bmc_log_path = bmc_log_path or []
        self.lcne_log_path = lcne_log_path or []
        self.bus_log_path = bus_log_path or []
        self.custom_log_list = custom_log_list or []

    def get_all_path(self):
        all_path_list = []
        for value in vars(self).values():
            if not value:
                continue
            if isinstance(value, list):
                all_path_list.extend(value)
            elif isinstance(value, dict):
                all_path_list.extend(chain(*value.values()))
        return all_path_list


@dataclass
class SuperPodInfo:
    """
    Parse Config
    """
    server_index: int
    super_pod_id: int
    super_pod_size: int


@dataclass
class LCNInfo:
    """
    Parse Config
    """
    level: int
    switch_id: int
    config_name: str
