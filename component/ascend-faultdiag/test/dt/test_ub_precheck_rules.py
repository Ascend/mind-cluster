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
from ascend_fd.utils.constant.ub_const import (
    AGE_PERIOD_KEYWORD,
    DAM_INTF_ALARM,
    DFX_TM_CRD_CTRL,
    HCOMM_TA_CTP_UB_TIMEOUT,
    ICRC_ERR_COUNT_METRICS,
    LQC_TAI_DFX_ALARM,
    PHY_REINIT_CNT,
    PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT,
    PRECHECK_RXDMA_ICRC_DATA,
    PRECHECK_UBCTL_DATA,
    PRECHECK_UBMEM_TIMEOUT,
    RC_FULL_QUEUE_CNT,
    RULE_DFX_TM_CRD_CTRL_INVALID,
    RULE_HCOMM_TA_CTP_UB_TIMEOUT,
    RULE_LQC_TAI_DFX_ALARM_BIT45,
    RULE_PHY_REINIT_CNT_EXCEED,
    RULE_RC_FULL_QUEUE_NONZERO,
    RULE_ROUTE_NO_CFG_BIT,
    RULE_RXDMA_ICRC_ERR_INCREASE,
    RULE_TAACK_ABNORM_HEADER_INCREASE,
    RULE_TAACK_ABNORM_SSN_INCREASE,
    RULE_TAI_COMPACT_TOP_BIT6,
    RULE_TP_RRP_ERR_BIT25,
    RULE_TP_RRP_ERR_BIT27_28,
    RULE_TP_RRP_ERR_BIT28,
    RULE_TWP_AE_DFX_BIT3,
    RULE_UBMEM_TIMEOUT_LOW,
    RULE_VLAN10_11_BALANCE,
    RULE_VLAN6_7_BALANCE,
    RX_VL10_PKT_NUM,
    RX_VL11_PKT_NUM,
    RX_VL6_PKT_NUM,
    RX_VL7_PKT_NUM,
    TAACK_ABNORM_HEADER,
    TAACK_ABNORM_SSN,
    TAI_COMPACT_ALARM,
    TAI_COMPACT_TOP_ALARM,
    TP_RRP_ERR_FLG_0,
    TWP_AE_DFX,
    TX_VL10_PKT_NUM,
    TX_VL11_PKT_NUM,
    TX_VL6_PKT_NUM,
    TX_VL7_PKT_NUM,
)


class PrecheckRuleTestBase(unittest.TestCase):
    """规则层公共构造：不依赖 schema 的干净实例"""

    def new_checker(self, single_device_precheck_event=None):
        checker = MergePrecheckCause(schema=None, precheck_info={})
        checker.single_device_precheck_event = single_device_precheck_event or {}
        return checker

    def assert_rule_flag(self, checker, rule_key, expect_flagged):
        if expect_flagged:
            self.assertEqual(checker.rule_flags[rule_key]["value"], True)
        else:
            self.assertNotIn(rule_key, checker.rule_flags)


class TestMergePrecheckHelper(PrecheckRuleTestBase):
    def test_to_int(self):
        self.assertEqual(MergePrecheckCause._to_int("0x10"), 16)
        self.assertEqual(MergePrecheckCause._to_int("16"), 16)
        # 非字符串（int/list）类型不支持带 base 转换，按约定返回 None
        for value in (16, None, "", "abc", [1, 2]):
            self.assertIsNone(MergePrecheckCause._to_int(value))

    def test_bit_is_set(self):
        self.assertTrue(MergePrecheckCause._bit_is_set(0x8, 3))
        self.assertFalse(MergePrecheckCause._bit_is_set(0x8, 2))
        self.assertTrue(MergePrecheckCause._bit_is_set(0x10000000, 28))

    def test_bits_are_set(self):
        self.assertTrue(MergePrecheckCause._bits_are_set(0x18000000, 27, 28))
        self.assertFalse(MergePrecheckCause._bits_are_set(0x10000000, 27, 28))
        self.assertFalse(MergePrecheckCause._bits_are_set(0x8000000, 27, 28))


