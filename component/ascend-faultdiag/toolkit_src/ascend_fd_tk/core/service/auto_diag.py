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

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.context.register import recursive_scan_and_register, get_analyzers
from ascend_fd_tk.core.service.base import DiagService


class AutoDiag(DiagService):
    async def run(self):
        recursive_scan_and_register("ascend_fd_tk.core.fault_analyzer")
        # cache 中 chip_generation 存为字符串，未写入时按 A3 处理
        gen_str = self.diag_ctx.cache.chip_generation
        generation = NpuType(gen_str) if gen_str else NpuType.A3
        # 代际确定后应用阈值覆盖配置（set_config_dir 时解析暂存，此处延迟应用）
        self.diag_ctx.threshold_loader.apply(gen_str)
        for cls in get_analyzers(generation):
            self.diag_ctx.diag_result.extend(cls(self.diag_ctx.cache).analyse())
