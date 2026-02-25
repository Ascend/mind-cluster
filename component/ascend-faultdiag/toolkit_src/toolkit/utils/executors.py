#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2026 Huawei Technologies Co., Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ==============================================================================

import abc
import asyncio
import os
import threading
import time
from pathlib import Path
from typing import Optional, Callable, List, Tuple

import paramiko
from paramiko.ssh_exception import SSHException, NoValidConnectionsError
from scp import SCPClient

from toolkit.core.common.json_obj import JsonObj
from toolkit.utils.logger import DIAG_LOGGER


class CommandResult(JsonObj):
    """命令执行结果封装"""

    def __init__(self, cmd: str, returncode: int, stdout: str = "", stderr: str = ""):
        self.cmd = cmd
        self.returncode = returncode
        self.stdout = stdout
        self.stderr = stderr

    def __repr__(self) -> str:
        return f"CommandResult(cmd={self.cmd!r}, success={self.is_success()})"

    def is_success(self) -> bool:
        return self.returncode == 0


CALLABLE_PARAM = Optional[Callable[[CommandResult], None]]


def default_error_callback(result: CommandResult):
    msg = (f"Command {result.cmd} failed with return code {result.returncode}, stdout: {result.stdout}, "
           f"stderr: {result.stderr}")
    DIAG_LOGGER.error(msg)


class CmdTask:

    def __init__(self, cmd: str, timeout=5, timeout_once=0.2, end_sign="", on_complete: CALLABLE_PARAM = None,
                 on_failed: CALLABLE_PARAM = default_error_callback):
        self.cmd = cmd
        self.timeout = timeout
        self.timeout_once = timeout_once
        self.end_sign = end_sign
        self.on_complete = on_complete
        self.on_failed = on_failed


class AsyncExecutor(abc.ABC):
    """异步执行器抽象基类"""

    def __init__(self, host):
        self.host = host

    async def __aenter__(self):
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()
        return False

    @abc.abstractmethod
    async def run_cmd(self, cmd_task: CmdTask) -> CommandResult:
        """执行单个命令"""
        pass

    @abc.abstractmethod
    async def run_parallel(self, cmd_tasks: List[CmdTask]) -> List[CommandResult]:
        """并行执行多个命令"""
        pass

    @abc.abstractmethod
    async def download_file(self, remote_path: str, local_path: str) -> None:
        """下载文件"""
        pass

    @abc.abstractmethod
    async def upload_file(self, local_path: str, remote_path: str,
                          permissions: Optional[int] = None) -> None:
        """上传文件"""
        pass

    async def close(self) -> None:
        """关闭执行器资源"""
        pass


