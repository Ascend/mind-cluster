#!/usr/bin/env python3
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
import unittest

from ascend_fd import parse_fault_type

MINDIE_LINE = (
    "[2024-11-05 12:00:00.123456] [11] [1] [MS] [1] [ERROR] : [MIE03E400008] [HttpServer] Http server on timeout."
)


def build_server(server="1.1.1.1", device=None, items=None):
    item = {"item_type": "MindIE", "log_lines": [MINDIE_LINE]}
    return {
        "log_domain": {"server": server, "device": ["0"] if device is None else device},
        "log_items": [item] if items is None else items,
    }


def build_invalid_log_lines():
    return build_server(items=[{"item_type": "MindIE"}])


class STParseFaultType(unittest.TestCase):
    def assert_parse_failed(self, payload, expected):
        results, err_msg_list = parse_fault_type(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn(expected, err_msg_list[0])

    def test_security_0021_255_servers_failed_one_success(self):
        """Ascend_Fd_SDK_Parse_Security_0021"""
        payload = [build_invalid_log_lines() for _ in range(255)] + [build_server()]
        results, err_msg_list = parse_fault_type(payload)
        self.assertEqual(len(results), 1)
        self.assertTrue(err_msg_list)
        self.assertIn("Some parsing works failed", err_msg_list[0])

    def test_security_0020_256_servers_all_failed(self):
        """Ascend_Fd_SDK_Parse_Security_0020"""
        results, err_msg_list = parse_fault_type([build_invalid_log_lines() for _ in range(256)])
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn("Some parsing works failed", err_msg_list[0])

    def test_security_0019_log_items_more_than_100(self):
        """Ascend_Fd_SDK_Parse_Security_0019"""
        items = [{"item_type": "MindIE", "log_lines": [MINDIE_LINE]} for _ in range(101)]
        self.assert_parse_failed([build_server(items=items)], "exceeds the limit, which is 100")

    def test_security_0018_servers_more_than_256(self):
        """Ascend_Fd_SDK_Parse_Security_0018"""
        self.assert_parse_failed([build_server() for _ in range(257)], "exceeds the limit, which is 256")

    def test_security_0001_server_value_empty(self):
        """Ascend_Fd_SDK_Parse_Security_0001"""
        self.assert_parse_failed([build_server(server="")], "'server' should be available and valid")

    def test_security_0002_item_type_value_invalid(self):
        """Ascend_Fd_SDK_Parse_Security_0002"""
        payload = [build_server(items=[{"item_type": "mindie", "log_lines": [MINDIE_LINE]}])]
        # An unknown but well-formed item type yields no fault instead of a hard failure.
        results, _ = parse_fault_type(payload)
        self.assertEqual(results, [])

    def test_security_0003_item_type_value_empty(self):
        """Ascend_Fd_SDK_Parse_Security_0003"""
        payload = [build_server(items=[{"item_type": "", "log_lines": [MINDIE_LINE]}])]
        results, err_msg_list = parse_fault_type(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn("both 'item_type' and 'log_lines'", err_msg_list[0])

    def test_security_0006_device_value_empty_list(self):
        """Ascend_Fd_SDK_Parse_Security_0006"""
        results, err_msg_list = parse_fault_type([build_server(device=[])])
        self.assertEqual(err_msg_list, [])
        self.assertTrue(results)

    def test_security_0005_device_element_not_number(self):
        """Ascend_Fd_SDK_Parse_Security_0005"""
        results, err_msg_list = parse_fault_type([build_server(device=["a"])])
        self.assertEqual(err_msg_list, [])
        self.assertTrue(results)

    def test_security_0004_server_value_not_ip(self):
        """Ascend_Fd_SDK_Parse_Security_0004"""
        results, err_msg_list = parse_fault_type([build_server(server="1.@.中.a")])
        self.assertEqual(err_msg_list, [])
        self.assertTrue(results)

    def test_security_0017_multiple_anomalies(self):
        """Ascend_Fd_SDK_Parse_Security_0017"""
        payload = [build_server(server=""), build_server(device=["a"]), build_invalid_log_lines()]
        results, err_msg_list = parse_fault_type(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn("Parse failed", err_msg_list[0])

    def test_security_0008_device_type_not_list(self):
        """Ascend_Fd_SDK_Parse_Security_0008"""
        self.assert_parse_failed([build_server(device="0")], "Invalid parameter type for 'device'")

    def test_security_0007_server_type_not_str(self):
        """Ascend_Fd_SDK_Parse_Security_0007"""
        self.assert_parse_failed([build_server(server=1)], "Invalid parameter type for 'server'")

    def test_security_0009_item_type_type_not_str(self):
        """Ascend_Fd_SDK_Parse_Security_0009"""
        payload = [build_server(items=[{"item_type": 1, "log_lines": [MINDIE_LINE]}])]
        results, err_msg_list = parse_fault_type(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn("Invalid parameter type for 'item_type'", err_msg_list[0])

    def test_security_0010_log_lines_type_not_list(self):
        """Ascend_Fd_SDK_Parse_Security_0010"""
        payload = [build_server(items=[{"item_type": "MindIE", "log_lines": MINDIE_LINE}])]
        results, err_msg_list = parse_fault_type(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn("Invalid parameter type for 'log_lines'", err_msg_list[0])

    def test_security_0013_device_missing(self):
        """Ascend_Fd_SDK_Parse_Security_0013"""
        payload = build_server()
        payload["log_domain"].pop("device")
        self.assert_parse_failed([payload], "'device' should be a filed of 'log_domain'")

    def test_security_0012_server_missing(self):
        """Ascend_Fd_SDK_Parse_Security_0012"""
        payload = build_server()
        payload["log_domain"].pop("server")
        self.assert_parse_failed([payload], "'server' should be available and valid")

    def test_security_0011_log_domain_missing(self):
        """Ascend_Fd_SDK_Parse_Security_0011"""
        payload = build_server()
        payload.pop("log_domain")
        self.assert_parse_failed([payload], "both 'log_domain' and 'log_items'")

    def test_security_0015_item_type_missing(self):
        """Ascend_Fd_SDK_Parse_Security_0015"""
        payload = [build_server(items=[{"log_lines": [MINDIE_LINE]}])]
        results, err_msg_list = parse_fault_type(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn("both 'item_type' and 'log_lines'", err_msg_list[0])

    def test_security_0016_log_lines_missing(self):
        """Ascend_Fd_SDK_Parse_Security_0016"""
        payload = [build_server(items=[{"item_type": "MindIE"}])]
        results, err_msg_list = parse_fault_type(payload)
        self.assertEqual(results, [])
        self.assertTrue(err_msg_list)
        self.assertIn("'log_lines' need to exist", err_msg_list[0])

    def test_security_0014_log_items_missing(self):
        """Ascend_Fd_SDK_Parse_Security_0014"""
        payload = build_server()
        payload.pop("log_items")
        self.assert_parse_failed([payload], "both 'log_domain' and 'log_items'")


if __name__ == "__main__":
    unittest.main()
