#!/usr/bin/env python3
# coding: utf-8
# Copyright (c) 2020, NVIDIA CORPORATION. All rights reserved.
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

# Modification Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# Modification description: Modify the logic of reestablishing process groups for MindIo.

import os
import time
import warnings
from datetime import timedelta

import torch
from megatron.core import mpu
from megatron.training import get_args, get_timers
from mindio_ttp.adaptor.utils import TRAIN_PARAM, MODEL_INDEX, OPTIM_INDEX
from mindio_ttp.controller_ttp import ttp_logger
from mindio_ttp.framework_ttp.ttp_decorator import get_mindio_export_version
from mindio_ttp.adaptor import tft_arf_group_repair, tft_replica_group

from . import common
from .common import destroy_sub_process_group


def scale_out_rebuild_process_group_callback(fault_ranks: list, train_args, params: str):
    ttp_logger.LOGGER.info(f"scale out strategy params: {params}, fault_ranks: {fault_ranks}")
    ttp_logger.LOGGER.debug(f"scale out train args: {train_args}")
    common.check_scale_out_params(params)
    t1 = time.time()
    if (len(train_args) <= TRAIN_PARAM or len(train_args[TRAIN_PARAM]) <= MODEL_INDEX or
            len(train_args[TRAIN_PARAM]) <= OPTIM_INDEX):
        raise RuntimeError(f"train_args error: {train_args}")
    models, optimizer = train_args[TRAIN_PARAM][MODEL_INDEX], train_args[TRAIN_PARAM][OPTIM_INDEX]
    args = get_args()
    timeout = timedelta(minutes=args.distributed_timeout_minutes)
    nccl_comm_cfgs = {}
    if args.nccl_communicator_config_path is not None:
        try:
            import yaml
        except ImportError as e:
            ttp_logger.LOGGER.error(f"import module yaml failed: {e}")
            raise e
        with open(args.nccl_communicator_config_path, 'r') as stream:
            nccl_comm_cfgs = yaml.safe_load(stream)
    if common.ORIGIN_DP_SIZE is not None:
        args.data_parallel_size = common.ORIGIN_DP_SIZE
        ttp_logger.LOGGER.info(f'rank:{args.rank} new DP size:{args.data_parallel_size}')
    timers = get_timers()
    timers('interval-time').reset()
    timers('interval-time', log_level=0).start(barrier=False)
    os.environ['TORCH_DIST_INIT_BARRIER'] = '1'
    tft_arf_group_repair.update_arf_reboot_flag(False)
    common.update_scale_in_flag(False)
    rebuild_process_group(args, timeout, nccl_comm_cfgs)
    update_model_and_optim_related_group(models, optimizer)
    os.environ['TORCH_DIST_INIT_BARRIER'] = '0'
    ttp_logger.LOGGER.info(f"[rebuild] rank:{args.rank}, rebuild total time consumed:{time.time() - t1:.3f}s")


def rebuild_process_group(args, timeout, nccl_comm_cfgs):
    ttp_logger.LOGGER.info(f"1.1 rank:{args.rank} start rebuild all process group")
    destroy_all_process_group()
    ttp_logger.LOGGER.info(f"1.1 rank:{args.rank} destroy_all_process_group done")
    init_all_process_group(args)
    ttp_logger.LOGGER.info(f"1.2 rank:{args.rank} init_all_process_group done")
    init_data_parallel_group(args, timeout, nccl_comm_cfgs)
    ttp_logger.LOGGER.info(f"1.3 rank:{args.rank} rebuild data parallel group done")
    all_dp_ranks_with_cp = init_data_parallel_with_cp_group(args, timeout, nccl_comm_cfgs)
    ttp_logger.LOGGER.info(f"1.4 rank:{args.rank} rebuild data parallel group with cp done")
    init_context_parallel_group(args, timeout, nccl_comm_cfgs)
    ttp_logger.LOGGER.info(f"1.5 rank:{args.rank} rebuild context parallel group done")
    init_model_parallel_group(args, timeout, nccl_comm_cfgs, all_dp_ranks_with_cp)
    ttp_logger.LOGGER.info(f"1.6 rank:{args.rank} rebuild model parallel group done")
    init_tensor_parallel_group(args, timeout, nccl_comm_cfgs)
    ttp_logger.LOGGER.info(f"1.7 rank:{args.rank} rebuild tensor parallel group done")
    init_pipeline_parallel_group(args, timeout, nccl_comm_cfgs)
    ttp_logger.LOGGER.info(f"1.8 rank:{args.rank} rebuild pipeline parallel group done")
    if common.SCALE_IN_WORLD_GROUP is not None:
        destroy_sub_process_group(common.SCALE_IN_WORLD_GROUP)
        common.SCALE_IN_WORLD_GROUP = None
        ttp_logger.LOGGER.info(f"1.9 rank:{args.rank} destroy scale in world group done")
    ttp_initialize_replica_dp_group(args.pipeline_model_parallel_size, args.tensor_model_parallel_size,
                                    args.context_parallel_size, args.world_size)
    destroy_sub_process_group(common.SCALE_IN_DP_CP_REPLICA_GROUP)
    destroy_sub_process_group(common.SCALE_IN_DP_CP_REPLICA_GROUP_GLOO)
    common.SCALE_IN_DP_CP_REPLICA_GROUP = None
    common.SCALE_IN_DP_CP_REPLICA_GROUP_GLOO = None
    ttp_logger.LOGGER.info(f"1.10 rank:{args.rank} destroy scale in replica group done")
    # build other group for gitee MindSpeed or MindSpeed-LLM
    if get_mindio_export_version() in ["MindSpeed", "MindSpeed-LLM"]:
        build_other_group(nccl_comm_cfgs)


