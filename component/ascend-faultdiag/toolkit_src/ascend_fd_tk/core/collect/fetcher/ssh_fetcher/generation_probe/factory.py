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

from typing import Optional

from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.generation_probe.base import GenerationProbe
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.generation_probe.host_probe import HostGenerationProbe
from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.host_ssh_fetcher import HostSshFetcher


def create_generation_probe(fetcher_type) -> Optional[GenerationProbe]:
    """根据 fetcher 类型返回对应的代际探测器，不需要探测时返回 None。"""
    if fetcher_type is HostSshFetcher:
        return HostGenerationProbe()
    # switch / bmc 命令与代际无关，暂不探测
    return None
