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
import gzip
import io
import os
import tarfile
import unittest
import zipfile
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
        self.assertIn(
            os.path.dirname(os.path.realpath(__file__)), tool.dir_check(os.path.dirname(os.path.realpath(__file__)))
        )
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
        self.assertIn(
            os.path.join(self.TESTCASE_DL_LOG_INPUT, "devicePlugin", "devicePlugin.log"), log_saver.device_plugin_list
        )
        self.assertIn(
            os.path.join(self.TESTCASE_DL_LOG_INPUT, "devicePlugin", "devicePlugin-2024-01-10T03-30-45.197.log"),
            log_saver.device_plugin_list,
        )

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
        with open(os.path.join(self.temp_dir, 'fruinfo.txt'), 'w', encoding='utf-8') as f:
            f.write('test')
        os.makedirs(os.path.join(self.temp_dir, "chassis"), exist_ok=True)
        os.makedirs(os.path.join(self.temp_dir, "AppDump"), exist_ok=True)
        os.makedirs(os.path.join(self.temp_dir, "DeviceDump"), exist_ok=True)
        os.makedirs(os.path.join(self.temp_dir, "LogDump"), exist_ok=True)
        with open(os.path.join(self.temp_dir, "chassis", 'mdb_info.log'), 'w', encoding='utf-8') as f:
            f.write('test')
        with open(os.path.join(self.temp_dir, "AppDump", 'app_dump.log'), 'w', encoding='utf-8') as f:
            f.write('test')
        with open(os.path.join(self.temp_dir, "DeviceDump", 'device_dump.log'), 'w', encoding='utf-8') as f:
            f.write('test')
        with open(os.path.join(self.temp_dir, "LogDump", 'remote_log'), 'w', encoding='utf-8') as f:
            f.write('test')
        with open(os.path.join(self.temp_dir, 'bmc.log'), 'w', encoding='utf-8') as f:
            f.write('test')

        self.instance.filter_log(self.temp_dir)
        self.assertEqual(self.instance.fruinfo_files, [os.path.join(self.temp_dir, 'fruinfo.txt')])
        self.assertEqual(self.instance.mdb_info_files, [os.path.join(self.temp_dir, "chassis", 'mdb_info.log')])
        self.assertEqual(self.instance.bmc_app_dump_log_list, [os.path.join(self.temp_dir, "AppDump", 'app_dump.log')])
        self.assertEqual(
            self.instance.bmc_device_dump_log_list, [os.path.join(self.temp_dir, "DeviceDump", 'device_dump.log')]
        )
        self.assertEqual(self.instance.bmc_log_dump_log_list, [os.path.join(self.temp_dir, "LogDump", 'remote_log')])
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
        expected_dict = {
            'worker1': 'path1',
            'worker2': 'path2',
            'worker3': 'path3',
            'worker4': 'path4',
            'worker5': 'path5',
            'worker6': 'path6',
        }
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
        expected_dict = {
            'worker1': 'path1',
            'worker2': 'path2',
            'worker3': 'path3',
            'worker4': 'path4',
            'worker5': 'path5',
        }
        self.assertEqual(analyzer_dict, expected_dict)
        # 验证 mock 函数被调用
        mock_bmc_func.assert_called()
        mock_lcne_func.assert_called()


class TestDeviceInfoMap(unittest.TestCase):
    """A5 plog: device_info.txt 解析测试"""

    def test_parse_device_info_file_empty(self):
        """空文件/不存在文件"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False) as f:
            f.write("")
        try:
            self.assertEqual(tool.parse_device_info_file(f.name), {})
        finally:
            os.unlink(f.name)
        self.assertEqual(tool.parse_device_info_file("/non/existent/path"), {})

    def test_parse_device_info_file_format(self):
        """标准格式"""
        content = """base info:
==============================================
device num: 0x8

