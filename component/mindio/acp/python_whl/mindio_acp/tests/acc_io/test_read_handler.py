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
from mindio_acp.acc_io.read_handler import create_read_chain
from mindio_acp.acc_io.read_handler import CHECKPOINT_TAIL_OFFSET, OCKIO_TAIL_OFFSET
from mindio_acp.acc_io.read_handler import ReadHandlerCtx

ACCACP_BYTES = b'3040506030405060000001accacp'
OCKIO_BYTES = b'3040506030405060mindio\x00'
WRONG_BYTES = b'3040506030405060708090100110'
read_str = "read"
open_str = "open"
size_str = "size"


@pytest.mark.parametrize(
    "memfs_read_res, nds_read_res, f_read_res, read_bytes",
    [
        pytest.param(0, 0, 0, OCKIO_BYTES, id='memfs read ockio bytes success'),
        pytest.param(0, 0, 0, ACCACP_BYTES, id='memfs read accacp bytes success'),
        pytest.param(-1, 0, 0, ACCACP_BYTES, id='f read accacp bytes success'),
        pytest.param(-1, -1, -1, ACCACP_BYTES, id='torch load success'),
    ])
def test_write_chain_handle_success(mocker, monkeypatch, memfs_read_res, f_read_res, nds_read_res, read_bytes):
    ckpt_obj = {"args": "test args"}
    memfs_reader = mocker.MagicMock()
    mock_read_res = 0
    checkpoint_tail_offset = CHECKPOINT_TAIL_OFFSET
    mock_record_map = {"data.pkl": (0, 10), "data/0": (11, 20)}
    mocker.patch("mindio_acp.acc_io.read_handler.SerializationMixin.unmarshal_checkpoint", return_value=ckpt_obj)

    monkeypatch.setattr(memfs_reader, read_str, lambda *args, **kwargs: mock_read_res)
    monkeypatch.setattr(memfs_reader, open_str, lambda *args, **kwargs: memfs_read_res)
    monkeypatch.setattr(memfs_reader, size_str, lambda *args, **kwargs: checkpoint_tail_offset)
    mocker.patch("mindio_acp.c2python_api.readable_file", return_value=memfs_reader)

    nds_reader = mocker.MagicMock()
    monkeypatch.setattr(nds_reader, read_str, lambda *args, **kwargs: mock_read_res)
    monkeypatch.setattr(nds_reader, open_str, lambda *args, **kwargs: nds_read_res)
    monkeypatch.setattr(nds_reader, size_str, lambda *args, **kwargs: checkpoint_tail_offset)
    mocker.patch("mindio_acp.c2python_api.nds_readable_file", return_value=nds_reader)

    f_reader = mocker.MagicMock()
    monkeypatch.setattr(f_reader, read_str, lambda *args, **kwargs: mock_read_res)
    monkeypatch.setattr(f_reader, open_str, lambda *args, **kwargs: f_read_res)
    monkeypatch.setattr(f_reader, size_str, lambda *args, **kwargs: checkpoint_tail_offset)
    mocker.patch("mindio_acp.c2python_api.freadable_file", return_value=f_reader)
    mocker.patch("mindio_acp.acc_io.read_handler.torch.load", return_value=ckpt_obj)

    import_mindio_sdk_api()

    def reader_handle():
        handler = create_read_chain("path", "memfs")
        res = handler.handle()
        assert res == ckpt_obj

    reader_handle()


@pytest.mark.parametrize(
    "memfs_read_res, nds_read_res, f_read_res",
    [
        pytest.param(0, 0, 0, id='memfs read ockio direct bytes success'),
        pytest.param(-1, 0, 0, id='f read ockio direct bytes success'),
    ])
