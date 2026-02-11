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

tiancheng_chip_0_npu_port_base = list(range(12, 20))
tiancheng_chip_0_npu_port = tiancheng_chip_0_npu_port_base + [x + 24 for x in tiancheng_chip_0_npu_port_base]
tiancheng_chip_0_cpu_port = [20, 21, 22, 23]
tiancheng_chip_0_sw_port_base = list(range(0, 4)) + list(range(6, 10))
tiancheng_chip_0_sw_port = tiancheng_chip_0_sw_port_base + [x + 24 for x in tiancheng_chip_0_sw_port_base]
tiancheng_chip_0_xpu_port = tiancheng_chip_0_cpu_port + tiancheng_chip_0_npu_port
tiancheng_chip_0_all_port = tiancheng_chip_0_sw_port + tiancheng_chip_0_xpu_port

tiancheng_chip_1_npu_port_base = list(range(0, 8))
tiancheng_chip_1_npu_port = tiancheng_chip_1_npu_port_base + [x + 24 for x in tiancheng_chip_1_npu_port_base]
tiancheng_chip_1_cpu_port = []
tiancheng_chip_1_sw_port_base = list(range(12, 16)) + list(range(18, 22))
tiancheng_chip_1_sw_port = tiancheng_chip_1_sw_port_base + [x + 24 for x in tiancheng_chip_1_sw_port_base]
tiancheng_chip_1_xpu_port = tiancheng_chip_1_cpu_port + tiancheng_chip_1_npu_port
tiancheng_chip_1_all_port = tiancheng_chip_1_sw_port + tiancheng_chip_1_xpu_port

tiancheng_chip_2_npu_port = tiancheng_chip_0_npu_port
tiancheng_chip_2_cpu_port = tiancheng_chip_0_cpu_port
tiancheng_chip_2_sw_port = tiancheng_chip_0_sw_port
tiancheng_chip_2_xpu_port = tiancheng_chip_2_cpu_port + tiancheng_chip_2_npu_port
tiancheng_chip_2_all_port = tiancheng_chip_2_sw_port + tiancheng_chip_2_xpu_port

tiancheng_chip_3_npu_port = tiancheng_chip_1_npu_port
tiancheng_chip_3_cpu_port = [8, 9, 10, 11]
tiancheng_chip_3_sw_port = tiancheng_chip_1_sw_port
tiancheng_chip_3_xpu_port = tiancheng_chip_3_cpu_port + tiancheng_chip_3_npu_port
tiancheng_chip_3_all_port = tiancheng_chip_3_sw_port + tiancheng_chip_3_xpu_port

tiancheng_chip_4_npu_port = tiancheng_chip_0_npu_port
tiancheng_chip_4_cpu_port = tiancheng_chip_0_cpu_port
tiancheng_chip_4_sw_port = tiancheng_chip_0_sw_port
tiancheng_chip_4_xpu_port = tiancheng_chip_4_cpu_port + tiancheng_chip_4_npu_port
tiancheng_chip_4_all_port = tiancheng_chip_4_sw_port + tiancheng_chip_4_xpu_port

tiancheng_chip_5_npu_port = tiancheng_chip_1_npu_port
tiancheng_chip_5_cpu_port = tiancheng_chip_3_cpu_port
tiancheng_chip_5_sw_port = tiancheng_chip_1_sw_port
tiancheng_chip_5_xpu_port = tiancheng_chip_5_cpu_port + tiancheng_chip_5_npu_port
tiancheng_chip_5_all_port = tiancheng_chip_5_sw_port + tiancheng_chip_5_xpu_port

tiancheng_chip_6_npu_port = tiancheng_chip_0_npu_port
tiancheng_chip_6_cpu_port = tiancheng_chip_1_cpu_port
tiancheng_chip_6_sw_port = tiancheng_chip_0_sw_port
tiancheng_chip_6_xpu_port = tiancheng_chip_6_cpu_port + tiancheng_chip_6_npu_port
tiancheng_chip_6_all_port = tiancheng_chip_6_sw_port + tiancheng_chip_6_xpu_port

tiancheng_cpu_port_list = [tiancheng_chip_0_cpu_port, tiancheng_chip_1_cpu_port, tiancheng_chip_2_cpu_port,
                           tiancheng_chip_3_cpu_port, tiancheng_chip_4_cpu_port, tiancheng_chip_5_cpu_port,
                           tiancheng_chip_6_cpu_port]
tiancheng_npu_port_list = [tiancheng_chip_0_npu_port, tiancheng_chip_1_npu_port, tiancheng_chip_2_npu_port,
                           tiancheng_chip_3_npu_port, tiancheng_chip_4_npu_port, tiancheng_chip_5_npu_port,
                           tiancheng_chip_6_npu_port]
tiancheng_xpu_port_list = [tiancheng_chip_0_xpu_port, tiancheng_chip_1_xpu_port, tiancheng_chip_2_xpu_port,
                           tiancheng_chip_3_xpu_port, tiancheng_chip_4_xpu_port, tiancheng_chip_5_xpu_port,
                           tiancheng_chip_6_xpu_port]
