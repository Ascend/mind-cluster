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
import logging
from collections import Counter
from datetime import datetime
from dataclasses import dataclass, field
from typing import List, Dict

from ascend_fd.model.diag_info import RCDiagResult
from ascend_fd.model.parse_info import TimeoutEvent, PlogBaseInfo, PlogErrorInfo, PlogShowLogs
from ascend_fd.utils import regular_table
from ascend_fd.utils.constant.num_util import TEN_THOUSAND
from ascend_fd.utils.constant.str_const import SUPER_POD_SCENE, AI_CPU
from ascend_fd.utils.i18n import get_label_for_language
from ascend_fd.utils.status import InfoNotFoundError, InnerError
from ascend_fd.model.node_info import FaultFilterTime
from ascend_fd.pkg.parse.root_cluster.parser import RemoteInfo
from ascend_fd.pkg.diag.root_cluster import fault_description
from ascend_fd.pkg.diag.message import (MULTI_RANK_NOTE_MSG, UNKNOWN_ROOT_ERROR_RANK, SOME_DEVICE_FAILED,
                                        NO_GROUP_RANK_INFO_NOTE)
from ascend_fd.model.cfg import DiagCFG

rc_logger = logging.getLogger("ROOT_CLUSTER")
lb = get_label_for_language()
NEGATIVE_ONE = "-1"
PART = lb.partial
ALL = lb.all_


class Device:
    """
    This class is used to store information about a single device.
    Includes pid, worker name, server ID, device ID, tls status...
    """

    def __init__(self, pid=NEGATIVE_ONE, worker_name="worker-NA", device_table=None):
        """
        Device Info
        :param pid: pid, default '-1'
        :param worker_name: worker name, default 'worker-NA'
        :param device_table: the device table instance for this diag job, default None
        """
        # base info
        self.pid = pid
        self.worker_name = worker_name
        self.device_table = device_table
        self.server_id = ""
        self.device_ip = ""
        self.logic_device_id = ""
        self.phy_device_id = ""
        self.tls_status = ""
        self.log_file_path = ""
        self.normal_logs_show = []
        self.error_logs_show = []
        # save device rank id in each identifier. e.g {identifier_name: {rank_num: xxx, rank_id: xxx}}
        self.identifier_map = dict()

        # error info:
        self.first_error_module = ""
        self.first_error_time = ""
        self.timeout_error_events_list = []
        self.transport_error_remote = None
        self.transport_init_error_happened = False
        # indicates whether a timeout error occurs first.
        # error type contain: socket, p2p, notify, ffts+, root_info init. Not contain normal
        self.first_timeout_error_event = None
        self.timeout_error_map = dict()

    def __hash__(self):
        return hash(self.pid + self.worker_name)

    def __eq__(self, other):
        if isinstance(other, self.__class__):
            return hash(self) == hash(other)
        return False

    def __str__(self):
        if self.worker_name == "worker-NA":
            return "Unknown Device"
        if self.device_id == NEGATIVE_ONE and self.device_table:
            for identifier_name in self.identifier_map.keys():
                device_id = self.device_table.get_lack_device_id(identifier_name, self.worker_name, self.pid)
                return f"{self.worker_name} device-{device_id}"
            return f"{self.worker_name} device-Unknown"
        return f"{self.worker_name} device-{self.device_id}"

    def __repr__(self):
        return self.__str__()

    @property
    def device_id(self):
        """
        Device id. Use phyDeviceID first, then logicDeviceID
        :return: device id
        """
        return self.phy_device_id or self.logic_device_id or NEGATIVE_ONE

    @property
    def err_time(self):
        """
        Rank first err time
        :return: first err time
        """
        return self.first_error_time or regular_table.MAX_TIME

    def is_error(self):
        """
        Determine if this device contains error logs
        :return: bool
        """
        return self.first_error_module and self.first_error_time != regular_table.MAX_TIME

    def record_identifier_instance(self, identifier_instance):
        """
        Save the identifier instance that related to itself
        :param identifier_instance: the identifier instance
        """
        self.identifier_map.update({identifier_instance.name: identifier_instance})

    def update_base_info(self, base_info: PlogBaseInfo):
        """
        Update the base device info from base info dict
        :param base_info: the base device info
        """
        if not base_info:
            return
        self.logic_device_id = base_info.logic_device_id
        self.phy_device_id = base_info.phy_device_id
        self.device_ip = base_info.device_ip
        self.server_id = base_info.server_id

    def update_error_info(self, error_info: PlogErrorInfo, resuming_training_time):
        """
        Update the error device info from base info dict
        :param error_info: the device error info
        :param resuming_training_time: the resuming training starting time after breakpoint
        """
        if not error_info:
            return
        self.first_error_time = error_info.first_error_time if error_info.first_error_time else regular_table.MAX_TIME
        self.first_error_module = error_info.first_error_module
        self.transport_error_remote = error_info.transport_error_remote
        self.transport_init_error_happened = error_info.transport_init_error_happened
        for timeout_event in error_info.timeout_error_events_list:
            if not timeout_event.error_time or timeout_event.error_time < resuming_training_time:
                continue
            self.timeout_error_events_list.append(TimeoutEvent.from_dict(timeout_event.to_dict()))
        if error_info.cqe_links:
            self.device_table.cqe_link_map.update({self: error_info.cqe_links})
        # find the first timeout error
        if not self.timeout_error_events_list:
            return
        self.timeout_error_events_list.sort(key=lambda event: event.error_time)
        for single_timeout_event in self.timeout_error_events_list:
            # if first is normal and have other type error, use the other type error
            if not self.first_timeout_error_event \
                    or self.first_timeout_error_event.error_type == regular_table.TIMEOUT_NORMAL:
                self.first_timeout_error_event = single_timeout_event
            if not single_timeout_event.identifier:
                if single_timeout_event.error_type == regular_table.TIMEOUT_NOTIFY:
                    self.device_table.no_group_rank_information_flag = True
                single_timeout_event.identifier = self.get_identifier_name_from_tag(single_timeout_event.tag)
            if single_timeout_event.identifier == regular_table.DEFAULT_IDENTIFIER:
                # if the identifier name is NA, use the max identifier name
                single_timeout_event.identifier = self.device_table.max_rank_num_identifier.get("identifier_name")
            # record timeout info for identifier and type
            index_field = (single_timeout_event.identifier, single_timeout_event.error_type)
            if index_field not in self.timeout_error_map \
                    or self.timeout_error_map.get(index_field).error_time > single_timeout_event.error_time:
                self.timeout_error_map.update({index_field: single_timeout_event})

    def get_identifier_name_from_tag(self, tag_name):
        """
        Get the identifier name from tag name. Mainly used in notify timeout error, other use default identifier
        :param tag_name: the tag name
        :return: identifier_name
        """
        if not tag_name:
            return regular_table.DEFAULT_IDENTIFIER
        for identifier_name in self.device_table.identifier_dict.keys():
            if identifier_name in tag_name:
                return identifier_name
        return regular_table.DEFAULT_IDENTIFIER

    def update_show_log(self, show_logs: PlogShowLogs):
        """
        Update the log of this device
        :param show_logs: record log
        """
        self.normal_logs_show.extend(show_logs.normal)
        self.error_logs_show.extend(show_logs.error)

    def is_socket_timeout_happened(self):
        """
        Check the socket timeout happen flag
        """
        for timeout_error_event in self.timeout_error_events_list:
            if timeout_error_event.error_type == regular_table.TIMEOUT_SOCKET:
                return True
        return False


