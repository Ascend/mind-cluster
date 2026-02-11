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
import os.path
import re
from typing import List

from toolkit.core.common.json_obj import JsonObj
from toolkit.core.log_parser.base import FindResult
from toolkit.core.log_parser.local_log_parser import LocalLogParser
from toolkit.core.log_parser.parse_config import swi_diag_info_log_config
from toolkit.core.model.switch import PortDownStatus, PortDownStatusLaneInfo
from toolkit.utils import helpers
from toolkit.utils.table_parser import TableParser


class DiagInfoParseResult(JsonObj):

    def __init__(self, swi_name="", find_log_results: List[FindResult] = None,
                 port_down_status: List[PortDownStatus] = None):
        self.swi_name = swi_name
        self.find_log_results = find_log_results
        self.port_down_status = port_down_status


class CollectDiagInfoLogParser:
    _NAME_LOG_RELA_PATH = os.path.join("logfile_slot_1", "tempdir", "diag.log", "diag.log")
    _PORT_DOWN_STATUS_PATH = os.path.join("slot_1", "tempdir", "port_down_status.log")
    _NAME_PATTERN = re.compile(r"\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}[^ ]{1,50} ([^ ]{1,50})")
    _LINK_SNR_INFO_START = "Diagnose Information Start----------"
    _LINK_SNR_INFO_PATTERN = (r"(?P<date>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}.\d{1,3}).{1,100}Unit:"
                              r"(?P<swi_chip_id>\d) Port:(?P<port_id>\d{1,3}).{1,200}Crc Error Cnt:"
                              r"\s{0,10}(?P<crc_error_cnt>\d{1,15})\nFec Error Cnt:\s{0,10}(?P<frc_error_cnt>\d{1,15})")

    @classmethod
    def parse(cls, log_dir: str) -> DiagInfoParseResult:
        swi_name = cls._find_swi_name(log_dir)
        pattern_map = {}
        for config in swi_diag_info_log_config.SWI_DIAG_INFO_LOG_CONFIG:
            pattern_map[config.keyword_config.pattern_key] = config
        find_log_results = asyncio.run(LocalLogParser().find(log_dir, pattern_map))
        port_down_status = cls._find_port_down_status_info(log_dir)
        return DiagInfoParseResult(swi_name, find_log_results, port_down_status)

    @classmethod
    def _find_swi_name(cls, log_dir: str) -> str:
        name_log_path = os.path.join(log_dir, cls._NAME_LOG_RELA_PATH)
        if not os.path.exists(name_log_path):
            return ""
        # 就搜100行
        max_search_line_num = 100
        search_cnt = 0
        with open(name_log_path, "r", encoding="utf8") as f:
            while f.readable():
                line = f.readline()
                search = cls._NAME_PATTERN.search(line)
                if search:
                    return search.group(1)
                search_cnt += 1
                if search_cnt >= max_search_line_num:
                    return ""
        return ""

    @classmethod
    def _find_port_down_status_info(cls, log_dir: str = "") -> List[PortDownStatus]:
        link_snr_log_path = os.path.join(log_dir, cls._PORT_DOWN_STATUS_PATH)
        if not os.path.exists(link_snr_log_path):
            return []
        try:
            with open(link_snr_log_path, 'r', encoding="utf8") as f:
                content = f.read()
        except Exception:
            return []
        if not content or cls._LINK_SNR_INFO_START not in content:
            return []
        parts = helpers.split_str(content, cls._LINK_SNR_INFO_START)
        if not parts or len(parts) % 2 != 0:
            return []
        results = []
        title_dict = {"lane_id": "laneId", "snr": "snr", "data_rate": "data-rate(MHz)",
                      "tx_amp_ctl_en": "tx-amp-ctl-en", "los_status": "losStatus"}
        for part in parts:
            part = part.strip()
            search_info = re.search(cls._LINK_SNR_INFO_PATTERN, part, re.DOTALL)
            info_dict = search_info and search_info.groupdict()
            if not info_dict:
                continue
            port_down_status = PortDownStatus.from_dict(info_dict)
            lane_infos = []
            for table in TableParser.parse(part, title_dict, end_sign="---"):
                lane_infos.append(PortDownStatusLaneInfo.from_dict(table))
            port_down_status.lane_info = lane_infos
            results.append(port_down_status)
        return results
