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
import mindio_acp


@pytest.mark.parametrize(
    "callback, check_dict, user_context, timeout_sec, checker_result, expect",
    [
        pytest.param(None, None, None, 0, None, None, id='timeout second too low'),
        pytest.param(None, None, None, 60 * 61, None, None, id='timeout second too high'),
        pytest.param(None, None, None, 10, 1, 1, id='normal register checker')
    ])
def test_mindio_register_checker(mocker, callback, check_dict, user_context, timeout_sec, checker_result, expect):
    mocker.patch("mindio_acp.acc_io.acc_io.torch_register_checker_helper", return_value=checker_result)
    assert mindio_acp.register_checker(callback, check_dict, user_context, timeout_sec) == expect


@pytest.mark.parametrize(
    "obj, path_list, multi_save_result, expect",
    [
        pytest.param({}, [], 0, 0, id='normal multi save')
    ])
def test_mindio_multi_save(mocker, obj, path_list, multi_save_result, expect):
    mocker.patch("mindio_acp.acc_io.acc_io.torch_multi_save_helper", return_value=multi_save_result)
    assert mindio_acp.multi_save(obj, path_list) == expect


@pytest.mark.parametrize(
    "obj, path, save_result, expect",
    [
        pytest.param({}, 'path', 0, 0, id='normal save')
    ])
def test_mindio_save(mocker, obj, path, save_result, expect):
    mocker.patch("mindio_acp.acc_io.acc_io.torch_save_helper", return_value=save_result)
    assert mindio_acp.save(obj, path) == expect


@pytest.mark.parametrize(
    "path, map_location, load_result, expect",
    [
        pytest.param('path', None, {}, {}, id='normal load'),
        pytest.param(None, 'npu', {}, {}, id='load with npu')
    ])
def test_mindio_load(mocker, path, map_location, load_result, expect):
    mocker.patch("mindio_acp.acc_io.acc_io.torch_load_helper", return_value=load_result)
    try:
        ret = mindio_acp.load(path, map_location=map_location)
        assert ret == expect
    except ValueError as e:
        assert str(e) == "param map_location only support value 'cpu'"


@pytest.mark.parametrize(
    "paths, preload_result, expect",
    [
        pytest.param('source', 0, 0, id='normal preload')
    ])
def test_mindio_preload(mocker, paths, preload_result, expect):
    mocker.patch("mindio_acp.acc_io.acc_io.torch_preload_helper", return_value=preload_result)
    mocker.patch("mindio_acp.acc_io.mindio_help.torch_initialize_helper", return_value=0)
    assert mindio_acp.preload(paths) == expect