UNKNOWN_DEVICE = Device()


class RankDevice(Device):
    """
    Used when root device only get the rank_id
    """

    def __init__(self, identifier, rank_id):
        super().__init__()
        self.identifier = identifier
        self.rank_id = rank_id

    def __hash__(self):
        return hash(self.identifier + self.rank_id)

    def __eq__(self, other):
        if isinstance(other, self.__class__):
            return hash(self) == hash(other)
        return False

    def __str__(self):
        return f"{self.identifier} rank-{self.rank_id}"


class Identifier:
    """
    This class is used to store information about a Identifier.
    Includes Identifier name, rank num, root device and rank id list.
    """
    RECORD_RANK_NUM_LINE = 8  # a log line record 8 rank_ids

    def __init__(self, identifier_name=regular_table.DEFAULT_IDENTIFIER, rank_num=int(NEGATIVE_ONE)):
        """
        Worker Info
        :param identifier_name: identifier name, default 'identifier-NA'
        :param rank_num: rank num, default '-1'
        """
        self.name = identifier_name
        self.rank_num = rank_num
        self.device_rank_id_map = {}  # e.g {device_instance: rank_id}
        self.rank_id_device_map = {}  # e.g {rank_id: device_instance}
        self.root_device = Device()

    def __hash__(self):
        return hash(self.name)

    def get_rank_id_list(self):
        """
        Get rank id list from device_rank_id_map
        """
        return list(self.device_rank_id_map.values())

    def get_device_list(self):
        """
        Get device list from device_rank_id_map
        """
        return list(self.device_rank_id_map.keys())

    def update_device(self, device_instance: Device, rank_id):
        self.device_rank_id_map.update({device_instance: rank_id})
        self.rank_id_device_map.update({rank_id: device_instance})
        device_instance.record_identifier_instance(self)

    def check_integrity_of_rank(self):
        """
        Check the integrity of the rank ID in the communication identifier. Use log to record result
        """
        lost_rank_list = []
        for rank_id in range(self.rank_num):
            if str(rank_id) not in self.get_rank_id_list():
                lost_rank_list.append(rank_id)
            if len(lost_rank_list) == self.RECORD_RANK_NUM_LINE:
                rc_logger.warning("Not found these rank id %s in communication identifier[%s].",
                                  lost_rank_list, self.name)
                lost_rank_list.clear()
        if lost_rank_list:
            rc_logger.warning("Not found these rank id %s in communication identifier[%s].", lost_rank_list, self.name)


