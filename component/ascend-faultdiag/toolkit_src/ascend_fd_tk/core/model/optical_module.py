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
import abc
from typing import List

from ascend_fd_tk.core.common.constants import SNR_LANE_DIFF_THRESHOLD
from ascend_fd_tk.core.common.diag_enum import PowerUnitType
from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.core.config.threshold_config import OpticalModuleThreshold
from ascend_fd_tk.core.model.threshold import Threshold
from ascend_fd_tk.utils import helpers


class LanePowerInfo(JsonObj):

    def __init__(self, lane_id: str = "", tx_power: str = "", rx_power: str = "", bias: str = "",
                 media_snr: str = "", host_snr: str = "", power_unit_type: PowerUnitType = PowerUnitType.MW):
        self.lane_id = lane_id
        self.tx_power = tx_power
        self.rx_power = rx_power
        self.bias = bias
        self.media_snr = media_snr
        self.host_snr = host_snr
        self.power_unit_type = power_unit_type

    @property
    def tx_power_dbm(self):
        return self.tx_power if self.power_unit_type == PowerUnitType.DBM else helpers.mw_to_dbm(self.tx_power)

    @property
    def rx_power_dbm(self):
        return self.rx_power if self.power_unit_type == PowerUnitType.DBM else helpers.mw_to_dbm(self.rx_power)


class OpticalModuleInfo(JsonObj):

    def __init__(self, lane_power_infos: List[LanePowerInfo] = None, sn="", op_id="", tx_los="", rx_los="",
                 log_time=""):
        self.lane_power_infos: List[LanePowerInfo] = lane_power_infos or []
        self.sn = sn
        self.op_id = op_id
        self.tx_los = tx_los
        self.rx_los = rx_los
        self.log_time = log_time

    def get_lane_diff_desc(self) -> str:
        check_snr_list = []
        for info in self.lane_power_infos:
            success, media_snr_float = helpers.to_float(info.media_snr)
            if success:
                check_snr_list.append([info.lane_id, media_snr_float])
        if len(check_snr_list) <= 1:
            return ""
        check_snr_list.sort(key=lambda x: x[1])
        min_lane_id, min_media_snr = check_snr_list[0]
        max_lane_id, max_media_snr = check_snr_list[-1]
        if max_media_snr - min_media_snr > SNR_LANE_DIFF_THRESHOLD:
            return (f"Lane最大值和最小值差值大于{SNR_LANE_DIFF_THRESHOLD}db，"
                    f"实际最大值lane{max_lane_id}：{max_media_snr}db，"
                    f"最小值lane{min_lane_id}：{min_media_snr}db")
        return ""

    def get_abnormal_snr_infos(self, host_th: Threshold, media_th: Threshold):
        abnormal_snr_list = []
        for info in self.lane_power_infos:
            host_desc = host_th.check_value_str(info.host_snr)
            if host_desc:
                abnormal_snr_list.append(f"Lane{info.lane_id} {host_desc}")
            media_desc = media_th.check_value_str(info.media_snr)
            if media_desc:
                abnormal_snr_list.append(f"Lane{info.lane_id} {media_desc}")
        return "\n".join(abnormal_snr_list)

    def get_abnormal_bias_infos(self, th: Threshold = None):
        abnormal_bias_list = []
        for info in self.lane_power_infos:
            desc = th.check_value_str(info.bias)
            if desc:
                abnormal_bias_list.append(f"Lane{info.lane_id} {desc}")
        return "\n".join(abnormal_bias_list)

    def get_abnormal_power_infos(self, th: OpticalModuleThreshold):
        abnormal_rx_power_list = []
        abnormal_tx_power_list = []
        th_tx = th.TX_POWER_THRESHOLD_CONFIG_DBM
        th_rx = th.RX_POWER_THRESHOLD_CONFIG_DBM
        for info in self.lane_power_infos:
            if info.power_unit_type == PowerUnitType.MW:
                th_tx = th.TX_POWER_THRESHOLD_CONFIG_MW
                th_rx = th.RX_POWER_THRESHOLD_CONFIG_MW
            desc_tx = th_tx.check_value_str(info.tx_power)
            if desc_tx:
                abnormal_tx_power_list.append(f"Lane{info.lane_id} {desc_tx}")
            desc_rx = th_rx.check_value_str(info.rx_power)
            if desc_rx:
                abnormal_rx_power_list.append(f"Lane{info.lane_id} {desc_rx}")
        return abnormal_rx_power_list, abnormal_tx_power_list

    def get_abnormal_txrx_infos(self):
        abnormal_txrx_info_list = []
        if self.tx_los and helpers.parse_hex(self.tx_los) > 0:
            abnormal_txrx_info_list.append(f"")


class OpticalModule(abc.ABC):

    @abc.abstractmethod
    def get_optical_module_info(self) -> OpticalModuleInfo:
        pass
