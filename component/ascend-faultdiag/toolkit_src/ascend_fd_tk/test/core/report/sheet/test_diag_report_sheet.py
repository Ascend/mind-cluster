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

"""DiagReportSheetGenerator 单元测试。

覆盖本次改动：
- _get_sort_key: None 值用 `or ""` 兜底，避免 A3 代际 optical_id=None 排序报 TypeError
- _compute_col_merge_ranges: None 值用 `or ""` 兜底，确保跨代际兼容
"""

import unittest

from ascend_fd_tk.core.common.constants import FAULT_TYPE_HOST
from ascend_fd_tk.core.report.sheet.diag_report_sheet import (
    DiagReportSheetGenerator,
    HostReportData,
)


class TestDiagReportSheetSort(unittest.TestCase):
    def test_get_sort_key_none_optical_id(self):
        """A3 代际 optical_id 为 None 时，_get_sort_key 应转为空串，不报 TypeError。"""
        # HostReportData 默认 optical_id=""，但模拟 A3 反序列化场景显式置 None
        data = HostReportData(host_id="host1", npu_id="0", optical_id=None, fault_domain="d1")
        key = DiagReportSheetGenerator._get_sort_key(data, FAULT_TYPE_HOST)
        # 6 个实体排序属性 + fault_domain + fault_time = 8 个元素
        self.assertEqual(key, ("host1", "0", "", "", "", "", "d1", ""))

    def test_sort_with_none_value_no_typeerror(self):
        """排序含 None 的数据不应报 TypeError。"""
        data_list = [
            HostReportData(host_id="host2", optical_id=None, fault_domain="d2"),
            HostReportData(host_id="host1", optical_id="0", fault_domain="d1"),
            HostReportData(host_id="host1", optical_id=None, fault_domain="d0"),
        ]
        sorted_data = sorted(data_list, key=lambda x: DiagReportSheetGenerator._get_sort_key(x, FAULT_TYPE_HOST))
        # host1 在前（含 optical_id="" 和 "0"），host2 在后
        self.assertEqual(sorted_data[0].host_id, "host1")
        self.assertEqual(sorted_data[-1].host_id, "host2")

    def test_sort_with_fault_time_ascending(self):
        """同实体+故障域内多条记录按故障时间升序排列（时间早的在前）。"""
        data_list = [
            HostReportData(host_id="host1", npu_id="0", fault_domain="d1", fault_time="2026-06-01 09:00:00"),
            HostReportData(host_id="host1", npu_id="0", fault_domain="d1", fault_time="2026-06-01 08:00:00"),
            HostReportData(host_id="host1", npu_id="0", fault_domain="d1", fault_time=""),
            HostReportData(host_id="host1", npu_id="1", fault_domain="d1", fault_time="2026-06-01 10:00:00"),
        ]
        sorted_data = sorted(data_list, key=lambda x: DiagReportSheetGenerator._get_sort_key(x, FAULT_TYPE_HOST))
        # host1/npu0/d1 组内按故障时间升序：空串最早，08:00 次之，09:00 最后
        self.assertEqual(
            [d.fault_time for d in sorted_data],
            ["", "2026-06-01 08:00:00", "2026-06-01 09:00:00", "2026-06-01 10:00:00"],
        )

    def test_compute_col_merge_ranges_none_grouping(self):
        """_compute_col_merge_ranges 对 None 属性分组合并，不报错。

        使用 __new__ 绕过 __init__ 避免依赖 cluster_info 等参数。
        """
        gen = DiagReportSheetGenerator.__new__(DiagReportSheetGenerator)
        data_list = [
            HostReportData(host_id="h1", optical_id=None),
            HostReportData(host_id="h1", optical_id=None),
            HostReportData(host_id="h2", optical_id="0"),
        ]
        # 按 host_id 分组（None optical_id 应被当作相同 key 处理）
        ranges = gen._compute_col_merge_ranges(data_list, col_idx=0, key_attrs=["host_id"])
        # 第一组 (h1, h1) 合并：seg_start=0, seg_end=2, _TWO_ROW=2 -> (2, 0, 3, 0)
        self.assertEqual(len(ranges), 1)
        self.assertEqual(ranges[0], (2, 0, 3, 0))


if __name__ == "__main__":
    unittest.main()
