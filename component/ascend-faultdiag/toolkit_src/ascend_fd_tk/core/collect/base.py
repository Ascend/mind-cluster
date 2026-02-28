#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import abc
from functools import wraps
from typing import Callable, Any

from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.utils import logger

_LOGGER = logger.DIAG_LOGGER


class Collector(abc.ABC):

    @abc.abstractmethod
    async def collect(self) -> JsonObj:
        pass

    @abc.abstractmethod
    async def get_id(self) -> str:
        pass


def log_collect_async_event(error_msg="", raise_exception=True):
    """
    类成员函数装饰器：自动添加函数开始/结束日志，并包含 self.host 信息

    """

    def decorator(func: Callable) -> Callable:
        @wraps(func)  # 保留原函数的元信息（名称、文档字符串等）
        async def wrapper(self, *args, **kwargs) -> Any:
            log = _LOGGER
            try:
                collector_id = await self.get_id()
            except Exception as e:
                log.error(e)
                collector_id = ""
            # 函数开始日志：包含主机、函数名、参数
            class_name = self.__class__.__name__
            func_name = func.__name__
            log.info(f"start collecting [{collector_id}] by class: [{class_name}] func: [{func_name}].")
            try:
                # 执行原函数并获取返回值
                result = await func(self, *args, **kwargs)
                # 函数成功结束日志
                log.info(f"collection of [{collector_id}] by "
                         f"class: [{class_name}] func: [{func_name}] is completed.")
                return result
            except Exception as e:
                # 函数异常日志（记录错误级别）
                log.error(f"failed to [{collector_id}] by "
                          f"class: [{class_name}] func: [{func_name}]. message: [{error_msg}] error: [{e}]",
                          exc_info=raise_exception)
                if raise_exception:
                    raise e  # 重新抛出异常，不影响原有异常处理逻辑

        return wrapper

    return decorator
