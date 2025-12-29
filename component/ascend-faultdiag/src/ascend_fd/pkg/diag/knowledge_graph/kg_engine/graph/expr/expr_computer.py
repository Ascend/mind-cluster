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
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.vertex import Vertex
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_condition import ExprCondition


class ExprComputer:
    SRC_KEY = "src"
    DEST_KEY = "dest"

    def __init__(self, src_event_v: Vertex, dest_event_v: Vertex):
        self._param_map = {
            self.SRC_KEY: src_event_v,
            self.DEST_KEY: dest_event_v
        }

    def compute(self, cond: ExprCondition) -> bool:
        if not cond:
            return False
        return cond.apply(self._param_map)