def destroy_all_process_group(group=None):
    from torch.distributed.distributed_c10d import GroupMember, _world
    if group == GroupMember.NON_GROUP_MEMBER:
        return
    pg = GroupMember.WORLD if group is None else group
    if pg is None:
        raise RuntimeError("Process group must not be None")
    if _world.pg_map.get(pg, None) is None:
        raise RuntimeError("Invalid process group specified")
    if pg.name().lower() == "nccl" and pg._has_hooks():
        pg._wait_for_pending_works()
    if group is None or group == GroupMember.WORLD:
        _world.default_pg = None
    del _world.pg_map[pg]
    del _world.pg_names[pg]
    del _world.pg_group_ranks[pg]
    del _world.pg_backend_config[pg]
    if hasattr(_world, 'pg_default_device') and pg in _world.pg_default_device:
        del _world.pg_default_device[pg]
    if pg in _world.pg_coalesce_state.keys():
        warnings.warn("Some coalesced collectives haven't been launched when"
                      " ProcessGroup is destroyed. They will be cleaned.")
        del _world.pg_coalesce_state[pg]
    tag = _world.pg_to_tag.get(pg)
    del _world.pg_to_tag[pg]
    if tag is not None:
        try:
            _world.tags_to_pg[tag].remove(pg)
            if tag.startswith("ptd:"):
                _world.tags_to_pg[""].remove(pg)
        except Exception:
            ttp_logger.LOGGER.warning(f"Failed to remove process group {pg} from _world.tags_to_pg")


def ttp_initialize_replica_dp_group(pipeline_model_parallel_size, tensor_model_parallel_size, context_parallel_size,
                                    world_size):
    if pipeline_model_parallel_size == 0 or tensor_model_parallel_size == 0 or context_parallel_size == 0:
        raise ValueError("pipeline_model_parallel_size, tensor_model_parallel_size, context_parallel_size "
                         "should not be zero")
    num_pipeline_model_parallel_groups: int = world_size // pipeline_model_parallel_size

    args = get_args()
    cur_rank = torch.distributed.get_rank()
    temp_replica_num = getattr(args, 'optimizer_replica_num', tft_replica_group.REPLICA_NUM)
    if temp_replica_num != 0 and temp_replica_num != tft_replica_group.REPLICA_NUM:
        tft_replica_group.REPLICA_NUM = temp_replica_num
    for i in range(pipeline_model_parallel_size):
        start_rank = i * num_pipeline_model_parallel_groups
        end_rank = (i + 1) * num_pipeline_model_parallel_groups
        for j in range(tensor_model_parallel_size):
            dp_cp_ranks = list(range(start_rank + j, end_rank, tensor_model_parallel_size))
            if cur_rank in dp_cp_ranks:
                tft_replica_group.DP_CP_ORIGIN_RANKS = dp_cp_ranks
            if args.use_distributed_optimizer:
                build_dp_cp_replica_group(dp_cp_ranks, cur_rank)


