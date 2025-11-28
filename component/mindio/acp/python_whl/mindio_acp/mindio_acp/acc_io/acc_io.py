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
from typing import Union, Dict
from collections import namedtuple

from mindio_acp.common import mindio_logger
from mindio_acp.acc_io.mindio_help import torch_initialize_helper
from mindio_acp.acc_io.mindio_help import torch_register_checker_helper
from mindio_acp.acc_io.mindio_help import torch_load_helper
from mindio_acp.acc_io.mindio_help import torch_multi_save_helper
from mindio_acp.acc_io.mindio_help import torch_preload_helper
from mindio_acp.acc_io.mindio_help import torch_save_helper
from mindio_acp.acc_io.mindio_help import torch_wait_flush_helper

logging = mindio_logger.LOGGER
DEFAULT_MAX_FILE_SIZE = 1024 * 1024 * 1024
ONE_HOUR_SEC = 60 * 60


def initialize(server_info: Dict[str, str] = None) -> int:
    """
    initialize mindio_acp module man manually.
    :param server_info: client launch server need default config info.
    :return: 0 for success,
             -1 for failed.
    """

    return torch_initialize_helper(server_info)


def register_checker(callback, check_dict, user_context, timeout_sec):
    """
    Register the callback function to C++ thread.
    :param callback: The function to be called, the callback function first param is result.
    :param check_dict: The dict to be checked, the key is path, the value is file nums in the key path.
    :param user_context: user input context as callback function second parameter.
    :param timeout_sec: Maximum execution time of the check thread.
    :return: None for failed, 1 for success
    """
    if not isinstance(timeout_sec, int) or timeout_sec < 1:
        logging.error("timeout_sec (%d) is abnormal. It should be int and greater than zero.", timeout_sec)
        return None
    if timeout_sec > ONE_HOUR_SEC:
        logging.error("timeout_sec (%d) is too big. It should be less than one hour.", timeout_sec)
        return None
    return torch_register_checker_helper(callback, check_dict, user_context, timeout_sec)


def multi_save(obj: Union[Dict, bytes], path_list: list):
    """
    Use mindio format to save torch ckpt from memfs or real paths
    :param obj: obj to save.
    :param path_list: file path list to save.
    :return: None for failed
             1 for memfs multi save success
             2 for fwrite multi save success
             0 for native torch save success
    """
    return torch_multi_save_helper(obj, path_list)


def save(obj: Union[Dict, bytes], path: str, open_way='memfs') -> int:
    """
    Use mindio format to save torch ckpt from memfs or real path
    Args:
        obj (dict/bytes): obj to save.
        path (str): file path to save.
        open_way (str): the way used to save.

    Returns:
        int: save checkpoint marker
    """
    path = os.path.realpath(path)
    return torch_save_helper(obj, path, open_way)


def load(path, open_way='memfs', map_location=None) -> Dict:
    """
    Use mindio format to load torch ckpt from memfs or real path
    Args:
        path (str): file path to load.
        open_way (str): the way used to load.
        map_location (str): specifying how to remap storage locations, only support 'cpu' now

    Returns: pytorch object from path, same as torch.load() return.
    """
    if map_location != 'cpu' and map_location:
        raise ValueError("param map_location only support value 'cpu'")
    path = os.path.realpath(path)
    return torch_load_helper(path, open_way, map_location)


def preload(*path) -> int:
    """
    Use mindio format async load torch ckpt from real path, and save it into memfs
    Args:
        *path: file path to load.

    Returns: 0 as succeeded, 1 as failed
    """
    return torch_preload_helper(*path)


def flush() -> int:
    """
    Wait mindio memfs background flush tasks finish
    Returns: 0 as succeeded, 1 as failed
    """
    return torch_wait_flush_helper()
