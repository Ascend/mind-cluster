#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
"""Exception Handler related classes and functions."""
import os
import gc
import re
from enum import Enum
import signal
import threading
import time
import atexit
import hashlib
from functools import wraps
from typing import Callable
import ttp_c2python_api
from ..controller_ttp import ttp_logger
from ..controller_ttp.ttp_controller import tft_destroy_controller
from ..controller_ttp.ttp_utils import is_zero_ip, input_ip_transform, get_env_var_int_safely
from ..utils import tft_can_do_uce_repair, get_l2_hbm_error_time
from ..utils.uce_utils import get_update_start_end_time

mindx_agent = os.getenv('MINDX_TASK_ID')
mind_spore = os.getenv('MINDIO_FOR_MINDSPORE', 'False')
mind_spore = mind_spore.lower() in ('true', '1')
if not mind_spore:
    import torch
    import torch_npu
    import torch.distributed
    from torch.distributed.elastic.multiprocessing import MultiprocessContext, SubprocessContext
    from torch.distributed.elastic.multiprocessing.subprocess_handler.subprocess_handler import SubprocessHandler
    from torch.distributed.rendezvous import rendezvous, _rendezvous_helper
    import torch.distributed.distributed_c10d as dist_c10d
else:
    import mindspore as ms

ACTION_TIMEOUT = get_env_var_int_safely("TTP_NORMAL_ACTION_TIME_LIMIT", 180, 30, 1800)
EXIT_WAIT_TIME = 10  # wait to exit time
RET_OK = 0
RET_ERROR = 1
RET_K8S = 2  # return not 0 value for k8s
RET_PASS = 3
RET_EXCEPTION = 4
TTP_MAX_WORLD_SIZE = 100000

rank_ = None
device_ = None
repair_id_ = None
sync_stream_ = None
mindio_export_function_version = None
force_stop_cond_ = threading.Condition()
uce_error_ = False
hccl_error_ = False
pause_step_ = 0
need_pause_ = None
need_pause_cond_ = threading.Condition()


class PauseType(Enum):
    """Pause train type."""
    PAUSE = 1
    RAISE = 2


class OptimizerType(Enum):
    """Optimizer type."""
    ATTENTION = 0
    MOE = 1


class Action(Enum):
    """wait next action return type"""
    RETRY = 0
    EXIT = 1


fault_dict = {'FORCE STOP': 'RS_NORMAL',
              'UCE ERROR': 'RS_UCE',
              'HBM MULTI BIT ECC': 'RS_UCE',
              'ARF FINISH': 'RS_PREREPAIR_FINISH',
              'STEP FINISH': 'RS_STEP_FINISH',
              'HCCL OP RETRY FAILED': 'RS_HCCL_FAILED',
              'SUSPECT REMOTE ERROR': 'RS_HCCL_FAILED'}


class ReportState(Enum):
    RS_NORMAL = ttp_c2python_api.ReportState_RS_NORMAL
    RS_UCE = ttp_c2python_api.ReportState_RS_UCE
    RS_UCE_CORRUPTED = ttp_c2python_api.ReportState_RS_UCE_CORRUPTED
    RS_HCCL_FAILED = ttp_c2python_api.ReportState_RS_HCCL_FAILED
    RS_UNKNOWN = ttp_c2python_api.ReportState_RS_UNKNOWN
    RS_INIT_FINISH = ttp_c2python_api.ReportState_RS_INIT_FINISH
    RS_PREREPAIR_FINISH = ttp_c2python_api.ReportState_RS_PREREPAIR_FINISH
    RS_STEP_FINISH = ttp_c2python_api.ReportState_RS_STEP_FINISH


class RepairType(Enum):
    RT_SEND = ttp_c2python_api.RepairType_RT_SEND
    RT_UCE_HIGHLEVEL = ttp_c2python_api.RepairType_RT_UCE_HIGHLEVEL
    RT_UCE_LOWLEVEL = ttp_c2python_api.RepairType_RT_UCE_LOWLEVEL
    RT_ROLLBACK = ttp_c2python_api.RepairType_RT_ROLLBACK
    RT_RECV_REPAIR = ttp_c2python_api.RepairType_RT_RECV_REPAIR
    RT_LOAD_CKPT = ttp_c2python_api.RepairType_RT_LOAD_CKPT
    RT_LOAD_REBUILD = ttp_c2python_api.RepairType_RT_LOAD_REBUILD


def get_device():
    return device_


def set_device():
    if mind_spore:
        return
    if device_ is None:
        ttp_logger.LOGGER.error("rank:%s device is null, nothing todo", rank_)
        raise ValueError("Npu device context is None")
    torch.npu.set_device(device_)


