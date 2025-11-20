#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import json
from copy import deepcopy


class Topo(object):

    def __init__(self, topo_file_path):
        self._topo = {}
        with open(topo_file_path) as topo_file:
            self._topo = json.load(topo_file)
        # Special handling for server 16p
        tmp = []
        for edge in self._topo["edge_list"]:
            if (
                edge["link_type"] == "PEER2PEER"
                and len(edge["local_a_ports"]) > 1
                and len(edge["local_b_ports"]) > 1
            ):
                tmp_edge = deepcopy(edge)
                tmp_edge["local_a"] = edge["local_b"]
                tmp_edge["local_b"] = edge["local_a"]
                tmp_edge["local_a_ports"] = edge["local_b_ports"]
                tmp_edge["local_b_ports"] = edge["local_a_ports"]
                tmp.append(tmp_edge)
        self._topo["edge_list"].extend(tmp)

    def get_ports_by_level_and_die(self, local_id, level, die_id):
        ports = set()
        for edge in self._topo["edge_list"]:
            filter_condition = (
                edge["local_a"] == local_id
                and edge["net_layer"] == level
                and len(edge["local_a_ports"]) > 1
            )
            protocol_condition = "UB_TP" in edge["protocols"] or "UB_CTP" in edge["protocols"]
            if filter_condition and protocol_condition:
                ps = edge["local_a_ports"]
                ports.update([port[-1] for port in ps if port.startswith(f"{die_id}/")])
        return sorted([int(i) for i in ports])


class TopoSingleFactory:
    _topo_path = None
    _topo = None

    @staticmethod
    def set_topo_path(topo_path):
        TopoSingleFactory._topo_path = topo_path

    @staticmethod
    def get_topo():
        if TopoSingleFactory._topo is None:
            TopoSingleFactory._topo = Topo(TopoSingleFactory._topo_path)
        return TopoSingleFactory._topo
