#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================
import datetime
import logging
import os
import re
import sys
from logging.handlers import RotatingFileHandler
from taskd.python.constants.constants import (LOG_DEFAULT_FILE_PATH, LOG_MAX_LINE_LENGTH, LOG_DATE_FORMAT,
                                              LOG_SIMPLE_FORMAT, LOG_DEFAULT_FILE, LOG_DEFAULT_FILE_NAME,
                                              LOG_DEFAULT_BACKUP_COUNT, LOG_DEFAULT_MAX_BYTES, LOG_BACKUP_FORMAT,
                                              LOG_PRIVILEGE, LOG_DIR_PRIVILEGE, LOG_BAK_PRIVILEGE, LOG_BACKUP_PATTERN,
                                              TASKD_LOG_LEVEL, TASKD_LOG_STDOUT, TASKD_LOG_PATH)
from taskd.python.utils.validator import FileValidator


class MaxLengthFormatter(logging.Formatter):
    '''
    Max Length Formatter format log max length
    '''

    def __init__(self, fmt, max_length, datefmt=None):
        super().__init__(fmt=fmt, datefmt=datefmt)
        self.max_length = max_length

    def format(self, record):
        msg = super().format(record)
        # The repr() function will escape characters like \r and \n.
        # The repr() function adds quotation marks at the beginning and end of a string; these need to be removed.
        msg = repr(msg)[1:-1]
        if len(msg) > self.max_length:
            return msg[:self.max_length] + '...'
        return msg


class CustomRotatingHandler(RotatingFileHandler):
    '''
    Custom RotatingFileHandler to backup same format log file
    '''

    def __init__(self, filename, maxBytes=0, backupCount=0, encoding=None, delay=None):
        super().__init__(
            filename, maxBytes=maxBytes,
            backupCount=backupCount,
            encoding=encoding,
            delay=delay
        )

    def rotation_filename(self, default_name):
        base, ext = os.path.splitext(self.baseFilename)
        back_time = datetime.datetime.now(tz=datetime.timezone.utc).strftime(LOG_BACKUP_FORMAT)[:-3]
        return f"{base}-{back_time}{ext}"

    def doRollover(self):
        """
        rewrite to do roll log file
        """
        if self.stream:
            self.stream.close()
            self.stream = None

        # create backup file name and rename the current log file
        backup_file_name = self.rotation_filename(self.baseFilename)
        if os.path.exists(backup_file_name):
            os.remove(backup_file_name)
        os.rename(self.baseFilename, backup_file_name)

        # clear backup files that exceed the file limit
        if self.backupCount > 0:
            dir_name = os.path.dirname(self.baseFilename)
            base_filename = os.path.basename(self.baseFilename)
            base, ext = os.path.splitext(base_filename)

            # match all backup files that match the format
            pattern = re.compile(rf'^({re.escape(base)}-{LOG_BACKUP_PATTERN}{re.escape(ext)})$')
            backups = []

            for filename in os.listdir(dir_name):
                match = pattern.match(filename)
                if match:
                    timestamp_str = match.group(1)
                    # get timestamp str for sort file
                    timestamp_str = timestamp_str[len(base) + 1:-len(ext)]

                    try:
                        # resolve timestamps in file names
                        ts = datetime.datetime.strptime(
                            timestamp_str, LOG_BACKUP_FORMAT)
                        backups.append((filename, ts))
                    except ValueError:
                        continue

            # sort by time (old → new)
            backups.sort(key=lambda x: x[1])

            # delete old backups that exceed the reserved quantity
            while len(backups) > self.backupCount:
                oldest = backups.pop(0)
                os.remove(os.path.join(dir_name, oldest[0]))

        # create new log file
        if not self.delay:
            self.stream = self._open()


