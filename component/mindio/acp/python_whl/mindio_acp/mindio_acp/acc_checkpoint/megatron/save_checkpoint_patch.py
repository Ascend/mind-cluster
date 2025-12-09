#!/usr/bin/env python
# coding=utf-8
# Copyright (c) 2024, NVIDIA CORPORATION. All rights reserved.
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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
#
# Modification description: Patch the Megatron framework save and load functions using
# MindIO's asynchronous save and load methods for acceleration.

import os
import math
from functools import wraps, partial
from time import time
from collections import defaultdict

import torch
from megatron.core import mpu, tensor_parallel, dist_checkpointing
from megatron.core.optimizer import ChainedOptimizer
from megatron.training import global_vars, utils, ft_integration, wandb_utils
from megatron.training.checkpointing import ensure_directory_exists
from megatron.training.checkpointing import generate_state_dict
from megatron.training.checkpointing import get_checkpoint_name
from megatron.training.checkpointing import get_checkpoint_tracker_filename
from megatron.training.checkpointing import get_distributed_optimizer_checkpoint_name
from megatron.training.checkpointing import get_rng_state
from mindio_acp.common import mindio_logger
from mindio_acp.acc_checkpoint.utils.utils import time_used

logging = mindio_logger.LOGGER


def import_torch_mindio():
    global torch_npu, print_rank_0, CheckpointHelper, mindio_acp_flush
    import torch_npu
    from mindio_acp import flush as mindio_acp_flush
    from mindio_acp.acc_checkpoint.utils.utils import print_rank_0
    from mindio_acp.acc_checkpoint.framework_acp import CheckpointHelper


def acp_save_checkpoint(iteration, model, optimizer, opt_param_scheduler, num_floating_point_operations_so_far,
                        pipeline_rank=None, expert_rank=None, tensor_rank=None, pipeline_parallel=None,
                        expert_parallel=None, train_data_iterator=None):
    start_ckpt = time()
    args = global_vars.get_args()

    import_torch_mindio()
    # Only rank zero of the data parallel writes to the disk.
    model = utils.unwrap_model(model)
    save_dir = args.save

    ckpt_format = 'torch'
    print_rank_0('saving checkpoint at iteration {:7d} to {} in {} format'.format(
        iteration, save_dir, ckpt_format))

    # Collect rng state across data parallel ranks.
    rng_state = get_rng_state(args.ckpt_format)

    # Checkpoint name.
    checkpoint_name = get_checkpoint_name(save_dir, iteration, release=False, pipeline_parallel=pipeline_parallel,
                                          tensor_rank=tensor_rank, pipeline_rank=pipeline_rank,
                                          expert_parallel=expert_parallel, expert_rank=expert_rank,
                                          return_base_dir=args.use_dist_ckpt)

    print_rank_0("Start waiting for all background tasks to be flushed to disks.")
    start = time()
    mindio_acp_flush()
    end = time()
    if torch.distributed.is_initialized():
        torch.distributed.barrier()
    print_rank_0(f'All background tasks have been flushed, cost: {end - start}s')

    # Save distributed optimizer's custom parameter state.
    save_optim_parameter = args.use_distributed_optimizer and not args.no_save_optim and \
                           optimizer is not None and not args.use_dist_ckpt
    if save_optim_parameter:
        optim_checkpoint_name = \
            get_distributed_optimizer_checkpoint_name(checkpoint_name)
        ensure_directory_exists(optim_checkpoint_name)
        save_parameter_state(args.rank, optimizer, optim_checkpoint_name)

    rank = torch.distributed.get_rank() if torch.distributed.is_initialized() else 0

    # Collect args, model, RNG.
    # Save.
    if not torch.distributed.is_initialized() \
            or mpu.get_expert_data_parallel_rank() == 0:
        state_dict = generate_state_dict(
            args,
            model,
            optimizer,
            opt_param_scheduler,
            rng_state,
            iteration=iteration,
            optim_sd_kwargs={},
        )
        state_dict['num_floating_point_operations_so_far'] = num_floating_point_operations_so_far
        end_ckpt = time()
        logging.debug(f"rank: {rank}, takes {end_ckpt - start_ckpt} to prepare state dict for ckpt ")
        ensure_directory_exists(checkpoint_name)
        CheckpointHelper(args.rank).save_model_checkpoint(checkpoint_name, state_dict)

    start_misc = time()

    # And update the latest iteration
    if not torch.distributed.is_initialized() \
            or torch.distributed.get_rank() == 0:
        iteration_dir = os.path.join(args.save, 'iter_{:07d}'.format(iteration))
        pp_size = 1 if args.pipeline_model_parallel_size < 0 else args.pipeline_model_parallel_size
        tp_size = 1 if args.tensor_model_parallel_size < 0 else args.tensor_model_parallel_size
        ep_size = 1 if args.expert_model_parallel_size < 0 else args.expert_model_parallel_size
        if all([args.use_distributed_optimizer, not args.no_save_optim, optimizer is not None, not args.use_dist_ckpt]):
            # 2: both model and optimizer files
            total_file_count = int(pp_size * tp_size * ep_size * 2)
        else:
            total_file_count = int(pp_size * tp_size * ep_size)
        tracker_filename = get_checkpoint_tracker_filename(args.save)
        CheckpointHelper(args.rank).async_write_tracker_file(iteration, iteration_dir, total_file_count,
                                                             tracker_filename)

    # Wait so everyone is done (not necessary)
    if torch.distributed.is_initialized():
        torch.distributed.barrier()

    end_misc = time()
    logging.debug(f"rank: {rank}, takes {end_misc - start_misc} to finalize ckpt save ")


