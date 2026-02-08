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
from typing import Dict

from diag_tool.core.common.json_obj import JsonObj
from diag_tool.core.common.path import CommonPath
from diag_tool.core.context.diag_ctx import DiagCtx
from diag_tool.core.service.base import DiagService


class OutputCache(DiagService):

    def __init__(self, diag_ctx: DiagCtx):
        super().__init__(diag_ctx)

    @staticmethod
    def _output_cache(cache_dir: str, cache_obj_map: Dict[str, JsonObj]):
        if not os.path.exists(cache_dir):
            os.makedirs(cache_dir, exist_ok=True)
        for name, cache_obj in cache_obj_map.items():
            with open(os.path.join(cache_dir, f"{name}.json"), "w", encoding="utf-8") as fs:
                fs.write(cache_obj.to_json())

    async def run(self):
        cache = self.diag_ctx.cache
        self._output_cache(CommonPath.COLLECT_BMC_CACHE_DIR, cache.bmcs_info)
        self._output_cache(CommonPath.COLLECT_SWITCH_CACHE_DIR, cache.swis_info)
        self._output_cache(CommonPath.COLLECT_HOST_CACHE_DIR, cache.hosts_info)
