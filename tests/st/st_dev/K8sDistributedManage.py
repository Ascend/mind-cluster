#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026. Huawei Technologies Co.,Ltd. All rights reserved.
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
import logging
from typing import List

from tests.st.st_dev.K8sNode import K8sNode
from tests.st.lib.dl.DLConf import K8S_VOLCANO
from tests.st.envs import ipv4_address, username, password


class K8sDistributedManage:
    def __init__(self):
        self.logger = logging.getLogger("mindcluster")
        self.env_info = {}
        self.nodes = self.get_current_nodes()
        self.master_nodes: List[K8sNode] = []
        self.worker_nodes: List[K8sNode] = []
        self.sim_worker_nodes: List = []
        self.master: K8sNode = None
        self._get_roles_nodes()
        self._init_node_name()

    def get_current_nodes(self):
        node = K8sNode(ipv4_address, username, password)
        return [node]

    def refresh_nodes_info(self):
        nodes_info = self.master.exec_command("kubectl get nodes").splitlines()
        nodes_info.pop(0)
        for node_info in nodes_info:
            node_info = node_info.split()
            node_name = node_info[0]
            node = self.get_node_by_name(node_name)
            node.status = node_info[1]
            node.role = node_info[2]
            node.version = node_info[4]

    def get_node_by_name(self, node_name):
        for node in self.nodes:
            if node.node_name == node_name:
                return node
        return None

    def get_volcano_version(self):
        k8s = self.master.exec_command("kubelet version")
        for k8s_version, volcano_version in K8S_VOLCANO.items():
            if k8s_version in k8s:
                return volcano_version
        raise RuntimeError("get volcano version failed")

    def get_task_nodes_ip_list(self, task_name):
        cmd = "kubectl get pods -A -owide | grep %s | awk '{print $8}'" % task_name
        node_info = self.master.exec_command(cmd)
        task_n_li = node_info.splitlines()
        self.logger.info("task node: %s", task_n_li)
        ip_list = []
        for node in self.nodes:
            for task_n in task_n_li:
                if node.node_name == task_n:
                    ip_list.append(node.ip)
        self.logger.info("return %s", ip_list)
        return ip_list

    def exec_command(self, cmd: str):
        return self.master.exec_command(cmd)

    def _get_roles_nodes(self):
        self.master_nodes = self.nodes
        self.worker_nodes = self.nodes
        self.master = self.nodes[0]

    def _init_node_name(self):
        node_info = self.master.exec_command("kubectl get nodes -o wide")
        for line in node_info.splitlines()[1:]:
            parts = line.split()
            node_name = parts[0]
            status = parts[1]
            role = parts[2]
            internal_ip = parts[5]
            for node in self.nodes:
                if node.ip == internal_ip:
                    node.node_name = node_name
                    node.status = status
                    node.role = role
                    self.logger.info("init node_name:%s", node_name)
