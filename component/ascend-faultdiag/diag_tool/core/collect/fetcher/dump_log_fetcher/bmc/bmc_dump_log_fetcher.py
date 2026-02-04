from typing import List

from diag_tool.core.collect.fetcher.bmc_fetcher import BmcFetcher
from diag_tool.core.collect.fetcher.dump_log_fetcher.bmc.base import BmcDumpLogDataType
from diag_tool.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData
from diag_tool.core.collect.parser.bmc_parser import BmcParser
from diag_tool.core.model.bmc import BmcSensorInfo, BmcSelInfo, BmcHealthEvents, \
    LinkDownOpticalModuleHistoryLog


class BmcDumpLogFetcher(BmcFetcher):

    def __init__(self, parse_dir: str, parsed_data: CliOutputParsedData):
        self.parse_dir = parse_dir
        self.parsed_data = parsed_data

    async def fetch_id(self) -> str:
        return self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.BMC_IP.name)

    async def fetch_bmc_sn(self) -> str:
        return self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.SN_NUM.name)

    async def fetch_bmc_sel_list(self) -> List[BmcSelInfo]:
        data = self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.SEL_INFO.name)
        return BmcParser.trans_sel_results(data)

    async def fetch_bmc_health_events(self) -> List[BmcHealthEvents]:
        data = self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.HEALTH_EVENTS.name)
        return BmcParser.trans_health_events_results(data)

    async def fetch_bmc_sensor_list(self) -> List[BmcSensorInfo]:
        data = self.parsed_data.fetch_data_by_name(BmcDumpLogDataType.SENSOR_INFO.name)
        return BmcParser.trans_sensor_results(data)

    async def fetch_bmc_date(self) -> str:
        return ""

    async def fetch_bmc_diag_info_log(self):
        pass

    async def fetch_bmc_optical_module_history_info_log(self) -> List[LinkDownOpticalModuleHistoryLog]:
        link_down_optical_module_history_log_data = self.parsed_data.fetch_data_by_name(
            BmcDumpLogDataType.OP_HISTORY_INFO_LOG.name)
        link_down_optical_module_history_logs = []
        for log_dict in link_down_optical_module_history_log_data:
            link_down_optical_module_history_logs.append(LinkDownOpticalModuleHistoryLog.from_dict(log_dict))
        return link_down_optical_module_history_logs
