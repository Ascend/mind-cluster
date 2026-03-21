#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import unittest

from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sTool import K8sTool


class MindclusterAscend800ta2Schedule0001(unittest.TestCase):
    base_dir = "/workspace/mind-cluster/tests/st/testcases/"
    resource_dir = base_dir + "basic_schedule/resources_0001/"
    job_yaml_path1 = resource_dir + "job_llama-1x8.yaml"
    job_yaml_path2 = resource_dir + "job_llama-2x8.yaml"
    model_yaml_path = resource_dir + "llama70b-pt.yaml"
    job_name1 = "default-test-1x8"
    job_name2 = "default-test-2x8"
    k8s_manager = K8sDistributedManage()

    @classmethod
    def setUpClass(cls) -> None:
        cls.k8s_manager.exec_command("kubectl apply -f %s" % cls.model_yaml_path)

    @classmethod
    def tearDownClass(cls):
        cls.k8s_manager.exec_command("kubectl delete -f %s" % cls.job_yaml_path1)
        cls.k8s_manager.exec_command("kubectl delete -f %s" % cls.job_yaml_path2)

    def setUp(self, methodName='mindcluster_ascend800ta2_schedule_0001'):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path2)

    def test_invalid_job_001(self):
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.job_yaml_path2)
        assert K8sTool.check_pod_deleted(self, self.job_name2) or \
               K8sTool.check_acjob_status(self, self.job_name2)

    def test_valid_job_001(self):
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.job_yaml_path1)
        assert K8sTool.check_pod_status(self, self.job_name1)

if __name__ == '__main__':
    unittest.main()
