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
from typing import List

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_holder import (
    Comparison,
    SingleExpression,
    DoubleExpression,
    VarHolder,
    ValueHolder,
)
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_token import Token

kg_logger = logging.getLogger("KG_ENGINE")


class ParseError(Exception):
    pass


class ExprParser:
    """递归下降解析器。

    语法规则（优先级从低到高）：
        expression -> or_expr
        or_expr    -> and_expr (LOP_OR and_expr)*
        and_expr   -> not_expr (LOP_AND not_expr)*
        not_expr   -> LOP_NOT not_expr | primary
        primary    -> BRACKET_OPEN expression BRACKET_CLOSE | comparison
        comparison -> value (COP_P6 | COP_P7) value
        value      -> VARIABLE | VALUE | NUMBER
    """

    def __init__(self, tokens: List[Token]):
        self._tokens = tokens
        self._pos = 0

    def _peek(self):
        if self._pos < len(self._tokens):
            return self._tokens[self._pos]
        return None

    def _consume(self, expected_type: str = None):
        token = self._peek()
        if token is None:
            raise ParseError("Unexpected end of input")
        if expected_type and token.type != expected_type:
            raise ParseError("Expected %s but got %s('%s')" % (expected_type, token.type, token.value))
        self._pos += 1
        return token

    def parse(self):
        result = self._parse_or()
        if self._peek() is not None:
            raise ParseError("trailing tokens after expression")
        return result

    def _parse_or(self):
        left = self._parse_and()
        while self._peek() and self._peek().type == 'LOP_OR':
            op = self._consume()
            right = self._parse_and()
            left = DoubleExpression(left, right, op.value)
        return left

    def _parse_and(self):
        left = self._parse_not()
        while self._peek() and self._peek().type == 'LOP_AND':
            op = self._consume()
            right = self._parse_not()
            left = DoubleExpression(left, right, op.value)
        return left

    def _parse_not(self):
        if self._peek() and self._peek().type == 'LOP_NOT':
            op = self._consume()
            expr = self._parse_not()
            return SingleExpression(expr, op.value)
        return self._parse_primary()

    def _parse_primary(self):
        if self._peek() and self._peek().type == 'BRACKET_OPEN':
            self._consume()
            expr = self._parse_or()
            self._consume('BRACKET_CLOSE')
            return expr
        return self._parse_comparison()

    def _parse_comparison(self):
        left = self._parse_value()
        peek = self._peek()
        if peek and peek.type in ('COP_P6', 'COP_P7'):
            op = self._consume()
            right = self._parse_value()
            return Comparison(left, right, op.value)
        raise ParseError("Expected comparison operator, got %s" % (peek.type if peek else 'EOF'))

    def _parse_value(self):
        token = self._peek()
        if token is None:
            raise ParseError("Unexpected end of input")
        if token.type == 'VARIABLE':
            self._consume()
            return VarHolder(token.value)
        if token.type in ('VALUE', 'NUMBER'):
            self._consume()
            return ValueHolder(token.value)
        raise ParseError("Unexpected token type '%s' with value '%s'" % (token.type, token.value))


class Parser:
    """解析器包装类，兼容旧接口。支持错误恢复：当遇到语法错误时，
    丢弃当前部分解析结果，跳过错误 token 和逻辑运算符，尝试重新解析后续表达式。
    """

    def parse(self, input_str: str, lexer=None):
        if lexer is None:
            from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_lexer import tokenize

            lexer = tokenize
        tokens = lexer(input_str)
        return self._parse_with_recovery(tokens)

    @staticmethod
    def _parse_with_recovery(tokens: List[Token]):
        # 第一次尝试完整解析
        parser = ExprParser(tokens)
        try:
            return parser.parse()
        except ParseError:
            kg_logger.error("Parser: Syntax error")

        # 错误恢复：跳过出错部分，从逻辑运算符之后重新解析
        pos = 0
        while pos < len(tokens):
            if tokens[pos].type in ('LOP_AND', 'LOP_OR'):
                pos += 1  # 跳过逻辑运算符
                if pos < len(tokens):
                    sub_parser = ExprParser(tokens[pos:])
                    try:
                        return sub_parser.parse()
                    except ParseError:
                        pass
            pos += 1
        return None


def get_parser():
    """返回 Parser 实例，兼容旧接口。"""
    return Parser()