def build_dp_cp_replica_group(dp_cp_ranks: list, cur_rank):
    if len(dp_cp_ranks) % tft_replica_group.REPLICA_NUM != 0:
        raise ValueError(f"size of dp_cp_ranks {len(dp_cp_ranks)} should be a multiple of replica"
                         f" num {tft_replica_group.REPLICA_NUM}")
    replica_group_size = len(dp_cp_ranks) // tft_replica_group.REPLICA_NUM
    replica_lists = [dp_cp_ranks[i*replica_group_size:(i+1) * replica_group_size]
                     for i in range(0, tft_replica_group.REPLICA_NUM)]
    for replica_list in replica_lists:
        if cur_rank in replica_list:
            replica_group = torch.distributed.new_group(replica_list, use_local_synchronization=True)
            replica_group_gloo = torch.distributed.new_group(replica_list, backend="gloo",
                                                             use_local_synchronization=True)
            destroy_sub_process_group(tft_replica_group.DP_CP_REPLICA_GROUP)
            destroy_sub_process_group(tft_replica_group.DP_CP_REPLICA_GROUP_GLOO)
            tft_replica_group.DP_CP_REPLICA_GROUP = replica_group
            tft_replica_group.DP_CP_REPLICA_GROUP_GLOO = replica_group_gloo
            return


def init_all_process_group(args):
    # call the init process
    torch.distributed.init_process_group(
        backend=args.distributed_backend,
        world_size=args.world_size,
        rank=args.rank,
        timeout=timedelta(minutes=args.distributed_timeout_minutes),
    )


def get_nccl_options(pg_name, nccl_comm_cfgs):
    if pg_name in nccl_comm_cfgs:
        nccl_options = torch.distributed.ProcessGroupNCCL.Options()
        nccl_options.config.cga_cluster_size = nccl_comm_cfgs[pg_name].get('cga_cluster_size', 4)
        nccl_options.config.max_ctas = nccl_comm_cfgs[pg_name].get('max_ctas', 32)
        nccl_options.config.min_ctas = nccl_comm_cfgs[pg_name].get('min_ctas', 1)
        return nccl_options
    return None


def init_data_parallel_group(args, timeout, nccl_comm_cfgs):
    rank = torch.distributed.get_rank()
    world_size = torch.distributed.get_world_size()
    tensor_model_parallel_size = args.tensor_model_parallel_size
    pipeline_model_parallel_size = args.pipeline_model_parallel_size
    context_parallel_size = args.context_parallel_size
    num_pipeline_model_parallel_size = world_size // pipeline_model_parallel_size

    all_data_parallel_group_ranks = []
    for i in range(pipeline_model_parallel_size):
        start_rank = i * num_pipeline_model_parallel_size
        end_rank = (i + 1) * num_pipeline_model_parallel_size
        for j in range(context_parallel_size * tensor_model_parallel_size):
            ranks = range(start_rank + j, end_rank, context_parallel_size * tensor_model_parallel_size)
            all_data_parallel_group_ranks.append(list(ranks))
            if rank in ranks:
                group = torch.distributed.new_group(ranks, timeout=timeout,
                                                    pg_options=get_nccl_options('dp', nccl_comm_cfgs),
                                                    use_local_synchronization=True)
                group_gloo = torch.distributed.new_group(ranks, timeout=timeout, backend='gloo',
                                                         use_local_synchronization=True)
                destroy_sub_process_group(mpu._DATA_PARALLEL_GROUP)
                destroy_sub_process_group(mpu._DATA_PARALLEL_GROUP_GLOO)
                mpu._DATA_PARALLEL_GROUP = group
                mpu._DATA_PARALLEL_GROUP_GLOO = group_gloo
                mpu._DATA_PARALLEL_GLOBAL_RANKS = ranks
    return all_data_parallel_group_ranks


def init_data_parallel_with_cp_group(args, timeout, nccl_comm_cfgs):
    rank = torch.distributed.get_rank()
    world_size = torch.distributed.get_world_size()
    tensor_model_parallel_size = args.tensor_model_parallel_size
    pipeline_model_parallel_size = args.pipeline_model_parallel_size
    num_pipeline_model_parallel_groups = world_size // pipeline_model_parallel_size

    all_data_parallel_group_ranks_with_cp = []
    for i in range(pipeline_model_parallel_size):
        start_rank = i * num_pipeline_model_parallel_groups
        end_rank = (i + 1) * num_pipeline_model_parallel_groups
        for j in range(tensor_model_parallel_size):
            ranks_with_cp = range(start_rank + j, end_rank, tensor_model_parallel_size)
            all_data_parallel_group_ranks_with_cp.append(list(ranks_with_cp))
            if rank in ranks_with_cp:
                group_with_cp = torch.distributed.new_group(ranks_with_cp, timeout=timeout,
                                                    pg_options=get_nccl_options('dp_cp', nccl_comm_cfgs),
                                                    use_local_synchronization=True)
                group_with_cp_gloo = torch.distributed.new_group(ranks_with_cp, timeout=timeout,
                                                            backend='gloo', use_local_synchronization=True)
                destroy_sub_process_group(mpu._DATA_PARALLEL_GROUP_WITH_CP)
                destroy_sub_process_group(mpu._DATA_PARALLEL_GROUP_WITH_CP_GLOO)
                mpu._DATA_PARALLEL_GROUP_WITH_CP = group_with_cp
                mpu._DATA_PARALLEL_GROUP_WITH_CP_GLOO = group_with_cp_gloo
                mpu._DATA_PARALLEL_GLOBAL_RANKS_WITH_CP = ranks_with_cp
    return all_data_parallel_group_ranks_with_cp


