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
import torch

from mindio_acp.acc_checkpoint.megatron.parallel_state import InitParallelResult


ckpt_str = 'ckpt'
load_checkpoint_for_model_str = 'load_checkpoint_for_model'


@pytest.mark.parametrize(
    "rank, global_ranks, dp_broadcast_threshold",
    [
        pytest.param(0, [0], 512, id=load_checkpoint_for_model_str),
        pytest.param(0, [0, 2, 4, 6], 512, id=load_checkpoint_for_model_str),
        pytest.param(0, [0, 2, 4, 6], 2, id=load_checkpoint_for_model_str),
    ])
def test_load_checkpoint_for_model_8npus_1host(mocker, monkeypatch, rank: int, global_ranks: list,
                                               dp_broadcast_threshold):
    pytest.skip("skip: need to migration to test_checkpoint_rapid_loader.py")
    expect_dict = {
        "args": {
            "DDP_impl": "local",
        },
    }
    monkeypatch.setattr(torch.distributed, "get_world_size", lambda: 8)

    mock_args = namedtuple("Args",
                           ["rank", "world_size", "replica_count",
                            "context_parallel_size", "expert_model_parallel_size",
                            "tensor_model_parallel_size", "pipeline_model_parallel_size"])
    mock_args.rank = rank
    mock_args.world_size = 8
    mock_args.replica_count = 2
    mock_args.context_parallel_size = 1
    mock_args.expert_model_parallel_size = 1
    mock_args.tensor_model_parallel_size = 2
    mock_args.pipeline_model_parallel_size = 1

    from megatron.training import global_vars
    monkeypatch.setattr(global_vars, "get_args", lambda: mock_args)
    monkeypatch.setattr(torch.distributed, "get_rank", lambda: rank)
    monkeypatch.setattr(torch.npu, "current_device", lambda: "npu")
    mocker.patch("mindio_acp.acc_checkpoint.core.checkpoint_loader.get_data_modulo_expert_parallel_group",
                 return_value="mock_group")
    mocker.patch("torch.distributed.get_process_group_ranks", return_value=global_ranks)
    mocker.patch("torch_npu.npu.Stream", return_value="npu_stream")
    mocker.patch("mindio_acp.load", return_value=expect_dict)
    mocker.patch("mindio_acp.acc_checkpoint.core.checkpoint_loader"
                 ".CheckpointLoaderMixin._CheckpointLoaderMixin__load_model_parameters", return_value=expect_dict)
    mocker.patch("torch.npu.synchronize", return_value=None)
    monkeypatch.setattr("mindio_acp.acc_checkpoint.core.checkpoint_loader.DP_BROADCAST_THRESHOLD",
                        dp_broadcast_threshold)

    monkeypatch.setattr(torch.distributed, "new_group", lambda *args, **kwargs: "processGroup")

    monkeypatch.setenv("LOCAL_WORLD_SIZE", str(8))

    loader = CheckpointLoaderMixin()
    model_state_dict = loader._load_model_checkpoint("ckpt.pt")
    assert model_state_dict == expect_dict


@pytest.mark.parametrize(
    "rank, global_ranks, dp_broadcast_threshold",
    [
        pytest.param(0, [0, 4, 8, 12], 1, id='load_checkpoint_for_model_2host'),
    ])
