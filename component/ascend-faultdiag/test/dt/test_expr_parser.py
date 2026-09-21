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

import unittest

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_lexer import tokenize
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_parser import ExprParser, Parser, ParseError
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_token import TokenType
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_holder import (
    Comparison,
    SingleExpression,
    DoubleExpression,
    VarHolder,
    ValueHolder,
)
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_compiler import ExprCompiler
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.vertex import Vertex
from ascend_fd.utils.load_kg_config import EntityAttribute


class TestLexer(unittest.TestCase):
    """词法分析器单元测试"""

    def _assert_token(self, token, value, lex_type, token_type):
        self.assertEqual(token.value, value)
        self.assertEqual(token.type, lex_type)
        self.assertEqual(token.token_type, token_type)

    def test_number_integer(self):
        tokens = tokenize("123")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 123, 'NUMBER', TokenType.VALUE)

    def test_number_negative(self):
        tokens = tokenize("-42")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], -42, 'NUMBER', TokenType.VALUE)

    def test_number_float(self):
        tokens = tokenize("3.14")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 3.14, 'NUMBER', TokenType.VALUE)

    def test_number_with_suffix_d(self):
        tokens = tokenize("1d")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 1.0, 'NUMBER', TokenType.VALUE)

    def test_number_with_suffix_f(self):
        tokens = tokenize("2.5f")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 2.5, 'NUMBER', TokenType.VALUE)

    def test_number_with_suffix_l(self):
        tokens = tokenize("100L")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 100, 'NUMBER', TokenType.VALUE)

    def test_variable_src(self):
        tokens = tokenize("src.count")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 'src.count', 'VARIABLE', TokenType.VALUE)

    def test_variable_dest(self):
        tokens = tokenize("dest.source")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 'dest.source', 'VARIABLE', TokenType.VALUE)

    def test_variable_with_underscore(self):
        tokens = tokenize("src.device_id")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 'src.device_id', 'VARIABLE', TokenType.VALUE)

    def test_value_plain(self):
        tokens = tokenize("nothing")
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 'nothing', 'VALUE', TokenType.VALUE)

    def test_string_value(self):
        tokens = tokenize('"hello world"')
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 'hello world', 'VALUE', TokenType.VALUE)

    def test_string_with_escape(self):
        tokens = tokenize(r'"hello \"world\""')
        self.assertEqual(len(tokens), 1)
        self._assert_token(tokens[0], 'hello "world"', 'VALUE', TokenType.VALUE)

    def test_cop_p6_lt(self):
        tokens = tokenize("1 < 2")
        self.assertEqual(len(tokens), 3)
        self._assert_token(tokens[1], '<', 'COP_P6', TokenType.COP)

    def test_cop_p6_le(self):
        tokens = tokenize("1 <= 2")
        self._assert_token(tokens[1], '<=', 'COP_P6', TokenType.COP)

    def test_cop_p6_gt(self):
        tokens = tokenize("1 > 2")
        self._assert_token(tokens[1], '>', 'COP_P6', TokenType.COP)

    def test_cop_p6_ge(self):
        tokens = tokenize("1 >= 2")
        self._assert_token(tokens[1], '>=', 'COP_P6', TokenType.COP)

    def test_cop_p7_eq(self):
        tokens = tokenize("a == b")
        self._assert_token(tokens[1], '==', 'COP_P7', TokenType.COP)

    def test_cop_p7_assign(self):
        tokens = tokenize("a = b")
        self._assert_token(tokens[1], '=', 'COP_P7', TokenType.COP)

    def test_cop_p7_ne(self):
        tokens = tokenize("a != b")
        self._assert_token(tokens[1], '!=', 'COP_P7', TokenType.COP)

    def test_cop_p7_contains(self):
        tokens = tokenize("a contains b")
        self._assert_token(tokens[1], 'contains', 'COP_P7', TokenType.COP)

    def test_cop_p7_not_contains(self):
        tokens = tokenize("a !contains b")
        self._assert_token(tokens[1], '!contains', 'COP_P7', TokenType.COP)

    def test_cop_p7_in(self):
        tokens = tokenize("a in b")
        self._assert_token(tokens[1], 'in', 'COP_P7', TokenType.COP)

    def test_cop_p7_not_in(self):
        tokens = tokenize("a !in b")
        self._assert_token(tokens[1], '!in', 'COP_P7', TokenType.COP)

    def test_cop_p7_out(self):
        tokens = tokenize("a out b")
        self._assert_token(tokens[1], 'out', 'COP_P7', TokenType.COP)

    def test_cop_p7_startwith(self):
        tokens = tokenize("a startWith b")
        self._assert_token(tokens[1], 'startWith', 'COP_P7', TokenType.COP)

    def test_cop_p7_startwith_lower(self):
        tokens = tokenize("a startwith b")
        self._assert_token(tokens[1], 'startwith', 'COP_P7', TokenType.COP)

    def test_cop_p7_endwith(self):
        tokens = tokenize("a endWith b")
        self._assert_token(tokens[1], 'endWith', 'COP_P7', TokenType.COP)

    def test_lop_and_double(self):
        tokens = tokenize("a && b")
        self._assert_token(tokens[1], '&&', 'LOP_AND', TokenType.LOP)

    def test_lop_and_keyword(self):
        tokens = tokenize("a and b")
        self._assert_token(tokens[1], 'and', 'LOP_AND', TokenType.LOP)

    def test_lop_or_double(self):
        tokens = tokenize("a || b")
        self._assert_token(tokens[1], '||', 'LOP_OR', TokenType.LOP)

    def test_lop_or_keyword(self):
        tokens = tokenize("a or b")
        self._assert_token(tokens[1], 'or', 'LOP_OR', TokenType.LOP)

    def test_lop_not_bang(self):
        tokens = tokenize("! a")
        self._assert_token(tokens[0], '!', 'LOP_NOT', TokenType.LOP)

    def test_lop_not_keyword(self):
        tokens = tokenize("not a")
        self._assert_token(tokens[0], 'not', 'LOP_NOT', TokenType.LOP)

    def test_bracket_open(self):
        tokens = tokenize("(a")
        self._assert_token(tokens[0], '(', 'BRACKET_OPEN', TokenType.VALUE)

    def test_bracket_close(self):
        tokens = tokenize("a)")
        self._assert_token(tokens[1], ')', 'BRACKET_CLOSE', TokenType.VALUE)

    def test_skip_whitespace(self):
        tokens = tokenize("  a   ==   b  ")
        self.assertEqual(len(tokens), 3)
        self._assert_token(tokens[0], 'a', 'VALUE', TokenType.VALUE)
        self._assert_token(tokens[1], '==', 'COP_P7', TokenType.COP)
        self._assert_token(tokens[2], 'b', 'VALUE', TokenType.VALUE)

    def test_and_keyword_not_confused_with_variable(self):
        """and 关键字后面跟空格或括号时才算 LOP_AND，否则算 VALUE"""
        tokens = tokenize("android == ios")
        self.assertEqual(len(tokens), 3)
        self._assert_token(tokens[0], 'android', 'VALUE', TokenType.VALUE)

    def test_empty_input(self):
        tokens = tokenize("")
        self.assertEqual(len(tokens), 0)


