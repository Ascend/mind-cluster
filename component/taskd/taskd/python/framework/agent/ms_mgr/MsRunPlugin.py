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
import subprocess
import time
from taskd.python.toolkit.fault_checker.fault_check import fault_processor, grace_exit_pids, stop_pids, FaultStatus, \
    force_exit_pids
from taskd.python.toolkit.constants import constants
from taskd.python.toolkit.constants.constants import KILL_ALL_WORKERS, KILL_ALL_WORKER_CALLBACK_NAME, \
    START_ALL_WORKER_CALLBACK_NAME, MONITOR_CALLBACK_NAME, KILL_INTERVAL
from taskd.python.toolkit.logger.log import run_log
from taskd.python.framework.agent.ms_mgr.MsUtils import check_monitor_res_valid, calculate_global_rank
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
    Rank_Status_UNHEALTHY = "UNHEALTHY"
    Rank_Status_UNKNOWN = "UNKNOWN"
    Rank_Status_INIT = "INIT"
    Rank_Status_HEALTHY = "HEALTHY"
    Rank_Status_STOPPED = "STOPPED"
    Rank_Status_SUCCEEDED = "SUCCEEDED"
    Rank_Status_FAILED = "FAILED"
    Framework_MS_Name = "mindspore"

    def __init__(self):
        # 该时间为死循环的间隔时间
        self.all_rank_succeed = False
        self.monitor_interval: float = 5
        # 用一个str标志所有global rank的健康状态，来确定是否需要杀进程
        self.rank_status = ""
        # local rank们的pid，即当前节点上的训练进程的进程号
        self.rank_pids = []
        # local rank的monitor返回的进程信息
        self.rank_info = {}
        # 本地的rank所对应的global rank，为global rank的数值
        self.node_global_rank_ids = []

        # 上一次记录的fault rank用于判定rank是否更新
        self.pre_fault_ranks = None
        # 记录来自于reset cm的所有的fault rank，不仅仅是local ranks
        self.fault_ranks = None
        self.retry_time = 0
        self.pre_retry_time = 0
        self.grace_exit = None
        self.restart_type = None
        self.__funcMap = {}
        self.rank_table_version = 0

        self.reset_cm_path = constants.RESET_CONFIG_PATH
        self.restart_type_path = constants.RESTART_TYPE_PATH
        self.rank_version_path = constants.RANK_TABLE_VERSION_PATH

        self.framework = self.Framework_MS_Name

    def register_callbacks(self, operator, func):
        self.__funcMap[operator] = func

    def start_mindspore_workers(self):
        start_worker_func = self.__funcMap[START_ALL_WORKER_CALLBACK_NAME]
        init_time = 0
        while True:
            if init_time >= constants.INIT_TIMEOUT:
                raise ValueError("failed to start workers, initialized timeout")
            run_log.warning(f"self.wait_to_start():{self.wait_to_start()}")
            if self.wait_to_start():
                run_log.info(f"nodeRank:{os.getenv('MS_NODE_RANK')} will start workers")
                start_worker_func()
                run_log.info("all training processes has been started")
                break
            time.sleep(constants.WAITING_INTERVAL)
            init_time = init_time + constants.WAITING_INTERVAL

    def _init_grpc_client_if_needed(self):
        if os.getenv("MS_NODE_RANK") == "0":
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
            self.__funcMap[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
            stop_pids(self.rank_pids)
        return True

    def _handle_fault_status(self, fault_status):
        if not fault_status.is_fault:
            return
        run_log.warning(f"nodeRank:{os.getenv('MS_NODE_RANK')}  entering fault_status.is_fault")
        self.__funcMap[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
        force_exit_pids(self.rank_pids)
        run_log.warning(f"local rank got fault, will stop worker{self.node_global_rank_ids}")
        exit(1)

    def _handle_process_retry_fault(self, fault_status):
        if fault_status.is_retried and not fault_status.is_unrecovered:
            run_log.warning(
                f"nodeRank:{os.getenv('MS_NODE_RANK')} entering fault_status.is_retried and not "
                f"fault_status.is_unrecovered")
            # 该场景下fault_rank没有内容，等待ranktable更新后即可重启训练
            if not self.all_fault_has_recovered():
                return True
            self.__funcMap[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
            run_log.warning(f"nodeRank:{os.getenv('MS_NODE_RANK')}  will sleep for 10 secs, after kill workers")
            time.sleep(KILL_INTERVAL)
            run_log.warning("sleep over, will start")
            if os.getenv("MS_NODE_RANK") == "0":
                run_log.warning("will kill mindio controller")
                shared_data.shared_data_inst.set_exit_flag(True)
            self.start_mindspore_workers()
            self.update_pre_fault_infos()
            return True
        return False

    def _handle_hardware_fault(self, fault_status):
        if not fault_status.is_unrecovered:
            return False
        run_log.warning(f"nodeRank:{os.getenv('MS_NODE_RANK')} entering fault_status.is_unrecovered")
        self.__funcMap[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
        if self.all_fault_has_recovered():
            self.__funcMap[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
            self.start_mindspore_workers()
        return True

    def _handle_all_process_succeed(self):
        if not self.all_rank_succeed:
            return
        run_log.info(
            f"nodeRank:{os.getenv('MS_NODE_RANK')} successfully finished."
        )
        shared_data.shared_data_inst.set_kill_flag(True)
        time.sleep(constants.WAITING_INTERVAL * constants.WAIT_TIMES)
        exit(0)

    def _handle_exist_unhealthy_process(self):
        if self.rank_status in {self.Rank_Status_UNHEALTHY}:
            run_log.warning(f"nodeRank:{os.getenv('MS_NODE_RANK')} some rank is unhealthy will stop workers, "
                            f"and exit this node")
            if os.getenv("MS_NODE_RANK") == "0":
                run_log.warning("will kill mindio controller")
                shared_data.shared_data_inst.set_kill_flag(True)
                time.sleep(constants.WAITING_INTERVAL * constants.WAIT_TIMES)
            stop_res = self.__funcMap[KILL_ALL_WORKER_CALLBACK_NAME]([KILL_ALL_WORKERS])
            run_log.warning(f"rank with pid {self.rank_pids} will be killed")
            if stop_res is not constants.RES_OK:
                run_log.error(
                    f"nodeRank:{os.getenv('MS_NODE_RANK')} failed to stop workers with return code:{stop_res}")
            exit(1)

    # start() should be called by mindspore msrun,to take over the control of training processes
    def start(self):
        kill_worker_func = self.__funcMap[KILL_ALL_WORKER_CALLBACK_NAME]
        start_worker_func = self.__funcMap[START_ALL_WORKER_CALLBACK_NAME]
        # {rank_0: {pid: pidNum, status: 状态码}，1：状态码 …..}
        monitor_func = self.__funcMap[MONITOR_CALLBACK_NAME]
        if kill_worker_func is None or start_worker_func is None or monitor_func is None:
            raise Exception(f"{self.Framework_MS_Name} hasn't fully registered all callbacks")

        self._init_grpc_client_if_needed()
        # 首先将mindspore的训练拉起来
        self.start_mindspore_workers()
        while True:
            if os.getenv("MS_NODE_RANK") == "0" and shared_data.shared_data_inst.get_kill_flag():
                run_log.info("master agent receive killMaster signal")
                kill_worker_func([KILL_ALL_WORKERS])
                exit(1)

            time.sleep(self.monitor_interval)
            # 进入循环后首先获取一次进程状态
            ms_proc_status = monitor_func([-1])
            run_log.info(f"nodeRank:{os.getenv('MS_NODE_RANK')} has got mindspore process status:{ms_proc_status}")
            if not check_monitor_res_valid(ms_proc_status):
                run_log.warning(f"monitor not return a valid result, but {ms_proc_status}")
                continue
            # 根据monitor接口返回信息更新本地进程号，全局rank号
            self.update_rank_status(ms_proc_status)

            # 进入循环后更新reset cm相关内容
            self.update_reset_info()
            fault_status = self.get_fault_status()
            run_log.info(f"nodeRank:{os.getenv('MS_NODE_RANK')}  fault status: is_fault:{fault_status.is_fault},"
                         f"is_unrecovered:{fault_status.is_unrecovered},is_retried:{fault_status.is_retried},"
                         f"local_ranks:{fault_status.local_ranks}")

            # reset cm中写了需要退出训练的话,亚健康使用
            if self._handle_grace_exit():
                continue
            # 有fault_rank的业务面故障   覆盖软件故障, 控制有故障的pod自己死亡退出，无故障rank pod的退出由以下两种场景覆盖
            # 软件故障场景下由clusterd会写入rank， if fault_status.is_retried and not fault_status.is_unrecovered
            # retry 由clusterd写入，status为fault,
            self._handle_fault_status(fault_status)
            # 没有fault_rank 但是retrytime增加了的 覆盖单pod重调度场景[业务面故障]，不开启进程级恢复场景，感知到故障后由volcano将
            # retrytime+1
            if self._handle_process_retry_fault(fault_status):
                continue
            # 有fault_rank的unrecover场景，覆盖硬件故障
            self._handle_hardware_fault(fault_status)
            # to exit while all training process has exit with succeed code
            self._handle_all_process_succeed()
            # 如果进程监控结果异常那么停止训练,并且退出使得pod error掉，由pod重调度写retrytime 其他节点重拉
            self._handle_exist_unhealthy_process()

    #  update_rank_status 根据monitor的返回值更新当前的所有rank的单一状态值，有err的rank状态就置为unhealthy
    #  同时更新所有rank对应的pid, 和本节点对应的所有global rank号
    def update_rank_status(self, rank_status_dict: dict):
        """
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
        for rank, details in rank_status_dict.items():
            # if process is in ok, not start yet[msrun taken over by taskd, monitor maybe called before training],
            # sleeping[during process recover]
            if details[constants.RANK_STATUS_KEY] not in {constants.rank_status_ok, constants.rank_status_not_start}:
                self.rank_status = self.Rank_Status_UNHEALTHY
                all_healthy = False
            if details[constants.RANK_STATUS_KEY] not in {constants.rank_status_complete}:
                all_succeed = False
            rank_pids.append(details[constants.RANK_PID_KEY])
            local_rank_ids.append(details[constants.GLOBAL_RANK_ID_KEY])
        self.rank_pids = rank_pids
        self.node_global_rank_ids = local_rank_ids
        self.all_rank_succeed = all_succeed
        if all_healthy:
            self.rank_status = self.Rank_Status_HEALTHY

    # 读取resetcm内容，并将相关内容进行更新
    def update_reset_info(self):
        reset_data = fault_processor._get_reset_info_from_cm()
        self.fault_ranks = reset_data.fault_ranks
        self.retry_time = reset_data.retry_time
        self.grace_exit = reset_data.grace_exit
        self.restart_type = reset_data.restart_type

    # 从更新后的reset cm内获取当前的fault状态
    def get_fault_status(self):
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

    # all_fault_has_recovered 判断是否所有故障已经恢复
    def all_fault_has_recovered(self) -> bool:
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

    def read_rank_table_version(self) -> int:
        version = safe_get_file_info(self.rank_version_path).strip()
        if not version.isdigit():
            return -1
        return int(version)

    def update_pre_fault_infos(self):
        self.pre_retry_time = self.retry_time
        self.pre_fault_ranks = []

    def wait_to_start(self) -> bool:
        reset_data = fault_processor._get_reset_info_from_cm()
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
