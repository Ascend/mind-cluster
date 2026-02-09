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

import re
from typing import List, Dict

from diag_tool.core.common.constants import NPU_LINK_DOWN, NPU_LONG_DOWN_TIME, NPU_LINK_UP, HIGH_POWER_ENABLE, \
    OP_PRESENT, OP_TX_DISABLE_STATUS
from diag_tool.core.common.diag_enum import TimeFormat, PowerUnitType
from diag_tool.core.common.json_obj import JsonObj
from diag_tool.core.log_parser.base import FindResult
from diag_tool.core.model.optical_module import OpticalModule, OpticalModuleInfo, LanePowerInfo, Threshold
from diag_tool.utils.date_tool import DateObj

NUMERIC_PATTERN = re.compile(r'([+-]?\d*\.?\d+)')


class HCCNOpticalInfo(JsonObj):

    def __init__(self, present="", temperature="", high_power_enable_reg="", vendor_name="", vendor_part_number="",
                 vendor_serial_number="", vendor_org_unique_id="", xsfp_identifier="", xsfp_wave_length="",
                 manufact_date_code="", vcc="",
                 tx_power0="", rx_power0="", tx_power1="", rx_power1="", tx_power2="", rx_power2="", tx_power3="",
                 rx_power3="", vcc_high_thres="",
                 vcc_low_thres="", temp_high_thres="", temp_low_thres="",
                 tx_power_high_thres="", tx_power_low_thres="", rx_power_high_thres="", rx_power_low_thres="",
                 tx_bias0="", tx_bias1="", tx_bias2="", tx_bias3="", tx_los_flag="",
                 rx_los_flag="", tx_lo_l_flag="", rx_lo_l_flag="",
                 host_snr_lane0="", host_snr_lane1="", host_snr_lane2="", host_snr_lane3="",
                 media_snr_lane0="", media_snr_lane1="", media_snr_lane3="", media_snr_lane2="",
                 physical_code="", vendor_rev="", specification_compliance="", control_link_unreachable=False):
        self.present = present
        self.temperature = self.extract_numeric_value(temperature)
        self.high_power_enable_reg = high_power_enable_reg
        self.vendor_name = vendor_name
        self.vendor_part_number = vendor_part_number
        self.vendor_serial_number = vendor_serial_number
        self.vendor_org_unique_id = vendor_org_unique_id
        self.xsfp_identifier = xsfp_identifier
        self.xsfp_wave_length = self.extract_numeric_value(xsfp_wave_length)
        self.manufact_date_code = manufact_date_code
        self.vcc = self.extract_numeric_value(vcc)
        self.tx_power0 = self.extract_numeric_value(tx_power0)
        self.rx_power0 = self.extract_numeric_value(rx_power0)
        self.tx_power1 = self.extract_numeric_value(tx_power1)
        self.rx_power1 = self.extract_numeric_value(rx_power1)
        self.tx_power2 = self.extract_numeric_value(tx_power2)
        self.rx_power2 = self.extract_numeric_value(rx_power2)
        self.tx_power3 = self.extract_numeric_value(tx_power3)
        self.rx_power3 = self.extract_numeric_value(rx_power3)
        self.vcc_high_thres = self.extract_numeric_value(vcc_high_thres)
        self.vcc_low_thres = self.extract_numeric_value(vcc_low_thres)
        self.temp_high_thres = self.extract_numeric_value(temp_high_thres)
        self.temp_low_thres = self.extract_numeric_value(temp_low_thres)
        self.tx_power_high_thres = self.extract_numeric_value(tx_power_high_thres)
        self.tx_power_low_thres = self.extract_numeric_value(tx_power_low_thres)
        self.rx_power_high_thres = self.extract_numeric_value(rx_power_high_thres)
        self.rx_power_low_thres = self.extract_numeric_value(rx_power_low_thres)
        self.tx_bias0 = self.extract_numeric_value(tx_bias0)
        self.tx_bias1 = self.extract_numeric_value(tx_bias1)
        self.tx_bias2 = self.extract_numeric_value(tx_bias2)
        self.tx_bias3 = self.extract_numeric_value(tx_bias3)
        self.tx_los_flag = tx_los_flag
        self.rx_los_flag = rx_los_flag
        self.tx_lo_l_flag = tx_lo_l_flag
        self.rx_lo_l_flag = rx_lo_l_flag
        self.host_snr_lane0 = self.extract_numeric_value(host_snr_lane0)
        self.host_snr_lane1 = self.extract_numeric_value(host_snr_lane1)
        self.host_snr_lane2 = self.extract_numeric_value(host_snr_lane2)
        self.host_snr_lane3 = self.extract_numeric_value(host_snr_lane3)
        self.media_snr_lane0 = self.extract_numeric_value(media_snr_lane0)
        self.media_snr_lane1 = self.extract_numeric_value(media_snr_lane1)
        self.media_snr_lane2 = self.extract_numeric_value(media_snr_lane2)
        self.media_snr_lane3 = self.extract_numeric_value(media_snr_lane3)
        self.physical_code = physical_code
        self.vendor_rev = vendor_rev
        self.specification_compliance = specification_compliance
        self.control_link_unreachable = control_link_unreachable

    @staticmethod
    def extract_numeric_value(value_str):
        """从字符串中提取数值部分，如 "0.2239 mW" -> 0.2239"""
        if isinstance(value_str, str):
            # 直接使用预编译好的对象进行搜索
            match = NUMERIC_PATTERN.search(value_str)
            if match:
                return match.group(1)
        return value_str

    @classmethod
    def _parse_to_py_key(cls):
        return True

    def is_low_tx_power(self) -> bool:
        return (self.tx_power0 < self.tx_power_low_thres
                or self.tx_power1 < self.tx_power_low_thres
                or self.tx_power2 < self.tx_power_low_thres
                or self.tx_power3 < self.tx_power_low_thres)

    def is_high_power_enable(self) -> bool:
        return self.high_power_enable_reg == HIGH_POWER_ENABLE

    def is_optical_present(self) -> bool:
        return self.present == OP_PRESENT


