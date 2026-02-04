#!/usr/bin/env python
# coding=utf-8
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
import os
from functools import wraps

from megatron.training import global_vars

from mindio_acp.common import mindio_logger
from mindio_acp.acc_checkpoint.megatron.parallel_state import InitParallelPolicy, CKPTStage
from mindio_acp.acc_checkpoint.utils.utils import time_used, retry

logging = mindio_logger.LOGGER

# the broadcast threshold in dp group
DP_BROADCAST_THRESHOLD = 512


def import_torch_mindio():
    global initialize, preload
    from mindio_acp import initialize, preload


@retry(wait_min=100, wait_max=800)
def get_iteration_from_tracker_file(load_path: str):
    """Get iteration code from tracker file."""
    tracker_filename = os.path.join(load_path, 'latest_checkpointed_iteration.txt')
    # If no tracker file, print log and raise exception
    if not os.path.isfile(tracker_filename):
        raise ValueError(f'could not find the metadata file {tracker_filename}')

    # Otherwise, read the tracker file and get the iteration
    with open(tracker_filename, 'r') as f:
        iteration = int(f.read().strip())
    return iteration


def get_model_parameters_checkpoint_name(checkpoints_path, iteration):
    """Get model parameters checkpoint name."""
    directory = 'iter_{:07d}'.format(iteration)

    args = global_vars.get_args()
    rank = args.rank
    world_size = args.world_size
    tp_size = args.tensor_model_parallel_size
    pp_size = args.pipeline_model_parallel_size
    ep_size = args.expert_model_parallel_size

    # Use both the tensor and pipeline MP rank.
    pipeline_parallel = (pp_size > 1)
    tensor_rank  = rank % tp_size
    pipeline_rank= (rank * pp_size) // world_size
    expert_parallel = (ep_size > 1)
    expert_rank  = (rank // tp_size) % ep_size

    # get common_path by identify pipeline_parallel.
    if not pipeline_parallel:
        common_path = os.path.join(checkpoints_path, directory,
                                   f'mp_rank_{tensor_rank:02d}')
    else:
        common_path = os.path.join(checkpoints_path, directory,
                                   f'mp_rank_{tensor_rank:02d}_{pipeline_rank:03d}')
    if expert_parallel:
        common_path = common_path + f'_{expert_rank:03d}'

    model_ckpt_name = os.path.join(common_path, "model_optim_rng.pt")

    return model_ckpt_name


def get_replica_count():
    """Get replica count to split dp_group, determined by parallel state and DP_BROADCAST_THRESHOLD"""
    args = global_vars.get_args()
    world_size = args.world_size
    tp_size = args.tensor_model_parallel_size
    pp_size = args.pipeline_model_parallel_size
    dp_cp_size = world_size // (tp_size * pp_size)
    replica_count = dp_cp_size // DP_BROADCAST_THRESHOLD
    return replica_count


@time_used
def acp_preload_checkpoint():
    args = global_vars.get_args()
    rank = args.rank

    iteration = get_iteration_from_tracker_file(args.load)
    if iteration is None:
        return
    model_checkpoint_name = get_model_parameters_checkpoint_name(args.load, iteration)
    optim_checkpoint_name = os.path.join(os.path.dirname(model_checkpoint_name), "distrib_optim.pt")

    # identify which rank need preload
    replica_count = get_replica_count()

    policy = InitParallelPolicy(replica_count, None, CKPTStage.PreLoad)

    if args.no_load_optim:
        policy.selected_optim_rank = None

    if rank == policy.selected_optim_rank and rank == policy.selected_model_rank:
        ret = preload(model_checkpoint_name, optim_checkpoint_name)
    elif rank == policy.selected_model_rank:
        ret = preload(model_checkpoint_name)
    elif rank == policy.selected_optim_rank:
        ret = preload(optim_checkpoint_name)
    else:
        return

    if ret != 0:
        logging.warning(f'rank: {rank}, failed preload checkpoint from {args.load} at iteration {iteration}.')


def setup_logging_wrapper(setup_logging):
    @wraps(setup_logging)
    def wrapper(*args, **kwargs):
        import_torch_mindio()
        initialize()
        acp_preload_checkpoint()
        setup_logging(*args, **kwargs)
    return wrapper
