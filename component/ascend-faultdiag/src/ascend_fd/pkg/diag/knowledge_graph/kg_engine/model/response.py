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
from typing import List, Dict

from ascend_fd.configuration.config import AICORE_ERRCODE_CONFIG
from ascend_fd.utils.fault_code import CANN_ERRCODE_CUSTOM, PYTORCH_ERRCODE_COMMON, RUNTIME_AICORE_EXECUTE_FAULT, \
    FAULT_WITH_COMPLEMENT_LIST, MINDIE_ERRCODE_COMMON
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.graph import Graph
from ascend_fd.pkg.diag.fault_entity import KG_DIAG_NORMAL_ENTITY
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.vertex import Vertex
from ascend_fd.utils.load_kg_config import EntityAttribute
from ascend_fd.utils.tool import safe_read_open, CONF_PATH
from ascend_fd.utils.i18n import get_label_for_language, LANG

lb = get_label_for_language()


class RootCause:
    COMPLEMENT = "complement"
    DESCRIPTION = f"description_{LANG}"
    ERROR_CODE = "error_code"
    KERNEL_NAME = "kernel_name"

    def __init__(self, code, entities_attribute: EntityAttribute, events_attribute: List[Dict] = None,
                 chains: dict = None):
        self.code = code
        self.entities_attribute = entities_attribute.to_json()
        self.events_attribute = events_attribute or []
        self.chains = chains or {}
        with safe_read_open(os.path.join(CONF_PATH, AICORE_ERRCODE_CONFIG), "r", encoding="UTF-8") as file_stream:
            self.aicore_errcode_map = json.load(file_stream)

    @staticmethod
    def _get_event_default_description(code: str) -> str:
        if code.startswith(PYTORCH_ERRCODE_COMMON):
            return lb.pytorch_custom_event_default_description
        if code.startswith(CANN_ERRCODE_CUSTOM):
            return lb.cann_custom_event_default_description
        if code.startswith(MINDIE_ERRCODE_COMMON):
            return lb.mindie_custom_event_default_description
        return ""

    def update_attribute_by_vertex(self, vertex: Vertex):
        """
        Update entities attribute by event vertex
        :param vertex: event vertex
        """
        if vertex.code != self.code:
            return
        if (vertex.code.startswith((PYTORCH_ERRCODE_COMMON, CANN_ERRCODE_CUSTOM, MINDIE_ERRCODE_COMMON))
                or vertex.code in FAULT_WITH_COMPLEMENT_LIST):
            self.entities_attribute[self.DESCRIPTION] = self._get_custom_event_info(
                vertex.event_attribute.get(self.COMPLEMENT, [])) or self._get_event_default_description(vertex.code)
        if vertex.code == RUNTIME_AICORE_EXECUTE_FAULT:
            self.entities_attribute[self.DESCRIPTION] = self._get_custom_event_info(
                self._get_aicore_fault_description_info(vertex))
        self.events_attribute.append(vertex.event_attribute)

    def _get_custom_event_info(self, complement: list) -> str:
        """
        Obtains the updated description of custom event
        :param complement: complementary param
        :return: description content
        """
        if not complement:
            return ""
        return self.entities_attribute.get(self.DESCRIPTION, "").format(*complement)

    def _get_aicore_fault_description_info(self, vertex):
        """
        Get the meaning of error code and fill them in chinese fault description
        """
        # if no errcode exists, display the default cause messages for ai core execute fault
        kernel_name = vertex.event_attribute.get(self.KERNEL_NAME, "")
        errcode = vertex.event_attribute.get(self.ERROR_CODE, "")
        description = self.aicore_errcode_map.get(errcode, lb.unrecorded_aicore_fault)
        errcode = lb.error_code + errcode if errcode else lb.suspected_op_script_error
        return [kernel_name, errcode, description]


