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

from ..controller_ttp import tft_init_controller, tft_start_controller, tft_destroy_controller
from .ttp_decorator import tft_init_processor, tft_start_processor, tft_destroy_processor, tft_pause_train
from .ttp_decorator import (tft_start_updating_os, tft_end_updating_os, tft_exception_handler, tft_set_step_args,
                            tft_start_copy_os, tft_set_optimizer_replica)
from .ttp_decorator import (tft_register_save_ckpt_handler, tft_register_rename_handler, tft_register_exit_handler,
                            tft_register_stop_handler, tft_register_clean_handler, tft_register_repair_handler,
                            tft_register_rollback_handler, tft_set_dp_group_info,
                            tft_register_zit_upgrade_rollback_handler, tft_register_rebuild_group_handler,
                            tft_register_zit_upgrade_repair_handler, tft_register_zit_upgrade_rebuild_handler,
                            tft_register_zit_downgrade_rebuild_handler, tft_register_decrypt_handler)
from .ttp_decorator import (tft_report_error, tft_wait_next_action, tft_get_repair_step, tft_get_repair_type,
                            tft_report_load_ckpt_step)
from .ttp_decorator import tft_is_reboot_node, tft_reset_limit_step, tft_get_reboot_type
from .ttp_decorator import ReportState, Action, OptimizerType, RepairType
from .ttp_decorator import set_mindio_export_version, tft_register_stream_sync_handler
from ..utils import tft_can_do_uce_repair

__all__ = [
    "tft_init_controller",
    "tft_start_controller",
    "tft_destroy_controller",
    "tft_init_processor",
    "tft_start_processor",
    "tft_destroy_processor",
    "tft_start_updating_os",
    "tft_start_copy_os",
    "tft_end_updating_os",
    "tft_set_optimizer_replica",
    "tft_exception_handler",
    "tft_set_step_args",
    "tft_register_rename_handler",
    "tft_register_save_ckpt_handler",
    "tft_register_exit_handler",
    "tft_register_stop_handler",
    "tft_register_clean_handler",
    "tft_register_rebuild_group_handler",
    "tft_register_repair_handler",
    "tft_register_rollback_handler",
    "tft_register_stream_sync_handler",
    "tft_register_zit_upgrade_rollback_handler",
    "tft_register_zit_upgrade_repair_handler",
    "tft_register_zit_upgrade_rebuild_handler",
    "tft_register_zit_downgrade_rebuild_handler",
    "tft_report_error",
    "tft_wait_next_action",
    "tft_get_repair_step",
    "tft_get_repair_type",
    "tft_is_reboot_node",
    "tft_get_reboot_type",
    "tft_can_do_uce_repair",
    "tft_reset_limit_step",
    "tft_pause_train",
    "tft_set_dp_group_info",
    "tft_report_load_ckpt_step",
    "tft_register_decrypt_handler",
    "OptimizerType",
    "Action",
    "ReportState",
    "RepairType",
]
