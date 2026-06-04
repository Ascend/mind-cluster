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
from typing import List, Optional, Tuple, Type

from ascend_fd_tk.core.config.port_mapping_config import L1InterfacePortMapping, get_port_mapping_config_instance
from ascend_fd_tk.core.config.threshold_config import OpticalModuleThreshold
from ascend_fd_tk.core.model.optical_module import OpticalModuleInfo
from ascend_fd_tk.core.model.switch import SwitchInfo

from ascend_fd_tk.core.root_cause.constants import PORT_SPEED_200G, PORT_SPEED_400G


class SnrChecker:
    """SNR信号检测器，负责链路SNR异常判定与格式化

    SNR数据来源:
        - hilink SNR: 来自交换机 hccs_info，包含 interface_snr_list 和 hccs_chip_port_snr_list
        - 光模块 SNR: 来自光模块 lane_power_infos，包含 host_snr 和 media_snr

    异常判定:
        通过阈值配置 (threshold) 中的 CDR_HOST_SNR_LINE、HOST_SNR_DB、MEDIA_SNR_DB
        对各SNR值进行越界检测
    """

    def __init__(self, threshold: Type[OpticalModuleThreshold]):
        self._threshold = threshold

    # ================================================================
    # 异常判定
    # ================================================================

    def check_hilink_snr_abnormal(self, swi_info: SwitchInfo, interface_name: str) -> Tuple[bool, bool]:
        """检测交换机端口hilink SNR是否异常

        分别检查200G和400G端口的SNR异常情况。
        SNR数据来源有两个: interface_snr_list 和 hccs_chip_port_snr_list，任一来源检测到异常即判定为异常。

        Returns:
            (is_200g_abnormal, is_400g_abnormal) 200G和400G端口是否异常
        """
        if not swi_info.hccs_info:
            return False, False

        is_200g_abnormal = False
        is_400g_abnormal = False
        port_speed = self.get_port_speed(interface_name)
        if port_speed == PORT_SPEED_200G:
            is_200g_abnormal = self._check_interface_snr_list(swi_info, interface_name)
        if port_speed == PORT_SPEED_400G:
            is_400g_abnormal = self._check_interface_snr_list(swi_info, interface_name)

        chip_port_abnormal = self._check_chip_port_snr(swi_info, interface_name, port_speed)
        if chip_port_abnormal == PORT_SPEED_200G:
            is_200g_abnormal = True
        elif chip_port_abnormal == PORT_SPEED_400G:
            is_400g_abnormal = True

        return is_200g_abnormal, is_400g_abnormal

    def _check_interface_snr_list(self, swi_info: SwitchInfo, interface_name: str) -> bool:
        """检查 interface_snr_list 中的SNR异常

        遍历 hccs_info.interface_snr_list，匹配目标接口名，
        对每个异常通道SNR值进行阈值检测
        """
        for interface_snr in swi_info.hccs_info.interface_snr_list:
            if interface_snr.interface_name != interface_name:
                continue
            for lane_snr in interface_snr.abnormal_lane_snr:
                if self._threshold.CDR_HOST_SNR_LINE.check_value_str(lane_snr.snr_value):
                    return True
            break
        return False

    def _check_chip_port_snr(
        self,
        swi_info: SwitchInfo,
        interface_name: str,
        port_speed: str,
    ) -> Optional[str]:
        """检查 hccs_chip_port_snr_list 中的SNR异常

        通过端口映射配置将 (chip_id, port_id) 映射为接口名，
        匹配目标接口后对SNR值进行阈值检测

        Returns:
            异常端口的速率 ("200G"/"400G")，无异常返回None
        """
        for chip_port_snr in swi_info.hccs_info.hccs_chip_port_snr_list:
            port_mapping = self._find_port_mapping(chip_port_snr.switch_chip_id, chip_port_snr.port_id)
            if port_mapping and port_mapping.swi_port != interface_name:
                continue
            if self._threshold.CDR_HOST_SNR_LINE.check_value_str(chip_port_snr.snr):
                return port_speed
            break
        return None

    @staticmethod
    def _find_port_mapping(swi_chip_id: str, port_id: str) -> Optional[L1InterfacePortMapping]:
        """查找端口映射配置，异常时返回None"""
        try:
            return get_port_mapping_config_instance().find_swi_port(swi_chip_id, port_id)
        except Exception:
            return None

    def check_optical_host_snr_abnormal(self, optical_info: OpticalModuleInfo) -> bool:
        """检测光模块host SNR是否异常

        遍历光模块的所有通道，任一通道的host_snr超过阈值即判定异常
        """
        if not optical_info or not optical_info.lane_power_infos:
            return False
        return any(
            self._threshold.HOST_SNR_DB.check_value_str(lane_info.host_snr)
            for lane_info in optical_info.lane_power_infos
        )

    def check_optical_media_snr_abnormal(self, optical_info: OpticalModuleInfo) -> bool:
        """检测光模块media SNR是否异常

        遍历光模块的所有通道，任一通道的media_snr超过阈值即判定异常
        """
        if not optical_info or not optical_info.lane_power_infos:
            return False
        return any(
            self._threshold.MEDIA_SNR_DB.check_value_str(lane_info.media_snr)
            for lane_info in optical_info.lane_power_infos
        )

    # ================================================================
    # SNR 格式化
    # ================================================================

    @staticmethod
    def format_hilink_snr(swi_info: SwitchInfo, interface_name: str) -> str:
        """格式化交换机端口的hilink SNR值

        从 interface_snr_list 和 hccs_chip_port_snr_list 两个来源收集SNR值，
        格式为 "通道名:SNR值"，多个值以分号分隔

        Returns:
            格式化后的SNR字符串，如 "Lane0:12.5;chip1_port0:11.8"
        """
        if not swi_info.hccs_info or not interface_name:
            return ""

        snr_values = SnrChecker._collect_interface_snr_values(swi_info, interface_name)
        snr_values.extend(SnrChecker._collect_chip_port_snr_values(swi_info, interface_name))

        return "; ".join(snr_values)

    @staticmethod
    def _collect_interface_snr_values(swi_info: SwitchInfo, interface_name: str) -> List[str]:
        """从 interface_snr_list 收集SNR格式化值"""
        values = []
        for interface_snr in swi_info.hccs_info.interface_snr_list:
            if interface_snr.interface_name == interface_name:
                for lane_snr in interface_snr.abnormal_lane_snr:
                    values.append(f"{lane_snr.lane_name}:{lane_snr.snr_value}")
                break
        return values

    @staticmethod
    def _collect_chip_port_snr_values(swi_info: SwitchInfo, interface_name: str) -> List[str]:
        """从 hccs_chip_port_snr_list 收集SNR格式化值"""
        values = []
        for chip_port_snr in swi_info.hccs_info.hccs_chip_port_snr_list:
            port_mapping = SnrChecker._find_port_mapping_static(chip_port_snr.switch_chip_id, chip_port_snr.port_id)
            if port_mapping and port_mapping.swi_port == interface_name:
                values.append(f"chip{chip_port_snr.switch_chip_id}_port{chip_port_snr.port_id}:{chip_port_snr.snr}")
                break
        return values

    @staticmethod
    def _find_port_mapping_static(swi_chip_id: str, port_id: str) -> Optional[L1InterfacePortMapping]:
        """静态方法版本的端口映射查找，供格式化方法使用"""
        try:
            return get_port_mapping_config_instance().find_swi_port(swi_chip_id, port_id)
        except Exception:
            return None

    @staticmethod
    def format_lane_snr(optical_info: Optional[OpticalModuleInfo], snr_type: str) -> str:
        """格式化光模块的通道SNR值

        Args:
            optical_info: 光模块信息
            snr_type: SNR类型，"host_snr" 或 "media_snr"

        Returns:
            格式化后的SNR字符串，如 "Lane0:12.5;Lane1:11.8"
        """
        if not optical_info or not optical_info.lane_power_infos:
            return ""

        snr_values = []
        for lane_info in optical_info.lane_power_infos:
            value = getattr(lane_info, snr_type, "")
            if value:
                snr_values.append(f"Lane{lane_info.lane_id}:{value}")
        return "; ".join(snr_values)

    # ================================================================
    # 端口速率识别
    # ================================================================

    @staticmethod
    def get_port_speed(interface_name: str) -> str:
        """根据接口名识别端口速率

        接口名以 "400G"开头 → 400G
        接口名以 "200G"开头 → 200G
        其他 → 空字符串
        """
        name_upper = interface_name.upper()
        if name_upper.startswith(PORT_SPEED_400G):
            return PORT_SPEED_400G
        if name_upper.startswith(PORT_SPEED_200G):
            return PORT_SPEED_200G
        return ""
