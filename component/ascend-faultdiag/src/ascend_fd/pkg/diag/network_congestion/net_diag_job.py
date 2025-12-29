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
import re
import warnings
from functools import partial

import joblib
import numpy as np
import pandas as pd

from ascend_fd.pkg.diag.fault_entity import NET_LINK_CONGESTION_FAULT_ENTITY, NET_DIAG_NORMAL_ENTITY
from ascend_fd.pkg.diag.message import NET_SINGLE_WORKER_MSG
from ascend_fd.pkg.parse.network_congestion.net_parse_job import safe_read_csv
from ascend_fd.utils.regular_table import NIC_OUT_FILENAME
from ascend_fd.utils.status import FileNotExistError, InfoIncorrectError, FileOpenError
from ascend_fd.utils.tool import CONF_PATH, safe_read_open, check_scikit_learn_version
from ascend_fd.model.cfg import DiagCFG

warnings.filterwarnings("ignore", category=UserWarning)
net_logger = logging.getLogger("NET_CONGESTION")


def start_net_diag_job(cfg: DiagCFG):
    """
    Start net diag job
    :param cfg: diag config
    :return: net diag result
    """
    result = {'analyze_success': True}
    result.update(net_congestion_detection(cfg))
    return result


def net_congestion_detection(cfg: DiagCFG):
    """
    Net congestion detection job
    :param cfg: config data
    :return: result for congestion workers and npu links
    """
    net_logger.info("Start network interface congestion fault diagnosis task.")
    processor = NetCongestionDetector()
    input_df = processor.get_nic_data(cfg.parsed_saver.get_all_worker_dir_path())
    if processor.check_single_worker():
        net_logger.info('Network congestion does not occur in single worker.')
        return {"fault": [NET_DIAG_NORMAL_ENTITY.attribute], "note_msgs": NET_SINGLE_WORKER_MSG}
    return {"fault": [processor.detection(input_df)]}


_WORKER_NUM_PATTERN = re.compile(r"([\w\-]{1,255})(\d{1,3})")
MAX_LEN_WORKER_NAME = 258
MAX_LEN_WORKER_NAME_PRINT = 50


def _fault_detail_sort_rule(item: dict):
    worker_name = item.get("worker", "")
    search_res = _WORKER_NUM_PATTERN.search(worker_name)
    if search_res:
        return search_res[1], int(search_res[2])
    return worker_name, -1


