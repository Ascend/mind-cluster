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

from ply.lex import LexToken


class TokenType:
    COP = 0
    LOP = 1
    VALUE = 2


class Token(LexToken):

    def __init__(self, value, lex_type, token_type):
        super().__init__()
        self.value = value
        self.type = lex_type
        self.token_type = token_type

    def __str__(self):
        return str(self.value) + "@" + str(self.token_type)

    def __repr__(self):
        return str(self)


class NumberToken(Token):

    def __init__(self, value, lex_type, token_type):
        super().__init__(value, lex_type, token_type)
        self.value = self._to_num(value)

    @staticmethod
    def _to_num(num_str):
        if num_str[-1] in ['f', 'F', 'd', 'D']:
            return float(num_str[:-1])
        elif '.' in num_str:
            return float(num_str)
        elif num_str[-1] in ['l', 'L']:
            return int(num_str[:-1])
        else:
            return int(num_str)