class SaveHandler:
    def __init__(self):
        """
        Init parameters for save handler class
        """

        self._args = None
        self._fm_save_ckpt_call = None
        self._save_ckpt_ctx = None
        self._fm_rename_call = None
        self._rename_ctx = None
        self._fm_exit_call = None
        self._exit_ctx = None
        self._fm_stop_call = None
        self._stop_ctx = None
        self._fm_clean_call = None
        self._clean_ctx = None
        self._fm_rebuild_group_call = None
        self._fm_sync_stream_call = None
        self._rebuild_group_ctx = None
        self._sync_stream_ctx = None
        self._fm_repair_call = None
        self._repair_ctx = None
        self._fm_rollback_call = None
        self._rollback_ctx = None
        self._zit_param = None

        self._zit_upgrade_rollback_ctx = None
        self._zit_upgrade_repair_ctx = None
        self._zit_upgrade_rebuild_ctx = None
        self._zit_downgrade_rebuild_ctx = None

        self._repair_step = 0
        self._need_rebuild = False
        self._error_ranks = None
        self._repair_info = None
        self._started_dump = False
        self._need_dump = False
        self._dump_info = None
        self._dump_step = 0
        self._exit_cond = threading.Condition()
        self._enable_uce = False

        self._fm_zit_upgrade_rollback_call = None
        self._fm_zit_upgrade_repair_call = None
        self._fm_zit_upgrade_rebuild_call = None
        self._fm_zit_downgrade_rebuild_call = None

    def get_repair_step(self):
        ttp_logger.LOGGER.info(f"tft return repair_step: {self._repair_step}")
        return self._repair_step

    def set_repair_step(self, repair_step):
        self._repair_step = repair_step

    def set_model_config(self, args):
        """
        Set Config set parameters for exception save
        """
        if self._started_dump:
            raise Exception("already start dump, can not set model param")
        self._args = args

    def set_dump_config(self, step: int, save_info: list):
        """
        Set dump parameter
        """
        if self._need_dump:
            ttp_logger.LOGGER.info("already set dump config, but receive new dump config!")
            return RET_ERROR

        self._dump_step = step
        self._dump_info = save_info
        self._need_dump = True
        return RET_OK

    def set_repair_rollback_config(self, args: ttp_c2python_api.RepairContext):

        repair_type, step, optim_idxs = args.type, args.step, args.group_idx
        src_ranks, dest_ranks, rank_list, zit_param = args.src_rank, args.dst_rank, args.rank_list, args.zit_param
        save_handler.set_repair_step(step)
        if repair_type in [RepairType.RT_UCE_HIGHLEVEL.value, RepairType.RT_LOAD_REBUILD.value]:
            need_rebuild = True
        else:
            need_rebuild = False
        error_ranks = dest_ranks

        # dict
        repair_info = {
            "type": optim_idxs,
            "repair_type": repair_type,
            "src": src_ranks,
            "dst": dest_ranks,
            "rank_list": rank_list
        }

        self._repair_step = step
        self._need_rebuild = need_rebuild
        self._error_ranks = error_ranks
        self._repair_info = repair_info
        self._zit_param = zit_param

    def wait_exit_cond(self):
        """
        Exception Save will save lastwords checkpoint when catch exception
        """
        with self._exit_cond:
            ttp_logger.LOGGER.info(f"rank:{rank_} reported exception state, exit wait...")
            self._exit_cond.wait(ACTION_TIMEOUT)
            ttp_logger.LOGGER.info(f"rank:{rank_} finished to wait dump, exit...")

    def register(self):
        """
        Register exception handler function
        """
        signal.signal(signal.SIGTERM, self.do_nothing)
        signal.signal(signal.SIGINT, self.do_nothing)

    def set_uce(self, enable_uce):
        self._enable_uce = enable_uce

    def get_uce(self):
        return self._enable_uce

    def do_nothing(self, signum=2, frame=None):
        ttp_logger.LOGGER.info("rank:%s catch exception signal, do nothing", rank_)

    def register_rename_handler(self, func, ctx):
        self._fm_rename_call = func
        self._rename_ctx = ctx

    def register_save_ckpt_handler(self, func, ctx):
        self._fm_save_ckpt_call = func
        self._save_ckpt_ctx = ctx

    def register_exit_handler(self, func, ctx):
        self._fm_exit_call = func
        self._exit_ctx = ctx

    def register_stop_handler(self, func, ctx):
        self._fm_stop_call = func
        self._stop_ctx = ctx

    def register_clean_handler(self, func, ctx):
        self._fm_clean_call = func
        self._clean_ctx = ctx

    def register_repair_handler(self, func, ctx):
        self._fm_repair_call = func
        self._repair_ctx = ctx

    def register_rollback_handler(self, func, ctx):
        self._fm_rollback_call = func
        self._rollback_ctx = ctx

    def register_rebuild_group_handler(self, func, ctx):
        self._fm_rebuild_group_call = func
        self._rebuild_group_ctx = ctx

    def register_sync_handler(self, func, ctx):
        self._fm_sync_stream_call = func
        self._sync_stream_ctx = ctx

    def register_zit_upgrade_rebuild_handler(self, func, ctx):
        self._fm_zit_upgrade_rebuild_call = func
        self._zit_upgrade_rebuild_ctx = ctx

    def register_zit_upgrade_rollback_handler(self, func, ctx):
        self._fm_zit_upgrade_rollback_call = func
        self._zit_upgrade_rollback_ctx = ctx

    def register_zit_upgrade_repair_handler(self, func, ctx):
        self._fm_zit_upgrade_repair_call = func
        self._zit_upgrade_repair_ctx = ctx

    def register_zit_downgrade_rebuild_handler(self, func, ctx):
        self._fm_zit_downgrade_rebuild_call = func
        self._zit_downgrade_rebuild_ctx = ctx

    def rename_ckpt_dir(self):
        if self._fm_rename_call is not None:
            try:
                if mind_spore:
                    self._fm_rename_call(self._dump_step, self._rename_ctx)
                else:
                    self._fm_rename_call(self._dump_step, self._args)
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
                return RET_ERROR
        return RET_OK

    def notify_exit(self):
        with self._exit_cond:
            self._exit_cond.notify()

    def execute_exit(self):
        if self._fm_exit_call is not None:
            try:
                self._fm_exit_call(self._exit_ctx)
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")

    def execute_stop(self):
        try:
            set_device()
            self._fm_stop_call(self._args, self._stop_ctx)
        except Exception as e:
            ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
            return RET_ERROR
        return RET_OK

    def execute_clean(self):
        if self._fm_clean_call is None:
            return RET_OK
        try:
            ret = self._fm_clean_call(uce_error_, self._args, self._clean_ctx)
        except Exception as e:
            ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
            return RET_EXCEPTION
        return ret

    def execute_repair(self):
        if self._fm_repair_call is not None:
            try:
                set_device()
                self._fm_repair_call(self._repair_step, self._need_rebuild, self._error_ranks, self._repair_info,
                                     self._args, self._repair_ctx)
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
                return RET_ERROR
        return RET_OK

    def execute_rollback(self):
        if self._fm_rollback_call is not None:
            try:
                set_device()
                self._fm_rollback_call(self._repair_step, self._args, self._rollback_ctx)
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
                return RET_ERROR
        return RET_OK

    def execute_zit_downgrade_rebuild(self, comm_groups, zit_param):
        if self._fm_zit_downgrade_rebuild_call is not None:
            try:
                self._fm_zit_downgrade_rebuild_call(comm_groups[1], comm_groups[0], self._args, zit_param)
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
                return RET_ERROR
        return RET_OK

    def execute_zit_upgrade_rebuild(self, rank_list, zit_param):
        if self._fm_zit_upgrade_rebuild_call is not None:
            try:
                self._fm_zit_upgrade_rebuild_call(rank_list, self._args, zit_param)
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
                return RET_ERROR
        return RET_OK

    def execute_zit_upgrade_rollback(self):
        if self._fm_zit_upgrade_rollback_call is not None:
            try:
                self._fm_zit_upgrade_rollback_call(self._repair_step, self._args, self._zit_param)
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
                return RET_ERROR
        return RET_OK

    def execute_zit_upgrade_repair(self):
        if self._fm_zit_upgrade_repair_call is not None:
            try:
                self._fm_zit_upgrade_repair_call(self._repair_step, self._need_rebuild, self._error_ranks,
                                                 self._repair_info,
                                                 self._args, self._zit_param)
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
                return RET_ERROR
        return RET_OK

    def execute_save(self):
        if not self._need_dump:
            ttp_logger.LOGGER.info("rank:%s don`t have dump config, nothing todo.", rank_)
            return
        if self._started_dump:
            ttp_logger.LOGGER.info("rank:%s already started dump work flow, no need execute again.", rank_)
            return
        ttp_logger.LOGGER.info("rank:%s begin to execute save checkpoint", rank_)
        self._started_dump = True
        try:
            set_device()
            if self._fm_save_ckpt_call is None:
                ttp_logger.LOGGER.error(f"rank:{rank_} failed to dump checkpoint due to callback is null")
                raise Exception(f"rank:{rank_} failed to dump checkpoint due to callback is null")
            if mind_spore:
                self._fm_save_ckpt_call(self._dump_step, self._dump_info, self._args, self._save_ckpt_ctx)
            else:
                global sync_stream_
                if sync_stream_ is None:
                    sync_stream_ = get_stream()
                with torch.npu.stream(sync_stream_):
                    self._fm_save_ckpt_call(self._dump_step, self._dump_info, self._args, self._save_ckpt_ctx)
        except Exception as e:
            ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
            ttp_c2python_api.set_dump_status(RET_ERROR)
            return
        ttp_c2python_api.set_dump_status(RET_OK)

    def execute_rebuild_group(self, fault_ranks):
        if self._fm_rebuild_group_call is not None:
            try:
                set_device()
                self._fm_rebuild_group_call(fault_ranks, self._args, self._rebuild_group_ctx)
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
                return RET_ERROR
        return RET_OK

    def execute_sync_stream(self):
        if self._fm_sync_stream_call is not None:
            try:
                set_device()
                self._fm_sync_stream_call()
            except Exception as e:
                ttp_logger.LOGGER.exception(f"An error occurred: {str(e)}")
                return RET_ERROR
        return RET_OK


