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

"""Host dump 日志解析器单元测试：V1/V2/V3 配置、名称目录与端到端 parse。"""

import os
import tempfile
import unittest

from ascend_fd_tk.core.collect.collect_config import ToolLogCollectionDataType, ToolLogCollectionSplitType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.host.host_dump_log_parser import (
    HostDumpLogParserV1,
    HostDumpLogParserV2,
    HostDumpLogParserV3,
)
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.host.host_log_parser_builder import HostLogParserBuilder


class TestHostDumpLogParserConfigs(unittest.TestCase):
    def test_v1_configs_and_patterns(self):
        parser = HostDumpLogParserV1("", "")
        self.assertTrue(parser.get_parse_configs())
        pattern_map = parser.get_pattern_map()
        self.assertIn(ToolLogCollectionSplitType.DEVICE_ID, pattern_map)
        self.assertIn(ToolLogCollectionSplitType.DEVICE_CHIP_ID, pattern_map)

    def test_v2_configs_and_patterns(self):
        parser = HostDumpLogParserV2("", "")
        self.assertTrue(parser.get_parse_configs())
        pattern_map = parser.get_pattern_map()
        self.assertIn(ToolLogCollectionSplitType.DEVICE_ID, pattern_map)
        self.assertIn(ToolLogCollectionSplitType.DEVICE_CHIP_ID, pattern_map)

    def test_v3_configs_and_patterns(self):
        parser = HostDumpLogParserV3("", "")
        self.assertTrue(parser.get_parse_configs())
        self.assertIn(ToolLogCollectionSplitType.DEVICE_ID, parser.get_pattern_map())

    def test_get_name_dir(self):
        with tempfile.TemporaryDirectory() as tmp:
            parse_dir = os.path.join(tmp, "logs")
            os.makedirs(parse_dir)
            self.assertEqual(HostDumpLogParserV1(tmp, parse_dir).get_name_dir(), os.path.abspath(parse_dir))
            self.assertEqual(HostDumpLogParserV2(tmp, parse_dir).get_name_dir(), os.path.abspath(tmp))
            self.assertEqual(HostDumpLogParserV3(tmp, parse_dir).get_name_dir(), os.path.abspath(parse_dir))


class TestHostDumpLogParserParse(unittest.TestCase):
    def test_v1_parse_hccn_tool_log(self):
        hccn_log = (
            "================ lldp ================\n"
            "hccn_tool -i 0 -lldp -g\n"
            "lldp data 0\n"
            "hccn_tool -i 1 -lldp -g\n"
            "lldp data 1\n"
            "----------------\n"
            "================ speed -g ================\n"
            "hccn_tool -i 0 -speed -g\n"
            "speed 100\n"
            "----------------\n"
        )
        with tempfile.TemporaryDirectory() as tmp:
            parse_dir = os.path.join(tmp, "logs")
            os.makedirs(parse_dir)
            with open(os.path.join(parse_dir, "hccn_tool.log"), "w", encoding="utf-8") as f:
                f.write(hccn_log)
            data = HostDumpLogParserV1(tmp, parse_dir).parse()
            self.assertEqual(data[ToolLogCollectionDataType.LLDP.name]["0"], "lldp data 0")
            self.assertEqual(data[ToolLogCollectionDataType.LLDP.name]["1"], "lldp data 1")
            self.assertEqual(data[ToolLogCollectionDataType.SPEED.name]["0"], "speed 100")
            self.assertEqual(data[ToolLogCollectionDataType.HOST_ID.name], "logs")

    def test_v2_parse_net_conf_log(self):
        net_conf = (
            "================ lldp ================\n"
            "device id=0\n"
            "lldp data 0\n"
            "device id=1\n"
            "lldp data 1\n"
            "----------------\n"
        )
        with tempfile.TemporaryDirectory() as tmp:
            parse_dir = os.path.join(tmp, "logs")
            hccn_dir = os.path.join(parse_dir, "hccn_log")
            os.makedirs(hccn_dir)
            with open(os.path.join(hccn_dir, "net_conf.log"), "w", encoding="utf-8") as f:
                f.write(net_conf)
            data = HostDumpLogParserV2(tmp, parse_dir).parse()
            self.assertEqual(data[ToolLogCollectionDataType.LLDP.name]["0"], "lldp data 0")
            self.assertEqual(data[ToolLogCollectionDataType.LLDP.name]["1"], "lldp data 1")

    def test_v3_parse_lldp_log(self):
        lldp_log = "============ lldp ============\n====> 0\nlldp data 0\n====> 1\nlldp data 1\n------------\n"
        with tempfile.TemporaryDirectory() as tmp:
            parse_dir = os.path.join(tmp, "logs")
            os.makedirs(parse_dir)
            with open(os.path.join(parse_dir, "lldp.log"), "w", encoding="utf-8") as f:
                f.write(lldp_log)
            data = HostDumpLogParserV3(tmp, parse_dir).parse()
            self.assertEqual(data[ToolLogCollectionDataType.LLDP.name]["0"], "lldp data 0")
            self.assertEqual(data[ToolLogCollectionDataType.LLDP.name]["1"], "lldp data 1")


class TestHostLogParserBuilder(unittest.TestCase):
    def test_build_finds_matching_parsers(self):
        with tempfile.TemporaryDirectory() as tmp:
            root_dir = os.path.join(tmp, "dump")
            ascend_dmi = os.path.join(root_dir, "a", "ascend-dmi")
            os.makedirs(ascend_dmi)
            parsers = HostLogParserBuilder.build(root_dir)
            self.assertTrue(parsers)
            self.assertIsInstance(parsers[0], HostDumpLogParserV1)
            self.assertTrue(all(p.parse_dir for p in parsers))


if __name__ == "__main__":
    unittest.main()
