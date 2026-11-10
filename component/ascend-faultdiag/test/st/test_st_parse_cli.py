#!/usr/bin/env python3
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
import argparse
import errno
import json
import os
import shutil
import sys
import tempfile
import unittest
from contextlib import redirect_stdout
from io import StringIO
from unittest.mock import patch

from ascend_fd.cli import command_line
from ascend_fd.controller.controller import ParseController
from ascend_fd.controller.job_worker import start_kg_parse_job, start_rc_parse_job
from ascend_fd.pkg.parse.parser_saver import (
    DevLogSaver,
    EnvInfoSaver,
    HostLogSaver,
    LCNELogSaver,
    ProcessLogSaver,
    TrainLogSaver,
)
from ascend_fd.utils.status import FileNotExistError, FileOpenError
from ascend_fd.utils.tool import dir_check, file_or_dir_check, safe_write_open, write_path_check

TEST_DIR = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
TESTCASE_INPUT = os.path.join(TEST_DIR, "st_testcase")


class ParseSTArgs:
    cmd = "parse"

    def __init__(self, input_path, output_path, **kwargs):
        self.input_path = input_path
        self.output_path = output_path
        self.performance = False
        self.task_id = "st_parse_cli"
        self.host_log = kwargs.get("host_log")
        self.train_log = kwargs.get("train_log")
        self.env_check = kwargs.get("env_check")
        self.device_log = kwargs.get("device_log")
        self.process_log = kwargs.get("process_log")
        self.dl_log = kwargs.get("dl_log")
        self.amct_log = kwargs.get("amct_log")
        self.mindie_log = kwargs.get("mindie_log")
        self.bmc_log = kwargs.get("bmc_log")
        self.lcne_log = kwargs.get("lcne_log")
        self.bus_log = kwargs.get("bus_log")
        self.custom_log = kwargs.get("custom_log")
        self.pymotor_vllm_log = kwargs.get("pymotor_vllm_log")


