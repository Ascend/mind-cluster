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
from dataclasses import dataclass, field
from typing import Dict


@dataclass
class Position:
    """
    Dataclass. Position contains the position of a node.
    """
    role: str
    server_rank: str
    process_rank: str


@dataclass
class TLSConfig:
    ca: str
    server_key: str
    server_crt: str
    client_key: str
    client_crt: str


@dataclass
class NetworkConfig:
    """
    Dataclass. NetworkConfig contains the network config of a node.
    """
    pos: Position
    upstream_addr: str
    listen_addr: str
    enable_tls: bool
    tls_conf: TLSConfig


@dataclass
class MessageInfo:
    """
    Dataclass. Message contains info sent to taskd manager.
    """
    uuid: str
    biz_type: str
    dst: Position
    body: str


@dataclass
class MsgBody:
    """
    Dataclass. MsgBody contains the body of a message.
    """
    msg_type: str
    code: int
    message: str
    extension: Dict[str, str]


@dataclass
class AgentReportInfo:
    """
    Dataclass. AgentReportInfo contains the report info.
    """
    fault_ranks: list = field(default_factory=list)
    restart_times: int = -1


LOCAL_HOST = "127.0.0.1"
CONFIG_UPSTREAMIP_KEY = "UpstreamAddr"
CONFIG_LISTENIP_KEY = "ListenAddr"
CONFIG_UPSTREAMPORT_KEY = "UpstreamPort"
CONFIG_LISTENPORT_KEY = "ListenPort"
CONFIG_ROLE_KEY = "Role"
CONFIG_SERVERRANK_KEY = "ServerRank"
CONFIG_PROCESSRANK_KEY = "ProcessRank"
DEFAULT_PROXY_UPSTREAMPORT = "9601"
DEFAULT_PRXOY_LISTENPORT = "9602"
DEFAULT_PROXY_ROLE = "Proxy"
DEFAULT_SERVERRANK = "0"
DEFAULT_PROCESSRANK = "-1"

CONFIG_FRAMEWORK_KEY = "Framework"
DEFAULT_AGENT_ROLE = "Agent"
DEFAULT_AGENT_UPSTREAMPORT = "9602"

DEFAULT_BIZTYPE = "DEFAULT"
