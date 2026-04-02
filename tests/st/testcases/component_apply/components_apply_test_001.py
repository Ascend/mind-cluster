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

from tests.st.lib.dl_deployer.install_manager import InstallManager
from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sTool import K8sTool
from tests.st.envs import ipv4_address, username, password, PR_OUTPUT_DIR


class MindclusterApplyTest(unittest.TestCase):
    installer = None
    k8s_manager = K8sDistributedManage()

    def get_manager(self, component_name):
        if self.installer:
            self.installer.component_name = component_name
            return
        ip = ipv4_address
        file_path = PR_OUTPUT_DIR
        self.installer = InstallManager(ip, username, password, file_path, component_name)

    def test_apply_dp(self):
        self.get_manager("device-plugin")
        self.installer.entry()
        self.assertTrue(self._check_pod_status("device-plugin"))

    def test_apply_volcano(self):
        self.get_manager("volcano")
        self.installer.entry()
        self.assertTrue(self._check_pod_status("volcano"))

    def test_apply_ascend_operator(self):
        self.get_manager("ascend-operator")
        self.installer.entry()
        self.assertTrue(self._check_pod_status("ascend-operator"))

    def test_apply_npu_exporter(self):
        self.get_manager("npu-exporter")
        self.installer.entry()
        self.assertTrue(self._check_pod_status("npu-exporter"))

    def test_apply_noded(self):
        self.get_manager("noded")
        self.installer.entry()
        self.assertTrue(self._check_pod_status("noded"))

    def test_apply_clusterd(self):
        self.get_manager("clusterd")
        self.installer.entry()
        self.assertTrue(self._check_pod_status("clusterd"))

    def _check_pod_status(self, component_name):
        return K8sTool.check_pod_status(self, component_name)