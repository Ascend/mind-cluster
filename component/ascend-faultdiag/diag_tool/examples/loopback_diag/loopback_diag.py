import asyncio

from diag_tool.core.context.diag_ctx import DiagCtx
from diag_tool.core.service.generate_diag_report import GenerateDiagReport
from diag_tool.core.service.load_cache import LoadCache
from diag_tool.core.service.loopback_diag import LoopbackDiag


class AutoLoopbackDiag:

    def __init__(self, diag_ctx=DiagCtx()):
        self.diag_ctx = diag_ctx

    async def main(self):
        await self.load_diag_info()
        await LoopbackDiag(self.diag_ctx).run()
        await GenerateDiagReport(self.diag_ctx).run()
        self.diag_ctx.close()

    async def load_diag_info(self):
        await asyncio.gather(
            LoadCache(self.diag_ctx).run()
        )


if __name__ == '__main__':
    asyncio.run(AutoLoopbackDiag().main())