class HCCNLinkHistory(JsonObj):
    def __init__(self, time="", link_status=""):
        self.time = time
        self.link_status = link_status


class HCCNLinkStatInfo(JsonObj):
    def __init__(self, current_time, link_up_count="", link_down_count="", link_history: List[HCCNLinkHistory] = None):
        self.current_time = current_time
        self.link_up_count = link_up_count
        self.link_down_count = link_down_count
        self.link_history = link_history or []

    def is_long_time_down(self):
        # 首次down与采集时间大于长时down 2倍则判定为长期down
        if not self.link_history:
            return False
        if NPU_LINK_DOWN == self.link_history[0].link_status:
            current_time = DateObj(self.current_time, TimeFormat.NPU_LINK_STAT_TIME.value)
            first_down_time = DateObj(self.link_history[0].time, TimeFormat.NPU_LINK_STAT_TIME.value)
            if current_time.diff_seconds(first_down_time) > (NPU_LONG_DOWN_TIME * 2):
                return True
        return False

    def get_down_interval(self):
        interval_list = []
        if not self.link_history:
            return interval_list
        up_time = None
        for i, link_history in enumerate(self.link_history):
            if NPU_LINK_UP == link_history.link_status:
                up_time = DateObj(link_history.time, TimeFormat.NPU_LINK_STAT_TIME.value)
            if NPU_LINK_DOWN == link_history.link_status:
                if i == 0:
                    up_time = DateObj(self.current_time, TimeFormat.NPU_LINK_STAT_TIME.value)
                down_time = DateObj(self.link_history[i].time, TimeFormat.NPU_LINK_STAT_TIME.value)
                if up_time:
                    interval_list.append(up_time.diff_seconds(down_time))
        return interval_list

    def is_repeat_up_down(self):
        interval_list = self.get_down_interval()
        for interval in interval_list:
            if interval > NPU_LONG_DOWN_TIME:
                return True
        return False

    def is_first_record_up(self):
        if not self.link_history:
            return False
        if NPU_LINK_UP == self.link_history[0].link_status:
            return True
        return False


