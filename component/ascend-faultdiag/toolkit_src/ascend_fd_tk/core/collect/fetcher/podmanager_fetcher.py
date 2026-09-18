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

from typing import List

from ascend_fd_tk.core.collect.fetcher.base import Fetcher


class PoDManagerFetcher(Fetcher):
    """PoDManager 采集器基类：每连接绑定一组 sfu_slot_ids/npu_slot_ids，串行切换槽位执行命令。"""

    def __init__(self, sfu_slot_ids: List[str], npu_slot_ids: List[str]):
        self.sfu_slot_ids = sfu_slot_ids
        self.npu_slot_ids = npu_slot_ids
        self.current_slot = (
            self.sfu_slot_ids[0] if self.sfu_slot_ids else (self.npu_slot_ids[0] if self.npu_slot_ids else None)
        )

    async def switch_slot(self, slot_id: str):
        self.current_slot = slot_id  # 切换到指定槽位，具体切换命令由子类实现。
