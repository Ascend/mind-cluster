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
import logging
import time
from typing import Dict

logger = logging.getLogger(__name__)

def wait_until(condition_fn, timeout=60, interval=5):
    start_time = time.time()
    while time.time() - start_time < timeout:
        if condition_fn():
            return
        time.sleep(interval)


class JobHelper(object):

    @staticmethod
    def delete_job(case, job_name=None):
        logger.info(f"删除任务{job_name}")
        acjobs = case.k8s_manager.master.exec_command("kubectl get acjob  --no-headers | awk '{print $1}'")
        for job in acjobs.splitlines():
            case.k8s_manager.master.exec_command(f"kubectl delete acjob {job}")
        if job_name:
            case.k8s_manager.master.exec_command(f"kubectl delete configmap -n default reset-config-{job_name}")
            wait_until(lambda: case.k8s_manager.master.exec_command(
                f"kubectl get pods -n default -l job-name={job_name} -o wide") == "No resources found in default namespace.",
                       timeout=180)

    @staticmethod
    def check_job_pods_all_running(case, job_name: str, pod_num: int, timeout=60):
        logger.info(f"查看训练任务{job_name}是否Running")
        cur_time = time.time()
        while time.time() - cur_time < timeout:
            res = case.k8s_manager.master.exec_command(
                f"kubectl get pods -n default -l job-name={job_name} -o jsonpath='{{.items[*].status.phase}}{{\"\\n\"}}'")
            pods_status = res.split()
            if len(pods_status) == pod_num and all(s == "Running" for s in pods_status):
                case.k8s_manager.master.exec_command(f"kubectl get pods -n default -l job-name={job_name} -o wide")
                return
            time.sleep(5)

        case.k8s_manager.master.exec_command(f"kubectl get pods -n default -l job-name={job_name} -o wide")
        case.k8s_manager.master.exec_command(f"kubectl describe pg")
        raise Exception(f"任务<{job_name}>的pod没有全部running！")

    @staticmethod
    def get_pod_node_mapping(case, job_name) -> Dict:
        logger.info("获取 pod name 和 node name 的映射关系")
        mapping = {}
        cmd = f"kubectl get pods -n default -l job-name={job_name} -o=jsonpath='{{range .items[*]}}{{.metadata.name}} {{.spec.nodeName}}{{\"\\n\"}}{{end}}'  "
        res = case.k8s_manager.master.exec_command(cmd)
        for line in res.splitlines():
            if line.strip():
                parts = line.strip().split()
                if len(parts) == 2:
                    pod_name, node_name = parts
                    mapping[pod_name] = node_name
        return mapping
