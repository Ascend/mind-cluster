#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import setuptools
from setuptools.dist import Distribution

# 消除whl压缩包的时间戳差异
os.environ['SOURCE_DATE_EPOCH'] = '0'
version = os.environ.get('BUILD_VERSION', '7.3.0')


class BinaryDistribution(Distribution):
    """Distribution which always forces a binary package with platform name"""
    def has_ext_modules(self):
        return True


setuptools.setup(
    name="mindio_ttp",
    version=version,
    author="",
    author_email="",
    description="python api for mindio ttp",
    packages=['mindio_ttp'],
    url="",
    license="",
    python_requires=">=3.7",
    package_data={
        "mindio_ttp": [
            "framework_ttp/**",
            "controller_ttp/**",
            "utils/**",
            "mindspore_api/**"
        ]
    },
    distclass=BinaryDistribution
)