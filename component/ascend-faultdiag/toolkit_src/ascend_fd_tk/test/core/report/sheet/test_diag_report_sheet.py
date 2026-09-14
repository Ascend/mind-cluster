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

"""DiagReportSheetGenerator 单元测试：None 兜底排序与 PoDManager 多槽位合并。"""

import unittest

from ascend_fd_tk.core.common.constants import FAULT_TYPE_HOST, FAULT_TYPE_SWITCH
from ascend_fd_tk.core.report.sheet.diag_report_sheet import (
    DiagReportSheetGenerator,
    HostReportData,
    SwitchReportData,
)


class TestDiagReportSheetSort(unittest.TestCase):
    def test_sort_with_none_optical_id(self):
        """A3 代际 optical_id=None 在排序键中按空串处理，不报 TypeError。"""
        data = HostReportData(host_id="host1", npu_id="0", optical_id=None, fault_domain="d1")
        self.assertEqual(
            DiagReportSheetGenerator._get_sort_key(data, FAULT_TYPE_HOST), ("host1", "0", "", "", "", "", "d1", "")
        )

    def test_sort_with_fault_time_ascending(self):
        """同实体+故障域内多条记录按故障时间升序排列（时间早的在前）。"""
        data_list = [
            HostReportData(host_id="host1", npu_id="0", fault_domain="d1", fault_time="2026-06-01 09:00:00"),
            HostReportData(host_id="host1", npu_id="0", fault_domain="d1", fault_time="2026-06-01 08:00:00"),
            HostReportData(host_id="host1", npu_id="0", fault_domain="d1", fault_time=""),
            HostReportData(host_id="host1", npu_id="1", fault_domain="d1", fault_time="2026-06-01 10:00:00"),
        ]
        sorted_data = sorted(data_list, key=lambda x: DiagReportSheetGenerator._get_sort_key(x, FAULT_TYPE_HOST))
        self.assertEqual(
            [d.fault_time for d in sorted_data],
            ["", "2026-06-01 08:00:00", "2026-06-01 09:00:00", "2026-06-01 10:00:00"],
        )

    def test_compute_col_merge_ranges_none_grouping(self):
        """_compute_col_merge_ranges 对 None 属性分组合并不报错。"""
        gen = DiagReportSheetGenerator.__new__(DiagReportSheetGenerator)
        data_list = [
            HostReportData(host_id="h1", optical_id=None),
            HostReportData(host_id="h1", optical_id=None),
            HostReportData(host_id="h2", optical_id="0"),
        ]
        ranges = gen._compute_col_merge_ranges(data_list, col_idx=0, key_attrs=["host_id"])
        self.assertEqual(ranges, [(2, 0, 3, 0)])

    def test_switch_entity_columns_merge_by_swi_id_and_slot_id(self):
        """switch sheet IP/名称/槽位号列按 swi_id+slot_id 合并：同 IP 不同槽位不合并。"""
        gen = DiagReportSheetGenerator.__new__(DiagReportSheetGenerator)
        header_mapping = DiagReportSheetGenerator._create_header_config(FAULT_TYPE_SWITCH)
        data_list = [
            SwitchReportData(swi_id="10.1.1.1", slot_id="61", interface="100GE1/1", swi_name="SW-A"),
            SwitchReportData(swi_id="10.1.1.1", slot_id="61", interface="100GE1/2", swi_name="SW-A"),
            SwitchReportData(swi_id="10.1.1.1", slot_id="62", interface="100GE1/1", swi_name="SW-B"),
            SwitchReportData(swi_id="10.1.1.1", slot_id="62", interface="100GE1/2", swi_name="SW-B"),
        ]
        ranges = gen._compute_merge_ranges(data_list, header_mapping, FAULT_TYPE_SWITCH)
        for title in ("交换机/PoDManager ID", "交换机名称", "槽位号"):
            col = list(header_mapping.values()).index(title) + 1
            self.assertEqual([r for r in ranges if r[1] == col == r[3]], [(2, col, 3, col), (4, col, 5, col)])


if __name__ == "__main__":
    unittest.main()
