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
import re
import os
from datetime import datetime, timedelta
from typing import List, Union

from ascend_fd.model.parse_info import PlogBaseInfo, PlogPidParseInfo, PlogErrorInfo, PlogShowLogs, RemoteInfo, \
    TimeoutEvent
from ascend_fd.utils import regular_table
from ascend_fd.utils.constant.str_const import TRANSPORT_INIT_ERROR
from ascend_fd.utils.regular_table import ERROR_CQE_LATEST, ERROR_CQE_LATEST_SPLIT, ERROR_CQE, ERROR_CQE_NEW, \
    ERROR_CQE_SPLIT, ERROR_CQE_NEW_SPLIT
from ascend_fd.utils.tool import safe_write_open, safe_read_open, SHOW_LINES_NUM, get_log_module_and_time
from ascend_fd.pkg.parse.blacklist.blacklist_op import BlackListManager

INVALID_ID = "-1"
INVALID_IP = "0.0.0.0"


class PidFileParser:
    def __init__(self, pid: str, device_ip_map: dict, resuming_training_time: str = regular_table.MIN_TIME,
                 recovery_time: str = regular_table.MIN_TIME):
        self.pid = pid
        self.start_train_time = regular_table.MAX_TIME
        self.end_train_time = regular_table.MIN_TIME
        self.resuming_training_time = resuming_training_time
        self.lagging_time = regular_table.MIN_TIME
        blacklist_manager = BlackListManager()
        self.base_info_parser = BaseInfoParser(device_ip_map, blacklist_manager)
        self.error_parser = ErrorParser(blacklist_manager)
        self.repeat_filter_set = set()
        self.recovery_success_time = recovery_time
        self.init_info_fetched = False

        self.normal_logs = []
        self.error_logs = []

    def get_result(self):
        """
        Get the parse result
        """
        base_info = self.base_info_parser.get_result()
        error_info = self.error_parser.get_result()
        # if there is no rank info and error info, it is invalid pid info
        if not base_info.rank_map and not error_info.first_error_module:
            return None
        plog_pid_parse_info = PlogPidParseInfo()
        plog_pid_parse_info.pid = self.pid
        plog_pid_parse_info.base = base_info
        plog_pid_parse_info.error = error_info
        self._shift_time_by_buffer()
        plog_pid_parse_info.start_train_time = self.start_train_time
        plog_pid_parse_info.end_train_time = self.end_train_time
        plog_pid_parse_info.lagging_time = self.lagging_time
        plog_pid_parse_info.recovery_success_time = self.recovery_success_time
        plog_pid_parse_info.start_resumable_training_time = self.resuming_training_time
        plog_pid_parse_info.plog_parsed_name = f"plog-parser-{self.pid}-{'1' if self.error_logs else '0'}.log"

        show_logs = PlogShowLogs()
        show_logs.error = self.error_logs[:SHOW_LINES_NUM]
        show_logs.normal = self.normal_logs[:SHOW_LINES_NUM]
        plog_pid_parse_info.show_logs = show_logs
        return plog_pid_parse_info

    def save_pid_log(self, output_path: str):
        """
        Save the origin error and useful logs to file.
        """
        if not self.normal_logs and not self.error_logs:
            return
        error_flag = "1" if self.error_logs else "0"
        dst_file = os.path.join(output_path, f"plog-parser-{self.pid}-{error_flag}.log")
        with safe_write_open(dst_file, mode="w+") as file_stream:
            file_stream.writelines(self.normal_logs + self.error_logs)

    def parse_log(self, log_source: Union[list, str]):
        """
        Parse a log of one pid
        """
        if isinstance(log_source, str):
            self._parse_file(log_source)
            return
        for line in log_source:
            try:
                module, log_time = get_log_module_and_time(line)
            except IndexError:
                continue
            self._parse_line(line, module, log_time)

    def _parse_file(self, file_path: str):
        """
        Parse a file of one pid
        """
        if not os.path.isfile(file_path):
            return
        with safe_read_open(file_path, "r", encoding="UTF-8") as file_stream:
            while True:
                line = file_stream.readline()
                if not line:
                    break
                try:
                    module, log_time = get_log_module_and_time(line)
                except IndexError:
                    continue  # If error, mean the log format is incorrect. Skip this line
                self._parse_line(line, module, log_time)

    def _parse_line(self, line, module, log_time):
        """
        Parse log line
        """
        # Use pre-fetched resuming training time when only attr init success exists
        # Manually update base info in n seconds recovery scenario
        if log_time < self.resuming_training_time and self.recovery_success_time == regular_table.MIN_TIME:
            return
        if log_time < self.start_train_time:
            self.start_train_time = log_time
        if log_time > self.end_train_time:
            self.end_train_time = log_time
        # 忽略场景："Process group work %s, seq_num %u dispatch sucess.This error log can be ignored."
        if "error log can be ignored" in line and "Process group work" not in line:
            self.error_parser.re_init()
            self.repeat_filter_set.clear()
            return
        if regular_table.ATTR_INIT_SUCCESS in line and self._info_need_be_updated():
            # reinit base info in multiple attr init success scenario
            self.normal_logs.clear()
            self.repeat_filter_set.clear()
            self.base_info_parser.re_init()
            return
        if regular_table.N_SECOND_RECOVERY_FINISH in line and self._info_need_be_updated():
            # keep base info in the multiple n seconds recovery scenario
            self.normal_logs.clear()
            self.repeat_filter_set.clear()
            return
        if regular_table.LAGGING_INFO_ON_WAITING in line:
            self.lagging_time = max(self.lagging_time, log_time)
        if self._check_repeat_line(line):
            return
        if self.base_info_parser.parse_line(line):
            self._add_origin_log(line)
        if self.error_parser.parse_line(line, module, log_time, self.resuming_training_time,
                                        self.recovery_success_time):
            self._add_origin_log(line)

    def _add_origin_log(self, log_line):
        """
        Add the origin log to dict data
        """
        if log_line.startswith(regular_table.ERROR_ALL) and log_line not in self.error_logs:
            self.error_logs.append(log_line)
            return
        if not log_line.startswith(regular_table.ERROR_ALL) and log_line not in self.normal_logs:
            self.normal_logs.append(log_line)

    def _check_repeat_line(self, line):
        """
        Check and skip the repeat line and
        """
        feature = line.split("[", 2)[-1]
        if feature in self.repeat_filter_set:
            return True
        self.repeat_filter_set.add(feature)
        return False

    def _info_need_be_updated(self):
        """
        Record first occurrence of init info or n seconds recovery
        Return true if the init info needs to be updated
        """
        if not self.init_info_fetched:
            self.init_info_fetched = True
            return False
        return True

    def _shift_time_by_buffer(self):
        """
        Subtract a buffer time to enlarge the time interval to avoid ignorance of errors in initialization
        """
        if self.resuming_training_time == regular_table.MIN_TIME:
            return
        buffer_time_in_seconds = 10
        time_format = "%Y-%m-%d-%H:%M:%S.%f"
        current_time = datetime.strptime(self.resuming_training_time, time_format)
        shifted_time = current_time - timedelta(seconds=buffer_time_in_seconds)
        self.resuming_training_time = shifted_time.strftime(time_format)