class DeviceTable:
    """
    This class use to save the device, worker, server, identifier relationship.
    (worker_name,pid) 1<----->1 device
    server            1<----->* device
    worker            1<----->1 server
    identifier        1<----->* device
    """

    def __init__(self):
        """
        DeviceTable: Use to store device information for the entire training.
        """
        self.worker_list = list()
        self.max_rank_num_identifier = {
            "rank_num": int(NEGATIVE_ONE), "identifier_name": regular_table.DEFAULT_IDENTIFIER
        }
        self.identifier_dict: Dict[str, Identifier] = {}

        self.device_map = dict()
        self.server_device_map = dict()
        self.device_ip_map = dict()
        self.device_id_instance_map = dict()
        self.vNic_ip_dev_id_relation = dict()

        self.err_device = list()
        self.no_err_device = list()

        self.cqe_link_map = {}
        self.timeouts = {regular_table.CONNECT_TIMEOUT: 120, regular_table.EXEC_TIMEOUT: 1800}
        self.no_group_rank_information_flag = False

        self.cluster_exception = {}
        # 同节点所有卡设备字典
        self.worker_name_to_device_map = {}
        # identifier+rank to device_ip
        self.identifier_rank_to_device_ip = {}
        # ai cpu notify wait relation
        self.aicpu_notify_wait_relation = {}
        self.transport_init_error_flag = False

    def update_max_identifier(self, identifier_name: str, rank_num: int):
        """
        Update the max rank num identifier info
        :param identifier_name: identifier name info
        :param rank_num: the rank nums in the identifier
        """
        if rank_num > self.max_rank_num_identifier.get("rank_num", int(NEGATIVE_ONE)):
            self.max_rank_num_identifier["rank_num"] = rank_num
            self.max_rank_num_identifier["identifier_name"] = identifier_name

    def update_hccl_cluster_exception(self, cluster_exception: dict):
        """
        Update the cluster exception info
        :param cluster_exception: cluster root device and root cause from hccl log line
        """
        if not cluster_exception:
            return
        self.cluster_exception.update(cluster_exception)

    def add_device(self, device: Device):
        """
        Add device and save device, sever and device relationship.
        device map: {"worker_name-pid": Device}
        server device map: {"server_id-device_id": Device}
        device ip map: {device_ip: Device}
        :param device: Device instance
        """
        self.worker_name_to_device_map.setdefault(device.worker_name, []).append(device)
        self.device_map.update({(device.worker_name, device.pid): device})
        if device.server_id and device.device_id:
            self.server_device_map.update({(device.server_id, device.device_id): device})
        if device.phy_device_id:
            self.device_id_instance_map.update({(device.worker_name, device.phy_device_id): device})
        if device.device_ip:
            self.device_ip_map.update({device.device_ip: device})
        if device.is_error():
            self.err_device.append(device)
        else:
            self.no_err_device.append(device)

    def get_device_by_server_device_id(self, server_id: str, device_id: str) -> Device:
        """
        Get device by server and device
        :param server_id: server id
        :param device_id: device id
        :return: Device instance
        """
        return self.server_device_map.get((server_id, device_id), Device())

    def get_device_by_device_ip(self, device_ip: str) -> Device:
        """
        Get device by device ip
        :param device_ip: device ip
        :return: Device instance
        """
        return self.device_ip_map.get(device_ip, Device())

    def get_lack_device_id(self, identifier_name, worker_name, pid):
        """
        Get the rank id by other rank
        :param identifier_name: identifier name
        :param worker_name: the worker name
        :param pid: the pid in plog
        :return: rank id
        """
        device = self.device_map.get((worker_name, pid), Device())
        identifier_instance = self.identifier_dict.get(identifier_name, Identifier())
        for store_device in identifier_instance.get_device_list():
            if device == store_device or device.worker_name != store_device.worker_name:
                continue
            store_device_rank_id = identifier_instance.device_rank_id_map.get(store_device, NEGATIVE_ONE)
            if store_device_rank_id != NEGATIVE_ONE and store_device.device_id != NEGATIVE_ONE:
                device_rank_id = identifier_instance.device_rank_id_map.get(device, NEGATIVE_ONE)
                if device_rank_id == NEGATIVE_ONE:
                    continue
                device_id = str(int(store_device.device_id) + (int(device_rank_id) - int(store_device_rank_id)))
                device.phy_device_id = device_id
                return device_id
        return "Unknown"

    def get_device_list_of_identifier(self, identifier_name):
        """
        Get device list of identifier
        :param identifier_name: identifier name
        :return: device list
        """
        return self.identifier_dict.get(identifier_name, Identifier()).get_device_list()

    def update_timeout(self, timeout_param: dict):
        """
        Update the timeout param
        :param timeout_param: timeout param value dict
        """
        for key in self.timeouts.keys():
            timeout_val_str = timeout_param.get(key)
            if timeout_val_str:
                self.timeouts.update({key: int(timeout_val_str)})

    def get_timeout(self, key: str):
        """
        Get timeout param by keyword
        :param key: keyword
        :return: timeout value
        """
        if self.timeouts.get(key) is None:
            rc_logger.error("The timeouts parameters don't contain %s.", key)
            raise InfoNotFoundError(f"The timeouts parameters don't contain {key}.")
        return self.timeouts.get(key)

    def get_transport_init_error_relation(self):
        relation_map = {}
        for device in self.server_device_map.values():
            if device.transport_init_error_happened:
                self.transport_init_error_flag = device.transport_init_error_happened
            remote_info = device.transport_error_remote
            if not remote_info or not remote_info.server_ip or not remote_info.phy_device_id:
                continue
            remote_device = self.get_device_by_server_device_id(remote_info.server_ip, remote_info.phy_device_id)
            if remote_device == UNKNOWN_DEVICE or not remote_device.device_ip:
                continue
            relation_map[device.device_ip] = remote_device.device_ip
        return relation_map


class BaseChecker:
    root_devices = [Device()]
    first_error_device = None
    last_error_device = None
    fault_description = None
    note_msg = []
    device_links = []
    remote_links = []
    mindie_error_device = []

    def __init__(self, device_table: DeviceTable):
        """
        Base rc error checker
        :param device_table: Device Table
        """
        self.device_table = device_table
        self.root_devices = [Device()]
        self.first_error_device = None
        self.last_error_device = None
        self.fault_description = None
        self.note_msg = []
        self.device_links = []
        self.remote_links = []

    def check(self):
        """
        Check rc error
        """
        pass

    def format_output(self, resuming_training_time=regular_table.MIN_TIME,
                      start_train_time=regular_table.MIN_TIME,
                      end_train_time=regular_table.MAX_TIME,
                      scene=None,
                      board_sn_exist=False) -> RCDiagResult:
        """
        Format the diag result
        :param resuming_training_time: the last resuming training time
        :param start_train_time: start train log time
        :param end_train_time: end train log time
        :return: rc diag result
        :param scene: hold "super_pod" only if in super pod scene, otherwise None
        :param board_sn_exist: board sn exist
        """
        self.root_devices.sort(key=lambda device: device.err_time)
        # when all ranks error, only print 'All Rank' info
        all_root = len(self.root_devices) == self.device_table.max_rank_num_identifier.get("rank_num")
        if all_root:
            detect_workers_devices = {worker: [] for worker in self.device_table.worker_list}
        else:
            detect_workers_devices = self._get_detect_worker_device_map()
        # fault filter time used for further filter in knowledge graph diagnosis
        filter_time = FaultFilterTime(regular_table.MIN_TIME, regular_table.MAX_TIME)
        # use start train time and end train time in the super pod scene
        # note such a time segment is smaller than the criteria of knowledge graph parse
        if scene == SUPER_POD_SCENE or board_sn_exist:
            filter_time.start_train_time = start_train_time
            filter_time.end_train_time = end_train_time
        # in resuming the training scene, use resuming training time to filter later in knowledge graph diagnosis
        if resuming_training_time != regular_table.MIN_TIME:
            filter_time.start_train_time = resuming_training_time

        root_cause_device = [str(device) for device in self.root_devices] if not all_root else ["ALL Device"]
        return RCDiagResult(
            analyze_success=True,
            fault_description=self.fault_description,
            root_cause_device=root_cause_device,
            device_link=self.device_links,
            remote_link=' -> '.join(self.remote_links),
            first_error_device=self.first_error_device,
            last_error_device=self.last_error_device,
            note_msgs=self.note_msg,
            fault_filter_time=filter_time,
            fault_description_list=[],
            mindie_error_device=self.mindie_error_device,
            show_device_info=self._add_original_plog(),
            detect_workers_devices=detect_workers_devices
        )

    def _get_detect_worker_device_map(self):
        detect_workers_devices = {}
        for device in self.root_devices:
            if device.worker_name == "worker-NA":
                continue
            detect_workers_devices.setdefault(device.worker_name, []).append(device.device_id)
        if len(self.root_devices) > 1:
            # Multiple root cluster may be diagnosed.
            self.note_msg.append(MULTI_RANK_NOTE_MSG)
        if not detect_workers_devices:
            # Device() means unknown device. Don't find root devices.
            self.note_msg.append(UNKNOWN_ROOT_ERROR_RANK)
            detect_workers_devices = {worker: [] for worker in self.device_table.worker_list}
        return detect_workers_devices

    def _add_original_plog(self) -> dict:
        """
        Add plog to the diag result dict
        :return: show the first device info dict
        """
        device_info = dict()
        show_device = self.root_devices[0]  # first root device
        device_type = "first_root_device"
        if isinstance(show_device, RankDevice):
            # RankDevice only get the rank_id and does not record plog
            return device_info
        if show_device.worker_name == 'worker-NA':
            show_device = self.first_error_device  # first error device
            device_type = "first_error_device"
        if not show_device:
            # there is no first root device and no first error device
            return device_info
        show_log = show_device.error_logs_show or show_device.normal_logs_show
        if show_log:
            device_info.update({
                "device_type": device_type,
                "device": str(show_device),
                "plog_file_path": show_device.log_file_path,
                "error_log": "".join(show_log)  # the origin log list has "\n"
            })
        return device_info


