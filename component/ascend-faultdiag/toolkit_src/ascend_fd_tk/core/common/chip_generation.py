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

import re
from typing import Dict, Tuple

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.model.cluster_mapping import DefaultNPUInfo


def parse_npu_mapping_and_generation(stdout: str) -> Tuple[NpuType, Dict[str, DefaultNPUInfo]]:
    """解析 ``npu-smi info -m`` 回显，同时产出代际与 npu_mapping。

    A3和A5回显格式（列由两个以上空格分隔）分别如下:
        NPU ID   Chip ID   Chip Logic ID   Chip Phy-ID  Chip Name
        0        0         0               0             Ascend910

        NPU ID   Slot ID   Chip ID   Chip Phy-ID  Chip Name
        0        0         0               0      Ascend950

    Args:
        stdout: ``npu-smi info -m`` 命令的完整回显文本

    Returns:
        (generation, npu_mapping):
          - generation: 首个有效芯片确定的代际，无法确定时回退 A3
          - npu_mapping: chip_phy_id → DefaultNPUInfo 映射
    """
    npu_mapping: Dict[str, DefaultNPUInfo] = {}
    generation = NpuType.A3
    generation_decided = False
    lines = stdout.strip().split('\n')
    start_line, total_col, chip_slot_id_index, chip_phy_id_index, chip_name_index = 2, 5, 2, 3, 4
    for line in lines[start_line:]:
        parts = re.split(r'\s{2,}', line.strip())
        if len(parts) < total_col:
            continue
        npu_id = parts[0]
        chip_phy_id = parts[chip_phy_id_index]
        chip_name = parts[chip_name_index]
        if chip_name == 'Mcu' or npu_id == '-':
            continue
        # 以首个有效芯片的 chip Name 确定代际，后续不再覆盖
        if not generation_decided and "Ascend950" in chip_name:
            generation = NpuType.A5
            generation_decided = True
        chip_id = parts[1] if generation == NpuType.A3 else parts[chip_slot_id_index]
        npu_mapping[chip_phy_id] = DefaultNPUInfo(npu_id=npu_id, chip_id=chip_id, chip_phy_id=chip_phy_id)
    return generation, npu_mapping