def test_write_chain_handle_ockio_direct_bytes_read(mocker, monkeypatch, memfs_read_res, nds_read_res, f_read_res):
    memfs_reader = mocker.MagicMock()
    mock_read_res = 0
    checkpoint_tail_offset = OCKIO_TAIL_OFFSET
    mock_record_map = {"mindio_save_data_type": 'directly_bytes', "bytes_length": len(OCKIO_BYTES)}
    mocker.patch("pickle.loads", return_value=mock_record_map)
    mocker.patch("mindio_acp.acc_io.read_handler.ReadHandlerCtx.read_file", return_value=OCKIO_BYTES)

    monkeypatch.setattr(memfs_reader, open_str, lambda *args, **kwargs: memfs_read_res)
    monkeypatch.setattr(memfs_reader, size_str, lambda *args, **kwargs: checkpoint_tail_offset)
    mocker.patch("mindio_acp.c2python_api.readable_file", return_value=memfs_reader)

    nds_reader = mocker.MagicMock()
    monkeypatch.setattr(nds_reader, open_str, lambda *args, **kwargs: nds_read_res)
    monkeypatch.setattr(nds_reader, size_str, lambda *args, **kwargs: checkpoint_tail_offset)
    mocker.patch("mindio_acp.c2python_api.nds_readable_file", return_value=nds_reader)

    f_reader = mocker.MagicMock()
    monkeypatch.setattr(f_reader, open_str, lambda *args, **kwargs: f_read_res)
    monkeypatch.setattr(f_reader, size_str, lambda *args, **kwargs: checkpoint_tail_offset)
    mocker.patch("mindio_acp.c2python_api.freadable_file", return_value=f_reader)
    import_mindio_sdk_api()

    def reader_handle():
        handler = create_read_chain("path", "memfs")
        handler.handle()

    reader_handle()


@pytest.mark.parametrize(
    "memfs_read_res, nds_read_res, f_read_res",
    [
        pytest.param(0, 0, 0, id='memfs read bytes failed'),
        pytest.param(-1, 0, 0, id='f read bytes failed'),
    ])
def test_write_chain_handle_wrong_format_error(mocker, monkeypatch, memfs_read_res, nds_read_res, f_read_res):
    memfs_reader = mocker.MagicMock()
    mock_read_res = (0, WRONG_BYTES)
    monkeypatch.setattr(memfs_reader, read_str, lambda *args, **kwargs: mock_read_res)
    monkeypatch.setattr(memfs_reader, open_str, lambda *args, **kwargs: memfs_read_res)
    mocker.patch("mindio_acp.c2python_api.readable_file", return_value=memfs_reader)

    nds_reader = mocker.MagicMock()
    monkeypatch.setattr(nds_reader, read_str, lambda *args, **kwargs: mock_read_res)
    monkeypatch.setattr(nds_reader, open_str, lambda *args, **kwargs: nds_read_res)
    mocker.patch("mindio_acp.c2python_api.nds_readable_file", return_value=nds_reader)

    f_reader = mocker.MagicMock()
    monkeypatch.setattr(f_reader, read_str, lambda *args, **kwargs: mock_read_res)
    monkeypatch.setattr(f_reader, open_str, lambda *args, **kwargs: f_read_res)
    mocker.patch("mindio_acp.c2python_api.freadable_file", return_value=f_reader)
    mocker.patch("mindio_acp.acc_io.read_handler.torch.load", return_value=b'test')
    import_mindio_sdk_api()

    def reader_handle():
        handler = create_read_chain("path", "memfs")
        handler.handle()

    reader_handle()


def test_read_chain_handle_error(mocker, monkeypatch):
    memfs_reader = mocker.MagicMock()
    monkeypatch.setattr(memfs_reader, "multi_read", lambda *args, **kwargs: -1)
    monkeypatch.setattr(memfs_reader, "read", lambda *args, **kwargs: -1)
    mocker.patch("mindio_acp.c2python_api.readable_file", return_value=memfs_reader)

    f_reader = mocker.MagicMock()
    monkeypatch.setattr(f_reader, "multi_read", lambda *args, **kwargs: -1)
    monkeypatch.setattr(f_reader, "read", lambda *args, **kwargs: -1)
    mocker.patch("mindio_acp.c2python_api.freadable_file", return_value=f_reader)

    mocker.patch("mindio_acp.acc_io.read_handler.torch.load", side_effect=RuntimeError('mocked error'))

    def reader_handle():
        handler = create_read_chain("path", "memfs")
        with pytest.raises(RuntimeError) as e:
            handler.handle()
        assert e.value.args[0] == "No read policy is available"

    import_mindio_sdk_api()
    reader_handle()