class RemoteRelation:
    MAX_TRAVERSED_NUM = 100000

    def __init__(self):
        self.remote_relation = dict()
        self.reverse_remote_relation = dict()
        # Nodes with an in-degree greater than 1
        self.in_degree_greater_than_one_set = set()
        self.dst_device_is_local_set = set()

        self.traversed_device_map = dict()
        self.traversed_num = 0
        self.remote_links = list()
        self.cycle_or_last_devices = []
        self.last_unknown_device = Device()

    def add_remote_relation(self, src_device: Device, dst_device: Device):
        """
        Add the relationship between the local device(src_device) and the remote device(dst_device)
        :param src_device: local device
        :param dst_device: remote device
        """
        if dst_device is None:
            return
        if src_device == dst_device:
            self.dst_device_is_local_set.add(src_device)
            return
        if dst_device not in self.remote_relation.get(src_device, []):
            self.remote_relation.setdefault(src_device, []).append(dst_device)

        if src_device not in self.reverse_remote_relation.get(dst_device, []):
            self.reverse_remote_relation.setdefault(dst_device, []).append(src_device)

        if len(self.reverse_remote_relation.get(dst_device, [])) > 1:
            self.in_degree_greater_than_one_set.add(dst_device)

    def check_notify_remote_relation(self):
        """
        Check the relationship between the devices at both ends when the notify times out
        :return: fault description, root devices, remote links
        """
        # filtering device that remote to local but have zero in-degree
        remote_local_devices = list(self.dst_device_is_local_set & set(self.reverse_remote_relation.keys()))
        # No loop at the end, remote to local
        if remote_local_devices:
            self._traverse_remote_cycle(remote_local_devices[0], self.reverse_remote_relation)
            self.remote_links.reverse()
            return fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT_REMOTE_LOCAL, remote_local_devices, self.remote_links
        # loop at the end, start from a device with in-degree greater than 1
        if self.in_degree_greater_than_one_set:
            self._traverse_remote_cycle(self.in_degree_greater_than_one_set.pop(), self.remote_relation)
            return fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT_REMOTE_CYCLE, [Device()], self.remote_links
        # all device in the loop, all device with in-degree is 1
        if self.remote_relation.keys() and all([len(dst_list) == 1 for dst_list in self.remote_relation.values()]):
            self._traverse_remote_cycle(list(self.remote_relation.keys())[0], self.remote_relation)
            return fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT_REMOTE_CYCLE, [Device()], self.remote_links
        return None, [], []

    def check_socket_remote_relation(self, first_device, parsed_saver):
        """
        Check the relationship between the devices at both ends when the socket times out
        :param first_device: local device
        :param parsed_saver: parsed saver
        :return: root devices, remote links, description msg
        """
        self._traverse_remote_cycle(first_device, self.remote_relation)
        # 根因节点：成环互等或等待链末端的设备
        root_devices = list(set(self.cycle_or_last_devices))
        if str(UNKNOWN_DEVICE) not in self.remote_links:
            return root_devices, self.remote_links

        # 找不到对端设备
        self.remote_links.remove("Unknown Device")
        # 找跨实例的device ip（A2场景）；找跨实例的 vNic ip（A3场景）
        worker_name, device_id = parsed_saver.infer_groups_device_map.get(self.last_unknown_device.device_ip, ("", ""))
        if parsed_saver.infer_task_flag and worker_name and device_id:
            self.remote_links.append("{} device-{}".format(worker_name, device_id))
            return root_devices, self.remote_links

        # 都找不到直接打印对端设备ip
        self.remote_links.append(self.last_unknown_device.device_ip)
        return root_devices, self.remote_links

    def _traverse_remote_cycle(self, start_device: Device, relation_map):
        """
        Traverse the waiting links and check whether the links are cyclic
        :param start_device: started device
        :param relation_map: map of the waiting relationship
        """
        cur_device = start_device
        traversed_num = 0
        first_appear_index = -1
        while True:
            if traversed_num > self.MAX_TRAVERSED_NUM:
                rc_logger.warning('notify or socket remote traversal times exceeds MAX_TRAVERSED_NUM, '
                                  'the traversal stops.')
                break
            cur_device_str = str(cur_device)
            self.remote_links.append(cur_device_str)
            if cur_device == UNKNOWN_DEVICE:
                self.last_unknown_device = cur_device
                break
            if cur_device_str in self.traversed_device_map:
                first_appear_index = self.traversed_device_map.get(cur_device_str)
                break
            self.cycle_or_last_devices.append(cur_device)
            self.traversed_device_map[cur_device_str] = traversed_num
            remote_device_list = relation_map.get(cur_device, [])
            if not remote_device_list:
                break
            # 如果对端设备有多个，优先找能确定的device
            known_remote_device_list = list(filter(lambda device: device != UNKNOWN_DEVICE, remote_device_list))
            cur_device = known_remote_device_list[0] if known_remote_device_list else remote_device_list[0]
            traversed_num += 1
        self.cycle_or_last_devices = self.cycle_or_last_devices[first_appear_index:]


