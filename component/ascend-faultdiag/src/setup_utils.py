#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025 Huawei Technologies Co., Ltd
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
import argparse
import shutil
import datetime
import sys
from pathlib import Path
from setuptools import find_packages

from ascend_fd.utils.tool import safe_write_open, safe_read_open


def clean():
    cache_folder = ('ascend_fd.egg-info', "build*")
    for pattern in cache_folder:
        for folder in Path().glob(pattern):
            if folder.exists():
                shutil.rmtree(folder)


def get_version_info(mode):
    src_path = Path(__file__).absolute().parent
    ci_version_file = src_path.parent.parent.parent
    ci_version_file = ci_version_file.joinpath('buildDesign', 'build', 'conf', 'config.yaml')
    version = '6.0.RC2'
    if ci_version_file.exists():
        file_stream = safe_read_open(ci_version_file, 'r')
        for line in file_stream:
            if "pkgversion" in line:
                version = line.split(':')[1].strip()
                break
        file_stream.close()
    if mode == "en":
        root_path = "alan_fd"
    else:
        root_path = "ascend_fd"
    version_file = src_path.joinpath(root_path, "Version.info")
    build_time = datetime.date.today()
    with safe_write_open(version_file, mode="w+") as file_stream:
        file_stream.writelines(f"{version}\n{build_time}")
    return version


def get_platform():
    system_info = platform.machine() or 'unknown'
    target_plat = {
        'AMD64': 'x86_64'
    }
    res_plat = target_plat.get(system_info)
    return res_plat or system_info


def parse_mode():
    parser = argparse.ArgumentParser(add_help=False)
    parser.add_argument("--mode", "-m", choices=["zh", "en"], default="zh")
    args, remain_args = parser.parse_known_args()
    sys.argv = [sys.argv[0]] + remain_args
    return args.mode


def get_setup_config(mode):
    """
    Get step config
    :param mode: Chinese or English mode
    :return: config info
    note: for install_requires in common_config
    If you want to use the performance (-p --performance) detection module,
    install the following modules: "scikit-learn>=1.3.0", "pandas>=1.3.5", "numpy>=1.21.6", "1.5.0>joblib>=1.2.0"
    """
    common_config = {
        "version": get_version_info(mode),
        "package_data": {"": ["**/*.so"]},
        "include_package_data": True,
        "python_requires": ">=3.7",
        "platforms": get_platform(),
        "packages": find_packages(),
        "install_requires": [
            "ply>=3.11"
        ],
    }

    if mode == "en":
        config = {
            "name": "alan-faultdiag",
            "description": "alan fault diag",
            "entry_points": {
                "console_scripts": [
                    "alan-fd=alan_fd.cli:main"
                ]
            }
        }
    else:
        config = {
            "name": "ascend-faultdiag",
            "description": "ascend fault diag",
            "author": "Huawei Technologies Co., Ltd",
            "entry_points": {
                "console_scripts": [
                    "ascend-fd=ascend_fd.cli:main"
                ]
            }
        }

    config.update(common_config)
    return config
