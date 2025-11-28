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
import os
import pytest
import mindio_acp
from mindio_acp.acc_checkpoint.framework_acp import CheckpointHelper


def test_checkpoint_helper_singleton():
    helper1 = CheckpointHelper(0)
    helper2 = CheckpointHelper(0)
    assert id(helper1) == id(helper2)
    helper3 = CheckpointHelper(1)
    assert id(helper1) != id(helper3)


def test_checkpoint_helper_init():
    ckpth = CheckpointHelper(0)
    assert ckpth._CheckpointAsyncSaverMixin__rank == 0
    assert ckpth._CheckpointRapidLoaderMixin__rank == 0


dir_str = '/iteration_dir'
testpt_str = 'test.pt'


@pytest.mark.parametrize(
    "iteration_dir, checker_result, expect",
    [
        pytest.param(dir_str, 0,
                     'the iteration_dir is not a directory.', id='iteration dir not exist'),
        pytest.param(dir_str * 1000, 0,
                     'the path length cannot exceed 1024 characters.', id='iteration dir too long'),
        pytest.param(dir_str, -1, None, id='register checker result not zero'),
        pytest.param(dir_str, 0, None, id='register checker result zero'),
    ])
def test_async_write_tracker_file(mocker, monkeypatch, iteration_dir, checker_result, expect):
    if not isinstance(expect, str):
        mocker.patch("os.path.isdir", return_value=True)

    def callback(iteration, result):
        pass

    def mock_register_checker(_save_post_process, check_dict, user_context, timeout_sec):
        _save_post_process(checker_result, 0)

    monkeypatch.setattr(mindio_acp, "register_checker", mock_register_checker)

    try:
        CheckpointHelper(0).async_write_tracker_file(
            0, iteration_dir, 0, testpt_str, callback)
        if checker_result == 0:
            assert os.path.exists(testpt_str)
            os.remove(testpt_str)
        else:
            assert not os.path.exists(iteration_dir)
    except Exception as e:
        assert str(e) == expect


def test_save_model_checkpoint(monkeypatch):
    monkeypatch.setattr(mindio_acp, "save", lambda *args, **kwargs: 1)
    monkeypatch.setattr(mindio_acp, "multi_save", lambda *args, **kwargs: 1)

    helper = CheckpointHelper(0)
    assert helper.save_model_checkpoint("/tmp/ckpt.pt", {"model": 1}) is None
    assert helper.save_model_checkpoint(["/tmp/ckpt1.pt", "/tmp/ckpt2.pt"], {"model": 1}) is None


def test_save_optimizer_checkpoint(monkeypatch):
    monkeypatch.setattr(mindio_acp, "save", lambda *args, **kwargs: 1)
    monkeypatch.setattr(mindio_acp, "multi_save", lambda *args, **kwargs: 1)

    def get_parameter_state_func():
        return {"param": 1}

    helper = CheckpointHelper(0)
    assert helper.save_optimizer_checkpoint("/tmp/ckpt.pt", get_parameter_state_func) is None

    def get_parameter_state_func2(notify_callback=None):
        return {"param": 1}

    assert helper.save_optimizer_checkpoint(["/tmp/ckpt1.pt", "/tmp/ckpt2.pt"], get_parameter_state_func2) is None