class TestRule1RcFullQueueNonzero(PrecheckRuleTestBase):
    def test_nonzero_dict_marks_flag(self):
        checker = self.new_checker()
        checker._check_rc_full_queue_nonzero({RC_FULL_QUEUE_CNT: "0x1"})
        self.assertTrue(checker.rule_flags[RULE_RC_FULL_QUEUE_NONZERO]["value"])

    def test_decimal_string_supported(self):
        checker = self.new_checker()
        checker._check_rc_full_queue_nonzero({RC_FULL_QUEUE_CNT: "2"})
        self.assertEqual(checker.rule_flags[RULE_RC_FULL_QUEUE_NONZERO]["value"], True)

    def test_zero_not_flagged(self):
        checker = self.new_checker()
        checker._check_rc_full_queue_nonzero({RC_FULL_QUEUE_CNT: "0x0"})
        self.assertNotIn(RULE_RC_FULL_QUEUE_NONZERO, checker.rule_flags)

    def test_missing_or_invalid_key_not_flagged(self):
        checker = self.new_checker()
        checker._check_rc_full_queue_nonzero({})
        checker._check_rc_full_queue_nonzero({RC_FULL_QUEUE_CNT: "abc"})
        self.assertNotIn(RULE_RC_FULL_QUEUE_NONZERO, checker.rule_flags)

    def test_already_flagged_skips(self):
        checker = self.new_checker()
        checker.rule_flags[RULE_RC_FULL_QUEUE_NONZERO] = {"value": True, "line": "first"}
        checker._check_rc_full_queue_nonzero({RC_FULL_QUEUE_CNT: "0x2"})
        # 已置位不再重复触发，flag 保持有效
        self.assertEqual(checker.rule_flags[RULE_RC_FULL_QUEUE_NONZERO]["value"], True)


class TestRule2TpRrpErrBit28(PrecheckRuleTestBase):
    def test_bit28_set(self):
        for value in ("0x10000000", "0x18000000"):
            checker = self.new_checker()
            checker._check_tp_rrp_err_bit28({TP_RRP_ERR_FLG_0: value})
            self.assertEqual(checker.rule_flags[RULE_TP_RRP_ERR_BIT28]["value"], True)
        # 未置位/非法值
        for value in ("0x0", "0x8000000", "0x1", "zzz"):
            checker = self.new_checker()
            checker._check_tp_rrp_err_bit28({TP_RRP_ERR_FLG_0: value})
            self.assertNotIn(RULE_TP_RRP_ERR_BIT28, checker.rule_flags)


class TestRule4RouteNoCfgBit(PrecheckRuleTestBase):
    def test_any_alarm_bit_marks_flag(self):
        # A(tai bit2)/B(lqc bit45)/C(dam bit01) 任一满足即置位，合并为一个用例覆盖全部触发路径
        cases = (
            {TAI_COMPACT_ALARM: "0x4", LQC_TAI_DFX_ALARM: "0x0", DAM_INTF_ALARM: "0x0"},
            {TAI_COMPACT_ALARM: "0x0", LQC_TAI_DFX_ALARM: "0x30", DAM_INTF_ALARM: "0x0"},
            {TAI_COMPACT_ALARM: "0x0", LQC_TAI_DFX_ALARM: "0x0", DAM_INTF_ALARM: "0x3"},
            {TAI_COMPACT_ALARM: "0x4", LQC_TAI_DFX_ALARM: "0x30", DAM_INTF_ALARM: "0x3"},
        )
        for alarm_index, alarm in enumerate(cases):
            with self.subTest(alarm_index=alarm_index):
                checker = self.new_checker()
                checker._check_route_no_cfg_bit(alarm)
                checker._post_process_route_no_cfg_bit()
                self.assertTrue(checker.rule_flags[RULE_ROUTE_NO_CFG_BIT]["value"])

    def test_none_flagged(self):
        checker = self.new_checker()
        checker._check_route_no_cfg_bit({TAI_COMPACT_ALARM: "0x0", LQC_TAI_DFX_ALARM: "0x0", DAM_INTF_ALARM: "0x0"})
        checker._post_process_route_no_cfg_bit()
        self.assertNotIn(RULE_ROUTE_NO_CFG_BIT, checker.rule_flags)

    def test_missing_all_alarm_keys(self):
        checker = self.new_checker()
        checker._check_route_no_cfg_bit({})
        checker._post_process_route_no_cfg_bit()
        self.assertNotIn(RULE_ROUTE_NO_CFG_BIT, checker.rule_flags)