class LogConfig:
    '''
    Log Config include logger configuration
    '''

    def __init__(self):
        self.log_max_line_length = LOG_MAX_LINE_LENGTH
        self.log_level = logging.INFO
        self.log_format = LOG_SIMPLE_FORMAT
        self.log_file = LOG_DEFAULT_FILE
        self.log_std_out = True
        self.log_backup_count = LOG_DEFAULT_BACKUP_COUNT
        self.log_max_bytes = LOG_DEFAULT_MAX_BYTES
        self.build_config()

    def build_config(self):
        self.build_log_path()
        self.build_loglevel()
        self.build_log_stdout()

    def build_log_path(self):
        log_path = os.getenv(TASKD_LOG_PATH)
        if log_path is None:
            log_path = LOG_DEFAULT_FILE_PATH
        if not os.path.exists(log_path):
            os.makedirs(log_path, mode=LOG_DIR_PRIVILEGE, exist_ok=True)
        log_file = os.path.join(log_path, LOG_DEFAULT_FILE_NAME)
        if not os.path.exists(log_file):
            os.mknod(log_file, mode=LOG_PRIVILEGE)
        self.log_file = log_file

    def build_loglevel(self):
        log_level = os.getenv(TASKD_LOG_LEVEL)
        if log_level is not None:
            self.log_level = log_level

    def build_log_stdout(self):
        log_stdout = os.getenv(TASKD_LOG_STDOUT)
        if log_stdout is not None and log_stdout is False:
            self.log_std_out = False


def _set_formatter(logger: logging.Logger, fmt: str):
    for handler in logger.handlers:
        formatter = MaxLengthFormatter(fmt, LOG_MAX_LINE_LENGTH, datefmt=LOG_DATE_FORMAT)
        handler.setFormatter(formatter)


def _set_loglevel(logger: logging.Logger, level: int):
    logger.setLevel(level)
    for handler in logger.handlers:
        handler.setLevel(level)


def _set_rotator(logger: logging.Logger, rotate_func: callable):
    for handler in logger.handlers:
        handler.rotator = rotate_func


def _log_rotator(source: str, dest: str) -> None:
    if os.path.exists(source):
        os.rename(source, dest)
        os.chmod(dest, mode=LOG_PRIVILEGE)
        if not os.path.exists(source):
            os.mknod(source, mode=LOG_PRIVILEGE)
        else:
            _exit_file_process(source)


def _exit_file_process(log_path: str) -> None:
    """
    Handle log file when file is already existed.
    :param log_path: log file path
    :return: None
    """
    # check is soft link or not
    if not FileValidator(log_path).check_not_soft_link().check().is_valid():
        raise ValueError("Run log file path is a soft link.")

    # check log file permission
    os.chmod(log_path, LOG_PRIVILEGE)


def _get_stream_handler(cfg: LogConfig):
    stream_handler = logging.StreamHandler(sys.stdout)
    stream_handler.setLevel(cfg.log_level)
    stream_handler.setFormatter(logging.Formatter(cfg.log_format, datefmt=LOG_DATE_FORMAT))
    return stream_handler


def _get_file_handler(cfg: LogConfig):
    file_handler = CustomRotatingHandler(cfg.log_file, maxBytes=cfg.log_max_bytes, backupCount=cfg.log_backup_count)
    file_handler.setLevel(cfg.log_level)
    file_handler.setFormatter(logging.Formatter(cfg.log_format, datefmt=LOG_DATE_FORMAT))
    return file_handler


def _get_logger() -> logging.Logger:
    # init logger and log config
    log_cfg = LogConfig()
    logger = logging.getLogger()

    # set log print std out
    if log_cfg.log_std_out:
        logger.addHandler(_get_stream_handler(log_cfg))

    # set log write to file
    logger.addHandler(_get_file_handler(log_cfg))

    _set_rotator(logger, rotate_func=_log_rotator)
    _set_formatter(logger, log_cfg.log_format)
    _set_loglevel(logger, log_cfg.log_level)
    return logger


run_log = _get_logger()
