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

import os.path
from typing import Dict, Type

from toolkit.core.common.json_obj import JsonObj
from toolkit.core.common.path import CommonPath
from toolkit.core.context.diag_ctx import DiagCtx
from toolkit.core.model.bmc import BmcInfo
from toolkit.core.model.host import HostInfo
from toolkit.core.model.switch import SwitchInfo
from toolkit.core.service.base import DiagService


class LoadCache(DiagService):

    def __init__(self, diag_ctx: DiagCtx):
        super().__init__(diag_ctx)

    @staticmethod
    def _load_cache(cache_dir: str, cache_type_class: Type[JsonObj], cache_obj_map: Dict):
        if not os.path.exists(cache_dir):
            return
        # 遍历cache_dir下所有.json文件
        for filename in os.listdir(cache_dir):
            if filename.endswith('.json'):
                file_path = os.path.join(cache_dir, filename)

                # 读取JSON文件内容
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
                    if content:
                        # 用cache_type_class的from_json函数将内容转为对象
                        obj = cache_type_class.from_json(content)
                        # 获取key（文件名不含扩展名）
                        key = os.path.splitext(filename)[0]
                        # 将对象添加到cache_obj_map字典中
                        cache_obj_map[key] = obj

    async def run(self):
        cache = self.diag_ctx.cache
        self._load_cache(CommonPath.COLLECT_BMC_CACHE_DIR, BmcInfo, cache.bmcs_info)
        self._load_cache(CommonPath.COLLECT_SWITCH_CACHE_DIR, SwitchInfo, cache.swis_info)
        self._load_cache(CommonPath.COLLECT_HOST_CACHE_DIR, HostInfo, cache.hosts_info)
        cache.init_diag_data()