def save_checkpoint_wrapper(fn):
    @wraps(fn)
    def wrapper(iteration, model, optimizer, opt_param_scheduler, num_floating_point_operations_so_far,
                checkpointing_context=None, pipeline_rank=None, expert_rank=None, tensor_rank=None,
                pipeline_parallel=None, expert_parallel=None, non_persistent_ckpt=False,
                train_data_iterator=None, preprocess_common_state_dict_fn=None):
        """Save a model checkpoint.

        Checkpointing context is used to persist some checkpointing state
        throughout a single job. Must be initialized externally (not used if None).
        """
        args = global_vars.get_args()

        if args.use_dist_ckpt or args.async_save or non_persistent_ckpt:
            return fn(iteration=iteration,
                      model=model,
                      optimizer=optimizer,
                      opt_param_scheduler=opt_param_scheduler,
                      num_floating_point_operations_so_far=num_floating_point_operations_so_far,
                      checkpointing_context=checkpointing_context,
                      pipeline_rank=pipeline_rank,
                      expert_rank=expert_rank,
                      tensor_rank=tensor_rank,
                      pipeline_parallel=pipeline_parallel,
                      expert_parallel=expert_parallel,
                      non_persistent_ckpt=non_persistent_ckpt,
                      train_data_iterator=train_data_iterator,
                      preprocess_common_state_dict_fn=preprocess_common_state_dict_fn)

        return acp_save_checkpoint(iteration=iteration,
                                   model=model,
                                   optimizer=optimizer,
                                   opt_param_scheduler=opt_param_scheduler,
                                   num_floating_point_operations_so_far=num_floating_point_operations_so_far,
                                   pipeline_rank=pipeline_rank,
                                   expert_rank=expert_rank,
                                   tensor_rank=tensor_rank,
                                   pipeline_parallel=pipeline_parallel,
                                   expert_parallel=expert_parallel,
                                   train_data_iterator=train_data_iterator)

    return wrapper


