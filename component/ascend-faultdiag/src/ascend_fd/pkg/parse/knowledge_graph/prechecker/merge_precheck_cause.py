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
from typing import Any, Dict, List, Optional

from ascend_fd.pkg.parse.knowledge_graph.prechecker.ub_cqe_checker import (
    Cqe0x2Checker,
    Cqe0x3Checker,
    Cqe0x5Checker,
    UBMemChecker,
)
from ascend_fd.pkg.parse.knowledge_graph.tools.utils import get_device_precheck_event
from ascend_fd.utils.constant.str_const import UNKNOWN_DEVICE_ID
from ascend_fd.utils.constant.ub_const import (
    AGE_PERIOD_KEYWORD,
    DAM_INTF_ALARM,
    DFX_TM_CRD_CTRL,
    HCOMM_TA_CTP_UB_TIMEOUT,
    LQC_TAI_DFX_ALARM,
    PHY_REINIT_CNT,
    PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT,
    PRECHECK_UBCTL_DATA,
    PRECHECK_UBMEM_TIMEOUT,
    RC_FULL_QUEUE_CNT,
    RULE_DFX_TM_CRD_CTRL_INVALID,
    RULE_HCOMM_TA_CTP_UB_TIMEOUT,
    RULE_LQC_TAI_DFX_ALARM_BIT45,
    RULE_PHY_REINIT_CNT_EXCEED,
    RULE_RC_FULL_QUEUE_NONZERO,
    RULE_ROUTE_NO_CFG_BIT,
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

reinit_max_cnt_persec = 5


class MergePrecheckCause:
    """
    对PRECHECK数据进行规则检查，返回问题详情和每条规则的触发状态。
    """

    def __init__(self, schema, precheck_info):
        self.udies = ['udie0', 'udie1']
        self.ports = list(range(9))
        # 使用描述性名称作为键
        self.rule_flags = {}

        # 规则4的组合标志
        self.rule4_A = {}
        self.rule4_B = {}
        self.rule4_C = {}

        # 规则7/8累加器（按 udie 组织：sums[udie]['rx'/'tx'][vlan]）
        self.sums = {
            udie: {
                'rx': {'v6': 0, 'v7': 0, 'v10': 0, 'v11': 0},
                'tx': {'v6': 0, 'v7': 0, 'v10': 0, 'v11': 0},
            }
            for udie in self.udies
        }
        self.schema = schema
        # 本节点所有设备的PRECHECK事件
        self.precheck_info = precheck_info
        # 本节点某个设备的PRECHECK事件
        self.single_device_precheck_event = {}
        self.unknown_device_event = {}

    # ---------- 辅助工具 ----------
    @staticmethod
    def _to_int(val: Any) -> Optional[int]:
        if val is None:
            return None
        try:
            return int(val, 0)
        except (ValueError, TypeError):
            return None

    @staticmethod
    def _bit_is_set(val: int, bit: int) -> bool:
        return ((val >> bit) & 1) == 1

    @staticmethod
    def _bits_are_set(val: int, bit1: int, bit2: int) -> bool:
        return MergePrecheckCause._bit_is_set(val, bit1) and MergePrecheckCause._bit_is_set(val, bit2)

    # ---------- 规则检查方法（均以 check_ 开头，名称描述检查内容） ----------

    # 规则1: rc_full_queue_cnt 非0
    # after_port: 示例 {'rc_full_queue_cnt': '0x0', 'tp_rrp_err_flg_0': '0x0', ... 其他指标数据}
    # before_port: 示例 {'rc_full_queue_cnt': '0x0', 'tp_rrp_err_flg_0': '0x0', ... 其他指标数据}
    def _check_rc_full_queue_nonzero(self, after_port: Dict):
        if self.rule_flags.get(RULE_RC_FULL_QUEUE_NONZERO, {}):
            return
        _val = after_port.get(RC_FULL_QUEUE_CNT)
        val = self._to_int(_val)
        if val is None:
            return
        if val == 0:
            return

        self.rule_flags[RULE_RC_FULL_QUEUE_NONZERO] = {"value": True, "line": RC_FULL_QUEUE_CNT + ": " + str(_val)}

    # 规则2: tp_rrp_err_flg_0 第28位为1
    def _check_tp_rrp_err_bit28(self, after_port: Dict):
        if self.rule_flags.get(RULE_TP_RRP_ERR_BIT28, {}):
            return
        _val = after_port.get(TP_RRP_ERR_FLG_0)
        val = self._to_int(_val)
        if val is None:
            return
        if not self._bit_is_set(val, 28):
            return

        self.rule_flags[RULE_TP_RRP_ERR_BIT28] = {"value": True, "line": TP_RRP_ERR_FLG_0 + ": " + str(_val)}

    # 规则3: twp_ae_dfx 第3位为1
    def _check_twp_ae_dfx_bit3(self, after_port: Dict):
        if self.rule_flags.get(RULE_TWP_AE_DFX_BIT3, {}):
            return
        _val = after_port.get(TWP_AE_DFX)
        val = self._to_int(_val)
        if val is None:
            return
        if not self._bit_is_set(val, 3):
            return
        self.rule_flags[RULE_TWP_AE_DFX_BIT3] = {"value": True, "line": TWP_AE_DFX + ": " + str(_val)}

    # 规则4: 组合告警（收集标志，后处理）
    def _check_route_no_cfg_bit(self, after_port: Dict):
        if self.rule_flags.get(RULE_ROUTE_NO_CFG_BIT, {}):
            return
        val = self._to_int(after_port.get(TAI_COMPACT_ALARM))
        if val is not None and self._bit_is_set(val, 2):
            self.rule4_A = {"value": True, "line": TAI_COMPACT_ALARM + ": " + str(after_port.get(TAI_COMPACT_ALARM))}

        val = self._to_int(after_port.get(LQC_TAI_DFX_ALARM))
        if val is not None and self._bits_are_set(val, 4, 5):
            self.rule4_B = {"value": True, "line": LQC_TAI_DFX_ALARM + ": " + str(after_port.get(LQC_TAI_DFX_ALARM))}

        val = self._to_int(after_port.get(DAM_INTF_ALARM))
        if val is not None and self._bits_are_set(val, 0, 1):
            self.rule4_C = {"value": True, "line": DAM_INTF_ALARM + ": " + str(after_port.get(DAM_INTF_ALARM))}

    # 规则5: tai_compact_top_alarm 第6位为1
    def _check_tai_compact_top_bit6(self, after_port: Dict):
        if self.rule_flags.get(RULE_TAI_COMPACT_TOP_BIT6, {}):
            return
        _val = after_port.get(TAI_COMPACT_TOP_ALARM)
        val = self._to_int(_val)
        if val is None:
            return
        if not self._bit_is_set(val, 6):
            return

        self.rule_flags[RULE_TAI_COMPACT_TOP_BIT6] = {"value": True, "line": TAI_COMPACT_TOP_ALARM + ": " + str(_val)}

    # 规则6: phy_reinit_cnt 增量 > 5
    def _check_phy_reinit_cnt_exceed(self, after_port: Dict, before_port: Dict):
        if self.rule_flags.get(RULE_PHY_REINIT_CNT_EXCEED, {}):
            return
        before_val = self._to_int(before_port.get(PHY_REINIT_CNT) if before_port else None)
        after_val = self._to_int(after_port.get(PHY_REINIT_CNT))
        if before_val is None or after_val is None:
            return

        diff = after_val - before_val

        if diff > reinit_max_cnt_persec:
            self.rule_flags[RULE_PHY_REINIT_CNT_EXCEED] = {
                "value": True,
                "line": f"{PHY_REINIT_CNT}: before={before_port.get(PHY_REINIT_CNT)}, after={after_port.get(PHY_REINIT_CNT)}",
            }

    # 规则9: taack_abnorm_ssn 增量 > 0
    def _check_taack_abnorm_ssn_increase(self, after_port: Dict, before_port: Dict):
        if self.rule_flags.get(RULE_TAACK_ABNORM_SSN_INCREASE, {}):
            return
        before_val = self._to_int(before_port.get(TAACK_ABNORM_SSN) if before_port else None)
        after_val = self._to_int(after_port.get(TAACK_ABNORM_SSN))
        if before_val is None or after_val is None:
            return

        diff = after_val - before_val
        if diff > 0:
            self.rule_flags[RULE_TAACK_ABNORM_SSN_INCREASE] = {
                "value": True,
                "line": f"{TAACK_ABNORM_SSN}: before={before_port.get(TAACK_ABNORM_SSN)}, after={after_port.get(TAACK_ABNORM_SSN)}",
            }

    # 规则10: taack_abnorm_header 增量 > 0
    def _check_taack_abnorm_header_increase(self, after_port: Dict, before_port: Dict):
        if self.rule_flags.get(RULE_TAACK_ABNORM_HEADER_INCREASE, {}):
            return
        before_val = self._to_int(before_port.get(TAACK_ABNORM_HEADER) if before_port else None)
        after_val = self._to_int(after_port.get(TAACK_ABNORM_HEADER))
        if before_val is None or after_val is None:
            return

        diff = after_val - before_val
        if diff > 0:
            self.rule_flags[RULE_TAACK_ABNORM_HEADER_INCREASE] = {
                "value": True,
                "line": f"{TAACK_ABNORM_HEADER}: before={before_port.get(TAACK_ABNORM_HEADER)}, after={after_port.get(TAACK_ABNORM_HEADER)}",
            }

    # 规则11: tp_rrp_err_flg_0 第25位为1
    def _check_tp_rrp_err_bit25(self, after_port: Dict):
        if self.rule_flags.get(RULE_TP_RRP_ERR_BIT25, {}):
            return
        _val = after_port.get(TP_RRP_ERR_FLG_0)
        val = self._to_int(_val)
        if val is None:
            return
        if not self._bit_is_set(val, 25):
            return
        self.rule_flags[RULE_TP_RRP_ERR_BIT25] = {"value": True, "line": TP_RRP_ERR_FLG_0 + ": " + str(_val)}

    # 规则12: tp_rrp_err_flg_0 第27和28位同时为1
    def _check_tp_rrp_err_bit27_28(self, after_port: Dict):
        if self.rule_flags.get(RULE_TP_RRP_ERR_BIT27_28, {}):
            return
        _val = after_port.get(TP_RRP_ERR_FLG_0)
        val = self._to_int(_val)
        if val is None:
            return
        if not self._bits_are_set(val, 27, 28):
            return
        self.rule_flags[RULE_TP_RRP_ERR_BIT27_28] = {"value": True, "line": TP_RRP_ERR_FLG_0 + ": " + str(_val)}

    # 规则13: dfx_tm_crd_ctrl != 1
    def _check_dfx_tm_crd_ctrl_invalid(self, after_port: Dict):
        if self.rule_flags.get(RULE_DFX_TM_CRD_CTRL_INVALID, {}):
            return
        _val = after_port.get(DFX_TM_CRD_CTRL)
        val = self._to_int(_val)
        if val is None:
            return
        if val == 1:
            return
        self.rule_flags[RULE_DFX_TM_CRD_CTRL_INVALID] = {"value": True, "line": DFX_TM_CRD_CTRL + ": " + str(_val)}

    # 规则14: lqc_tai_dfx_alarm 第4和5位同时为1
    def _check_lqc_tai_dfx_alarm_bit45(self, after_port: Dict):
        if self.rule_flags.get(RULE_LQC_TAI_DFX_ALARM_BIT45, {}):
            return
        _val = after_port.get(LQC_TAI_DFX_ALARM)
        val = self._to_int(_val)
        if val is None:
            return
        if not self._bits_are_set(val, 4, 5):
            return
        self.rule_flags[RULE_LQC_TAI_DFX_ALARM_BIT45] = {"value": True, "line": LQC_TAI_DFX_ALARM + ": " + str(_val)}

    # ---------- 规则7/8 累加 ----------
    def _accumulate_vlan_balance(self, udie: str, after_port: Dict, before_port: Dict):
        def get_pair(key_rx: str, key_tx: str):
            before_rx = self._to_int(before_port.get(key_rx) if before_port else None)
            after_rx = self._to_int(after_port.get(key_rx))
            before_tx = self._to_int(before_port.get(key_tx) if before_port else None)
            after_tx = self._to_int(after_port.get(key_tx))
            # 此处有校验所有值是否都不能为None
            if all(x is not None for x in (before_rx, after_rx, before_tx, after_tx)):
                return after_rx - before_rx, after_tx - before_tx
            return None, None

        for vlan, rx_key, tx_key in [
            ('v6', RX_VL6_PKT_NUM, TX_VL6_PKT_NUM),
            ('v7', RX_VL7_PKT_NUM, TX_VL7_PKT_NUM),
            ('v10', RX_VL10_PKT_NUM, TX_VL10_PKT_NUM),
            ('v11', RX_VL11_PKT_NUM, TX_VL11_PKT_NUM),
        ]:
            rx_delta, tx_delta = get_pair(rx_key, tx_key)
            if rx_delta is not None and tx_delta is not None:
                self.sums[udie]['rx'][vlan] += rx_delta
                self.sums[udie]['tx'][vlan] += tx_delta

    # ---------- 后处理 ----------
    def _post_process_route_no_cfg_bit(self):
        if self.rule_flags.get(RULE_ROUTE_NO_CFG_BIT, {}):
            return

        if self.rule4_A or self.rule4_B or self.rule4_C:
            details = []
            if self.rule4_A:
                details.append("tai_compact_alarm 值二进制第2位为1, " + self.rule4_A.get("line", ""))
            if self.rule4_B:
                details.append('lqc_tai_dfx_alarm 值二进制第4、5位同时为1, ' + self.rule4_B.get("line", ""))
            if self.rule4_C:
                details.append('dam_intf_alarm 值二进制第0、1位同时为1, ' + self.rule4_C.get("line", ""))
            self.rule_flags[RULE_ROUTE_NO_CFG_BIT] = {
                "value": True,
                "line": '; '.join(details),
            }

    def _post_process_vlan_balance(self):
        # 规则7: v10/v11 —— udie0 块与 udie1 块的 v10、v11 各自 RX 增量==TX 增量同时满足才算正常
        self._set_vlan_balance_flag(['v10', 'v11'], RULE_VLAN10_11_BALANCE)
        # 规则8: v6/v7 —— 同理
        self._set_vlan_balance_flag(['v6', 'v7'], RULE_VLAN6_7_BALANCE)

    def _set_vlan_balance_flag(self, vlans: List[str], rule_key: str):
        """
        判断指定 vlan 组是否丢包：每个 udie 的每个 vlan 都要 RX 增量==TX 增量，
        任一组合不平衡即视为丢包，并把所有不平衡项收集进 line。
        """
        unbalance_lines = []
        for vlan in vlans:
            for udie in self.udies:
                rx_sum = self.sums[udie]['rx'][vlan]
                tx_sum = self.sums[udie]['tx'][vlan]
                if rx_sum != tx_sum:
                    unbalance_lines.append(f'{vlan} {udie} RX增量总和({rx_sum}) != TX增量总和({tx_sum})')
        if unbalance_lines:
            self.rule_flags[rule_key] = {
                "value": True,
                "line": '; '.join(unbalance_lines),
            }

    def _check_hcomm_ta_ctp_ub_timeout(self):
        hcomm_timeout = self.unknown_device_event.get(
            PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT, {}
        ) or self.single_device_precheck_event.get(PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT, {})
        if not hcomm_timeout:
            return

        _val = hcomm_timeout.get("attribute", {}).get(HCOMM_TA_CTP_UB_TIMEOUT, "")
        val = self._to_int(_val)
        if val is None:
            return
        # 1档位4秒， 每8计数为1个档位阈值
        level = val / 8
        if level >= 1:
            return

        self.rule_flags[RULE_HCOMM_TA_CTP_UB_TIMEOUT] = {
            "value": True,
            "line": HCOMM_TA_CTP_UB_TIMEOUT + ": " + str(_val),
        }

    def _check_ubmem_timeout_low(self):
        if self.rule_flags.get(RULE_UBMEM_TIMEOUT_LOW, {}):
            return
        ubmem_timeout = self.single_device_precheck_event.get(PRECHECK_UBMEM_TIMEOUT, {})
        if not ubmem_timeout:
            return
        # age_period 单位是微秒，换算成秒后小于 4 秒才视为超时配置过低
        _val = ubmem_timeout.get("attribute", {}).get(AGE_PERIOD_KEYWORD)
        val = self._to_int(_val)
        if val is None:
            return
        sec = val / 1_000_000
        if sec >= 4:
            return

        self.rule_flags[RULE_UBMEM_TIMEOUT_LOW] = {
            "value": True,
            "line": f"{AGE_PERIOD_KEYWORD}: {_val}",
        }

    def _prepare_ubctl_log_rule(self):
        data = self.single_device_precheck_event.get(PRECHECK_UBCTL_DATA, {})
        ubctl_data_after = data.get("attribute", {}).get('after', {})
        ubctl_data_before = data.get("attribute", {}).get('before', {})

        # 仅依赖 after_port 的检查方法（单参数）
        after_only_methods = [
            self._check_rc_full_queue_nonzero,
            self._check_tp_rrp_err_bit28,
            self._check_twp_ae_dfx_bit3,
            self._check_route_no_cfg_bit,
            self._check_tai_compact_top_bit6,
            self._check_tp_rrp_err_bit25,
            self._check_tp_rrp_err_bit27_28,
            self._check_dfx_tm_crd_ctrl_invalid,
            self._check_lqc_tai_dfx_alarm_bit45,
        ]
        # 依赖 before/after 差值的方法（双参数）
        delta_methods = [
            self._check_phy_reinit_cnt_exceed,
            self._check_taack_abnorm_ssn_increase,
            self._check_taack_abnorm_header_increase,
        ]

        # 单次遍历所有端口
        for udie in self.udies:
            after_udie = ubctl_data_after.get(udie, {})
            before_udie = ubctl_data_before.get(udie, {})
            for port in self.ports:
                after_port = after_udie.get(port)
                if after_port is None:
                    continue
                before_port = before_udie.get(port)
                for method in after_only_methods:
                    method(after_port)
                for method in delta_methods:
                    method(after_port, before_port)
                self._accumulate_vlan_balance(udie, after_port, before_port)

        # 后处理跨端口规则
        self._post_process_route_no_cfg_bit()
        self._post_process_vlan_balance()

    def _prepare_unknown_device_rule(self):
        self._check_hcomm_ta_ctp_ub_timeout()

    def single_device_analyze(self, source_device, device_causes):
        """
        执行全部规则检查
        :param source_device: 设备ID
        :param device_causes: 设备故障原因{
                code: {
                    "code": code,
                    "entities_attribute": entities_attribute,
                    "events_attribute": events_attribute,
                    "chains": {source_device: link}
                }
            }
        """

        # 1. 所有PRECHECK规则的预计算，得到是否违反规则 ----------
        # ubctl_log.txt中数据的指标预检查
        self.single_device_precheck_event = get_device_precheck_event(self.precheck_info, source_device)
        if source_device != UNKNOWN_DEVICE_ID:
            device_causes.update(self.single_device_precheck_event)
        self._prepare_ubctl_log_rule()
        self._check_ubmem_timeout_low()

        self.unknown_device_event = get_device_precheck_event(self.precheck_info, UNKNOWN_DEVICE_ID)
        # unknown_device预检查
        self._prepare_unknown_device_rule()

        # 2.开始分析、合并PRECHECK和其他故障
        checker_list = [Cqe0x2Checker, Cqe0x3Checker, UBMemChecker, Cqe0x5Checker]
        for cls in checker_list:
            checker = cls(source_device, self)
            checker.analyze(device_causes)
