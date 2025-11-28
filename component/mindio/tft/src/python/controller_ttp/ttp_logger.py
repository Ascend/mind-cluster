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
import inspect
import os
import threading
import traceback
from enum import Enum

import ttp_c2python_api


class LEVEL(Enum):
    DEBUG = 0
    INFO = 1
    WARNING = 2
    ERROR = 3

TID = 100000


class Logger:
    @staticmethod
    def format(msg, *args, **kwargs):
        frame = inspect.currentframe().f_back
        info = inspect.getframeinfo(frame.f_back)
        prefix = f"[{threading.get_ident()%TID}][PYH {os.path.basename(info.filename)}:{info.lineno}] {msg}"
        return prefix % args

    def debug(self, msg, *args, **kwargs):
        ttp_c2python_api.log(LEVEL.DEBUG.value, self.format(msg, *args, **kwargs))

    def info(self, msg, *args, **kwargs):
        ttp_c2python_api.log(LEVEL.INFO.value, self.format(msg, *args, **kwargs))

    def warning(self, msg, *args, **kwargs):
        ttp_c2python_api.log(LEVEL.WARNING.value, self.format(msg, *args, **kwargs))

    def error(self, msg, *args, **kwargs):
        ttp_c2python_api.log(LEVEL.ERROR.value, self.format(msg, *args, **kwargs))

    def exception(self, msg, *args, **kwargs):
        ttp_c2python_api.log(LEVEL.ERROR.value, self.format(msg, *args, **kwargs)+"\n"+traceback.format_exc())

LOGGER = Logger()