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
from mindio_acp.acc_checkpoint.core.checkpoint_rapid_loader import CheckpointRapidLoaderMixin


def test_async_preload_checkpoint(mocker, monkeypatch, tmpdir):
    latest_checkpointed_iteration = str(tmpdir.join("latest_checkpointed_iteration.txt"))
    with open(latest_checkpointed_iteration, "w") as file:
        file.write(str(10))
    tmpdir.mkdir("iter_{:07d}".format(10))

    load_dir = str(tmpdir)
    mocker.patch("mindio_acp.preload", return_value=None)

    assert CheckpointRapidLoaderMixin(1).async_preload_checkpoint(load_dir) is None


def test_rapid_load_model_checkpoint(mocker, monkeypatch, tmpdir):
    mocker.patch("mindio_acp.load", return_value={"data": 1})

    model_states = CheckpointRapidLoaderMixin(1).rapid_load_model_checkpoint("test", 1, None)
    assert model_states == {"data": 1}
