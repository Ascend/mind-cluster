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

import platform
import subprocess
from typing import Tuple, Optional


class PingTool:
    @staticmethod
    def ping(host: str, count: int = 4, timeout: int = 5) -> Tuple[bool, Optional[str]]:
        """
        跨平台 ping 工具
        :param host: 目标主机（IP或域名）
        :param count: 发送包数量
        :param timeout: 超时时间（秒）
        :return: (是否可达, 结果摘要)
        """
        # 根据系统设置 ping 参数
        os_name = platform.system().lower()

        if os_name == "windows":
            # Windows: ping -n 次数 -w 超时毫秒
            param = ['ping', '-n', str(count), '-w', str(timeout * 1000), host]
        elif os_name in ["linux", "darwin"]:  # darwin 是 macOS
            # Linux/macOS: ping -c 次数 -W 超时秒数
            param = ['ping', '-c', str(count), '-W', str(timeout), host]
        else:
            return False, f"不支持的操作系统: {os_name}"

        try:
            # 执行 ping 命令，捕获输出
            result = subprocess.run(
                param,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                timeout=timeout + 2  # 总超时比单个包超时多2秒
            )

            # 解析结果（Windows 成功返回 0，Linux 成功返回 0）
            is_success = result.returncode == 0

            # 提取关键信息（如丢包率）
            output = result.stdout
            if os_name == "windows":
                # Windows 结果示例："丢失 = 0 (0% 丢失)"
                loss_line = [line for line in output.splitlines() if "丢失" in line or "Lost" in line][-1]
            else:
                # Linux 结果示例："4 packets transmitted, 4 received, 0% packet loss"
                loss_line = [line for line in output.splitlines() if "packet loss" in line][0]

            return is_success, loss_line

        except subprocess.TimeoutExpired:
            return False, f"超时（{timeout}秒）"
        except Exception as e:
            return False, f"执行失败: {str(e)}"
