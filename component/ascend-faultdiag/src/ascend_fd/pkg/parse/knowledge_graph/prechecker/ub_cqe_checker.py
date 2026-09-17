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
from abc import ABC
import copy
from datetime import datetime

from ascend_fd.utils.constant.ub_const import (
    Comp_CANN_HCCL_Custom_CQE0x2,
    Comp_CANN_HCCL_Custom_CQE0x5,
    PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT,
    PRECHECK_KERNEL_AE_TYPE23,
    PRECHECK_UBCTL_DATA,
    PRECHECK_UBMEM_TIMEOUT,
    REMOTE_MEM_ACCESS_NET_TIMEOUT,
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
    UB_RAS_CODES,
)
from ascend_fd.utils.i18n import LANG, get_label_for_language

lb = get_label_for_language()


class Checker(ABC):
    base_code = None

    def __init__(self, source_device, merge_obj):
        # merge_obj: MergePrecheckCause 实例（provides schema / rule_flags / precheck 事件）
        super().__init__()
        self.source_device = source_device
        self.schema = merge_obj.schema
        # 使用描述性名称作为键
        self.rule_flags = merge_obj.rule_flags
        # 本节点某个设备的PRECHECK事件
        self.single_device_precheck_event = merge_obj.single_device_precheck_event
        self.unknown_device_event = merge_obj.unknown_device_event

    def analyze(self, device_causes):
        pass

    @staticmethod
    def _format_link_key(event_name, event_code):
        """
        Format link key, for example, '故障编号（故障名称）'.
        :param event_name: event cause
        :param event_code: event ID
        :return: event_code（event_name）
        """
        return event_code + lb.left_bracket + event_name + lb.right_bracket

    def _build_link(self, code, base_code, suffix=""):
        """拼链文本：细分码（细分cause）-> 基础码（基础cause）[suffix]。"""
        code_attribute = self.schema.get_schema_entity(code).attribute.to_json()
        base_attribute = self.schema.get_schema_entity(base_code).attribute.to_json()
        link = (
            self._format_link_key(code_attribute.get(f"cause_{LANG}", ""), code)
            + "-> "
            + self._format_link_key(base_attribute.get(f"cause_{LANG}", ""), base_code)
        )
        return link + suffix

    def _build_cqe_cause(self, device_causes, code, rule=None, event_key=PRECHECK_UBCTL_DATA):
        """构建并写入细分故障码的公共逻辑（基础码取本类属性 self.base_code）。

        :param device_causes: 输出的故障原因字典
        :param code: 细分故障码（Comp_Custom_*）
        :param rule: 命中规则的 rule_flags 条目（含 "line"）；为 None 时不写 key_info/occurrence
        :param event_key: 预检事件键，默认 PRECHECK_UBCTL_DATA
        """
        # 先拿 code 的实体，再从本类基础码 self.base_code 取 unknown 事件
        base_code = self.base_code
        code_entity = self.schema.get_schema_entity(code)
        entities_attribute = code_entity.attribute.to_json()  # {'component':..., 'cause_zh':..., ...}
        link = self._build_link(code, base_code)
        base_event = self.unknown_device_event.get(base_code, {}) or device_causes.get(base_code, {})
        event_attribute = copy.deepcopy(base_event)

        tmp = self.single_device_precheck_event.get(event_key, {})
        precheck_event = copy.deepcopy(tmp)
        if rule:
            precheck_event["key_info"] = rule.get("line", "")
            precheck_event["occurrence"] = [
                [tmp.get("occur_time", datetime.min.strftime("%Y-%m-%d %H:%M:%S")), rule.get("line", "")]
            ]
        events_attribute = [precheck_event, event_attribute]

        device_causes.update(
            {
                code: {
                    "code": code,
                    "entities_attribute": entities_attribute,
                    "events_attribute": events_attribute,
                    "chains": {self.source_device: link},
                }
            }
        )


