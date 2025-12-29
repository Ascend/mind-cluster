#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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
import os
import re
import logging
from datetime import datetime
from abc import ABC, abstractmethod
from itertools import chain
from typing import Dict, Tuple, Union
from dataclasses import dataclass

from ascend_fd.model.context import KGParseCtx
from ascend_fd.pkg.parse.parser_saver import LogInfoSaver
from ascend_fd.utils.regular_table import DATETIME_REGEX, KG_MAX_TIME, NPU_INFO_SOURCE
from ascend_fd.utils.status import FileNotExistError, InfoIncorrectError
from ascend_fd.utils.tool import safe_read_open_with_size
from ascend_fd.pkg.parse.knowledge_graph.parser.file_parser import FileParser
from ascend_fd.pkg.diag.root_cluster.utils import NEGATIVE_ONE
from ascend_fd.utils.fault_code import LINK_DOWN_FAULT, HBM_ABNORMAL_FAULT, ABNORMAL_FEC_MODE_FAULT, \
    GENERAL_NET_HEALTH_FAULT, OPTICAL_MODULE_NOT_PRESENT, PHYSICAL_CARD_DROPPING, \
    SOFTWARE_CARD_DROPPING, IP_NOT_CONFIG_FAULT, OPTICAL_POWER_FAULT, OPTICAL_MODULE_NOT_RX_OR_TX_FAULT, \
    OPTICAL_MODULE_OUT_OF_LOCK_FAULT, NPU_DRIVER_FAULT

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
LOGIC_ID_CONFIG: Dict[Tuple[str, str], str] = dict()


class NpuInfoLineParser:
    def __init__(self, id_regex: str, info_regex: str, detail_regex: str, detail_func):
        """
        Line info parser for npu info
        :param id_regex: id info regex
        :param info_regex: base info regex
        :param detail_regex: detail info regex
        :param detail_func: detail info lambda func
        """
        self.npu_id_regex = re.compile(id_regex)
        self.regex = re.compile(info_regex)
        self.detail_regex = re.compile(detail_regex) if detail_regex else None
        self.detail_func = detail_func
        self.npu_id = None
        self.desc = None

    def parse(self, desc: str):
        """
        Parse each paragraph in npu info file
        :param desc: single paragraph
        :return: base info dict
        """
        ret = self.regex.findall(desc)
        if not ret:
            self.desc = None
            return {}
        # Regular matches the NPU ID
        self.desc = desc
        ret = self.npu_id_regex.findall(desc)
        if not ret:
            self.npu_id = None
            kg_logger.warning("Can't find npu id in npu info file, please check the npu info file.")
            return {}
        npu_id = ret[0]
        if isinstance(npu_id, tuple):
            # if LOGIC_ID_CONFIG has been initialized, obtain logic id through (npu_id, chip_id)
            if LOGIC_ID_CONFIG and npu_id[0] and npu_id[1]:
                self.npu_id = LOGIC_ID_CONFIG.get((npu_id[0], npu_id[1]), npu_id[0])
            # otherwise in case of -i n -c 0, keep the first place only
            else:
                self.npu_id = npu_id[0]
        else:
            self.npu_id = npu_id
        # Regular matches the detail info (detail_func)
        if not self.detail_regex and self.detail_func:
            detail_info = self.detail_func(desc)
            if not detail_info:
                kg_logger.warning("Can't find detail info in npu info file by detail func.")
                return {}
            return {"npu_id": self.npu_id, "detail_info": detail_info, "key_info": desc}
        if not self.detail_regex:
            kg_logger.warning("Can't find details because 'detail_regex' is not defined.")
            return {}
        # Regular matches the detail info (detail_regex and detail_func)
        ret = self.detail_regex.findall(desc)
        if not ret:
            kg_logger.warning("Can't find detail info in npu info file by regex '%s'.", self.detail_regex.pattern)
            return {}
        detail_info = [self.detail_func(single_ret) if self.detail_func else single_ret for single_ret in ret]
        return {"npu_id": self.npu_id, "detail_info": detail_info, "key_info": desc}


