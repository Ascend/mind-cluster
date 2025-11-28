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
import struct
import pickle

from torch import serialization

from mindio_acp.common import utils
from mindio_acp.common.utils import FileCheckPolicy as fcp
from mindio_acp.common.mindio_logger import LOGGER

FLAG_MINDIO = b'mindio\x00'
FILE_FLAG_LENGTH = 7
MAP_SIZE_LENGTH = 8
MAP_START_LENGTH = 8


def convert(src, dst) -> int:
    """
    convert file format from mindio to torch
    Args:
        src: mindio format saved file
        dst: torch format saved file

    Returns:
        0 for success,
       -1 for failed
    """
    if not utils.file_path_check(src):
        return -1
    if not utils.file_path_check(dst, fcp.CHECK_NOT_EMPTY | fcp.CHECK_LENGTH | fcp.CHECK_SYMBOLIC_LINK):
        return -1
    src = os.path.realpath(src)
    dst = os.path.realpath(dst)
    if src == dst:
        LOGGER.error("The src file path and dst file path can't be the same.")
        return -1

    with open(src, 'rb') as f:
        flag = read_flag(f)
        if flag != FLAG_MINDIO:
            LOGGER.error("The flag is incorrect, should = %s, real = %s", FLAG_MINDIO, flag)
            return -1

        start, size = read_tails(f)
        record_map = read_record_map(f, start, size)

        with serialization._open_zipfile_writer(dst) as opened_zipfile:
            for key, (start, size) in record_map.items():
                opened_zipfile.write_record(key, read_bytes(f, start, size), size)
    return 0


def read_flag(f):
    f.seek(-FILE_FLAG_LENGTH, os.SEEK_END)
    flag = f.read(FILE_FLAG_LENGTH)
    return flag


def read_uint64(f, offset, length):
    f.seek(offset, os.SEEK_END)
    size_b = f.read(length)
    size = struct.unpack('Q', size_b)
    if len(size) == 0:
        raise ValueError("invalid file format when decode uint64")
    size = size[0]
    return size


def read_tails(f):
    size = read_uint64(f, -FILE_FLAG_LENGTH - MAP_SIZE_LENGTH, MAP_SIZE_LENGTH)
    start = read_uint64(f, -FILE_FLAG_LENGTH - MAP_SIZE_LENGTH - MAP_START_LENGTH, MAP_START_LENGTH)
    return start, size


def read_record_map(f, start, size):
    record_bytes = read_bytes(f, start, size)
    record_map = pickle.loads(record_bytes)
    return record_map


def read_bytes(f, start, size):
    f.seek(start, os.SEEK_SET)
    return f.read(size)
