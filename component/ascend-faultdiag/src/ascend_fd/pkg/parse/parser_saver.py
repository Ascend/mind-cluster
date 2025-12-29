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
import argparse
import bisect
import glob
import json
import logging
import os
import re
import shutil
import subprocess
from dataclasses import dataclass, field
from typing import Dict, List
from abc import ABC, abstractmethod

from ascend_fd.configuration.config import CUSTOM_CONFIG_PATH
from ascend_fd.model.mindie_info import MindIEParseResult, MindIEDiagResult
from ascend_fd.pkg.customize.custom_config.config_info import get_config_info, ConfigInfo, CustomFileInfo
from ascend_fd.utils import regular_table
from ascend_fd.utils.status import ParamError, InnerError, PathError, FileNotExistError
from ascend_fd.utils.tool import safe_walk, get_log_module_and_time, path_check, fd_logger, SERIAL_NUMBER, \
    check_and_format_time_str, echo, safe_list_dir, safe_write_open, safe_read_json, BOARD_SERIAL_NUMBER, \
    collect_parse_results, WORKER_MAX_NUM, check_symlink

logger = logging.getLogger("FAULT_DIAG")


class BaseLogSaver(ABC):
    # the log type used to indicate it in the log
    LOG_TYPE = None
    # declare if the log has a preset directory to store
    CENTRALIZED_STORAGE_DIRECTORY = None
    # the variable name of cmd arg, which is configured in cli.py
    CMD_ARG_KEYS = None

    def __init__(self):
        self.log_map: Dict[str, List[LogInfoSaver]] = {}

    def __init_subclass__(cls, **kwargs):
        super().__init_subclass__(**kwargs)
        SaverFactory.register_saver(cls.__name__, cls)

    @property
    def is_sdk_input(self):
        return bool(self.log_map)

    @abstractmethod
    def filter_log(self, file_dir: str):
        pass

    def update_log(self, source: dict):
        """
        Update log directly from sdk source.
        :param source: data from sdk, format: {source_file: LogInfoSaver}
        """
        self.log_map.update(source)


class SaverFactory:
    __registry = {}

    @classmethod
    def register_saver(cls, saver_name: str, saver_class: type):
        if not issubclass(saver_class, BaseLogSaver):
            raise ParamError(f"{saver_class} must be a subclass of BaseLogSaver")
        cls.__registry[saver_name] = saver_class

    @classmethod
    def create_saver(cls, saver_name: str):
        saver_class = cls.__registry.get(saver_name)
        if saver_class is None:
            raise ParamError(f"Invalid saver name: {saver_name}")
        return saver_class()

    @classmethod
    def batch_create_savers(cls, saver_names):
        return [cls.create_saver(name) for name in saver_names]

    @classmethod
    def list_savers_classes(cls):
        return list(cls.__registry.values())

    @classmethod
    def get_saver_class(cls, saver_name: str):
        saver_class = cls.__registry.get(saver_name)
        if saver_class is None:
            raise KeyError(f"Invalid saver name: {saver_name}")
        return saver_class


