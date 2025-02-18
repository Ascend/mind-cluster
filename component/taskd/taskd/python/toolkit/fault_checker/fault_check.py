#!/usr/bin/python3
# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
import os
import json
import shutil
import signal
import time

from taskd.python.toolkit.config.path import TORCH_EXTENSIONS_CACHE_DIR
from taskd.python.toolkit.constants import constants
from taskd.python.toolkit.constants.constants import (WATCHDOG_ENV, WATCHDOG_ON,
                                                                              WATCHDOG_OFF, ENABLE_RANKTABLE_ENV)
from taskd.python.toolkit.validator.file_process import safe_get_file_info
from taskd.python.toolkit.logger.log import run_log


def clean_before_restart():
    """
    Clear related resources before restarting the process.
    """

    if not os.path.exists(TORCH_EXTENSIONS_CACHE_DIR):
        return

    if os.path.isfile(TORCH_EXTENSIONS_CACHE_DIR) or os.path.islink(TORCH_EXTENSIONS_CACHE_DIR):
        os.remove(TORCH_EXTENSIONS_CACHE_DIR)
    if os.path.isdir(TORCH_EXTENSIONS_CACHE_DIR):
        shutil.rmtree(TORCH_EXTENSIONS_CACHE_DIR, ignore_errors=True)


def grace_exit_pids(pids):
    """
    grace exit pid list
    """
    if not isinstance(pids, dict):
        raise ValueError("pids type is invalid")

    for pid in pids.values():
        try:
            process_dir = os.path.join('/proc', str(pid))
            if os.path.exists(process_dir):
                os.kill(pid, signal.SIGTERM)
        except ProcessLookupError:
            run_log.warning(f"{ProcessLookupError} occur when kill the process of {pid}")
        except Exception as e:
            run_log.error(f"An unexpected error {e} occur when kill the process of {pid}")
            raise e


def force_exit_pids(pids):
    """
    force exit pid list
    """
    if isinstance(pids, dict):
        pids = pids.values()
    for pid in pids:
        try:
            process_dir = os.path.join('/proc', str(pid))
            if os.path.exists(process_dir):
                os.kill(pid, signal.SIGKILL)
        except ProcessLookupError:
            run_log.warning(f"{ProcessLookupError} occur when kill the process of {pid}")
        except Exception as e:
            run_log.error(f"An unexpected error {e} occur when kill the process of {pid}")
            raise e


def all_pid_stopped(pids):
    """
    Return true if all target process stopped
    """
    if isinstance(pids, dict):
        pids = pids.values()
    for pid in pids:
        process_dir = os.path.join('/proc', str(pid))
        if os.path.exists(process_dir):
            return False
    return True


def stop_pids(pids):
    start_wait = time.time()
    while (not all_pid_stopped(pids)) and (time.time() - start_wait < constants.GRACE_TIME_OUT):
        time.sleep(constants.SLEEP_GAP)
    if not all_pid_stopped(pids):
        run_log.warning("wait grace exit time-out")
        force_exit_pids(pids)


class ResetCmData:
    def __init__(self, fault_ranks, retry_time, grace_exit, restart_type, fault_flush: bool = False):
        self.fault_ranks = fault_ranks
        self.retry_time = retry_time
        self.grace_exit = grace_exit
        self.restart_type = restart_type
        self.fault_flush = fault_flush


class FaultStatus:
    def __init__(self, fault_ranks: list, fault_status: bool, unrecovered_status: bool, retry_status: bool):
        self.local_ranks = fault_ranks
        self.is_fault = fault_status
        self.is_unrecovered = unrecovered_status
        self.is_retried = retry_status


