#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
from tests.st.lib.dl_deployer.dl import Installer


class NpuExporterInstaller(Installer):
    component_name = 'npu-exporter'

    def get_modified_yaml_contents(self):
        lines = self._get_yaml_contents()
        for index, line in enumerate(lines):
            if "-containerMode=docker" in line:
                lines[index] = line.replace("-containerMode=docker",
                                            "-containerMode=containerd -containerd=/run/containerd/containerd.sock "
                                            "-endpoint=/run/containerd/containerd.sock")
                break
        return lines