class BaseInfoParser:
    TIMEOUT_KEYWORD = {
        regular_table.CONNECT_TIMEOUT: regular_table.CONNECT_TIMEOUT_KEYWORD,
        regular_table.EXEC_TIMEOUT: regular_table.EXEC_TIMEOUT_KEYWORD,
        regular_table.RDMA_TIMEOUT: regular_table.RDMA_TIMEOUT_KEYWORD,
        regular_table.RDMA_RETRY_CNT: regular_table.RDMA_RETRY_CNT_KEYWORD
    }
    DEFAULT_TIMEOUT_SET_KEYWORD = {
        regular_table.CONNECT_TIMEOUT: regular_table.DEFAULT_CONNECT_TIMEOUT_SET_KEYWORD,
        regular_table.EXEC_TIMEOUT: regular_table.DEFAULT_EXEC_TIMEOUT_SET_KEYWORD,
        regular_table.RDMA_TIMEOUT: regular_table.DEFAULT_RDMA_TIMEOUT_SET_KEYWORD,
        regular_table.RDMA_RETRY_CNT: regular_table.DEFAULT_RDMA_RETRY_CNT_SET_KEYWORD
    }
    TIMEOUT_PATTERN = re.compile(regular_table.TIME_OUT)

    def __init__(self, device_ip_map: dict, blacklist_manager: BlackListManager):
        self.device_ip_map = device_ip_map
        self.blacklist_manager = blacklist_manager

        self.logic_device_id = ""
        self.phy_device_id = ""
        self.device_ip = ""
        self.vNic_ip = ""
        self.server_id = ""
        self.rank_map = {}  # {identifier_name: {"rank_id": rank id str, "rank_num": int}}
        self.root_for_identifiers = set()
        self.server_name = ""

        self.timeout_params = {}

    def re_init(self):
        """
        Re init class
        """
        self.__init__(self.device_ip_map, self.blacklist_manager)

    def parse_line(self, line):
        """
        Parse base info from log line
        """
        if regular_table.GET_ROOT_INFO in line and not self.blacklist_manager.is_log_line_need_ignore(line):
            self._parse_root_init_info(line)
            return True
        if regular_table.ENTRY_ROOT_INFO in line and not self.blacklist_manager.is_log_line_need_ignore(line):
            self._parse_entry_init_info(line)
            return True
        if regular_table.RANK_NUM_INFO in line and regular_table.RANK_INFO in line \
                and not self.blacklist_manager.is_log_line_need_ignore(line):
            self._parse_common_init_info(line)
            return True
        if regular_table.TOTAL_RANK_INFO in line and regular_table.SERVER_ID_INFO in line \
                and not self.blacklist_manager.is_log_line_need_ignore(line):
            self._parse_common_init_info(line)
            return True
        for cate, keyword in self.DEFAULT_TIMEOUT_SET_KEYWORD.items():
            default_timeout_key_in_line = regular_table.EXTERNAL_INPUT_KEYWORD in line and keyword in line \
                                          and not self.blacklist_manager.is_log_line_need_ignore(line)
            if default_timeout_key_in_line:
                self._parse_default_timeout_set_info(line, cate)
                return True
        for cate, keyword in self.TIMEOUT_KEYWORD.items():
            timeout_key_in_line = keyword in line and not self.blacklist_manager.is_log_line_need_ignore(line)
            if timeout_key_in_line:
                self._parse_timeout_info(line, cate)
                return True
        if regular_table.SOCKET_VIRTUAL_NIC_IP_INFO in line and regular_table.SOCKET_PHY_ID_INFO in line \
                and not self.blacklist_manager.is_log_line_need_ignore(line):
            self._parse_vNic_ip_phy_device_id(line)
            return True
        return False

    def get_result(self) -> PlogBaseInfo:
        """
        Get the parse result
        """
        plog_base_info = PlogBaseInfo()
        plog_base_info.device_ip = self.device_ip or self.device_ip_map.get(self.phy_device_id or self.logic_device_id,
                                                                            "")
        plog_base_info.vNic_ip = self.vNic_ip
        plog_base_info.logic_device_id = self.logic_device_id
        plog_base_info.phy_device_id = self.phy_device_id
        plog_base_info.server_id = self.server_id
        plog_base_info.root_list = list(self.root_for_identifiers)
        plog_base_info.timeout_param = self.timeout_params
        plog_base_info.rank_map = self.rank_map
        plog_base_info.server_name = self.server_name
        return plog_base_info

    def supplement_base_info(self, device_id: int, server_id: str):
        """
        Supplement lacked base info
        :param device_id: given device id
        :param server_id: given server id to complement
        """
        str_dev_id = str(device_id)
        if not self.logic_device_id and str_dev_id != INVALID_ID:
            self.logic_device_id = str(device_id)
        if not self.phy_device_id and str_dev_id != INVALID_ID:
            self.phy_device_id = str(device_id)
        if not self.server_id:
            self.server_id = server_id

    def supplement_rank_info(self, instance_id: str, rank_map: dict, server_name: str):
        """
        Supplement lacked rank_map
        :param instance_id: given instance id used for identify a special identifier
        :param rank_map: a rank map that is going to update
        :param server_name: input server name
        """
        self.rank_map.update(rank_map)
        self.root_for_identifiers.add(instance_id)
        self.server_name = server_name

    def _parse_vNic_ip_phy_device_id(self, line: str):
        """
        Parse vNic IP and physic device id of device
        """
        self.vNic_ip = filter_single_rank_info(line, regular_table.SOCKET_VIRTUAL_NIC_IP_INFO)
        self.phy_device_id = self.phy_device_id or filter_single_rank_info(line, regular_table.SOCKET_PHY_ID_INFO)

    def _parse_timeout_info(self, line: str, cate: str):
        """
        Parse the timeout info from plog log
        """
        if self.timeout_params.get(cate):
            return
        timeout_re = self.TIMEOUT_PATTERN.search(line)
        if timeout_re:
            self.timeout_params.update({cate: int(timeout_re[1])})

    def _parse_default_timeout_set_info(self, line: str, cate: str):
        """
        Parse through splitting the line to obtain and set timeout info
        The log line is in the following format
        TIMEOUT_PARA set by (default|environment) to [timeout]
        """
        if self.timeout_params.get(cate):
            return
        timeout = filter_single_rank_info(line, "to [")
        if timeout:
            self.timeout_params.update({cate: int(timeout)})

    def _parse_root_init_info(self, line: str):
        """
        Parse and save the Device root init info.
        Log e.g:
        [HCCL_TRACE]HcclGetRootInfo success, take time [5015]us, identifier[*]
        """
        identifier_name = filter_single_rank_info(line, regular_table.IDENTIFIER_INFO)
        if identifier_name:
            self.root_for_identifiers.add(identifier_name)

    def _parse_entry_init_info(self, line: str):
        """
        Parse and save the Device Entry-HcclCommInitRootInfo from plog HCCL logs.
        Log e.g:
        Entry-HcclCommInitRootInfo:ranks[*], rank[*], rootinfo: host ip[*] port[*] nicDeploy[*] identifier[*],
        deviceLogicId[*]
        """
        identifier_name = filter_single_rank_info(line, regular_table.IDENTIFIER_INFO) or \
                          regular_table.DEFAULT_IDENTIFIER
        rank_id = filter_single_rank_info(line, regular_table.RANK_INFO)
        if rank_id:
            self.rank_map.setdefault(identifier_name, dict()).update({"rank_id": rank_id})
        rank_num_str = filter_single_rank_info(line, regular_table.ENTRY_RANKS_INFO)
        if rank_num_str:
            self.rank_map.setdefault(identifier_name, dict()).update({"rank_num": int(rank_num_str)})
        logic_device_id = filter_single_rank_info(line, regular_table.ENTRY_DEVICE_INFO)
        if logic_device_id == INVALID_ID:
            logic_device_id = ""
        self.logic_device_id = self.logic_device_id or logic_device_id

    def _parse_common_init_info(self, line: str):
        """
        Filter and save the Device HcclCommInitRootInfo from plog HCCL logs or Hccl.
        Log e.g:
        - Old info:
        Entry-HcclCommInitRootInfo:ranks[*], rank[*], rootinfo: host ip[*] port[*] nicDeploy[*] identifier[*],
        deviceLogicId[*]
        HcclCommInitRootInfo success,take time [237063]us, rankNum[*], rank[*],  rootInfo  identifier[*],
        server[*], device[*]
        HcclCommInitRootInfo failed, rankNum[*], rank[*], server[*], return[0x0000000005000004], rootInfo
        identifier[*]
        - New info:
        [HCCL_TRACE]SetupAgent rankNum[*], rank[*], rootInfo identifier[*], server[*], logicDevId[*],
        phydevId[*], deviceIp[*]
        [HCCL_TRACE]HcclCommInitRootInfo success,take time [227304]us, rankNum[*], rank[*],rootInfo identifier
        [*],  server[*], logicDevId[*]
        [HCCL_TRACE]HcclCommInitRootInfo failed, return 0x50000001, rankNum[*], rank[*],rootInfo identifier
        [*],  server[*], logicDevId[*]
        - 20240927 New info:
        hcclCommInitInfo:commId[*], rank[*], totalRanks[*], serverId[*], deviceType[*],logicDevId[*], identifier[*]
        """
        identifier_name = filter_single_rank_info(line, regular_table.IDENTIFIER_INFO) or \
                          regular_table.DEFAULT_IDENTIFIER
        rank_id = filter_single_rank_info(line, regular_table.RANK_INFO)
        rank_num_str = filter_single_rank_info(line, regular_table.RANK_NUM_INFO) or \
                       filter_single_rank_info(line, regular_table.TOTAL_RANK_INFO)
        if (identifier_name != regular_table.DEFAULT_IDENTIFIER or regular_table.INIT_ROOT_INFO not in line) \
                and rank_num_str:
            # skip line when not found identifier and use root info to init
            try:
                rank_num = int(rank_num_str)
            except ValueError:
                rank_num = -1
            self.rank_map.setdefault(identifier_name, dict()).update({"rank_num": rank_num, "rank_id": rank_id})
        logic_device_id = filter_single_rank_info(line, regular_table.OLD_DEVICE_INFO) or \
                          filter_single_rank_info(line, regular_table.LOGIC_DEVICE_INFO)
        if logic_device_id == INVALID_ID:
            logic_device_id = ""
        self.logic_device_id = self.logic_device_id or logic_device_id
        phy_device_id = filter_single_rank_info(line, regular_table.PHY_DEVICE_INFO)
        if phy_device_id == INVALID_ID:
            phy_device_id = ""
        self.phy_device_id = self.phy_device_id or phy_device_id
        device_ip = filter_single_rank_info(line, regular_table.DEVICE_IP_INFO)
        if device_ip == INVALID_IP:
            device_ip = ""
        self.device_ip = self.device_ip or device_ip
        server_info = filter_single_rank_info(line, regular_table.SERVER_INFO)
        server_info = server_info.split("%")[0]  # remove the network adapter info
        new_server_id = filter_single_rank_info(line, regular_table.SERVER_ID_INFO)
        self.server_id = self.server_id or server_info or new_server_id


