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
import re
import time

from tests.st.st_dev.K8sTool import K8sTool

volcano_prefix = "volcano-v1.9.0"


class MindClusterHandler:

    @staticmethod
    def restart_dp(k8s_manager, yaml_path, device_plugin_yaml_prefix="device-plugin-volcano"):
        k8s_manager.exec_command("kubectl get nodes --no-headers | grep work | awk '{print $1}' | xargs -I {} kubectl "
                                 "delete cm -n kube-system mindx-dl-deviceinfo-{}")
        k8s_manager.exec_command('find %s -name "%s*.yaml" | xargs -I {} kubectl delete -f {}'
                                 % (yaml_path, device_plugin_yaml_prefix))
        k8s_manager.exec_command('find %s -name "%s*.yaml" | xargs -I {} kubectl apply -f {}'
                                 % (yaml_path, device_plugin_yaml_prefix))

    @staticmethod
    def restart_volcano(k8s_manager, yaml_path):
        k8s_manager.exec_command('find %s -name "%s*" | xargs -I {} kubectl delete -f {}' % (yaml_path, volcano_prefix))
        time.sleep(5)
        k8s_manager.exec_command('find %s -name "%s*" | xargs -I {} kubectl apply -f {}' % (yaml_path, volcano_prefix))

    @staticmethod
    def check_is_dp_service_allocatable(case, work, timeout=30):
        cur_time = time.time()
        while time.time() - cur_time < timeout:
            real_key = False
            node_info = case.k8s_manager.exec_command("kubectl describe node %s" % work)
            for line in node_info.splitlines():
                if "Allocatable" in line:
                    real_key = True
                if real_key and re.match(r"\s+huawei.com/Ascend\s+:\s+\b[1-9]]\d*\b", line):
                    return True
            time.sleep(2)
        return False

    @staticmethod
    def check_dp_service(case):
        assert K8sTool.all_worker_execute_func(case, MindClusterHandler.check_is_dp_service_allocatable), "dp start failed"
