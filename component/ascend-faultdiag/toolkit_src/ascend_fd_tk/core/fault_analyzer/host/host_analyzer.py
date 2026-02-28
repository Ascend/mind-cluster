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

from ascend_fd_tk.core.common import diag_enum
from ascend_fd_tk.core.common.constants import KEY_IIC_ERROR, KEY_UNCORR_CW, UNCORR_CW_THRESHOLD, KEY_PCS_LINK
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.diag_result import DiagResult, Domain
from ascend_fd_tk.core.model.host import HostInfo, NpuChipInfo, UncorrCwCntInfo, RfLfPcsLinkInfo
from ascend_fd_tk.utils.date_tool import DateObj


@register_analyzer
class HostAnalyzer(Analyzer):
    NUM_TWO = 2

    def __init__(self, cluster_info: ClusterInfoCache):
        super().__init__(cluster_info)
        self._threshold = cluster_info.get_threshold()

    @staticmethod
    def get_npu_chip_domain(ip, npu_id, chip_id):
        return [Domain(diag_enum.DeviceType.SERVER, ip),
                Domain(diag_enum.DeviceType.NPU, npu_id),
                Domain(diag_enum.DeviceType.CHIP, chip_id)]

    def analyse(self) -> List[DiagResult]:
        hosts_info = self.cluster_info.hosts_info
        diag_results = []
        for host_info in hosts_info.values():
            diag_results.extend(self.analyze_single_host(host_info))
        return diag_results

    def analyze_single_host(self, host_info: HostInfo) -> List[DiagResult]:
        diag_results = []
        for npu_chip_info in host_info.npu_chip_info.values():
            diag_results.extend(self.analyze_single_npu_chip(host_info, npu_chip_info))
        diag_results.extend(self._analyze_iic_fault(host_info))
        return diag_results

    def analyze_single_npu_chip(self, host_info: HostInfo, npu_chip_info: NpuChipInfo) -> List[DiagResult]:
        diag_results = []
        diag_results.extend(self._analyze_optical_status(host_info, npu_chip_info))
        diag_results.extend(self._analyze_power(host_info, npu_chip_info))
        diag_results.extend(self._analyze_optical_snr(host_info, npu_chip_info))
        diag_results.extend(self._analyze_cdr(host_info, npu_chip_info))
        diag_results.extend(self._analyze_uncorr_cw_cnt_fault(host_info, npu_chip_info))
        return diag_results

    def _analyze_optical_status(self, host_info: HostInfo, npu_chip_info: NpuChipInfo) -> List[DiagResult]:
        optical_info = npu_chip_info.hccn_optical_info
        if not optical_info:
            return []
        domain = self.get_npu_chip_domain(host_info.host_id, npu_chip_info.npu_id, npu_chip_info.chip_phy_id)
        if not optical_info.is_optical_present():
            return [
                DiagResult(domain, f"光模块未在位, 状态: {optical_info.present or 'NA'}", "光模块可能松动，请重新插拔光模块")]
        if optical_info.control_link_unreachable:
            return [DiagResult(domain, "光模块Control link unreachable", "光模块可能故障，请联系技术支持人员")]
        if not optical_info.is_high_power_enable():
            fault_info = f"光模块处于低功率模式，high power enable reg:{optical_info.high_power_enable_reg}"
            return [DiagResult(domain, fault_info, "光模块处于低功率模式，建议打开高功率模式")]

        dfx_cfg_info = npu_chip_info.hccn_dfx_cfg
        if not dfx_cfg_info:
            return []
        if dfx_cfg_info.is_tx_disable():
            fault_info = f"光模块处于关光状态，tx disable status：{dfx_cfg_info.tx_disable_status}"
            return [DiagResult(domain, fault_info, "建议保证光模块为开光状态")]
        return []

    def _analyze_power(self, host_info: HostInfo, npu_chip_info: NpuChipInfo) -> List[DiagResult]:
        optical_info = npu_chip_info.get_optical_module_info()
        if not optical_info or optical_info.lane_power_infos:
            return []
        domain = self.get_npu_chip_domain(host_info.host_id, npu_chip_info.npu_id, npu_chip_info.chip_phy_id)
        abn_rx_power_infos, abn_tx_power_infos = optical_info.get_abnormal_power_infos(
            self._threshold.TX_POWER_THRESHOLD_CONFIG_MW,
            self._threshold.RX_POWER_THRESHOLD_CONFIG_MW
        )
        if not abn_rx_power_infos and not abn_tx_power_infos:
            return []
        abn_power_infos = "\n".join(abn_rx_power_infos + abn_tx_power_infos)
        fault_info = f"光模块光功率异常，{abn_power_infos}"
        suggestion = "光功率异常，建议排查Los/Lol"
        return [DiagResult(domain, fault_info, suggestion)]

    def _analyze_optical_snr(self, host_info: HostInfo, npu_chip_info: NpuChipInfo) -> List[DiagResult]:
        diag_results = []
        optical_info = npu_chip_info.get_optical_module_info()
        if not optical_info or optical_info.lane_power_infos:
            return diag_results
        domain = self.get_npu_chip_domain(host_info.host_id, npu_chip_info.npu_id, npu_chip_info.chip_phy_id)
        abn_snr_infos = optical_info.get_abnormal_snr_infos(self._threshold.HOST_SNR_DB, self._threshold.MEDIA_SNR_DB)
        fault_info = f"光模块SNR异常：\n{abn_snr_infos}"
        suggestion = "建议更换交换机侧光模块"
        diag_results.append(DiagResult(domain, fault_info, suggestion))
        diff_value_desc = optical_info.get_lane_diff_desc()
        if diff_value_desc:
            fault_info = f"光模块SNR LANE间差值异常：{diff_value_desc}"
            suggestion = "光模块SNR LANE间差值异常，优先排查SNR异常的LANE"
            diag_results.append(DiagResult(domain, fault_info, suggestion))
        return diag_results

    def _analyze_cdr(self, host_info: HostInfo, npu_chip_info: NpuChipInfo) -> List[DiagResult]:
        diag_results = []
        cdr_snr_info = npu_chip_info.cdr_snr_info
        if not cdr_snr_info:
            return diag_results
        fault_info = cdr_snr_info.get_snr_abnormal_desc(self._threshold.CDR_HOST_SNR_DB,
                                                        self._threshold.CDR_MEDIA_SNR_DB)
        if not fault_info:
            return diag_results
        domain = self.get_npu_chip_domain(host_info.host_id, npu_chip_info.npu_id, npu_chip_info.chip_phy_id)
        suggestion = "CDR SNR异常，请根据SNR异常排查NPU侧或者光模块"
        diag_results.append(DiagResult(domain, fault_info, suggestion))
        return diag_results

    def _analyze_uncorr_cw_cnt_fault(self, host_info: HostInfo, npu_chip_info: NpuChipInfo) -> List[DiagResult]:
        diag_results = []
        if not host_info or not host_info.msnpureport_log:
            return diag_results
        uncorr_cw_cnt_infos = host_info.get_msn_logs_by_type(KEY_UNCORR_CW)
        if not uncorr_cw_cnt_infos:
            return diag_results
        uncorr_cw_cnt_list = []
        for uncorr_cw_cnt_info in uncorr_cw_cnt_infos:
            uncorr_cw_cnt = UncorrCwCntInfo.from_dict(uncorr_cw_cnt_info.info_dict)
            if (uncorr_cw_cnt.device_id != npu_chip_info.npu_id or uncorr_cw_cnt.die_id != npu_chip_info.chip_id or
                    uncorr_cw_cnt.count_check()):
                continue
            uncorr_cw_cnt_list.append(uncorr_cw_cnt)
        if not uncorr_cw_cnt_list:
            return diag_results
        uncorr_cw_cnt_list.sort(key=lambda x: x.date_time, reverse=True)
        domain = self.get_npu_chip_domain(host_info.host_id, npu_chip_info.npu_id, npu_chip_info.chip_id)
        for i in range(len(uncorr_cw_cnt_list) - self.NUM_TWO):
            first_item = DateObj(uncorr_cw_cnt_list[i].date_time, "%Y-%m-%d %H:%M:%S.%f")
            second_item = DateObj(uncorr_cw_cnt_list[i + self.NUM_TWO].date_time, "%Y-%m-%d %H:%M:%S.%f")
            # 连续3次>10、24小时出现多次>10且伴随闪断、24小时出现1~2次>100且伴随闪断
            if first_item.diff_seconds(second_item) < UNCORR_CW_THRESHOLD:
                fault_info = (f"持续连续3次出现uncorr_cw_cnt > 10，发生时间：{uncorr_cw_cnt_list[i].date_time}，"
                              f"{uncorr_cw_cnt_list[i + 1].date_time}，{uncorr_cw_cnt_list[i + self.NUM_TWO].date_time}")
                suggestion = "连续3次出现uncorr_cw_cnt > 10，请优先排查光纤脏污，然后排查两端光模块"
                diag_results.append(DiagResult(domain, fault_info, suggestion))
                return diag_results
        return diag_results

    def _analyze_rf_lf_pcs_link_fault(self, host_info: HostInfo, npu_chip_info: NpuChipInfo) -> List[DiagResult]:
        diag_results = []
        if not host_info or not host_info.msnpureport_log:
            return diag_results
        rf_lf_pcs_link_infos = host_info.get_msn_logs_by_type(KEY_PCS_LINK)
        if not rf_lf_pcs_link_infos:
            return diag_results
        rf_lf_pcs_link_list = []
        for rf_lf_pcs_link_info in rf_lf_pcs_link_infos:
            rf_lf_pcs_link = RfLfPcsLinkInfo.from_dict(rf_lf_pcs_link_info.info_dict)
            if not rf_lf_pcs_link.pcs_link and not rf_lf_pcs_link.rf_lf:
                rf_lf_pcs_link_list.append(rf_lf_pcs_link)
        rf_lf_pcs_link_list.sort(key=lambda x: x.date_time, reverse=True)
        return diag_results

    def _analyze_iic_fault(self, host_info: HostInfo):
        diag_results = []
        if not host_info or not host_info.msnpureport_log:
            return diag_results
        iic_error_infos = host_info.get_msn_logs_by_type(KEY_IIC_ERROR)
        if not iic_error_infos:
            return diag_results
        for iic_error_info in iic_error_infos:
            device_id = iic_error_info.info_dict.get('device_id', '')
            die_id = iic_error_info.info_dict.get('die_id', '')
            domain = self.get_npu_chip_domain(host_info.host_id, device_id, die_id)
            fault_info = "检测到IIC异常: trans status[0x40], error status[0x10], NPU板载光模块转接器可能存在故障"
            diag_results.append(DiagResult(domain, fault_info, "建议更换NPU板载光模块转接器"))
        return diag_results
