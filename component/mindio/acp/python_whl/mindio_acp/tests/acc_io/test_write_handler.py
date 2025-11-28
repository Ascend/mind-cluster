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

from mindio_acp.acc_io.write_handler import import_mindio_sdk_api
from mindio_acp.acc_io.write_handler import create_write_chain


@pytest.mark.parametrize(
    "check_process, memfs_write_res, f_write_res, torch_save_res, saver_marker",
    [
        pytest.param(True, 0, 0, None, 1, id='memfs write success'),
        pytest.param(False, -1, 0, None, 2, id='f write success'),
        pytest.param(False, -1, -1, None, 0, id='torch save success'),
    ])
def test_write_chain_handle_success(mocker, monkeypatch, check_process, memfs_write_res, f_write_res, torch_save_res,
                                    saver_marker):
    memfs_writer = mocker.MagicMock()
    mocker.patch("mindio_acp.acc_io.write_handler.check_process", return_value=check_process)
    monkeypatch.setattr(memfs_writer, "write_list", lambda *args, **kwargs: memfs_write_res)
    monkeypatch.setattr(memfs_writer, "write", lambda *args, **kwargs: memfs_write_res)
    monkeypatch.setattr(memfs_writer, "create", lambda *args, **kwargs: 0)
    mocker.patch("mindio_acp.c2python_api.writeable_file", return_value=memfs_writer)

    f_writer = mocker.MagicMock()
    monkeypatch.setattr(f_writer, "write_list", lambda *args, **kwargs: f_write_res)
    monkeypatch.setattr(f_writer, "write", lambda *args, **kwargs: f_write_res)
    monkeypatch.setattr(f_writer, "create", lambda *args, **kwargs: 0)

    mocker.patch("mindio_acp.c2python_api.fwriteable_file", return_value=f_writer)

    mocker.patch("mindio_acp.acc_io.write_handler.torch.save", return_value=None)

    import_mindio_sdk_api()

    def writer_handle(write_content):
        ckpt_obj = {"args": "test args"}
        handler = create_write_chain(ckpt_obj, "path", "memfs")
        res = handler.handle(write_content)
        assert res == saver_marker

    # case 1: write list
    writer_handle([("data.pkl", 0, 10), ("data/0", 11, 20)])
    # case 2:  write bytes
    writer_handle(b'102030405060')


def test_write_chain_handle_error(mocker, monkeypatch):
    memfs_writer = mocker.MagicMock()
    monkeypatch.setattr(memfs_writer, "write_list", lambda *args, **kwargs: -1)
    monkeypatch.setattr(memfs_writer, "write", lambda *args, **kwargs: -1)
    mocker.patch("mindio_acp.c2python_api.writeable_file", return_value=memfs_writer)

    f_writer = mocker.MagicMock()
    monkeypatch.setattr(f_writer, "write_list", lambda *args, **kwargs: -1)
    monkeypatch.setattr(f_writer, "write", lambda *args, **kwargs: -1)
    mocker.patch("mindio_acp.c2python_api.fwriteable_file", return_value=f_writer)

    mocker.patch("mindio_acp.acc_io.write_handler.torch.save", side_effect=RuntimeError('mocked error'))

    def writer_handle(write_content):
        ckpt_obj = {"args": "test args"}
        handler = create_write_chain(ckpt_obj, "path", "memfs")
        with pytest.raises(RuntimeError) as e:
            handler.handle(write_content)
        assert e.value.args[0] == "No write policy is available"

    import_mindio_sdk_api()
    # case 1: write list
    writer_handle([("data.pkl", 0, 10), ("data/0", 11, 20)])
    # case 2:  write bytes
    writer_handle(b'102030405060')