class NetCongestionDetector:
    PRE_LABEL_ABNORMAL = 1
    INPUT_COL_NAMES = [
        'mac_tx_pfc_pkt_num', 'mac_rx_pfc_pkt_num', 'roce_rx_rc_pkt_num', 'roce_rx_all_pkt_num',
        'roce_tx_rc_pkt_num', 'roce_tx_all_pkt_num', 't_congestion_index', 'r_congestion_index'
    ]
    NET_LATEST_MODEL_PATH = os.path.join(CONF_PATH, 'model', "net_rf_model_latest.pt")  # model for sklearn >= 1.3.0
    NET_OLD_MODEL_PATH = os.path.join(CONF_PATH, 'model', "net_rf_model_102.pt")  # model for sklearn < 1.3.0
    _NET_RF_MODEL_MODULES = frozenset([('sklearn.ensemble._forest', 'RandomForestClassifier'),
                                       ('sklearn.tree._classes', 'DecisionTreeClassifier'),
                                       ('joblib.numpy_pickle', 'NumpyArrayWrapper'), ('numpy', 'ndarray'),
                                       ('numpy', 'dtype'),
                                       ('numpy.core.multiarray', 'scalar'), ('sklearn.tree._tree', 'Tree')])

    def __init__(self):
        self.worker_num = -1
        self.model = None
        self._model_load()

    @staticmethod
    def _wrap_fault_detail(detect_result: dict):
        fault_detail = list()
        for worker, device_list in detect_result.items():
            if len(worker) >= MAX_LEN_WORKER_NAME:
                worker = worker[:MAX_LEN_WORKER_NAME]
                net_logger.warning("worker name: %s... exceeds the maximum length limit, it will be truncated.",
                                   worker[:MAX_LEN_WORKER_NAME_PRINT])
            fault_detail.append({
                'worker': worker,
                'device_list': device_list
            })
        return sorted(fault_detail, key=_fault_detail_sort_rule)

    @staticmethod
    def _remove_nan_row(npu_df: pd.DataFrame) -> pd.DataFrame:
        """
        Remove the row containing nan. Rows containing Nan may be generated after feature extraction
        :param npu_df:
        :return:  dataframe after deleting the row containing nan.
        """
        npu_df.replace([np.inf, -np.inf], np.nan)  # replace inf data to NaN
        drop_na_df = npu_df.dropna(axis=0, how='any')  # drop all rows that have any NaN values
        if drop_na_df.empty:
            net_logger.error("Nan data exists in all rows, please check whether the values is valid.")
            raise InfoIncorrectError("Please check whether the values is valid.")
        if len(drop_na_df) != len(npu_df):
            net_logger.warning("Please check whether the values is valid in some rows.")
        return drop_na_df

    def get_nic_data(self, worker_path_dict) -> pd.DataFrame:
        """
        Get and read 'NIC_OUT_FILENAME' csv file
        :param: worker_path_dict: worker path directory, the format of key is "{worker_id}", e.g. "0"
        """
        self.worker_num = len(worker_path_dict)
        nic_df_list = list()

        for worker, worker_dir in worker_path_dict.items():
            nic_file = os.path.join(worker_dir, NIC_OUT_FILENAME)
            if not os.path.exists(nic_file):
                net_logger.warning(f"The %s don't have %s.", worker, NIC_OUT_FILENAME)
                continue
            nic_df = safe_read_csv(nic_file, dtype=self._get_df_column_type(), header=0,
                                   encoding='unicode_escape')
            nic_df["worker_name"] = worker
            if not self._check_nic_df(nic_df, worker):
                continue
            nic_df_list.append(nic_df)

        if not nic_df_list:
            net_logger.error("No nic_clean csv file that meets the path specification is found.")
            raise FileNotExistError("No nic_clean csv file that meets the path specification is found.")

        concat_df = pd.concat(nic_df_list)  # concat_df index format: worker-{worker_id}_rank-{rank_id}
        return self._remove_nan_row(concat_df.sort_index())

    def detection(self, data: pd.DataFrame):
        """
        detect congested links
        :param data: input dataframe
        :return: fault attribute, contain fault_details
        """
        detect_result = dict()

        for worker_name, device_name, pre_label in zip(list(data['worker_name']), list(data['device_id']),
                                                       self.model.predict(data[self.INPUT_COL_NAMES]).astype(int)):
            if pre_label != self.PRE_LABEL_ABNORMAL:
                continue
            detect_result.setdefault(worker_name, list()).append(device_name)

        if not detect_result:
            return NET_DIAG_NORMAL_ENTITY.attribute
        NET_LINK_CONGESTION_FAULT_ENTITY.update_attribute({"fault_details": self._wrap_fault_detail(detect_result)})
        return NET_LINK_CONGESTION_FAULT_ENTITY.attribute

    def check_single_worker(self):
        """
        Check whether it's a single worker environment
        """
        return self.worker_num == 1

    def _get_df_column_type(self):
        """
        Get dataframe column dtype, to restrict the type of input data
        :return: data type dictionary.
        """
        data_type = {'device_id': str}
        for metric in self.INPUT_COL_NAMES:
            data_type[metric] = np.float64
        return data_type

    def _check_nic_df(self, nic_df, worker) -> bool:
        """
        check whether metric data type is valid
        :param nic_df: origin dataframe
        :param worker: worker id
        :return: is the dataframe valid
        """
        if nic_df.empty:
            net_logger.warning(f"Data in %s is empty.", worker)
            return False

        df_columns_name = nic_df.columns.values
        missing_columns = (set(self.INPUT_COL_NAMES) | {"device_id"}) - set(df_columns_name)
        if missing_columns:
            net_logger.warning(f"Metrics %s don't exist in %s npu_detail.csv.", missing_columns, worker)
            return False
        return True

    def _model_load(self):
        """
        Load network congestion model
        """
        check_scikit_learn_version()
        self.model = safe_joblib_load(self.NET_LATEST_MODEL_PATH, self._NET_RF_MODEL_MODULES)


class RestrictedUnpickler(joblib.numpy_pickle.NumpyUnpickler):

    def __init__(self, filename, file_handle, mmap_mode=None, supported_module_set=None):
        super().__init__(filename, file_handle, mmap_mode)
        self._supported_module_set = supported_module_set or {}

    def find_class(self, module, name):
        if (module, name) not in self._supported_module_set:
            raise ImportError("Restricted unpickling; cannot load module: {}, name: {}".format(module, name))
        return super().find_class(module, name)


def safe_joblib_load(model_path: str, supported_module_set):
    """
    Safe joblib load model
    Will use safe joblib func to load model
    :param model_path: model path
    :param supported_module_set: supported model file module set
    :return: model file
    """
    with safe_read_open(model_path, 'rb') as file_stream:
        origin_unpickler = joblib.numpy_pickle.NumpyUnpickler
        try:
            joblib.numpy_pickle.NumpyUnpickler = partial(RestrictedUnpickler, supported_module_set=supported_module_set)
            model = joblib.load(file_stream)
        except Exception as err:
            raise FileOpenError(f"Failed to load model {os.path.basename(model_path)}: {err}") from err
        finally:
            joblib.numpy_pickle.NumpyUnpickler = origin_unpickler
    return model