class TestRule6PhyReinitCntExceed(PrecheckRuleTestBase):
    def test_diff_greater_than_5(self):
        checker = self.new_checker()
        checker._check_phy_reinit_cnt_exceed({PHY_REINIT_CNT: "0x6"}, {PHY_REINIT_CNT: "0x0"})
        self.assertEqual(checker.rule_flags[RULE_PHY_REINIT_CNT_EXCEED]["value"], True)

    def test_diff_not_greater_than_5_not_flagged(self):
        for after, before in (("0x5", "0x0"), ("0x1", "0x6")):
            checker = self.new_checker()
            checker._check_phy_reinit_cnt_exceed({PHY_REINIT_CNT: after}, {PHY_REINIT_CNT: before})
            self.assertNotIn(RULE_PHY_REINIT_CNT_EXCEED, checker.rule_flags)

    def test_before_missing_or_empty_not_flagged(self):
        # before 侧缺失（None）或空 dict 均不应触发规则
        checker = self.new_checker()
        checker._check_phy_reinit_cnt_exceed({PHY_REINIT_CNT: "0x6"}, None)
        checker._check_phy_reinit_cnt_exceed({PHY_REINIT_CNT: "0x6"}, {})
        self.assertNotIn(RULE_PHY_REINIT_CNT_EXCEED, checker.rule_flags)


class TestSingleBitAndValueRules(PrecheckRuleTestBase):
    """单 bit/单值判定规则，数据驱动覆盖置位与未置位两个分支"""

    def test_twp_ae_dfx_bit3(self):
        for value, expect in (("0x8", True), ("0x7", False)):
            checker = self.new_checker()
            checker._check_twp_ae_dfx_bit3({TWP_AE_DFX: value})
            self.assert_rule_flag(checker, RULE_TWP_AE_DFX_BIT3, expect)

    def test_tai_compact_top_bit6(self):
        for value, expect in (("0x40", True), ("0x0", False)):
            checker = self.new_checker()
            checker._check_tai_compact_top_bit6({TAI_COMPACT_TOP_ALARM: value})
            self.assert_rule_flag(checker, RULE_TAI_COMPACT_TOP_BIT6, expect)

    def test_taack_abnorm_ssn_increase(self):
        for value, expect in (("0x1", True), ("0x0", False)):
            checker = self.new_checker()
            checker._check_taack_abnorm_ssn_increase({TAACK_ABNORM_SSN: value}, {TAACK_ABNORM_SSN: "0x0"})
            self.assert_rule_flag(checker, RULE_TAACK_ABNORM_SSN_INCREASE, expect)

    def test_taack_abnorm_header_increase(self):
        for value, expect in (("0x2", True), ("0x0", False)):
            checker = self.new_checker()
            checker._check_taack_abnorm_header_increase({TAACK_ABNORM_HEADER: value}, {TAACK_ABNORM_HEADER: "0x0"})
            self.assert_rule_flag(checker, RULE_TAACK_ABNORM_HEADER_INCREASE, expect)

    def test_tp_rrp_err_bit25(self):
        for value, expect in (("0x2000000", True), ("0x1000000", False)):
            checker = self.new_checker()
            checker._check_tp_rrp_err_bit25({TP_RRP_ERR_FLG_0: value})
            self.assert_rule_flag(checker, RULE_TP_RRP_ERR_BIT25, expect)

    def test_tp_rrp_err_bit27_28(self):
        for value, expect in (("0x18000000", True), ("0x10000000", False)):
            checker = self.new_checker()
            checker._check_tp_rrp_err_bit27_28({TP_RRP_ERR_FLG_0: value})
            self.assert_rule_flag(checker, RULE_TP_RRP_ERR_BIT27_28, expect)

    def test_dfx_tm_crd_ctrl_invalid(self):
        for value, expect in (("0x0", True), ("0x1", False)):
            checker = self.new_checker()
            checker._check_dfx_tm_crd_ctrl_invalid({DFX_TM_CRD_CTRL: value})
            self.assert_rule_flag(checker, RULE_DFX_TM_CRD_CTRL_INVALID, expect)

    def test_lqc_tai_dfx_alarm_bit45(self):
        for value, expect in (("0x30", True), ("0x10", False)):
            checker = self.new_checker()
            checker._check_lqc_tai_dfx_alarm_bit45({LQC_TAI_DFX_ALARM: value})
            self.assert_rule_flag(checker, RULE_LQC_TAI_DFX_ALARM_BIT45, expect)


