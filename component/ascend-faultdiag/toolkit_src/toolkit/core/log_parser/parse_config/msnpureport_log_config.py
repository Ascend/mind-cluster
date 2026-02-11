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

from toolkit.core.log_parser.base import LogParsePattern

MS_NPU_REPORT_PARSE_CONFIG = [
    LogParsePattern.build("slog_device_error_status", r"error status.*fifo state",
                          r"slog/dev-os-\d+/debug/device-\d+/device-\d+.*\.log",
                          r"device_id:(?P<device_id>\d+) die_id:(?P<die_id>\d+).*error status\[(?P<error_status>.*?)\]"),

    # 添加 uncorr_cw_cnt 的配置
    LogParsePattern.build("uncorr_cw_cnt_check", r"uncorr_cw_cnt\s+(\d+)",
                          r"slog/dev-os-\d+/.*\.log",
                          r"\[EVENT\].*?(?P<timestamp>\d{4}-\d{2}-\d{2}-\d{2}:\d{2}:\d{2})\..*?device_id:(?P<device_id>\d+).*die_id:(?P<die_id>\d+).*?uncorr_cw_cnt\s+(?P<count>\d+)"),

    # 添加 IIC 故障的配置
    LogParsePattern.build("iic_error_check", r"trans status\[0x40\].*error status\[0x10\]",
                          r"slog/dev-os-\d+/.*\.log",
                          r"device_id:(?P<device_id>\d+).*die_id:(?P<die_id>\d+).*trans status\[0x40\].*error status\[0x10\]"),

    # 添加 rf_lf 和 pcs_link 的配置
    LogParsePattern.build("rf_lf_check", r"rf_lf\s+(\d+)",
                          r"slog/dev-os-\d+/.*\.log",
                          r"rf_lf\s+(?P<rf_lf_value>\d+)"),

    LogParsePattern.build("pcs_link_check", r"pcs_link\s+(\d+)",
                          r"slog/dev-os-\d+/.*\.log",
                          r"pcs_link\s+(?P<pcs_link_value>\d+)"),
]
