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

from ascend_fd_tk.core.common import constants
from ascend_fd_tk.core.common.diag_enum import PowerUnitType
from ascend_fd_tk.core.common.json_obj import JsonObj
from ascend_fd_tk.core.model.optical_module import OpticalModule, OpticalModuleInfo, LanePowerInfo
from ascend_fd_tk.utils import list_tool
from ascend_fd_tk.utils import helpers
from ascend_fd_tk.utils.list_tool import safe_get


class BmcSelInfo(JsonObj):
    def __init__(self, sel_id: str = "", generation_time: str = "", severity: str = "", event_code: str = "",
                 status: str = "", event_description: str = ""):
        self.sel_id = sel_id
        self.generation_time = generation_time
        self.severity = severity
        self.event_code = event_code
        self.status = status
        self.event_description = event_description


class BmcSensorInfo(JsonObj):
    def __init__(self, sensor_id: str = "", sensor_name: str = "", value: str = "", unit: str = "", status: str = "",
                 lnr: str = "", lc: str = "", lnc: str = "", unc: str = "", uc: str = "", unr: str = "", phys: str = "",
                 n_hys: str = ""):
        self.sensor_id = sensor_id
        self.sensor_name = sensor_name
        self.value = value
        self.unit = unit
        self.status = status
        self.lnr = lnr
        self.lc = lc
        self.lnc = lnc
        self.unc = unc
        self.uc = uc
        self.unr = unr
        self.phys = phys
        self.n_hys = n_hys


class BmcHealthEvents(JsonObj):
    def __init__(self, event_num: str = "", event_time: str = "", alarm_level: str = "", event_code: str = "",
                 event_description: str = ""):
        self.event_num = event_num
        self.event_time = event_time
        self.alarm_level = alarm_level
        self.event_code = event_code
        self.event_description = event_description


class OpticalModuleHistoryLog(JsonObj, abc.ABC):
    @classmethod
    def _mapping_rules(cls, src_dict):
        return helpers.strip_keys_from_parentheses(src_dict)

    @classmethod
    def _parse_to_py_key(cls):
        return True


# 周期记录, 太长没有价值, 暂时不用
class PeriodicRecordingOpticalModuleHistoryLog(OpticalModuleHistoryLog):

    def __init__(self, log_time="", location="", optical_module_id="", tx_power_current_max="", tx_power_current_min="",
                 rx_power_current_max="", rx_power_current_min="", tx_bias_current_max="", tx_bias_current_min="",
                 tx_los="", rx_los="", host_snr_max="",
                 host_snr_min="", media_snr_max="", media_snr_min=""):
        self.log_time = log_time
        self.location = location
        self.optical_module_id = optical_module_id
        self.tx_power_current_max = tx_power_current_max
        self.tx_power_current_min = tx_power_current_min
        self.rx_power_current_max = rx_power_current_max
        self.rx_power_current_min = rx_power_current_min
        self.tx_bias_current_max = tx_bias_current_max
        self.tx_bias_current_min = tx_bias_current_min
        self.tx_los = tx_los
        self.rx_los = rx_los
        self.host_snr_max = host_snr_max
        self.host_snr_min = host_snr_min
        self.media_snr_max = media_snr_max
        self.media_snr_min = media_snr_min


# linkdown记录
class LinkDownOpticalModuleHistoryLog(OpticalModuleHistoryLog):

    def __init__(self, log_time="", location="", optical_module_id="", tx_power_current="", rx_power_current="",
                 tx_bias_current="", tx_los="", rx_los="", host_snr="", media_snr=""):
        self.log_time = log_time
        self.location = location
        self.optical_module_id = optical_module_id
        self.tx_power_current = tx_power_current
        self.rx_power_current = rx_power_current
        self.tx_bias_current = tx_bias_current
        self.tx_los = tx_los
        self.rx_los = rx_los
        self.host_snr = host_snr
        self.media_snr = media_snr