@dataclass
class TimeoutErrorOfIdentifier:
    """
    Used to save the 5 type timeout error info in an identifier
    """
    identifier = Identifier()
    error_type = ""
    first_time = '9999-12-31-23:59:59.999999'
    last_time = '1999-12-31-23:59:59.999999'
    first_device: Device = field(default_factory=Device)
    last_device: Device = field(default_factory=Device)
    error_devices: set = field(default_factory=set)
    remote_relation: RemoteRelation = field(default_factory=RemoteRelation)

    @staticmethod
    def get_remote_device(remote_info: RemoteInfo, device: Device, device_table: DeviceTable):
        """
        Get the remote device
        :param remote_info: timeout events
        :param device: device on which a timeout evnet occurs
        :param device_table: Device Table
        """
        if remote_info.device_ip:
            # 根据device_ip找对端设备
            remote_device = device_table.device_ip_map.get(remote_info.device_ip)
            if remote_device:
                return remote_device

            # 根据vNic_ip找对端设备
            phy_device_id = device_table.vNic_ip_dev_id_relation.get((device.worker_name, remote_info.device_ip), "")
            remote_device = device_table.device_id_instance_map.get((device.worker_name, phy_device_id))
            if remote_device:
                return remote_device

        if remote_info.phy_device_id:
            # 根据phy_device_id找对端设备
            remote_device = device_table.device_id_instance_map.get((device.worker_name, remote_info.phy_device_id))
            if remote_device:
                return remote_device
            # 机内p2p超时，找不到，直接为当前worker下device id为phy_device_id的设备
            remote_device = Device(worker_name=device.worker_name)
            remote_device.phy_device_id = remote_info.phy_device_id
            return remote_device

        # 未找到对端设备，保存未知设备的device_ip
        remote_device = Device()
        remote_device.device_ip = remote_info.device_ip
        return remote_device

    def update_event(self, timeout_event: TimeoutEvent, device: Device, device_table: DeviceTable):
        """
        Update the timeout event info from each
        :param timeout_event: timeout events
        :param device: device on which a timeout evnet occurs
        :param device_table: Device Table
        """
        self.error_devices.add(device)
        if self.first_time > timeout_event.error_time:
            self.first_time = timeout_event.error_time
            self.first_device = device
        if self.last_time < timeout_event.error_time:
            self.last_time = timeout_event.error_time
            self.last_device = device
        if timeout_event.error_type == regular_table.TIMEOUT_SOCKET:
            for info in timeout_event.remote_info:
                remote_device = self.get_remote_device(info, device, device_table)
                self.remote_relation.add_remote_relation(device, remote_device)
            return
        if timeout_event.error_type == regular_table.TIMEOUT_NOTIFY:
            remote_rank = timeout_event.remote_rank
            remote_device = device if remote_rank == "local" else self.identifier.rank_id_device_map.get(remote_rank)
            self.remote_relation.add_remote_relation(device, remote_device)

    def union(self, timeout_error):
        """
        Union the another TimeoutErrorOfIdentifier instance data
        :param timeout_error: TimeoutErrorOfIdentifier instance
        """
        if not isinstance(timeout_error, self.__class__):
            return
        if timeout_error.identifier != self.identifier:
            return
        if timeout_error.first_time < self.first_time and self.first_device != timeout_error.first_device:
            self.first_time = timeout_error.first_time
            self.first_device = timeout_error.first_device
        if timeout_error.last_time > self.last_time and self.last_device != timeout_error.last_device:
            self.last_time = timeout_error.last_time
            self.last_device = timeout_error.last_device
        self.error_devices |= timeout_error.error_devices

    def check_index_and_tag(self, timeout):
        """
        Check the operator index and tag in Notify timeout.
        If not find root device, return empty list and None
        :param timeout: the HCCL_EXEC_TIMEOUT parameter value
        :return: root device and fault description.
        """
        index_dict = {}
        tag_dict = {}
        for device in self.error_devices:
            error_event = device.timeout_error_map.get((self.identifier.name, regular_table.TIMEOUT_NOTIFY))
            if error_event is None:  # not found error event
                continue
            if error_event.index:
                index_dict.setdefault(error_event.index, []).append(device)
            if error_event.tag:
                tag_dict.setdefault(error_event.tag, []).append(device)
        if len(index_dict) > 1:
            # if index is digit, sorted. if index not is digit, put it in the back
            index_sort_list = sorted(index_dict.keys(), key=lambda x: (0, int(x)) if x.isdigit() else (1, x))
            min_index_device = index_dict.get(index_sort_list[0])
            description = fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT_INDEX_ERR.format(timeout,
                                                                                          index_sort_list,
                                                                                          index_sort_list[0])
            return min_index_device, description
        if len(tag_dict) > 1:
            tag_sort_list = sorted(tag_dict.items(), key=lambda x: len(x[1]))
            min_tag = tag_sort_list[0][0]
            min_tag_device = tag_sort_list[0][1]
            description = fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT_TAG_ERR.format(timeout,
                                                                                        list(tag_dict.keys()),
                                                                                        min_tag)
            return min_tag_device, description
        return [], None


