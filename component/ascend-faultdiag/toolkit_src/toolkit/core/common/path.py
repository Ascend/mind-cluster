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

import os
import platform


class CommonPath:
    # 项目根目录
    ROOT_DIR = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

    # 当前目录
    CUR_PATH = os.getcwd()

    # 用户目录
    USER_HOME = os.path.expanduser("~")

    # 临时目录(windows下用户路径可能无法删除
    TEMP_DIR = CUR_PATH if platform.system().lower() == "windows" else USER_HOME

    # 用户路径工具根目录
    TOOL_HOME = os.path.join(TEMP_DIR, ".ascend-faultdiag-toolkit")

    LOG_DIR = os.path.join(TOOL_HOME, "logs")

    # 缓存目录
    TOOL_HOME_CACHE_DIR = os.path.join(TOOL_HOME, "cache")

    PROJECT_CACHE_DIR = os.path.join(ROOT_DIR, "cache")

    # 清洗缓存
    PARSE_CACHE_DIR = os.path.join(TOOL_HOME_CACHE_DIR, "parse_cache")

    PARSE_CONFIG_FILE = os.path.join(PARSE_CACHE_DIR, "parse_config.json")

    # bmc日志导出缓存
    PROJECT_BMC_DUMP_CACHE_DIR = os.path.join(PROJECT_CACHE_DIR, "bmc_dump_cache")

    TOOL_HOME_BMC_DUMP_CACHE_DIR = os.path.join(TOOL_HOME_CACHE_DIR, "bmc_dump_cache")

    # host日志导出缓存
    HOST_DUMP_DIR_CACHE_DIR = os.path.join(PROJECT_CACHE_DIR, "host_dump_cache")

    # 交换设备命令行采集缓存
    SWI_DUMP_LOG_CACHE = os.path.join(PROJECT_CACHE_DIR, "switch_cli_output_cache")

    # 自动收集缓存
    COLLECT_CACHE = os.path.join(TOOL_HOME_CACHE_DIR, "collect_cache")

    COLLECT_BMC_CACHE_DIR = os.path.join(COLLECT_CACHE, "bmc")

    COLLECT_HOST_CACHE_DIR = os.path.join(COLLECT_CACHE, "host")

    COLLECT_SWITCH_CACHE_DIR = os.path.join(COLLECT_CACHE, "switch")

    # 报告
    REPORT_DIR = os.path.join(TOOL_HOME, "report")

    REPORT_FILE = os.path.join(REPORT_DIR, "diag_report.csv")

    INSPECTION_ERRORS_REPORT_FILE = os.path.join(REPORT_DIR, "inspection_errors.csv")

    # 连接配置
    CUR_PATH_CONN_CONFIG_PATH = os.path.join(CUR_PATH, "conn.ini")

    # bmc日志导出目录
    CUR_PATH_BMC_DUMP_LOG_DIR = os.path.join(CUR_PATH, "bmc_dump_log")

    # host日志导出目录
    CUR_PATH_HOST_DUMP_LOG_DIR = os.path.join(CUR_PATH, "host_dump_log")

    # 交换机
    CUR_PATH_SWITCH_DUMP_LOG_DIR = os.path.join(CUR_PATH, "switch_dump_log")
    
    # 加密配置文件路径
    ENCRYPTED_CONN_CONFIG_PATH = os.path.join(TOOL_HOME, "encrypted_conn_config")


class ConfigPath:
    CONN_CONFIG_DEFAULT_PATH = os.path.join(CommonPath.ROOT_DIR, "conn.ini")

    CONFIG_DIR = os.path.join(CommonPath.ROOT_DIR, "core", "config")

    L1_GLOBAL_ADDR_CONFIG_PATH = os.path.join(CONFIG_DIR, "L1_global_addr_config.json")

    L1_INTERFACE_PORT_CONFIG = os.path.join(CONFIG_DIR, "L1_interface_port_config.json")

    L1_LOCAL_ADDR_CONFIG = os.path.join(CONFIG_DIR, "L1_local_addr_config.json")
