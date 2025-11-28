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
import time
import torch

from mindio_acp.acc_checkpoint.core.checkpoint_async_saver import CheckpointAsyncSaverMixin


def test_async_save_model_checkpoint(monkeypatch):
    from mindio_acp.acc_checkpoint.core.checkpoint_async_saver import mindio_acp
    monkeypatch.setattr(mindio_acp, "save", lambda *args, **kwargs: 1)

    # first save
    saver = CheckpointAsyncSaverMixin(0)
    saver.async_save_model_checkpoint("ckpt.pt", {"checkpoint_version": 3.0})
    time.sleep(0.01)
    assert saver._CheckpointAsyncSaverMixin__running_count == 0

    # second time save
    saver.async_save_model_checkpoint("ckpt2.pt", {"checkpoint_version": 4.0})
    time.sleep(0.01)
    assert saver._CheckpointAsyncSaverMixin__running_count == 0


def test_async_save_optimizer_checkpoint(monkeypatch):
    from mindio_acp.acc_checkpoint.core.checkpoint_async_saver import mindio_acp
    monkeypatch.setattr(mindio_acp, "save", lambda *args, **kwargs: 1)

    def get_parameter_state_func(notify_callback=None):
        if notify_callback:
            notify_callback()
        return {"stat": 1}

    # first save
    saver = CheckpointAsyncSaverMixin(0)
    saver.async_save_optimizer_checkpoint("ckpt.pt", get_parameter_state_func, True)
    time.sleep(0.01)
    assert saver._CheckpointAsyncSaverMixin__running_count == 0

    def get_parameter_state_func2():
        return {"stat": 1}

    # second time save
    saver.async_save_optimizer_checkpoint("ckpt2.pt", get_parameter_state_func2, False)
    time.sleep(0.01)
    assert saver._CheckpointAsyncSaverMixin__running_count == 0


def test_checkpoint_wait_async_finished(mocker):
    mocker.patch("torch.distributed.get_rank", return_value=0)
    mocker.patch("torch.distributed.barrier", return_value=None)

    ckpt_saver = CheckpointAsyncSaverMixin(0)
    ckpt_saver._CheckpointAsyncSaverMixin__background_saving_checkpoint = True
    ckpt_saver._CheckpointAsyncSaverMixin__running_count = 0
    ckpt_saver.wait_d2h_checkpoint_finished()

    assert torch.distributed.barrier.call_count == 1
