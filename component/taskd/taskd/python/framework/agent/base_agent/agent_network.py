#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright 2025. Huawei Technologies Co.,Ltd. All rights reserved.
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
import ctypes
import json
import os
import threading
import time
import queue
import uuid

from dataclasses import asdict
from taskd.python.utils.log import run_log
from taskd.python.cython_api import cython_api
from taskd.python.framework.common.type import MsgBody, MessageInfo, Position, DEFAULT_BIZTYPE
from taskd.python.toolkit.constants.constants import SEND_RETRY_TIMES
from taskd.python.toolkit.constants import constants


class AgentMessageManager():
    """
    AgentMessageManager transfers message between agent and taskd manager.
    """
    instance = None

    def __new__(cls, *args, **kwargs):
        if not cls.instance:
            cls.instance = super().__new__(cls)
        return cls.instance

    def __init__(self, network_config, msg_queue, logger):
        if cython_api.lib is None:
            run_log.error("the libtaskd.so has not been loaded")
            raise Exception("the libtaskd.so has not been loaded")
        if msg_queue is None:
            run_log.error("msg_queue is None")
            raise Exception("msg_queue is None")
        self.lib = cython_api.lib
        self.rank = int(network_config.pos.server_rank)
        self.msg_queue = msg_queue
        self._network_instance = None
        self._init_network(network_config, logger)

    def register(self, rank: str):
        """
        Register agent to taskd manager.
        """
        dst = Position(
            role="Mgr",
            server_rank="0",
            process_rank="-1"
        )
        msg_body = MsgBody(
            msg_type="REGISTER",
            code=0,
            message="",
            extension={}
        )
        body_json = json.dumps(asdict(msg_body))
        msg = MessageInfo(
            uuid=str(uuid.uuid4()),
            biz_type=DEFAULT_BIZTYPE,
            dst=dst,
            body=body_json
        )
        run_log.info(f"agent register: {msg}")
        self.send_message(msg)

    def send_message(self, message: MessageInfo, code: int = 0):
        """
        Send message to taskd manager.
        """
        msg_json = json.dumps(asdict(message)).encode('utf-8')
        self.lib.DestroyNetTool.argtypes = [ctypes.c_void_p]
        send_times = 0
        self.lib.SyncSendMessage.argtypes = [ctypes.c_void_p, ctypes.c_char_p]
        self.lib.SyncSendMessage.restype = ctypes.c_int
        while True:
            if send_times >= SEND_RETRY_TIMES:
                run_log.error(f"agent send message failed, msg: {message.uuid}")
                break
            result = self.lib.SyncSendMessage(self._network_instance, msg_json)
            if result == 0:
                if code != 0:
                    run_log.info(f"agent send message success, msg: {message.uuid} code: {code}")
                if code == constants.EXITAGENTCODE:
                    run_log.info(f"agent send exit message, msg: {message.uuid}")
                    self.lib.DestroyNetTool(self._network_instance)
                break
            run_log.warning(f"agent send message failed, result: {result}")
            send_times += 1
            time.sleep(1)

    def receive_message(self):
        """
        Receive message from taskd manager.
        """
        self.lib.ReceiveMessageC.argtypes = [ctypes.c_void_p]
        self.lib.ReceiveMessageC.restype = ctypes.c_void_p
        self.lib.FreeCMemory.argtypes = [ctypes.c_void_p]
        self.lib.FreeCMemory.restype = None
        while True:
            msg_ptr = self.lib.ReceiveMessageC(self._network_instance)
            if msg_ptr is None:
                continue
            msg_str = ctypes.cast(msg_ptr, ctypes.c_char_p).value.decode('utf-8')
            run_log.info(f"agent recv message: {msg_str}")
            msg = self._parse_msg(msg_str)
            if msg is None:
                self.lib.FreeCMemory(msg_ptr)
                continue
            self.msg_queue.put(msg)
            self.lib.FreeCMemory(msg_ptr)

    def get_network_instance(self):
        """
        Get network instance.
        """
        return self._network_instance

    def _parse_msg(self, msg_json) -> MsgBody:
        """
        Parse message from taskd manager.
        """
        try:
            msg_json = json.loads(msg_json)
            msg_body_json = msg_json["body"]
            msg_body = json.loads(msg_body_json)
            msg = MsgBody(
                msg_type=msg_body["msg_type"],
                code=msg_body["code"],
                message=msg_body["message"],
                extension=msg_body["extension"]
            )
        except Exception as e:
            run_log.error(f"agent parse message failed, reason: {e}")
            return None
        run_log.info(f"agent parse message body: {msg}")
        return msg

    def _init_network(self, network_config, logger):
        """
        Initialize network.
        """ 
        run_log.info(f"network config: {network_config}")
        config_json = json.dumps(asdict(network_config)).encode('utf-8')

        init_network_func = self.lib.InitNetwork
        if init_network_func is None:
            run_log.error("init_network: func InitNetwork has not been loaded from libtaskd.so")
            raise Exception("init_network: func InitNetwork has not been loaded from libtaskd.so")
        init_network_func.argtypes = [ctypes.c_char_p, ctypes.c_void_p]
        init_network_func.restype = ctypes.c_void_p
        self._network_instance = init_network_func(config_json, logger)
        if self._network_instance is None:
            run_log.error("init_network_func failed")
            raise Exception("init_network_func failed")


def init_network_client(network_config, msg_queue, logger):
    start_process = threading.Thread(target=init_message_manager, args=(network_config, msg_queue, logger))
    start_process.daemon = True
    start_process.start()


def init_message_manager(network_config, msg_queue, logger):
    """
    Initialize message manager.
    """
    if network_config is None or msg_queue is None or logger is None:
        run_log.error(f"network_config/msg_queue/logger  is None, \
        network_config:{network_config}, msg_queue:{msg_queue}, logger:{logger}")
        raise Exception("network_config is None")
    msg_manager = AgentMessageManager(network_config, msg_queue, logger)

    time_use = 0
    while True:
        if time_use > 60:
            run_log.error("init message manager failed")
            return
        if msg_manager.get_network_instance() is not None:
            run_log.info("init message manager success")
            break
        time.sleep(1)
        time_use += 1
        run_log.info("wait get_network_instance")

    msg_manager.register(network_config.pos.server_rank)
    msg_manager.receive_message()


def get_message_manager() -> AgentMessageManager:
    """
    Get message manager instance.
    """
    return AgentMessageManager.instance


def network_send_message(msg: MessageInfo, code: int = 0):
    """
    Send message to taskd manager.
    """
    msg_manager = get_message_manager()
    if msg_manager.get_network_instance() is None:
        run_log.warning("network instance is None")
        return
    msg_manager.send_message(msg, code)


def get_msg_network_instance():
    """
    Get network instance.
    """
    msg_manager = get_message_manager()
    if msg_manager is None:
        return None
    return msg_manager.get_network_instance()
