#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import os
import unittest

from tests.st.lib.dl_deployer.install_manager import InstallManager
from tests.st.st_dev.K8sTool import K8sTool


class MindclusterApplyTest(unittest.TestCase):

    def setUp(self, methodName='mindcluster_components_apply_0001'):
        pass

    def get_manager(self, component_name):
        if self.installer:
            return self.installer
        ip = os.environ["ipv4_address"]
        username = os.environ["username"]
        password = os.environ["password"]
        file_path = os.environ["PR_OUTPUT_DIR"]
        return InstallManager(ip, username, password, file_path, component_name)

    def test_apply_dp(self):
        manager = self.get_manager("device-plugin")
        manager.entry()
        assert self._check_pod_status("device-plugin")

    def test_apply_volcano(self):
        manager = self.get_manager("volcano")
        manager.entry()
        assert self._check_pod_status("volcano")

    def test_apply_ascend_operator(self):
        manager = self.get_manager("ascend-operator")
        manager.entry()
        assert self._check_pod_status("ascend-operator")

    def test_apply_npu_exporter(self):
        manager = self.get_manager("npu-exporter")
        manager.entry()

    def test_apply_noded(self):
        manager = self.get_manager("noded")
        manager.entry()
        assert self._check_pod_status("noded")

    def test_apply_clusterd(self):
        manager = self.get_manager("clusterd")
        manager.entry()
        assert self._check_pod_status("clusterd")

    def _check_pod_status(self, component_name):
        return K8sTool.check_pod_status(self, component_name)


if __name__ == '__main__':
    unittest.main()