def init_context_parallel_group(args, timeout, nccl_comm_cfgs):

    world_size = torch.distributed.get_world_size()
    tensor_model_parallel_size = args.tensor_model_parallel_size
    pipeline_model_parallel_size = args.pipeline_model_parallel_size
    context_parallel_size = args.context_parallel_size
    num_pipeline_model_parallel_groups = world_size // pipeline_model_parallel_size
    data_parallel_size = (world_size //
                          (pipeline_model_parallel_size * tensor_model_parallel_size * context_parallel_size))
    for i in range(pipeline_model_parallel_size):
        for j in range(data_parallel_size):
            start_rank = (i * num_pipeline_model_parallel_groups +
                          j * tensor_model_parallel_size * context_parallel_size)
            end_rank = (i * num_pipeline_model_parallel_groups +
                        (j + 1) * tensor_model_parallel_size * context_parallel_size)
            if create_context_group(args, start_rank, end_rank, timeout, nccl_comm_cfgs):
                return


def create_context_group(args, start_rank, end_rank, timeout, nccl_comm_cfgs):
    rank = torch.distributed.get_rank()
    tensor_model_parallel_size = args.tensor_model_parallel_size
    for k in range(tensor_model_parallel_size):
        ranks = range(start_rank + k, end_rank, tensor_model_parallel_size)
        if rank in ranks:
            group = torch.distributed.new_group(ranks, timeout=timeout,
                                                pg_options=get_nccl_options('cp', nccl_comm_cfgs),
                                                use_local_synchronization=True)
            destroy_sub_process_group(mpu._CONTEXT_PARALLEL_GROUP)
            mpu._CONTEXT_PARALLEL_GROUP = group
            mpu._CONTEXT_PARALLEL_GLOBAL_RANKS = ranks
            return True
    return False


def init_model_parallel_group(args, timeout, nccl_comm_cfgs, all_dp_ranks_with_cp):
    rank = torch.distributed.get_rank()
    world_size = torch.distributed.get_world_size()
    tensor_model_parallel_size = args.tensor_model_parallel_size
    pipeline_model_parallel_size = args.pipeline_model_parallel_size
    context_parallel_size = args.context_parallel_size
    data_parallel_size = (world_size //
                          (pipeline_model_parallel_size * tensor_model_parallel_size * context_parallel_size))
    for i in range(data_parallel_size * context_parallel_size):
        ranks = [data_parallel_group_ranks_with_cp[i] for data_parallel_group_ranks_with_cp in all_dp_ranks_with_cp]
        if rank in ranks:
            group = torch.distributed.new_group(ranks, timeout=timeout,
                                                    pg_options=get_nccl_options('mp', nccl_comm_cfgs),
                                                    use_local_synchronization=True)
            destroy_sub_process_group(mpu._MODEL_PARALLEL_GROUP)
            mpu._MODEL_PARALLEL_GROUP = group
            return


def init_tensor_parallel_group(args, timeout, nccl_comm_cfgs):
    rank = torch.distributed.get_rank()
    world_size = torch.distributed.get_world_size()
    tensor_model_parallel_size = args.tensor_model_parallel_size
    num_tensor_model_parallel_groups = world_size // tensor_model_parallel_size
    for i in range(num_tensor_model_parallel_groups):
        ranks = range(i * tensor_model_parallel_size, (i + 1) * tensor_model_parallel_size)
        if rank in ranks:
            group = torch.distributed.new_group(ranks, timeout=timeout,
                                                pg_options=get_nccl_options('tp', nccl_comm_cfgs),
                                                use_local_synchronization=True)
            destroy_sub_process_group(mpu._TENSOR_MODEL_PARALLEL_GROUP)
            mpu._TENSOR_MODEL_PARALLEL_GROUP = group
            return


