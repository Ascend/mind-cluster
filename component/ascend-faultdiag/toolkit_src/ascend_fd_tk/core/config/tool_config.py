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

from importlib.metadata import version, PackageNotFoundError

from ascend_fd_tk.core.common.constants import PACKAGE_NAME
from ascend_fd_tk.utils.logger import DIAG_LOGGER


class ToolConfig:
    def __init__(self, version_info: str = "v0.10"):
        self.version = self.get_version() or version_info

    @staticmethod
    def get_version():
        try:
            return version(PACKAGE_NAME)
        except PackageNotFoundError:
            DIAG_LOGGER.warning("包 %s 未安装", PACKAGE_NAME)
            return ""
        except Exception as e:
            DIAG_LOGGER.warning("获取版本号时发生未知错误: %s", str(e))
            return ""
