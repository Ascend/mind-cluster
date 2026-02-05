from diag_tool.core.collect.base import Collector, log_collect_async_event
from diag_tool.core.collect.fetcher.host_fetcher import HostFetcher
from diag_tool.core.collect.parser.host_parser import HostParser
from diag_tool.core.model.host import HostInfo, NpuChipInfo, HCCNOpticalInfo, HCCNLinkStatInfo, HCCNStatInfo, \
    HCCNLLDPInfo, HccnPortHccsInfo, SpodInfo, CdrSnrInfo, HCCNDfxCfgInfo


class HostCollector(Collector):

    def __init__(self, fetcher: HostFetcher):
        self.fetcher = fetcher
        self.parser = HostParser()

    @log_collect_async_event()
    async def collect(self) -> HostInfo:
        host_id = await self.get_id()
        host_name = await self.fetcher.fetch_hostname()
        msnpureport_log = await self.fetcher.fetch_msnpureport_log()
        npu_mapping = await self.fetcher.fetch_npu_mapping()
        npu_type = await self.collect_npy_type()
        sn_num = await self.fetcher.fetch_sn_num()
        npu_chip_info = {}
        server_superpod_id = None
        server_index = None
        for npu_id, chips in npu_mapping.items():
            for chip_id, chip_phy_id in chips.items():
                hccn_optical_info = await self.collect_optical_info(chip_phy_id)
                hccn_link_stat_info = await self.collect_link_stat_info(chip_phy_id)
                hccn_stat_info = await self.collect_stat_info(chip_phy_id)
                hccn_lldp_info = await self.collect_lldp_info(chip_phy_id)
                hccs_info = await self.collect_hccs_info(npu_id, chip_id)
                spod_info = await self.collect_spod_info(npu_id, chip_id)
                speed = await self.collect_roce_speed(chip_phy_id)
                duplex = await self.fetcher.fetch_roce_duplex(chip_phy_id)
                net_health = await self.collect_hccn_tool_net_health(chip_phy_id)
                link_status = await self.collect_hccn_tool_link_status(chip_phy_id)
                cdr_snr_info = await self.collect_hccn_tool_cdr(chip_phy_id)
                hccn_dfx_cfg = await self.collect_hccn_dfx_cfg(chip_phy_id)
                if server_superpod_id is None and spod_info:
                    server_superpod_id = spod_info.super_pod_id
                if server_index is None and spod_info:
                    server_index = spod_info.server_index
                npu_chip_info[chip_phy_id] = NpuChipInfo(hccn_lldp_info=hccn_lldp_info,
                                                         hccn_optical_info=hccn_optical_info,
                                                         hccn_link_stat_info=hccn_link_stat_info,
                                                         hccn_stat_info=hccn_stat_info,
                                                         hccs_info=hccs_info,
                                                         spod_info=spod_info,
                                                         npu_type=npu_type, npu_id=npu_id, chip_id=chip_id,
                                                         chip_phy_id=chip_phy_id, speed=speed, duplex=duplex,
                                                         net_health=net_health, link_status=link_status,
                                                         cdr_snr_info=cdr_snr_info, hccn_dfx_cfg=hccn_dfx_cfg)
        host_info = HostInfo(host_id, sn_num, hostname=host_name,
                             server_superpod_id=server_superpod_id, server_index=server_index,
                             msnpureport_log=msnpureport_log,
                             npu_chip_info=npu_chip_info)
        return host_info

    async def get_id(self) -> str:
        return await self.fetcher.fetch_id()

    async def collect_npy_type(self) -> str:
        recv = await self.fetcher.fetch_npu_type()
        return self.parser.parse_npu_type(recv)

    async def collect_optical_info(self, chip_phy_id) -> HCCNOpticalInfo:
        recv = await self.fetcher.fetch_optical_info(chip_phy_id)
        return self.parser.parse_optical_info(recv)

    async def collect_link_stat_info(self, chip_phy_id) -> HCCNLinkStatInfo:
        recv = await self.fetcher.fetch_link_stat_info(chip_phy_id)
        return self.parser.parse_link_stat_info(recv)

    async def collect_stat_info(self, chip_phy_id) -> HCCNStatInfo:
        recv = await self.fetcher.fetch_stat_info(chip_phy_id)
        return self.parser.parse_stat_info(recv)

    async def collect_lldp_info(self, chip_phy_id) -> HCCNLLDPInfo:
        recv = await self.fetcher.fetch_lldp_info(chip_phy_id)
        return self.parser.parse_lldp_info(recv)

    async def collect_hccs_info(self, npu_id, chip_id) -> HccnPortHccsInfo:
        recv = await self.fetcher.fetch_hccs_info(npu_id, chip_id)
        return self.parser.parse_hccs_info(recv)

    async def collect_roce_speed(self, chip_phy_id) -> str:
        recv = await self.fetcher.fetch_roce_speed(chip_phy_id)
        return self.parser.parse_roce_speed(recv)

    async def collect_spod_info(self, npu_id, chip_id) -> SpodInfo:
        recv = await self.fetcher.fetch_spod_info(npu_id, chip_id)
        return self.parser.parse_spod_info(recv)

    async def collect_hccn_tool_net_health(self, chip_phy_id) -> str:
        recv = await self.fetcher.fetch_hccn_tool_net_health(chip_phy_id)
        return self.parser.parse_hccn_tool_net_health(recv)

    async def collect_hccn_tool_link_status(self, chip_phy_id) -> str:
        recv = await self.fetcher.fetch_hccn_tool_link_status(chip_phy_id)
        return self.parser.parse_hccn_tool_link_status(recv)

    async def collect_hccn_tool_cdr(self, chip_phy_id) -> CdrSnrInfo:
        recv = await self.fetcher.fetch_hccn_tool_cdr(chip_phy_id)
        return self.parser.parse_hccn_tool_cdr(recv)

    async def collect_hccn_dfx_cfg(self, chip_phy_id) -> HCCNDfxCfgInfo:
        recv = await self.fetcher.fetch_hccn_dfx_cfg(chip_phy_id)
        return self.parser.parse_hccn_dfx_cfgr(recv)
