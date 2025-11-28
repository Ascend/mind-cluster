#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
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
from typing import Dict
from mindio_acp.c2python_api import readable_file
from mindio_acp.c2python_api import freadable_file
from mindio_acp.c2python_api import writeable_file
from mindio_acp.c2python_api import fwriteable_file
from mindio_acp.acc_io.mindio_help import torch_initialize_helper
from mindio_acp.common.utils import get_relative_path
from mindio_acp.common import mindio_logger
from mindio_acp.launch_server_conf.default_memfs_conf import default_server_info
from mindio_acp.launch_server_conf.launch_server_param import ockiod_path, server_worker_dir
logging = mindio_logger.LOGGER


class _ReadableFileWrapper:
    def __init__(self, path: str):
        self._file = readable_file(path=path)
        self._path = path
        ret = self._file.open()
        if ret != 0:
            logging.warning(f'[mindio_acp] open file failed, using [fopen] open file path:{get_relative_path(self._path)}')
            self._file = freadable_file(path=self._path)
            ret = self._file.open()
            if ret != 0:
                logging.error(f'[fopen] file failed, path: {get_relative_path(self._path)}')
                raise RuntimeError(ret, f'fopen read file return error')

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        return self.close()

    def read(self, offset=0, count=-1):
        if self._file is None:
            raise RuntimeError(-1, f'_file is None')

        if count < 0:
            count = self._file.size()

        if offset < 0:
            raise RuntimeError(-1, 'offset is less than zero.')

        if count == 0 or offset == self._file.size() or offset + count > self._file.size():
            raise RuntimeError(-1, 'read params invalid.')

        data = bytes(count)
        ret = self._file.read(data, count, offset)
        if ret < 0:
            if isinstance(self._file, readable_file):
                logging.warning(f'[memfs] read file path: {get_relative_path(self._path)} failed: {ret}, use fopen')
                self._file.close()
                self._file = freadable_file(path=self._path)
                ret = self._file.open()
                if ret == 0:
                    ret = self._file.read(data, count, offset)
                    if ret == -1:
                        logging.error(f'[fopen] read file path: {get_relative_path(self._path)} failed: {ret}')
                        raise RuntimeError(-1, f'fopen read data failed: {ret}')
                    return data
                else:
                    raise RuntimeError(-1, f'fopen read open failed: {ret}')
            else:
                logging.error(f'[fopen] read file path: {get_relative_path(self._path)} failed: {ret}')
                raise RuntimeError(-1, f'memfs read data failed: {ret}')
        return data

    def close(self):
        if self._file is None:
            return None
        ret = self._file.close()
        if ret != 0:
            raise RuntimeError(-1, f'close file failed: {ret}')
        self._file = None
        return ret


class _WriteableFileWrapper:
    def __init__(self, path: str, mode: int = 0o600):
        self._file = writeable_file(path=path, mode=mode)
        self._path = path
        self._mode = mode
        self._success_datas = []
        ret = self._file.create()
        if ret != 0:
            logging.warning(f'[mindio_acp] creat file failed, using [fopen] creat file path: '
                            f'{get_relative_path(self._path)}')
            self._file = fwriteable_file(path=self._path, mode=self._mode)
            ret = self._file.create()
            if ret != 0:
                logging.error(f'[fopen] create file failed, path: {get_relative_path(self._path)}')
                raise RuntimeError(ret, f'[fopen] create file failed')

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        return self.close()

    def write(self, data: bytes):
        if self._file is None:
            raise RuntimeError(-1, f'_file is None')

        ret = self._file.write(data, len(data))
        if ret < 0:
            if isinstance(self._file, writeable_file):
                logging.warning(f'[memfs] write file path: {get_relative_path(self._path)} failed: {ret}, use fopen')
                self._file.close()
                self._file = fwriteable_file(path=self._path, mode=self._mode)
                ret = self._file.create()
                if ret == 0:
                    for success_data in self._success_datas:
                        ret = self._file.write(success_data, len(success_data))
                        if ret < 0:
                            logging.error(f'[fopen] write data failed: {ret}')
                            raise RuntimeError(-1, f'[fopen] write data failed: {ret}')

                    self._success_datas.clear()
                    ret = self._file.write(data, len(data))
                    if ret < 0:
                        logging.error(f'[fopen] write data failed: {ret}')
                        raise RuntimeError(-1, f'[fopen] write data failed: {ret}')
                else:
                    raise RuntimeError(-1, f'fopen write create failed: {ret}')
            else:
                logging.error(f'[fopen] write file path: {get_relative_path(self._path)} failed: {ret}')
                raise RuntimeError(-1, f'memfs write data failed: {ret}')
        else:
            if isinstance(self._file, writeable_file):
                self._success_datas.append(data)
        return ret

    def drop(self):
        if self._file is None:
            return None
        return self._file.drop()

    def close(self):
        if self._file is None:
            return None
        ret = self._file.close()
        if ret != 0:
            raise RuntimeError(-1, f'close file failed: {ret}')
        self._file = None
        return ret


def open_file(path: str):
    ret = torch_initialize_helper(None)
    if ret != 0:
        logging.warning(f"[mindio_acp] default initialize failed.")
    return _ReadableFileWrapper(path=os.path.realpath(path))


def create_file(path: str, mode: int = 0o600):
    ret = torch_initialize_helper(None)
    if ret != 0:
        logging.warning(f"[mindio_acp] default initialize failed.")
    return _WriteableFileWrapper(path=os.path.realpath(path), mode=mode)

