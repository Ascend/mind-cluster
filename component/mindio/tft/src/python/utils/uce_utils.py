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
import re
from ..controller_ttp import ttp_logger
mind_spore = os.getenv('MINDIO_FOR_MINDSPORE', 'False')
mind_spore = mind_spore.lower() in ('true', '1')
if not mind_spore:
    import torch
    import torch_npu


TRAIN_STEP_TIME = {}
TIME_US_PATTERN = r"time us=(\d+)"


def tft_set_update_start_time(start_time: int = None):
    """
    save optimizer update start time
    start_time: int: if None, only for pytorch
    """
    start_time = get_event_time() if start_time is None else start_time
    TRAIN_STEP_TIME["START_TIME"] = start_time


def tft_set_update_end_time(end_time: int = None):
    """
    save optimizer update end time
    end_time: int: if None, only for pytorch
    """
    end_time = get_event_time() if end_time is None else end_time
    TRAIN_STEP_TIME["END_TIME"] = end_time


def tft_can_do_uce_repair(hbm_error_time: int, start_time: int = None, end_time: int = None) -> bool:
    """
    check can do uce repair
    hbm_error_time: int: the time of hbm error occur
    start_time: int: the time of optimizer start update
    end_time: int: the time of optimizer end update
    """
    start_time = TRAIN_STEP_TIME.get("START_TIME", None) if start_time is None else start_time
    end_time = TRAIN_STEP_TIME.get("END_TIME", None) if end_time is None else end_time

    ttp_logger.LOGGER.info(f"check can do uce repair, hbm error time:{hbm_error_time}, "
                           f"start time:{start_time}, end time:{end_time}, TRAIN_STEP_TIME:{TRAIN_STEP_TIME}")
    if not hbm_error_time:
        return False
    return _can_do_repair(hbm_error_time, start_time, end_time)


def get_event_time():
    event = torch.npu.Event(enable_timing=True)
    event.record()
    event_time = 0
    if hasattr(event, "recorded_time"):
        event_time = event.recorded_time()
    else:
        ttp_logger.LOGGER.debug(f"torch.npu.Event has no attribute 'recorded_time', "
                                f"unable get the time of Event occur, please update torch_npu. "
                                f"default time is {event_time} .")
    return event_time


def get_l2_hbm_error_time(log_string: str):
    match = re.search(TIME_US_PATTERN, log_string)
    if match:
        return int(match.group(1))
    return None


def _can_do_repair(hbm_error_time: int, start_time: int, end_time: int) -> bool:
    if start_time and end_time:  # step > 0
        return max(start_time, end_time) < hbm_error_time or start_time < hbm_error_time < end_time

    return False


def get_update_start_end_time():
    return TRAIN_STEP_TIME.get('START_TIME', None), TRAIN_STEP_TIME.get('END_TIME', None)