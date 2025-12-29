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
import logging
import time

import pandas as pd
import numpy as np
from sklearn.preprocessing import MinMaxScaler
from ascend_fd.utils.status import InfoNotFoundError
from ascend_fd.utils.regular_table import PROCESS_VALID_COLUMNS

node_logger = logging.getLogger("NODE_ANOMALY")
MAX_VALUE_PERCENT = 90
SMOOTH_LEN = 5
VALID_PROCESS_RATIO = 0.75


def check_column(process_df: pd.DataFrame):
    """
    Check whether the column name exists
    :param process_df: process data frame
    :return: check result
    """
    process_columns_name = process_df.columns.values
    process_missing_columns = set(PROCESS_VALID_COLUMNS) - set(process_columns_name)
    if process_missing_columns:
        node_logger.error(f"Metrics %s don't exist in npu_detail.csv, please check file.", process_missing_columns)
        raise InfoNotFoundError(f"Metrics {process_missing_columns} don't exist in npu_detail.csv. Please check file.")
    return process_df


def filter_process(process_df: pd.DataFrame):
    """
    Filtering invalid processes
    :param process_df: the process data
    :return: the process data after filter process
    """
    min_process_len = len(set(process_df.loc[:, "time"].to_list())) * VALID_PROCESS_RATIO
    for _, pid_df in process_df.groupby("pid"):
        # if valid pid len is less than 3/4 of the collect time len, drop the pid.
        if len(pid_df.index) < min_process_len:
            process_df = process_df.drop(index=pid_df.index)
    return process_df


def change_time_format(data_frame: pd.DataFrame, feature_name):
    """
    Change the pandas time format to '%Y-%m-00 00:00:00'
    :param data_frame: data frame
    :param feature_name: the time feature column name
    :return: data of time formatted
    """
    origin_format = "%Y-%m-%d %H:%M:%S"
    change_format = "%Y-%m-%d %H:%M:00"
    if data_frame.empty:
        node_logger.error("The process data is abnormal, please check the process.csv file data.")
        raise InfoNotFoundError("The process data is abnormal. Please check the process.csv file data.")

    data_frame[feature_name] = data_frame[feature_name].apply(
        lambda x: time.strftime(change_format, time.strptime(x, origin_format)))
    return data_frame


def smooth_series(series: pd.Series, window=SMOOTH_LEN):
    """
    Use rolling windows to smooth the data.
    :param series: the time series data
    :param window: the window size
    :return: the smoothed data
    """
    median_smooth = series.rolling(window=window, min_periods=1, center=True).median()
    mean_smooth = median_smooth.rolling(window=window, min_periods=1, center=True).mean()
    return mean_smooth


def cpu_normalization(process_df: pd.DataFrame):
    """
    Cpu normalization.
    Group data by 'pid' and perform normalization, and smoothing operations.
    :param process_df: the process data
    :return: the data after normalization
    """
    pid_dfs = []
    minmax_scaler = MinMaxScaler()

    for _, pid_df in process_df.groupby("pid"):
        # upper limit of cpu normalization: 90th value, if value > 90th value, smooth the value.
        cpu_str = 'cpu'
        max_value = np.percentile(pid_df[cpu_str].values, MAX_VALUE_PERCENT)
        pid_df.loc[pid_df[cpu_str] > max_value, cpu_str] = max_value
        # scale (map) data to a fixed range.
        pid_df[cpu_str] = minmax_scaler.fit_transform(np.append(pid_df[cpu_str].values, [0]).reshape(-1, 1))[:-1]
        pid_df[cpu_str] = smooth_series(pid_df[cpu_str])
        pid_dfs.append(pid_df)

    if not pid_dfs:
        node_logger.error("Metrics [cpu] don't exist in process.csv, please check the file.")
        raise InfoNotFoundError("Metrics [cpu] don't exist in process.csv. Please check the file.")
    return pd.concat(pid_dfs)