class TestParserAST(unittest.TestCase):
    """解析器 AST 结构单元测试"""

    def _parse(self, expr):
        tokens = tokenize(expr)
        parser = ExprParser(tokens)
        return parser.parse()

    def test_simple_comparison_eq(self):
        result = self._parse("src.count == dest.count")
        self.assertIsInstance(result, Comparison)
        self.assertIsInstance(result.left, VarHolder)
        self.assertIsInstance(result.right, VarHolder)
        self.assertEqual(result.left.var_name, 'src')
        self.assertEqual(result.left.prop_key, 'count')
        self.assertEqual(result.right.var_name, 'dest')
        self.assertEqual(result.right.prop_key, 'count')
        self.assertEqual(result.compare_type, '==')

    def test_comparison_with_number(self):
        result = self._parse("1 > 0.1")
        self.assertIsInstance(result, Comparison)
        self.assertIsInstance(result.left, ValueHolder)
        self.assertIsInstance(result.right, ValueHolder)
        self.assertEqual(result.left.value, 1)
        self.assertEqual(result.right.value, 0.1)
        self.assertEqual(result.compare_type, '>')

    def test_comparison_with_string(self):
        result = self._parse('dest.source == "every thing"')
        self.assertIsInstance(result, Comparison)
        self.assertIsInstance(result.right, ValueHolder)
        self.assertEqual(result.right.value, 'every thing')

    def test_comparison_contains(self):
        result = self._parse("src.num contains dest.num")
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.compare_type, 'contains')

    def test_comparison_not_contains(self):
        result = self._parse("src.num !contains dest.num")
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.compare_type, '!contains')

    def test_comparison_in(self):
        result = self._parse("src.num in dest.num")
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.compare_type, 'in')

    def test_comparison_not_in(self):
        result = self._parse("src.num !in dest.num")
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.compare_type, '!in')

    def test_comparison_out(self):
        result = self._parse("src.num out dest.num")
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.compare_type, 'out')

    def test_comparison_startwith(self):
        result = self._parse("src.num startWith dest.num")
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.compare_type, 'startwith')

    def test_comparison_endwith(self):
        result = self._parse("src.num endWith dest.num")
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.compare_type, 'endwith')

    def test_logical_and(self):
        result = self._parse("1 <= 2.3 and 2 > 1")
        self.assertIsInstance(result, DoubleExpression)
        self.assertEqual(result.lop_type, 'and')
        self.assertIsInstance(result.left, Comparison)
        self.assertIsInstance(result.right, Comparison)

    def test_logical_and_double_amp(self):
        result = self._parse("2 <= src.num && 2 <= dest.num")
        self.assertIsInstance(result, DoubleExpression)
        self.assertEqual(result.lop_type, '&&')

    def test_logical_or(self):
        result = self._parse("1 > 2.3 or 2 < 1")
        self.assertIsInstance(result, DoubleExpression)
        self.assertEqual(result.lop_type, 'or')

    def test_logical_not(self):
        result = self._parse("! 1 > 2.3")
        self.assertIsInstance(result, SingleExpression)
        self.assertEqual(result.log_type, '!')
        self.assertIsInstance(result.exp, Comparison)

    def test_logical_not_keyword(self):
        result = self._parse("not 1 > 2.3")
        self.assertIsInstance(result, SingleExpression)
        self.assertEqual(result.log_type, 'not')

    def test_bracket_grouping(self):
        result = self._parse("(1 > 2.3 or 2 > 1) and (1 == 1 or 2 > 3)")
        self.assertIsInstance(result, DoubleExpression)
        self.assertEqual(result.lop_type, 'and')
        self.assertIsInstance(result.left, DoubleExpression)
        self.assertIsInstance(result.right, DoubleExpression)

    def test_nested_brackets(self):
        result = self._parse("((1 > 2))")
        self.assertIsInstance(result, Comparison)

    def test_precedence_and_before_or(self):
        """a or b and c 应解析为 a or (b and c)"""
        result = self._parse("1 > 2.3 || 2 > 1 && 1 == 1")
        self.assertIsInstance(result, DoubleExpression)
        self.assertEqual(result.lop_type, '||')
        self.assertIsInstance(result.right, DoubleExpression)
        self.assertEqual(result.right.lop_type, '&&')

    def test_not_precedence(self):
        """! a && b 应解析为 (!a) && b"""
        result = self._parse("! src.count <= dest.count && src.count > 0")
        self.assertIsInstance(result, DoubleExpression)
        self.assertEqual(result.lop_type, '&&')
        self.assertIsInstance(result.left, SingleExpression)

    def test_negative_number(self):
        result = self._parse("-1 > -2")
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.left.value, -1)
        self.assertEqual(result.right.value, -2)

    def test_number_with_suffix(self):
        result = self._parse("1d <= 2.3f")
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.left.value, 1.0)
        self.assertEqual(result.right.value, 2.3)

    def test_complex_expression(self):
        result = self._parse('(src.count > 0 && dest.count > 0) || src.source == "nothing"')
        self.assertIsInstance(result, DoubleExpression)
        self.assertEqual(result.lop_type, '||')
        self.assertIsInstance(result.left, DoubleExpression)
        self.assertIsInstance(result.right, Comparison)


