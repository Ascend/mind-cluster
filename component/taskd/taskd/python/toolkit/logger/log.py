#!/usr/bin/python3
# -*- coding: utf-8 -*-
#  Copyright (C)  2022. Huawei Technologies Co., Ltd. All rights reserved.
import logging.config
import logging.handlers
import os
import yaml

from taskd.python.toolkit.config import path
from taskd.python.toolkit.constants.constants import (LOG_PRIVILEGE,
                                                                              LOG_DIR_PRIVILEGE, LOG_BAK_PRIVILEGE, LOG_FILE_PATH_ENV)
from taskd.python.toolkit.constants.constants import MAX_SIZE
from taskd.python.toolkit.validator.file_process import safe_open
from taskd.python.toolkit.validator.validators import FileValidator
from taskd.python.toolkit.validator.validators import DirectoryValidator
from taskd.python.toolkit.constants.constants import LOG_MAX_LINE_LENGTH
from taskd.python.toolkit.constants.constants import LOG_SIMPLE_FORMAT
from taskd.python.toolkit.logger.custom_formatter import MaxLengthFormatter


def init_sys_log():
    config_path = path.LOG_CFG_FILE
    real_config_path = os.path.realpath(config_path)
    if real_config_path != config_path:
        raise ValueError("Config path not correct.")

    if not os.path.exists(real_config_path):
        raise ValueError("Config path does not exist.")

    if not FileValidator(real_config_path).check_file_size().check().is_valid():
        raise ValueError("Config file size is not valid.")

    with safe_open(real_config_path, mode='r', encoding='UTF-8') as fp:
        data = fp.read(MAX_SIZE)
        log_cfg = yaml.safe_load(data)

    log_path = os.getenv(LOG_FILE_PATH_ENV)
    if log_path is not None:
        log_file = os.path.join(log_path, "run.log")
        handlers = log_cfg.get('handlers')
        for handler_name in handlers:
            handler_dict = handlers.get(handler_name)
            handler_dict['filename'] = log_file
    else:
        os.environ[LOG_FILE_PATH_ENV] = "/var/log/mindx-dl/elastic"

    init_log_dir_for_dt(log_cfg)

    logging.config.dictConfig(log_cfg)


def init_log_dir_for_dt(log_cfg):
    """Create log directory.

    :param log_cfg: log configuration dictionary from yml file.
    :return: None
    """
    handlers = log_cfg.get('handlers')
    if not handlers:
        return

    for handler_name in handlers:
        handler_dict = handlers.get(handler_name)
        log_file = handler_dict.get('filename')

        if not log_file:
            continue

        log_file_standard = os.path.realpath(log_file)
        if log_file_standard != log_file:
            continue

        log_dir = os.path.dirname(log_file_standard)
        if not DirectoryValidator(log_dir) \
                .check_is_not_none() \
                .check_dir_name() \
                .should_not_contains_sensitive_words() \
                .with_blacklist() \
                .check() \
                .is_valid():
            continue

        _process_log(log_dir)


def _process_log(log_dir: str) -> None:
    """
    Create or check run log and service log file
    :param log_dir: log directory
    :return: None
    """
    if os.path.exists(log_dir) and not DirectoryValidator(log_dir).check_user_group().check().is_valid():
        raise ValueError("Invalid run log dir permissions.")

    os.makedirs(log_dir, mode=LOG_DIR_PRIVILEGE, exist_ok=True)
    run_log_path = os.path.join(log_dir, "run.log")
    if run_log_path != os.path.join(os.getenv(LOG_FILE_PATH_ENV), "run.log"):
        raise ValueError("Run log file path is not correct.")
    if not os.path.exists(run_log_path):
        os.mknod(os.path.join(log_dir, "run.log"), mode=LOG_PRIVILEGE)
    else:
        _exist_file_process(run_log_path)

    os.chmod(log_dir, LOG_DIR_PRIVILEGE)


def _exist_file_process(log_path: str) -> None:
    """
    Handle log file when file is already existed.
    :param log_path: log file path
    :return: None
    """
    # check is soft link or not
    if not FileValidator(log_path).check_not_soft_link().check().is_valid():
        raise ValueError("Run log file path is a soft link.")

    # check process user and group with log file
    if not FileValidator(log_path).check_user_group().check().is_valid():
        raise ValueError("Invalid run log file permissions.")

    # check log file permission
    os.chmod(log_path, LOG_PRIVILEGE)


def _log_rotator(source: str, dest: str) -> None:
    if os.path.exists(source):
        os.rename(source, dest)
        os.chmod(dest, mode=LOG_BAK_PRIVILEGE)
        if not os.path.exists(source):
            os.mknod(source, mode=LOG_PRIVILEGE)
        else:
            _exist_file_process(source)


def _set_rotator_func(handlers: logging.handlers, rotate_func: callable) -> logging.handlers:
    for handler in handlers:
        handler.rotator = rotate_func

    return handlers


def _set_formatter(logger: logging.Logger, fmt: str):
    for handler in logger.handlers:
        formatter = MaxLengthFormatter(fmt, LOG_MAX_LINE_LENGTH)
        handler.setFormatter(formatter)


init_sys_log()

run_log = logging.getLogger("logRun")
_set_formatter(run_log, LOG_SIMPLE_FORMAT)
run_log.handlers = _set_rotator_func(run_log.handlers, _log_rotator)
