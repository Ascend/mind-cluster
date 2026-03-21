#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import logging
import os

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
        if err_str in ret and ret[err_str] is not None:
            self.logger.warning(ret[err_str])
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



