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
import re
import logging

import pandas as pd

from ascend_fd.utils.tool import safe_list_dir
from ascend_fd.utils.regular_table import PROCESS_CORE_CSV
from ascend_fd.pkg.parse.network_congestion.net_parse_job import safe_read_csv
from ascend_fd.utils.status import FileNotExistError, PathError
from ascend_fd.pkg.diag.node_anomaly.resource_preemption.utils import (CpuAbnormalDetector,
                                                                       change_time_format, ResultProbComputer,
                                                                       cpu_normalization, ResultSmoother,
                                                                       FaultTimePeriodFormatter, wrap_error_result,
                                                                       check_column, filter_process)

node_logger = logging.getLogger("NODE_ANOMALY")


def resource_preemption_job(worker_dir, worker_name):
    """
    The resource preemption detect job. It includes the following four parts:
    1. preprocess the input data;
    2. use model to predict;
    3. organize predict output and extract the detail error information;
    4. format the error result
    :param worker_dir: the worker dir path
    :param worker_name: the worker name
    :return: resource preemption fault code and fault details info
    """
    process_name = ""
    # match the process.csv file from the parse dir.
    if not worker_dir or not os.path.isdir(worker_dir):
        node_logger.error("The %s is empty or is not a directory.", worker_dir)
        raise PathError(f"The {worker_dir} is empty or is not a directory.")
    for parse_file in safe_list_dir(worker_dir):
        process_re = re.match(PROCESS_CORE_CSV, parse_file)
        if not process_re:
            continue
        process_name = process_re[0]
        break
    if not process_name:
        node_logger.warning("The %s don't have process_num.csv file.", worker_name)
        raise FileNotExistError(f"The {worker_name} don't have process_num.csv file.")
    process_file = os.path.join(worker_dir, process_name)

    process_df = safe_read_csv(process_file)
    process_df = preprocess_data(process_df)
    detect_result_list = CpuAbnormalDetector(process_df).detect()
    error_info_list = _organize_and_extract_result(detect_result_list)
    return wrap_error_result(error_info_list, worker_name)


def preprocess_data(process_df: pd.DataFrame):
    """
    Preprocess the input data: process data
    1. check whether the column exists;
    2. check and change time format;
    3. cpu column normalization.
    :param process_df: the process data
    :return: process data after process
    """
    check_column(process_df)
    process_df = filter_process(process_df)
    process_df = change_time_format(process_df, "time")
    process_df = cpu_normalization(process_df)
    return process_df


def _organize_and_extract_result(res_list):
    """
    Organize the model output, extract the key info and return the format result of cpu,
    merge res_list by time column and deal with abnormal result
    :param res_list: cpu_res model predict result list
    :return: error info list
    """
    smooth_res = ResultSmoother().cpu_smooth(res_list)
    compute_prob_res = ResultProbComputer().compute(smooth_res, "cpu")
    error_info_list = FaultTimePeriodFormatter().format(compute_prob_res, "cpu")
    return error_info_list