class GeneralInfoParser(ABC):
    def __init__(self, end_time: str = ""):
        self.info_dict = dict()
        if end_time:
            self.end_time = end_time

    def parse(self, snippets_param: dict) -> list:
        """
        Collecting, storing and processing all info in general case
        :param snippets_param: npu info dict include key for when it was collected and value for a list of snippets
        :return: event list
        """
        event_list = []
        for name, event_message_list in snippets_param.items():
            for event_message in event_message_list:
                self.parse_info(name, event_message)
        if self.info_dict:
            event_list.extend(self.processing_info())
        return event_list

    @abstractmethod
    def parse_info(self, name: str, event_message: str):
        """
        Parse info into info dict based on name and event_message
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        pass

    @abstractmethod
    def processing_info(self) -> list:
        """
        Processing info and return a list of events
        :return: event list
        """
        pass


class PairInfoParser(ABC):
    def __init__(self, end_time: str):
        self.end_time = end_time

    def parse(self, snippets_param: dict) -> list:
        """
        Collecting, storing and processing paired info in general case
        :param snippets_param: npu info dict include key for when it was collected and value for a list of snippets
        :return: event list
        """
        event_list = []
        for event_message_list in snippets_param.values():
            distinguishable_info_dict = {}
            definitive_info_dict = {}
            for event_message in event_message_list:
                # distinguishable_info_dict and definitive_info_dict are both input and output parameters
                self.parse_pair_info(distinguishable_info_dict, definitive_info_dict, event_message)
            event_list.extend(self.process_pair_info(distinguishable_info_dict, definitive_info_dict))
        return event_list

    @abstractmethod
    def parse_pair_info(self, distinguishable_info_dict: dict, definitive_info_dict: dict, event_message: str):
        """
        Parse paired info store them in two dicts
        :param distinguishable_info_dict: the info for distinguishing the target info
        :param definitive_info_dict: the info that is definitive for the wondering result
        :param event_message: event message used for parsing
        """
        pass

    @abstractmethod
    def check_event_existence(self, npu_id: str, distinguishable_info_dict: dict, definitive_info_dict: dict) -> list:
        """
        Check the existence of event depends on logics of subclass
        :param npu_id: str to determining npu id
        :param distinguishable_info_dict: the info for distinguishing the target info
        :param definitive_info_dict: the info that is definitive for the wondering result
        :return:
        """
        pass

    def process_pair_info(self, distinguishable_info_dict: dict, definitive_info_dict: dict) -> list:
        """
        process paired info and present them for further logics of judgement
        :param distinguishable_info_dict: the info for distinguishing the target info
        :param definitive_info_dict: the info that is definitive for the wondering result
        :return: event list
        """
        event_list = []
        for npu_id in distinguishable_info_dict.keys():
            distinguishable_event_dict = distinguishable_info_dict.get(npu_id, dict())
            definitive_event_dict = definitive_info_dict.get(npu_id, dict())
            if not distinguishable_event_dict and not definitive_event_dict:
                kg_logger.warning("Paired info not matched or lost when checking device '%s'.", npu_id)
                continue
            events = self.check_event_existence(npu_id, distinguishable_event_dict, definitive_event_dict)
            event_list.extend(events)
        return event_list


class HBMInfoParser(PairInfoParser):
    CHIP_MODEL_PARSER = NpuInfoLineParser(id_regex=r"npu-smi info -i (\d{1,3}) -c (\d)",
                                          info_regex=r"npu-smi info -i \d{1,3} -c \d -t board",
                                          detail_regex=r"\s{2}Chip Name\s{22}: (\w{1,15})",
                                          detail_func=None)

    HBM_INFO_PARSER = NpuInfoLineParser(id_regex=r"npu-smi info -i (\d{1,3})(?: -c (\d))?",
                                        info_regex=r"npu-smi info -i \d{1,3}(?: -c \d)? -t usages",
                                        detail_regex=r"\s{2}HBM Capacity\(MB\)\s{15}: (\d{4,6})",
                                        detail_func=None)

    RATED_HBM_DICT = {
        "300I A2": 32768,
        "300T A2": 32768,
        "910A": 32768,
        "910ProA": 32768,
        "910ProB": 32768,
        "910B": 32768,
        "910B1": 65536,
        "910B2": 65536,
        "910B3": 65536,
        "910B4": 32768,
        "Ascend910_939": 65536
    }

    HBM_ABNORMAL_THRESHOLD = 0.95

    def __init__(self, end_time: str):
        super().__init__(end_time)

    def parse_pair_info(self, distinguishable_info_dict: dict, definitive_info_dict: dict, event_message: str):
        """
        Parse chip model and hbm, store them in two dicts
        :param distinguishable_info_dict: the info of chip model for distinguishing
        :param definitive_info_dict: the info of current hbm that is definitive for the wondering result
        :param event_message: event message used for parsing
        """
        model_info = self.CHIP_MODEL_PARSER.parse(event_message)
        if model_info:
            distinguishable_info_dict[model_info.get("npu_id")] = model_info
            return
        hbm_info = self.HBM_INFO_PARSER.parse(event_message)
        if hbm_info:
            definitive_info_dict[hbm_info.get("npu_id")] = hbm_info

    def check_event_existence(self, npu_id: str, distinguishable_info_dict: dict, definitive_info_dict: dict) -> list:
        """
        Check the existence of hbm abnormal fault depends on logics of subclass
        :param npu_id: str to determining npu id
        :param distinguishable_info_dict: the info of chip model for distinguishing the target info
        :param definitive_info_dict: the info of hbm that is definitive for the wondering result
        :return: event list or empty list
        """
        event_list = []
        model = distinguishable_info_dict.get("detail_info", ["unknown"])[0]
        hbm = definitive_info_dict.get("detail_info", ["0"])[0]
        # lower than 95% of rated hbm are considered as abnormal
        if int(hbm) < int(self.RATED_HBM_DICT.get(model, 0)) * self.HBM_ABNORMAL_THRESHOLD:
            event_list.append({
                "occur_time": self.end_time,
                "key_info": definitive_info_dict.get("key_info"),
                "event_code": HBM_ABNORMAL_FAULT,
                "source_device": npu_id,
                "source_file": "npu_info_before/after.txt"
            })
        return event_list


class OpticalInfoParser(PairInfoParser):
    POD_BOARD_ID = {"0x30", "0x31", "0x32", "0x34"}

    def __init__(self, end_time: str):
        super().__init__(end_time)
        self.optical_info_parse = NpuInfoLineParser(id_regex=r"hccn_tool -i (\d{1,3})",
                                                    info_regex=r"hccn_tool -i \d{1,3} -optical -g",
                                                    detail_regex="",
                                                    detail_func=self.regex_detail_func)
        self.board_info_parse = NpuInfoLineParser(id_regex=r"npu-smi info -i (\d{1,3}) -c (\d)",
                                                  info_regex=r"npu-smi info -i \d{1,3} -c \d -t board",
                                                  detail_regex=r"\s{2}Board ID\s{23}: (\w{4})",
                                                  detail_func=None)

    @staticmethod
    def _regex_power(type_str: str, desc) -> tuple:
        """
        Matching optical power information
        :param type_str: TX or Rx
        :param desc: original information
        :return: check whether the fault occurs, key info
        """
        power_keyword = [f"{type_str} Power", f"{type_str}Power High Thres", f"{type_str}Power Low Thres"]
        pattern = r"\s{0,20}: (\d{0,10}\.?\d{0,10}) mW"
        power_regex = re.compile(power_keyword[0] + r"\d{0,3}" + pattern)
        power_ret = power_regex.findall(desc)
        power_high_regex = re.compile(power_keyword[1] + pattern)
        power_high_th = power_high_regex.findall(desc)
        power_low_regex = re.compile(power_keyword[2] + pattern)
        power_low_th = power_low_regex.findall(desc)
        if not power_ret or not power_high_th or not power_low_th:
            return False, []
        key_info = [line for line in desc.split('\n') if any(keyword in line for keyword in power_keyword)]
        for value in power_ret:
            value_f = float(value)
            if value_f < float(power_low_th[0]) or value_f > float(power_high_th[0]):
                return True, key_info
        return False, []

    @staticmethod
    def _regex_los_or_lol(type_str: str, desc) -> tuple:
        """
        Matching Tx Los/LoL or Rx Los/LoL information
        :param type_str: Tx Los/LoL or Rx Los/LoL
        :param desc: original information
        :return: check whether the fault occurs, key info
        """
        normal_value = "0x0"
        keyword = f"{type_str} Flag"
        los_regex = re.compile(keyword + r"\s{0,20}: (.{0,15})")
        los_ret = los_regex.findall(desc)
        if not los_ret:
            return False, []
        key_info = [line for line in desc.split('\n') if keyword in line]
        if los_ret[0] == normal_value or los_ret[0] == "NA":
            return False, key_info
        return True, key_info

    def regex_detail_func(self, desc: str) -> list:
        """
        Matching multiple fault details
        :param desc: original information
        :return: list of fault details
        """
        details = []
        cmd = [f"/usr/local/Ascend/driver/tools/hccn_tool -i {self.optical_info_parse.npu_id} -optical -g"]
        present_regex = re.compile(r"present\s{0,20}: not present")
        present_ret = present_regex.findall(desc)
        if present_ret:
            details.append((OPTICAL_MODULE_NOT_PRESENT, "\n".join(cmd + present_ret)))

        tx_flag, tx_info = self._regex_power("Tx", desc)
        rx_flag, rx_info = self._regex_power("Rx", desc)
        if tx_flag or rx_flag:
            details.append((OPTICAL_POWER_FAULT, "\n".join(cmd + tx_info + rx_info)))

        tx_los_flag, tx_los_info = self._regex_los_or_lol("Tx Los", desc)
        rx_los_flag, rx_los_info = self._regex_los_or_lol("Rx Los", desc)
        if tx_los_flag or rx_los_flag:
            details.append((OPTICAL_MODULE_NOT_RX_OR_TX_FAULT, "\n".join(cmd + tx_los_info + rx_los_info)))

        tx_lol_flag, tx_lol_info = self._regex_los_or_lol("Tx LoL", desc)
        rx_lol_flag, rx_lol_info = self._regex_los_or_lol("Rx LoL", desc)
        if tx_lol_flag or rx_lol_flag:
            details.append((OPTICAL_MODULE_OUT_OF_LOCK_FAULT, "\n".join(cmd + tx_lol_info + rx_lol_info)))
        return details

    def parse_pair_info(self, distinguishable_info_dict: dict, definitive_info_dict: dict, event_message: str):
        """
        Parse board id and presence of optical module, store them in two dicts
        :param distinguishable_info_dict: the info of board id for distinguishing the target info
        :param definitive_info_dict: the info of presence of optical module that is definitive for the wondering result
        :param event_message: event message used for parsing
        """
        board_info = self.board_info_parse.parse(event_message)
        if board_info:
            distinguishable_info_dict[board_info.get("npu_id")] = board_info
            return
        optical_info = self.optical_info_parse.parse(event_message)
        if optical_info:
            definitive_info_dict[optical_info.get("npu_id")] = optical_info

    def check_event_existence(self, npu_id: str, distinguishable_info_dict: dict, definitive_info_dict: dict) -> list:
        """
        Check the existence of event of optical module not present depends on logics of subclass
        :param npu_id: str to determining npu id
        :param distinguishable_info_dict: the info of board id for distinguishing the target info
        :param definitive_info_dict: the info of presence of optical module that is definitive for the wondering result
        :return: event list or empty list
        """
        event_list = []
        board_id = distinguishable_info_dict.get("detail_info", ["unknown"])[0]
        if board_id in self.POD_BOARD_ID:
            return event_list
        detail_info = definitive_info_dict.get("detail_info", [])
        for event_code, key_info in detail_info:
            event_list.append({
                "occur_time": self.end_time,
                "key_info": key_info,
                "event_code": event_code,
                "source_device": npu_id,
                "source_file": "npu_info_before/after.txt"
            })
        return event_list


class NetHealthInfoParser(GeneralInfoParser):
    NET_HEALTH_PARSER = NpuInfoLineParser(id_regex=r"hccn_tool -i (\d{1,3})",
                                          info_regex=r"hccn_tool -i \d{1,3} -net_health -g",
                                          detail_regex=r"net health status: "
                                                       r"(\w{6}\s\w{2}\s\w{3}|\w{4,7}\s\w{4,8}|\w{4,11})",
                                          detail_func=None)

    def __init__(self, end_time: str):
        super().__init__(end_time)

    def parse_info(self, name: str, event_message: str):
        """
        Parse net health info into info dict based on name and event_message
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        net_health_info = self.NET_HEALTH_PARSER.parse(event_message)
        if net_health_info:
            self.info_dict.setdefault(name, []).append(net_health_info)

    def processing_info(self) -> list:
        """
        Processing net health info and return a list of events
        :return: event list
        """
        event_list = []
        for net_health_info_list in self.info_dict.values():
            for net_health_info in net_health_info_list:
                net_health_event = self._check_net_health(net_health_info)
                if net_health_event:
                    event_list.append(net_health_event)
        return event_list

    def _check_net_health(self, net_health_info: dict):
        """
        Check net health, return non-empty dict only if the status are focused
        :param net_health_info: dict of net health info
        :return: event dict
        """
        focused_status = {"Socket failed", "Receive timeout", "Unreachable", "Detect ip set"}
        if net_health_info.get("detail_info", ["unknown"])[0] in focused_status:
            return {
                "occur_time": self.end_time,
                "key_info": net_health_info.get("key_info"),
                "event_code": GENERAL_NET_HEALTH_FAULT,
                "complement": net_health_info.get("detail_info"),
                "source_device": net_health_info.get("npu_id"),
                "source_file": "npu_info_before/after.txt"
            }
        return {}


