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

import asyncio
import os

from toolkit.core.context.diag_ctx import DiagCtx
from toolkit.core.service.collect_loopback_info import CollectLoopbackInfo
from toolkit.core.service.init_fetcher import InitFetcher
from toolkit.core.service.output_cache import OutputCache


class LoopbackCollect:

    def __init__(self, diag_ctx=DiagCtx()):
        self.diag_ctx = diag_ctx

    async def main(self):
        await InitFetcher(self.diag_ctx).run()
        # 目前只支持Host输入
        await asyncio.gather(
            CollectLoopbackInfo(self.diag_ctx).run()
        )
        await OutputCache(self.diag_ctx).run()

if __name__ == '__main__':
    confirm = input(f"现在将对配置文件中的[host]进行光模块环回，光模块环回对环境会有影响，请确认是否继续执行 (y/n): {os.linesep}")
    if confirm.lower() == "y":
        asyncio.run(LoopbackCollect().main())

