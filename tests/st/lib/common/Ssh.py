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
import socket
from time import sleep, time
import traceback
import paramiko
from tests.st.envs import SSH_LOG_LEVEL


RESULT_CODE = "rc"
INPUT_LIST = "input"
STD_OUT = "stdout"
STD_ERR = "stderr"


class ClassSsh(object):
    BASIC_WAIT_STR = "[>#]"
    MSG_TOO_LONG = "MSG_TOO_LONG"
    SSH_SERVER_TYPE_LINUX = "Linux"
    SSH_SERVER_TYPE_WINDOWS = "Windows"
    RESULT_MSG_MAX_LEN = 10000
    CMD_SENDONLY_FLAG = "sendonly"

    def __init__(self, ip, username, password, port=22, password_hidden="***", **kwargs):
        self.logger = logging.getLogger("ssh-mindcluster")
        logging_level = SSH_LOG_LEVEL
        logging.basicConfig(level=logging_level, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
        self.ip = ip
        self.username = username
        self.password = password
        self.port = port
        self.linesep = '\n'
        self.__sshClient = None
        self.__channel = None
        self.sshServerType = ""
        self.DEFAULT_WAIT_STR = ClassSsh.BASIC_WAIT_STR

        self.RECV_TMOUT = False
        self.SEND_ONLY = False

        self.keyType = ""
        self.keyFile = ""
        if kwargs and len(kwargs) != 0:
            pass
        if password == "" and self.keyType == "":
            raise Exception("no passwd")

    def login(self, ssh_server_type=SSH_SERVER_TYPE_LINUX):
        self.sshServerType = ssh_server_type
        try:
            connectInfo = self.connect()
            self.__sshClient = connectInfo.get("client")
            self.__channel = connectInfo.get("channel")
            return connectInfo.get(STD_OUT)
        except Exception as e:
            self.logger.error(f"Login failed: {e}")
            return None

    def connect(self):
        self.logger.info(
            "Connect to %s" % self.ip
        )
        max_con = 5
        flag = False
        count = 0
        while count < max_con:
            count += 1
            try:
                ssh = paramiko.SSHClient()
                ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
                if self.password != "":
                    ssh.connect(self.ip, self.port, self.username, self.password, look_for_keys=False)
                elif self.keyType != "":
                    ssh.connect(self.ip,
                                self.port,
                                self.username,
                                key_filename=self.keyFile, look_for_keys=False)
                flag = True
            except (socket.error, paramiko.SSHException):
                self.logger.error(traceback.format_exc())
            finally:
                pass
            if flag:
                break
            else:
                sleep(2)
                continue
        if count >= max_con:
            raise Exception("ssh connection failed")
        
        self.logger.info("Connected successfully using stateless exec_command mode.")
        return {"client": ssh, "channel": None, "stdout": "Successfully connected"}

    def close(self):
        if self.__channel:
            self.__channel.shutdown(2)
            self.__channel.close()
        if self.__sshClient:
            self.__sshClient.close()

    def cmd(self, cmd_spec, timeout=60):
        """
        send command function (Stateless refactor using exec_command)
        """
        result = {RESULT_CODE: -1, STD_ERR: None, STD_OUT: ""}

        # Extract command components
        cmd_parts = cmd_spec.get("command", [])
        if not cmd_parts:
            return result

        cmdstr = " ".join(cmd_parts)
        directory = cmd_spec.get("directory", "")

        # Combine command with the desired directory
        if directory:
            cmdstr = f"cd {directory} && {cmdstr}"

        self.logger.info(f"Sending ssh cmd(ip={self.ip}): {cmdstr}")

        max_retries = 3
        for retry in range(max_retries):
            try:
                if not self.__sshClient:
                    self.logger.info("SSH client not found, establishing connection...")
                    self.login()
                    
                stdin, stdout, stderr = self.__sshClient.exec_command(cmdstr, timeout=timeout)

                # Sub-commands (inputs for interactive, rare in stateless but kept as a simple write pattern)
                inputs = cmd_spec.get(INPUT_LIST, [])
                if inputs:
                    for idx in range(0, len(inputs), 2):
                        sub_cmd = inputs[idx]
                        if not isinstance(sub_cmd, str):
                            continue
                        if not sub_cmd.endswith("\n"):
                            sub_cmd += "\n"
                        stdin.write(sub_cmd)
                        stdin.flush()

                # Decode output cleanly
                out = stdout.read().decode('utf-8', 'ignore').strip()
                err = stderr.read().decode('utf-8', 'ignore').strip()

                # Get execution status directly from channel
                rc = stdout.channel.recv_exit_status()

                if rc == 0 and not out and err:
                    out = err
                    err = ""

                if rc != 0:
                    self.logger.warning(f"cmd failed with rc={rc}. stderr: {err}")

                result[RESULT_CODE] = rc
                result[STD_OUT] = out
                if rc != 0 and err:
                    result[STD_ERR] = err

                return result

            except socket.timeout:
                self.logger.error(f"cmd timeout out after {timeout}s: {cmdstr}")
                result[STD_ERR] = "TIMEOUT"
                return result
            except Exception as e:
                self.logger.error(f"cmd exception parsing: {e}")
                if "SSH session not active" in str(e) or "not established" in str(e) or "broken pipe" in str(e).lower():
                    if retry < max_retries - 1:
                        self.logger.info(f"SSH session broken, reconnecting... (attempt {retry+1}/{max_retries})")
                        self.close()
                        self.__sshClient = None
                        self.__channel = None
                        time.sleep(2)
                        continue
                result[STD_ERR] = str(e)
                return result
