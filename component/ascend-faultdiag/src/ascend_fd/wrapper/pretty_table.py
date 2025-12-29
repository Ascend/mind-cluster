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
import itertools
import unicodedata
from typing import Dict, Optional, List, Set
from dataclasses import dataclass, field


class DefaultStyle:
    vertical_char = "|"
    horizontal_char = "-"
    junction_char = "+"
    default_align = "c"
    start = 0
    end = None
    left_border = True
    right_border = True
    inner_border = True
    top_border = True
    bottom_border = True
    align = {}
    max_width = {}
    dividers = []
    max_omit_show = 3000


@dataclass
class Style:
    vertical_char: str = DefaultStyle.vertical_char
    horizontal_char: str = DefaultStyle.horizontal_char
    junction_char: str = DefaultStyle.junction_char
    default_align: str = DefaultStyle.default_align
    start: int = DefaultStyle.start
    end: Optional[int] = DefaultStyle.end
    left_border: bool = DefaultStyle.left_border
    right_border: bool = DefaultStyle.right_border
    inner_border: bool = DefaultStyle.inner_border
    top_border: bool = DefaultStyle.top_border
    bottom_border: bool = DefaultStyle.bottom_border
    align: Dict[str, Set[str]] = field(default_factory=dict)
    max_width: Dict[str, int] = field(default_factory=dict)
    dividers: List[bool] = field(default_factory=list)
    max_omit_show: int = DefaultStyle.max_omit_show

    def __post_init__(self):
        validation_rules = {
            "default_align": self._validate_default_align,
            "vertical_char": self._validate_single_char,
            "horizontal_char": self._validate_single_char,
            "junction_char": self._validate_single_char
        }
        for attr, validator in validation_rules.items():
            validator(attr, getattr(self, attr))

    @staticmethod
    def _validate_default_align(attr, val):
        """
        Validate default align
        :param attr: attribute name
        :param val: attribute value
        """
        check_key = set(val) - {"l", "c", "r"}
        if check_key:
            raise ValueError(f"The '{attr}' attribute is not supported: {list(check_key)}")

    @staticmethod
    def _validate_single_char(attr: str, val: str):
        """
        Validate single char
        :param attr: attribute name
        :param val: attribute value
        """
        if sum(get_char_width(val)) != 1:
            raise ValueError(f"Invalid value for {attr}. Must be a string of length 1.")

    @staticmethod
    def _validate_value_base_header(value: set, field_names: list):
        """
        Validate attribute values based on table headers
        :param value: attribute name
        :param field_names: table header data
        """
        check_value = value - set(field_names)
        if check_value:
            raise ValueError(f"Table headers do not include: {list(check_value)}")

    @staticmethod
    def _validate_int(attr: str, val: int):
        """
        Validate positive integers
        :param attr: attribute name
        :param val: attribute value
        """
        if val <= 0:
            raise ValueError(f"The '{attr}' attribute must be an integer greater than or equal to 0.")

    @staticmethod
    def _align_string(text: str, width: int, align: str = 'c') -> str:
        """
        Align Strings
        :param text: text
        :param width: width after alignment
        :param align: alignment method
        :return: aligned text
        """
        half_pad = 2
        actual_width = sum(get_char_width(text))
        if actual_width >= width:
            return text
        if align == 'r':
            return ' ' * (width - actual_width) + text
        if align == 'l':
            return text + ' ' * (width - actual_width)
        if align == 'c':
            pad = (width - actual_width) // half_pad
            return ' ' * pad + text + ' ' * (width - actual_width - pad)
        raise ValueError("The 'align' attribute supports only these three types: ['l', 'c', 'r'].")

    @staticmethod
    def _custom_textwrap(text: str, width: int) -> str:
        """
        Custom textwrap. If the line width exceeds the maximum width, start a new line
        :param text: text
        :param width: row width of each colum
        :return: adjusted text
        """
        lines = []
        current_line = []
        current_width = 0
        for char, char_width in zip(text, get_char_width(text)):
            if char == '\n':
                lines.append(''.join(current_line))
                current_line = []
                current_width = 0
                continue
            # if the current line width plus the character width exceeds the maximum width, start a new line.
            if current_width + char_width > width:
                lines.append(''.join(current_line))
                current_line = [char]
                current_width = char_width
            else:
                current_line.append(char)
                current_width += char_width
        if current_line:
            lines.append(''.join(current_line))
        return '\n'.join(lines)

    def format_title(self, title: str, field_names: list, column_widths: list) -> list:
        """
        Formatting the title data
        :param title: title
        :param field_names: table header data
        :param column_widths: row width of each colum
        :return: formatted title data
        """
        width_vert_separator = len(f" {self.vertical_char} ")
        lines = []
        if self.top_border:
            lines.append(self.format_border(column_widths, "top"))
        widths = [sum(column_widths) + width_vert_separator * (len(column_widths) - 1)]
        lines.extend(self.format_single_row([title], field_names, widths, row_type="title"))
        lines.append(self.format_border(column_widths))
        return lines

    def format_border(self, column_widths: list, border_type: str = "common") -> str:
        """
        Formatting borderlines
        :param column_widths: row width of each colum
        :param border_type: borderlines type: 'common' or 'top'
        :return: formatted borderlines
        """
        width_fill_space = 2  # fill a space on the left and right of the cell data
        if border_type not in ("common", "top"):
            raise ValueError("'border_type' supports only these two types: ['top', 'common'].")
        join_str = self.horizontal_char if border_type == "top" else self.junction_char
        return self.junction_char + \
            join_str.join([self.horizontal_char * (width + width_fill_space) for width in column_widths]) + \
            self.junction_char

    def format_rows(self, rows: list, field_names: list, column_widths: list) -> list:
        """
        Formatting rows
        :param rows: rows of the table
        :param field_names: table header data
        :param column_widths: rows width of each colum
        :return: formatted rows
        """
        lines = []
        spec_rows, spec_dividers = self._get_specify_row(rows)
        if not spec_rows or not spec_dividers:
            return lines
        separator = self.format_border(column_widths)
        for index, row in enumerate(spec_rows):
            lines.extend(self.format_single_row(row, field_names, column_widths))
            if not spec_dividers[index]:
                continue
            if index != len(spec_dividers) - 1:
                lines.append(separator)
        lines.append(separator)
        return lines

    def format_single_row(self, row: list, field_names: list, column_widths: list, row_type: str = "common") -> list:
        """
        Formatting a single row
        :param row: a single row
        :param field_names: table header data
        :param column_widths: row width of each colum
        :param row_type: row type: 'common' or 'title'
        :return: formatted single row
        e.g:
            field_names: ['A', 'B', 'C']
            row: ['', '说明情况', '1.xxx2.xxx3.xxx']
            self.max_width: {'B': 4, 'C': 5}
        """
        left_bottom = self.vertical_char if self.left_border else " "
        right_bottom = self.vertical_char if self.right_border else " "
        inner_bottom = f" {self.vertical_char} " if self.inner_border else "   "
        # wrap when the maximum width is exceeded.
        wrapped_row = self._set_wrapped_row(row, column_widths)
        if not wrapped_row:
            return wrapped_row
        # align data in each cell
        result = []
        align = self._get_align(field_names)
        for row in wrapped_row:
            formatted_columns = []
            if not field_names:
                for cell, width in zip(row, column_widths):
                    formatted_columns.append(self._align_string(cell, width, self.default_align))
                result.append(left_bottom + " " + inner_bottom.join(formatted_columns) + " " + right_bottom)
                continue
            for cell, col_data, width in zip(row, field_names, column_widths):
                align_type = self.default_align if row_type == "title" else align.get(col_data, self.default_align)
                formatted_columns.append(self._align_string(cell, width, align_type))
            result.append(left_bottom + " " + inner_bottom.join(formatted_columns) + " " + right_bottom)
        return result

    def get_max_width(self, field_names: list, col_data: str) -> int:
        """
        Get the maximum width
        :param field_names: table header data
        :param col_data: table header data of a column
        :return: maximum width
        """
        self._validate_value_base_header(set(self.max_width.keys()), field_names)
        for val in self.max_width.values():
            self._validate_int("max_width", val)
        return self.max_width.get(col_data, 0)

    def _get_dividers(self, rows: list) -> list:
        """
        Get dividers
        :param rows: rows
        :return: dividers
        """
        if len(self.dividers) != len(rows):
            raise ValueError(
                "Dividers list has incorrect number of values, "
                f"(actual) len({len(self.dividers)})!=len({len(rows)}) (expected)."
            )
        return self.dividers

    def _get_align(self, field_names: list) -> dict:
        """
        Get align
        :param field_names: table header data
        :return: align
        """
        self._validate_default_align("align", self.align.keys())
        # check whether all values are in filed_name.
        self._validate_value_base_header(set(itertools.chain(*self.align.values())), field_names)
        align_all = {}
        for align, fields in self.align.items():
            for col_data in fields:
                align_all[col_data] = align
        return align_all

    def _get_specify_row(self, rows: list) -> tuple:
        """
        Get the specified row data
        :param rows: rows
        :return: the specified row data, the specified dividers
        """
        spec_rows = rows[self.start: self.end] if self.end else rows[self.start:]
        dividers = self._get_dividers(rows)
        spec_dividers = dividers[self.start: self.end] if self.end else dividers[self.start:]
        return spec_rows, spec_dividers

    def _set_wrapped_row(self, row: list, column_widths: list) -> list:
        """
        Set wrapped row
        :param row: a single row
        :param column_widths: row width of each colum
        :return: wrapped rows
        """
        wrapped_row = []  # e.g: ['', '说明\n情况', '1.xxx\n2.xxx\n3.xxx']
        for cell, max_width in zip(row, column_widths):
            cell = str(cell)
            if not max_width or sum(get_char_width(cell)) <= max_width:
                wrapped_row.append(cell)
                continue
            line = self._custom_textwrap(cell, max_width)
            wrapped_row.append(line)
        if not wrapped_row:
            return wrapped_row
        # split a single line into multiple lines by '\n'
        max_height = max(len(item.split('\n')) for item in wrapped_row)
        multi_wrapped_row = []  # e.g: [['', '说明', '1.xxx'], ['', '情况', '2.xxx'], ['', '', '3.xxx']]
        for height in range(max_height):
            split_rows = []
            for item in wrapped_row:
                parts = item.split('\n')
                split_rows.append(parts[height] if height < len(parts) else '')
            multi_wrapped_row.append(split_rows)
        return multi_wrapped_row


