#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
def get_args():
    return args_test()


def get_timers():
    return None


class args_test():
    load = None
    iteration = None
    num_query_groups = None
    curr_iteration = None
    do_train = False
    do_valid = False
    do_test = False
    consumed_train_samples = None
    finetune = False
    optim_nums = 2
    ckpt_format = None
    num_floating_point_operations_so_far = None

    @staticmethod
    def state_dict_by_idx(optim_idx):
        return None

    def load_state_dict_memory(self, dict_memory):
        pass

    def load_state_dict(self, state_dict):
        pass

    def load_state_dict_by_idx(self, optimizer, optim_idx):
        pass

    def state_dict(self):
        pass