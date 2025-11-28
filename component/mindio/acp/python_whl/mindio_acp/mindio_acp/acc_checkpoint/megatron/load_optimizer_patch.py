#!/usr/bin/env python
# coding=utf-8
# Copyright (c) 2024, NVIDIA CORPORATION. All rights reserved.
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
# Modification description: Patch the Megatron framework load_parameter_state functions using
# MindIO's load methods by HCCL communication for acceleration.
from collections import defaultdict
from typing import Callable, Optional

import torch
from megatron.core.optimizer import ChainedOptimizer
from megatron.core.parallel_state import get_data_modulo_expert_parallel_group

import mindio_acp
from mindio_acp.common import mindio_logger
from mindio_acp.acc_checkpoint.megatron.preload_checkpoint_patch import get_replica_count
from mindio_acp.acc_checkpoint.megatron.parallel_state import InitParallelPolicy, CKPTStage

logging = mindio_logger.LOGGER


def import_torch_npu():
    global torch_npu
    import torch_npu


def load_parameter_state(optimizer, filename: str):
    """Load the distributed parameter state from disk.

    Args:
        optimizer     : the optimizer object
        filename (str): path to load parameter state from.
    """
    import_torch_npu()
    if isinstance(optimizer, ChainedOptimizer) and len(optimizer.chained_optimizers) == 1:
        single_optimizer = optimizer.chained_optimizers[0]
    else:
        single_optimizer = optimizer
    if not isinstance(single_optimizer, ChainedOptimizer):
        policy = _get_load_optimizer_policy(single_optimizer, False)
        state_dict = mindio_acp.load(filename)
        _load_optimizer_parameters(single_optimizer, filename, policy, state_dict)
        return

    states = None
    for i, opt in enumerate(optimizer.chained_optimizers):
        if hasattr(opt, 'load_parameter_state_from_dp_zero'):
            policy = _get_load_optimizer_policy(opt, True)
            if torch.distributed.get_rank() == policy.selected_optim_rank and states is None:
                states = mindio_acp.load(filename)
            state_dict = states[i] if states else None
            _load_optimizer_parameters(opt, filename, policy, state_dict)


def _get_load_optimizer_policy(optimizer, is_chained_optimizers: bool):
    replica_count = get_replica_count()

    dp_ep_group = get_data_modulo_expert_parallel_group()
    dp_ep_ranks = torch.distributed.get_process_group_ranks(dp_ep_group)
    optim_dp_ranks = torch.distributed.get_process_group_ranks(optimizer.data_parallel_group)

    if is_chained_optimizers and optim_dp_ranks != dp_ep_ranks:
        policy = InitParallelPolicy(replica_count, optimizer.data_parallel_group, CKPTStage.LoadDPCP)
    else:
        policy = InitParallelPolicy(replica_count, optimizer.data_parallel_group, CKPTStage.LoadDPEP)
    return policy


def _load_optimizer_parameters(optimizer, filename, policy: InitParallelPolicy, state_dict):
    """Load parameter state (i.e., parameter & optimizer tensors).

    This method performs the reverse of save_parameter_state():
    - Load world buffers from disk (i.e., distrib_opt.pt).
    - Scatter contiguous buffers from DP rank 0 to each DP rank (each DP
      rank receives its relevant subset of the world buffers).
    - For each DP rank, copy param & optimizer shards from contiguous CPU
      buffers. (e.g., one buffer each for main_param, exp_avg, and
      exp_avg_sq).
    """

    def scatter(recv_tensor, send_tensors):
        torch.distributed.scatter(
            recv_tensor,
            send_tensors,
            src=policy.selected_optim_rank,
            group=policy.process_group
        )

    npu_copy_stream = torch_npu.npu.Stream(device=torch.npu.current_device())

    group_gloo = optimizer.data_parallel_group_gloo
    data_parallel_rank = torch.distributed.get_rank(group_gloo)
    if data_parallel_rank == 0:
        # Do nothing if "--fp8-param-gather" is not used.
        optimizer.split_state_dict_if_needed(state_dict)

    h2d_tensors = defaultdict(lambda: defaultdict(lambda: defaultdict(lambda: defaultdict())))
    if torch.distributed.get_rank() == policy.selected_optim_rank:
        with torch_npu.npu.stream(npu_copy_stream):
            _h2d_data_copy(h2d_tensors, optimizer, state_dict)
    npu_copy_stream.synchronize()

    kwargs = {"policy": policy}
    _scatter_tensors(h2d_tensors, optimizer, scatter, **kwargs)
    del h2d_tensors
    torch_npu.npu.empty_cache()


def _h2d_data_copy(h2d_tensors, optimizer, state_dict):
    for gbuf_idx, gbuf_range_maps in enumerate(optimizer.gbuf_ranges):
        for dtype, gbuf_range_map_for_all_buckets in gbuf_range_maps.items():
            buffer_numel_unpadded = optimizer.buffers[gbuf_idx].numel_unpadded
            checkpoint_numel_unpadded = state_dict[gbuf_idx][dtype]["numel_unpadded"]
            if buffer_numel_unpadded != checkpoint_numel_unpadded:
                raise ValueError(
                    f"Number of unpadded elements must be same in current run "
                    f"({buffer_numel_unpadded}) and checkpoint ({checkpoint_numel_unpadded})"
                )
            for key in ("param", "exp_avg", "exp_avg_sq"):
                offset_in_world_tensors = 0
                kwargs = {
                    "gbuf_idx": gbuf_idx, "dtype": dtype,
                    "gbuf_range_map_for_all_buckets": gbuf_range_map_for_all_buckets, "key": key
                }
                _h2d_data_copy_each_bucket(h2d_tensors, optimizer, state_dict,
                                           offset_in_world_tensors, **kwargs)


