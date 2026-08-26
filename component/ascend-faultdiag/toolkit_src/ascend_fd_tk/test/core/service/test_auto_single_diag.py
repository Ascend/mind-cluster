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

"""AutoSingleDiag 单元测试：全量诊断后按 IP+端口过滤诊断结果"""

import asyncio
import dataclasses
import unittest
from unittest.mock import AsyncMock, MagicMock, patch

from ascend_fd_tk.core.config.threshold_config import A5Threshold
from ascend_fd_tk.core.cli_module.cli_model import AutoSingleDiagCliModel
from ascend_fd_tk.core.model.diag_result import BmcDomain, DiagResult, HostDomain, SwitchDomain
from ascend_fd_tk.core.report.sheet.optical_module_sheet import (
    HostToSwitchOpticalModuleData,
    HostToSwitchOpticalModuleSheetGenerator,
)
from ascend_fd_tk.core.report.sheet.signal_link_mapping_sheet import SignalLinkMappingSheetGenerator
from ascend_fd_tk.core.report.sheet.switch_optical_module_sheet import (
    SwitchOpticalModuleData,
    SwitchOpticalModuleSheetGenerator,
)
from ascend_fd_tk.core.root_cause.model import HostToL1LinkData, L1ToL2LinkData
from ascend_fd_tk.examples.auto_diag.auto_single_diag import AutoSingleDiag

_HOST_IP = "10.1.1.1"
_SWI_IP = "10.1.1.2"
_BMC_IP = "10.1.1.3"
_SWI_IFACE = "400GE1/0/1:10"


def _make_cache():
    cache = MagicMock()
    chip0 = MagicMock(npu_id="0", chip_phy_id="0")
    chip5 = MagicMock(npu_id="5", chip_phy_id="9")
    host = MagicMock()
    host.npu_chip_info = {"0": chip0, "9": chip5}
    swi = MagicMock()
    swi.interface_full_infos = {_SWI_IFACE: MagicMock()}
    cache.hosts_info = {_HOST_IP: host}
    cache.swis_info = {_SWI_IP: swi}
    cache.bmcs_info = {_BMC_IP: MagicMock()}
    return cache


def _make_diag_ctx(cache=None, diag_results=None):
    diag_ctx = MagicMock()
    diag_ctx.cache = cache or _make_cache()
    diag_ctx.diag_result = diag_results or []
    return diag_ctx


def _fill_required(cls, **kwargs):
    """构造dataclass实例，无默认值的字段以空串占位"""
    values = {
        field.name: ""
        for field in dataclasses.fields(cls)
        if field.default is dataclasses.MISSING and field.default_factory is dataclasses.MISSING
    }
    values.update(kwargs)
    return cls(**values)


class TestCheckDevice(unittest.TestCase):
    """设备IP与端口校验"""

    def test_unknown_ip(self):
        """IP不在缓存中时报错并提示"""
        diag = AutoSingleDiag(_make_diag_ctx(), "10.9.9.9", "0")
        self.assertFalse(diag._check_device())
        self.assertIn("未在缓存中找到设备 10.9.9.9", diag.error_msg)

    def test_host_port_not_exist(self):
        """主机端口不存在时报错并列出可用端口（仅npu_id）"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "15")
        self.assertFalse(diag._check_device())
        self.assertIn("不存在NPU端口 15", diag.error_msg)
        self.assertIn("0, 5", diag.error_msg)

    def test_host_port_by_npu_id(self):
        """端口匹配npu_id（非npu_chip_info键）时校验通过"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        self.assertTrue(diag._check_device())

    def test_host_port_chip_phy_id_not_valid(self):
        """端口仅命中chip_phy_id或npu_chip_info键时校验失败（不做chip_phy_id兜底）"""
        # chip5的chip_phy_id为9，且9为npu_chip_info键
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "9")
        self.assertFalse(diag._check_device())
        self.assertIn("不存在NPU端口 9", diag.error_msg)

    def test_switch_port_not_exist(self):
        """交换机端口不存在时报错并列出端口示例"""
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, "400GE1/0/1:99")
        self.assertFalse(diag._check_device())
        self.assertIn("不存在端口 400GE1/0/1:99", diag.error_msg)
        self.assertIn(_SWI_IFACE, diag.error_msg)

    def test_switch_port_exist(self):
        """交换机端口存在时校验通过"""
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, _SWI_IFACE)
        self.assertTrue(diag._check_device())
        self.assertEqual(diag.error_msg, "")

    def test_bmc_skip_port_check(self):
        """BMC端口不做存在性校验（端口为关联主机NPU端口号）"""
        diag = AutoSingleDiag(_make_diag_ctx(), _BMC_IP, "5")
        self.assertTrue(diag._check_device())


