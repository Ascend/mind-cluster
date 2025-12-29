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
import logging

from ascend_fd.model.context import KGParseCtx
from ascend_fd.model.node_info import MindieClusterInfo
from ascend_fd.pkg.parse.knowledge_graph.utils.data_descriptor import DataDescriptor
from ascend_fd.utils.status import FileNotExistError
from ascend_fd.pkg.parse.knowledge_graph.utils.package_parser import PackageParser
from ascend_fd.utils.tool import safe_read_line
from ascend_fd.utils.regular_table import SERVER_INFO_FILE
from ascend_fd.configuration.config import KG_PASER_DUMP_NAME, KG_ANALYZER_DUMP_NAME

kg_logger = logging.getLogger("KNOWLEDGE_GRAPH")


def find_mindie_cluster_info(log_line, mindie_cluster_info_map):
    """
    Find mindie cluster info
    :param log_line: log line
    :param mindie_cluster_info_map: mindie_cluster_info_map
    """
    if not log_line:
        return
    json_data = None
    try:
        json_data = json.loads(log_line)
    except json.JSONDecodeError as error:
        kg_logger.info("The mindie cluster info loads failed: %s, log detail: %s.", error, log_line)
    if json_data:
        mindie_cluster_info_map.update(MindieClusterInfo.from_dict(json_data).trans_to_map())


class SingleJsonFileProcessing:
    MINDIE_CLUSTER_INFO = "mindie-cluster-info.json"

    def __init__(self, kg_parse_ctx: KGParseCtx):
        """
        Single json file process class
        :param kg_parse_ctx: kg parse ctx
        """
        self.kg_parse_ctx = kg_parse_ctx

    def export_json_file(self, result_path, task_id):
        """
        Export json file
        :param result_path: the path to export json result
        :param task_id: the task unique id
        """
        if not os.path.isdir(result_path):
            kg_logger.error("Result path %s not found.", os.path.basename(result_path))
            raise FileNotExistError(f"Result path {os.path.basename(result_path)} not found.")
        package_parser = PackageParser(self.kg_parse_ctx)
        package_parser.parse(task_id)

        json_path = os.path.join(result_path, KG_PASER_DUMP_NAME)
        package_parser.desc.dump_to_json_file(json_path)
        kg_logger.info("The kg parsing result is saved in dir %s.", os.path.basename(result_path))

        output_path = os.path.join(result_path, SERVER_INFO_FILE)
        package_parser.desc.export_server_info_file(output_path)

        file_path = os.path.join(result_path, KG_ANALYZER_DUMP_NAME)
        package_parser.desc.single_worker_fault_analysis(file_path)
        kg_logger.info("The results of the early kg analysis are saved in %s.", os.path.basename(file_path))

    def get_parse_data(self, task_id):
        """
        Get parse data
        :param task_id: the task unique id
        """
        package_parser = PackageParser(self.kg_parse_ctx)
        package_parser.parse(task_id)
        return package_parser.desc.get_single_worker_fault_analysis()

    def export_mindie_cluster_info(self, result_path):
        """
        Export mindie cluster info
        :param result_path: the path to export json result
        """
        mindie_cluster_info_map = {}
        for mindie_cluster_log_file in self.kg_parse_ctx.parse_file_path.mindie_cluster_log_path:
            temp_log_line = None
            for log_line in safe_read_line(mindie_cluster_log_file):
                if log_line.startswith("{") and "completed" in log_line and "server_list" in log_line:
                    temp_log_line = log_line
                    break
            find_mindie_cluster_info(temp_log_line, mindie_cluster_info_map)
        if mindie_cluster_info_map:
            json_path = os.path.join(result_path, self.MINDIE_CLUSTER_INFO)
            DataDescriptor.write_to_json_file(json_path, mindie_cluster_info_map)
