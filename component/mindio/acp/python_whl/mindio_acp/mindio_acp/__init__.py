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
import sys

current_path = os.path.abspath(__file__)
current_dir = os.path.dirname(current_path)
sys.path.append(current_dir)

from .acc_io import initialize, load, save, multi_save, register_checker, preload, convert, flush
from .ms.ms import open_file, create_file
from .acc_checkpoint.framework_acp import CheckpointHelper

__all__ = [
    'initialize',
    'load',
    'save',
    'multi_save',
    'register_checker',
    'convert',
    'preload',
    'open_file',
    'create_file',
    'flush',
]
