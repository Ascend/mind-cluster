#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# Copyright (c) 2025, Huawei Technologies Co., Ltd. All rights reserved.
from megatron.training.utils import test_optimizer


def get_args():
    return args_test()


class timer():
    def reset(self):
        pass

    def set_barrier_group(self, group):
        pass


class timers():
    _timers = {"0": timer()}

    def reset(self):
        pass

    def start(self, barrier):
        pass


def get_timers(*args, **kwargs):
    return timers()


class bucket():
    data_parallel_group = 'dp'
    data_parallel_world_size = 4
    data_parallel_rank = 0
    intra_distributed_optimizer_instance_group = None
    intra_distributed_optimizer_instance_size = 4
    intra_distributed_optimizer_instance_rank = 0


class buffer():
    data_parallel_group = 'dp'
    data_parallel_world_size = 4
    buckets = [bucket()]


class options_config():
    cga_cluster_size = 4
    max_ctas = 32
    min_ctas = 1


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
    num_steps = 1
    train_samples = None
    global_batch_size = 4
    rampup_batch_size = [1, 1, 1]
    micro_batch_size = 1
    rank = 0
    virtual_pipeline_model_parallel_size = 1
    log_params_norm = 1
    param_groups = [{'is_decoupled_lr': 1.0, 'lr': 1.0}]
    swap_attention = True
    num_experts = 1
    moe_permutation_async_comm = True
    distributed_backend = 'nccl'
    expert_model_parallel_size = 1
    context_parallel_size = 1
    data_parallel_size = 4
    buffers = [buffer()]
    param_to_bucket_group = {"0": bucket()}
    distributed_timeout_minutes = 1
    nccl_communicator_config_path = None
    world_size = 16
    tensor_model_parallel_size = 4
    pipeline_model_parallel_size = 1
    use_distributed_optimizer = True
    use_nd_matmul = False
    tp_2d = True
    tp_x = 4
    tp_y = 4
    chained_optimizers = [test_optimizer(), test_optimizer()]
    config = options_config()

    @staticmethod
    def state_dict_by_idx(optim_idx):
        return None

    @staticmethod
    def step(global_batch_size):
        return None

    @staticmethod
    def sync_gather_all_model_params(force_sync=False):
        pass

    @staticmethod
    def get_loss_scale():
        return test_optimizer()

    @staticmethod
    def _copy_main_params_to_model_params():
        pass

    def load_state_dict_memory(self, dict_memory):
        pass

    def load_state_dict(self, state_dict):
        pass

    def load_state_dict_by_idx(self, optimizer, optim_idx):
        pass

    def state_dict(self):
        pass