class FecModeParser(GeneralInfoParser):
    FEC_INFO_PARSER = NpuInfoLineParser(id_regex=r"hccn_tool -i (\d{1,3})",
                                        info_regex=r"hccn_tool -i \d{1,3} -fec -g",
                                        detail_regex=r"fec mode: (\w{2}) FEC mode",
                                        detail_func=None)

    def __init__(self, end_time: str):
        super().__init__(end_time)

    def parse_info(self, name: str, event_message: str):
        """
        Parse fec info into info dict based on name and event_message
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        fec_mode_info = self.FEC_INFO_PARSER.parse(event_message)
        if fec_mode_info:
            self.info_dict.setdefault(name, []).append(fec_mode_info)

    def processing_info(self) -> list:
        """
        Processing fec info and return a list of events
        :return: event list
        """
        event_list = []
        for fec_info_list in self.info_dict.values():
            for fec_info in fec_info_list:
                fec_event = self._check_fec_status(fec_info)
                if fec_event:
                    event_list.append(fec_event)
        return event_list

    def _check_fec_status(self, fec_mode_info: dict):
        """
        Check fec mode status and report if fault exists
        :param fec_mode_info: dict of fec mode info
        :return: event dict or empty dict
        """
        if "no" in fec_mode_info.get("detail_info", ["unknown"])[0]:
            return {
                "occur_time": self.end_time,
                "key_info": fec_mode_info.get("key_info"),
                "event_code": ABNORMAL_FEC_MODE_FAULT,
                "source_device": fec_mode_info.get("npu_id"),
                "source_file": "npu_info_before/after.txt"
            }
        return {}


class LinkInfoParse(GeneralInfoParser):
    LINK_DOWN_PARSERS = NpuInfoLineParser(id_regex=r"hccn_tool -i (\d{1,3})",
                                          info_regex=r"hccn_tool -i \d{1,3} -link_stat -g",
                                          detail_regex=r"\[devid \d{1,3}\]    (\w{3} \w{3} .?[0-9] [0-9]{2}:[0-9]{2}:"
                                                       r"[0-9]{2} [0-9]{4})    LINK (DOWN|UP)",
                                          detail_func=None)

    def __init__(self, end_time: str):
        super().__init__(end_time)

    @staticmethod
    def _convert_time_format(date_time: str):
        """
        Convert time format
        :param date_time: original time string
        :return: converted time string
        """
        time_format = '%a %b %d %H:%M:%S %Y'
        try:
            date_time_convert = datetime.strptime(date_time, time_format)
            return date_time_convert
        except ValueError as error:
            kg_logger.warning("Time format converting in npu_info failed by %s.", str(error))
            raise InfoIncorrectError(f"Time format converting in npu_info failed by {str(error)}.") from error

    def parse_info(self, name: str, event_message: str):
        """
        Parse link info into info dict based on name and event_message
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        link_down_info = self.LINK_DOWN_PARSERS.parse(event_message)
        if link_down_info:
            self.info_dict.setdefault(name, []).append(link_down_info)

    def processing_info(self):
        """
        Compare the link down info in before file and after file
        :return: filtered after training link down info list
        """
        event_dict_list = []
        compare_dict = dict()
        before_info_dict = self.info_dict.get("before", [])
        for event_dict in before_info_dict:
            npu_id_before = event_dict.get("npu_id")
            occur_before = event_dict.get("detail_info")
            compare_dict[npu_id_before] = occur_before
        after_info_dict = self.info_dict.get("after", [])
        for event_dict in after_info_dict:
            occur_after = event_dict.get("detail_info")
            npu_id = event_dict.get("npu_id")
            occur_before = compare_dict.get(npu_id, "")
            occur_time = self._check_before_and_after(occur_before, occur_after)
            if occur_time:
                event_dict_list.append({
                    "occur_time": occur_time,
                    "source_device": npu_id,
                    "key_info": event_dict.get("key_info"),
                    "event_code": LINK_DOWN_FAULT,
                    "source_file": "npu_info_before/after.txt"
                })
        return event_dict_list

    def _check_before_and_after(self, before_list, after_list):
        """
        Check the link down record by compare before list and after list
        :param before_list: link down record list before training
        :param after_list: link down record list after training
        :return: err time str
        """
        if not before_list or before_list[0] not in after_list:
            check_list = after_list
        else:
            check_list = after_list[:after_list.index(before_list[0])]
        pre_record = None
        # check list: [(time, UP/DOWN),...]
        for record in check_list:
            if record[1] == "UP":
                pre_record = record
                continue
            down_time = self._convert_time_format(record[0])
            if str(down_time) > self.end_time:
                pre_record = record
                continue
            if not pre_record:
                return str(down_time)
            up_time = self._convert_time_format(pre_record[0])
            if (up_time - down_time).seconds > 30:  # valid only when the up and down duration exceeds 30s.
                return str(down_time)
            pre_record = record
        return ""


