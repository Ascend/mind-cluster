#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
import json
import os
import time
from dataclasses import asdict
from taskd.python.cython_api import cython_api
from taskd.python.utils.log import run_log
from taskd.python.framework.common.type import CONFIG_SERVERRANK_KEY, Position, NetworkConfig, LOCAL_HOST, DEFAULT_PROXY_UPSTREAMPORT, \
     DEFAULT_PRXOY_LISTENPORT, DEFAULT_PROXY_ROLE, DEFAULT_SERVERRANK, DEFAULT_PROCESSRANK, CONFIG_UPSTREAMIP_KEY, \
     CONFIG_LISTENIP_KEY, CONFIG_UPSTREAMPORT_KEY, CONFIG_LISTENPORT_KEY
from taskd.python.toolkit.constants import constants


def init_taskd_proxy(config: dict) -> bool:
    if cython_api.lib is None:
        run_log.error("init_taskd_proxy: the libtaskd.so has not been loaded")
        return False

    default_values = {
        CONFIG_UPSTREAMIP_KEY: LOCAL_HOST,
        CONFIG_LISTENIP_KEY: LOCAL_HOST,
        CONFIG_UPSTREAMPORT_KEY: DEFAULT_PROXY_UPSTREAMPORT,
        CONFIG_LISTENPORT_KEY: DEFAULT_PRXOY_LISTENPORT,
        CONFIG_SERVERRANK_KEY: os.getenv("RANK") or os.getenv("MS_NODE_RANK") or DEFAULT_SERVERRANK
    }

    config_values = {}
    for key, default in default_values.items():
        config_values[key] = config.get(key, default)

    configs = NetworkConfig(
        pos=Position(
            role=DEFAULT_PROXY_ROLE,
            server_rank=config_values.get(CONFIG_SERVERRANK_KEY),
            process_rank=str(int(time.time()))
        ),
        upstream_addr=config_values.get(CONFIG_UPSTREAMIP_KEY) + ":" + config_values.get(CONFIG_UPSTREAMPORT_KEY),
        listen_addr=config_values.get(CONFIG_LISTENIP_KEY) + ":" + config_values.get(CONFIG_LISTENPORT_KEY),
        enable_tls=False,
        tls_conf=None
    )

    use_local_proxy = os.getenv(constants.LOCAL_PROXY_ENABLE)
    if use_local_proxy == "on":
        run_log.info(f"taskd_proxy use local proxy connect mgr")
        configs.upstream_addr = constants.LOCAL_PROXY_IP + ":" + config_values.get(CONFIG_UPSTREAMPORT_KEY)
    run_log.info(f"init_taskd_proxy: configs is {configs}")
    config_json = json.dumps(asdict(configs)).encode('utf-8')
    try:
        res = cython_api.lib.InitTaskdProxy(config_json)
        if res != 0:
            run_log.error("init_taskd_proxy: init_taskd_proxy fail, reason in taskd proxy log")
            return False
    except Exception as e:
        run_log.error(f"init_taskd_proxy: encounter exception: {e}")
        return False
    return True


def destroy_taskd_proxy() -> bool:
    if cython_api.lib is None:
        run_log.error("destroy_taskd_proxy: the libtaskd.so has not been loaded")
        return False
    try:
        destroy_proxy_func = cython_api.lib.DestroyTaskdProxy
        if destroy_proxy_func is None:
            run_log.error("destroy_taskd_proxy: func DestroyTaskdProxy has not been loaded from libtaskd.so")
            return False
        destroy_proxy_func()
    except Exception as e:
        run_log.error(f"destroy_taskd_proxy: encounter exception: {e}")
        return False
    return True
