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

"""
报告Sheet生成器基类
"""

from abc import ABC, abstractmethod
from typing import Optional

from ascend_fd_tk.core.model.cluster_info_cache import ClusterInfoCache
from ascend_fd_tk.utils.excel_tool import ExcelGenerator


class BaseSheetGenerator(ABC):
    """报告Sheet生成器基类"""

    def __init__(self, cluster_info: ClusterInfoCache, excel_gen: Optional[ExcelGenerator] = None):
        """
        初始化Sheet生成器

        :param cluster_info: 集群信息缓存对象
        :param excel_gen: Excel生成器对象，如果不提供则创建新实例
        """
        self.cluster_info = cluster_info
        self.excel_gen = excel_gen or ExcelGenerator()

    @abstractmethod
    def generate_sheet(self) -> None:
        """
        生成Excel Sheet的抽象方法，子类必须实现
        """
        pass

    def get_excel_gen(self) -> ExcelGenerator:
        """
        获取Excel生成器对象

        :return: ExcelGenerator实例
        """
        return self.excel_gen

    def save_excel(self, output_path: str) -> None:
        """
        保存Excel文件

        :param output_path: 输出文件路径
        """
        self.excel_gen.generate_excel(output_path)
