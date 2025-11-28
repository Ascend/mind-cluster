#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
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

import os
import setuptools
from setuptools import find_namespace_packages
from setuptools.dist import Distribution

# 消除whl压缩包的时间戳差异
os.environ['SOURCE_DATE_EPOCH'] = '0'

current_version = os.getenv('BUILD_VERSION', '7.3.0')


class BinaryDistribution(Distribution):
    """Distribution which always forces a binary package with platform name"""

    def has_ext_modules(foo):
        return True


setuptools.setup(
    name="mindio_acp",
    version=current_version,
    author="",
    author_email="",
    description="python api for mindio_acp",
    packages=find_namespace_packages(exclude=("tests*",)),
    url="",
    license="",
    python_requires=">=3.7",
    package_data={"mindio_acp": ["_c2python_api.so", "lib/**", "bin/**", "launch_server_conf/**", "VERSION"]},
    distclass=BinaryDistribution
)
