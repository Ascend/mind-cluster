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

import re
from typing import List, Dict

from ascend_fd_tk.core.common.json_obj import JsonObj


class Conn(JsonObj):
    def __init__(self, host, port, username, password, private_key):
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.private_key = private_key


class ConnConfig(JsonObj):
    def __init__(self, host_conn: List[Conn], bmc_conn: List[Conn], switch_conn: List[Conn]):
        self.host_conn = host_conn
        self.bmc_conn = bmc_conn
        self.switch_conn = switch_conn


class DeviceConfigParser:
    _DEFAULT_GROUPS = ("host", "switch", "bmc", "config")

    def __init__(self, config_data: str):
        self._raw_lines = config_data.split("\n")  # 读取原始文本行
        self._groups = self._parse_groups()  # 按分组整理行
        self._comm_config = {}

    @staticmethod
    def _ip_to_int(ip: str) -> int:
        """将 IPv4 地址转为 32 位整数"""
        octets = list(map(int, ip.strip().split('.')))
        return (octets[0] << 24) | (octets[1] << 16) | (octets[2] << 8) | octets[3]

    @staticmethod
    def _int_to_ip(num: int) -> str:
        """将 32 位整数转回 IPv4 地址格式"""
        return f"{(num >> 24) & 0xFF}.{(num >> 16) & 0xFF}.{(num >> 8) & 0xFF}.{num & 0xFF}"

    @staticmethod
    def _is_valid_ip(ip: str) -> bool:
        ip_pattern = re.compile(r'\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$')
        if not ip_pattern.match(ip):
            return False
        octets = list(map(int, ip.split('.')))
        return all(0 <= octet <= 255 for octet in octets)

    def parse(self) -> ConnConfig:
        """解析所有分组，返回ConnConfig对象"""
        self._parse_comm_config()
        return ConnConfig(self._get_group_conns("host"), self._get_group_conns("bmc"), self._get_group_conns("switch"))

    def _parse_groups(self) -> Dict[str, List[str]]:
        """手动解析分组（[host] [bmc] [switch] [config]），将行归类到对应分组"""
        groups = {group: [] for group in self._DEFAULT_GROUPS}
        current_group = None

        for line in self._raw_lines:
            # 跳过空行和注释
            if not line or line.startswith((';', '#')):
                continue

            # 检测分组标记
            if line.startswith('[') and line.endswith(']'):
                group_name = line[1:-1].strip().lower()
                if group_name in groups:
                    current_group = group_name
                continue

            # 将行添加到当前分组
            if current_group:
                groups[current_group].append(line)

        return groups

    def _parse_line(self, line: str) -> List[Conn]:
        """解析单行配置：IP username=xxx password=xxx [port=xxx]"""
        parts = line.split()
        if len(parts) < 2:
            raise ValueError(f"无效格式: {line}（至少需要 IP、username）")
        ip_range = parts[0]
        port_str = "port"
        step_str = "step"
        params = {port_str: 22, step_str: 1}
        for part in parts[1:]:
            if "=" not in part:
                continue
            key, value = part.split("=", 1)
            key = key.strip().lower()
            # 去除引号
            if value.startswith(('"', "'")) and value.endswith(('"', "'")):
                value = value[1:-1].strip()
            # 转换端口/步长为整数
            if key == port_str or key == step_str:
                try:
                    value = int(value)
                except ValueError as e:
                    raise ValueError(f"{key}必须是整数: {value}（行: {line}）") from e
                if value <= 0:
                    raise ValueError(f"{key}必须是正整数: {value}（行: {line}）")
            params[key] = value
        # 验证必填参数
        if "username" not in params:
            raise ValueError(f"缺少参数: {line}（必须包含username）")

        conn_list = []
        for ip in self._generate_ips_from_range(line, ip_range, params[step_str]):
            conn_list.append(Conn(
                host=ip,
                port=params[port_str],
                username=params["username"],
                password=params.get("password", ""),
                private_key=params.get("private_key", "") or self._comm_config.get("private_key", "")
            ))
        return conn_list

    def _generate_ips_from_range(self, line: str, ip_range: str, step: int = 1) -> List[str]:
        """
        Generate all IPv4 addresses based on the IP range and step size
        :param line: data line
        :param ip_range: IP range string in the format such as "1.1.1.0-1.1.255.255" or "1.1.1.0"
        :param step: step size (positive integer, default: 1)
        :return: list of all valid generated IP addresses
        """
        if '-' not in ip_range:
            if not self._is_valid_ip(ip_range):
                raise ValueError(
                    f"无效 IP 地址，或者 IP 不合法，各字节需在 0-255 之间。或者 IP 段配置未使用'-'连接：{ip_range}（行: {line}）"
                )
            return [ip_range]
        start_ip_str, end_ip_str = ip_range.split('-', 1)
        if not self._is_valid_ip(start_ip_str):
            raise ValueError(f"无效 IP 地址，或者 IP 不合法，各字节需在 0-255 之间：{start_ip_str}（行: {line}）")
        if not self._is_valid_ip(end_ip_str):
            raise ValueError(f"无效 IP 地址，或者 IP 不合法，各字节需在 0-255 之间：{end_ip_str}（行: {line}）")
        start_num = self._ip_to_int(start_ip_str)
        end_num = self._ip_to_int(end_ip_str)
        if start_num > end_num:
            raise ValueError(f"起始 IP 大于结束 IP（行: {line}）")
        ip_list = []
        for num in range(start_num, end_num + 1, step):
            ip_list.append(self._int_to_ip(num))
        return ip_list

    def _get_group_conns(self, type_group: str) -> List[Conn]:
        conn_list = []
        for line in self._groups.get(type_group, []):
            conn_list.extend(self._parse_line(line))
        return conn_list

    def _parse_comm_config(self):
        for line in self._groups.get("config", []):
            if line.startswith("private_key="):
                key_path = line[len("private_key="):]
                if key_path.startswith(('"', "'")) and key_path.endswith(('"', "'")):
                    key_path = key_path[1:-1].strip()
                self._comm_config.update({"private_key": key_path})