class AsyncCmdExecutor(AsyncExecutor):
    """本地命令异步执行器"""

    def __init__(self):
        super().__init__("localhost")

    async def run_cmd(self, cmd_task: CmdTask) -> CommandResult:
        """执行单个本地异步命令"""
        try:
            # 创建子进程执行命令
            proc = await asyncio.create_subprocess_shell(
                cmd_task.cmd,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            # 等待命令完成并获取输出（带超时）
            stdout, stderr = await asyncio.wait_for(
                proc.communicate(),
                timeout=cmd_task.timeout
            )

            res = CommandResult(
                cmd=cmd_task.cmd,
                returncode=proc.returncode,
                stdout=stdout.decode().strip(),
                stderr=stderr.decode().strip()
            )

            # 执行回调
            if res.is_success() and cmd_task.on_complete:
                cmd_task.on_complete(res)
            elif not res.is_success() and cmd_task.on_failed:
                cmd_task.on_failed(res)

            return res

        except asyncio.TimeoutError:
            # 超时处理
            proc.terminate()
            return CommandResult(
                cmd=cmd_task.cmd,
                returncode=-1,
                stdout="",
                stderr=f"Command timed out after {cmd_task.timeout} seconds"
            )
        except Exception as e:
            return CommandResult(
                cmd=cmd_task.cmd,
                returncode=-2,
                stdout="",
                stderr=f"Execution error: {str(e)}"
            )

    async def run_parallel(self, cmd_tasks: List[CmdTask]) -> List[CommandResult]:
        """并行执行多个本地命令"""
        # 创建所有异步任务
        tasks = [self.run_cmd(task) for task in cmd_tasks]

        # 并发执行并收集结果
        results = []
        for future in asyncio.as_completed(tasks):
            result = await future
            results.append(result)

        return results

    async def download_file(self, remote_path: str, local_path: str) -> None:
        """本地文件复制（模拟下载）"""
        # 实际场景可能是从网络位置下载到本地
        import shutil
        loop = asyncio.get_event_loop()

        def _copy_sync():
            # 确保目标目录存在
            local_dir = os.path.dirname(local_path)
            if local_dir and not os.path.exists(local_dir):
                os.makedirs(local_dir, exist_ok=True)
            shutil.copy2(remote_path, local_path)

        await loop.run_in_executor(None, _copy_sync)

    async def upload_file(self, local_path: str, remote_path: str,
                          permissions: Optional[int] = None) -> None:
        """本地文件复制（模拟上传）"""
        import shutil
        loop = asyncio.get_event_loop()

        def _copy_sync():
            # 确保目标目录存在
            remote_dir = os.path.dirname(remote_path)
            if remote_dir and not os.path.exists(remote_dir):
                os.makedirs(remote_dir, exist_ok=True)
            shutil.copy2(local_path, remote_path)

            # 设置权限
            if permissions is not None:
                os.chmod(remote_path, permissions)

        await loop.run_in_executor(None, _copy_sync)


class AsyncSSHExecutor(AsyncExecutor):
    """SSH执行器（极简超时逻辑）
    - 每0.1秒尝试读取一次通道
    - 仅当无新内容且超过任务设置的超时时间后终止
    - 不添加额外超时错误信息
    """

    def __init__(self, host: str, port: int = 22, username: str = "root", password: Optional[str] = None,
                 private_key: Optional[str] = None, passphrase: Optional[str] = None, timeout: float = 10.0,
                 look_for_keys: bool = True, allow_agent: bool = True):
        """初始化SSH连接参数"""
        super().__init__(host)
        self.host = host
        self.port = port
        self.username = username
        self.password = password
        self.timeout = timeout  # 连接超时时间
        self.look_for_keys = look_for_keys
        self.allow_agent = allow_agent
        # 私钥处理
        self.private_key_obj = self._load_private_key(private_key, passphrase)
        # 单连接模式的客户端和shell通道
        self.ssh_client: Optional[paramiko.SSHClient] = None
        self.shell_channel: Optional[paramiko.Channel] = None
        self._lock = threading.Lock()
        self.last_prompt = ""  # 记录上次命令提示符

    @staticmethod
    def _load_private_key(private_key, passphrase):
        """加载私钥（Ed25519/RSA兼容）"""
        private_key_obj = None
        if not private_key:
            return private_key_obj
        private_key_path = str(Path(private_key).expanduser().resolve())
        if not private_key_path or not os.path.exists(private_key_path):
            return private_key_obj
        try:
            private_key_obj = paramiko.Ed25519Key.from_private_key_file(private_key_path, passphrase)
        except (SSHException, NotImplementedError):
            try:
                private_key_obj = paramiko.RSAKey.from_private_key_file(private_key_path, passphrase)
            except SSHException as err:
                DIAG_LOGGER.warning(f"Ed25519密钥和RSA密钥加载失败[{private_key}]：{str(err)}")
        return private_key_obj

    def ensure_shell_session(self):
        """确保shell会话已建立并可用"""
        with self._lock:
            if not self.ssh_client or not self.ssh_client.get_transport().is_active():
                self.ssh_client = self._create_ssh_client()
                self.shell_channel = self._create_shell_channel(self.ssh_client)
            elif not self.shell_channel or not self.shell_channel.active:
                self.shell_channel = self._create_shell_channel(self.ssh_client)

    async def run_cmd(self, cmd_task: CmdTask) -> CommandResult:
        """执行单个命令（使用任务的超时参数作为无新数据等待时间）"""
        loop = asyncio.get_event_loop()

        def _run_sync():
            return self._shell_execute(cmd_task)

        start_time = time.time()
        exit_code, stdout, stderr = await loop.run_in_executor(None, _run_sync)
        DIAG_LOGGER.info(f"Host: {self.host}, cmd: {cmd_task.cmd} cost {time.time() - start_time} seconds")
        result = CommandResult(
            cmd=cmd_task.cmd,
            returncode=exit_code,
            stdout=stdout or "",
            stderr=stderr or ""
        )

        # 执行回调
        if result.is_success() and cmd_task.on_complete:
            cmd_task.on_complete(result)
        elif not result.is_success() and cmd_task.on_failed:
            cmd_task.on_failed(result)

        return result

    async def run_parallel(self, cmd_tasks: List[CmdTask]) -> List[CommandResult]:
        """多连接并行执行命令（每个命令使用自身的超时参数）"""

        async def _parallel_task(task: CmdTask):
            ssh = self._create_ssh_client()
            channel = self._create_shell_channel(ssh)
            prompt = self.last_prompt

            try:
                def _sync_exec():
                    all_stdout = []
                    all_stderr = []
                    returncode = -1
                    last_data_time = time.time()

                    channel.send(f"{task.cmd}\n".encode())

                    while True:
                        # 读取数据
                        stdout, stderr, has_new_data = self._read_channel_output(channel)

                        if has_new_data:
                            last_data_time = time.time()
                            if stdout:
                                all_stdout.append(stdout)
                            if stderr:
                                all_stderr.append(stderr)

                            # 检查命令完成
                            full_output = ''.join(all_stdout)
                            if prompt in full_output and full_output.endswith(prompt):
                                if channel.exit_status_ready():
                                    returncode = channel.recv_exit_status()
                                break
                        else:
                            # 无新数据时检查超时
                            if time.time() - last_data_time > task.timeout:
                                break  # 超时后直接终止，不添加错误信息

                        # 每0.1秒尝试一次
                        time.sleep(0.1)

                    # 清理输出
                    full_output = ''.join(all_stdout)
                    cleaned_output = full_output.replace(f"{task.cmd}\n", "", 1).rsplit(prompt, 1)[0].strip()
                    return returncode, cleaned_output, ''.join(all_stderr).strip()

                loop = asyncio.get_event_loop()
                exit_code, stdout, stderr = await loop.run_in_executor(None, _sync_exec)

                return CommandResult(
                    cmd=task.cmd,
                    returncode=exit_code,
                    stdout=stdout,
                    stderr=stderr
                )

            except Exception as e:
                return CommandResult(
                    cmd=task.cmd,
                    returncode=-1,
                    stdout="",
                    stderr=f"并行执行错误: {str(e)}"
                )
            finally:
                channel.close()
                ssh.close()

        tasks = [_parallel_task(task) for task in cmd_tasks]
        return await asyncio.gather(*tasks)

    async def upload_file(self, local_path: str, remote_path: str,
                          permissions: Optional[int] = None) -> None:
        """
        异步上传文件（SCP模式）
        :param local_path: 本地文件路径
        :param remote_path: 远程文件路径
        :param permissions: 远程文件权限（如 0o644）
        """
        loop = asyncio.get_event_loop()

        def _upload_sync():
            with self._lock:  # 同步锁保护SSH连接
                # 确保SSH连接有效
                if not self.ssh_client or not self.ssh_client.get_transport().is_active():
                    self.ssh_client = self._create_ssh_client()

                # 使用SCPClient上传
                with SCPClient(self.ssh_client.get_transport()) as scp:
                    # 上传文件（preserve_times=True 保留文件时间戳）
                    scp.put(
                        local_path,
                        remote_path=remote_path,
                        preserve_times=True
                    )

                # 设置权限（如果指定）
                if permissions is not None:
                    with self.ssh_client.open_sftp() as sftp:  # 用SFTP设置权限（SCP不直接支持）
                        sftp.chmod(remote_path, permissions)

        # 在 executor 中执行同步操作
        await loop.run_in_executor(None, _upload_sync)

    async def download_file(self, remote_path: str, local_path: str) -> None:
        """
        异步下载文件（SCP模式）
        :param remote_path: 远程文件路径
        :param local_path: 本地文件路径
        """
        loop = asyncio.get_event_loop()

        def _download_sync():
            with self._lock:  # 同步锁保护SSH连接
                # 确保SSH连接有效
                if not self.ssh_client or not self.ssh_client.get_transport().is_active():
                    self.ssh_client = self._create_ssh_client()

                # 确保本地目录存在
                local_dir = os.path.dirname(local_path)
                if local_dir and not os.path.exists(local_dir):
                    os.makedirs(local_dir, exist_ok=True)

                # 使用SCPClient下载
                with SCPClient(self.ssh_client.get_transport()) as scp:
                    scp.get(
                        remote_path=remote_path,
                        local_path=local_path,
                        preserve_times=True  # 保留文件时间戳
                    )

        # 在 executor 中执行同步操作
        await loop.run_in_executor(None, _download_sync)

    async def close(self) -> None:
        loop = asyncio.get_event_loop()

        def _close_sync():
            with self._lock:
                if self.shell_channel and self.shell_channel.active:
                    try:
                        self.shell_channel.close()
                    except SSHException:
                        pass
                    self.shell_channel = None

                if self.ssh_client:
                    try:
                        self.ssh_client.close()
                    except SSHException:
                        pass
                    self.ssh_client = None

        await loop.run_in_executor(None, _close_sync)

    def _create_ssh_client(self) -> paramiko.SSHClient:
        """创建新的SSH客户端连接"""
        # 优先尝试密钥登录
        ssh = self._connect_with_key()
        if ssh:
            return ssh
        ssh = self._connect_with_password()
        if ssh:
            return ssh
        DIAG_LOGGER.error(f"{self.username}@{self.host}:{self.port} 密钥和密码方式都登录失败")
        return None

    def _connect_with_key(self) -> paramiko.SSHClient:
        if not self.private_key_obj:
            return None
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        try:
            ssh.connect(
                hostname=self.host,
                port=self.port,
                username=self.username,
                pkey=self.private_key_obj,
                timeout=self.timeout,
                look_for_keys=self.look_for_keys,
                allow_agent=self.allow_agent
            )
            return ssh
        except NoValidConnectionsError:
            DIAG_LOGGER.warning(f"密钥方式登录：无法连接到 {self.host}:{self.port}")
        except SSHException as e:
            DIAG_LOGGER.warning(f"密钥方式登录：SSH连接 {self.host}:{self.port} 失败: {str(e)}")
        return None

    def _connect_with_password(self) -> paramiko.SSHClient:
        ssh = paramiko.SSHClient()
        ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())
        try:
            ssh.connect(
                hostname=self.host,
                port=self.port,
                username=self.username,
                password=self.password,
                timeout=self.timeout,
                look_for_keys=self.look_for_keys,
                allow_agent=self.allow_agent
            )
            return ssh
        except NoValidConnectionsError:
            DIAG_LOGGER.warning(f"密码方式登录：无法连接到 {self.host}:{self.port}")
        except SSHException as e:
            DIAG_LOGGER.warning(f"密码方式登录：SSH连接 {self.host}:{self.port} 失败: {str(e)}")
        return None

    def _create_shell_channel(self, ssh: paramiko.SSHClient) -> paramiko.Channel:
        """创建交互式shell通道"""
        if not ssh:
            return None
        channel = ssh.invoke_shell()
        channel.settimeout(0.1)  # 短超时便于频繁检查
        # 读取初始提示符
        time.sleep(3)
        self._read_initial_prompt(channel)
        return channel

    def _read_initial_prompt(self, channel: paramiko.Channel) -> None:
        """读取初始命令提示符"""
        output = []
        while True:
            if channel.recv_ready():
                output.append(channel.recv(4096).decode('utf-8', errors='replace'))
            else:
                break
        if output:
            self.last_prompt = ''.join(output).splitlines()[-1]

    def _read_channel_output(self, channel: paramiko.Channel) -> Tuple[str, str, bool]:
        """读取通道输出，返回是否有新数据"""
        stdout = []
        stderr = []
        has_new_data = False

        if channel.recv_ready():
            stdout.append(channel.recv(4096).decode('utf-8', errors='replace'))
            has_new_data = True
        if channel.recv_stderr_ready():
            stderr.append(channel.recv_stderr(4096).decode('utf-8', errors='replace'))
            has_new_data = True

        return ''.join(stdout), ''.join(stderr), has_new_data

    def _shell_execute(self, cmd_task: CmdTask) -> Tuple[int, str, str]:
        """极简的shell执行逻辑
        - 每0.1秒尝试读取一次通道
        - 仅当无新内容且超过超时时间后终止
        - 不添加额外超时错误信息
        """
        self.ensure_shell_session()
        if not self.shell_channel:
            return 0, "", "shell会话建立失败"
        all_stdout = []
        all_stderr = []
        last_data_time = time.time()  # 最后一次收到数据的时间

        # 发送命令
        self.shell_channel.send(f"{cmd_task.cmd}\n".encode())
        end_sign = cmd_task.end_sign or self.last_prompt
        full_output = ""
        while True:
            # 读取通道数据
            stdout, stderr, has_new_data = self._read_channel_output(self.shell_channel)

            if has_new_data:
                # 有新数据，更新时间并保存输出
                last_data_time = time.time()
                if stdout:
                    all_stdout.append(stdout)
                if stderr:
                    all_stderr.append(stderr)

                # 检查命令是否完成（通过提示符判断）
                full_output = ''.join(all_stdout)
                if end_sign and end_sign in full_output and full_output.endswith(end_sign):
                    # 更新最后提示符
                    self.last_prompt = full_output.splitlines()[-1]
                    break
            else:
                # 无新数据，检查是否超过超时时间
                if time.time() - last_data_time > cmd_task.timeout:
                    if full_output:
                        self.last_prompt = full_output.splitlines()[-1]
                    break  # 超时后直接终止，不添加错误信息

            # 每0.1秒尝试读取一次
            time.sleep(cmd_task.timeout_once)

        # 清理输出
        full_output = ''.join(all_stdout)
        cleaned_output = full_output.replace(f"{cmd_task.cmd}\n", "", 1)
        if self.last_prompt:
            cleaned_output = cleaned_output.rsplit(self.last_prompt, 1)[0].strip()
        return -int(len(all_stderr) > 0), cleaned_output, ''.join(all_stderr).strip()
