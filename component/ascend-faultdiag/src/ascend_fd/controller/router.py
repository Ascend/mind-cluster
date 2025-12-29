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
import logging
import os

from ascend_fd.pkg import start_blacklist_job
from ascend_fd.pkg.customize import start_entity_job
from ascend_fd.controller.controller import ParseController, DiagController, SingleDiagController
from ascend_fd.configuration.config import RUN_LOG_FORMAT
from ascend_fd.pkg.customize.custom_config import start_config_job

echo = logging.getLogger("ECHO")


def router(args):
    """
    Perform parsing or diagnostic tasks based on command-line arguments
    :param args: the command-line arguments
    """
    if args.cmd in ["parse", "diag", "single-diag"]:
        echo.info("The %s job starts. Please wait. Job id: [%s], run log file is [%s].",
                  args.cmd, args.task_id, RUN_LOG_FORMAT.format(os.getpid()))
        _controller_func(args)
        echo.info(f"The %s job is complete.", args.cmd)
        return
    if args.cmd == "blacklist":
        start_blacklist_job(args)
        return
    if args.cmd == "entity":
        start_entity_job(args)
    if args.cmd == "config":
        start_config_job(args)


def _controller_func(args):
    """
    Parse and diag cmd function
    :param args: the command-line arguments
    """
    if args.cmd == "parse":
        controller = ParseController(args)
    elif args.cmd == "diag":
        controller = DiagController(args)
    else:
        controller = SingleDiagController(args)
    controller.start_job()
