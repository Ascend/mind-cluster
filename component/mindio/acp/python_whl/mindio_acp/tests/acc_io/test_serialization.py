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
import types
import pickle
from collections import OrderedDict

import torch
import pytest

from mindio_acp.acc_io.mindio_help import _TorchSaveHelp
from mindio_acp.acc_io.serialization import SerializationMixin, _get_restore_func

if torch.__version__ > torch.torch_version.TorchVersion('1.12'):
    from torch.storage import TypedStorage as TypedStorage
else:
    from torch.storage import _TypedStorage as TypedStorage


def test_marshal_checkpoint(mocker, monkeypatch):
    mocker.patch("mindio_acp.acc_io.serialization.location_tag", return_value=None)

    tensor1 = torch.tensor([[1, 2], [3, 4]], dtype=torch.float32)
    tensor2 = torch.tensor([[1, 3], [2, 4]], dtype=torch.float32)
    ckpt_obj = {
        "model_state1": tensor1,
        "model_state2": tensor2,
        "optimizer_state": {"lr": 0.001, "momentum": 0.9},
        "epoch": 5
    }

    torch_save_helper = _TorchSaveHelp()
    data_value, serialized_storages, record_buff = torch_save_helper.marshal_checkpoint(ckpt_obj)
    record_map = pickle.loads(record_buff)

    assert isinstance(data_value, bytes), "data_value should be of type bytes"
    assert isinstance(serialized_storages, OrderedDict), "serialized_storages should be an OrderedDict"
    assert isinstance(record_map, dict), "record_map should be of type dict"

    assert 'data.pkl' in record_map, "record_map should contain 'data.pkl'"
    assert 'data/0' in record_map, "record_map should contain 'data/0'"
    assert 'data/1' in record_map, "record_map should contain 'data/1'"
    assert len(serialized_storages) == 2, "Expected serialized_storages to have two stored elements"


TENSOR_DICTS = OrderedDict([
    ('key1', TypedStorage([1, 2])),
    ('key2', TypedStorage([2, 3])),
    ('key3', TypedStorage([3, 4])),
])

EXPECTED_RES_WITH_DICT = [(b'data_buff_test_str', 18),
                          (b'data.pkl:0,10;data/0:11,20*\x00\x00\x00\x00\x00\x00\x00'
                           b'\x1a\x00\x00\x00\x00\x00\x00\x00mindio\x00', 49)]
EXPECTED_RES_WITHOUT_DICT = (b'data_buff_test_strdata.pkl:0,10;data/0:11,20'
                             b'\x12\x00\x00\x00\x00\x00\x00\x00\x1a\x00\x00\x00\x00\x00\x00\x00mindio\x00')


@pytest.mark.parametrize(
    "tensors_dict, expected_result",
    [
        pytest.param(TENSOR_DICTS, EXPECTED_RES_WITH_DICT, id='get write content with tensors dict'),
        pytest.param(None, EXPECTED_RES_WITHOUT_DICT, id='get write content without tensors dict'),
    ])
def test_get_write_content(tensors_dict, expected_result):
    if tensors_dict:
        for key in reversed(tensors_dict):
            expected_result.insert(1, (tensors_dict[key].data_ptr(), tensors_dict[key].nbytes()))

    data_buff = b'data_buff_test_str'
    record_map_bytes = b'data.pkl:0,10;data/0:11,20'
    torch_save_helper = _TorchSaveHelp()
    write_content = torch_save_helper.get_write_content(data_buff, tensors_dict, record_map_bytes)
    assert write_content == expected_result


@pytest.mark.parametrize(
    "tensors_dict, expected_result",
    [
        pytest.param(TENSOR_DICTS, EXPECTED_RES_WITH_DICT, id='get write content with tensors dict'),
        pytest.param(None, EXPECTED_RES_WITHOUT_DICT, id='get write content without tensors dict'),
    ])
def test_get_write_content(tensors_dict, expected_result):
    if tensors_dict:
        for key in reversed(tensors_dict):
            expected_result.insert(1, (tensors_dict[key].data_ptr(), tensors_dict[key].nbytes()))

    data_buff = b'data_buff_test_str'
    record_map_bytes = b'data.pkl:0,10;data/0:11,20'
    torch_save_helper = _TorchSaveHelp()
    write_content = torch_save_helper.get_write_content(data_buff, tensors_dict, record_map_bytes)
    assert write_content == expected_result


@pytest.mark.parametrize(
    "record_map, data_pkl_bytes, map_location, expect",
    [
        pytest.param({'data.pkl': (0, 15)}, b'\x80\x02X\x05\x00\x00\x00helloq\x00.', None,
                     'hello', id='unmarshal normal data'),
        pytest.param({'data.pkl': (0, 141), 'data/0': (141, 8)},
                     b'\x80\x02ctorch._utils\n_rebuild_tensor_v2\nq\x00((X\x07\x00\x00\x00storageq\x01ctorch'
                     b'\nLongStorage\nq\x02X\x01\x00\x00\x000q\x03X\x03\x00\x00\x00cpuq\x04K\x01tq\x05QK\x00))'
                     b'\x89ccollections\nOrderedDict\nq\x06)Rq\x07tq\x08Rq\t.',
                     None, torch.Tensor, id='unmarshal tensor'),
    ])
def test_unmarshal_checkpoint(mocker, monkeypatch, record_map, data_pkl_bytes, map_location, expect):
    memfs_reader = mocker.MagicMock()
    memfs_reader.file = mocker.MagicMock()
    monkeypatch.setattr(memfs_reader.file, "multi_read", lambda *args, **kwargs: -1)
    serialize = SerializationMixin()
    ret = serialize.unmarshal_checkpoint(memfs_reader, record_map, data_pkl_bytes, None)
    if isinstance(ret, str):
        assert ret == expect
    else:
        assert isinstance(ret, expect)


@pytest.mark.parametrize(
    "map_location",
    [
        pytest.param({}, id='map_location is dict'),
        pytest.param('test', id='map_location is str'),
        pytest.param(torch.device('cpu'), id='map_location is torch.device'),
        pytest.param(10, id='map_location is int'),
    ]
)
def test_get_restore_func(map_location):
    assert isinstance(_get_restore_func(map_location), types.FunctionType)
