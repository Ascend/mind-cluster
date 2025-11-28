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
from typing import Union, Dict, List
import torch
from mindio_acp.common import mindio_logger
from mindio_acp.common.check_process import check_process
from mindio_acp.acc_io.write_handler import WriteHandler, MemFsWriteHandler, FWriteHandler, TorchWriteHandler
from mindio_acp.c2python_api import initialize, readable_file, nds_readable_file, freadable_file, \
    writeable_file, fwriteable_file, link

logging = mindio_logger.LOGGER


class MemFsMultiWriteHandler(MemFsWriteHandler):
    def __init__(self, path: list):
        super().__init__(None)
        self.path_list = path
        self.format_path_list = [os.path.basename(path) for path in self.path_list]
        self.path = None
        self.file = None

    def __repr__(self):
        return f"MemFsMultiWriteHandler(path_list={self.format_path_list})"

    def handle(self, write_content: Union[List, bytes]) -> int:
        try:
            if not check_process():
                raise RuntimeError(-1, '[mindio_acp] ockiod service not available.')
            self.path = os.path.realpath(self.path_list[0])
            with self:
                self.write(write_content)
                self.flush()
                for path in self.path_list[1:]:
                    link_ret = link(os.path.realpath(self.path_list[0]), os.path.realpath(path))
                    if link_ret != 0:
                        raise RuntimeError(-1,
                                           f'[mindio_acp] link data path: {self.format_path_list[0]} failed.')
        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle(write_content)
        return self.saver_marker

    def flush(self) -> int:
        ret = self.file.flush()
        if ret != 0:
            raise RuntimeError(-1, f'[mindio_acp] flush data path: {self.format_path_list[0]} failed: {ret}, '
                                   f'using fopen to save.')
        return ret


class FMultiWriteHandler(FWriteHandler):

    def __init__(self, path: list):
        super().__init__(None)
        self.path_list = path
        self.format_path_list = [os.path.basename(path) for path in self.path_list]
        self.path = None
        self.file = None

    def __repr__(self):
        return f"FMultiWriteHandler(path_list={self.format_path_list})"

    def handle(self, write_content: Union[List, bytes]) -> int:
        try:
            for path in self.path_list:
                self.path = os.path.realpath(path)
                with self:
                    self.write(write_content)

        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle(write_content)

        return self.saver_marker


class TorchMultiWriteHandler(TorchWriteHandler):

    def __init__(self, ckpt_obj: Union[Dict, bytes], path: list):
        super().__init__(None, None)
        self.ckpt_obj = ckpt_obj
        self.path_list = path
        self.format_path_list = [os.path.basename(path) for path in self.path_list]

    def __repr__(self):
        return f"TorchMultiWriteHandler(path_list={self.format_path_list})"

    def torch_write(self, _: Union[List, bytes], path):
        """
        Use the native torch save to write ckpt.
        """
        torch.save(self.ckpt_obj, path)

    def handle(self, write_content: Union[List, bytes]) -> int:
        try:
            for path in self.path_list:
                with self:
                    self.torch_write(write_content, os.path.realpath(path))
        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle(write_content)
        return self.saver_marker


def create_multi_write_chain(ckpt_obj: Union[Dict, bytes], path: list, open_way: str) -> WriteHandler:
    memfs_write = MemFsMultiWriteHandler(path)
    f_write = FMultiWriteHandler(path)
    torch_write = TorchMultiWriteHandler(ckpt_obj, path)
    memfs_write.set_next(f_write).set_next(torch_write)
    return memfs_write if open_way == "memfs" else f_write
