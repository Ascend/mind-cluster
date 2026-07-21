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

from tests.st.lib.dl_deployer.dl import Installer


class NodedInstaller(Installer):
    component_name = 'noded'

    @staticmethod
    def get_labels():
        return [
            "nodeDEnable=on",
        ]

    def get_yaml_path(self):
        """pick noded deployment yaml, exclude fdConfig.yaml and other config files"""
        for root, _, files in os.walk(self.extract_dir):
            for filename in files:
                if filename.startswith('noded-v') and filename.endswith('.yaml'):
                    return os.path.join(root, filename)
        raise RuntimeError("Failed to find noded yaml in {}".format(self.extract_dir))
