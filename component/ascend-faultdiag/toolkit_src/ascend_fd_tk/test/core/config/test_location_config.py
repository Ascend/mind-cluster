#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
# Licensed under the Apache License, Version 2.0

import unittest
from unittest.mock import MagicMock, patch

from ascend_fd_tk.core.config.location_config import (
    HostLocationInfo,
    SwitchLocationInfo,
    LocationConfig,
    LldConfigReader,
)


class TestLocationConfigFind(unittest.TestCase):
    def _config(self):
        c = LocationConfig()
        c.host_sn_location_map["SN001"] = HostLocationInfo(host_sn="SN001", room_name="机房A", cabinet_id="柜1")
        c.host_name_location_map["host2"] = HostLocationInfo(host_name="host2", room_name="机房B", cabinet_id="柜2")
        c.switch_sn_location_map["SW-SN"] = SwitchLocationInfo(sn="SW-SN", room_name="机房X", cabinet_id="柜X")
        c.switch_ip_location_map["10.0.0.1"] = SwitchLocationInfo(
            switch_ip="10.0.0.1", room_name="机房Y", cabinet_id="柜Y"
        )
        c.switch_name_location_map["sw1"] = SwitchLocationInfo(switch_name="sw1", room_name="机房Z", cabinet_id="柜Z")
        return c

    def test_find_host(self):
        c = self._config()
        self.assertEqual(c._find_host_location("SN001", ""), ("机房A", "柜1"))
        self.assertEqual(c._find_host_location("", "host2"), ("机房B", "柜2"))
        self.assertEqual(c._find_host_location("SN001", "host2"), ("机房A", "柜1"))  # SN优先
        self.assertEqual(c._find_host_location("x", "x"), ("", ""))

    def test_find_switch(self):
        c = self._config()
        self.assertEqual(c._find_switch_location("SW-SN", "", ""), ("机房X", "柜X"))
        self.assertEqual(c._find_switch_location("", "10.0.0.1", ""), ("机房Y", "柜Y"))
        self.assertEqual(c._find_switch_location("", "", "sw1"), ("机房Z", "柜Z"))
        self.assertEqual(c._find_switch_location("SW-SN", "10.0.0.1", "sw1"), ("机房X", "柜X"))  # SN优先

    def test_enrich_host(self):
        c = LocationConfig()
        c.host_sn_location_map["SN001"] = HostLocationInfo(host_sn="SN001", room_name="机房A", cabinet_id="柜1")
        host = MagicMock(sn_num="SN001", hostname="")
        c.enrich_host_info(host)
        self.assertEqual(host.room_name, "机房A")
        c.enrich_host_info(None)  # 不报错

    def test_enrich_switch(self):
        c = LocationConfig()
        c.switch_sn_location_map["SW-SN"] = SwitchLocationInfo(sn="SW-SN", room_name="机房X", cabinet_id="柜X")
        swi = MagicMock(sn="SW-SN", swi_id="", name="")
        c.enrich_switch_info(swi)
        self.assertEqual(swi.room_name, "机房X")
        c.enrich_switch_info(None)


class TestLldConfigReader(unittest.TestCase):
    def test_helpers(self):
        idx = LldConfigReader._build_col_index(["服务器", "主机SN"], ["服务器"])
        self.assertEqual(idx, {"服务器": 0})
        self.assertEqual(len(LldConfigReader._build_col_index(["x"], ["y"])), 0)
        wb = MagicMock(sheetnames=["灵衢L1网络对应关系"])
        self.assertIsNotNone(LldConfigReader._find_sheet_by_keyword(wb, "L1网络"))
        self.assertIsNone(LldConfigReader._find_sheet_by_keyword(wb, "不存在"))

    def test_read_l1_sheet(self):
        wb = MagicMock()
        ws = MagicMock()
        ws.iter_rows.return_value = iter(
            [
                ("服务器", "主机SN", "机房名称", "机柜编号", "L1名称", "L1_IP", "L1_SN"),
                ("host1", "SN001", "机房A", "柜1", "L1-1", "10.0.0.1", "L1SN001"),
            ]
        )
        wb.__getitem__ = MagicMock(return_value=ws)
        wb.sheetnames = ["灵衢L1网络对应关系"]
        config = LldConfigReader()._read(wb)
        self.assertIn("host1", config.host_name_location_map)
        self.assertIn("10.0.0.1", config.switch_ip_location_map)

    def test_read_l2_sheet(self):
        wb = MagicMock()
        ws = MagicMock()
        ws.iter_rows.return_value = iter(
            [
                ("设备名", "SN", "机房名称", "机柜编号", "管理IP配置"),
                ("L2-1", "L2SN001", "机房C", "柜3", "10.0.1.1"),
            ]
        )
        wb.__getitem__ = MagicMock(return_value=ws)
        wb.sheetnames = ["灵衢L2网络对应关系"]
        config = LldConfigReader()._read(wb)
        self.assertIn("L2-1", config.switch_name_location_map)
        self.assertIn("L2SN001", config.switch_sn_location_map)

    def test_file_not_exist(self):
        self.assertIsNone(LldConfigReader().read_from_dir("/nonexistent/path"))

    @patch("ascend_fd_tk.core.config.location_config.os.path.exists", return_value=True)
    @patch("ascend_fd_tk.core.config.location_config.load_workbook")
    def test_read_from_dir(self, mock_load, mock_exists):
        mock_load.side_effect = Exception("error")
        self.assertIsNone(LldConfigReader().read_from_dir("."))


if __name__ == "__main__":
    unittest.main()