class ProcessLogSaver(BaseLogSaver):
    LOG_TYPE = "process log"
    CENTRALIZED_STORAGE_DIRECTORY = "process_log"
    CMD_ARG_KEYS = ["process_log"]

    MAX_LOG_NUM = 2
    PLOG_KEY = "plog-"
    DEVICE_LOG_KEY = "device-"

    def __init__(self):
        """
        Process Log Saver
        """
        super().__init__()
        # plog save dict
        self.plog_dict = {
            "run": dict(),
            "debug": dict(),
            "default": dict()
        }

        # device log save dict
        self.device_log_dict = {
            "debug": dict(),
            "run": dict(),
            "default": dict()
        }

        self.resuming_training_time = regular_table.MIN_TIME
        self.resuming_training_record = dict()
        self.n_seconds_recovery_record = dict()
        self.plogs_to_be_append = set()

    @staticmethod
    def _allocate_unique_pid(unique_pids: set):
        fake_unique_pid_upper_bound = 50
        for fake_pid in range(fake_unique_pid_upper_bound):
            if fake_pid not in unique_pids:
                unique_pids.add(fake_pid)
                return fake_pid
        return 0

    def filter_log(self, file_dir: str):
        """
        Add files recursively;
        :param file_dir: the directory to find
        """
        if not file_dir or not os.path.isdir(file_dir):
            return
        run_dir = ""
        for root, dirs, files in safe_walk(file_dir):
            # record the dir "plog" under the dir "run"
            if not run_dir and "plog" in dirs and root.endswith("run"):
                run_dir = os.path.join(root, "plog")
            for file in files:
                self._add_single_file(root, file)
        if run_dir:
            self._extract_resuming_training_time(run_dir)
        for pid, file_path in self.plogs_to_be_append:
            self._add_log(pid, file_path, self.plog_dict.get("run", dict()), False)

    def get_plog_dict(self):
        """
        Get the list of plog files categorized by PID
        :return: plog files dict
        """
        new_log_dict = dict()
        if self.is_sdk_input:
            self._get_sdk_log_dict(regular_table.CANN_PLOG_SOURCE, self.plog_dict)
            self._extract_resuming_training_time_by_item()
        for log_dict in self.plog_dict.values():
            for pid, plog_list in log_dict.items():
                new_log_dict.setdefault(pid, []).extend(plog_list)
        return new_log_dict

    def get_device_log_dict(self):
        """
        Get the list of device log files categorized by PID
        :return: device log files dict
        """
        new_log_dict = dict()
        if self.is_sdk_input:
            self._get_sdk_log_dict(regular_table.CANN_DEVICE_SOURCE, self.device_log_dict)
        for log_dict in self.device_log_dict.values():
            for pid, device_log_list in log_dict.items():
                new_log_dict.setdefault(pid, []).extend(device_log_list)
        return new_log_dict

    def get_resuming_training_time(self):
        return self.resuming_training_time

    def _get_sdk_log_dict(self, key: str, saved_dict: dict):
        """
        For sdk input, ensure plog can be added properly
        """
        item_list = self.log_map.get(key, [])
        if not item_list:
            return
        unique_pids = set()
        for item in item_list:
            match_re = re.match(regular_table.PLOG_ORIGIN, item.filename) or \
                       re.match(regular_table.PLOG_DEVICE_ORIGIN, item.filename)
            pid = match_re[1] if match_re else self._allocate_unique_pid(unique_pids)
            unique_pids.add(pid)
            if item.has_parent("run"):
                self._add_log(pid, item, saved_dict.get("run", {}))
                continue
            if item.has_parent("debug"):
                self._add_log(pid, item, saved_dict.get("debug", {}))
                continue
            self._add_log(pid, item, saved_dict.get("default", {}))

    def _extract_resuming_training_time_by_item(self):
        """
        Extract resuming training time by item for sdk input only
        """
        run_plog = self.plog_dict.get("run")
        if not run_plog:
            return
        pid_info_dict = dict()
        for pid, item_list in run_plog.items():
            all_lines = (
                line
                for item in item_list
                for line in item.log_lines
            )
            for line in all_lines:
                self._handle_single_resuming_training_info(line, pid, pid_info_dict)
        self._update_resuming_training_time_by_pid_info(pid_info_dict)

    def _handle_single_resuming_training_info(self, line, pid, pid_info_dict):
        if regular_table.ATTR_INIT_SUCCESS not in line:
            return
        try:
            _, log_time = get_log_module_and_time(line)
        except IndexError:
            return
        self._handle_stored_attr_init_info(str(pid), pid_info_dict, log_time)

    def _add_single_file(self, dir_path: str, file_name: str):
        """
        Add process log files by name regex, contain plog and device log;
        support two directory structures:
        1) xxx/debug or run/xxx/plog-{pid}_{time}.log or device-{pid}_{time}.log
        2) xxx/plog-{pid}_{time}.log or device-{pid}_{time}.log
        :param dir_path: the log file folder path
        :param file_name: the file name
        """
        file_path = os.path.join(dir_path, file_name)
        # add plog
        if file_name.startswith(self.PLOG_KEY):
            plog_re = re.match(regular_table.PLOG_ORIGIN, file_name)
            if plog_re:
                pid = plog_re[1]
                self._add_process_log(pid, file_path, self.plog_dict)
                return
        # add device_log
        if file_name.startswith(self.DEVICE_LOG_KEY):
            device_log_re = re.match(regular_table.DEVICE_LOG_ORIGIN, file_name)
            if device_log_re:
                pid = device_log_re[1]
                self._add_process_log(pid, file_path, self.device_log_dict)

    def _add_log(self, pid: int, log_source, log_dict: dict, has_limit: bool = True):
        """
        Stores the log file name by time.
        If the number of files exceeds the maximum, older files are deleted and the most recent files are stored;
        :param pid: PID
        :param log_source: log file path or log item
        :param log_dict: plog save dict
        """
        bisect.insort(log_dict.setdefault(pid, []), log_source)
        if has_limit and len(log_dict.get(pid, [])) > self.MAX_LOG_NUM + 1:  # choose the first and last two log
            log_dict[pid] = log_dict.get(pid, [])[:1] + log_dict.get(pid, [])[-2:]

    def _add_process_log(self, pid: int, file_path: str, log_dict: dict):
        """
        Add plog or device log file path based on the log type
        :param pid: PID
        :param file_path: plog log file path
        :param log_dict: log dict
        """
        upper_dir_name = os.path.basename(os.path.dirname(os.path.dirname(file_path)))
        if upper_dir_name in ["debug", "run"]:
            self._add_log(pid, file_path, log_dict.get(upper_dir_name, dict()))
            return
        self._add_log(pid, file_path, log_dict.get("default", dict()))

    def _extract_resuming_training_time(self, run_plog_dir):
        """
        pre-extract resuming training time for further root cluster
        :param run_plog_dir:
        """
        timeout_limit = "10"
        try:
            timeout_cmd_path = path_check(shutil.which("timeout"))
            grep_cmd_path = path_check(shutil.which("grep"))
        except argparse.ArgumentTypeError:
            return
        if not timeout_cmd_path or not grep_cmd_path:
            return
        result = subprocess.run(
            [
                timeout_cmd_path,
                timeout_limit,
                grep_cmd_path,
                "-rHE",
                "--include=plog-*",
                regular_table.ATTR_INIT_SUCCESS + "|" + regular_table.N_SECOND_RECOVERY_FINISH,
                run_plog_dir
            ],
            text=True,
            capture_output=True
        )
        # cmd return code 0 if executed successfully
        if result.returncode != 0:
            fd_logger.warning("Failed to fetching plog initialization info.")
            return
        self._update_resuming_training_time(result)

    def _update_resuming_training_time(self, result):
        """
        Update resuming training time according to the results from grep
        :param result: the results of grep cmd
        """
        pid_info_dict = dict()
        for line in result.stdout.splitlines():
            split_parts = line.split(":", 1)
            if len(split_parts) < 2:
                continue
            filename, log_line = split_parts
            plog_re = re.match(regular_table.PLOG_ORIGIN, os.path.basename(filename))
            if not plog_re:
                continue
            pid = plog_re[1]
            try:
                _, log_time = get_log_module_and_time(log_line)
            except IndexError:
                continue
            self.plogs_to_be_append.add((pid, filename))
            if regular_table.N_SECOND_RECOVERY_FINISH in log_line:
                if log_time > self.n_seconds_recovery_record.get(pid, regular_table.MIN_TIME):
                    self.n_seconds_recovery_record[pid] = log_time
                continue
            self._handle_stored_attr_init_info(pid, pid_info_dict, log_time)
        self._update_resuming_training_time_by_pid_info(pid_info_dict)

    def _update_resuming_training_time_by_pid_info(self, pid_info_dict):
        for pid, info in pid_info_dict.items():
            log_time, count = info
            log_time = regular_table.MIN_TIME if count == 1 else log_time
            self.resuming_training_record.update({pid: log_time})
        # 1. Format the original plog time to the knowledge graph standard
        # e.g., 0000-01-01-00:00:00.000000 -> 0000-01-01 00:00:00.000000
        self.resuming_training_time = check_and_format_time_str(self.resuming_training_time)

    def _handle_stored_attr_init_info(self, pid: str, pid_info_dict: dict, log_time: str):
        """
        Handle stored attr init success info, record the time on duplicated occurrence on a same pid.
        :param pid: process id
        :param pid_info_dict: a structured pid info dict
        :param log_time: log time
        """
        if pid not in pid_info_dict:
            pid_info_dict[pid] = (log_time, 1)
            return
        store_time, count = pid_info_dict.get(pid)
        if log_time <= store_time:
            pid_info_dict[pid] = (store_time, count + 1)
            return
        pid_info_dict[pid] = (log_time, count + 1)
        if log_time > self.resuming_training_time:
            self.resuming_training_time = log_time


class EnvInfoSaver(BaseLogSaver):
    LOG_TYPE = "env check"
    CENTRALIZED_STORAGE_DIRECTORY = "environment_check"
    CMD_ARG_KEYS = ["env_check"]

    NPU_INFO_KEY = ".txt"
    NPU_INFO_NAME = "npu_info_"
    HOST_METRICS_NAME = "host_metrics_"
    HOST_METRICS_KEY = ".json"

    def __init__(self):
        """
        Env Info Saver
        """
        super().__init__()
        self.npu_smi_detail_list = []
        self.npu_info_list = []
        self.host_metrics_path = ""
        self.npu_detail_list = []

    def filter_log(self, file_dir: str):
        """
        Filter env check file
        :param file_dir: env_check dir path
        """
        if not file_dir or not os.path.isdir(file_dir):
            return
        for root, _, files in safe_walk(file_dir):
            for file in files:
                self._add_single_file(root, file)

    def get_npu_info_list(self) -> list:
        """
        Get npu info list for kg parse
        :return: npu info list
        """
        return self.npu_info_list if not self.is_sdk_input else self.log_map.get(regular_table.NPU_INFO_SOURCE, [])

    def get_npu_smi_detail_list(self) -> list:
        """
        Get npu smi detail list for node parse
        :return: npu smi detail list
        """
        return self.npu_smi_detail_list

    def get_host_metrics_path(self) -> str:
        """
        Get host metrics list for kg parse
        :return: host metrics list
        """
        return self.host_metrics_path

    def get_npu_detail_list(self) -> list:
        """
        Get npu detail list for net parse
        :return: npu detail list
        """
        return self.npu_detail_list

    def _add_single_file(self, dir_path: str, file_name: str):
        """
        Add environment check files by name regex, contain npu_info、 npu_smi_detail、host_metrics、and npu_detail info;
        support four directory structures:
        1) xxx/npu_info_before/after.txt
        2) xxx/npu_smi_{npu_id}_details.csv
        3) xxx/npu_{npu_id}_details.csv
        4) xxx/host_metrics_{core_num}.json
        :param dir_path: the log file folder path
        :param file_name: the file name
        """
        if file_name.startswith(self.NPU_INFO_NAME) and file_name.endswith(self.NPU_INFO_KEY):
            self.npu_info_list.append(os.path.join(dir_path, file_name))
            return
        if re.match(regular_table.NPU_SMI_DETAILS_CSV, file_name):
            self.npu_smi_detail_list.append(os.path.join(dir_path, file_name))
            return
        if re.match(regular_table.NPU_DETAILS_CSV, file_name):
            self.npu_detail_list.append(os.path.join(dir_path, file_name))
            return
        if not self.host_metrics_path and file_name.startswith(self.HOST_METRICS_NAME) \
                and file_name.endswith(self.HOST_METRICS_KEY):
            self.host_metrics_path = os.path.join(dir_path, file_name)