class HCCNLLDPInfo(JsonObj):
    def __init__(self, port_id_tlv="", system_name_tlv=""):
        self.port_id_tlv = port_id_tlv
        self.system_name_tlv = system_name_tlv


class HccnPortHccsInfo(JsonObj):
    def __init__(self, hccs_health_status="", hccs_lane_mode="", hccs_link_speed="", hccs_first_err_lane=""):
        self.hccs_health_status = hccs_health_status
        self.hccs_lane_mode = self.trans_to_list(hccs_lane_mode)
        self.hccs_link_speed = self.trans_to_list(hccs_link_speed)
        self.hccs_first_err_lane = hccs_first_err_lane

    @staticmethod
    def trans_to_list(value_str):
        """将形如 '[2 4 4 4 4 4]' 的字符串转换为列表"""
        if isinstance(value_str, str):
            value_str = value_str.strip()
            # 去除方括号并按空格分割
            if value_str.startswith('[') and value_str.endswith(']'):
                # 提取括号内的内容
                content = value_str[1:-1].strip()
                if content:
                    # 按空格分割并转换为整数列表
                    return [int(x) for x in content.split()]
                else:
                    return []
        return value_str

    @classmethod
    def _parse_to_py_key(cls):
        return True


class HCCNDfxCfgInfo(JsonObj):
    def __init__(self, loopback_status: str = "", tx_disable_status: str = ""):
        self.loopback_status = loopback_status
        self.tx_disable_status = tx_disable_status

    @classmethod
    def _parse_to_py_key(cls):
        return True

    def is_tx_disable(self) -> bool:
        return not self.tx_disable_status or self.tx_disable_status != OP_TX_DISABLE_STATUS


class SpodInfo(JsonObj):
    def __init__(self, sdid="", super_pod_size="", super_pod_id="", server_index=""):
        self.sdid = sdid
        self.super_pod_size = super_pod_size
        self.super_pod_id = super_pod_id
        self.server_index = server_index
        _, self.npu_id, self.chip_id, self.chip_phy_id, self.sdid_four_part = self.sdid_convert(sdid)

    @staticmethod
    def sdid_convert(sdid):
        if not sdid:
            return "", "", "", "", ""
        binary = bin(int(sdid))[2:].zfill(32)
        sdid_segment1 = int(binary[0:8], 2)
        sdid_segment2 = int(binary[8:16], 2)
        sdid_segment3 = int(binary[16:24], 2)
        sdid_segment4 = int(binary[24:32], 2)
        # 以IP地址格式显示
        sdid_four_part = f"{sdid_segment1}.{sdid_segment2}.{sdid_segment3}.{sdid_segment4}"

        server_index = int(binary[0:10], 2)
        npu_id = int(binary[10:14], 2)
        chip_id = int(binary[14:16], 2)
        chip_phy_id = int(binary[16:30], 2)
        return str(server_index), str(npu_id), str(chip_id), str(chip_phy_id), sdid_four_part

    @classmethod
    def _parse_to_py_key(cls):
        return True


