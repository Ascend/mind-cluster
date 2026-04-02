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
from tests.st.st_dev.CaseRoutines import CaseRoutines


class MindclusterSingleNodeA2Cases(unittest.TestCase):
    base_dir = "/workspace/mind-cluster/tests/st/testcases/basic_schedule/training_910b/"
    resource_dir = base_dir + "resources_0001/"
    job_yaml_path1 = resource_dir + "basic_training_1x8.yaml"
    job_name1 = "basic-training-1x8"
    k8s_manager = K8sDistributedManage()
    logger = k8s_manager.logger

    def test_mindcluster_single_node_a2_0001(self):
        assert CaseRoutines.check_mind_cluster(self)

    def test_mindcluster_single_node_a2_0002(self):
        K8sTool.apply_yaml_by_file(self, self.job_yaml_path1)
        assert K8sTool.check_pod_status(self, self.job_name1), "job not running!"

    def test_mindcluster_single_node_a2_0003(self):
        assert CaseRoutines.check_pod_in_train_iters(self, rank_index='0', job_name=self.job_name1)

    def test_mindcluster_single_node_a2_0004(self):
        K8sTool.delete_yaml_by_file(self, self.job_yaml_path1)

    def test_mindcluster_single_node_a2_0005(self):
        assert K8sTool.check_pod_deleted(self, self.job_name1)
