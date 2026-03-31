#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import logging
import os
import time

logger = logging.getLogger(__name__)
class K8sTool(object):

    @staticmethod
    def check_pods_status(case, names, status="Running", timeout=30):
        for name in names:
            assert K8sTool.check_pod_status(case, name, status, timeout), "pod %s not all %s" % (name, status)

        return True

    @staticmethod
    def check_pod_status(case, pod_name, status="Running", timeout=30):
        is_all_status = False
        status_cmd = "kubectl get pods -A | grep %s | awk '{print$4}'" % pod_name
        cur_time = time.time()
        while time.time() - cur_time < timeout:
            ret = case.k8s_manager.exec_command(status_cmd)
            for pod_status in ret.splitlines():
                if pod_status != status:
                    time.sleep(5)
                    break
            else:
                is_all_status = True
                break
        return is_all_status

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
    def find_volcano_yaml(k8s_manager):
        yaml = k8s_manager.master.exec_command(
            f'find {os.environ.get("MIND_CLUSTER_YAML_DIR")} -name "volcano-*.yaml"')
        if not yaml:
            yaml = k8s_manager.master.exec_command(f'find /home/update/yamls -name "volcano-*.yaml"')
            if not yaml:
                raise Exception("未找到volcano组件yaml！")
        return yaml

    @staticmethod
    def modify_volcano_yaml(case, super_pod_size="512", useClusterInfoManager="false"):
        logger.info("修改volcano yaml配置")
        volcano_yaml_path = K8sTool.find_volcano_yaml(case.k8s_manager)
        if super_pod_size is not None:
            case.k8s_manager.master.exec_command(
                f"sed -i 's/\"super-pod-size\": \"[0-9]\\+\"/\"super-pod-size\": \"{super_pod_size}\"/g'"
                f" \"{volcano_yaml_path}\"")
        if useClusterInfoManager is not None:
            case.k8s_manager.master.exec_command(
                f"sed -i 's/\"useClusterInfoManager\":\"\\(false\\|true\\)\"/\"useClusterInfoManager\":\"{useClusterInfoManager}\"/g'"
                f" \"{volcano_yaml_path}\"")
        K8sTool.restart_volcano(case.k8s_manager, volcano_yaml_path)

    @staticmethod
    def reset_volcano_yaml(case):
        volcano_yaml_path = K8sTool.find_volcano_yaml(case.k8s_manager)
        case.k8s_manager.master.exec_command(f"sed -i 's/\"super-pod-size\": \"[0-9]\\+\"/"
                                             f"\"super-pod-size\": \"48\"/g' \"{volcano_yaml_path}\"")
        case.k8s_manager.master.exec_command(f"sed -i 's/\"useClusterInfoManager\":\"\\(false\\|true\\)\"/"
                                             f"\"useClusterInfoManager\":\"true\"/g' \"{volcano_yaml_path}\"")
        K8sTool.restart_volcano(case.k8s_manager, volcano_yaml_path)

    @staticmethod
    def fault_inject(case):
        case.k8s_manager.master.exec_command(f"bash {case._fault_inject_file_path} --fault_pod_name "
                                             f"{case._fault_pod} --card_num 8 "
                                             f"--card_unhealthy \"npu-0\"")

    @staticmethod
    def restart_volcano(k8s_manager, volcano_yaml_path):
        logger.info("重启volcano")
        k8s_manager.master.exec_command(f'kubectl delete -f {volcano_yaml_path}')
        time.sleep(5)
        k8s_manager.master.exec_command(f'kubectl apply -f {volcano_yaml_path}')

    @staticmethod
    def apply_mindcluster(case, yaml_path="/home/update/yamls"):
        case.k8s_manager.master.exec_command("kubectl create ns mindx-dl && kubectl create ns volcano-system && kubectl create ns cluster-system")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f device-plugin-npu-volcano-*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f ascend-operator-*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f volcano-*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f clusterd-*.yaml")
        case.k8s_manager.exec_command(f"cd {yaml_path} && kubectl apply -f noded-*.yaml")
