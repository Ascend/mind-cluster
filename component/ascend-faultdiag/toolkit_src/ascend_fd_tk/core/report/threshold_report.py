#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
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

from dataclasses import dataclass, field
from enum import Enum
from typing import List, Dict, Any, Optional, TypeVar, Generic

from ascend_fd_tk.core.model.threshold import Threshold, ThresholdStatus
from ascend_fd_tk.utils.excel_tool import ExcelGenerator, Color, CellStyle, StyledCell

T = TypeVar('T')


class ThresholdColorMap(Enum):
    """阈值状态与颜色的映射关系"""
    NORMAL = Color.LIGHT_SUCCESS
    LOW_THRESHOLD_ALARM = Color.LIGHT_ERROR
    LOW_THRESHOLD_WARN = Color.LIGHT_WARNING
    HIGH_THRESHOLD_ALARM = Color.LIGHT_ERROR
    HIGH_THRESHOLD_WARN = Color.LIGHT_WARNING
    NOT_EQUAL_THRESHOLD_ALARM = Color.LIGHT_ERROR
    NOT_EQUAL_THRESHOLD_WARN = Color.LIGHT_WARNING


@dataclass
class ThresholdConfig:
    """阈值配置类，用于配置某个字段的阈值信息"""
    field_name: str  # 对象中的字段名
    threshold: Threshold  # 对应的阈值对象
    display_name: Optional[str] = None  # 显示在Excel中的列名
    value_converter: Optional[callable] = str  # 值转换器，将对象字段值转换为字符串
    desc: Optional[str] = None  # 字段描述

    @property
    def column_name(self) -> str:
        """获取列名"""
        base_name = self.display_name or self.field_name
        if self.threshold and self.threshold.unit:
            return f"{base_name} ({self.threshold.unit})"
        return base_name


@dataclass
class ReportSheet(Generic[T]):
    """报告Sheet类，对应Excel中的一个Sheet"""
    sheet_name: str  # Sheet名称
    data_list: List[T]  # 数据列表
    header_mapping: Dict[str, str]  # header与对象字段的映射关系 {field_name: header_name}
    header_order: List[str]  # header的顺序（使用header_name）
    threshold_configs: List[ThresholdConfig]  # 阈值配置列表
    na_rep: str = ""  # 空值替换字符串
    header_widths: Optional[Dict[str, int]] = field(default_factory=dict)  # 列宽配置

    def __post_init__(self):
        # 验证header_order中的header是否都在header_mapping.values()中
        for header in self.header_order:
            if header not in self.header_mapping.values():
                raise ValueError(f"header_order中的{header}不在{self.header_mapping.values()}中")

        # 验证threshold_configs中的field_name是否都在header_mapping.keys()中
        for config in self.threshold_configs:
            if config.field_name not in self.header_mapping:
                raise ValueError(f"threshold_configs中的{config.field_name}不在{self.header_mapping.keys()}中")

        # 更新header_mapping以包含单位信息
        new_header_mapping = {}
        header_name_map = {}  # 记录原始header_name到新header_name的映射

        for field_name, header_name in self.header_mapping.items():
            # 检查是否有阈值配置
            config = next((c for c in self.threshold_configs if c.field_name == field_name), None)
            if config and config.threshold and config.threshold.unit:
                # 有单位信息，更新列名
                new_header_name = f"{header_name} ({config.threshold.unit})"
                new_header_mapping[field_name] = new_header_name
                header_name_map[header_name] = new_header_name
            else:
                # 没有单位信息，保持原列名
                new_header_mapping[field_name] = header_name

        # 更新header_order以包含单位信息
        new_header_order = []
        for header_name in self.header_order:
            new_header_order.append(header_name_map.get(header_name, header_name))

        # 更新属性
        self.header_mapping = new_header_mapping
        self.header_order = new_header_order


@dataclass
class ReportData:
    """报告数据类，用于封装报告的所有数据"""
    sheets: List[ReportSheet] = field(default_factory=list)  # 所有Sheet数据

    def add_sheet(self, sheet: ReportSheet) -> None:
        """添加一个Sheet"""
        self.sheets.append(sheet)


