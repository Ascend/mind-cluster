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

from tests.st.envs import MIND_CLUSTER_YAML_DIR

logger = logging.getLogger(__name__)


class K8sTool(object):

    @staticmethod
    def check_pods_status(case, names, status="Running", timeout=30):
        for name in names:
            assert K8sTool.check_pod_status(case, name, status, timeout), "pod %s not all %s" % (name, status)

        return True

    @staticmethod
    def check_pod_status(case, pod_name, status=None, timeout=30):
        if status is None:
            status = ["Running"]
        elif isinstance(status, str):
            status = [status]
        is_all_status = False
        status_cmd = "kubectl get pods -A | grep %s | awk '{print $4}'" % pod_name
        cur_time = time.time()
        while time.time() - cur_time < timeout:
            ret = case.k8s_manager.exec_command(status_cmd)
            logger.info(f"{pod_name}--pod_status: {ret}")
            if not ret or not ret.strip():
                logger.info(f"Warning: No pods found for {pod_name}")
                time.sleep(5)
                continue
            for pod_status in ret.splitlines():
                if pod_status not in status:
                    time.sleep(5)
                    break
            else:
                is_all_status = True
                break
        return is_all_status

    @staticmethod
    def check_all_pods_status(case, pod_names=None, status=None, timeout=30, check_interval=2):
        if not pod_names:
            logger.error("error: pod_names should not be nil!")
            return False
        elif isinstance(pod_names, str):
            pod_names = [pod_names]

        if status is None:
            expected_status = ["Running"]
        elif isinstance(status, str):
            expected_status = [status]

        logger.info(
            f"start check pod status | target pod: {pod_names} | expected status: {expected_status} | timeout: {timeout}s")

        pod_commands = {}
        for pod in pod_names:
            cmd = f"kubectl get pods -A | grep -w {pod} | awk '{{print $4}}'"
            pod_commands[pod] = cmd

        start_time = time.time()
        while time.time() - start_time < timeout:
            all_pods_ok = True
            for pod_name, cmd in pod_commands.items():
                try:
                    ret = case.k8s_manager.exec_command(cmd)
                    ret = ret.strip()
                    logger.info(f"[{pod_name}] current status: {ret if ret else 'not found'}")
                    if not ret:
                        logger.warning(f"[{pod_name}] not found pod, wait...")
                        all_pods_ok = False
                        break

                    if ret not in expected_status:
                        logger.warning(
                            f"[{pod_name}] status not expected, current status: {ret}, expected status: {expected_status}")
                        all_pods_ok = False
                        break

                except Exception as e:
                    logger.error(f"[{pod_name}] exec failed: {str(e)}")
                    all_pods_ok = False
                    break

            if all_pods_ok:
                cost_time = round(time.time() - start_time, 2)
                logger.info(f"all Pods have reached the desired state! Time taken: {cost_time}s")
                return True

            time.sleep(check_interval)

        logger.error(f"Check timed out! Not all Pods reached the desired state within {timeout}s: {expected_status}")
        return False

    @staticmethod
    def check_pg_info(case, pod_group_name, info, timeout=30):
        cur_time = time.time()
        while time.time() - cur_time < timeout:
            infos = case.k8s_manager.exec_command("kubectl describe pg %s | grep %s " % (pod_group_name, info))
            if infos and info in infos:
                case.logger.info("pg %s info is %s" % (pod_group_name, infos))
                return
            else:
                time.sleep(5)
                continue
        else:
            raise Exception("pg not exist")

    @staticmethod
    def check_acjob_status(case, pod_name, status='Pending', timeout=30):
        cur_time = time.time()
        namespace = case.k8s_manager.exec_command("kubectl get acjob -A| grep %s |awk '{print $1}'" % (pod_name))
        while time.time() - cur_time < timeout:
            state = case.k8s_manager.exec_command(
                "kubectl get acjob -n %s %s |awk '{print $1}'" % (namespace, pod_name))
            if status in state:
                return True
            else:
                time.sleep(5)
                continue
        else:
            raise Exception("the status of acjob is not %s" % status)

    @staticmethod
    def check_device_info_cm_fault_code(case, device_name, timeout=30):
        cur_time = time.time()
        while time.time() - cur_time < timeout:
            device_info_cm = case.k8s_manager.exec_command(
                "kubectl get acjob -A| grep %s |awk '{print $1}'" % (device_name))
            if device_info_cm and "fault_code" in device_info_cm:
                case.logger.info("fault inject success!")
                return device_info_cm
            else:
                time.sleep(5)
        return False

    @staticmethod
    def check_pod_env(case, job_name):
        operator_env = case.k8s_manager.exec_command("kubectl get pod %s -o jsonpath='{{.spec.containers[0].env}} && "
                                                     "ll'")
        for item in ("LOCAL_RANK", "MASTER_ADDR", "WORLD_SIZE", "LOCAL_WORLD_SIZE", "MASTER_PORT"):
            assert operator_env.find(item) != -1, case.logger.error("ascend-operator inject env failed")

    @staticmethod
    def check_pod_deleted(case, job_name, timeout=30):
        cur_time = time.time()
        while time.time() - cur_time < timeout:
            job_info = case.k8s_manager.exec_command("kubectl get pod -A| grep %s" % job_name)
            if job_name not in job_info:
                return True
            else:
                time.sleep(5)
        return False

    @staticmethod
    def check_all_device_available(case, work_num=16):
        for idx in range(1, 1 + work_num):
            device_name = "work%s" % idx
            device_info_configmap = case.k8s_manager.exec_command("kubectl get cm -n kube-system "
                                                                  "mindx-dl-deviceinfo-%s -o json" % device_name)
            if device_info_configmap and "fault_code" in device_info_configmap:
                case.logger.error("fault at work %s not recovered" % device_name)
                return False
            else:
                case.logger.info("fault all recovered")
                return True

    @staticmethod
    def all_worker_execute_func(case, func, *args, **kwargs):
        workers = case.k8s_manager.exec_command("kubectl get nodes | grep work | awk '{print$1}'")
        try:
            for worker in range(workers):
                func(case, worker, *args, **kwargs)
        except Exception as e:
            return False
        else:
            return True

    @staticmethod
    def apply_yaml_by_file(case, yaml_path):
        return case.k8s_manager.exec_command("kubectl apply -f %s" % yaml_path)

    @staticmethod
    def delete_yaml_by_file(case, yaml_path):
        return case.k8s_manager.exec_command("kubectl delete -f %s" % yaml_path)

    @staticmethod
    def find_volcano_yaml(case):
        yaml = case.k8s_manager.master.exec_command(
            f'find {MIND_CLUSTER_YAML_DIR} -name "volcano-*.yaml"')
        if not yaml:
            raise Exception("volcano component yaml not found!")
        return yaml

    @staticmethod
    def modify_volcano_yaml(case, super_pod_size="512", useClusterInfoManager="false"):
        logger.info("Modifying volcano yaml configuration")
        volcano_yaml_path = K8sTool.find_volcano_yaml(case)
        if super_pod_size is not None:
            case.k8s_manager.master.exec_command(
                f"sed -i 's/\"super-pod-size\": \"[0-9]\\+\"/\"super-pod-size\": \"{super_pod_size}\"/g'"
                f" \"{volcano_yaml_path}\"")
        if useClusterInfoManager is not None:
            case.k8s_manager.master.exec_command(
                f"sed -i 's/\"useClusterInfoManager\":\"\\(false\\|true\\)\"/\"useClusterInfoManager\":\"{useClusterInfoManager}\"/g'"
                f" \"{volcano_yaml_path}\"")
        K8sTool.restart_volcano(case.k8s_manager)

    @staticmethod
    def reset_volcano_yaml(case):
        volcano_yaml_path = K8sTool.find_volcano_yaml(case)
        case.k8s_manager.master.exec_command(f"sed -i 's/\"super-pod-size\": \"[0-9]\\+\"/"
                                             f"\"super-pod-size\": \"48\"/g' \"{volcano_yaml_path}\"")
        case.k8s_manager.master.exec_command(f"sed -i 's/\"useClusterInfoManager\":\"\\(false\\|true\\)\"/"
                                             f"\"useClusterInfoManager\":\"true\"/g' \"{volcano_yaml_path}\"")
        K8sTool.restart_volcano(case.k8s_manager)

    @staticmethod
    def fault_inject(case):
        case.k8s_manager.master.exec_command(f"bash {case._fault_inject_file_path} --fault_pod_name "
                                             f"{case._fault_pod} --card_num 8 "
                                             f"--card_unhealthy \"npu-0\"")

    @staticmethod
    def restart_volcano(k8s_manager):
        logger.info("Restarting volcano")
        k8s_manager.master.exec_command("kubectl delete pod -n volcano-system -l app=volcano-scheduler")
        time.sleep(5)

    @staticmethod
    def apply_mindcluster(case, yaml_path=MIND_CLUSTER_YAML_DIR):
        case.k8s_manager.exec_command("chmod 777 /user/mindx-dl")
        case.k8s_manager.exec_command(
            "kubectl create ns mindx-dl && kubectl create ns volcano-system && kubectl create ns cluster-system")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f device-plugin-volcano-v*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f ascend-operator-v*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f volcano-v*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f clusterd-v*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f noded-v*.yaml")

    @staticmethod
    def insert_software_fault(case, ns="default", pod_name=""):
        case.k8s_manager.exec_command(f"kubectl label pod -n {ns} {pod_name} software-fault=occur")

    @staticmethod
    def apply_mindcluster_v2(case, yaml_path=MIND_CLUSTER_YAML_DIR):
        case.k8s_manager.master.exec_command(
            "kubectl create ns mindx-dl && kubectl create ns volcano-system && kubectl create ns cluster-system")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl delete -f device-plugin-npu-volcano-*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f device-plugin-volcano-*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f ascend-operator-*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f volcano-*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f clusterd-*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f noded-*.yaml")
