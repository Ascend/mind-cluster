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

from ascend_fd.controller.controller import DiagController
from ascend_fd.model.diag_info import RCDiagResult
from ascend_fd.pkg.diag.root_cluster.rc_diag_job import RCDiagWorker
from ascend_fd.pkg.diag.root_cluster.utils import DeviceTable, ErrorChecker, Device, Identifier, RankDevice, \
    NoPlogChecker, NoValidPlogInfoErrorChecker, ResumingTrainingInvalidBaseInfoChecker
from ascend_fd.pkg.parse.parser_saver import ParsedDataSaver
from ascend_fd.model.cfg import DiagCFG
from ascend_fd.configuration.config import HOME_PATH
from ascend_fd.pkg.diag.root_cluster import fault_description

TEST_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))


class CommonTestCase(unittest.TestCase):
    testcase_name = ""
    code = ""
    device = ""

    @classmethod
    def setUpClass(cls) -> None:
        os.makedirs(HOME_PATH, 0o700, exist_ok=True)
        ErrorChecker.timeout_error_info_map = {}
        args = DiagSTArgs(os.path.join(TEST_DIR, "st_module_testcase", "rc_diag", cls.testcase_name), "")
        cls.diag = DiagController(args)
        cls.diag._exec_root_cluster_job()

    @classmethod
    def tearDownClass(cls):
        super().tearDownClass()

    def execute_assert(self):
        result = RCDiagResult.from_dict(self.diag.origin_results.get("Rc"))
        self.assertTrue(result.analyze_success)
        self.assertEqual(result.fault_description.code, self.code)
        self.assertEqual(result.root_cause_device, self.device)


class STTestRC101(CommonTestCase):
    testcase_name = "rc_101"
    code = fault_description.PART_ERROR_WITH_NO_TIMEOUT.code
    device = ["ALL Device"]

    def test_start(self):
        self.execute_assert()


class STTestRC107(CommonTestCase):
    testcase_name = "rc_107"
    code = fault_description.ALL_SOCKET_ERROR_WITH_TIMEOUT.code
    device = ["worker-0 device-6"]

    def test_start(self):
        self.execute_assert()


class STTestRC117(CommonTestCase):
    testcase_name = "rc_117"
    code = fault_description.ALL_INIT_FAILED_WITH_TIMEOUT.code
    device = ['worker-1 device-4']

    def test_start(self):
        self.execute_assert()


class STTestRC118(CommonTestCase):
    testcase_name = "rc_118"
    code = fault_description.ALL_INIT_FAILED_NOT_TIMEOUT.code
    device = ["Unknown Device"]

    def test_start(self):
        self.execute_assert()


class STTestRC119(CommonTestCase):
    testcase_name = "rc_119"
    code = fault_description.PART_INIT_FAILED.code
    device = ['worker-0 device-7']

    def test_start(self):
        self.execute_assert()


class STTestRC120(CommonTestCase):
    testcase_name = "rc_120"
    code = fault_description.TLS_SWITCH_DIFFERENT.code
    device = ['worker-0 device-5']

    def test_start(self):
        self.execute_assert()


class STTestRC121(CommonTestCase):
    testcase_name = "rc_121"
    code = fault_description.INIT_FAILED_WITH_NO_CONN_NO_LOG.code
    device = ["51.38.67.141%enp189s0f0_60004_4_1711521369616458 rank-1"]

    def test_start(self):
        self.execute_assert()


class STTestRC122(CommonTestCase):
    testcase_name = "rc_122"
    code = fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT_INDEX_ERR.code
    device = ["w_1 device-1"]

    def test_start(self):
        self.execute_assert()


class STTestRC123(CommonTestCase):
    testcase_name = "rc_123"
    code = fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT_TAG_ERR.code
    device = ["w_1 device-1"]

    def test_start(self):
        self.execute_assert()


class STTestRC124(CommonTestCase):
    testcase_name = "rc_124"
    code = fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT_REMOTE_CYCLE.code
    device = ['Unknown Device']

    def test_start(self):
        self.execute_assert()


class STTestRC125(CommonTestCase):
    testcase_name = "rc_remote_local"
    code = fault_description.ALL_NOTIFY_ERROR_NOT_TIMEOUT_REMOTE_LOCAL.code
    device = ['worker-0 device-2']

    def test_start(self):
        self.execute_assert()


