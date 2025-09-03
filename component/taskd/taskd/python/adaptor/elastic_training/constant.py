#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.

class send_or_recieve_rank_repair_args:
    def __init__(self, src_ranks=None, dest_ranks=None, optim_idxs=None, step=None, rank_list=None):
        self.src_ranks = src_ranks
        self.dest_ranks = dest_ranks
        self.optim_idxs = optim_idxs
        self.step = step
        self.rank_list = rank_list


class build_model_and_optimizer_result:
    def __init__(self, model=None, optimizer=None, lr_scheduler=None, config=None):
        self.model = model
        self.optimizer = optimizer
        self.lr_scheduler = lr_scheduler
        self.config = config