class ErrorParser:
    OTHER_TIMEOUT_KEYWORDS = {
        regular_table.TIMEOUT_FFTS: ["FFTS+ run failed"],
        regular_table.TIMEOUT_NORMAL: ["Wait timeout for sockets recv", "get rasocket timeout", "recv fail"]
    }
    _Transport_error_info_pattern = re.compile(r"remoteIpAddr\[(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})/(\d)]")

    def __init__(self, blacklist_manager: BlackListManager):
        self.first_error_time = regular_table.MAX_TIME  # default max time
        self.first_error_module = ""
        self.blacklist_manager = blacklist_manager
        self.cqe_links = set()
        self.root_info_error_data_cache = RootInfoErrorDataCache()  # root info error need to find the identifier name
        self.notify_error_data_cache = NotifyErrorDataCache()  # notify error data cache, record tag\index\group
        self.socket_error_data_cache = SocketErrorDataCache()  # p2p and socket error data cache, record tag
        self.notify_event_data = dict()  # {"tag_name": error_event} notify need to fix remote rank
        self.timeout_error_events_list = []  # each item is TimeOutEvent
        self.cluster_exception = {}  # key: root device; value: error cause
        self.transport_error_remote = None
        self.transport_init_error_happened = False

    def parse_line(self, line, module, log_time, resuming_training_time, recovery_time):
        """
        Parse error info from plog log
        """
        error_need_ignore = self.blacklist_manager.is_log_line_need_ignore(
            line) or log_time < resuming_training_time or log_time < recovery_time
        if not line.startswith(regular_table.ERROR_ALL) or error_need_ignore:
            return False
        self._parse_error_info(line, module, log_time)
        return True

    def get_result(self) -> PlogErrorInfo:
        """
        Get the parse result
        """
        if self.notify_error_data_cache.exist_data_flag:  # if notify timeout error cache has data, record it
            error_event = self.notify_error_data_cache.format_data()
            self.notify_error_data_cache.re_init()
            self._save_and_refactor_notify_event(error_event)
        all_events = self.timeout_error_events_list + list(self.notify_event_data.values())
        if self.socket_error_data_cache.exist_data_flag:
            all_events.append(self.socket_error_data_cache.format_data())

        plog_error_info = PlogErrorInfo()
        plog_error_info.first_error_module = self.first_error_module
        plog_error_info.first_error_time = self.first_error_time
        plog_error_info.cqe_links = list(self.cqe_links)
        plog_error_info.timeout_error_events_list = [error_event.to_dict() for error_event in all_events]
        plog_error_info.cluster_exception = self.cluster_exception
        plog_error_info.transport_error_remote = self.transport_error_remote
        plog_error_info.transport_init_error_happened = self.transport_init_error_happened
        return plog_error_info

    def re_init(self):
        """
        Re init class
        """
        self.__init__(self.blacklist_manager)

    def _parse_error_info(self, line, err_module, err_time):
        """
        Parse error info from plog log
        """
        if err_time < self.first_error_time:
            self.first_error_time = err_time
            self.first_error_module = err_module
        if not line.startswith(regular_table.ERROR_HCCL):  # extract more only for hccl errors
            return
        self._filter_cqe_error_from_log(line)  # get cqe err info
        self._filter_cluster_exception(line)  # get cluster exception info
        self._filter_root_info_timeout_error_from_log(line, err_time)  # get root info timeout err info
        self._filter_notify_timeout_error_from_log(line, err_time)  # get notify timeout err info
        self._filter_other_timeout_error_from_log(line, err_time)  # get other timeout err info
        self._filter_socket_timeout_error_from_log(line, err_time)  # get socket timeout err info
        self._filter_transport_error_from_log(line)

    def _filter_cqe_error_from_log(self, line: str):
        """
        Filter the cqe error link info from log
        :return:
        """
        for keyword, split_key in zip(
                [ERROR_CQE, ERROR_CQE_NEW],
                [ERROR_CQE_SPLIT, ERROR_CQE_NEW_SPLIT]
        ):
            if keyword in line:
                cqe_ip = filter_single_rank_info(line, split_key)
                if ERROR_CQE_LATEST in line:
                    key_log = line.split(ERROR_CQE_LATEST)[-1]
                    cqe_ip = filter_single_rank_info(key_log, ERROR_CQE_LATEST_SPLIT)
                if not cqe_ip:
                    return
                self.cqe_links.add(cqe_ip)

    def _filter_root_info_timeout_error_from_log(self, line: str, err_time: str):
        """
        Filter the root info timeout error info. Contain connected ranks, identifier and error info
        """
        if ("DisplayConnectionedRank" in line or "DispalyConnectionedRank" in line) and \
                "connected rankinfo" in line:
            for connect_fail_info in line.split("]:")[-1].strip(';\n').split(','):
                rank_id_str = connect_fail_info.split(":")[0].strip().lstrip('[').rstrip(']')
                self.root_info_error_data_cache.add_connected_rank_info(rank_id_str)
            self.root_info_error_data_cache.add_key_info(line)
            return
        for key_id, keyword in enumerate(["topo exchange server get socket timeout", "GetRankBasicInfo from rank[",
                                          "topo exchange agent get socket timeout", "receive from fdhandle failed",
                                          "receive msg length from fdhandle failed"]):
            if keyword in line:
                # the first two errors indicate that the device is a root rank.
                self.root_info_error_data_cache.add_timeout_info(err_time, key_id <= 1)
                self.root_info_error_data_cache.add_key_info(line)
                return
        if "rootInfo identifier[" not in line:  # parse identifier name
            return
        identifier_name = filter_single_rank_info(line, regular_table.IDENTIFIER_INFO)
        if not identifier_name or not self.root_info_error_data_cache.exist_data_flag:
            return
        self.root_info_error_data_cache.add_key_info(line)
        error_info = self.root_info_error_data_cache.format_data(identifier_name)
        self.timeout_error_events_list.append(error_info)

    def _filter_notify_timeout_error_from_log(self, line, err_time):
        """
        Filter the notify timeout error info
        """
        if regular_table.NOTIFY_TASK_EXCEPTION not in line:
            return
        if "taskType[Notify Wait]" in line:  # notify log line:
            if self.notify_error_data_cache.exist_data_flag:
                error_event = self.notify_error_data_cache.format_data()
                self.notify_error_data_cache.re_init()
                self._save_and_refactor_notify_event(error_event)
            tag_name = filter_single_rank_info(line, regular_table.TAG_INFO)
            tag_index = filter_single_rank_info(line, regular_table.NOTIFY_INDEX_INFO)
            self.notify_error_data_cache.add_tag_name(tag_name, tag_index, err_time)
            self.notify_error_data_cache.add_key_info(line)
            return
        # notify remote rank line:
        if "para information" in line:
            remote_rank = filter_single_rank_info(line, regular_table.NOTIFY_REMOTE_RANK_INFO)
            self.notify_error_data_cache.add_remote_rank(remote_rank)
            self.notify_error_data_cache.add_key_info(line)
            return
        # notify group line:
        if "groupRank information" in line:
            identifier = filter_single_rank_info(line, regular_table.NOTIFY_IDENTIFIER_INFO)
            self.notify_error_data_cache.add_identifier_name(identifier)
            self.notify_error_data_cache.add_key_info(line)

    def _filter_other_timeout_error_from_log(self, line: str, err_time: str):
        """
        Filter the common timeout error info
        """
        for error_type, keywords_list in self.OTHER_TIMEOUT_KEYWORDS.items():
            for keyword in keywords_list:
                if keyword not in line:
                    continue
                error_event = TimeoutEvent(error_type=error_type, error_time=err_time, key_info=line)
                self.timeout_error_events_list.append(error_event)

    def _filter_socket_timeout_error_from_log(self, line: str, err_time: str):
        """
        Filter the socket timeout error info
        :param line: one-line log
        :param err_time: time of printing error log
        """
        # socket timeout key log
        if "the connection failure" in line:
            self.socket_error_data_cache.add_error_type(regular_table.TIMEOUT_SOCKET, err_time)
            self.socket_error_data_cache.add_key_info(line)
            return
        # p2p timeout key log
        if "connected p2p timeout" in line and "remote physic id:" in line:
            self.socket_error_data_cache.add_error_type(regular_table.TIMEOUT_SOCKET, err_time)
            self.socket_error_data_cache.add_key_info(line)
            split_info_list = line.split("remote physic id:")
            has_phy_id_len = 2
            if len(split_info_list) != has_phy_id_len:
                return
            phy_device_id = split_info_list[1].strip(" .\n")
            self.socket_error_data_cache.add_remote_info(RemoteInfo("", phy_device_id))
            return
        # get dest_ip(rank_id) and src_ip(rank_id)
        if "| no connect |" not in line and "| time out |" not in line:
            return
        regex = re.compile(r"\|\s{1,5}(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\((\d{1,5})\)")
        device_ip_rank_id = regex.findall(line)
        link_num = 2  # src and dest
        if len(device_ip_rank_id) != link_num:
            return
        self.socket_error_data_cache.add_key_info(line)
        self.socket_error_data_cache.add_remote_info(RemoteInfo(device_ip_rank_id[0][0], ""))

    def _filter_cluster_exception(self, line):
        """
        Filter the cluster exception info
        """
        if regular_table.CLUSTER_EXCEPTION_ROOT_DEVICE not in line:
            return
        root_device = filter_single_rank_info(line, regular_table.CLUSTER_EXCEPTION_ROOT_DEVICE)
        if not root_device:
            return
        root_cause = filter_single_rank_info(line, regular_table.CLUSTER_EXCEPTION_ROOT_CAUSE)
        self.cluster_exception.update({root_device: root_cause})

    def _save_and_refactor_notify_event(self, notify_event):
        """
        Save the notify event and refactor. Some event have same tag name but have different remote rank (digital or
        'local'). Only one event is stored, and digital data is preferentially recorded
        """
        tag_name = notify_event.tag
        if tag_name not in self.notify_event_data:
            self.notify_event_data.update({tag_name: notify_event})
            return
        store_notify_event = self.notify_event_data.get(tag_name)
        if store_notify_event.remote_rank == "local" and notify_event.remote_rank != "local":
            store_notify_event.remote_rank = notify_event.remote_rank
            store_notify_event.key_info = store_notify_event.key_info + notify_event.key_info

    def _filter_transport_error_from_log(self, line):
        if self.transport_error_remote or TRANSPORT_INIT_ERROR not in line:
            return
        self.transport_init_error_happened = True
        match = re.search(self._Transport_error_info_pattern, line)
        if match:
            self.transport_error_remote = RemoteInfo(phy_device_id=match[2], server_ip=match[1])


