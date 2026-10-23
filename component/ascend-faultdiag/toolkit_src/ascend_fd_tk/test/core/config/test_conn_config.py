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

"""conn_config 单元测试：DeviceConfigParser 解析连接配置文件。"""

import unittest

from ascend_fd_tk.core.config.conn_config import DeviceConfigParser


class TestDeviceConfigParser(unittest.TestCase):
    def test_parse_single_ip(self):
        config = """
[host]
1.1.1.1 username=root password=123456
[bmc]
2.2.2.2 username=admin
[switch]
3.3.3.3 username=root port=22
"""
        conn_config = DeviceConfigParser(config).parse()
        self.assertEqual(len(conn_config.host_conn), 1)
        self.assertEqual(conn_config.host_conn[0].host, "1.1.1.1")
        self.assertEqual(conn_config.host_conn[0].username, "root")
        self.assertEqual(conn_config.host_conn[0].password, "123456")
        self.assertEqual(conn_config.host_conn[0].port, 22)
        self.assertEqual(len(conn_config.bmc_conn), 1)
        self.assertEqual(len(conn_config.switch_conn), 1)

    def test_parse_ip_range_with_step(self):
        config = """
[host]
1.1.1.1-1.1.1.3 username=root step=1
"""
        conn_config = DeviceConfigParser(config).parse()
        self.assertEqual([c.host for c in conn_config.host_conn], ["1.1.1.1", "1.1.1.2", "1.1.1.3"])

    def test_parse_invalid_ip(self):
        config = "[host]\n999.1.1.1 username=root\n"
        with self.assertRaises(ValueError):
            DeviceConfigParser(config).parse()

    def test_start_greater_than_end(self):
        config = "[host]\n1.1.1.3-1.1.1.1 username=root\n"
        with self.assertRaises(ValueError):
            DeviceConfigParser(config).parse()

    def test_missing_username(self):
        config = "[host]\n1.1.1.1 password=123\n"
        with self.assertRaises(ValueError):
            DeviceConfigParser(config).parse()

    def test_invalid_port(self):
        config = "[host]\n1.1.1.1 username=root port=abc\n"
        with self.assertRaises(ValueError):
            DeviceConfigParser(config).parse()

    def test_private_key_from_config_group(self):
        config = """
[config]
private_key="/path/to/key"
[host]
1.1.1.1 username=root
"""
        conn_config = DeviceConfigParser(config).parse()
        self.assertEqual(conn_config.host_conn[0].private_key, "/path/to/key")

    def test_comments_and_empty_lines_ignored(self):
        config = """
# 注释行
; 分号注释

[host]
1.1.1.1 username=root
"""
        conn_config = DeviceConfigParser(config).parse()
        self.assertEqual(len(conn_config.host_conn), 1)


if __name__ == "__main__":
    unittest.main()
