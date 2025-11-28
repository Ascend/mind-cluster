#!/usr/bin/env python
# coding=utf-8
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
from collections import namedtuple

import pytest

from mindio_acp.acc_checkpoint.megatron.parallel_state import InitParallelPolicy, CKPTStage

DistConfig = namedtuple("DistConfig", "rank, world_size, pp_size, tp_size, ep_size, nproc_per_node, replica_count")
ExpectPartResult = namedtuple("ExpectPartResult", "group, rank_model, rank_optim")


@pytest.mark.parametrize(
    "dist_config, expect_part_result, group_ranks",
    [
        # 16npu 2host replica 1 load
        # DP CP Group [0, 2, 4, 6, 8, 10, 12, 14]
        # DP CP Group [0, 2, 4, 6, 8, 10, 12, 14]
        pytest.param(
            DistConfig(rank=0, world_size=16, pp_size=1, tp_size=2, ep_size=1, nproc_per_node=8, replica_count=1),
            ExpectPartResult(group=list(range(0, 16, 2)), rank_model=0, rank_optim=0), list(range(0, 16, 2)),
            id='dp:8/tp:2/pp:1/cp:1/ep:1/rank:0/replica:1/selected model and optim'),
        pytest.param(
            DistConfig(rank=2, world_size=16, pp_size=1, tp_size=2, ep_size=1, nproc_per_node=8, replica_count=1),
            ExpectPartResult(group=list(range(0, 16, 2)), rank_model=0, rank_optim=0), list(range(0, 16, 2)),
            id='dp:8/tp:2/pp:1/cp:1/ep:1/rank:2/replica:1/not selected'),
        pytest.param(
            DistConfig(rank=10, world_size=16, pp_size=1, tp_size=2, ep_size=1, nproc_per_node=8, replica_count=1),
            ExpectPartResult(group=list(range(0, 16, 2)), rank_model=0, rank_optim=0), list(range(0, 16, 2)),
            id='dp:8/tp:2/pp:1/cp:1/ep:1/rank:10/replica:1/not selected'),
        # 32npu 4host replica 2 load
        # DP CP GROUP [[0, 2, 4, 6, 8, 10, 12, 14], ... , [17, 19, 21, 23, 25, 27, 29, 31]]
        # DP EP GROUP [[0, 4, 8, 12], [1, 5, 9, 13], ..., [18, 22, 26, 30], [19, 23, 27, 31]]
        pytest.param(
            DistConfig(rank=0, world_size=32, pp_size=2, tp_size=2, ep_size=2, nproc_per_node=8, replica_count=2),
            ExpectPartResult(group=list(range(0, 8, 4)), rank_model=4, rank_optim=0), list(range(0, 16, 4)),
            id='dp:8/tp:2/pp:2/cp:1/ep:2/rank:0/replica:2/selected optim'),
        pytest.param(
            DistConfig(rank=2, world_size=32, pp_size=2, tp_size=2, ep_size=2, nproc_per_node=8, replica_count=2),
            ExpectPartResult(group=list(range(2, 8, 4)), rank_model=6, rank_optim=2), list(range(2, 16, 4)),
            id='dp:8/tp:2/pp:2/cp:1/ep:2/rank:2/replica:2/selected optim'),
        pytest.param(
            DistConfig(rank=4, world_size=32, pp_size=2, tp_size=2, ep_size=2, nproc_per_node=8, replica_count=2),
            ExpectPartResult(group=list(range(0, 8, 4)), rank_model=4, rank_optim=0), list(range(0, 16, 4)),
            id='dp:8/tp:2/pp:2/cp:1/ep:2/rank:4/replica:2/selected model'),
        pytest.param(
            DistConfig(rank=6, world_size=32, pp_size=2, tp_size=2, ep_size=2, nproc_per_node=8, replica_count=2),
            ExpectPartResult(group=list(range(2, 8, 4)), rank_model=6, rank_optim=2), list(range(2, 16, 4)),
            id='dp:8/tp:2/pp:2/cp:1/ep:2/rank:6/replica:2/selected model'),])
