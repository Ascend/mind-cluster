#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version  2.0 (the "License");
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

from typing import List, Tuple

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.context.register import register_analyzer
from ascend_fd_tk.core.fault_analyzer.base import Analyzer
from ascend_fd_tk.core.model.diag_result import DiagResult, HostDomain
from ascend_fd_tk.core.model.host_a5 import NICInfoA5, NICPortLaneInfoA5


@register_analyzer(generation=[NpuType.A5])
class HostNicLaneAnalyzerA5(Analyzer):
    """A5 网卡 SFP lane 级故障分析器。"""

    # 未在位/未采集 lane 的占位值
    _UNUSED_LANE_VALUES = ("", "-inf", "-inf dbm")

    def __init__(self, cluster_info):
        super().__init__(cluster_info)
        self._threshold = cluster_info.get_threshold()

    def _check_nic_port_lanes(
        self,
        host_id: str,
        nic_info: NICInfoA5,
        port_lane_info: NICPortLaneInfoA5,
    ) -> List[DiagResult]:
        results = []
        if not port_lane_info:
            return results
        domain = HostDomain(host_id=host_id, nic_id=nic_info.card_name, port_id=port_lane_info.port_id)
        results.extend(self._check_flag_lanes(domain, port_lane_info))
        results.extend(self._check_numeric_lanes(domain, port_lane_info))
        return results

    def _check_flag_lanes(self, domain: HostDomain, port_lane: NICPortLaneInfoA5) -> List[DiagResult]:
        """检查 tx_los / rx_los / tx_cdr_lol / rx_cdr_lol flag 异常。

        回显中 flag 为 "0" 表示正常，非 "0" 表示异常；空串表示未采集（跳过）。
        """
        results: List[DiagResult] = []
        flag_specs: List[Tuple[str, List[str]]] = [
            ("TxLos", port_lane.tx_los_lanes),
            ("RxLos", port_lane.rx_los_lanes),
            ("TxCdrLol", port_lane.tx_cdr_lol_lanes),
            ("RxCdrLol", port_lane.rx_cdr_lol_lanes),
        ]
        th = self._threshold.NIC_LANE_FLAG
        for flag_name, lanes in flag_specs:
            if not lanes:
                continue
            for lane_idx, lane_val in enumerate(lanes):
                if not lane_val:
                    continue
                desc = th.check_value_str(lane_val)
                if desc:
                    # lane_idx 0~7 对应回显 lane 1~8
                    results.append(
                        DiagResult(
                            domain=domain,
                            fault_info=f"网卡端口 Lane{lane_idx + 1} {flag_name} 异常：实际值 {lane_val}（正常值 0）",
                            suggestion=f"请检查网卡端口 Lane{lane_idx + 1} {flag_name} 指标",
                        )
                    )
        return results

    def _check_numeric_lanes(self, domain: HostDomain, port_lane: NICPortLaneInfoA5) -> List[DiagResult]:
        """检查 bias / tx_power / rx_power / host_snr / media_snr 异常。"""
        results: List[DiagResult] = []
        numeric_specs = [
            ("bias", port_lane.bias_lanes, self._threshold.NIC_TX_BIAS_MA),
            ("tx_power", port_lane.tx_power_lanes, self._threshold.NIC_TX_POWER_DBM),
            ("rx_power", port_lane.rx_power_lanes, self._threshold.NIC_RX_POWER_DBM),
            ("host_snr", port_lane.host_snr_lanes, self._threshold.NIC_HOST_SNR_DB),
            ("media_snr", port_lane.media_snr_lanes, self._threshold.NIC_MEDIA_SNR_DB),
        ]
        for item_name, lanes, th in numeric_specs:
            if not lanes:
                continue
            abnormal_infos: List[str] = []
            for lane_idx, lane_val in enumerate(lanes):
                if not self._is_valid_numeric(lane_val):
                    continue
                desc = th.check_value_str(lane_val)
                if desc:
                    abnormal_infos.append(f"Lane{lane_idx + 1} {desc}")
            if not abnormal_infos:
                continue
            fault_info = f"网卡端口 {item_name} 异常：\n" + "\n".join(abnormal_infos)
            results.append(
                DiagResult(
                    domain=domain,
                    fault_info=fault_info,
                    suggestion=f"网卡端口 {item_name} 异常，建议排查 Los/Lol",
                )
            )
        return results

    @classmethod
    def _is_valid_numeric(cls, value: str) -> bool:
        """判断数值类字段是否需要参与阈值检查（空串或 -inf 等未在位占位值跳过）。"""
        if not value:
            return False
        return value.strip().lower() not in cls._UNUSED_LANE_VALUES

    def analyse(self) -> List[DiagResult]:
        results = []
        for host_info in self.cluster_info.hosts_info.values():
            for nic_info in host_info.nic_info_list:
                if not nic_info:
                    continue
                for port_lane_info in nic_info.port_lane_info_list:
                    results.extend(
                        self._check_nic_port_lanes(
                            host_id=host_info.host_id,
                            nic_info=nic_info,
                            port_lane_info=port_lane_info,
                        )
                    )
        return results