class PrettyTable:
    def __init__(self, field_names=None) -> None:
        """
        Return a new PrettyTable instance
        :param field_names: table header
        """
        self._field_names: List[str] = field_names if field_names else []
        self._rows: List[List[str]] = []
        self._title: str = ""
        self._style: Style = Style()

    def __str__(self) -> str:
        return self.get_string()

    @property
    def field_names(self) -> list:
        """
        Table header
        :return: table title
        """
        return self._field_names

    @field_names.setter
    def field_names(self, field_names: List[str]):
        if self._field_names and len(self._field_names) != len(field_names):
            raise ValueError(
                "Field name list has incorrect number of values, "
                f"(actual) {len(self._field_names)}!={len(field_names)} (expected)."
            )
        self._field_names = field_names[:]

    @property
    def title(self) -> str:
        """
        Table title
        :return: table title
        """
        return self._title

    @title.setter
    def title(self, val: str):
        self._title = val

    @property
    def style(self) -> Style:
        """
        Table style
        :return: table style
        """
        return self._style

    @style.setter
    def style(self, style: Style):
        for key, value in style.__dict__.items():
            if value != DefaultStyle.__dict__.get(key):
                setattr(self._style, key, value)

    def add_rows(self, rows: list):
        """
        Adding multiple rows of data
        :param rows: rows data
        """
        for row in rows:
            self.add_row(row)

    def add_row(self, row: list, divider: bool = False):
        """
        Add a single row of data
        :param row: row data
        :param divider: whether to set a separator line
        """
        if self._field_names and len(row) != len(self._field_names):
            raise ValueError("Row must have the same number of elements as field_names.")
        if not self._field_names:
            self._field_names = [f"Field {n + 1}" for n in range(len(row))]
        self._rows.append([str(cell) for cell in row])
        self.style.dividers.append(divider)

    def get_string(self) -> str:
        """
        Get the string format of a table
        :return: the string format of a table
        """
        self._set_omit_cell()
        lines = []
        column_widths = self._get_column_widths()
        separator = self.style.format_border(column_widths)
        # add title
        if self._title:
            lines.extend(self.style.format_title(self._title, self._field_names, column_widths))
        # add header
        if self._field_names:
            if not self._title and self.style.top_border:
                lines.append(separator)
            lines.extend(self.style.format_single_row(self._field_names, self._field_names, column_widths))
            lines.append(separator)
        # add rows
        lines.extend(self.style.format_rows(self._rows, self._field_names, column_widths))
        if lines and lines[-1] == separator and not self.style.bottom_border:
            lines.pop(-1)
        return "\n".join(lines)

    def _set_omit_cell(self):
        """
        Set omit cell
        """
        spec_rows = []
        for row in self._rows:
            spec_cell = []
            for cell in row:
                if sum(get_char_width(cell)) > self.style.max_omit_show:
                    spec_cell.append(cell[:self.style.max_omit_show] + "...")
                    continue
                spec_cell.append(cell)
            spec_rows.append(spec_cell)
        self._rows = spec_rows[:]

    def _get_column_widths(self) -> list:
        """
        Get the maximum width of each column
        :return: a list that stores the width of each column
        """
        widths = []
        calculate_width = self._calculate_table_width()
        if not self._field_names:
            return calculate_width
        for width, col_data in zip(calculate_width, self._field_names):
            max_width = self.style.get_max_width(self._field_names, col_data)
            if not max_width or width <= max_width:
                col_width = width
            else:
                col_width = max_width
            col_width = 2 if col_width <= 1 else col_width
            widths.append(col_width)
        return widths

    def _calculate_table_width(self) -> list:
        """
        Calculate the width of each column in the table by row
        :return: a list that stores the width of each column
        """
        # calculate the header width
        widths = [sum(get_char_width(col_data)) for col_data in self._field_names]
        # calculate row width
        for row in self._rows:
            for index, cell in enumerate(row):
                max_width = max(sum(get_char_width(line)) for line in cell.split("\n"))
                widths[index] = max(widths[index], max_width)
        # calculate tile width
        width_vert_separator = len(f"{self.style.vertical_char} ")
        width_row = len(widths) * width_vert_separator + sum(widths)
        width_title = sum(get_char_width(self._title))
        if width_title < width_row:
            return widths
        len_widths = len(widths)
        if len_widths == 0:
            return [width_title]
        excess_width = width_title - width_row
        add_width = excess_width // len_widths
        if excess_width % len_widths:
            add_width += 1
        return [width + add_width for width in widths]


def get_char_width(text: str) -> list:
    """
    Gets the width of each character
    :param text: text
    :return: list of widths per character
    """
    wf_width = 2  # the width of wide and full-width characters is 2
    other_char_width = 1
    widths = []
    for char in text:
        east_asian_width = unicodedata.east_asian_width(char)
        if east_asian_width in "WF":
            widths.append(wf_width)
        else:
            widths.append(other_char_width)
    return widths
