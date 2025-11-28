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

from collections import namedtuple

import mindio_acp
from mindio_acp.acc_checkpoint.megatron import save_checkpoint_patch  # for mocking
from mindio_acp.acc_checkpoint.megatron.save_checkpoint_patch import save_checkpoint_wrapper


def test_save_checkpoint(mocker, monkeypatch, tmp_path, tmpdir):
    save_checkpoint_patch.import_torch_mindio()
    tmpdir.mkdir('iter_0000010')

    Args = namedtuple("Args",
                      ["use_dist_ckpt", "async_save", "save", "use_distributed_optimizer", "no_save_optim", "rank",
                       "pipeline_model_parallel_size", "tensor_model_parallel_size", "expert_model_parallel_size",
                       "ckpt_format"])
    mock_args = Args(False, False, str(tmpdir), True, False, 0, 2, 4, 2, "torch")
    monkeypatch.setattr(save_checkpoint_patch.global_vars, "get_args", lambda: mock_args)
    monkeypatch.setattr(save_checkpoint_patch.utils, "unwrap_model", lambda model: "model")
    monkeypatch.setattr(save_checkpoint_patch, "generate_state_dict", lambda *args, **kwargs: {"state": 1})
    monkeypatch.setattr(save_checkpoint_patch.torch.distributed, "get_process_group_ranks",
                        lambda group: "group_ranks")

    monkeypatch.setattr(save_checkpoint_patch.torch.distributed, "is_initialized", lambda: True)
    monkeypatch.setattr(save_checkpoint_patch.torch.distributed, "get_rank", lambda: 0)
    monkeypatch.setattr(save_checkpoint_patch.torch.distributed, "barrier", lambda: None)

    monkeypatch.setattr(save_checkpoint_patch.mpu, "get_expert_data_parallel_rank", lambda: 0)

    monkeypatch.setattr(save_checkpoint_patch, "get_parameter_state",
                        lambda chain_optimizer, notify_callback=None: {"states": 1})

    # patch checkpoint path
    optimizer_checkpoint_path = str(tmp_path / "distrib_optimizer.pt")
    model_checkpoint_path = str(tmp_path / "model.pt")
    monkeypatch.setattr(save_checkpoint_patch, "get_distributed_optimizer_checkpoint_name",
                        lambda *args, **kwargs: optimizer_checkpoint_path)
    monkeypatch.setattr(save_checkpoint_patch, "get_checkpoint_name", lambda *args, **kwargs: model_checkpoint_path)

    # patch mindio_acp api
    monkeypatch.setattr(mindio_acp, "flush", lambda: None)
    monkeypatch.setattr(mindio_acp, "save", lambda *args, **kwargs: 1)
    monkeypatch.setattr(mindio_acp, "register_checker", lambda *args, **kwargs: None)

    model = {"version": 3.0}
    optimizer = mocker.MagicMock()
    mock_original_fn = mocker.MagicMock()
    wrapped_fn = save_checkpoint_wrapper(mock_original_fn)

    res = wrapped_fn(iteration=10,
                     model=model,
                     optimizer=optimizer,
                     opt_param_scheduler="opt_param_scheduler",
                     num_floating_point_operations_so_far="num_floating_point_operations_so_far")
    assert res is None