class TrainLogSaver(BaseLogSaver):
    LOG_TYPE = "train log"
    CMD_ARG_KEYS = ["train_log"]

    RANK_KEY = "rank-"
    WORKER_KEY = "worker-"
    TXT_END_KEY = ".txt"
    LOG_END_KEY = ".log"
    TRAIN_LOG_LOAD_LIMIT = 20

    def __init__(self):
        """
        Train Log Saver
        """
        super().__init__()
        self.train_log_files = []
        self.logger = logging.getLogger("FAULT_DIAG")

    def get_train_log(self) -> list:
        """
        Get the mindspore train log list
        :return: train log list
        """
        return self.train_log_files if not self.is_sdk_input else self.log_map.get(regular_table.TRAIN_LOG_SOURCE, [])

    def filter_log(self, file_src):
        """
        Filter the train log.
        :param file_src: str of dir or list of dirs and filenames for train log
        """
        if file_src is None:
            return
        if isinstance(file_src, str):
            self._recognize_full_formatted_train_log(file_src)
            return
        for src in sorted(file_src, key=lambda path: os.path.isdir(path)):
            if os.path.isfile(src):
                self.train_log_files.append(src)
            else:
                self._recognize_partial_formatted_train_log(src)
            if len(self.train_log_files) > self.TRAIN_LOG_LOAD_LIMIT:
                self.train_log_files = self.train_log_files[:self.TRAIN_LOG_LOAD_LIMIT]
                self.logger.warning("As the quantity of input train logs exceed the limit of 20, "
                                    "excess logs are discarded.")
                echo.warning("As the quantity of input train logs exceed the limit of 20, "
                             "excess logs are discarded.")
                break

    def _recognize_full_formatted_train_log(self, file_dir):
        """
        Recursively recognizing and collect full formatted train logs under file_dir
        :param file_dir: train log save dir
        """
        for root, _, files in safe_walk(file_dir):
            for file_name in files:
                if not (self._has_valid_key(file_name) and self._has_valid_suffix(file_name)):
                    continue
                self.train_log_files.append(os.path.join(root, file_name))

    def _recognize_partial_formatted_train_log(self, file_dir):
        """
        Recognize and collect valid train log only in the first-level directory
        :param file_dir: train log save dir
        """
        file_list = safe_list_dir(file_dir)
        for file_name in sorted(file_list):
            if not self._has_valid_suffix(file_name):
                continue
            self.train_log_files.append(os.path.join(file_dir, file_name))

    def _has_valid_key(self, file_name: str) -> bool:
        """
        The valid filename is supposed to have one of the certain keys
        :param file_name: the given file name
        :return: boolean of whether the file name is valid
        """
        return self.RANK_KEY in file_name or self.WORKER_KEY in file_name

    def _has_valid_suffix(self, file_name: str) -> bool:
        """
        The valid filename is supposed to have one of the certain suffixes
        :param file_name: the given file name
        :return: boolean of whether the file name is valid
        """
        return file_name.endswith(self.TXT_END_KEY) or file_name.endswith(self.LOG_END_KEY)


class HostLogSaver(BaseLogSaver):
    LOG_TYPE = "host os log"
    CMD_ARG_KEYS = ["host_log"]

    HOST_KEY = "messages"
    DMESG_KEY = "dmesg"
    SYSMON_KEY = "sysmonitor.log"
    VMCORE_KEY = "vmcore-dmesg.txt"
    SKIP_KEY = "dev-os-"
    DMIDECODE_KEY = "dmidecode.txt"

    def __init__(self):
        """
        Host Log Saver
        """
        super().__init__()
        self.host_log_files = []
        self.host_dmesg_files = []
        self.host_sysmon_files = []
        self.host_vmcore_dmesg_files = []
        self.host_dmidecode_files = []

    def filter_log(self, file_dir: str):
        """
        Filter the host os log
        :param file_dir: host os log save dir
        """
        if not file_dir or not os.path.isdir(file_dir):
            return
        for root, _, files in safe_walk(file_dir):
            # if message log in device_log, skip
            if os.path.basename(root).startswith(self.SKIP_KEY):
                continue
            for file_name in files:
                file_path = os.path.join(root, file_name)
                if file_name.startswith(self.HOST_KEY):
                    self.host_log_files.append(file_path)
                    continue
                if file_name == self.DMESG_KEY:
                    self.host_dmesg_files.append(file_path)
                    continue
                if file_name == self.SYSMON_KEY:
                    self.host_sysmon_files.append(file_path)
                    continue
                if file_name == self.DMIDECODE_KEY:
                    self.host_dmidecode_files.append(file_path)
                    continue
                if file_name == self.VMCORE_KEY:
                    self.host_vmcore_dmesg_files.append(file_path)

    def get_host_log(self) -> list:
        """
        Get the host os log list
        :return: host os log list
        """
        return self.host_log_files if not self.is_sdk_input else self.log_map.get(regular_table.OS_SOURCE, [])

    def get_dmesg_log(self) -> list:
        """
        Get the dmesg log list
        :return: dmesg log list
        """
        return self.host_dmesg_files if not self.is_sdk_input else self.log_map.get(regular_table.OS_DEMESG_SOURCE, [])

    def get_sysmon_log(self) -> list:
        """
        Get the sysmonitor.log list
        :return: sysmonitor.log list
        """
        return self.host_sysmon_files if not self.is_sdk_input else self.log_map.get(regular_table.OS_SYSMON_SOURCE, [])

    def get_vmcore_dmesg_log(self) -> list:
        """
        Get the vmcore-dmesg.txt list
        :return: vmcore-dmesg.txt list
        """
        return self.host_vmcore_dmesg_files if not self.is_sdk_input \
            else self.log_map.get(regular_table.OS_VMCORE_DMESG_SOURCE, [])

    def get_dmidecode_log(self) -> list:
        """
        Get the dmidecode.txt list
        :return: dmidecode.txt list
        """
        return self.host_dmidecode_files if not self.is_sdk_input else self.log_map.get(regular_table.DMI_DECODE_SOURCE,
                                                                                        [])