def test_partition_dp_group(mocker, monkeypatch, dist_config, expect_part_result, group_ranks):
    mock_args = namedtuple("Args",
                           ["rank", "world_size", "replica_count",
                            "context_parallel_size", "expert_model_parallel_size",
                            "tensor_model_parallel_size", "pipeline_model_parallel_size"])
    mock_args.rank = dist_config.rank
    mock_args.world_size = dist_config.world_size
    mock_args.replica_count = dist_config.replica_count
    mock_args.context_parallel_size = 1
    mock_args.expert_model_parallel_size = dist_config.ep_size
    mock_args.tensor_model_parallel_size = dist_config.tp_size
    mock_args.pipeline_model_parallel_size = dist_config.pp_size
    selected_rank_model_expect = expect_part_result.rank_model
    selected_rank_optim_expect = expect_part_result.rank_optim
    group_expect = expect_part_result.group

    from megatron.training import global_vars
    monkeypatch.setattr(global_vars, "get_args", lambda: mock_args)
    mocker.patch("torch.distributed.new_group", return_value=None)
    monkeypatch.setenv("LOCAL_WORLD_SIZE", str(dist_config.nproc_per_node))

    policy = InitParallelPolicy(dist_config.replica_count, None, CKPTStage.LoadDPEP)

    assert policy.group_ranks == group_expect
    assert policy.selected_model_rank == selected_rank_model_expect
    assert policy.selected_optim_rank == selected_rank_optim_expect


model_128k_pp_total = 128 * 1024 // 4  # 32768, PP: 4
world_size_128k = int(128 * 8 * 8 * 16)


@pytest.mark.parametrize(
    "dist_config, expect_part_result, group_ranks",
    [
        # DP-EP 0: [0, 512, 1024, 1536, 2048, ..., 16384, 16896, 17408, ..., 31232, 31744, 32256]
        pytest.param(
            DistConfig(rank=16896, world_size=world_size_128k, pp_size=4, tp_size=16, ep_size=32,
                       nproc_per_node=16, replica_count=2),
            ExpectPartResult(group=list(range(16384, 16384 + model_128k_pp_total // 2, 512)),
                             rank_model=16896, rank_optim=16384),
            list(range(16384, model_128k_pp_total, 512)),
            id='dp:2048/tp:16/pp:4/cp:1/ep:32/rank:16384/replica:2/selected model'),
        # DP-EP 0: [0, 512, 1024, 1536, 2048, ..., 16384, 16400, 16416, ..., 31232, 31744, 32256]
        pytest.param(
            DistConfig(rank=16384, world_size=world_size_128k, pp_size=4, tp_size=16, ep_size=32,
                       nproc_per_node=16, replica_count=2),
            ExpectPartResult(group=list(range(16384, 16384 + model_128k_pp_total // 2, 512)),
                             rank_model=16896, rank_optim=16384),
            list(range(16384, 16384 + model_128k_pp_total // 2, 512)),
            id='dp:2048/tp:16/pp:4/cp:1/ep:32/rank:16384/replica:2/selected optim'),
    ])
def test_partition_dp_group_for_128k(mocker, monkeypatch, dist_config, expect_part_result, group_ranks):
    mock_args = namedtuple("Args",
                           ["rank", "world_size", "replica_count",
                            "context_parallel_size", "expert_model_parallel_size",
                            "tensor_model_parallel_size", "pipeline_model_parallel_size"])
    mock_args.rank = dist_config.rank
    mock_args.world_size = dist_config.world_size
    mock_args.replica_count = dist_config.replica_count
    mock_args.context_parallel_size = 1
    mock_args.expert_model_parallel_size = dist_config.ep_size
    mock_args.tensor_model_parallel_size = dist_config.tp_size
    mock_args.pipeline_model_parallel_size = dist_config.pp_size
    selected_rank_model_expect = expect_part_result.rank_model
    selected_rank_optim_expect = expect_part_result.rank_optim
    group_expect = expect_part_result.group

    from megatron.training import global_vars
    monkeypatch.setattr(global_vars, "get_args", lambda: mock_args)
    mocker.patch("torch.distributed.new_group", return_value=None)
    monkeypatch.setenv("LOCAL_WORLD_SIZE", str(dist_config.nproc_per_node))

    policy = InitParallelPolicy(dist_config.replica_count, None, CKPTStage.LoadDPEP)

    assert policy.selected_model_rank == selected_rank_model_expect
    assert policy.selected_optim_rank == selected_rank_optim_expect
    assert policy.group_ranks == group_expect