class HealthInfoParser(GeneralInfoParser):
    HEALTH_CODE_PARSERS = NpuInfoLineParser(id_regex=r"npu-smi info -i (\d{1,3}) -c (\d)",
                                            info_regex=r"npu-smi info -i \d{1,3} -c \d -t health",
                                            detail_regex=r"(Error Code\s{1,100}: .[a-zA-Z0-9 ]{1,100})",
                                            detail_func=lambda p: p.split(": ")[1])  # index 1 is used to get Error Code

    def __init__(self, end_time: str):
        super().__init__(end_time)

    def parse_info(self, name: str, event_message: str):
        """
        Parse health info into info dict based on name and event_message
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        health_info = self.HEALTH_CODE_PARSERS.parse(event_message)
        if health_info:
            self.info_dict.setdefault(name, []).append(health_info)

    def processing_info(self) -> list:
        """
        Processing health info and return a list of events
        :return: event list
        """
        before_list = self.info_dict.get("before", [])
        after_list = self.info_dict.get("after", [])
        npu_id_key = "npu_id"
        detail_info_str = "detail_info"
        empty_default_value = "NA"
        event_list = []
        if len(before_list) != len(after_list):
            kg_logger.warning("The before npu health info list's length not equal to "
                              "the after npu health info list.")
            return event_list
        before_list.sort(key=lambda x: x.get(npu_id_key, NEGATIVE_ONE))
        after_list.sort(key=lambda x: x.get(npu_id_key, NEGATIVE_ONE))
        for err_before, err_after in zip(before_list, after_list):
            if err_before.get(npu_id_key, NEGATIVE_ONE) != err_after.get(npu_id_key, NEGATIVE_ONE):
                kg_logger.warning("The npu id in the before npu info and after file do not match, "
                                  "please check the npu info files.")
                continue

            err_before_detail_info = err_before.get(detail_info_str, [empty_default_value])
            err_after_detail_info = err_after.get(detail_info_str, [empty_default_value])
            if err_after_detail_info[0] == empty_default_value:
                continue
            error_code_add = set(err_after_detail_info[0].split(" ")) - set(err_before_detail_info[0].split(" "))
            if not error_code_add:
                continue
            for error_code in error_code_add:
                error_code = error_code if error_code.startswith("0x") else "0x" + error_code
                event_list.append({"event_code": error_code,
                                   "occur_time": self.end_time,
                                   "source_device": err_after.get(npu_id_key),
                                   "key_info": err_after.get("key_info"),
                                   "source_file": "npu_info_before/after.txt"
                                   })
        return event_list


class PhysicalCardInfoParser(GeneralInfoParser):
    CARD_STATUS_PARSER = NpuInfoLineParser(id_regex=r"(lspci \| grep acce)",
                                           info_regex=r"lspci \| grep acce",
                                           detail_regex=r"(\w{2}:\w{2}\.\w{1})",
                                           detail_func=None)

    BOARD_INFO_PARSER = NpuInfoLineParser(id_regex=r"npu-smi info -i (\d{1,3})(?: -c (\d))?",
                                          info_regex=r"npu-smi info -i \d{1,3}(?: -c \d)? -t board",
                                          detail_regex=r"\s{2}PCIe Bus Info\s{18}: 0000:(\w{2}:\w{2}.\w{1})",
                                          detail_func=str.lower)

    def __init__(self, end_time: str):
        super().__init__(end_time)

    @staticmethod
    def _obtain_physical_dropping_cards(cards_before: dict, cards_after: dict, bus_id_map: dict):
        """
        Use the difference between set to identify and return dropping cards
        :param cards_before: list of cards info after
        :param cards_after: list of cards info after
        :param bus_id_map: map between bus id and device id
        :return: a list of dropping cards
        """
        cards_before_list = cards_before.get("detail_info", [])
        dropping_cards_set = set(cards_before_list) - set(cards_after.get("detail_info", []))
        dropping_cards = []
        for card in dropping_cards_set:
            if card in bus_id_map:
                dropping_cards.append(bus_id_map.get(card))
        return dropping_cards

    def parse_info(self, name: str, event_message: str):
        """
        Parsing and storing card info and board info in info_dict
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        card_info = self.CARD_STATUS_PARSER.parse(event_message)
        if card_info:
            self.info_dict[name] = card_info
            return
        board_info = self.BOARD_INFO_PARSER.parse(event_message)
        if board_info:
            self.info_dict[board_info.get("detail_info")[0]] = board_info.get("npu_id")

    def processing_info(self) -> list:
        """
        Processing card info and return a list of events
        :return: event list
        """
        event_list = []
        cards_before = self.info_dict.get("before", dict())
        cards_after = self.info_dict.get("after", dict())
        if len(cards_before.get("detail_info", [])) > len(cards_after.get("detail_info", [])):
            dropping_cards = self._obtain_physical_dropping_cards(cards_before, cards_after, self.info_dict)
            for card in dropping_cards:
                event_list.append({
                    "occur_time": self.end_time,
                    "key_info": cards_after.get("key_info"),
                    "event_code": PHYSICAL_CARD_DROPPING,
                    "complement": [",".join(dropping_cards)],
                    "source_device": card,
                    "source_file": "npu_info_before/after.txt"
                })
        return event_list


