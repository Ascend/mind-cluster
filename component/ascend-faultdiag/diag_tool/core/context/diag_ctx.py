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

import asyncio
import os
from collections import defaultdict
from concurrent.futures import ProcessPoolExecutor
from typing import List, Dict, Callable

from diag_tool.core.collect.fetcher.bmc_fetcher import BmcFetcher
from diag_tool.core.collect.fetcher.host_fetcher import HostFetcher
from diag_tool.core.collect.fetcher.switch_fetcher import SwitchFetcher
from diag_tool.core.common.constants import CPU_UTILIZATION_RATIO
from diag_tool.core.common.path import CommonPath, ConfigPath
from diag_tool.core.config.conn_config import DeviceConfigParser
from diag_tool.core.config.dump_log_dir_config import DumpLogDirConfig
from diag_tool.core.config.tool_config import ToolConfig
from diag_tool.core.crypto.crypto import RootKeyCrypto
from diag_tool.core.crypto.key_generator import KeyGenerator
from diag_tool.core.log_parser.base import LogParsePattern
from diag_tool.core.model.cluster_info_cache import ClusterInfoCache
from diag_tool.core.model.diag_result import DiagResult
from diag_tool.core.model.inspection import InspectionErrorItem


class DiagCtx:

    def __init__(self):
        self.conn_config = None
        self.dump_log_dir_config = DumpLogDirConfig()
        self.tool_config = ToolConfig()
        self.host_fetchers: Dict[str, HostFetcher] = {}
        self.switch_fetchers: Dict[str, SwitchFetcher] = {}
        self.bmcs_fetchers: Dict[str, BmcFetcher] = {}
        self.parse_log_result_map: Dict[str, List[LogParsePattern]] = defaultdict(list)
        self.cache = ClusterInfoCache()
        self.diag_result: List[DiagResult] = []
        self.inspection_result: List[InspectionErrorItem] = []
        self.process_pool = ProcessPoolExecutor(max_workers=int(os.cpu_count() * CPU_UTILIZATION_RATIO))
        self.crypto = RootKeyCrypto(KeyGenerator().generate_complex_password())

    def close(self):
        if self.process_pool:
            self.process_pool.shutdown(wait=True)

    def submit_multi_process_task(self, task: Callable, *args, **kwargs):
        return asyncio.wrap_future(self.process_pool.submit(task, *args, **kwargs))

    def encrypt_conn_config(self, config_path=ConfigPath.CONN_CONFIG_DEFAULT_PATH):
        if not os.path.exists(config_path):
            config_path = CommonPath.CUR_PATH_CONN_CONFIG_PATH

        # 读取配置文件内容
        with open(config_path, 'r') as f:
            config_content = f.read()

        # 加密配置文件内容
        encrypted_content = self.crypto.encrypt_with_salt(config_content)

        # 确保TOOL_HOME目录存在
        if not os.path.exists(CommonPath.TOOL_HOME):
            os.makedirs(CommonPath.TOOL_HOME)

        # 保存加密后的配置到TOOL_HOME目录
        with open(CommonPath.ENCRYPTED_CONN_CONFIG_PATH, 'w') as f:
            f.write(encrypted_content)

    def load_conn_config(self):
        if not os.path.exists(CommonPath.ENCRYPTED_CONN_CONFIG_PATH):
            return "加密配置文件不存在"
        try:
            # 从加密缓存加载
            with open(CommonPath.ENCRYPTED_CONN_CONFIG_PATH, 'r') as f:
                encrypted_data = f.read()
            # 解密数据
            decrypted_data = self.crypto.decrypt_with_salt(encrypted_data)
            # 解析临时文件
            conn_config = DeviceConfigParser(decrypted_data).parse()
            self.conn_config = conn_config
            return ""
        except Exception as e:
            return str(e)
