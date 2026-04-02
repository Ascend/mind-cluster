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

from tests.st.lib.dl_deployer.dl import Installer


class DevicePluginInstaller(Installer):
    component_name = 'device-plugin'
    accelerator_labels = ["910"]

    def __init__(self, ip, username, password, resource_dir):
        super(DevicePluginInstaller, self).__init__(ip, username, password, resource_dir)
        self.all_yaml_files = []

    def get_yaml_path(self):
        yaml_files = []
        for root, _, files in os.walk(self.extract_dir):
            for filename in files:
                if filename.endswith('.yaml') and "1usoc" not in filename and "volcano" in filename:
                    yaml_files.append(os.path.join(root, filename))
        if not yaml_files:
            raise Exception('failed to find the yaml about volcano in {}'.format(self.extract_dir))
        self.all_yaml_files.extend(sorted(yaml_files, reverse=self.use_new_k8s))
        matching_yaml_files = []
        for line in self.iter_cmd_output('lspci'):
            if 'Processing accelerators' in line:
                if 'Device d500' in line:
                    substring = 'device-plugin-310P-'
                    matching_yaml_files = [file for file in yaml_files if substring in file]
                elif 'Device d100' in line or 'Device d107' in line:
                    substring = 'device-plugin-310-'
                    matching_yaml_files = [file for file in yaml_files if substring in file]
                elif 'Device d801' in line or 'Device d802' in line or 'Device d803' in line:
                    substring = 'device-plugin-volcano-'
                    matching_yaml_files = [file for file in yaml_files if substring in file]
        if not matching_yaml_files:
            matching_yaml_files.append(yaml_files[0])
        return matching_yaml_files[0]

    def apply_yaml(self):
        if not os.path.exists(self.yaml_dir):
            os.makedirs(self.yaml_dir, 0o755)
        accelerator_labels = self.accelerator_labels
        for yaml_file in self.all_yaml_files:
            device_met = False
            for device_type in accelerator_labels:
                if device_type == "910" and "device-plugin-volcano" in yaml_file:
                    device_met = True
                    break
            if not device_met:
                continue
            basename = os.path.basename(yaml_file)
            blank_yaml_path = os.path.join(self.yaml_dir, basename)
            with open(blank_yaml_path, 'w') as f:
                f.writelines(self.get_modified_yaml_contents())
            self.clear_previous_pod(blank_yaml_path)
            cmd = 'kubectl apply -f {}'.format(blank_yaml_path)
            self.module.execute_command(cmd)
            self.module.logger.info('apply yaml: {} for component: {}'.format(blank_yaml_path, self.component_name))

