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

from diag_tool.core.common.diag_enum import Customer
from diag_tool.core.context.diag_ctx import DiagCtx
from diag_tool.core.service.auto_inspection import AutoInspection
from diag_tool.core.service.load_cache import LoadCache


class Inspection:

    def __init__(self, diag_ctx=DiagCtx(), customer: Customer=Customer.Mayi):
        self.diag_ctx = diag_ctx
        self.customer = customer

    async def main(self):
        await LoadCache(self.diag_ctx).run()
        await AutoInspection(self.diag_ctx, self.customer).run()
        self.diag_ctx.close()



if __name__ == '__main__':
    asyncio.run(Inspection().main())
