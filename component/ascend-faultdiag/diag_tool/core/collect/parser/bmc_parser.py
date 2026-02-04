from typing import List

from diag_tool.core.model.bmc import BmcSelInfo, BmcSensorInfo, BmcHealthEvents, PeriodicRecordingOpticalModuleHistoryLog
from diag_tool.utils import csv_tool
from diag_tool.utils.table_parser import TableParser


class BmcParser:

    @classmethod
    def trans_sel_results(cls, cmd_res: str):
        titles_dict = {
            "id": "ID", "generation_time": "Generation Time", "severity": "Severity", "event_code": "Event Code",
            "status": "Status", "event_description": "Event Description"
        }
        parse_data_list = TableParser.parse(cmd_res, titles_dict, col_separator="| ")
        sel_info_list = []
        for data in parse_data_list:
            sel_info_list.append(BmcSelInfo.from_dict(data))
        return sel_info_list

    @classmethod
    def trans_sensor_results(cls, cmd_res: str):
        titles_dict = {
            "sensor_id": "sensor id", "sensor_name": "sensor name", "value": "value", "unit": "unit",
            "status": "status", "lnr": "lnr", "lc": "lc", "lnc": "lnc", "unc": "unc  ", "uc": "uc",
            "unr": "unr", "phys": "phys", "n_hys": "nhys"
        }
        parse_data_list = TableParser.parse(cmd_res, titles_dict, col_separator="| ")
        sensor_info_list = []
        for data in parse_data_list:
            sensor_info_list.append(BmcSensorInfo.from_dict(data))
        return sensor_info_list

    @classmethod
    def trans_health_events_results(cls, cmd_res: str):
        titles_dict = {
            "event_num": "Event Num", "event_time": "Event Time", "alarm_level": "Alarm Level",
            "event_code": "Event Code", "event_description": "Event Description"
        }
        parse_data_list = TableParser.parse(cmd_res, titles_dict, col_separator="| ")
        health_events = []
        for data in parse_data_list:
            health_events.append(BmcHealthEvents.from_dict(data))
        return health_events

