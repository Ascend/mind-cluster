#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2024. All rights reserved.
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
import os
from mindio_acp.common.mindio_logger import LOGGER
MAX_FILE_PATH = 1024


class FileCheckPolicy:
    CHECK_NOT_EMPTY = 1
    CHECK_LENGTH = 1 << 1
    CHECK_SYMBOLIC_LINK = 1 << 2
    CHECK_EXIST = 1 << 3


CHECK_ALL = (FileCheckPolicy.CHECK_NOT_EMPTY | FileCheckPolicy.CHECK_LENGTH |
             FileCheckPolicy.CHECK_SYMBOLIC_LINK | FileCheckPolicy.CHECK_EXIST)


def file_path_check(path, check_policy=CHECK_ALL):
    fcp = FileCheckPolicy
    if (check_policy & fcp.CHECK_NOT_EMPTY == fcp.CHECK_NOT_EMPTY) and not path:
        LOGGER.error('Error: missing file path.')
        return False

    if (check_policy & fcp.CHECK_LENGTH == fcp.CHECK_LENGTH) and len(path) > MAX_FILE_PATH:
        LOGGER.error('Error: Path length cannot exceed %d characters.', MAX_FILE_PATH)
        return False

    if (check_policy & fcp.CHECK_SYMBOLIC_LINK == fcp.CHECK_SYMBOLIC_LINK) and is_symlink_in_path(path):
        LOGGER.error('Error: the file path contains a symbolic link.')
        return False

    if (check_policy & fcp.CHECK_EXIST == fcp.CHECK_EXIST) and not os.path.isfile(path):
        LOGGER.error('Error: the file not exist.')
        return False

    return True


def is_symlink_in_path(path):
    parts = os.path.normpath(path).split(os.sep)
    for i in range(1, len(parts) + 1):
        current_path = os.sep.join(parts[:i])
        if os.path.islink(current_path):
            return True
    return False


def get_relative_path(path):
    path = os.path.normpath(path)
    if not os.path.isabs(path):
        return path

    path = path.lstrip(os.path.sep)
    next_slash_index = path.find(os.path.sep)
    if next_slash_index != -1:
        return path[next_slash_index + 1:]

    return path
