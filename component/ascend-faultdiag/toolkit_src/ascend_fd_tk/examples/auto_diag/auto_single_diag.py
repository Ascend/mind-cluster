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

import asyncio

from ascend_fd_tk.core.context.diag_ctx import DiagCtx
from ascend_fd_tk.core.model.diag_result import BmcDomain, DiagResult, HostDomain, SwitchDomain
from ascend_fd_tk.core.report.sheet.optical_module_sheet import HostToSwitchOpticalModuleData
from ascend_fd_tk.core.report.sheet.switch_optical_module_sheet import SwitchOpticalModuleData
from ascend_fd_tk.core.root_cause.model import HostToL1LinkData, L1ToL2LinkData
from ascend_fd_tk.core.service.auto_diag import AutoDiag
from ascend_fd_tk.core.service.generate_diag_report import GenerateDiagReport
from ascend_fd_tk.core.service.load_cache import LoadCache
from ascend_fd_tk.core.service.root_cause_analysis import RootCauseAnalysis
from ascend_fd_tk.utils import logger

_DIAG_LOGGER = logger.DIAG_LOGGER


class AutoSingleDiag:
    """指定单条链路进行诊断（设备IP + 端口）。

    执行全量诊断与根因分析，在结果展示阶段仅保留指定链路（故障域本端或对端命中IP+端口）
    的诊断结果，再生成诊断报告。
    """

    def __init__(self, diag_ctx=DiagCtx(), ip="", port=""):
        self.diag_ctx = diag_ctx
        self.ip = ip
        self.port = port
        self.error_msg = ""

    async def main(self):
        await LoadCache(self.diag_ctx).run()
        if not self._check_device():
            return
        _DIAG_LOGGER.info("单链路诊断，设备IP：%s，端口：%s", self.ip, self.port)
        await AutoDiag(self.diag_ctx).run()
        await RootCauseAnalysis(self.diag_ctx).run()
        # 结果展示阶段过滤：仅保留指定链路的诊断结果
        self.diag_ctx.diag_result = [result for result in self.diag_ctx.diag_result if self._is_specified_link(result)]
        if not self.diag_ctx.diag_result:
            self.error_msg = f"链路 {self.ip} 端口 {self.port} 诊断完成，未发现故障"
            return
        await GenerateDiagReport(self.diag_ctx, link_filter=self._is_specified_link_row).run()

    def _check_device(self) -> bool:
        """校验设备IP及端口是否存在于缓存中"""
        cache = self.diag_ctx.cache
        if self.ip in cache.hosts_info:
            host_info = cache.hosts_info[self.ip]
            ports = self._host_valid_ports(host_info)
            if self.port not in ports:
                self.error_msg = f"主机 {self.ip} 不存在NPU端口 {self.port}，可用端口：{', '.join(sorted(ports))}"
                return False
        elif self.ip in cache.swis_info:
            swi_info = cache.swis_info[self.ip]
            if self.port not in swi_info.interface_full_infos:
                samples = ", ".join(list(swi_info.interface_full_infos)[:5])
                self.error_msg = f"交换机 {self.ip} 不存在端口 {self.port}，可用端口示例：{samples}"
                return False
        elif self.ip not in cache.bmcs_info:
            self.error_msg = f"未在缓存中找到设备 {self.ip}，请确认为已采集的host/switch/bmc设备IP"
            return False
        return True

    @staticmethod
    def _host_valid_ports(host_info) -> set:
        """主机可用端口：各芯片的npu_id（端口输入为npu_id，不做chip_phy_id兜底）"""
        return {chip_info.npu_id for chip_info in host_info.npu_chip_info.values() if getattr(chip_info, "npu_id", "")}

    def _is_specified_link(self, result: DiagResult) -> bool:
        """诊断结果是否属于指定链路（故障域本端或对端命中IP+端口）"""
        domain = result.domain
        if isinstance(domain, HostDomain):
            if domain.host_id == self.ip:
                return self._hit_host_port(domain.npu_id)
            return self._hit(domain.peer_switch_id, domain.peer_interface)
        if isinstance(domain, SwitchDomain):
            if domain.swi_id == self.ip:
                return domain.interface == self.port
            return self._hit(domain.peer_switch_id, domain.peer_switch_interface)
        if isinstance(domain, BmcDomain):
            return domain.bmc_id == self.ip and self._hit_host_port(domain.npu_id)
        return False

    def _is_specified_link_row(self, row) -> bool:
        """报告数据行（链路分析/光模块sheet行）是否属于指定链路，按本端或对端IP+端口匹配"""
        if isinstance(row, HostToL1LinkData):
            if row.host_id == self.ip:
                return self._hit_host_port(row.npu_id)
            return self._hit(row.l1_switch_id, row.l1_interface)
        if isinstance(row, L1ToL2LinkData):
            if row.l1_switch_id == self.ip:
                return row.l1_interface == self.port
            return self._hit(row.l2_switch_id, row.l2_interface)
        if isinstance(row, HostToSwitchOpticalModuleData):
            if row.host_id == self.ip:
                return self._hit_host_port(row.npu_id)
            return self._hit(row.peer_switch_id, row.peer_switch_port)
        if isinstance(row, SwitchOpticalModuleData):
            if row.local_switch_id == self.ip:
                return row.local_interface == self.port
            return self._hit(row.peer_switch_id, row.peer_interface)
        return False

    def _hit_host_port(self, npu_id: str) -> bool:
        """主机/BMC端口匹配：端口为npu_id，仅匹配npu_id，不做chip_phy_id兜底"""
        return npu_id == self.port

    def _hit(self, device_id: str, port: str) -> bool:
        return device_id == self.ip and port == self.port


if __name__ == '__main__':
    asyncio.run(AutoSingleDiag(ip="127.0.0.1", port="0").main())
