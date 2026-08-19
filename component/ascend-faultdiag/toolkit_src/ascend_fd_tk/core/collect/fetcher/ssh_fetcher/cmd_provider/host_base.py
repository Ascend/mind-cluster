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


class HostBaseProvider:
    """主机在线采集命令提供者基类（代际无关命令默认实现）。"""

    # 代际无关命令（基类提供默认实现，子类按需重写）
    def npu_mapping_cmd(self) -> str:
        """npu-smi info -m 映射查询命令"""
        return "npu-smi info -m"

    def npu_type_cmd(self) -> str:
        """lspci NPU 型号查询命令"""
        return "lspci |grep 'Device d80' --color=never"

    def sn_num_cmd(self) -> str:
        """dmidecode 序列号查询命令"""
        return "dmidecode -s system-serial-number"

    def hostname_cmd(self) -> str:
        """主机名查询命令"""
        return "hostname"

    def msnpureport_cmd(self) -> str:
        """msnpureport 日志采集命令"""
        return "msnpureport"

    def hccs_cmd(self, npu_id, chip_id) -> str:
        """npu-smi hccs 采集命令"""
        return "npu-smi info -t hccs -i {} -c {}".format(npu_id, chip_id)

    def spod_cmd(self, npu_id, chip_id) -> str:
        """npu-smi spod-info 采集命令"""
        return "npu-smi info -t spod-info -i {} -c {}".format(npu_id, chip_id)
