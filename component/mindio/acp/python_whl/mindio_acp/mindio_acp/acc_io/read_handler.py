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
from __future__ import annotations

import os
import sys
import pickle
import struct
from abc import ABCMeta, abstractmethod
from typing import Tuple

import torch
from mindio_acp.acc_io.serialization import SerializationMixin
from mindio_acp.common import mindio_logger

logging = mindio_logger.LOGGER

FILE_FLAG_LENGTH = 7
VERSION_FLAG_LENGTH = 6
RECORD_MAP_SIZE = 8
RECORD_MAP_START = 8
CHECKPOINT_TAIL_OFFSET = RECORD_MAP_SIZE + RECORD_MAP_START + FILE_FLAG_LENGTH + VERSION_FLAG_LENGTH
OCKIO_TAIL_OFFSET = RECORD_MAP_SIZE + RECORD_MAP_START + FILE_FLAG_LENGTH

FILE_TYPE_ACCACP = 1
FILE_TYPE_OCKIO = 0
FILE_TYPE_UNKHOWN = -1

OCKIO_TAIL_BYTE = b'mindio\x00'


class ReadHandler(metaclass=ABCMeta):
    """
    The ReadHandler interface declares a method for building the chain of handlers.
    It also declares a method for executing a request.
    """

    @abstractmethod
    def set_next(self, handler: ReadHandler) -> ReadHandler:
        pass

    @abstractmethod
    def handle(self) -> object:
        pass


class ReadHandlerCtx(ReadHandler, SerializationMixin):
    """
    The default chaining behavior can be implemented inside a base handler class.
    """
    _next_handler: ReadHandler = None

    def __init__(self, path: str, map_location):
        self.path = path
        self.map_location = map_location
        self.file = None

    @abstractmethod
    def __enter__(self):
        pass

    @abstractmethod
    def __exit__(self, exc_type, exc_val, exc_tb):
        pass

    @abstractmethod
    def read(self):
        pass

    def set_next(self, handler: ReadHandler) -> ReadHandler:
        """
        Returning a handler from here will let us link handlers in a
        convenient way like this:
        memfs_read.set_next(f_read).set_next(torch_read)
        """
        self._next_handler = handler
        return handler

    @abstractmethod
    def handle(self) -> object:
        if self._next_handler:
            return self._next_handler.handle()

        raise RuntimeError("No read policy is available")

    def read_file(self, offset, size) -> bytes:
        malloc_size = min(self.file.size() - offset, size)
        buffer = bytes(malloc_size)
        ret = self.file.read(buffer, size, offset)
        if ret == -1:
            raise RuntimeError("read file failed.")
        return buffer

    def get_tail_bytes(self) -> Tuple[int, bytes]:
        file_size = self.file.size()
        tail_bytes = self.read_file(file_size - CHECKPOINT_TAIL_OFFSET, CHECKPOINT_TAIL_OFFSET)
        if tail_bytes[-FILE_FLAG_LENGTH:] == OCKIO_TAIL_BYTE:
            tail_bytes = self.read_file(file_size - OCKIO_TAIL_OFFSET, OCKIO_TAIL_OFFSET)
            return FILE_TYPE_OCKIO, tail_bytes
        return FILE_TYPE_UNKHOWN, None

    def get_record_bytes(self, tail_bytes) -> bytes:
        m_record_start_bytes = tail_bytes[:RECORD_MAP_START]
        m_record_size_bytes = tail_bytes[RECORD_MAP_START:RECORD_MAP_START + RECORD_MAP_SIZE]
        m_start = struct.unpack('Q', m_record_start_bytes)
        m_size = struct.unpack('Q', m_record_size_bytes)
        if len(m_start) == 0 or len(m_size) == 0:
            raise ValueError("invalid file format when decode uint64")
        m_start = m_start[0]
        m_size = m_size[0]
        # read record map
        record_bytes = self.read_file(m_start, m_size)
        return record_bytes

    def get_record_map(self, file_type, record_bytes):
        record_map = pickle.loads(record_bytes)
        return record_map

    def get_read_result(self, record_map):
        data_pkl_start, data_pkl_size = record_map["data.pkl"]
        # read data.pkl
        data_pkl_bytes = self.read_file(data_pkl_start, data_pkl_size)
        result = self.unmarshal_checkpoint(self, record_map, data_pkl_bytes, map_location=self.map_location)
        return result

    def direct_read_legacy_bytes(self, record_map):
        record_bytes = self.read_file(0, record_map['bytes_length'])
        return record_bytes