class ReportGenerator:
    """报告生成器类，用于生成Excel报告"""

    def __init__(self, excel_gen: Optional[ExcelGenerator] = None):
        self.excel_gen = excel_gen or ExcelGenerator()
        self.report_data = ReportData()

    @staticmethod
    def _get_field_value(obj: Any, field_name: str) -> Any:
        """获取对象的字段值，支持嵌套字段（使用点号分隔）"""
        value = obj
        for part in field_name.split('.'):
            if hasattr(value, part):
                value = getattr(value, part)
            elif isinstance(value, dict) and part in value:
                value = value[part]
            else:
                return None
        return value

    def add_sheet(self, sheet: ReportSheet) -> None:
        """添加一个Sheet"""
        self.report_data.add_sheet(sheet)

    def generate_excel(self, output_path: Optional[str] = None) -> Optional[ExcelGenerator]:
        """生成Excel报告

        :param output_path: 输出Excel文件路径，如果提供则保存文件，否则返回ExcelGenerator实例
        :return: 如果没有提供output_path，返回修改后的ExcelGenerator实例；否则返回None
        """
        if not self.report_data.sheets:
            raise ValueError("没有添加任何Sheet数据")

        # 为每个Sheet添加数据
        for sheet in self.report_data.sheets:
            # 转换为带样式的字典列表
            styled_data = self._convert_to_styled_dict(sheet)
            # 添加Sheet到Excel生成器
            self.excel_gen.add_sheet(
                sheet_name=sheet.sheet_name,
                data=styled_data,
                columns=sheet.header_order,
                na_rep=sheet.na_rep,
                header_widths=sheet.header_widths
            )

        # 生成Excel文件
        if output_path:
            self.excel_gen.generate_excel(output_path)
            return None
        else:
            return self.excel_gen

    def _convert_to_styled_dict(self, sheet: ReportSheet) -> List[Dict[str, Any]]:
        """将数据列表转换为带样式的字典列表，用于Excel生成"""
        styled_data = []

        # 创建阈值配置映射 {field_name: ThresholdConfig}
        threshold_config_map = {config.field_name: config for config in sheet.threshold_configs}

        for obj in sheet.data_list:
            row_data = {}
            for field_name, header_name in sheet.header_mapping.items():
                # 获取字段值
                value = self._get_field_value(obj, field_name)

                # 应用阈值检查和样式
                if field_name in threshold_config_map:
                    config = threshold_config_map[field_name]
                    # 使用值转换器转换值
                    value_str = config.value_converter(value) if value is not None else sheet.na_rep

                    # 检查阈值
                    if not value_str or value_str == sheet.na_rep:
                        # 空值，不应用颜色
                        color = None
                        display_value = value_str
                    else:
                        th_status, th_value = config.threshold.check_value(value_str)
                        # 获取颜色
                        color = ThresholdColorMap[th_status.name].value

                        # 格式化显示内容
                        display_value = value_str
                        if th_status != ThresholdStatus.NORMAL and th_value:
                            # 根据阈值状态确定比较符号
                            if th_status in [ThresholdStatus.HIGH_THRESHOLD_ALARM, ThresholdStatus.HIGH_THRESHOLD_WARN]:
                                comparator = ">"
                            elif th_status in [ThresholdStatus.LOW_THRESHOLD_ALARM, ThresholdStatus.LOW_THRESHOLD_WARN]:
                                comparator = "<"
                            elif th_status in [ThresholdStatus.NOT_EQUAL_THRESHOLD_ALARM,
                                               ThresholdStatus.NOT_EQUAL_THRESHOLD_WARN]:
                                comparator = "!="
                            else:
                                comparator = "=="
                            # 格式化显示为 "value | > threshold" 或 "value | < threshold" 或 "value | != threshold"
                            display_value = f"{value_str} | {comparator} {th_value}"

                    # 创建带样式的单元格
                    if color:
                        row_data[header_name] = StyledCell(display_value, CellStyle(bg_color=color))
                    else:
                        row_data[header_name] = StyledCell(display_value)
                else:
                    # 普通字段，不应用样式
                    row_data[header_name] = value if value is not None else sheet.na_rep

            styled_data.append(row_data)

        return styled_data


def create_threshold_report(
        sheet_name: str,
        data_list: List[Any],
        header_mapping: Dict[str, str],
        header_order: List[str],
        threshold_configs: List[ThresholdConfig],
        na_rep: str = "",
        header_widths: Optional[Dict[str, int]] = None
) -> ReportSheet:
    """创建阈值报告Sheet的便捷函数"""
    return ReportSheet(
        sheet_name=sheet_name,
        data_list=data_list or [],
        header_mapping=header_mapping or {},
        header_order=header_order or [],
        threshold_configs=threshold_configs or [],
        na_rep=na_rep,
        header_widths=header_widths or {}
    )


def generate_threshold_excel(
        output_path: Optional[str] = None,
        sheets: Optional[List[ReportSheet]] = None,
        excel_gen: Optional[ExcelGenerator] = None
) -> Optional[ExcelGenerator]:
    """生成阈值Excel报告的便捷函数

    :param output_path: 输出Excel文件路径，如果提供则保存文件，否则返回ExcelGenerator实例
    :param sheets: 要添加的Sheet列表
    :param excel_gen: 现有的ExcelGenerator实例，如果不提供则创建新实例
    :return: 如果没有提供output_path，返回修改后的ExcelGenerator实例；否则返回None
    """
    report_gen = ReportGenerator(excel_gen=excel_gen)

    # 添加所有Sheet
    if sheets:
        for sheet in sheets:
            report_gen.add_sheet(sheet)

    # 生成Excel报告
    return report_gen.generate_excel(output_path=output_path)