def _h2d_data_copy_each_bucket(h2d_tensors, optimizer, state_dict, offset_in_world_tensors, **kwargs):
    gbuf_idx = kwargs.get("gbuf_idx")
    key = kwargs.get("key")
    dtype = kwargs.get("dtype")
    gbuf_range_map_for_all_buckets = kwargs.get("gbuf_range_map_for_all_buckets")
    data_parallel_world_size = optimizer.data_parallel_group_gloo.size()
    for bucket_idx, _ in enumerate(gbuf_range_map_for_all_buckets):
        # Compute local DP contiguous shard's size.
        gbuf_world_numel = (
            optimizer.buffers[gbuf_idx].buckets[bucket_idx].grad_data.numel()
        )
        gbuf_world_numel_unpadded = (
            optimizer.buffers[gbuf_idx].buckets[bucket_idx].numel_unpadded
        )

        world_tensors = state_dict[gbuf_idx][dtype][key]
        start = offset_in_world_tensors
        end = offset_in_world_tensors + gbuf_world_numel_unpadded
        if not (0 <= start < end <= world_tensors.numel()):
            raise ValueError("start and end values are out of the valid range.")
        world_tensor = world_tensors[start:end].to('npu')
        offset_in_world_tensors += gbuf_world_numel_unpadded

        # Pad world_tensor to gbuf_world_numel. Don't pad at the front, pad at the back.
        world_tensor = torch.nn.functional.pad(
            world_tensor, (0, gbuf_world_numel - gbuf_world_numel_unpadded)
        )
        if world_tensor.numel() != gbuf_world_numel:
            raise ValueError("world_tensor numel is invalid.")

        h2d_tensors[gbuf_idx][dtype][bucket_idx][key] = world_tensor


def _scatter_tensors(h2d_tensors, optimizer, scatter: Callable[[torch.Tensor, Optional[torch.Tensor]], None],
                     **kwargs):
    for gbuf_idx, gbuf_range_maps in enumerate(optimizer.gbuf_ranges):
        for dtype, gbuf_range_map_for_all_buckets in gbuf_range_maps.items():
            recv_tensors = {}
            for key in ("param", "exp_avg", "exp_avg_sq"):
                kwargs["gbuf_idx"] = gbuf_idx

                kwargs["dtype"] = dtype
                kwargs["gbuf_range_map_for_all_buckets"] = gbuf_range_map_for_all_buckets
                kwargs["key"] = key
                kwargs["recv_tensors"] = recv_tensors
                _scatter_tensors_each_bucket(h2d_tensors, optimizer, scatter, **kwargs)
            for model_param, tensors in recv_tensors.items():
                optimizer._set_main_param_and_optimizer_states(model_param, tensors)


def _scatter_tensors_each_bucket(h2d_tensors, optimizer,
                                 scatter: Callable[[torch.Tensor, Optional[torch.Tensor]], None], **kwargs):
    policy = kwargs.get("policy")
    load_optimizer_rank = policy.selected_optim_rank
    data_parallel_world_size = optimizer.data_parallel_group_gloo.size()
    gbuf_idx = kwargs.get("gbuf_idx")
    key = kwargs.get("key")
    dtype = kwargs.get("dtype")
    recv_tensors = kwargs.get("recv_tensors")
    gbuf_range_map_for_all_buckets = kwargs.get("gbuf_range_map_for_all_buckets")
    partition_group_len = len(policy.group_ranks)
    partition_group_idx = torch.distributed.get_group_rank(optimizer.data_parallel_group,
                                                           policy.selected_optim_rank) / partition_group_len

    for bucket_idx, gbuf_range_map in enumerate(gbuf_range_map_for_all_buckets):
        # Compute local DP contiguous shard's size.
        gbuf_world_numel = (
            optimizer.buffers[gbuf_idx].buckets[bucket_idx].grad_data.numel()
        )
        if gbuf_world_numel % data_parallel_world_size != 0:
            raise ValueError("gbuf_world_numel is invalid.")
        gbuf_local_numel = gbuf_world_numel // data_parallel_world_size
        gbuf_world_numel_unpadded = (
            optimizer.buffers[gbuf_idx].buckets[bucket_idx].numel_unpadded
        )
        if not gbuf_world_numel_unpadded <= gbuf_world_numel:
            raise ValueError("gbuf_world_numel_unpadded is invalid.")

        # Contiguous local shards (received from DP rank 0).
        recv_tensor = torch.zeros((gbuf_local_numel,), dtype=torch.float32, device="npu")
        if torch.distributed.get_rank() == load_optimizer_rank:
            start_numel = int(gbuf_local_numel * partition_group_idx * partition_group_len)
            end_numel = int(start_numel + partition_group_len * gbuf_local_numel)
            gbuf_start_idxs = list(range(start_numel, end_numel, gbuf_local_numel))
            world_tensor = h2d_tensors[gbuf_idx][dtype][bucket_idx][key]
            send_tensors = [world_tensor[i:(i + gbuf_local_numel)] for i in gbuf_start_idxs]
        else:
            send_tensors = None

        # Scatter.
        scatter(recv_tensor, send_tensors)

        # Copy local contiguous shards to param/optim shards.
        for model_param, param_range_map in gbuf_range_map["param_map"].items():
            # Copy states into contiguous shard.
            gbuf_local_start = param_range_map["gbuf_local"].start
            gbuf_local_end = param_range_map["gbuf_local"].end
            if model_param not in recv_tensors:
                recv_tensors[model_param] = {}
            recv_tensors[model_param][key] = recv_tensor[gbuf_local_start:gbuf_local_end]
    torch_npu.npu.empty_cache()
