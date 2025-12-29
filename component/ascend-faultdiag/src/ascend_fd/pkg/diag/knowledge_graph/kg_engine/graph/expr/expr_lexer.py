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

import ply

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_token import Token, TokenType, NumberToken

kg_logger = logging.getLogger("KG_ENGINE")


class LexerError(Exception):
    pass


class ExprLexer:
    tokens = [
        'BRACKET_OPEN',
        'BRACKET_CLOSE',
        'LOP_NOT',
        'COP_P6',
        'COP_P7',
        'LOP_AND',
        'LOP_OR',
        'NUMBER',
        'VARIABLE',
        'VALUE'
    ]

    states = (
        ("doublequote", "exclusive"),
    )

    t_BRACKET_OPEN = r'\('
    t_BRACKET_CLOSE = r'\)'
    # 忽略空格和制表符
    t_ignore = ' \t'
    t_doublequote_ignore = ''

    # 定义每个词的正则表达式
    @staticmethod
    def t_LOP_NOT(token):
        r'\!|not(?=\s)'
        return Token(token.value, token.type, TokenType.LOP)

    @staticmethod
    def t_COP_P6(token):
        r'<=|<|>=|>'
        return Token(token.value, token.type, TokenType.COP)

    @staticmethod
    def t_COP_P7(token):
        r'!=|==|=|\!contains|contains|in|out|\!in|startWith|startwith|endWith|endwith'
        return Token(token.value, token.type, TokenType.COP)

    @staticmethod
    def t_LOP_AND(token):
        r'\&\&|and(?=\s)'
        return Token(token.value, token.type, TokenType.LOP)

    @staticmethod
    def t_LOP_OR(token):
        r'\|\||or(?=\s)'
        return Token(token.value, token.type, TokenType.LOP)

    @staticmethod
    def t_NUMBER(token):
        r'-?\d{1,200}(\.\d{1,200})?[dDfFLl]?'
        return NumberToken(token.value, token.type, TokenType.VALUE)

    @staticmethod
    def t_VARIABLE(token):
        r'(src|dest)\.[a-zA-Z0-9_]{1,200}'
        return Token(token.value, token.type, TokenType.VALUE)

    @staticmethod
    def t_VALUE(token):
        r'\w{1,200}'
        return Token(token.value, token.type, TokenType.VALUE)

    @staticmethod
    def t_doublequote(token):
        r'"'
        token.lexer.push_state('doublequote')
        token.lexer.string_start = token.lexer.lexpos
        token.lexer.string_value = ''

    @staticmethod
    def t_doublequote_escape(token):
        r'\\.'
        token.lexer.string_value += token.value[1]

    @staticmethod
    def t_doublequote_content(token):
        r'[^"\\]{1,200}'
        token.lexer.string_value += token.value

    @staticmethod
    def t_doublequote_end(token):
        r'"'
        token.lexer.pop_state()
        token.value = token.lexer.string_value
        token.lexer.string_value = None
        return Token(token.value, 'VALUE', TokenType.VALUE)

    @staticmethod
    def t_doublequote_error(token):
        raise LexerError(
            'Error on line %s, col %s while lexing backquoted operator: Unexpected character: %s ' % (
                token.lexer.lineno, token.lexpos - token.lexer.latest_newline, token.value[0]))

    # 可选的错误处理函数
    @staticmethod
    def t_error(token):
        kg_logger.error("Lexer illegal character '%s'", token.value[0])
        token.lexer.skip(1)


def get_lexer():
    expr_lexer = ExprLexer()
    return ply.lex.lex(object=expr_lexer)
