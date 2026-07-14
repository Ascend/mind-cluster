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

import unittest

from ascend_fd_tk.core.fault_analyzer.bmc.bmc_err_code_analyzer import BmcErrCodeAnalyzer

from .helpers import BMC_ID, build_bmc_info, build_cluster_info, build_sel_event


class TestBmcErrCodeAnalyzer(unittest.TestCase):
    def test_analyse_creates_diag_result_for_matching_sel_event(self):
        sel = build_sel_event(
            sel_id="1",
            event_code="0x2C000007",
            event_description="NPU Board2 NPU2-1 V_AVDD12_HVCC abnormal",
        )
        cluster_info = build_cluster_info(build_bmc_info(bmc_sel_list=[sel]), include_host=True)

        results = BmcErrCodeAnalyzer(cluster_info).analyse()

        self.assertEqual(len(results), 1)
        result = results[0]
        self.assertEqual(result.domain.bmc_id, BMC_ID)
        self.assertEqual(result.domain.npu_id, "1")
        self.assertEqual(result.domain.chip_phy_id, "2")
        self.assertEqual(result.err_code, "0x2C000007")
        self.assertEqual(result.suggestion, "20A PSIP故障或者12V电容失效，请联系运维处理")
        self.assertIn("服务器server-1", result.fault_info)
        self.assertIn("系统异常下电告警，主板有电压跌落（V_AVDD12_HVCC）", result.fault_info)
        self.assertIn("NPU Board2 NPU2-1 V_AVDD12_HVCC abnormal", result.fault_info)
        self.assertIn("2026-06-01 10:00:00", result.fault_info)

    def test_analyse_ignores_unknown_code_and_keyword_mismatch(self):
        mismatch_keyword = build_sel_event(
            sel_id="1",
            event_code="0x2C000007",
            event_description="NPU Board2 NPU2-1 V_NOT_CONFIGURED abnormal",
        )
        unknown_code = build_sel_event(
            sel_id="2",
            event_code="0xFFFFFFFF",
            event_description="NPU Board1 V_AVDD12_HVCC abnormal",
            generation_time="2026-06-01 10:01:00",
        )
        cluster_info = build_cluster_info(build_bmc_info(bmc_sel_list=[mismatch_keyword, unknown_code]))

        self.assertEqual(BmcErrCodeAnalyzer(cluster_info).analyse(), [])

    def test_analyse_parses_domain_from_npu_and_ai_module_description(self):
        npu_event = build_sel_event(
            sel_id="1",
            event_code="0x56000005",
            event_description="NPU3 connection has been lost",
        )
        ai_module_event = build_sel_event(
            sel_id="2",
            event_code="0x56000009",
            event_description="AI Module4 over temperature shutdown",
            generation_time="2026-06-01 10:01:00",
        )
        cluster_info = build_cluster_info(build_bmc_info(bmc_sel_list=[npu_event, ai_module_event]))

        results = BmcErrCodeAnalyzer(cluster_info).analyse()

        self.assertEqual(len(results), 2)
        self.assertEqual(results[0].domain.bmc_id, BMC_ID)
        self.assertEqual(results[0].domain.npu_id, "2")
        self.assertEqual(results[0].domain.chip_phy_id, "")
        self.assertEqual(results[0].err_code, "0x56000005")
        self.assertIn("NPU connection has been lost告警", results[0].fault_info)
        self.assertEqual(results[1].domain.npu_id, "3")
        self.assertEqual(results[1].domain.chip_phy_id, "")
        self.assertEqual(results[1].err_code, "0x56000009")
        self.assertIn("NPU 过热关机", results[1].fault_info)


if __name__ == "__main__":
    unittest.main()
