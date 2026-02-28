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

from ascend_fd_tk.core.common.path import CommonPath


class DumpLogDirConfig:

    def __init__(self, bmc_dump_log_dir=CommonPath.PROJECT_BMC_DUMP_CACHE_DIR,
                 host_dump_log_dir=CommonPath.HOST_DUMP_DIR_CACHE_DIR,
                 switch_dump_log_dir=CommonPath.SWI_DUMP_LOG_CACHE):
        self.bmc_dump_log_dir = bmc_dump_log_dir
        self.host_dump_log_dir = host_dump_log_dir
        self.switch_dump_log_dir = switch_dump_log_dir