class BmcNpuInfo(JsonObj, OpticalModule):

    def __init__(self, npu_id="", chip_id="", chip_phy_id="",
                 link_down_optical_module_history_log: LinkDownOpticalModuleHistoryLog = None, lane_len=8):
        self.npu_id = npu_id
        self.chip_id = chip_id
        self.chip_phy_id = chip_phy_id
        self.link_down_optical_module_history_log = link_down_optical_module_history_log
        self._optical_module_info: OpticalModuleInfo = None
        self.lane_len = lane_len

    @staticmethod
    def _u_to_m(u_num: str, radio=1.0) -> str:
        success, res = helpers.to_float(u_num)
        if not success:
            return ""
        return str(round(res / 1000 * radio, 2))

    def get_optical_module_info(self) -> OpticalModuleInfo:
        optical_module_history_log = self.link_down_optical_module_history_log
        if not optical_module_history_log:
            return None
        tx_power_current = optical_module_history_log.tx_power_current.split()
        rx_power_current = optical_module_history_log.rx_power_current.split()
        tx_bias_current = optical_module_history_log.tx_bias_current.split()
        host_snr = optical_module_history_log.host_snr.split()
        media_snr = optical_module_history_log.media_snr.split()
        lane_power_infos: List[LanePowerInfo] = []
        for i in range(self.lane_len):
            lane_power_info = LanePowerInfo(
                lane_id=str(i),
                tx_power=self._u_to_m(safe_get(tx_power_current, i, ""), 0.1),
                rx_power=self._u_to_m(safe_get(rx_power_current, i, ""), 0.1),
                bias=self._u_to_m(safe_get(tx_bias_current, i, "")),
                host_snr=safe_get(host_snr, i, ""),
                media_snr=safe_get(media_snr, i, ""),
                power_unit_type=PowerUnitType.MW
            )
            lane_power_infos.append(lane_power_info)
        res = OpticalModuleInfo(lane_power_infos=lane_power_infos,
                                op_id=optical_module_history_log.optical_module_id,
                                tx_los=optical_module_history_log.tx_los,
                                rx_los=optical_module_history_log.rx_los,
                                log_time=optical_module_history_log.log_time)
        return res


class BmcInfo(JsonObj):
    _LANE_LEN_LIST = [8, 4, 2, 1]

    def __init__(self, bmc_id: str, sn_num: str,
                 bmc_sel_list: List[BmcSelInfo] = None,
                 sensor_info_list: List[BmcSensorInfo] = None,
                 health_events: List[BmcHealthEvents] = None,
                 link_down_optical_module_history_logs: List[LinkDownOpticalModuleHistoryLog] = None,
                 bmc_date=""):
        self.bmc_id = bmc_id
        self.sn_num = sn_num
        self.bmc_sel_list = bmc_sel_list or []
        self.sensor_info_list = sensor_info_list or []
        self.health_events = health_events or []
        self.link_down_optical_module_history_logs = link_down_optical_module_history_logs or []
        self.bmc_date = bmc_date
        self._bmc_npu_infos: List[BmcNpuInfo] = []

    def get_bmc_npu_infos(self) -> List[BmcNpuInfo]:
        if self._bmc_npu_infos:
            return self._bmc_npu_infos
        self._bmc_npu_infos = []
        group_dict = list_tool.group_by_to_dict(self.link_down_optical_module_history_logs,
                                                key=lambda log: log.location.replace("NPU", ""))
        # NPU序号从1开始
        for idx in range(1, constants.MAX_NPU_SIZE + 1):
            npu_id = str(idx)
            optical_module_history_log = group_dict.get(npu_id)
            if not optical_module_history_log:
                continue
            last_record = max(optical_module_history_log, key=lambda log: log.log_time)
            # 取最近一次
            bmc_npu_info = BmcNpuInfo(npu_id=npu_id, link_down_optical_module_history_log=last_record,
                                      lane_len=self._analyse_lane_len())
            self._bmc_npu_infos.append(bmc_npu_info)
        return self._bmc_npu_infos

    def _analyse_lane_len(self):
        none_power = '0'
        for lane_len in self._LANE_LEN_LIST:
            for opm_history_log in self.link_down_optical_module_history_logs:
                tx_parts = opm_history_log.tx_power_current.split()
                rx_parts = opm_history_log.rx_power_current.split()
                if lane_len <= len(tx_parts) and (
                        tx_parts[lane_len - 1] != none_power or rx_parts[lane_len - 1] != none_power):
                    return lane_len
        return 0
