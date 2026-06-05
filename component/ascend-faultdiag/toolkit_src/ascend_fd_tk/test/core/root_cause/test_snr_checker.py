#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
# Licensed under the Apache License, Version 2.0

import unittest
from unittest.mock import MagicMock

from ascend_fd_tk.core.root_cause.constants import PORT_SPEED_200G, PORT_SPEED_400G
from ascend_fd_tk.core.root_cause.snr_checker import SnrChecker


class TestSnrChecker(unittest.TestCase):
    def test_get_port_speed(self):
        self.assertEqual(SnrChecker.get_port_speed("200G-1"), PORT_SPEED_200G)
        self.assertEqual(SnrChecker.get_port_speed("400G-1"), PORT_SPEED_400G)
        self.assertEqual(SnrChecker.get_port_speed("100G-1"), "")

    def test_check_hilink_snr_abnormal(self):
        mock_threshold = MagicMock()
        checker = SnrChecker(mock_threshold)
        # 无hccs_info
        swi = MagicMock(hccs_info=None)
        self.assertEqual(checker.check_hilink_snr_abnormal(swi, "200G-1"), (False, False))
        # 200G异常
        mock_threshold.CDR_HOST_SNR_LINE.check_value_str.return_value = True
        swi2 = MagicMock()
        swi2.hccs_info = MagicMock()
        swi2.hccs_info.interface_snr_list = [
            MagicMock(interface_name="200G-1", abnormal_lane_snr=[MagicMock(snr_value="12.5")])
        ]
        swi2.hccs_info.hccs_chip_port_snr_list = []
        self.assertEqual(checker.check_hilink_snr_abnormal(swi2, "200G-1"), (True, False))

    def test_check_optical_snr_abnormal(self):
        mock_threshold = MagicMock()
        checker = SnrChecker(mock_threshold)
        # None
        self.assertFalse(checker.check_optical_host_snr_abnormal(None))
        # host异常
        mock_threshold.HOST_SNR_DB.check_value_str.return_value = True
        optical = MagicMock(lane_power_infos=[MagicMock(host_snr="8.0")])
        self.assertTrue(checker.check_optical_host_snr_abnormal(optical))
        # media异常
        mock_threshold.MEDIA_SNR_DB.check_value_str.return_value = True
        optical2 = MagicMock(lane_power_infos=[MagicMock(media_snr="5.0")])
        self.assertTrue(checker.check_optical_media_snr_abnormal(optical2))

    def test_format_hilink_snr(self):
        self.assertEqual(SnrChecker.format_hilink_snr(MagicMock(hccs_info=None), "200G-1"), "")

    def test_format_lane_snr(self):
        self.assertEqual(SnrChecker.format_lane_snr(None, "host_snr"), "")
        optical = MagicMock(lane_power_infos=[MagicMock(lane_id=0, host_snr="12.5", media_snr="10.0")])
        self.assertEqual(SnrChecker.format_lane_snr(optical, "host_snr"), "Lane0:12.5")
        self.assertEqual(SnrChecker.format_lane_snr(optical, "media_snr"), "Lane0:10.0")


if __name__ == "__main__":
    unittest.main()