class BMCLogSaver(BaseLogSaver):
    LOG_TYPE = "bmc log"
    CENTRALIZED_STORAGE_DIRECTORY = "bmc_log"
    CMD_ARG_KEYS = ["bmc_log"]

    FRU_INFO_KEY = "fruinfo.txt"
    MDB_INFO_KEY = "mdb_info.log"
    APP_DUMP_KEY = "AppDump"
    DEVICE_DUMP_KEY = "DeviceDump"
    LOG_DUMP_KEY = "LogDump"

    def __init__(self):
        """
        BMC Log Saver
        """
        super().__init__()
        self.fruinfo_files = []
        self.mdb_info_files = []
        self.bmc_log_list = []
        self.bmc_app_dump_log_list = []
        self.bmc_device_dump_log_list = []
        self.bmc_log_dump_log_list = []

    def filter_log(self, file_dir: str):
        """
        Filter the bmc log
        :param file_dir: bmc log save dir
        """
        if not file_dir or not os.path.isdir(file_dir):
            return
        for root, _, files in safe_walk(file_dir):
            for file_name in files:
                file_path = os.path.join(root, file_name)
                if file_name == self.FRU_INFO_KEY:
                    self.fruinfo_files.append(file_path)
                    continue
                if file_name == self.MDB_INFO_KEY and "chassis" in file_path:
                    self.mdb_info_files.append(file_path)
                    continue
                if not file_name.endswith(".log"):
                    continue
                if self.APP_DUMP_KEY in file_path:
                    self.bmc_app_dump_log_list.append(file_path)
                elif self.DEVICE_DUMP_KEY in file_path:
                    self.bmc_device_dump_log_list.append(file_path)
                elif self.LOG_DUMP_KEY in file_path:
                    self.bmc_log_dump_log_list.append(file_path)

    def get_fruinfo_log(self) -> list:
        """
        Get the bmc fruinfo log list
        :return: fruinfo log list
        """
        return self.fruinfo_files

    def get_mdb_info_log(self) -> list:
        """
        Get the bmc mdb_info log list
        :return: bmc mdb_info log list
        """
        return self.mdb_info_files

    def get_bmc_log_list(self) -> list:
        return self.bmc_log_list if not self.is_sdk_input else self.log_map.get(regular_table.BMC_SOURCE, [])

    def get_bmc_app_dump_log_list(self) -> list:
        return self.bmc_app_dump_log_list if not self.is_sdk_input \
            else self.log_map.get(regular_table.BMC_APP_DUMP_SOURCE, [])

    def get_bmc_device_dump_log_list(self) -> list:
        return self.bmc_device_dump_log_list if not self.is_sdk_input \
            else self.log_map.get(regular_table.BMC_DEVICE_DUMP_SOURCE, [])

    def get_bmc_log_dump_log_list(self) -> list:
        return self.bmc_log_dump_log_list if not self.is_sdk_input \
            else self.log_map.get(regular_table.BMC_LOG_DUMP_SOURCE, [])


class LCNELogSaver(BaseLogSaver):
    """
        采集目录
           |-- lcne_log
                  |-- logfile   # CPU/NPU多个LCNE日志目录名称设置为不同名称
                         |-- log.log
                         |-- log_xxx.log.zip
                  |-- logfile1
                         |-- log.log
                         |-- log_xxx.log.zip
    """
    LOG_TYPE = "lcne log"
    CENTRALIZED_STORAGE_DIRECTORY = "lcne_log"
    CMD_ARG_KEYS = ["lcne_log", "bus_log"]

    DEVM_BDDRVADP_KEY = "devm_bddrvadp.log"
    DEVM_BDDRVADP_DIR = "slot_1/tempdir"
    DIAG_DISPLAY_INFO_KEY = "diag_display_info.txt"
    LOG_PATTERN = r'log_1_\d{13,15}\.log$'

    BUS_KEY = "log"
    LOG_SUFFIX = ".log"
    ZIP_SUFFIX = ".log.zip"

    def __init__(self):
        """
        LCNE Log Saver
        """
        super().__init__()
        self.devm_bddvadp_files = []
        self.diag_display_info_files = []
        self.lcne_log_list = []
        self.bus_log_dict = {}

    def filter_log(self, file_dir: str):
        """
        Filter the lcne log
        :param file_dir: lcne log save dir
        """
        if not file_dir or not os.path.isdir(file_dir):
            return
        for root, _, files in safe_walk(file_dir):
            dir_files = []
            for file_name in files:
                file_path = os.path.join(root, file_name)
                if file_name == self.DEVM_BDDRVADP_KEY and self.DEVM_BDDRVADP_DIR in file_path:
                    self.devm_bddvadp_files.append(file_path)
                    continue
                if file_name == self.DIAG_DISPLAY_INFO_KEY:
                    self.diag_display_info_files.append(file_path)
                    continue
                if file_name.endswith("log.log") or re.fullmatch(self.LOG_PATTERN, file_name):
                    self.lcne_log_list.append(file_path)
                    continue
                if file_name.startswith(self.BUS_KEY) and (
                        file_name.endswith(self.LOG_SUFFIX) or
                        file_name.endswith(self.ZIP_SUFFIX)
                ):
                    # 拼接完整路径并添加到当前目录列表
                    dir_files.append(os.path.join(root, file_name))
                if dir_files:
                    self.bus_log_dict[root] = dir_files

    def get_devm_bddvadp_log(self) -> list:
        return self.devm_bddvadp_files

    def get_diag_display_info_log(self) -> list:
        return self.diag_display_info_files

    def get_lcne_log_list(self) -> list:
        return self.lcne_log_list if not self.is_sdk_input else self.log_map.get(regular_table.LCNE_SOURCE, [])

    def get_bus_log_dict(self):
        return self.bus_log_dict


