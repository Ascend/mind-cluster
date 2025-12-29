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
class BaseError(Exception):
    code = 500
    description = "Internal Server Error."

    def __init__(self, description=None):
        if description:
            self.description = description
        super().__init__(description)

    def __str__(self):
        return f"{self.__class__.__name__}({self.code}): {self.description}"


class PathError(BaseError):
    code = 501
    description = "Invalid Path."


class FileNotExistError(BaseError):
    code = 502
    description = "File not exist."


class InfoNotFoundError(BaseError):
    code = 503
    description = "Information not found."


class InfoIncorrectError(BaseError):
    code = 504
    description = "Information not correct."


class FileOpenError(BaseError):
    code = 505
    description = "Open file failed."


class InnerError(BaseError):
    code = 506
    description = "Inner error."


class ParamError(BaseError):
    code = 507
    description = "ParamError."


class FileTooLarge(BaseError):
    code = 508
    description = "The number of files is too large."


class SuccessRet:
    code = 200
    description = "Successful operation."
