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
import os
import unittest

from tests.st.envs import BASE_DIR
from tests.st.st_dev.CaseRoutines import CaseRoutines
from tests.st.st_dev.ClusterSimulatorTool import ClusterSimulator
from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sNode import K8sNode
from tests.st.st_dev.K8sTool import K8sTool


class MindclusterMultiPoda3Schedule0003(unittest.TestCase):
    base_dir = BASE_DIR
    resource_dir = os.path.join(base_dir, "multi_pod_schedule/resources_0002/")
    deployment_yaml = resource_dir + "deployment-2x16.yaml"
    deployment_name = "default-test-deployment-2pod-16npu"
    statefulset_yaml = resource_dir + "statefulset-2x16.yaml"
    statefulset_name = "default-test-statefulset-2pod-16npu"
    k8s_manager = K8sDistributedManage()
    logger = k8s_manager.logger

    @classmethod
    def setUpClass(cls):
        ClusterSimulator.create_kwok_cluster_a3(
            cls, container_name="a3_container", node_name="910csuperpod", super_pod_num=1, super_pod_size=3
        )
        K8sTool.modify_volcano_yaml(cls, super_pod_size=3)

    def setUp(self) -> None:
        self.test_method_name = self._testMethodName
        self.logger.info("test method: %s", self.test_method_name)

    def test_multipod_schedule_common_000(self):
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), 3, "kwok nodes are not ready")

    def test_multipod_schedule_common_001(self):
        K8sNode.set_accelerator_type_a3(
            self, node_name="910csuperpod", node_num=3, accelerator_type="module-a3-16-super-pod"
        )
        self.assertIs(
            ClusterSimulator.get_kwok_nodes_with_accelerator_type(self, "module-a3-16-super-pod"),
            3,
            "kwok nodes with a3 accelerator type are not ready",
        )

    def test_multipod_schedule_common_002(self):
        K8sTool.apply_mindcluster_v2(self)
        self.assertTrue(CaseRoutines.check_mind_cluster(self), "mind cluster is not ready")

    def test_multipod_schedule_deployment_003(self):
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.deployment_yaml)
        self.assertTrue(K8sTool.check_pod_status(self, self.deployment_name), "pod is not running")

    def test_multipod_schedule_deployment_004(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.deployment_yaml)
        self.assertTrue(K8sTool.check_pod_deleted(self, self.deployment_name), "job are still running")

    def test_multipod_schedule_statefulset_005(self):
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.statefulset_yaml)
        self.assertTrue(K8sTool.check_pod_status(self, self.statefulset_name), "pod is not running")

    def test_multipod_schedule_statefulset_006(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.statefulset_yaml)
        self.assertTrue(K8sTool.check_pod_deleted(self, self.statefulset_name), "job are still running")

    @classmethod
    def tearDownClass(cls):
        cls.k8s_manager.exec_command("kubectl delete -f %s" % cls.deployment_yaml)
        cls.k8s_manager.exec_command("kubectl delete -f %s" % cls.statefulset_yaml)
        ClusterSimulator.stop_kwok_cluster(cls, "a3_container")
        K8sTool.reset_volcano_yaml(cls)
