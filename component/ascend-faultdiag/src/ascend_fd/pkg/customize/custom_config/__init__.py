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
from ascend_fd.pkg.customize.custom_config.config_info import ConfigInfo, update_config_info, check_config_info, \
    show_config_info


def start_config_job(args):
    """
    Start the custom config job, contain update, show and check cmd
    :param args: the command-line arguments
    """
    if args.update:
        update_config_info(args.update)
        return
    if args.check:
        check_config_info()
        return
    if args.show:
        show_config_info()