devices info:
==============================================
dir        phy-id        logic-id   status
device-0   0             3          os running
device-1   1             4          os running
device-2   2             6          os running
device-3   3             7          os running
device-4   4             2          os running
device-5   5             0          os running
device-6   6             1          os running
device-7   7             5          os running
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False) as f:
            f.write(content)
        try:
            expected = {"3": "0", "4": "1", "6": "2", "7": "3", "2": "4", "0": "5", "1": "6", "5": "7"}
            self.assertEqual(tool.parse_device_info_file(f.name), expected)
        finally:
            os.unlink(f.name)

    def test_parse_device_info_file_malformed(self):
        """格式异常行"""
        content = """devices info:
==============================================
dir        phy-id        logic-id   status
device-0   0             3          os running
device-X   a             b          os running
device-1   1             4          os running
"""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".txt", delete=False) as f:
            f.write(content)
        try:
            self.assertEqual(tool.parse_device_info_file(f.name), {"3": "0", "4": "1"})
        finally:
            os.unlink(f.name)

    def test_load_device_info_map_empty(self):
        """空列表"""
        self.assertEqual(tool.load_device_info_map([]), {})

    def test_load_device_info_map_found(self):
        """找到 device_info.txt"""
        with tempfile.NamedTemporaryFile(mode="w", suffix="device_info.txt", delete=False) as f:
            f.write("""devices info:
==============================================
dir        phy-id        logic-id   status
device-0   0             3          os running
device-1   1             4          os running
""")
        try:
            hisi_logs = ["/hisi_logs/device-0/kernel.log", f.name, "/hisi_logs/device-1/kernel.log"]
            self.assertEqual(tool.load_device_info_map(hisi_logs), {"3": "0", "4": "1"})
        finally:
            os.unlink(f.name)

    def test_load_device_info_map_not_found(self):
        """没有 device_info.txt"""
        self.assertEqual(tool.load_device_info_map(["/hisi_logs/device-0/kernel.log"]), {})


class TestPatternMatcherAll(unittest.TestCase):
    """PatternMatcher 的 all 语法测试（单行匹配）"""

    def setUp(self):
        self.matcher = tool.PatternMatcher()

    def test_all_single_group_same_line(self):
        """单组关键字在同一行内匹配成功"""
        self.assertTrue(self.matcher.compare({"all": [["key1", "key2"]]}, "xxx key1 yyy key2 zzz"))

    def test_all_single_group_missing_keyword(self):
        """单组关键字缺少一个，匹配失败"""
        self.assertFalse(self.matcher.compare({"all": [["key1", "key2"]]}, "xxx key1 yyy zzz"))

    def test_all_multi_groups_all_matched_same_line(self):
        """多组关键字全部在同一行匹配成功"""
        self.assertTrue(self.matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, "key1 key2 key3 key4"))

    def test_all_multi_groups_one_not_matched(self):
        """多组关键字中有一组未匹配，匹配失败"""
        self.assertFalse(self.matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, "key1 key2 key3"))

    def test_all_flat_form_single_group(self):
        """扁平形式 ["k1", "k2"] 等价于 [["k1", "k2"]]"""
        self.assertTrue(self.matcher.compare({"all": ["key1", "key2"]}, "key1 key2"))

    def test_all_empty(self):
        """all 为空列表时匹配失败"""
        self.assertFalse(self.matcher.compare({"all": []}, "key1 key2"))

    def test_all_empty_inner_group_consistency(self):
        """all 含空内层列表时过滤空组：全空返回 False，部分空按剩余组匹配"""
        self.assertFalse(self.matcher._match_all([[]], "key1 key2"))
        self.assertFalse(self.matcher.compare({"all": [[]]}, "key1 key2"))
        self.assertTrue(self.matcher.compare({"all": [[], ["key1", "key2"]]}, "key1 key2"))
        self.assertFalse(self.matcher.compare({"all": [[], ["key9"]]}, "key1 key2"))

    def test_all_combined_with_in(self):
        """all 与 in 共存，in 命中即成功"""
        self.assertTrue(self.matcher.compare({"in": [["key1"]], "all": [["key3", "key4"]]}, "key1 only"))

    def test_all_combined_with_regex(self):
        """all 与 regex 共存，regex 命中即成功"""
        self.assertTrue(self.matcher.compare({"regex": "key1", "all": [["key3", "key4"]]}, "key1 only"))


