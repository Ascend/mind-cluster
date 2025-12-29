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

from ascend_fd.utils.load_kg_config import EntityAttribute


class Edge:

    def __init__(self, form_id, to_id, r_type="trigger"):
        """
        Init edge
        :param form_id: id of source vertex
        :param to_id: id of destination vertex
        :param r_type: edge type
        """
        self.form_id = form_id
        self.to_id = to_id
        self.r_type = r_type

    def get_from_id(self) -> str:
        return self.form_id

    def get_to_id(self) -> str:
        return self.to_id

    def get_type(self) -> str:
        return self.r_type


class Vertex:
    def __init__(self, code, event_attribute: Dict, entity_attribute: EntityAttribute):
        """
        Init Vertex
        :param code: vertex code
        :param event_attribute: event attribute
        :param entity_attribute: entity attribute for vertex code
        """
        self.v_id = event_attribute.get("event_id")
        self.code = code
        self.event_attribute = event_attribute
        self.entity_attribute = entity_attribute
        self.src_dev = event_attribute.get("source_device", "Unknown")
        self.out_edges: List[Edge] = []
        self.in_edges: List[Edge] = []

    @staticmethod
    def _check_add(edge: Edge, edges: List[Edge]):
        """
        Check whether edges are added, and add an edge
        :param edge: edge to be added
        :param edges: edges of vertex
        :return:
        """
        if not edge:
            return False
        for out_edge in edges:
            if out_edge.get_type() == edge.get_type() and out_edge.get_to_id() == edge.get_to_id():
                return False
        return True

    def out_size(self) -> int:
        return len(self.out_edges)

    def in_size(self) -> int:
        return len(self.in_edges)

    def add_out(self, edge: Edge):
        self.out_edges.append(edge)

    def add_in(self, edge: Edge):
        self.in_edges.append(edge)

    def check_out_add(self, edge: Edge):
        return self._check_add(edge, self.out_edges)

    def check_in_add(self, edge: Edge):
        return self._check_add(edge, self.in_edges)

    def eval(self, prop_key):
        # get the attributes of the event first
        return self.event_attribute.get(prop_key, None) or getattr(self.entity_attribute, prop_key, None)

    def get_code(self):
        return self.code

    def get_id(self):
        return self.v_id
