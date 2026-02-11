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

import unittest
from unittest.mock import patch, MagicMock

from toolkit.core.collect.parser.switch_parser import SwitchParser

TIME_STR = "time"


class TestSwitchParser(unittest.TestCase):

    def test_parse_op_state_flag_diag_info(self):
        # 测试parse_op_state_flag_diag_info方法
        cmd_res = """Items   Status
--------------
TX LOS   Normal
RX LOS   Alarm
------------"""
        result = SwitchParser.parse_op_state_flag_diag_info(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].items, "TX LOS")
        self.assertEqual(result[0].status, "Normal")
        self.assertEqual(result[1].items, "RX LOS")
        self.assertEqual(result[1].status, "Alarm")

    def test_trans_opt_module_results(self):
        # 测试trans_opt_module_results方法
        cmd_res = """Items    Value  HighAlarm  HighWarn  LowAlarm  LowWarn  Status
------------------------------------------------------------------------
TxPower  -10.0  -5.0       -7.0      -15.0     -12.0     Normal
RxPower  -11.5  -6.0       -8.0      -16.0     -13.0     Normal
------------"""
        result = SwitchParser.trans_opt_module_results(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].items, "TxPower")
        self.assertEqual(result[0].value, "-10.0")
        self.assertEqual(result[1].items, "RxPower")
        self.assertEqual(result[1].status, "Normal")

    def test_filter_opt_module_info(self):
        # 测试filter_opt_module_info方法
        mock_log_info1 = MagicMock()
        mock_log_info1.info_dict = {
            "chip_id": "1",
            "port_id": "1",
            "items": "TxPower",
            "lane_id": "0",
            "mode": "",
            TIME_STR: "2026-02-01 10:00:00"
        }
        mock_log_info2 = MagicMock()
        mock_log_info2.info_dict = {
            "chip_id": "1",
            "port_id": "1",
            "items": "TxPower",
            "lane_id": "0",
            "mode": "",
            TIME_STR: "2026-02-01 11:00:00"
        }
        result = SwitchParser.filter_opt_module_info([mock_log_info1, mock_log_info2])
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0][TIME_STR], "2026-02-01 11:00:00")

    @patch('toolkit.core.collect.parser.switch_parser.FormParser')
    def test_parse_bit_err_rate(self, mock_form_parser):
        # 测试parse_bit_err_rate方法
        mock_form = MagicMock()
        test_num = "1e-5"
        mock_form.parse.return_value = {"Bit error rate": test_num}
        mock_form_parser.return_value = mock_form
        cmd_res = "display interface troubleshooting eth0\ndisplay interface troubleshooting eth1"
        mock_interface_briefs = [
            MagicMock(interface="eth0"),
            MagicMock(interface="eth1")
        ]
        result = SwitchParser.parse_bit_err_rate(cmd_res, mock_interface_briefs)
        self.assertEqual(len(result), 2)  # 两个interface的bit error rate都超过了阈值
        self.assertEqual(result[0].interface_name, "eth0")
        self.assertEqual(result[0].bit_err_rate, test_num)
        self.assertEqual(result[1].interface_name, "eth1")
        self.assertEqual(result[1].bit_err_rate, test_num)

    def test_parse_lldp_nei_brief(self):
        # 测试parse_lldp_nei_brief方法
        cmd_res = """Local Interface  Exptime(s)  Neighbor Interface  Neighbor Device
----------------------------------------------------------------------------
eth0              120         eth1                host-01
eth1              120         eth0                host-02"""
        result = SwitchParser.parse_lldp_nei_brief(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].local_interface_name, "eth0")
        self.assertEqual(result[0].remote_device_interface.interface, "eth1")
        self.assertEqual(result[0].remote_device_interface.device_name, "host-01")

    def test_parse_alarms(self):
        # 测试parse_alarms方法
        cmd_res = """Sequence  AlarmId  Severity  Date Time         Description
------------------------------------------------------------------------
1         ALM-001  Critical  2026-02-01 10:00:00  Interface eth0 down
2         ALM-002  Warning   2026-02-01 11:00:00  High temperature
------"""
        result = SwitchParser.parse_alarms(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].alarm_id, "ALM-001")
        self.assertEqual(result[0].severity, "Critical")
        self.assertEqual(result[1].alarm_id, "ALM-002")

    @patch('toolkit.utils.form_parser.FormParser')
    def test_parse_interface_info(self, mock_form_parser):
        # 测试parse_interface_info方法
        mock_form = MagicMock()
        mock_form.parse.return_value = {"Speed": "200G"}
        mock_form_parser.return_value = mock_form
        cmd_res = "eth0 current state : UP (ifindex 1)\nSpeed: 200G\neth1 current state : DOWN (ifindex 2)\nSpeed: 200G"
        result = SwitchParser.parse_interface_info(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].interface_name, "eth0")
        self.assertEqual(result[1].interface_name, "eth1")

    def test_parse_datetime(self):
        # 测试parse_datetime方法
        cmd_res = """System time:
2026-02-01 10:00:00
Timezone: UTC+8"""
        result = SwitchParser.parse_datetime(cmd_res)
        self.assertEqual(result, "2026-02-01 10:00:00")

    @patch('toolkit.utils.form_parser.FormParser')
    def test_parse_esn(self, mock_form_parser):
        # 测试parse_esn方法
        mock_form = MagicMock()
        mock_form.parse.return_value = {"ESN": "ABC123"}
        mock_form_parser.return_value = mock_form
        cmd_res = "ESN: ABC123"
        result = SwitchParser.parse_esn(cmd_res)
        self.assertEqual(result, "ABC123")

    @patch('toolkit.core.collect.parser.switch_parser.FormParser')
    def test_parse_transceiver_info(self, mock_form_parser):
        # 测试parse_transceiver_info方法
        mock_form = MagicMock()
        mock_form.parse.return_value = {
            "transceiver information eth0": {"Vendor": "Huawei", "Part Number": "02310MNY"}
        }
        mock_form_parser.return_value = mock_form
        cmd_res = "transceiver information eth0\nVendor: Huawei\nPart Number: 02310MNY"
        result = SwitchParser.parse_transceiver_info(cmd_res)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].interface, "eth0")
        self.assertEqual(result[0].manufacture_information.manu_serial_number, "")  # 默认值
        # 注意：TransceiverInfo类没有直接的vendor属性，vendor信息应该在manufacture_information中

    @patch('toolkit.utils.form_parser.FormParser')
    def test_parse_alarm_verbose(self, mock_form_parser):
        # 测试parse_alarm_verbose方法
        mock_form = MagicMock()
        mock_form.parse.return_value = {"AlarmId": "ALM-001", "Severity": "Critical"}
        mock_form_parser.return_value = mock_form
        cmd_res = "AlarmId: ALM-001\nSeverity: Critical\n\nAlarmId: ALM-002\nSeverity: Warning"
        result = SwitchParser.parse_alarm_verbose(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].alarm_id, "ALM-001")
        self.assertEqual(result[1].alarm_id, "ALM-002")

    @patch('toolkit.core.collect.parser.switch_parser.port_mapping_config')
    def test_parse_port_mapping(self, mock_port_config):
        # 测试parse_port_mapping方法 - 从命令结果解析
        cmd_res = """Interface  IfIndex  TB  TP  Chip  Port  Core
--------------------------------------------------------
eth0        1        0   0    0     0     0
eth1        2        0   0    0     1     0
------"""
        result = SwitchParser.parse_port_mapping(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result["0/0"], "eth0")
        self.assertEqual(result["0/1"], "eth1")

    def test_parse_interface_brief(self):
        # 测试parse_interface_brief方法
        cmd_res = """Interface    PHY    Protocol    InUti    OutUti    inErrors    outErrors
eth0(200GE)  up     up          0.1%     0.2%      0           0
eth1         down   down        0.0%     0.0%      10          5"""
        result = SwitchParser.parse_interface_brief(cmd_res)
        self.assertEqual(len(result), 2)
        self.assertEqual(result[0].interface, "eth0")
        self.assertEqual(result[0].phy, "up")
        self.assertEqual(result[1].interface, "eth1")
        self.assertEqual(result[1].in_errors, "10")


if __name__ == '__main__':
    unittest.main()
