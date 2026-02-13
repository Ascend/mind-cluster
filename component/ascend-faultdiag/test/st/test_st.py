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
import json
import os
import shutil
import unittest
import logging
import logging.handlers
from unittest.mock import patch

from ascend_fd.pkg.diag.knowledge_graph.kg_diag_job import _kg_diag_job

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.model.package_data import PackageData
from ascend_fd.pkg.diag.root_cluster.utils import ErrorChecker
from ascend_fd.utils.tool import CONF_PATH
from ascend_fd.pkg.parse.parser_saver import ProcessLogSaver, EnvInfoSaver, TrainLogSaver, HostLogSaver, BMCLogSaver, \
    LCNELogSaver, DevLogSaver, DlLogSaver, AMCTLogSaver, MindieLogSaver, ParsedDataSaver
from ascend_fd.model.cfg import DiagCFG, ParseCFG
from ascend_fd.configuration.config import HOME_PATH
from ascend_fd.controller import router
from ascend_fd.cli import run
from ascend_fd.controller.controller import ParseController, DiagController
from ascend_fd.controller.job_worker import start_rc_parse_job, start_kg_parse_job, start_kg_diag_job
from ascend_fd.pkg.parse.node_anomaly import start_node_parse_job
from ascend_fd.pkg.parse.network_congestion import start_net_parse_job
from ascend_fd.pkg.diag.node_anomaly import start_node_diag_job
from ascend_fd.pkg.diag.root_cluster import start_rc_diag_job
from ascend_fd.pkg.diag.network_congestion import start_net_diag_job
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.kg_engine_main import kg_engine_analyze
from ascend_fd.pkg.diag.node_anomaly.npu_anomaly import npu_anomaly_job
from ascend_fd.pkg.parse.node_anomaly.node_parse_job import NpuAnomalyParser, ParseHostMetrics
from ascend_fd.pkg.diag.node_anomaly.resource_preemption import resource_preemption_job
from ascend_fd.pkg.diag.fault_entity import SINGLE_PROCESS_PREEMPT_FAULT_ENTITY, NPU_STATUS_ABNORMAL_ENTITY
from ascend_fd.pkg.diag.root_cluster.fault_description import ALL_SOCKET_ERROR_NOT_TIMEOUT, \
    AI_CPU_NOTIFY_TIMEOUT, TRANSPORT_INIT_ERROR, TRANSPORT_INIT_ERROR_NO_DEVICE_ID

TEST_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
TESTCASE_INPUT = os.path.join(TEST_DIR, "st_testcase")
TESTCASE_CQE_INPUT = os.path.join(TEST_DIR, "st_cqe_testcase")
TESTCASE_HOST_OS_INPUT = os.path.join(TEST_DIR, "st_host_os_testcase")
TESTCASE_TIMEOUT_EVENT_INPUT = os.path.join(TEST_DIR, "st_event_timeout_testcase")
TESTCASE_TRANSPORT_TIMEOUT_INPUT = os.path.join(TEST_DIR, "st_transport_timeout_testcase")
TESTCASE_AICPU_NOTIFY_INPUT = os.path.join(TEST_DIR, "st_aicpu_notify_testcase")
TESTCASE_TRANSPORT_INIT_INPUT = os.path.join(TEST_DIR, "st_transport_init_testcase")


def set_logger(name_list):
    """
    reset the logger handler to StreamHandler.
    :param name_list: logger name list
    """
    stream_handler = logging.StreamHandler()
    for name in name_list:
        logger = logging.getLogger(name)
        logger.handlers = [stream_handler]


