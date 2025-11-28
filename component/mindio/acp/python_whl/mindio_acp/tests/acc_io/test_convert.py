#!/usr/bin/env python
# coding=utf-8
#  Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
import pickle

import torch

import mindio_acp


def test_convert(tmp_path):
    mindio_file = tmp_path / "mindio_format.pt"
    # mindio_acp save the data:
    tt = {"name-0": torch.ones([1, 2]), "name-1": torch.ones([2, 3])}
    flag = b'mindio\00'
    map_size = b'?\x00\x00\x00\x00\x00\x00\x00'
    map_start = b'\x0e\x01\x00\x00\x00\x00\x00\x00'
    record_map = pickle.dumps({'data.pkl': (0, 238), 'data/0': (238, 8), 'data/1': (246, 24)})
    tensor1 = b'\x00\x00\x80?\x00\x00\x80?\x00\x00\x80?\x00\x00\x80?\x00\x00\x80?\x00\x00\x80?'
    tensor0 = b'\x00\x00\x80?\x00\x00\x80?'
    data_pkl = (b'\x80\x02}q\x00(X\x06\x00\x00\x00name-0q\x01ctorch._utils\n_rebuild_tensor_v2\nq\x02((X\x07\x00\x00'
                b'\x00storageq\x03ctorch\nFloatStorage\nq\x04X\x01\x00\x00\x000q\x05X\x03\x00\x00\x00cpuq\x06K\x02'
                b'tq\x07QK\x00K\x01K\x02\x86q\x08K\x02K\x01\x86q\t\x89ccollections\nOrderedDict\nq\n)Rq\x0btq\x0cRq\r'
                b'X\x06\x00\x00\x00name-1q\x0eh\x02((h\x03h\x04X\x01\x00\x00\x001q\x0fh\x06K\x06tq\x10QK\x00K\x02'
                b'K\x03\x86q\x11K\x03K\x01\x86q\x12\x89h\n)Rq\x13tq\x14Rq\x15u.')

    mindio_data = b''.join([data_pkl, tensor0, tensor1, record_map, map_start, map_size, flag])
    mindio_file.write_bytes(mindio_data)

    torch_file = str(tmp_path / "torch_format.pt")
    res = mindio_acp.convert(str(mindio_file), torch_file)
    assert res == 0

    torch_data = torch.load(torch_file, map_location="cpu")
    assert len(torch_data.keys()) == len(tt.keys())
    for key in tt:
        assert torch.equal(tt[key], torch_data[key])
