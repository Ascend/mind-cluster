#  Copyright (C)  2024. Huawei Technologies Co., Ltd. All rights reserved.
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
