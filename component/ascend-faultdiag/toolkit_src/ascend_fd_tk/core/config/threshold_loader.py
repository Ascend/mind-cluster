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

"""阈值加载逻辑：代际Profile注册表 + 阈值配置文件加载器（解析暂存 + 延迟应用）

阈值类定义见 threshold_config.py；本模块负责按代际取Profile类，
以及从 set_config_dir 设置的目录读取 threshold_config.json 覆盖默认阈值。
"""

import inspect
import os
from typing import Dict, Optional, Tuple, Type

from ascend_fd_tk.core.common.diag_enum import NpuType
from ascend_fd_tk.core.config.threshold_config import A5Threshold, BaseThreshold
from ascend_fd_tk.core.model.threshold import Threshold
from ascend_fd_tk.utils import helpers
from ascend_fd_tk.utils.file_tool import safe_read_json
from ascend_fd_tk.utils.logger import DIAG_LOGGER

# 代际 → 阈值Profile注册表（新增代际如 A6 只需定义 A6Threshold 子类并注册）
THRESHOLD_REGISTRY: Dict[str, Type[BaseThreshold]] = {
    NpuType.A3.value: BaseThreshold,
    NpuType.A5.value: A5Threshold,
}


def get_threshold_cls(chip_generation: str = None) -> Type[BaseThreshold]:
    """按芯片代际取对应阈值Profile类，未知代际或未指定时按 A3 处理"""
    return THRESHOLD_REGISTRY.get(chip_generation or NpuType.A3.value, BaseThreshold)


# 阈值配置文件名（放在 set_config_dir 设置的配置目录下）
THRESHOLD_CONFIG_FILE_NAME = "threshold_config.json"

# 阈值配置文件中允许覆盖的字段：从 Threshold 构造参数动态派生（排除self，跳过 *args/**kwargs），
# 保证配置字段集与构造参数恒一致：Threshold 增减参数时此处自动同步，无需维护两份清单
_THRESHOLD_FIELDS = frozenset(
    name
    for name, param in inspect.signature(Threshold.__init__).parameters.items()
    if param.name != "self" and param.kind not in (param.VAR_POSITIONAL, param.VAR_KEYWORD)
)

# 数值阈值字段（可转为float参与比较）；其余为字符串字段（normal_value_*/desc/unit），不做数值校验
_NUMERICAL_THRESHOLD_FIELDS = frozenset(
    {
        "low_value_alarm",
        "high_value_alarm",
        "low_value_warn",
        "high_value_warn",
    }
)

# 字符串字段长度上限：desc 用于阈值报告/日志展示，unit 为短单位串，防止超长配置污染输出
DESC_MAX_LEN = 128
UNIT_MAX_LEN = 16


# 全部Profile类：默认值快照与重置范围
_PROFILE_CLASSES = (BaseThreshold, A5Threshold)

# 代码内默认阈值快照：模块导入时拍一次（此时类必然未被覆盖），供各加载器实例共享；
# 阈值覆盖只发生在运行期 apply()，故快照始终是干净的代码默认值
_DEFAULT_THRESHOLDS = {
    cls: {name: th for name, th in vars(cls).items() if isinstance(th, Threshold)} for cls in _PROFILE_CLASSES
}


