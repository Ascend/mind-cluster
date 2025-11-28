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

import pytest
from mindio_acp.acc_io.mindio_help import torch_multi_save_helper
from mindio_acp.acc_io.mindio_help import multi_save_helper
from mindio_acp.acc_io.mindio_help import torch_register_checker_helper
from mindio_acp.acc_io.mindio_help import torch_initialize_helper
from mindio_acp.acc_io.mindio_help import torch_preload_helper
from mindio_acp.acc_io.multi_write_handler import create_multi_write_chain


path1_str = 'path1'
path2_str = 'path2'


@pytest.mark.parametrize(
    "check_process, path_list, handle_value, result",
    [
        pytest.param(False, "not_a_list", 0, None, id='path_list is not list'),
        pytest.param(False, [], 0, 0, id='path_list is empty'),
        pytest.param(False, [path1_str, path2_str], 0, 0, id='check_process is false'),
        pytest.param(True, [path1_str, path2_str], 1, 1, id='flink is 1'),
        pytest.param(True, [path1_str, path2_str], 0, 0, id='memfs link is 0'),
    ])
def test_torch_multi_save_helper(mocker, monkeypatch, check_process, path_list, handle_value, result):
    mocker.patch('os.path.realpath', side_effect=lambda path: path)

    mocker.patch("mindio_acp.acc_io.mindio_help.check_process", return_value=check_process)

    mock_writer = mocker.MagicMock()
    mock_writer.handle.return_value = handle_value
    mocker.patch("mindio_acp.acc_io.mindio_help.create_multi_write_chain", return_value=mock_writer)
    mocker.patch("mindio_acp.acc_io.mindio_help.torch_initialize_helper", return_value=0)
    assert torch_multi_save_helper({}, path_list) == result


@pytest.mark.parametrize(
    "callback, check_dict, _register_checker, result",
    [
        pytest.param(None, {}, 1, None, id='invalid_callback'),
        pytest.param(lambda x: x, None, 1, None, id='invalid_check_dict'),
        pytest.param(lambda x: x, {}, 1, None, id='empty_check_dict()'),
        pytest.param(lambda x: x, {'key': 'value'}, 1, None, id='register fail'),
        pytest.param(lambda x: x, {'key': 'value'}, 0, 1, id='register success'),
    ])
def test_torch_register_checker_helper(mocker, monkeypatch, callback, check_dict, _register_checker, result):
    mocker.patch("mindio_acp.c2python_api.register_checker", return_value=_register_checker)
    assert torch_register_checker_helper(callback, check_dict, 1, 1) == result


@pytest.mark.parametrize(
    "_initialize, result",
    [
        pytest.param(-1, -1, id='initialize failed'),
        pytest.param(0, 0, id='initialize succeed'),
    ])
def test_torch_initialize_helper(mocker, monkeypatch, _initialize, result):
    mocker.patch("mindio_acp.c2python_api.initialize", return_value=_initialize)
    assert torch_initialize_helper(None) == result


@pytest.mark.parametrize(
    "path, check_process, _preload, result",
    [
        pytest.param((), False, 0, 1, id='path empty tuple'),
        pytest.param("", False, 0, 1, id='path empty string'),
        pytest.param(None, False, 0, 1, id='path None'),
        pytest.param(1, False, 0, 1, id='path invalid'),
        pytest.param(path1_str, False, 0, 1, id='check_process is false'),
        pytest.param(path1_str, True, 1, 1, id='preload failed'),
        pytest.param(path1_str, True, 0, 0, id='preload success'),
    ])
def test_torch_preload_helper(mocker, monkeypatch, path, check_process, _preload, result):
    mocker.patch("mindio_acp.acc_io.mindio_help.check_process", return_value=check_process)
    mocker.patch("mindio_acp.acc_io.mindio_help.torch_initialize_helper", return_value=0)
    mocker.patch("mindio_acp.c2python_api.preload", return_value=_preload)
    assert torch_preload_helper(path) == result