save_handler: SaveHandler = SaveHandler()


def change_timeout(timeout):
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            kwargs['timeout'] = timeout
            return func(*args, **kwargs)

        return wrapper

    return decorator


def subprocess_handler_close(self, death_sig=None) -> None:
    if not death_sig:
        death_sig = _get_default_signal()
    self.proc.send_signal(death_sig)
    if self._stdout:
        self._stdout.close()
    if self._stderr:
        self._stderr.close()


if not mind_spore:
    MultiprocessContext._close = change_timeout(timeout=ACTION_TIMEOUT)(MultiprocessContext._close)
    SubprocessContext._close = change_timeout(timeout=ACTION_TIMEOUT)(SubprocessContext._close)
    SubprocessHandler.close = subprocess_handler_close


def save_checkpoint_callback(step: int, repairId: int, optimidxs: list, ranks: list):
    """
    Save a model checkpoint.
    callback function use to register to processor
    """
    global repair_id_
    repair_id_ = repairId
    ttp_logger.LOGGER.info("rank:%s set dump config step:%s, repairId:%s, optimidxs:%s, ranks:%s",
                           rank_, step, repairId, optimidxs, ranks)
    save_info = []
    for list_idx, rank_list in enumerate(ranks):
        save_info.append({
            "type": optimidxs[list_idx],
            "ranks": rank_list
        })
    save_handler.set_dump_config(step, save_info)
    save_thread = threading.Thread(target=save_data_thread)
    save_thread.start()
    return RET_OK


def save_data_thread():
    """
    Save ckpt in new thread
    """
    save_handler.execute_save()


def get_stream():
    if hasattr(torch.npu, "SyncLaunchStream"):
        stream_ = torch.npu.SyncLaunchStream()
    else:
        stream_ = torch.npu.Stream()
        ttp_logger.LOGGER.warning("update torch_npu to use SyncStream...")
    return stream_


def rename_callback():
    ttp_logger.LOGGER.info("rank:%s begin to rename callback", rank_)
    return save_handler.rename_ckpt_dir()


def exit_callback():
    ttp_logger.LOGGER.debug(f"rank:{rank_} start to exit callback")
    save_handler.notify_exit()
    if not mind_spore:  # to remove for different framework
        exit_thread = threading.Thread(target=exit_proc_thread)
        exit_thread.start()
    else:
        save_handler.execute_exit()


def exit_proc_thread():
    """
    Exit proc thread
    """
    ttp_logger.LOGGER.debug("sleep 10 seconds, then exit process")
    time.sleep(EXIT_WAIT_TIME)
    os._exit(RET_K8S)


def launch_tcp_store_client(url: str, save_rank: int = -1, world_size: int = -1):
    # in case pytorch agent is broken and loose tcp store
    ttp_logger.LOGGER.info(f"ttp_launch_tcp_store_client, url:{url}, save_rank:{save_rank}, world_size:{world_size}")
    os.environ["TORCHELASTIC_USE_AGENT_STORE"] = "True"
    try:
        rendezvous_iterator = rendezvous(url, save_rank, world_size)
    except Exception:
        ttp_logger.LOGGER.error(f"launch tcp store client failed. rank:{save_rank}, world_size:{world_size}, url:{url}")
        return RET_ERROR
    store, _, _ = next(rendezvous_iterator)
    default_pg = dist_c10d._get_default_group()
    dist_c10d._world.pg_map[default_pg] = default_pg, store
    ttp_logger.LOGGER.info(f"cur_rank:{save_rank} successfully launch tcp store")
    return RET_OK


