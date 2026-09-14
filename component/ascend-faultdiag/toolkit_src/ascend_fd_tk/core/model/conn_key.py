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


class ConnKey:
    """所有 fetcher map 的统一 key（host/bmc/switch/PoDManager）。

    ip 为设备唯一标识（IP 或名称），conn_idx 区分同一设备的多个采集分片：
    多连接设备（switch/PoDManager 分组并行采集）用 conn_idx=0..N-1；
    单连接设备与离线 file fetcher 直接用默认 conn_idx=0。
    作为 dict key 使用，按 ip/conn_idx 值判等与哈希。
    """

    def __init__(self, ip: str, conn_idx: int = 0):
        self.ip = ip
        self.conn_idx = conn_idx

    def __eq__(self, other):
        return isinstance(other, ConnKey) and self.ip == other.ip and self.conn_idx == other.conn_idx

    def __hash__(self):
        return hash((self.ip, self.conn_idx))
