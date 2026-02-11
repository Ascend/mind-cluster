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
import os

LANGUAGE = "zh"
NAME = "ascend"
COMPONENT = "ascend-fd"
DEFAULT_HOME_PATH = os.path.realpath(os.path.expanduser("~/.ascend_faultdiag/"))
ENV_VAR_KEY = "ASCEND_FD_HOME_PATH"
HOME_PATH = os.path.realpath(os.environ.get(ENV_VAR_KEY, DEFAULT_HOME_PATH))
DEFAULT_USER_CONF = os.path.join(HOME_PATH, "custom-ascend-kg-config.json")
CUSTOM_CONFIG_PATH = os.path.join(HOME_PATH, "custom-fd-config.json")
RUN_LOG_FORMAT = "ascend_faultdiag_{}.log"
OP_LOG_FILE = "ascend_faultdiag_operation.log"
RC_PARSER_DUMP_NAME = "ascend-rc-parser.json"
KG_PASER_DUMP_NAME = "ascend-kg-parser.json"
KG_ANALYZER_DUMP_NAME = "ascend-kg-analyzer.json"
AICORE_ERRCODE_CONFIG = "aicore-error-code-config-zh.json"
KNOWLEDGE_GRAPH_CONF = os.path.join(os.path.dirname(os.path.realpath(__file__)), "kg-config.json")
