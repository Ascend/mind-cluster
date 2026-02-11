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

from typing import List

from toolkit.core.model.diag_result import Domain, DiagResult
from toolkit.core.model.optical_module import OpticalModuleInfo


class OpticalFaultChecker:
    COMM_SUGGEST = "建议排查本端和对端光模块本体是否失效。"

    def __init__(self, optical_module_threshold):
        self._threshold = optical_module_threshold
        self._power_rule_map = {  # 1表示异常；0表示正常
            0b1111: ("本端和对端收发光均异常。",
                     "两端光模块异常关断，建议先排查光链路中的端面/光纤和模块是否虚插。"),
            0b1110: ("本端收发光均异常，对端收光异常。",
                     "建议排查本端光模块是否虚插、是否被异常关断和是否本体失效。"),
            0b1101: ("本端收发光均异常，对端发光异常。", self.COMM_SUGGEST),
            0b1100: ("本端收发光均异常，对端收发光均正常。",
                     "可能是鸳鸯纤，建议排查本端和对端光模块连接是否正确。"),
            0b1011: ("本端收光异常，对端收发光均异常。",
                     "建议排查对端光模块是否虚插、是否被异常关断和是否本体失效。"),
            0b1010: ("本端收光异常，对端收光异常。",
                     "建议排查本端和对端光模块双向光链路端面/光纤是否异常。"),
            0b1001: ("本端收光异常，对端发光异常。",
                     "建议排查对端光模块本体发端是否失效。"),
            0b1000: ("本端收光异常，对端收发光均正常。",
                     "建议排查本端光模块本体接收端是否异常，光链路端面/光纤是否异常。"),
            0b0111: ("本端发光异常，对端收发光均异常。", self.COMM_SUGGEST),
            0b0110: ("本端发光异常，对端收光异常。",
                     "建议排查本端光模块是否虚插、是否被异常关断和是否本体失效。"),
            0b0101: ("本端发光异常，对端发光异常。", self.COMM_SUGGEST),
            0b0100: ("本端发光异常，对端收发光均正常。",
                     "本端光模块上报有误，建议排查本端光模块是否本体失效。"),
            0b0011: ("本端收发光均正常，对端收发光均异常。",
                     "可能是鸳鸯纤，建议排查本段和对端光模块连接是否正确。"),
            0b0010: ("本端收发光均正常，对端收光异常。",
                     "建议排查对端光模块本体接收端是否异常，光链路端面/光纤是否异常。"),
            0b0001: ("本端收发光均正常，对端发光异常。",
                     "对端光模块上报有误，建议排查对端光模块是否本体失效。")
        }

        self._single_power_rule_map = {
            0b11: ("本端收发光均异常，未收集到对端交换机或者交换机端口的光模块信息。",
                   "建议排查本端光模块本体是否失效，并收集对端信息后再分析。"),
            0b10: ("本端收光异常，未收集到对端交换机或者交换机端口的光模块信息。",
                   "建议排查本端光模块本体接收端是否异常，并收集对端信息后再分析。"),
            0b01: ("本端发光异常，未收集到对端交换机或者交换机端口的光模块信息。",
                   "建议排查本端光模块本体发送端是否异常，并收集对端信息后再分析。")
        }

        self._snr_rule_map = {  # 1表示异常；0表示正常
            0b11: (f"本端和对端光模块信噪比均异常。", "建议更换两端光模块。"),
            0b10: (f"本端光模块信噪比异常。", "建议更换对端光模块。"),
            0b01: (f"对端光模块信噪比异常。", "建议更换本端光模块。")
        }

        self._bias_rule_map = {  # 1表示异常；0表示正常
            0b11: (f"本端和对端光模块电流均异常。", "建议更换两端光模块。"),
            0b10: (f"本端光模块电流异常。", "建议更换本端光模块。"),
            0b01: (f"对端光模块电流异常。", "建议更换对端光模块。")
        }

    def power_analyze_single_ended(self, info: OpticalModuleInfo, domain_list: List[Domain]) -> List[DiagResult]:
        abn_rx_power_infos, abn_tx_power_infos = info.get_abnormal_power_infos(
            self._threshold.TX_POWER_THRESHOLD_CONFIG_DBM,
            self._threshold.RX_POWER_THRESHOLD_CONFIG_DBM
        )
        if not abn_rx_power_infos and not abn_tx_power_infos:
            return []
        description, suggest = self._check_single_power_value(bool(abn_rx_power_infos), bool(abn_tx_power_infos))
        if not description or not suggest:
            return []
        abn_power_infos = "\n".join(abn_rx_power_infos + abn_tx_power_infos)
        description = f"{description}\n本端：{abn_power_infos}\n对端：NA"
        return [DiagResult(domain_list, f"光模块光功率异常：{description}", suggest)]

    def snr_analyze_single_ended(self, info: OpticalModuleInfo, domain_list: List[Domain]) -> List[DiagResult]:
        res_list = []
        abnormal_snr_infos = info.get_abnormal_snr_infos(self._threshold.HOST_SNR_DB, self._threshold.MEDIA_SNR_DB)
        if abnormal_snr_infos:
            description = f"本端信噪比SNR异常，未收集到对端信息。\n本端：{abnormal_snr_infos}\n对端：NA"
            suggest = "建议收集对端信息并排查。"
            res_list.append(DiagResult(domain_list, description, suggest))
        return res_list

    def bias_analyze_single_ended(self, info: OpticalModuleInfo, domain_list: List[Domain]) -> List[DiagResult]:
        res_list = []
        abnormal_bias_infos = info.get_abnormal_bias_infos(self._threshold.TX_BIAS_MA)
        if abnormal_bias_infos:
            description = f"本端电流异常，未收集到对端信息。\n本端：{abnormal_bias_infos}\n对端：NA"
            suggest = "建议更换本端光模块或者收集对端信息并排查。"
            res_list.append(DiagResult(domain_list, description, suggest))
        return res_list

    def power_analyze(self, local_info: OpticalModuleInfo, remote_info: OpticalModuleInfo,
                      domain_list: List[Domain]) -> List[DiagResult]:
        local_abn_rx_power_infos, local_abn_tx_power_infos = local_info.get_abnormal_power_infos(
            self._threshold.TX_POWER_THRESHOLD_CONFIG_DBM,
            self._threshold.RX_POWER_THRESHOLD_CONFIG_DBM)
        remote_abn_rx_power_infos, remote_abn_tx_power_infos = remote_info.get_abnormal_power_infos(
            self._threshold.TX_POWER_THRESHOLD_CONFIG_DBM,
            self._threshold.RX_POWER_THRESHOLD_CONFIG_DBM)
        description, suggest = self._check_power_value(bool(local_abn_rx_power_infos), bool(local_abn_tx_power_infos),
                                                       bool(remote_abn_rx_power_infos), bool(remote_abn_tx_power_infos))
        if not description or not suggest:
            return []
        local_desc, remote_desc = "", ""
        if local_abn_rx_power_infos or local_abn_tx_power_infos:
            local_abn_power_infos = "\n".join(local_abn_rx_power_infos + local_abn_tx_power_infos)
            local_desc = f"\n本端：{local_abn_power_infos}"
        if remote_abn_rx_power_infos or remote_abn_tx_power_infos:
            remote_abn_power_infos = "\n".join(remote_abn_rx_power_infos + remote_abn_tx_power_infos)
            remote_desc = f"\n对端：{remote_abn_power_infos}"
        fault_info = f"光模块光功率异常：{description}{local_desc}{remote_desc}"
        return [DiagResult(domain_list, fault_info, suggest)]

    def snr_analyze(self, local_info: OpticalModuleInfo, remote_info: OpticalModuleInfo,
                    domain_list: List[Domain]) -> List[DiagResult]:
        local_abn_snr_infos = local_info.get_abnormal_snr_infos(self._threshold.HOST_SNR_DB,
                                                                self._threshold.MEDIA_SNR_DB)
        remote_abn_snr_infos = remote_info.get_abnormal_snr_infos(self._threshold.HOST_SNR_DB,
                                                                  self._threshold.MEDIA_SNR_DB)
        description, suggest = self._check_snr_value(bool(local_abn_snr_infos), bool(remote_abn_snr_infos))
        if not description or not suggest:
            return []
        local_desc, remote_desc = "", ""
        if local_abn_snr_infos:
            local_desc = f"\n本端：{local_abn_snr_infos}"
        if remote_abn_snr_infos:
            remote_desc = f"\n对端：{remote_abn_snr_infos}"
        return [DiagResult(domain_list, f"{description}{local_desc}{remote_desc}", suggest)]

    def bias_analyze(self, local_info: OpticalModuleInfo, remote_info: OpticalModuleInfo,
                     domain_list: List[Domain]) -> List[DiagResult]:
        local_abn_bias_infos = local_info.get_abnormal_bias_infos(self._threshold.TX_BIAS_MA)
        remote_abn_bias_infos = remote_info.get_abnormal_bias_infos(self._threshold.TX_BIAS_MA)
        description, suggest = self._check_bias_value(bool(local_abn_bias_infos), bool(remote_abn_bias_infos))
        if not description or not suggest:
            return []
        local_desc, remote_desc = "", ""
        if local_abn_bias_infos:
            local_desc = f"\n本端：{local_abn_bias_infos}"
        if remote_abn_bias_infos:
            remote_desc = f"\n对端：{remote_abn_bias_infos}"
        return [DiagResult(domain_list, f"{description}{local_desc}{remote_desc}", suggest)]

    def _check_power_value(self, is_local_rx_abnormal: bool, is_local_tx_abnormal: bool,
                           is_remote_rx_abnormal: bool, is_remote_tx_abnormal: bool):
        bits = (int(is_local_rx_abnormal) << 3) | (int(is_local_tx_abnormal) << 2) | (
                int(is_remote_rx_abnormal) << 1) | int(is_remote_tx_abnormal)
        return self._power_rule_map.get(bits, ("", ""))

    def _check_single_power_value(self, is_local_rx_abnormal: bool, is_local_tx_abnormal: bool):
        bits = (int(is_local_rx_abnormal) << 1) | (int(is_local_tx_abnormal))
        return self._single_power_rule_map.get(bits, ("", ""))

    def _check_snr_value(self, local_snr_abnormal: bool, remote_snr_abnormal: bool):
        bits = (int(local_snr_abnormal) << 1) | int(remote_snr_abnormal)
        return self._snr_rule_map.get(bits, ("", ""))

    def _check_bias_value(self, local_bias_abnormal: bool, remote_bias_abnormal: bool):
        bits = (int(local_bias_abnormal) << 1) | int(remote_bias_abnormal)
        return self._bias_rule_map.get(bits, ("", ""))
