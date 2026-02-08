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

from diag_tool.core.context.diag_ctx import DiagCtx
from diag_tool.core.context.register import recursive_scan_and_register, ANALYZER_STORE
from diag_tool.core.service.base import DiagService


class AutoDiag(DiagService):

    def __init__(self, diag_ctx: DiagCtx):
        super().__init__(diag_ctx)

    async def run(self):
        recursive_scan_and_register("diag_tool.core.fault_analyzer")
        for cls in ANALYZER_STORE:
            self.diag_ctx.diag_result.extend(cls(self.diag_ctx.cache).analyse())