def launch_tcp_store_server(url: str, world_size: int):
    ttp_logger.LOGGER.info(f"ttp_launch_tcp_store_server, world_size:{world_size}, url:{url}")
    os.environ["TORCHELASTIC_USE_AGENT_STORE"] = "False"
    try:
        store, _, _ = next(_rendezvous_helper(url, 0, world_size))
    except Exception:
        ttp_logger.LOGGER.error(f"launch tcp store server failed. world_size:")
        return RET_ERROR
    ttp_logger.LOGGER.info(f"start tcp store server success, world_size:{world_size}, url:{url}")
    return RET_OK


def _process_group_name(ranks, use_hashed_name):
    from torch.distributed.distributed_c10d import _world
    global repair_id_, rank_
    if use_hashed_name:
        # The hashlib.sha1 is used only as a hash algorithm to prevent group name conflicts.
        # It is irrelevant to the encryption of sensitive information.
        pg_name = hashlib.sha1(bytes("_".join(map(str, ranks)), "utf-8")).hexdigest() + '_' + str(repair_id_)
        while pg_name in _world.pg_names.values():
            pg_name = hashlib.sha1(bytes(pg_name + "_", "utf-8")).hexdigest()
        ttp_logger.LOGGER.debug(f"[new group] rank:{rank_} repair_id_:{repair_id_} ranks:{ranks} pg_name:{pg_name}")
    else:
        if ranks == []:
            pg_name = str(repair_id_) + '__'
            ttp_logger.LOGGER.debug('[torch] new_all_group, pg_name:%s  group_count:%s', pg_name, _world.group_count)
        else:
            pg_name = str(_world.group_count) + '_' + str(repair_id_)
            _world.group_count += 1
    return pg_name


def stop_callback():
    global uce_error_, hccl_error_
    if need_pause_ == PauseType.PAUSE:
        continue_callback()  # for pause status when error occurred
    with force_stop_cond_:
        start_time = time.time()
        ret = save_handler.execute_stop()
        tft_reset_limit_step()
        if uce_error_ or hccl_error_:
            ttp_logger.LOGGER.info(f"rank:{rank_} uce or hccl error , no need wait, end stop callback")
            return ret

        if not force_stop_cond_.wait(ACTION_TIMEOUT):
            ttp_logger.LOGGER.error(f"rank:{rank_} wait stop timeout.")
            ret = RET_ERROR

        ttp_logger.LOGGER.info(f'[stop] rank:{rank_} stop consumed:{time.time() - start_time:.3f}s, ret:{ret}')

    return ret


def clean_callback():
    global uce_error_, hccl_error_
    ret = save_handler.execute_clean()
    uce_error_, hccl_error_ = False, False
    return ret


def repair_callback(repair_args: ttp_c2python_api.RepairContext):
    global repair_id_
    repair_id_ = repair_args.repair_id
    save_handler.set_repair_rollback_config(repair_args)
    ttp_logger.LOGGER.info(f"rank:{rank_} start to repair callback, repair_id_:{repair_id_}")
    return save_handler.execute_repair()


def rollback_callback(rollback_args: ttp_c2python_api.RepairContext):
    global repair_id_
    repair_id_ = rollback_args.repair_id
    save_handler.set_repair_rollback_config(rollback_args)
    ttp_logger.LOGGER.info(f"rank:{rank_} start to rollback callback, repair_id_:{repair_id_}")
    return save_handler.execute_rollback()


def pause_callback(pause_step: int, hot_switch: bool):
    global need_pause_, need_pause_cond_, pause_step_
    ttp_logger.LOGGER.info(f"rank:{rank_} start to pause callback, pause_step_:{pause_step}")
    start_time = time.time()
    ret = RET_OK
    with need_pause_cond_:
        pause_step_ = pause_step
        need_pause_ = PauseType.RAISE if hot_switch else PauseType.PAUSE
        if not need_pause_cond_.wait(ACTION_TIMEOUT):
            ret = RET_ERROR
    ttp_logger.LOGGER.info(f'[Pause] rank:{rank_} pause training consumed:{time.time() - start_time:.3f}s, ret:{ret}')
    return save_handler.execute_sync_stream()


def continue_callback():
    global need_pause_, need_pause_cond_
    with need_pause_cond_:
        need_pause_ = None
        need_pause_cond_.notify()
    return RET_OK


def tft_pause_train(cur_step: int):
    """
    Suspend training when necessary
    """
    if cur_step < 0:
        ttp_logger.LOGGER.error(f"pause train failed, cur_step: {cur_step} invalid!")
        return
    global need_pause_, need_pause_cond_, pause_step_
    ttp_logger.LOGGER.debug(f"[pause] rank: {rank_} need_pause_: {need_pause_}, "
                            f"pause_step_: {pause_step_}, cur_step:{cur_step}. ")
    with need_pause_cond_:
        if need_pause_ in [PauseType.PAUSE, PauseType.RAISE] and cur_step == pause_step_:
            ttp_logger.LOGGER.info("[pause] training paused, rank:%s", rank_)
            need_pause_cond_.notify()
            if need_pause_ == PauseType.RAISE:
                raise RuntimeError("STEP FINISH")
            need_pause_cond_.wait()


def execute_rebuild_group(fault_ranks, repair_id):
    global repair_id_
    repair_id_ = repair_id

    global rank_
    ttp_logger.LOGGER.info("[ARF] rebuild group, rank:%s, repairId:%s", rank_, repair_id_)
    return save_handler.execute_rebuild_group(fault_ranks)


def zit_downgrade_rebuild_callback(comm_group_idx: list, comm_groups: list, repair_id: int, zit_param: str):
    global repair_id_
    repair_id_ = repair_id

    global rank_
    ttp_logger.LOGGER.info("[ZIT] degrade rebuild group, rank:%s, repairId:%s", rank_, repair_id)
    return save_handler.execute_zit_downgrade_rebuild(comm_groups, zit_param)


