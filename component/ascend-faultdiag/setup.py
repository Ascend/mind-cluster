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

from setuptools import setup, find_packages

from diag_tool.core.config.tool_config import ToolConfig


def get_platform():
    system_info = platform.machine() or 'unknown'
    target_plat = {
        'AMD64': 'x86_64'
    }
    res_plat = target_plat.get(system_info)
    return res_plat or system_info


setup(
    name="mindcluster_diag_tool",
    version=ToolConfig().version,
    description="MindCluster device link diagnostic tool",
    author="Huawei Technologies Co., Ltd",
    url="https://gitcode.com/Ascend/mind-cluster",
    packages=find_packages(),
    include_package_data=True,
    package_data={
        '': ['*.ini', '*.json'],
    },
    install_requires=[],
    entry_points={
        'console_scripts': [
            'mindcluster_diag_tool=diag_tool.cli:main',
        ],
    },
    platforms=get_platform(),
    python_requires='>=3.7',
)