class FaultProcessor:
    def __init__(self):
        self.reset_cm_path = constants.RESET_CONFIG_PATH
        self.rank_version_path = constants.RANK_TABLE_VERSION_PATH
        self.restart_type_path = constants.RESTART_TYPE_PATH
        self.pre_retry_time = 0
        self.retry_time = 0
        self.grace_exit = 0
        self.restart_type = ""
        self.pre_fault_ranks = []
        self.fault_ranks = []
        self.rank_table_version = 0
        self.check_watchdog()

    @staticmethod
    def check_watchdog():
        if os.environ.get(WATCHDOG_ENV, WATCHDOG_ON) == WATCHDOG_ON:
            run_log.warning("The watchdog features may cause process-level recovery to fail."
                            "It is recommended to be turned off. "
                            f"Please set the environment variable {WATCHDOG_ENV}={WATCHDOG_OFF}")

    @staticmethod
    def _get_rank_id(fault_ranks: dict) -> int:
        return fault_ranks.get(constants.KEY_RANK_ID)

    def read_rank_table_version(self) -> int:
        version = safe_get_file_info(self.rank_version_path).strip()
        if not version.isdigit():
            return -1
        return int(version)

    def wait_to_start(self, worker_group) -> bool:
        local_ranks = [worker.global_rank for worker in worker_group.workers]
        reset_data = self._get_reset_info_from_cm()
        fault_ranks, retry_time = reset_data.fault_ranks, reset_data.retry_time
        fault_flush = reset_data.fault_flush
        self.pre_retry_time = retry_time
        if fault_flush:
            return False
        if not fault_ranks:
            return True

        for fault_rank in fault_ranks:
            if constants.KEY_RANK_ID not in fault_rank or constants.KEY_STATUS not in fault_rank:
                continue
            rank_id = fault_rank.get(constants.KEY_RANK_ID)
            status = fault_rank.get(constants.KEY_STATUS)
            if rank_id in local_ranks and status == constants.VALUE_FAULT:
                return False
        return True

    def is_recovered(self) -> bool:
        for fault_rank in self.fault_ranks:
            if constants.KEY_STATUS not in fault_rank.keys():
                warn_info = f"can not get status from {fault_rank}, skipping checking reset phrase for this rank"
                run_log.warning(warn_info)
                continue
            status = fault_rank.get(constants.KEY_STATUS)
            if status != constants.VALUE_RECOVERED:
                uncovered_info = f"{fault_rank} is not recovered yet"
                run_log.warning(uncovered_info)
                return False

        # only judge rank table version while rank file exists and training job has enabled ranktable
        # pod rescheduling will try to wait for new version, so need to judge restartType as well,
        # of rank table version, only vcjob depends on ranktable version, acjob will not have such file
        run_log.info(f"ranktable version exist:{os.path.exists(self.rank_version_path)}")
        run_log.info(f"self.restart_type: {self.restart_type}")
        if os.getenv(ENABLE_RANKTABLE_ENV) != "" and os.path.exists(self.rank_version_path) \
                and self.restart_type == constants.VALUE_RESTART_RESCHEDULE_TYPE:
            file_rank_version = self.read_rank_table_version()
            run_log.info(f"file_rank_version:{file_rank_version}")
            run_log.info(f"self.rank_table_version:{self.rank_table_version}")
            file_rank_version = self.read_rank_table_version()
            if file_rank_version <= self.rank_table_version:
                warn_info = f"rank table version is {file_rank_version} while self.rank_version " \
                            f"is {self.rank_table_version}, maybe rank table file in container is " \
                            f"still not updated in path {self.rank_version_path}"
                run_log.warning(warn_info)
                return False
            self.rank_table_version = file_rank_version

        # if all fault ranks are recovered, torch fault_checker will restart workers. update recorded retry time and fault ranks
        recovered_infos = f'all fault recovered, updating fault_ranks={self.fault_ranks},' \
                          f' retry_time={self.retry_time}, restart_type={self.restart_type}'
        run_log.warning(recovered_infos)
        self.pre_retry_time = self.retry_time
        self.pre_fault_ranks = self.fault_ranks
        return True

    def get_fault_status(self, worker_group) -> FaultStatus:
        fault_local_ranks = []
        fault_status = False
        unrecovered_status = False
        retry_status = False
        local_worker_ranks = [worker.global_rank for worker in worker_group.workers]
        self._update_reset_info()
        if self.retry_time > self.pre_retry_time:
            retry_status = True
        if self.pre_fault_ranks != self.fault_ranks:
            for fault_rank in self.fault_ranks:
                if constants.KEY_STATUS not in fault_rank.keys():
                    warn_info = f"can not get status from {fault_rank},skipping checking reset phrase for this rank"
                    run_log.warning(warn_info)
                    continue
                rank_id = fault_rank.get(constants.KEY_RANK_ID)
                status = fault_rank.get(constants.KEY_STATUS)
                if status == constants.VALUE_FAULT and rank_id in local_worker_ranks:
                    fault_local_ranks.append(rank_id)
                    fault_status = True
                if status == constants.VALUE_UNRECOVERED or status == constants.VALUE_RECOVERED:
                    unrecovered_status = True
        return FaultStatus(fault_local_ranks, fault_status, unrecovered_status, retry_status)

    def update_fault_info(self):
        self.pre_retry_time = self.retry_time
        self.pre_fault_ranks = []

    def get_remain_retry_time(self, max_retry_times: int) -> int:
        if max_retry_times > self.retry_time:
            return max_retry_times - self.retry_time
        return 0

    def _get_reset_info_from_cm(self) -> ResetCmData:
        file_content = self._get_reset_config()
        fault_ranks = []
        retry_time = 0
        grace_exit = 0
        restart_type = ""
        fault_flush = False
        if constants.KEY_FAULT_FLUSH in file_content.keys() \
                and isinstance(file_content[constants.KEY_FAULT_FLUSH], bool):
            fault_flush = file_content[constants.KEY_FAULT_FLUSH]
        if constants.KEY_RANK_LIST in file_content.keys() and isinstance(file_content[constants.KEY_RANK_LIST], list):
            fault_ranks = []
            for raw_rank in file_content[constants.KEY_RANK_LIST]:
                if isinstance(raw_rank, dict):
                    fault_ranks.append(raw_rank)
        if constants.KEY_RETRY_TIME in file_content.keys() and isinstance(file_content[constants.KEY_RETRY_TIME], int):
            retry_time = file_content[constants.KEY_RETRY_TIME]
        if constants.KEY_GRACE_EXIT in file_content.keys() and isinstance(file_content[constants.KEY_GRACE_EXIT], int):
            grace_exit = file_content[constants.KEY_GRACE_EXIT]
        if (constants.KEY_RESTART_TYPE in file_content.keys() and
                isinstance(file_content[constants.KEY_RESTART_TYPE], str)):
            restart_type = file_content[constants.KEY_RESTART_TYPE]
        fault_ranks.sort(key=self._get_rank_id)
        cm_infos = f'get reset config from file, retry_time={retry_time}, restart_type={restart_type},' \
                   f' grace_exit={grace_exit}, fault_flush={fault_flush}, fault_ranks={fault_ranks}'
        run_log.info(cm_infos)
        return ResetCmData(fault_ranks, retry_time, grace_exit, restart_type, fault_flush)

    def _update_reset_info(self):
        reset_data = self._get_reset_info_from_cm()
        self.fault_ranks = reset_data.fault_ranks
        self.retry_time = reset_data.retry_time
        self.grace_exit = reset_data.grace_exit
        self.restart_type = reset_data.restart_type

    def _get_reset_config(self) -> dict:
        try:
            reset_file_content = json.loads(safe_get_file_info(self.reset_cm_path))
        except Exception as err:
            run_log.warning(f"json load config failed, , because {err}")
            return dict()

        restart_type_content = safe_get_file_info(self.restart_type_path).strip()
        run_log.info(f"got restart_type_content:{restart_type_content}")
        reset_file_content[constants.KEY_RESTART_TYPE] = restart_type_content
        return reset_file_content


fault_processor = FaultProcessor()