class Cqe0x2Checker(Checker):
    base_code = Comp_CANN_HCCL_Custom_CQE0x2

    def analyze(self, device_causes):
        fault_codes = list(
            self.unknown_device_event.keys() | [] if not isinstance(device_causes, dict) else device_causes.keys()
        )
        # 基础码由 plog 解析，unknown_device_event或已识别的故障列表中无基础码则直接返回
        if self.base_code not in fault_codes:
            return

        keep_cqe = [
            self.add_rc_not_enough_causes(device_causes),
            self.add_rc_queue_not_enough_causes(device_causes),
            self.add_tp_psn_error_causes(device_causes),
            self.add_no_router_causes(device_causes),
            self.add_ctp_no_close_causes(device_causes),
        ]
        if any(keep_cqe):
            # 有故障匹配成功，移除基础码，正常情况是没有的，做兜底
            device_causes.pop(self.base_code, None)

    def add_rc_not_enough_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x2_RC_NOT_ENOUGH"
        if code in device_causes:
            return True
        rule = self.rule_flags.get(RULE_RC_FULL_QUEUE_NONZERO, {})
        if not rule.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule)
        return True

    def add_rc_queue_not_enough_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x2_RC_QUEUE_NOT_ENOUGH"
        if code in device_causes:
            return True
        rule = self.rule_flags.get(RULE_TP_RRP_ERR_BIT28, {})
        if not rule.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule)
        return True

    def add_tp_psn_error_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x2_TP_PSN_ERROR"
        if code in device_causes:
            return True
        rule = self.rule_flags.get(RULE_TWP_AE_DFX_BIT3, {})
        ae_event = self.single_device_precheck_event.get(PRECHECK_KERNEL_AE_TYPE23, {})
        if not rule.get("value", False) or not ae_event:
            return False
        self._build_cqe_cause(device_causes, code, rule)
        return True

    def add_no_router_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x2_NO_ROUTER"
        if code in device_causes:
            return True
        rule = self.rule_flags.get(RULE_ROUTE_NO_CFG_BIT, {})
        if not rule.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule)
        return True

    def add_ctp_no_close_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x2_CTP_NO_CLOSE"
        if code in device_causes:
            return True
        rule = self.rule_flags.get(RULE_TAI_COMPACT_TOP_BIT6, {})
        if not rule.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule)
        return True


class Cqe0x3Checker(Checker):
    def analyze(self, device_causes):
        # 基础码由 plog 解析，unknown_device_event或已识别的故障列表中无基础码则直接返回，此处为后续预留
        return