class RootInfoErrorDataCache:
    """
    Parse and record the root info timeout error for the identifier that obtain later
    """

    def __init__(self):
        self.error_time = regular_table.MAX_TIME  # default max time
        self.connected_ranks_set = set()
        self.root_rank_flag = False
        self.exist_data_flag = False
        self.key_info = []

    def add_connected_rank_info(self, rank_id):
        self.connected_ranks_set.add(rank_id)
        self.exist_data_flag = True

    def add_timeout_info(self, error_time, root_rank_flag):
        self.exist_data_flag = True
        if error_time < self.error_time:
            self.error_time = error_time
            self.root_rank_flag = self.root_rank_flag or root_rank_flag

    def add_key_info(self, line):
        self.key_info.append(line)

    def format_data(self, identifier_name):
        key_info_str = "".join(self.key_info)  # the origin log list has "\n"
        return TimeoutEvent(
            error_type=regular_table.TIMEOUT_ROOT_INFO, error_time=self.error_time, identifier=identifier_name,
            root_flag=self.root_rank_flag, connected_ranks=list(self.connected_ranks_set), key_info=key_info_str
        )

    def re_init(self):
        self.__init__()


class NotifyErrorDataCache:
    """
    Record the notify timeout error info, include operator tag, index, remote rank and identifier group.
    """

    def __init__(self):
        self.tag_name = ""
        self.remote_rank = ""
        self.error_time = regular_table.MAX_TIME  # default max time
        self.key_info = []
        # tag_index and identifier_name info need
        self.tag_index = ""
        self.identifier_name = ""
        self.exist_data_flag = False

    def add_tag_name(self, tag_name, tag_index, error_time):
        self.exist_data_flag = True
        self.tag_name = tag_name
        self.tag_index = tag_index
        if error_time < self.error_time:
            self.error_time = error_time

    def add_remote_rank(self, remote_rank):
        if not self.remote_rank or self.remote_rank == "local":
            self.remote_rank = remote_rank

    def add_identifier_name(self, identifier_name):
        self.identifier_name = identifier_name

    def add_key_info(self, line: str):
        self.key_info.append(line)

    def format_data(self):
        key_info_str = "".join(self.key_info)  # the origin log list has "\n"
        return TimeoutEvent(error_type=regular_table.TIMEOUT_NOTIFY, error_time=self.error_time, tag=self.tag_name,
                            identifier=self.identifier_name, key_info=key_info_str, index=self.tag_index,
                            remote_rank=self.remote_rank)

    def re_init(self):
        self.__init__()


