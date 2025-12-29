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
import argparse
import os
import unittest
import tempfile
import shutil
from unittest.mock import patch, MagicMock

import ascend_fd.pkg.parse.parser_saver
from ascend_fd.pkg.parse.parser_saver import TrainLogSaver, ProcessLogSaver
from ascend_fd.utils import tool
from ascend_fd.pkg.diag.knowledge_graph.kg_diag_job import get_super_pod_analyzer_dict


class VersionTestCase(unittest.TestCase):
    TEST_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
    TESTCASE_DL_LOG_INPUT = os.path.join(TEST_DIR, "st_module_testcase", "kg_parse", "dl_log")
    TESTCASE_PLOG_INPUT = os.path.join(TEST_DIR, "st_testcase", "modelarts-job-testdt", "ascend", "process_log")

    def setUp(self) -> None:
        pass

    def test_get_version(self):
        self.assertIsNotNone(tool.get_version())

    def test_path_check(self):
        self.assertIn(os.path.realpath(__file__), tool.file_check(os.path.realpath(__file__)))
        self.assertIn(os.path.dirname(os.path.realpath(__file__)),
                      tool.dir_check(os.path.dirname(os.path.realpath(__file__))))
        self.assertIn(os.path.realpath(__file__), tool.file_or_dir_check(os.path.realpath(__file__)))
        self.assertIn(os.path.dirname(os.path.realpath(__file__)), tool.file_or_dir_check(os.path.dirname(__file__)))
        # non-existent file
        self.assertRaises(argparse.ArgumentTypeError, tool.path_check, "11.txt")
        # illegal named file
        self.assertRaises(argparse.ArgumentTypeError, tool.path_check, "11%%%.txt")
        # file name length is less than 1
        self.assertRaises(argparse.ArgumentTypeError, tool.path_check, "")

    def test_get_user_info(self):
        tool.get_user_info()

    def test_init_home_path_by_env(self):
        self.assertTrue(tool._init_home_path_by_env())

    def test_train_log_saver(self):
        log_saver = TrainLogSaver()
        log_saver.filter_log(None)
        self.assertEqual([], log_saver.get_train_log())
        log_saver.filter_log([os.path.dirname(__file__)])
        self.assertEqual([], log_saver.get_train_log())
        log_saver.filter_log([os.path.realpath(__file__)])
        self.assertEqual([os.path.realpath(__file__)], log_saver.get_train_log())

    def test_dl_log_saver(self):
        log_saver = ascend_fd.pkg.parse.parser_saver.DlLogSaver()
        log_saver.filter_log(os.path.realpath(__file__))
        self.assertEqual([], log_saver.device_plugin_list)
        log_saver.filter_log(self.TESTCASE_DL_LOG_INPUT)
        self.assertIn(os.path.join(self.TESTCASE_DL_LOG_INPUT, "devicePlugin", "devicePlugin.log"),
                      log_saver.device_plugin_list)
        self.assertIn(os.path.join(self.TESTCASE_DL_LOG_INPUT, "devicePlugin",
                                   "devicePlugin-2024-01-10T03-30-45.197.log"), log_saver.device_plugin_list)

    def test_resuming_training_fetch(self):
        log_saver = ProcessLogSaver()
        log_saver.filter_log(self.TESTCASE_PLOG_INPUT)
        self.assertEqual(log_saver.resuming_training_time, "2023-01-01 02:00:00.000000")

    def tearDown(self) -> None:
        pass


