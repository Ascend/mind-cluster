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
import copy

import torch
from megatron.core import mpu, num_microbatches_calculator
from megatron.training import get_args, get_timers
from megatron.core.num_microbatches_calculator import get_num_microbatches
from mindio_ttp.controller_ttp import ttp_logger
from mindio_ttp.adaptor import tft_replica_group, utils

from . import common
from .common import destroy_sub_process_group


def scale_in_rebuild_callback(new_dp_ranks: list, new_world_ranks: list, args, params: str):
    ttp_logger.LOGGER.info(f"scale-in strategy params: {params}, new_dp_ranks: {new_dp_ranks}, new_world_ranks: {new_world_ranks}")
    common.check_scale_in_params(params)
    cur_rank = torch.distributed.get_rank()
    if len(args) <= utils.TRAIN_PARAM or len(args[utils.TRAIN_PARAM]) <= utils.MODEL_INDEX:
        raise RuntimeError(f"args error: {args}")
    models = args[utils.TRAIN_PARAM][utils.MODEL_INDEX]
    arguments = get_args()
    if arguments.expert_model_parallel_size > 1 or arguments.context_parallel_size > 1:
        raise RuntimeError(f"not support ep or cp bigger than 1, but got ep: {arguments.expert_model_parallel_size} "
                           f"cp: {arguments.context_parallel_size} ")
    common.SCALE_IN_WORLD_GROUP = torch.distributed.new_group(new_world_ranks, use_local_synchronization=True)
    ttp_logger.LOGGER.info(f"backend: {arguments.distributed_backend}, rank: {cur_rank}, world_size: "
                           f"{len(new_world_ranks)}, new_world_ranks: {new_world_ranks}")
    old_dp_ranks = torch.distributed.get_process_group_ranks(
        mpu.get_data_parallel_group(with_context_parallel=True))
    dp_cp_replica_group = tft_replica_group.ttp_get_dp_cp_replica_group()
    dp_cp_replica_ranks = torch.distributed.get_process_group_ranks(dp_cp_replica_group)
    if len(dp_cp_replica_ranks) == 0:
        raise RuntimeError(f"dp_cp_replica_ranks is empty")
    common.ORIGIN_DP_SIZE = len(old_dp_ranks)
    common.ORIGIN_NUM_MICRO_BATCHES = get_num_microbatches()
    both_replica_group_fault, changed_old_dp_ranks = get_changed_old_dp_ranks(cur_rank, old_dp_ranks, new_dp_ranks,
                                                                   dp_cp_replica_ranks)
    fault_idxs, fault_local_idxs, fault_first_group = get_fault_msgs(cur_rank, old_dp_ranks, changed_old_dp_ranks,
                                                                     new_dp_ranks,
                                                                   dp_cp_replica_ranks)
    build_scale_in_dp_cp_replica_group(fault_local_idxs, fault_first_group,
                                       both_replica_group_fault, changed_old_dp_ranks)
    rebuild_not_changed_group(cur_rank, both_replica_group_fault, arguments)
    ttp_logger.LOGGER.info(f"rank: {cur_rank} rebuild not changed group done")
    change_model_group(models)
    change_num_micro_batches(old_dp_ranks, new_dp_ranks, arguments)
    common.update_scale_in_flag(True)
    timers = get_timers()
    for _, timer in timers._timers.items():
        timer.set_barrier_group(common.SCALE_IN_WORLD_GROUP)
        timer.reset()
    ttp_logger.LOGGER.info(f"rank:{cur_rank},"
                           f"zit_is_fault_replica_rank:{common.zit_is_fault_replica_rank()},"
                           f"zit_fault_rank_in_dp_cp_replica_group:{common.zit_fault_rank_in_dp_cp_replica_group()},"
                           f"FAULT_REPLICA_RANK:{common.FAULT_REPLICA_RANK}")
    ttp_logger.LOGGER.info(f"rank:{cur_rank} start to build dataset")
    common.build_dataset(args)
    ttp_logger.LOGGER.info(f"rank:{cur_rank} finished build dataset")
    from megatron.core.rerun_state_machine import destroy_rerun_state_machine
    destroy_rerun_state_machine()
    ttp_logger.LOGGER.info(f"rank:{cur_rank} destroy_rerun_state_machine dataset")