class SoftwareCardInfoParser(GeneralInfoParser):
    BOARD_CMD_ERROR_PARSER = NpuInfoLineParser(id_regex=r"npu-smi info -i (\d{1,3})(?: -c (\d))?",
                                               info_regex=r"npu-smi info -i \d{1,3}(?: -c \d)? -t board",
                                               detail_regex=r"(npu-smi info -i \d{1,3}(?: -c \d)? -t board\n"
                                                            r"Invalid card id.\n"
                                                            r"Error parameter of (?:-i|-c)\n)",
                                               detail_func=None)

    def __init__(self, end_time: str):
        super().__init__(end_time)

    def parse_info(self, name: str, event_message: str):
        """
        Parse board cmd err info into info dict based on name and event_message
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        err_info = self.BOARD_CMD_ERROR_PARSER.parse(event_message)
        if err_info:
            self.info_dict.setdefault(name, []).append(err_info)

    def processing_info(self) -> list:
        """
        Processing err info and return a list of events
        :return: event list
        """
        event_list = []
        err_info_before = self.info_dict.get("before", [])
        err_info_after = self.info_dict.get("after", [])
        if len(err_info_after) > len(err_info_before):
            dropping_cards = set()
            npu_id_before = [npu_id["npu_id"] for npu_id in err_info_before]
            npu_id_after = [npu_id["npu_id"] for npu_id in err_info_after]
            for npu_id in npu_id_after:
                if npu_id not in npu_id_before:
                    dropping_cards.add(npu_id)
            for card in dropping_cards:
                event_list.append({
                    "occur_time": self.end_time,
                    "key_info": "\n".join(err_info_after[0].get("detail_info")),
                    "event_code": SOFTWARE_CARD_DROPPING,
                    "complement": [",".join(dropping_cards)],
                    "source_device": card,
                    "source_file": "npu_info_before/after.txt"
                })
        return event_list


class IpInfoParser(GeneralInfoParser):
    IP_INFO_PARSER = NpuInfoLineParser(id_regex=r"hccn_tool -i (\d{1,3})",
                                       info_regex=r"hccn_tool -i \d{1,3} -ip -g",
                                       detail_regex=r"ipaddr:(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\n"
                                                    r"netmask:(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})",
                                       detail_func=None)

    def __init__(self, end_time: str):
        super().__init__(end_time)

    def parse_info(self, name: str, event_message: str):
        """
        Parse ip info into info dict based on name and event_message
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        ip_info = self.IP_INFO_PARSER.parse(event_message)
        if ip_info:
            self.info_dict.setdefault(name, []).append(ip_info)
            return
        if not self.IP_INFO_PARSER.desc:
            return
        self.info_dict.setdefault(name, []).append({
            "npu_id": self.IP_INFO_PARSER.npu_id, "key_info": self.IP_INFO_PARSER.desc
        })

    def processing_info(self) -> list:
        """
        Processing ip info and return a list of events
        :return: event list
        """
        event_list = []
        for ip_info_list in self.info_dict.values():
            ip_events = self._contrast_ip_info(ip_info_list)
            event_list.extend(ip_events)
        return event_list

    def _contrast_ip_info(self, ip_info_list: list) -> list:
        """
        Check the consistency of the first three segments of ip address and netmask
        :param ip_info_list: list of ip info dict
        :return: event list
        """
        ip_events = []
        if not ip_info_list:
            return ip_events
        ip_list = []
        netmask_list = []
        define_key_info = ("执行以下命令无结果：\n/usr/local/Ascend/driver/tools/hccn_tool -i ${device id} -ip -g\n"
                           "预期执行示例：\n/usr/local/Ascend/driver/tools/hccn_tool -i ${device id} -ip -g\n"
                           "ipaddr:*.*.*.*\nnetmask:*.*.*.*")
        for ip_info in ip_info_list:
            detail_info = ip_info.get("detail_info")
            if detail_info:
                netmask_list.append(detail_info[0][1])
                ip_list.append(detail_info[0][0])
                continue
            ip_events.append({
                "occur_time": self.end_time,
                "key_info": define_key_info,
                "event_code": IP_NOT_CONFIG_FAULT,
                "source_device": ip_info.get("npu_id", "Unknown"),
                "source_file": "npu_info_before/after.txt"
            })
        return ip_events


