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

from ascend_fd.pkg.parse.knowledge_graph.prechecker.merge_precheck_cause import MergePrecheckCause
from ascend_fd.pkg.parse.knowledge_graph.tools.utils import get_device_precheck_event
from ascend_fd.utils.constant.str_const import UNKNOWN_DEVICE_ID
from ascend_fd.utils.constant.ub_const import (
    AGE_PERIOD_KEYWORD,
    Comp_CANN_HCCL_Custom_CQE0x2,
    Comp_CANN_HCCL_Custom_CQE0x5,
    DFX_TM_CRD_CTRL,
    HCOMM_TA_CTP_UB_TIMEOUT,
    ICRC_ERR_COUNT_METRICS,
    LQC_TAI_DFX_ALARM,
    PHY_REINIT_CNT,
    PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT,
    PRECHECK_KERNEL_AE_TYPE23,
    PRECHECK_RXDMA_ICRC_DATA,
    PRECHECK_UBCTL_DATA,
    PRECHECK_UBMEM_TIMEOUT,
    RC_FULL_QUEUE_CNT,
    REMOTE_MEM_ACCESS_NET_TIMEOUT,
    RX_VL6_PKT_NUM,
    RX_VL7_PKT_NUM,
    RX_VL10_PKT_NUM,
    RX_VL11_PKT_NUM,
    TAI_COMPACT_ALARM,
    TAI_COMPACT_TOP_ALARM,
    TAACK_ABNORM_HEADER,
    TAACK_ABNORM_SSN,
    TP_RRP_ERR_FLG_0,
    TWP_AE_DFX,
    TX_VL6_PKT_NUM,
    TX_VL7_PKT_NUM,
    TX_VL10_PKT_NUM,
    TX_VL11_PKT_NUM,
    UB_RAS_CODES,
)
from ascend_fd.utils.load_kg_config import Schema, SchemaEntity

CQE0X2_SUB_CODES = (
    "Comp_Custom_CQE0x2_RC_NOT_ENOUGH",
    "Comp_Custom_CQE0x2_RC_QUEUE_NOT_ENOUGH",
    "Comp_Custom_CQE0x2_TP_PSN_ERROR",
    "Comp_Custom_CQE0x2_NO_ROUTER",
    "Comp_Custom_CQE0x2_CTP_NO_CLOSE",
)
CQE0X5_SUB_CODES = (
    "Comp_Custom_CQE0x5_LOST_PKG",
    "Comp_Custom_CQE0x5_TIMEOUT_TA_CTP_LOW",
    "Comp_Custom_CQE0x5_TIMEOUT_RETRAINING",
    "Comp_Custom_CQE0x5_ABN_PKT_SSN",
    "Comp_Custom_CQE0x5_ABN_PKT_HEADER",
    "Comp_Custom_CQE0x5_RQE_NOT_ENOUGH",
    "Comp_Custom_CQE0x5_RC_NOT_ENOUGH",
    "Comp_Custom_CQE0x5_FLOW_CFG_ERR",
    "Comp_Custom_CQE0x5_TPM_CFG_ERR",
    "Comp_Custom_CQE0x5_RXDMA_ICRC_ERR",
)
UBMEM_SUB_CODES = (
    "Comp_Custom_UBMEM_TIMEOUT_LOW",
    "Comp_Custom_UBMEM_TIMEOUT_RETRAINING",
    "Comp_Custom_UBMEM_TIMEOUT_LOST_PKG",
)


def make_schema():
    """构造最小 schema：为全部基础码/细分码注册结构化属性"""
    schema = Schema(config_pkg_list=[])
    all_codes = (
        (Comp_CANN_HCCL_Custom_CQE0x2,)
        + (Comp_CANN_HCCL_Custom_CQE0x5,)
        + (REMOTE_MEM_ACCESS_NET_TIMEOUT,)
        + CQE0X2_SUB_CODES
        + CQE0X5_SUB_CODES
        + UBMEM_SUB_CODES
    )
    for code in all_codes:
        schema.add_custom_event_to_schema(
            code,
            SchemaEntity(
                code,
                {
                    "class": "3",
                    "component": "UB",
                    "module": "UB",
                    "cause_zh": "故障_" + code,
                    "cause_en": "fault_" + code,
                    "description_zh": code,
                    "description_en": code,
                },
            ),
        )
    # RAS 故障码也可能作为 UBMem 链路源端，需可查
    for ras_code in UB_RAS_CODES:
        schema.add_custom_event_to_schema(
            ras_code,
            SchemaEntity(
                ras_code,
                {
                    "class": "3",
                    "component": "UB",
                    "module": "UB",
                    "cause_zh": "故障_" + ras_code,
                    "cause_en": "fault_" + ras_code,
                },
            ),
        )
    return schema