class TestParserErrorHandling(unittest.TestCase):
    """解析器错误处理单元测试"""

    def test_parse_error_on_incomplete(self):
        """不完整的表达式应抛出 ParseError"""
        tokens = tokenize("src.count ==")
        parser = ExprParser(tokens)
        with self.assertRaises(ParseError):
            parser.parse()

    def test_parse_error_on_trailing(self):
        """多余 token 应抛出 ParseError"""
        tokens = tokenize("a == b c")
        parser = ExprParser(tokens)
        with self.assertRaises(ParseError):
            parser.parse()

    def test_parse_error_no_comparison(self):
        """只有 value 没有比较运算符应抛出 ParseError"""
        tokens = tokenize("src.count")
        parser = ExprParser(tokens)
        with self.assertRaises(ParseError):
            parser.parse()

    def test_parse_error_unmatched_bracket(self):
        """括号不匹配应抛出 ParseError"""
        tokens = tokenize("(a == b")
        parser = ExprParser(tokens)
        with self.assertRaises(ParseError):
            parser.parse()

    def test_parse_error_empty(self):
        """空输入应抛出 ParseError"""
        tokens = tokenize("")
        parser = ExprParser(tokens)
        with self.assertRaises(ParseError):
            parser.parse()


class TestParserWrapper(unittest.TestCase):
    """Parser 包装类（错误恢复）单元测试"""

    def setUp(self):
        self.parser = Parser()

    def test_normal_parse(self):
        result = self.parser.parse("a == b")
        self.assertIsInstance(result, Comparison)

    def test_error_recovery_skip_to_and(self):
        """错误恢复：跳过错误部分，从 && 之后重新解析"""
        result = self.parser.parse("a == b c && d == e")
        self.assertIsInstance(result, Comparison)
        self.assertIsInstance(result.left, ValueHolder)
        self.assertEqual(result.left.value, 'd')

    def test_error_recovery_skip_to_or(self):
        """错误恢复：跳过错误部分，从 || 之后重新解析"""
        result = self.parser.parse("x y z || a == b")
        self.assertIsInstance(result, Comparison)
        self.assertIsInstance(result.left, ValueHolder)
        self.assertEqual(result.left.value, 'a')

    def test_error_recovery_all_invalid(self):
        """所有部分都无效时返回 None"""
        result = self.parser.parse("x y z")
        self.assertIsNone(result)

    def test_wrapper_with_lexer_param(self):
        """测试传入 lexer 参数"""
        result = self.parser.parse("a == b", lexer=tokenize)
        self.assertIsInstance(result, Comparison)


