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

from diag_tool.core.log_parser.base import LogParsePattern

SWI_DIAG_INFO_LOG_CONFIG = [
    LogParsePattern.build(
        "interface_info",
        r"Optical Interface=\[",
        r"slot_1/tempdir/devm_picm.log",
        r"\[(?P<time>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}:\d{1,3})\].{1,100}Optical Interface=\[\d/(?P<chip_id>\d)/"
        r"(?P<port_id>\d{1,3})\] (?P<items>[a-z]{1,20})\[(?P<lane_id>[0-7])\]=(?P<value>[0-9.]{1,10})\."),
    LogParsePattern.build(
        "snr",
        r"type/card/port/bitmap=",
        r"slot_1/tempdir/devm_picm.log",
        r"\[(?P<time>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}:\d{1,3})\].{1,100}type/card/port/bitmap=\d/"
        r"(?P<chip_id>\d)/(?P<port_id>\d{1,3})/\d{1,3}, (?P<items>SNR)=(?P<value>[0-9.]{1,10}), "
        r"index=(?P<lane_id>[0-7]), mode=(?P<mode>[01])\.")
]