class ErrorChecker(BaseChecker):
    """
    Checker for Error log
    """
    timeout_error_info_map = {}  # each dict is {(identifier_name, timeout_type): TimeoutErrorOfIdentifier()}
    # the connected device dict. eg. {identifier_name: connected_success_devices_set}, only use for root info timeout
    connected_device_with_identifier = {}

    def __init__(self, device_table: DeviceTable, suspected_lagging_devices: list = None, cfg: DiagCFG = None):
        super().__init__(device_table)
        self.suspected_lagging_devices = suspected_lagging_devices
        self.cfg = cfg

    @staticmethod
    def cycle_find_err_device(start_device_ip, relation_map):
        """
        Cycle find error device
        :param start_device_ip: started device
        :param relation_map: map of the waiting relationship
        """
        step_num = 0
        result_list = []
        traversed_device_list = []

        while True:
            if step_num > TEN_THOUSAND:
                break
            # 存在循环等待
            if start_device_ip in traversed_device_list:
                repeat_index = traversed_device_list.index(start_device_ip)
                result_list = traversed_device_list[repeat_index:]
                # 循环等待在展示等待链后增加重复的节点
                traversed_device_list.append(start_device_ip)
                traversed_device_list = traversed_device_list[repeat_index:]
                return result_list, traversed_device_list

            traversed_device_list.append(start_device_ip)
            remote_device_ip = relation_map.get(start_device_ip, "")
            if not remote_device_ip:
                break
            start_device_ip = remote_device_ip
            step_num += 1
        # 没有循环等待，最后一个设备为问题设备
        if traversed_device_list:
            result_list.append(traversed_device_list[-1])
        return result_list, traversed_device_list

    def check_error_relation(self, relation_map: Dict[str, str]):
        check_result = False
        if not relation_map:
            return check_result
        first_err_device_ip = self.device_table.err_device[0].device_ip
        # 如果首报错不在等待关系中，取排序后的第一个
        if first_err_device_ip not in relation_map:
            first_err_device_ip = max(relation_map.keys())
        root_device_list, remote_link_list = self.cycle_find_err_device(first_err_device_ip, relation_map)
        device_ip_map = self.device_table.device_ip_map
        # 等待关系要两个节点及以上
        if remote_link_list and len(remote_link_list) > 1:
            for device_ip in remote_link_list:
                device = device_ip_map.get(device_ip, None)
                if not device:
                    self.remote_links.append(device_ip)
                    continue
                self.remote_links.append(str(device))
        temp_root_devices = []
        for device_ip in root_device_list:
            device = device_ip_map.get(device_ip, None)
            if not device:
                continue
            temp_root_devices.append(device)
        if temp_root_devices:
            self.root_devices = temp_root_devices
            self.fault_description = fault_description.TRANSPORT_INIT_ERROR
            check_result = True
        return check_result

    def check(self):
        if not self.device_table.err_device:
            if self.suspected_lagging_devices:
                self.fault_description = fault_description.LAGGING_ON_WAITING_TIMEOUT_REPORT
                self.root_devices = self.suspected_lagging_devices
                return
            self.fault_description = fault_description.ALL_NO_ERROR
            self.root_devices = [Device()]
            return
        # parse their own timeout error content
        self._parse_err_content()
        first_err_device = self.device_table.err_device[0]
        self.first_error_device = first_err_device
        self.last_error_device = self.device_table.err_device[-1]
        if self._check_cqe_error():
            return
        if self._check_cluster_exception():
            return
        # Transport init error检查
        if self.check_error_relation(self.device_table.get_transport_init_error_relation()):
            return
        if self._check_timeout_err():
            return
        all_time_out_devices = self._get_all_timeout_devices()
        if not all_time_out_devices:
            # aicpu notify wait检查
            if self.check_error_relation(self.device_table.aicpu_notify_wait_relation):
                self.fault_description = fault_description.AI_CPU_NOTIFY_TIMEOUT.format(AI_CPU)
                return
            self.fault_description = fault_description.PART_ERROR_WITH_NO_TIMEOUT
            self.root_devices = self.device_table.err_device
            return
        self.fault_description = fault_description.FIRST_ERROR_RANK
        self.root_devices = list(set(self.device_table.err_device).difference(all_time_out_devices))

    def _check_cqe_error(self):
        """
        Check cqe error
        """

        def string_to_device(*device_name: str) -> List[Device]:
            """
            Map device name to `Device` object
            """
            return list(filter(lambda x: str(x) in device_name, err_devices))

        def cqe_from_one_device() -> List[Device]:
            """
            Return the root device of cqe when all target ranks are the same.
            """
            device_links = [device_link.split(' -> ') for device_link in self.device_links]
            # count target_rank
            rank_count, target_rank_count = dict(), dict()
            for rank, target_rank in device_links:
                rank_count[rank] = rank_count.get(rank, 0) + 1
                target_rank_count[target_rank] = target_rank_count.get(target_rank, 0) + 1

            # dealing with a special situation of A -> B & B -> A
            if set(rank_count.keys()) == set(target_rank_count.keys()) or len(device_links) == 1:
                return []

            most_common_target = max(target_rank_count, key=target_rank_count.get)
            # filter device_link which rank device != most_common_target
            filtered_target_devices = {target_rank for rank, target_rank in device_links if rank != most_common_target}
            if len(filtered_target_devices) != 1:
                return []

            return string_to_device(most_common_target)

        if not self.device_table.cqe_link_map:
            return False

        err_devices = set()
        for device, device_ip_list in sorted(self.device_table.cqe_link_map.items(), key=lambda x: x[0].err_time):
            err_devices.add(device)
            target_err_devices = self._add_device_links(device, device_ip_list)
            err_devices.update(target_err_devices)

        self.root_devices = cqe_from_one_device() or list(err_devices)
        self.fault_description = fault_description.CQE_ERROR
        return True

    def _check_cluster_exception(self):
        """
        Check the cluster exception info from HCCL
        """
        if not self.device_table.cluster_exception:
            return False
        root_devices = []
        error_causes = set()
        for device_info, error_cause in self.device_table.cluster_exception.items():
            server_device_id = device_info.split("/")
            if len(server_device_id) != 2:  # the device info is "IP/ID" format
                continue
            root_device = self.device_table.get_device_by_server_device_id(*server_device_id)
            if root_device == Device():
                continue
            root_devices.append(root_device)
            error_causes.add(error_cause)
        if not root_devices:
            return False
        self.root_devices = root_devices
        error_cause_str = ""
        for idx, cause in enumerate(error_causes):
            error_cause_str += f"{idx + 1}. {cause};"
        self.fault_description = fault_description.CLUSTER_EXCEPTION_LOCATION_ERROR.format(error_cause_str)
        return True

    def _check_timeout_err(self):
        """
        Check hccl timeout err
        :return:
        """
        first_time_error_event = self.first_error_device.first_timeout_error_event
        if not first_time_error_event:  # not record timeout error
            return False
        timeout_identifier = first_time_error_event.identifier
        timeout_type = first_time_error_event.error_type
        timeout_error_info = self.timeout_error_info_map.get((timeout_identifier, timeout_type))
        func_map = {
            regular_table.TIMEOUT_SOCKET: self._check_socket_timeout,
            regular_table.TIMEOUT_NOTIFY: self._check_notify_timeout,
            regular_table.TIMEOUT_FFTS: self._check_notify_timeout,
            regular_table.TIMEOUT_ROOT_INFO: self._check_init_timeout
        }
        if not timeout_error_info:  # No timeout errors occurred in the communication domain
            return False
        if timeout_type in [regular_table.TIMEOUT_SOCKET, regular_table.TIMEOUT_ROOT_INFO]:
            normal_error_info = self.timeout_error_info_map.get((timeout_identifier, regular_table.TIMEOUT_NORMAL))
            timeout_error_info.union(normal_error_info)
        if timeout_type == regular_table.TIMEOUT_NORMAL:
            socket_error_info = self.timeout_error_info_map.get((timeout_identifier, regular_table.TIMEOUT_SOCKET))
            if not socket_error_info and self.device_table.transport_init_error_flag:
                self.fault_description = fault_description.TRANSPORT_INIT_ERROR_NO_DEVICE_ID
                self.root_devices = [Device()]
                return True
            timeout_error_info.union(socket_error_info)
            timeout_type = regular_table.TIMEOUT_SOCKET
        first_error_time = timeout_error_info.first_time
        last_error_time = timeout_error_info.last_time
        interval_times = (datetime.strptime(last_error_time, '%Y-%m-%d-%H:%M:%S.%f') -
                          datetime.strptime(first_error_time, '%Y-%m-%d-%H:%M:%S.%f')).total_seconds()
        func = func_map.get(timeout_type)
        if not func:
            return False
        func(timeout_error_info, interval_times)
        return True

    def _check_init_timeout(self, error_identifier_info: TimeoutErrorOfIdentifier, interval_times: float):
        """
        Check the init timeout root ranks and err reason
        :param error_identifier_info: all init error info of error_identifier
        :param interval_times: time difference between the earliest init error card and the latest init error card
        """
        # Check connected failed and lost logs rank
        identifier_instance = error_identifier_info.identifier
        connected_ranks = self.connected_device_with_identifier.get(identifier_instance.name, set())
        lost_rank = []
        if identifier_instance and connected_ranks:
            for rank_id in range(int(identifier_instance.rank_num)):
                if str(rank_id) in identifier_instance.get_rank_id_list():
                    continue
                if str(rank_id) in connected_ranks:
                    continue
                lost_rank.append(RankDevice(identifier=identifier_instance.name, rank_id=rank_id))
        if lost_rank:
            self.fault_description = fault_description.INIT_FAILED_WITH_NO_CONN_NO_LOG
            self.root_devices = lost_rank
            return

        timeout = self.device_table.get_timeout('CONNECT_TIMEOUT')
        # If some ranks do not report this error, these ranks are the root cause node.
        # Otherwise, all ranks may have problems
        all_devices = set(self.device_table.get_device_list_of_identifier(identifier_instance.name))
        no_this_err_devices = all_devices - error_identifier_info.error_devices
        if no_this_err_devices:
            self.fault_description = fault_description.PART_INIT_FAILED
            self.root_devices = list(no_this_err_devices)
            return
        if float(timeout) < interval_times:
            self.fault_description = fault_description.ALL_INIT_FAILED_WITH_TIMEOUT.format(timeout)
            self.root_devices = [error_identifier_info.last_device]
            return

        self.fault_description = fault_description.ALL_INIT_FAILED_NOT_TIMEOUT.format(timeout)
        self.root_devices = [identifier_instance.root_device]

    def _check_socket_timeout(self, error_identifier_info: TimeoutErrorOfIdentifier, interval_times: float):
        """
        Check the socket timeout root ranks and err reason
        :param error_identifier_info: all socket error info of error_identifier
        :param interval_times: time difference between the earliest socket error card and the latest socket error card
        """
        identifier_instance = error_identifier_info.identifier
        timeout = self.device_table.get_timeout('CONNECT_TIMEOUT')
        # check the device TLS SWITCH
        min_tls_diff_rank, minority_tls_status, majority_tls_status = self._filter_min_tls_switch_rank()
        if min_tls_diff_rank:
            self.fault_description = fault_description.TLS_SWITCH_DIFFERENT.format(
                minority_tls_status[0],
                majority_tls_status[0], )
            self.root_devices = list(min_tls_diff_rank)
            return
        # If some ranks do not report this error, these ranks are the root cause node.
        # Otherwise, all ranks may have problems
        all_devices = set(self.device_table.get_device_list_of_identifier(identifier_instance.name))
        error_devices = error_identifier_info.error_devices
        no_this_err_devices = all_devices - error_devices
        if not no_this_err_devices and float(timeout) < interval_times:
            self.fault_description = fault_description.ALL_SOCKET_ERROR_WITH_TIMEOUT.format(timeout)
            self.root_devices = [error_identifier_info.last_device]
            return

        self.root_devices, self.remote_links = error_identifier_info.remote_relation.check_socket_remote_relation(
            error_identifier_info.first_device, self.cfg.parsed_saver)
        timeout_scope = PART if no_this_err_devices else ALL
        self.fault_description = fault_description.ALL_SOCKET_ERROR_NOT_TIMEOUT.format(timeout_scope, timeout)

    def _check_notify_timeout(self, error_identifier_info: TimeoutErrorOfIdentifier, interval_times: float):
        """
        Check the notify (ffts+ run faile) timeout root ranks and err reason
        :param error_identifier_info: all notify error info of error_identifier
        :param interval_times: time difference between the earliest notify error card and the latest notify error card
        """
        identifier_instance = error_identifier_info.identifier
        timeout = self.device_table.get_timeout('EXEC_TIMEOUT')
        # If some ranks do not report this error, these ranks are the root cause node.
        # Otherwise, all ranks may have problems
        all_devices = set(self.device_table.get_device_list_of_identifier(identifier_instance.name))
        error_devices = error_identifier_info.error_devices
        no_this_err_device = all_devices - error_devices
        if no_this_err_device:
            self.fault_description = fault_description.PART_NOTIFY_ERROR
            self.root_devices = list(no_this_err_device)
            return
        if 0 < int(timeout) < interval_times:
            self.fault_description = fault_description.ALL_NOTIFY_ERROR_WITH_TIMEOUT.format(timeout)
            self.root_devices = [error_identifier_info.last_device]
            return
        if error_identifier_info.error_type == regular_table.TIMEOUT_NOTIFY:
            # check the index and tag
            root_device, description = error_identifier_info.check_index_and_tag(timeout)
            if root_device:
                self.fault_description = description
                self.root_devices = root_device
                return
            self.fault_description, self.root_devices, self.remote_links = \
                error_identifier_info.remote_relation.check_notify_remote_relation()
            if self.device_table.no_group_rank_information_flag:
                self.note_msg.append(NO_GROUP_RANK_INFO_NOTE)
        if not self.fault_description:
            self.fault_description = fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT.format(timeout)
            self.root_devices = [Device()]
            return

    def _get_all_timeout_devices(self):
        """
        Get all timeout devices
        :return:
        """
        all_time_out_devices = set()
        for error_identifier_info in self.timeout_error_info_map.values():
            all_time_out_devices |= error_identifier_info.error_devices
        return all_time_out_devices

    def _add_device_links(self, rank: Device, device_ip_list: list):
        """
        Add device to device_links
        :param rank: Rank instance
        :param device_ip_list: device ip list
        :return err_rank: the set to record error ranks
        """
        err_rank = set()
        if not device_ip_list:
            return err_rank
        for device_ip in device_ip_list:
            target_rank = self.device_table.get_device_by_device_ip(device_ip)
            if target_rank != Device():
                self.device_links.append("{} device-{} -> {} device-{}".format
                                         (rank.worker_name, str(rank.device_id), target_rank.worker_name,
                                          str(target_rank.device_id)))
                err_rank.add(target_rank)
            else:
                self.device_links.append(
                    "{} device-{} -> {}".format(rank.worker_name, str(rank.device_id), str(device_ip)))
                if SOME_DEVICE_FAILED not in self.note_msg:
                    self.note_msg.append(SOME_DEVICE_FAILED)
        return err_rank

    def _parse_err_content(self):
        """
        Each device parse their own timeout error content.
        """
        self.device_table.err_device.sort(key=lambda x: x.err_time)
        for device in self.device_table.err_device:
            for identifier_type_key, timeout_event in device.timeout_error_map.items():
                identifier_name, timeout_type = identifier_type_key
                err_identifier_info = self.timeout_error_info_map.setdefault(identifier_type_key,
                                                                             TimeoutErrorOfIdentifier())
                err_identifier_info.identifier = self.device_table.identifier_dict.get(identifier_name, Identifier())
                err_identifier_info.error_type = timeout_type
                # Update the timeout event info
                err_identifier_info.update_event(timeout_event, device, self.device_table)
                if timeout_type == regular_table.TIMEOUT_ROOT_INFO and timeout_event.connected_ranks:
                    self.connected_device_with_identifier.update({identifier_name: set(timeout_event.connected_ranks)})

    def _filter_min_tls_switch_rank(self):
        all_ranks = set(self.device_table.device_map.values())
        tls_switch_counts = Counter(rank.tls_status for rank in all_ranks if rank.tls_status)
        if not tls_switch_counts:
            rc_logger.warning("TLS SWITCH status not found.")
            return "", "", ""
        if len(tls_switch_counts) > 2:
            rc_logger.error("TLS SWITCH status is abnormal, please check.")
            raise InnerError("TLS SWITCH status is abnormal, please check.")
        if len(tls_switch_counts) != 1:
            minority_tls_status, majority_tls_status = sorted(tls_switch_counts.items(), key=lambda x: x[1])
            minority_ranks = [rank for rank in all_ranks if rank.tls_status == minority_tls_status[0]]
            return minority_ranks, minority_tls_status, majority_tls_status
        return "", "", ""


class InvalidDeviceChecker(BaseChecker):
    """
    Checker when only one device
    """

    def check(self):
        self.fault_description = fault_description.INVALID_DEVICE_ERROR
        self.root_devices = [Device()]


class ResumingTrainingInvalidBaseInfoChecker(BaseChecker):
    """
    Checker when partial device lack of base info after resuming training
    """

    def __init__(self, device_table: DeviceTable, lack_info_worker_set: set):
        super().__init__(device_table)
        self.root_devices = list(lack_info_worker_set)

    def check(self):
        self.fault_description = fault_description.LACK_OF_BASE_INFO_AFTER_RESUMING_TRAINING


class NoPlogChecker(BaseChecker):
    """
    Checker when there is no plog
    """

    def check(self):
        self.fault_description = fault_description.NO_PLOG_ERROR
        self.root_devices = [Device()]


class NoValidPlogInfoErrorChecker(BaseChecker):
    """
    Checker when no valid plog info
    """

    def check(self):
        self.fault_description = fault_description.NO_VALID_PLOG_INFO_ERROR
        self.root_devices = [Device()]
