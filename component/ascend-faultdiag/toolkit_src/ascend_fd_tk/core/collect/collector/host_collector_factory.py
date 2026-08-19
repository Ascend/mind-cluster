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

from ascend_fd_tk.core.collect.collector.host_collector import HostCollector
from ascend_fd_tk.core.context.register import HOST_COLLECTOR_REGISTRY
from ascend_fd_tk.core.collect.fetcher.host_fetcher import HostFetcher
from ascend_fd_tk.core.common.diag_enum import NpuType


def create_host_collector(fetcher: HostFetcher) -> HostCollector:
    """根据 fetcher 携带的芯片代际返回对应的主机采集器。"""
    # 触发 A5 collector 模块导入，确保其装饰器完成注册
    import ascend_fd_tk.core.collect.collector.host_collector_a5  # noqa: F401

    generation = getattr(fetcher, 'chip_generation', NpuType.A3)
    cls = HOST_COLLECTOR_REGISTRY.get(generation, HostCollector)
    return cls(fetcher)
