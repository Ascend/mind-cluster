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
from setuptools import setup, find_packages
from mindcluster_tools import __version__


install_requires = ["bitarray", "netifaces"]


setup(
    name="mindcluster_tools",
    version=__version__,
    packages=find_packages(),
    description="",
    author="mindcluster",
    python_requires=">=3.6",
    install_requires=install_requires,
    entry_points={
        "console_scripts": [
            "mindcluster-tools=mindcluster_tools.tools_parser:main",
        ]
    },
)
