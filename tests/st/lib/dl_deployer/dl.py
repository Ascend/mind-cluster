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
import platform
import re
import shutil
import zipfile

from tests.st.lib.common.CLI import ClassCLI

name_list = [
    'device-plugin',
    'ascend-operator',
    'noded',
    'npu-exporter',
    'volcano',
    'clusterd'
]


class Installer(object):
    component_name = ''
    user = 'hwMindX'
    user_id = 9000
    group = 'hwMindX'
    group_id = 9000
    namespace = 'mindx-dl'

    def __init__(self, client, resource_dir):
        self.image_tags = []
        self.component_installer = None
        self.module = client
        self.resources_dir = resource_dir
        self.package_name = ""
        self.arch = platform.machine()
        self.arch = "aarch64"
        self.dl_dir = os.path.join(os.path.dirname(self.resources_dir), 'mindxdl')
        self.package_dir = os.path.join('/workspace/', 'dlPackage', self.arch, self.component_name)
        self.extract_dir = os.path.join('/workspace/', 'dlDeployPackages', self.arch, self.component_name)
        self.yaml_dir = os.path.join(self.dl_dir, 'yamls', self.arch)
        self.dl_log = '/var/log/mindx-dl'
        self.registry_port = 5000
        self.use_new_k8s = True
        self.import_cmd = ''
        self.yaml_file_path = ''
        self.images = dict()
        self.facts = dict()

    def is_new_k8s_version(self):
        ret = self.module.execute_command('kubelet --version')
        if 'Kubernetes' not in ret['stdout']:
            raise Exception('failed to get kubelet version, ret:{}'.format(ret['stdout']))
        version = re.search(r'(?<=v)\d+\.\d+(\.\d+)?', ret['stdout']).group()
        version_tuple = tuple(map(int, version.split('.')))
        return version_tuple > (1, 19, 16)

    def get_yaml_path(self):
        """ pick the right yaml file and return file path """
        for root, _, files in os.walk(self.extract_dir):
            for filename in files:
                if filename.endswith('.yaml') and 'without' not in filename and '1usoc' not in filename:
                    return os.path.join(root, filename)
        raise Exception('failed to find yaml in {}'.format(self.extract_dir))

    def check_and_prepare(self):
        if self.component_name not in name_list:
            raise Exception('invalid component name, choice from {}'.format(name_list))
        self.use_new_k8s = self.is_new_k8s_version()
        src = ''
        if os.path.exists(self.package_dir):
            shutil.rmtree(self.package_dir)
        os.makedirs(self.package_dir, 0o755)
        if os.path.exists(self.package_dir):
            shutil.rmtree(self.package_dir)
        shutil.copytree(self.resources_dir, self.package_dir)
        src = os.path.join(self.package_dir, self.package_name)
        if os.path.exists(self.extract_dir):
            shutil.rmtree(self.extract_dir)
        with zipfile.ZipFile(src) as zf:
            zf.extractall(self.extract_dir)
        yaml_file = self.get_yaml_path()
        if not os.path.exists(yaml_file):
            raise Exception('failed to find yaml file: {}'.format(yaml_file))
        self.yaml_file_path = yaml_file

    def get_image_tags(self):
        keyword = 'image:'
        image_tags = []
        with open(self.yaml_file_path) as f:
            for line in f:
                if keyword in line and line.strip() != keyword:
                    # like"      - image: ascend-k8sdeviceplugin:v5.0.0"
                    image_tag = line.replace(keyword, '').replace(' - ', '').strip()
                    if ':' in image_tag:
                        image_tags.append(image_tag)
        if not image_tags:
            raise Exception('failed to find image name in file: {}'.format(self.yaml_file_path))
        return image_tags

    def build_images(self):
        build_dir = os.path.dirname(self.yaml_file_path)
        for tag, save_name in self.images.items():
            full_tag = self.module.ip + ":{}/".format(self.registry_port) + tag
            self.image_tags.append(full_tag)
            self.module.execute_command('docker build -t {} .'.format(full_tag), path=build_dir)
            self.module.logger.info('build image file: {} in {} successfully'.format(save_name, build_dir))

    def is_images_exists(self):
        for image_tag in self.get_image_tags():
            image_name, image_version = image_tag.split(':')
            image_name = image_name.split('/')[-1]
            image_save_name = '{}_{}.tar'.format(image_name, self.arch)
            self.images[image_tag] = image_save_name
        image_path_list = []
        exist = True
        for save_name in self.images.values():
            image_path = os.path.join(self.yaml_file_path, save_name)
            image_path_list.append(image_path)
            if not os.path.exists(image_path):
                exist = False
        return exist

    def build(self):
        self.check_and_prepare()
        if self.is_images_exists():
            self.module.logger.warning('images already exists, skip build')
            return
        self.build_images()

    def push(self):
        self.module.logger.info("push the image to %s" % self.module.ip)
        for image_tag in self.image_tags:
            self.module.execute_command("docker push %s" % image_tag, path=self.extract_dir)

    def ensure_group_exist(self):
        cmd = 'groupmod -g {} {}'.format(self.group_id, self.group)
        self.module.execute_command(cmd)

    def create_log_dir(self):
        """ do jobs such as creating log dir and logrotate file """
        log_dir_names = (self.component_name,)
        for log_dir in log_dir_names:
            log_path = os.path.join(self.dl_log, log_dir)
            if not os.path.exists(log_path):
                os.makedirs(log_path, 0o750)
                os.chown(log_path, self.user_id, self.group_id)

    def install(self):
        if not os.path.exists(self.dl_log):
            os.makedirs(self.dl_log, 0o755)
        self.ensure_group_exist()
        self.create_log_dir()

    def get_modified_yaml_contents(self):
        lines = self._get_yaml_contents()
        for index, line in enumerate(lines):
            if "image: " in line:
                replace_line = line.replace("image: ",
                                            "image: {}:{}/".format(self.module.ip, self.registry_port))
                lines[index] = replace_line
            if "imagePullPolicy:" in line:
                lines[index] = line.replace("imagePullPolicy: Never", "imagePullPolicy: IfNotPresent")
        return lines

    def create_namespace(self):
        cmd = 'kubectl create namespace {}'.format(self.namespace)
        self.module.execute_command(cmd)
        self.module.logger.info('create namespace: {} for component: {}'.format(self.namespace, self.component_name))
        if self.component_name == 'npu-exporter':
            self.module.execute_command('kubectl create namespace npu-exporter')

    def clear_previous_pod(self, yaml_path):
        cmd = 'kubectl delete -f {}'.format(yaml_path)
        self.module.execute_command(cmd)

    def apply_yaml(self):
        if not os.path.exists(self.yaml_dir):
            os.makedirs(self.yaml_dir, 0o755)
        yaml_path = os.path.join(self.yaml_dir, os.path.basename(self.yaml_file_path))
        with open(yaml_path, 'w') as f:
            f.writelines(self.get_modified_yaml_contents())
        self.clear_previous_pod(yaml_path)
        cmd = 'kubectl apply -f {}'.format(yaml_path)
        self.module.execute_command(cmd)
        self.module.logger.info('apply yaml: {} for component: {}'.format(yaml_path, self.component_name))

    def apply(self):
        self.create_namespace()
        self.apply_yaml()

    def run(self):
        steps = {
            'build': self.build,
            'push': self.push,
            'install': self.install,
            'apply': self.apply
        }

        steps.get(self.step)()

    def _get_yaml_contents(self):
        with open(self.yaml_file_path) as f:
            return f.readlines()
