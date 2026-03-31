#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import logging
import time

logger = logging.getLogger(__name__)


class ClusterSimulator(object):
    @staticmethod
    def create_kwok_cluster(case, node_name, node_num):
        case.k8s_manager.exec_command(
            f"docker run -d -v /root/.kube/config:/root/.kube/config --rm cluster_simulator:v1.7 simulate {node_name} --node_num {node_num}")
        # Wait for nodes to be created completely
        time.sleep(5)

    @staticmethod
    def delete_kwok_cluster(case):
        case.k8s_manager.exec_command(
            f"docker run -d -v /root/.kube/config:/root/.kube/config --rm cluster_simulator:v1.7 cleanup")
        # Wait for nodes to be created completely
        time.sleep(5)

    @staticmethod
    def inject_kwok_software_fault(case, namespace, pod_name):
        case.k8s_manager.exec_command(f"kubectl label pod {pod_name} -n {namespace} software-fault=occur")

    @staticmethod
    def inject_kwok_hardware_fault(case, pod_name):
        case.k8s_manager.exec_command(f"bash inject_fault_for_910a.sh {pod_name}")

    @staticmethod
    def get_ready_kwok_node_count(case):
        ready_kwok_node_cmd = "kubectl get nodes | grep kwok-node | grep Ready | wc -l"
        ready_node_count = int(case.k8s_manager.exec_command(ready_kwok_node_cmd))
        return ready_node_count

    @staticmethod
    def get_kwok_nodes_with_accelerator_type(case, accelerator_type="module-910b-8"):
        # Get the number of kwok nodes with specified accelerator-type label within 10 seconds
        cmd = f"kubectl get nodes -l accelerator-type={accelerator_type} | grep kwok-node | wc -l"
        start_time = time.time()
        timeout = 10

        while time.time() - start_time < timeout:
            try:
                node_count = int(case.k8s_manager.exec_command(cmd))
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

    # 设置超节点大小，参数为框的数目，并且让原有节点不可调度
    @staticmethod
    def start_cluster_simulator(case, *superpod_sizes: int):
        total_nodes = sum(superpod_sizes) * 8
        superpod_sizes = ' '.join(list(map(str, superpod_sizes)))
        ClusterSimulator.stop_cluster_simulator(case)
        res = case.k8s_manager.master.exec_command(f"docker run -d  --name my_container -v /root/.kube/config:/root/.kube/config --rm \
                                        cluster_simulator:v1.7 simulate davidx8superpod --super_pod_sizes \
                                        {superpod_sizes}")

        if "Unable to find image" in res:
            raise Exception("cluster_simulator 镜像不存在！")

        kwok_nodes, i = 0, 0
        while i < 120:
            i += 1
            kwok_nodes = int(case.k8s_manager.master.exec_command(f"kubectl get nodes | grep kwok | wc -l"))
            if kwok_nodes >= total_nodes:
                break
            time.sleep(1)
        if kwok_nodes < total_nodes:
            raise Exception(f"kwok nodes less than {total_nodes}")

        case.k8s_manager.master.exec_command(
            "kubectl get nodes | awk {'print $1'} | grep work | xargs -I {}  kubectl cordon {}")

    @staticmethod
    def stop_cluster_simulator(case):
        case.k8s_manager.master.exec_command("docker rm -f my_container")
        case.k8s_manager.master.exec_command(
            "docker run  --rm -v /root/.kube/config:/root/.kube/config cluster_simulator:v1.7 cleanup")
        case.k8s_manager.master.exec_command(
            "kubectl get nodes | awk {'print $1'} | grep work | xargs -I {}  kubectl uncordon {}")