def zit_upgrade_rebuild_callback(rank_list: list, repair_id: int, zit_param: str):
    global repair_id_
    repair_id_ = repair_id

    global rank_
    ttp_logger.LOGGER.info("[ZIT] upgrade rebuild group, rank:%s, repairId:%s", rank_, repair_id_)
    return save_handler.execute_zit_upgrade_rebuild(rank_list, zit_param)


def zit_upgrade_repair_callback(repair_args: ttp_c2python_api.RepairContext):
    global repair_id_
    repair_id_ = repair_args.repair_id
    save_handler.set_repair_rollback_config(repair_args)
    ttp_logger.LOGGER.info(f"[ZIT] rank:{rank_} start upgrade to repair callback, repair_id_:{repair_id_}")

    return save_handler.execute_zit_upgrade_repair()


def zit_upgrade_rollback_callback(rollback_args: ttp_c2python_api.RepairContext):
    global repair_id_
    repair_id_ = rollback_args.repair_id
    save_handler.set_repair_rollback_config(rollback_args)
    ttp_logger.LOGGER.info(f"[ZIT] rank:{rank_} start to rollback callback, repair_id_:{repair_id_}")

    return save_handler.execute_zit_upgrade_rollback()


def set_mindio_export_version(framework_type: str):
    # set framework type and version
    if framework_type not in ["MindSpeed", "MindSpeed-LLM"]:
        raise Exception(f"This framework: {framework_type} is not supported by mindio tft now.\n")

    global mindio_export_function_version

    mindio_export_function_version = framework_type
    os.environ["TTP_STOP_CLEAN_BEFORE_DUMP"] = "1"
    if rank_ == 0:
        ttp_logger.LOGGER.info(f"set framework: {framework_type} for mindio tft done\n")


def get_mindio_export_version():
    global mindio_export_function_version
    return mindio_export_function_version


def tft_set_optimizer_replica(rank: int, replica_info: list):
    rank_list = []
    replica_cnt = []
    replica_shift = []
    if rank < 0:
        raise Exception(f"set_optimizer_replica failed, rank:{rank}, rank must be greater than 0")
    if rank >= TTP_MAX_WORLD_SIZE:
        raise Exception(f"set_optimizer_replica failed, rank:{rank}, rank must be less than MAX_WORLD_SIZE(100000)")

    for _, replica_dict in enumerate(replica_info):
        tmp_rank_list = replica_dict.get('rank_list', None)
        tmp_replica_cnt = replica_dict.get('replica_cnt', None)
        tmp_replica_shift = replica_dict.get('replica_shift', None)

        valid_flag = (isinstance(tmp_rank_list, list) and isinstance(tmp_replica_cnt, int) and
                      isinstance(tmp_replica_shift, int))
        if not valid_flag:
            raise Exception(f"set_optimizer_replica failed, rank:{rank}, replica_info is invalid")
        for element in tmp_rank_list:
            if not isinstance(element, int):
                raise Exception(f"set_optimizer_replica failed, rank:{rank}, replica_info is invalid")

        rank_list.append(tmp_rank_list)
        replica_cnt.append(tmp_replica_cnt)
        replica_shift.append(tmp_replica_shift)

    ret = ttp_c2python_api.set_optimizer_replica(rank_list, replica_cnt, replica_shift)
    if ret != RET_OK:
        ttp_logger.LOGGER.error("set_optimizer_replica failed, rank:%s, replica_info:%s", rank, replica_info)
        raise Exception()


def tft_set_dp_group_info(rank: int, dp_rank_list: list):
    check_valid = True
    if rank < 0 or rank >= TTP_MAX_WORLD_SIZE:
        check_valid = False
    elif isinstance(dp_rank_list, list):
        for element in dp_rank_list:
            if not isinstance(element, int):
                check_valid = False
    else:
        check_valid = False
    if not check_valid:
        raise Exception(f"set_dp_group_info input params invalid, rank:{rank}, dp_rank_list:{dp_rank_list}")

    ret = ttp_c2python_api.set_dp_group_info(dp_rank_list)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"set_dp_group_info failed, rank:{rank}, dp_rank_list:{dp_rank_list}")
        raise Exception()


def is_valid_ip(ip_str):
    if len(ip_str) > 15:
        ttp_logger.LOGGER.warning(f"rank:{rank_} input illegal ipv4 address: length exceeds 15.")
        return False
    ip_pattern = r'^(?:(?:25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:$|(?!\.$)\.)){4}$'
    return bool(re.match(ip_pattern, ip_str))


def is_loopback_address(ip):
    if len(ip) > 15:
        ttp_logger.LOGGER.warning(f"rank:{rank_} input illegal ipv4 address: length exceeds 15.")
        return False
    lb_pattern = r"^127\.0*\.0*\.0*1$"
    return re.match(lb_pattern, ip) is not None


def get_local_ip(master_ip, local_ip):
    if local_ip != '' or is_loopback_address(master_ip) or is_zero_ip(master_ip):
        return local_ip

    try:
        ips1 = master_ip.split('.')
        base = int(ips1[0]) * 256 + int(ips1[1])
        cmd = os.popen('hostname -I')
        ip_list = cmd.read().strip().split(' ')
        cmd.close()
        for ip in ip_list:
            if not is_valid_ip(ip):
                return local_ip

            ips2 = ip.split('.')
            val = int(ips2[0]) * 256 + int(ips2[1])
            if base == val:
                return ip

        if ip_list:
            return ip_list[0]
        else:
            return local_ip
    except Exception as e:
        return local_ip


def wrap_exit():
    exit_proc_thread()


