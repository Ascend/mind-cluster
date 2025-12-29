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
from typing import List, Dict

from ascend_fd.pkg.diag.knowledge_graph.kg_engine.graph.vertex import Vertex, Edge


class Graph:
    def __init__(self):
        """
        Init graph
        """
        self.vertex_map: Dict[str, Vertex] = {}  # key is vertex id
        self.edges_map: Dict[str, List[Edge]] = {}
        self.v_indexes_map: Dict[str, List[Vertex]] = {}  # key is entity code

    def add_vertex(self, vertex: Vertex):
        """
        Add a vertex to graph
        :param vertex:  vertex of fault event
        """
        if vertex.get_id() in self.vertex_map:
            return
        self.vertex_map[vertex.get_id()] = vertex
        self.v_indexes_map.setdefault(vertex.get_code(), []).append(vertex)

    def add_edge(self, edge: Edge):
        """
        Add an edge to graph
        :param edge: edge of source vertex to destination vertex
        """
        if edge.get_from_id() not in self.vertex_map or edge.get_to_id() not in self.vertex_map:
            return
        if self.vertex_map.get(edge.get_from_id()).check_out_add(edge):
            self.vertex_map.get(edge.get_from_id()).add_out(edge)  # source vertex add an out edge
            self.vertex_map.get(edge.get_to_id()).add_in(edge)  # destination vertex add an in edge
            self.edges_map.setdefault(edge.get_type(), []).append(edge)
