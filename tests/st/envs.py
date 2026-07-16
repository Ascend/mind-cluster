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

DEFAULT_WAIT_TIME = "3"

# the place where the testcases are located, usually the testcases directory of the project, e.g.,
# tests/st/testcases.
BASE_DIR = os.getenv("BASE_DIR", None)
# the directory where the kwok node spec templates are located, e.g., tests/st/spec.
SPEC_DIR = os.getenv("SPEC_DIR", None)
# the directory where the valid mindcluster yaml files are located, its usually tested already.
MIND_CLUSTER_YAML_DIR = os.getenv("MIND_CLUSTER_YAML_DIR", None)
# the directory where the pull request output files are located.
PR_OUTPUT_DIR = os.getenv("PR_OUTPUT_DIR", None)
# the ipv4 address of the node
ipv4_address = os.getenv("ipv4_address", None)
# the username of the node
username = os.getenv("username", None)
# the password of the node
password = os.getenv("password", None)
# the logging level of the ssh connection
SSH_LOG_LEVEL = os.getenv("SSH_LOG_LEVEL", "INFO")
BACKUP_YAML_DIR = os.getenv("BACKUP_YAML_DIR", None)
WAIT_TIME = int(os.getenv("WAIT_TIME", DEFAULT_WAIT_TIME))
