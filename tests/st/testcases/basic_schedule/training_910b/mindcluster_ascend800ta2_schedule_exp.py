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

from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sTool import K8sTool

from tests.st.envs import BASE_DIR


class MindclusterAscend800ta2Schedule0001(unittest.TestCase):
    base_dir = BASE_DIR + "basic_schedule/training_910b/"
    resource_dir = base_dir + "resources_0001/"
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

    def setUp(self):
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path1)
        self.k8s_manager.exec_command("kubectl delete -f %s" % self.job_yaml_path2)

    def test_invalid_job_001(self):
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.job_yaml_path2)
        assert K8sTool.check_pod_deleted(self, self.job_name2) or \
               K8sTool.check_acjob_status(self, self.job_name2)

    def test_valid_job_001(self):
        self.k8s_manager.exec_command("kubectl apply -f %s" % self.job_yaml_path1)
        assert K8sTool.check_pod_status(self, self.job_name1)
