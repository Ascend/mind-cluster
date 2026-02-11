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

import logging
import os
from logging.handlers import RotatingFileHandler

from toolkit.core.common import constants
from toolkit.core.common.path import CommonPath


class SimpleLogger:
    def __init__(self,
                 logger_name: str = constants.DEFAULT_LOGGER_NAME,
                 log_dir: str = CommonPath.LOG_DIR,
                 log_filename: str = "ascend-fd-tk.log",
                 max_size: int = 10 * 1024 * 1024,  # 10MB
                 backup_count: int = 5,  # 最多保留5个归档文件
                 level: int = logging.INFO):
        """
        初始化日志模块
        :param logger_name: 日志器名称
        :param log_dir: 日志目录
        :param log_filename: 主日志文件名
        :param max_size: 单个日志文件最大大小（字节），默认10MB
        :param backup_count: 归档文件保留数量
        :param level: 日志级别（DEBUG/INFO/WARNING/ERROR/CRITICAL）
        """
        self.logger = logging.getLogger(logger_name)
        self.logger.setLevel(level)
        self.logger.handlers = []  # 清除已有处理器，避免重复输出

        # 确保日志目录存在
        os.makedirs(log_dir, exist_ok=True)
        self.log_path = os.path.join(log_dir, log_filename)

        # 日志格式（包含时间、级别、模块、消息）
        formatter = logging.Formatter(
            f'[%(levelname)s] {logger_name}:%(asctime)s [%(module)s:%(lineno)d] %(message)s',
            datefmt='%Y-%m-%d %H:%M:%S'
        )

        # 1. 添加文件处理器（支持大小限制和归档）
        file_handler = RotatingFileHandler(
            filename=self.log_path,
            maxBytes=max_size,
            backupCount=backup_count,
            encoding='utf-8'
        )
        # 归档文件命名格式：app.log.1, app.log.2...（最新的归档编号最小）
        file_handler.setFormatter(formatter)
        self.logger.addHandler(file_handler)

        # 2. 添加控制台处理器
        console_handler = logging.StreamHandler()
        console_handler.setFormatter(formatter)
        self.logger.addHandler(console_handler)

    def get_logger(self) -> logging.Logger:
        """获取配置好的日志器"""
        return self.logger


DIAG_LOGGER = SimpleLogger().get_logger()


class ConsoleLogger:
    def __init__(self, logger_name: str = constants.DEFAULT_CONSOLE_LOGGER_NAME, level: int = logging.INFO):
        """
        初始化简单控制台日志模块
        :param logger_name: 日志器名称
        :param level: 日志级别（DEBUG/INFO/WARNING/ERROR/CRITICAL）
        """
        self.logger = logging.getLogger(logger_name)
        self.logger.setLevel(level)
        self.logger.handlers = []

        # 添加控制台处理器（只打印内容）
        console_handler = logging.StreamHandler()
        console_handler.setFormatter(logging.Formatter('%(message)s'))
        self.logger.addHandler(console_handler)

    def get_logger(self) -> logging.Logger:
        """获取配置好的日志器"""
        return self.logger


CONSOLE_LOGGER = ConsoleLogger().get_logger()

# 使用示例
if __name__ == "__main__":
    # 方式1：自定义配置
    custom_logger = DIAG_LOGGER

    custom_logger.debug("这是调试信息（仅DEBUG级别可见）")
    custom_logger.info("这是普通信息日志")
    custom_logger.warning("这是警告日志")
    custom_logger.error("这是错误日志")
    custom_logger.critical("这是严重错误日志")

    # 方式2：简单控制台日志（只打印内容）
    simple_logger = CONSOLE_LOGGER
    simple_logger.info("这是简单普通信息日志")
