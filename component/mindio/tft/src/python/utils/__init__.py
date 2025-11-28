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
import sys

current_path = os.path.abspath(__file__)
sys.path.append(os.path.dirname(current_path))

from .uce_utils import tft_can_do_uce_repair, tft_set_update_start_time, tft_set_update_end_time, get_l2_hbm_error_time

__all__ = [
    "tft_can_do_uce_repair",
    "tft_set_update_start_time",
    "tft_set_update_end_time",
]