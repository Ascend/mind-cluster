import asyncio
import os

from diag_tool.core.context.diag_ctx import DiagCtx
from diag_tool.core.service.collect_loopback_info import CollectLoopbackInfo
from diag_tool.core.service.init_fetcher import InitFetcher
from diag_tool.core.service.output_cache import OutputCache


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