class UBMemChecker(Checker):
    base_code = REMOTE_MEM_ACCESS_NET_TIMEOUT

    def _build_ubmem_cause(self, device_causes, code, rule=None, event_key=PRECHECK_UBCTL_DATA, suffix=""):
        """UBMem 细分码构建公共逻辑（基础码取本类属性 self.base_code）。

        与 CQE 的差异：基础码事件从 single_device_precheck_event 取（UBMem 基础码事件归属当前设备，而非 Unknown）；
        suffix 非空时追加到链尾部（如 "+lost package(命中行)"）。
        """
        entities_attribute = self.schema.get_schema_entity(code).attribute.to_json()
        link = self._build_link(code, self.base_code, suffix)
        base_event = self.single_device_precheck_event.get(self.base_code, {})
        event_attribute = copy.deepcopy(base_event)

        tmp = self.single_device_precheck_event.get(event_key, {})
        precheck_event = copy.deepcopy(tmp)
        if rule:
            precheck_event["key_info"] = rule.get("line", "")
            precheck_event["occurrence"] = [
                [tmp.get("occur_time", datetime.min.strftime("%Y-%m-%d %H:%M:%S")), rule.get("line", "")]
            ]
        events_attribute = [precheck_event, event_attribute]

        device_causes.update(
            {
                code: {
                    "code": code,
                    "entities_attribute": entities_attribute,
                    "events_attribute": events_attribute,
                    "chains": {self.source_device: link},
                }
            }
        )

    def analyze(self, device_causes):
        self.add_ubmem_ub_ras_causes(device_causes)
        self.add_ubmem_timeout_low_causes(device_causes)
        self.add_ubmem_timeout_retraining_causes(device_causes)
        self.add_ubmem_timeout_lost_pkg_causes(device_causes)

    def add_ubmem_timeout_low_causes(self, device_causes):
        code = "Comp_Custom_UBMEM_TIMEOUT_LOW"
        if code in device_causes:
            return

        if self.base_code not in device_causes:
            return

        rule = self.rule_flags.get(RULE_VLAN10_11_BALANCE, {})
        # 有丢包
        if rule.get("value", False):
            return

        rule_timeout_low = self.rule_flags.get(RULE_UBMEM_TIMEOUT_LOW, {})
        # 规则层已按 age_period（微秒换算秒）小于 4 秒判定并置位
        if not rule_timeout_low.get("value", False):
            return

        self._build_ubmem_cause(device_causes, code, event_key=PRECHECK_UBMEM_TIMEOUT)

    def add_ubmem_timeout_retraining_causes(self, device_causes):
        code = "Comp_Custom_UBMEM_TIMEOUT_RETRAINING"
        if code in device_causes:
            return

        if self.base_code not in device_causes:
            return

        rule_lost_pkt = self.rule_flags.get(RULE_VLAN10_11_BALANCE, {})
        # 有丢包
        if rule_lost_pkt.get("value", False):
            return

        rule_reinit = self.rule_flags.get(RULE_PHY_REINIT_CNT_EXCEED, {})
        # 1秒内没有多次retraining
        if not rule_reinit.get("value", False):
            return

        self._build_ubmem_cause(device_causes, code, rule_reinit)

    def add_ubmem_ub_ras_causes(self, device_causes):
        if self.base_code not in device_causes:
            return

        # 没UB_RAS故障直接返回
        ras_code = list(set(UB_RAS_CODES) & device_causes.keys())
        if not ras_code:
            return

        rule_lost_pkt = self.rule_flags.get(RULE_VLAN10_11_BALANCE, {})
        # 没有丢包
        if not rule_lost_pkt.get("value", False):
            return

        for code in ras_code:
            # 一定存在cause
            cause = device_causes.get(code)
            link = self._build_link(code, self.base_code, "+lost package({})".format(rule_lost_pkt.get("line", "")))
            cause.get("chains").update({self.source_device: link})

    def add_ubmem_timeout_lost_pkg_causes(self, device_causes):
        code = "Comp_Custom_UBMEM_TIMEOUT_LOST_PKG"
        if code in device_causes:
            return

        if self.base_code not in device_causes:
            return

        # 有UB_RAS故障直接返回,isdisjoint表示没有交集返回True
        if not set(UB_RAS_CODES).isdisjoint(device_causes.keys()):
            return

        rule_lost_pkt = self.rule_flags.get(RULE_VLAN10_11_BALANCE, {})
        # 没有丢包
        if not rule_lost_pkt.get("value", False):
            return

        self._build_ubmem_cause(
            device_causes,
            code,
            rule_lost_pkt,
            suffix="+lost package({})".format(rule_lost_pkt.get("line", "")),
        )