class TestPatternSingleOrMultiLineMatcherAll(unittest.TestCase):
    """PatternSingleOrMultiLineMatcher 的 all 语法测试（跨行匹配）"""

    @staticmethod
    def _build_matcher_with_lines(lines, idx=0):
        matcher = tool.PatternSingleOrMultiLineMatcher(log_lines=lines)
        matcher.update_line_index(idx)
        return matcher

    def test_all_multi_groups_cross_line(self):
        """多组关键字跨行匹配：group1 在当前行，group2 在下一行"""
        lines = ["key1 key2", "key3 key4", "other line"]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertTrue(matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, lines[0]))

    def test_all_single_group_cross_line(self):
        """单组关键字跨行匹配：key1 在当前行，key2 在下一行"""
        lines = ["key1 something", "key2 something", "other"]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertTrue(matcher.compare({"all": [["key1", "key2"]]}, lines[0]))

    def test_all_group_beyond_window(self):
        """第二组关键字超出 10 行窗口，匹配失败"""
        lines = ["key1 key2"] + ["filler"] * 10 + ["key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertFalse(matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, lines[0]))

    def test_all_group_within_window(self):
        """第二组关键字在 10 行窗口内，匹配成功"""
        lines = ["key1 key2"] + ["filler"] * 8 + ["key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertTrue(matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, lines[0]))

    def test_all_anchor_no_first_keyword_on_current_line(self):
        """当前行不包含任何组的首关键字，匹配失败（不读取窗口）"""
        lines = ["no anchor here", "key1 key2", "key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertFalse(matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, lines[0]))

    def test_all_anchor_via_group2_first_keyword(self):
        """当前行包含 group2 的首关键字作为锚点，两组在窗口内匹配成功"""
        lines = ["key3 key4", "key1 key2"]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertTrue(matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, lines[0]))

    def test_all_flat_form_multi_line(self):
        """扁平形式在多行匹配器下跨行匹配成功"""
        lines = ["key1 abc", "key2 def"]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertTrue(matcher.compare({"all": ["key1", "key2"]}, lines[0]))

    def test_all_with_file_stream(self):
        """通过 file_stream 进行跨行匹配"""
        stream = io.StringIO("key1 key2\nkey3 key4\nother\n")
        matcher = tool.PatternSingleOrMultiLineMatcher(file_stream=stream)
        self.assertTrue(matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, "key1 key2"))

    def test_all_cross_line_at_different_index(self):
        """从中间行开始跨行匹配：锚点行后续窗口内包含另一组"""
        lines = ["filler", "key1 key2", "filler", "key3 key4", "filler"]
        matcher = self._build_matcher_with_lines(lines, 1)
        self.assertTrue(matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, lines[1]))

    def test_all_one_group_not_in_window(self):
        """多组中其中一组在窗口内，另一组完全不在窗口内，匹配失败"""
        lines = [
            "key1 key2",
            "filler",
            "filler",
            "filler",
            "filler",
            "filler",
            "filler",
            "filler",
            "filler",
            "filler",
            "filler",
            "key3 key4",
        ]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertFalse(matcher.compare({"all": [["key1", "key2"], ["key3", "key4"]]}, lines[0]))

    def test_all_max_lines_smaller_window_miss(self):
        """max_lines 限制窗口为 2 行，第二组在第 3 行（窗口外），匹配失败"""
        lines = ["key1 key2", "filler", "key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        conf = {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 2}
        self.assertFalse(matcher.compare(conf, lines[0]))

    def test_all_max_lines_smaller_window_hit(self):
        """max_lines 限制窗口为 3 行，第二组在第 3 行（窗口内），匹配成功"""
        lines = ["key1 key2", "filler", "key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        conf = {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 3}
        self.assertTrue(matcher.compare(conf, lines[0]))

    def test_all_max_lines_zero_only_current_line(self):
        """max_lines 为 0 时只匹配当前行，第二组在下一行匹配失败"""
        lines = ["key1 key2", "key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        conf = {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 0}
        self.assertFalse(matcher.compare(conf, lines[0]))

    def test_all_max_lines_negative_falls_back_to_default(self):
        """max_lines 小于 0 时回退到默认 10 行，第二组在默认窗口内匹配成功"""
        lines = ["key1 key2"] + ["filler"] * 8 + ["key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        conf = {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": -1}
        self.assertTrue(matcher.compare(conf, lines[0]))

    def test_all_max_lines_negative_default_window_boundary(self):
        """max_lines 小于 0 时回退到默认 10 行，第二组超出默认窗口匹配失败"""
        lines = ["key1 key2"] + ["filler"] * 10 + ["key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        conf = {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": -5}
        self.assertFalse(matcher.compare(conf, lines[0]))

    def test_all_max_lines_with_file_stream(self):
        """max_lines 在 file_stream 模式下生效：窗口为 2 行时第二组在第 3 行匹配失败"""
        stream = io.StringIO("key1 key2\nfiller\nkey3 key4\nother\n")
        matcher = tool.PatternSingleOrMultiLineMatcher(file_stream=stream)
        conf = {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 2}
        self.assertFalse(matcher.compare(conf, "key1 key2"))

    def test_all_max_lines_with_file_stream_hit(self):
        """max_lines 在 file_stream 模式下生效：窗口为 3 行时第二组在第 3 行匹配成功"""
        stream = io.StringIO("key1 key2\nfiller\nkey3 key4\nother\n")
        matcher = tool.PatternSingleOrMultiLineMatcher(file_stream=stream)
        conf = {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 3}
        self.assertTrue(matcher.compare(conf, "key1 key2"))

    def test_all_max_lines_not_affect_single_line_matcher(self):
        """max_lines 对单行 PatternMatcher 无影响，仅按同行匹配"""
        matcher = tool.PatternMatcher()
        conf = {"all": [["key1", "key2"]], "max_lines": 0}
        self.assertTrue(matcher.compare(conf, "key1 key2"))
        self.assertFalse(matcher.compare(conf, "key1 other"))

    def test_all_max_lines_null_falls_back_to_default(self):
        """max_lines 显式为 null 时回退到默认窗口，不抛 TypeError"""
        lines = ["key1 key2"] + ["filler"] * 8 + ["key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        conf = {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": None}
        self.assertTrue(matcher.compare(conf, lines[0]))

    def test_all_empty_inner_group_consistency_multi(self):
        """多行匹配器下 all 含空内层列表时过滤空组：全空返回 False，部分空按剩余组匹配"""
        lines = ["key1 key2", "key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertFalse(matcher._match_all([[]], lines[0]))
        self.assertFalse(matcher.compare({"all": [[]]}, lines[0]))
        self.assertTrue(matcher.compare({"all": [[], ["key1", "key2"]]}, lines[0]))
        self.assertFalse(matcher.compare({"all": [[], ["key9"]]}, lines[0]))


class TestPatternMatcherOpt(unittest.TestCase):
    """opt 语法测试：opt 为选项列表，选项间 OR，每选项内部 in/regex/all"""

    def test_opt_single_option_all_matched_same_line(self):
        """opt 单选项，all 两组关键字同行全部匹配成功"""
        matcher = tool.PatternMatcher()
        conf = {"opt": [{"all": [["key1", "key2"], ["key3", "key4"]]}]}
        self.assertTrue(matcher.compare(conf, "key1 key2 key3 key4"))

    def test_opt_single_option_all_not_matched(self):
        """opt 单选项，all 缺一组匹配失败"""
        matcher = tool.PatternMatcher()
        conf = {"opt": [{"all": [["key1", "key2"], ["key3", "key4"]]}]}
        self.assertFalse(matcher.compare(conf, "key1 key2 key3"))

    def test_opt_multi_options_first_matches(self):
        """opt 多选项，第一个选项命中即成功"""
        matcher = tool.PatternMatcher()
        conf = {"opt": [{"all": [["key1", "key2"]]}, {"in": [["key3", "key4"]]}]}
        self.assertTrue(matcher.compare(conf, "key1 key2"))

    def test_opt_multi_options_second_matches(self):
        """opt 多选项，第一个失败第二个命中即成功（OR 逻辑）"""
        matcher = tool.PatternMatcher()
        conf = {"opt": [{"all": [["key1", "key2"]]}, {"in": [["key3", "key4"]]}]}
        self.assertTrue(matcher.compare(conf, "key3 key4"))

    def test_opt_multi_options_none_matches(self):
        """opt 多选项，全部不匹配则失败"""
        matcher = tool.PatternMatcher()
        conf = {"opt": [{"all": [["key1", "key2"]]}, {"in": [["key3", "key4"]]}]}
        self.assertFalse(matcher.compare(conf, "unrelated text"))

    def test_opt_empty_list_falls_back_to_top_level(self):
        """opt 为空列表时回退到顶层 in 匹配（向后兼容）"""
        matcher = tool.PatternMatcher()
        conf = {"opt": [], "in": [["key1", "key2"]]}
        self.assertTrue(matcher.compare(conf, "key1 key2"))

    def test_opt_absent_uses_top_level(self):
        """无 opt 时使用顶层 in/regex/all（向后兼容）"""
        matcher = tool.PatternMatcher()
        self.assertTrue(matcher.compare({"in": [["key1"]]}, "key1"))
        self.assertTrue(matcher.compare({"all": [["key1", "key2"]]}, "key1 key2"))

    def test_opt_null_does_not_crash(self):
        """opt 显式为 null 时安全回退到顶层匹配，不抛 TypeError"""
        matcher = tool.PatternMatcher()
        self.assertFalse(matcher.compare({"opt": None}, "key1 key2"))
        self.assertTrue(matcher.compare({"opt": None, "in": [["key1"]]}, "key1"))

    def test_null_fields_does_not_crash(self):
        """in/regex/all/max_lines 显式为 null 时安全处理，不抛 TypeError"""
        matcher = tool.PatternMatcher()
        self.assertFalse(matcher.compare({"in": None}, "key1"))
        self.assertFalse(matcher.compare({"regex": None}, "key1"))
        self.assertFalse(matcher.compare({"all": None}, "key1"))


class TestPatternSingleOrMultiLineMatcherOpt(unittest.TestCase):
    """opt 语法在多行匹配器下的跨行匹配测试"""

    @staticmethod
    def _build_matcher_with_lines(lines, idx=0):
        matcher = tool.PatternSingleOrMultiLineMatcher(log_lines=lines)
        matcher.update_line_index(idx)
        return matcher

    def test_opt_all_cross_line_within_max_lines(self):
        """opt 选项内 all 两组跨行匹配，max_lines=20 窗口内成功"""
        lines = ["key1 key2", "filler", "key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        conf = {"opt": [{"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 20}]}
        self.assertTrue(matcher.compare(conf, lines[0]))

    def test_opt_all_cross_line_beyond_max_lines(self):
        """opt 选项内 all 第二组超出 max_lines 窗口匹配失败"""
        lines = ["key1 key2", "filler", "key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        conf = {"opt": [{"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 2}]}
        self.assertFalse(matcher.compare(conf, lines[0]))

    def test_opt_per_option_max_lines_independent(self):
        """opt 多选项各自 max_lines 独立：选项1窗口小失败，选项2窗口大成功"""
        lines = ["key1 key2", "filler", "filler", "key3 key4"]
        matcher = self._build_matcher_with_lines(lines, 0)
        conf = {
            "opt": [
                {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 2},
                {"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 20},
            ]
        }
        self.assertTrue(matcher.compare(conf, lines[0]))

    def test_opt_real_config_mooncake_002(self):
        """真实配置 MOONCAKE_002：opt 内 all 两组跨行匹配"""
        conf = {
            "opt": [
                {
                    "all": [
                        ["send request to", "connection refused"],
                        [
                            "Initialize mooncake failed",
                            "metadata_server",
                            "P2PHANDSHAKE",
                            "Check mooncake config and network",
                        ],
                    ],
                    "max_lines": 20,
                }
            ]
        }
        lines = [
            "send request to connection refused",
            "Initialize mooncake failed",
            "metadata_server P2PHANDSHAKE",
            "Check mooncake config and network",
        ]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertTrue(matcher.compare(conf, lines[0]))

    def test_opt_real_config_mooncake_002_missing_group(self):
        """真实配置 MOONCAKE_002：仅一组匹配，all 要求两组故失败"""
        conf = {
            "opt": [
                {
                    "all": [
                        ["send request to", "connection refused"],
                        [
                            "Initialize mooncake failed",
                            "metadata_server",
                            "P2PHANDSHAKE",
                            "Check mooncake config and network",
                        ],
                    ],
                    "max_lines": 20,
                }
            ]
        }
        lines = ["send request to connection refused"]
        matcher = self._build_matcher_with_lines(lines, 0)
        self.assertFalse(matcher.compare(conf, lines[0]))

    def test_opt_with_file_stream(self):
        """opt 语法在 file_stream 模式下跨行匹配"""
        stream = io.StringIO("key1 key2\nfiller\nkey3 key4\nother\n")
        matcher = tool.PatternSingleOrMultiLineMatcher(file_stream=stream)
        conf = {"opt": [{"all": [["key1", "key2"], ["key3", "key4"]], "max_lines": 20}]}
        self.assertTrue(matcher.compare(conf, "key1 key2"))


class TestDecompress(unittest.TestCase):
    """decompress_gz / decompress_zip / decompress_tar_gz 测试"""

    def setUp(self):
        self.temp_dir = tempfile.mkdtemp()

    def tearDown(self):
        shutil.rmtree(self.temp_dir, ignore_errors=True)

    def _make_gz(self, name, content=b"hello"):
        path = os.path.join(self.temp_dir, name)
        with gzip.open(path, "wb") as f:
            f.write(content)
        return path

    def _make_zip(self, name, files):
        path = os.path.join(self.temp_dir, name)
        with zipfile.ZipFile(path, "w") as zf:
            for fname, data in files.items():
                zf.writestr(fname, data)
        return path

    def _make_tar_gz(self, name, files):
        path = os.path.join(self.temp_dir, name)
        with tarfile.open(path, "w:gz") as tf:
            for fname, data in files.items():
                info = tarfile.TarInfo(name=fname)
                info.size = len(data)
                tf.addfile(info, io.BytesIO(data))
        return path

    def test_decompress_gz_success(self):
        gz_path = self._make_gz("a.log.gz", b"hello world")
        out = tool.decompress_gz(gz_path)
        self.assertTrue(out and os.path.isfile(out))
        with open(out, "rb") as f:
            self.assertEqual(f.read(), b"hello world")

    def test_decompress_gz_not_exist(self):
        self.assertEqual(tool.decompress_gz(os.path.join(self.temp_dir, "no.gz")), "")

    def test_decompress_gz_wrong_suffix(self):
        path = os.path.join(self.temp_dir, "a.txt")
        with open(path, "w", encoding="utf-8") as f:
            f.write("x")
        self.assertEqual(tool.decompress_gz(path), "")

    @patch("ascend_fd.utils.tool.MAX_SIZE", 256)
    def test_decompress_gz_bomb(self):
        """压缩炸弹：压缩后很小，解压后超过 MAX_SIZE，应被流式拦截"""
        big_content = b"x" * 1024  # 解压后 1KB，超过 patch 后的 256
        gz_path = self._make_gz("bomb.log.gz", big_content)
        out = tool.decompress_gz(gz_path)
        self.assertEqual(out, "")
        self.assertFalse(os.path.exists(os.path.join(self.temp_dir, "bomb.log")))

    def test_decompress_zip_success(self):
        zip_path = self._make_zip("b.zip", {"f1.log": "data1", "sub/f2.log": "data2"})
        out = tool.decompress_zip(zip_path)
        self.assertTrue(out and os.path.isdir(out))
        with open(os.path.join(out, "f1.log"), encoding="utf-8") as f:
            self.assertEqual(f.read(), "data1")
        with open(os.path.join(out, "sub", "f2.log"), encoding="utf-8") as f:
            self.assertEqual(f.read(), "data2")

    def test_decompress_zip_not_exist(self):
        self.assertEqual(tool.decompress_zip(os.path.join(self.temp_dir, "no.zip")), "")

    def test_decompress_zip_wrong_suffix(self):
        path = os.path.join(self.temp_dir, "b.txt")
        with open(path, "w", encoding="utf-8") as f:
            f.write("x")
        self.assertEqual(tool.decompress_zip(path), "")

    def test_decompress_zip_path_traversal(self):
        zip_path = self._make_zip("evil.zip", {"../escape.log": "bad"})
        out = tool.decompress_zip(zip_path)
        self.assertEqual(out, "")
        self.assertFalse(os.path.exists(os.path.join(self.temp_dir, "escape.log")))

    def test_decompress_tar_gz_success(self):
        tar_path = self._make_tar_gz("c.tar.gz", {"f1.log": b"d1", "sub/f2.log": b"d2"})
        out = tool.decompress_tar_gz(tar_path)
        self.assertTrue(out and os.path.isdir(out))
        with open(os.path.join(out, "f1.log"), "rb") as f:
            self.assertEqual(f.read(), b"d1")
        with open(os.path.join(out, "sub", "f2.log"), "rb") as f:
            self.assertEqual(f.read(), b"d2")

    def test_decompress_tgz_success(self):
        tar_path = self._make_tar_gz("c.tgz", {"f.log": b"x"})
        out = tool.decompress_tar_gz(tar_path)
        self.assertTrue(out and os.path.isfile(os.path.join(out, "f.log")))

    def test_decompress_tar_gz_not_exist(self):
        self.assertEqual(tool.decompress_tar_gz(os.path.join(self.temp_dir, "no.tar.gz")), "")

    def test_decompress_tar_gz_wrong_suffix(self):
        path = os.path.join(self.temp_dir, "c.txt")
        with open(path, "w", encoding="utf-8") as f:
            f.write("x")
        self.assertEqual(tool.decompress_tar_gz(path), "")

    def test_decompress_tar_gz_path_traversal(self):
        tar_path = self._make_tar_gz("evil.tar.gz", {"../escape.log": b"bad"})
        out = tool.decompress_tar_gz(tar_path)
        self.assertEqual(out, "")
        self.assertFalse(os.path.exists(os.path.join(self.temp_dir, "escape.log")))


if __name__ == '__main__':
    unittest.main()