class TestIsSpecifiedLink(unittest.TestCase):
    """诊断结果链路匹配"""

    def test_host_match_npu_id(self):
        """host入口：本端host_id+npu_id命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        domain = HostDomain(host_id=_HOST_IP, npu_id="5", chip_phy_id="9")
        self.assertTrue(diag._is_specified_link(DiagResult(domain)))

    def test_host_chip_phy_id_no_fallback(self):
        """npu_id缺失时不回退匹配chip_phy_id（端口仅支持npu_id）"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        domain = HostDomain(host_id=_HOST_IP, npu_id="", chip_phy_id="5")
        self.assertFalse(diag._is_specified_link(DiagResult(domain)))

    def test_host_peer_switch_match(self):
        """switch入口：host故障域携带对端交换机信息时命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, _SWI_IFACE)
        domain = HostDomain(
            host_id="10.1.1.9",
            chip_phy_id="3",
            peer_switch_id=_SWI_IP,
            peer_interface=_SWI_IFACE,
        )
        self.assertTrue(diag._is_specified_link(DiagResult(domain)))

    def test_switch_match_local(self):
        """switch入口：本端swi_id+interface命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, _SWI_IFACE)
        domain = SwitchDomain(swi_id=_SWI_IP, interface=_SWI_IFACE)
        self.assertTrue(diag._is_specified_link(DiagResult(domain)))

    def test_switch_match_peer(self):
        """switch入口：对端交换机端口命中（L1-L2链路另一端）"""
        diag = AutoSingleDiag(_make_diag_ctx(), "10.1.1.8", "200GE1/0/2:3")
        domain = SwitchDomain(
            swi_id=_SWI_IP,
            interface=_SWI_IFACE,
            peer_switch_id="10.1.1.8",
            peer_switch_interface="200GE1/0/2:3",
        )
        self.assertTrue(diag._is_specified_link(DiagResult(domain)))

    def test_bmc_match(self):
        """bmc入口：bmc_id+npu_id命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _BMC_IP, "5")
        domain = BmcDomain(bmc_id=_BMC_IP, npu_id="5", chip_phy_id="9")
        self.assertTrue(diag._is_specified_link(DiagResult(domain)))

    def test_no_match(self):
        """IP或端口不匹配时过滤掉"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        cases = [
            HostDomain(host_id=_HOST_IP, npu_id="2", chip_phy_id="6"),
            # npu_id存在且不等于端口时，chip_phy_id等于端口也不命中（端口为npu_id）
            HostDomain(host_id=_HOST_IP, npu_id="2", chip_phy_id="5"),
            HostDomain(host_id="10.1.1.9", npu_id="2", chip_phy_id="5"),
            SwitchDomain(swi_id=_SWI_IP, interface=_SWI_IFACE),
            BmcDomain(bmc_id=_BMC_IP, npu_id="2", chip_phy_id="6"),
        ]
        for domain in cases:
            self.assertFalse(diag._is_specified_link(DiagResult(domain)))


