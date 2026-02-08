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

from typing import List

from diag_tool.core.common.json_obj import JsonObj


class Domain(JsonObj):

    def __init__(self, domain_type, value):
        self.value = value
        self.domain_type = domain_type

    def __str__(self):
        return f"{self.domain_type}:{self.value}"


class DiagResult(JsonObj):

    def __init__(self, domain: List[Domain], fault_info: str="", suggestion: str="", err_code=""):
        self.domain = domain
        self.fault_info = fault_info
        self.suggestion = suggestion
        self.err_code = err_code

    def get_domain_desc(self):
        if isinstance(self.domain, list):
            return "->".join([str(item) for item in self.domain])
        return str(self.domain)

    def to_dict(self):
        return {
            "故障域": str(self.get_domain_desc()),
            "故障码": self.err_code,
            "故障信息": self.fault_info,
            "处理建议": self.suggestion,
        }