class CdrSnrInfo(JsonObj):
    def __init__(self, cdr_host_snr_lane1="",
                 cdr_host_snr_lane2="",
                 cdr_host_snr_lane3="",
                 cdr_host_snr_lane4="",
                 cdr_media_snr_lane1="",
                 cdr_media_snr_lane2="",
                 cdr_media_snr_lane3="",
                 cdr_media_snr_lane4=""):
        self.cdr_host_snr_lane1 = cdr_host_snr_lane1
        self.cdr_host_snr_lane2 = cdr_host_snr_lane2
        self.cdr_host_snr_lane3 = cdr_host_snr_lane3
        self.cdr_host_snr_lane4 = cdr_host_snr_lane4
        self.cdr_media_snr_lane1 = cdr_media_snr_lane1
        self.cdr_media_snr_lane2 = cdr_media_snr_lane2
        self.cdr_media_snr_lane3 = cdr_media_snr_lane3
        self.cdr_media_snr_lane4 = cdr_media_snr_lane4

    @classmethod
    def _mapping_rules(cls, src_dict):
        return {k: str(v).replace("dB", "").strip() for k, v in src_dict.items()}

    @classmethod
    def _parse_to_py_key(cls):
        return True

    def get_snr_abnormal_desc(self, host_threshold: Threshold, media_threshold: Threshold) -> str:
        snr_abnormal_desc = []
        for name, snr_value in vars(self).items():
            desc = ""
            if "host" in name:
                desc = host_threshold.check_value_str(snr_value)
            elif "media" in name:
                desc = media_threshold.check_value_str(snr_value)
            if desc:
                snr_abnormal_desc.append(f"{name} {desc}")
        return "\n".join(snr_abnormal_desc)


class HCCNStatExtraInfo(JsonObj):

    def __init__(self, cw_total_cnt="", cw_before_correct_cnt="", cw_correct_cnt="", cw_uncorrect_cnt="", cw_bad_cnt="",
                 trans_total_bit="", cw_total_correct_bit="", rx_full_drop_cnt="", pcs_err_cnt="",
                 rx_send_app_good_pkts="",
                 rx_send_app_bad_pkts="", correcting_bit_rate=""):
        self.cw_total_cnt = cw_total_cnt
        self.cw_before_correct_cnt = cw_before_correct_cnt
        self.cw_correct_cnt = cw_correct_cnt
        self.cw_uncorrect_cnt = cw_uncorrect_cnt
        self.cw_bad_cnt = cw_bad_cnt
        self.trans_total_bit = trans_total_bit
        self.cw_total_correct_bit = cw_total_correct_bit
        self.rx_full_drop_cnt = rx_full_drop_cnt
        self.pcs_err_cnt = pcs_err_cnt
        self.rx_send_app_good_pkts = rx_send_app_good_pkts
        self.rx_send_app_bad_pkts = rx_send_app_bad_pkts
        self.correcting_bit_rate = correcting_bit_rate

    @classmethod
    def _parse_to_py_key(cls):
        return True


