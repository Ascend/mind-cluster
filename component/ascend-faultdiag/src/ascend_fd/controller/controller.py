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
import logging
from typing import Dict, List, Type

from ascend_fd.controller.job_worker import generate_parse_job, generate_diag_job
from ascend_fd.pkg.diag.root_cluster import start_rc_diag_job
from ascend_fd.pkg.diag.root_cluster.mindie_diag_job import MindIEDiagWorker
from ascend_fd.utils.constant.str_const import SUPER_POD_SCENE
from ascend_fd.wrapper import PrintWrapper, JsonWrapper
from ascend_fd.utils.status import InnerError, PathError, ParamError, BaseError
from ascend_fd.utils.tool import (
    safe_write_open,
    safe_walk,
    get_version,
    get_build_time,
    MultiProcessJob,
    SHOW_IP_MAX,
    DOUBLE_SEP,
    SHOE_INFER_GROUP_MAX,
)
from ascend_fd.pkg.parse.parser_saver import (
    ParsedDataSaver,
    SaverFactory,
    BaseLogSaver,
    HostLogSaver,
    TrainLogSaver,
    CustomLogSaver,
)
from ascend_fd.model.cfg import DiagCFG, ParseCFG
from ascend_fd.pkg.parse.knowledge_graph.kg_parse_job import get_single_parse_data
from ascend_fd.pkg.diag.knowledge_graph.kg_diag_job import single_diag_job
from ascend_fd.utils.i18n import get_label_for_language

logger = logging.getLogger("FAULT_DIAG")
echo = logging.getLogger("ECHO")
lb = get_label_for_language()


class ParseController:
    """
    The parse job controller.
    The input_path needs to be specified to the job directory.
    """

    INPUT_DIR_DEPTH = 10

    def __init__(self, args):
        """
        Parse Controller
        :param args: command-line interface args. (contain: input_path, output_path)
        """
        logger.info("Start the log-parse job.")
        self._check_input_cmd(args)
        self.input_path = args.input_path
        self.output_path = self._check_output_path_data(args)
        self.performance_flag = False if args.cmd == "single-diag" else args.performance
        self.cfg = self.init_cfg(args)
        self.origin_results = []

    @staticmethod
    def _check_input_cmd(args):
        """
        Check inpt cmd when parsing the file
        :param args: the args
        """
        skip_args = ["output_path", "performance", "task_id", "cmd"]
        args_to_validate = [value for key, value in vars(args).items() if key not in skip_args]
        if not any(args_to_validate):
            logger.error("All input path parameters are empty.")
            raise ParamError("All input path parameters are empty.")

    @staticmethod
    def _check_output_path_data(args):
        """
        Check parse output path
        :param args: the args
        :return: whether the parse output dir is empty
        """
        output_path = args.output_path
        if args.cmd == "single-diag":
            return output_path
        if os.listdir(output_path):
            logger.error("The output path already contains data file, it should be empty.")
            raise PathError("The output path already contains data file, it should be empty.")
        return output_path

    @staticmethod
    def _find_paths_by_sub_cmd(args) -> Dict[Type[BaseLogSaver], str]:
        """
        Find log paths based on command line arguments for different saver types.
        """
        founded_path = dict()
        for saver in SaverFactory.list_savers_classes():
            saver_log_path = ParseController._get_saver_log_path(args, saver.CMD_ARG_KEYS)
            if not saver_log_path and saver in [TrainLogSaver, HostLogSaver] and args.input_path is not None:
                founded_path[saver] = args.input_path
                continue
            if saver_log_path:
                founded_path[saver] = saver_log_path
        return founded_path

    @staticmethod
    def _get_saver_log_path(args, cmd_arg_keys: List[str]) -> str:
        """
        Get the log path from command line arguments based on the provided argument keys.
        """
        for cmd in cmd_arg_keys:
            saver_log_path = getattr(args, cmd, "")
            if saver_log_path:
                return saver_log_path
        return ""

    @staticmethod
    def _build_saver_directory_mapping(found_paths: Dict[Type[BaseLogSaver], str]) -> Dict[str, Type[BaseLogSaver]]:
        """
        Create mapping from target directory names to saver classes that haven't been matched yet.
        """
        target_dir_to_saver_map = dict()
        for saver in SaverFactory.list_savers_classes():
            if saver.CENTRALIZED_STORAGE_DIRECTORY is not None and saver not in found_paths:
                target_dir_to_saver_map[saver.CENTRALIZED_STORAGE_DIRECTORY] = saver
        return target_dir_to_saver_map

    def init_cfg(self, args):
        """
        Init parse config
        :param args: command args
        :return: ParseCFG
        """
        collected_paths = self._deep_find_input_path(args)

        savers_for_current_task = []
        for saver_class in SaverFactory.list_savers_classes():
            log_path = collected_paths.get(saver_class, "")

            if log_path:
                saver = SaverFactory.create_saver(saver_class.__name__)
                saver.filter_log(log_path)
                savers_for_current_task.append(saver)
                logger.info("Obtain the %s from the %s folder.", saver.LOG_TYPE, log_path)

        return ParseCFG.cmd_config(args, saver_list=savers_for_current_task)

    def start_job(self):
        """
        Use multiprocessing to start parse tasks
        """
        logger.info("Component Version: %s. Build time: %s", get_version(), get_build_time())
        multiprocess_job = MultiProcessJob(
            "FAULT_DIAG", pool_size=4, task_id=self.cfg.task_id, daemon=False, failed_raise=False
        )
        parse_jobs = generate_parse_job(self.performance_flag)
        for job_name, job_func in parse_jobs.items():
            multiprocess_job.add_security_job(job_name, job_func, self.cfg)
        _, failed_details = multiprocess_job.join_and_get_results()
        success_job = list(set(parse_jobs.keys()) - set(failed_details.keys()))
        if success_job:
            echo.info("These job %s succeeded.", success_job)
        for job_name, error_info in failed_details.items():
            echo.warning("The job %s failed. The error is: [%s].", job_name, error_info)
        logger.info("The log-parse job is complete.")
        if len(failed_details) == len(parse_jobs):
            logger.error("All parse subjobs failed.")
            raise InnerError("All parse subjobs failed.")

    def start_single_parse(self):
        """
        start single diag task
        """
        return get_single_parse_data(self.cfg)

    def _deep_find_input_path(self, args) -> Dict[Type[BaseLogSaver], str]:
        """
        Find various log dir name based on folder-traversal

        :param args: command args
        :return: a dict of various log paths, keys: saver(child of BaseLogSaver), values: path
            example:
            {
                ProcessLogSaver: "",
                EnvInfoSaver: "",
                TrainLogSaver: "",
                HostLogSaver: "",
                ...
            }
        """
        found_paths = self._find_paths_by_sub_cmd(args)
        if not args.input_path or not os.path.isdir(args.input_path):
            # 自定义清洗日志只支持--custom_log，不支持-i
            custom_log = getattr(args, "custom_log", "")
            if custom_log:
                found_paths.update({CustomLogSaver: custom_log})
            return found_paths
        # those savers still not have a log path are the target,
        # and the target dirs are their CENTRALIZED_STORAGE_DIRECTORY
        target_dir_to_saver_map = self._build_saver_directory_mapping(found_paths)
        for root, dirs, _ in safe_walk(args.input_path, self.INPUT_DIR_DEPTH):
            for target in set(target_dir_to_saver_map.keys()) & set(dirs):
                corresponding_saver = target_dir_to_saver_map[target]
                if corresponding_saver not in found_paths:
                    found_paths[corresponding_saver] = os.path.join(root, target)
        return found_paths


