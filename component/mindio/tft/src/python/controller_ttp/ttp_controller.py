#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
import atexit
import ttp_c2python_api
import ttp_logger
import ttp_mindx_api
from ttp_utils import is_zero_ip, input_ip_transform

RET_OK = 0
RET_ERROR = 1
RET_K8S = 2


def tft_init_controller(rank: int, world_size: int, enable_local_copy: bool, enable_arf=False, enable_zit=False):
    ret = ttp_c2python_api.init_controller(rank, world_size, enable_local_copy, enable_arf, enable_zit)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init controller failed, error num:{ret}")
        raise Exception(f"init controller failed, error num:{ret}")
    init_mindx_callback()


def init_mindx_callback():
    ret = ttp_c2python_api.set_register_check_callback(ttp_mindx_api.register_check_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init controller failed to set register callback, error num:{ret}")
        raise Exception(f"init controller failed to set register callback, error num:{ret}")
    ret = ttp_c2python_api.set_report_stop_complete_callback(ttp_mindx_api.report_stop_complete_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init controller failed to set report stop complete callback, error num:{ret}")
        raise Exception(f"init controller failed to set report stop complete callback, error num:{ret}")
    ret = ttp_c2python_api.set_report_strategies_callback(ttp_mindx_api.report_strategies_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init controller failed to set report recover strategy callback, error num:{ret}")
        raise Exception(f"init controller failed to set report recover strategy callback, error num:{ret}")
    ret = ttp_c2python_api.set_report_result_callback(ttp_mindx_api.report_result_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init controller failed to set report recover status callback, error num:{ret}")
        raise Exception(f"init controller failed to set report recover status callback, error num:{ret}")
    ret = ttp_c2python_api.set_report_fault_ranks_callback(ttp_mindx_api.report_fault_ranks_callback)
    if ret != RET_OK:
        ttp_logger.LOGGER.error(f"init controller failed to set report process fault callback, error num:{ret}")
        raise Exception(f"init controller failed to set report process fault callback, error num:{ret}")


def tft_start_controller(bind_ip: str, port: int, enable_tls=True, tls_info=''):
    if is_zero_ip(bind_ip):
        ttp_logger.LOGGER.error(f"start controller failed, all-zero ip is not supported ")
        raise SystemExit(RET_K8S)
    ip = input_ip_transform(bind_ip)
    ret = ttp_c2python_api.start_controller(ip, port, enable_tls, tls_info)
    if ret != RET_OK:
        ttp_logger.LOGGER.error("start controller failed, error:%s", ret)
        raise SystemExit(RET_K8S)


def tft_destroy_controller():
    ret = ttp_c2python_api.destroy_controller()
    if ret != RET_OK:
        ttp_logger.LOGGER.error("destroy controller failed, error num:%s", ret)


def tft_query_high_availability_switch():
    ttp_logger.LOGGER.info("query high availability switch...")
    return ttp_c2python_api.mindx_query_high_availability_switch()


atexit.register(tft_destroy_controller)