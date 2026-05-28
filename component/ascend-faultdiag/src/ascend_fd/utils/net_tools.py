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
import ipaddress


class IPAddress:
    """
    IP address utility class providing IPv4 / IPv6 detection and validation.
    """

    # ========== IPv4 / IPv6 Detection ==========

    @classmethod
    def is_ipv4(cls, address: str) -> bool:
        """
        Check whether the given address is a valid IPv4 address.

        Args:
            address (str): IP address string to check

        Returns:
            bool: True if valid IPv4 address, False otherwise.
                  Empty string returns False.
        """
        try:
            return isinstance(ipaddress.ip_address(address), ipaddress.IPv4Address)
        except (ValueError, TypeError):
            return False

    @classmethod
    def is_ipv6(cls, address: str) -> bool:
        """
        Check whether the given address is a valid IPv6 address.

        Args:
            address (str): IP address string to check

        Returns:
            bool: True if valid IPv6 address (full or compressed format), False otherwise.
                  Empty string returns False.
        """
        try:
            return isinstance(ipaddress.ip_address(address), ipaddress.IPv6Address)
        except (ValueError, TypeError):
            return False

    @classmethod
    def is_valid_ip(cls, address: str) -> bool:
        """
        Check whether the given address is a valid IP (IPv4 or IPv6).

        Args:
            address (str): IP address string to check

        Returns:
            bool: True if a valid IPv4 or IPv6 address, False otherwise.
        """
        try:
            ipaddress.ip_address(address)
            return True
        except (ValueError, TypeError):
            return False
