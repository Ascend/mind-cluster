#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
class TTPReplicaOptimizer():
    def __init__(self, dump):
        self.error_dump = dump
        self.save_args = {'rank': 0}
        self.ori_dp_group = None
        self.filename = ''

    def save_parameter_state_impl(self):
        pass

    def save_parameter_state(self, filename):
        self.filename = filename