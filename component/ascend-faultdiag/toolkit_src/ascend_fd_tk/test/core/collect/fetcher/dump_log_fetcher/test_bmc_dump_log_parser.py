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

"""BMC dump 日志解析器单元测试：IP/SN 解析、光模块历史 CSV 与整体 parse。"""

import os
import tempfile
import unittest

from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.bmc.base import BmcDumpLogDataType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.bmc.bmc_dump_log_parser import BmcDumpLogParser


class TestBmcDumpLogParserMethods(unittest.TestCase):
    def test_parse_bmc_ip(self):
        parser = BmcDumpLogParser("", "")
        self.assertEqual(parser.parse_bmc_ip(["IP Address : 10.1.1.1\n"]), "10.1.1.1")
        self.assertEqual(parser.parse_bmc_ip(["no ip here"]), "")

    def test_parse_host_sn_num(self):
        lines = [
            "FRU Device Description : Builtin FRU Device (FRUID 0, BMC)",
            "Product Manufacturer : HUAWEI",
            "System Serial Number : 2102311ABCD",
        ]
        parser = BmcDumpLogParser("", "")
        self.assertEqual(parser.parse_host_sn_num(lines), "2102311ABCD")
        self.assertEqual(parser.parse_host_sn_num(["no fru here"]), "")

    def test_parse_optical_module_history_info_log(self):
        with tempfile.TemporaryDirectory() as tmp:
            parse_dir = os.path.join(tmp, "logs")
            log_dir = os.path.join(parse_dir, "AppDump", "network_adapter", "optical_module")
            os.makedirs(log_dir)
            csv_path = os.path.join(log_dir, "optical_module_history_info_log.csv")
            with open(csv_path, "w", newline="", encoding="utf-8") as f:
                f.write(
                    "LogTime,Current,LinkStatus\n"
                    "2026-08-29 10:00:00,0.1,LinkDown\n"
                    "LogTime,CurrentMax,Optical\n"
                    "2026-08-29 11:00:00,1.2,OK\n"
                )
            result = BmcDumpLogParser("", parse_dir).parse_optical_module_history_info_log()
            self.assertEqual(len(result), 1)
            self.assertEqual(result[0]["LinkStatus"], "LinkDown")
            self.assertEqual(result[0]["LogTime"], "2026-08-29 10:00:00")

    def test_parse_optical_module_history_info_log_missing(self):
        with tempfile.TemporaryDirectory() as tmp:
            parser = BmcDumpLogParser("", os.path.join(tmp, "no_logs"))
            self.assertEqual(parser.parse_optical_module_history_info_log(), [])


class TestBmcDumpLogParserParse(unittest.TestCase):
    def _make_logs(self, parse_dir):
        network_dir = os.path.join(parse_dir, "AppDump", "bmc_network")
        os.makedirs(network_dir)
        with open(os.path.join(network_dir, "network_info.txt"), "w", encoding="utf-8") as f:
            f.write("IP Address : 10.1.1.1\n")

        fru_dir = os.path.join(parse_dir, "AppDump", "frudata")
        os.makedirs(fru_dir)
        with open(os.path.join(fru_dir, "fruinfo.txt"), "w", encoding="utf-8") as f:
            f.write("FRU Device Description : Builtin FRU Device (FRUID 0, BMC)\nSystem Serial Number : 2102311ABCD\n")

        event_dir = os.path.join(parse_dir, "AppDump", "event")
        os.makedirs(event_dir)
        with open(os.path.join(event_dir, "sel.txt"), "w", encoding="utf-8") as f:
            f.write("1 | 2026-08-29 | event one\n")
        with open(os.path.join(event_dir, "current_event.txt"), "w", encoding="utf-8") as f:
            f.write("health: ok\n")

        sensor_dir = os.path.join(parse_dir, "AppDump", "sensor")
        os.makedirs(sensor_dir)
        with open(os.path.join(sensor_dir, "sensor_info.txt"), "w", encoding="utf-8") as f:
            f.write("temp 25\n")

        optical_dir = os.path.join(parse_dir, "AppDump", "network_adapter", "optical_module")
        os.makedirs(optical_dir)
        csv_file_name = "optical_module_history_info_log.csv"
        with open(os.path.join(optical_dir, csv_file_name), "w", newline="", encoding="utf-8") as f:
            f.write("LogTime,Current,LinkStatus\n2026-08-29 10:00:00,0.1,LinkDown\n")

    def test_parse_full_dir(self):
        with tempfile.TemporaryDirectory() as tmp:
            parse_dir = os.path.join(tmp, "logs")
            os.makedirs(parse_dir)
            self._make_logs(parse_dir)
            data = BmcDumpLogParser("", parse_dir).parse()
            self.assertEqual(data[BmcDumpLogDataType.BMC_IP.name], "10.1.1.1")
            self.assertEqual(data[BmcDumpLogDataType.SN_NUM.name], "2102311ABCD")
            self.assertEqual(data[BmcDumpLogDataType.SEL_INFO.name], "1 | 2026-08-29 | event one\n")
            self.assertEqual(data[BmcDumpLogDataType.SENSOR_INFO.name], "temp 25\n")
            self.assertEqual(data[BmcDumpLogDataType.HEALTH_EVENTS.name], "health: ok\n")
            self.assertEqual(len(data[BmcDumpLogDataType.OP_HISTORY_INFO_LOG.name]), 1)

    def test_parse_missing_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            data = BmcDumpLogParser("", tmp).parse()
            self.assertEqual(data[BmcDumpLogDataType.OP_HISTORY_INFO_LOG.name], [])


if __name__ == "__main__":
    unittest.main()
