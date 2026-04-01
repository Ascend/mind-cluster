#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import os
import unittest

from tests.st.lib.dl_deployer.install_manager import InstallManager
from tests.st.st_dev.K8sDistributedManage import K8sDistributedManage
from tests.st.st_dev.K8sTool import K8sTool


class MindclusterApplyTest(unittest.TestCase):
    installer = None
    k8s_manager = K8sDistributedManage()

    def get_manager(self, component_name):
        if self.installer:
            self.installer.component_name = component_name
            return
        ip = os.environ.get("ipv4_address")
        username = os.environ.get("username")
        password = os.environ.get("password")
        file_path = os.environ.get("PR_OUTPUT_DIR")
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