def ubctl_event(source_device="0", before=None, after=None, occur_time="2026-01-01 00:00:00"):
    return {
        "event_code": PRECHECK_UBCTL_DATA,
        "source_device": source_device,
        "occur_time": occur_time,
        "is_custom_event": False,
        "type": "NPU_UBCTL",
        "source_file": "ubctl_log.txt",
        "key_info": "ubctl before/after raw data",
        "occurrence": [],
        "attribute": {
            "before": before or {"udie0": {}, "udie1": {}},
            "after": after or {"udie0": {}, "udie1": {}},
            "source_lines": {"before": [], "after": []},
        },
    }


class PrecheckCheckerTestBase(unittest.TestCase):
    def setUp(self):
        self.schema = make_schema()
        self.source_device = "0"

    # ---------- 构造辅助 ----------
    def base_entry(self, code, chain_key="0"):
        """模拟 plog 解析出的基础故障条目"""
        return {
            code: {
                "code": code,
                "entities_attribute": {"class": "0", "component": ""},
                "events_attribute": [{"event_code": code, "key_info": "plog line"}],
                "chains": {chain_key: "base chain"},
            }
        }

    def run_analyze(self, precheck_info, device_causes, source_device=None):
        merge = MergePrecheckCause(self.schema, precheck_info)
        merge.single_device_analyze(source_device or self.source_device, device_causes)
        return device_causes


class TestGetDevicePrecheckEvent(unittest.TestCase):
    def test_unknown_device_takes_all_events(self):
        info = {
            UNKNOWN_DEVICE_ID: [
                {"event_code": "PRECHECK_UBCTL_DATA"},
                {"event_code": "Comp_CANN_HCCL_Custom_CQE0x2"},
                {"event_code": "0x81AF8009"},
            ]
        }
        events = get_device_precheck_event(info, UNKNOWN_DEVICE_ID)
        self.assertEqual(set(events.keys()), {"PRECHECK_UBCTL_DATA", "Comp_CANN_HCCL_Custom_CQE0x2", "0x81AF8009"})

    def test_normal_device_takes_precheck_only(self):
        info = {"0": [{"event_code": "PRECHECK_UBCTL_DATA"}, {"event_code": "Comp_CANN_HCCL_Custom_CQE0x2"}]}
        events = get_device_precheck_event(info, "0")
        self.assertEqual(set(events.keys()), {"PRECHECK_UBCTL_DATA"})

    def test_missing_device_returns_empty(self):
        self.assertEqual(get_device_precheck_event({}, "0"), {})
        self.assertEqual(get_device_precheck_event({"0": []}, "0"), {})


