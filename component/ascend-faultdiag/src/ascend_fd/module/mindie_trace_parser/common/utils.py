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
import os

from ascend_fd.utils.status import FileOpenError
from ascend_fd.utils.tool import safe_read_open, safe_walk


def read_file_lines(file_path):
    try:
        with safe_read_open(file_path, "r", encoding="utf8") as f:
            return f.readlines()
    except FileOpenError:
        return []


def get_dir_files(dir_path):
    res = []
    for root, _, files in safe_walk(dir_path):
        for file in files:
            res.append(os.path.join(root, file))
    return res