def get_changed_old_dp_ranks(cur_rank, old_dp_ranks, new_dp_ranks, dp_cp_replica_ranks):
    """
    Check whether the fault ranks in two replica groups, if it is true,
    use the replica rank of fault rank in second replica group.
    """
    fault_idxs, fault_local_idxs = [], []
    left_replica_group_fault, right_replica_group_fault = False, False
    for idx, rank in enumerate(old_dp_ranks):
        if rank in new_dp_ranks:
            continue
        fault_idxs.append(idx)
        fault_local_idxs.append(idx % len(dp_cp_replica_ranks))
        if idx < len(dp_cp_replica_ranks):
            left_replica_group_fault = True
        else:
            right_replica_group_fault = True
    both_replica_group_fault = left_replica_group_fault and right_replica_group_fault
    ttp_logger.LOGGER.info(f"rank: {cur_rank}, new_dp_ranks: {new_dp_ranks}, fault_idxs: {fault_idxs},"
                           f" fault_local_idxs: {fault_local_idxs}, "
                           f"both_replica_group_fault: {both_replica_group_fault}")
    changed_old_dp_ranks = copy.deepcopy(old_dp_ranks)
    if both_replica_group_fault:
        changed_old_dp_ranks = get_ranks_after_change_left(len(dp_cp_replica_ranks), fault_idxs,
                                                           changed_old_dp_ranks, cur_rank)
    return both_replica_group_fault, changed_old_dp_ranks


def get_ranks_after_change_left(dp_cp_replica_ranks_length, fault_idxs, changed_old_dp_ranks, cur_rank):
    """
    When fault rank list in two replica groups, use the replica rank of fault rank in second replica group.
    For example:
    old dp rank list [0, 8, 16, 24] left replica group [0, 8] right replica group [0, 8]
    fault rank is rank 8 and 16. After changed, dp rank list is [0, 24, 16, 24].
    Then we can perform scale-in replica group reconstruction based on
    the assumption that task failures only occur within the second replica group.
    """
    for idx in fault_idxs:
        if idx >= dp_cp_replica_ranks_length:
            continue
        if idx + dp_cp_replica_ranks_length >= len(changed_old_dp_ranks):
            continue
        changed_old_dp_ranks[idx] = changed_old_dp_ranks[idx + dp_cp_replica_ranks_length]
        if cur_rank == changed_old_dp_ranks[idx + dp_cp_replica_ranks_length]:
            common.IS_FAULT_REPLICA_RANK = True
    return changed_old_dp_ranks


def get_fault_msgs(cur_rank, old_dp_ranks, changed_old_dp_ranks, new_dp_ranks, dp_cp_replica_ranks):
    fault_idxs, fault_local_idxs = [], []
    for idx, rank in enumerate(old_dp_ranks):
        if rank not in new_dp_ranks:
            fault_idxs.append(idx)
            fault_local_idxs.append(idx % len(dp_cp_replica_ranks))
    build_new_dp_cp_group(fault_idxs)
    ttp_logger.LOGGER.info(f"rank: {cur_rank}, new_dp_ranks: {new_dp_ranks}, fault_idxs: {fault_idxs},"
                           f" fault_local_idxs: {fault_local_idxs}, build dp_cp done")
    fault_idxs, fault_local_idxs = [], []
    for idx, rank in enumerate(changed_old_dp_ranks):
        if rank not in new_dp_ranks:
            fault_idxs.append(idx)
            fault_local_idxs.append(idx % len(dp_cp_replica_ranks))
    fault_first_group = False
    for idx, local_idx in zip(fault_idxs, fault_local_idxs):
        if dp_cp_replica_ranks[local_idx] in new_dp_ranks:
            common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP = False
            if cur_rank == dp_cp_replica_ranks[local_idx]:
                common.IS_FAULT_REPLICA_RANK = True
        else:
            common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP = True
        if old_dp_ranks[local_idx] not in new_dp_ranks:
            fault_first_group = True
            common.FAULT_REPLICA_RANK = old_dp_ranks[local_idx + len(dp_cp_replica_ranks)]
        elif old_dp_ranks[idx] not in new_dp_ranks:
            fault_first_group = False
            common.FAULT_REPLICA_RANK = old_dp_ranks[local_idx]
    ttp_logger.LOGGER.info(f"rank: {cur_rank}, new_dp_ranks: {new_dp_ranks}, fault_idxs: {fault_idxs},"
                           f" fault_local_idxs: {fault_local_idxs}, fault_first_group: {fault_first_group}")
    return fault_idxs, fault_local_idxs, fault_first_group


