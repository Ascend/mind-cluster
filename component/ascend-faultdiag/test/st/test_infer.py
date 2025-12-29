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
import unittest

from ascend_fd.pkg.diag.root_cluster.mindie_diag_job import diag_link_error
from ascend_fd.pkg.parse.parser_saver import ParsedDataSaver

TEST_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))


class InferTestController(unittest.TestCase):

    def setUp(self) -> None:
        """
        Single-diag Controller
        """
        self.mindie_link_error_info = {
            '1': ['2'],
            '2': ['3', '4'],
            '4': ['2']
        }

    def test_start_job_with_link_errors(self):
        result = diag_link_error(self.mindie_link_error_info, {}, {})
        self.assertIn("3", result)

    def test_parsed_data(self):
        parsed_data_path = os.path.join(TEST_DIR, "st_module_testcase", "rc_diag", "rc_infer")
        parsed_saver = ParsedDataSaver(parsed_data_path, args=DiagSTArgs("", ""))
        self.assertEqual(len(parsed_saver.pid_device_dict), 16)


class DiagSTArgs:
    cmd = "diag"

    def __init__(self, input_dir, output_dir, mode=0, task_id="test_uuid", scene="host"):
        self.input_path = input_dir
        self.output_path = output_dir
        self.mode = mode
        self.task_id = task_id
        self.performance = False
        self.scene = scene
