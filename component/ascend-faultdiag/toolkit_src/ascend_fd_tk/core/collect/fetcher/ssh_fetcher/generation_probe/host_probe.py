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

from typing import Dict, Tuple

from ascend_fd_tk.core.collect.fetcher.ssh_fetcher.generation_probe.base import GenerationProbe
from ascend_fd_tk.core.common.chip_generation import parse_npu_mapping_and_generation
from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.utils import logger
from ascend_fd_tk.utils.executors import CmdTask

_CONSOLE_LOGGER = logger.CONSOLE_LOGGER

_PROBE_CMD = "npu-smi info -m"


class HostGenerationProbe(GenerationProbe):
    """主机代际探测器，复用 npu-smi info -m 的回显。"""

    async def probe(self, executor) -> Tuple[NpuType, Dict]:
        try:
            cmd_res = await executor.run_cmd(CmdTask(_PROBE_CMD))
            if not cmd_res.is_success():
                _CONSOLE_LOGGER.warning("host 代际探测命令执行失败：%s", cmd_res.stderr)
                return NpuType.A3, {}
            generation, npu_mapping = parse_npu_mapping_and_generation(cmd_res.stdout)
            return generation, npu_mapping
        except Exception as e:
            _CONSOLE_LOGGER.info("host 代际探测异常：%s", e)
            return NpuType.A3, {}
