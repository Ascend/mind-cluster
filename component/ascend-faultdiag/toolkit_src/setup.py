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

import argparse
import sys

from setuptools import setup, find_packages


def parse_args():
    parser = argparse.ArgumentParser(add_help=False)
    parser.add_argument("--version", "-v")
    args, remain_args = parser.parse_known_args()
    sys.argv = [sys.argv[0]] + remain_args
    return args.version


version = parse_args()

setup(
    name="ascend-faultdiag-toolkit",
    version=version,
    description="MindCluster ascend faultdiag diagnostic toolkit",
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
            'ascend-fd-tk=toolkit.cli:main',
        ],
    },
    python_requires='>=3.8',
)
