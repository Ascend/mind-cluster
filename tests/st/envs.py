#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
import os
from typing import Dict, Callable, Any

env_variables: Dict[str, Callable[[], Any]] = {
    # the place where the testcases are located, usually the testcases directory of the project, e.g.,
    # tests/st/testcases.
    "BASE_DIR":
        lambda: os.getenv("BASE_DIR", None),
    # the directory where the valid mindcluster yaml files are located, its usually tested already
    "MIND_CLUSTER_YAML_DIR":
        lambda: os.getenv("MIND_CLUSTER_YAML_DIR", None),
    # the directory where the pull request output files are located.
    "PR_OUTPUT_DIR":
        lambda: os.getenv("PR_OUTPUT_DIR", None),
    # the ipv4 address of the node
    "ipv4_address":
        lambda: os.getenv("ipv4_address", None),
    # the username of the node
    "username":
        lambda: os.getenv("username", None),
    # the password of the node
    "password":
        lambda: os.getenv("password", None),
    # the logging level of the ssh connection
    "SSH_LOG_LEVEL":
        lambda: os.getenv("SSH_LOG_LEVEL", "INFO"),
}


def __getattr__(name: str):
    # lazy evaluation of environment variables
    if name in env_variables:
        return env_variables[name]()
    raise AttributeError(f"module {__name__!r} has no attribute {name!r}")


def __dir__():
    return list(env_variables.keys())
