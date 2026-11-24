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
import re
import time

from tests.st.lib.dl_deployer.dl import Installer

VOLCANO_YAML_RE = re.compile(r'^volcano-v(\d+)\.(\d+)\.(\d+)\.yaml$')

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
        """pick the volcano deployment yaml for the cluster k8s version.

        New k8s picks the newest volcano version; old k8s the oldest; prefer openeuler variant.
        """
        yaml_files = []
        for root, _, files in os.walk(self.extract_dir):
            for filename in files:
                if VOLCANO_YAML_RE.match(filename):
                    yaml_files.append(os.path.join(root, filename))
        if not yaml_files:
            raise RuntimeError('failed to find volcano yaml in {}'.format(self.extract_dir))

        def version_of(path):
            match = VOLCANO_YAML_RE.match(os.path.basename(path))
            if not match:
                raise RuntimeError('invalid volcano yaml name: {}'.format(path))
            return tuple(int(part) for part in match.groups())

        def is_openeuler_variant(path):
            relative_dir = os.path.relpath(os.path.dirname(path), self.extract_dir)
            return 'openeuler' in [part.lower() for part in relative_dir.split(os.sep)]

        yaml_files.sort(key=version_of)
        target_version = version_of(yaml_files[-1]) if self.use_new_k8s else version_of(yaml_files[0])
        same_version_files = [path for path in yaml_files if version_of(path) == target_version]
        openeuler_files = [path for path in same_version_files if is_openeuler_variant(path)]
        if openeuler_files:
            return openeuler_files[0]
        self.module.logger.warning(
            'volcano %s has no openeuler variant in %s, fallback to %s',
            target_version,
            self.extract_dir,
            same_version_files[0],
        )
        return same_version_files[0]

    def docker_build_with_retry(self, tag, docker_file_name, build_dir, max_retries=3, retry_delay=20):
        for attempt in range(max_retries):
            out = self.module.execute_command(
                'docker build -q -t {} -f {} .'.format(tag, docker_file_name), path=build_dir
            )
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
