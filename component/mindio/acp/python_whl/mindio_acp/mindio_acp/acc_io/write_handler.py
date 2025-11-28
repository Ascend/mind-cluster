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
from abc import ABCMeta, abstractmethod
from typing import Union, Dict, List

import torch
from mindio_acp.common import mindio_logger
from mindio_acp.common.utils import get_relative_path
from mindio_acp.common.check_process import check_process
logging = mindio_logger.LOGGER


def import_mindio_sdk_api():
    global initialize, readable_file, nds_readable_file, freadable_file, \
        writeable_file, fwriteable_file, preload, link, register_checker, check_background_task, un_initialize
    from mindio_acp.c2python_api import (initialize, readable_file, nds_readable_file, freadable_file, \
        writeable_file, fwriteable_file, preload, link, register_checker, check_background_task, un_initialize)


class WriteHandler(metaclass=ABCMeta):
    """
    The WriteHandler interface declares a method for building the chain of handlers.
    It also declares a method for executing a request.
    """

    @abstractmethod
    def set_next(self, handler: WriteHandler) -> WriteHandler:
        pass

    @abstractmethod
    def handle(self, write_content: Union[List, bytes]) -> int:
        pass


class WriteHandlerCtx(WriteHandler):
    """
    The default chaining behavior can be implemented inside a base handler class.
    """
    mode = 0o600
    _next_handler: WriteHandler = None

    def __init__(self, path: str):
        self.path = path
        self.file = None

    @abstractmethod
    def __enter__(self):
        pass

    @abstractmethod
    def __exit__(self, exc_type, exc_val, exc_tb):
        pass

    @abstractmethod
    def write(self, write_content: Union[List, bytes]):
        pass

    def set_next(self, handler: WriteHandler) -> WriteHandler:
        """
        Returning a handler from here will let us link handlers in a
        convenient way like this:
        memfs_write.set_next(f_write)
        """
        self._next_handler = handler
        return handler

    @abstractmethod
    def handle(self, write_content: Union[List, bytes]) -> int:
        if self._next_handler:
            return self._next_handler.handle(write_content)

        raise RuntimeError("No write policy is available")


class MemFsWriteHandler(WriteHandlerCtx):
    saver_marker = 1

    def __repr__(self):
        return f"MemFsWriteHandler(path={os.path.basename(self.path)})"

    def __enter__(self):
        self.file = writeable_file(self.path, self.mode)
        ret = self.file.create()
        if ret < 0:
            raise RuntimeError(-1, f'[mindio_acp] open file path: {get_relative_path(self.path)} failed: {ret}, using '
                                   f'fopen to save.')
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_type:
            self.file.drop()
            return
        self.file.close()

    def write(self, write_content: Union[List, bytes]):
        if isinstance(write_content, bytes):
            ret = self.file.write(write_content, len(write_content))
        else:
            ret = self.file.write_list(write_content)
        if ret < 0:
            raise RuntimeError(-1, f'[mindio_acp] write file path: {get_relative_path(self.path)} failed: {ret}, using '
                                   f'fopen to save.')

    def handle(self, write_content: Union[List, bytes]) -> int:
        try:
            if not check_process():
                raise RuntimeError(-1, '[mindio_acp] ockiod service not available.')
            with self:
                self.write(write_content)
        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle(write_content)
        return self.saver_marker


class FWriteHandler(WriteHandlerCtx):
    saver_marker = 2

    def __repr__(self):
        return f"FWriteHandler(path={os.path.basename(self.path)})"

    def __enter__(self):
        self.file = fwriteable_file(self.path, self.mode)
        ret = self.file.create()
        if ret < 0:
            raise RuntimeError(-1, f'[mindio_acp] open file path param is invalid, path: {get_relative_path(self.path)}'
                                   f'using torch.save""to save.')
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if exc_type:
            self.file.drop()
            return
        self.file.close()

    def write(self, write_content: Union[List, bytes]):
        if isinstance(write_content, bytes):
            ret = self.file.write(write_content, len(write_content))
        else:
            ret = self.file.write_list(write_content)
        if ret < 0:
            raise RuntimeError(-1, f'[mindio_acp] write file path: {get_relative_path(self.path)} failed: {ret}, '
                                   f'using torch.save to save.')

    def handle(self, write_content: Union[List, bytes]) -> int:
        try:
            with self:
                self.write(write_content)
        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle(write_content)
        return self.saver_marker


class TorchWriteHandler(WriteHandlerCtx):
    saver_marker = 0

    def __init__(self, ckpt_obj: Union[Dict, bytes], path: str):
        self.ckpt_obj = ckpt_obj
        super().__init__(path)

    def __repr__(self):
        return f"TorchWriteHandler(path={os.path.basename(self.path)})"

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        pass

    def write(self, _: Union[List, bytes]):
        """
        Use the native torch save to write ckpt.
        """
        torch.save(self.ckpt_obj, self.path)

    def handle(self, write_content: Union[List, bytes]) -> int:
        try:
            with self:
                self.write(write_content)
        except Exception as e:
            logging.warning("%s error is %s", repr(self), e)
            return super().handle(write_content)
        return self.saver_marker


def create_write_chain(ckpt_obj: Union[Dict, bytes], path: str, open_way: str) -> WriteHandler:
    memfs_write = MemFsWriteHandler(path)
    f_write = FWriteHandler(path)
    torch_write = TorchWriteHandler(ckpt_obj, path)
    memfs_write.set_next(f_write).set_next(torch_write)
    return memfs_write if open_way == "memfs" else f_write