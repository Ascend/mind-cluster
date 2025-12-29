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
import time
import logging

import pandas as pd

from ascend_fd.utils.status import FileNotExistError, InfoNotFoundError
from ascend_fd.utils.tool import MultiProcessJob
from ascend_fd.model.cfg import ParseCFG
from ascend_fd.pkg.parse.network_congestion.net_parse_job import safe_read_csv, safe_save_csv
from ascend_fd.utils.regular_table import NAD_OUT_FILENAME
from ascend_fd.pkg.parse.node_anomaly.host_metrics_parse import ParseHostMetrics

node_logger = logging.getLogger("NODE_ANOMALY")


def start_node_parse_job(cfg: ParseCFG):
    """
    Start node parse job
    :param cfg: parse config
    """
    node_logger.info("Start node anomaly parsing job.")
    jobs = {
        "NPU_ANOMALY_PARSE": NpuAnomalyParser(cfg).parse,
        "HOST_METRICS_PARSE": ParseHostMetrics(cfg).parse
    }
    multiprocess_job = MultiProcessJob("NODE_ANOMALY", pool_size=2, task_id=cfg.task_id)
    for job_name, job_func in jobs.items():
        multiprocess_job.add_security_job(job_name, job_func)
    multiprocess_job.join_and_get_results()
    node_logger.info("Node anomaly parsing job is complete.")


class NpuAnomalyParser:
    NAD_VALID_COLUMNS = {"time", "dev_id", "power", "freq", "temp"}
    NAD_SORT_COLUMNS = ["dev_id", "time"]

    def __init__(self, cfg: ParseCFG):
        """
        Npu Anomaly Parser
        :param cfg: parse config
        """
        self.cfg = cfg
        self.output_dir = cfg.output_path
        self.output_npu_file = os.path.join(self.output_dir, NAD_OUT_FILENAME)
        self.npu_smi_files = cfg.env_info_saver.get_npu_smi_detail_list() if cfg.env_info_saver else []

    @staticmethod
    def _parse_single_file(file: str, valid_col: set):
        """
        Parse single npu detail file
        :param file: single npu smi file
        :param valid_col: valid columns for csv
        :return: parsed dataframe
        """
        time_key = "time"
        data_frame = safe_read_csv(file, dtype={time_key: int, "dev_id": int, "power": float,
                                                "freq": int, "temp": int})
        if data_frame.empty:
            node_logger.warning("No data in %s, please check file.", file)
            raise InfoNotFoundError("No data in %s. Please check file." % file)

        missing_columns = valid_col - set(data_frame.columns)
        if missing_columns:
            node_logger.warning("Metrics %s don't exist in %s, please check file.",
                              missing_columns, os.path.basename(file))
            raise InfoNotFoundError("Metrics %s don't exist in %s. Please check file."
                                    % (missing_columns, os.path.basename(file)))
        data_frame[time_key] = data_frame[time_key].apply(
            lambda x: time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(x)))
        return data_frame

    def parse(self):
        """
        Executing npu anomaly parsing
        """
        node_logger.info("Start parse npu smi file.")
        if not self.npu_smi_files:
            node_logger.warning("No npu_smi csv file that meets the path specification is found.")
            raise FileNotExistError("No npu_smi csv file that meets the path specifications is found")

        npu_concat_list = []
        for file in self.npu_smi_files:
            npu_df = self._parse_single_file(file, self.NAD_VALID_COLUMNS)
            npu_concat_list.append(npu_df)
        npus_df = pd.concat(npu_concat_list)
        npus_df.sort_values(by=self.NAD_SORT_COLUMNS, inplace=True)
        safe_save_csv(npus_df, self.output_npu_file, mode="w+", newline="")
        node_logger.info("The npu smi detail files is parsed.")

        node_logger.info(
            "The npu_anomaly parsing result is saved in dir %s.",
            os.path.basename(self.output_dir)
        )
