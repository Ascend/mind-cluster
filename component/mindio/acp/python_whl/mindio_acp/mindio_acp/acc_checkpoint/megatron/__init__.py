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

# get the environment variable: MINDIO_AUTO_PATCH_MEGATRON
# If MINDIO_AUTO_PATCH_MEGATRON is 1 or true, then auto patch megatron.
# Otherwise, disable auto patch megatron.
if str(os.getenv('MINDIO_AUTO_PATCH_MEGATRON', "false")).strip().lower() in ["1", "true"]:
    from .megatron_patch import exec_adaptation
    exec_adaptation()