class HCCNStatInfo(JsonObj):

    def __init__(self, mac_rx_mac_pause_num="", mac_tx_mac_pause_num="", mac_tx_pfc_pkt_num="",
                 mac_tx_pfc_pri0_pkt_num="",
                 mac_tx_pfc_pri1_pkt_num="", mac_tx_pfc_pri2_pkt_num="", mac_tx_pfc_pri3_pkt_num="",
                 mac_tx_pfc_pri4_pkt_num="",
                 mac_tx_pfc_pri5_pkt_num="", mac_tx_pfc_pri6_pkt_num="", mac_tx_pfc_pri7_pkt_num="",
                 mac_rx_pfc_pkt_num="", mac_rx_pfc_pri0_pkt_num="", mac_rx_pfc_pri1_pkt_num="",
                 mac_rx_pfc_pri2_pkt_num="",
                 mac_rx_pfc_pri3_pkt_num="", mac_rx_pfc_pri4_pkt_num="", mac_rx_pfc_pri5_pkt_num="",
                 mac_rx_pfc_pri6_pkt_num="",
                 mac_rx_pfc_pri7_pkt_num="", mac_tx_total_pkt_num="", mac_tx_total_oct_num="", mac_tx_bad_pkt_num="",
                 mac_tx_bad_oct_num="", mac_rx_total_pkt_num="", mac_rx_total_oct_num="", mac_rx_bad_pkt_num="",
                 mac_rx_bad_oct_num="",
                 max_rx_fcs_err_pkt_num="", roce_rx_rc_pkt_num="", roce_rx_all_pkt_num="", roce_rx_err_pkt_num="",
                 roce_tx_rc_pkt_num="", roce_tx_all_pkt_num="", roce_tx_err_pkt_num="", roce_cqe_num="",
                 roce_rx_cnp_pkt_num="",
                 roce_tx_cnp_pkt_num="", roce_unexpected_ack_num="", roce_out_of_order_num="",
                 roce_verificaton_err_num="",
                 roce_qp_status_err_num="",
                 roce_new_pkt_rty_num="", roce_ecn_db_num="", nic_tx_all_pkg_num="", nic_tx_all_oct_num="",
                 nic_rx_all_pkg_num="",
                 nic_rx_all_oct_num="",
                 ):
        self.mac_rx_mac_pause_num = mac_rx_mac_pause_num
        self.mac_tx_mac_pause_num = mac_tx_mac_pause_num
        self.mac_tx_pfc_pkt_num = mac_tx_pfc_pkt_num
        self.mac_tx_pfc_pri0_pkt_num = mac_tx_pfc_pri0_pkt_num
        self.mac_tx_pfc_pri1_pkt_num = mac_tx_pfc_pri1_pkt_num
        self.mac_tx_pfc_pri2_pkt_num = mac_tx_pfc_pri2_pkt_num
        self.mac_tx_pfc_pri3_pkt_num = mac_tx_pfc_pri3_pkt_num
        self.mac_tx_pfc_pri4_pkt_num = mac_tx_pfc_pri4_pkt_num
        self.mac_tx_pfc_pri5_pkt_num = mac_tx_pfc_pri5_pkt_num
        self.mac_tx_pfc_pri6_pkt_num = mac_tx_pfc_pri6_pkt_num
        self.mac_tx_pfc_pri7_pkt_num = mac_tx_pfc_pri7_pkt_num
        self.mac_rx_pfc_pkt_num = mac_rx_pfc_pkt_num
        self.mac_rx_pfc_pri0_pkt_num = mac_rx_pfc_pri0_pkt_num
        self.mac_rx_pfc_pri1_pkt_num = mac_rx_pfc_pri1_pkt_num
        self.mac_rx_pfc_pri2_pkt_num = mac_rx_pfc_pri2_pkt_num
        self.mac_rx_pfc_pri3_pkt_num = mac_rx_pfc_pri3_pkt_num
        self.mac_rx_pfc_pri4_pkt_num = mac_rx_pfc_pri4_pkt_num
        self.mac_rx_pfc_pri5_pkt_num = mac_rx_pfc_pri5_pkt_num
        self.mac_rx_pfc_pri6_pkt_num = mac_rx_pfc_pri6_pkt_num
        self.mac_rx_pfc_pri7_pkt_num = mac_rx_pfc_pri7_pkt_num
        self.mac_tx_total_pkt_num = mac_tx_total_pkt_num
        self.mac_tx_total_oct_num = mac_tx_total_oct_num
        self.mac_tx_bad_pkt_num = mac_tx_bad_pkt_num
        self.mac_tx_bad_oct_num = mac_tx_bad_oct_num
        self.mac_rx_total_pkt_num = mac_rx_total_pkt_num
        self.mac_rx_total_oct_num = mac_rx_total_oct_num
        self.mac_rx_bad_pkt_num = mac_rx_bad_pkt_num
        self.mac_rx_bad_oct_num = mac_rx_bad_oct_num
        self.max_rx_fcs_err_pkt_num = max_rx_fcs_err_pkt_num
        self.roce_rx_rc_pkt_num = roce_rx_rc_pkt_num
        self.roce_rx_all_pkt_num = roce_rx_all_pkt_num
        self.roce_rx_err_pkt_num = roce_rx_err_pkt_num
        self.roce_tx_rc_pkt_num = roce_tx_rc_pkt_num
        self.roce_tx_all_pkt_num = roce_tx_all_pkt_num
        self.roce_tx_err_pkt_num = roce_tx_err_pkt_num
        self.roce_cqe_num = roce_cqe_num
        self.roce_rx_cnp_pkt_num = roce_rx_cnp_pkt_num
        self.roce_tx_cnp_pkt_num = roce_tx_cnp_pkt_num
        self.roce_unexpected_ack_num = roce_unexpected_ack_num
        self.roce_out_of_order_num = roce_out_of_order_num
        self.roce_verificaton_err_num = roce_verificaton_err_num
        self.roce_qp_status_err_num = roce_qp_status_err_num
        self.roce_new_pkt_rty_num = roce_new_pkt_rty_num
        self.roce_ecn_db_num = roce_ecn_db_num
        self.nic_tx_all_pkg_num = nic_tx_all_pkg_num
        self.nic_tx_all_oct_num = nic_tx_all_oct_num
        self.nic_rx_all_pkg_num = nic_rx_all_pkg_num
        self.nic_rx_all_oct_num = nic_rx_all_oct_num