class DevLogSaver(BaseLogSaver):
    LOG_TYPE = "device log"
    CENTRALIZED_STORAGE_DIRECTORY = "device_log"
    CMD_ARG_KEYS = ["device_log"]

    def __init__(self):
        """
        Device Log Saver
        """
        super().__init__()
        self.slog_dict = dict()
        self.hisi_logs_list = []

    def filter_log(self, dev_log_dir: str):
        """
        Filter device log
        :param dev_log_dir: the device log root dir
        :return: slog and hisi_logs filter result
        """
        if not dev_log_dir or not os.path.isdir(dev_log_dir):
            return
        hisi_log_path, slog_path = "", ""
        for root, dirs, _ in safe_walk(dev_log_dir):
            if not hisi_log_path and "hisi_logs" in dirs:
                hisi_log_path = os.path.join(root, "hisi_logs")
                self._filter_hisi_logs(hisi_log_path)
            if not slog_path and "slog" in dirs:
                slog_path = os.path.join(root, "slog")
                self._filter_slog(slog_path)
            if hisi_log_path and slog_path:
                break

    def get_hisi_logs_list(self):
        """
        Get the hisi logs files list
        :return: device hisi logs files list
        """
        return self.hisi_logs_list if not self.is_sdk_input else self.log_map.get(regular_table.NPU_HISTORY_SOURCE, [])

    def get_slog_dict(self):
        """
        Get the slog files dict
        :return: device slog files dict
        """
        if self.is_sdk_input:
            self._filter_sdk_slog()
        return self.slog_dict

    def _filter_sdk_slog(self):
        """
        Filter slog for slog saved by sdk
        """
        slog_keys = [regular_table.NPU_DEVICE_SOURCE, regular_table.NPU_OS_SOURCE]
        slog_dict = {k: v for k, v in self.log_map.items() if k in slog_keys}
        for source_list in slog_dict.values():
            for file_source in source_list:
                if file_source.path and not file_source.parent_is("event") and not \
                        re.match(r"device-(\d{1,3}|os)", file_source.filename):
                    continue
                self.slog_dict.setdefault(file_source.dir_name, []).append(file_source)
        for file_list in self.slog_dict.values():
            file_list.sort()

    def _filter_hisi_logs(self, hisi_logs_path):
        """
        Filter the hisi_logs
        :param hisi_logs_path: the hisi logs path
        """
        if not hisi_logs_path or not os.path.isdir(hisi_logs_path):
            return
        for device_dir in safe_list_dir(hisi_logs_path):
            history_path = os.path.join(hisi_logs_path, device_dir, regular_table.DEV_NPU_HISI_HISTORY_ORIGIN)
            if os.path.exists(history_path) and os.path.isfile(history_path):
                self.hisi_logs_list.append(history_path)

    def _filter_slog(self, slog_path):
        """
        Filter the slog
        :param slog_path: the slog path
        """
        if not slog_path or not os.path.isdir(slog_path):
            return
        for dev_os_dir in safe_list_dir(slog_path):
            dev_os_path = os.path.join(slog_path, dev_os_dir)
            if not dev_os_dir.startswith(regular_table.DEV_OS_INFO) or not os.path.isdir(dev_os_path):
                continue
            self._update_slog_dev_os_files(dev_os_path)

    def _update_slog_dev_os_files(self, dev_os_path):
        """
        Update the log dir and log list in slog
        :param dev_os_path: dev os dir path in slog
        :return: the files dict in each dev os
        """
        if not dev_os_path or not os.path.isdir(dev_os_path):
            return
        for dir_name in safe_list_dir(dev_os_path):
            if dir_name in ["debug", "run"]:
                dir_path = os.path.join(dev_os_path, dir_name)
                self._update_device_dir(dir_path)
                continue
            if dir_name.startswith(regular_table.DEV_NPU_INFO) and os.path.isdir(os.path.join(dev_os_path, dir_name)):
                dir_path = os.path.join(dev_os_path, dir_name)
                dir_list = safe_list_dir(dir_path)
                self.slog_dict.update({dir_path: sorted(dir_list)})

    def _update_device_dir(self, device_path):
        """
        Update the log dir and log list in debug or run
        :param device_path: dev os dir path in debug or run
        :return: the files dict in each dev os
        """
        if not device_path or not os.path.isdir(device_path):
            return
        for sub_dir in safe_list_dir(device_path):
            if sub_dir == "event" or re.match(r"device-(\d{1,3}|os)", sub_dir):
                device_dir = os.path.join(device_path, sub_dir)
                device_list = safe_list_dir(device_dir)
                self.slog_dict.update({device_dir: sorted(device_list)})


class DlLogSaver(BaseLogSaver):
    LOG_TYPE = "dl log"
    CENTRALIZED_STORAGE_DIRECTORY = "dl_log"
    CMD_ARG_KEYS = ["dl_log"]

    DEVICE_PLUGIN_KEY = "devicePlugin"
    NODED_KEY = "noded"
    VOLCANO_SCHEDULER_KEY = "volcano-scheduler"
    VOLCANO_CONTROLLER_KEY = "volcano-controller"
    DOCKER_RUNTIME_RUN_KEY = "runtime-run"
    DOCKER_HOOK_RUN_KEY = "hook-run"
    DOCKER_RUNTIME_DIRECTORY = "docker-runtime"
    NPU_EXPORTER_KEY = "npu-exporter"
    MINDIO_KEY = "ttp_log"
    LOG_SUFFIX = ".log"
    DL_LOG_LABEL_LIST = [
        DEVICE_PLUGIN_KEY, NODED_KEY, VOLCANO_SCHEDULER_KEY, VOLCANO_CONTROLLER_KEY, NPU_EXPORTER_KEY, MINDIO_KEY
    ]

    def __init__(self):
        """
        DL Log Saver
        """
        super().__init__()
        self.device_plugin_list = []
        self.noded_log_list = []
        self.volcano_scheduler_list = []
        self.volcano_controller_list = []
        self.docker_runtime_list = []
        self.npu_exporter_list = []
        self.mindio_log_list = []

    def filter_log(self, file_dir: str):
        """
        Distribute a valid dl directories path to corresponding handling functions
        :param file_dir: a dir to store various dl log dir
        """
        if not file_dir or not os.path.isdir(file_dir):
            return
        dir_list = safe_list_dir(file_dir)
        for dir_name in dir_list:
            if dir_name in self.DL_LOG_LABEL_LIST or self.DOCKER_RUNTIME_DIRECTORY in dir_name:
                self._record_dl_log(os.path.join(file_dir, dir_name))

    def get_device_plugin_list(self):
        return self.device_plugin_list if not self.is_sdk_input \
            else self.log_map.get(regular_table.DEVICEPLUGIN_SOURCE, [])

    def get_volcano_scheduler_list(self):
        return self.volcano_scheduler_list if not self.is_sdk_input \
            else self.log_map.get(regular_table.VOLCANO_SCHEDULER_SOURCE, [])

    def get_volcano_controller_list(self):
        return self.volcano_controller_list if not self.is_sdk_input \
            else self.log_map.get(regular_table.VOLCANO_CONTROLLER_SOURCE, [])

    def get_docker_runtime_list(self):
        return self.docker_runtime_list if not self.is_sdk_input \
            else self.log_map.get(regular_table.DOCKER_RUNTIME_SOURCE, [])

    def get_npu_exporter_list(self):
        return self.npu_exporter_list if not self.is_sdk_input \
            else self.log_map.get(regular_table.NPU_EXPORTER_SOURCE, [])

    def get_noded_list(self):
        return self.noded_log_list if not self.is_sdk_input else self.log_map.get(regular_table.NODEDLOG_SOURCE, [])

    def get_mindio_log_list(self):
        return self.mindio_log_list if not self.is_sdk_input else self.log_map.get(regular_table.MINDIO_SOURCE, [])

    def _record_dl_log(self, dir_path):
        """
        Record valid dl log file paths to the responding list
        :param dir_path: a dir stores device plugin logs
        """
        for root, _, files in safe_walk(dir_path):
            for filename in files:
                if filename.startswith(self.DEVICE_PLUGIN_KEY) and filename.endswith(self.LOG_SUFFIX):
                    self.device_plugin_list.append(os.path.join(root, filename))
                    continue
                if filename.startswith(self.NODED_KEY) and filename.endswith(self.LOG_SUFFIX):
                    self.noded_log_list.append(os.path.join(root, filename))
                    continue
                if filename.startswith(self.VOLCANO_SCHEDULER_KEY) and filename.endswith(self.LOG_SUFFIX):
                    self.volcano_scheduler_list.append(os.path.join(root, filename))
                    continue
                if filename.startswith(self.VOLCANO_CONTROLLER_KEY) and filename.endswith(self.LOG_SUFFIX):
                    self.volcano_controller_list.append(os.path.join(root, filename))
                    continue
                if (filename.startswith(self.DOCKER_RUNTIME_RUN_KEY) or filename.startswith(self.DOCKER_HOOK_RUN_KEY)) \
                        and filename.endswith(self.LOG_SUFFIX):
                    self.docker_runtime_list.append(os.path.join(root, filename))
                    continue
                if filename.startswith(self.NPU_EXPORTER_KEY) and filename.endswith(self.LOG_SUFFIX):
                    self.npu_exporter_list.append(os.path.join(root, filename))
                    continue
                if filename.startswith(self.MINDIO_KEY):
                    self.mindio_log_list.append(os.path.join(root, filename))


