#!/usr/bin/python3
# -*- coding: utf-8 -*-

#   Copyright (C)  2025. Huawei Technologies Co., Ltd. All rights reserved.
import os
import stat

from taskd.python.toolkit.constants import constants


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
