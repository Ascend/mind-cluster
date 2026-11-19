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
from unittest.mock import patch

from ascend_fd.configuration.config import KNOWLEDGE_GRAPH_CONF
from ascend_fd.pkg.parse.knowledge_graph.parser.lcne_parser import LCNEParser
from ascend_fd.pkg.parse.knowledge_graph.prechecker.merge_precheck_cause import MergePrecheckCause
from ascend_fd.pkg.parse.knowledge_graph.tools.utils import get_device_precheck_event
from ascend_fd.utils.constant.str_const import UNKNOWN_DEVICE_ID
from ascend_fd.utils.constant.ub_const import (
    AGE_PERIOD_KEYWORD,
    Comp_CANN_HCCL_Custom_CQE0x2,
    Comp_CANN_HCCL_Custom_CQE0x5,
    Comp_Custom_CQE0x5_LOST_PKG_EAD,
    Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD,
    DFX_TM_CRD_CTRL,
    HCOMM_TA_CTP_UB_TIMEOUT,
    ICRC_ERR_COUNT_METRICS,
    LQC_TAI_DFX_ALARM,
    NIC_TX_L2_ERR_DROP_PKT_CNT,
    PHY_REINIT_CNT,
    PPP_RLT_DROP_NUM,
    PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT,
    PRECHECK_KERNEL_AE_TYPE23,
    PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA,
    PRECHECK_RXDMA_ICRC_DATA,
    PRECHECK_UBCTL_DATA,
    PRECHECK_UBMEM_TIMEOUT,
    RC_FULL_QUEUE_CNT,
    REMOTE_MEM_ACCESS_NET_TIMEOUT,
    RX_VL6_PKT_NUM,
    RX_VL7_PKT_NUM,
    RX_VL10_PKT_NUM,
    RX_VL11_PKT_NUM,
    SNAPSHOOT_TAG,
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
    Comp_Custom_CQE0x5_LOST_PKG_EAD,
)
UBMEM_SUB_CODES = (
    "Comp_Custom_UBMEM_TIMEOUT_LOW",
    "Comp_Custom_UBMEM_TIMEOUT_RETRAINING",
    "Comp_Custom_UBMEM_TIMEOUT_LOST_PKG",
    Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD,
)


def udie0(metrics):
    """udie0/port0 指标树，udie1 为空"""
    return {"udie0": {0: metrics}, "udie1": {}}


