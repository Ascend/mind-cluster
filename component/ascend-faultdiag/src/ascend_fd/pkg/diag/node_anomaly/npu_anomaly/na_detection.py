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
import logging

import numpy as np
import pandas as pd

from ascend_fd.pkg.parse.network_congestion.net_parse_job import safe_read_csv
from ascend_fd.utils.regular_table import NAD_OUT_FILENAME
from ascend_fd.utils.fault_code import NODE_DIAGNOSIS_NORMAL
from ascend_fd.utils.status import InfoNotFoundError, FileNotExistError
from ascend_fd.pkg.diag.fault_entity import NODE_DIAG_NORMAL_ENTITY, NPU_OVER_TEMPERATURE_ENTITY, \
    NPU_STATUS_ABNORMAL_ENTITY

node_logger = logging.getLogger("NODE_ANOMALY")


class NPUChecker:
    RATE_FREQ_THRESHOLD = 1000
    TEMP_THRESHOLD = 105
    PERSISTENT_LEN = 4
    NAD_VALID_COLUMNS = {"time", "dev_id", "power", "freq", "temp"}

    def __init__(self, npu_data: pd.DataFrame):
        """
        NPU checker
        :param npu_data: the dataframe of npu data
        """
        self.npu_data = npu_data

        self.anomaly_npu = pd.DataFrame()
        self.anomaly_index = []
        self.fault_detail = dict()

        self.rate_freq = self.RATE_FREQ_THRESHOLD
        if 'rated_freq' in npu_data.columns:
            self.rate_freq = np.max(npu_data['rated_freq'])

    def detection(self):
        """
        Detect whether NPU overload frequency reduction occurs
        :return: fault code and fault details info
        """
        missing_columns = self.NAD_VALID_COLUMNS - set(self.npu_data.columns)
        if missing_columns:
            node_logger.error("Metrics %s don't exist in nad_clean.csv, please check the file.", missing_columns)
            raise InfoNotFoundError("Metrics %s don't exist in nad_clean.csv, please check the file." % missing_columns)
        self.fault_detail = self.get_fault_detail_info()
        if not self.fault_detail:
            return NODE_DIAG_NORMAL_ENTITY, self.fault_detail
        if not self.temp_high_check():
            return NPU_STATUS_ABNORMAL_ENTITY, self.fault_detail
        return NPU_OVER_TEMPERATURE_ENTITY, self.fault_detail

    def get_fault_detail_info(self) -> dict:
        """
        Obtain data of npus that frequency is below the threshold
        :return: fault details info
        """
        aicore_rate_str = 'aicore_rate'
        anomaly_npu_all = self.npu_data[self.npu_data['freq'] < self.rate_freq]
        if aicore_rate_str in anomaly_npu_all.columns:  # delete the range where the aicore usage is 0.
            anomaly_npu_all[aicore_rate_str] = anomaly_npu_all[aicore_rate_str].astype(int)
            anomaly_npu_all = anomaly_npu_all[anomaly_npu_all[aicore_rate_str] > 0]
        if anomaly_npu_all.empty:
            return dict()
        anomaly_period = self._get_anomaly_period(anomaly_npu_all)
        if not anomaly_period:
            return dict()
        return {
            "periods": anomaly_period
        }

    def temp_high_check(self):
        """
        Obtain data of npus that temperature is too high in the fault time interval
        :return: Whether there are overly high temperature NPUs
        """
        temp_high = self.anomaly_npu[(self.anomaly_npu['temp'] >= self.TEMP_THRESHOLD)]
        return not temp_high.empty

    def _get_anomaly_period(self, anomaly_npu_all: pd.DataFrame):
        """
        Get the anomaly period of npu. Only frequency reduction lasting for one minute is considered really reduction
        :param anomaly_npu_all: all anomaly npu that freq lower than threshold
        :return: anomaly period list
        """
        period_list = []
        anomaly_index_all = list(anomaly_npu_all.index)
        if len(anomaly_index_all) < self.PERSISTENT_LEN:
            return period_list

        anomaly_index_all.append(-1)  # add a final index flag
        pre_index = anomaly_index_all[0]
        continuity_length = 1
        for now_index in anomaly_index_all[1:]:
            if now_index == pre_index + 1:
                continuity_length += 1
                pre_index = now_index
                continue
            if continuity_length >= self.PERSISTENT_LEN:
                period_list.append([anomaly_npu_all["time"].get(pre_index - continuity_length + 1),
                                    anomaly_npu_all["time"].get(pre_index)])
                self.anomaly_index.extend([idx for idx in range(pre_index - continuity_length + 1, pre_index + 1)])
            continuity_length = 1
            pre_index = now_index

        self.anomaly_npu = self.npu_data.iloc[self.anomaly_index]
        return period_list


def npu_anomaly_job(worker_dir: str, worker_name: str):
    """
    Npu anomaly detection for one node
    :param worker_dir: the worker dir path
    :param worker_name: the worker name
    :return: fault code and fault details info
    """
    dev_id_key = "dev_id"
    result = dict()
    npu_file = os.path.join(worker_dir, NAD_OUT_FILENAME)
    if not os.path.exists(npu_file):
        node_logger.warning("%s don't have npu file.", worker_name)
        raise FileNotExistError("%s don't have npu file." % worker_name)

    npu_data = safe_read_csv(npu_file, dtype={"time": str, dev_id_key: int, "power": float,
                                              "freq": int, "temp": int})
    npu_set = set(npu_data[dev_id_key])
    for npu_id in npu_set:
        single_npu_data = npu_data[npu_data[dev_id_key] == npu_id].reset_index(drop=True)
        npu_checker = NPUChecker(single_npu_data)
        fault_entity, fault_detail = npu_checker.detection()
        if fault_entity.code == NODE_DIAGNOSIS_NORMAL:
            result[fault_entity] = []
            continue
        fault_detail.update({"device": (worker_name, npu_id)})
        result.setdefault(fault_entity, []).append(fault_detail)
    return result
