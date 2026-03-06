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
import logging
import multiprocessing
import os
import shutil
import argparse
import subprocess
import re

LOG_FORMAT = "%(message)s"
logging.basicConfig(level=logging.INFO, format=LOG_FORMAT)

PRINT_WIDTH = 109
DEVICE_NUM = 8


class LocalDiag:
    def __init__(self, input_path, output_path):
        """
        Local split data, parse, diag.
        :param input_path: modelarts log input Path,
        :param output_path: result output Path.
        """
        self.input_path = os.path.realpath(input_path)
        self.output_path = os.path.realpath(output_path)
        if os.path.exists(self.output_path) and os.listdir(self.output_path):
            raise Exception("The output path is not empty. Please specify the empty folder.")
        os.makedirs(self.output_path, 0o700, exist_ok=True)

        # create the parse output path in the output path.
        self.parse_output_path = os.path.join(self.output_path, "fault_diag_data")
        os.makedirs(self.parse_output_path, exist_ok=True)

        # create the split tmp dir in the output path.
        self.spilt_dir_path = os.path.join(self.output_path, "split_log_dir")
        os.makedirs(self.spilt_dir_path, 0o700, exist_ok=True)

        self.modelarts_dir_name = ""
        self.origin_ascend_path = ""
        self.split_job_dict = {}
        self.process_list = []

    @staticmethod
    def print_info(info):
        """
        Print log.
        """
        logging.info(info)

    @staticmethod
    def print_line(line):
        """
        Print format.
        """
        logging.info("\n")
        logging.info(line.center(PRINT_WIDTH, "*"))

    def get_split_path(self):
        """
        Get and create split path.
        :return: new modelarts_dir_name, new origin_ascend_path and new split_job_dict
        """
        for file_name in os.listdir(self.input_path):
            modelarts_path = os.path.join(self.input_path, file_name)
            if "modelarts-job-" not in file_name and "-worker-" not in file_name:
                continue
            # record the path of the modelarts job dir name and ascend dir path.
            if os.path.isdir(modelarts_path):
                self.modelarts_dir_name = file_name
                self.origin_ascend_path = os.path.join(modelarts_path, "ascend")
                continue
            worker_match = r"modelarts-job-[^,]{1,200}-worker-(\d{1,5}).log$"
            worker_re = re.match(worker_match, file_name)
            if not worker_re:
                continue
            worker_num = int(worker_re[1])
            # create split path and parse output path for each worker and record the paths.
            split_worker_path = os.path.join(self.spilt_dir_path, f"job_worker_{worker_num}")
            os.makedirs(split_worker_path, exist_ok=True)
            parse_output_path = os.path.join(self.parse_output_path, f"worker-{worker_num}")
            os.makedirs(parse_output_path, exist_ok=True)
            self.split_job_dict[worker_num] = {
                "split_path": split_worker_path,
                "parse_output_path": parse_output_path,
                "worker_log_name": file_name
            }

    def split_worker_log(self, worker_num, worker_name):
        """
        split modelarts-job-*-worker-*.log file every 1 worker.
        """
        origin_path = os.path.join(self.input_path, worker_name)
        split_path = os.path.join(self.spilt_dir_path, f"job_worker_{worker_num}", worker_name)
        shutil.copy(origin_path, split_path)

    def split_rank_txt(self, worker_num, split_path):
        """
        split modelarts-job-*-proc-rank-*-device-*.txt file every 8 device
        """
        for rank_num in range(DEVICE_NUM):
            split_rank_num = worker_num * DEVICE_NUM + rank_num
            rank_file_name = f"{self.modelarts_dir_name}-proc-rank-{split_rank_num}-device-{rank_num}.txt"
            origin_path = os.path.join(self.input_path, rank_file_name)
            spilt_path = os.path.join(split_path, rank_file_name)
            if os.path.exists(origin_path):
                shutil.copy(origin_path, spilt_path)

    def split_process_log(self, worker_num, split_path):
        """
        Split files in the modelarts-job-*/ascend/process_log/rank_* dir every 8 device.
        """
        rank_min = worker_num * DEVICE_NUM
        rank_max = worker_num * DEVICE_NUM + DEVICE_NUM
        for rank_num in range(rank_min, rank_max):
            rank_name = f"rank_{rank_num}"
            origin_path = os.path.join(self.origin_ascend_path, "process_log", rank_name)
            if not os.path.exists(origin_path):
                continue
            spilt_path = os.path.join(split_path, "process_log", rank_name)
            shutil.copytree(origin_path, spilt_path, True)

    def split_env_check(self, worker_num, split_path):
        """
        Split files in the modelarts-job-*/ascend/environment_check/worker-* dir every 1 worker.
        """
        worker_name = f"worker-{worker_num}"
        origin_path = os.path.join(self.origin_ascend_path, "environment_check", worker_name)
        if os.path.exists(origin_path):
            spilt_path = os.path.join(split_path, "environment_check", worker_name)
            shutil.copytree(origin_path, spilt_path, True)

    def split_and_parse_single_worker(self, worker_num, split_dict):
        """
        Split and parse single worker logs based on worker_num
        """
        split_path = split_dict.get("split_path", "")
        self.split_worker_log(worker_num, split_dict.get("worker_log_name", ""))
        self.split_rank_txt(worker_num, split_path)
        self.split_process_log(worker_num, split_path)
        self.split_env_check(worker_num, split_path)

        parse_output_path = split_dict.get("parse_output_path", "")
        parse_cmd = f"ascend-fd parse -i {split_path} -o {parse_output_path}"
        self.get_result("parse", parse_cmd)

    def split_and_parse_multi_worker(self):
        """
        Split and parse multi worker logs.
        """
        self.get_split_path()
        self.print_line("split log and parse start")
        for worker_num, split_dict in self.split_job_dict.items():
            process = multiprocessing.Process(target=self.split_and_parse_single_worker, name=worker_num,
                                              args=(worker_num, split_dict,))
            self.process_list.append(process)
            process.start()
        for p in self.process_list:
            p.join()
        # all parse worker is complete, delete the temporary split files.
        shutil.rmtree(os.path.join(self.spilt_dir_path))

    def diag_worker(self):
        """
        Diag files in the parse output_path/fault_diag_data.
        :return: diag result file.
        """
        self.print_line("diag start")
        diag_cmd = f"ascend-fd diag -i {self.parse_output_path} -o {self.output_path}"
        self.get_result("diag", diag_cmd)
        fault_diag_result_path = os.path.join(self.output_path, "fault_diag_result", "diag_report.json")
        self.print_info(f"please read the diag report from {fault_diag_result_path}")

    def get_result(self, tag, cmd):
        """
        Get and print the command execution result.
        """
        pipe = subprocess.Popen(cmd.split(), shell=False, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        res = pipe.stdout.read().decode("utf-8").strip()
        if res:
            self.print_info(f"{cmd}: {tag} success.\n{res}\n")
        else:
            self.print_info(f"{cmd}: {tag} failed.\n{res}\n")

    def split_go(self):
        """
        Start to perform the split, parse, and diag tasks and obtain results.
        """
        self.split_and_parse_multi_worker()
        self.diag_worker()
        self.print_line("Auto split diag Test Complete!")


def command_line():
    """
        The command line interface. Commands contain:
        -i, --input_path, the input path of modelarts log file,
        -o, --output_path, the input path of diag result file.
    """
    arg_cmd = argparse.ArgumentParser(add_help=True, description="Local run modelsArts log diag")
    arg_cmd.add_argument("-i", "--input_path", type=str, required=True, metavar='',
                         help="the input path of modelarts log file.")
    arg_cmd.add_argument("-o", "--output_path", type=str, required=True, metavar='',
                         help="the output path of diag result file.")
    return arg_cmd.parse_args()


if __name__ == "__main__":
    args = command_line()
    local_diag = LocalDiag(args.input_path, args.output_path)
    local_diag.split_go()
