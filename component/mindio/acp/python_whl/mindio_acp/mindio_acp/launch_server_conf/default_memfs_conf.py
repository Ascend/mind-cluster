#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
from typing import Dict

default_server_info: Dict = {
    # memfs
    'memfs.data_block_pool_capacity_in_gb': '128',
    'memfs.data_block_size_in_mb': '128',
    'memfs.write.parallel.enabled': 'true',
    'memfs.write.parallel.thread_num': '16',
    'memfs.write.parallel.slice_in_mb': '16',

    # background
    'background.backup.thread_num': '32',
}
