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
import json
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
        logger.info(f"Deleting job {job_name}")
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
        logger.info(f"Checking if training job {job_name} is Running")
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
        raise Exception(f"Not all pods of job <{job_name}> are running!")

    @staticmethod
    def get_pod_node_mapping(case, job_name) -> Dict:
        logger.info("Getting mapping relationship between pod name and node name")
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

    @staticmethod
    def check_pod_label_exist(case, pod_name, label_name, ns="default") -> bool:
        cmd = "kubectl get pod -n {} {} --show-labels | grep -o \"{}\"".format(ns, pod_name, label_name)
        res = case.k8s_manager.master.exec_command(cmd)
        if label_name in res:
            return True
        return False

    @staticmethod
    def get_server_count_from_ranktable(case, ranktable_path):
        """
        Get server_count from hccl.json and return as integer
        """
        cmd = f"cat {ranktable_path}"
        res = case.k8s_manager.master.exec_command(cmd)

        if not res:
            print("Error: Unable to get file content or content is empty")
            return 0

        try:
            data = json.loads(res)
            server_count = int(data.get("server_count", 0))
            return server_count

        except json.JSONDecodeError as e:
            print(f"Error: JSON parsing failed - {e}")
            return 0
        except (ValueError, TypeError) as e:
            print(f"Error: server_count format conversion failed - {e}")
            return 0