class Cqe0x5Checker(Checker):
    base_code = Comp_CANN_HCCL_Custom_CQE0x5

    def analyze(self, device_causes):
        fault_codes = list(
            self.unknown_device_event.keys() | [] if not isinstance(device_causes, dict) else device_causes.keys()
        )
        # 基础码由 plog 解析，unknown_device_event或已识别的故障列表中无基础码则直接返回
        if self.base_code not in fault_codes:
            return

        keep_cqe = [
            self.add_cqe0x5_lost_pkg_causes(device_causes),
            self.add_cqe0x5_timeout_ta_ctp_low_causes(device_causes),
            self.add_cqe0x5_timeout_retraining_causes(device_causes),
            self.add_cqe0x5_abn_pkt_ssn_causes(device_causes),
            self.add_cqe0x5_abn_pkt_header_causes(device_causes),
            self.add_cqe0x5_rqe_not_enough_causes(device_causes),
            self.add_cqe0x5_rc_not_enough_causes(device_causes),
            self.add_cqe0x5_flow_cfg_err_causes(device_causes),
            self.add_cqe0x5_tpm_cfg_err_causes(device_causes),
        ]
        if any(keep_cqe):
            # 有故障匹配成功，移除基础码，正常情况是没有的，做兜底
            device_causes.pop(self.base_code, None)

    def add_cqe0x5_lost_pkg_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x5_LOST_PKG"
        if code in device_causes:
            return True
        rule = self.rule_flags.get(RULE_VLAN6_7_BALANCE, {})
        if not rule.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule)
        return True

    def add_cqe0x5_timeout_ta_ctp_low_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x5_TIMEOUT_TA_CTP_LOW"
        if code in device_causes:
            return True
        # 有UB_RAS故障直接返回,isdisjoint表示没有交集返回True
        if not set(UB_RAS_CODES).isdisjoint(device_causes.keys()):
            return False
        rule_lost_pkt = self.rule_flags.get(RULE_VLAN6_7_BALANCE, {})
        # 有丢包
        if rule_lost_pkt.get("value", False):
            return False
        rule_ta_timeout = self.rule_flags.get(RULE_HCOMM_TA_CTP_UB_TIMEOUT, {})
        if not rule_ta_timeout.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, event_key=PRECHECK_HCOMM_TA_CTP_UB_TIMEOUT)
        return True

    def add_cqe0x5_timeout_retraining_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x5_TIMEOUT_RETRAINING"
        if code in device_causes:
            return True
        # 有UB_RAS故障直接返回,isdisjoint表示没有交集返回True
        if not set(UB_RAS_CODES).isdisjoint(device_causes.keys()):
            return False
        rule_lost_pkt = self.rule_flags.get(RULE_VLAN6_7_BALANCE, {})
        # 有丢包
        if rule_lost_pkt.get("value", False):
            return False
        rule_reinit = self.rule_flags.get(RULE_PHY_REINIT_CNT_EXCEED, {})
        # 1秒内没有多次retraining
        if not rule_reinit.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule_reinit)
        return True

    def add_cqe0x5_abn_pkt_ssn_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x5_ABN_PKT_SSN"
        if code in device_causes:
            return True
        # 有UB_RAS故障直接返回,isdisjoint表示没有交集返回True
        if not set(UB_RAS_CODES).isdisjoint(device_causes.keys()):
            return False
        rule_lost_pkt = self.rule_flags.get(RULE_VLAN6_7_BALANCE, {})
        # 有丢包
        if rule_lost_pkt.get("value", False):
            return False
        rule_taack = self.rule_flags.get(RULE_TAACK_ABNORM_SSN_INCREASE, {})
        if not rule_taack.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule_taack)
        return True

    def add_cqe0x5_abn_pkt_header_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x5_ABN_PKT_HEADER"
        if code in device_causes:
            return True
        # 有UB_RAS故障直接返回,isdisjoint表示没有交集返回True
        if not set(UB_RAS_CODES).isdisjoint(device_causes.keys()):
            return False
        rule_lost_pkt = self.rule_flags.get(RULE_VLAN6_7_BALANCE, {})
        # 有丢包
        if rule_lost_pkt.get("value", False):
            return False
        rule_taack = self.rule_flags.get(RULE_TAACK_ABNORM_HEADER_INCREASE, {})
        if not rule_taack.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule_taack)
        return True

    def add_cqe0x5_rqe_not_enough_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x5_RQE_NOT_ENOUGH"
        if code in device_causes:
            return True
        rule_flag = self.rule_flags.get(RULE_TP_RRP_ERR_BIT25, {})
        if not rule_flag.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule_flag)
        return True

    def add_cqe0x5_rc_not_enough_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x5_RC_NOT_ENOUGH"
        if code in device_causes:
            return True
        rule_flag = self.rule_flags.get(RULE_TP_RRP_ERR_BIT27_28, {})
        if not rule_flag.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule_flag)
        return True

    def add_cqe0x5_flow_cfg_err_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x5_FLOW_CFG_ERR"
        if code in device_causes:
            return True
        rule_flag = self.rule_flags.get(RULE_DFX_TM_CRD_CTRL_INVALID, {})
        if not rule_flag.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule_flag)
        return True

    def add_cqe0x5_tpm_cfg_err_causes(self, device_causes):
        """返回值仅表示是否本故障匹配成功"""
        code = "Comp_Custom_CQE0x5_TPM_CFG_ERR"
        if code in device_causes:
            return True
        rule_flag = self.rule_flags.get(RULE_LQC_TAI_DFX_ALARM_BIT45, {})
        if not rule_flag.get("value", False):
            return False
        self._build_cqe_cause(device_causes, code, rule_flag)
        return True