def change_num_micro_batches(old_dp_ranks, new_dp_ranks, arguments):
    old_dp_size = len(old_dp_ranks)
    new_dp_size = len(new_dp_ranks)
    total_num_microbatches = get_num_microbatches() * old_dp_size
    new_num_microbatches = total_num_microbatches // new_dp_size
    common.HAS_DATA = total_num_microbatches % new_dp_size
    if common.HAS_DATA and torch.distributed.get_rank() in new_dp_ranks[:common.HAS_DATA]:
        new_num_microbatches += 1
    ttp_logger.LOGGER.info(f"new num_micro_batches: {new_num_microbatches}, new_dp_size: {new_dp_size},"
                           f"_GLOBAL_NUM_MICROBATCHES_CALCULATOR:"
                           f" {num_microbatches_calculator._GLOBAL_NUM_MICROBATCHES_CALCULATOR}")
    if arguments.rampup_batch_size is not None and len(arguments.rampup_batch_size) == 3:
        new_micro_bsz_times_dp_size = arguments.micro_batch_size * new_dp_size
        num_microbatches_calculator._GLOBAL_NUM_MICROBATCHES_CALCULATOR.data_parallel_size = new_dp_size
        num_microbatches_calculator._GLOBAL_NUM_MICROBATCHES_CALCULATOR.micro_batch_times_data_parallel_size = new_micro_bsz_times_dp_size
    num_microbatches_calculator._GLOBAL_NUM_MICROBATCHES_CALCULATOR.num_micro_batches = new_num_microbatches


def change_model_group(models):
    new_dp_cp_group = mpu.get_data_parallel_group(with_context_parallel=True)
    new_dp_cp_group_size = torch.distributed.get_world_size(group=new_dp_cp_group)
    new_dp_cp_group_rank = torch.distributed.get_rank(group=new_dp_cp_group)
    for model in models:
        for buffer in model.buffers:
            buffer.data_parallel_group = new_dp_cp_group
            buffer.data_parallel_world_size = new_dp_cp_group_size
            for bucket in buffer.buckets:
                bucket.data_parallel_group = new_dp_cp_group
                bucket.data_parallel_world_size = new_dp_cp_group_size
                bucket.data_parallel_rank = new_dp_cp_group_rank
        for _, bucket_group in model.param_to_bucket_group.items():
            bucket_group.intra_distributed_optimizer_instance_group = new_dp_cp_group
            bucket_group.intra_distributed_optimizer_instance_size = new_dp_cp_group_size
            bucket_group.intra_distributed_optimizer_instance_rank = new_dp_cp_group_rank


def build_new_dp_cp_group(fault_idxs):
    reversed_idxs = list(reversed(fault_idxs))
    rank = torch.distributed.get_rank()
    pipeline_model_parallel_size = mpu.get_pipeline_model_parallel_world_size()
    tensor_model_parallel_size = mpu.get_tensor_model_parallel_world_size()
    context_parallel_size = mpu.get_context_parallel_world_size()
    num_pipeline_model_parallel_groups = torch.distributed.get_world_size() // pipeline_model_parallel_size

    for i in range(pipeline_model_parallel_size):
        start_rank = i * num_pipeline_model_parallel_groups
        end_rank = (i + 1) * num_pipeline_model_parallel_groups
        # build new dp group
        for j in range(context_parallel_size * tensor_model_parallel_size):
            dp_ranks = list(range(start_rank + j, end_rank, context_parallel_size * tensor_model_parallel_size))
            dp_ranks = delete_ranks_from_src_by_ids(dp_ranks, reversed_idxs)
            if rank in dp_ranks:
                ttp_logger.LOGGER.info(f"rank:{rank}, dp_ranks:{dp_ranks}")
                group = torch.distributed.new_group(dp_ranks, use_local_synchronization=True)
                group_gloo = torch.distributed.new_group(dp_ranks, backend='gloo', use_local_synchronization=True)
                destroy_sub_process_group(mpu._DATA_PARALLEL_GROUP)
                destroy_sub_process_group(mpu._DATA_PARALLEL_GROUP_GLOO)
                mpu._DATA_PARALLEL_GROUP = group
                mpu._DATA_PARALLEL_GROUP_GLOO = group_gloo
                mpu._DATA_PARALLEL_GLOBAL_RANKS = dp_ranks
                get_args().data_parallel_size = len(dp_ranks)
                destroy_sub_process_group(mpu._DATA_PARALLEL_GROUP_WITH_CP)
                destroy_sub_process_group(mpu._DATA_PARALLEL_GROUP_WITH_CP_GLOO)
                mpu._DATA_PARALLEL_GROUP_WITH_CP = group
                mpu._DATA_PARALLEL_GROUP_WITH_CP_GLOO = group_gloo
                mpu._DATA_PARALLEL_GLOBAL_RANKS_WITH_CP = dp_ranks