class VersionInfoParser(GeneralInfoParser):
    DRIVER_VERSION_PARSER = NpuInfoLineParser(id_regex=r"(cat /usr/local/Ascend/driver/version.info)",
                                              info_regex=r"cat /usr/local/Ascend/driver/version.info",
                                              detail_regex=r"Version=([A-Za-z0-9\.]{1,20})",
                                              detail_func=None)

    FIRM_VERSION_PARSER = NpuInfoLineParser(id_regex=r"(cat /usr/local/Ascend/firmware/version.info)",
                                            info_regex=r"cat /usr/local/Ascend/firmware/version.info",
                                            detail_regex=r"Version=([A-Za-z0-9\.]{1,20})",
                                            detail_func=None)

    NNAE_VERSION_PARSER = NpuInfoLineParser(id_regex=r"(cat /usr/local/Ascend/nnae/latest/ascend_nnae_install.info)",
                                            info_regex=r"cat /usr/local/Ascend/nnae/latest/ascend_nnae_install.info",
                                            detail_regex=r"version=([A-Za-z0-9\.]{1,20})",
                                            detail_func=None)

    CAAN_X86_64_VERSION_PARSER = NpuInfoLineParser(id_regex=r"(cat /usr/local/Ascend/ascend-toolkit/latest/"
                                                            r"x86_84-linux/ascend_toolkit_install.info)",
                                                   info_regex=r"cat /usr/local/Ascend/ascend-toolkit/latest/"
                                                              r"x86_84-linux/ascend_toolkit_install.info",
                                                   detail_regex=r"version=([A-Za-z0-9\.]{1,20})",
                                                   detail_func=None)

    CAAN_AARCH64_VERSION_PARSER = NpuInfoLineParser(id_regex=r"(cat /usr/local/Ascend/ascend-toolkit/latest/"
                                                             r"aarch64-linux/ascend_toolkit_install.info)",
                                                    info_regex=r"cat /usr/local/Ascend/ascend-toolkit/latest/"
                                                               r"aarch64-linux/ascend_toolkit_install.info",
                                                    detail_regex=r"version=([A-Za-z0-9\.]{1,20})",
                                                    detail_func=None)

    PYTORCH_VERSION_PARSER = NpuInfoLineParser(id_regex=r'(pip list \| grep "torch ")',
                                               info_regex=r'pip list \| grep "torch "',
                                               detail_regex=r"[\d\.]{1,10}",
                                               detail_func=None)

    TORCH_NPU_VERSION_PARSER = NpuInfoLineParser(id_regex=r'(pip list \| grep torch-npu)',
                                                 info_regex=r'pip list \| grep torch-npu',
                                                 detail_regex=r"[a-zA-Z\d\.]{1,30}",
                                                 detail_func=None)

    MINDSPORE_VERSION_PARSER = NpuInfoLineParser(id_regex=r'(pip list \| grep "mindspore ")',
                                                 info_regex=r'pip list \| grep "mindspore "',
                                                 detail_regex=r"[\d\.]{1,10}",
                                                 detail_func=None)

    def __init__(self):
        super().__init__()

    def parse_info(self, name: str, event_message: str):
        """
        Parsing and storing version info in info_dict
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        driver_version = self.DRIVER_VERSION_PARSER.parse(event_message).get("detail_info", [])
        if driver_version:
            self.info_dict["driver_version"] = driver_version[0]
            return
        firm_version = self.FIRM_VERSION_PARSER.parse(event_message).get("detail_info", [])
        if firm_version:
            self.info_dict["firm_version"] = firm_version[0]
            return
        nnae_version = self.NNAE_VERSION_PARSER.parse(event_message).get("detail_info", [])
        if nnae_version:
            self.info_dict["nnae_version"] = nnae_version[0]
            return
        x86_64_cann_version = self.CAAN_X86_64_VERSION_PARSER.parse(event_message).get("detail_info", [])
        if x86_64_cann_version:
            self.info_dict["cann_version"] = x86_64_cann_version[0]
            return
        aarch64_cann_version = self.CAAN_AARCH64_VERSION_PARSER.parse(event_message).get("detail_info", [])
        if aarch64_cann_version:
            self.info_dict["cann_version"] = aarch64_cann_version[0]
            return
        pytorch_version = self.PYTORCH_VERSION_PARSER.parse(event_message).get("detail_info", [])
        if pytorch_version:
            self.info_dict["pytorch_version"] = pytorch_version[0]
            return
        torch_npu_version = self.TORCH_NPU_VERSION_PARSER.parse(event_message).get("detail_info", [])
        if torch_npu_version and torch_npu_version[-1] != 'npu':
            self.info_dict["torch_npu_version"] = torch_npu_version[-1]
            return
        mindspore_version = self.MINDSPORE_VERSION_PARSER.parse(event_message).get("detail_info", [])
        if mindspore_version:
            self.info_dict["mindspore_version"] = mindspore_version[0]

    def processing_info(self) -> list:
        """
        Assemble and return a composed version dict with keys of version label and values of version
        """
        return [dict(event_code="VERSION_INFO", **self.info_dict)]


class NpuSmiInfoParser(GeneralInfoParser):
    NPU_SMI_INFO_PARSER = NpuInfoLineParser(id_regex=r"(/usr/local/bin/npu-smi info)",
                                            info_regex=r"/usr/local/bin/npu-smi info$|/usr/local/bin/npu-smi info\n",
                                            detail_regex=r"npu-smi [A-Za-z0-9\.]{0,20}\s{0,30}Version: [A-Za-z0-9\.]{"
                                                         r"0,20}",
                                            detail_func=None)

    def __init__(self, end_time: str):
        super().__init__(end_time)

    def parse_info(self, name: str, event_message: str):
        """
        Parsing and storing npu-smi info in info_dict
        :param name: before or after refer to the time of conducting collecting cmds
        :param event_message: event message used for parsing
        """
        if self.NPU_SMI_INFO_PARSER.parse(event_message).get("detail_info", []):
            return
        if self.NPU_SMI_INFO_PARSER.desc:
            self.info_dict.setdefault(name, []).append({"key_info": self.NPU_SMI_INFO_PARSER.desc})

    def processing_info(self) -> list:
        """
        Processing npu-smi info and return a list of events
        :return: event list
        """
        event_list = []
        define_key_info = "执行以下命令无结果或报错：\n"
        for info in chain(*self.info_dict.values()):
            if info.get("detail_info"):
                continue
            event_list.append({
                "occur_time": self.end_time,
                "key_info": define_key_info + info.get("key_info", ""),
                "event_code": NPU_DRIVER_FAULT,
                "source_device": "Unknown",
                "source_file": "npu_info_before/after.txt"
            })
        return event_list


@dataclass
class DeviceInfo:
    _info_line: str = ""
    npu_id: str = ""
    chip_id: str = ""
    chip_logic_id: str = ""

    def __post_init__(self):
        self.npu_id, self.chip_id, self.chip_logic_id, _ = self._info_line.strip().split(None, 3)


class ConfigInitializer:
    @staticmethod
    def initialize_config(snippet: str):
        """
        Initialize the global config of config dict, provide a dict of (npu_id, chip_id) -> chip_logic_id
        """
        local_config = dict()
        least_unpack_num = 4
        for line in snippet.strip().splitlines()[1:]:
            if len(line.split()) < least_unpack_num:
                continue
            device_info = DeviceInfo(line.strip())
            if device_info.chip_logic_id.isdigit():
                local_config[(device_info.npu_id, device_info.chip_id)] = device_info.chip_logic_id
        global LOGIC_ID_CONFIG
        LOGIC_ID_CONFIG = local_config


class NpuInfoParser(FileParser):
    SOURCE_FILE = NPU_INFO_SOURCE
    DATETIME_STR = "Datetime"
    TARGET_FILE_PATTERNS = "npu_info_path"
    COMMAND_PREFIX_SET = {"/usr/local", "cat /usr", "lspci |", "pip list"}
    DEVICE_INFO_CMD = "npu-smi info -m"
    DATETIME_PATTERN = re.compile(DATETIME_REGEX)

    def __init__(self, params):
        """
        The NPU Info parser.
        Note: This parser does not filter by time.
        """
        super().__init__(params)

    def parse(self, parse_ctx: KGParseCtx, task_id):
        """
        Parse log file
        :param parse_ctx: file path
        :param task_id: the task unique id
        :return: parse descriptor result
        """
        npu_info = self.find_log(parse_ctx.parse_file_path)
        if not npu_info:
            return [], {}
        events_list = []
        snippets_param = dict()
        file_time_after = None
        kwd_before = "before"
        kwd_after = "after"
        self.is_sdk_input = parse_ctx.is_sdk_input
        for file_source in npu_info:
            filename = self._get_filename(file_source)
            is_npu_before = (self.is_sdk_input and filename == "npu_info_before.txt") or \
                            (filename == "npu_info_before.txt" and os.path.exists(file_source))
            if is_npu_before:
                snippets_param[kwd_before], self.start_time = self._split_into_snippets(file_source)
                continue
            is_npu_after = (self.is_sdk_input and filename == "npu_info_after.txt") or \
                           (filename == "npu_info_after.txt" and os.path.exists(file_source))
            if is_npu_after:
                snippets_param[kwd_after], self.end_time = self._split_into_snippets(file_source)
                if self.is_sdk_input:
                    file_time_after = getattr(file_source, "modification_time", "") or KG_MAX_TIME
                else:
                    file_time_after = datetime.fromtimestamp(
                        os.path.getmtime(file_source)).strftime("%Y-%m-%d %H:%M:%S.%f")
        if kwd_before not in snippets_param or kwd_after not in snippets_param:
            kg_logger.warning("The npu info files is incomplete. The before or after file does not exist.")
            raise FileNotExistError("The npu info files is incomplete. The before or after file does not exist.")
        log_before = snippets_param.get(kwd_before, [])
        device_info_log_idx = 1
        if log_before and len(log_before) > 1:
            if log_before[device_info_log_idx] and self.DEVICE_INFO_CMD in log_before[device_info_log_idx]:
                ConfigInitializer.initialize_config(log_before[device_info_log_idx])
        self._update_train_time()
        end_time = self.params.get("end_time") or file_time_after  # have end_time, use end_time as occur_time
        for parser in [LinkInfoParse(end_time), HealthInfoParser(end_time), HBMInfoParser(end_time),
                       IpInfoParser(end_time), PhysicalCardInfoParser(end_time), SoftwareCardInfoParser(end_time),
                       OpticalInfoParser(end_time), FecModeParser(end_time), NetHealthInfoParser(end_time),
                       VersionInfoParser(), NpuSmiInfoParser(end_time)]:
            events_list.extend(parser.parse(snippets_param))
        return events_list, {}

    def process_single_snippet(self, log_lines: list):
        snippet_list = []
        current_snippet = ""
        date_time = None
        for line in log_lines:
            if line.startswith(self.DATETIME_STR):
                date_time_str = self.DATETIME_PATTERN.search(line)
                date_time = date_time_str[0] if date_time_str else date_time
            if any(line.startswith(cmd_prefix) for cmd_prefix in self.COMMAND_PREFIX_SET) \
                    and current_snippet.strip():
                snippet_list.append(current_snippet.strip())
                current_snippet = line
                continue
            current_snippet += line
        if current_snippet.strip():
            snippet_list.append(current_snippet.strip())
        return snippet_list, date_time

    def _split_into_snippets(self, file_source: Union[str, LogInfoSaver]):
        """
        Split file source into snippets then store them in a list
        :param file_source: path of npu info file
        :return: list of divided snippets, datetime
        """
        if self.is_sdk_input:
            return self.process_single_snippet(file_source.log_lines)
        with safe_read_open_with_size(file_source, mode='r', encoding='utf-8') as file_stream:
            return self.process_single_snippet(file_stream.readlines())

    def _update_train_time(self):
        """
        Update the train time interval of this train job
        """
        pre_start_time = self.params.get("start_time")
        pre_end_time = self.params.get("end_time")
        if not pre_start_time or self.start_time and self.start_time < pre_start_time:
            self.params.update({"start_time": self.start_time})
        if not pre_end_time or self.end_time and self.end_time > pre_end_time:
            self.params.update({"end_time": self.end_time})
