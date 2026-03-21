#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import logging
import os
import time
from typing import List

from tests.st.st_dev.K8sNode import K8sNode
from tests.st.lib.dl.DLConf import K8S_VOLCANO


class InitEnvironment:
    pass


class K8sDistributedManage(object):

    def __init__(self):
        self.logger = logging.getLogger("mindcluster")
        self.env_info = {}
        self.nodes = self.get_current_nodes()
        self.master_nodes: List[K8sNode] = []
        self.worker_nodes: List[K8sNode] = []
        self.sim_worker_nodes: List = []
        self.master: K8sNode = None
        self.__get_roles_nodes()

    def get_current_nodes(self):
        node = K8sNode(os.environ['ipv4_address'], os.environ['username'], os.environ['password'])
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
        raise Exception("get volcano version failed")

    def check_all_nodes_ready(self):
        for node in self.nodes:
            if "NotReady" in node.status:
                raise Exception("node %s not ready" % node.node_name)
            if "SchedulingDisabled" in node.status:
                cmd = "kubectl uncordon %s" % node.node_name
                self.master.exec_command(cmd)
            cmd = "npu-smi info | grep 910 | awk '{print $3}' | wc -l"
            assert node.exec_command(cmd) == "8", f"{node.ip} npu num is not 8"
            for dev_id in range(8):
                cmd = "hccn_tool -i %s -net_health -g" % dev_id
                info = node.exec_command(cmd)
                if "Success" not in info:
                    self.build_node_link_up(node.ip, dev_id)
                    time.sleep(5)

    def get_task_nodes_ip_list(self, task_name):
        cmd = "kubectl get pods -A -owide | grep %s | awk '{print $8}'" % task_name
        node_info = self.master.exec_command(cmd)
        task_n_li = node_info.splitlines()
        self.logger.info("task node: %s" % task_n_li)
        ip_list = list()
        for node in self.nodes:
            [
                ip_list.append(node.ip) for task_n in task_n_li
                if node.node_name == task_n
            ]
        self.logger.info("return %s" % ip_list)
        return ip_list

    def exec_command(self, cmd: str):
        return self.master.exec_command(cmd)

    def __get_roles_nodes(self):
        worker_list = []
        master_list = []
        for node in self.nodes:
            if "master" in node.role:
                master_list.append(node)
            if "worker" in node.role:
                worker_list.append(node)
            if "control-plane" in node.role and node not in master_list:
                master_list.append(node)
        self.master_nodes = master_list
        self.worker_nodes = worker_list
        self.master = master_list[0]
