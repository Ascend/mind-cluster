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
# pylint: disable=duplicate-code
import unittest

from tests.st.envs import BASE_DIR
from tests.st.st_dev.CaseRoutines import CaseRoutines
from tests.st.st_dev.ClusterSimulatorTool import ClusterSimulator
from tests.st.st_dev.JobTool import JobHelper
from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sTool import K8sTool
from tests.st.st_dev.K8sNode import K8sNode

NODE_NUM = 8
NODE_NAME = "910ax8"
MODULE_910B_8 = "module-910b-8"


class MindclusterAscend800ta2MutliNodeSchedule0001(unittest.TestCase):
    resource_dir = BASE_DIR + "multi_node_schedule/resources_0001/"
    job_yaml_path1 = resource_dir + "job_kwok_simulator-8x8.yaml"
    job_name1 = "default-test-8x8"
    k8s_manager = K8sDistributedManage()
    logger = k8s_manager.logger
    ranktable_path = "/user/mindx-dl/ranktable/default.default-test-8x8/hccl.json"

    @classmethod
    def setUpClass(cls):
        cls.k8s_manager.exec_command("kubectl delete -f %s" % cls.job_yaml_path1)

    def setUp(self) -> None:
        self.test_method_name = self._testMethodName
        self.logger.info("test method: %s", self.test_method_name)

    def test_multinode_schedule_acjob_000(self):
        K8sTool.apply_mindcluster(self)
        self.assertTrue(CaseRoutines.check_mind_cluster(self), "mindcluster is not ready")

    def test_multinode_schedule_acjob_001(self):
        ClusterSimulator.create_kwok_cluster(self, "a2_container", NODE_NAME, node_num=NODE_NUM + 2)

    def test_multinode_schedule_acjob_002(self):
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), NODE_NUM + 2)

    def test_multinode_schedule_acjob_003(self):
        K8sNode.set_accelerator_type(self, node_name=NODE_NAME, node_num=NODE_NUM + 2, accelerator_type=MODULE_910B_8)
        self.assertIs(ClusterSimulator.get_kwok_nodes_with_accelerator_type(self), NODE_NUM + 2)

    def test_multinode_schedule_acjob_004(self):
        K8sTool.cordon_all_nodes(self)
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.job_yaml_path1)
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1), "pod is not running")

    def test_multinode_schedule_acjob_005(self):
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1), "pod is not running")
        master_node = K8sNode.get_node_by_pod_name(self, self.job_name1 + "-master-0")
        ClusterSimulator.inject_kwok_software_fault(self, namespace="default", pod_name=self.job_name1 + "-worker-0")
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1), "pod is not running")

        master_node_after = K8sNode.get_node_by_pod_name(self, self.job_name1 + "-master-0")
        self.assertTrue(master_node == master_node_after, "master_node should equal")

    def test_multinode_schedule_acjob_006(self):
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1), "pod is not running")
        ClusterSimulator.inject_kwok_software_fault(self, namespace="default", pod_name=self.job_name1 + "-master-0")
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1, timeout=120), "pod is not running")

    def test_multinode_schedule_acjob_007(self):
        server_count = JobHelper.get_server_count_from_ranktable(self, self.ranktable_path)
        self.assertTrue(server_count == 8, "ranktable check error")

    def test_multinode_schedule_acjob_008(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        ret = self.k8s_manager.exec_command(f"kubectl get pod {self.job_name1}'")
        self.assertTrue(len(ret) == 0, "job delete fail")

    def test_multinode_schedule_acjob_009(self):
        ClusterSimulator.stop_kwok_cluster(self, "a2_container")
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), 0)

    @classmethod
    def tearDownClass(cls):
        cls.k8s_manager.exec_command("kubectl delete -f %s" % cls.job_yaml_path1)
        ClusterSimulator.stop_kwok_cluster(cls, "a2_container")
        K8sTool.uncordon_all_nodes(cls)
