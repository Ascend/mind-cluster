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
import time

from tests.st.st_dev.K8sTool import K8sTool


class CaseRoutines(object):

    @staticmethod
    def check_driver_status(case):
        driver_cmd = "npu-smi info"
        npu_infos = case.k8s_manager.exec_command(driver_cmd)
        if npu_infos and "dcmi module initialize failed" in npu_infos:
            return False
        else:
            return True

    @staticmethod
    def check_mind_cluster(case):
        dls = ["ascend-operator", "clusterd", "noded", "device-plugin", "volcano"]
        case.logger.info("checking mindcluster component status %s" % dls)
        K8sTool.check_pods_status(case, dls)
        return CaseRoutines.check_driver_status(case)

    @staticmethod
    def get_pod_name_and_rank_index(case, job_name, namespace):

        pod_names = case.k8s_manager.exec_command(
            "kubectl get pod -A -owide | grep %s | awk '{print $2}'" % job_name).splitlines()
        train_dict = {}
        for pod_name in pod_names:
            ret = case.k8s_manager.exec_command("kubectl get po -n %s %s -ojson | egrep 'hccl/rankIndex'" %
                                                (namespace, pod_name)).strip(",")
            key = [i.strip() for i in ret.replace('"', '').split(":")][1]
            train_dict[key] = pod_name
        return train_dict

    @staticmethod
    def clean_dirty_data(raw_data: str) -> str:
        dirty_data = ['[root', 'Error from server', 'root@']
        data = raw_data
        for item in dirty_data:
            if item in raw_data:
                data = data[:data.find(item)]

        return data

    @staticmethod
    def check_pod_in_train_iters(case, rank_index: str, job_name="default-test", time_interval=20, times=5,
                                 is_mindio=False):
        """
        check either pod is in iteration in train cases
        :param case: testcase object
        :param rank_index: rank index
        :param job_name: job name
        :param time_interval: check interval
        :param times: times to check
        :param is_mindio: is mindio situation
        :return:
        """
        namespace = case.k8s_manager.exec_command(
            "kubectl get pod -A -owide | grep %s | awk '{print $1}'" % job_name).splitlines()[0]
        train_dict = CaseRoutines.get_pod_name_and_rank_index(case, job_name, namespace)

        for i in range(times):
            case.logger.info("current checking times: %s" % int(i + 1))
            if is_mindio:
                ret = case.k8s_manager.exec_command("kubectl logs -n %s %s | grep \"initialize success\""
                                                    % (namespace, train_dict[rank_index]))
                ret = CaseRoutines.clean_dirty_data(ret)
                if ret == "":
                    time.sleep(time_interval)
                    continue

                ret = case.k8s_manager.exec_command("kubectl logs -n %s %s | grep -E \"Connect to.*success\""
                                                    % (namespace, train_dict[rank_index]))
                ret = CaseRoutines.clean_dirty_data(ret)
                if ret == "":
                    time.sleep(time_interval)
                    continue
            else:
                ret = case.k8s_manager.exec_command("kubectl logs -n %s %s | grep -E \"iteration\""
                                                    % (namespace, train_dict[rank_index]))
                ret = CaseRoutines.clean_dirty_data(ret)
                if ret == "":
                    time.sleep(time_interval)
                    continue
            case.logger.info("memory allocated done: maybe iteration has started ...")
            return True

        case.k8s_manager.exec_command("kubectl logs -n %s %s --tail 30" % (namespace, train_dict[rank_index]))
        return True