class TestVlanBalanceRules(PrecheckRuleTestBase):
    def _accumulate_and_post(self, before_tree, after_tree):
        checker = self.new_checker()
        for udie in ("udie0", "udie1"):
            for port in (1, 2):
                before_port = before_tree.get(udie, {}).get(port)
                after_port = after_tree.get(udie, {}).get(port)
                checker._accumulate_vlan_balance(udie, after_port or {}, before_port or {})
        checker._post_process_vlan_balance()
        return checker

    def test_balanced_v10v11_not_flagged(self):
        after_tree = {
            "udie0": {
                1: {RX_VL10_PKT_NUM: "0x2", TX_VL10_PKT_NUM: "0x2", RX_VL11_PKT_NUM: "0x1", TX_VL11_PKT_NUM: "0x1"},
                2: {RX_VL10_PKT_NUM: "0x3", TX_VL10_PKT_NUM: "0x3", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"},
            },
            "udie1": {},
        }
        before_tree = {
            "udie0": {
                1: {RX_VL10_PKT_NUM: "0x0", TX_VL10_PKT_NUM: "0x0", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"},
                2: {RX_VL10_PKT_NUM: "0x0", TX_VL10_PKT_NUM: "0x0", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"},
            },
            "udie1": {},
        }
        checker = self._accumulate_and_post(before_tree, after_tree)
        self.assertNotIn(RULE_VLAN10_11_BALANCE, checker.rule_flags)
        self.assertNotIn(RULE_VLAN6_7_BALANCE, checker.rule_flags)

    def test_unbalanced_v10_udie0_flagged(self):
        after_tree = {
            "udie0": {
                1: {RX_VL10_PKT_NUM: "0x2", TX_VL10_PKT_NUM: "0x1", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"}
            },
            "udie1": {},
        }
        before_tree = {
            "udie0": {
                1: {RX_VL10_PKT_NUM: "0x0", TX_VL10_PKT_NUM: "0x0", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"}
            },
            "udie1": {},
        }
        checker = self._accumulate_and_post(before_tree, after_tree)
        self.assertTrue(checker.rule_flags[RULE_VLAN10_11_BALANCE]["value"])

    def test_unbalanced_udie1_side_detected(self):
        after_tree = {
            "udie0": {
                1: {RX_VL10_PKT_NUM: "0x1", TX_VL10_PKT_NUM: "0x1", RX_VL11_PKT_NUM: "0x1", TX_VL11_PKT_NUM: "0x1"}
            },
            "udie1": {
                1: {RX_VL10_PKT_NUM: "0x5", TX_VL10_PKT_NUM: "0x5", RX_VL11_PKT_NUM: "0x1", TX_VL11_PKT_NUM: "0x0"}
            },
        }
        before_tree = {
            "udie0": {
                1: {RX_VL10_PKT_NUM: "0x0", TX_VL10_PKT_NUM: "0x0", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"}
            },
            "udie1": {
                1: {RX_VL10_PKT_NUM: "0x0", TX_VL10_PKT_NUM: "0x0", RX_VL11_PKT_NUM: "0x0", TX_VL11_PKT_NUM: "0x0"}
            },
        }
        checker = self._accumulate_and_post(before_tree, after_tree)
        self.assertTrue(checker.rule_flags[RULE_VLAN10_11_BALANCE]["value"])

    def test_v6v7_unbalanced_flagged(self):
        after_tree = {
            "udie0": {1: {RX_VL6_PKT_NUM: "0x4", TX_VL6_PKT_NUM: "0x4", RX_VL7_PKT_NUM: "0x3", TX_VL7_PKT_NUM: "0x2"}},
            "udie1": {},
        }
        before_tree = {
            "udie0": {1: {RX_VL6_PKT_NUM: "0x0", TX_VL6_PKT_NUM: "0x0", RX_VL7_PKT_NUM: "0x0", TX_VL7_PKT_NUM: "0x0"}},
            "udie1": {},
        }
        checker = self._accumulate_and_post(before_tree, after_tree)
        self.assertTrue(checker.rule_flags[RULE_VLAN6_7_BALANCE]["value"])

    def test_partial_metrics_ignored(self):
        # 仅 after 侧有值、before 缺失时，不参与累加（不误报）
        after_tree = {"udie0": {1: {RX_VL10_PKT_NUM: "0x9", TX_VL10_PKT_NUM: "0x9"}}, "udie1": {}}
        checker = self._accumulate_and_post({"udie0": {}, "udie1": {}}, after_tree)
        self.assertNotIn(RULE_VLAN10_11_BALANCE, checker.rule_flags)


class TestRule15HcommTaCtpUbTimeout(PrecheckRuleTestBase):
    def test_value_below_one_level_flagged(self):
        event = {PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT: {"attribute": {HCOMM_TA_CTP_UB_TIMEOUT: "4"}}}
        checker = self.new_checker(event)
        checker._prepare_unknown_device_rule()
        self.assertEqual(checker.rule_flags[RULE_HCOMM_TA_CTP_UB_TIMEOUT]["value"], True)

    def test_value_level_exact_one_not_flagged(self):
        event = {PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT: {"attribute": {HCOMM_TA_CTP_UB_TIMEOUT: "8"}}}
        checker = self.new_checker(event)
        checker._prepare_unknown_device_rule()
        self.assertNotIn(RULE_HCOMM_TA_CTP_UB_TIMEOUT, checker.rule_flags)

    def test_no_event_not_flagged(self):
        checker = self.new_checker()
        checker._prepare_unknown_device_rule()
        self.assertNotIn(RULE_HCOMM_TA_CTP_UB_TIMEOUT, checker.rule_flags)

    def test_invalid_value_not_flagged(self):
        event = {PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT: {"attribute": {HCOMM_TA_CTP_UB_TIMEOUT: "abc"}}}
        checker = self.new_checker(event)
        checker._prepare_unknown_device_rule()
        self.assertNotIn(RULE_HCOMM_TA_CTP_UB_TIMEOUT, checker.rule_flags)


class TestRule16UbmemTimeoutLow(PrecheckRuleTestBase):
    def test_age_period_below_4s_flagged(self):
        event = {PRECHECK_UBMEM_TIMEOUT: {"attribute": {AGE_PERIOD_KEYWORD: "3000000"}}}
        checker = self.new_checker(event)
        checker._check_ubmem_timeout_low()
        self.assertEqual(checker.rule_flags[RULE_UBMEM_TIMEOUT_LOW]["value"], True)

    def test_age_period_not_below_4s_not_flagged(self):
        for value in ("5000000", "abc"):
            event = {PRECHECK_UBMEM_TIMEOUT: {"attribute": {AGE_PERIOD_KEYWORD: value}}}
            checker = self.new_checker(event)
            checker._check_ubmem_timeout_low()
            self.assertNotIn(RULE_UBMEM_TIMEOUT_LOW, checker.rule_flags)

    def test_no_event_or_missing_attribute_not_flagged(self):
        checker = self.new_checker()
        checker._check_ubmem_timeout_low()
        self.assertNotIn(RULE_UBMEM_TIMEOUT_LOW, checker.rule_flags)

        checker = self.new_checker({PRECHECK_UBMEM_TIMEOUT: {"attribute": {}}})
        checker._check_ubmem_timeout_low()
        self.assertNotIn(RULE_UBMEM_TIMEOUT_LOW, checker.rule_flags)


class TestPrepareIcrcRule(PrecheckRuleTestBase):
    def _icrc_event(self, before, after):
        return {
            PRECHECK_RXDMA_ICRC_DATA: {
                "source_device": "0",
                "attribute": {"before": before, "after": after},
            }
        }

    def test_after_greater_flags(self):
        # 指标增长即触发：常规增长与从 0 变为正值均属于同一判定，合并覆盖
        before = {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: 5}}}
        after = {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: 10}}}
        checker = self.new_checker(self._icrc_event(before, after))
        checker._prepare_icrc_rule()
        self.assertTrue(checker.rule_flags[RULE_RXDMA_ICRC_ERR_INCREASE]["value"])

    def test_after_not_greater_not_flagged(self):
        for before, after in ((10, 10), (10, 3)):
            checker = self.new_checker(
                self._icrc_event(
                    {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: before}}}, {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: after}}}
                )
            )
            checker._prepare_icrc_rule()
            self.assertNotIn(RULE_RXDMA_ICRC_ERR_INCREASE, checker.rule_flags)

    def test_before_data_incomplete_skipped_without_flag(self):
        # before 侧缺端口数据或缺指标 key 均跳过，合并覆盖
        missing_port = self._icrc_event({}, {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: 10}}})
        checker = self.new_checker(missing_port)
        checker._prepare_icrc_rule()
        self.assertNotIn(RULE_RXDMA_ICRC_ERR_INCREASE, checker.rule_flags)

        missing_metric = self._icrc_event({"0": {"0": {}}}, {"0": {"0": {ICRC_ERR_COUNT_METRICS[0]: 10}}})
        checker = self.new_checker(missing_metric)
        checker._prepare_icrc_rule()
        self.assertNotIn(RULE_RXDMA_ICRC_ERR_INCREASE, checker.rule_flags)

    def test_no_event_not_flagged(self):
        checker = self.new_checker()
        checker._prepare_icrc_rule()
        self.assertNotIn(RULE_RXDMA_ICRC_ERR_INCREASE, checker.rule_flags)


