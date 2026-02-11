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

import abc
from typing import List

from toolkit.core.common.diag_enum import Customer
from toolkit.core.model.cluster_info_cache import ClusterInfoCache
from toolkit.core.model.inspection import InspectionErrorItem


class InspectionCheckItem(abc.ABC):

    def __init__(self, cluster_info: ClusterInfoCache, customer: Customer):
        self.customer = customer
        self.cluster_info = cluster_info

    @abc.abstractmethod
    def check(self) -> List[InspectionErrorItem]:
        pass
