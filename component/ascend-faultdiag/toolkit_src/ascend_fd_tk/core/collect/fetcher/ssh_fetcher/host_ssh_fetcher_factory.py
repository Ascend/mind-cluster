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

"""按芯片代际创建主机 SSH 采集器实例。"""

from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.host_ssh_fetcher import HostSshFetcher
from ascend_fd_tk.core.context.register import HOST_FETCHER_REGISTRY
from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.utils.executors import AsyncSSHExecutor


def create_host_ssh_fetcher(
    executor: AsyncSSHExecutor, generation: NpuType, npu_mapping_cache: dict = None
) -> HostSshFetcher:
    """根据芯片代际返回对应的主机 SSH 采集器。

    通过 HOST_FETCHER_REGISTRY 注册表查找，新增代际无需修改本函数。
    npu_mapping_cache 为代际探测阶段已解析的 npu_mapping，注入后 fetch_npu_mapping 直接复用。
    """
    # 触发 A5 fetcher 模块导入，确保其装饰器完成注册
    import ascend_fd_tk.core.collect.fetcher.ssh_fetcher.host_ssh_fetcher_a5  # noqa: F401

    cls = HOST_FETCHER_REGISTRY.get(generation, HostSshFetcher)
    return cls(executor, npu_mapping_cache)
