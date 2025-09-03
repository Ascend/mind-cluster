#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
import array


def print_rank_0(msg):
    pass


def unwrap_model(model):
    pass


def calc_params_l2_norm():
    pass


class test_optimizer():
    @staticmethod
    def numpy():
        return array.array('i', [1, 2, 3])

    def send_optim_param_state(self, dest_rank, group, optim_idx):
        pass

    def item(self):
        pass

    def cpu(self):
        return self
