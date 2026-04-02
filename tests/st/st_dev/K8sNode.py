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

from tests.st.lib.common.CLI import ClassCLI


class K8sNode(ClassCLI):

    def __init__(self, ip, username, password):
        super().__init__(ip, username, password)
        self.node_name = ""
        self.status = ""
        self.role = ""
        self.version = ""
        self.SSH_connect = None
        self.SFTP_connect = None
        self.npu_type = ""
        self.__init_device(ip, username, password)
        self.__init_sftp(ip, username, password)
        self.logger = logging.getLogger("k8s-mindcluster")
        logging.basicConfig(level=logging.DEBUG, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')

    def exec_command(self, cmd, path="", waitstr=None, timeout=30, inputList=None):
        err_str = "stderr"
        out_str = "stdout"
        ret = self.SSH_connect.execute_command(cmd, path=path, waitstr=waitstr, timeout=timeout, inputList=inputList)
        if err_str in ret and ret['rc'] != 0:
            self.logger.warning("command: %s => %s" % (cmd, ret[err_str]))
        return ret[out_str] if out_str in ret else None

    def execute_command(self, cmd, path="", waitstr=None, timeout=30, inputList=None):
        ret = self.SSH_connect.execute_command(cmd, path=path, waitstr=waitstr, timeout=timeout, inputList=inputList)
        return ret

    def net_down_and_recover(self, interval_time=5, node_ip=""):
        network_name = self.exec_command("ip route | grep %s | awk -F '[ \\t*]' '{print $3}'" % node_ip)
        self.exec_command(f"ifconfig {network_name} down && sleep {interval_time} && ifconfig {network_name} up")

    def get_wait_str(self):
        return self.SSH_connect.get_wait_str()

    def sftp_transfer_folder(self):
        pass

    def __init_device(self, ip, username, password):
        SSH_connect = ClassCLI(ip, username, password)
        SSH_connect.login()
        self.SSH_connect = SSH_connect

    def __init_sftp(self, ip, username, password):
        self.SFTP_connect = None

    @staticmethod
    def set_accelerator_type(case, node_name, node_num, accelerator_type):
        # Set accelerator-type label for all kwok nodes
        cmd = (f"kubectl label node {' '.join(f'kwok-node-{node_name}-{i}' for i in range(node_num))} "
               f"accelerator-type={accelerator_type} --overwrite")
        case.k8s_manager.exec_command(cmd)

    @staticmethod
    def get_node_by_pod_name(case, pod_name, namespace="default"):
        cmd = f"kubectl get pod {pod_name} -n {namespace} -owide --no-headers | awk '{{print $7}}'"
        return case.k8s_manager.exec_command(cmd).strip()



