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

from abc import ABC, abstractmethod
from typing import Dict


class ToDict(ABC):
    """All serialized objects need to inherit from ToDict and implement the to_dict method"""

    @abstractmethod
    def to_dict(self) -> Dict:
        pass


class Queryable(ABC):
    """All objects querying EID need to inherit from Queryable and implement the query method"""

    @abstractmethod
    def query(self, *args, **kwargs) -> Dict:
        pass
