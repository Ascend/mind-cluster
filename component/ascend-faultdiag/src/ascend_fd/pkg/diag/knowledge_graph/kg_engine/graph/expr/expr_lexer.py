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
import logging
import re
from typing import List

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_token import Token, TokenType, NumberToken

kg_logger = logging.getLogger("KG_ENGINE")


# 组合正则：按优先级顺序，长匹配优先
_TOKEN_RE = re.compile(
    r'(?P<SKIP>[ \t]+)'
    r'|(?P<STRING>"(?:[^"\\]|\\.)*")'
    r'|(?P<COP_P7>!contains|!in\b|contains|startWith|startwith|endWith|endwith|in\b|out\b|!=|==|=)'
    r'|(?P<COP_P6><=|>=|<|>)'
    r'|(?P<LOP_AND>&&|and(?=[\s)$"]))'
    r'|(?P<LOP_OR>\|\||or(?=[\s)$"]))'
    r'|(?P<LOP_NOT>!|not(?=[\s)$"(]))'
    r'|(?P<BRACKET_OPEN>\()'
    r'|(?P<BRACKET_CLOSE>\))'
    r'|(?P<NUMBER>-?\d+(?:\.\d+)?[dDfFLl]?)'
    r'|(?P<VARIABLE>(?:src|dest)\.[a-zA-Z0-9_]+)'
    r'|(?P<VALUE>\w+)'
)


def _handle_string_token(value: str) -> str:
    """去除双引号并处理转义字符。"""
    inner = value[1:-1]
    result = []
    i = 0
    while i < len(inner):
        if inner[i] == '\\' and i + 1 < len(inner):
            result.append(inner[i + 1])
            i += 2
        else:
            result.append(inner[i])
            i += 1
    return ''.join(result)


_SKIP = "SKIP"
_STRING = "STRING"
_VALUE = "VALUE"
_NUMBER = "NUMBER"
_COP_P6 = "COP_P6"
_COP_P7 = "COP_P7"
_LOP_AND = "LOP_AND"
_LOP_OR = "LOP_OR"
_LOP_NOT = "LOP_NOT"


def tokenize(text: str) -> List[Token]:
    """将输入字符串解析为 Token 列表。"""
    tokens = []
    pos = 0
    while pos < len(text):
        m = _TOKEN_RE.match(text, pos)
        if not m:
            kg_logger.error("Lexer illegal character '%s'", text[pos])
            pos += 1
            continue
        pos = m.end()
        token_type = m.lastgroup
        if token_type == _SKIP:
            continue
        value = m.group()
        if token_type == _STRING:
            value = _handle_string_token(value)
            tokens.append(Token(value, _VALUE, TokenType.VALUE))
        elif token_type == _NUMBER:
            tokens.append(NumberToken(value, token_type, TokenType.VALUE))
        elif token_type in (_COP_P6, _COP_P7):
            tokens.append(Token(value, token_type, TokenType.COP))
        elif token_type in (_LOP_AND, _LOP_OR, _LOP_NOT):
            tokens.append(Token(value, token_type, TokenType.LOP))
        else:
            tokens.append(Token(value, token_type, TokenType.VALUE))
    return tokens


def get_lexer():
    """返回 tokenize 函数，兼容旧接口。"""
    return tokenize
