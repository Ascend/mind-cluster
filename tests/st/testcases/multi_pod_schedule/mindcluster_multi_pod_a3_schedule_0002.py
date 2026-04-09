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
from tests.st.st_dev.JobTool import JobHelper
from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sNode import K8sNode
from tests.st.st_dev.K8sTool import K8sTool


class MindclusterA3JobReschedule(unittest.TestCase):
    base_dir = BASE_DIR
    resource_dir = os.path.join(base_dir, "multi_pod_schedule/resources_0002/")
    job_yaml = resource_dir + "job_llama-2x16.yaml"
    job_name = ["default-test-2x16"]
    k8s_manager = K8sDistributedManage()
    logger = k8s_manager.logger
    master_pod_name = job_name[0] + "-master-0"
    worker_pod_name = job_name[0] + "-worker-0"

    @classmethod
    def setUpClass(cls):
        ClusterSimulator.mock_kwok_cluster_a3(cls, "910csuperpod", 1, 3)
        K8sTool.modify_volcano_yaml(cls, 3)

    def setUp(self) -> None:
        self.test_method_name = self._testMethodName
        self.logger.info("test method: %s", self.test_method_name)

    def test_mindcluster_a3_job_reschedule_001(self):
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), 3, "kwok nodes are not ready")

    def test_mindcluster_a3_job_reschedule_002(self):
        K8sNode.set_accelerator_type_a3(self, "910csuperpod", 3, "module-a3-16-super-pod")
        self.assertIs(ClusterSimulator.get_kwok_nodes_with_accelerator_type(self, "module-a3-16-super-pod"),
                      3, "kwok nodes with a3 accelerator type are not ready")

    def test_mindcluster_a3_job_reschedule_003(self):
        K8sTool.apply_mindcluster_v2(self)
        self.assertTrue(CaseRoutines.check_mind_cluster(self), "mind cluster is not ready")

    def test_mindcluster_a3_job_reschedule_004(self):
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.job_yaml)
        self.assertTrue(K8sTool.check_pod_status(self, self.master_pod_name, timeout=60), "master pod is not ready")
        self.assertTrue(K8sTool.check_pod_status(self, self.worker_pod_name, timeout=60), "worker pod is not ready")

    def test_mindcluster_a3_job_reschedule_005(self):
        K8sTool.insert_software_fault(self, ns="default", pod_name=self.worker_pod_name)
        self.assertTrue(K8sTool.check_all_pods_status(self, pod_names=[self.worker_pod_name, self.master_pod_name],
                                                      status=["Error", "Pending"]), "master or worker pod is not error")

    def test_mindcluster_a3_job_reschedule_006(self):
        self.assertTrue(K8sTool.check_pod_status(self, self.master_pod_name, timeout=60), "master pod is not running")
        self.assertTrue(K8sTool.check_pod_status(self, self.worker_pod_name, timeout=60), "worker pod is not running")

    def test_mindcluster_a3_job_reschedule_007(self):
        ret = JobHelper.check_pod_label_exist(self, self.worker_pod_name, "software-fault")
        self.assertFalse(ret, "worker pod is not rebuild")

    def test_mindcluster_a3_job_reschedule_008(self):
        K8sTool.insert_software_fault(self, ns="default", pod_name=self.master_pod_name)
        self.assertTrue(K8sTool.check_all_pods_status(self, pod_names=[self.worker_pod_name, self.master_pod_name],
                                                      status=["Error", "Pending"]), "master or worker pod is not error")

    def test_mindcluster_a3_job_reschedule_009(self):
        self.assertTrue(K8sTool.check_pod_status(self, self.master_pod_name, timeout=60), "master pod is not running")
        self.assertTrue(K8sTool.check_pod_status(self, self.worker_pod_name, timeout=60), "worker pod is not running")

    def test_mindcluster_a3_job_reschedule_010(self):
        ret = JobHelper.check_pod_label_exist(self, self.master_pod_name, "software-fault")
        self.assertFalse(ret, "master pod is not rebuild")

    def test_mindcluster_a3_job_reschedule_011(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml)
        self.assertTrue(K8sTool.check_pod_deleted(self, self.job_name[0]), "job are still running")

    @classmethod
    def tearDownClass(cls):
        ClusterSimulator.stop_cluster_simulator_a3(cls)
        cls.k8s_manager.exec_command("kubectl delete -f %s" % cls.job_yaml)
        K8sTool.reset_volcano_yaml(cls)
