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
from typing import Dict

from ascend_fd.utils.regular_table import TRACEBACK_FAULT_ENTITY_ATTR
from ascend_fd.utils.fault_code import PRE_TRACEBACK_FAULT, CANN_ERRCODE_CUSTOM, PYTORCH_ERRCODE_COMMON, \
    MINDIE_ERRCODE_COMMON
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.graph import Graph
from ascend_fd.utils.load_kg_config import Schema, SchemaEntity
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.vertex import Edge, Vertex
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.model.package_data import PackageData
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_compiler import ExprCompiler
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_condition import ExprCondition
from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.expr.expr_computer import ExprComputer


class GraphBuilder:
    _EVENT_CODE = "event_code"

    def __init__(self, schema: Schema, pkg_data: PackageData):
        """
        Init Graph builder
        :param schema: entities from kg-config.json
        :param pkg_data: the data package path
        """
        self.schema = schema
        self.pkg_data = pkg_data
        self.has_effective_event = False
        self._judge_event()
        self.graph = Graph()
        self._expr_cache: Dict[str, ExprCondition] = {}
        self._compiler = ExprCompiler()
        self.has_root_device = False

    def build_graph(self):
        self._add_vertices()
        self._add_edges()
        self._infer_device()
        return self.graph

    def _judge_event(self):
        """
        Check whether events contain lower-layer faults
        """
        for event in self.pkg_data.event_map.values():
            event_code = event.get(self._EVENT_CODE, "")
            is_custom_event = event.get("is_custom_event", False)
            # 有效的故障事件：非traceback故障并且非自定义故障
            if not event_code.startswith(PRE_TRACEBACK_FAULT) and not is_custom_event:
                self.has_effective_event = True
                break

    def _add_vertices(self):
        """
        Create vertex for all pkg_data event and add to graph
        """
        for event in self.pkg_data.event_map.values():
            event_code = event.get(self._EVENT_CODE)
            schema_entity = self._handle_schema_entity(event_code)
            if not isinstance(event_code, str) or not schema_entity:
                continue
            self.graph.add_vertex(Vertex(event_code, event, schema_entity.attribute))

    def _handle_schema_entity(self, event_code):
        """
        Handle schema_entity of different fault events
        """
        if event_code.startswith(CANN_ERRCODE_CUSTOM):
            schema_entity = self.schema.get_schema_entity(CANN_ERRCODE_CUSTOM)
            self.schema.add_custom_event_to_schema(event_code, schema_entity)
            return schema_entity
        if event_code.startswith(PYTORCH_ERRCODE_COMMON):
            schema_entity = self.schema.get_schema_entity(PYTORCH_ERRCODE_COMMON)
            self.schema.add_custom_event_to_schema(event_code, schema_entity)
            return schema_entity
        if event_code.startswith(MINDIE_ERRCODE_COMMON):
            schema_entity = self.schema.get_schema_entity(MINDIE_ERRCODE_COMMON)
            self.schema.add_custom_event_to_schema(event_code, schema_entity)
            return schema_entity
        # traceback faults are displayed only when there is no valid fault event.
        if not self.has_effective_event and event_code.startswith(PRE_TRACEBACK_FAULT):
            return SchemaEntity(entity_code=event_code, attribute=TRACEBACK_FAULT_ENTITY_ATTR)
        return self.schema.get_schema_entity(event_code)

    def _add_edges(self):
        """
        Create edges between all vertices based on schema rules, and add to graph
        """
        for event_code, event_list in self.graph.v_indexes_map.items():
            event_entity = self.schema.get_schema_entity(event_code)
            if not event_entity:
                return
            for single_rule in event_entity.rule:
                dst_event_code, expression = single_rule.dst_code, single_rule.expression
                dst_event_list = self.graph.v_indexes_map.get(dst_event_code, None)
                self._match_vertex(dst_event_list, event_list, expression)

    def _match_vertex(self, dst_event_list, event_list, expression):
        """
        Check whether source events and destination events meet the expression, and add edges.
        :param dst_event_list: destination events list with the same code
        :param event_list: source events list with the same code
        :param expression: expression rule between source events and destination events
        """
        if not dst_event_list:
            return
        for dst_event in dst_event_list:
            for src_event in event_list:
                if expression and expression not in self._expr_cache:
                    self._expr_cache[expression] = ExprCondition(expression, self._compiler)
                expr_computer = ExprComputer(src_event, dst_event)
                if not (dst_event.src_dev == "Unknown" or src_event.src_dev == "Unknown" or
                        dst_event.src_dev == src_event.src_dev):
                    continue
                if not expression or expr_computer.compute(self._expr_cache.get(expression, None)):
                    self.graph.add_edge(Edge(src_event.get_id(), dst_event.get_id()))

    def _infer_device(self):
        """
        Infer the device where the fault event occurs
        """
        if not self.pkg_data.root_device_list:
            # all source_devices are 'Unknown'
            return
        for vertex in self.graph.vertex_map.values():
            if vertex.in_size() != 0 and vertex.out_size() == 0:
                self.has_root_device = False
                device = vertex.event_attribute.get("source_device", "Unknown")
                self._in_vertex_set_device(vertex, device)

    def _in_vertex_set_device(self, vertex: Vertex, device: str):
        """
        Traverse the forward vertex and set device
        :param vertex: event vertex
        :param device: source device
        """
        if device == self.pkg_data.root_device_list[0]:
            # 只要链中有一个故障事件的device不为Unknown，has_root_device就会置为True
            self.has_root_device = True
        for in_edge in vertex.in_edges:
            in_vertex = self.graph.vertex_map.get(in_edge.form_id)
            in_device = in_vertex.event_attribute.get("source_device", "Unknown")
            self._in_vertex_set_device(in_vertex, in_device)
        # 该vertex事件的device为Unknown且故障链路中有root_device，则推断出该事件的device也为root_device
        if device == "Unknown" and self.has_root_device:
            vertex.event_attribute["source_device"] = self.pkg_data.root_device_list[0]
            vertex.src_dev = self.pkg_data.root_device_list[0]