class UncorrCwCntInfo(JsonObj):
    UNCORR_CW_THRESHOLD = 10

    def __init__(self, device_id="", die_id="", count="", date_time=""):
        self.device_id = device_id
        self.die_id = die_id
        self.count = count
        self.date_time = date_time

    def count_check(self):
        return not self.count or int(self.count) <= self.UNCORR_CW_THRESHOLD


class RfLfPcsLinkInfo(JsonObj):
    def __init__(self, device_id="", die_id="", rf_lf="", pcs_link="", mac_link="", date_time=""):
        self.device_id = device_id
        self.die_id = die_id
        self.rf_lf = rf_lf
        self.pcs_link = pcs_link
        self.mac_link = mac_link
        self.date_time = date_time


class NpuChipInfo(JsonObj, OpticalModule):

    def __init__(self, hccn_lldp_info: HCCNLLDPInfo = None,
                 hccn_optical_info: HCCNOpticalInfo = None,
                 hccn_link_stat_info: HCCNLinkStatInfo = None,
                 hccn_stat_info: HCCNStatInfo = None,
                 hccn_dfx_cfg: HCCNDfxCfgInfo = None,
                 hccs_info: HccnPortHccsInfo = None,
                 spod_info: SpodInfo = None,
                 cdr_snr_info: CdrSnrInfo = None,
                 npu_type="", npu_id="", chip_id="", chip_phy_id="", speed="", duplex="", net_health="",
                 link_status=""):
        self.hccn_lldp_info = hccn_lldp_info
        self.hccn_optical_info = hccn_optical_info
        self.hccn_link_stat_info = hccn_link_stat_info
        self.hccn_stat_info = hccn_stat_info
        self.hccn_dfx_cfg = hccn_dfx_cfg
        self.hccs_info = hccs_info
        self.spod_info = spod_info
        self.cdr_snr_info = cdr_snr_info
        self.speed = speed
        self.duplex = duplex
        self.net_health = net_health
        self.link_status = link_status
        self._optical_module_info: OpticalModuleInfo = None
        # 关系属性
        self.npu_type = npu_type
        self.npu_id = npu_id
        self.chip_id = chip_id
        self.chip_phy_id = chip_phy_id

    def get_optical_module_info(self) -> OpticalModuleInfo:
        if self._optical_module_info:
            return self._optical_module_info
        if not self.hccn_optical_info:
            return None
        op_info = self.hccn_optical_info
        lane_power_infos = [
            LanePowerInfo("0", tx_power=op_info.tx_power0, rx_power=op_info.rx_power0, bias=op_info.tx_bias0,
                          media_snr=op_info.media_snr_lane0, host_snr=op_info.host_snr_lane0,
                          power_unit_type=PowerUnitType.MW),
            LanePowerInfo("1", tx_power=op_info.tx_power1, rx_power=op_info.rx_power1, bias=op_info.tx_bias1,
                          media_snr=op_info.media_snr_lane1, host_snr=op_info.host_snr_lane1,
                          power_unit_type=PowerUnitType.MW),
            LanePowerInfo("2", tx_power=op_info.tx_power2, rx_power=op_info.rx_power2, bias=op_info.tx_bias2,
                          media_snr=op_info.media_snr_lane2, host_snr=op_info.host_snr_lane2,
                          power_unit_type=PowerUnitType.MW),
            LanePowerInfo("3", tx_power=op_info.tx_power3, rx_power=op_info.rx_power3, bias=op_info.tx_bias3,
                          media_snr=op_info.media_snr_lane3, host_snr=op_info.host_snr_lane3,
                          power_unit_type=PowerUnitType.MW),
        ]
        self._optical_module_info = OpticalModuleInfo(lane_power_infos, op_info.vendor_serial_number)

        return self._optical_module_info