class TestFilterBmcLog(unittest.TestCase):
    def setUp(self):
        self.temp_dir = tempfile.mkdtemp()
        self.instance = ascend_fd.pkg.parse.parser_saver.BMCLogSaver()

    def tearDown(self):
        shutil.rmtree(self.temp_dir)

    def test_no_directory(self):
        self.instance.filter_log(None)
        self.assertEqual(self.instance.fruinfo_files, [])
        self.assertEqual(self.instance.mdb_info_files, [])
        self.assertEqual(self.instance.bmc_app_dump_log_list, [])
        self.assertEqual(self.instance.bmc_device_dump_log_list, [])
        self.assertEqual(self.instance.bmc_log_dump_log_list, [])
        self.assertEqual(self.instance.bmc_log_list, [])

    def test_empty_directory(self):
        self.instance.filter_log(self.temp_dir)
        self.assertEqual(self.instance.fruinfo_files, [])
        self.assertEqual(self.instance.mdb_info_files, [])
        self.assertEqual(self.instance.bmc_app_dump_log_list, [])
        self.assertEqual(self.instance.bmc_device_dump_log_list, [])
        self.assertEqual(self.instance.bmc_log_dump_log_list, [])
        self.assertEqual(self.instance.bmc_log_list, [])

    @patch('os.path.isdir', return_value=False)
    def test_not_directory(self, mock_isdir):
        self.instance.filter_log(self.temp_dir)
        self.assertEqual(self.instance.fruinfo_files, [])
        self.assertEqual(self.instance.mdb_info_files, [])
        self.assertEqual(self.instance.bmc_app_dump_log_list, [])
        self.assertEqual(self.instance.bmc_device_dump_log_list, [])
        self.assertEqual(self.instance.bmc_log_dump_log_list, [])
        self.assertEqual(self.instance.bmc_log_list, [])

    def test_with_files(self):
        with open(os.path.join(self.temp_dir, 'fruinfo.txt'), 'w') as f:
            f.write('test')
        os.makedirs(os.path.join(self.temp_dir, "chassis"), exist_ok=True)
        os.makedirs(os.path.join(self.temp_dir, "AppDump"), exist_ok=True)
        os.makedirs(os.path.join(self.temp_dir, "DeviceDump"), exist_ok=True)
        os.makedirs(os.path.join(self.temp_dir, "LogDump"), exist_ok=True)
        with open(os.path.join(self.temp_dir, "chassis", 'mdb_info.log'), 'w') as f:
            f.write('test')
        with open(os.path.join(self.temp_dir, "AppDump", 'app_dump.log'), 'w') as f:
            f.write('test')
        with open(os.path.join(self.temp_dir, "DeviceDump", 'device_dump.log'), 'w') as f:
            f.write('test')
        with open(os.path.join(self.temp_dir, "LogDump", 'log_dump.log'), 'w') as f:
            f.write('test')
        with open(os.path.join(self.temp_dir, 'bmc.log'), 'w') as f:
            f.write('test')

        self.instance.filter_log(self.temp_dir)
        self.assertEqual(self.instance.fruinfo_files, [os.path.join(self.temp_dir, 'fruinfo.txt')])
        self.assertEqual(self.instance.mdb_info_files, [os.path.join(self.temp_dir, "chassis", 'mdb_info.log')])
        self.assertEqual(self.instance.bmc_app_dump_log_list, [os.path.join(self.temp_dir, "AppDump", 'app_dump.log')])
        self.assertEqual(self.instance.bmc_device_dump_log_list,
                         [os.path.join(self.temp_dir, "DeviceDump", 'device_dump.log')])
        self.assertEqual(self.instance.bmc_log_dump_log_list, [os.path.join(self.temp_dir, "LogDump", 'log_dump.log')])
        self.assertEqual(self.instance.bmc_log_list, [])


class TestGetSuperPodAnalyzerDict(unittest.TestCase):
    def setUp(self):
        self.cfg = MagicMock()
        self.cfg.root_worker_devices = {'worker1': 'path1', 'worker2': 'path2'}
        self.parsed_saver = MagicMock()
        self.parsed_saver.infer_task_flag = False
        self.parsed_saver.bmc_path_dict = {'worker3': 'path3', 'worker4': 'path4'}
        self.parsed_saver.lcne_path_dict = {'worker5': 'path5', 'worker6': 'path6'}

    @patch('ascend_fd.pkg.diag.knowledge_graph.kg_diag_job.get_host_worker_name_by_bmc_worker_name')
    @patch('ascend_fd.pkg.diag.knowledge_graph.kg_diag_job.get_host_worker_name_by_lcne_worker_name')
    def test_infer_task_flag_false(self, mock_lcne_func, mock_bmc_func):
        analyzer_dict = get_super_pod_analyzer_dict(self.cfg, self.parsed_saver)
        expected_dict = {'worker1': 'path1', 'worker2': 'path2', 'worker3': 'path3', 'worker4': 'path4',
                         'worker5': 'path5', 'worker6': 'path6'}
        self.assertEqual(analyzer_dict, expected_dict)
        # 验证 mock 函数没有被调用（因为 infer_task_flag 为 False）
        mock_bmc_func.assert_not_called()
        mock_lcne_func.assert_not_called()

    @patch('ascend_fd.pkg.diag.knowledge_graph.kg_diag_job.get_host_worker_name_by_bmc_worker_name')
    @patch('ascend_fd.pkg.diag.knowledge_graph.kg_diag_job.get_host_worker_name_by_lcne_worker_name')
    def test_infer_task_flag_true(self, mock_lcne_func, mock_bmc_func):
        self.parsed_saver.infer_task_flag = True
        self.parsed_saver.infer_instance = 'instance1'
        self.parsed_saver.cluster_info = {'instance1': ['ip1', 'ip2']}
        self.parsed_saver.container_worker_map = {'ip1': 'worker1', 'ip2': 'worker2'}
        self.parsed_saver.bmc_path_dict = {'worker3': 'path3', 'worker4': 'path4'}
        self.parsed_saver.lcne_path_dict = {'worker5': 'path5'}

        # 配置 mock 函数的返回值
        mock_bmc_func.return_value = 'worker7'
        mock_lcne_func.return_value = 'worker7'

        analyzer_dict = get_super_pod_analyzer_dict(self.cfg, self.parsed_saver)
        expected_dict = {'worker1': 'path1', 'worker2': 'path2'}
        self.assertEqual(analyzer_dict, expected_dict)

        mock_bmc_func.return_value = 'worker1'
        mock_lcne_func.return_value = 'worker1'

        analyzer_dict = get_super_pod_analyzer_dict(self.cfg, self.parsed_saver)
        expected_dict = {'worker1': 'path1', 'worker2': 'path2', 'worker3': 'path3', 'worker4': 'path4',
                         'worker5': 'path5'}
        self.assertEqual(analyzer_dict, expected_dict)
        # 验证 mock 函数被调用
        mock_bmc_func.assert_called()
        mock_lcne_func.assert_called()


if __name__ == '__main__':
    unittest.main()