class STParseCli(unittest.TestCase):
    def setUp(self):
        self.base = tempfile.mkdtemp()
        self.input_dir = os.path.join(self.base, "input")
        self.output_dir = os.path.join(self.base, "output")
        os.makedirs(self.input_dir)
        os.makedirs(self.output_dir)

    def tearDown(self):
        shutil.rmtree(self.base, ignore_errors=True)

    def make_files(self, relative_dir, names, content="test log"):
        target = os.path.join(self.base, relative_dir)
        os.makedirs(target, exist_ok=True)
        for name in names:
            with open(os.path.join(target, name), "w", encoding="utf-8") as file_stream:
                file_stream.write(content)
        return target

    def make_args(self, **kwargs):
        return ParseSTArgs(kwargs.pop("input_path", self.input_dir), self.output_dir, **kwargs)

    def assert_invalid_path_rejected(self, check_func, name):
        """Create a real directory with invalid characters so only the character check can reject it."""
        invalid_path = os.path.join(self.base, name)
        os.makedirs(invalid_path, exist_ok=True)
        with self.assertRaises(argparse.ArgumentTypeError):
            check_func(invalid_path)

    def test_parse_0001_input_without_required_log(self):
        """Ascend_Fd_Parse_0001"""
        unknown_dir = self.make_files("unknown_input", ["unknown.dat"])
        args = ParseSTArgs(unknown_dir, self.output_dir)
        controller = ParseController(args)
        with self.assertRaises(FileNotExistError):
            start_kg_parse_job(controller.cfg)

    def test_parse_0002_input_path_is_symlink(self):
        """Ascend_Fd_Parse_0002"""
        link_path = os.path.join(self.base, "input_link")
        os.symlink(self.input_dir, link_path)
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check(link_path)

    def test_parse_0005_input_path_invalid_value(self):
        """Ascend_Fd_Parse_0005"""
        self.assert_invalid_path_rejected(dir_check, "invalid:path")

    def test_parse_0012_parsed_directory_and_content(self):
        """Ascend_Fd_Parse_0012"""
        args = ParseSTArgs(TESTCASE_INPUT, self.output_dir)
        controller = ParseController(args)
        start_rc_parse_job(controller.cfg)
        start_kg_parse_job(controller.cfg)
        output_files = os.listdir(self.output_dir)
        self.assertIn("ascend-kg-parser.json", output_files)
        self.assertIn("ascend-kg-analyzer.json", output_files)
        self.assertIn("ascend-rc-parser.json", output_files)
        with open(os.path.join(self.output_dir, "ascend-kg-parser.json"), encoding="utf-8") as file_stream:
            self.assertTrue(json.load(file_stream))

    def test_parse_0015_log_file_is_symlink(self):
        """Ascend_Fd_Parse_0015"""
        target = os.path.join(self.base, "target.log")
        with open(target, "w", encoding="utf-8") as file_stream:
            file_stream.write("origin")
        link_path = os.path.join(self.base, "fd_op.log")
        os.symlink(target, link_path)
        with self.assertRaises(FileOpenError):
            safe_write_open(link_path, mode="w+")

    def test_parse_0011_output_disk_no_space(self):
        """Ascend_Fd_Parse_0011"""
        with (
            patch("os.open", side_effect=OSError(errno.ENOSPC, "No space left on device")),
            self.assertRaises(OSError) as context,
        ):
            safe_write_open(os.path.join(self.output_dir, "result.json"), mode="w+")
        self.assertEqual(context.exception.errno, errno.ENOSPC)

    def test_parse_0010_output_path_not_exist(self):
        """Ascend_Fd_Parse_0010"""
        with self.assertRaises(argparse.ArgumentTypeError):
            write_path_check(os.path.join(self.base, "not_exist_output"))

    def test_parse_0009_output_path_invalid_value(self):
        """Ascend_Fd_Parse_0009"""
        self.assert_invalid_path_rejected(write_path_check, "output:invalid")

    def test_para_check_0013_device_log_path_name(self):
        """Ascend_Fd_Para_Check_0013"""
        valid_dir = self.make_files("device_log", ["device_info.txt"])
        self.assertEqual(dir_check(valid_dir), valid_dir)
        self.assert_invalid_path_rejected(dir_check, "device|log")

    def test_para_check_0024_device_log_path_not_exist(self):
        """Ascend_Fd_Para_Check_0024"""
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check(os.path.join(self.base, "no_such_device_log"))

    def test_para_check_0003_device_log_with_input_abnormal(self):
        """Ascend_Fd_Para_Check_0003"""
        file_path = os.path.join(self.make_files("device_log", ["device_info.txt"]), "device_info.txt")
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check(file_path)

    def test_para_check_0019_device_log_with_input_normal(self):
        """Ascend_Fd_Para_Check_0019"""
        device_dir = self.make_files("device_log", ["device_info.txt"])
        args = self.make_args(device_log=device_dir)
        found = ParseController._find_paths_by_sub_cmd(args)
        self.assertIn(device_dir, found.values())

    def test_para_check_0012_env_check_path_name(self):
        """Ascend_Fd_Para_Check_0012"""
        valid_dir = self.make_files("env_check", ["npu_info_before.txt"])
        self.assertEqual(dir_check(valid_dir), valid_dir)
        self.assert_invalid_path_rejected(dir_check, "env|check")

    def test_para_check_0023_env_check_path_not_exist(self):
        """Ascend_Fd_Para_Check_0023"""
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check(os.path.join(self.base, "no_such_env_check"))

    def test_para_check_0002_env_check_with_input_abnormal(self):
        """Ascend_Fd_Para_Check_0002"""
        file_path = os.path.join(self.make_files("env_check", ["npu_info_before.txt"]), "npu_info_before.txt")
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check(file_path)

    def test_para_check_0018_env_check_with_input_normal(self):
        """Ascend_Fd_Para_Check_0018"""
        env_dir = self.make_files("env_check", ["npu_info_before.txt"])
        args = self.make_args(env_check=env_dir)
        found = ParseController._find_paths_by_sub_cmd(args)
        self.assertIn(env_dir, found.values())

    def test_para_check_0028_env_check_data_mismatch(self):
        """Ascend_Fd_Para_Check_0028"""
        mismatch_dir = self.make_files("env_check_mismatch", ["unknown.txt", "other.log"])
        saver = EnvInfoSaver()
        saver.filter_log(mismatch_dir)
        self.assertEqual(saver.get_npu_info_list(), [])
        self.assertEqual(saver.get_npu_smi_detail_list(), [])
        self.assertEqual(saver.get_npu_detail_list(), [])
        self.assertEqual(saver.get_host_metrics_path(), "")

    def test_para_check_0015_host_log_path_name(self):
        """Ascend_Fd_Para_Check_0015"""
        valid_dir = self.make_files("host_log", ["messages"])
        self.assertEqual(dir_check(valid_dir), valid_dir)
        self.assert_invalid_path_rejected(dir_check, "host|log")

    def test_para_check_0005_host_log_with_input_abnormal(self):
        """Ascend_Fd_Para_Check_0005"""
        file_path = os.path.join(self.make_files("host_log", ["messages"]), "messages")
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check(file_path)

    def test_para_check_0011_process_log_path_name(self):
        """Ascend_Fd_Para_Check_0011"""
        valid_dir = self.make_files("process_log", ["plog-1_20240101000000.log"])
        self.assertEqual(dir_check(valid_dir), valid_dir)
        self.assert_invalid_path_rejected(dir_check, "process|log")

    def test_para_check_0022_process_log_path_not_exist(self):
        """Ascend_Fd_Para_Check_0022"""
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check(os.path.join(self.base, "no_such_process_log"))

    def test_para_check_0001_process_log_with_input_abnormal(self):
        """Ascend_Fd_Para_Check_0001"""
        log_name = "plog-1_20240101000000.log"
        file_path = os.path.join(self.make_files("process_log", [log_name]), log_name)
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check(file_path)

    def test_para_check_0017_process_log_with_input_normal(self):
        """Ascend_Fd_Para_Check_0017"""
        process_dir = self.make_files("process_log", ["plog-1_20240101000000.log"])
        args = self.make_args(process_log=process_dir)
        found = ParseController._find_paths_by_sub_cmd(args)
        self.assertIn(process_dir, found.values())

    def test_para_check_0014_train_log_path_name(self):
        """Ascend_Fd_Para_Check_0014"""
        valid_dir = self.make_files("train_log", ["rank-0.log"])
        self.assertEqual(file_or_dir_check(valid_dir), valid_dir)
        self.assert_invalid_path_rejected(file_or_dir_check, "train|log")

    def test_para_check_0025_train_log_path_not_exist(self):
        """Ascend_Fd_Para_Check_0025"""
        with self.assertRaises(argparse.ArgumentTypeError):
            file_or_dir_check(os.path.join(self.base, "no_such_train_log"))

    def test_para_check_0004_train_log_with_input_abnormal(self):
        """Ascend_Fd_Para_Check_0004"""
        with self.assertRaises(argparse.ArgumentTypeError):
            file_or_dir_check(os.path.join(self.base, "no_such_train_file.log"))

    def test_para_check_0020_train_log_with_input_normal(self):
        """Ascend_Fd_Para_Check_0020"""
        train_dir = self.make_files("train_log", ["rank-0.log"])
        args = self.make_args(train_log=[train_dir])
        found = ParseController._find_paths_by_sub_cmd(args)
        self.assertIn([train_dir], list(found.values()))

    def test_no_force_match_directory_0006_messages_under_dev_os(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0006"""
        host_dir = self.make_files("host_log/dev-os-0", ["messages"])
        saver = HostLogSaver()
        saver.filter_log(host_dir)
        self.assertEqual(saver.get_host_log(), [])

    def test_no_force_match_directory_0008_messages_in_different_levels(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0008"""
        host_dir = self.base
        self.make_files("host/a", ["messages"])
        self.make_files("host/b/c", ["messages-1"])
        saver = HostLogSaver()
        saver.filter_log(host_dir)
        self.assertEqual(len(saver.get_host_log()), 2)

    def test_no_force_match_directory_0010_env_files_in_different_levels(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0010"""
        env_dir = self.base
        self.make_files("env/a", ["npu_info_before.txt"])
        self.make_files("env/b/c", ["npu_info_after.txt"])
        self.make_files("env/d/e", ["npu_smi_0_details.csv"])
        self.make_files("env/f/g/h", ["npu_0_details.csv"])
        self.make_files("env/i/j", ["host_metrics_64.json"])
        saver = EnvInfoSaver()
        saver.filter_log(env_dir)
        self.assertEqual(len(saver.get_npu_info_list()), 2)
        self.assertEqual(len(saver.get_npu_smi_detail_list()), 1)
        self.assertEqual(len(saver.get_npu_detail_list()), 1)
        self.assertTrue(saver.get_host_metrics_path().endswith("host_metrics_64.json"))

    def test_no_force_match_directory_0004_device_logs_in_different_levels(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0004"""
        device_dir = self.base
        self.make_files("device/slog/dev-os-0/debug/device-0", ["device-os_20240101000000.log"])
        self.make_files("device/hisi_logs/0/20240101000000-123456789/log", ["kernel.log"])
        dev_saver = DevLogSaver()
        dev_saver.filter_log(device_dir)
        self.assertTrue(dev_saver.get_slog_dict())
        self.assertTrue(dev_saver.get_hisi_logs_list())

    def test_no_force_match_directory_0001_process_logs_in_different_levels(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0001"""
        process_dir = self.base
        self.make_files("p/a/run/plog", ["plog-100_20240101000000.log"])
        self.make_files("p/b/debug/plog", ["plog-101_20240101000000.log"])
        self.make_files("p/c/device", ["device-102_20240101000000.log"])
        self.make_files("p/d/run/device", ["device-103_20240101000000.log"])
        saver = ProcessLogSaver()
        saver.filter_log(process_dir)
        self.assertEqual(len(saver.get_plog_dict()), 2)
        self.assertEqual(len(saver.get_device_log_dict()), 2)

    def test_no_force_match_directory_0002_train_logs_in_different_levels(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0002"""
        first = self.make_files("train/a", ["rank-0.log"])
        second = self.make_files("train/b/c", ["rank-1.log"])
        saver = TrainLogSaver()
        saver.filter_log([first, second])
        self.assertEqual(len(saver.get_train_log()), 2)

    def test_no_force_match_directory_0005_input_path_with_modified_structure(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0005"""
        input_root = self.base
        for target_dir in (
            "process_log",
            "device_log",
            "environment_check",
            "bmc_log",
            "lcne_log",
            "dl_log",
            "amct_log",
            "mindie",
        ):
            os.makedirs(os.path.join(input_root, target_dir, "nested"), exist_ok=True)
        args = ParseSTArgs(input_root, self.output_dir)
        found = ParseController(args)._deep_find_input_path(args)
        self.assertGreaterEqual(len(found), 8)

    def test_no_force_match_directory_0009_run_and_debug_max_9_files(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0009"""
        for index in range(12):
            self.make_files("process/run/plog", [f"plog-100_{20240101000000 + index}.log"])
            self.make_files("process/debug/plog", [f"plog-100_{20240102000000 + index}.log"])
            self.make_files("process/plog", [f"plog-100_{20240103000000 + index}.log"])
        saver = ProcessLogSaver()
        saver.filter_log(os.path.join(self.base, "process"))
        plog_dict = saver.get_plog_dict()
        self.assertEqual(sum(len(file_list) for file_list in plog_dict.values()), 9)

    def test_no_force_match_directory_0003_run_or_debug_max_6_files(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0003"""
        for index in range(12):
            self.make_files("process/run/plog", [f"plog-100_{20240101000000 + index}.log"])
            self.make_files("process/plog", [f"plog-100_{20240103000000 + index}.log"])
        saver = ProcessLogSaver()
        saver.filter_log(os.path.join(self.base, "process"))
        plog_dict = saver.get_plog_dict()
        self.assertEqual(sum(len(file_list) for file_list in plog_dict.values()), 6)

    def test_no_force_match_directory_0007_no_run_debug_max_3_files(self):
        """Ascend_Fd_Parse_No_Force_Match_Directory_0007"""
        for index in range(12):
            self.make_files("process/plog", [f"plog-100_{20240103000000 + index}.log"])
        saver = ProcessLogSaver()
        saver.filter_log(os.path.join(self.base, "process"))
        plog_dict = saver.get_plog_dict()
        self.assertEqual(sum(len(file_list) for file_list in plog_dict.values()), 3)

    def test_func_para_buslog_0002_help_contains_bus_log(self):
        """Ascend_Fd_Func_Para_Buslog_0002"""
        stdout = StringIO()
        with (
            patch.object(sys, "argv", ["ascend-fd", "parse", "-h"]),
            redirect_stdout(stdout),
            self.assertRaises(SystemExit),
        ):
            command_line()
        self.assertIn("--bus_log", stdout.getvalue())

    def test_func_para_buslog_0007_empty_dir_and_empty_value(self):
        """Ascend_Fd_Func_Para_Buslog_0007"""
        empty_dir = self.make_files("empty_bus_log", [])
        self.assertEqual(dir_check(empty_dir), empty_dir)
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check("")

    def test_func_para_buslog_0008_input_path_cleans_bus_log(self):
        """Ascend_Fd_Func_Para_Buslog_0008"""
        input_root = self.base
        self.make_files("input_root/lcne_log", ["log_1_1234567890123.log"])
        args = ParseSTArgs(input_root, self.output_dir)
        found = ParseController(args)._deep_find_input_path(args)
        self.assertIn(LCNELogSaver, found)
        saver = LCNELogSaver()
        saver.filter_log(found[LCNELogSaver])
        self.assertEqual(len(saver.get_lcne_log_list()), 1)

    def test_func_para_buslog_0003_bus_log_with_plog_failed(self):
        """Ascend_Fd_Func_Para_Buslog_0003"""
        plog_dir = self.make_files("plog_only", ["plog-1_20240101000000.log"])
        saver = LCNELogSaver()
        saver.filter_log(plog_dir)
        self.assertEqual(saver.get_lcne_log_list(), [])
        self.assertEqual(saver.get_bus_log_dict(), {})

    def test_func_para_buslog_0005_bus_log_symlink_failed(self):
        """Ascend_Fd_Func_Para_Buslog_0005"""
        target = self.make_files("bus_target", ["log_1_1234567890123.log"])
        link_path = os.path.join(self.base, "bus_link")
        os.symlink(target, link_path)
        with self.assertRaises(argparse.ArgumentTypeError):
            dir_check(link_path)

    def test_func_para_buslog_0001_bus_log_standalone(self):
        """Ascend_Fd_Func_Para_Buslog_0001"""
        bus_dir = self.make_files("bus_log", ["log_1_1234567890123.log"])
        saver = LCNELogSaver()
        saver.filter_log(bus_dir)
        self.assertEqual(len(saver.get_lcne_log_list()), 1)


if __name__ == "__main__":
    unittest.main()
