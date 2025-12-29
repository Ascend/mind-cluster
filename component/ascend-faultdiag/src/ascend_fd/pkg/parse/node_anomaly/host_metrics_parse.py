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
import os.path
import logging
import re
import time

import pandas as pd

from ascend_fd.utils.tool import safe_read_json
from ascend_fd.pkg.parse.network_congestion.net_parse_job import safe_save_csv
from ascend_fd.model.cfg import ParseCFG
from ascend_fd.utils.status import FileNotExistError, InfoNotFoundError, InfoIncorrectError
from ascend_fd.utils.regular_table import PROCESS_FILE, MEM_USED_FILE

logger = logging.getLogger("NODE_ANOMALY")


def process_time_format(values):
    """
    The timestamp is converted from second to year-month-day hour:minute:second format.
    :param values: 2D list containing timestamps
    """
    if not values:
        raise InfoIncorrectError("Input data list is empty.")
    for value in values:
        new_time = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(value[0]))
        value[0] = new_time
    return values


class ParseHostMetrics:

    def __init__(self, cfg: ParseCFG):
        self.cfg = cfg
        self.save_dir = cfg.output_path
        self.mem_used = pd.DataFrame()
        self.core_num = 0
        self.metrics = {'rss': pd.DataFrame(), 'cpu': pd.DataFrame()}
        self.host_metrics_file = cfg.env_info_saver.get_host_metrics_path() if cfg.env_info_saver else ""

    def parse_data(self):
        """
        Parse host metrics json data
        """
        if not os.path.isfile(self.host_metrics_file):
            logger.warning("Can not find the 'host_metrics.json' file. Please check the input.")
            raise FileNotExistError("Can not find the 'host_metrics.json' file. Please check the input.")
        data_json = safe_read_json(self.host_metrics_file)
        for key, values in data_json.items():
            try:
                values = process_time_format(values)
            except Exception as error:
                logger.warning("Failed to parse the %s data in 'host_metrics.json'. Because of: %s", key, error)
                continue
            if "node_mem_used" in key:
                self.mem_used = pd.DataFrame(values, columns=['time', 'mem_used'])
                continue
            pid_result = re.search(r'node_(\w{3})_(\d{1,10})', key)
            if not pid_result:
                continue
            metrics_name, pid = pid_result[1], pid_result[2]
            if metrics_name not in ['cpu', 'rss']:
                continue
            for data in values:
                data.extend([pid, pid])
            new_metrics = pd.DataFrame(values, columns=['time', metrics_name, 'pid', 'cpu_affinity'])
            self.metrics[metrics_name] = pd.concat([self.metrics.get(metrics_name, pd.DataFrame()), new_metrics],
                                                   ignore_index=True)

        core_num_res = re.search(r'host_metrics_(\d{1,10}).json', self.host_metrics_file)
        if not core_num_res:
            logger.warning("Parse core number failed.")
            return
        self.core_num = core_num_res[1]

    def save_process_csv(self):
        """
        Save process rss and cpu metrics to csv
        :return: Whether the task is successful
        """
        rss = self.metrics.get('rss', pd.DataFrame())
        cpu = self.metrics.get('cpu', pd.DataFrame())
        if rss.empty or cpu.empty:
            logger.warning("Rss or Cpu data is empty.")
            return False
        if rss.shape != cpu.shape:
            logger.warning("Rss and Cpu data shape in 'host_metrics.json' is different.")
            return False
        process = pd.merge(cpu, rss)
        process = process[['time', 'pid', 'cpu', 'rss', 'cpu_affinity']]
        safe_save_csv(process, os.path.join(self.save_dir, PROCESS_FILE + str(self.core_num) + '.csv'),
                      mode="w+", newline="")
        return True

    def save_mem_csv(self):
        """
        Save memory metrics to csv
        :return: Whether the task is successful
        """
        if self.mem_used.empty:
            logger.warning("The mem_used data in 'host_metrics.json' is empty.")
            return False
        safe_save_csv(self.mem_used, os.path.join(self.save_dir, MEM_USED_FILE), mode="w+", newline="")
        return True

    def parse(self):
        self.parse_data()
        process_flag = self.save_process_csv()
        mem_flag = self.save_mem_csv()
        if not process_flag and not mem_flag:
            logger.warning("Can not parse effective data from 'host_metrics.json'.")
            raise InfoNotFoundError("Can not parse effective data from 'host_metrics.json'.")
        logger.info("The host_metrics file is parsed.")
        logger.info("The host_metrics parsing result is saved in dir %s.", os.path.basename(self.save_dir))