# VL10/11 不平衡（UBMem 丢包规则）与 VL6/7 不平衡（CQE0x5 丢包规则）的公共指标树
VL10_BEFORE = udie0({RX_VL10_PKT_NUM: "0x0", TX_VL10_PKT_NUM: "0x0", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"})
VL10_UNBALANCED = udie0(
    {RX_VL10_PKT_NUM: "0x2", TX_VL10_PKT_NUM: "0x1", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"}
)
VL67_BEFORE = udie0({RX_VL7_PKT_NUM: "0x0", TX_VL7_PKT_NUM: "0x0", RX_VL6_PKT_NUM: "0x0", TX_VL6_PKT_NUM: "0x0"})
VL67_UNBALANCED = udie0({RX_VL7_PKT_NUM: "0x3", TX_VL7_PKT_NUM: "0x2", RX_VL6_PKT_NUM: "0x1", TX_VL6_PKT_NUM: "0x1"})
VL10_BALANCED = udie0({RX_VL10_PKT_NUM: "0x2", TX_VL10_PKT_NUM: "0x2"})
VL67_BALANCED = udie0({RX_VL7_PKT_NUM: "0x2", TX_VL7_PKT_NUM: "0x2"})
# 同时触发 VL10/11 与 VL6/7 不平衡（UBMem + CQE0x5 两条丢包规则）
BOTH_BEFORE = udie0({**VL10_BEFORE["udie0"][0], **VL67_BEFORE["udie0"][0]})
BOTH_UNBALANCED = udie0({**VL10_UNBALANCED["udie0"][0], **VL67_UNBALANCED["udie0"][0]})
# snapshoot 端侧丢包增长/未增长
SNAP_GROWTH_BEFORE = udie0({PPP_RLT_DROP_NUM: "0x1"})
SNAP_GROWTH_AFTER = udie0({PPP_RLT_DROP_NUM: "0x2"})
SNAP_FLAT = udie0({NIC_TX_L2_ERR_DROP_PKT_CNT: "0x2"})


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


def snapshoot_event(before, after, source_device=UNKNOWN_DEVICE_ID):
    """构造 lcne nsdk_hccl_shapshot 预检事件（默认来自未知设备，按全局事件生效）"""
    return {
        "event_code": PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA,
        "source_device": source_device,
        "occur_time": "2026-01-01 00:00:00",
        "is_custom_event": False,
        "type": "LCNELog",
        "source_file": "nsdk_hccl_shapshot.log",
        "key_info": "snapshoot before/after raw data",
        "occurrence": [],
        "attribute": {
            "before": before,
            "after": after,
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

    def test_ubctl_and_snapshoot_events_coexist(self):
        # 8.2-6：snapshoot 独立事件码与 PRECHECK_UBCTL_DATA 同设备共存，按 event_code 分别保留不互踩
        info = {
            "0": [
                {"event_code": PRECHECK_UBCTL_DATA},
                {"event_code": PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA},
            ]
        }
        events = get_device_precheck_event(info, "0")
        self.assertEqual(set(events.keys()), {PRECHECK_UBCTL_DATA, PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA})


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

    def test_lost_pkg_without_ead_diagnoses_old_code(self):
        """场景 1：未命中端侧主动丢包时，仍输出旧丢包码"""
        before, after = self.vlan_unbalanced_tree()
        result = self.run_analyze({"0": [ubctl_event(before=before, after=after)]}, self.base_entry(self.BASE_CODE))
        self.assertIn("Comp_Custom_UBMEM_TIMEOUT_LOST_PKG", result)
        self.assertNotIn(Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD, result)

    def test_lost_pkg_with_ead_shows_only_new_code(self):
        """场景 2：VL 不平衡 + snapshoot 端侧丢包增长，只输出 EAD 新码"""
        before, after = self.vlan_unbalanced_tree()
        precheck_info = {
            "0": [ubctl_event(before=before, after=after)],
            UNKNOWN_DEVICE_ID: [snapshoot_event(SNAP_GROWTH_BEFORE, SNAP_GROWTH_AFTER)],
        }
        result = self.run_analyze(precheck_info, self.base_entry(self.BASE_CODE))
        self.assertIn(Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD, result)
        self.assertNotIn("Comp_Custom_UBMEM_TIMEOUT_LOST_PKG", result)
        self.assertNotIn(self.BASE_CODE, result)
        precheck_event, _ = result[Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD]["events_attribute"]
        self.assertEqual(precheck_event["event_code"], PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA)
        chain = result[Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD]["chains"][self.source_device]
        # 与旧 UBMem 丢包码链路保持一致：保留 lost package，再追加端侧主动丢包证据
        self.assertIn("+lost package(", chain)
        self.assertIn("end-side active drop(", chain)

    def test_ead_without_vlan_imbalance_keeps_base_code(self):
        """snapshoot 有增长但 VL 平衡：不产生任何丢包细分码"""
        result = self.run_analyze(
            {
                "0": [ubctl_event(before=VL10_BEFORE, after=VL10_BALANCED)],
                UNKNOWN_DEVICE_ID: [snapshoot_event(SNAP_GROWTH_BEFORE, SNAP_GROWTH_AFTER)],
            },
            self.base_entry(self.BASE_CODE),
        )
        self.assertIn(self.BASE_CODE, result)
        self.assertNotIn(Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD, result)
        self.assertNotIn("Comp_Custom_UBMEM_TIMEOUT_LOST_PKG", result)

    def test_ead_without_growth_keeps_old_code(self):
        """snapshoot 计数未增长：回退旧丢包码"""
        before, after = self.vlan_unbalanced_tree()
        result = self.run_analyze(
            {
                "0": [ubctl_event(before=before, after=after)],
                UNKNOWN_DEVICE_ID: [snapshoot_event(SNAP_FLAT, SNAP_FLAT)],
            },
            self.base_entry(self.BASE_CODE),
        )
        self.assertIn("Comp_Custom_UBMEM_TIMEOUT_LOST_PKG", result)
        self.assertNotIn(Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD, result)

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
        self.assertNotIn(Comp_Custom_CQE0x5_LOST_PKG_EAD, result)

    def test_lost_pkg_with_ead_shows_only_new_code(self):
        """CQE0x5：VL6/7 不平衡 + snapshoot 端侧丢包增长，只输出 EAD 新码"""
        precheck_info = {
            "0": [ubctl_event(before=VL67_BEFORE, after=VL67_UNBALANCED)],
            UNKNOWN_DEVICE_ID: [snapshoot_event(SNAP_GROWTH_BEFORE, SNAP_GROWTH_AFTER)],
        }
        result = self.run_analyze(precheck_info, self.base_entry(self.BASE_CODE))
        self.assertIn(Comp_Custom_CQE0x5_LOST_PKG_EAD, result)
        self.assertNotIn("Comp_Custom_CQE0x5_LOST_PKG", result)
        self.assertNotIn(self.BASE_CODE, result)
        precheck_event, _ = result[Comp_Custom_CQE0x5_LOST_PKG_EAD]["events_attribute"]
        self.assertEqual(precheck_event["event_code"], PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA)
        chain = result[Comp_Custom_CQE0x5_LOST_PKG_EAD]["chains"][self.source_device]
        # 与旧 CQE0x5 丢包码链路保持一致：旧链路不带 lost package，只追加端侧主动丢包证据
        self.assertNotIn("+lost package(", chain)
        self.assertIn("end-side active drop(", chain)

    def test_ead_without_vlan_imbalance_keeps_base_code(self):
        """snapshoot 有增长但 VL6/7 平衡：不产生丢包细分码"""
        result = self.run_analyze(
            {
                "0": [ubctl_event(before=VL67_BEFORE, after=VL67_BALANCED)],
                UNKNOWN_DEVICE_ID: [snapshoot_event(SNAP_GROWTH_BEFORE, SNAP_GROWTH_AFTER)],
            },
            self.base_entry(self.BASE_CODE),
        )
        self.assertIn(self.BASE_CODE, result)
        self.assertNotIn(Comp_Custom_CQE0x5_LOST_PKG_EAD, result)
        self.assertNotIn("Comp_Custom_CQE0x5_LOST_PKG", result)

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


class TestEadOutputGating(PrecheckCheckerTestBase):
    """8.2-4：端侧主动丢包细分码的输出门禁（基础码在场才出码，双基础码出双码，均不在场静默）"""

    def _both_rule_info(self):
        return {
            "0": [ubctl_event(before=BOTH_BEFORE, after=BOTH_UNBALANCED)],
            UNKNOWN_DEVICE_ID: [snapshoot_event(SNAP_GROWTH_BEFORE, SNAP_GROWTH_AFTER)],
        }

    def test_both_base_codes_produce_both_ead_codes(self):
        device_causes = self.base_entry(REMOTE_MEM_ACCESS_NET_TIMEOUT)
        device_causes.update(self.base_entry(Comp_CANN_HCCL_Custom_CQE0x5))
        result = self.run_analyze(self._both_rule_info(), device_causes)
        self.assertIn(Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD, result)
        self.assertIn(Comp_Custom_CQE0x5_LOST_PKG_EAD, result)
        self.assertNotIn(REMOTE_MEM_ACCESS_NET_TIMEOUT, result)
        self.assertNotIn(Comp_CANN_HCCL_Custom_CQE0x5, result)

    def test_rule_hit_without_base_codes_is_silent(self):
        # 规则命中（VL 不平衡 + snapshoot 增长）但两个基础码均不在场：不产出任何 EAD 细分码
        result = self.run_analyze(self._both_rule_info(), {"AISW_CANN_OTHER": {"code": "AISW_CANN_OTHER"}})
        self.assertIn("AISW_CANN_OTHER", result)
        self.assertNotIn(Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD, result)
        self.assertNotIn(Comp_Custom_CQE0x5_LOST_PKG_EAD, result)


class TestEadEndToEnd(PrecheckCheckerTestBase):
    """snapshoot 原始日志 -> 解析事件 -> 规则置位 -> 出码 全链路诊断验证"""

    SNAP_TAG = "2026-11-03 09:29:06.920 [INFO]" + SNAPSHOOT_TAG

    @staticmethod
    def _pair_log_lines():
        after_tag = "2026-11-03 10:29:06.920 [INFO]" + SNAPSHOOT_TAG
        return [
            f"{TestEadEndToEnd.SNAP_TAG}devid = 0, dieid = 0, start_end = 0, hccl_pid = 1, hccp_pid = 2",
            f"{TestEadEndToEnd.SNAP_TAG}port              0            1",
            f"{TestEadEndToEnd.SNAP_TAG}PPP_RLT_DROP_NUM  10           20",
            f"{after_tag}devid = 0, dieid = 0, start_end = 1, hccl_pid = 1, hccp_pid = 2",
            f"{after_tag}port              0            1",
            f"{after_tag}PPP_RLT_DROP_NUM  11           20",
        ]

    def parse_snapshoot_events(self):
        """走真实 parser，把日志行解析成 PRECHECK_LCNE_HCCL_SNAPSHOOT_DATA 事件"""
        parser = LCNEParser({"default_conf": {}, "user_conf": {}})

        class CtxStub:
            is_sdk_input = False
            parse_file_path = None

        with (
            patch.object(parser, "find_log", return_value=["nsdk_hccl_shapshot.log"]),
            patch.object(parser, "_yield_log", return_value=self._pair_log_lines()),
        ):
            events, error = parser.parse(CtxStub(), "task_id")
        self.assertEqual(error, {})
        self.assertEqual(len(events), 1)
        return events

    def test_ead_diagnosed_from_raw_log(self):
        cases = [
            (
                REMOTE_MEM_ACCESS_NET_TIMEOUT,
                VL10_BEFORE,
                VL10_UNBALANCED,
                Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD,
                "Comp_Custom_UBMEM_TIMEOUT_LOST_PKG",
            ),
            (
                Comp_CANN_HCCL_Custom_CQE0x5,
                VL67_BEFORE,
                VL67_UNBALANCED,
                Comp_Custom_CQE0x5_LOST_PKG_EAD,
                "Comp_Custom_CQE0x5_LOST_PKG",
            ),
        ]
        for base_code, before, after, ead_code, old_code in cases:
            with self.subTest(base_code=base_code):
                result = self.run_analyze(
                    {"0": [ubctl_event(before=before, after=after)], UNKNOWN_DEVICE_ID: self.parse_snapshoot_events()},
                    self.base_entry(base_code),
                )
                self.assertIn(ead_code, result)
                self.assertNotIn(old_code, result)
                self.assertNotIn(base_code, result)


class TestEadSchemaEntities(unittest.TestCase):
    """两个 EAD 细分码必须在默认 kg-config.json 中有实体，否则出码会 AttributeError"""

    def test_ead_codes_exist_in_default_config(self):
        schema = Schema([KNOWLEDGE_GRAPH_CONF])
        for code in (Comp_Custom_UBMEM_TIMEOUT_LOST_PKG_EAD, Comp_Custom_CQE0x5_LOST_PKG_EAD):
            entity = schema.get_schema_entity(code)
            self.assertIsNotNone(entity, code)
            attribute = entity.attribute.to_json()
            self.assertTrue(attribute.get("cause_zh"), code)
            self.assertTrue(attribute.get("description_zh"), code)
            self.assertTrue(attribute.get("suggestion_zh"), code)
