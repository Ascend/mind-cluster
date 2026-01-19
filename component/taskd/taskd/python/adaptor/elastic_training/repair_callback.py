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
import io
import random
import time

import numpy as np
import torch
from megatron.core import mpu, tensor_parallel
from megatron.core.utils import get_model_config
from megatron.core.num_microbatches_calculator import update_num_microbatches
from megatron.training import get_args, get_timers
from megatron.training.checkpointing import (get_rng_state, set_checkpoint_version, check_checkpoint_args,
                                            get_checkpoint_version, fix_query_key_value_ordering, load_checkpoint)
from megatron.training.training import setup_model_and_optimizer
from megatron.training.utils import print_rank_0, unwrap_model
from mindio_ttp.adaptor.tft_replica_group import get_repair_group, build_repair_group
from mindio_ttp.adaptor.utils import TRAIN_PARAM, MODEL_INDEX, OPTIM_INDEX, SCHEDULER_INDEX, CONFIG_INDEX
from mindio_ttp.framework_ttp.ttp_decorator import get_device, tft_report_load_ckpt_step
from mindio_ttp.framework_ttp import OptimizerType, RepairType
from mindio_ttp.controller_ttp import ttp_logger
from mindio_ttp.adaptor import tft_optimizer_data_repair

from . import common
from .constant import send_or_recieve_rank_repair_args, build_model_and_optimizer_result


def repair_callback(step: int, need_rebuild: bool, error_ranks: list, repair_info: dict, train_args, params: str):
    ttp_logger.LOGGER.info(f"repair strategy params: {params}, step: {step}, need_rebuild:{need_rebuild}, "
                           f"error_ranks: {error_ranks}, repair_info: {repair_info}")
    common.check_scale_out_params(params)
    t1 = time.time()
    rank = torch.distributed.get_rank()
    torch.npu.set_device(get_device())
    optim_idxs = repair_info.get('type', OptimizerType.ATTENTION.value)
    repair_type = repair_info.get('repair_type', None)
    src_ranks = repair_info.get('src', None)
    dest_ranks = repair_info.get('dst', None)
    rank_list = repair_info.get('rank_list', None)
    ttp_logger.LOGGER.info(f"repair rank {rank}, repair type {repair_type},src ranks {src_ranks}, dest ranks "
                           f"{dest_ranks}, dest ranks {dest_ranks}, rank_list {rank_list} optim idx {optim_idxs}, "
                           f"step {step}")
    if step <= 0:
        raise ValueError(f"repair step {step} is not valid")
    if repair_type == RepairType.RT_SEND.value:
        send_rank_repair(send_or_recieve_rank_repair_args(src_ranks, dest_ranks, optim_idxs, step, rank_list), train_args[
            TRAIN_PARAM], rank)
    elif repair_type == RepairType.RT_RECV_REPAIR.value:
        recv_rank_repair(send_or_recieve_rank_repair_args(src_ranks, dest_ranks, optim_idxs, step, rank_list), need_rebuild, train_args[TRAIN_PARAM],
                         rank)
    else:
        ttp_logger.LOGGER.error(f"rank:{rank} repair type {repair_type} not supported")
        raise ValueError(f"rank:{rank} repair type {repair_type} not supported")
    ttp_logger.LOGGER.info(f"repair rank {rank} repair total time consumed: {time.time()-t1:.3f}s")


def send_rank_repair(send_rank_repair_args: send_or_recieve_rank_repair_args, train_args, rank):
    t1 = time.time()
    send_rank_repair_args.rank_list.sort()
    build_repair_group(send_rank_repair_args.rank_list)
    group = get_repair_group()
    t2 = time.time()
    for idx, src_rank in enumerate(send_rank_repair_args.src_ranks):
        dest_rank, optim_idx = send_rank_repair_args.dest_ranks[idx], send_rank_repair_args.optim_idxs[idx]
        if src_rank != rank:
            ttp_logger.LOGGER.error(f"repair current rank {rank} is not equal to src rank {src_rank}")
            raise ValueError("src rank is not equal to current rank")
        if src_rank == dest_rank:
            ttp_logger.LOGGER.error(f"repair src rank {src_rank} and dest rank {dest_rank} are not allowed to be the"
                                    f"same")
            raise ValueError("src rank is equal to dest rank")
        save_and_send_ckpt(dest_rank, send_rank_repair_args.step, optim_idx, train_args, rank)

    t3 = time.time()
    for idx, _ in enumerate(send_rank_repair_args.src_ranks):
        dest_rank, optim_idx = send_rank_repair_args.dest_ranks[idx], send_rank_repair_args.optim_idxs[idx]
        train_args[OPTIM_INDEX].send_optim_param_state(dest_rank, group, optim_idx)

    t4 = time.time()
    tft_optimizer_data_repair.convert_log_args_to_tensors()
    for dest_rank in send_rank_repair_args.dest_ranks:
        tft_optimizer_data_repair.send_log_args(dest_rank)
    tft_optimizer_data_repair.convert_log_tensors_to_args()

    t5 = time.time()
    ttp_logger.LOGGER.info(f"repair rank {rank} send total time consumed: {t5-t1:.3f}s, "
                           f"build repair group: {t2 - t1:.3f}s, "
                           f"save and send ckpt: {t3 - t2:.3f}s, "
                           f"send optim: {t4 - t3:.3f}s, "
                           f"send log args: {t5 - t4:.3f}s")


