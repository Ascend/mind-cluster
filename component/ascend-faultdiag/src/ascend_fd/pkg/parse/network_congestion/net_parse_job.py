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
import csv
import logging
import os
import re

import numpy as np
import pandas as pd

from ascend_fd.utils.tool import safe_read_open, safe_write_open
from ascend_fd.model.cfg import ParseCFG
from ascend_fd.utils.regular_table import NPU_DETAILS_CSV, NIC_OUT_FILENAME
from ascend_fd.utils.status import FileNotExistError, InfoNotFoundError, InfoIncorrectError, InnerError

net_logger = logging.getLogger("NET_CONGESTION")


def start_net_parse_job(cfg: ParseCFG):
    """
    start net parse job
    """
    net_logger.info("Start net parse job.")
    parse_df = NetParser(cfg).parse()
    safe_save_csv(parse_df, os.path.join(cfg.output_path, NIC_OUT_FILENAME), mode='w+', newline='')
    net_logger.info("The parsing result is saved in dir %s.", os.path.basename(cfg.output_path))


class NetParser:
    NIC_VALID_COLUMNS_TYPE = np.int64
    NIC_VALID_COLUMNS = [
        'mac_tx_pfc_pkt_num', 'mac_rx_pfc_pkt_num', 'mac_tx_total_pkt_num', 'mac_rx_total_pkt_num',
        'roce_rx_rc_pkt_num', 'roce_rx_all_pkt_num', 'roce_tx_rc_pkt_num', 'roce_tx_all_pkt_num'
    ]

    def __init__(self, cfg: ParseCFG):
        self.cfg = cfg
        self.file_dict = dict()
        self.get_parse_files()

    @staticmethod
    def feature_extract_cal_percentage(npu_df: pd.DataFrame):
        # 1. calculate the congestion factor. when the divisor is 0, the result is NaN.
        npu_df['t_congestion_index'] = npu_df['mac_rx_pfc_pkt_num'] / npu_df['roce_tx_all_pkt_num']
        npu_df['r_congestion_index'] = npu_df['mac_tx_pfc_pkt_num'] / npu_df['roce_rx_all_pkt_num']

        # 2. calculate the percentage of received or send packets
        for direct in ["rx", "tx"]:
            # mac_rx/tx_total_pkt_num: MAC接收/发送总报文数
            mac_total_pkt = npu_df.loc[:, f'mac_{direct}_total_pkt_num'].copy()
            cols = [col for col in npu_df.columns if direct in col]
            for col in cols:
                # when the divisor is 0, the result is NaN.
                npu_df.loc[:, col] = npu_df.loc[:, col] / mac_total_pkt
            npu_df = npu_df.drop(columns=[f'mac_{direct}_total_pkt_num'])
        return npu_df

    @staticmethod
    def feature_extract_cal_mean(npu_df: pd.DataFrame) -> pd.DataFrame:
        """
        Feature extract
        :param npu_df: dataframe
        :return: dataframe after feature extraction, only one row feature data.
        """
        # 1. replaced NAN data
        # the NAN data replaced with the next data by bfill(), and the last row replaced with the before data by ffill()
        npu_df[npu_df < 0] = np.nan
        npu_df = npu_df.bfill().ffill()

        # 2. mean calculation
        # compress data shape from (n, m) to (1, m), 'n' means the number of data , 'm' means the number of metrics
        npu_df = npu_df.mean(axis=0).to_frame().T
        return npu_df

    @staticmethod
    def filter_valid_data_interval(npu_df: pd.DataFrame) -> pd.DataFrame:
        """
        Filter valid data interval
        :param npu_df: dataframe
        :return: dataframe after filtering valid data
        """
        # difference calculation.
        npu_df = npu_df.diff()

        metric = "roce_tx_rc_pkt_num"  # ROCEE发送的RC类型报文数
        first_nonzero_index = npu_df[metric].iloc[::].ne(0).idxmax()
        last_nonzero_index = npu_df[metric].iloc[::-1].ne(0).idxmax()
        interval_len = last_nonzero_index - first_nonzero_index

        # 平滑范围边界，保留部分"roce_tx_rc_pkt_num"为0的数据, r_expend_len为右边界扩展的长度
        r_expend_len = 10
        interval_len = ((interval_len + r_expend_len // 2) // r_expend_len + 1) * r_expend_len
        end_index = last_nonzero_index + r_expend_len
        if len(npu_df) <= end_index:
            end_index = last_nonzero_index
        start_index = max(end_index - interval_len, 0)
        return npu_df[start_index:end_index].reset_index(drop=True)

    def drop_invalid_metric(self, npu_df: pd.DataFrame) -> pd.DataFrame:
        """
        Drop invalid metric and check whether metric data type is valid
        :param npu_df: origin dataframe
        :return: dataframe after dropping invalid metric
        """
        df_columns_name = npu_df.columns.values
        missing_columns = set(self.NIC_VALID_COLUMNS) - set(df_columns_name)
        if missing_columns:
            net_logger.warning("Metrics %s don't exist in npu_detail.csv, please check file.", missing_columns)
            raise InfoNotFoundError(f"Metrics {missing_columns} don't exist in npu_detail.csv. Please check file.")

        for metric in self.NIC_VALID_COLUMNS:
            if not np.issubdtype(npu_df[metric].dtypes, self.NIC_VALID_COLUMNS_TYPE):
                net_logger.warning("Metric [%s] has non-int64 invalid data, please check file.", metric)
                raise InfoIncorrectError(f"Metric {metric} has non-int64 invalid data. Please check file.")
        return npu_df[self.NIC_VALID_COLUMNS]

    def parse_single_file(self, device_id, npu_file):
        """
        Start parse single npu details file
        :param device_id: device id, the format is 'device-{device id}'
        :param npu_file: npu details file path
        :return: dataframe of single npu details file
        """
        npu_df = safe_read_csv(npu_file)
        if npu_df.empty:
            net_logger.error("No data in %s, please check file.", os.path.basename(npu_file))
            raise InfoNotFoundError(f"No data in {os.path.basename(npu_file)}, please check file.")

        npu_df = self.drop_invalid_metric(npu_df)
        npu_df = self.filter_valid_data_interval(npu_df)
        npu_df = self.feature_extract_cal_mean(npu_df)
        npu_df = self.feature_extract_cal_percentage(npu_df)
        npu_df.insert(0, 'device_id', device_id)
        return npu_df

    def parse(self) -> pd.DataFrame:
        """
        Start parse all npu details files
        :return: parse result
        """
        net_logger.info("Start parse npu_details csv files.")
        parse_df_list = list()
        for device_id, npu_file in self.file_dict.items():
            parse_df_list.append(self.parse_single_file(device_id, npu_file))
            net_logger.info("%s npu_detail csv file parse succeed.", device_id)
        return pd.concat(parse_df_list, axis=0).sort_values('device_id')

    def get_parse_files(self):
        """
        Get parse files
        npu_details csv files directory structure:
        ———— environment_check
           ———— {worker_name}
             ———— npu_{rank_id}_details.csv
        """
        npu_detail_files = self.cfg.env_info_saver.get_npu_detail_list() if self.cfg.env_info_saver else []
        for file in npu_detail_files:
            device_re = re.match(NPU_DETAILS_CSV, os.path.basename(file))
            if not device_re:
                continue
            self.file_dict[f"device-{device_re[1]}"] = file

        if not self.file_dict:
            net_logger.error("No npu_detail csv file that meets the path specifications is found.")
            raise FileNotExistError("No npu_detail csv file that meets the path specifications is found.")


def safe_read_csv(file: str, *args, **kwargs) -> pd.DataFrame:
    """
    Safe read csv file by pandas.
    Will use safe open func to open csv file
    :param file: file path
    :param args: func parameters
    :param kwargs: func parameters
    :return: pandas DataFrame
    """
    with safe_read_open(file) as file_stream:
        try:
            dataframe = pd.read_csv(file_stream, *args, **kwargs)
        except Exception as err:
            raise InnerError(f"Open {os.path.basename(file)} csv to pandas failed: {err}") from err
    return dataframe


def safe_save_csv(data_frame: pd.DataFrame, file_path: str, *args, **kwargs):
    """
    Safe save csv file by pandas.
    :param data_frame: pandas data frame
    :param file_path: save file path
    :param args: func parameters
    :param kwargs: func parameters
    """
    df_columns = data_frame.columns.values.tolist()
    df_contents = data_frame.values.tolist()
    df_all = [df_columns, *df_contents]
    with safe_write_open(file_path, *args, **kwargs) as file_stream:
        writer = csv.writer(file_stream, quoting=csv.QUOTE_ALL)
        writer.writerows(df_all)
