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

from tests.st.envs import WAIT_TIME

logger = logging.getLogger(__name__)


class ClusterSimulator(object):
    @staticmethod
    def create_kwok_cluster(case, container_name, node_name, node_num):
        case.k8s_manager.exec_command(
            f"docker run -d --name {container_name} -v /root/.kube/config:/root/.kube/config "
            f"-v /workspace/mind-cluster/tests/st/specs/:/cluster_simulator/specs "
            f"--rm cluster_simulator:base simulate {node_name} --node_num {node_num}")
        time.sleep(WAIT_TIME)

    @staticmethod
    def create_kwok_cluster_a3(case, container_name, node_name, super_pod_num, super_pod_size):
        case.k8s_manager.exec_command(
            f"docker run -d --name {container_name} -v /root/.kube/config:/root/.kube/config "
            f"-v /workspace/mind-cluster/tests/st/specs/:/cluster_simulator/specs --rm cluster_simulator:base "
            f"simulate {node_name} --super_pod_num {super_pod_num} --super_pod_size {super_pod_size}")
        time.sleep(WAIT_TIME)

    @staticmethod
    def stop_kwok_cluster(case, container_name):
        case.k8s_manager.master.exec_command(f"docker rm -f {container_name}")
        case.k8s_manager.exec_command(
            f"docker run -d -v /root/.kube/config:/root/.kube/config --rm cluster_simulator:base cleanup")
        time.sleep(WAIT_TIME)

    @staticmethod
    def inject_kwok_software_fault(case, namespace, pod_name):
        case.k8s_manager.exec_command(f"kubectl label pod {pod_name} -n {namespace} software-fault=occur")

    @staticmethod
    def inject_kwok_hardware_fault(case, pod_name):
        case.k8s_manager.exec_command(f"bash inject_fault_for_910a.sh {pod_name}")

    @staticmethod
    def get_ready_kwok_node_count(case):
        ready_kwok_node_cmd = "kubectl get nodes | grep kwok-node | grep Ready | wc -l"
        ready_node_count_str = case.k8s_manager.exec_command(ready_kwok_node_cmd)
        ready_node_count = int(ready_node_count_str.splitlines()[0])
        return ready_node_count

    @staticmethod
    def get_kwok_nodes_with_accelerator_type(case, accelerator_type="module-910b-8"):
        # Get the number of kwok nodes with specified accelerator-type label within 10 seconds
        cmd = f"kubectl get nodes -l accelerator-type={accelerator_type} | grep kwok-node | wc -l"
        start_time = time.time()
        timeout = 10

        while time.time() - start_time < timeout:
            try:
                node_count_str = case.k8s_manager.exec_command(cmd)
                node_count = int(node_count_str.splitlines()[0])
                if node_count > 0:
                    return node_count
            except (ValueError, Exception) as e:
                case.logger.warning(f"Error getting kwok nodes with accelerator-type {accelerator_type}: {e}")
            time.sleep(2)

        # Return the final count after timeout
        try:
            return int(case.k8s_manager.exec_command(cmd))
        except (ValueError, Exception) as e:
            case.logger.error(f"Failed to get kwok nodes with accelerator-type {accelerator_type} after timeout: {e}")
            return 0
