from diag_tool.core.collect.collector.host_collector import HostCollector
from diag_tool.core.context.diag_ctx import DiagCtx
from diag_tool.core.service.base import DiagService


class CollectHostsInfo(DiagService):

    def __init__(self, diag_ctx: DiagCtx):
        super().__init__(diag_ctx)

    async def run(self):
        if not self.diag_ctx.host_fetchers:
            return
        async_tasks = []
        for fetcher in self.diag_ctx.host_fetchers.values():
            async_tasks.append(HostCollector(fetcher).collect())
        for task in async_tasks:
            host_info = await task
            self.diag_ctx.cache.hosts_info.update({host_info.host_id: host_info})