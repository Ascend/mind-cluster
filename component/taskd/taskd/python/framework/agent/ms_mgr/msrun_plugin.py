#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import os
import time

from taskd.python.toolkit.fault_checker.fault_check import fault_processor, grace_exit_pids, stop_pids, FaultStatus, \
    force_exit_pids
from taskd.python.toolkit.constants import constants
from taskd.python.toolkit.constants.constants import KILL_ALL_WORKERS, KILL_ALL_WORKER_CALLBACK_NAME, \
    START_ALL_WORKER_CALLBACK_NAME, MONITOR_CALLBACK_NAME, KILL_INTERVAL
from taskd.python.utils.log import run_log
from taskd.python.framework.agent.ms_mgr.ms_utils import check_monitor_res_valid, calculate_global_rank
from taskd.python.toolkit.recover_module import shared_data
from taskd.python.toolkit.recover_module.recover_manager import init_grpc_client, register_callback_func, \
    init_grpc_recover_manager, init_grpc_process
from taskd.python.toolkit.validator.file_process import safe_get_file_info


class MSRunPlugin:
    """
    MSRunPlugin class is for manager process-rescheduling、pod-level-rescheduling and grace tolerance
    it is called by mindspore to register relative callbacks to controller the life-cycle of its msrun processes
    and parse the faults including software and hardware by the file inject by device-plugin reset configmap
    """
    RANK_STATUS_UNHEALTHY = "UNHEALTHY"
    RANK_STATUS_UNKNOWN = "UNKNOWN"
    RANK_STATUS_INIT = "INIT"
    RANK_STATUS_HEALTHY = "HEALTHY"
    RANK_STATUS_STOPPED = "STOPPED"
    RANK_STATUS_SUCCEEDED = "SUCCEEDED"
    Rank_Status_FAILED = "FAILED"
    FRAMEWORK_MS_NAME = "mindspore"

    def __init__(self):
        # This time is the interval time of the infinite loop.
        self.all_rank_succeed = False
        self.monitor_interval = 5
        # Use a string to mark the health status of all global ranks to determine
        # whether it is necessary to kill the processes.
        self.rank_status = ""
        # The PIDs of the local ranks, that is, the process IDs of the training processes on the current node.
        self.rank_pids = []
        # The process information returned by the monitor of the local rank.
        self.rank_info = {}
        # The global rank corresponding to the local rank, which is the value of the global rank.
        self.node_global_rank_ids = []

        # The previously recorded faulty ranks are used to determine whether the ranks have been updated.
        self.pre_fault_ranks = None
        # Record all the faulty ranks from the reset CM, not just the local ranks.
        self.fault_ranks = None
        self.retry_time = 0
        self.pre_retry_time = 0
        self.grace_exit = None
        self.restart_type = None
        self.__func_map = {}
        self.rank_table_version = 0

        self.reset_cm_path = constants.RESET_CONFIG_PATH
        self.restart_type_path = constants.RESTART_TYPE_PATH
        self.rank_version_path = constants.RANK_TABLE_VERSION_PATH

        self.framework = self.FRAMEWORK_MS_NAME
        self.ms_node_rank = os.getenv("MS_NODE_RANK")

    def register_callbacks(self, operator, func):
        self.__func_map[operator] = func

    def start_mindspore_workers(self):
        start_worker_func = self.__func_map[START_ALL_WORKER_CALLBACK_NAME]
        init_time = 0
        while True:
            if init_time >= constants.INIT_TIMEOUT:
                raise ValueError("failed to start workers, initialized timeout")
            run_log.warning(f"self.wait_to_start():{self.wait_to_start()}")
            if self.wait_to_start():
                run_log.info(f"nodeRank:{self.ms_node_rank} will start workers")
                start_worker_func()
                run_log.info("all training processes has been started")
                break
            time.sleep(constants.WAITING_INTERVAL)
            init_time = init_time + constants.WAITING_INTERVAL

    def all_fault_has_recovered(self) -> bool:
        """
        all_fault_has_recovered is a function or operation that checks whether all faults have been resolved.
        """
        for fault_rank in self.fault_ranks:
            if "Status" not in fault_rank.keys():
                run_log.warning(f"can not get status from {fault_rank}, skipping checking reset phrase for this rank")
                continue
            if fault_rank.get("Status") != "recovered":
                run_log.warning(f"{fault_rank} is not recovered yet")
                return False

        if os.path.exists(self.rank_version_path) and self.restart_type == constants.VALUE_RESTART_RESCHEDULE_TYPE:
            file_rank_version = self.read_rank_table_version()
            if file_rank_version <= self.rank_table_version:
                warn_info = f"rank table version is {file_rank_version} while self.rank_version " \
                            f"is {self.rank_table_version}, maybe rank table file in container is " \
                            f"still not updated in path {self.rank_version_path}"
                run_log.warning(warn_info)
                return False
            self.rank_table_version = file_rank_version

        # if all fault ranks are recovered, should restart workers. update recorded retry time and fault ranks
        recovered_infos = f'all fault recovered, updating fault_ranks={self.fault_ranks},' \
                          f' retry_time={self.retry_time}, restart_type={self.restart_type}'
        run_log.info(recovered_infos)
        self.pre_retry_time = self.retry_time
        self.pre_fault_ranks = self.fault_ranks
        return True

    def get_fault_status(self):
        """
        Obtain the current fault status from the updated reset cm.
        """
        fault_local_ranks = []
        fault_status = False
        unrecovered_status = False
        retry_status = False
        local_worker_ranks = self.node_global_rank_ids
        self.update_reset_info()
        # retry time被更新了
        if self.retry_time > self.pre_retry_time:
            retry_status = True
        # fault rank有更新
        if self.pre_fault_ranks != self.fault_ranks:
            for fault_rank in self.fault_ranks:
                if "Status" not in fault_rank.keys():
                    warn_info = f"can not get Status from {fault_rank},skipping checking reset phrase for this rank"
                    run_log.warning(warn_info)
                    continue
                rank_id = fault_rank.get("RankId")
                status = fault_rank.get("Status")
                run_log.info(
                    f"status:{status},rankId:{rank_id},local:{local_worker_ranks}, {rank_id in local_worker_ranks}")
                if status == "fault" and rank_id in local_worker_ranks:
                    fault_local_ranks.append(rank_id)
                    fault_status = True
                if status == "unrecovered" or status == "recovered":
                    unrecovered_status = True
        return FaultStatus(fault_local_ranks, fault_status, unrecovered_status, retry_status)

    def read_rank_table_version(self) -> int:
        version = safe_get_file_info(self.rank_version_path).strip()
        if not version.isdigit():
            return -1
        return int(version)

    # start() should be called by mindspore msrun,to take over the control of training processes
    def start(self):
        kill_worker_func = self.__func_map[KILL_ALL_WORKER_CALLBACK_NAME]
        start_worker_func = self.__func_map[START_ALL_WORKER_CALLBACK_NAME]
        # {rank_0: {pid: pidNum, status: 状态码}，1：状态码 …..}
        monitor_func = self.__func_map[MONITOR_CALLBACK_NAME]
        if kill_worker_func is None or start_worker_func is None or monitor_func is None:
            raise Exception(f"{self.FRAMEWORK_MS_NAME} hasn't fully registered all callbacks")

        self._init_grpc_client_if_needed()
        # First, start the MindSpore training.
        self.start_mindspore_workers()
        while True:
            if self.ms_node_rank == "0" and shared_data.shared_data_inst.get_kill_flag():
                run_log.info("master agent receive killMaster signal")
                kill_worker_func([KILL_ALL_WORKERS])
                exit(1)

            time.sleep(self.monitor_interval)
            # After entering the loop, first obtain the process status once.
            ms_proc_status = monitor_func([-1])
            run_log.info(f"nodeRank:{self.ms_node_rank} has got mindspore process status:{ms_proc_status}")
            if not check_monitor_res_valid(ms_proc_status):
                run_log.warning(f"monitor not return a valid result, but {ms_proc_status}")
                continue
            # Update the local process IDs and global rank numbers
            # based on the information returned by the monitor interface.
            self.update_rank_status(ms_proc_status)

            # 进入循环后更新reset cm相关内容
            self.update_reset_info()
            fault_status = self.get_fault_status()
            run_log.info(f"nodeRank:{self.ms_node_rank}  fault status: is_fault:{fault_status.is_fault},"
                         f"is_unrecovered:{fault_status.is_unrecovered},is_retried:{fault_status.is_retried},"
                         f"local_ranks:{fault_status.local_ranks}")

            # If the "reset cm" indicates that the training needs to be exited, use it in the sub-healthy state.
            if self._handle_grace_exit():
                continue
            # There are business-side faults with fault_rank, covering software faults.
            # Faulty pods are controlled to terminate and exit on their own.
            # The exit of pods with non-faulty ranks is covered by the following two scenarios.
            # process fault clusterd will write rank， if fault_status.is_retried and not fault_status.is_unrecovered
            # retry be wrotten by clusterd，status as fault,
            self._handle_fault_status(fault_status)
            # When there is no fault_rank but the retry_time has increased,
            # it covers the single-Pod rescheduling scenario [business-side fault].
            # The process-level recovery scenario is not enabled. After a fault is detected,
            # [the operation is carried out] by Volcano.
            # retrytime+1
            if self._handle_process_retry_fault(fault_status):
                continue
            # There is an unrecoverable scenario with fault_rank, covering hardware failures.
            if self._handle_hardware_fault(fault_status):
                continue
            # to exit while all training process has exit with succeed code
            self._handle_all_process_succeed()
            # If the result of the process monitoring is abnormal,
            # stop the training and exit to make the pod go into an error state.
            # Let the pod rescheduling mechanism update the retry_time, and then restart the training on other nodes.
            self._handle_exist_unhealthy_process()

    def update_rank_status(self, rank_status_dict: dict):
        """
        update_rank_status updates the single status value of all current ranks based on
        the return value of the monitor. If a rank has an error, its status is set to unhealthy.
        At the same time, update the PIDs corresponding to all ranks and all global rank
        numbers corresponding to the current node.
        data = {
            {0: {'pid': 101, 'status': None, 'global_rank': 16}, 1: {'pid': 110, 'status': None, 'global_rank': 17},
            2: {'pid': 119, 'status': None, 'global_rank': 18}, 3: {' 129, 'status': None, 'global_rank': 19},
            4: {'pid': 143, 'status': None, 'global_rank': 20}, 5: {'pid': 155, 'status': None, 'global_rank': 21},
            6: {'pid': 167, 'status': None, 'global_rank': 22}, 7: {'pid': 176, 'status': None, 'global_rank': 23}}
        }
        """
        self.rank_info = rank_status_dict
        all_healthy = True
        all_succeed = True
        rank_pids = []
        local_rank_ids = []
        for _, details in rank_status_dict.items():
            # if process is in ok, not start yet[msrun taken over by taskd, monitor maybe called before training],
            # sleeping[during process recover]
            if details[constants.RANK_STATUS_KEY] not in {constants.RANK_STATUS_OK, constants.RANK_STATUS_NOT_START,
                                                          constants.RANK_STATUS_COMPLETE}:
                self.rank_status = self.RANK_STATUS_UNHEALTHY
                all_healthy = False
            if details[constants.RANK_STATUS_KEY] not in {constants.RANK_STATUS_COMPLETE}:
                all_succeed = False
            rank_pids.append(details[constants.RANK_PID_KEY])
            local_rank_ids.append(details[constants.GLOBAL_RANK_ID_KEY])
        self.rank_pids = rank_pids
        self.node_global_rank_ids = local_rank_ids
        self.all_rank_succeed = all_succeed
        if all_healthy:
            self.rank_status = self.RANK_STATUS_HEALTHY

    # Read the content of resetcm and update the relevant content.
    def update_reset_info(self):
        reset_data = fault_processor.get_reset_info_from_cm()
        self.fault_ranks = reset_data.fault_ranks
        self.retry_time = reset_data.retry_time
        self.grace_exit = reset_data.grace_exit
        self.restart_type = reset_data.restart_type

    def update_pre_fault_infos(self):
        self.pre_retry_time = self.retry_time
        self.pre_fault_ranks = []

    def wait_to_start(self) -> bool:
        reset_data = fault_processor.get_reset_info_from_cm()
        # 通过环境变量计算 global ranks
        self.node_global_rank_ids = calculate_global_rank()
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
            if rank_id in self.node_global_rank_ids and status == constants.VALUE_FAULT:
                return False
        return True

    def _init_grpc_client_if_needed(self):
        if self.ms_node_rank == "0":
            run_log.info("rank 0 will start controller grpc client")
            init_grpc_client(self.framework)

    def _handle_grace_exit(self):
        if self.grace_exit != 1:
            return False
        try:
            grace_exit_pids(self.rank_pids)
        except Exception as e:
            run_log.info(f"failed to gracefully kill worker process, {e}")
        finally:
            self.__func_map[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
            stop_pids(self.rank_pids)
        return True

    def _handle_fault_status(self, fault_status):
        if not fault_status.is_fault:
            return
        run_log.warning(f"nodeRank:{self.ms_node_rank}  entering fault_status.is_fault")
        self.__func_map[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
        force_exit_pids(self.rank_pids)
        run_log.warning(f"local rank got fault, will stop worker{self.node_global_rank_ids}")
        exit(1)

    def _handle_process_retry_fault(self, fault_status):
        if fault_status.is_retried and not fault_status.is_unrecovered:
            run_log.warning(
                f"nodeRank:{self.ms_node_rank} entering fault_status.is_retried and not "
                f"fault_status.is_unrecovered")
            # In this scenario, there is no content in the fault_rank.
            # restart the training after the rank table is updated.
            if not self.all_fault_has_recovered():
                return True
            self.__func_map[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
            if self.ms_node_rank == "0":
                run_log.warning("will kill mindio controller")
                shared_data.shared_data_inst.set_exit_flag(True)
            run_log.warning(f"nodeRank:{self.ms_node_rank}  will sleep for 10 secs, after kill workers")
            time.sleep(KILL_INTERVAL)
            run_log.warning("sleep over, will start")
            self.start_mindspore_workers()
            self.update_pre_fault_infos()
            return True
        return False

    def _handle_hardware_fault(self, fault_status):
        if not fault_status.is_unrecovered:
            return False
        run_log.warning(f"nodeRank:{self.ms_node_rank} entering fault_status.is_unrecovered")
        self.__func_map[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
        if self.all_fault_has_recovered():
            self.__func_map[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
            self.start_mindspore_workers()
        return True

    def _handle_all_process_succeed(self):
        if not self.all_rank_succeed:
            return
        run_log.info(
            f"nodeRank:{self.ms_node_rank} successfully finished."
        )
        shared_data.shared_data_inst.set_kill_flag(True)
        time.sleep(constants.WAITING_INTERVAL * constants.WAIT_TIMES)
        exit(0)

    def _handle_exist_unhealthy_process(self):
        if self.rank_status in {self.RANK_STATUS_UNHEALTHY}:
            run_log.warning(f"nodeRank:{self.ms_node_rank} some rank is unhealthy will stop workers, "
                            f"and exit this node")
            if self.ms_node_rank == "0":
                run_log.warning("will kill mindio controller")
                shared_data.shared_data_inst.set_kill_flag(True)
                time.sleep(constants.WAITING_INTERVAL * constants.WAIT_TIMES)
            stop_res = self.__func_map[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
            run_log.warning(f"rank with pid {self.rank_pids} will be killed")
            if stop_res is not constants.RES_OK:
                run_log.error(
                    f"nodeRank:{self.ms_node_rank} failed to stop workers with return code:{stop_res}")
            exit(1)
