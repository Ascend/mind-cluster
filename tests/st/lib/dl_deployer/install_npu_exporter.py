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
