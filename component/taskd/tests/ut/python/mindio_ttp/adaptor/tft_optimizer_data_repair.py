#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
from dataclasses import dataclass


@dataclass
class LogArgs:
    losses_reduced_ = [1, 1]
    grad_norm_ = [1, 1]
    num_zeros_in_grad_ = 0


def get_build_data_args():
    return None, None, None


def unset_memory_ckpt():
    pass


def set_load_ckpt():
    pass


def average_losses_across_microbatches(losses_reduced_):
    return [1, 1]


def get_load_ckpt():
    return False


def convert_log_args_to_tensors():
    pass


def send_log_args(dest_rank):
    pass


def convert_log_tensors_to_args():
    pass


def set_memory_ckpt():
    pass


def recv_log_args():
    pass