class AMCTLogSaver(BaseLogSaver):
    LOG_TYPE = "amct log"
    CENTRALIZED_STORAGE_DIRECTORY = "amct_log"
    CMD_ARG_KEYS = ["amct_log"]

    AMCT_KEY = "amct"
    LOG_SUFFIX = ".log"

    def __init__(self):
        """
        AMCT Log Saver
        """
        super().__init__()
        self.amct_log_list = []

    def filter_log(self, file_dir: str):
        """
        Distribute valid amct directories path to corresponding handling functions
        :param file_dir: a dir to store various amct log dir
        """
        if not file_dir or not os.path.isdir(file_dir):
            return
        for root, _, files in safe_walk(file_dir):
            for filename in files:
                if filename.startswith(self.AMCT_KEY) and filename.endswith(self.LOG_SUFFIX):
                    self.amct_log_list.append(os.path.join(root, filename))

    def get_amct_log(self):
        return self.amct_log_list if not self.is_sdk_input else self.log_map.get(regular_table.AMCT_SOURCE, [])


class MindieLogSaver(BaseLogSaver):
    LOG_TYPE = "mindie log"
    CENTRALIZED_STORAGE_DIRECTORY = "mindie"
    CMD_ARG_KEYS = ["mindie"]

    MINDIE_KEY = "mindie-"
    LOG_SUFFIX = ".log"
    CERT_MODULE_KEY = "mindie-cert"
    MINDIE_CLUSTER_KEY = "mindie_cluster_log"

    def __init__(self):
        super().__init__()
        self.mindie_log_list = []
        self.mindie_cluster_log_list = []

    def filter_log(self, file_dir: str):
        """
        Filter MindIE log file into the list, excluded faults are not listed
        :param file_dir: a dir to store mindie log files
        """
        if not file_dir or not os.path.isdir(file_dir):
            return
        for root, dirs, files in safe_walk(file_dir):
            # 添加 mindIE 通过k8s导出的日志，日志包含实例信息，用于切分实例
            if self.MINDIE_CLUSTER_KEY in dirs:
                mindie_cluster_dir = os.path.join(root, self.MINDIE_CLUSTER_KEY)
                self.filter_mindie_cluster_log(mindie_cluster_dir)
            for filename in files:
                if os.path.basename(root) == self.MINDIE_CLUSTER_KEY:
                    continue
                if filename.startswith(self.MINDIE_KEY) and filename.endswith(self.LOG_SUFFIX) and \
                        not filename.startswith(self.CERT_MODULE_KEY):
                    self.mindie_log_list.append(os.path.join(root, filename))

    def filter_mindie_cluster_log(self, file_dir: str):
        for root, _, files in safe_walk(file_dir):
            for filename in files:
                self.mindie_cluster_log_list.append(os.path.join(root, filename))

    def get_mindie_log_list(self):
        return self.mindie_log_list if not self.is_sdk_input else self.log_map.get(regular_table.MINDIE_SOURCE, [])

    def get_mindie_clu_log_list(self):
        return self.mindie_cluster_log_list if not self.is_sdk_input \
            else self.log_map.get(regular_table.MINDIE_CLUSTER_SOURCE, [])


@dataclass
class LogInfoSaver:
    source_file: str
    path: str
    device_id: int
    log_lines: list
    modification_time: str
    component: str

    def __lt__(self, other) -> bool:
        if not isinstance(other, LogInfoSaver):
            return NotImplemented
        return self.filename < other.filename

    @property
    def filename(self) -> str:
        if os.sep not in self.path:
            return self.path
        return os.path.basename(self.path)

    @property
    def dir_name(self) -> str:
        if os.sep not in self.path:
            return self.path
        return os.path.dirname(self.path)

    @property
    def device_id_str(self) -> str:
        return str(self.device_id) if self.device_id != -1 else "Unknown"

    def has_parent(self, keyword: str) -> bool:
        if os.sep not in self.path:
            return False
        split_path = self.path.split(os.sep)
        return keyword in split_path

    def parent_is(self, keyword: str) -> bool:
        return os.path.basename(os.path.dirname(self.path.rstrip(os.sep))) == keyword


class HostSnInfo:
    def __init__(self, log_dir, serial_number):
        self.log_dir = log_dir
        self.serial_number = serial_number


class BMCSNInfo:
    def __init__(self, log_dir, serial_number, board_serial_number):
        self.log_dir = log_dir
        self.serial_number = serial_number
        self.board_serial_number = board_serial_number


class LCNESNInfo:
    def __init__(self, log_dir, board_serial_number):
        self.log_dir = log_dir
        self.board_serial_number = board_serial_number


class EmptyObject:
    def __init__(self):
        self.__dict__ = {}


class SuperpodInfoSaver:
    def __init__(self):
        self.host_info_by_sn = {}  # {sn:HostSnInfo}
        self.bmc_info_by_board_sn = {}  # {board_sn:BMCSNInfo}
        self.bmc_info_by_sn = {}  # {sn:BMCSNInfo}
        self.lcne_info_by_board_sn = {}  # {board_sn:LCNESNInfo}

    def add_host_info(self, host_info_instance):
        if host_info_instance.serial_number:
            self.host_info_by_sn[host_info_instance.serial_number] = host_info_instance

    def add_bmc_sn_info(self, bmc_sn_info_instance):
        if bmc_sn_info_instance.board_serial_number:
            self.bmc_info_by_board_sn[bmc_sn_info_instance.board_serial_number] = bmc_sn_info_instance
        if bmc_sn_info_instance.serial_number:
            self.bmc_info_by_sn[bmc_sn_info_instance.serial_number] = bmc_sn_info_instance

    def add_lcne_sn_info(self, lcne_sn_info_instance):
        if lcne_sn_info_instance.board_serial_number:
            self.lcne_info_by_board_sn[lcne_sn_info_instance.board_serial_number] = lcne_sn_info_instance

    def get_related(self, sn_info_instance):
        if isinstance(sn_info_instance, HostSnInfo):
            return self._find_from_host(sn_info_instance)
        elif isinstance(sn_info_instance, BMCSNInfo):
            return self._find_from_bmc(sn_info_instance)
        elif isinstance(sn_info_instance, LCNESNInfo):
            return self._find_from_lcne(sn_info_instance)
        return None, None

    def find_from_bmc_worker_name(self, bmc_worker_name):
        for bmc_sn_info_instance in self.bmc_info_by_sn.values():
            log_dir_split = bmc_sn_info_instance.log_dir.split("/")
            instance_worker_name = log_dir_split[1] if len(log_dir_split) > 1 else None
            if instance_worker_name == bmc_worker_name:
                return self._find_from_bmc(bmc_sn_info_instance)
        return None, None

    def find_from_lcne_worker_name(self, lcne_worker_name):
        for lcne_sn_info_instance in self.lcne_info_by_board_sn.values():
            log_dir_split = lcne_sn_info_instance.log_dir.split("/")
            insinstance_worker_name = log_dir_split[1] if len(log_dir_split) > 1 else None
            if insinstance_worker_name == lcne_worker_name:
                return self._find_from_lcne(lcne_sn_info_instance)
        return None, None

    def save_to_json(self, save_path, file_name):
        topo_infos = []
        # Traverse all bmc_info_by_board_sn as connection hubs
        for bmc_sn_info_info_instance in self.bmc_info_by_board_sn.values():
            topo_info = {
                "host": (self.host_info_by_sn.get(bmc_sn_info_info_instance.serial_number) or EmptyObject()).__dict__,
                "bmc": (bmc_sn_info_info_instance or EmptyObject()).__dict__,
                "lcne": (self.lcne_info_by_board_sn.get(
                    bmc_sn_info_info_instance.board_serial_number) or EmptyObject()).__dict__
            }
            topo_infos.append(topo_info)
        with safe_write_open(os.path.join(save_path, file_name), mode='w+', encoding='utf-8') as file_stream:
            file_stream.write(json.dumps({"super_pod_topo_info": topo_infos}, ensure_ascii=False))

    def _find_from_bmc(self, bmc_sn_info_info_instance):
        host_info_instance = self.host_info_by_sn.get(bmc_sn_info_info_instance.serial_number)
        lcne_sn_info_instance = self.lcne_info_by_board_sn.get(bmc_sn_info_info_instance.board_serial_number)
        return host_info_instance, lcne_sn_info_instance

    def _find_from_lcne(self, lcne_sn_info_instance):
        bmc_sn_info_info_instance = self.bmc_info_by_board_sn.get(lcne_sn_info_instance.board_serial_number)
        host_info_instance = bmc_sn_info_info_instance and self.host_info_by_sn.get(
            bmc_sn_info_info_instance.serial_number)
        return host_info_instance, bmc_sn_info_info_instance

    def _find_from_host(self, host_info_instance):
        bmc_sn_info_info_instance = self.bmc_info_by_sn.get(host_info_instance.serial_number)
        lcne_sn_info_instance = bmc_sn_info_info_instance and self.lcne_info_by_board_sn.get(
            bmc_sn_info_info_instance.board_serial_number)
        return bmc_sn_info_info_instance, lcne_sn_info_instance


