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

from ttp_mindx_api import (tft_notify_controller_dump, tft_notify_controller_stop_train, tft_register_mindx_callback,
                           tft_notify_controller_on_global_rank, tft_notify_controller_change_strategy,
                           tft_notify_controller_prepare_action)
from .ttp_controller import (tft_init_controller, tft_start_controller, tft_destroy_controller,
                             tft_query_high_availability_switch)


__all__ = [
    "tft_notify_controller_dump",
    "tft_notify_controller_stop_train",
    "tft_notify_controller_on_global_rank",
    "tft_notify_controller_prepare_action",
    "tft_notify_controller_change_strategy",
    "tft_register_mindx_callback",
    "tft_query_high_availability_switch"
]