def delete_ranks_from_src_by_ids(src_ranks, reversed_idxs):
    for index in reversed_idxs:
        del src_ranks[index]
    return src_ranks


def build_scale_in_dp_cp_replica_group(fault_local_idxs, fault_first_group,
                                       both_replica_group_fault, changed_old_dp_ranks):
    replica_group_size = len(changed_old_dp_ranks) // tft_replica_group.ttp_get_replica_dp_num()
    if not both_replica_group_fault and len(fault_local_idxs) == replica_group_size:
        rank = torch.distributed.get_rank()
        common.IS_FAULT_REPLICA_RANK = False
        ttp_logger.LOGGER.info(f"rank {rank} the full replica group is fault, length of fault_local_idxs"
                               f" is {len(fault_local_idxs)}, common.IS_FAULT_REPLICA_RANK is {common.IS_FAULT_REPLICA_RANK}")
        return
    ranks_left = changed_old_dp_ranks[:replica_group_size]
    ranks_right = changed_old_dp_ranks[replica_group_size:]
    for fault_local_idx in fault_local_idxs:
        ranks_left[fault_local_idx], ranks_right[fault_local_idx] = \
            ranks_right[fault_local_idx], ranks_left[fault_local_idx]
    create_scale_in_replica_group(fault_first_group, ranks_left, ranks_right)
    if both_replica_group_fault:
        create_new_replica_group_for_changed_left(changed_old_dp_ranks[:replica_group_size])


def create_scale_in_replica_group(fault_first_group, ranks_left, ranks_right):
    rank = torch.distributed.get_rank()
    if fault_first_group and rank in ranks_left:
        ttp_logger.LOGGER.info(f"rank:{rank} in ranks_left, replica dp ranks:{ranks_left}")
        group_left = torch.distributed.new_group(ranks_left, use_local_synchronization=True)
        group_left_gloo = torch.distributed.new_group(ranks_left, backend="gloo",
                                                      use_local_synchronization=True)
        common.SCALE_IN_DP_CP_REPLICA_GROUP = group_left
        common.SCALE_IN_DP_CP_REPLICA_GROUP_GLOO = group_left_gloo
        if not common.IS_FAULT_REPLICA_RANK:
            destroy_sub_process_group(tft_replica_group.DP_CP_REPLICA_GROUP)
            destroy_sub_process_group(tft_replica_group.DP_CP_REPLICA_GROUP_GLOO)
            tft_replica_group.DP_CP_REPLICA_GROUP = common.SCALE_IN_DP_CP_REPLICA_GROUP
            tft_replica_group.DP_CP_REPLICA_GROUP_GLOO = common.SCALE_IN_DP_CP_REPLICA_GROUP_GLOO
    elif not fault_first_group and rank in ranks_right:
        ttp_logger.LOGGER.info(f"rank:{rank} in ranks_right, replica dp ranks:{ranks_right}")
        group_right = torch.distributed.new_group(ranks_right, use_local_synchronization=True)
        group_right_gloo = torch.distributed.new_group(ranks_right, backend="gloo",
                                                       use_local_synchronization=True)
        common.SCALE_IN_DP_CP_REPLICA_GROUP = group_right
        common.SCALE_IN_DP_CP_REPLICA_GROUP_GLOO = group_right_gloo
        if not common.IS_FAULT_REPLICA_RANK:
            destroy_sub_process_group(tft_replica_group.DP_CP_REPLICA_GROUP)
            destroy_sub_process_group(tft_replica_group.DP_CP_REPLICA_GROUP_GLOO)
            tft_replica_group.DP_CP_REPLICA_GROUP = common.SCALE_IN_DP_CP_REPLICA_GROUP
            tft_replica_group.DP_CP_REPLICA_GROUP_GLOO = common.SCALE_IN_DP_CP_REPLICA_GROUP_GLOO