class ThresholdConfigLoader:
    """阈值配置文件加载器：嵌套格式解析暂存 + 按代际延迟应用

    实例由诊断上下文（DiagCtx）持有；默认值快照为模块级共享（见 _DEFAULT_THRESHOLDS），
    多个实例先后应用互不污染，重置始终恢复代码内默认值。

    配置文件 threshold_config.json 放在 set_config_dir 设置的目录下，阈值字段统一套在 threshold 层内：
        {"TX_BIAS_MA": {"threshold": {"high_value_alarm": "13"}},
         "NIC_TX_BIAS_MA": {"threshold": {"low_value_alarm": "0.8"}}}
    配置键归属由类结构决定（见 _resolve_target）：全部配置键按名在本次诊断代际的 Profile 类上解析
    （getattr 走继承链，含 A5 新增网卡类别的 NIC_ 前缀阈值名）。

    使用方式（两步）：
        1. set_config_dir 时调用 parse()：仅解析校验并暂存覆盖配置，不修改阈值；
        2. 诊断启动、集群代际确定后调用 apply()：先重置全部Profile为代码内默认值，
           再按代际应用覆盖（同一份配置在 A3/A5 集群上分别覆盖各自代际Profile的阈值）。
    """

    def __init__(self):
        # parse() 暂存的覆盖配置 {配置键: {字段名: 值}} 及其来源路径
        self._overrides: Dict[str, Dict[str, str]] = {}
        self._config_path: str = ""

    def parse(self, config_dir: str) -> Dict[str, Dict[str, str]]:
        """解析配置目录下的阈值配置文件并暂存（延迟应用）

        仅解析校验并暂存覆盖配置，不修改阈值；实际应用在诊断启动、集群代际确定后调用 apply()。
        """
        self._config_path = os.path.join(config_dir, THRESHOLD_CONFIG_FILE_NAME)
        self._overrides = self._load_overrides(config_dir)
        return self._overrides

    def apply(self, chip_generation: str = None) -> None:
        """按代际应用暂存的阈值覆盖配置：配置的字段覆盖默认值，未配置的保持默认值

        先将全部Profile类重置为代码内默认值再应用覆盖（支持多次诊断、更换配置目录的场景）。
        :param chip_generation: 本次诊断的芯片代际（如 "A3"/"A5"），None 按 A3 处理
        """
        for cls, defaults in _DEFAULT_THRESHOLDS.items():
            self._reset_profile(cls, defaults)
        for name, fields in self._overrides.items():
            cls, attr_name = self._resolve_target(name, chip_generation)
            if cls is None:
                DIAG_LOGGER.warning(
                    "阈值配置项 %s 不是有效的阈值名（代际：%s），已忽略", name, chip_generation or NpuType.A3.value
                )
                continue
            setattr(cls, attr_name, getattr(cls, attr_name).merged(**fields))
            DIAG_LOGGER.info(
                "阈值 %s -> %s.%s 已按 %s 覆盖: %s", name, cls.__name__, attr_name, self._config_path, fields
            )

    @staticmethod
    def _reset_profile(th_cls: Type[BaseThreshold], defaults: Dict[str, Threshold]) -> None:
        """将Profile类重置为代码内默认值，并清除历史覆盖新增的属性"""
        own_attrs = {name for name, value in vars(th_cls).items() if isinstance(value, Threshold)}
        for name in own_attrs - set(defaults):  # 清除历史覆盖新增的属性
            delattr(th_cls, name)
        for name, default_th in defaults.items():
            setattr(th_cls, name, default_th)

    def _load_overrides(self, config_dir: str) -> Dict[str, Dict[str, str]]:
        """读取配置目录下的阈值覆盖配置文件，返回 {配置键: {字段名: 值}}；文件不存在或解析失败时返回空字典"""
        config_path = os.path.join(config_dir, THRESHOLD_CONFIG_FILE_NAME)
        if not os.path.isfile(config_path):
            return {}
        try:
            config = safe_read_json(config_path) or {}
        except Exception as e:
            DIAG_LOGGER.warning("读取阈值配置文件 %s 失败，使用默认阈值：%s", config_path, e)
            return {}
        if not isinstance(config, dict):
            DIAG_LOGGER.warning("阈值配置文件 %s 内容应为JSON对象，使用默认阈值", config_path)
            return {}
        overrides = {}
        for name, fields in config.items():
            if not isinstance(fields, dict):
                DIAG_LOGGER.warning("阈值配置项 %s 的值应为对象，已忽略", name)
                continue
            # 阈值字段统一套在 threshold 子对象内，读取该子对象中的字段
            threshold_fields = fields.get("threshold")
            if not isinstance(threshold_fields, dict):
                DIAG_LOGGER.warning("阈值配置项 %s 缺少 threshold 对象，已忽略", name)
                continue
            invalid_fields = [field for field in threshold_fields if field not in _THRESHOLD_FIELDS]
            if invalid_fields:
                DIAG_LOGGER.warning("阈值配置项 %s 含无效字段 %s，已忽略这些字段", name, invalid_fields)
            valid_fields = {}
            for field, value in threshold_fields.items():
                if field not in _THRESHOLD_FIELDS:
                    continue
                str_value = self._validate_field(name, field, value)
                if str_value is not None:
                    valid_fields[field] = str_value
            if valid_fields:
                overrides[name] = valid_fields
        return overrides

    @staticmethod
    def _validate_field(name: str, field: str, value) -> Optional[str]:
        """校验单个阈值字段，返回合法字符串值；值为空/类型错误/非法数值/超长时返回 None（该字段已忽略）"""
        # 空值（null/空白字符串）视为未配置该字段，保持代码默认值，避免误清空默认阈值
        if value is None or (isinstance(value, str) and not value.strip()):
            DIAG_LOGGER.warning("阈值配置项 %s 的字段 %s 配置值为空，使用默认值", name, field)
            return None
        if isinstance(value, str):
            str_value = value
        else:
            str_value = str(value)
            # 字符串字段（normal_value_*/desc/unit）要求字符串类型，数字/bool/数组/对象告警忽略，避免被静默强转
            if field not in _NUMERICAL_THRESHOLD_FIELDS:
                DIAG_LOGGER.warning("阈值配置项 %s 的字段 %s 值 %s 不是字符串，已忽略该字段", name, field, str_value)
                return None
        # 数值字段在解析期校验合法性：无法转数值的值（如 "abc"）告警忽略，避免静默失效
        if field in _NUMERICAL_THRESHOLD_FIELDS and not helpers.to_float(str_value)[0]:
            DIAG_LOGGER.warning("阈值配置项 %s 的字段 %s 值 %s 不是合法数值，已忽略该字段", name, field, str_value)
            return None
        # desc/unit 长度上限：防止超长字符串污染阈值报告/日志
        max_len = DESC_MAX_LEN if field == "desc" else (UNIT_MAX_LEN if field == "unit" else None)
        if max_len is not None and len(str_value) > max_len:
            DIAG_LOGGER.warning(
                "阈值配置项 %s 的字段 %s 值超长（%d/%d 字符），已忽略该字段", name, field, len(str_value), max_len
            )
            return None
        return str_value

    @staticmethod
    def _resolve_target(name: str, chip_generation: str) -> Tuple[Optional[Type[BaseThreshold]], Optional[str]]:
        """解析配置键归属：返回 (Profile类, 属性名)，无法解析返回 (None, None)

        归属由类结构决定，不做名称推断：全部配置键按名在本次诊断代际的 Profile 类上解析
        （getattr 走继承链，含 A5 新增网卡类别的 NIC_ 前缀阈值名）。
        """
        cls, attr = get_threshold_cls(chip_generation), name
        if isinstance(getattr(cls, attr, None), Threshold):
            return cls, attr
        return None, None