class STTestRC126(CommonTestCase):
    testcase_name = "rc_126"
    code = fault_description.CLUSTER_EXCEPTION_LOCATION_ERROR.code
    device = ["w_1 device-4"]

    def test_start(self):
        self.execute_assert()


class STTestRC128(CommonTestCase):
    testcase_name = "rc_128"
    code = fault_description.LACK_OF_BASE_INFO_AFTER_RESUMING_TRAINING.code
    device = ["worker-0 device-Unknown"]

    def test_start(self):
        self.execute_assert()


class DiagSTArgs:
    cmd = "diag"

    def __init__(self, input_dir, output_dir, mode=0, task_id="test_uuid", scene="host"):
        self.input_path = input_dir
        self.output_path = output_dir
        self.mode = mode
        self.task_id = task_id
        self.performance = False
        self.scene = scene


class STTestRC(unittest.TestCase):
    def setUp(self) -> None:
        self.device_table = DeviceTable()
        self.rc_diag_worker = RCDiagWorker(
            cfg=DiagCFG(task_id="test_dt", input_path="", output_path="",
                        parsed_saver=ParsedDataSaver("", args=DiagSTArgs("", ""))))

    def test_102(self):
        err_checker = ErrorChecker(self.device_table)
        self.check_and_assert(err_checker, fault_description.ALL_NO_ERROR.code)

    def test_115(self):
        err_checker = NoPlogChecker(self.device_table)
        self.check_and_assert(err_checker, fault_description.NO_PLOG_ERROR.code)

    def test_127(self):
        err_checker = NoValidPlogInfoErrorChecker(self.device_table)
        self.check_and_assert(err_checker, fault_description.NO_VALID_PLOG_INFO_ERROR.code)

    def test_114(self):
        err_checker = self.rc_diag_worker._generate_checker()
        self.check_and_assert(err_checker, fault_description.INVALID_DEVICE_ERROR.code)

    def test_128(self):
        err_checker = ResumingTrainingInvalidBaseInfoChecker(self.device_table, {Device()})
        self.check_and_assert(err_checker, fault_description.LACK_OF_BASE_INFO_AFTER_RESUMING_TRAINING.code)

    def check_and_assert(self, err_checker, code):
        err_checker.check()
        result = err_checker.format_output()
        self.assertTrue(result.analyze_success)
        self.assertEqual(result.fault_description.code, code)
        self.assertEqual(result.root_cause_device, ["Unknown Device"])

    def test_parse_plog_for_old_version(self):
        plog_list = [os.path.join(TEST_DIR, "st_module_testcase", "rc_diag", "init_failed_with_timeout", "worker-0",
                                  "plog-parser-123-1.log")]
        device_info_dict = {'device_ip': {}, 'device_tls': {'123': '1'}}
        result = self.rc_diag_worker._parse_plog_for_old_version(plog_list, device_info_dict)
        pid_info = result.get("123", {})
        base_result = pid_info.get("base", {})
        error_result = pid_info.get("error", {})
        self.assertEqual(base_result.get("logic_device_id"), "2")
        self.assertEqual("HCCL", error_result.get("first_error_module"))
        self.assertEqual("1", pid_info.get("tls_status"))


class STTestClass(unittest.TestCase):

    def setUp(self) -> None:
        self.identifier_instance = Identifier(identifier_name="hccl_world_group")
        store_device = Device(pid="pid_1", worker_name="worker-NA", device_table=DeviceTable())
        store_device.phy_device_id = "1"
        self.identifier_instance.update_device(store_device, "2")
        self.identifier_instance.update_device(Device(pid="pid_2", worker_name="worker-1",
                                                      device_table=DeviceTable()), "3")
        device_table = DeviceTable()
        identifier_dict = {self.identifier_instance.name: self.identifier_instance}
        device_table.identifier_dict.update(identifier_dict)
        self.device = Device(pid="pid_1", worker_name="worker-0", device_table=device_table)
        self.device.identifier_map.update(identifier_dict)

        self.rank_device = RankDevice(identifier=self.identifier_instance.name, rank_id="1")

    def test_start(self):
        self.assertEqual(repr(self.device), "worker-0 device-Unknown")
        self.assertEqual(hash(self.identifier_instance),
                         hash(Identifier(identifier_name="hccl_world_group", rank_num=2)))
        self.assertEqual(self.rank_device, RankDevice(identifier=self.identifier_instance.name, rank_id="1"))
        self.assertEqual(repr(self.rank_device), "hccl_world_group rank-1")
