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

"""交换机 CLI 输出解析器单元测试：TXT 解析、diag info 解析与入口分发。"""

import os
import tempfile
import unittest

from ascend_fd_tk.core.collect.collect_config import SwiCliOutputDataType
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.cli_output_txt.cli_output_txt_parser import (
    SwiCliOutputTxtParser,
)
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.diag_info_output.diag_info_output_parser import (
    SwitchDiagnoseInformationParser,
)
from ascend_fd_tk.core.collect.fetcher.dump_log_fetcher.switch.swi_cli_output_parser import SwiCliOutputParser

_TXT_CONTENT = (
    "<RCK A-2>\n"
    "display clock\n"
    "clock 2026-08-29 10:00:00 Time Zone CST\n"
    "interface MEth0/0/0\n"
    " ip address 10.1.1.1 255.255.255.0\n"
)


class TestSwiCliOutputTxtParser(unittest.TestCase):
    def test_parse_with_ip(self):
        data = SwiCliOutputTxtParser(_TXT_CONTENT).parse()
        self.assertEqual(data[SwiCliOutputDataType.SWI_NAME.name], "RCK A-2")
        self.assertEqual(data[SwiCliOutputDataType.SWI_IP.name], "10.1.1.1")
        self.assertIn(SwiCliOutputDataType.CLOCK.name, data)

    def test_parse_without_ip(self):
        content = "<RCK A-2>\ndisplay clock\nclock 2026-08-29 10:00:00 Time Zone CST\n"
        self.assertEqual(SwiCliOutputTxtParser(content).parse(), {})

    def test_find_name_without_match(self):
        parser = SwiCliOutputTxtParser("no angle bracket")
        parser.find_name("no angle bracket")
        self.assertNotIn(SwiCliOutputDataType.SWI_NAME.name, parser.parse_data.get_data_dict())


class TestSwitchDiagnoseInformationParser(unittest.TestCase):
    def _make_content(self):
        sep = "=" * 79
        return (
            sep + "\n"
            "display current-configuration\n"
            "sysname SW0\n"
            "interface MEth0/0/0\n"
            " ip address 10.1.1.1 255.255.255.0\n"
            "display clock\n"
            "clock 2026-08-29 10:00:00 Time Zone CST\n" + sep + "\n"
            "other section\n"
        )

    def test_parse(self):
        data = SwitchDiagnoseInformationParser(self._make_content()).parse()
        self.assertEqual(data[SwiCliOutputDataType.SWI_NAME.name], "SW0")
        self.assertEqual(data[SwiCliOutputDataType.SWI_IP.name], "10.1.1.1")
        self.assertIn(SwiCliOutputDataType.CLOCK.name, data)

    def test_parse_odd_parts(self):
        content = "display current-configuration\nsysname SW0\n" + "=" * 79 + "\n"
        self.assertEqual(SwitchDiagnoseInformationParser(content).parse(), {})


class TestSwiCliOutputParser(unittest.TestCase):
    def test_parse_missing_file(self):
        self.assertEqual(SwiCliOutputParser.parse(os.path.join(tempfile.gettempdir(), "no_such_file")), {})

    def test_parse_dispatch_diag_info(self):
        with tempfile.TemporaryDirectory() as tmp:
            file_path = os.path.join(tmp, "diag.txt")
            sep = "=" * 79
            with open(file_path, "w", encoding="utf-8") as f:
                f.write(
                    sep + "\n"
                    "display current-configuration\n"
                    "sysname SW0\n"
                    "interface MEth0/0/0\n"
                    " ip address 10.1.1.1 255.255.255.0\n" + sep + "\n"
                )
            data = SwiCliOutputParser.parse(file_path)
            self.assertEqual(data[SwiCliOutputDataType.SWI_NAME.name], "SW0")
            self.assertEqual(data[SwiCliOutputDataType.SWI_IP.name], "10.1.1.1")

    def test_parse_dispatch_txt(self):
        with tempfile.TemporaryDirectory() as tmp:
            file_path = os.path.join(tmp, "cli.txt")
            with open(file_path, "w", encoding="utf-8") as f:
                f.write(_TXT_CONTENT)
            data = SwiCliOutputParser.parse(file_path)
            self.assertEqual(data[SwiCliOutputDataType.SWI_NAME.name], "RCK A-2")


if __name__ == "__main__":
    unittest.main()