class NpuChipLoopBackInfo(JsonObj):
    def __init__(self, hccn_lldp_info: HCCNLLDPInfo = None,
                 hccn_optical_info: HCCNOpticalInfo = None,
                 spod_info: SpodInfo = None,
                 host_input_link_stat: HCCNLinkStatInfo = None,
                 media_output_link_stat: HCCNLinkStatInfo = None,
                 host_input_enable=False,
                 media_output_enable=False,
                 host_output_enable=False,
                 media_input_enable=False,
                 npu_type="", npu_id="", chip_id="", chip_phy_id=""):
        # 基本信息
        self.hccn_lldp_info = hccn_lldp_info
        self.hccn_optical_info = hccn_optical_info
        self.spod_info = spod_info
        self.npu_type = npu_type
        self.npu_id = npu_id
        self.chip_id = chip_id
        self.chip_phy_id = chip_phy_id
        # 环回使能场景
        self.host_input_enable = host_input_enable
        self.media_output_enable = media_output_enable
        self.host_output_enable = host_output_enable
        self.media_input_enable = media_input_enable
        # 环回使能端口状态
        self.host_input_link_stat = host_input_link_stat
        self.media_output_link_stat = media_output_link_stat


class NpuInfo(JsonObj):

    def __init__(self, hccn_lldp_info: Dict[str, HCCNLLDPInfo], hccn_optical_info: Dict[str, HCCNOpticalInfo] = None,
                 hccn_link_stat_info: Dict[str, HCCNLinkStatInfo] = None,
                 hccn_stat_info: Dict[str, HCCNStatInfo] = None,
                 hccs_info: Dict[str, HccnPortHccsInfo] = None,
                 spod_info: Dict[str, SpodInfo] = None,
                 npu_type="", npu_id="", speed="", duplex=""):
        self.hccn_lldp_info = hccn_lldp_info
        self.hccn_optical_info = hccn_optical_info
        self.hccn_link_stat_info = hccn_link_stat_info
        self.hccn_stat_info = hccn_stat_info
        self.hccs_info = hccs_info or {}
        self.spod_info = spod_info or {}
        self.npu_type = npu_type
        self.npu_id = npu_id
        self.speed = speed
        self.duplex = duplex


class HostInfo(JsonObj):
    def __init__(self, host_id: str, sn_num: str, hostname="", server_superpod_id="", server_index="",
                 msnpureport_log: List[FindResult] = None,
                 npu_chip_info: Dict[str, NpuChipInfo] = None,
                 loopback_info: List[NpuChipLoopBackInfo] = None):
        self.host_id = host_id
        self.sn_num = sn_num
        self.hostname = hostname
        self.server_superpod_id = server_superpod_id
        self.server_index = server_index
        self.msnpureport_log = msnpureport_log
        self.npu_chip_info = npu_chip_info
        self.loopback_info = loopback_info

    def get_msn_logs_by_type(self, type_key: str) -> List[FindResult]:
        if not type_key:
            return []
        return [find_result for find_result in self.msnpureport_log if find_result.pattern_key == type_key]
