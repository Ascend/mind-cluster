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
import time

from tests.st.lib.dl_deployer.dl import Installer

logrotate_content = """/var/log/mindx-dl/volcano-*/*.log{
   daily
   rotate 8
   size 50M
   compress
   dateext
   missingok
   notifempty
   copytruncate
   create 0640 hwMindX hwMindX
   sharedscripts
   postrotate
       chmod 640 /var/log/mindx-dl/volcano-*/*.log
       chmod 440 /var/log/mindx-dl/volcano-*/*.log-*
   endscript
}"""


class VolcanoInstaller(Installer):
    component_name = 'volcano'

    def get_yaml_path(self):
        yaml_files = []
        for root, _, files in os.walk(self.extract_dir):
            for filename in files:
                if filename.endswith('.yaml'):
                    yaml_files.append(os.path.join(root, filename))
        if not yaml_files:
            raise Exception('failed to find yaml in {}'.format(self.extract_dir))
        return sorted(yaml_files, reverse=self.use_new_k8s)[0]

    def docker_build_with_retry(self, tag, docker_file_name, build_dir, max_retries=3, retry_delay=20):
        for attempt in range(max_retries):
            out = self.module.execute_command('docker build -q -t {} -f {} .'.format(tag, docker_file_name),
                                              path=build_dir)
            if not out:
                time.sleep(retry_delay)
            else:
                self.module.logger.info("Docker build successful on attempt {}".format(attempt + 1))
                return

    def build_images(self):
        build_dir = os.path.dirname(self.get_yaml_path())
        for tag, save_name in self.images.items():
            full_tag = self.module.ip + ":{}/".format(self.registry_port) + tag
            self.image_tags.append(full_tag)
            docker_file_name = 'Dockerfile-scheduler'
            if 'controller' in tag:
                docker_file_name = 'Dockerfile-controller'
            try:
                self.docker_build_with_retry(full_tag, docker_file_name, build_dir)
            except Exception as e:
                raise e
            self.module.execute_command('docker save -o {} {}'.format(save_name, full_tag), path=build_dir)
