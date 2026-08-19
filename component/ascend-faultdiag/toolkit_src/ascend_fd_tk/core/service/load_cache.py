#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# pylint: disable=R0801
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

import json
import os.path
from typing import Callable, Dict, Type

from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.core.common.path import CommonPath
from ascend_fd_tk.core.context.host_registry import HOST_INFO_REGISTRY
from ascend_fd_tk.core.model.bmc import BmcInfo
from ascend_fd_tk.core.model.host import HostInfo
from ascend_fd_tk.core.model.switch import SwitchInfo
from ascend_fd_tk.core.service.base import DiagService
from ascend_fd_tk.utils.file_tool import convert_log_path


class LoadCache(DiagService):
    @staticmethod
    def _load_cache(
        cache_dir: str,
        cache_type_class: Type[JsonObj],
        cache_obj_map: Dict,
        class_resolver: Callable[[dict], Type[JsonObj]] = None,
    ):
        """加载 cache JSON 文件并反序列化。

        Args:
            class_resolver: 可选，根据 JSON 内容决定实际反序列化类。
                            host cache 用此参数按 chip_generation 在 HostInfo/HostInfoA5 间选择。
        """
        cache_dir = convert_log_path(cache_dir)
        if not cache_dir:
            return
        # 遍历cache_dir下所有.json文件
        for filename in os.listdir(cache_dir):
            if not filename.endswith('.json'):
                continue
            file_path = os.path.join(cache_dir, filename)

            # 读取JSON文件内容
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
                if not content:
                    continue
                json_dict = json.loads(content)
                actual_cls = class_resolver(json_dict) if class_resolver else cache_type_class
                obj = actual_cls.from_dict(json_dict)
                # 获取key（文件名不含扩展名）
                key = os.path.splitext(filename)[0]
                # 将对象添加到cache_obj_map字典中
                cache_obj_map[key] = obj

    @staticmethod
    def _resolve_host_class(json_dict: dict) -> Type[JsonObj]:
        """按 chip_generation 选择 host 反序列化类。

        通过 HOST_INFO_REGISTRY 注册表查找，新增代际无需修改本方法。
        """
        chip_generation = json_dict.get("chip_generation", "A3")
        return HOST_INFO_REGISTRY.get(chip_generation, HostInfo)

    async def run(self):
        cache = self.diag_ctx.cache
        self._load_cache(CommonPath.COLLECT_BMC_CACHE_DIR, BmcInfo, cache.bmcs_info)
        self._load_cache(CommonPath.COLLECT_SWITCH_CACHE_DIR, SwitchInfo, cache.swis_info)
        self._load_cache(CommonPath.COLLECT_HOST_CACHE_DIR, HostInfo, cache.hosts_info, self._resolve_host_class)
        # 排序，屏蔽系统原生读取顺序差异，也保证后续分析有序
        cache.sort_info()
        # 从首个 HostInfo 提取代际（集群默认同代际），供诊断阶段分流 analyzer。旧 cache 无 chip_generation 字段时为空串，默认 A3
        for host_info in cache.hosts_info.values():
            if host_info.chip_generation:
                cache.chip_generation = host_info.chip_generation
                break
        cache.init_diag_data()
