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

from ascend_fd.wrapper.pretty_table import PrettyTable, Style


class TableTestCase(unittest.TestCase):

    def setUp(self) -> None:
        self.table = PrettyTable()
        self.table.title = "Ascend Fault-Diag Report"
        self.table.field_names = ["Info", "Type", "Version"]
        self.table.add_row(["Ascend", "Fault-Diag", "7.0.RC1"])

    def test_common(self):
        ######################
        # 创建一个表格
        ######################
        table = PrettyTable()
        # 表格的标题
        table.title = "Ascend Fault-Diag Report"
        # 表格的字段名，即表格的列标题
        table.field_names = ["Info", "Type", "Version"]
        # 一次性添加多行
        table.add_rows([["Ascend", "Fault-Diag", "7.0.RC1"], ["Ascend", "Dmi", "7.0.RC1"]])
        # 添加单行。添加行时，若未定义field_names，或自定创建field_names：Field 1,Field 2,Field 3...
        table.add_row(["Ascend", "Deployer", "7.0.RC1"], divider=True)
        # 添加单行。该行后展示分割线
        table.add_row(["Ascend", "Operator", "123456\n789abcdefghijk"], divider=True)
        table.add_row(["Ascend", "Other", "7.0.RC1"])
        ######################
        # 设置表格风格
        ######################
        style = Style(
            vertical_char="|",  # 绘制表格垂直线的单个字符
            horizontal_char="-",  # 绘制表格水平线的单个字符
            junction_char="+",  # 绘制表格线交汇点的单个字符
            align={"r": {"Info"}, "c": {"Type"}, "l": {"Version"}},  # 根据field_names设置每列的对齐方式
            default_align="c",  # 默认的对齐方式。可一次性设置所有列的对齐方式。title的对齐方式为default_align
            start=0,  # 对输出行进行切片，头索引
            end=None,  # 对输出行进行切片，尾索引
            left_border=True,  # 是否展示左边框
            right_border=True,  # 是否展示右边框
            inner_border=True,  # 是否展示内部边框
            top_border=True,  # 是否展示上边框
            bottom_border=True,  # 是否展示下边框
            max_width={"Version": 10},  # 根据field_names设置每列的最大列宽
            dividers=[]  # 行后是否展示分割线
        )
        table.style = style
        expect_result = ('+----------------------------------+\n'
                         '|     Ascend Fault-Diag Report     |\n'
                         '+--------+------------+------------+\n'
                         '|   Info |    Type    | Version    |\n'
                         '+--------+------------+------------+\n'
                         '| Ascend | Fault-Diag | 7.0.RC1    |\n'
                         '| Ascend |    Dmi     | 7.0.RC1    |\n'
                         '| Ascend |  Deployer  | 7.0.RC1    |\n'
                         '+--------+------------+------------+\n'
                         '| Ascend |  Operator  | 123456     |\n'
                         '|        |            | 789abcdefg |\n'
                         '|        |            | hijk       |\n'
                         '+--------+------------+------------+\n'
                         '| Ascend |   Other    | 7.0.RC1    |\n'
                         '+--------+------------+------------+')
        self.assertEqual(table.get_string(), expect_result)

    def test_long_title(self):
        table = PrettyTable()
        table.title = "Ascend Fault-Diag Report-12345678910"
        table.field_names = ["Info", "Type", "Version"]
        table.add_row(["Ascend", "Fault-Diag", "7.0.RC1"])
        table.add_row(["Ascend", "Operator", "123456\n789abcdefghijk"], divider=True)
        style = Style(
            align={"l": {"Version"}},
            max_width={"Version": 10}
        )
        table.style = style
        expect_result = ('+----------------------------------+\n'
                         '| Ascend Fault-Diag Report-1234567 |\n'
                         '|               8910               |\n'
                         '+--------+------------+------------+\n'
                         '|  Info  |    Type    | Version    |\n'
                         '+--------+------------+------------+\n'
                         '| Ascend | Fault-Diag | 7.0.RC1    |\n'
                         '| Ascend |  Operator  | 123456     |\n'
                         '|        |            | 789abcdefg |\n'
                         '|        |            | hijk       |\n'
                         '+--------+------------+------------+')
        self.assertEqual(table.get_string(), expect_result)

    def test_long_header(self):
        table = PrettyTable()
        table.title = "Ascend Fault-Diag Report"
        table.field_names = ["Info", "Type", "Version-12345678910"]
        table.add_row(["Ascend", "Fault-Diag", "7.0.RC1"])
        table.add_row(["Ascend", "Operator", "123456\n789abcdefghijk"], divider=True)
        style = Style(
            align={"l": {"Version-12345678910"}},
            max_width={"Version-12345678910": 10}
        )
        table.style = style
        expect_result = ('+----------------------------------+\n'
                         '|     Ascend Fault-Diag Report     |\n'
                         '+--------+------------+------------+\n'
                         '|  Info  |    Type    | Version-12 |\n'
                         '|        |            | 345678910  |\n'
                         '+--------+------------+------------+\n'
                         '| Ascend | Fault-Diag | 7.0.RC1    |\n'
                         '| Ascend |  Operator  | 123456     |\n'
                         '|        |            | 789abcdefg |\n'
                         '|        |            | hijk       |\n'
                         '+--------+------------+------------+')
        self.assertEqual(table.get_string(), expect_result)

    def test_cz(self):
        table = PrettyTable()
        table.title = "Ascend Fault-Diag Report"
        table.field_names = ["版本信息", "Type", "版本"]
        table.add_row(["", "Fault-Diag", "7.0.RC1"], divider=True)
        table.add_row(["根因节点分析", "类型", "描述"])
        table.add_row(["", "根因节点", "['worker-0 device-1']"])
        expect_result = ('+---------------------------------------------------+\n'
                         '|             Ascend Fault-Diag Report              |\n'
                         '+--------------+------------+-----------------------+\n'
                         '|   版本信息   |    Type    |         版本          |\n'
                         '+--------------+------------+-----------------------+\n'
                         '|              | Fault-Diag |        7.0.RC1        |\n'
                         '+--------------+------------+-----------------------+\n'
                         '| 根因节点分析 |    类型    |         描述          |\n'
                         "|              |  根因节点  | ['worker-0 device-1'] |\n"
                         '+--------------+------------+-----------------------+')
        self.assertEqual(table.get_string(), expect_result)

    def test_left_right_border(self):
        style = Style(
            left_border=False,
            right_border=False
        )
        self.table.style = style
        expect_result = ('+-------------------------------+\n'
                         '    Ascend Fault-Diag Report     \n'
                         '+--------+------------+---------+\n'
                         '   Info  |    Type    | Version  \n'
                         '+--------+------------+---------+\n'
                         '  Ascend | Fault-Diag | 7.0.RC1  \n'
                         '+--------+------------+---------+')
        self.assertEqual(self.table.get_string(), expect_result)

    def test_top_bottom_border(self):
        style = Style(
            left_border=True,
            right_border=True,
            top_border=False,
            bottom_border=False
        )
        self.table.style = style
        expect_result = ('|   Ascend Fault-Diag Report    |\n'
                         '+--------+------------+---------+\n'
                         '|  Info  |    Type    | Version |\n'
                         '+--------+------------+---------+\n'
                         '| Ascend | Fault-Diag | 7.0.RC1 |')
        self.assertEqual(self.table.get_string(), expect_result)

    def test_inner_border(self):
        style = Style(
            left_border=True,
            right_border=True,
            top_border=True,
            bottom_border=True,
            inner_border=False
        )
        self.table.add_row(["Ascend", "Other", "7.0.RC1"])
        self.table.style = style
        expect_result = ('+-------------------------------+\n'
                         '|   Ascend Fault-Diag Report    |\n'
                         '+--------+------------+---------+\n'
                         '|  Info       Type      Version |\n'
                         '+--------+------------+---------+\n'
                         '| Ascend   Fault-Diag   7.0.RC1 |\n'
                         '| Ascend     Other      7.0.RC1 |\n'
                         '+--------+------------+---------+')
        self.assertEqual(self.table.get_string(), expect_result)

    def test_type_border(self):
        table = PrettyTable()
        table.title = "Ascend Fault-Diag Report"
        table.field_names = ["Info", "Type", "Version"]
        table.add_row(["Ascend", "Fault-Diag", "7.0.RC1"], divider=True)
        table.add_row(["Ascend", "Other", "7.0.RC1"])
        style = Style(
            horizontal_char='^',
            vertical_char='>',
            junction_char='~'
        )
        table.style = style
        expect_result = ('~^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^~\n'
                         '>   Ascend Fault-Diag Report    >\n'
                         '~^^^^^^^^~^^^^^^^^^^^^~^^^^^^^^^~\n'
                         '>  Info  >    Type    > Version >\n'
                         '~^^^^^^^^~^^^^^^^^^^^^~^^^^^^^^^~\n'
                         '> Ascend > Fault-Diag > 7.0.RC1 >\n'
                         '~^^^^^^^^~^^^^^^^^^^^^~^^^^^^^^^~\n'
                         '> Ascend >   Other    > 7.0.RC1 >\n'
                         '~^^^^^^^^~^^^^^^^^^^^^~^^^^^^^^^~')
        self.assertEqual(table.get_string(), expect_result)

    def test_escape_cha(self):
        def format_rows(title_name, indicator_name, description):
            description = repr(description).strip("'\"")
            return [title_name, indicator_name, description]

        table = PrettyTable()
        table.title = "Ascend Fault-Diag Report"
        table.field_names = ["Info", "Type", "Version"]
        table.add_row(format_rows("Ascend", "Fault-Diag", "7\t1\v1\\w\x41"), divider=True)
        table.add_row(["Ascend", "Other", "7.0.RC1"])
        expect_result = ('+-------------------------------------+\n'
                         '|      Ascend Fault-Diag Report       |\n'
                         '+--------+------------+---------------+\n'
                         '|  Info  |    Type    |    Version    |\n'
                         '+--------+------------+---------------+\n'
                         '| Ascend | Fault-Diag | 7\\t1\\x0b1\\\\wA |\n'
                         '+--------+------------+---------------+\n'
                         '| Ascend |   Other    |    7.0.RC1    |\n'
                         '+--------+------------+---------------+')
        self.assertEqual(table.get_string(), expect_result)

    def test_omit_show(self):
        table = PrettyTable()
        table.title = ""
        table.field_names = ["Info", "Type", "Version"]
        table.add_row(["Ascend", "Fault-Diag", "123456789abcdefghijk"], divider=True)
        table.add_row(["Ascend", "Other", "7.0.RC1"])
        style = Style(
            max_omit_show=10,
            max_width={"Version": 10}
        )
        table.style = style
        expect_result = ('+--------+------------+------------+\n'
                         '|  Info  |    Type    |  Version   |\n'
                         '+--------+------------+------------+\n'
                         '| Ascend | Fault-Diag | 123456789a |\n'
                         '|        |            |    ...     |\n'
                         '+--------+------------+------------+\n'
                         '| Ascend |   Other    |  7.0.RC1   |\n'
                         '+--------+------------+------------+')
        self.assertEqual(table.get_string(), expect_result)

    def test_only_title(self):
        table = PrettyTable()
        table.title = "Ascend Fault-Diag Report"
        expect_result = ('+--------------------------+\n'
                         '| Ascend Fault-Diag Report |\n'
                         '+--------------------------+')
        self.assertEqual(table.get_string(), expect_result)

    def test_only_header(self):
        table = PrettyTable()
        table.field_names = ["Info", "Type", "Version"]
        expect_result = ('+------+------+---------+\n'
                         '| Info | Type | Version |\n'
                         '+------+------+---------+')
        self.assertEqual(table.get_string(), expect_result)

    def test_only_row(self):
        table = PrettyTable()
        table.add_row(["Ascend", "Fault-Diag", "7.0.RC1"])
        expect_result = ('+---------+------------+---------+\n'
                         '| Field 1 |  Field 2   | Field 3 |\n'
                         '+---------+------------+---------+\n'
                         '| Ascend  | Fault-Diag | 7.0.RC1 |\n'
                         '+---------+------------+---------+')
        self.assertEqual(table.get_string(), expect_result)

    def test_only1_row(self):
        table = PrettyTable()
        table.title = " "
        table.field_names = [" ", " ", " "]
        table.add_row([" ", " ", " "])
        expect_result = ('+--------------+\n'
                         '|              |\n'
                         '+----+----+----+\n'
                         '|    |    |    |\n'
                         '+----+----+----+\n'
                         '|    |    |    |\n'
                         '+----+----+----+')
        self.assertEqual(table.get_string(), expect_result)

    def tearDown(self) -> None:
        pass
