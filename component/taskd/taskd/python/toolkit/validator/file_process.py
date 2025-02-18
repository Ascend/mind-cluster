#!/usr/bin/python3
# -*- coding: utf-8 -*-

#   Copyright (C)  2025. Huawei Technologies Co., Ltd. All rights reserved.
import os
import stat

from taskd.python.toolkit.constants import constants
from taskd.python.toolkit.constants.constants import MAX_SIZE


def safe_open(file, mode="r", encoding='utf-8', errors=None, newline=None):
    """
    Check open file validality.
    :param file: open file
    :param mode: file open mode
    :param encoding: the encoding format
    :param errors: string specifying how to handle encoding/decoding errors
    :param newline: how newlines mode works
    :return: file stream
    """
    file_real_path = os.path.realpath(file)
    file_stream = open(file=file_real_path, mode=mode, encoding=encoding,
                       errors=errors, newline=newline, closefd=True)
    file_info = os.stat(file_stream.fileno())
    if stat.S_ISLNK(file_info.st_mode):
        file_stream.close()
        raise ValueError(f"{os.path.basename(file)} should not be a symbolic link file")

    if file_info.st_size > MAX_SIZE:
        file_stream.close()
        raise ValueError(f"File {os.path.basename(file)} size should be less than {MAX_SIZE} bytes.")

    if file_info.st_uid != os.geteuid():
        file_stream.close()
        raise ValueError(f"{os.path.basename(file)} is not owned by current user or root.")
    return file_stream


def safe_get_file_info(file_path: str) -> str:
    if not os.path.exists(file_path):
        return ""
    try:
        with open(file_path, "r") as f:
            if os.path.islink(file_path) or not os.path.isfile(file_path):
                return ""

            if os.path.getsize(file_path) > constants.MAX_FILE_SIZE:
                return ""
            return f.read()
    except Exception as err:
        return ""
