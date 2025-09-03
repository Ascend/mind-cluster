#!/usr/bin/python3
# coding: utf-8
# Copyright (c) Huawei Technologies Co., Ltd. 2025. All rights reserved.

def tft_init_controller(*args):
    return

def tft_start_controller(*args):
    return

def tft_notify_controller_dump(*args) -> int:
    return 0

def tft_notify_controller_stop_train(*args) -> int:
    return 0

def tft_notify_controller_on_global_rank(*args) -> int:
    return 0

def tft_notify_controller_change_strategy(*args) -> int:
    return 0

def tft_register_mindx_callback(*args):
    return

def tft_destroy_controller(*args):
    return

def tft_query_high_availability_switch(*args) -> bool:
    return True


class logger:
    def info(self, *args):
        pass

    def debug(self, *args):
        pass

    def warning(self, *args):
        pass

    def error(self, *args):
        pass


class ttp_logger_class:
    def __init__(self):
        self.LOGGER = logger()


ttp_logger = ttp_logger_class()