class TestExprEval(unittest.TestCase):
    """表达式求值端到端测试"""

    def setUp(self):
        self.compiler = ExprCompiler()
        src_event = Vertex(
            "",
            {
                "event_id": "",
                "count": 1,
                "source": "nothing",
                "num": 2,
                "source_device": "2",
                "name": "hello",
            },
            EntityAttribute({}),
        )
        dest_event = Vertex(
            "",
            {
                "event_id": "",
                "count": 1,
                "source": "every thing",
                "num": 3,
                "source_device": "2",
                "name": "world",
            },
            EntityAttribute({}),
        )
        self.param = {
            "src": src_event,
            "dest": dest_event,
        }

    # -- 比较运算符 --
    def test_eq_true(self):
        self.assertTrue(self.compiler.compile('aaa == aaa').eval({}))
        self.assertTrue(self.compiler.compile('src.count == dest.count').eval(self.param))

    def test_eq_false(self):
        self.assertFalse(self.compiler.compile('aaa == 1').eval({}))
        self.assertFalse(self.compiler.compile('src.count == dest.source').eval(self.param))

    def test_ne_via_not_equals(self):
        """!= 运算符验证：_COMPARE_FUNCS 不支持 !=，但 parser 可正确解析 AST"""
        result = self.compiler.compile('3 != 0')
        self.assertIsInstance(result, Comparison)
        self.assertEqual(result.compare_type, '!=')

    def test_gt(self):
        self.assertTrue(self.compiler.compile('1 > 0.1').eval({}))
        self.assertTrue(self.compiler.compile('dest.num > src.num').eval(self.param))

    def test_gte(self):
        self.assertTrue(self.compiler.compile('1 >= 1').eval({}))
        self.assertTrue(self.compiler.compile('src.count >= src.count').eval(self.param))

    def test_lt(self):
        self.assertTrue(self.compiler.compile('1 < 2.3').eval({}))
        self.assertTrue(self.compiler.compile('src.num < dest.num').eval(self.param))

    def test_lte(self):
        self.assertTrue(self.compiler.compile('1 <= 1').eval({}))
        self.assertTrue(self.compiler.compile('src.count <= dest.count').eval(self.param))

    def test_contains(self):
        self.assertTrue(self.compiler.compile('src.source contains "thing"').eval(self.param))

    def test_not_contains(self):
        self.assertTrue(self.compiler.compile('src.source !contains "xyz"').eval(self.param))

    def test_in(self):
        self.assertTrue(self.compiler.compile('"thing" in src.source').eval(self.param))

    def test_not_in(self):
        self.assertTrue(self.compiler.compile('"xyz" !in src.source').eval(self.param))

    def test_out(self):
        self.assertTrue(self.compiler.compile('"xyz" out src.source').eval(self.param))

    def test_startwith(self):
        self.assertTrue(self.compiler.compile('src.name startWith "he"').eval(self.param))

    def test_endwith(self):
        self.assertTrue(self.compiler.compile('src.name endWith "lo"').eval(self.param))

    # -- 逻辑运算符 --
    def test_and(self):
        self.assertTrue(self.compiler.compile('1 <= 2.3 and 2 > 1').eval({}))
        self.assertTrue(self.compiler.compile('2 <= src.num && 2 <= dest.num').eval(self.param))

    def test_or(self):
        self.assertTrue(self.compiler.compile('1 > 2.3 or 2 > 1').eval({}))
        self.assertTrue(self.compiler.compile('src.count > src.num || src.count <= dest.count').eval(self.param))

    def test_not(self):
        self.assertTrue(self.compiler.compile('! 1 > 2.3').eval({}))
        self.assertTrue(self.compiler.compile('not 1 > 2.3').eval({}))

    def test_not_with_and(self):
        self.assertTrue(self.compiler.compile('! src.count > dest.count && src.count > 0').eval(self.param))

    # -- 复杂表达式 --
    def test_bracket(self):
        self.assertTrue(self.compiler.compile('1 > 2.3 or (2 > 1 and 1 == 1)').eval({}))

    def test_complex_nested(self):
        self.assertTrue(self.compiler.compile('(1 > 2.3 or 2 > 1) and (1 == 1 or 2 > 3)').eval({}))

    def test_precedence_and_before_or(self):
        self.assertTrue(self.compiler.compile('1 > 2.3 || 2 > 1 && 1 == 1').eval({}))

    def test_not_with_or(self):
        self.assertTrue(
            self.compiler.compile('not (src.source == dest.source || src.count > dest.count)').eval(self.param)
        )

    def test_complex_with_variable(self):
        self.assertTrue(
            self.compiler.compile('(src.count > 0 && dest.count > 0) || src.source == "nothing"').eval(self.param)
        )


if __name__ == '__main__':
    unittest.main()
