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

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, HostDomain
from ascend_fd_tk.core.model.host_a5 import HostInfoA5, NpuChipInfoA5, HCCNOpticalInfoA5


@register_analyzer(generation=[NpuType.A5])
class HostAnalyzerA5(Analyzer):
    HOST_SNR_TYPE, MEDIA_SNR_TYPE = "HostSNR Lane", "MediaSNR Lane"

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self._threshold = cluster_info.get_threshold()

    @staticmethod
    def _analyze_hardware_attr(domain: HostDomain, optical_info: HCCNOpticalInfoA5):
        res = []
        for hard_info in optical_info.hardware_attr:
            if not hard_info.is_optical_present():
                res.append(
                    DiagResult(
                        domain=domain,
                        fault_info=f"光模块未在位，状态：{hard_info.present or 'NA'}",
                        suggestion="光模块可能松动，请重新插拔光模块",
                    )
                )
                continue
            if not hard_info.is_high_power():
                res.append(
                    DiagResult(
                        domain=domain,
                        fault_info=f"光模块处于低功率模式，high power enable reg:{hard_info.high_power}",
                        suggestion="光模块处于低功率模式，建议打开高功率模式",
                    )
                )
        return res

    def _analyze_optical_status(self, host_info: HostInfoA5, npu_chip_info: NpuChipInfoA5) -> List[DiagResult]:
        res = []
        optical_info_dict = npu_chip_info.hccn_optical_info
        if not optical_info_dict:
            return res
        for optical_id, optical_info in optical_info_dict.items():
            # pylint: disable=duplicate-code
            if not optical_info:
                continue
            domain = HostDomain(
                host_id=host_info.host_id,
                npu_id=npu_chip_info.npu_id,
                chip_phy_id=npu_chip_info.chip_phy_id,
                optical_id=optical_id,
            )
            res.extend(self._analyze_hardware_attr(domain, optical_info))
            res.extend(self._analyze_bias(domain, optical_info))
            res.extend(self._analyze_power(domain, optical_info))
            res.extend(self._analyze_optical_snr(domain, optical_info))
        return res

    def _analyze_bias(self, domain: HostDomain, optical_info: HCCNOpticalInfoA5) -> List[DiagResult]:
        if not optical_info or not optical_info.monitor_item:
            return []
        th = self._threshold.TX_BIAS_MA
        abnormal_infos = []
        for item in optical_info.monitor_item:
            items_name = item.items or ""
            if not items_name.startswith("Bias Lane"):
                continue
            lane_id = items_name.replace("Bias Lane", "").split("(")[0]
            desc = th.check_value_str(item.value)
            if desc:
                abnormal_infos.append(f"Lane{lane_id} {desc}")
        if not abnormal_infos:
            return []
        fault_info = "光模块Bias异常：\n" + "\n".join(abnormal_infos)
        return [DiagResult(domain=domain, fault_info=fault_info, suggestion="光模块电流异常，建议更换本端光模块")]

    def _analyze_power(self, domain: HostDomain, optical_info: HCCNOpticalInfoA5) -> List[DiagResult]:
        if not optical_info or not optical_info.monitor_item:
            return []
        abn_tx_infos = []
        abn_rx_infos = []
        for item in optical_info.monitor_item:
            items_name = item.items or ""
            if items_name.startswith("TxPower Lane"):
                lane_id = items_name.replace("TxPower Lane", "").split("(")[0]
                desc = self._threshold.TX_POWER_DBM.check_value_str(item.value)
                if desc:
                    abn_tx_infos.append(f"Lane{lane_id} {desc}")
            elif items_name.startswith("RxPower Lane"):
                lane_id = items_name.replace("RxPower Lane", "").split("(")[0]
                desc = self._threshold.RX_POWER_DBM.check_value_str(item.value)
                if desc:
                    abn_rx_infos.append(f"Lane{lane_id} {desc}")
        if not abn_tx_infos and not abn_rx_infos:
            return []
        abnormal_infos = abn_tx_infos + abn_rx_infos
        fault_info = "光模块光功率异常：\n" + "\n".join(abnormal_infos)
        return [DiagResult(domain=domain, fault_info=fault_info, suggestion="光功率异常，建议排查Los/Lol")]

    def _analyze_optical_snr(self, domain: HostDomain, optical_info: HCCNOpticalInfoA5) -> List[DiagResult]:
        diag_results = []
        if not optical_info or not optical_info.monitor_item:
            return diag_results
        abnormal_host_snr_infos, abnormal_media_snr_infos = [], []
        host_snr_list, media_snr_list = [], []

        for item in optical_info.monitor_item:
            items_name = item.items or ""
            if items_name.startswith(self.HOST_SNR_TYPE):
                lane_id = items_name.replace(self.HOST_SNR_TYPE, "").split("(")[0]
                desc = self._threshold.HOST_SNR_DB.check_value_str(item.value)
                if desc:
                    abnormal_host_snr_infos.append(f"Lane{lane_id} {desc}")
                host_snr_list.append([lane_id, item.value])
                continue
            if items_name.startswith(self.MEDIA_SNR_TYPE):
                lane_id = items_name.replace(self.MEDIA_SNR_TYPE, "").split("(")[0]
                desc = self._threshold.MEDIA_SNR_DB.check_value_str(item.value)
                if desc:
                    abnormal_media_snr_infos.append(f"Lane{lane_id} {desc}")
                media_snr_list.append([lane_id, item.value])
        if abnormal_media_snr_infos:
            fault_info = "光模块Media SNR异常：\n" + "\n".join(abnormal_media_snr_infos)
            diag_results.append(DiagResult(domain=domain, fault_info=fault_info, suggestion="建议更换交换机侧光模块"))
        if abnormal_host_snr_infos:
            fault_info = "光模块Host SNR异常：\n" + "\n".join(abnormal_host_snr_infos)
            diag_results.append(DiagResult(domain=domain, fault_info=fault_info, suggestion="建议更换本端光模块"))
        diag_results.extend(self._analyse_snr_lane_diff(media_snr_list, host_snr_list, domain))
        return diag_results

    def _analyse_snr_lane_diff(self, media_snr_list, host_snr_list, domain: HostDomain) -> List[DiagResult]:
        diff_desc_list = []
        for snr_list, value_type in ((media_snr_list, "Media SNR"), (host_snr_list, "Host SNR")):
            diff_desc_list.extend(self._threshold.SNR_LANE_DIFF_DB.check_lane_diff_desc(snr_list, value_type))
        if not diff_desc_list:
            return []
        fault_info = "光模块SNR Lane间差值异常：\n" + "\n".join(diff_desc_list)
        suggestion = "光模块SNR Lane间差值异常，优先排查SNR异常的Lane"
        return [DiagResult(domain=domain, fault_info=fault_info, suggestion=suggestion)]

    def analyse(self) -> List[DiagResult]:
        hosts_info = self.cluster_info.hosts_info
        diag_results = []
        for host_info in hosts_info.values():
            for npu_chip_info in host_info.npu_chip_info.values():
                diag_results.extend(self._analyze_optical_status(host_info, npu_chip_info))
        return diag_results