class SocketErrorDataCache:
    """
    Record the socket timeout error info, include operator tag.
    """

    def __init__(self):
        self.error_type = ""
        self.error_time = regular_table.MAX_TIME  # default max time
        self.key_info = []
        self.remote_info: List[RemoteInfo] = []
        self.exist_data_flag = False

    def add_error_type(self, error_type, error_time):
        self.exist_data_flag = True
        self.error_type = error_type
        if error_time < self.error_time:
            self.error_time = error_time

    def add_key_info(self, line: str):
        self.key_info.append(line)

    def add_remote_info(self, info: RemoteInfo):
        self.remote_info.append(info)

    def format_data(self):
        key_info_str = "".join(self.key_info)  # the origin log list has "\n"
        if self.error_type == regular_table.TIMEOUT_SOCKET:
            return TimeoutEvent(error_type=self.error_type, error_time=self.error_time, key_info=key_info_str,
                                remote_info=self.remote_info)
        return TimeoutEvent(error_type=self.error_type, error_time=self.error_time, key_info=key_info_str)


def filter_single_rank_info(line, key):
    """
    Get rank info by split func
    :param line: plog line
    :param key: keyword
    :return: key info
    """
    if key not in line:
        return ""
    key_info_list = line.split(key, 1)
    if len(key_info_list) < 2:
        return ""
    return key_info_list[1].split("]")[0]
