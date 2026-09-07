#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
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
from ascend_fd.utils.constant.ub_const import PRECHECK_PREFIX


def get_device_precheck_event(precheck_info, source_device: str):
    precheck = precheck_info.get(source_device, [])
    # 多个PRECHECK事件以各自的event_code为key，全部放入root_causes
    return {event.get("event_code"): event for event in precheck if event.get("event_code").startswith(PRECHECK_PREFIX)}
