#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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

import ipaddress
from typing import Tuple, Optional

from taskd.python.toolkit.constants.constants import IPVERSION_6


def pre_handle_ip(address: str) -> Tuple[str, Optional[Exception]]:
    if not address:
        return "", ValueError("address is empty")
    ip_str = address
    if address.startswith("[") and address.endswith("]"):
        ip_str = address[1:-1]
    try:
        ip = ipaddress.ip_address(ip_str)
        if ip.version == IPVERSION_6:
            return f"[{ip_str}]", None
        return ip_str, None
    except ValueError as e:
        return "", ValueError(f"invalid IP address: {address}")