def init_pipeline_parallel_group(args, timeout, nccl_comm_cfgs):
    rank = torch.distributed.get_rank()
    world_size = torch.distributed.get_world_size()
    pipeline_model_parallel_size = args.pipeline_model_parallel_size
    num_pipeline_model_parallel_groups = world_size // pipeline_model_parallel_size
    for i in range(num_pipeline_model_parallel_groups):
        ranks = range(i, world_size, num_pipeline_model_parallel_groups)
        if rank in ranks:
            group = torch.distributed.new_group(ranks, timeout=timeout,
                                                pg_options=get_nccl_options('pp', nccl_comm_cfgs),
                                                use_local_synchronization=True)
            destroy_sub_process_group(mpu._PIPELINE_MODEL_PARALLEL_GROUP)
            mpu._PIPELINE_MODEL_PARALLEL_GROUP = group
            mpu._PIPELINE_GLOBAL_RANKS = ranks
        if len(ranks) > 1:
            embedding_ranks = [ranks[0], ranks[-1]]
            position_embedding_ranks = [ranks[0]]
        else:
            embedding_ranks = ranks
            position_embedding_ranks = ranks
        if rank in embedding_ranks:
            group = torch.distributed.new_group(embedding_ranks, timeout=timeout,
                                                pg_options=get_nccl_options('embd', nccl_comm_cfgs),
                                                use_local_synchronization=True)
            destroy_sub_process_group(mpu._EMBEDDING_GROUP)
            mpu._EMBEDDING_GROUP = group
        if rank in position_embedding_ranks:
            group = torch.distributed.new_group(position_embedding_ranks, timeout=timeout,
                                                pg_options=get_nccl_options('embd', nccl_comm_cfgs),
                                                use_local_synchronization=True)
            destroy_sub_process_group(mpu._POSITION_EMBEDDING_GROUP)
            mpu._POSITION_EMBEDDING_GROUP = group
        if rank in ranks:
            mpu._EMBEDDING_GLOBAL_RANKS = embedding_ranks
            mpu._POSITION_EMBEDDING_GLOBAL_RANKS = position_embedding_ranks


def build_other_group(nccl_comm_cfgs):
    args = get_args()
    # rebuild groups in MindSpeed-LLM
    from mindspeed.core import parallel_state as mindspeed_mpu
    if hasattr(mindspeed_mpu, 'initialize_context_parallel_group_for_send_recv_overlap'):
        ttp_logger.LOGGER.info(f"rank:{args.rank} initialize context parallel group for send recv overlap")
        mindspeed_mpu.initialize_context_parallel_group_for_send_recv_overlap(args.tensor_model_parallel_size,
                                                                              args.pipeline_model_parallel_size,
                                                                              args.context_parallel_size,
                                                                              nccl_comm_cfgs)

    if hasattr(mindspeed_mpu, 'initialize_context_parallel_group_for_hybrid_cp'):
        ttp_logger.LOGGER.info(f'rank:{args.rank} initialize context parallel group for hybrid cp')
        mindspeed_mpu.initialize_context_parallel_group_for_hybrid_cp(args.tensor_model_parallel_size,
                                                                      args.pipeline_model_parallel_size,
                                                                      args.context_parallel_size,
                                                                      nccl_comm_cfgs)

    if hasattr(mindspeed_mpu, 'initialize_context_parallel_group_for_double_ring'):
        ttp_logger.LOGGER.info(f'rank:{args.rank} initialize context parallel group for double ring')
        mindspeed_mpu.initialize_context_parallel_group_for_double_ring(args.tensor_model_parallel_size,
                                                                      args.pipeline_model_parallel_size,
                                                                      args.context_parallel_size,
                                                                      nccl_comm_cfgs)

    if mindspeed_mpu._PIPELINE_MODEL_PARALLEL_GROUP_FOR_NEW_STREAM is not None:
        ttp_logger.LOGGER.info(f'rank:{args.rank} initialize pipeline model parallel group for new stream')
        initialize_context_parallel_group_for_hybrid_cp(args, nccl_comm_cfgs)

    use_nd_matmul_str = 'use_nd_matmul'
    if (getattr(args, use_nd_matmul_str, False) or args.tp_2d) and hasattr(mindspeed_mpu,
                                                                         'initialize_ndmm_parallel_group'):
        ttp_logger.LOGGER.info(f'rank:{args.rank} initialize ndmm parallel group')
        nd1_dim1_sz = args.nd1_dim1_size if getattr(args, use_nd_matmul_str, False) else args.tp_x
        nd2_dim1_sz = args.nd2_dim1_size if getattr(args, use_nd_matmul_str, False) else args.tp_y
        mindspeed_mpu.initialize_ndmm_parallel_group(nccl_comm_cfgs,
                                                    tensor_model_parallel_size=args.tensor_model_parallel_size,
                                                    nd1_dim1_size=nd1_dim1_sz,
                                                    nd2_dim1_size=nd2_dim1_sz)


