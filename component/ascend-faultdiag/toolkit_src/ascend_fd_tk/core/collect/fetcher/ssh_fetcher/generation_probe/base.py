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
from typing import Dict, Tuple

from ascend_fd_tk.core.common.diag_enum import NpuType


class GenerationProbe(abc.ABC):
    """设备代际探测器抽象基类。"""

    @abc.abstractmethod
    async def probe(self, executor) -> Tuple[NpuType, Dict]:
        """执行代际探测命令并解析。

        Args:
            executor: 已建立 SSH 会话的异步执行器

        Returns:
            (generation, extra_data):
              - generation: 探测到的芯片代际，探测失败时回退到 A3
              - extra_data: 可复用的命令解析数据，供对应 fetcher 注入缓存，
                避免重复执行同一条命令。无附加数据时返回空 dict。
        """