class ParseSTArgs:
    cmd = "parse"

    def __init__(self, input_dir, output_dir, host_log=None, train_log=None, env_check=None,
                 device_log=None, process_log=None, dl_log=None, amct_log=None, mindie_log=None, bmc_log=None,
                 lcne_log=None, bus_log=None, custom_log=None, scene="host", task_id="test_uuid"):
        self.input_path = input_dir
        self.output_path = output_dir
        self.host_log = host_log
        self.train_log = train_log
        self.env_check = env_check
        self.device_log = device_log
        self.process_log = process_log
        self.dl_log = dl_log
        self.amct_log = amct_log
        self.mindie_log = mindie_log
        self.bmc_log = bmc_log
        self.lcne_log = lcne_log
        self.bus_log = bus_log
        self.scene = scene
        self.task_id = task_id
        self.performance = False
        self.custom_log = custom_log


class DiagSTArgs:
    cmd = "diag"

    def __init__(self, input_dir, output_dir, mode=0, task_id="test_uuid", scene="host"):
        self.input_path = input_dir
        self.output_path = output_dir
        self.mode = mode
        self.task_id = task_id
        self.performance = False
        self.scene = scene


class STTestController(unittest.TestCase):
    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TEST_DIR, "st_output")
        self.parse_output = os.path.join(self.output, "fault_diag_data", "worker-0")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.output):
            os.makedirs(self.output)
        if not os.path.exists(self.parse_output):
            os.makedirs(self.parse_output)
        self.parse_args = ParseSTArgs(TESTCASE_INPUT, self.parse_output)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)

    def test_parse_and_diag(self):
        self.execute_parse_and_diag(self.parse_args, self.diag_args)

    def test_parse_and_diag_p(self):
        self.parse_args.performance = True
        self.diag_args.performance = True
        self.execute_parse_and_diag(self.parse_args, self.diag_args)

    def execute_parse_and_diag(self, parse_args, diag_args):
        self.run_parse_or_diag(args=parse_args)
        self.run_parse_or_diag(args=diag_args)
        fault_diag_result_dir = os.path.join(self.output, "fault_diag_result")
        diag_report = os.path.join(fault_diag_result_dir, "diag_report.json")
        with open(diag_report, 'r') as file_stream:
            report = json.loads(file_stream.read())
            self.assertTrue(report["Knowledge_Graph"]["analyze_success"])
            self.assertTrue(report["Root_Cluster"]["analyze_success"])

    @patch("ascend_fd.cli.command_line")
    def run_parse_or_diag(self, command_line, args):
        command_line.return_value = args
        run(args.cmd)

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestSingleController(unittest.TestCase):
    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TEST_DIR, "st_single_output")
        self.single_diag_args = ParseSTArgs(TESTCASE_INPUT, self.output)
        self.single_diag_args.cmd = "single-diag"

    def test_parse_and_diag(self):
        router(self.single_diag_args)
        single_diag_result_dir = os.path.join(self.output, "fault_diag_result")
        single_diag_report = os.path.join(single_diag_result_dir, "diag_report.json")
        with open(single_diag_report, 'r') as file_stream:
            report = json.loads(file_stream.read())
            self.assertTrue(report["Knowledge_Graph"]["analyze_success"])

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestRCAndKg(unittest.TestCase):

    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TEST_DIR, "st_output_1")
        self.parse_output = os.path.join(self.output, "fault_diag_data", "worker-0")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.output):
            os.makedirs(self.output)
        if not os.path.exists(self.parse_output):
            os.makedirs(self.parse_output)
        self.parse_args = ParseSTArgs(TESTCASE_INPUT, self.parse_output)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["ROOT_CLUSTER", "KNOWLEDGE_GRAPH", "KG_ENGINE"])

    def test_parse_and_diag(self):
        controller = ParseController(self.parse_args)
        cfg = controller.cfg
        start_rc_parse_job(cfg)
        start_kg_parse_job(cfg)
        # parse test
        parsed_file_list = os.listdir(self.parse_output)
        self.assertIn("ascend-kg-parser.json", parsed_file_list)
        self.assertIn("ascend-kg-analyzer.json", parsed_file_list)
        self.assertIn("plog-parser-10972-1.log", parsed_file_list)
        self.assertIn("plog-parser-11044-1.log", parsed_file_list)
        self.assertIn("plog-parser-11116-1.log", parsed_file_list)
        self.assertIn("plog-parser-11190-1.log", parsed_file_list)
        self.assertIn("plog-parser-11262-1.log", parsed_file_list)
        self.assertIn("plog-parser-11335-1.log", parsed_file_list)
        self.assertIn("plog-parser-11406-1.log", parsed_file_list)
        self.assertIn("plog-parser-11481-1.log", parsed_file_list)
        ascend_kg_parser_json = os.path.join(self.parse_output, "ascend-kg-parser.json")
        with open(ascend_kg_parser_json, 'r') as file_stream:
            kg_json = json.loads(file_stream.read())
            self.assertIn("AISW_MindSpore_Compiler_01", kg_json)
            self.assertIn("AISW_PyTorch_Env_02", kg_json)
            self.assertIn("Comp_Network_Custom_01", kg_json)
            self.assertIn("AISW_TRACEBACK_NameError", kg_json)
            self.assertIn("AISW_TRACEBACK_Exception", kg_json)
        # diag test
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_rc_diag_job(cfg)
        self.assertTrue(result.analyze_success)
        self.assertIn("worker-0", result.detect_workers_devices)
        cfg.root_worker_devices = result.detect_workers_devices
        kg_result = start_kg_diag_job(cfg)
        code_list = []
        for fault in kg_result.get("fault"):
            code_list.append(fault.get("code"))
        self.assertIn("AISW_MindSpore_Compiler_01", code_list)
        self.assertTrue(kg_result.get("analyze_success"))

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestCQE(unittest.TestCase):
    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TEST_DIR, "st_output_cqe")
        self.parse_output = os.path.join(self.output, "fault_diag_data", "worker-9")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.output):
            os.makedirs(self.output)
        if not os.path.exists(self.parse_output):
            os.makedirs(self.parse_output)
        self.parse_args = ParseSTArgs(TESTCASE_CQE_INPUT, self.parse_output)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["ROOT_CLUSTER"])

    def test_parse_and_diag(self):
        controller = ParseController(self.parse_args)
        cfg = controller.cfg
        start_rc_parse_job(cfg)
        # parse test
        parsed_file_list = os.listdir(self.parse_output)
        self.assertIn("device_ip_info.json", parsed_file_list)
        self.assertIn("plog-parser-12339-1.log", parsed_file_list)
        self.assertIn("plog-parser-12337-1.log", parsed_file_list)
        # diag test
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_rc_diag_job(cfg)
        self.assertTrue(result.analyze_success)
        self.assertIn("worker-9", result.detect_workers_devices)

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestAllSocketTimeout(unittest.TestCase):
    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        ErrorChecker.timeout_error_info_map = {}
        self.output = os.path.join(TEST_DIR, "st_output_socket")
        self.parse_output_0 = os.path.join(self.output, "fault_diag_data", "worker-0")
        self.parse_output_1 = os.path.join(self.output, "fault_diag_data", "worker-1")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.parse_output_0):
            os.makedirs(self.parse_output_0)
        if not os.path.exists(self.parse_output_1):
            os.makedirs(self.parse_output_1)
        self.parse_args_0 = ParseSTArgs(os.path.join(TESTCASE_TIMEOUT_EVENT_INPUT, "worker-0"), self.parse_output_0)
        self.parse_args_1 = ParseSTArgs(os.path.join(TESTCASE_TIMEOUT_EVENT_INPUT, "worker-1"), self.parse_output_1)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["ROOT_CLUSTER"])

    def test_cycle_device(self):
        start_rc_parse_job(ParseController(self.parse_args_0).cfg)
        start_rc_parse_job(ParseController(self.parse_args_1).cfg)
        # parse test
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_0))
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_1))
        # diag test
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_rc_diag_job(cfg)
        self.assertTrue(result.analyze_success)
        self.assertEqual(result.root_cause_device, ['worker-1 device-0', 'worker-0 device-0'])
        self.assertEqual(result.remote_link, "worker-1 device-0 -> worker-0 device-0 -> worker-1 device-0")
        self.assertEqual(result.fault_description.code, ALL_SOCKET_ERROR_NOT_TIMEOUT.code)

    def test_unknown_devices(self):
        start_rc_parse_job(ParseController(self.parse_args_1).cfg)
        # parse test
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_1))
        # diag test
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_rc_diag_job(cfg)
        self.assertTrue(result.analyze_success)
        self.assertEqual(result.root_cause_device, ['worker-1 device-0'])
        self.assertEqual(result.remote_link, "worker-1 device-0 -> 124.0.0.21")
        self.assertEqual(result.fault_description.code, ALL_SOCKET_ERROR_NOT_TIMEOUT.code)

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestTransportSocketTimeout(unittest.TestCase):
    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        ErrorChecker.timeout_error_info_map = {}
        self.output = os.path.join(TEST_DIR, "st_output_transport_socket")
        self.parse_output_0 = os.path.join(self.output, "fault_diag_data", "worker-0")
        self.parse_output_1 = os.path.join(self.output, "fault_diag_data", "worker-1")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.parse_output_0):
            os.makedirs(self.parse_output_0)
        if not os.path.exists(self.parse_output_1):
            os.makedirs(self.parse_output_1)
        self.parse_args_0 = ParseSTArgs(os.path.join(TESTCASE_TRANSPORT_TIMEOUT_INPUT, "worker-0"), self.parse_output_0)
        self.parse_args_1 = ParseSTArgs(os.path.join(TESTCASE_TRANSPORT_TIMEOUT_INPUT, "worker-1"), self.parse_output_1)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["ROOT_CLUSTER"])

    def test_transport_Socket_diag(self):
        start_rc_parse_job(ParseController(self.parse_args_0).cfg)
        start_rc_parse_job(ParseController(self.parse_args_1).cfg)
        # parse test
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_0))
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_1))
        # diag test
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_rc_diag_job(cfg)
        self.assertTrue(result.analyze_success)
        self.assertEqual(['worker-1 device-1'], result.root_cause_device)
        self.assertEqual("worker-0 device-2 -> worker-1 device-2 -> worker-1 device-0 -> worker-1 device-1",
                         result.remote_link)
        self.assertEqual(result.fault_description.code, ALL_SOCKET_ERROR_NOT_TIMEOUT.code)

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestHostOs(unittest.TestCase):
    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TESTCASE_HOST_OS_INPUT, "st_output_host_os")
        self.parse_output = os.path.join(self.output, "fault_diag_data", "worker-0")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.output):
            os.makedirs(self.output)
        if not os.path.exists(self.parse_output):
            os.makedirs(self.parse_output)
        self.parse_args = ParseSTArgs(TESTCASE_HOST_OS_INPUT, self.parse_output)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["KNOWLEDGE_GRAPH"])

    def test_parse_and_diag(self):
        controller = ParseController(self.parse_args)
        cfg = controller.cfg
        start_kg_parse_job(cfg)
        # parse test
        ascend_kg_parser_json = os.path.join(self.parse_output, "ascend-kg-parser.json")
        with open(ascend_kg_parser_json, 'r') as file_stream:
            kg_json = json.loads(file_stream.read())
            self.assertIn("Comp_OS_Kernel_FS_02", kg_json)
            self.assertIn("Comp_OS_Kernel_Mem_06", kg_json)
            self.assertIn("Comp_OS_Custom_03", kg_json)
            self.assertIn("Comp_OS_Custom_04", kg_json)
            self.assertIn("Comp_OS_Service_State_05", kg_json)
        # diag test
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_rc_diag_job(cfg)
        self.assertTrue(result.analyze_success)
        self.assertIn("worker-0", result.detect_workers_devices)
        cfg.root_worker_devices = result.detect_workers_devices
        kg_result = start_kg_diag_job(cfg)
        code_list = []
        for fault in kg_result.get("fault"):
            code_list.append(fault.get("code"))
        self.assertIn("Comp_OS_Kernel_FS_02", code_list)
        self.assertIn("Comp_OS_Kernel_Mem_06", code_list)
        self.assertIn("Comp_OS_Custom_03", code_list)
        self.assertIn("Comp_OS_Custom_04", code_list)
        self.assertIn("Comp_OS_Service_State_05", code_list)
        self.assertTrue(kg_result.get("analyze_success"))

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestNode(unittest.TestCase):
    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TEST_DIR, "st_output_2")
        self.parse_output = os.path.join(self.output, "fault_diag_data", "worker-0")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.output):
            os.makedirs(self.output)
        if not os.path.exists(self.parse_output):
            os.makedirs(self.parse_output)
        self.parse_args = ParseSTArgs(TESTCASE_INPUT, self.parse_output)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["NODE_ANOMALY"])

    def test_parse_and_diag(self):
        controller = ParseController(self.parse_args)
        cfg = controller.cfg
        start_node_parse_job(cfg)
        parsed_file_list = os.listdir(self.parse_output)
        self.assertIn("nad_clean.csv", parsed_file_list)
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_node_diag_job(cfg)
        self.assertTrue(result.get("analyze_success"))

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestNetDiag(unittest.TestCase):
    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TEST_DIR, "st_output_net")
        self.parse_input = os.path.join(TESTCASE_INPUT)
        self.parse_output = os.path.join(self.output, "fault_diag_data", "worker-0")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.output):
            os.makedirs(self.output)
        if not os.path.exists(self.parse_output):
            os.makedirs(self.parse_output)
        self.parse_args = ParseSTArgs(self.parse_input, self.parse_output)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["NET_CONGESTION"])


    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestSingleModule(unittest.TestCase):

    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TEST_DIR, "st_output_module")
        if not os.path.exists(self.output):
            os.makedirs(self.output)
        self.kg_repo = os.path.join(CONF_PATH, "kg-config.json")
        set_logger(["KG_ENGINE", "NODE_ANOMALY", "ROOT_CLUSTER"])

    def test_single_na_and_rp_parse(self):
        env_info_path = os.path.join(TEST_DIR, "st_testcase", "modelarts-job-testdt",
                                     "ascend", "environment_check")
        env_info = EnvInfoSaver()
        env_info.filter_log(env_info_path)
        cfg = ParseCFG(
            task_id="test_uuid",
            input_path="",
            output_path=self.output,
            lcne_log="",
            bmc_log="",
            log_saver=ProcessLogSaver(),
            env_info_saver=env_info,
            train_log_saver=TrainLogSaver(),
            host_log_saver=HostLogSaver(),
            dev_log_saver=DevLogSaver(),
            dl_log_saver=DlLogSaver(),
            amct_log_saver=AMCTLogSaver(),
            mindie_log_saver=MindieLogSaver(),
            bmc_log_saver=BMCLogSaver(),
            lcne_log_saver=LCNELogSaver(),
        )
        NpuAnomalyParser(cfg).parse()
        ParseHostMetrics(cfg).parse()
        result_files = os.listdir(self.output)
        self.assertIn("nad_clean.csv", result_files)
        self.assertIn("process_64.csv", result_files)

    def test_single_kg_diag(self):
        input_file = os.path.join(TEST_DIR, "st_module_testcase", "kg_diag",
                                  "fault_diag_data", "worker-0", "ascend-kg-parser.json")
        package_data = PackageData([], input_file)
        resp = kg_engine_analyze([self.kg_repo], package_data)

        self.assertTrue(resp.analyze_success)
        self.assertIn("AISW_CANN_Runtime_014", resp.root_causes)
        self.assertIn("AISW_CANN_Runtime_033", resp.root_causes)
        self.assertIn("AISW_PyTorch_ERRCODE_Common_ERR01002", resp.root_causes)

    def test_only_traceback_kg_diag(self):
        # traceback faults are displayed only when there is no valid fault event
        input_file = os.path.join(TEST_DIR, "st_module_testcase", "kg_diag",
                                  "fault_diag_data", "worker-1", "ascend-kg-parser.json")
        package_data = PackageData([], input_file)
        resp = kg_engine_analyze([self.kg_repo], package_data)
        self.assertTrue(resp.analyze_success)
        self.assertIn("AISW_TRACEBACK_FileNotFoundError", resp.root_causes)

    def test_old_kg_diag_job(self):
        input_dir = os.path.join(TEST_DIR, "st_module_testcase", "kg_diag", "fault_diag_data")
        parsed_saver = ParsedDataSaver(input_dir, ParseSTArgs(input_dir, ""))
        worker_name = "worker-0"
        job_name = f"KNOWLEDGE_GRAPH_WORKER_{worker_name}"
        root_causes = _kg_diag_job(worker_name, ["0"], parsed_saver, job_name).get("root_causes")
        self.assertIn("AISW_CANN_Runtime_014", root_causes)
        self.assertIn("AISW_CANN_Runtime_033", root_causes)  # no 'source_device'
        self.assertIn("AISW_PyTorch_ERRCODE_Common_ERR01002", root_causes)  # has 'source_device'

    def test_single_na_and_rp_diag(self):
        input_dir = os.path.join(TEST_DIR, "st_module_testcase", "node_diag",
                                 "fault_diag_data", "worker-0")
        result_dict = npu_anomaly_job(input_dir, "0")
        self.assertIn(NPU_STATUS_ABNORMAL_ENTITY, result_dict)
        result_dict = resource_preemption_job(input_dir, "0")
        self.assertIn(SINGLE_PROCESS_PREEMPT_FAULT_ENTITY, result_dict)

    def test_single_net_diag(self):
        expect_fault_details = [
            {
                "worker": "worker-0",
                "device_list": ["device-0", "device-1", "device-2", "device-3", "device-4", "device-5", "device-6",
                                "device-7"]
            },
            {
                "worker": "worker-1",
                "device_list": ["device-0", "device-1", "device-2", "device-3", "device-4", "device-5", "device-6",
                                "device-7"]
            }
        ]
        input_dir = os.path.join(TEST_DIR, "st_module_testcase", "net_diag",
                                 "fault_diag_data")
        controller = DiagCFG("test_uuid", input_dir, self.output, ParsedDataSaver(input_dir, DiagSTArgs(input_dir, "")))
        result = start_net_diag_job(controller)
        self.assertTrue(result['analyze_success'])
        self.assertEqual(result['fault'][0]['code'], "NET_CONGESTION_ABNORMAL_01")
        self.assertEqual(result['fault'][0]['fault_details'], expect_fault_details)

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestAiCpuNotify(unittest.TestCase):
    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        ErrorChecker.timeout_error_info_map = {}
        self.output = os.path.join(TEST_DIR, "st_output_aicpu_notify")
        self.parse_output_0 = os.path.join(self.output, "fault_diag_data", "worker-0")
        self.parse_output_1 = os.path.join(self.output, "fault_diag_data", "worker-1")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.parse_output_0):
            os.makedirs(self.parse_output_0)
        if not os.path.exists(self.parse_output_1):
            os.makedirs(self.parse_output_1)
        self.parse_args_0 = ParseSTArgs(os.path.join(TESTCASE_AICPU_NOTIFY_INPUT, "worker-0"), self.parse_output_0)
        self.parse_args_1 = ParseSTArgs(os.path.join(TESTCASE_AICPU_NOTIFY_INPUT, "worker-1"), self.parse_output_1)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["ROOT_CLUSTER"])

    def test_transport_Socket_diag(self):
        start_rc_parse_job(ParseController(self.parse_args_0).cfg)
        start_rc_parse_job(ParseController(self.parse_args_1).cfg)
        # parse test
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_0))
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_1))
        # diag test
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_rc_diag_job(cfg)
        self.assertTrue(result.analyze_success)
        self.assertEqual(result.root_cause_device, ['worker-0 device-11'])
        self.assertEqual(result.remote_link, "worker-1 device-3 -> worker-0 device-11")
        self.assertEqual(result.fault_description.code, AI_CPU_NOTIFY_TIMEOUT.code)

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestTransportInitError(unittest.TestCase):

    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TEST_DIR, "st_output_transport_init")
        self.parse_output_0 = os.path.join(self.output, "fault_diag_data", "worker-0")
        self.parse_output_1 = os.path.join(self.output, "fault_diag_data", "worker-1")
        self.parse_output_2 = os.path.join(self.output, "fault_diag_data", "worker-2")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.parse_output_0):
            os.makedirs(self.parse_output_0)
        if not os.path.exists(self.parse_output_1):
            os.makedirs(self.parse_output_1)
        if not os.path.exists(self.parse_output_2):
            os.makedirs(self.parse_output_2)
        self.parse_args_0 = ParseSTArgs(os.path.join(TESTCASE_TRANSPORT_INIT_INPUT, "worker-0"), self.parse_output_0)
        self.parse_args_1 = ParseSTArgs(os.path.join(TESTCASE_TRANSPORT_INIT_INPUT, "worker-1"), self.parse_output_1)
        self.parse_args_2 = ParseSTArgs(os.path.join(TESTCASE_TRANSPORT_INIT_INPUT, "worker-2"), self.parse_output_2)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["ROOT_CLUSTER"])

    def test_transport_Socket_diag(self):
        start_rc_parse_job(ParseController(self.parse_args_0).cfg)
        start_rc_parse_job(ParseController(self.parse_args_1).cfg)
        start_rc_parse_job(ParseController(self.parse_args_2).cfg)
        # parse test
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_0))
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_1))
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_2))
        # diag test
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_rc_diag_job(cfg)
        self.assertTrue(result.analyze_success)
        self.assertEqual(result.root_cause_device, ['worker-0 device-0'])
        self.assertEqual(result.remote_link, "worker-2 device-2 -> worker-1 device-1 -> worker-0 device-0")
        self.assertEqual(result.fault_description.code, TRANSPORT_INIT_ERROR.code)

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)


