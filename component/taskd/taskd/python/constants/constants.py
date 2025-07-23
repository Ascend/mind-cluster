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

# build dependency
LIB_SO_NAME = 'libtaskd.so'
LIB_SO_PATH = 'libs'

# log dir permission
LOG_PRIVILEGE = 0o640
MUSK_PRIVILEGE = 0o777
LOG_DIR_PRIVILEGE = 0o750
LOG_BAK_PRIVILEGE = 0o400

# log level for flush into log file
TASKD_FILE_LOG_LEVEL = "TASKD_FILE_LOG_LEVEL"
# log level for print to screen
TASKD_STD_LOG_LEVEL = "TASKD_STD_LOG_LEVEL"
# to determine whether to print log into screen
TASKD_LOG_STDOUT = 'TASKD_LOG_STDOUT'
TASKD_LOG_PATH = 'TASKD_LOG_PATH'

# logger default config
LOG_MAX_LINE_LENGTH = 1023
LOG_SIMPLE_FORMAT = '[%(levelname)s]     %(asctime)s.%(msecs)06d %(process)d   %(filename)s:%(lineno)d     %(message)s'
LOG_DATE_FORMAT = '%Y/%m/%d %H:%M:%S'
LOG_BACKUP_FORMAT = '%Y-%m-%dT%H-%M-%S.%f'
LOG_BACKUP_PATTERN = '\\d{4}-\\d{2}-\\d{2}T\\d{2}-\\d{2}-\\d{2}\\.\\d{3}'
LOG_DEFAULT_FILE = "./taskd_log/taskd.log"
LOG_DEFAULT_FILE_PATH = "./taskd_log/"
LOG_DEFAULT_FILE_NAME = "taskd.log"
LOG_DEFAULT_BACKUP_COUNT = 30
LOG_DEFAULT_MAX_BYTES = 1024 * 1024 * 20

# valid boundary value
MIN_RANK_SIZE = 0
MAX_RANK_SIZE = 4095
MAX_FILE_NUMS = 4096
MIN_DEVICE_NUM = 1
MAX_DEVICE_NUM = 4096
MAX_SIZE = 1024 * 1024
MIN_SIZE = 0

# AI framework
PYTORCH = "pytorch"
MINDSPORE = "mindspore"

# check step out period
CHECK_STEP_PERIOD = 10

# jobid env key
JOB_ID_KEY = "MINDX_TASK_ID"

# default group info base dir
DEFAULT_GROUP_DIR = "/user/cluster-info/profiling"

# group info dir mode
PROFILING_DIR_MODE = 0o755

# group info constants
GROUP_INFO_NAME = "parallel_group.json"
GROUP_INFO_KEY = "parallel_group_info"
GROUP_NAME_KEY = "group_name"
GROUP_RANK_KEY = "group_rank"
GLOBAL_RANKS_KEY = "global_ranks"
DEFAULT_GROUP = "default_group"
GROUP_BASE_DIR_ENV = "GROUP_BASE_DIR"