def create_new_replica_group_for_changed_left(left_ranks):
    """
    When fault rank list in two replica groups, use the replica rank of fault rank in second replica group in func
    get_ranks_after_change_left. So here we need change FAULT_RANK_IN_DP_CP_REPLICA_GROUP to false.
    """
    rank = torch.distributed.get_rank()
    if rank not in left_ranks:
        return
    if common.IS_FAULT_REPLICA_RANK:
        common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP = False
    group_left = torch.distributed.new_group(left_ranks, use_local_synchronization=True)
    group_left_gloo = torch.distributed.new_group(left_ranks, backend="gloo",
                                                  use_local_synchronization=True)
    destroy_sub_process_group(tft_replica_group.DP_CP_REPLICA_GROUP)
    destroy_sub_process_group(tft_replica_group.DP_CP_REPLICA_GROUP_GLOO)
    tft_replica_group.DP_CP_REPLICA_GROUP = group_left
    tft_replica_group.DP_CP_REPLICA_GROUP_GLOO = group_left_gloo


def rebuild_not_changed_group(cur_rank, both_replica_group_fault, args):
    init_context_parallel_group()
    init_model_parallel_group()
    init_tensor_parallel_group()
    init_pipeline_parallel_group(cur_rank)
    if not both_replica_group_fault and not common.FAULT_RANK_IN_DP_CP_REPLICA_GROUP:
        from .scale_out_rebuild_process_group_callback import ttp_initialize_replica_dp_group
        ttp_initialize_replica_dp_group(args.pipeline_model_parallel_size, args.tensor_model_parallel_size,
                                        args.context_parallel_size, args.world_size)


def init_context_parallel_group():
    if mpu._CONTEXT_PARALLEL_GROUP is not None:
        torch.distributed.reinit_process_group(mpu._CONTEXT_PARALLEL_GROUP, rebuild_link=True)
    else:
        ttp_logger.LOGGER.warning("CONTEXT_PARALLEL_GROUP is None")


def init_model_parallel_group():
    if mpu._MODEL_PARALLEL_GROUP is not None:
        torch.distributed.reinit_process_group(mpu._MODEL_PARALLEL_GROUP, rebuild_link=True)
    else:
        ttp_logger.LOGGER.warning("MODEL_PARALLEL_GROUP is None")


def init_tensor_parallel_group():
    if mpu._TENSOR_MODEL_PARALLEL_GROUP is not None:
        torch.distributed.reinit_process_group(mpu._TENSOR_MODEL_PARALLEL_GROUP, rebuild_link=True)
    else:
        ttp_logger.LOGGER.warning("TENSOR_MODEL_PARALLEL_GROUP is None")


def init_pipeline_parallel_group(rank):
    ranks = mpu._PIPELINE_GLOBAL_RANKS
    if rank in ranks and mpu._PIPELINE_MODEL_PARALLEL_GROUP is not None:
        torch.distributed.reinit_process_group(mpu._PIPELINE_MODEL_PARALLEL_GROUP, rebuild_link=True)
    else:
        ttp_logger.LOGGER.warning(f"rank: {rank} PIPELINE_MODEL_PARALLEL_GROUP is None")
    embedding_ranks = mpu._EMBEDDING_GLOBAL_RANKS
    if embedding_ranks is not None and rank in embedding_ranks and mpu._EMBEDDING_GROUP is not None:
        torch.distributed.reinit_process_group(mpu._EMBEDDING_GROUP, rebuild_link=True)
    else:
        ttp_logger.LOGGER.warning(
            f"rank: {rank} not in ranks {embedding_ranks} or EMBEDDING_GROUP {mpu._EMBEDDING_GROUP} is None")
    position_embedding_ranks = mpu._POSITION_EMBEDDING_GLOBAL_RANKS
    if position_embedding_ranks is not None and rank in position_embedding_ranks \
            and mpu._POSITION_EMBEDDING_GROUP is not None:
        torch.distributed.reinit_process_group(mpu._POSITION_EMBEDDING_GROUP, rebuild_link=True)
    else:
        ttp_logger.LOGGER.warning(
            f"rank: {rank} not in ranks {position_embedding_ranks} or EMBEDDING_GROUP {mpu._POSITION_EMBEDDING_GROUP} is None")