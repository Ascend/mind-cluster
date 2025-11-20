#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import enum


class ProductType(enum.Enum):
    """Super pod type code for all device forms. Standard card use its main board id."""

    SERVER_8P = 0
    POD_1D = 1
    POD_2D = 2
    SERVER_16P = 3
    STANDARD_1P = 104
    STANDARD_2P = 106
    STANDARD_4P = 108
