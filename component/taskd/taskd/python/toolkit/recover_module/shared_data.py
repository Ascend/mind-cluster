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
import threading
lock = threading.Lock()


class SharedData:
    _instance = None

    def __new__(cls, *args, **kwargs):
        if not cls._instance:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self):
        self.exit_flag = False
        self.kill_flag = False

    def get_exit_flag(self):
        lock.acquire()
        flag = self.exit_flag
        lock.release()
        return flag

    def set_exit_flag(self, flag: bool):
        lock.acquire()
        self.exit_flag = flag
        lock.release()

    def get_kill_flag(self):
        lock.acquire()
        flag = self.kill_flag
        lock.release()
        return flag

    def set_kill_flag(self, flag: bool):
        lock.acquire()
        self.kill_flag = flag
        lock.release()


shared_data_inst = SharedData()
