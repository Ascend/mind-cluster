#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import random
import time
import unittest

from tests.st.st_dev.ClusterSimulatorTool import ClusterSimulator
from tests.st.st_dev.JobTool import JobHelper
from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sTool import K8sTool
from tests.st.st_dev.K8sNode import K8sNode

NODE_NUM = 8
NODE_NAME = "910ax8"
MODULE_910B_8 = "module-910b-8"
BASE_DIR = "/workspace/mind-cluster/tests/st/testcases/"

class MindclusterAscend800ta2MutliNodeSchedule0002(unittest.TestCase):
    resource_dir = BASE_DIR + "multi_node_schedule/resources_0001/"
    job_yaml_path1 = resource_dir + "deployment_kwok_simulator-8x8.yaml"
    job_name1 = "default-test-pytorch-8pod-8npu"
    k8s_manager = K8sDistributedManage()
    logger = k8s_manager.logger

    @classmethod
    def setUpClass(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)

    def test_multinode_schedule_deployment_000(self):
        K8sTool.apply_mindcluster(self)

    def test_multinode_schedule_deployment_001(self):
        ClusterSimulator.create_kwok_cluster(self, node_name=NODE_NAME, node_num=NODE_NUM + 2)

    def test_multinode_schedule_deployment_002(self):
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), NODE_NUM + 2)

    def test_multinode_schedule_deployment_003(self):
        K8sNode.set_accelerator_type(self, node_name=NODE_NAME, node_num=NODE_NUM + 2, accelerator_type=MODULE_910B_8)
        self.assertIs(ClusterSimulator.get_kwok_nodes_with_accelerator_type(self), NODE_NUM + 2)

    def test_multinode_schedule_deployment_004(self):
        self.k8s_manager.exec_command("kubectl cordon localhost.localdomain master")
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.job_yaml_path1)
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1), "pod is not running")

    def test_multinode_schedule_deployment_005(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        ret = self.k8s_manager.exec_command(f"kubectl get pod {self.job_name1}'")
        self.assertTrue(len(ret) == 0, "job delete fail")


    def test_multinode_schedule_deployment_006(self):
        ClusterSimulator.delete_kwok_cluster(self)
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), 0)

    @classmethod
    def tearDownClass(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        ClusterSimulator.delete_kwok_cluster(self)
        self.k8s_manager.exec_command("kubectl uncordon localhost.localdomain master")

if __name__ == '__main__':
    unittest.main()
