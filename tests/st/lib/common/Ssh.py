#!/usr/bin/env python3
# coding: utf-8
# Copyright 2026 Huawei Technologies Co., Ltd
import logging
import re
import socket
from time import sleep, time
import traceback
import string
import paramiko

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
        logging.basicConfig(level=logging.DEBUG, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
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

            if self.sshServerType == ClassSsh.SSH_SERVER_TYPE_LINUX:
                self.__cmd(self.__channel, 'PS1="\\u@#>"', waitstr='%s@#' % self.username)

                self.DEFAULT_WAIT_STR = self.username + "@#>"

                self.__recv(self.__channel, self.DEFAULT_WAIT_STR, timeout=3, noTimeout=True)
                self.RECV_TMOUT = False
            elif self.sshServerType == ClassSsh.SSH_SERVER_TYPE_WINDOWS:
                pass

            return connectInfo.get(STD_OUT)
        except Exception as e:
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
        trans = ssh.get_transport()
        trans.windows_size = 214748

        count = 0
        while count < max_con:
            count += 1
            try:
                channel = ssh.invoke_shell(width=200)
                channel.settimeout(20)
                break
            except Exception:
                self.logger.error(traceback.format_exc())
                ssh.close()
                return self.connect()
            finally:
                pass

        if count >= max_con:
            raise Exception("get channel failed")

        result = self.__recv(channel, self.DEFAULT_WAIT_STR, 600, 10)
        self.logger.info(result)
        return {"client": ssh, "channel": channel, "stdout": result}

    def close(self):
        if self.__channel:
            self.__channel.shutdown(2)
            self.__channel.close()
        if self.__sshClient:
            self.__sshClient.close()

    def cmd(self, cmd_spec, timeout=60):
        """
        send command function
        :param cmd_spec(dict) cmd_spec:
        {
          command: []
          input: []
          waitstr: []
          directory: ""
          timeout: 60
          username: ""
          password: ""
        }
        :param timeout(int) timeout: the command wait time
        :return:
          result(str): the full command output structure
        """
        if not self.__channel:
            raise Exception("call login first")

        result = {"rc": None, "stderr": None, "stdout": ""}

        waitstr = cmd_spec["waitstr"]
        if waitstr == "":
            waitstr = self.DEFAULT_WAIT_STR

        if "directory" in cmd_spec:
            ret = self.__cmd(self.__channel, "cd" + cmd_spec["directory"])
            if not self.__mergeResult(result, ret):
                self.__discardLine0(result)
                return result

        cmdstr = " ".join(cmd_spec["command"])
        self.logger.info("Sending ssh cmd(ip=%s, waitstr: %s): %s" % (self.ip, waitstr, cmdstr))
        ret = self.__cmd(self.__channel, cmdstr, waitstr, timeout)

        if not self.__mergeResult(result, ret):
            self.__discardLine0(result)
            return result

        # analyze rsult
        if INPUT_LIST not in cmd_spec:
            if self.sshServerType == ClassSsh.SSH_SERVER_TYPE_LINUX:
                ret = self.__cmd(self.__channel, "echo $?", waitstr, timeout)
                result[RESULT_CODE] = self.__parseRC(ret)
            self.__discardLine0(result)

        waitstr = self.DEFAULT_WAIT_STR
        length = len(cmd_spec[INPUT_LIST])
        tmp = 0
        while tmp < length:
            sub_cmd = cmd_spec[INPUT_LIST][tmp]
            sub_wait = cmd_spec[INPUT_LIST][tmp + 1]
            if sub_wait == "":
                sub_wait = self.DEFAULT_WAIT_STR
            self.logger.info("sending ssh sub cmd %s" % sub_cmd)
            ret = self.__cmd(self.__channel, sub_cmd, sub_wait, timeout)
            ret = self.__removeEscape(cmdstr, ret)

        if self.sshServerType == ClassSsh.SSH_SERVER_TYPE_LINUX:
            ret = self.__cmd(self.__channel, "echo $?", waitstr, timeout)
            result[RESULT_CODE] = self.__parseRC(ret)
        self.__discardLine0(result)
        return result

    def __cmd(self, channel, cmd, waitstr, timeout, sendonly=False):
        if not self.__send(channel, cmd, timeout):
            return None
        ret = self.__recv(channel, waitstr, 40000, timeout)
        return ret

    def __mergeResult(self, result, ret):
        if result.get(STD_OUT) is not None:
            result[STD_OUT] = ret
        if ret is None:
            return False
        if result[STD_OUT] == self.MSG_TOO_LONG:
            return True

        result[STD_OUT] += ret

        if len(result[STD_OUT]) >= self.RESULT_MSG_MAX_LEN:
            result[STD_OUT] = self.MSG_TOO_LONG
            return True

        return True

    def __discardLine0(self, result):
        split_str = "\x0d\x0a"
        if RESULT_CODE in result and result[RESULT_CODE] != 0:
            result[STD_ERR] = result[STD_OUT].strip()
        if STD_OUT in result and isinstance(result[STD_OUT], str):
            result[STD_OUT] = result[STD_OUT].strip()
            tmp_list = re.split("\x0d?\x0a|\x0d", result[STD_OUT])
            if len(tmp_list) > 2:
                tmp_list.pop(0)
                tmp_list.pop(len(tmp_list) - 1)
                if len(tmp_list) == 0:
                    result[STD_OUT] = ""
                else:
                    result[STD_OUT] = split_str.join(tmp_list)
            elif len(tmp_list) == 2:
                tmp_list.pop(0)
                result[STD_OUT] = split_str.join(tmp_list)

        if STD_ERR in result and isinstance(result[STD_ERR], str):
            result[STD_ERR] = result[STD_ERR].strip()
            tmp_list = re.split("\x0d?\x0a|\x0d", result[STD_ERR])
            if len(tmp_list) > 2:
                tmp_list.pop(0)
                tmp_list.pop(len(tmp_list) - 1)
                if len(tmp_list) == 0:
                    result[STD_ERR] = ""
                else:
                    result[STD_ERR] = split_str.join(tmp_list)
            elif len(tmp_list) == 2:
                tmp_list.pop(0)
                result[STD_ERR] = split_str.join(tmp_list)

    def __send(self, channel, cmd, timeout):
        nowtime = time()
        endtime = nowtime + timeout
        while nowtime < endtime:
            if not channel.send_ready:
                sleep(1)
                nowtime = time()
                continue
            try:
                channel.send(cmd + self.linesep)
                return True
            except socket.timeout:
                self.logger.error("time out err %s" % cmd)
            finally:
                pass
            sleep(1)
            nowtime = time()

    def __recv(self, channel, waitstr, nbytes, timeout, noTimeout=False):
        recv = ""
        nowtime = time()
        endtime = nowtime + timeout
        match = None
        while nowtime < endtime:
            if not channel.recv_ready:
                sleep(1)
                nowtime = time()
                continue
            try:
                strGet = channel.recv(nbytes).decode(errors='ignores')
            except socket.timeout:
                strGet = ""
            finally:
                pass
            recv += strGet
            if not waitstr:
                waitstr = self.DEFAULT_WAIT_STR
            match = re.search(r'' + str(waitstr) + '', recv)
            if match is not None:
                break
            sleep(1)
            nowtime = time()
        if match is None and not noTimeout:
            self.RECV_TMOUT = True
            self.logger.warning("TIMEOUT occurred")

        if recv == "":
            recv = None
        return recv

    def __removeEscape(self, cmdstr, ret):
        return ret

    def __parseRC(self, ret):
        if ret is None or ret == "":
            return None
        ret_list = re.split("\x0d?\x0a|\x0d", ret)
        if len(ret_list) != 3:
            return None
        if ret_list[1].isdigit():
            return int(ret_list[1])
        return None