class TestCqe0x2CheckerIntegration(PrecheckCheckerTestBase):
    BASE_CODE = Comp_CANN_HCCL_Custom_CQE0x2

    def test_rc_not_enough_by_rule1(self):
        sub_code = "Comp_Custom_CQE0x2_RC_NOT_ENOUGH"
        after = {"udie0": {0: {RC_FULL_QUEUE_CNT: "0x1"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)
        entry = result[sub_code]
        self.assertEqual(entry["code"], sub_code)
        # 仅核对故障链/事件已生成，不比对具体文案，避免文案调整导致用例失效
        self.assertTrue(entry["chains"][self.source_device])
        precheck_event, base_event = entry["events_attribute"]
        self.assertTrue(precheck_event["key_info"])
        self.assertTrue(precheck_event["occurrence"])
        self.assertEqual(base_event, self.base_entry(self.BASE_CODE)[self.BASE_CODE])
        self.assertEqual(entry["entities_attribute"], self.schema.get_schema_entity(sub_code).attribute.to_json())

    def test_rc_queue_not_enough_by_rule2(self):
        sub_code = "Comp_Custom_CQE0x2_RC_QUEUE_NOT_ENOUGH"
        after = {"udie0": {0: {TP_RRP_ERR_FLG_0: "0x10000000"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_tp_psn_error_requires_rule3_and_ae_event(self):
        sub_code = "Comp_Custom_CQE0x2_TP_PSN_ERROR"
        after = {"udie0": {0: {TWP_AE_DFX: "0x8"}}, "udie1": {}}
        ae_event = {
            "event_code": PRECHECK_KERNEL_AE_TYPE23,
            "source_device": "0",
            "key_info": "UDMA ae event type is 2 sub type is 3",
            "attribute": {},
            "is_custom_event": False,
        }
        # 仅规则命中、无 AE 事件：不输出
        result_no_ae = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertNotIn(sub_code, result_no_ae)
        self.assertIn(self.BASE_CODE, result_no_ae)
        # 规则命中 + AE 事件：输出
        result_with_ae = self.run_analyze({"0": [ubctl_event(after=after), ae_event]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result_with_ae)
        self.assertNotIn(self.BASE_CODE, result_with_ae)

    def test_no_router_by_rule4(self):
        sub_code = "Comp_Custom_CQE0x2_NO_ROUTER"
        after = {"udie0": {0: {TAI_COMPACT_ALARM: "0x4"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_ctp_no_close_by_rule5(self):
        sub_code = "Comp_Custom_CQE0x2_CTP_NO_CLOSE"
        after = {"udie0": {0: {TAI_COMPACT_TOP_ALARM: "0x40"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_base_code_not_present_returns_early(self):
        after = {"udie0": {0: {RC_FULL_QUEUE_CNT: "0x1"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, {"AISW_CANN_OTHER": {"code": "AISW_CANN_OTHER"}})
        # 规则命中但无基础码：不产生细分码；PRECHECK 事件会并入 device_causes
        for sub_code in CQE0X2_SUB_CODES:
            self.assertNotIn(sub_code, result)
        self.assertIn("AISW_CANN_OTHER", result)
        self.assertIn(PRECHECK_UBCTL_DATA, result)

    def test_no_precheck_event_keeps_base_code(self):
        # 无 PRECHECK_UBCTL_DATA 事件：规则不触发，基础码保留、不出细分码
        result = self.run_analyze({"0": []}, self.base_entry(self.BASE_CODE))
        self.assertEqual(set(result.keys()), {self.BASE_CODE})

    def test_no_rule_hit_keeps_base_code(self):
        after = {"udie0": {0: {RC_FULL_QUEUE_CNT: "0x0"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(self.BASE_CODE, result)
        for sub_code in CQE0X2_SUB_CODES:
            self.assertNotIn(sub_code, result)


class TestCqe0x3Checker(PrecheckCheckerTestBase):
    def test_no_op(self):
        # 预留实现：始终不产生任何变更
        from ascend_fd.pkg.parse.knowledge_graph.prechecker.ub_cqe_checker import Cqe0x3Checker

        merge = MergePrecheckCause(self.schema, {})
        device_causes = {"X": {"code": "X"}}
        checker = Cqe0x3Checker("0", merge)
        checker.analyze(device_causes)
        self.assertEqual(device_causes, {"X": {"code": "X"}})


class TestUBMemCheckerIntegration(PrecheckCheckerTestBase):
    BASE_CODE = REMOTE_MEM_ACCESS_NET_TIMEOUT

    def vlan_unbalanced_tree(self):
        # v10 RX 增量 2 != TX 增量 1，v11 平衡，触发 RULE_VLAN10_11_BALANCE
        after = {
            "udie0": {
                0: {RX_VL10_PKT_NUM: "0x2", TX_VL10_PKT_NUM: "0x1", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"}
            },
            "udie1": {},
        }
        before = {
            "udie0": {
                0: {RX_VL10_PKT_NUM: "0x0", TX_VL10_PKT_NUM: "0x0", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"}
            },
            "udie1": {},
        }
        return before, after

    def test_ub_ras_chain_appended_with_lost_package(self):
        ras_code = UB_RAS_CODES[0]
        device_causes = self.base_entry(self.BASE_CODE)
        device_causes[ras_code] = {"code": ras_code, "entities_attribute": {}, "events_attribute": [], "chains": {}}
        before, after = self.vlan_unbalanced_tree()
        result = self.run_analyze({"0": [ubctl_event(before=before, after=after)]}, device_causes)
        # 基础码被移除，RAS 故障链已被更新（生成含说明的故障链）
        self.assertNotIn(self.BASE_CODE, result)
        self.assertTrue(result[ras_code]["chains"][self.source_device])
        # 无 RAS 损耗情况下，不应输出 LOST_PKG 细分码
        self.assertNotIn("Comp_Custom_UBMEM_TIMEOUT_LOST_PKG", result)

    def test_ubmem_timeout_low_by_rule16(self):
        sub_code = "Comp_Custom_UBMEM_TIMEOUT_LOW"
        ubmem_event = {
            "event_code": PRECHECK_UBMEM_TIMEOUT,
            "source_device": self.source_device,
            "occur_time": "2026-01-01 00:00:00",
            "key_info": "age_period lines",
            "attribute": {AGE_PERIOD_KEYWORD: "3000000"},
            "is_custom_event": False,
        }
        result = self.run_analyze({"0": [ubctl_event(), ubmem_event]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)
        precheck_event, base_event = result[sub_code]["events_attribute"]
        self.assertEqual(precheck_event["event_code"], PRECHECK_UBMEM_TIMEOUT)
        # _build_ubmem_cause 基础码事件取 unknown/设备事件；本场景未知前缀事件为空
        self.assertEqual(base_event, {})

    def test_ubmem_timeout_retraining_by_rule6(self):
        sub_code = "Comp_Custom_UBMEM_TIMEOUT_RETRAINING"
        after = {"udie0": {0: {PHY_REINIT_CNT: "0x6"}}, "udie1": {}}
        before = {"udie0": {0: {PHY_REINIT_CNT: "0x0"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(before=before, after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_ubmem_timeout_lost_pkg_without_ras(self):
        sub_code = "Comp_Custom_UBMEM_TIMEOUT_LOST_PKG"
        before, after = self.vlan_unbalanced_tree()
        result = self.run_analyze({"0": [ubctl_event(before=before, after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)
        # 仅核对包含期望来源设备的故障链已生成
        self.assertTrue(result[sub_code]["chains"][self.source_device])

    def test_no_base_code_no_output(self):
        before, after = self.vlan_unbalanced_tree()
        result = self.run_analyze({"0": [ubctl_event(before=before, after=after)]}, {"X": {"code": "X"}})
        # X 保留、PRECHECK 事件并入，但无 UBMem 基础码时不产出细分码
        self.assertIn("X", result)
        for sub_code in UBMEM_SUB_CODES:
            self.assertNotIn(sub_code, result)


class TestCqe0x5CheckerIntegration(PrecheckCheckerTestBase):
    BASE_CODE = Comp_CANN_HCCL_Custom_CQE0x5

    def test_lost_pkg_by_rule8(self):
        sub_code = "Comp_Custom_CQE0x5_LOST_PKG"
        after = {
            "udie0": {0: {RX_VL7_PKT_NUM: "0x3", TX_VL7_PKT_NUM: "0x2", RX_VL6_PKT_NUM: "0x1", TX_VL6_PKT_NUM: "0x1"}},
            "udie1": {},
        }
        before = {
            "udie0": {0: {RX_VL7_PKT_NUM: "0x0", TX_VL7_PKT_NUM: "0x0", RX_VL6_PKT_NUM: "0x0", TX_VL6_PKT_NUM: "0x0"}},
            "udie1": {},
        }
        result = self.run_analyze({"0": [ubctl_event(before=before, after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_timeout_ta_ctp_low_by_rule15(self):
        sub_code = "Comp_Custom_CQE0x5_TIMEOUT_TA_CTP_LOW"
        hcomm_event = {
            "event_code": PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT,
            "source_device": self.source_device,
            "occur_time": "2026-01-01 00:00:00",
            "key_info": "hcomm env line",
            "attribute": {HCOMM_TA_CTP_UB_TIMEOUT: "4"},
            "is_custom_event": False,
        }
        result = self.run_analyze({"0": [ubctl_event(), hcomm_event]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)
        precheck_event, _ = result[sub_code]["events_attribute"]
        self.assertEqual(precheck_event["event_code"], PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT)

    def test_timeout_retraining_by_rule6(self):
        sub_code = "Comp_Custom_CQE0x5_TIMEOUT_RETRAINING"
        after = {"udie0": {0: {PHY_REINIT_CNT: "0x6"}}, "udie1": {}}
        before = {"udie0": {0: {PHY_REINIT_CNT: "0x0"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(before=before, after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_abn_pkt_ssn_by_rule9(self):
        sub_code = "Comp_Custom_CQE0x5_ABN_PKT_SSN"
        after = {"udie0": {0: {TAACK_ABNORM_SSN: "0x1"}}, "udie1": {}}
        before = {"udie0": {0: {TAACK_ABNORM_SSN: "0x0"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(before=before, after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_abn_pkt_header_by_rule10(self):
        sub_code = "Comp_Custom_CQE0x5_ABN_PKT_HEADER"
        after = {"udie0": {0: {TAACK_ABNORM_HEADER: "0x2"}}, "udie1": {}}
        before = {"udie0": {0: {TAACK_ABNORM_HEADER: "0x0"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(before=before, after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_rqe_not_enough_by_rule11(self):
        sub_code = "Comp_Custom_CQE0x5_RQE_NOT_ENOUGH"
        after = {"udie0": {0: {TP_RRP_ERR_FLG_0: "0x2000000"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_rc_not_enough_by_rule12(self):
        sub_code = "Comp_Custom_CQE0x5_RC_NOT_ENOUGH"
        after = {"udie0": {0: {TP_RRP_ERR_FLG_0: "0x18000000"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_flow_cfg_err_by_rule13(self):
        sub_code = "Comp_Custom_CQE0x5_FLOW_CFG_ERR"
        after = {"udie0": {0: {DFX_TM_CRD_CTRL: "0x2"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_tpm_cfg_err_by_rule14(self):
        sub_code = "Comp_Custom_CQE0x5_TPM_CFG_ERR"
        after = {"udie0": {0: {LQC_TAI_DFX_ALARM: "0x30"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)

    def test_icrc_err_by_icrc_growth(self):
        sub_code = "Comp_Custom_CQE0x5_RXDMA_ICRC_ERR"
        icrc_event = {
            "event_code": PRECHECK_RXDMA_ICRC_DATA,
            "source_device": self.source_device,
            "key_info": "icrc raw",
            "attribute": {
                "before": {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: 5}}},
                "after": {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: 10}}},
            },
            "is_custom_event": False,
        }
        result = self.run_analyze({"0": [ubctl_event(), icrc_event]}, self.base_entry(self.BASE_CODE))
        self.assertIn(sub_code, result)
        self.assertNotIn(self.BASE_CODE, result)
        precheck_event, _ = result[sub_code]["events_attribute"]
        self.assertEqual(precheck_event["event_code"], PRECHECK_RXDMA_ICRC_DATA)

    def test_icrc_no_growth_no_output(self):
        sub_code = "Comp_Custom_CQE0x5_RXDMA_ICRC_ERR"
        icrc_event = {
            "event_code": PRECHECK_RXDMA_ICRC_DATA,
            "source_device": self.source_device,
            "attribute": {
                "before": {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: 10}}},
                "after": {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: 10}}},
            },
        }
        result = self.run_analyze({"0": [ubctl_event(), icrc_event]}, self.base_entry(self.BASE_CODE))
        self.assertNotIn(sub_code, result)

    def test_no_rule_hit_keeps_base_code(self):
        after = {"udie0": {0: {RC_FULL_QUEUE_CNT: "0x0"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn(self.BASE_CODE, result)
        for sub_code in CQE0X5_SUB_CODES:
            self.assertNotIn(sub_code, result)

    def test_base_code_not_present_returns_early(self):
        after = {"udie0": {0: {RC_FULL_QUEUE_CNT: "0x1"}}, "udie1": {}}
        result = self.run_analyze({"0": [ubctl_event(after=after)]}, {"AISW_CANN_OTHER": {"code": "AISW_CANN_OTHER"}})
        # 无基础码：不产出任何 Cqe0x5 细分码；PRECHECK 事件并入 device_causes
        for sub_code in CQE0X5_SUB_CODES:
            self.assertNotIn(sub_code, result)
        self.assertIn("AISW_CANN_OTHER", result)
        self.assertIn(PRECHECK_UBCTL_DATA, result)