def initialize_context_parallel_group_for_hybrid_cp(args, nccl_comm_cfgs):
    from mindspeed.core import parallel_state as mindspeed_mpu
    import megatron
    rank = torch.distributed.get_rank()
    world_size: int = torch.distributed.get_world_size()
    num_pipeline_model_parallel_groups: int = world_size // args.pipeline_model_parallel_size
    for i in range(num_pipeline_model_parallel_groups):
        ranks = range(i, world_size, num_pipeline_model_parallel_groups)
        if rank in ranks:
            group = torch.distributed.new_group(ranks, pg_options=megatron.core.parallel_state.get_nccl_options(
                'pp_new_stream', nccl_comm_cfgs))
            destroy_sub_process_group(mindspeed_mpu._PIPELINE_MODEL_PARALLEL_GROUP_FOR_NEW_STREAM)
            mindspeed_mpu._PIPELINE_MODEL_PARALLEL_GROUP_FOR_NEW_STREAM = group
            return


def update_model_and_optim_related_group(models, optimizer):
    if not get_args().use_distributed_optimizer:
        return

    # fix optimizer attributes
    if hasattr(optimizer, 'optim_nums') and optimizer.optim_nums > 1:
        optimizer.chained_optimizers[0].ori_dp_group = mpu._DATA_PARALLEL_GROUP
        optimizer.chained_optimizers[0].data_parallel_group = tft_replica_group.ttp_get_dp_cp_replica_group()
        optimizer.chained_optimizers[0].data_parallel_group_gloo = tft_replica_group.ttp_get_dp_cp_replica_group_gloo()
        optimizer.chained_optimizers[0].ori_dp_list = torch.distributed.get_process_group_ranks(
            mpu._DATA_PARALLEL_GROUP)
        optimizer.chained_optimizers[1].data_parallel_group = tft_replica_group.ttp_get_dp_ep_replica_group()
        optimizer.chained_optimizers[1].data_parallel_group_gloo = tft_replica_group.ttp_get_dp_ep_replica_group_gloo()
        optimizer.chained_optimizers[1].ori_dp_group = mpu._DATA_MODULO_EXPERT_PARALLEL_GROUP
        optimizer.chained_optimizers[1].ori_dp_list = torch.distributed.get_process_group_ranks(
            mpu._DATA_MODULO_EXPERT_PARALLEL_GROUP)
    else:
        optimizer.data_parallel_group = tft_replica_group.ttp_get_dp_cp_replica_group()
        optimizer.data_parallel_group_gloo = tft_replica_group.ttp_get_dp_cp_replica_group_gloo()
        optimizer.ori_dp_group = mpu._DATA_PARALLEL_GROUP
        optimizer.ori_dp_list = torch.distributed.get_process_group_ranks(mpu._DATA_PARALLEL_GROUP)
    for model in models:
        for buffer in model.buffers:
            buffer.data_parallel_group = mpu._DATA_PARALLEL_GROUP
            buffer.data_parallel_world_size = torch.distributed.get_world_size(group=mpu._DATA_PARALLEL_GROUP)
            for bucket in buffer.buckets:
                bucket.data_parallel_group = mpu._DATA_PARALLEL_GROUP
                bucket.data_parallel_world_size = torch.distributed.get_world_size(group=mpu._DATA_PARALLEL_GROUP)
                bucket.data_parallel_rank = torch.distributed.get_rank(group=mpu._DATA_PARALLEL_GROUP)
        for _, bucket_group in model.param_to_bucket_group.items():
            bucket_group.intra_distributed_optimizer_instance_group = mpu._DATA_PARALLEL_GROUP
            bucket_group.intra_distributed_optimizer_instance_size = torch.distributed.get_world_size(
                group=mpu._DATA_PARALLEL_GROUP)
            bucket_group.intra_distributed_optimizer_instance_rank = torch.distributed.get_rank(
                group=mpu._DATA_PARALLEL_GROUP)








    