class Response:
    NORMAL_ROOT_CAUSES = {
        KG_DIAG_NORMAL_ENTITY.code:
            RootCause(KG_DIAG_NORMAL_ENTITY.code, EntityAttribute(KG_DIAG_NORMAL_ENTITY.attribute))
    }
    MAX_LINK_DEPTH = 10

    def __init__(self):
        """
        Inference result model
        """
        self.analyze_success = False
        self.root_causes: Dict[str, RootCause] = dict()
        self.error = None
        self.all_out_edges = dict()

    def __repr__(self):
        return json.dumps(self.__dict__)

    def __str__(self):
        return self.__repr__()

    @staticmethod
    def get_root_causes(graph: Graph) -> dict:
        """
        Obtain root_causes based on event graph
        :param graph: event graph
        :return: root_causes
        """
        root_causes = dict()
        for vertex in graph.vertex_map.values():
            # 特殊处理，由于Aicore场景，上游CCAE需要解析AISW_CANN_Runtime_032及fault_kernel name字段。加链的同时，保证将该事件返回
            if vertex.in_size() == 0 or vertex.code == RUNTIME_AICORE_EXECUTE_FAULT:
                root_cause = root_causes.setdefault(vertex.code, RootCause(vertex.code, vertex.entity_attribute, []))
                root_cause.update_attribute_by_vertex(vertex)
        return root_causes

    @staticmethod
    def _format_link_key(event_name, event_code):
        """
        Format link key, for example, '故障编号（故障名称）'.
        :param event_name: event cause
        :param event_code: event ID
        :return: event_code（event_name）
        """
        return event_code + lb.left_bracket + event_name + lb.right_bracket

    def get_information(self, graph: Graph):
        self.analyze_success = True
        self.root_causes = self.get_root_causes(graph) or self.NORMAL_ROOT_CAUSES
        self._get_fault_chains(graph)
        return self

    def _get_fault_chains(self, graph: Graph):
        """
        Format link key, for example, '故障名称（故障编号）'.
        :param graph: event graph
        :return: fault chains list
        """
        for vertex in graph.vertex_map.values():
            # use out_edges to determine whether a link exists
            if vertex.out_size() == 0:
                continue
            self._get_out_edges(graph, vertex)
        self._get_fault_link(self.all_out_edges.keys(), [], [])

    def _get_out_edges(self, graph: Graph, vertex: Vertex):
        """
        Get out edges of the vertex
        :param graph: event graph
        :param vertex: event vertex
        """
        edges_list = []
        # find all destination vertices(to id) for each out vertex, form id is self
        for out_vertex in vertex.out_edges:
            vertex_to_id = graph.vertex_map.get(out_vertex.to_id)
            event_code = vertex_to_id.code
            event_name = getattr(getattr(vertex_to_id, "entity_attribute"), f"cause_{LANG}", "")
            edges_list.append((self._format_link_key(event_name, event_code), vertex_to_id.src_dev))
        self.all_out_edges[
            (self._format_link_key(getattr(getattr(vertex, "entity_attribute"),
                                           f"cause_{LANG}", ""), vertex.code), vertex.src_dev)] = edges_list

    def _get_fault_link(self, root_keys: [], link_list: [], devices: [], max_depth=0):
        """
        Get all fault link with root_keys as the root link by recursion
        :param root_keys: root link vertex list
        :param link_list: a fault link
        :param devices: event devices
        :param max_depth: maximum depth of the chain
        """
        for root_key, src_dev in root_keys:
            new_link_list = link_list + [root_key]
            new_devices = devices + [src_dev]
            # root_code e.g. 'fault_code（fault_name）', need 'fault_code'
            root_code = new_link_list[0].split(lb.left_bracket)[0]
            if root_code not in self.root_causes.keys():
                continue
            out_edge_list = self.all_out_edges.get((root_key, src_dev), [])
            # if over max link depth, no recursive
            if out_edge_list and max_depth < self.MAX_LINK_DEPTH:
                self._get_fault_link(out_edge_list, new_link_list, new_devices, max_depth + 1)
                continue
            # root_cause and link_str e.g. fault_code1（fault_name1）-> fault_code2（fault_name2）-> ...
            root_cause = self.root_causes.setdefault(root_code, RootCause(root_code, EntityAttribute(dict()), []))
            # only the longest fault chain is retained.
            chain_key = new_devices[0] if new_devices else ""
            if len(root_cause.chains.get(chain_key, "").split("-> ")) < len(new_link_list):
                link_str = "-> ".join(new_link_list)
                root_cause.chains[chain_key] = link_str
