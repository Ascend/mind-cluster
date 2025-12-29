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
import os

import pandas as pd

from ascend_fd.utils.status import InfoIncorrectError
from ascend_fd.utils.tool import CONF_PATH, check_scikit_learn_version
from ascend_fd.pkg.diag.network_congestion.net_diag_job import safe_joblib_load

node_logger = logging.getLogger("NODE_ANOMALY")

ALL_PROCESS_FAULT = 10
SINGE_PROCESS_FAULT = 20
PART_PROCESS_FAULT = 30
RANDOM_PROCESS_FAULT = 80


class AbnormalDetector:
    MODEL_PATH = ""
    MODEL_TYPE = "decision_tree"
    NAME = "base detector"
    FEATURES = set()
    _DECISION_TREE_MODULES = frozenset([('sklearn.ensemble._forest', 'RandomForestClassifier'),
                                        ('sklearn.tree._classes', 'DecisionTreeClassifier'),
                                        ('joblib.numpy_pickle', 'NumpyArrayWrapper'), ('numpy', 'ndarray'),
                                        ('numpy', 'dtype'),
                                        ('numpy.core.multiarray', 'scalar'), ('sklearn.tree._tree', 'Tree')])

    def model_predict(self, x_pred):
        """
        Load the model and predict input data.
        :param x_pred: model features
        :return: model predict result
        """
        check_scikit_learn_version()
        model_path = '%s/%s_%s.pkl' % (self.MODEL_PATH, self.MODEL_TYPE, "latest")
        model = safe_joblib_load(model_path, self._DECISION_TREE_MODULES)
        return model.predict(x_pred)

    def select_and_predict(self, data):
        """
        Select the data column data and predict.
        :param data: the predict data
        :return: the model prediction result
        """
        if data.empty:
            node_logger.error('The predict data is empty when %s detecting, please check input data.', self.NAME)
            raise InfoIncorrectError(f'The predict data is empty when {self.NAME} detecting. Please check input data.')

        columns = set(data.columns.array)
        if not self.FEATURES.issubset(columns):
            node_logger.error(f'Some features[%s] are lost when %s detecting.', self.FEATURES - columns, self.NAME)
            raise InfoIncorrectError(f'Some features[{self.FEATURES - columns}] are lost when {self.NAME} detecting.')

        df_selected = data[list(self.FEATURES)]
        return self.model_predict(df_selected)

    def detect(self):
        pass


class CpuAbnormalDetector(AbnormalDetector):
    MODEL_PATH = os.path.join(CONF_PATH, "model")
    MODEL_TYPE = "cpu_decision_tree"
    NAME = "CPU detector"
    FEATURES = {"cpu"}

    def __init__(self, process_pd):
        """
        Resource preemption data init.
        :param process_pd: process data
        """
        super().__init__()
        self.process_pd = process_pd.copy()

    def detect(self):
        """
        Detect cpu abnormal.
        :return: abnormal cpu detect result list
        """
        time_str = "time"
        all_abnormal_cpu = []
        res_pd = pd.DataFrame()
        time_columns = sorted(list(set(self.process_pd[time_str].tolist())))
        for time in time_columns:
            abnormal_cpu_affinities = self._detect_cpu_in_specified_time(time)
            all_abnormal_cpu.append(abnormal_cpu_affinities)

        process_num = len(set(self.process_pd["cpu_affinity"].tolist()))
        abnormal_cpu_number = self._predict_cpu(all_abnormal_cpu, process_num)

        res_pd[time_str] = time_columns
        res_pd["fault_cpu_flag"] = abnormal_cpu_number
        res_pd["fault_pid"] = all_abnormal_cpu
        res_pd.sort_values(by=time_str, inplace=True)
        res_pd.reset_index(inplace=True, drop=True)
        return res_pd

    def _predict_cpu(self, all_abnormal_cpu, process_num):
        """
        Processes model detect results.
        :param all_abnormal_cpu: the abnormal cpu list
        :param process_num: the process num
        :return: the CPU result after predict and the abnormal process id
        """
        result = list()
        last_time_list = list()

        if process_num == 0:
            node_logger.error('The cpu data incorrect when %s detecting, please check [cpu_affinity].', self.NAME)
            raise InfoIncorrectError(
                f'The cpu data incorrect when {self.NAME} detecting. Please check [cpu_affinity].')

        for cpu_affinity_list in all_abnormal_cpu:
            # no process is preempted at the current time
            if not cpu_affinity_list:
                result.append(0)
            # random preemption
            elif last_time_list and set(last_time_list) != set(cpu_affinity_list):
                result[-1] = RANDOM_PROCESS_FAULT
                result.append(RANDOM_PROCESS_FAULT)
            # only one process is preempted
            elif len(cpu_affinity_list) == 1:
                result.append(SINGE_PROCESS_FAULT)
            # all process is preempted
            elif len(cpu_affinity_list) == process_num:
                result.append(ALL_PROCESS_FAULT)
            # 2-7 process are preempted
            else:
                result.append(PART_PROCESS_FAULT)
            last_time_list = cpu_affinity_list
        return result

    def _detect_cpu_in_specified_time(self, current_time):
        """
        Detect cpu abnormal in specified time(one row data).
        :param current_time: the time to be detected
        :return: abnormal cpu affinities after model detect
        """
        fetch_index = (self.process_pd["time"] == current_time)
        current_time_process_pds = self.process_pd[fetch_index].drop_duplicates(subset="cpu_affinity", keep="first")

        if current_time_process_pds.empty:
            node_logger.error('The data is lost when cpu_affinity drop duplicates in %s time.', current_time)
            raise InfoIncorrectError(f'The data is lost when cpu_affinity drop duplicates in {current_time} time.')

        predict_flag = self.select_and_predict(current_time_process_pds)
        if len(predict_flag) == 0:
            node_logger.error('The predict result is empty when %s detecting, please check predict model.', self.NAME)
            raise InfoIncorrectError(
                f'The predict result is empty when {self.NAME} detecting. Please check predict model.')

        cpu_affinity_list = current_time_process_pds["cpu_affinity"].tolist()
        abnormal_cpu_affinities = [cpu_affinity_list[i] for i, flag in enumerate(predict_flag) if flag != 0]
        return abnormal_cpu_affinities
