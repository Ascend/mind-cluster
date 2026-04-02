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
import os.path
import unittest

from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sTool import K8sTool
from tests.st.envs import MIND_CLUSTER_YAML_DIR


class MindClusterComponentRecoverCases(unittest.TestCase):
    k8s_manager = K8sDistributedManage()
    base_dir = MIND_CLUSTER_YAML_DIR
    dp_yaml = "device-plugin-volcano-*.yaml"
    volcano_yaml = "volcano-*.yaml"
    clusterd_yaml = "clusterd-*.yaml"
    ascend_operator_yaml = "ascend-operator-*.yaml"
    noded_yaml = "noded-*.yaml"
    npu_exporter_yaml = "npu-exporter-*.yaml"

    def test_mindcluster_recover_dp(self):
        dp_path = os.path.join(self.base_dir, self.dp_yaml)
        K8sTool.apply_yaml_by_file(self, dp_path)
        self.assertTrue(K8sTool.check_pod_status(self, "device-plugin"))

    def test_mindcluster_recover_volcano(self):
        volcano_path = os.path.join(self.base_dir, self.volcano_yaml)
        K8sTool.apply_yaml_by_file(self, volcano_path)
        self.assertTrue(K8sTool.check_pod_status(self, "volcano"))

    def test_mindcluster_recover_clusterd(self):
        clusterd_path = os.path.join(self.base_dir, self.clusterd_yaml)
        K8sTool.apply_yaml_by_file(self, clusterd_path)
        self.assertTrue(K8sTool.check_pod_status(self, "clusterd"))

    def test_mindcluster_recover_ascend_operator(self):
        ascend_operator_path = os.path.join(self.base_dir, self.ascend_operator_yaml)
        K8sTool.apply_yaml_by_file(self, ascend_operator_path)
        self.assertTrue(K8sTool.check_pod_status(self, "ascend-operator"))

    def test_mindcluster_recover_node_yaml(self):
        noded_path = os.path.join(self.base_dir, self.noded_yaml)
        K8sTool.apply_yaml_by_file(self, noded_path)
        self.assertTrue(K8sTool.check_pod_status(self, "noded"))

    def test_mindcluster_recover_npu_exporter(self):
        npu_exporter_path = os.path.join(self.base_dir, self.npu_exporter_yaml)
        K8sTool.apply_yaml_by_file(self, npu_exporter_path)
        self.assertTrue(K8sTool.check_pod_status(self, "npu-exporter"))