class TestIsSpecifiedLinkRow(unittest.TestCase):
    """报告数据行（链路分析/光模块sheet行）链路匹配"""

    def test_host_to_l1_row_host_side(self):
        """host入口：主机到L1链路行按本端host_id+npu_id命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        row = HostToL1LinkData(host_id=_HOST_IP, npu_id="5", chip_phy_id="9")
        self.assertTrue(diag._is_specified_link_row(row))

    def test_host_to_l1_row_chip_phy_id_no_fallback(self):
        """npu_id缺失时不回退匹配chip_phy_id（端口仅支持npu_id）"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        row = HostToL1LinkData(host_id=_HOST_IP, npu_id="", chip_phy_id="5")
        self.assertFalse(diag._is_specified_link_row(row))

    def test_host_to_l1_row_l1_side(self):
        """switch入口：主机到L1链路行按L1交换机IP+端口命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, _SWI_IFACE)
        row = HostToL1LinkData(host_id="10.1.1.9", l1_switch_id=_SWI_IP, l1_interface=_SWI_IFACE)
        self.assertTrue(diag._is_specified_link_row(row))

    def test_l1_to_l2_row_l1_side(self):
        """switch入口：L1到L2链路行按本端L1交换机IP+端口命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, _SWI_IFACE)
        row = L1ToL2LinkData(l1_switch_id=_SWI_IP, l1_interface=_SWI_IFACE)
        self.assertTrue(diag._is_specified_link_row(row))

    def test_l1_to_l2_row_l2_side(self):
        """switch入口：L1到L2链路行按对端L2交换机IP+端口命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), "10.1.1.8", "200GE1/0/2:3")
        row = L1ToL2LinkData(
            l1_switch_id=_SWI_IP, l1_interface=_SWI_IFACE, l2_switch_id="10.1.1.8", l2_interface="200GE1/0/2:3"
        )
        self.assertTrue(diag._is_specified_link_row(row))

    def test_host_optical_row_host_side(self):
        """host入口：主机光模块行按本端host_id+npu_id命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        row = _fill_required(HostToSwitchOpticalModuleData, host_id=_HOST_IP, npu_id="5", chip_phy_id="9")
        self.assertTrue(diag._is_specified_link_row(row))

    def test_host_optical_row_peer_switch_side(self):
        """switch入口：主机光模块行按对端交换机IP+端口命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, _SWI_IFACE)
        row = _fill_required(
            HostToSwitchOpticalModuleData, host_id="10.1.1.9", peer_switch_id=_SWI_IP, peer_switch_port=_SWI_IFACE
        )
        self.assertTrue(diag._is_specified_link_row(row))

    def test_switch_optical_row_local_side(self):
        """switch入口：交换机间光模块行按本端交换机IP+端口命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, _SWI_IFACE)
        row = _fill_required(SwitchOpticalModuleData, local_switch_id=_SWI_IP, local_interface=_SWI_IFACE)
        self.assertTrue(diag._is_specified_link_row(row))

    def test_switch_optical_row_peer_side(self):
        """switch入口：交换机间光模块行按对端交换机IP+端口命中"""
        diag = AutoSingleDiag(_make_diag_ctx(), "10.1.1.8", "200GE1/0/2:3")
        row = _fill_required(
            SwitchOpticalModuleData,
            local_switch_id=_SWI_IP,
            local_interface=_SWI_IFACE,
            peer_switch_id="10.1.1.8",
            peer_interface="200GE1/0/2:3",
        )
        self.assertTrue(diag._is_specified_link_row(row))

    def test_row_no_match(self):
        """IP或端口不匹配时行被过滤掉"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        cases = [
            HostToL1LinkData(host_id=_HOST_IP, npu_id="2", chip_phy_id="6"),
            # npu_id存在且不等于端口时，chip_phy_id等于端口也不命中（端口为npu_id）
            HostToL1LinkData(host_id=_HOST_IP, npu_id="2", chip_phy_id="5"),
            HostToL1LinkData(host_id="10.1.1.9", npu_id="2", chip_phy_id="5"),
            L1ToL2LinkData(
                l1_switch_id=_SWI_IP, l1_interface=_SWI_IFACE, l2_switch_id="10.1.1.8", l2_interface="200GE1/0/2:3"
            ),
            _fill_required(HostToSwitchOpticalModuleData, host_id="10.1.1.9", npu_id="2", chip_phy_id="5"),
            _fill_required(SwitchOpticalModuleData, local_switch_id=_SWI_IP, local_interface=_SWI_IFACE),
        ]
        for row in cases:
            self.assertFalse(diag._is_specified_link_row(row))

    def test_unknown_row_type(self):
        """未知行类型不匹配"""
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        self.assertFalse(diag._is_specified_link_row(object()))


class TestSheetLinkFilter(unittest.TestCase):
    """报告sheet生成器按链路过滤数据行"""

    @staticmethod
    def _make_cluster_info():
        cluster_info = MagicMock()
        cluster_info.get_threshold.return_value = A5Threshold
        return cluster_info

    def test_signal_link_sheet_filters_rows(self):
        """链路分析sheet仅保留指定链路的上/下行数据行"""
        down_hit = HostToL1LinkData(host_id="10.1.1.9", l1_switch_id=_SWI_IP, l1_interface=_SWI_IFACE)
        down_miss = HostToL1LinkData(host_id="10.1.1.9", l1_switch_id="10.1.1.7", l1_interface="400GE1/0/2:1")
        up_hit = L1ToL2LinkData(l1_switch_id=_SWI_IP, l1_interface=_SWI_IFACE)
        up_miss = L1ToL2LinkData(l1_switch_id="10.1.1.9", l1_interface="400GE1/0/2:1")
        root_cause_filter = MagicMock()
        root_cause_filter.host_to_l1_links = [down_hit, down_miss]
        root_cause_filter.l1_to_l2_links = [up_hit, up_miss]
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, _SWI_IFACE)
        with (
            patch("ascend_fd_tk.core.report.sheet.signal_link_mapping_sheet.create_threshold_report") as ctr,
            patch("ascend_fd_tk.core.report.sheet.signal_link_mapping_sheet.generate_threshold_excel"),
        ):
            gen = SignalLinkMappingSheetGenerator(
                MagicMock(),
                excel_gen=MagicMock(),
                root_cause_filter=root_cause_filter,
                link_filter=diag._is_specified_link_row,
            )
            gen.generate_sheet()
        self.assertEqual(ctr.call_args_list[0].kwargs["data_list"], [down_hit])
        self.assertEqual(ctr.call_args_list[1].kwargs["data_list"], [up_hit])

    def test_host_optical_sheet_filters_rows(self):
        """主机光模块sheet仅保留指定链路的数据行"""
        hit = _fill_required(HostToSwitchOpticalModuleData, host_id=_HOST_IP, npu_id="5")
        miss = _fill_required(HostToSwitchOpticalModuleData, host_id="10.1.1.9", npu_id="5")
        diag = AutoSingleDiag(_make_diag_ctx(), _HOST_IP, "5")
        with (
            patch.object(
                HostToSwitchOpticalModuleSheetGenerator, "_collect_optical_module_data", return_value=[hit, miss]
            ),
            patch("ascend_fd_tk.core.report.sheet.optical_module_sheet.create_threshold_report") as ctr,
            patch("ascend_fd_tk.core.report.sheet.optical_module_sheet.generate_threshold_excel"),
        ):
            gen = HostToSwitchOpticalModuleSheetGenerator(
                self._make_cluster_info(), excel_gen=MagicMock(), link_filter=diag._is_specified_link_row
            )
            gen.generate_sheet()
        self.assertEqual(ctr.call_args.kwargs["data_list"], [hit])

    def test_switch_optical_sheet_filters_rows(self):
        """交换机间光模块sheet仅保留指定链路的数据行"""
        hit = _fill_required(SwitchOpticalModuleData, local_switch_id=_SWI_IP, local_interface=_SWI_IFACE)
        miss = _fill_required(SwitchOpticalModuleData, local_switch_id="10.1.1.9", local_interface="400GE1/0/2:1")
        diag = AutoSingleDiag(_make_diag_ctx(), _SWI_IP, _SWI_IFACE)
        with (
            patch.object(SwitchOpticalModuleSheetGenerator, "_collect_optical_module_data", return_value=[hit, miss]),
            patch("ascend_fd_tk.core.report.sheet.switch_optical_module_sheet.create_threshold_report") as ctr,
            patch("ascend_fd_tk.core.report.sheet.switch_optical_module_sheet.generate_threshold_excel"),
        ):
            gen = SwitchOpticalModuleSheetGenerator(
                self._make_cluster_info(), excel_gen=MagicMock(), link_filter=diag._is_specified_link_row
            )
            gen.generate_sheet()
        self.assertEqual(ctr.call_args.kwargs["data_list"], [hit])


class TestCliRunTask(unittest.TestCase):
    """CLI入口run_task参数校验与透传"""

    def _run(self, *args):
        """构造CLI模型并执行run_task，返回(结果, AutoSingleDiag构造Mock)"""
        with (
            patch("ascend_fd_tk.core.cli_module.cli_model.AutoSingleDiag") as single_cls,
            patch("ascend_fd_tk.core.cli_module.cli_model.asyncio.run"),
        ):
            single_cls.return_value.error_msg = ""
            cli_model = AutoSingleDiagCliModel(MagicMock(), MagicMock())
            result = cli_model.run_task(*args)
        return result, single_cls

    def test_insufficient_params(self):
        """仅传ip时提示参数不足，不执行诊断"""
        result, single_cls = self._run("1.1.1.1")
        self.assertIn("参数不足", result)
        single_cls.assert_not_called()

    def test_two_params(self):
        """两参数语法透传ip与端口"""
        result, single_cls = self._run(_HOST_IP, "5")
        self.assertIn("诊断完成", result)
        single_cls.assert_called_once()
        self.assertEqual(single_cls.call_args.args[1:], (_HOST_IP, "5"))

    def test_error_msg_returned(self):
        """诊断返回错误信息时直接透传给用户"""
        with (
            patch("ascend_fd_tk.core.cli_module.cli_model.AutoSingleDiag") as single_cls,
            patch("ascend_fd_tk.core.cli_module.cli_model.asyncio.run"),
        ):
            single_cls.return_value.error_msg = "未在缓存中找到设备 10.9.9.9"
            cli_model = AutoSingleDiagCliModel(MagicMock(), MagicMock())
            result = cli_model.run_task("10.9.9.9", "5")
        self.assertEqual(result, "未在缓存中找到设备 10.9.9.9")


class TestMainFlow(unittest.TestCase):
    """main流程：全量诊断后仅保留指定链路结果"""

    @staticmethod
    def _patch_services(diag_results):
        """将四个服务替换为Mock，AutoDiag写入指定诊断结果"""
        load_cache = MagicMock()
        load_cache.return_value.run = AsyncMock()

        def _auto_diag_factory(diag_ctx):
            instance = MagicMock()

            async def _run():
                diag_ctx.diag_result = list(diag_results)

            instance.run = _run
            return instance

        auto_diag = MagicMock(side_effect=_auto_diag_factory)
        root_cause = MagicMock()
        root_cause.return_value.run = AsyncMock()
        report = MagicMock()
        report.return_value.run = AsyncMock()
        patchers = [
            patch("ascend_fd_tk.examples.auto_diag.auto_single_diag.LoadCache", load_cache),
            patch("ascend_fd_tk.examples.auto_diag.auto_single_diag.AutoDiag", auto_diag),
            patch("ascend_fd_tk.examples.auto_diag.auto_single_diag.RootCauseAnalysis", root_cause),
            patch("ascend_fd_tk.examples.auto_diag.auto_single_diag.GenerateDiagReport", report),
        ]
        return patchers, report

    def test_filter_results_and_generate_report(self):
        """仅保留指定链路结果并生成报告"""
        link_result = DiagResult(HostDomain(host_id=_HOST_IP, npu_id="5", chip_phy_id="9"))
        other_result = DiagResult(HostDomain(host_id="10.1.1.9", npu_id="2", chip_phy_id="5"))
        diag_ctx = _make_diag_ctx()
        patchers, report = self._patch_services([link_result, other_result])
        for p in patchers:
            p.start()
        try:
            diag = AutoSingleDiag(diag_ctx, _HOST_IP, "5")
            asyncio.run(diag.main())
        finally:
            for p in patchers:
                p.stop()
        self.assertEqual(diag.error_msg, "")
        self.assertEqual(diag_ctx.diag_result, [link_result])
        report.assert_called_once_with(diag_ctx, link_filter=diag._is_specified_link_row)
        report.return_value.run.assert_awaited_once()

    def test_no_fault_on_link(self):
        """指定链路无故障时不生成报告并提示未发现故障"""
        other_result = DiagResult(HostDomain(host_id="10.1.1.9", npu_id="2", chip_phy_id="5"))
        diag_ctx = _make_diag_ctx()
        patchers, report = self._patch_services([other_result])
        for p in patchers:
            p.start()
        try:
            diag = AutoSingleDiag(diag_ctx, _HOST_IP, "5")
            asyncio.run(diag.main())
        finally:
            for p in patchers:
                p.stop()
        self.assertIn("未发现故障", diag.error_msg)
        self.assertEqual(diag_ctx.diag_result, [])
        report.return_value.run.assert_not_awaited()

    def test_invalid_device_stops_before_diag(self):
        """设备校验失败时不执行诊断"""
        diag_ctx = _make_diag_ctx()
        patchers, _ = self._patch_services([])
        for p in patchers:
            p.start()
        try:
            diag = AutoSingleDiag(diag_ctx, "10.9.9.9", "5")
            asyncio.run(diag.main())
        finally:
            for p in patchers:
                p.stop()
        self.assertIn("未在缓存中找到设备", diag.error_msg)


if __name__ == "__main__":
    unittest.main()
