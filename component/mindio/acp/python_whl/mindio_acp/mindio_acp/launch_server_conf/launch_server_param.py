#!/usr/bin/env python
# coding=utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.
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
from mindio_acp.common import utils
from mindio_acp.common.utils import FileCheckPolicy as fcp

# default server worker dir， for example : ~/.mindio/server
home_dir = os.path.expanduser("~")
DEFAULT_SERVICE_REPOSITORY = '.mindio'
service_repository = os.path.join(home_dir, DEFAULT_SERVICE_REPOSITORY)
DEFAULT_SERVICE_NAME = 'server'
server_worker_dir_default = os.path.join(service_repository, DEFAULT_SERVICE_NAME)
server_worker_dir = os.environ.get("MINDIO_ACP_LOG_PATH", server_worker_dir_default)

current_path = os.path.abspath(__file__)
current_dir = os.path.dirname(current_path)
ockiod_path = os.path.join(os.path.dirname(current_dir), "bin", "ockiod")

if not utils.file_path_check(server_worker_dir, fcp.CHECK_NOT_EMPTY | fcp.CHECK_LENGTH | fcp.CHECK_SYMBOLIC_LINK):
    raise FileNotFoundError

if not utils.file_path_check(ockiod_path, fcp.CHECK_NOT_EMPTY | fcp.CHECK_LENGTH | fcp.CHECK_SYMBOLIC_LINK):
    raise FileNotFoundError

server_worker_dir = os.path.realpath(server_worker_dir)
ockiod_path = os.path.realpath(ockiod_path)

# server ockiod lib
lib_path = os.path.join(os.path.dirname(current_dir), "lib")
if "LD_LIBRARY_PATH" in os.environ:
    os.environ["LD_LIBRARY_PATH"] = lib_path + ":" + os.environ["LD_LIBRARY_PATH"]
else:
    os.environ["LD_LIBRARY_PATH"] = lib_path
