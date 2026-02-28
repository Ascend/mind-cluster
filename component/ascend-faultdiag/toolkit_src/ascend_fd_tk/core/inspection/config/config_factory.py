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

from ascend_fd_tk.core.common import diag_enum
from ascend_fd_tk.core.inspection.config import mayi_config
from ascend_fd_tk.core.inspection.config.base import InspectionConfig


class InspectionConfigFactory:
    _CONFIG_MAPPING = {
        diag_enum.Customer.Mayi: mayi_config.MayiConfig,
    }

    @classmethod
    def get_inspection_config(cls, customer: diag_enum.Customer) -> InspectionConfig:
        return cls._CONFIG_MAPPING[customer]
