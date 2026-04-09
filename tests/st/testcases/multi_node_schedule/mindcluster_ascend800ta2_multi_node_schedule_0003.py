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
import unittest

from tests.st.envs import BASE_DIR
from tests.st.st_dev.CaseRoutines import CaseRoutines
from tests.st.st_dev.ClusterSimulatorTool import ClusterSimulator
from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sTool import K8sTool
from tests.st.st_dev.K8sNode import K8sNode

NODE_NUM = 8
NODE_NAME = "910ax8"
MODULE_910B_8 = "module-910b-8"

class MindclusterAscend800ta2MutliNodeSchedule0003(unittest.TestCase):
    resource_dir = BASE_DIR + "multi_node_schedule/resources_0001/"
    vcjob_yaml_path = resource_dir + "vcjob_kwok_simulator-8x8.yaml"
    vcjob_name = "default-test-vcjob-8pod-8npu"
    k8s_manager = K8sDistributedManage()
    logger = k8s_manager.logger

    @classmethod
    def setUpClass(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.vcjob_yaml_path)
        ClusterSimulator.create_kwok_cluster(self, node_name=NODE_NAME, node_num=NODE_NUM)

    def test_multinode_schedule_vcjob_000(self):
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), NODE_NUM, "kwok nodes are not ready")

    def test_multinode_schedule_vcjob_001(self):
        K8sTool.apply_mindcluster(self)
        self.assertTrue(CaseRoutines.check_mind_cluster(self), "mindcluster is not ready")

    def test_multinode_schedule_vcjob_002(self):
        K8sNode.set_accelerator_type(self, node_name=NODE_NAME, node_num=NODE_NUM, accelerator_type=MODULE_910B_8)
        self.assertIs(ClusterSimulator.get_kwok_nodes_with_accelerator_type(self), NODE_NUM, "kwok nodes with accelerator type are not ready")

    def test_multinode_schedule_vcjob_003(self):
        self.k8s_manager.exec_command("kubectl delete stage pod-complete")
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.vcjob_yaml_path)
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.vcjob_yaml_path)
        self.assertTrue(K8sTool.check_pod_status(self, self.vcjob_name), "pod is not running")

    def test_multinode_schedule_vcjob_004(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.vcjob_yaml_path)
        ret = self.k8s_manager.exec_command(f"kubectl get pod {self.vcjob_name}'")
        self.assertTrue(len(ret) == 0, "job delete fail")

    def test_multinode_schedule_vcjob_005(self):
        ClusterSimulator.delete_kwok_cluster(self)
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), 0)

    @classmethod
    def tearDownClass(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.vcjob_yaml_path)
        ClusterSimulator.delete_kwok_cluster(self)