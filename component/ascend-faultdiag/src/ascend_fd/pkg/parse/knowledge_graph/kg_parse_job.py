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

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.parse_info import KGParseFilePath
from ascend_fd.utils.status import FileNotExistError
from ascend_fd.utils.tool import check_file_num_and_size
from ascend_fd.model.cfg import ParseCFG
from ascend_fd.pkg.parse.knowledge_graph.utils import SingleJsonFileProcessing
from ascend_fd.utils.regular_table import KG_MIN_TIME

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")
echo = logging.getLogger("ECHO")


def start_kg_parse_job(cfg: ParseCFG):
    """
    Execute the knowledge graph parsing task and invoke the knowledge graph parsing code
    :param cfg: parse config
    """
    output_path = cfg.output_path
    worker = SingleJsonFileProcessing(get_parse_ctx(cfg))
    kg_logger.info("Start knowledge graph parse job.")
    worker.export_json_file(output_path, cfg.task_id)
    worker.export_mindie_cluster_info(output_path)


def get_single_parse_data(cfg: ParseCFG):
    """
    Execute the knowledge graph parsing task and invoke the knowledge graph parsing code
    :param cfg: parse config
    """
    parse_ctx = get_parse_ctx(cfg)
    worker = SingleJsonFileProcessing(parse_ctx)
    kg_logger.info("Start knowledge graph parse job.")
    return worker.get_parse_data(cfg.task_id)


def get_parse_ctx(cfg: ParseCFG) -> KGParseCtx:
    """
    Generate context for kg parse
    :param cfg: parse config
    :return: kg parse context
    """
    collector = SaverCollector(cfg)
    if not cfg.is_sdk_input:
        check_file_num_and_size(collector.parse_file_path.get_all_path(), kg_logger)
    collector.validate_all_emtpy()
    return collector.parse_ctx


class SaverCollector:
    def __init__(self, cfg):
        self.cfg = cfg
        self.parse_file_path = KGParseFilePath(
            plog_path=self.safe_get("log_saver", "get_plog_dict", default={}),
            device_log_path=self.safe_get("log_saver", "get_device_log_dict", default={}),
            npu_info_path=self.safe_get("env_info_saver", "get_npu_info_list", default=[]),
            train_log_path=self.safe_get("train_log_saver", "get_train_log", default=[]),
            host_log_path=self.safe_get("host_log_saver", "get_host_log", default=[]),
            host_dmesg_path=self.safe_get("host_log_saver", "get_dmesg_log", default=[]),
            host_sysmon_path=self.safe_get("host_log_saver", "get_sysmon_log", default=[]),
            host_vmcore_dmesg_path=self.safe_get("host_log_saver", "get_vmcore_dmesg_log", default=[]),
            hisi_logs_path=self.safe_get("dev_log_saver", "get_hisi_logs_list", default=[]),
            slog_path=self.safe_get("dev_log_saver", "get_slog_dict", default={}),
            noded_log_path=self.safe_get("dl_log_saver", "get_noded_list", default=[]),
            device_plugin_path=self.safe_get("dl_log_saver", "get_device_plugin_list", default=[]),
            volcano_scheduler_path=self.safe_get("dl_log_saver", "get_volcano_scheduler_list", default=[]),
            volcano_controller_path=self.safe_get("dl_log_saver", "get_volcano_controller_list", default=[]),
            mindio_log_path=self.safe_get("dl_log_saver", "get_mindio_log_list", default=[]),
            docker_runtime_path=self.safe_get("dl_log_saver", "get_docker_runtime_list", default=[]),
            npu_exporter_path=self.safe_get("dl_log_saver", "get_npu_exporter_list", default=[]),
            amct_path=self.safe_get("amct_log_saver", "get_amct_log", default=[]),
            mindie_log_path=self.safe_get("mindie_log_saver", "get_mindie_log_list", default=[]),
            mindie_cluster_log_path=self.safe_get("mindie_log_saver", "get_mindie_clu_log_list", default=[]),
            bmc_app_dump_log_path=self.safe_get("bmc_log_saver", "get_bmc_app_dump_log_list", default=[]),
            bmc_device_dump_log_path=self.safe_get("bmc_log_saver", "get_bmc_device_dump_log_list", default=[]),
            bmc_log_dump_log_path=self.safe_get("bmc_log_saver", "get_bmc_log_dump_log_list", default=[]),
            bmc_log_path=self.safe_get("bmc_log_saver", "get_bmc_log_list", default=[]),
            lcne_log_path=self.safe_get("lcne_log_saver", "get_lcne_log_list", default=[]),
            bus_log_path=self.safe_get("lcne_log_saver", "get_bus_log_dict", default={}),
            custom_log_list=self.safe_get("custom_log_saver", "get_custom_log_list", default=[])
        )
        # sdk断点续训时间属性需要在self.plog_dict后。（先要plog解析时间）
        self.parse_ctx = KGParseCtx(
            parse_file_path=self.parse_file_path,
            resuming_training_time=self.safe_get("log_saver", "get_resuming_training_time", default=KG_MIN_TIME),
            is_sdk_input=cfg.is_sdk_input,
            custom_info_list=self.safe_get("custom_log_saver", "get_custom_info_list", default=[]))

    def safe_get(self, attr_name: str, method_name: str, default=None):
        sub_obj = getattr(self.cfg, attr_name, None)
        if sub_obj is None:
            return default

        method = getattr(sub_obj, method_name, None)
        if method is None or not callable(method):
            return default

        return method()

    def validate_all_emtpy(self):
        if not self.parse_file_path.get_all_path() and not self.parse_ctx.custom_info_list:
            kg_logger.error("No log that meets the specifications is found.")
            raise FileNotExistError("No log that meets the specifications is found.")