class TestPrepareUbctlLogRuleIntegration(PrecheckRuleTestBase):
    def build_ubctl_event(self, before_tree, after_tree):
        return {
            PRECHECK_UBCTL_DATA: {
                "attribute": {
                    "before": before_tree,
                    "after": after_tree,
                    "source_lines": {"before": [], "after": []},
                }
            }
        }

    def test_single_port_all_rules_trigger(self):
        after_port = {
            RC_FULL_QUEUE_CNT: "0x1",
            TP_RRP_ERR_FLG_0: "0x10000000",
            TWP_AE_DFX: "0x8",
            TAI_COMPACT_ALARM: "0x4",
            LQC_TAI_DFX_ALARM: "0x30",
            DAM_INTF_ALARM: "0x3",
            TAI_COMPACT_TOP_ALARM: "0x40",
            DFX_TM_CRD_CTRL: "0x2",
            PHY_REINIT_CNT: "0x6",
            TAACK_ABNORM_SSN: "0x1",
            TAACK_ABNORM_HEADER: "0x1",
        }
        before_port = {
            RC_FULL_QUEUE_CNT: "0x0",
            TP_RRP_ERR_FLG_0: "0x0",
            TWP_AE_DFX: "0x0",
            TAI_COMPACT_ALARM: "0x0",
            LQC_TAI_DFX_ALARM: "0x0",
            DAM_INTF_ALARM: "0x0",
            TAI_COMPACT_TOP_ALARM: "0x0",
            DFX_TM_CRD_CTRL: "0x1",
            PHY_REINIT_CNT: "0x0",
            TAACK_ABNORM_SSN: "0x0",
            TAACK_ABNORM_HEADER: "0x0",
        }
        after_tree = {"udie0": {3: after_port}, "udie1": {}}
        before_tree = {"udie0": {3: before_port}, "udie1": {}}
        checker = self.new_checker(self.build_ubctl_event(before_tree, after_tree))
        checker._prepare_ubctl_log_rule()

        for rule_key in (
            RULE_RC_FULL_QUEUE_NONZERO,
            RULE_TP_RRP_ERR_BIT28,
            RULE_TWP_AE_DFX_BIT3,
            RULE_ROUTE_NO_CFG_BIT,
            RULE_TAI_COMPACT_TOP_BIT6,
            RULE_PHY_REINIT_CNT_EXCEED,
            RULE_TAACK_ABNORM_SSN_INCREASE,
            RULE_TAACK_ABNORM_HEADER_INCREASE,
            RULE_DFX_TM_CRD_CTRL_INVALID,
            RULE_LQC_TAI_DFX_ALARM_BIT45,
        ):
            self.assertEqual(checker.rule_flags[rule_key]["value"], True, rule_key)

        # 位判断规则应保持未触发
        self.assertNotIn(RULE_TP_RRP_ERR_BIT25, checker.rule_flags)
        self.assertNotIn(RULE_TP_RRP_ERR_BIT27_28, checker.rule_flags)

    def test_rules_triggered_from_any_port(self):
        after_tree = {
            "udie0": {1: {}, 2: {}, 3: {RC_FULL_QUEUE_CNT: "0x1"}},
            "udie1": {5: {TP_RRP_ERR_FLG_0: "0x8000000"}},
        }
        before_tree = {"udie0": {}, "udie1": {}}
        checker = self.new_checker(self.build_ubctl_event(before_tree, after_tree))
        checker._prepare_ubctl_log_rule()
        self.assertEqual(checker.rule_flags[RULE_RC_FULL_QUEUE_NONZERO]["value"], True)

    def test_empty_trees_no_flags(self):
        checker = self.new_checker(self.build_ubctl_event({"udie0": {}, "udie1": {}}, {"udie0": {}, "udie1": {}}))
        checker._prepare_ubctl_log_rule()
        self.assertEqual(checker.rule_flags, {})