@dataclass
class ServiceInfo:
    pid: str
    worker_name: str
    logic_device_id = ""
    phy_device_id = ""
    device_ip = ""

    def update_device_info(self, info_dict):
        """
        Update the base device info from base info dict
        :param info_dict: the base info
        """
        self.logic_device_id = info_dict.get("logic_device_id", "")
        self.phy_device_id = info_dict.get("phy_device_id", "")
        self.device_ip = info_dict.get("device_ip", "")

    def get_device_info_dict(self):
        """
        Get device info
        """
        return {
            "logic_device_id": self.logic_device_id,
            "phy_device_id": self.phy_device_id,
            "device_ip": self.device_ip
        }


@dataclass
class InferGroup:
    infer_group_name: str
    device_num: int


class ParsedDataSaver:
    def __init__(self, data_path, args):
        """
        Parse data saver;
        :param data_path: fault diag data dir path
        """
        self.data_path = data_path
        self.worker_path_dict = dict()

        self.infer_task_flag = False
        self.board_sn_exist_tag = False
        self.bmc_dir_exist = False
        self.lcne_dir_exist = False

        self.cluster_info = dict()
        self.pid_device_dict = dict()
        self.ip_infer_group = dict()
        self.container_worker_map = dict()
        self.collect_infer_group = []
        self.infer_instance = ""
        self.infer_group_2_worker_path = dict()
        self.mindie_parse_result = MindIEParseResult()
        self.mindie_diag_result = MindIEDiagResult()
        self.infer_groups_device_map = dict()

        self.bmc_path_dict = dict()
        self.lcne_path_dict = dict()
        self.all_worker_path_dict = dict()
        self.scene = args.scene if hasattr(args, 'scene') else None
        self.output_path = args.output_path if hasattr(args, 'output_path') else None
        self.super_pod_info_saver = SuperpodInfoSaver()
        self.init_worker_data()
        self.init_infer_task()
        if self.scene == "super_pod":
            fault_diag_result_dir = os.path.join(self.output_path, "fault_diag_result")
            os.makedirs(fault_diag_result_dir, 0o700, exist_ok=True)
            self.super_pod_info_saver.save_to_json(fault_diag_result_dir, "topo_info.json")

    @staticmethod
    def get_server_info_dict(worker_dir):
        """
        Get the server info from server_info.json in each worker dir
        :param worker_dir: worker dir
        :return: device info
        """
        server_info_file = os.path.join(worker_dir, regular_table.SERVER_INFO_FILE)
        if not os.path.exists(server_info_file) or not os.path.isfile(server_info_file):
            return dict()
        server_info_dict = safe_read_json(server_info_file)
        return server_info_dict

    def init_worker_data(self):
        """
        Init parsed data by worker dir
        """
        if not self.data_path or not os.path.isdir(self.data_path):
            return
        for worker_dir in safe_list_dir(self.data_path):
            server_info_file = os.path.join(self.data_path, worker_dir, regular_table.SERVER_INFO_FILE)
            if os.path.isfile(server_info_file):
                server_info_dict = safe_read_json(server_info_file)
                self.board_sn_exist_tag = BOARD_SERIAL_NUMBER in server_info_dict

            if self.scene == "super_pod":
                self.init_super_pod(worker_dir)
            else:
                worker_dir_path = os.path.join(self.data_path, worker_dir)
                if not os.path.isdir(worker_dir_path):
                    continue
                self.worker_path_dict.update({worker_dir: worker_dir_path})
                self.all_worker_path_dict.update({worker_dir: worker_dir_path})
        if self.scene == "super_pod" and not (self.bmc_dir_exist and self.lcne_dir_exist):
            fd_logger.error(
                "The bmc or lcne dir not exist in {}, not applicable to super_pod scenario diagnosis.".format(
                    self.data_path))
            raise InnerError(
                "The bmc or lcne dir not exist in {}, not applicable to super_pod scenario diagnosis.".format(
                    self.data_path))

    def init_super_pod(self, worker_dir):
        for super_dir in safe_list_dir(os.path.join(self.data_path, worker_dir)):
            worker_dir_path = os.path.join(self.data_path, worker_dir, super_dir)
            server_info_dict = self.get_server_info_dict(worker_dir_path)
            if not os.path.isdir(worker_dir_path):
                continue
            if 'bmc' in worker_dir:
                self.bmc_dir_exist = True
                self.bmc_path_dict.update({f"BMC:{super_dir}": worker_dir_path})
                self.all_worker_path_dict.update({f"BMC:{super_dir}": worker_dir_path})
                self.super_pod_info_saver.add_bmc_sn_info(
                    BMCSNInfo(os.path.join(worker_dir, super_dir),
                              server_info_dict.get(SERIAL_NUMBER, ""),
                              server_info_dict.get(BOARD_SERIAL_NUMBER, "")))
            elif 'lcne' in worker_dir:
                self.lcne_dir_exist = True
                self.lcne_path_dict.update({f"LCNE:{super_dir}": worker_dir_path})
                self.all_worker_path_dict.update({f"LCNE:{super_dir}": worker_dir_path})
                self.super_pod_info_saver.add_lcne_sn_info(LCNESNInfo(os.path.join(worker_dir, super_dir),
                                                                      server_info_dict.get(
                                                                          BOARD_SERIAL_NUMBER, "")))
            else:
                self.worker_path_dict.update({super_dir: worker_dir_path})
                self.all_worker_path_dict.update({super_dir: worker_dir_path})
                self.super_pod_info_saver.add_host_info(
                    HostSnInfo(os.path.join(worker_dir, super_dir), server_info_dict.get(SERIAL_NUMBER, "")))

    def init_infer_task(self):
        """
        Init infer task
        """
        # 查找mindie-cluster-info.json文件，判断是否为推理任务
        for worker_dir_path in self.worker_path_dict.values():
            rc_infer_file = os.path.join(worker_dir_path, regular_table.INFER_FILE)
            if os.path.exists(rc_infer_file):
                self.cluster_info = safe_read_json(rc_infer_file)
                self.infer_task_flag = True
                break
        if not self.infer_task_flag:
            return
        # 解析server-info.json，保存container_ip和device实例
        ip_device_dict = {}  # {container_ip: [device_instance...]}
        for worker_name, worker_dir_path in self.worker_path_dict.items():
            container_info = safe_read_json(os.path.join(worker_dir_path, regular_table.CONTAINER_FILE))
            container_ip = container_info.get("container_ip", "")
            self.container_worker_map.update({container_ip: worker_name})
            for pid, device_info in container_info.get("device_info", {}).items():
                device_instance = ServiceInfo(pid, worker_name)
                device_instance.update_device_info(device_info)
                ip_device_dict.setdefault(container_ip, []).append(device_instance)
                # 使用worker_name+pid作为key防止不同节点同pid覆盖
                self.pid_device_dict.update({(worker_name + pid): device_instance})
            self.mindie_parse_result.link_error_info_map.update(container_info.get("link_error_info", {}))
            self.mindie_parse_result.pull_kv_error_map.update(container_info.get("pull_kv_error_info", {}))
        # 处理mindie-cluster-info.json内容，保存container_ip和infer_group实例
        infer_group_dict = {}
        for infer_group_name, ip_list in self.cluster_info.items():
            for ip in ip_list:
                device_instance_list = ip_device_dict.get(ip, [])
                # 计算每个推理组的device_num
                if infer_group_name in infer_group_dict:
                    infer_group_instance = infer_group_dict.get(infer_group_name)
                    infer_group_instance.device_num += len(device_instance_list)
                else:
                    infer_group_instance = InferGroup(infer_group_name, len(device_instance_list))
                    infer_group_dict.update({infer_group_name: infer_group_instance})
                self.ip_infer_group.update({ip: infer_group_instance})
        # 获取采集到日志的推理组
        for worker_dir_path in self.worker_path_dict.values():
            container_info = safe_read_json(os.path.join(worker_dir_path, regular_table.CONTAINER_FILE))
            container_ip = container_info.get("container_ip", "")
            infer_group_instance = self.ip_infer_group.get(container_ip)
            if not infer_group_instance:
                raise InnerError("Divide P/D instance failed. Please check the rankTable info from Pod log")
            self.infer_group_2_worker_path.setdefault(infer_group_instance.infer_group_name, []).append(worker_dir_path)
            if not infer_group_instance or infer_group_instance.infer_group_name in self.collect_infer_group:
                continue
            self.collect_infer_group.append(infer_group_instance.infer_group_name)
        self.coll_global_device_ip_info()

    def coll_global_device_ip_info(self):
        """
        Collecting Global vNIC information
        """
        for worker_name, worker_dir_path in self.worker_path_dict.items():
            path_list = collect_parse_results(worker_dir_path, "rc-parser.json")
            rc_parser_file = path_list[0] if path_list else ""
            rc_parser_dict = safe_read_json(rc_parser_file)
            for pid_info in rc_parser_dict.values():
                base_info = pid_info.get("base", dict())
                phy_device_id = base_info.get("phy_device_id", "")
                device_ip = base_info.get("device_ip")
                if device_ip:
                    device_id = phy_device_id or base_info.get("logic_device_id", "")
                    self.infer_groups_device_map.update({device_ip: (worker_name, device_id)})
                vNic_ip = base_info.get("vNic_ip")
                if vNic_ip:
                    self.infer_groups_device_map.update({vNic_ip: (worker_name, phy_device_id)})

    def get_worker_plog_dict(self):
        """
        Get the parsed plog path list categorized by worker id;
        :return: worker plog dict
        """
        worker_plog_dict = dict()
        for worker_name, worker_dir in self.worker_path_dict.items():
            plog_list = []
            for file in safe_list_dir(worker_dir):
                if file.startswith("plog-parser-"):
                    plog_list.append(os.path.join(worker_dir, file))
            worker_plog_dict.update({worker_name: plog_list})
        return worker_plog_dict

    def get_device_info_from_json(self, worker_name):
        """
        Get the device info from device_ip_info.json in each worker dir
        :param worker_name: worker name
        :return: device info
        """
        worker_dir = self.get_worker_dir_path(worker_name)
        device_file = os.path.join(worker_dir, regular_table.DEVICE_IP_FILE)
        if not os.path.exists(device_file) or not os.path.isfile(device_file):
            return dict()
        device_info_dict = safe_read_json(device_file)
        return device_info_dict

    def get_worker_dir_path(self, worker_name):
        """
        Get the worker dir path by worker id;
        :param worker_name: worker name
        :return: worker dir
        """
        return self.all_worker_path_dict.get(worker_name, "")

    def get_all_worker_dir_path(self):
        """
        Get all worker dir path categorized by worker id;
        :return: worker dir dict
        """
        return self.worker_path_dict


