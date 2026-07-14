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

from typing import List, Optional

from ascend_fd_tk.core.model.bmc import BmcInfo, BmcSelInfo, LinkDownOpticalModuleHistoryLog
from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.core.model.host import HostInfo


BMC_ID = "bmc-1"
HOST_ID = "server-1"
SN_NUM = "SN001"


def build_sel_event(
    sel_id: str,
    event_code: str,
    event_description: str,
    generation_time: str = "2026-06-01 10:00:00",
) -> BmcSelInfo:
    return BmcSelInfo(
        sel_id=sel_id,
        generation_time=generation_time,
        severity="Critical",
        event_code=event_code,
        status="Active",
        event_description=event_description,
    )


def build_linkdown_log(
    log_time: str,
    optical_module_id: str,
    tx_power_current: str,
    rx_power_current: str,
    tx_bias_current: str,
    tx_los: str,
    rx_los: str,
    host_snr: str,
    media_snr: str,
) -> LinkDownOpticalModuleHistoryLog:
    return LinkDownOpticalModuleHistoryLog(
        log_time=log_time,
        location="NPU1",
        optical_module_id=optical_module_id,
        tx_power_current=tx_power_current,
        rx_power_current=rx_power_current,
        tx_bias_current=tx_bias_current,
        tx_los=tx_los,
        rx_los=rx_los,
        host_snr=host_snr,
        media_snr=media_snr,
    )


def build_bmc_info(
    bmc_sel_list: Optional[List[BmcSelInfo]] = None,
    linkdown_logs: Optional[List[LinkDownOpticalModuleHistoryLog]] = None,
) -> BmcInfo:
    return BmcInfo(
        bmc_id=BMC_ID,
        sn_num=SN_NUM,
        bmc_sel_list=bmc_sel_list or [],
        link_down_optical_module_history_logs=linkdown_logs or [],
    )


def build_cluster_info(bmc_info: BmcInfo, include_host: bool = False) -> ClusterInfoCache:
    hosts_info = {HOST_ID: HostInfo(host_id=HOST_ID, sn_num=SN_NUM)} if include_host else {}
    return ClusterInfoCache(hosts_info=hosts_info, bmcs_info={BMC_ID: bmc_info})