def save_and_send_ckpt(dest_rank, step, optim_idx, train_args, rank):
    """
    save memory checkpoint and send to dest rank.
    """
    t1 = time.time()
    state_dict = save_memory_ckpt(train_args[OPTIM_INDEX], train_args[SCHEDULER_INDEX], step, rank, optim_idx)
    buffer = io.BytesIO()
    torch.save(state_dict, buffer)
    state_dict_bytes = buffer.getvalue()
    state_dict_tensor = torch.ByteTensor(torch.ByteStorage.from_buffer(state_dict_bytes)).to('npu')

    t2 = time.time()
    # send tensor size first
    size_tensor = torch.tensor([state_dict_tensor.numel()], dtype=torch.long).to('npu')
    torch.distributed.send(size_tensor, dst=dest_rank, group=get_repair_group())

    # send the serialized state_dict tensor
    torch.distributed.send(state_dict_tensor, dst=dest_rank, group=get_repair_group())
    ttp_logger.LOGGER.info(f"repair rank {rank} save ckpt: {t2 - t1:.3f}s, "
                           f"send ckpt: {time.time() - t1:.3f}s")


def recv_rank_repair(recv_rank_repair_args: send_or_recieve_rank_repair_args, need_rebuild: bool, train_args, rank):
    recv_rank_repair_args.rank_list.sort()
    build_repair_group(recv_rank_repair_args.rank_list)
    if need_rebuild:
        result = build_model_and_optimizer(
            tft_optimizer_data_repair.model_provider_, tft_optimizer_data_repair.model_type_, True)
        train_args[MODEL_INDEX] = result.model
        train_args[OPTIM_INDEX] = result.optimizer
        train_args[SCHEDULER_INDEX] = result.lr_scheduler
        train_args[CONFIG_INDEX] = result.config
    group = get_repair_group()
    for idx, src_rank in enumerate(recv_rank_repair_args.src_ranks):
        dest_rank, optim_idx = recv_rank_repair_args.dest_ranks[idx], recv_rank_repair_args.optim_idxs[idx]
        if dest_rank != rank:
            ttp_logger.LOGGER.error(f"repair rank {rank} is not equal to dest rank")
            raise ValueError("dest rank is not equal current rank")
        if src_rank == dest_rank:
            ttp_logger.LOGGER.error(f"repair src rank {src_rank} and dest rank {dest_rank} is not allowed to be the "
                                    f"same")
            raise ValueError("src rank is equal to dest rank")
        recv_ckpt_from_peer(src_rank, dest_rank, recv_rank_repair_args.step, recv_rank_repair_args.rank_list)
    # combine state_dict and once load, fix precision problem
    state_dict = tft_optimizer_data_repair.temp_memory_ckpt
    load_memory_ckpt(train_args[MODEL_INDEX], train_args[OPTIM_INDEX], train_args[SCHEDULER_INDEX], state_dict,
                     None)

    for idx, src_rank in enumerate(recv_rank_repair_args.src_ranks):
        dest_rank, optim_idx = recv_rank_repair_args.dest_ranks[idx], recv_rank_repair_args.optim_idxs[idx]
        train_args[OPTIM_INDEX].recv_and_load_optim_param_state(src_rank, group, recv_rank_repair_args.step, optim_idx)

    tft_optimizer_data_repair.convert_log_args_to_tensors()
    for src_rank in recv_rank_repair_args.src_ranks:
        tft_optimizer_data_repair.recv_log_args(src_rank)
    tft_optimizer_data_repair.convert_log_tensors_to_args()

    ttp_logger.LOGGER.info(f"repair rank {rank} recv finish")


def build_model_and_optimizer(model_provider, model_type, skip_load):
    args = get_args()
    if skip_load:
        load, args.load = args.load, None
    from mindio_ttp.adaptor.tft_replica_group import get_local_embedding_group
    ori_embedding_group = mpu._EMBEDDING_GROUP
    mpu._EMBEDDING_GROUP = get_local_embedding_group()
    model, optimizer, lr_scheduler = setup_model_and_optimizer(model_provider, model_type)
    mpu._EMBEDDING_GROUP = ori_embedding_group
    if skip_load:
        args.load = load
    config = get_model_config(model[0])
    return build_model_and_optimizer_result(model, optimizer, lr_scheduler, config)