@dataclass
class MatchedCustomInfo:
    custom_file_info: CustomFileInfo = None
    custom_log_list: List[str] = field(default_factory=list)
    sdk_custom_log_map: Dict[str, List[LogInfoSaver]] = None


class CustomLogSaver(BaseLogSaver):
    LOG_TYPE = "custom log"
    CMD_ARG_KEYS = ["custom_log"]

    def __init__(self):
        super().__init__()
        self.custom_config = get_config_info() if os.path.exists(CUSTOM_CONFIG_PATH) else ConfigInfo()
        self.custom_info_list = []

    def filter_log(self, file_dir: str):
        """
        Filter custom log file
        :param file_dir: custom log dir path
        """
        if not file_dir or not os.path.isdir(file_dir):
            return
        for file_info in self.custom_config.custom_parse_file:
            file_path_glob = os.path.join(file_dir, file_info.file_path_glob)
            loop_num = 0
            each_custom_log_list = []
            for file_path in glob.iglob(file_path_glob):
                try:
                    file_path = check_symlink(file_path)
                except (PathError, FileNotExistError) as err:
                    logger.warning("Failed to check custom parsing file path: [%s], error: %s", file_path, str(err))
                    continue
                if not os.path.isfile(file_path):
                    logger.warning("The path should be a file: %s", file_path)
                    continue
                if loop_num > WORKER_MAX_NUM:
                    logger.warning("The number of files matched by %s exceeds %s.", file_path_glob, WORKER_MAX_NUM)
                    break
                each_custom_log_list.append(file_path)
                loop_num += 1

            matched_info = MatchedCustomInfo(custom_file_info=file_info, custom_log_list=each_custom_log_list)
            self.custom_info_list.append(matched_info)

    def get_custom_log_list(self) -> list:
        """
        Get custom log list for kg parse
        :return: custom log list
        """
        log_list = []
        for info in self.custom_info_list:
            log_list.extend(info.custom_log_list)
        return log_list

    def get_custom_info_list(self) -> list:
        """
        Get custom info list for kg parse
        :return: custom info list
        """
        if self.is_sdk_input:
            matched_info = MatchedCustomInfo(sdk_custom_log_map=self.log_map)
            self.custom_info_list.append(matched_info)
        return self.custom_info_list
