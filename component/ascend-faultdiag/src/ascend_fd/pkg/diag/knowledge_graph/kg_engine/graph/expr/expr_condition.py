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

from typing import Dict

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.vertex import Vertex
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_compiler import ExprCompiler


class ExprCondition:

    def __init__(self, expr: str, compiler: ExprCompiler):
        self.expr = expr
        self.holder = compiler.compile(expr)

    def apply(self, element: Dict[str, Vertex]):
        return self.holder.eval(element)