def test_load_checkpoint_for_model_16npus_2host(mocker, monkeypatch, rank: int, global_ranks: list,
                                                dp_broadcast_threshold):
    pytest.skip("skip: need to migration to test_checkpoint_rapid_loader.py")
    expect_dict = {
        "args": {
            "DDP_impl": "local",
        },
    }
    monkeypatch.setattr(torch.distributed, "get_world_size", lambda: 16)

    mock_args = namedtuple("Args",
                           ["rank", "world_size", "replica_count",
                            "context_parallel_size", "expert_model_parallel_size",
                            "tensor_model_parallel_size", "pipeline_model_parallel_size"])
    mock_args.rank = rank
    mock_args.world_size = 16
    mock_args.replica_count = 2
    mock_args.context_parallel_size = 1
    mock_args.expert_model_parallel_size = 1
    mock_args.tensor_model_parallel_size = 2
    mock_args.pipeline_model_parallel_size = 4

    from megatron.training import global_vars
    monkeypatch.setattr(global_vars, "get_args", lambda: mock_args)
    monkeypatch.setattr(torch.distributed, "get_rank", lambda: rank)
    monkeypatch.setattr(torch.npu, "current_device", lambda: "npu")
    mocker.patch("torch_npu.npu.Stream", return_value="npu_stream")
    mocker.patch("mindio_acp.acc_checkpoint.core.checkpoint_loader.get_data_modulo_expert_parallel_group",
                 return_value="mock_group")
    mocker.patch("torch.distributed.get_process_group_ranks", return_value=global_ranks)
    mocker.patch("mindio_acp.load", return_value=expect_dict)
    mocker.patch("mindio_acp.acc_checkpoint.core.checkpoint_loader"
                 ".CheckpointLoaderMixin._CheckpointLoaderMixin__load_model_parameters", return_value=expect_dict)
    mocker.patch("torch.npu.synchronize", return_value=None)
    monkeypatch.setattr("mindio_acp.acc_checkpoint.core.checkpoint_loader.DP_BROADCAST_THRESHOLD",
                        dp_broadcast_threshold)

    monkeypatch.setattr(torch.distributed, "new_group", lambda *args, **kwargs: "processGroup")

    monkeypatch.setenv("LOCAL_WORLD_SIZE", str(8))

    loader = CheckpointLoaderMixin()
    model_state_dict = loader._load_model_checkpoint("ckpt.pt")
    assert model_state_dict == expect_dict


@pytest.mark.parametrize(
    "checkpoint_name, policy, exception",
    [
        pytest.param(ckpt_str, InitParallelResult(selected_model_rank=0, selected_optim_rank=0,
                                                process_group=None, group_ranks=[0]), None, id='group ranks size is 1'),
        pytest.param(ckpt_str, InitParallelResult(selected_model_rank=0, selected_optim_rank=0,
                                                process_group=None, group_ranks=[0, 1]), None, id='rank to broadcast'),
        pytest.param(ckpt_str, InitParallelResult(selected_model_rank=1, selected_optim_rank=0,
                                                process_group=None, group_ranks=[0, 1]), None, id='rank to receive'),
        pytest.param(ckpt_str, InitParallelResult(selected_model_rank=1, selected_optim_rank=0, process_group=None,
                                                group_ranks=[0, 1]), 'ModuleNotFoundError', id='load not found error')
    ]
)
def test_load_model_parameters(mocker, checkpoint_name, policy, exception):
    pytest.skip("skip: need to migration to test_checkpoint_rapid_loader.py")
    expect_dict = {
        "args": {
            "DDP_impl": "local",
        },
    }
    mocker.patch("torch.load", return_value=expect_dict)
    mocker.patch("mindio_acp.load", return_value=expect_dict)
    mocker.patch("mindio_acp.acc_checkpoint.core.checkpoint_loader"
                 ".CheckpointLoaderMixin._CheckpointLoaderMixin__broadcast_checkpoint_from_src_rank",
                 return_value=expect_dict)
    mocker.patch("mindio_acp.acc_checkpoint.core.checkpoint_loader"
                 ".CheckpointLoaderMixin._CheckpointLoaderMixin__receive_checkpoint_at_dst_rank",
                 return_value=expect_dict)
    if exception is None:
        mocker.patch("torch.distributed.get_rank", return_value=0)
    else:
        mocker.patch("torch.distributed.get_rank", side_effect=ModuleNotFoundError)

    assert CheckpointLoaderMixin._CheckpointLoaderMixin__load_model_parameters(checkpoint_name, policy) == expect_dict
