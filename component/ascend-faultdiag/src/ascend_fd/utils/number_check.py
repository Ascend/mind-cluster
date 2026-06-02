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


class NumberCheck:
    """Utility class for number validation checks"""

    @staticmethod
    def is_non_negative_integer(value) -> bool:
        """
        Check if value represents a non-negative integer (greater than or equal to 0)
        :param value: string or other type to check
        :return: True if value can be converted to int and >= 0, False otherwise
        """
        if not value or not str(value).strip():
            return False
        try:
            int_value = int(str(value).strip())
            return int_value >= 0
        except (ValueError, TypeError):
            return False
