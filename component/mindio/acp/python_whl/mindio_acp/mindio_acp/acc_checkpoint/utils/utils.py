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

import random
import time
import weakref
from functools import wraps

import torch


from mindio_acp.common import mindio_logger

logging = mindio_logger.LOGGER


def time_used(func):
    def wrapper(*args, **kwargs):
        rank = -1
        if torch.distributed.is_initialized():
            rank = torch.distributed.get_rank()
        start_time = time.time()
        result = func(*args, **kwargs)
        end_time = time.time()
        logging.debug(f'rank: {rank}, func_name: {func.__name__}, took {end_time - start_time:.6f} seconds to execute')
        return result

    return wrapper


def print_rank_0(message):
    """If distributed is initialized, print only on rank 0."""
    if torch.distributed.is_initialized():
        if torch.distributed.get_rank() == 0:
            logging.info(message)
    else:
        logging.info(message)


def retry(wait_min=100, wait_max=800):
    """
    A retry function. The number of retry times is attempts.
    The waiting time for each retry is random. The default retry
    interval is [100, 200], [200, 400], [400, 800].

    :param wait_min: Minimum retry wait time. Default value is 100ms.
    :param wait_max: Maximum retry wait time. Default value is 800ms.
    """

    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            attempt = 0
            while True:
                try:
                    result = func(*args, **kwargs)
                    return result
                except Exception as e:
                    lower = wait_min * (2 ** attempt)
                    if lower >= wait_max:
                        logging.warning(str(e))
                        break
                    upper = lower * 2
                    attempt += 1
                    current_delay = random.uniform(lower, upper)
                    time.sleep(current_delay / 1000)

        return wrapper

    return decorator


class SingletonBase(object):
    """
    The Singleton class is implemented by `base class`.

    Examples:
        class Singleton(SingletonBase):
            def __init__(self):
                ...

            def some_business_logic(self):
                ...
    """

    _instance = None

    def __new__(cls, *args, **kw):
        if not cls._instance:
            orig = super(SingletonBase, cls)
            cls._instance = orig.__new__(cls, *args, **kw)
        return cls._instance


class SingletonMeta(type):
    """
    The Singleton class is implemented by `metaclass`.

    Examples:
        class Singleton(metaclass=SingletonMeta):
            def __init__(self, arg1: int, arg2: float, ...):
                ...

            def some_business_logic(self):
                ...
    """

    _instances = weakref.WeakValueDictionary()

    def __call__(cls, *args):
        """
        Possible changes to the value of the `__init__` argument do not affect
        the returned instance.
        """
        if args in cls._instances:
            return cls._instances[args]
        else:
            obj = super().__call__(*args)
            cls._instances[args] = obj
            return obj