def recv_ckpt_from_peer(src_rank, dest_rank, step, rank_list: list):
    """
    receive memory checkpoint and repair train() param
    """
    # receive tensor size first
    size_tensor = torch.tensor([0], dtype=torch.long, device='npu')
    torch.distributed.recv(size_tensor, src=src_rank, group=get_repair_group())
    size = size_tensor.item()

    # receive the serialized state_dict tensor
    state_dict_tensor = torch.empty(size, dtype=torch.uint8, device='npu')
    torch.distributed.recv(state_dict_tensor, src=src_rank, group=get_repair_group())

    # deserialize the state_dict
    state_dict_bytes = state_dict_tensor.cpu().numpy().tobytes()
    buffer = io.BytesIO(state_dict_bytes)

    device_count = torch.npu.device_count()
    if device_count == 0:
        raise ValueError("device count is 0")
    map_location = {'npu:' + str(src_rank % device_count): 'npu:' + str(dest_rank % device_count)}
    loaded_state_dict = torch.load(buffer, map_location=map_location, weights_only=False)
    tft_optimizer_data_repair.set_memory_ckpt(loaded_state_dict)


def load_memory_ckpt(model, optimizer, opt_param_scheduler, state_dict, optim_idx):
    rank = torch.distributed.get_rank()
    args = get_args()
    model = unwrap_model(model)
    if state_dict is None:
        return 0
    args_str = 'args'
    set_checkpoint_version(state_dict.get('checkpoint_version', 0))
    args.iteration = state_dict['iteration']
    args.num_query_groups = state_dict[args_str].num_query_groups
    args.curr_iteration = state_dict[args_str].curr_iteration
    args.do_train, args.do_valid, args.do_test = \
        state_dict[args_str].do_train, state_dict[args_str].do_valid, state_dict[args_str].do_test
    args.num_floating_point_operations_so_far = state_dict['num_floating_point_operations_so_far']
    # check arguments
    if args_str in state_dict and not args.finetune:
        checkpoint_args = state_dict[args_str]
        check_checkpoint_args(checkpoint_args)
        args.consumed_train_samples = getattr(checkpoint_args, 'consumed_train_samples', 0)
        update_num_microbatches(consumed_samples=args.consumed_train_samples)
        args.consumed_valid_samples = getattr(checkpoint_args, 'consumed_valid_samples', 0)
    else:
        print_rank_0('could not find arguments in the checkpoint')

    # fix up query/key/value matrix ordering if needed
    checkpoint_version = get_checkpoint_version()
    print_rank_0(f'checkpoint version: {checkpoint_version}')
    fix_query_key_value_ordering(model, checkpoint_version)
    # optimizer
    if hasattr(optimizer, 'optim_nums') and optimizer.optim_nums > 1 and optim_idx is not None:
        optimizer.load_state_dict_by_idx(state_dict['optimizer'], optim_idx)
    else:
        optimizer.load_state_dict_memory(state_dict['optimizer'])

    opt_param_scheduler.load_state_dict(state_dict['opt_param_scheduler'])
    # rng states
    if 'rng_state' in state_dict:
        rng_state = state_dict['rng_state'][0]
        random.setstate(rng_state['random_rng_state'])
        np.random.set_state(rng_state['np_rng_state'])
        torch.set_rng_state(rng_state['torch_rng_state'])
        torch.cuda.set_rng_state(rng_state['cuda_rng_state'])
        tensor_parallel.get_cuda_rng_tracker().set_states(rng_state['rng_tracker_states'])
    ttp_logger.LOGGER.info(f"rank:{rank} successfully load checkpoint at iteration {args.iteration} to memory")
    return args.iteration


def save_memory_ckpt(optimizer, opt_param_scheduler, step, rank, optim_idx=None):
    args = get_args()
    state_dict = {}
    if hasattr(optimizer, 'optim_nums') and optimizer.optim_nums > 1 and optim_idx is not None:
        state_dict['optimizer'] = optimizer.state_dict_by_idx(optim_idx)
    else:
        state_dict['optimizer'] = optimizer.state_dict_memory()
    rng_state = get_rng_state(args.ckpt_format)
    state_dict['args'] = args
    state_dict['iteration'] = args.iteration
    state_dict['checkpoint_version'] = 3.0
    state_dict['rng_state'] = rng_state
    state_dict['opt_param_scheduler'] = opt_param_scheduler.state_dict()
    state_dict['num_floating_point_operations_so_far'] = args.num_floating_point_operations_so_far
    ttp_logger.LOGGER.info(f'rank:{rank} successfully save checkpoint at iteration {step} to memory')
    return state_dict
