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
import os
import re
import logging
from datetime import datetime, timedelta

from ascend_fd.model.diag_info import RCDiagResult
from ascend_fd.model.parse_info import PlogPidParseInfo, PlogBaseInfo

from ascend_fd.pkg.diag.root_cluster import fault_description
from ascend_fd.utils import regular_table
from ascend_fd.utils.status import InfoIncorrectError
from ascend_fd.utils.tool import safe_read_json, collect_parse_results, MultiProcessJob
from ascend_fd.model.cfg import DiagCFG
from ascend_fd.pkg.parse.root_cluster.rc_parse_job import check_device_id_repeat
from ascend_fd.pkg.parse.root_cluster.parser import PidFileParser
from ascend_fd.pkg.diag.root_cluster.utils import ErrorChecker, InvalidDeviceChecker, NoPlogChecker, Device, \
    DeviceTable, Identifier, NoValidPlogInfoErrorChecker, ResumingTrainingInvalidBaseInfoChecker

rc_logger = logging.getLogger("ROOT_CLUSTER")


class RCDiagWorker:

    def __init__(self, cfg: DiagCFG = None, sdk_input: dict = None):
        """
        Init rc diag job

        At least one of 'cfg' and 'sdk_input' MUST be provided for initialization, otherwise this class has no effect
        If them both provided, sdk input would have a higher priority

        :param cfg: Diag Config
        :param
        """
        self.cfg = cfg
        self.pid_find_flag = False
        self.rc_parser_find_flag = True if sdk_input else False  # rc-parser.json文件是否存在标记
        self.cluster_training_flag = False  # flag for cluster training job
        self.device_table = DeviceTable()
        self.resuming_training_time = regular_table.MIN_TIME
        self.recovery_time = regular_table.MIN_TIME
        self.latest_lagging_time = regular_table.MIN_TIME
        self.all_worker_has_recovered = True
        self.lack_base_info_device_set = set()
        self.suspected_lagging_devices = []
        self.sdk_input_rc_parser = sdk_input
        # start time and end time are used for further comparison
        # if they stay in default, the start would be the min time, and the end would be the max time
        self.start_train_time = regular_table.MAX_TIME
        self.end_train_time = regular_table.MIN_TIME

    @staticmethod
    def _parse_plog_for_old_version(plog_list, device_info_dict):
        """
        Parse the plog file and obtain the device info
        :param plog_list: the plog-parser-xxx-0/1.log files list
        :param device_info_dict: additional information about the device, contain device_ip and device_tls_status
        :return: rc parser dict
        """
        rc_parser_dict = dict()
        device_pid_map = dict()  # use device id and pid to filter the latest task
        for plog_file in plog_list:
            pid_re = re.match(regular_table.PLOG_PARSED, os.path.basename(plog_file))
            if not pid_re:
                continue
            pid = pid_re[1]
            device_ip_map = device_info_dict.get("device_ip", {})
            device_tls_map = device_info_dict.get("device_tls", {})
            pid_file_parser = PidFileParser(pid, device_ip_map)
            pid_file_parser.parse_log(plog_file)
            this_pid_result = pid_file_parser.get_result()
            if not this_pid_result:
                continue
            if check_device_id_repeat(pid, this_pid_result, device_pid_map, rc_parser_dict):
                continue
            tls_status = device_tls_map.get(pid)
            if tls_status:
                this_pid_result.tls_status = tls_status
            rc_parser_dict[pid] = this_pid_result.to_dict()
        return rc_parser_dict

    def start_job(self) -> RCDiagResult:
        """
        Start rc diag job
        :return: rc diag result
        """
        if self.sdk_input_rc_parser:
            worker_parser_dict = self.sdk_input_rc_parser
        else:
            worker_parser_dict = self.assemble_rc_parser()

        self.update_cluster_level_parameters(worker_parser_dict)
        if self.latest_lagging_time != regular_table.MIN_TIME:
            self.filter_lagging_devices(worker_parser_dict)

        aicpu_notify_wait_dict = {}
        for worker_name in sorted(worker_parser_dict.keys()):
            rc_parser_dict = worker_parser_dict.get(worker_name, {})
            worker_dir_path = "" if self.sdk_input_rc_parser else self.cfg.parsed_saver.get_worker_dir_path(worker_name)
            for info in rc_parser_dict.values():
                self.pid_find_flag = True
                pid_parse_info = PlogPidParseInfo.from_dict(info)
                self._get_single_pid_parser_info(worker_name, worker_dir_path, pid_parse_info, aicpu_notify_wait_dict)
                self.coll_vnic_info(worker_name, pid_parse_info)
        self.add_aicpu_notify_wait_relation(aicpu_notify_wait_dict)
        self._check_integrity_of_rank_in_identifier()
        if not self.pid_find_flag:
            rc_logger.warning("No valid PID information is found in the parser json file.")
            err_checker = NoValidPlogInfoErrorChecker(self.device_table) if self.rc_parser_find_flag \
                else NoPlogChecker(self.device_table)
        else:
            err_checker = self._generate_checker()
        err_checker.check()
        # 添加 MindIE 建链故障设备
        self.add_mindie_error_device(err_checker)
        return err_checker.format_output(self.resuming_training_time, self.start_train_time,
                                         self.end_train_time, self.cfg.parsed_saver.scene,
                                         self.cfg.parsed_saver.board_sn_exist_tag)

    def add_aicpu_notify_wait_relation(self, aicpu_notify_wait_dict):
        identifier_rank_to_device_ip = self.device_table.identifier_rank_to_device_ip
        for device_ip, identifier_rank_list in aicpu_notify_wait_dict.items():
            if not identifier_rank_list:
                continue
            remote_device_ip = identifier_rank_to_device_ip.get(max(identifier_rank_list))
            if not remote_device_ip:
                continue
            self.device_table.aicpu_notify_wait_relation.update({device_ip: remote_device_ip})

    def coll_vnic_info(self, worker_name, pid_parse_info: PlogPidParseInfo):
        if not pid_parse_info or not pid_parse_info.base:
            return
        phy_device_id = pid_parse_info.base.phy_device_id
        vNic_ip = pid_parse_info.base.vNic_ip
        if vNic_ip and phy_device_id:
            self.device_table.vNic_ip_dev_id_relation.update({(worker_name, vNic_ip): phy_device_id})

    def add_mindie_error_device(self, err_checker):
        pull_kv_error_happened_flag = False
        for pull_kv_error_ip in self.cfg.parsed_saver.mindie_diag_result.pull_kv_error_ip_list:
            pull_kv_error_device = self.device_table.get_device_by_device_ip(pull_kv_error_ip)
            if pull_kv_error_device and pull_kv_error_device.worker_name == "worker-NA":
                continue
            err_checker.root_devices.append(pull_kv_error_device)
            err_checker.mindie_error_device.append(str(pull_kv_error_device))
            pull_kv_error_happened_flag = True
        if pull_kv_error_happened_flag:
            err_checker.fault_description.string += os.linesep + fault_description.MINDIE_PULL_KV_ERROR.string
            err_checker.fault_description.code = fault_description.MINDIE_PULL_KV_ERROR.code
            err_checker.root_devices = [
                root_device
                for root_device in err_checker.root_devices
                if root_device and root_device.worker_name != "worker-NA"
            ]
            return
        link_error_happened_flag = False
        for link_error_ip in self.cfg.parsed_saver.mindie_diag_result.link_error_ip_list:
            link_error_device = self.device_table.get_device_by_device_ip(link_error_ip)
            if not link_error_device or link_error_device.worker_name == "worker-NA":
                continue
            err_checker.root_devices.append(link_error_device)
            err_checker.mindie_error_device.append(str(link_error_device))
            link_error_happened_flag = True
        if link_error_happened_flag:
            err_checker.fault_description.string += os.linesep + fault_description.MINDIE_LINK_ERROR.string
        # 过滤 Unknown Device,防止后面进行数量判定出现错误
        if len(err_checker.root_devices) > 2:
            err_checker.root_devices = [
                root_device
                for root_device in err_checker.root_devices
                if root_device and root_device.worker_name != "worker-NA"
            ]

    def update_cluster_level_parameters(self, worker_parser_dict):
        """
        Update cluster level parameters including the following:
        1. The latest resuming training time
        2. The latest lagging time
        :param worker_parser_dict: dictionary for storing parsing results
        """
        for worker_name in sorted(worker_parser_dict.keys()):
            rc_parser_dict = worker_parser_dict.get(worker_name, {})
            worker_resuming_training_time, worker_recovery_time = \
                self._fetch_worker_lever_parameters(worker_name, rc_parser_dict)
            # if any worker has not been recovered, then recovery time would be ignored
            if self.all_worker_has_recovered and not worker_recovery_time:
                self.all_worker_has_recovered = False
                self.recovery_time = regular_table.MIN_TIME
            if self.all_worker_has_recovered and worker_recovery_time > self.recovery_time:
                self.recovery_time = worker_recovery_time
            if worker_resuming_training_time > self.resuming_training_time:
                self.resuming_training_time = worker_resuming_training_time

        # if all workers have recovered, then the latest recovery time would influence the latest resuming training time
        if self.all_worker_has_recovered:
            self.resuming_training_time = max(self.resuming_training_time, self.recovery_time)

        # no plog scene when both the start time and the end time stay in default
        # in this case, assign default min time and max time respectively
        if self.start_train_time == regular_table.MAX_TIME and self.end_train_time == regular_table.MIN_TIME:
            self.start_train_time = regular_table.MIN_TIME
            self.end_train_time = regular_table.MAX_TIME

    def filter_lagging_devices(self, worker_parser_dict):
        """
        Filter those devices report timeout at the early stage with respect to the latest
        """
        time_format = "%Y-%m-%d-%H:%M:%S.%f"
        latest_time = datetime.strptime(self.latest_lagging_time, time_format)
        # the threshold to determine if a device only reports in an early stage
        threshold_time_in_minutes = 15
        lagging_devices = set()
        for worker_name, rc_parser_dict in worker_parser_dict.items():
            for pid, pid_info in rc_parser_dict.items():
                device_instance = Device(pid, worker_name, self.device_table)
                pid_parse_info = PlogPidParseInfo.from_dict(pid_info)
                device_instance.update_base_info(pid_parse_info.base)
                lagging_time = pid_parse_info.lagging_time
                if lagging_time == regular_table.MIN_TIME:
                    lagging_devices.add(device_instance)
                    continue
                cur_lagging_time = datetime.strptime(lagging_time, time_format)
                if abs(latest_time - cur_lagging_time) > timedelta(minutes=threshold_time_in_minutes):
                    lagging_devices.add(device_instance)
        self.suspected_lagging_devices = list({str(device): device for device in lagging_devices}.values())

    def assemble_rc_parser(self):
        """
        Assemble a rc parser dict for further use if rc parser exists
        Otherwise multiprocessing plog-parsers for the old version
        """
        # plog parser file path dict for the old version
        plog_parsed_dict = self.cfg.parsed_saver.get_worker_plog_dict()
        worker_parser_dict = {}
        multiprocess_job = MultiProcessJob("ROOT_CLUSTER", pool_size=20, task_id=self.cfg.task_id, failed_raise=False)
        worker_path_list = self.cfg.parsed_saver.infer_group_2_worker_path.get(self.cfg.parsed_saver.infer_instance)
        for worker_name, worker_dir_path in self.cfg.parsed_saver.get_all_worker_dir_path().items():
            if self.cfg.parsed_saver.infer_task_flag and worker_path_list and worker_dir_path not in worker_path_list:
                continue
            if worker_name not in self.device_table.worker_list:
                self.device_table.worker_list.append(worker_name)
            path_list = collect_parse_results(worker_dir_path, "rc-parser.json")
            rc_parser_file = path_list[0] if path_list else ""
            if not os.path.exists(rc_parser_file):
                device_info_dict = self.cfg.parsed_saver.get_device_info_from_json(worker_name)
                multiprocess_job.add_security_job(worker_name, self._parse_plog_for_old_version,
                                                  plog_parsed_dict.get(worker_name, []), device_info_dict)
                continue
            self.rc_parser_find_flag = True
            worker_parser_dict.update({worker_name: safe_read_json(rc_parser_file)})
        multi_results, _ = multiprocess_job.join_and_get_results()  # not handle the parse error
        worker_parser_dict.update(multi_results)
        return worker_parser_dict

    def _get_single_pid_parser_info(self, worker_name, worker_dir_path, pid_parse_info, aicpu_notify_wait_dict):
        """
        Get the base device info, error info from rc_parser_dict
        :param worker_name: worker dir name
        :param worker_dir_path: worker dir path
        :param pid_parse_info: the rc parser dict contain plog device info
        :param aicpu_notify_wait_dict: the all aicpu notify wait
        """
        device_instance = self._get_device_instance(worker_name, worker_dir_path, pid_parse_info)
        if not device_instance:
            return
        # error info
        if pid_parse_info.error:
            device_instance.update_error_info(pid_parse_info.error, self.resuming_training_time)
            self.device_table.update_hccl_cluster_exception(pid_parse_info.error.cluster_exception)
        # tls status
        device_instance.tls_status = pid_parse_info.tls_status
        # show logs
        device_instance.update_show_log(pid_parse_info.show_logs)
        self.device_table.add_device(device_instance)
        # parser log file path
        log_file_path = pid_parse_info.plog_parsed_name
        if log_file_path:
            device_instance.log_file_path = os.path.join(worker_dir_path, log_file_path)
        # aicpu notify wait
        aicpu_notify_wait_info = pid_parse_info.aicpu_notify_wait_remote
        if aicpu_notify_wait_info:
            aicpu_notify_wait_dict.setdefault(device_instance.device_ip, []).append(aicpu_notify_wait_info)

    def _get_device_instance(self, worker_name, worker_dir_path, info: PlogPidParseInfo) -> Device:
        """
        Get device instance
        :param worker_name: worker dir name
        :param worker_dir_path: worker dir path
        :param info: the rc parser dict contain plog device info
        """
        device_instance = Device(info.pid, worker_name, self.device_table)
        if not self.sdk_input_rc_parser and self.cfg.parsed_saver.infer_task_flag:
            if self.cfg.parsed_saver.pid_device_dict.get(worker_name + info.pid):
                return self._infer_task_initialization(device_instance, worker_dir_path, info, worker_name)
        # device base info
        if not info.base or not info.base.rank_map:
            return None
        device_instance.update_base_info(info.base)
        # record the identifier info
        for identifier_name, rank_info in info.base.rank_map.items():
            if not rank_info or rank_info.rank_num < 1:
                continue
            self.cluster_training_flag = True
            rank_num = rank_info.rank_num
            identifier_instance = self.device_table.identifier_dict.setdefault(
                identifier_name, Identifier(identifier_name, rank_num))
            if identifier_name != regular_table.DEFAULT_IDENTIFIER and identifier_instance.rank_num != rank_num:
                # the identifier info in two plogs are not same
                error_msg = (f"The identifier [{identifier_name}] have two different rank num values "
                             f"({identifier_instance.rank_num} and {rank_num}).")
                rc_logger.error(error_msg)
                raise InfoIncorrectError(error_msg)
            identifier_instance.rank_num = rank_num

            if not rank_info.rank_id:
                continue
            identifier_instance.update_device(device_instance, rank_info.rank_id)
            self.device_table.update_max_identifier(identifier_name, rank_num)
            self.device_table.identifier_rank_to_device_ip.update(
                {"{}:{}".format(identifier_name, rank_info.rank_id): device_instance.device_ip})
        # record the root rank info to identifier
        for identifier_name in info.base.root_list:
            if identifier_name not in self.device_table.identifier_dict:
                rc_logger.warning(
                    "The identifier [%s] has root rank: worker[%s] pid[%s]. But no related information "
                    "is found in this device's log.", identifier_name, worker_name, info.pid)
                continue
            self.device_table.identifier_dict.get(identifier_name).root_device = device_instance
        self.device_table.update_timeout(info.base.timeout_param)
        return device_instance

    def _infer_task_initialization(self, device_instance, worker_dir_path, info, worker_name) -> Device:
        # device base info
        sever_info_instance = self.cfg.parsed_saver.pid_device_dict.get(f"{worker_name}{info.pid}")
        device_instance.update_base_info(PlogBaseInfo.from_dict(sever_info_instance.get_device_info_dict()))
        device_instance.device_table = self.device_table
        self.device_table.update_timeout(info.base.timeout_param)
        # record the identifier info
        container_info = safe_read_json(os.path.join(worker_dir_path, regular_table.CONTAINER_FILE))
        container_ip = container_info.get("container_ip", "")
        infer_group = self.cfg.parsed_saver.ip_infer_group.get(container_ip)
        if not infer_group or infer_group.infer_group_name != self.cfg.parsed_saver.infer_instance:
            return None
        identifier_instance = self.device_table.identifier_dict.setdefault(
            infer_group.infer_group_name, Identifier(infer_group.infer_group_name, infer_group.device_num)
        )
        identifier_instance.update_device(device_instance, device_instance.phy_device_id)
        self.device_table.update_max_identifier(infer_group.infer_group_name, infer_group.device_num)
        identifier_instance.root_device = device_instance
        self.cluster_training_flag = True
        return device_instance

    def _generate_checker(self):
        """
        Generate the error checker
        :return: error checker.
        """
        if self.cluster_training_flag and self.lack_base_info_device_set:
            return ResumingTrainingInvalidBaseInfoChecker(self.device_table, self.lack_base_info_device_set)
        if not self.cluster_training_flag:
            return InvalidDeviceChecker(self.device_table)
        return ErrorChecker(self.device_table, suspected_lagging_devices=self.suspected_lagging_devices, cfg=self.cfg)

    def _check_integrity_of_rank_in_identifier(self):
        """
        Check the integrity of the rank ID in the communication identifier. Use log to record result
        """
        for identifier_instance in self.device_table.identifier_dict.values():
            identifier_instance.check_integrity_of_rank()

    def _fetch_worker_lever_parameters(self, worker_name: str, rc_parse_dict: dict):
        """
        Get the latest resuming training starting time and the latest lagging time for the worker level
        """
        latest_resumable_starting_time = regular_table.MIN_TIME
        latest_recovery_time = regular_table.MIN_TIME
        all_pid_has_recovered_flag = True
        if not rc_parse_dict:
            return latest_resumable_starting_time, latest_recovery_time
        for pid, pid_info in rc_parse_dict.items():
            end_train_time = pid_info.get("end_train_time", regular_table.MAX_TIME)
            resume_time = pid_info.get("start_resumable_training_time", regular_table.MIN_TIME)
            recovery_time = pid_info.get("recovery_success_time", regular_table.MIN_TIME)
            lagging_time = pid_info.get("lagging_time", regular_table.MIN_TIME)
            start_train_time = pid_info.get("start_train_time", regular_table.MIN_TIME)
            self.latest_lagging_time = max(self.latest_lagging_time, lagging_time)
            self.start_train_time = min(self.start_train_time, start_train_time)
            self.end_train_time = max(self.end_train_time, end_train_time)
            # if any pid has no recovery time record, raise the false flag
            if all_pid_has_recovered_flag and recovery_time == regular_table.MIN_TIME:
                all_pid_has_recovered_flag = False
            if resume_time > end_train_time:
                continue
            # check if any base info is re inited but no further base info filled in
            if resume_time != regular_table.MIN_TIME and \
                    all(not bool(value) for value in pid_info.get("base", {}).values()):
                self.lack_base_info_device_set.add(Device(pid, worker_name, self.device_table))
            if resume_time > latest_resumable_starting_time:
                latest_resumable_starting_time = resume_time
            if all_pid_has_recovered_flag and recovery_time > latest_recovery_time:
                latest_recovery_time = recovery_time
        # if all pids has recovered, return both time records
        if all_pid_has_recovered_flag:
            return latest_resumable_starting_time, latest_recovery_time
        # if any pid has not recovered, return only the resuming time by the keyword 'attr init success'
        return latest_resumable_starting_time, ""


def start_rc_diag_job(cfg) -> RCDiagResult:
    """
    Start rc diag job
    :param cfg: diag config
    :return: rc diag result
    """
    rc_logger.info("Start root cluster diagnosis task.")
    return RCDiagWorker(cfg).start_job()
