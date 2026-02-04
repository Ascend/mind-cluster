import os
from typing import List, Dict, Tuple

from diag_tool.core.collect.fetcher.dump_log_fetcher.base import DumpLogDirParser
from diag_tool.core.collect.fetcher.dump_log_fetcher.bmc.base import BmcDumpLogDataType
from diag_tool.core.collect.fetcher.dump_log_fetcher.cli_output_parsed_data import CliOutputParsedData
from diag_tool.utils import csv_tool, helpers, list_tool


class BmcDumpLogParser(DumpLogDirParser):
    _SPLIT_LEN = 2
    _BMC_CONFIGS = {
        BmcDumpLogDataType.BMC_IP.name: os.path.join("AppDump", "bmc_network", "network_info.txt"),
        BmcDumpLogDataType.SN_NUM.name: os.path.join("AppDump", "frudata", "fruinfo.txt"),
        BmcDumpLogDataType.SEL_INFO.name: os.path.join("AppDump", "event", "sel.txt"),
        BmcDumpLogDataType.SENSOR_INFO.name: os.path.join("AppDump", "sensor", "sensor_info.txt"),
        BmcDumpLogDataType.HEALTH_EVENTS.name: os.path.join("AppDump", "event", "current_event.txt"),
    }
    _OPTICAL_MODULE_HISTORY_LOG_PATH = [
        os.path.join("AppDump", "network_adapter", "optical_module", "optical_module_history_info_log.csv"),
        os.path.join("AppDump", "CpuMem", "NpuIO", "optical_module_history_info_log.csv"),
    ]

    def __init__(self, root_dir: str, parse_dir: str):
        super().__init__(root_dir, parse_dir)

    def parse(self) -> dict:
        parse_data = CliOutputParsedData()
        for config, path in self._BMC_CONFIGS.items():
            config_file_path = os.path.join(self.parse_dir, path)
            with open(config_file_path, "r", encoding="utf-8") as f:
                file_lines = f.readlines()
            if config == BmcDumpLogDataType.BMC_IP.name:
                parse_data.add_data([config], self.parse_bmc_ip(file_lines))
            elif config == BmcDumpLogDataType.SN_NUM.name:
                parse_data.add_data([config], self.parse_host_sn_num(file_lines))
            else:
                parse_data.add_data([config], "\n".join(file_lines))
        parse_data.add_data([BmcDumpLogDataType.OP_HISTORY_INFO_LOG.name], self.parse_optical_module_history_info_log())
        return parse_data.get_data_dict()

    def parse_bmc_ip(self, file_lines) -> str:
        for line in file_lines:
            if "IP Address" in line:
                part_line = line.split(":")
                if len(part_line) == self._SPLIT_LEN:
                    return part_line[-1].strip()
        return ""

    def parse_host_sn_num(self, file_lines) -> str:
        sn_flag = False
        for line in file_lines:
            if "FRU Device Description : Builtin FRU Device (FRUID 0, BMC)" in line:
                sn_flag = True
                continue
            if sn_flag and "System Serial Number" in line:
                part_line = line.split(":")
                if len(part_line) == self._SPLIT_LEN:
                    return part_line[-1].strip()
        return ""

    def parse_lcne_sn_num(self, file_lines) -> str:
        pass

    def parse_optical_module_history_info_log(self) -> List[Dict]:
        """
        该文件分为2类title, link down和periodic recording(中断或周期)记录
        聚合可能分散的表格, 其中periodic recording关键字CurrentMax
        """
        title_start = "LogTime"
        periodic_recording_tag = "CurrentMax"
        for log_path in self._OPTICAL_MODULE_HISTORY_LOG_PATH:
            log_path = os.path.join(self.parse_dir, log_path)
            if not os.path.exists(log_path):
                continue
            link_down_header = None
            link_down_list = []
            is_link_down_part = False
            lists = csv_tool.csv_to_list_of_lists(log_path)
            for lst in lists:
                if not lst:
                    continue
                if lst[0] == title_start:
                    if any(periodic_recording_tag in item for item in lst):
                        is_link_down_part = False
                    else:
                        if not link_down_header:
                            link_down_header = lst
                        is_link_down_part = True
                elif is_link_down_part:
                    link_down_list.append(lst)
            link_down_maps = list_tool.list_of_lists_to_dict_list(link_down_list, link_down_header)
            return link_down_maps
        return []