def tft_exception_handler(func: Callable):
    """
    Save Final Ckpt: A wrapper decorator on a training function to catch exception
    """
    if not callable(func):
        ttp_logger.LOGGER.error("tft_exception_handler: func must be a callable")
        raise TypeError("func must be a callable")

    def ttp_destroy():
        if rank_ == 0 and mindx_agent is None:
            tft_destroy_controller()
        tft_destroy_processor()
        ttp_logger.LOGGER.debug(f"rank:{rank_} tft destroy end.")

    def handle_l2_hbm_error(err_str):
        handle_hbm_err = "HBM MULTI BIT ECC" in err_str
        can_repair = True
        if handle_hbm_err:
            hbm_error_time = get_l2_hbm_error_time(err_str)
            ttp_logger.LOGGER.info(f"rank:{rank_} handle hbm error time: {hbm_error_time}")
            can_repair = tft_can_do_uce_repair(hbm_error_time)
        return handle_hbm_err, can_repair

    def report_and_wait(err_str, can_repair):
        for dict_str in fault_dict:
            if dict_str in err_str:
                if not can_repair and dict_str == 'HBM MULTI BIT ECC':
                    ttp_logger.LOGGER.warning(f"rank:{rank_} catch HBM exception, but can't repair "
                                              f"ex instance:{err_str}")
                    tft_report_error(ReportState.RS_UCE_CORRUPTED.value)
                else:
                    ttp_logger.LOGGER.warning(f"rank:{rank_} catch {dict_str} exception, "
                                              f"ex instance:{err_str}")
                    tft_report_error(ReportState[fault_dict[dict_str]].value)
                # wait stop & clean finish
                if dict_str not in ['ARF FINISH', 'STEP FINISH']:
                    ret = tft_wait_next_action()
                    if ret != Action.RETRY.value:
                        wrap_exit()

    @wraps(func)
    def wrapper(*args, **kwargs):
        if mindio_export_function_version not in ["MindSpeed", "MindSpeed-LLM"]:
            args = list(args)
        save_handler.set_model_config(args)
        memory_allocated_before = 0
        max_memory_allocated_before = 0
        memory_reserved_before = 0
        max_memory_reserved_before = 0
        byte_to_gb = 1024 * 1024 * 1024
        while True:
            wait_next = False
            try:
                iteration = func(*args, **kwargs)
                ttp_destroy()
                return iteration
            except RuntimeError as e:
                err_str = str(e)
                handle_l2_hbm_err, can_repair = handle_l2_hbm_error(err_str)
                handle_uce_err = save_handler.get_uce() and ("UCE ERROR" in err_str or handle_l2_hbm_err)
                start_time, end_time = get_update_start_end_time()
                ttp_logger.LOGGER.info(f"rank:{rank_} optimizer start update time:{start_time}, "
                                       f"end update time:{end_time}")
                if handle_uce_err or any(s in err_str for s in
                        {"ARF FINISH", "STEP FINISH", "FORCE STOP", "HCCL OP RETRY FAILED", "SUSPECT REMOTE ERROR"}):
                    memory_allocated_before = torch_npu.npu.memory_allocated() / byte_to_gb
                    max_memory_allocated_before = torch_npu.npu.max_memory_allocated() / byte_to_gb
                    memory_reserved_before = torch_npu.npu.memory_reserved() / byte_to_gb
                    max_memory_reserved_before = torch_npu.npu.max_memory_reserved() / byte_to_gb
                    report_and_wait(err_str, can_repair)
                    wait_next = True
                else:
                    ttp_logger.LOGGER.exception("rank:%s catch other exception", rank_)
                    ttp_logger.LOGGER.info(f"other exception str is : {err_str}")
                    err_type = ReportState.RS_UCE_CORRUPTED.value if (handle_l2_hbm_err and not can_repair) \
                        else ReportState.RS_UNKNOWN.value
                    tft_report_error(err_type)  # run ttp
                    raise e
            except Exception as e:
                ttp_logger.LOGGER.exception("rank:%s catch other exception", rank_)
                tft_report_error(ReportState.RS_UNKNOWN.value)
                raise e
            finally:
                # release memory for repair stage
                gc.collect()
                start_time = time.time()
                torch_npu.npu.empty_cache()
                ttp_logger.LOGGER.info(f'[repair stage] rank:{rank_} empty_cache cost:{time.time() - start_time}s')
                if wait_next:
                    # wait repair & rollback finish
                    ret = tft_wait_repair_action()
                    if ret != Action.RETRY.value:
                        wrap_exit()
                # release memory for normal stage
                gc.collect()
                start_time = time.time()
                torch_npu.npu.empty_cache()
                ttp_logger.LOGGER.info(f'[normal stage] rank:{rank_} empty_cache cost:{time.time() - start_time}s')

                memory_allocated_after = torch_npu.npu.memory_allocated() / byte_to_gb
                max_memory_allocated_after = torch_npu.npu.max_memory_allocated() / byte_to_gb
                memory_reserved_after = torch_npu.npu.memory_reserved() / byte_to_gb
                max_memory_reserved_after = torch_npu.npu.max_memory_reserved() / byte_to_gb

                ttp_logger.LOGGER.info(f"[memory] rank:{rank_}"
                                       f", memory_allocated_before:{format(memory_allocated_before, '.4f')} GB"
                                       f", max_memory_allocated_before:{format(max_memory_allocated_before, '.4f')} GB"
                                       f", memory_reserved_before:{format(memory_reserved_before, '.4f')} GB"
                                       f", max_memory_reserved_before:{format(max_memory_reserved_before, '.4f')} GB"
                                       f", memory_allocated_after:{format(memory_allocated_after, '.4f')} GB"
                                       f", max_memory_allocated_after:{format(max_memory_allocated_after, '.4f')} GB"
                                       f", memory_reserved_after:{format(memory_reserved_after, '.4f')} GB"
                                       f", max_memory_reserved_after:{format(max_memory_reserved_after, '.4f')} GB")

    return wrapper


def tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=True,
                       tls_info='', enable_uce=True, enable_arf=False, enable_zit=False):
    """
    init processor
    set checkpoint callback function use to ckpt when have train task failed
    set rename callback function use to rename when ckpt finished
    enable_local_copy: false for MS
    :param enable_arf: false for default
    :param enablez_zit: false for default
    """
    if rank < 0 or rank >= world_size or world_size > TTP_MAX_WORLD_SIZE:
        ttp_logger.LOGGER.error(f"init processor {rank} failed, world_size {world_size}, Out of range")
        wrap_exit()

    global rank_
    global device_
    global repair_id_
    rank_ = rank
    repair_id_ = 0
    save_handler.register()
    save_handler.set_uce(enable_uce)
    if not mind_spore:
        device_ = torch.npu.current_device()
        torch.npu.SyncLaunchStream()
        torch.distributed.distributed_c10d._process_group_name = _process_group_name
    else:
        device_ = ms.context.get_context("device_id")

    ret = ttp_c2python_api.init_processor(rank, world_size, enable_local_copy, enable_tls,
                                          tls_info, enable_uce, enable_arf, enable_zit)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor {rank_} failed, error:{ret}")
        wrap_exit()
    set_processor_callback()


def set_zit_processor_callback():
    ret = ttp_c2python_api.set_zit_downgrade_rebuild_callback(zit_downgrade_rebuild_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set degrade rebuild callback, error num:{ret}")
        raise Exception(f"init processor failed to set degrade rebuild callback, error num:{ret}")
    ret = ttp_c2python_api.set_zit_upgrade_rebuild_callback(zit_upgrade_rebuild_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set upgrade rebuild callback, error num:{ret}")
        raise Exception(f"init processor failed to set upgrade rebuild callback, error num:{ret}")
    ret = ttp_c2python_api.set_zit_upgrade_repair_callback(zit_upgrade_repair_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set upgrade repair callback, error num:{ret}")
        raise Exception(f"init processor failed to set upgrade repair callback, error num:{ret}")
    ret = ttp_c2python_api.set_zit_upgrade_rollback_callback(zit_upgrade_rollback_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set upgrade rollback callback, error num:{ret}")
        raise Exception(f"init processor failed to set upgrade rollback callback, error num:{ret}")


def set_processor_callback():
    ret = ttp_c2python_api.set_ckpt_callback(save_checkpoint_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set ckpt callback, error num:{ret}")
        raise Exception(f"init processor failed to set ckpt callback, error num:{ret}")
    ret = ttp_c2python_api.set_rename_callback(rename_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set rename callback, error num:{ret}")
        raise Exception(f"init processor failed to set rename callback, error num:{ret}")
    ret = ttp_c2python_api.set_exit_callback(exit_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set exit callback, error num:{ret}")
        raise Exception(f"init processor failed to set exit callback, error num:{ret}")
    ret = ttp_c2python_api.set_stop_device_callback(stop_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set stop device callback, error num:{ret}")
        raise Exception(f"init processor failed to set stop device callback, error num:{ret}")
    ret = ttp_c2python_api.set_clean_device_callback(clean_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set clean device callback, error num:{ret}")
        raise Exception(f"init processor failed to set clean device callback, error num:{ret}")
    ret = ttp_c2python_api.set_repair_callback(repair_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set repair callback, error num:{ret}")
        raise Exception(f"init processor failed to set repair callback, error num:{ret}")
    ret = ttp_c2python_api.set_communication_operate_callback(execute_rebuild_group)
    if ret != RET_OK:
        raise Exception(f"init processor failed to set communication operate callback, error num:{ret}")
    ret = ttp_c2python_api.set_launch_tcp_store_client_callback(launch_tcp_store_client)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set launch tcp store client callback, error num:{ret}")
        raise Exception(f"init processor failed to set launch tcp store client callback, error num:{ret}")
    ret = ttp_c2python_api.set_launch_tcp_store_server_callback(launch_tcp_store_server)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set launch tcp store server callback, error num:{ret}")
        raise Exception(f"init processor failed to set launch tcp store server callback, error num:{ret}")
    ret = ttp_c2python_api.set_rollback_callback(rollback_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set rollback callback, error num:{ret}")
        raise Exception(f"init processor failed to set rollback callback, error num:{ret}")
    ret = ttp_c2python_api.set_pause_callback(pause_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set pause callback, error num:{ret}")
        raise Exception(f"init processor failed to set pause callback, error num:{ret}")
    ret = ttp_c2python_api.set_continue_callback(continue_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init processor failed to set continue callback, error num:{ret}")
        raise Exception(f"init processor failed to set continue callback, error num:{ret}")
    set_zit_processor_callback()


def tft_start_processor(master_ip: str, port: int, local_ip=''):
    global rank_
    global repair_id_
    if is_zero_ip(master_ip) or is_zero_ip(local_ip):
        ttp_logger.LOGGER.error(f"start processor: {rank_} failed, all-zero ip is not supported ")
        wrap_exit()
    controller_ip = input_ip_transform(master_ip)
    processor_ip = input_ip_transform(local_ip)
    processor_ip = get_local_ip(controller_ip, processor_ip)
    ret = ttp_c2python_api.start_processor(controller_ip, port, processor_ip)
    if ret != RET_OK:
        ttp_logger.LOGGER.error("start processor failed, error:%s", ret)
        wrap_exit()


def tft_is_reboot_node():
    global rank_, repair_id_
    processor_repair_id = ttp_c2python_api.get_repair_id()
    reboot_flag = processor_repair_id != 0
    if reboot_flag and not check_pytorch_version():
        processor_repair_id = processor_repair_id + 0.5
        repair_id_ = processor_repair_id
    if reboot_flag:
        ttp_logger.LOGGER.info("[TFT] node reboot, rank:%s repair_id:%s ", rank_, repair_id_)
    return reboot_flag


def tft_get_reboot_type():
    hot_switch = ttp_c2python_api.get_hot_switch()
    return "arf" if not hot_switch else "hot switch"


def check_pytorch_version():
    if mind_spore:
        return True
    version_parts = torch_npu.__version__.split('.')
    ttp_logger.LOGGER.debug(f"PTA version:{torch_npu.__version__} ")
    version_prefix = '.'.join(version_parts[:3])

    def cast_version_number(version):
        return [int(c) if c.isdigit() else c for c in re.split(r'(\d+)', version)]

    if version_prefix == '2.1.0' and (cast_version_number(version_parts[3]) < cast_version_number('post11')):
        ttp_logger.LOGGER.warning(f"Current PTA version may not support reinit_process_group API")
        return False
    ttp_logger.LOGGER.debug(f"Current PTA version support reinit_process_group API")
    return True


def tft_destroy_processor():
    ret = ttp_c2python_api.destroy_processor()
    if ret != RET_OK:
        ttp_logger.LOGGER.error("destroy processor failed, error num:%s", ret)
        raise Exception(f"destroy processor failed, error num:{ret}")


def tft_start_copy_os():
    """
    notify processor beging to local copy optimizer state
    """
    ret = ttp_c2python_api.start_copying()
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"notify start to local copy os failed, error num:{ret}")
        raise Exception(f"notify start to local copy os failed, error num:{ret}")


def tft_start_updating_os(backup_step: int = -1):
    """
    notify processor begin to updating optimizer state
    if processor is start to ckpt, then return failed
    """
    ret = ttp_c2python_api.start_updating(backup_step)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"notify start to updating os failed, error num:{ret}")
        raise Exception(f"notify start to updating os failed, error num:{ret}")


def tft_end_updating_os(step: int):
    """
    notify processor already finished updating optimizer state
    """
    ret = ttp_c2python_api.end_updating(step)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"notify end to updating os failed, error num:{ret}")
        raise Exception(f"notify end to updating os failed, error num:{ret}")


def tft_report_load_ckpt_step(step: int):
    ret = ttp_c2python_api.report_load_ckpt_step(step)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"report step failed while load from ckpt, error num:{ret}")
        raise Exception(f"report step failed while load from ckpt, error num:{ret}")


def tft_set_step_args(args):
    """
    set args after every train step
    """
    if args is None:
        ttp_logger.LOGGER.warning(f"set step args input param is None")
        return

    if mindio_export_function_version in ["MindSpeed", "MindSpeed-LLM"]:
        ttp_logger.LOGGER.warning(f"MindSpeed or MindSpeed-LLM no need set args")
        return

    save_handler.set_model_config(args)


def tft_report_error(error_type: ReportState):
    """
    report error state && unify condition variable post process
    """
    global uce_error_, hccl_error_
    ret = ttp_c2python_api.report_status(error_type)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"report status: {error_type} failed, error num:{ret}")
        raise Exception(f"report status: {error_type} failed, error num:{ret}")

    if error_type in [ReportState.RS_UCE.value, ReportState.RS_UCE_CORRUPTED.value]:
        uce_error_ = True
    elif error_type == ReportState.RS_HCCL_FAILED.value:
        hccl_error_ = True
        notify_stop_callback_return()
    elif error_type == ReportState.RS_NORMAL.value:
        notify_stop_callback_return()
    elif error_type in [ReportState.RS_STEP_FINISH.value, ReportState.RS_INIT_FINISH.value,
                        ReportState.RS_PREREPAIR_FINISH.value]:
        pass
    else:
        ttp_logger.LOGGER.error(f"rank:{rank_} catch other exception")
        save_handler.wait_exit_cond()


def notify_stop_callback_return():
    with force_stop_cond_:
        force_stop_cond_.notify()
        ttp_logger.LOGGER.info(f"rank:{rank_} notify stop callback end waiting")


def tft_reset_limit_step():
    ret = ttp_c2python_api.reset_limit_step()
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"end waiting at end_updating failed, ret: {ret}")


def tft_wait_repair_action():
    ret = ttp_c2python_api.wait_repair_action()
    if ret != RET_OK:
        ret = Action.EXIT.value
        ttp_logger.LOGGER.error("wait repair action failed, error num:%s", ret)
    else:
        ret = Action.RETRY.value
    return ret


def tft_wait_next_action():
    ret = ttp_c2python_api.wait_next_action()
    if ret != RET_OK:
        ret = Action.EXIT.value
        ttp_logger.LOGGER.error("wait_next_action failed, error num:%s", ret)
    else:
        ret = Action.RETRY.value
    return ret


def tft_register_exit_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_exit_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_exit_handler(func, ctx)


def tft_register_rename_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_rename_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_rename_handler(func, ctx)


def tft_register_save_ckpt_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_save_ckpt_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_save_ckpt_handler(func, ctx)


def tft_register_stop_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_stop_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_stop_handler(func, ctx)


def tft_register_clean_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_clean_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_clean_handler(func, ctx)


def tft_register_repair_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_repair_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_repair_handler(func, ctx)


def tft_register_rollback_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_rollback_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_rollback_handler(func, ctx)


def tft_register_rebuild_group_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_rebuild_group_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_rebuild_group_handler(func, ctx)


def tft_register_stream_sync_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft register sync stream handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_sync_handler(func, ctx)


def tft_register_zit_upgrade_rollback_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_zit_upgrade_rollback_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_zit_upgrade_rollback_handler(func, ctx)


def tft_register_zit_upgrade_repair_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_zit_upgrade_repair_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_zit_upgrade_repair_handler(func, ctx)


def tft_register_zit_upgrade_rebuild_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_zit_upgrade_rebuild_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_zit_upgrade_rebuild_handler(func, ctx)


def tft_register_zit_downgrade_rebuild_handler(func: Callable, ctx=None):
    if not callable(func):
        ttp_logger.LOGGER.error("tft_register_zit_downgrade_rebuild_handler: func must be a callable")
        raise TypeError("func must be a callable")
    save_handler.register_zit_downgrade_rebuild_handler(func, ctx)


def tft_get_repair_step():
    return save_handler.get_repair_step()


def tft_get_repair_type():
    return ttp_c2python_api.get_repair_type()


def tft_register_decrypt_handler(decryptor: Callable):
    if decryptor is None:
        ttp_logger.LOGGER.info(f"tft_register_decrypt_handler: decryptor is None")
        return
    if not callable(decryptor):
        ttp_logger.LOGGER.error("tft_register_decrypt_handler: func must be a callable")
        raise TypeError("func must be a callable")
    ret = ttp_c2python_api.set_decrypt_callback(decryptor)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"set decryptor failed, error num:{ret}")
        raise Exception(f"set decryptor failed, error num:{ret}")


atexit.register(tft_destroy_processor)