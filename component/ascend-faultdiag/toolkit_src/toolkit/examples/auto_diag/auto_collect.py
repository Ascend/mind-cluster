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

from toolkit.core.context.diag_ctx import DiagCtx
from toolkit.core.service.collect_bmc_info import CollectBmcsInfo
from toolkit.core.service.collect_host_info import CollectHostsInfo
from toolkit.core.service.collect_l1_hccs_info import CollectL1HccsInfo
from toolkit.core.service.collect_swi_info import CollectSwiInfo
from toolkit.core.service.init_fetcher import InitFetcher
from toolkit.core.service.output_cache import OutputCache


class AutoCollect:

    def __init__(self, diag_ctx=DiagCtx()):
        self.diag_ctx = diag_ctx

    async def main(self):
        await InitFetcher(self.diag_ctx).run()
        await asyncio.gather(
            CollectHostsInfo(self.diag_ctx).run(),
            CollectBmcsInfo(self.diag_ctx).run(),
            self.collect_swi_info()
        )
        await OutputCache(self.diag_ctx).run()

    async def collect_swi_info(self):
        await CollectSwiInfo(self.diag_ctx).run()
        await CollectL1HccsInfo(self.diag_ctx).run()


if __name__ == '__main__':
    asyncio.run(AutoCollect().main())
