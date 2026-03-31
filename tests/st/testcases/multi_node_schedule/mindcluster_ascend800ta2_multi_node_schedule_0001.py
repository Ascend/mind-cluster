#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import unittest

from tests.st.st_dev.CaseRoutines import CaseRoutines
from tests.st.st_dev.ClusterSimulatorTool import ClusterSimulator
from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sTool import K8sTool
from tests.st.st_dev.K8sNode import K8sNode

NODE_NUM = 8
NODE_NAME = "910ax8"
MODULE_910B_8 = "module-910b-8"
BASE_DIR = "/workspace/mind-cluster/tests/st/testcases/"

class MindclusterAscend800ta2MutliNodeSchedule0001(unittest.TestCase):
    resource_dir = BASE_DIR + "multi_node_schedule/resources_0001/"
    job_yaml_path1 = resource_dir + "job_llama-8x8.yaml"
    job_name1 = "default-test-8x8"
    k8s_manager = K8sDistributedManage()
    logger = k8s_manager.logger

    @classmethod
    def setUpClass(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)

    def test_valid_job_000(self):
        K8sTool.apply_mindcluster(self)
        self.assertTrue(CaseRoutines.check_mind_cluster(self), "mindcluster is not ready")

    def test_valid_job_001(self):
        ClusterSimulator.create_kwok_cluster(self, node_name=NODE_NAME, node_num=NODE_NUM + 2)

    def test_valid_job_002(self):
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), NODE_NUM + 2)

    def test_valid_job_003(self):
        K8sNode.set_accelerator_type(self, node_name=NODE_NAME, node_num=NODE_NUM + 2, accelerator_type=MODULE_910B_8)
        self.assertIs(ClusterSimulator.get_kwok_nodes_with_accelerator_type(self), NODE_NUM + 2)

    def test_valid_job_004(self):
        self.k8s_manager.exec_command("kubectl cordon localhost.localdomain master")
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.job_yaml_path1)
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1), "pod is not running")


    def test_valid_job_005(self):
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1), "pod is not running")
        master_node = K8sNode.get_node_by_pod_name(self, self.job_name1+"-master-0")
        ClusterSimulator.inject_kwok_software_fault(self, namespace="default", pod_name=self.job_name1+"-worker-0")
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1), "pod is not running")

        master_node_after = K8sNode.get_node_by_pod_name(self, self.job_name1+"-master-0")
        self.assertTrue(master_node == master_node_after,"master_node should equal")

    def test_valid_job_006(self):
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1), "pod is not running")
        ClusterSimulator.inject_kwok_software_fault(self, namespace="default", pod_name=self.job_name1+"-master-0")
        self.assertTrue(K8sTool.check_pod_status(self, self.job_name1, timeout=120), "pod is not running")

    def test_valid_job_007(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        ret = self.k8s_manager.exec_command(f"kubectl get pod {self.job_name1}'")
        self.assertTrue(len(ret) == 0, "job delete fail")


    def test_valid_job_008(self):
        ClusterSimulator.delete_kwok_cluster(self)
        self.assertIs(ClusterSimulator.get_ready_kwok_node_count(self), 0)

    @classmethod
    def tearDownClass(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        ClusterSimulator.delete_kwok_cluster(self)
        self.k8s_manager.exec_command("kubectl uncordon localhost.localdomain master")

if __name__ == '__main__':
    unittest.main()
