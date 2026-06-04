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

from ascend_fd_tk.core.root_cause.filter import RootCauseFilter
from ascend_fd_tk.core.service.base import DiagService
from ascend_fd_tk.utils.logger import DIAG_LOGGER


class RootCauseAnalysis(DiagService):
    """根因分析服务，构建信号链路并执行故障分析"""

    async def run(self):
        DIAG_LOGGER.info("正在执行链路根因分析...")
        self.diag_ctx.root_cause_filter = RootCauseFilter(self.diag_ctx.cache)
        self.diag_ctx.root_cause_filter.build_and_analyze()
        DIAG_LOGGER.info("链路根因分析完成")
