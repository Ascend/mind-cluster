from typing import List, Dict

from diag_tool.core.collect.base import Collector, log_collect_async_event
from diag_tool.core.collect.fetcher.switch_fetcher import SwitchFetcher
from diag_tool.core.collect.parser.switch_parser import SwitchParser
from diag_tool.core.model.switch import SwitchInfo, InterfaceBrief, SwiOpticalModel, InterfaceMapping, AlarmInfo, \
    InterfaceInfo, BitErrRate, TransceiverInfo


class SwitchCollector(Collector):

    def __init__(self, fetcher: SwitchFetcher):
        self.fetcher = fetcher
        self.parser = SwitchParser()

    @log_collect_async_event()
    async def collect(self) -> SwitchInfo:
        await self.fetcher.init_fetcher()
        switch_name = await self.fetcher.get_switch_name()
        switch_ip = await self.get_id()
        sn = await self.coll_serial_num()
        interface_briefs = await self.coll_interface_brief()
        optical_models = await self.coll_optical_module_info(interface_briefs)
        active_alarm_info = await self.coll_active_alarms()
        history_alarm_info = await self.coll_history_alarms()
        interface_info = await self.coll_interface_info()
        interface_mapping = await self.coll_lldp_nei_brief()
        date_time = await self.coll_datetime()
        bit_error_rate_list = await self.coll_bit_error_rate(interface_briefs)
        transceiver_infos = await self.coll_transceiver_info()
        switch_info = SwitchInfo(switch_name, switch_ip, sn=sn,
                                 optical_models=optical_models,
                                 interface_briefs=interface_briefs,
                                 active_alarm_info=active_alarm_info,
                                 history_alarm_info=history_alarm_info,
                                 interface_info=interface_info,
                                 interface_mapping=interface_mapping,
                                 date_time=date_time,
                                 bit_error_rate=bit_error_rate_list,
                                 transceiver_infos=transceiver_infos, )
        return switch_info

    async def get_id(self) -> str:
        return await self.fetcher.fetch_id()

    async def coll_serial_num(self):
        sn_num_table = await self.fetcher.fetch_serial_num()
        return self.parser.parse_esn(sn_num_table)

    async def coll_interface_brief(self) -> List[InterfaceBrief]:
        interface_brief_str = await self.fetcher.fetch_interface_brief()
        return self.parser.parse_interface_brief(interface_brief_str)

    async def coll_optical_module_info(self, interface_briefs: List[InterfaceBrief]) -> List[SwiOpticalModel]:
        opt_module_info = await self.fetcher.fetch_optical_module_info(interface_briefs)
        table_parse_list = self.parser.parse_opt_module_info_from_table(opt_module_info, interface_briefs)

        switch_log_info = await self.fetcher.fetch_switch_log_info()
        port_mapping = await self.coll_port_mapping()
        line_parse_list = self.parser.parse_opt_module_info_from_line(switch_log_info, port_mapping)
        return table_parse_list + line_parse_list

    async def coll_bit_error_rate(self, interface_briefs: List[InterfaceBrief]) -> List[BitErrRate]:
        bit_error_rate = await self.fetcher.fetch_bit_error_rate(interface_briefs)
        return self.parser.parse_bit_err_rate(bit_error_rate, interface_briefs)

    async def coll_lldp_nei_brief(self) -> List[InterfaceMapping]:
        lldp_nei_brief = await self.fetcher.fetch_lldp_nei_brief()
        return self.parser.parse_lldp_nei_brief(lldp_nei_brief)

    async def coll_active_alarms(self) -> List[AlarmInfo]:
        active_alarms = await self.fetcher.fetch_active_alarms()
        active_alarms_objs = self.parser.parse_alarms(active_alarms)
        active_alarms_verbose = await self.fetcher.fetch_active_alarms_verbose()
        active_alarms_verbose_objs = self.parser.parse_alarm_verbose(active_alarms_verbose)
        return active_alarms_objs + active_alarms_verbose_objs

    async def coll_history_alarms(self) -> List[AlarmInfo]:
        history_alarms = await self.fetcher.fetch_history_alarms()
        result = self.parser.parse_alarms(history_alarms)
        history_alarms_verbose = await self.fetcher.fetch_history_alarms_verbose()
        result.extend(self.parser.parse_alarm_verbose(history_alarms_verbose))
        return result

    async def coll_interface_info(self) -> List[InterfaceInfo]:
        interface_info = await self.fetcher.fetch_interface_info()
        return self.parser.parse_interface_info(interface_info)

    async def coll_datetime(self) -> str:
        datetime_str = await self.fetcher.fetch_datetime()
        return self.parser.parse_datetime(datetime_str)

    async def coll_transceiver_info(self) -> List[TransceiverInfo]:
        cmd_res = await self.fetcher.fetch_transceiver_info()
        return self.parser.parse_transceiver_info(cmd_res)

    async def coll_port_mapping(self) -> Dict[str, str]:
        cmd_res = await self.fetcher.fetch_interface_port_mapping()
        return self.parser.parse_port_mapping(cmd_res)