class STTestTransportInitErrorNoDeviceId(unittest.TestCase):

    def setUp(self) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        self.output = os.path.join(TEST_DIR, "st_output_transport_init_no_device_id")
        self.parse_output_0 = os.path.join(self.output, "fault_diag_data", "worker-3")
        self.diag_input = os.path.join(self.output, "fault_diag_data")
        if not os.path.exists(self.parse_output_0):
            os.makedirs(self.parse_output_0)
        self.parse_args_0 = ParseSTArgs(os.path.join(TESTCASE_TRANSPORT_INIT_INPUT, "worker-3"), self.parse_output_0)
        self.diag_args = DiagSTArgs(self.diag_input, self.output)
        set_logger(["ROOT_CLUSTER"])

    def test_transport_Socket_diag(self):
        start_rc_parse_job(ParseController(self.parse_args_0).cfg)
        # parse test
        self.assertIn("ascend-rc-parser.json", os.listdir(self.parse_output_0))
        # diag test
        controller = DiagController(self.diag_args)
        cfg = controller.cfg
        result = start_rc_diag_job(cfg)
        self.assertTrue(result.analyze_success)
        self.assertEqual(['Unknown Device'], result.root_cause_device)
        self.assertEqual(TRANSPORT_INIT_ERROR_NO_DEVICE_ID.code, result.fault_description.code)

    def tearDown(self) -> None:
        if os.path.exists(self.output):
            shutil.rmtree(self.output)
