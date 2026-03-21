#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import os
import platform
import re
import time

from tests.st.lib.common.Ssh import ClassSsh
from tests.st.lib.common.Ssh import STD_OUT


class ClassCLI(ClassSsh):

    def __init__(self, ip, username, password):
        super(ClassCLI, self).__init__(ip, username, password)
        self.wait_str = "root@"

    def execute_command(self, command, timeout=300, waitstr=None, path=None, inputList=None):
        inputList = inputList if inputList else []
        if waitstr is None:
            wait_str = self.wait_str
        if path is None:
            pass
        else:
            command_cd = {
                "command": ["cd", path],
                "waitstr": waitstr,
                "timeout": timeout,
                "input_list": inputList
            }
            self.cmd(command_cd)
        command = {
            "command": [command],
            "waitstr": waitstr,
            "timeout": timeout,
            "input_list": inputList
        }
        result = self.cmd(command)
        if not result:
            return result
        comp = re.compile(r'\x1b[@-_][0-?]*[ -/]*[@-~]]')
        result['stdout'] = comp.sub('', result['stdout']).strip()
        return result

    def get_cmd_result(self):
        command = {"command": ["echo $?"]}
        ret = self.cmd(command)
        return ret[STD_OUT]

    def wait_for_reboot(self, ip_addr=None, wait_time=600, timout=20, pack=3, type_par=""):
        os_type = platform.system()
        if not ip_addr:
            ip_addr = self.ip
        if os_type == "Linux":
            cmd_str = "ping {} -c {}".format(ip_addr, pack)
        elif os_type == "Windows":
            raise Exception("wrong os")
        startTime = time.time()
        endTime = startTime + wait_time
        count = 1
        while time.time() <= endTime:
            self.logger.info("ping {} ,wait time {}".format(ip_addr, str(time.time() - startTime)))
            ping_info = os.popen(cmd_str)
            ping_list = list(filter(lambda line: "ms" in line.lower() and "ttl" in line.lower(),
                                    ping_info.readlines()))
            if len(ping_list) >= int(pack):
                self.logger.info("{} connected".format(ip_addr))
                break
            else:
                count += 1
                time.sleep(5)
        else:
            self.logger.warning("out of time connected")
            return False
        for _ in range(5):
            result = self.login(type_par)
            if result:
                self.logger.info("login {} success".format(ip_addr))
                res = self.syn_system_time()
                if res:
                    self.logger.info("Success in synchronizing time")
                else:
                    self.logger.warning("failed to synchronizing time")
        else:
            self.logger.warning("login {} failed".format(ip_addr))
            return False

    def syn_system_time(self):
        modify_time = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime())
        res = self.cmd({"command": ["date", "-s", "'" + modify_time + "'"]})
        from tests.st.lib.common.Ssh import RESULT_CODE
        return res[RESULT_CODE] == 0

    def view_file_directory(self, folder, timeout=60, wait_str=""):
        command = {"command": ["ll %s| awk '{print$9}'" % folder],
                   "wait_str": wait_str,
                   "timeout": timeout}
        ret = self.cmd(command)
        if "No Such file or directory" not in ret['stdout']:
            ls_file_lists = ret.get('stdout').split('\r\n')
            self.logger.info("all the files in the folder are %s" % ls_file_lists)
            return ls_file_lists
        else:
            self.logger.warning("not find the folder")
            return None

    def delete_dir_path(self, dir_f, wait_str=""):
        if wait_str is None:
            wait_str = self.wait_str
        command = {'command': ["rm -rf %s" % dir_f],
                   'waitstr': wait_str}
        self.cmd(command)
        command = {"command": ["cd %s" % dir_f]}
        ret = self.cmd(command)['stdout']
        if "No Such file or directory" not in ret['stdout']:
            self.logger.warning("delete success")
            return True
        else:
            self.logger.warning("delete failed")
            return None

    def get_wait_str(self):
        return self.wait_str