class DiagController:
    """
    The diag job controller
    """

    OUT_DIR = "fault_diag_result"
    NORMAL_RC_CODE = 102

    def __init__(self, args):
        """
        Parse Controller
        :param args: command args. (contain: input_path, output_path, mode, task, is_print)
        """
        logger.info("Start the fault-diag job.")
        self.cfg = self.init_cfg(args)
        self.input_path = self.cfg.input_path
        self.output_path = self.cfg.output_path
        os.makedirs(self.output_path, 0o700, exist_ok=True)
        self.performance_flag = False if args.cmd == "single-diag" else args.performance
        self.single_diag_flag = bool(args.cmd == "single-diag")
        self.origin_results = dict()
        self.failed_details = dict()

    def init_cfg(self, args):
        """
        Init diag config. The config contains: mode, input_path, output_path, parsed data saver
        :param args: command args
        :return: DiagCFG
        """
        input_path = args.input_path
        output_path = os.path.join(args.output_path, self.OUT_DIR)

        parsed_saver = ParsedDataSaver(input_path, args)
        return DiagCFG(args.task_id, input_path, output_path, parsed_saver)

    def start_job(self):
        """
        Use multiprocessing to start diag tasks
        """
        # 训练场景诊断
        if not self.cfg.parsed_saver.infer_task_flag:
            results = self.start_train_task()
            self._export_results(results)
            return
        # 推理场景诊断
        # 推理场景诊断-MindIE相关诊断
        MindIEDiagWorker(self.cfg).start_job()
        # 推理场景诊断-plog相关诊断
        count = 1
        echo_show_flag = True
        for infer_group_name in self.cfg.parsed_saver.collect_infer_group:
            if count > SHOE_INFER_GROUP_MAX:
                echo_show_flag = False
            count += 1
            self.cfg.parsed_saver.infer_instance = infer_group_name
            results = self.start_train_task()
            self._echo_server_info(infer_group_name, echo_show_flag)
            self._export_results(results, f"diag_report_{infer_group_name}.json", echo_show_flag)
            self.origin_results = dict()
            self.failed_details = dict()

    def start_train_task(self):
        """
        Use multiprocessing to start train diag tasks
        """
        logger.info("Component Version: %s. Build time: %s", get_version(), get_build_time())
        self._exec_root_cluster_job()  # execute root cluster diag job first
        multiprocess_job = MultiProcessJob(
            "FAULT_DIAG", pool_size=3, task_id=self.cfg.task_id, daemon=False, failed_raise=False
        )
        diag_jobs = generate_diag_job(self.performance_flag)
        for job_name, job_func in diag_jobs.items():
            multiprocess_job.add_security_job(job_name, job_func, self.cfg)
        results, failed_details = multiprocess_job.join_and_get_results()
        self.failed_details.update(failed_details)
        logger.info("The fault-diag job is complete.")
        if len(failed_details) == len(diag_jobs):
            logger.error("All diag subjobs failed.")
            raise InnerError("All diag subjobs failed.")
        return results

    def start_single_diag_job(self, parsed_data):
        """
        Start single diag task
        :param parsed_data: parsed data
        """
        if not parsed_data:
            return
        results = {"KNOWLEDGE_GRAPH": {"Kg": single_diag_job(parsed_data, self.cfg)}}
        self._export_results(results)

    def _exec_root_cluster_job(self):
        """
        Diag job first execute root cluster job to check:
        1. check whether the training task is faulty;
        2. If fault occurs, which cluster is the root cause node;
        """
        try:
            result = start_rc_diag_job(self.cfg)
        except BaseError as err:
            err_msg = f"Root Cluster diag job failed. The reason is: {err}"
            logger.error(err_msg)
            self.failed_details.update({"ROOT_CLUSTER": err_msg})
            raise err
        self.origin_results.update({"Rc": result.to_dict()})
        if not result.detect_workers_devices:
            logger.error("The list of workers to be checked is empty, please check the root cluster diag result.")
            raise InnerError("The list of workers to be checked is empty. Please check the root cluster diag result.")
        self.cfg.root_worker_devices = result.detect_workers_devices
        self.cfg.fault_filter_time = result.fault_filter_time
        fault_description = result.fault_description
        if fault_description:
            return
        raise InnerError("Root Cluster diag job failed. Can't get the result.")

    def _echo_server_info(self, infer_group_name, echo_show_flag):
        """
        Display service information on the screen
        :param infer_group_name: the name of infer group
        """
        if not echo_show_flag:
            return
        container_ip_list = self.cfg.parsed_saver.cluster_info.get(infer_group_name, [])
        show_ip_list = container_ip_list[:SHOW_IP_MAX]
        if len(container_ip_list) > SHOW_IP_MAX:
            show_ip_list.append("...")
        echo.info("\n%s", DOUBLE_SEP)
        echo.info("%s：%s", lb.instance_name, infer_group_name)
        echo.info("%s：%s", lb.node_name, show_ip_list)

    def _export_results(self, results: Dict[str, dict], out_file_name="diag_report.json", echo_show_flag=True):
        """
        Sort the diagnostic results and save results to output path.
        If print parameter is true, func will print the results
        :param results: the diag result for all job
        :param out_file_name: the name of output file
        """
        for _, job_result in results.items():
            # don't need the key(job name), job_result is : {Kg/Node/Net: fault_detail_dict}
            self.origin_results.update(job_result)
        out_file = os.path.join(self.output_path, out_file_name)
        format_table = PrintWrapper(
            self.origin_results, self.failed_details, self.performance_flag, self.single_diag_flag
        ).get_format_table()
        if echo_show_flag:
            echo.info(format_table)
        json_wrapper = JsonWrapper(
            self.origin_results, self.failed_details, self.performance_flag, self.cfg.task_id, self.single_diag_flag
        )
        json_wrapper.format_json()
        json_file = json_wrapper.get_format_json()
        with safe_write_open(out_file, mode="w+", encoding="utf-8") as file_stream:
            file_stream.write(json_file)
            file_stream.write('\r\n')

        if self.cfg.parsed_saver.scene == SUPER_POD_SCENE:
            self.cfg.parsed_saver.super_pod_info_saver.save_to_json(self.output_path, "topo_info.json")


class SingleDiagController:
    """
    The single diag job controller
    """

    def __init__(self, args):
        """
        Single-diag Controller
        :param args: command args. (contain: input_path, output_path, mode, task, is_print)
        """
        logger.info("Start the single-diag job.")
        self.parse_controller = ParseController(args)
        self.diag_controller = DiagController(args)

    def start_job(self):
        """
        Start single diag task
        """
        parsed_data = self.parse_controller.start_single_parse()
        self.diag_controller.start_single_diag_job(parsed_data)
