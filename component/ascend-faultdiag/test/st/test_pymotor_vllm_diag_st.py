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
# pylint: disable=duplicate-code
import json
import os
import shutil
import tempfile
import unittest

from ascend_fd.configuration.config import KG_ANALYZER_DUMP_NAME, KG_PASER_DUMP_NAME
from ascend_fd.controller.controller import ParseController
from ascend_fd.controller.job_worker import start_kg_parse_job

MOONCAKE_EVENT_CODE = "AISW_VLLM_ASCEND_KV_POOL_MOONCAKE_002"

# The plog time window is 10:40:00 - 10:50:00, the fault lines below are at 10:46:50.
PLOG_CONTENT = (
    "[INFO] RUNTIME(123,python):2026-06-24-10:40:00.000.000 [task_info.cc:5261]123 plog start\n"
    "[INFO] RUNTIME(123,python):2026-06-24-10:50:00.000.000 [task_info.cc:5261]123 plog end\n"
)
PYMOTOR_CONTENT = (
    "\x1b[93m2026-06-24 10:46:50.237365 WARNING  \x1b[0m\x1b[0K[3591] "
    "[client_pool.hpp:445] send request to 80.48.37.140:50060 failed. connection refused. \n"
    "E20260624 10:46:50.237413  3591 master_client.cpp:234] Client not available \n"
    "(Worker_TP1 pid=3193) ERROR 06-24 10:46:50 [mooncake_backend.py:146] "
    "[vllm-ascend] [distributed] - Initialize mooncake failed. ret=-600, "
    "metadata_server=P2PHANDSHAKE. Check mooncake config and network. \n"
)


class ParseArgs:
    """Minimal ascend-fd parse command args"""

    cmd = "parse"
    task_id = "st_pymotor_vllm_diag"

    def __init__(self, input_path, output_path):
        self.input_path = input_path
        self.output_path = output_path
        self.performance = False
        self.host_log = ""
        self.train_log = ""
        self.env_check = ""
        self.device_log = ""
        self.process_log = ""
        self.dl_log = ""
        self.amct_log = ""
        self.mindie_log = ""
        self.bmc_log = ""
        self.lcne_log = ""
        self.bus_log = ""
        self.custom_log = ""
        self.pymotor_vllm_log = ""


class STPyMotorVLLMDiag(unittest.TestCase):
    """End-to-end parse and diag for PyMotor/vLLM logs with a plog time window"""

    def setUp(self):
        self.base = tempfile.mkdtemp()
        self.input_dir = os.path.join(self.base, "input")
        self.output_dir = os.path.join(self.base, "output")
        process_log_dir = os.path.join(self.input_dir, "process_log")
        pymotor_log_dir = os.path.join(self.input_dir, "pymotor_vllm_log")
        os.makedirs(process_log_dir)
        os.makedirs(pymotor_log_dir)
        os.makedirs(self.output_dir)
        with open(os.path.join(process_log_dir, "plog-123_20260624104000.log"), "w", encoding="utf-8") as file_stream:
            file_stream.write(PLOG_CONTENT)
        with open(os.path.join(pymotor_log_dir, "vllm-mooncake.log"), "w", encoding="utf-8") as file_stream:
            file_stream.write(PYMOTOR_CONTENT)

    def tearDown(self):
        shutil.rmtree(self.base, ignore_errors=True)

    def test_parse_and_diag_cross_line_fault_with_plog_window(self):
        """The cross-line opt/all fault must survive the plog window and appear in the diag result"""
        controller = ParseController(ParseArgs(self.input_dir, self.output_dir))
        start_kg_parse_job(controller.cfg)

        with open(os.path.join(self.output_dir, KG_PASER_DUMP_NAME), encoding="utf-8") as file_stream:
            parse_result = json.load(file_stream)
        self.assertIn(MOONCAKE_EVENT_CODE, json.dumps(parse_result))

        with open(os.path.join(self.output_dir, KG_ANALYZER_DUMP_NAME), encoding="utf-8") as file_stream:
            diag_result = json.load(file_stream)
        self.assertTrue(diag_result.get("response"))
        self.assertIn(MOONCAKE_EVENT_CODE, json.dumps(diag_result))


if __name__ == "__main__":
    unittest.main()