def get_optimizer_save_rank(optimizer) -> int:
    args = global_vars.get_args()
    replica_count = 2 if getattr(args, "replica_count", 1) > 1 else 1
    # adaptor with TTP, get origin dp group
    group_gloo = optimizer.data_parallel_group_gloo if not hasattr(optimizer,
                                                                   'ori_dp_group') else optimizer.ori_dp_group
    group_ranks = torch.distributed.get_process_group_ranks(group_gloo)
    if len(group_ranks) == 0:
        raise RuntimeError("optimizer group_ranks can not be empty")
    if len(group_ranks) == 1:
        return group_ranks[0]

    local_rank = torch.distributed.get_rank(group_gloo)
    replica_group_len = len(group_ranks) // replica_count
    replica_group_idx = math.floor(local_rank // replica_group_len)
    selected_local_rank = (replica_group_idx * replica_group_len) % len(group_ranks)
    selected_optim_rank = group_ranks[selected_local_rank]
    return selected_optim_rank


def check_save_ranks(optimizer):
    save_rank = get_optimizer_save_rank(optimizer)
    # adaptor with TTP, group_gloo is new dp group
    group_gloo = optimizer.data_parallel_group_gloo
    group_ranks = torch.distributed.get_process_group_ranks(group_gloo)
    self_rank = global_vars.get_args().rank
    if save_rank not in group_ranks or self_rank not in group_ranks:
        return False
    return True


def default_callback():
    pass


def get_parameter_state(chain_optimizer, notify_callback=None):
    if notify_callback is None:
        notify_callback = default_callback

    is_single = isinstance(chain_optimizer, ChainedOptimizer) and len(chain_optimizer.chained_optimizers) == 1
    single_optimizer = chain_optimizer.chained_optimizers[0] if is_single else chain_optimizer

    if not isinstance(single_optimizer, ChainedOptimizer):
        if not check_save_ranks(single_optimizer):
            notify_callback()
            return None

        d2h_tensors = d2h_optimizer(single_optimizer)
        notify_callback()
        state_dict = gather_optimizer_async(d2h_tensors, single_optimizer)
        if torch.distributed.get_rank(single_optimizer.data_parallel_group) == 0:
            return state_dict
        return None

    d2h_tensors_list = []
    for optimizer in chain_optimizer.chained_optimizers:
        if not check_save_ranks(optimizer):
            continue
        if hasattr(optimizer, 'get_parameter_state_dp_zero'):
            d2h_tensors = d2h_optimizer(optimizer)
            d2h_tensors_list.append(d2h_tensors)

    notify_callback()
    if not d2h_tensors_list:
        return None

    save_states = False
    states = []
    for d2h_tensors, optimizer in zip(d2h_tensors_list, chain_optimizer.chained_optimizers):
        if not check_save_ranks(optimizer):
            continue
        if hasattr(optimizer, 'get_parameter_state_dp_zero'):
            state_dict = gather_optimizer_async(d2h_tensors, optimizer)

            # Save checkpoint economically, only when DP rank = 0, state dict
            # needs to be saved.
            if torch.distributed.get_rank(optimizer.data_parallel_group) == 0:
                states.append(state_dict)
                save_states = True
            else:
                states.append(None)
        else:
            states.append(None)

    if save_states:
        return states
    return None


def save_parameter_state(rank, optimizer, filename: str):
    """Save the distributed parameter state on DP rank 0.

    Args:
        rank          : the rank number
        optimizer     : the optimizer object
        filename (str): path to save parameter state to.
    """
    get_parameter_state_func = partial(get_parameter_state, optimizer)
    CheckpointHelper(rank).save_optimizer_checkpoint(filename, get_parameter_state_func)


@time_used
def d2h_optimizer(optimizer):
    d2h_tensors = defaultdict(lambda: defaultdict(lambda: defaultdict(list)))
    try:
        copy_stream = torch_npu.npu.Stream(device=torch.npu.current_device())
        with torch_npu.npu.stream(copy_stream):
            d2h_optimizer_async(d2h_tensors, optimizer)
        copy_stream.synchronize()
    except RuntimeError as e:
        if 'FORCE STOP' in str(e):
            logging.warning('[torch_mindio] async thread stream with ttp err force stop conflict. Exception is: '
                            '{}'.format(e))
        else:
            raise
    return d2h_tensors


def d2h_optimizer_async(d2h_tensors, optimizer):
    data_parallel_world_size = optimizer.data_parallel_group_gloo.size()
    for gbuf_idx, gbuf_range_maps in enumerate(optimizer.gbuf_ranges):

        # Iterate grad buffers (by data type).
        if len(gbuf_range_maps) != 1:
            raise AssertionError("single dtype supported, for now.")
        for dtype, gbuf_range_map_for_all_buckets in gbuf_range_maps.items():
            # Create coalesced tensors for all state related to parameters in this buffer.
            for bucket_idx, gbuf_range_map in enumerate(gbuf_range_map_for_all_buckets):

                # Compute local DP contiguous shard's size.
                gbuf_world_numel = optimizer.buffers[gbuf_idx].buckets[bucket_idx].grad_data.numel()
                if gbuf_world_numel % data_parallel_world_size != 0:
                    raise AssertionError(f"gbuf_world_numel % data_parallel_world_size should equal 0")

                gbuf_world_numel_unpadded = (
                    optimizer.buffers[gbuf_idx].buckets[bucket_idx].numel_unpadded
                )
                if gbuf_world_numel_unpadded > gbuf_world_numel:
                    raise AssertionError(f"gbuf_world_numel_unpadded should <= gbuf_world_numel")
                key_list = ["param", "exp_avg", "exp_avg_sq"]
                # Build contiguous DP rank shards (for param + optim states).
                for model_param, param_range_map in gbuf_range_map["param_map"].items():
                    tensors = optimizer._get_main_param_and_optimizer_states(model_param)

                    # Copy states into contiguous shard.
                    gbuf_local_start = param_range_map["gbuf_local"].start
                    gbuf_local_end = param_range_map["gbuf_local"].end
                    for key in key_list:
                        d2h_tensors[gbuf_idx][dtype][bucket_idx].append((key,
                                                                         gbuf_local_start,
                                                                         gbuf_local_end,
                                                                         tensors[key].to("cpu", non_blocking=True)))


@time_used
def gather_optimizer_async(d2h_tensors, optimizer):
    data_parallel_world_size = optimizer.data_parallel_group_gloo.size()
    data_parallel_rank = torch.distributed.get_rank(optimizer.data_parallel_group_gloo)
    data_parallel_group_gloo = optimizer.data_parallel_group_gloo
    data_parallel_global_ranks = torch.distributed.get_process_group_ranks(
        optimizer.data_parallel_group_gloo
    )

    state = {"buckets_coalesced": True}
    for gbuf_idx, gbuf_range_maps in enumerate(optimizer.gbuf_ranges):

        # Iterate grad buffers (by data type).
        dtype_state = {}
        if len(gbuf_range_maps) != 1:
            raise AssertionError("single dtype supported, for now.")
        for dtype, gbuf_range_map_for_all_buckets in gbuf_range_maps.items():
            buffer_numel_unpadded = optimizer.buffers[gbuf_idx].numel_unpadded
            # Create coalesced tensors for all state related to parameters in this buffer.
            world_tensors = {}
            if data_parallel_rank == 0:
                world_tensors = {
                    key: torch.zeros(
                        (buffer_numel_unpadded,), dtype=torch.float32, device="cpu"
                    )
                    for key in ("param", "exp_avg", "exp_avg_sq")
                }
                world_tensors["numel_unpadded"] = buffer_numel_unpadded
            offset_in_world_tensors = 0
            for bucket_idx, _ in enumerate(gbuf_range_map_for_all_buckets):

                # Compute local DP contiguous shard's size.
                gbuf_world_numel = optimizer.buffers[gbuf_idx].buckets[bucket_idx].grad_data.numel()
                if gbuf_world_numel % data_parallel_world_size != 0:
                    raise AssertionError(f"gbuf_world_numel % data_parallel_world_size should equal 0")
                gbuf_local_numel = gbuf_world_numel // data_parallel_world_size

                gbuf_world_numel_unpadded = (
                    optimizer.buffers[gbuf_idx].buckets[bucket_idx].numel_unpadded
                )
                if gbuf_world_numel_unpadded > gbuf_world_numel:
                    raise AssertionError(f"gbuf_world_numel_unpadded should <= gbuf_world_numel")
                local_shards = {
                    key: torch.zeros((gbuf_local_numel,), dtype=torch.float32, device="cpu")
                    for key in ("param", "exp_avg", "exp_avg_sq")
                }

                for key, start, end, tensor in d2h_tensors[gbuf_idx][dtype][bucket_idx]:
                    local_shards[key][start:end] = tensor

                # Gather contiguous shards on DP rank 0.
                for key, send_tensor in local_shards.items():

                    # Gather tensor list.
                    if data_parallel_rank == 0:
                        total_size = data_parallel_world_size * gbuf_local_numel
                        recv_single_tensor = torch.zeros(total_size, dtype=torch.float32, device="cpu")
                        recv_tensors = [
                            recv_single_tensor[i:(i + gbuf_local_numel)]
                            for i in range(0, total_size, gbuf_local_numel)
                        ]
                    else:
                        recv_single_tensor = None
                        recv_tensors = None

                    # Gather.
                    torch.distributed.gather(
                        send_tensor,
                        recv_tensors,
                        data_parallel_global_ranks[0],
                        data_parallel_group_gloo,
                    )

                    # Concatenate.
                    if data_parallel_rank == 0:
                        recv_tensors_concatenated = recv_single_tensor
                        # Copy this bucket's collected all-gather tensors into the right place
                        # in the tensor for the buffer. The tensor for the buffer gets rid of
                        # the padding between buckets.
                        start = offset_in_world_tensors
                        end = offset_in_world_tensors + gbuf_world_numel_unpadded
                        try:
                            world_tensors[key][start:end].copy_(
                                recv_tensors_concatenated[:gbuf_world_numel_unpadded]
                            )
                        except KeyError as e:
                            print_rank_0(f"KeyError : {e}")
                            print_rank_0(f"KeyError: The key '{key}' does not exist in world_tensors.")

                offset_in_world_tensors += gbuf_world_numel_unpadded

            # Collect world state.
            dtype_state[dtype] = world_tensors
        state[gbuf_idx] = dtype_state

    return state


def distrib_optimizer_step_wrapper(step):
    @wraps(step)
    def wrapper(*args, **kwargs):
        import_torch_mindio()
        CheckpointHelper(global_vars.get_args().rank).wait_d2h_finished()
        return step(*args, **kwargs)

    return wrapper
