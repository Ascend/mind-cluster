#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import os

from tests.st.lib.dl_deployer.dl import Installer
from tests.st.lib.dl_deployer.install_device_plugin import DevicePluginInstaller
from tests.st.lib.dl_deployer.install_volcano import VolcanoInstaller
from tests.st.lib.common.CLI import ClassCLI


class InstallManager:
    arch = "aarch64"

    def __init__(self, ip, username, password, resource_dir, component_name):
        self.cli = ClassCLI(ip, username, password)
        self.component_name = component_name
        self.component_installer = None
        self.resource_dir = resource_dir
        self.installer_dict = {
            'device-plugin': DevicePluginInstaller(ip, username, password, resource_dir),
            'ascend-operator': Installer(ip, username, password, resource_dir),
            'noded': Installer(ip, username, password, resource_dir),
            'npu-exporter': Installer(ip, username, password, resource_dir),
            'volcano': VolcanoInstaller(ip, username, password, resource_dir),
            'clusterd': Installer(ip, username, password, resource_dir),
        }

    def execute(self):
        if self.component_installer:
            self.component_installer.step = 'build'
            self.component_installer.run()
            self.component_installer.step = 'push'
            self.component_installer.run()
            self.component_installer.step = 'install'
            self.component_installer.run()
            self.component_installer.step = 'apply'
            self.component_installer.run()

    def entry(self):
        pkgs = os.listdir(self.resource_dir)
        for pkg in pkgs:
            if self.component_name in pkg and self.arch in pkg:
                installer = self.installer_dict[self.component_name]
                installer.component_name = self.component_name
                installer.package_name = pkg
                self.component_installer = installer
                self.execute()
