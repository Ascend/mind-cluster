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

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.model.response import Response
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.graph_builder import GraphBuilder
from ascend_fd.utils.load_kg_config import Schema

kg_logger = logging.getLogger("KG_ENGINE")


def start_analyze(schema_pkg_list, package_data):
    """
    Start engine analyze
    :param schema_pkg_list: kg-parser file path
    :param package_data: the data package path
    :return: interface response self
    """
    schema = Schema(schema_pkg_list)
    graph = GraphBuilder(schema, package_data).build_graph()
    return Response().get_information(graph)


def kg_engine_analyze(schema_pkg_list, package_data):
    """
    Inference engine main function
    :param schema_pkg_list: kg-parser file path
    :param package_data: the data package path
    :return: inference result
    """
    resp = Response()
    try:
        resp = start_analyze(schema_pkg_list, package_data)
    except Exception as error:
        kg_logger.error(str(error))
        resp.error = error
        resp.analyze_success = False
    return resp
