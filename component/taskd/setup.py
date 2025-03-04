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
"""
build config
"""
import os
import stat
import glob
import shutil

from setuptools import setup, find_packages
from setuptools.command.build_py import build_py

build_folder = ('build/bdist*', 'build/lib')
cache_folder = ('taskd.egg-info', '_package_output')

pwd = os.path.dirname(os.path.relpath(__file__))
pkg_dir = os.path.join(pwd, "build/lib")
version = os.getenv("BUILD_VERSION", "v7.0.RC1")


def _write_version(file):
    file.write(f"__version__ = '{version}'\n")


def build_dependencies():
    """generate python file"""
    version_file = os.path.join(pkg_dir, 'taskd', 'version.py')
    version_file_dir = os.path.join(pkg_dir, 'taskd')
    if not os.path.exists(version_file):
        os.makedirs(version_file_dir, exist_ok=True)

    with os.fdopen(os.open(version_file, os.O_WRONLY | os.O_CREAT, mode=stat.S_IRUSR | stat.S_IWUSR), 'w') as f:
        _write_version(f)


def clean():
    for folder in cache_folder:
        if os.path.exists(folder):
            shutil.rmtree(folder)
    for pattern in build_folder:
        for name in glob.glob(pattern):
            if os.path.exists(name):
                shutil.rmtree(name)


def get_required_packages():
    with open(os.path.join(pwd, 'requirements.txt'), encoding='UTF-8') as f:
        lines = f.readlines()
    lines = [line.strip('\n') for line in lines]
    return lines

clean()
build_dependencies()
required_packages = get_required_packages()

package_data = {
    '':
        ['fault_checker/**',
         'python/**',
         '*.yaml',
        ]
}

setup(
    name='taskd',
    version=version,
    platforms=['linux', ],
    description='Ascend MindCluster taskd is a new library for training management',
    python_requires='>=3.7',
    install_requires=required_packages,
    ext_requires={"torch":["torch"]},
    package_data=package_data,
    packages=find_packages(exclude=["**test**"]),
    include_package_data=True,
    cmdclass={
        'build_py': build_py,
    }
)

clean()