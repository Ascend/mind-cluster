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
import argparse
import os
import json
import logging
from dataclasses import dataclass, field
from typing import List

from ascend_fd.configuration.config import CUSTOM_CONFIG_PATH
from ascend_fd.utils.comm_valid import valid_int, char_check, valid_list_len, LINE_WHITE_LIST
from ascend_fd.utils.status import ParamError
from ascend_fd.utils.json_dict import JsonObj
from ascend_fd.utils.tool import safe_write_open, safe_read_json, PATH_WHITE_LIST_LIN, white_check, MAX_PATH_LEN

echo = logging.getLogger("ECHO")
logger = logging.getLogger("FAULT_DIAG")
TRAIN_LOG_READ_SIZE = 1024 * 1024
WHITE_LIST_TIME_FORMAT = "YmdHMSf" + "%- :,."
WHITE_LIST_FILE_GLOB = PATH_WHITE_LIST_LIN + "*"


@dataclass
class CustomFileInfo(JsonObj):
    file_path_glob: str = ""
    log_time_format: str = ""
    source_file: List[str] = field(default_factory=list)

    def __post_init__(self):
        validation_rules = {
            "file_path_glob": self._valid_file_path_glob,
            "log_time_format": self._valid_time_format,
            "source_file": self._valid_source_file,
        }
        for attr, validator in validation_rules.items():
            validator(attr, getattr(self, attr))

    @staticmethod
    def _valid_file_path_glob(attr: str, path: str):
        """
        Validate glob pattern file path
        :param attr: attribute name
        :param path: attribute value, glob pattern file path
        """
        if not path:
            raise ValueError(f"Invalid value for {attr}: '{path}'. Input is empty.")
        if not white_check(path, WHITE_LIST_FILE_GLOB):
            raise argparse.ArgumentTypeError(
                f"The {attr} is invalid.\n"
                "The path can contain only digits, uppercase and lowercase letters, "
                "and following characters: ['~', '+', '-', '_', '.', ' ', '*', '/']")
        if len(path) < 1 or len(path) > MAX_PATH_LEN:
            raise argparse.ArgumentTypeError(
                f"The {attr} is invalid.\n"
                f"The path length exceeds the maximum path length({MAX_PATH_LEN}).")

    @staticmethod
    def _valid_time_format(attr: str, val: str):
        """
        Validate the time format
        :param attr: attribute name
        :param val: attribute value, the time format
        """
        if not char_check(val, length_range=(0, 50), white_list=WHITE_LIST_TIME_FORMAT, allow_zh=False):
            raise ValueError(f"Invalid value for {attr}: {val}. The time format is invalid.")

    @staticmethod
    def _valid_source_file(attr: str, val: list):
        """
        Validate the source file
        :param attr: attribute name
        :param val: attribute value, the source file
        """
        if not val:
            raise ValueError(f"Invalid value for {attr}: {val}. The input is empty.")
        valid_list_len(attr, val)
        for each_str in val:
            if char_check(each_str, length_range=(1, 50), white_list=LINE_WHITE_LIST, allow_zh=False):
                continue
            raise ValueError(f"Invalid value for {attr}: {val}. Verification of '{each_str}' failed.")


@dataclass
class TimeZoneCFG(JsonObj):
    """
    Time Zone Config
    """
    lcne: bool = False
    mindie: bool = False

    def get_trans_flag_by_type(self, log_type):
        return getattr(self, log_type, False)


@dataclass
class ConfigInfo(JsonObj):
    enable_model_asrt: bool = False
    train_log_size: int = TRAIN_LOG_READ_SIZE
    custom_parse_file: List[CustomFileInfo] = field(default_factory=list)
    timezone_config: TimeZoneCFG = None

    def __post_init__(self):
        validation_rules = {
            "train_log_size": valid_int,
            "custom_parse_file": valid_list_len
        }
        for attr, validator in validation_rules.items():
            validator(attr, getattr(self, attr))


def get_config_info(data_path: str = CUSTOM_CONFIG_PATH):
    """
    Get the custom configuration file info
    :param data_path: str, the custom configuration file path
    :return: ConfigInfo, the custom configuration file info
    """
    update_data = read_config_info(data_path)
    config_info = validate_config_info(update_data)
    return config_info


def read_config_info(data_path: str = CUSTOM_CONFIG_PATH):
    update_data = safe_read_json(data_path)
    return update_data


def validate_config_info(update_data):
    try:
        config_info = ConfigInfo.from_dict(update_data, check_parameter=True)
    except (TypeError, ValueError) as err:
        logger.error("Failed to convert the custom config file content. The reason is: %s", err)
        raise ParamError(f"Failed to convert the config file content: {err}") from err
    return config_info


def update_config_info(data_path: str):
    """
    Update the custom configuration file by input data file
    :param data_path: str, the updated data file path
    """
    config_info = get_config_info(data_path)
    config_data = config_info.to_dict()
    with safe_write_open(CUSTOM_CONFIG_PATH, mode="w+", encoding="utf-8") as f_dump:
        f_dump.write(json.dumps(config_data, sort_keys=False, separators=(',', ':'), ensure_ascii=False, indent=4))
    echo.info("The custom config file was updated successfully.")
    logger.info("The custom config file was updated successfully.")


def check_config_info():
    """
    Check the custom configuration file
    """
    _ = validate_config_info(read_config_info())
    echo.info("The custom config file was checked successfully.")
    logger.info("The custom config file was checked successfully.")


def show_config_info():
    """
    Show the custom configuration file content
    """
    show_data = {}
    if os.path.exists(CUSTOM_CONFIG_PATH):
        show_data = json.dumps(get_config_info().to_dict(), ensure_ascii=False, indent=4)
    echo.info(show_data)
    logger.info("Show that the custom configuration file content is executed successfully.")
