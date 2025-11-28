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
from functools import wraps

from megatron.training import global_vars

from mindio_acp.acc_checkpoint.megatron.parallel_state import InitParallelPolicy, CKPTStage

# the broadcast threshold in dp group
DP_BROADCAST_THRESHOLD = 512


def import_torch_mindio():
    global CheckpointHelper, initialize
    from mindio_acp.acc_checkpoint.framework_acp import CheckpointHelper
    from mindio_acp import initialize


def get_replica_count():
    """Get replica count to split dp_group, determined by parallel state and DP_BROADCAST_THRESHOLD"""
    args = global_vars.get_args()
    world_size = args.world_size
    tp_size = args.tensor_model_parallel_size
    pp_size = args.pipeline_model_parallel_size
    dp_cp_size = world_size // (tp_size * pp_size)
    replica_count = dp_cp_size // DP_BROADCAST_THRESHOLD
    return replica_count


def acp_preload_checkpoint():
    args = global_vars.get_args()

    # identify which rank need preload
    replica_count = get_replica_count()

    policy = InitParallelPolicy(replica_count, None, CKPTStage.PreLoad)
    load_rank = policy.selected_optim_rank
    if args.rank == load_rank:
        CheckpointHelper(args.rank).async_preload(args.load)


def initialize_model_parallel_wrapper(initialize_model_parallel):
    @wraps(initialize_model_parallel)
    def wrapper(*args, **kwargs):
        import_torch_mindio()
        initialize()
        acp_preload_checkpoint()
        initialize_model_parallel(*args, **kwargs)

    return wrapper
