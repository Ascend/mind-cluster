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


import time
from datetime import datetime

class DateObj:

    def __init__(self, date_str: str, date_fmt: str):
        self.date_str = date_str
        self.date_fmt = date_fmt
        self._timestamp = None

    @property
    def timestamp(self):
        if self._timestamp is None:
            dt = datetime.strptime(self.date_str, self.date_fmt)
            self._timestamp = time.mktime(dt.timetuple()) + dt.microsecond / 1000000
        return self._timestamp

    def diff_seconds(self, other: 'DateObj') -> float:
        """
        计算两个DateObj对象之间的时间差（以秒为单位）
        返回值为正数表示self的时间晚于other的时间
        返回值为负数表示self的时间早于other的时间
        """
        return self.timestamp - other.timestamp

    