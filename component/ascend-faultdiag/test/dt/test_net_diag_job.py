#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
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
import unittest
import tempfile
import shutil
from unittest.mock import patch

import pandas as pd

from ascend_fd.pkg.diag.network_congestion.net_diag_job import NetCongestionDetector
from ascend_fd.pkg.parse.network_congestion.net_parse_job import safe_save_csv, safe_read_csv
from ascend_fd.utils.regular_table import NIC_OUT_FILENAME
from ascend_fd.utils.status import FileNotExistError

INPUT_COL_NAMES = NetCongestionDetector.INPUT_COL_NAMES


class NetDiagJobTestCase(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp_dir = tempfile.mkdtemp(prefix="net_diag_test_")

    def tearDown(self) -> None:
        if os.path.exists(self.tmp_dir):
            shutil.rmtree(self.tmp_dir)

    def _build_nic_df(self, rows=None):
        if rows is None:
            rows = [
                {"device_id": "0", **{col: 1.0 for col in INPUT_COL_NAMES}},
                {"device_id": "1", **{col: 2.0 for col in INPUT_COL_NAMES}},
            ]
        return pd.DataFrame(rows)

    def _write_nic_csv(self, worker_dir, df=None):
        os.makedirs(worker_dir, exist_ok=True)
        df = df if df is not None else self._build_nic_df()
        safe_save_csv(df, os.path.join(worker_dir, NIC_OUT_FILENAME), mode='w+', newline='')
        return os.path.join(worker_dir, NIC_OUT_FILENAME)

    def test_safe_read_csv_utf8_without_encoding_arg(self):
        nic_file = self._write_nic_csv(self.tmp_dir)
        dataframe = safe_read_csv(nic_file, dtype={'device_id': str}, header=0)
        self.assertFalse(dataframe.empty)
        self.assertIn('device_id', dataframe.columns)
        self.assertEqual(len(dataframe), 2)
        self.assertEqual(str(dataframe.loc[0, 'device_id']), '0')

    def _build_detector_without_model(self):
        with patch.object(NetCongestionDetector, '_model_load', return_value=None):
            detector = NetCongestionDetector.__new__(NetCongestionDetector)
            detector.worker_num = -1
            detector.model = None
        return detector

    def test_get_nic_data_reads_utf8_csv_without_encoding_error(self):
        worker_dir = os.path.join(self.tmp_dir, "worker-0")
        self._write_nic_csv(worker_dir)
        worker_path_dict = {"0": worker_dir}

        detector = self._build_detector_without_model()
        with patch.object(NetCongestionDetector, '_model_load', return_value=None):
            nic_df = detector.get_nic_data(worker_path_dict)
        self.assertFalse(nic_df.empty)
        self.assertIn('worker_name', nic_df.columns)
        self.assertTrue((nic_df['worker_name'] == '0').all())

    def test_get_nic_data_multi_worker(self):
        worker_path_dict = {}
        for worker_id in ("0", "1"):
            worker_dir = os.path.join(self.tmp_dir, f"worker-{worker_id}")
            self._write_nic_csv(worker_dir)
            worker_path_dict[worker_id] = worker_dir

        detector = self._build_detector_without_model()
        with patch.object(NetCongestionDetector, '_model_load', return_value=None):
            nic_df = detector.get_nic_data(worker_path_dict)
        self.assertEqual(detector.worker_num, 2)
        self.assertEqual(set(nic_df['worker_name']), {"0", "1"})

    def test_get_nic_data_missing_csv_raises(self):
        worker_path_dict = {"0": self.tmp_dir}
        detector = self._build_detector_without_model()
        with patch.object(NetCongestionDetector, '_model_load', return_value=None):
            with self.assertRaises(FileNotExistError):
                detector.get_nic_data(worker_path_dict)

    def test_get_nic_data_skips_invalid_df(self):
        worker_dir = os.path.join(self.tmp_dir, "worker-0")
        bad_df = pd.DataFrame([{"device_id": "0"}])
        self._write_nic_csv(worker_dir, bad_df)
        worker_path_dict = {"0": worker_dir}

        detector = self._build_detector_without_model()
        with patch.object(NetCongestionDetector, '_model_load', return_value=None):
            with self.assertRaises(FileNotExistError):
                detector.get_nic_data(worker_path_dict)


if __name__ == '__main__':
    unittest.main()