class MemFsReadHandler(ReadHandlerCtx):
    def __repr__(self):
        return f"MemFsReadHandler(path={os.path.basename(self.path)})"

    def __enter__(self):
        from mindio_acp.acc_io.write_handler import readable_file
        self.file = readable_file(self.path)
        ret = self.file.open()
        if ret < 0:
            raise RuntimeError(-1, f'[memfs] read file failed: {ret}, using nds to open.')
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.file.close()

    def read(self):
        file_type, tail_bytes = self.get_tail_bytes()
        if file_type == FILE_TYPE_UNKHOWN:
            raise RuntimeError(-1, f'[memfs] read file failed, using nds to open.')
        record_bytes = self.get_record_bytes(tail_bytes)
        record_map = self.get_record_map(file_type, record_bytes)
        if 'mindio_save_data_type' in record_map and record_map['mindio_save_data_type'] == 'directly_bytes':
            return self.direct_read_legacy_bytes(record_map)
        result = self.get_read_result(record_map)
        return result

    def handle(self) -> object:
        try:
            with self:
                result = self.read()
        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle()
        return result


class NdsReadHandler(ReadHandlerCtx, SerializationMixin):
    def __repr__(self):
        return f"NdsReadHandler(path={os.path.basename(self.path)})"

    def __enter__(self):
        from mindio_acp.acc_io.write_handler import nds_readable_file
        self.file = nds_readable_file(self.path)
        ret = self.file.open()
        if ret < 0:
            raise RuntimeError(-1, f'[nds] read file failed: {ret}, using fopen to open.')
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.file.close()

    def read(self):
        file_type, tail_bytes = self.get_tail_bytes()
        if file_type == FILE_TYPE_UNKHOWN:
            raise RuntimeError(-1, f'[nds] read file failed, using fopen to open.')
        record_bytes = self.get_record_bytes(tail_bytes)
        record_map = self.get_record_map(file_type, record_bytes)
        if 'mindio_save_data_type' in record_map and record_map['mindio_save_data_type'] == 'directly_bytes':
            return self.direct_read_legacy_bytes(record_map)
        result = self.get_read_result(record_map)
        return result

    def handle(self) -> object:
        try:
            with self:
                result = self.read()
        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle()
        return result


class FReadHandler(ReadHandlerCtx, SerializationMixin):
    def __repr__(self):
        return f"FReadHandler(path={os.path.basename(self.path)})"

    def __enter__(self):
        from mindio_acp.acc_io.write_handler import freadable_file
        self.file = freadable_file(self.path)
        ret = self.file.open()
        if ret < 0:
            raise RuntimeError(-1, f'[fopen] read file failed: {ret}, using torch.load to open.')
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.file.close()

    def read(self):
        file_type, tail_bytes = self.get_tail_bytes()
        if file_type == FILE_TYPE_UNKHOWN:
            raise RuntimeError(-1, f'[fopen] read file failed, using torch.load to open.')
        record_bytes = self.get_record_bytes(tail_bytes)
        record_map = self.get_record_map(file_type, record_bytes)
        if 'mindio_save_data_type' in record_map and record_map['mindio_save_data_type'] == 'directly_bytes':
            return self.direct_read_legacy_bytes(record_map)
        result = self.get_read_result(record_map)
        return result

    def handle(self) -> object:
        try:
            with self:
                result = self.read()
        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle()
        return result


class TorchReadHandler(ReadHandlerCtx):
    def __repr__(self):
        return f"TorchReadHandler(path={os.path.basename(self.path)})"

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        pass

    def read(self):
        return torch.load(self.path, map_location=self.map_location)

    def handle(self) -> object:
        try:
            with self:
                result = self.read()
        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle()
        return result


def create_read_chain(path: str, open_way: str, map_location=None) -> ReadHandler:
    memfs_read = MemFsReadHandler(path, map_location)
    nds_read = NdsReadHandler(path, map_location)
    f_read = FReadHandler(path, map_location)
    torch_read = TorchReadHandler(path, map_location)
    memfs_read.set_next(nds_read).set_next(f_read).set_next(torch_read)
    return memfs_read if open_way == "memfs" else nds_read
