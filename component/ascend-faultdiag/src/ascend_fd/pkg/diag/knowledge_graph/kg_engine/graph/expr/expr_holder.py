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

import abc
from typing import Dict

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.vertex import Vertex
from ascend_fd.utils.status import ParamError


class Holder:

    @abc.abstractmethod
    def eval(self, param):
        pass


class VarHolder(Holder):

    def __init__(self, value):
        value_splits = value.split('.')
        self.var_name = value_splits[0]
        self.prop_key = value_splits[1]

    def __str__(self):
        return "%s.%s" % (self.var_name, self.prop_key)

    def __repr__(self):
        return str(self)

    def eval(self, param):
        if isinstance(param, dict):
            return self._eval_context(param)
        raise ParamError("Param type error. It must be a variable or dict.")

    def _eval_context(self, context: Dict[str, Vertex]):
        var = context.get(self.var_name)
        return None if not var else var.eval(self.prop_key)


class ValueHolder(Holder):

    def __init__(self, value):
        self.value = value

    def eval(self, param):
        return self.value


class Comparison(Holder):
    _COMPARE_FUNCS = {
        '>': lambda a, b: a > b,
        '<': lambda a, b: a < b,
        '>=': lambda a, b: a >= b,
        '<=': lambda a, b: a <= b,
        '=': lambda a, b: a == b,
        '==': lambda a, b: a == b,
        'startwith': lambda a, b: a and str(a).startswith(str(b)),
        'endwith': lambda a, b: a and str(a).endswith(str(b)),
        'contains': lambda a, b: a and b in a,
        '!contains': lambda a, b: a and b not in a,
        'in': lambda a, b: b and a in b,
        '!in': lambda a, b: b and a not in b,
        'out': lambda a, b: b and a not in b,
    }

    def __init__(self, left, right, compare_type):
        self.left = left
        self.right = right
        self.compare_type = compare_type.lower()

    def eval(self, param):
        if self.compare_type not in self._COMPARE_FUNCS:
            return False
        new_left = self.left.eval(param)
        new_right = self.right.eval(param)
        return self._COMPARE_FUNCS.get(self.compare_type)(new_left, new_right)


_LOP_FUNCS = {
    '!': lambda a: not a,
    'not': lambda a: not a,
    '||': lambda a, b: a or b,
    'or': lambda a, b: a or b,
    '&&': lambda a, b: a and b,
    'and': lambda a, b: a and b
}


class DoubleExpression(Holder):

    def __init__(self, left: Holder, right: Holder, lop_type):
        self.left = left
        self.right = right
        self.lop_type = lop_type

    def eval(self, param):
        if self.lop_type not in _LOP_FUNCS:
            return False
        new_left = self.left.eval(param)
        new_right = self.right.eval(param)
        return _LOP_FUNCS.get(self.lop_type)(new_left, new_right)


class SingleExpression(Holder):

    def __init__(self, exp: Holder, lop_type):
        self.exp = exp
        self.log_type = lop_type

    def eval(self, param):
        if self.log_type not in _LOP_FUNCS:
            return False
        res = self.exp.eval(param)
        return _LOP_FUNCS.get(self.log_type)(res)
